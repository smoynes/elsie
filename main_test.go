package main_test

import (
	"bufio"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/vm"
)

var logBuffer bufio.Writer

func init() {
}

type testHarness struct {
	*testing.T
}

func (testHarness) Make() *vm.LC3 {
	return vm.New()
}

var (
	// timeout is how long to wait for the machine to stop running. It is very likely to take
	// less than 200 ms.
	timeout    = 1 * time.Second
	statusTick = 25 * time.Millisecond
)

// Context creates a test context. The context is cancelled after a timeout.
func (testHarness) Context() (ctx context.Context,
	cause context.CancelCauseFunc,
	cancel context.CancelFunc,
) {
	ctx = context.Background()
	ctx, cause = context.WithCancelCause(ctx)
	ctx, cancel = context.WithTimeout(ctx, timeout)

	return ctx, func(err error) {
		logBuffer.Flush()
		cause(err)
	}, cancel
}

func TestMain(tt *testing.T) {
	t := testHarness{tt}
	start := time.Now()
	machine := t.Make()
	// Buffer log output. Without buffering, for each emitted log call, a write is issued to the
	// output stream. By buffering a little bit, the test is about 10x faster.
	log.LogLevel.Set(log.Error)

	ctx, cause, cancel := t.Context()
	defer cancel()

	go func() {
		for {
			select {
			case <-time.After(statusTick):
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
		select {
		case <-ctx.Done():
			return
		default:
		}

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
		t.Errorf("test: error: %s: elapsed: %s, %s", err, elapsed, timeout)
	}
}
