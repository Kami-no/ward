package main

import (
	"log"
)

func detectDeadBrunches(cfg config) {
	undead := detectDead(cfg)
	for rcpt, v := range undead.Authors {
		if rcpt == "unidentified@any.local" {
			continue
		}
		v.Projects = undead.Projects
		msg, err := deadAuthorTemplate(v)
		if err != nil {
			log.Printf("Templating error: %v", err)
			return
		}

		subj := "Dead branch notification"
		if err := mailSend(cfg, []string{rcpt}, subj, msg); err != nil {
			log.Printf("Failed to send mail: %v", err)
		}
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
