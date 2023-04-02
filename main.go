package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/mpoegel/rsvp.pizza/internal/pizza"
	"go.uber.org/zap"
)

func main() {
	configFile := flag.String("config", "configs/pizza.yaml", "config file")
	flag.Parse()
	config, err := pizza.LoadConfig(*configFile)
	if err != nil {
		pizza.Log.Fatal("could not load config", zap.Error(err))
	}
	if err := pizza.InitCalendarClient(config.Calendar.CredentialFile, config.Calendar.TokenFile, config.Calendar.ID, context.Background()); err != nil {
		pizza.Log.Fatal("failed to init calendar client", zap.Error(err))
	}
	server, err := pizza.NewServer(config)
	if err != nil {
		pizza.Log.Fatal("could not create server", zap.Error(err))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pizza.Log.Info("shutting down")
		server.Stop()
	}()

	server.Start()
}
