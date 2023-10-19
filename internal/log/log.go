// Package log provides logging output.
package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	// DefaultLogger returns the default, global logger. During application startup components can
	// call DefaultLogger and cache the result. The default will not change at runtime.
	DefaultLogger = func() *Logger { return NewFormattedLogger(os.Stderr) }

	// SetDefault overrides the default logger.
	SetDefault = slog.SetDefault

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

// Options for log handlers.
var Options = &slog.HandlerOptions{
	AddSource:   true,
	Level:       LogLevel,
	ReplaceAttr: func(_ []string, attr Attr) Attr { return attr },
}

// NewHandler creates and initializes a Handler with a writer.
func NewHandler(out io.Writer) *Handler {
	h := Handler{
		out:  out,
		mut:  new(sync.Mutex),
		opts: Options,
	}

	return &h
}

// Enabled returns true if the level is greater than the current logging level.
func (h *Handler) Enabled(ctx context.Context, level Level) bool {
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
		fmt.Fprintf(out, "%10s : %s\n", "TIMESTAMP", rec.Time.Format(time.RFC3339Nano))
	}

	fmt.Fprintf(out, "%10s : %s\n", "LEVEL", rec.Level.String())

	if h.opts.AddSource && rec.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{rec.PC})
		f, _ := frames.Next()
		_, file := path.Split(f.File)
		fmt.Fprintf(out, "%10s : %s:%d\n", "SOURCE", file, f.Line)

		if f.Func != nil {
			splits := strings.Split(f.Function, "/")
			fmt.Fprintf(out, "%10s : %s\n", "FUNCTION", splits[len(splits)-1])
		}
	}

	fmt.Fprintf(out, "%10s : %s\n", "MESSAGE", rec.Message)

	for _, a := range h.attrs {
		err := h.appendAttr(out, a, false)
		if err != nil {
			panic(err)
		}
	}

	rec.Attrs(func(attr Attr) bool {
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
func (h *Handler) WithAttrs(attrs []Attr) slog.Handler {
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
	case attr.Equal(Attr{}):
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
		for _, a := range value.Group() {
			err := h.appendAttr(out, a, grouped)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Loggable interface {
	WithLogger(*Logger)
}

type (
	Attr   = slog.Attr
	Level  = slog.Level
	Logger = slog.Logger
	Value  = slog.Value
)

var (
	String      = slog.String
	StringValue = slog.StringValue
	Group       = slog.Group
	GroupValue  = slog.GroupValue
	Any         = slog.Any
	AnyValue    = slog.AnyValue
)

const (
	Debug = slog.LevelDebug
	Info  = slog.LevelInfo
	Warn  = slog.LevelWarn
	Error = slog.LevelError
)
