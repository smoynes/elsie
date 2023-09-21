package main // tty

import (
	"context"
	"errors"
	"log"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Type aliases to reduce symbol stutter.
type (
	Context         = context.Context
	ConsoleDoneFunc = context.CancelCauseFunc
)

// Console is simulated serial console using Unix terminal I/O for teletype emulation.
type Console struct {
	in     *os.File
	fd     int
	state  *term.State
	cancel ConsoleDoneFunc
	keyCh  chan uint16
	termCh chan []byte
}

var (
	errNoTTY error = errors.New("console: not a TTY")
)

func WithConsole(parent Context) (Context, *Console, ConsoleDoneFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	console, err := newConsole(os.Stdin, os.Stdout, cancel)

	if err != nil {
		cancel(err)
		return ctx, console, cancel
	}

	go console.readKeys(ctx, console.restore)

	return ctx, console, console.restore
}

func newConsole(in *os.File, out *os.File, cancel ConsoleDoneFunc) (*Console, error) {
	fd := int(in.Fd())

	if !term.IsTerminal(fd) {
		return nil, errNoTTY
	}

	saved, err := term.MakeRaw(fd)
	if err != nil {
		return nil, errNoTTY
	}

	cons := Console{
		in:     in,
		fd:     fd,
		keyCh:  make(chan uint16),
		termCh: make(chan []byte, 8),
		state:  saved,
		cancel: cancel,
	}

	err = cons.setTerminalParams(1, 0)
	if err != nil {
		return nil, err
	}

	return &cons, nil
}

func (c *Console) setTerminalParams(vmin, vtime byte) error {
	_ = syscall.SetNonblock(c.fd, true)

	termIO, err := unix.IoctlGetTermios(c.fd, unix.TIOCGETA)
	if err != nil {
		return err
	}

	termIO.Cc[unix.VMIN] = vmin
	termIO.Cc[unix.VTIME] = vtime

	err = unix.IoctlSetTermios(c.fd, unix.TIOCSETAW, termIO)
	if err != nil {
		return err
	}

	_ = os.Stdin.SetReadDeadline(time.Time{})

	return nil
}

func (c Console) readKeys(ctx Context, cancel ConsoleDoneFunc) {
	buf := make([]byte, 8)

	for { // ever and ever
		select {
		case <-ctx.Done():
			return
		default:
			_ = syscall.SetNonblock(c.fd, false)
			n, err := c.in.Read(buf)

			if err != nil {
				log.Printf("read error %#v", err)
				cancel(err)

				return
			}

			for i := 0; i < n; i++ {
				c.keyCh <- uint16(buf[i])
			}
		}
	}
}

func (c *Console) restore(err error) {
	_ = os.Stdin.SetReadDeadline(time.Now()) // Cancel any in progress blocking reads.
	_ = term.Restore(c.fd, c.state)
	c.cancel(err)
}

func (c Console) Keys() <-chan uint16 {
	return c.keyCh
}
