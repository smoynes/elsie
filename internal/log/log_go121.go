//go:build 1.21

package log

import (
	"log/slog"
)

type (
	Logger = slog.Logger
	Value  = slog.Value
	Attr   = slog.Attr

	levelVar = slog.LevelVar
	options  = slog.HandlerOptions
)

var (
	AnyValue      = slog.AnyValue
	BoolValue     = slog.BoolValue
	DurationValue = slog.DurationValue
	Float64Value  = slog.Float64Value
	GroupValue    = slog.GroupValue
	Int64Value    = slog.Int64Value
	IntValue      = slog.IntValue
	StringValue   = slog.StringValue
	TimeValue     = slog.TimeValue
	Uint64Value   = slog.Uint64Value

	Any    = slog.Any
	String = slog.String
	Group  = slog.Group

	SetDefault = slog.SetDefault
	newSlog    = slog.New
)

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelWarn
)
