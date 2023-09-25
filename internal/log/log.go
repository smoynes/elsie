// Package log provides logging output.
package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// DefaultLogger returns the default, global logger. Application components can call DefaultLogger
// and cache the result. The default will not change at runtime.
var (
	DefaultLogger func() *Logger = slog.Default
)

// NewFormattedLogger returns a logger that uses a Handler to format and write logs to a Writer.
func NewFormattedLogger(out io.Writer) *Logger {
	handler := NewHandler(out)
	return slog.New(handler)
}

// Handler implements slog.Handler to produce formatted log output.
//
// (It exists as an exercise in learning about the slog module.)
type Handler struct {
	out io.Writer
	mut *sync.Mutex // Synchronizes writer.

	opts    slog.HandlerOptions
	grouped bool
	attrs   []Attr
}

func NewHandler(out io.Writer) *Handler {
	h := Handler{
		out:  out,
		mut:  new(sync.Mutex),
		opts: *logOptions,
	}

	return &h
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *Handler) Handle(ctx context.Context, rec slog.Record) error {
	buf := make([]byte, 0, 128) // TODO: buffer pool
	out := bytes.NewBuffer(buf)

	if !rec.Time.IsZero() {
		fmt.Fprintf(out, "%10s : %s\n", "TIMESTAMP", rec.Time.Format(time.RFC3339))
	}

	fmt.Fprintf(out, "%10s : %s\n", "LEVEL", rec.Level.String())

	if h.opts.AddSource && rec.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{rec.PC})
		f, _ := frames.Next()

		fmt.Fprintf(out, "%10s : %s:%d\n", "SOURCE", f.File, f.Line)

		if f.Func != nil {
			fmt.Fprintf(out, "%10s : %s\n", "FUNCTION", f.Function)
		}
	}

	fmt.Fprintf(out, "%10s : %s\n", "MESSAGE", rec.Message)

	for _, a := range h.attrs {
		h.appendAttr(out, a, false)
	}

	rec.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(out, attr, false)
		return true
	})

	fmt.Fprintln(out)

	h.mut.Lock()
	defer h.mut.Unlock()

	_, err := h.out.Write(out.Bytes())

	return err
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	attrs := make([]Attr, len(h.attrs))
	copy(attrs, h.attrs)

	return &Handler{
		mut:   h.mut,
		out:   h.out,
		opts:  h.opts,
		attrs: attrs,
	}
}

// WithAttrs returns a new handler that combines the handler's attributes and those in the argument.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	as := make([]Attr, 0, len(h.attrs)+len(attrs))
	copy(as, h.attrs)
	as = append(as, attrs...)

	return &Handler{
		out:   h.out,
		mut:   h.mut,
		opts:  h.opts,
		attrs: as,
	}
}

func (h *Handler) appendAttr(out io.Writer, attr slog.Attr, grouped bool) error {
	attr.Value = attr.Value.Resolve()

	switch {
	case attr.Equal(slog.Attr{}):
		return nil

	case attr.Value.Kind() != slog.KindGroup:
		if grouped {
			fmt.Fprint(out, "  ")
		}
		fmt.Fprintf(out, "%10s : %v\n", attr.Key, attr.Value.Any())
	case attr.Value.Kind() == slog.KindGroup && attr.Key != "":
		fmt.Fprintf(out, "%10s :\n", attr.Key)
		grouped = true
		fallthrough
	case attr.Value.Kind() == slog.KindGroup && attr.Key == "":
		for _, a := range attr.Value.Group() {
			h.appendAttr(out, a, grouped)
		}
	}

	return nil
}

type Loggable interface {
	WithLogger(*Logger)
}

var (
	LogLevel   = &slog.LevelVar{}
	logOptions = &slog.HandlerOptions{
		AddSource: true,
		Level:     LogLevel,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			// string paths and packages
			return attr // TODO
		},
	}
)

type (
	Logger = slog.Logger
	Value  = slog.Value
	Attr   = slog.Attr
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
)
