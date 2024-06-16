package config

import (
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"
	"time"
)

type AppConfig struct {
	App  App  `mapstructure:"app"`
	Gpas Gpas `mapstructure:"gpas"`
	Fhir Fhir `mapstructure:"fhir"`
}

type App struct {
	LogLevel string `mapstructure:"log-level"`
	Env      string `mapstructure:"env"`
}

type Gpas struct {
	Url string `mapstructure:"url"`
}

type Fhir struct {
	Pseudonymizer Pseudonymizer `mapstructure:"pseudonymizer"`
	Provider      Provider      `mapstructure:"provider"`
}

type Provider struct {
	MongoDb MongoDb `mapstructure:"mongodb"`
}

type MongoDb struct {
	Connection string `mapstructure:"connection"`
}

type Pseudonymizer struct {
	Url   string `mapstructure:"url"`
	Retry Retry  `mapstructure:"retry"`
}

type Retry struct {
	Count   int `mapstructure:"count"`
	Timeout int `mapstructure:"timeout"`
	Wait    int `mapstructure:"wait"`
	MaxWait int `mapstructure:"max-wait"`
}

func LoadConfig() error {

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))

	return viper.ReadInConfig()
}

func ConfigureLogger(c AppConfig) {

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      lvl,
			TimeFormat: time.Kitchen,
		}),
	))

	// set configured log level
	err := lvl.UnmarshalText([]byte(c.App.LogLevel))
	if err != nil {
		slog.Error("Unable to set Log level from application properties", "level", c.App.LogLevel, "error", err)
	}
}
