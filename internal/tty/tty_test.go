package main // tty

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testHarness struct {
	*testing.T
}

const timeout = time.Second

func (testHarness) Context() (ctx context.Context, cancel context.CancelFunc) {
	ctx = context.Background()
	ctx, cancel = context.WithTimeout(ctx, timeout)

	return ctx, cancel
}

func TestTerminal(tt *testing.T) {
	t := testHarness{tt}

	ctx, cancel := t.Context()
	defer cancel()

	ctx, console, cause := WithConsole(ctx)
	defer cause(nil)

	if err := context.Cause(ctx); errors.Is(err, errNoTTY) {
		t.Skipf("error: %s", context.Cause(ctx))
	}

	select {
	case ev := <-console.Keys():
		cause(nil)
		t.Logf("%v", ev)

	case <-ctx.Done():
		cause(nil)
		break
	}

	if err := context.Cause(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("cause: %s", err)
	}
}
