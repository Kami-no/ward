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

	s, err := scheduler.NewScheduler(1000)
	if err != nil {
		panic(err)
	}
	s.Every().Second(0).Do(trackOpenedMR, cfg)
	s.Every().Second(30).Do(trackMergedMR, cfg)

	http.HandleFunc("/", handler)

	server := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
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
