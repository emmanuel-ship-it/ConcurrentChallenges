package main

import (
	"context"
	"fmt"
	"os"

	"ConcurrentChallenges/internal/primepalindrome"
	"ConcurrentChallenges/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg := primepalindrome.NewConfig()

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

	finder := primepalindrome.NewFinder(cfg, logg)
	numbers, sum, err := finder.Find(ctx)
	if err != nil {
		logg.Error("Finding prime palindromes failed", zap.Error(err))
		os.Exit(1)
	}

	logg.Info("Found prime palindromic numbers",
		zap.Int("count", len(numbers)),
		zap.Ints("numbers", numbers),
		zap.Int("sum", sum),
	)
	fmt.Printf("The first %d prime palindromic numbers are: %v\n", cfg.N, numbers)
	fmt.Printf("Their sum is: %d\n", sum)
}
