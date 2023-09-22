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
	ctx := context.Background()
	keyboard := vm.NewKeyboard()

	ctx, console, cancel := tty.WithConsole(ctx, keyboard)
	defer cancel()

	intr := time.Tick(100 * time.Millisecond)
	poll := time.Tick(500 * time.Millisecond)
	timeout := time.After(10 * time.Second)

	select {
	case <-ctx.Done():
		log.Fatal(context.Cause(ctx))
	default:
	}

	log.SetOutput(console.Writer())
	log.Printf("polling keyboard")

	for {
		select {
		case <-intr:
			if keyboard.InterruptRequested() {
				log.Printf("INT: %s", keyboard)
			}
		case <-poll:
			key, err := keyboard.Read(vm.KBDRAddr)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("polled: %c", rune(key))
		case <-timeout:
			log.Print("timeout")
			cancel()

			return
		case <-ctx.Done():
			log.Printf("done: %s", ctx.Err())
		}
	}
}
