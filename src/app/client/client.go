package client

import (
	"github.com/Kami-no/ward/src/app/client/gitlabclient"
	"github.com/Kami-no/ward/src/config"
	"github.com/xanzy/go-gitlab"
	"log"
	"strings"
)

type GitlabClient interface {
	CheckPrjRequests(projects map[int]*config.Project, list string) (map[int]MrProject, error)
}

var _ GitlabClient = (*client)(nil)

func NewGitlabClient(config *config.Config, gitlabClient gitlabclient.GitlabClient) *client {
	return &client{
		Cfg:    config,
		Client: gitlabClient,
	}
}

type client struct {
	Cfg    *config.Config
	Client gitlabclient.GitlabClient
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
