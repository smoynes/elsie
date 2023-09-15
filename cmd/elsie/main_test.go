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

const timeout = time.Second

func TestMain(t *testing.T) {
	ctx := context.Background()
	mach := vm.New(
		vm.WithLogger(log.Default()),
	)

	ctx, cancel := context.WithTimeoutCause(ctx, timeout, errors.New("bad"))
	defer cancel()

	go func() {
		t.Logf("running")
		err := mach.Run(ctx)

		if !errors.Is(err, vm.ErrNoDevice) {
			t.Error(err)
		}

		t.Logf("ranned: err: %s, ctx: %v", err, ctx.Err())

		cancel()
	}()

	start := time.Now()

loop:
	for {
		t.Log(mach.String())
		t.Log(mach.Reg.String())
		t.Log(mach.PC.String())
		t.Log(mach.IR.String())
		t.Log(mach.MCR.String())
		t.Log(mach.Mem.MAR.String())
		t.Log(mach.Mem.MDR.String())
		t.Log(mach.SSP.String())
		t.Log(mach.USP.String())
		t.Log("")
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			break loop
		}
	}

	cause, err := context.Cause(ctx), ctx.Err()
	switch {
	case errors.Is(cause, context.Canceled):
		// ok
	case errors.Is(err, context.DeadlineExceeded):
		// too slow
		t.Errorf("run took more than %v", timeout)
	default:
		t.Errorf("unexpected error: %s", err)
	}

	t.Logf("took %s", time.Since(start))
	t.Fail()
}
