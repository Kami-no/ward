package client

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/Kami-no/ward/app/client/gitlabclient"
	"github.com/Kami-no/ward/app/ldap"
	"github.com/Kami-no/ward/config"
)

type GitlabClient interface {
	CheckPrjRequests(projects map[int]*config.Project, list string) (map[int]MrProject, error)
	DetectDead() DeadResults
}

var _ GitlabClient = (*client)(nil)

func NewGitlabClient(config *config.Config, gitlabClient gitlabclient.GitlabClient, service ldap.Service) *client {
	return &client{
		Cfg:         config,
		Client:      gitlabClient,
		LdapService: service,
	}
}

type client struct {
	Cfg         *config.Config
	Client      gitlabclient.GitlabClient
	LdapService ldap.Service
}

func (c *client) CheckPrjRequests(projects map[int]*config.Project, list string) (map[int]MrProject, error) {
	var mrsOpts *gitlab.ListProjectMergeRequestsOptions
	MrProjects := make(map[int]MrProject)

	switch list {
	case "opened":
		mrsOpts = &gitlab.ListProjectMergeRequestsOptions{
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
		mrsOpts = &gitlab.ListProjectMergeRequestsOptions{
			State:   gitlab.String("merged"),
			OrderBy: gitlab.String("updated_at"),
			Scope:   gitlab.String("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    1,
			},
		}
	default:
		mrsOpts = &gitlab.ListProjectMergeRequestsOptions{
			OrderBy: gitlab.String("updated_at"),
			Scope:   gitlab.String("all"),
			ListOptions: gitlab.ListOptions{
				PerPage: 10,
				Page:    1,
			},
		}
	}

	// Process projects
	for pid, project := range projects {
		var mrProject MrProject
		var consensus int

		if project.Votes > 0 {
			consensus = project.Votes
		} else {
			if len(project.Teams) < 2 {
				consensus = 2
			} else {
				consensus = 1
			}
		}

		// Get the list of protected branches
		pbsOpts := &gitlab.ListProtectedBranchesOptions{}
		pbs, _, err := c.Client.ListProtectedBranches(pid, pbsOpts)
		if err != nil {
			log.Printf("Failed to get list of protected branches for %v: %v", pid, err)
			continue
		}
		var protectedBranches = make(map[string]struct{}, len(pbs))

		for _, pb := range pbs {
			protectedBranches[pb.Name] = struct{}{}
		}

		// Get Merge Requests for project
		mrs, _, err := c.Client.ListProjectMergeRequests(pid, mrsOpts)
		if err != nil {
			log.Printf("Failed to list Merge Requests for %v: %v", pid, err)
			break
		}

		// Process Merge Requests
		for _, mr := range mrs {
			// Ignore MR if target branch is not protected
			if _, ok := protectedBranches[mr.TargetBranch]; !ok {
				continue
			}

			var MRequest MergeRequest
			likes := make(map[string]int)

			awards, _, err := c.Client.ListMergeRequestAwardEmoji(pid, mr.IID, &gitlab.ListAwardEmojiOptions{})
			if err != nil {
				log.Printf("Failed to list MR awards: %v", err)
				break
			}

			// Process awards
			for _, award := range awards {
				// Check group awards
				username := strings.ToLower(award.User.Username)

				if award.User.Username != mr.Author.Username {
					switch award.Name {
					case c.Cfg.Awards.Like:
						for team, members := range project.GetTeamsWithMembers() {
							if likes[team] < consensus {
								if _, ok := members[username]; ok {
									likes[team]++
								}
							}
						}
					case c.Cfg.Awards.Dislike:
						MRequest.Awards.Dislike = true
					}
				}

				// Check service awards
				if username == c.Cfg.Credentials.User {
					switch award.Name {
					case c.Cfg.Awards.Ready:
						MRequest.Awards.Ready = award.ID
					case c.Cfg.Awards.NotReady:
						MRequest.Awards.NotReady = award.ID
					case c.Cfg.Awards.NonCompliant:
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

			MRequest.Author = mr.Author.Username

			if mr.MergedBy != nil {
				MRequest.MergedBy = mr.MergedBy.Username
			}

			if mrProject.MR == nil {
				mrProject.MR = make(map[int]MergeRequest)
			}
			mrProject.MR[mr.IID] = MRequest
		}

		MrProjects[pid] = mrProject
	}

	return MrProjects, nil
}

func (c *client) DetectDead() DeadResults {
	var undead DeadResults
	undead.Authors = make(map[string]DeadAuthor)
	undead.Projects = make(map[int]DeadProject)
	trueMail := make(map[string]string)

	projects := c.Cfg.Projects

	now := time.Now()

	branchesOpts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	for pid, project := range projects {
		var owners []string
		for _, team := range project.Teams {
			owners = append(owners, team...)
		}

		// Process all branches for the project not just latest
		for {
			branches, response, err := c.Client.ListBranches(pid, branchesOpts)
			if err != nil {
				log.Printf("Failed to list branches: %v", err)
				break
			}

			for _, branch := range branches {
				var name string
				var mail string

				// Ignore protected branches
				if branch.Protected {
					continue
				}

				updated := *branch.Commit.AuthoredDate
				if now.Sub(updated).Hours() >= 7*24 {
					if _, found := trueMail[branch.Commit.AuthorEmail]; !found {
						// Validate true mail
						if c.LdapService.Check(branch.Commit.AuthorEmail) {
							name = branch.Commit.AuthorName
							mail = branch.Commit.AuthorEmail
						} else {
							var rcptUsers []string

							rcptUser := strings.Split(branch.Commit.AuthorEmail, "@")
							rcptUsers = append(rcptUsers, rcptUser[0])
							rcptEmails := c.LdapService.ListMails(rcptUsers)

							if len(rcptEmails) > 0 {
								name = branch.Commit.AuthorName
								mail = rcptEmails[0]
							} else {
								rcptEmails := c.LdapService.ListMails([]string{branch.Commit.AuthorName})
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
							undead.Authors[mail] = DeadAuthor{
								Name:     name,
								Branches: make(map[int][]string),
							}
						}
					} else {
						mail = trueMail[branch.Commit.AuthorEmail]
					}
					undead.Authors[mail].Branches[pid] = append(undead.Authors[mail].Branches[pid], branch.Name)

					// Fill in data for a the project
					if _, found := undead.Projects[pid]; !found {
						var prjName string
						var prjUrl string

						prjOpts := &gitlab.GetProjectOptions{}
						prj, _, err := c.Client.GetProject(pid, prjOpts)
						if err != nil {
							prjName = fmt.Sprintf("%v", pid)
							prjUrl = c.Cfg.Endpoints.GitLab
							log.Printf("Failed to get project info: %v", err)
						} else {
							prjName = prj.NameWithNamespace
							prjUrl = prj.WebURL
						}

						undead.Projects[pid] = DeadProject{
							Branches: make(map[string]DeadBranch),
							Owners:   owners,
							URL:      prjUrl,
							Name:     prjName,
						}
					}
					undead.Projects[pid].Branches[branch.Name] = DeadBranch{
						Age:    int(now.Sub(updated).Hours()) / 24,
						Author: branch.Commit.AuthorName,
					}
				}
			}

			if response.CurrentPage >= response.TotalPages {
				break
			}
			branchesOpts.Page = response.NextPage
		}
	}
	return undead

}
