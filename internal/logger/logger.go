package logger

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/odysseythink/mlog"
)

func formatPairs(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	var b strings.Builder
	b.WriteString(msg)
	for i := 0; i < len(args)-1; i += 2 {
		b.WriteString(" ")
		key, ok := args[i].(string)
		if ok {
			b.WriteString(key)
			b.WriteString("=")
		}
		b.WriteString(fmt.Sprintf("%v", args[i+1]))
	}
	return b.String()
}

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
		flag.Set("logtostderr", "true")
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
	mlog.Debug(formatPairs(msg, args...))
}

func (l *Logger) Info(msg string, args ...any) {
	mlog.Info(formatPairs(msg, args...))
}

func (l *Logger) Warn(msg string, args ...any) {
	mlog.Warning(formatPairs(msg, args...))
}

func (l *Logger) Error(msg string, args ...any) {
	mlog.Error(formatPairs(msg, args...))
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
