package main_test

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/smoynes/elsie/internal/vm"
)

func init() {
	log.Default().SetOutput(io.Discard)
}

type testHarness struct {
	*testing.T
}

func (testHarness) Make() *vm.LC3 {
	return vm.New(
		vm.WithLogger(log.Default()),
	)
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
	// less than 100 ms.
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
			case <-time.After(timeout / 6):
				// This seems... racy.
				t.Log(machine.String())
				t.Log(machine.REG.String())
				t.Log("")
			case <-ctx.Done():
				cancel()
			}
		}
	}()

	go func() {
		t.Logf("running")

		err := machine.Run(ctx)

		if !errors.Is(err, vm.ErrNoDevice) {
			t.Error(err)
			cause(err)
		} else if ctx.Err() != nil {
			cause(ctx.Err())
		}

		t.Logf("ranned: err: %s", context.Cause(ctx))
		cancel()
	}()

	<-ctx.Done()

	elapsed := time.Since(start)
	err := context.Cause(ctx)

	switch {
	case err == nil:
		t.Logf("test: ok, elapsed time: %s", elapsed)
	case errors.Is(err, context.Canceled):
		t.Log(machine.String())
		t.Log(machine.REG.String())
		t.Log(machine.PC.String())
		t.Log(machine.IR.String())
		t.Log(machine.MCR.String())
		t.Log(machine.Mem.MAR.String())
		t.Log(machine.Mem.MDR.String())
		t.Log(machine.SSP.String())
		t.Log(machine.USP.String())
		t.Log("")
	case errors.Is(err, errTestTimeout):
		t.Error(errTestTimeout)
	case errors.Is(err, context.DeadlineExceeded):
		t.Errorf("%s: elapsed: %s", err, elapsed)
	default:
		t.Errorf("unexpected error: %s", err)
	}
}
