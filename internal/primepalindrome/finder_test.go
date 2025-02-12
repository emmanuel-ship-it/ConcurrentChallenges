package primepalindrome

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFinder_Find(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name     string
		n        int
		expected []int
		wantSum  int
		wantErr  bool
	}{
		{
			name:     "first 5 prime palindromes",
			n:        5,
			expected: []int{2, 3, 5, 7, 11},
			wantSum:  28,
			wantErr:  false,
		},
		{
			name:     "invalid N",
			n:        -1,
			expected: nil,
			wantSum:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				N:       tt.n,
				Timeout: 5 * time.Second,
			}
			finder := NewFinder(cfg, logger)
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			numbers, sum, err := finder.Find(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			sort.Ints(numbers)
			assert.Equal(t, tt.expected, numbers)
			assert.Equal(t, tt.wantSum, sum)
		})
	}
}

func TestIsPrime(t *testing.T) {
	tests := []struct {
		num      int
		expected bool
	}{
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{9, false},
		{11, true},
		{0, false},
		{1, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, isPrime(tt.num))
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		num      int
		expected bool
	}{
		{11, true},
		{121, true},
		{123, false},
		{12321, true},
		{12345, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, isPalindrome(tt.num))
	}
}
