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

func TestMain(t *testing.T) {
	ctx := context.Background()
	mach := vm.New(
		vm.WithLogger(log.Default()),
	)

	t.Log(mach.String())
	t.Log(mach.Reg.String())
	t.Log(mach.PC.String())
	t.Log(mach.IR.String())
	t.Log(mach.MCR.String())
	t.Log(mach.Mem.MAR.String())
	t.Log(mach.Mem.MDR.String())
	t.Log(mach.SSP.String())
	t.Log(mach.USP.String())

	ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second, errors.New("bad"))
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

	select {
	case <-ctx.Done():
		err := ctx.Err()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Error(err)
		} else if errors.Is(err, context.Canceled) {
			return
		}
	}
}
