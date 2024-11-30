package main

import (
	"chunky/cmd"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-signalChan
		cancel()
	}()

	err := cmd.Execute(ctx)
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
}
