package main_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

func init() {
	log.DefaultLogger = func() *log.Logger {
		return log.New(io.Discard)
	}
}

type testHarness struct {
	*testing.T
}

func (testHarness) Make() *vm.LC3 {
	return vm.New()
}

// Context creates a test context. The context is cancelled after a timeout.
func (testHarness) Context() (ctx context.Context,
	cause context.CancelCauseFunc,
	cancel context.CancelFunc,
) {
	ctx = context.Background()
	ctx, cause = context.WithCancelCause(ctx)
	ctx, cancel = context.WithTimeout(ctx, timeout)

	return ctx, cause, cancel
}

var (
	// timeout is how long to wait for the machine to stop running. It is very likely to take
	// less than 200 ms.
	timeout = time.Second

	// errTestTimeout is the cause of a context cancellation for a timeout.
	errTestTimeout = errors.New("test: timeout")
)

func TestMain(tt *testing.T) {
	t := testHarness{tt}
	start := time.Now()
	machine := t.Make()

	ctx, cause, cancel := t.Context()
	defer cancel()

	go func() {
		for {
			select {
			case <-time.After(80 * time.Millisecond):
				// This seems... racy.
				t.Log("in progress, PC:", machine.PC.String(), "MCR:", machine.MCR.String())
			case <-ctx.Done():
				cancel()
			}
		}
	}()

	go func() {
		t.Logf("running")

		err := machine.Run(ctx)

		// We expect the program to eventually reach protected I/O address space.
		if !errors.Is(err, vm.ErrNoDevice) {
			t.Error(err)
			cause(err)
		} else if ctx.Err() != nil {
			cause(ctx.Err())
		}

		cancel()
	}()

	<-ctx.Done()

	elapsed := time.Since(start)
	err := context.Cause(ctx)

	switch {
	case err == nil:
		t.Logf("test: ok, elapsed: %s", elapsed)
	case errors.Is(err, context.Canceled):
		t.Logf("test: ok, err: %s, elapsed: %s", err, elapsed)
	default:
		err = context.Cause(ctx)
		t.Errorf("%s: elapsed: %s", err, elapsed)
	}
}
