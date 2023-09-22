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

	"github.com/smoynes/elsie/cmd/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

type testHarness struct {
	*testing.T
}

const timeout = 100 * time.Millisecond

func (testHarness) Context() (context.Context, context.CancelCauseFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, context.DeadlineExceeded)

	return ctx, func(err error) {
		cancel()
	}
}

func TestTerminal(tt *testing.T) {
	t := testHarness{tt}
	kbd := vm.NewKeyboard()

	ctx, cancel := t.Context()
	defer cancel(nil)

	ctx, console, cancel := tty.WithConsole(ctx, kbd)
	defer cancel(nil)

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
			cancel(err)
			return
		}

		kbd.Wait()
	}()

	go func() {
		console.Press('!')
	}()

	select {
	case <-ctx.Done(): // Just wait.
	case <-pressed:
	}

	cancel(nil)

	if err := context.Cause(ctx); err != nil {
		t.Errorf("cause: %s", err)
	}
}
