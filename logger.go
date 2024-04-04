package gopxgrid

import (
	"context"
	"log/slog"
)

type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

var _ Logger = (*slogWrapper)(nil)

type slogWrapper struct {
	*slog.Logger
}

func (s *slogWrapper) With(args ...any) Logger {
	return &slogWrapper{s.Logger.With(args...)}
}

func FromSlog(logger *slog.Logger) Logger {
	return &slogWrapper{logger}
}
