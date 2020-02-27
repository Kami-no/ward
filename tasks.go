package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

func detectDeadBrunches(cfg config) {
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

				if ldapCheck(cfg, branch.Commit.AuthorEmail) {
					rcpt = append(rcpt, branch.Commit.AuthorEmail)
				} else {
					var rcptUsers []string

					rcptUser := strings.Split(branch.Commit.AuthorEmail, "@")
					rcptUsers = append(rcptUsers, rcptUser[0])
					rcptEmails := ldapMail(cfg, rcptUsers)

					if len(rcptEmails) > 0 {
						rcpt = append(rcpt, rcptEmails[0])
					} else {
						rcpt = append(rcpt, cfg.SMail)
					}
				}

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

func detectMR(cfg config) []mrAction {
	mrsOpened, err := checkPrjRequests(cfg, cfg.Projects, "opened")
	if err != nil {
		log.Println(err)
	}
	actionsOpened := evalOpenedRequests(mrsOpened)

	mrsMerged, err := checkPrjRequests(cfg, cfg.Projects, "merged")
	if err != nil {
		log.Println(err)
	}
	actionsMerged := evalMergedRequests(mrsMerged)

	actions := append(actionsOpened, actionsMerged...)

	processMR(cfg, actions)

	return actions
}
