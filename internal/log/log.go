// Package log provides logging output.
package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DefaultLogger returns the default, global logger. During application startup components can call
// DefaultLogger and cache the result. The default will not change at runtime.
var (
	DefaultLogger func() *Logger = func() *Logger {
		return NewFormattedLogger(os.Stderr)
	}

	// LogLevel is a variable holding the log level. It can be changed at runtime.
	LogLevel = &slog.LevelVar{}
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
	mut *sync.Mutex // Synchronizes writer.
	out io.Writer

	opts  *slog.HandlerOptions
	group string
	attrs []Attr
}

// The log options for Handlers.
var logOptions = &slog.HandlerOptions{
	AddSource:   true,
	Level:       LogLevel,
	ReplaceAttr: cleanAttr,
}

// NewHandler creates and initializes a Handler with a writer.
func NewHandler(out io.Writer) *Handler {
	h := Handler{
		out:  out,
		mut:  new(sync.Mutex),
		opts: logOptions,
	}

	return &h
}

// Enabled returns true if the level is greater than the current logging level.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle formats and writes a log record to the handler's writer. There are some subtle rules about
// how it ought to behave. See the [slog handler guide].
//
// [slog handler guide]: https://github.com/golang/example/tree/d9923f6970e9ba7e0d23aa9448ead71ea57235ae/slog-handler-guide
func (h *Handler) Handle(ctx context.Context, rec slog.Record) error {
	buf := make([]byte, 0, 4096) // TODO: buffer pool
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
		err := h.appendAttr(out, a, false)
		if err != nil {
			panic(err)
		}
	}

	rec.Attrs(func(attr slog.Attr) bool {
		err := h.appendAttr(out, attr, false)
		if err != nil {
			panic(err)
		}
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
		group: name,
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
	var err error

	attr.Value = attr.Value.Resolve()
	attr = h.opts.ReplaceAttr([]string{h.group}, attr)

	key, value := strings.ToUpper(attr.Key), attr.Value

	switch {
	case attr.Equal(slog.Attr{}):
		return nil

	case value.Kind() != slog.KindGroup:
		if grouped {
			fmt.Fprint(out, "  ")
		}
		_, err = fmt.Fprintf(out, "%10s : %v\n", key, value.Any())
		return err
	case value.Kind() == slog.KindGroup && key != "":
		_, err = fmt.Fprintf(out, "%10s :\n", key)
		grouped = true
		h.group = key

		if err != nil {
			return err
		}

		for _, a := range value.Group() {
			if err := h.appendAttr(out, a, grouped); err != nil {
				return err
			}
		}

	case attr.Value.Kind() == slog.KindGroup && key == "":
		if key == "STATE" {
			fmt.Printf("h %+v\nk %v\nv %+v\n", h, value.Kind(), value.Group())
		}
		for _, a := range value.Group() {
			err := h.appendAttr(out, a, grouped)
			if err != nil {
				fmt.Printf("%v\n", err)
				return err
			}
		}
	}

	return nil
}

type Loggable interface {
	WithLogger(*Logger)
}

func cleanAttr(groups []string, attr slog.Attr) slog.Attr {
	// string paths and packages
	return attr // TODO
}

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
