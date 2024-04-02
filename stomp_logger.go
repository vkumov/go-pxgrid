package gopxgrid

import (
	"fmt"
	"log/slog"

	"github.com/go-stomp/stomp/v3"
)

type stompLogger struct {
	*slog.Logger
}

func fromSlogLogger(logger *slog.Logger) stomp.Logger {
	return &stompLogger{logger}
}

func (l *stompLogger) Debugf(format string, value ...interface{}) {
	l.Debug(fmt.Sprintf(format, value...))
}
func (l *stompLogger) Infof(format string, value ...interface{}) {
	l.Info(fmt.Sprintf(format, value...))
}
func (l *stompLogger) Warningf(format string, value ...interface{}) {
	l.Warning(fmt.Sprintf(format, value...))
}
func (l *stompLogger) Errorf(format string, value ...interface{}) {
	l.Error(fmt.Sprintf(format, value...))
}

func (l *stompLogger) Debug(message string)   { l.Debug(message) }
func (l *stompLogger) Info(message string)    { l.Info(message) }
func (l *stompLogger) Warning(message string) { l.Warning(message) }
func (l *stompLogger) Error(message string)   { l.Error(message) }
