package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prprprus/scheduler"
)

func main() {
	var cfg config
	cfg.getConfig()

	log.SetOutput(os.Stdout)

	s, err := scheduler.NewScheduler(1000)
	if err != nil {
		panic(err)
	}
	s.Every().Second(15).Do(detectMR, cfg)
	s.Every().Second(0).Minute(0).Hour(2).Weekday(1).Do(detectDeadBrunches, cfg)

	http.HandleFunc("/", handler)
	http.HandleFunc("/mr", handleMR)
	http.HandleFunc("/mr/opened", handleMROpened)
	http.HandleFunc("/mr/merged", handleMRMerged)
	http.HandleFunc("/mr/apply", handleMRApply)
	http.HandleFunc("/dead", handleDead)
	http.HandleFunc("/dead/letter", handleDeadLetter)

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
