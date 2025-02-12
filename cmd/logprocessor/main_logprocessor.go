package main

import (
	"context"
	"fmt"
	"os"

	"ConcurrentChallenges/internal/logprocessor"
	"ConcurrentChallenges/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg := logprocessor.NewConfig()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	logg, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer func(logg *zap.Logger) {
		err := logg.Sync()
		if err != nil {

		}
	}(logg)

	logg.Info("Starting log processing",
		zap.String("file", cfg.FilePath),
		zap.Int("workers", cfg.WorkerCount),
		zap.Int("chunkSize", cfg.ChunkSize),
	)

	processor := logprocessor.NewProcessor(cfg, logg)
	if err := processor.Process(ctx); err != nil {
		logg.Error("Processing failed", zap.Error(err))
		os.Exit(1)
	}

	logg.Info("Log processing completed successfully")
}
