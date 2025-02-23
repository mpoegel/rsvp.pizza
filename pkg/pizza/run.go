package pizza

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
)

func Run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Parse(args)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx := context.Background()

	config := LoadConfigEnv()
	slog.Info("using the sqlite accessor")
	accessor, err := NewSQLAccessor(config.DBFile, false)
	if err != nil {
		return err
	}

	googleCal, err := NewGoogleCalendar(config.Calendar.CredentialFile, config.Calendar.TokenFile, config.Calendar.ID, ctx)
	if err != nil {
		return err
	}

	keycloak, err := NewKeycloak(ctx, config.OAuth2)
	if err != nil {
		slog.Error("keycloak failure", "error", err)
		return err
	}

	metricsReg := NewPrometheusRegistry()
	server, err := NewServer(config, accessor, googleCal, keycloak, metricsReg)
	if err != nil {
		slog.Error("could not create server", "error", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		slog.Info("shutting down")
		server.Stop()
	}()

	if config.MetricsPort != 0 {
		go metricsReg.Serve(config.MetricsPort)
	}

	return server.Start()
}
