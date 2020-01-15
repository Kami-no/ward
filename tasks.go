package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

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

		awards, _, _ := git.AwardEmoji.ListMergeRequestAwardEmoji(mr.ProjectID, mr.IID, &gitlab.ListAwardEmojiOptions{})
		for _, award := range awards {
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

			if award.User.Username == cfg.GUser {
				if award.Name == cfg.MGood {
					state_good = award.ID
				}
				if award.Name == cfg.MBad {
					state_bad = award.ID
				}
			}

		}

		if vip_f+vip_b == 2 {
			if state_bad != 0 {
				_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, state_bad)
			}
			if state_good == 0 {
				award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MGood}
				_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
				log.Printf("MR %v is ready.", mr.IID)
			}
		} else {
			if state_good != 0 {
				_, _ = git.AwardEmoji.DeleteMergeRequestAwardEmoji(mr.ProjectID, mr.IID, state_good)
			}
			if state_bad == 0 {
				award_opts := &gitlab.CreateAwardEmojiOptions{Name: cfg.MBad}
				_, _, _ = git.AwardEmoji.CreateMergeRequestAwardEmoji(mr.ProjectID, mr.IID, award_opts)
				log.Printf("MR %v is not ready.", mr.IID)
			}
		}
	}
}
