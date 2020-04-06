package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

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
	Pid      int
	Mid      int
	Aid      int
	Award    string
	MergedBy string
	Path     string
	State    bool
}

type deadBranch struct {
	Author string
	Age    int
}

type deadProject struct {
	Name     string
	URL      string
	Owners   []string
	Branches map[string]deadBranch
}

type deadAuthor struct {
	Name     string
	Branches map[int][]string
	Projects map[int]deadProject
}

type deadResults struct {
	Projects map[int]deadProject
	Authors  map[string]deadAuthor
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

			MRequest.Path = mr.WebURL

			if mr.MergedBy != nil {
				MRequest.MergedBy = mr.MergedBy.Username
			}

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
						Pid:   pid,
						Mid:   mid,
						Aid:   mr.Awards.NotReady,
						Award: "notready",
						State: false}
					actions = append(actions, action)
				}
				if mr.Awards.Ready == 0 {
					action := mrAction{
						Pid:   pid,
						Mid:   mid,
						Aid:   mr.Awards.Ready,
						Award: "ready",
						State: true}
					actions = append(actions, action)
				}
			} else {
				if mr.Awards.Ready != 0 {
					action := mrAction{
						Pid:   pid,
						Mid:   mid,
						Aid:   mr.Awards.Ready,
						Award: "ready",
						State: false}
					actions = append(actions, action)
				}
				if mr.Awards.NotReady == 0 {
					action := mrAction{
						Pid:      pid,
						Mid:      mid,
						Aid:      mr.Awards.NotReady,
						Award:    "notready",
						MergedBy: mr.MergedBy,
						State:    true}
					actions = append(actions, action)
				}
			}

			if mr.Awards.NonCompliant != 0 {
				action := mrAction{
					Pid:   pid,
					Mid:   mid,
					Aid:   mr.Awards.NonCompliant,
					Award: "nc",
					State: false}
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
			if mr.Awards.Dislike || !mr.Awards.Like {
				if mr.Awards.NonCompliant == 0 {
					action := mrAction{
						Pid:      pid,
						Mid:      mid,
						Aid:      mr.Awards.NonCompliant,
						Award:    "nc",
						MergedBy: mr.MergedBy,
						Path:     mr.Path,
						State:    true}
					actions = append(actions, action)
				}
			} else {
				if mr.Awards.NonCompliant != 0 {
					action := mrAction{
						Pid:   pid,
						Mid:   mid,
						Aid:   mr.Awards.NonCompliant,
						Award: "nc",
						State: false}
					actions = append(actions, action)
				}
			}

			if mr.Awards.NotReady != 0 {
				action := mrAction{
					Pid:   pid,
					Mid:   mid,
					Aid:   mr.Awards.NotReady,
					Award: "notready",
					State: false}
				actions = append(actions, action)
			}

			if mr.Awards.Ready != 0 {
				action := mrAction{
					Pid:   pid,
					Mid:   mid,
					Aid:   mr.Awards.Ready,
					Award: "ready",
					State: false}
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
		if action.State {
			award_opts := &gitlab.CreateAwardEmojiOptions{Name: award[action.Award]}
			_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(action.Pid, action.Mid, award_opts)

			// Notify about non-compiant merge
			if action.Award == "nc" {
				var users []string
				var emails []string
				var subj string
				var msg string
				var ownersEmail []string
				var ownersUsers []string

				log.Printf("Non-compliant MR detected: %v@%v", action.Mid, action.Pid)

				users = append(users, action.MergedBy)
				emails = ldapMail(cfg, users)
				subj = "Code of Conduct failure incident"
				msg = fmt.Sprintf("Hello,"+
					"<p>By merging <a href='%v'>Merge Request #%v</a> without 2 qualified approves"+
					" or negative review you've failed repository's Code of Conduct.</p>"+
					"<p>This incident will be reported.</p>", action.Path, action.Mid)

				if err := mailSend(cfg, emails, subj, msg); err != nil {
					log.Printf("Failed to send mail to recipient: %v", err)
				}

				for _, prj := range cfg.Projects {
					if prj.ID == action.Pid {
						for _, team := range prj.Teams {
							ownersUsers = append(ownersUsers, team...)
						}
					}
				}

				ownersEmail = ldapMail(cfg, ownersUsers)

				subj = fmt.Sprintf("MR %v has failed requirements!", action.Mid)
				msg = fmt.Sprintf(
					"<p><a href='%v'>Merge Request #%v</a> does not meet requirements but it was merged!</p>",
					action.Path, action.Mid)

				if err := mailSend(cfg, ownersEmail, subj, msg); err != nil {
					log.Printf("Failed to send mail to owners: %v", err)
				}
			}
		} else {
			_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(action.Pid, action.Mid, action.Aid)
		}
	}
}

func detectDead(cfg config) deadResults {
	var undead deadResults
	undead.Authors = make(map[string]deadAuthor)
	undead.Projects = make(map[int]deadProject)
	trueMail := make(map[string]string)

	projects := cfg.Projects

	now := time.Now()

	git, err := gitlab.NewBasicAuthClient(nil, cfg.GURL, cfg.GUser, cfg.LPass)
	if err != nil {
		fmt.Printf("Failed to connect to GitLab: %v", err)
	}

	branches_opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	for _, project := range projects {
		// Process all branches for the project not just latest
		for {
			branches, response, err := git.Branches.ListBranches(project.ID, branches_opts)
			if err != nil {
				log.Printf("Failed to list branches: %v", err)
				break
			}

			for _, branch := range branches {
				var name string
				var mail string

				updated := *branch.Commit.AuthoredDate
				if now.Sub(updated).Hours() >= 7*24 {
					if _, found := trueMail[branch.Commit.AuthorEmail]; !found {
						// Validate true mail
						if ldapCheck(cfg, branch.Commit.AuthorEmail) {
							name = branch.Commit.AuthorName
							mail = branch.Commit.AuthorEmail
						} else {
							var rcptUsers []string

							rcptUser := strings.Split(branch.Commit.AuthorEmail, "@")
							rcptUsers = append(rcptUsers, rcptUser[0])
							rcptEmails := ldapMail(cfg, rcptUsers)

							if len(rcptEmails) > 0 {
								name = branch.Commit.AuthorName
								mail = rcptEmails[0]
							} else {
								rcptEmails := ldapMail(cfg, []string{branch.Commit.AuthorName})
								if len(rcptEmails) > 0 {
									name = branch.Commit.AuthorName
									mail = rcptEmails[0]
								} else {
									name = "Unidentified"
									mail = "unidentified@any.local"
									log.Printf("Unidentified author: %v - %v",
										branch.Commit.AuthorName, branch.Commit.AuthorEmail)
								}

							}
						}

						trueMail[branch.Commit.AuthorEmail] = mail

						if _, found := undead.Authors[mail]; !found {
							undead.Authors[mail] = deadAuthor{
								Name:     name,
								Branches: make(map[int][]string),
							}
						}
					} else {
						mail = trueMail[branch.Commit.AuthorEmail]
					}
					undead.Authors[mail].Branches[project.ID] = append(undead.Authors[mail].Branches[project.ID], branch.Name)

					// Fill in data for a the project
					if _, found := undead.Projects[project.ID]; !found {
						var prj_name string
						var prj_url string

						prj_opts := &gitlab.GetProjectOptions{}
						prj, _, err := git.Projects.GetProject(project.ID, prj_opts)
						if err != nil {
							prj_name = fmt.Sprintf("%v", project.ID)
							prj_url = cfg.Endpoints.GitLab
							log.Printf("Failed to get project info: %v", err)
						} else {
							prj_name = prj.NameWithNamespace
							prj_url = prj.WebURL
						}

						undead.Projects[project.ID] = deadProject{
							Branches: make(map[string]deadBranch),
							Owners:   append(cfg.VBackend, cfg.VFrontend...),
							URL:      prj_url,
							Name:     prj_name,
						}
					}
					undead.Projects[project.ID].Branches[branch.Name] = deadBranch{
						Age:    int(now.Sub(updated).Hours()) / 24,
						Author: branch.Commit.AuthorName,
					}
				}
			}

			if response.CurrentPage >= response.TotalPages {
				break
			}
			branches_opts.Page = response.NextPage
		}
	}
	return undead
}

func deadAuthorTemplate(dAuthor deadAuthor) (string, error) {
	var buffer bytes.Buffer
	var output string

	tmpl := template.Must(template.ParseFiles("templates/dead-branches-author.gohtml"))
	err := tmpl.Execute(&buffer, dAuthor)
	if err != nil {
		return output, err
	}
	output = buffer.String()

	return output, nil
}
