// Termtest is a testing tool for Unix terminal I/O. Lacking simple PTY support, running this tool
// manually is easier than writing a automated test.
package main

import (
	"context"
	"log"
	"time"

	"github.com/smoynes/elsie/cmd/internal/tty"
	"github.com/smoynes/elsie/internal/vm"
)

func main() {
	var (
		ctx      = context.Background()
		keyboard = vm.NewKeyboard()
		display  = vm.NewDisplay()
	)

	display.Init(nil, nil)

	ctx, console, cancel := tty.WithConsole(ctx, keyboard, display)
	defer cancel()

	log.SetOutput(console.Writer())

	poll := time.Tick(100 * time.Millisecond)
	timeout := time.After(5 * time.Second)

	select {
	case <-ctx.Done():
		log.Fatal(context.Cause(ctx))
	default:
	}

	log.Printf("polling keyboard")

	display.Write(vm.Register('\n'))

	for {
		select {
		case <-poll:
			key, err := keyboard.Read(vm.KBDRAddr)
			if err != nil {
				log.Fatal(err)
			}

			if key != 0x0000 {
				display.Write(vm.Register(key))
			}
		case <-timeout:
			cancel()
			return
		case <-ctx.Done():
			log.Printf("done: %s", ctx.Err())
		}
	}
}
