package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prprprus/scheduler"

	"github.com/Kami-no/ward/app"
	"github.com/Kami-no/ward/app/client"
	"github.com/Kami-no/ward/app/client/gitlabclient"
	"github.com/Kami-no/ward/app/ldap"
	"github.com/Kami-no/ward/config"
	"github.com/xanzy/go-gitlab"
)

func main() {
	log.SetOutput(os.Stdout)

	cfg := config.NewConfig()
	gitOpts := gitlab.WithBaseURL(cfg.Endpoints.GitLab)
	httpGitlabClient, err := gitlab.NewBasicAuthClient(cfg.Credentials.User, cfg.Credentials.Password, gitOpts)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to GitLab: %v", err))
	}

	ldapService := ldap.NewLdapServiceImpl(cfg)

	gitlabClient := client.NewGitlabClient(cfg, gitlabclient.NewDefaultGitlabClient(httpGitlabClient), ldapService)
	controller := app.NewMRController(cfg, gitlabClient, ldapService)

	s, err := scheduler.NewScheduler(1000)
	if err != nil {
		panic(err)
	}
	s.Every().Second(15).Do(app.DetectMR, ldapService, gitlabClient, cfg)
	s.Every().Second(0).Minute(0).Hour(2).Weekday(1).Do(app.DetectDeadBrunches, gitlabClient, cfg)

	http.HandleFunc("/", controller.Handler)
	http.HandleFunc("/mr", controller.HandleMR)
	http.HandleFunc("/mr/opened", controller.HandleMROpened)
	http.HandleFunc("/mr/merged", controller.HandleMRMerged)
	http.HandleFunc("/mr/apply", controller.HandleMRApply)
	http.HandleFunc("/dead", controller.HandleDead)
	http.HandleFunc("/dead/letter", controller.HandleDeadLetter)

	server := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	// Wait for an interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Attempt a graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
