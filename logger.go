package gopxgrid

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-stomp/stomp/v3"
)

type internalLogger struct {
	l *slog.Logger
}

var _ stomp.Logger = (*internalLogger)(nil)

func newInternalLogger(level slog.Level) stomp.Logger {
	lvl := new(slog.LevelVar)
	lvl.Set(level)

	return &internalLogger{
		l: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: lvl,
		})),
	}
}

func (l *internalLogger) Debugf(format string, value ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, value...))
}

func (l *internalLogger) Infof(format string, value ...interface{}) {
	l.l.Info(fmt.Sprintf(format, value...))
}

func (l *internalLogger) Warningf(format string, value ...interface{}) {
	l.l.Warn(fmt.Sprintf(format, value...))
}

func (l *internalLogger) Errorf(format string, value ...interface{}) {
	l.l.Error(fmt.Sprintf(format, value...))
}

func (l *internalLogger) Debug(message string) {
	l.l.Debug(message)
}

func (l *internalLogger) Info(message string) {
	l.l.Info(message)
}

func (l *internalLogger) Warning(message string) {
	l.l.Warn(message)
}

func (l *internalLogger) Error(message string) {
	l.l.Error(message)
}
