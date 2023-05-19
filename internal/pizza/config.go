package pizza

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var Log *zap.Logger

func init() {
	var err error
	Log, err = zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("could not create logger: %v", err))
	}

	if val := os.Getenv("PIZZA_STATIC_DIR"); len(val) > 0 {
		StaticDir = val
	}
}

type Config struct {
	Port            int            `yaml:"port"`
	ReadTimeout     time.Duration  `yaml:"readTimeout"`
	WriteTimeout    time.Duration  `yaml:"writeTimeout"`
	ShutdownTimeout time.Duration  `yaml:"shutdownTimeout"`
	Calendar        CalendarConfig `yaml:"calendar"`
	MetricsPort     int            `yaml:"metricsPort"`
}

type CalendarConfig struct {
	CredentialFile string `yaml:"credentialFile"`
	TokenFile      string `yaml:"tokenFile"`
	ID             string `yaml:"id"`
}

func LoadConfig(filename string) (Config, error) {
	config := Config{}
	rawBytes, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(rawBytes, &config)
	return config, err
}
