package main

import (
	"context"

	"acc-dp/producer/pipeline"
)

func runProducer(ctx context.Context) error {
	return pipeline.Run(ctx)
}
