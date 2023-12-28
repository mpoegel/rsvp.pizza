package pizza

import (
	"flag"
	"os"
	"os/signal"

	"go.uber.org/zap"
)

func Run(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.Parse(args)

	config := LoadConfigEnv()
	metricsReg := NewPrometheusRegistry()
	server, err := NewServer(config, metricsReg)
	if err != nil {
		Log.Fatal("could not create server", zap.Error(err))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		Log.Info("shutting down")
		server.Stop()
	}()

	if config.MetricsPort != 0 {
		go metricsReg.Serve(config.MetricsPort)
	}

	server.Start()
}
