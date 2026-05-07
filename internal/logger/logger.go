package logger

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/odysseythink/mlog"
)

var (
	defaultLogger *Logger
	initOnce      sync.Once
	initErr       error
)

// Logger wraps mlog functionality with a compatible API.
type Logger struct {
	fields []mlog.Field
}

func Init(level string, output string) error {
	initOnce.Do(func() {
		mlog.SetEncoder(mlog.NewJSONEncoder())

		switch strings.ToLower(level) {
		case "debug", "info", "warn", "warning", "error", "fatal":
			// valid
		default:
			initErr = fmt.Errorf("init logger: invalid level %q", level)
			return
		}

		if output == "stdout" || output == "stderr" {
			// Best effort: mlog always registers a stderr sink.
		}

		mlog.SetLogLevel(-1)

		defaultLogger = &Logger{}
	})
	return initErr
}

func Default() *Logger {
	if defaultLogger == nil {
		panic("logger not initialized")
	}
	return defaultLogger
}

func (l *Logger) Debug(msg string, args ...any) {
	if len(args) > 0 {
		mlog.Debugf(msg, args...)
	} else {
		mlog.Debug(msg)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if len(args) > 0 {
		mlog.Infof(msg, args...)
	} else {
		mlog.Info(msg)
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if len(args) > 0 {
		mlog.Warningf(msg, args...)
	} else {
		mlog.Warning(msg)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	if len(args) > 0 {
		mlog.Errorf(msg, args...)
	} else {
		mlog.Error(msg)
	}
}

func (l *Logger) With(key string, val any) *Logger {
	return &Logger{fields: append(l.fields, mlog.Any(key, val))}
}

func Debug(msg string, args ...any) { Default().Debug(msg, args...) }
func Info(msg string, args ...any)  { Default().Info(msg, args...) }
func Warn(msg string, args ...any)  { Default().Warn(msg, args...) }
func Error(msg string, args ...any) { Default().Error(msg, args...) }

func WithContext(ctx context.Context) *Logger {
	if reqID := ctx.Value("request_id"); reqID != nil {
		return Default().With("request_id", reqID)
	}
	return Default()
}
