package main

import (
	"context"
	"github.com/Kami-no/ward/src/app"
	"github.com/Kami-no/ward/src/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prprprus/scheduler"
)

func main() {
	cfg := config.NewConfig()
	controller := app.NewMRController(cfg)

	log.SetOutput(os.Stdout)

	s, err := scheduler.NewScheduler(1000)
	if err != nil {
		panic(err)
	}
	s.Every().Second(15).Do(app.DetectMR, cfg)
	s.Every().Second(0).Minute(0).Hour(2).Weekday(1).Do(app.DetectDeadBrunches, cfg)

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
