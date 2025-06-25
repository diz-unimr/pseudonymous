package config

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"time"
)

type AppConfig struct {
	App  App  `mapstructure:"app"`
	Gpas Gpas `mapstructure:"gpas"`
	Fhir Fhir `mapstructure:"fhir"`
}

type App struct {
	LogLevel    string `mapstructure:"log-level"`
	Concurrency int    `mapstructure:"concurrency"`
}

type Gpas struct {
	Url     string            `mapstructure:"url"`
	Auth    *Auth             `mapstructure:"auth"`
	Domains map[string]string `mapstructure:"domains"`
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
	BatchSize  int    `mapstructure:"batch-size"`
}

type Pseudonymizer struct {
	Url   string `mapstructure:"url"`
	Retry Retry  `mapstructure:"retry"`
	Auth  *Auth  `mapstructure:"auth"`
}

type Auth struct {
	Basic *Basic `mapstructure:"basic"`
}

type Basic struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Retry struct {
	Count   int `mapstructure:"count"`
	Timeout int `mapstructure:"timeout"`
	Wait    int `mapstructure:"wait"`
	MaxWait int `mapstructure:"max-wait"`
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
