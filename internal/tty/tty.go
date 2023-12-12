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

// Console is a serial console for the machine simulated using Unix terminal I/O[^1]. It adapts the
// machine's (virtual) keyboard and display devices for use on contemporary systems[^2].
//
// Keys pressed on the console are copied to the keyboard device, after waiting for device
// interrupts to be enabled. Likewise, writes to the display device are output on the terminal.
//
// [1]: See: tty(4), termios(4).
// [2]: These systems, themselves, emulating electromecahnical teletype devices, of course.
type Console struct {
	in    *os.File
	out   *term.Terminal
	fd    int
	state *term.State

	// I/O buffers.
	keyCh  chan uint8
	termCh chan rune
}

// ErrNoTTY is returned if standard input is not a terminal. In this case, asynchronous I/O is
// not supported by the console.
var ErrNoTTY error = errors.New("console: not a TTY")

// ConsoleContext creates a Console context with the standard streams. Calling cancel will restore the
// terminal state and release resources.
func ConsoleContext(parent context.Context, keyboard *vm.Keyboard, display *vm.DisplayDriver) (
	context.Context, *Console, context.CancelFunc,
) {
	ctx, cause := context.WithCancelCause(parent)

	console, err := NewConsole(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		cause(err)

		return ctx, console, func() { cause(err) }
	}

	go console.readTerminal(ctx, cause)
	go console.updateKeyboard(ctx, keyboard, cause)
	go console.updateTerminal(ctx, display, cause)

	return ctx, console, console.Restore
}

// WithTerminal returns an option function that configures the machine to use the console.
func (c Console) WithTerminal(parent context.Context) vm.OptionFn {
	ctx, cause := context.WithCancelCause(parent)

	return func(machine *vm.LC3, late bool) {
		if !late {
			go c.readTerminal(ctx, cause)

			kbd := machine.Mem.Devices.Get(vm.KBDRAddr)
			if kbd == nil {
				panic("keyboard not found")
			} else if kbd, ok := kbd.(*vm.Keyboard); !ok {
				panic(kbd)
			} else {
				go c.updateKeyboard(ctx, kbd, cause)
			}

			disp := machine.Mem.Devices.Get(vm.DDRAddr)
			if disp, ok := disp.(*vm.DisplayDriver); !ok {
				panic(disp)
			} else {
				go c.updateTerminal(ctx, disp, cause)
			}
		}
	}
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
		return nil, fmt.Errorf("%w: %w", ErrNoTTY, err)
	}

	cons := Console{
		fd:     fd,
		in:     sin,
		out:    term.NewTerminal(sin, ""),
		state:  saved,
		keyCh:  make(chan uint8, 1),
		termCh: make(chan rune, 80),
	}

	err = cons.setTerminalParams(1, 0)
	if err != nil {
		return nil, err
	}

	return &cons, nil
}

// Press injects a key press into the input stream.
func (c Console) Press(key byte) {
	c.keyCh <- key
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

// readTerminal reads bytes from the terminal and writes them to the key channel until the context
// is cancelled. If reading from the terminal fails, the cancel is called.
func (c Console) readTerminal(ctx context.Context, cancel context.CancelCauseFunc) {
	buf := bufio.NewReader(c.in)

	// Make terminal input block on reads.
	_ = syscall.SetNonblock(c.fd, false)

	for { // ever and ever
		select {
		case <-ctx.Done():
			return
		default:
		}

		b, err := buf.ReadByte()
		if err != nil {
			cancel(err) // TODO: Is it right to cancel the context on errors?
			return
		}

		select {
		case <-ctx.Done():
			return
		case c.keyCh <- b:
		}
	}
}

// updateKeyboard takes keys from the key channel and updates the keyboard device for each key. The
// function blocks until the context is cancelled.
func (c Console) updateKeyboard(ctx context.Context, kbd *vm.Keyboard, _ context.CancelCauseFunc) {
	for { // you, a gift.
		select {
		case <-ctx.Done():
			return
		case key := <-c.keyCh:
			// Blocks until there is space in keyboard buffer.
			kbd.Update(uint16(key))
		}
	}
}

// updateTerminal waits for writes to the display and outputs the display data to the terminal.
func (c Console) updateTerminal(ctx context.Context, disp *vm.DisplayDriver, cancel context.CancelCauseFunc) {
	// Listen to the display device.
	disp.Listen(
		func(char uint16) {
			select {
			case <-ctx.Done():
			case c.termCh <- rune(char):
			default:
				// dropped signal
			}
		},
	)

	for { // SPARTA!
		select {
		case char := <-c.termCh:
			if _, err := fmt.Fprintf(c.out, "%c", char); err != nil {
				cancel(err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
