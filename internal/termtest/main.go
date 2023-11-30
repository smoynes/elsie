// Termtest is a testing tool for Unix terminal I/O. Lacking simple PTY support, running this tool
// manually is easier than writing automated tests.
package main

import (
	"context"
	"os"
	"time"

	"github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

var logger = log.DefaultLogger()

func main() {
	var (
		ctx           = context.Background()
		keyboard      = vm.NewKeyboard()
		display       = vm.NewDisplay()
		displayDriver = vm.NewDisplayDriver(display)
	)

	display.Init(nil, nil)

	ctx, _, cancel := tty.ConsoleContext(ctx, keyboard, displayDriver)
	defer cancel()

	poll := time.Tick(100 * time.Millisecond)
	timeout := time.After(5 * time.Second)

	select {
	case <-ctx.Done():
		logger.Debug("cause", context.Cause(ctx))
	default:
	}

	logger.Info("Polling keyboard. Type keys.")

	display.Write(vm.Register('\n'))

	for {
		select {
		case <-poll:
			key, err := keyboard.Read(vm.KBDRAddr)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			if key != 0x0000 {
				display.Write(vm.Register(key))
			}
		case <-timeout:
			cancel()
			return
		case <-ctx.Done():
			if ctx.Err() != nil {
				cause := context.Cause(ctx)
				logger.Error(cause.Error())
			} else {
				logger.Info("Done")
			}
		}
	}
}
