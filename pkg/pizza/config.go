package pizza

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	zap "go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

var Log *zap.Logger

func init() {
	var err error
	Log, err = zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("could not create logger: %v", err))
	}
	AllPatches = []func(*SQLAccessor) error{
		func(*SQLAccessor) error { return nil },
		Patch001,
		Patch002,
	}
}

type Config struct {
	Port            int            `yaml:"port"`
	StaticDir       string         `yaml:"staticDir"`
	ReadTimeout     time.Duration  `yaml:"readTimeout"`
	WriteTimeout    time.Duration  `yaml:"writeTimeout"`
	ShutdownTimeout time.Duration  `yaml:"shutdownTimeout"`
	Calendar        CalendarConfig `yaml:"calendar"`
	MetricsPort     int            `yaml:"metricsPort"`
	DBFile          string         `yaml:"dbFile"`
	OAuth2          OAuth2Config
}

type CalendarConfig struct {
	CredentialFile string `yaml:"credentialFile"`
	TokenFile      string `yaml:"tokenFile"`
	ID             string `yaml:"id"`
}

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	KeycloakURL  string
	Realm        string
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

func LoadConfigEnv() Config {
	return Config{
		Port:            loadIntEnv("PORT", 5000),
		StaticDir:       loadStrEnv("PIZZA_STATIC_DIR", "static"),
		ReadTimeout:     time.Duration(loadIntEnv("READ_TIMEOUT", 3)) * time.Second,
		WriteTimeout:    time.Duration(loadIntEnv("WRITE_TIMEOUT", 3)) * time.Second,
		ShutdownTimeout: time.Duration(loadIntEnv("SHUTDOWN_TIMEOUT", 5)) * time.Second,
		Calendar: CalendarConfig{
			CredentialFile: loadStrEnv("CREDENTIAL_FILE", "credentials.json"),
			TokenFile:      loadStrEnv("TOKEN_FILE", "token.json"),
			ID:             loadStrEnv("CALENDAR_ID", "primary"),
		},
		MetricsPort: loadIntEnv("METRICS_PORT", 5050),
		DBFile:      loadStrEnv("DBFILE", "pizza.db"),
		OAuth2: OAuth2Config{
			ClientID:     loadStrEnv("OAUTH2_CLIENT_ID", ""),
			ClientSecret: loadStrEnv("OAUTH2_CLIENT_SECRET", ""),
			RedirectURL:  loadStrEnv("OAUTH2_REDIRECT", "http://localhost"),
			KeycloakURL:  loadStrEnv("KEYCLOAK_URL", "http://localhost:8080"),
			Realm:        loadStrEnv("OAUTH2_REALM", "pizza"),
		},
	}
}
