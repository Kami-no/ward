package main

import (
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

	actions := evalMergedRequests(mrs)
	output := fmt.Sprintf("%v", actions)
	fmt.Fprint(w, output)
}

func handleMRApply(w http.ResponseWriter, r *http.Request) {
	actions := detectMR()

	output := fmt.Sprintf("%v", actions)
	fmt.Fprint(w, output)
}
