package app

import (
	"github.com/Kami-no/ward/src/config"
	"log"
)

func DetectDeadBrunches(cfg *config.Config) {
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
