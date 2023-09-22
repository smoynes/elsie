// Package tty provides terminal emulation.
package tty

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/smoynes/elsie/internal/vm"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Console is simulated serial console using Unix terminal I/O for teletype emulation. It adapts the
// machine's keyboard and display devices for use on modern systems that pretend to have much older
// electromechanical devices.
type Console struct {
	in    *os.File
	out   *term.Terminal
	fd    int
	state *term.State
	keyCh chan uint16
}

var (
	// ErrNoTTY is returned if standard input is not a terminal.
	ErrNoTTY error = errors.New("console: not a TTY")
)

// WithConsole creates a Console context with the standard streams. Calling cancel will restore the
// terminal state and release resources.
func WithConsole(parent Context, keyboard *vm.Keyboard) (Context, *Console, ConsoleDoneFunc) {
	ctx, cause := context.WithCancelCause(parent)
	console, err := NewConsole(os.Stdin, os.Stdout, os.Stderr)

	if err != nil {
		cause(err)
		return ctx, console, func() { cause(context.Canceled) }
	}

	go console.readTerminal(ctx, console.Restore)
	go console.updateKeyboard(ctx, keyboard, console.Restore)

	return ctx, console, console.Restore
}

// NewConsole creates a Console using the provided streams. If the input stream is not a terminal,
// ErrNoTTY is returned. Callers are responsible for calling [Restore] to return the terminal to its
// initial state.
func NewConsole(sin, sout, serr *os.File) (*Console, error) {
	fd := int(sin.Fd())

	if !term.IsTerminal(fd) {
		return nil, ErrNoTTY
	}

	saved, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNoTTY, err)
	}

	cons := Console{
		fd:    fd,
		in:    sin,
		out:   term.NewTerminal(sin, ""),
		state: saved,
		keyCh: make(chan uint16, 1),
	}

	err = cons.setTerminalParams(1, 0)
	if err != nil {
		return nil, err
	}

	return &cons, nil
}

// Press injects a key press into the input stream.
func (c Console) Press(key byte) {
	c.keyCh <- uint16(key)
}

// Writer returns an io.Writer that writes to the terminal.
func (c Console) Writer() io.Writer {
	return c.out
}

// Restore returns the terminal to its initial state and cancels in-progress reads.
func (c *Console) Restore() {
	_ = os.Stdin.SetReadDeadline(time.Now())
	_ = term.Restore(c.fd, c.state)
}

func (c *Console) setTerminalParams(vmin, vtime byte) error {
	_ = syscall.SetNonblock(c.fd, true)

	termIO, err := unix.IoctlGetTermios(c.fd, getTermiosIoctl)
	if err != nil {
		return err
	}

	termIO.Cc[unix.VMIN] = vmin
	termIO.Cc[unix.VTIME] = vtime

	err = unix.IoctlSetTermios(c.fd, setTermiosIoctl, termIO)
	if err != nil {
		return err
	}

	_ = os.Stdin.SetReadDeadline(time.Time{})

	return nil
}

func (c Console) readTerminal(ctx Context, cancel ConsoleDoneFunc) {
	buf := bufio.NewReader(c.in)

	_ = syscall.SetNonblock(c.fd, false)

	for { // ever and ever
		select {
		case <-ctx.Done():
			return
		default:
			b, err := buf.ReadByte()

			if err != nil {
				cancel()
				return
			}

			c.keyCh <- uint16(b)
		}
	}
}

func (c Console) updateKeyboard(ctx Context, kbd *vm.Keyboard, cancel ConsoleDoneFunc) {
	for { // you, a gift.
		select {
		case key := <-c.keyCh:
			kbd.Update(key)
			continue
		case <-ctx.Done():
			return
		}
	}
}

// Type aliases to reduce symbol stutter.
type (
	Context         = context.Context
	ConsoleDoneFunc = context.CancelFunc
)
