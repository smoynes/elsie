package main //tty

import (
	"context"
	"log"
	"time"

	"golang.org/x/term"
)

func main() {
	ctx := context.Background()
	ctx, console, cancel := WithConsole(ctx)
	defer cancel(nil)

	term := term.NewTerminal(console.in, "&^=>")

	log.SetOutput(term)

loop:
	for {
		select {
		case key := <-console.Keys():
			log.Printf("key: %x", key)
			continue loop

		case <-time.After(5 * time.Second):
			log.Print("timeout")
			cancel(context.DeadlineExceeded)
			break loop

		case <-ctx.Done():
			log.Printf("done: %s", ctx.Err())
			break loop
		}
	}
}
