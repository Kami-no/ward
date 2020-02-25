package main

import (
	"fmt"
	"log"

	"github.com/xanzy/go-gitlab"
)

type MrProject struct {
	Name string
	Path string
	MR   map[int]MergeRequest
}

type MergeRequest struct {
	Name     string
	Path     string
	MergedBy string
	Awards   struct {
		Like         bool
		Dislike      bool
		Ready        int
		NotReady     int
		NonCompliant int
	}
}

type mrAction struct {
	pid      int
	mid      int
	aid      int
	award    string
	mergedBy string
	path     string
	state    bool
}

func checkPrjRequests(cfg config, projects []*Project, list string) (map[int]MrProject, error) {
	var mrs_opts *gitlab.ListProjectMergeRequestsOptions
	MrProjects := make(map[int]MrProject)

	git, err := gitlab.NewBasicAuthClient(
		nil, cfg.Endpoints.GitLab, cfg.Credentials.User, cfg.Credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to GitLab: %v", err)
	}

	switch list {
	case "opened":
		mrs_opts = &gitlab.ListProjectMergeRequestsOptions{
			State:   gitlab.String("opened"),
			OrderBy: gitlab.String("updated_at"),
			Scope:   gitlab.String("all"),
			WIP:     gitlab.String("no"),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    1,
			},
		}
	case "merged":
		mrs_opts = &gitlab.ListProjectMergeRequestsOptions{
			State:   gitlab.String("merged"),
			OrderBy: gitlab.String("updated_at"),
			Scope:   gitlab.String("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    1,
			},
		}
	default:
		mrs_opts = &gitlab.ListProjectMergeRequestsOptions{
			OrderBy: gitlab.String("updated_at"),
			Scope:   gitlab.String("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    1,
			},
		}
	}

	// Process projects
	for _, project := range projects {
		var MrPrj MrProject
		var consensus int

		if len(project.Teams) < 2 {
			consensus = 2
		} else {
			consensus = 1
		}

		mrs, _, err := git.MergeRequests.ListProjectMergeRequests(project.ID, mrs_opts)
		if err != nil {
			log.Printf("Failed to list Merge Requests for %v: %v", project.ID, err)
			break
		}

		// Process Merge Requests
		for _, mr := range mrs {
			var MRequest MergeRequest
			likes := make(map[string]int)

			awards, _, err := git.AwardEmoji.ListMergeRequestAwardEmoji(project.ID, mr.IID, &gitlab.ListAwardEmojiOptions{})
			if err != nil {
				log.Printf("Failed to list MR awards: %v", err)
				break
			}

			// Process awards
			for _, award := range awards {
				// Check group awards
				if award.User.Username != mr.Author.Username {
					switch award.Name {
					case cfg.Awards.Like:
						for team, members := range project.Teams {
							if likes[team] < consensus {
								if contains(members, award.User.Username) {
									likes[team]++
								}
							}
						}
					case cfg.Awards.Dislike:
						MRequest.Awards.Dislike = true
					}
				}

				// Check service awards
				if award.User.Username == cfg.Credentials.User {
					switch award.Name {
					case cfg.Awards.Ready:
						MRequest.Awards.Ready = award.ID
					case cfg.Awards.NotReady:
						MRequest.Awards.NotReady = award.ID
					case cfg.Awards.NonCompliant:
						MRequest.Awards.NonCompliant = award.ID
						MRequest.MergedBy = mr.MergedBy.Username
						MRequest.Path = mr.WebURL
					}
				}
			}

			// Deside if MR meets Likes requirement
			mrLike := true
			for tid := range project.Teams {
				if mrLike {
					if v, found := likes[tid]; found {
						if v < consensus {
							mrLike = false
							break
						}
					} else {
						mrLike = false
						break
					}
				}
			}
			MRequest.Awards.Like = mrLike

			if MrPrj.MR == nil {
				MrPrj.MR = make(map[int]MergeRequest)
			}
			MrPrj.MR[mr.IID] = MRequest
		}

		MrProjects[project.ID] = MrPrj
	}

	return MrProjects, nil
}

func evalOpenedRequests(MRProjects map[int]MrProject) []mrAction {
	var actions []mrAction

	for pid, project := range MRProjects {
		for mid, mr := range project.MR {
			if mr.Awards.Like && !mr.Awards.Dislike {
				if mr.Awards.NotReady != 0 {
					action := mrAction{
						pid:   pid,
						mid:   mid,
						aid:   mr.Awards.NotReady,
						award: "notready",
						state: false}
					actions = append(actions, action)
				}
				if mr.Awards.Ready == 0 {
					action := mrAction{
						pid:   pid,
						mid:   mid,
						aid:   mr.Awards.Ready,
						award: "ready",
						state: true}
					actions = append(actions, action)
				}
			} else {
				if mr.Awards.Ready != 0 {
					action := mrAction{
						pid:   pid,
						mid:   mid,
						aid:   mr.Awards.Ready,
						award: "ready",
						state: false}
					actions = append(actions, action)
				}
				if mr.Awards.NotReady == 0 {
					action := mrAction{
						pid:   pid,
						mid:   mid,
						aid:   mr.Awards.NotReady,
						award: "notready",
						state: true}
					actions = append(actions, action)
				}
			}

			if mr.Awards.NonCompliant != 0 {
				action := mrAction{
					pid:   pid,
					mid:   mid,
					aid:   mr.Awards.NonCompliant,
					award: "nc",
					state: false}
				actions = append(actions, action)
			}
		}
	}

	return actions
}

func evalMergedRequests(MRProjects map[int]MrProject) []mrAction {
	var actions []mrAction

	for pid, project := range MRProjects {
		for mid, mr := range project.MR {
			if mr.Awards.Dislike {
				if mr.Awards.NonCompliant == 0 {
					action := mrAction{
						pid:      pid,
						mid:      mid,
						aid:      mr.Awards.NonCompliant,
						award:    "nc",
						mergedBy: mr.MergedBy,
						path:     mr.Path,
						state:    true}
					actions = append(actions, action)
				}
			} else if mr.Awards.Like {
				if mr.Awards.NonCompliant != 0 {
					action := mrAction{
						pid:   pid,
						mid:   mid,
						aid:   mr.Awards.NonCompliant,
						award: "nc",
						state: false}
					actions = append(actions, action)
				}
			}

			if mr.Awards.NotReady != 0 {
				action := mrAction{
					pid:   pid,
					mid:   mid,
					aid:   mr.Awards.NotReady,
					award: "notready",
					state: false}
				actions = append(actions, action)
			}

			if mr.Awards.Ready != 0 {
				action := mrAction{
					pid:   pid,
					mid:   mid,
					aid:   mr.Awards.Ready,
					award: "ready",
					state: false}
				actions = append(actions, action)
			}
		}
	}

	return actions
}

func processMR(cfg config, actions []mrAction) {
	award := map[string]string{
		"ready":    cfg.Awards.Ready,
		"notready": cfg.Awards.NotReady,
		"nc":       cfg.Awards.NonCompliant,
	}

	git, err := gitlab.NewBasicAuthClient(nil, cfg.GURL, cfg.GUser, cfg.LPass)
	if err != nil {
		fmt.Printf("Failed to connect to GitLab: %v", err)
	}

	for _, action := range actions {
		if action.state {
			award_opts := &gitlab.CreateAwardEmojiOptions{Name: award[action.award]}
			_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(action.pid, action.mid, award_opts)

			// Notify about non-compiant merge
			if action.award == "nc" {
				var users []string
				var emails []string
				var subj string
				var msg string
				var ownersEmail []string
				var ownersUsers []string

				users = append(users, action.mergedBy)
				emails = ldapMail(cfg, users)
				subj = "Code of Conduct failure incident"
				msg = fmt.Sprintf("Hello,"+
					"<p>By merging <a href='%v'>Merge Request #%v</a> without 2 qualified approves"+
					" or negative review you've failed repository's Code of Conduct.</p>"+
					"<p>This incident will be reported.</p>", action.path, action.mid)

				if err := mailSend(cfg, emails, subj, msg); err != nil {
					log.Printf("Failed to send mail: %v", err)
					break
				}

				ownersUsers = append(cfg.VBackend, cfg.VFrontend...)

				ownersEmail = ldapMail(cfg, ownersUsers)

				subj = fmt.Sprintf("MR %v has failed requirements!", action.mid)
				msg = fmt.Sprintf(
					"<p><a href='%v'>Merge Request #%v</a> does not meet requirements but it was merged!</p>",
					action.path, action.mid)

				if err := mailSend(cfg, ownersEmail, subj, msg); err != nil {
					log.Printf("Failed to send mail: %v", err)
					break
				}

				log.Printf("MR %v is merged and has failed CC !!!", action.mid)
			}
		} else {
			_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(action.pid, action.mid, action.aid)
		}
	}
}
