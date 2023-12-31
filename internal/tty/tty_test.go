// Package tty_test tries to test ttys.
//
// The test is skipped when stdin is not a terminal (ErrNoTTY). Notably, this includes when run with
// "go test" because it redirects tests' standard input/output streams. You can test it by building
// a test binary and running it directly:
//
//	$ go test -c && ./tty.test
package tty_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smoynes/elsie/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

type testHarness struct {
	*testing.T
}

const timeout = 100 * time.Millisecond

func (testHarness) Context() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeoutCause(ctx, timeout, context.DeadlineExceeded)
}

func TestTerminal(tt *testing.T) {
	t := testHarness{tt}
	kbd := vm.NewKeyboard()
	display := vm.NewDisplay()
	driver := vm.NewDisplayDriver(display)

	display.Init(nil, nil)

	ctx, cancel := t.Context()
	defer cancel()

	ctx, console, cancel := tty.ConsoleContext(ctx, kbd, driver)
	defer cancel()

	if err := context.Cause(ctx); errors.Is(err, tty.ErrNoTTY) {
		t.Skipf("error: %s", context.Cause(ctx))
		t.SkipNow()
	}

	pressed := make(chan struct{})
	_, _ = kbd.Read(vm.KBDRAddr)

	go func() {
		defer close(pressed)

		_, err := kbd.Read(vm.KBDRAddr)
		if err != nil {
			cancel()
			return
		}
	}()

	go func() {
		console.Press('!')
	}()

	display.Write('\n')
	display.Write('⍝')
	display.Write('\n')

	select {
	case <-ctx.Done(): // Just wait.
	case <-pressed:
	}

	cancel()

	if err := ctx.Err(); err != nil {
		t.Errorf("cause: %s", err)
	}
}
