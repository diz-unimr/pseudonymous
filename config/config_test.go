package config

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestConfigureLoggerSetsLogLevel(t *testing.T) {

	expected := "debug"

	ConfigureLogger(AppConfig{App: App{LogLevel: expected}})

	assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))
}
