package primepalindrome

import (
	"flag"
	"fmt"
	"time"
)

// Config holds the configuration for the prime palindrome challenge.
type Config struct {
	N       int
	Timeout time.Duration
}

// NewConfig parses flags and returns a Config.
func NewConfig() *Config {
	n := flag.Int("N", 10, "Number of prime palindromic numbers to find")
	timeout := flag.Duration("timeout", 10*time.Second, "Timeout for processing")
	flag.Parse()

	if *n <= 0 {
		return &Config{
			N:       10,
			Timeout: *timeout,
		}
	}

	return &Config{
		N:       *n,
		Timeout: *timeout,
	}
}

// Validate checks the configuration.
func (c *Config) Validate() error {
	if c.N <= 0 {
		return fmt.Errorf("N must be positive, got %d", c.N)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Timeout)
	}
	return nil
}
