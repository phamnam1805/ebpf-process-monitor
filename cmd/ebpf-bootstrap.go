package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ebpf-bootstrap/internal/probe"
)

var (
    minDuration = flag.Int("d", 0, "Minimum process duration (ms) to report")
)

func signalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nCaught SIGINT... Exiting")
		cancel()
	}()
}

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalHandler(cancel)

	if err := probe.Run(ctx, *minDuration); err != nil {
		log.Fatalf("Failed running the probe: %v", err)
	}
}