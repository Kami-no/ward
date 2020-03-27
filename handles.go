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
	fmt.Fprint(w, output)
}

func handleMRApply(w http.ResponseWriter, r *http.Request) {
	var cfg config
	cfg.getConfig()

	actions := detectMR(cfg)

	output := fmt.Sprintf("%v", actions)
	fmt.Fprint(w, output)
}
