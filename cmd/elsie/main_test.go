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
func (testHarness) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(context.Background(), timeout, errTestTimeout)
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
	ctx, done := t.Context()
	defer done()

	machine := t.Make()

	go func() {
		t.Logf("running")

		err := machine.Run(ctx)

		if !errors.Is(err, vm.ErrNoDevice) {
			t.Error(err)
		}
		time.Sleep(2 * time.Second)

		done()

		t.Logf("ranned: err: %s, ctx: %v", err, ctx.Err())
	}()

	start := time.Now()

loop:
	for {
		select {
		case <-time.After(20 * time.Millisecond):
		case <-ctx.Done():
			break loop
		}

		// This seems... racy.
		t.Log(machine.String())
		t.Log(machine.Reg.String())
		t.Log(machine.PC.String())
		t.Log(machine.IR.String())
		t.Log(machine.MCR.String())
		t.Log(machine.Mem.MAR.String())
		t.Log(machine.Mem.MDR.String())
		t.Log(machine.SSP.String())
		t.Log(machine.USP.String())
		t.Log("")
	}

	elapsed := time.Since(start)
	err := context.Cause(ctx)

	switch {
	case errors.Is(err, context.Canceled):
		t.Log(machine.String())
		t.Log(machine.Reg.String())
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

	t.Logf("test: elapsed: %s", elapsed)
}
