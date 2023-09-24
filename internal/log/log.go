package log

import (
	"log/slog"
	"os"
)

var (
	DefaultLogger func() *Logger = makeLogger
)

func makeLogger() *Logger {
	handler := slog.NewTextHandler(
		os.Stderr,
		logOptions,
	)
	return slog.New(handler)
}

func NewTestLogger() *Logger {
	handler := slog.NewTextHandler(
		os.Stdout,
		logOptions,
	)
	return slog.New(handler)
}

var (
	LogLevel   = &slog.LevelVar{}
	logOptions = &slog.HandlerOptions{
		AddSource:   true,
		Level:       LogLevel,
		ReplaceAttr: nil,
	}
)

type Loggable interface {
	WithLogger(*Logger)
}

type (
	Logger = slog.Logger
	Value  = slog.Value
)

var (
	GroupValue = slog.GroupValue
	String     = slog.String
	Group      = slog.Group
)
