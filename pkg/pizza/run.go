package pizza

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
)

func Run(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Parse(args)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	config := LoadConfigEnv()
	metricsReg := NewPrometheusRegistry()
	server, err := NewServer(config, metricsReg)
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

	server.Start()
}
