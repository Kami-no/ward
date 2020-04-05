package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func handleMR(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	mrs, err := checkPrjRequests(cfg, cfg.Projects, "any")
	if err != nil {
		log.Println(err)
	}
	output := fmt.Sprintf("%v", mrs)
	fmt.Fprint(w, output)
}

func handleMROpened(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	mrs, err := checkPrjRequests(cfg, cfg.Projects, "opened")
	if err != nil {
		log.Println(err)
	}

	actions := evalOpenedRequests(mrs)
	output := fmt.Sprintf("%v", actions)
	fmt.Fprint(w, output)
}

func handleMRMerged(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	mrs, err := checkPrjRequests(cfg, cfg.Projects, "merged")
	if err != nil {
		log.Println(err)
	}

	data := evalMergedRequests(mrs)

	out, _ := json.Marshal(data)
	output := fmt.Sprintf("%v", string(out))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func handleMRApply(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	data := detectMR(cfg)

	out, _ := json.Marshal(data)
	output := fmt.Sprintf("%v", string(out))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func handleDead(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	data := detectDead(cfg)
	undead, _ := json.Marshal(data)

	output := fmt.Sprintf("%v", string(undead))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func handleDeadLetter(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	undead := detectDead(cfg)
	for _, v := range undead.Authors {
		v.Projects = undead.Projects
		template, err := deadAuthorTemplate(v)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, template)
		break
	}
}
