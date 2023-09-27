// Termtest is a testing tool for Unix terminal I/O. Lacking simple PTY support, running this tool
// manually is easier than writing automated tests.
package main

import (
	"context"
	"os"
	"time"

	logl "github.com/smoynes/elsie/internal/log"
	"github.com/smoynes/elsie/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

var log = logl.DefaultLogger()

func main() {
	var (
		ctx      = context.Background()
		keyboard = vm.NewKeyboard()
		display  = vm.NewDisplay()
	)

	display.Init(nil, nil)

	ctx, _, cancel := tty.WithConsole(ctx, keyboard, display)
	defer cancel()

	poll := time.Tick(100 * time.Millisecond)
	timeout := time.After(5 * time.Second)

	select {
	case <-ctx.Done():
		log.Debug("cause", context.Cause(ctx))
	default:
	}

	log.Info("Polling keyboard. Type keys.")

	display.Write(vm.Register('\n'))

	for {
		select {
		case <-poll:
			key, err := keyboard.Read(vm.KBDRAddr)
			if err != nil {
				log.Error(err.Error())
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
				log.Error(cause.Error())
			} else {
				log.Info("Done")
			}
		}
	}
}
