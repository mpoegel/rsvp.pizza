package pizza

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	UseSQLite       bool           `yaml:"useSQLite"`
	DBFile          string         `yaml:"dbFile"`
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

func loadStrEnv(name, defaultVal string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return defaultVal
	}
	return val
}

func loadBoolEnv(name string, defaultVal bool) bool {
	val, ok := os.LookupEnv(name)
	if !ok {
		return defaultVal
	}
	return strings.ToLower(val) == "true" || val == "1"
}

func loadIntEnv(name string, defaultVal int) int {
	valStr, ok := os.LookupEnv(name)
	if !ok {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func LoadConfigEnv(filename string) Config {
	return Config{
		Port:            loadIntEnv("PORT", 5000),
		ReadTimeout:     time.Duration(loadIntEnv("READ_TIMEOUT", 3)) * time.Second,
		WriteTimeout:    time.Duration(loadIntEnv("WRITE_TIMEOUT", 3)) * time.Second,
		ShutdownTimeout: time.Duration(loadIntEnv("SHUTDOWN_TIMEOUT", 5)) * time.Second,
		Calendar: CalendarConfig{
			CredentialFile: loadStrEnv("CREDENTIAL_FILE", "credentials.json"),
			TokenFile:      loadStrEnv("TOKEN_FILE", "token.json"),
			ID:             loadStrEnv("CALENDAR_ID", "primary"),
		},
		MetricsPort: loadIntEnv("METRICS_PORT", 5050),
		UseSQLite:   loadBoolEnv("USE_SQLITE", true),
		DBFile:      loadStrEnv("DBFILE", "pizza.db"),
	}
}
