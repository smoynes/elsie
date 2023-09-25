package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
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
		AddSource: true,
		Level:     LogLevel,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			return attr // TODO
		},
	}
)

// Handler implements slog.Handler to produce formatted log output.
type Handler struct {
	out io.Writer
	mut *sync.Mutex // Synchronizes writer.

	opts   slog.HandlerOptions
	indent int
	group  string
}

func FormattedLogger(out io.Writer) *Logger {
	return slog.New(NewHandler(out))
}

func NewHandler(out io.Writer) *Handler {
	h := Handler{
		out:    out,
		mut:    new(sync.Mutex),
		opts:   *logOptions,
		indent: 0,
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

	rec.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(out, attr)
		return true
	})

	fmt.Fprintln(out)
	h.mut.Lock()
	defer h.mut.Unlock()

	_, err := h.out.Write(out.Bytes())

	return err
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		mut:   h.mut,
		out:   h.out,
		opts:  h.opts,
		group: h.group + "::" + name,
	}
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		out:   h.out,
		mut:   h.mut,
		opts:  h.opts,
		group: h.group,
	}
}

func (h *Handler) appendAttr(out io.Writer, attr slog.Attr) error {
	attr.Value = attr.Value.Resolve()

	if h.group != "" {
		fmt.Fprintf(out, "  ")
	}

	switch {
	case attr.Equal(slog.Attr{}):
		return nil
	case attr.Value.Kind() != slog.KindGroup:
		//println("1", attr.String())
		fmt.Fprintf(out, "%10s : %v\n", attr.Key, attr.Value.Any())
	case attr.Value.Kind() == slog.KindGroup && attr.Key != "":
		//println("2", attr.String())
		fmt.Fprintf(out, "%10s :\n", attr.Key)
		h.group = attr.Key
		fallthrough

	case attr.Value.Kind() == slog.KindGroup && attr.Key == "":
		//println("3", attr.String())
		for _, attr := range attr.Value.Group() {
			h.appendAttr(out, attr)
		}
		h.group = ""
	}

	return nil
}

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
