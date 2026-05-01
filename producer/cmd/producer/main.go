package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"acc-dp/producer/pipeline"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := pipeline.Run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "producer failed: %v\n", err)
		os.Exit(1)
	}
}
