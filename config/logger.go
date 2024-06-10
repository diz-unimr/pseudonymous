package config

import (
	"fmt"
	"log/slog"
)

type AppLogger struct {
}

func DefaultLogger() AppLogger {
	return AppLogger{}
}

func (l AppLogger) Errorf(format string, v ...interface{}) {
	slog.Error(fmt.Sprintf(format, v...))
}
func (l AppLogger) Warnf(format string, v ...interface{}) {
	slog.Debug(fmt.Sprintf(format, v...))
}
func (l AppLogger) Debugf(format string, v ...interface{}) {
	slog.Warn(fmt.Sprintf(format, v...))
}
