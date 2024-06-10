package config

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseEnv(t *testing.T) {
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("KAFKA_SECURITY_PROTOCOL", "plaintext")

	cases := []struct {
		name     string
		env      string
		config   string
		expected string
	}{
		{
			name:     "matchesDelimiter",
			config:   "app.name",
			env:      "APP_NAME",
			expected: "app.name",
		},
		{
			name:     "matchesHyphen",
			config:   "kafka.security-protocol",
			env:      "KAFKA_SECURITY_PROTOCOL",
			expected: "kafka.security-protocol",
		},
		{
			name:     "matchesOnlyExisting",
			config:   "other.value",
			env:      "APP_NAME",
			expected: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			k := koanf.New(".")
			_ = k.Set(c.config, "test")

			actual := parseEnv(k, c.env)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestLoadConfig(t *testing.T) {

	_, b, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(b), "../..")

	c, _ := LoadConfig(base + "/app.yaml")

	assert.Equal(t, c.App.Name, "consent-to-fhir")
}

func TestLoadConfig_invalidPath(t *testing.T) {

	_, b, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(b), "../..")

	c, err := LoadConfig(base + "/invalid.yml")

	assert.Nil(t, c)
	assert.Error(t, err)
}
