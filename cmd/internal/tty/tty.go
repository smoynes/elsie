// Package tty provides terminal emulation.
package tty

import (
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
// devices.
type Console struct {
	in     *os.File
	out    *term.Terminal
	fd     int
	state  *term.State
	cancel ConsoleDoneFunc
	keyCh  chan uint16
}

var (
	// ErrNoTTY is returned
	ErrNoTTY error = errors.New("console: not a TTY")
)

// WithConsole creates a Console context with the standard streams. Calling cancel will restore the
// terminal state and release resources.
func WithConsole(parent Context, keyboard *vm.Keyboard) (
	ctx Context, console *Console, cancel ConsoleDoneFunc,
) {
	ctx, cancel = context.WithCancelCause(parent)
	console, err := NewConsole(os.Stdin, os.Stdout, os.Stderr)

	if err != nil {
		cancel(err)
		return ctx, console, cancel
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
		fd:     fd,
		in:     sin,
		out:    term.NewTerminal(sin, ""),
		state:  saved,
		cancel: func(_ error) {},

		keyCh: make(chan uint16),
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

// Restore returns the terminal to its initial state, cancels in-progress reads, and cancels the
// context.
func (c *Console) Restore(err error) {
	_ = os.Stdin.SetReadDeadline(time.Now()) // Cancel any in progress blocking reads.
	_ = term.Restore(c.fd, c.state)
	c.cancel(err)
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
	buf := make([]byte, 8)
	_ = syscall.SetNonblock(c.fd, false)

	for { // ever and ever
		select {
		case <-ctx.Done():
			return
		default:
			n, err := c.in.Read(buf)

			if err != nil {
				cancel(err)
				return
			}

			for i := 0; i < n; i++ {
				c.keyCh <- uint16(buf[i])
			}
		}
	}
}

func (c Console) updateKeyboard(ctx Context, kbd *vm.Keyboard, cancel ConsoleDoneFunc) {
	for { // you, a gift.
		var key uint16

		select {
		case <-ctx.Done():
			return
		case key = <-c.keyCh:
			break
		}

		kbd.Update(key)
	}
}

// Type aliases to reduce symbol stutter.
type (
	Context         = context.Context
	ConsoleDoneFunc = context.CancelCauseFunc
)
