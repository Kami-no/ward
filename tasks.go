package main

import (
	"fmt"
	"log"
	"time"

	"github.com/xanzy/go-gitlab"
)

func detectOpenedMR(cfg config) {
	git := gitlab.NewClient(nil, cfg.GToken)
	if err := git.SetBaseURL(cfg.GURL); err != nil {
		log.Printf("Failed to setup GitLab: %v", err)
		return
	}

	mrs_opts := &gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("opened"),
		OrderBy: gitlab.String("updated_at"),
		Scope:   gitlab.String("all"),
		WIP:     gitlab.String("no"),
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	mrs, _, err := git.MergeRequests.ListProjectMergeRequests(cfg.GProject, mrs_opts)
	if err != nil {
		log.Printf("Failed to list Merge Requests: %v", err)
		return
	}

	for _, mr := range mrs {
		vip_f := 0
		vip_b := 0

		state_good := 0
		state_bad := 0
		state_down := 0

		awards, _, err := git.AwardEmoji.ListMergeRequestAwardEmoji(mr.ProjectID, mr.IID, &gitlab.ListAwardEmojiOptions{})
		if err != nil {
			log.Printf("Failed to list MR awards: %v", err)
			break
		}

	Loop:
		for _, award := range awards {
			if award.User.Username != mr.Author.Username {
				switch award.Name {
				case cfg.MUp:
					if vip_f == 0 {
						if contains(cfg.VFrontend, award.User.Username) {
							vip_f = 1
						}
					}
					if vip_b == 0 {
						if contains(cfg.VBackend, award.User.Username) {
							vip_b = 1
						}
					}
				case cfg.MDown:
					state_down = 1
					break Loop
				}
			}

			if award.User.Username == cfg.GUser {
				switch award.Name {
				case cfg.MGood:
					state_good = award.ID
				case cfg.MBad:
					state_bad = award.ID
				}
			}
		}

		switch {
		case vip_f+vip_b != 2, state_down == 1:
			if state_good != 0 {
				_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, state_good)
			}
			if state_bad == 0 {
				award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MBad}
				_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
				log.Printf("MR %v is not ready.", mr.IID)
			}
		default:
			if state_bad != 0 {
				_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, state_bad)
			}
			if state_good == 0 {
				award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MGood}
				_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
				log.Printf("MR %v is ready.", mr.IID)
			}
		}
	}
}

func detectMergedMR(cfg config) {
	var users []string
	var emails []string
	var ownersEmail []string
	var subj string
	var msg string

	git := gitlab.NewClient(nil, cfg.GToken)
	if err := git.SetBaseURL(cfg.GURL); err != nil {
		log.Printf("Failed to setup GitLab: %v", err)
		return
	}

	mrs_opts := &gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("merged"),
		OrderBy: gitlab.String("updated_at"),
		Scope:   gitlab.String("all"),
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	mrs, _, err := git.MergeRequests.ListProjectMergeRequests(cfg.GProject, mrs_opts)
	if err != nil {
		log.Printf("Failed to list Merge Requests: %v", err)
		return
	}

	for _, mr := range mrs {
		vip_f := 0
		vip_b := 0

		state_down := 0
		state_fail := 0

		awards, _, err := git.AwardEmoji.ListMergeRequestAwardEmoji(mr.ProjectID, mr.IID, &gitlab.ListAwardEmojiOptions{})
		if err != nil {
			log.Printf("Failed to list MR awards: %v", err)
			break
		}

		for _, award := range awards {
			if award.User.Username != mr.Author.Username {
				switch award.Name {
				case cfg.MUp:
					if vip_f == 0 {
						if contains(cfg.VFrontend, award.User.Username) {
							vip_f = 1
						}
					}
					if vip_b == 0 {
						if contains(cfg.VBackend, award.User.Username) {
							vip_b = 1
						}
					}
				case cfg.MDown:
					state_down = 1
				}
			}

			if award.User.Username == cfg.GUser {
				switch award.Name {
				case cfg.MGood:
					_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award.ID)
				case cfg.MBad:
					_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award.ID)
				case cfg.MFail:
					state_fail = award.ID
				}
			}
		}

		switch {
		case state_fail != 0:
		case vip_f+vip_b != 2, state_down == 1:
			users = append(users, mr.MergedBy.Username)
			emails = ldapMail(cfg, users)
			subj = "Code of Conduct failure incident"
			msg = fmt.Sprintf("Hello,"+
				"<p>By merging <a href='%v'>Merge Request #%v</a> without 2 qualified approves"+
				" or negative review you've failed repository's Code of Conduct.</p>"+
				"<p>This incident will be reported.</p>", mr.WebURL, mr.IID)

			if err := mailSend(cfg, emails, subj, msg); err != nil {
				log.Printf("Failed to send mail: %v", err)
				break
			}

			ownersUsers := cfg.VBackend
			ownersUsers = append(ownersUsers, cfg.VFrontend...)

			ownersEmail = ldapMail(cfg, ownersUsers)

			subj = fmt.Sprintf("MR %v has failed requirements!", mr.IID)
			msg = fmt.Sprintf(
				"<p><a href='%v'>Merge Request #%v</a> does not meet requirements but it was merged!</p>",
				mr.WebURL, mr.IID)

			if err := mailSend(cfg, ownersEmail, subj, msg); err != nil {
				log.Printf("Failed to send mail: %v", err)
				break
			}

			log.Printf("MR %v is merged and has failed CC !!!", mr.IID)

			award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MFail}
			_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
		}
	}
}

func detectDeadBrunches(cfg config) {
	now := time.Now()

	git := gitlab.NewClient(nil, cfg.GToken)
	if err := git.SetBaseURL(cfg.GURL); err != nil {
		log.Printf("Failed to setup GitLab: %v", err)
		return
	}

	branches_opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	for {
		branches, response, err := git.Branches.ListBranches(cfg.GProject, branches_opts)
		if err != nil {
			log.Printf("Failed to list branches: %v", err)
			break
		}

		for _, branch := range branches {
			updated := *branch.Commit.AuthoredDate
			if now.Sub(updated).Hours() >= 7*24 {
				var rcpt []string
				rcpt = append(rcpt, branch.Commit.AuthorEmail)
				url := fmt.Sprintf("%v/-/branches/all?utf8=✓&search=%v", cfg.GPUrl, branch.Name)
				subj := fmt.Sprint("Dead branch detected!")
				msg := fmt.Sprintf(
					"<p>I see dead branches: <a href='%v'>%v</a></p>"+
						"<p>If you don't need it anymore, you should delete it.</p>",
					url, branch.Name)

				if err := mailSend(cfg, rcpt, subj, msg); err != nil {
					log.Printf("Failed to send mail: %v", err)
					break
				}
				log.Printf("Branch is too old: %v %v %v", branch.Name, branch.Commit.AuthorEmail, updated)
			}
		}

		if response.CurrentPage >= response.TotalPages {
			break
		}
		branches_opts.Page = response.NextPage
	}
}
