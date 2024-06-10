package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"regexp"
	"strings"
)

type AppConfig struct {
	App  App  `koanf:"app"`
	Gpas Gpas `koanf:"gpas"`
	Fhir Fhir `koanf:"fhir"`
}

type App struct {
	Name     string `koanf:"name"`
	LogLevel string `koanf:"log-level"`
	Env      string `koanf:"env"`
}

type Gpas struct {
	Url string `koanf:"url"`
}

type Fhir struct {
	Pseudonymizer Pseudonymizer `koanf:"pseudonymizer"`
	Provider      Provider      `koanf:"provider"`
}

type Provider struct {
	MongoDb MongoDb `koanf:"mongodb"`
}

type MongoDb struct {
	Connection string `koanf:"connection"`
}

type Pseudonymizer struct {
	Url   string `koanf:"url"`
	Retry Retry  `koanf:"retry"`
}

type Retry struct {
	Count   int `koanf:"count"`
	Timeout int `koanf:"timeout"`
	Wait    int `koanf:"wait"`
	MaxWait int `koanf:"max-wait"`
}

func LoadConfig(path string) (*AppConfig, error) {

	// load config file
	var k = koanf.New(".")
	f := file.Provider(path)
	if err := k.Load(f, yaml.Parser()); err != nil {
		return nil, err
	}
	// replace env vars
	_ = k.Load(env.Provider("", ".", func(s string) string {
		return parseEnv(k, s)
	}), nil)

	return parseConfig(k), nil
}

func parseEnv(k *koanf.Koanf, s string) string {
	r := "^" + strings.Replace(strings.ToLower(s), "_", "(.|-)", -1) + "$"

	for _, p := range k.Keys() {
		match, _ := regexp.MatchString(r, p)
		if match {
			return p
		}
	}
	return ""
}

func parseConfig(k *koanf.Koanf) (config *AppConfig) {
	_ = k.Unmarshal("", &config)
	return config
}
