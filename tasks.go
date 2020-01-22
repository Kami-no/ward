package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func trackOpenedMR(cfg config) {
	git := gitlab.NewClient(nil, cfg.GToken)
	_ = git.SetBaseURL(cfg.GURL)

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

	mrs, _, _ := git.MergeRequests.ListProjectMergeRequests(cfg.GProject, mrs_opts)

	for _, mr := range mrs {
		vip_f := 0
		vip_b := 0

		state_good := 0
		state_bad := 0
		state_down := 0

		awards, _, _ := git.AwardEmoji.ListMergeRequestAwardEmoji(mr.ProjectID, mr.IID, &gitlab.ListAwardEmojiOptions{})

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

func trackMergedMR(cfg config) {
	var email []string
	var ownersEmail []string

	git := gitlab.NewClient(nil, cfg.GToken)
	_ = git.SetBaseURL(cfg.GURL)

	mrs_opts := &gitlab.ListProjectMergeRequestsOptions{
		State:   gitlab.String("merged"),
		OrderBy: gitlab.String("updated_at"),
		Scope:   gitlab.String("all"),
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	mrs, _, _ := git.MergeRequests.ListProjectMergeRequests(cfg.GProject, mrs_opts)

	for _, mr := range mrs {
		vip_f := 0
		vip_b := 0

		state_down := 0
		state_fail := 0

		awards, _, _ := git.AwardEmoji.ListMergeRequestAwardEmoji(mr.ProjectID, mr.IID, &gitlab.ListAwardEmojiOptions{})

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
			email = ldapMail(cfg, mr.MergedBy.Username)
			mailMergeman(cfg, email[0], mr.WebURL, mr.IID)

			for _, i := range cfg.VBackend {
				ownersEmail = append(ownersEmail, ldapMail(cfg, i)...)
			}
			for _, i := range cfg.VFrontend {
				ownersEmail = append(ownersEmail, ldapMail(cfg, i)...)
			}
			mailMaintainers(cfg, ownersEmail, mr.WebURL, mr.IID)

			log.Printf("MR %v is merged and has failed CC !!!", mr.IID)

			award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MFail}
			_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
		}
	}
}
