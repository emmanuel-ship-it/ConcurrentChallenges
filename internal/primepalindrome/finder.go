package primepalindrome

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Finder is responsible for finding prime palindromic numbers.
type Finder struct {
	config *Config
	logger *zap.Logger
}

// NewFinder returns a new Finder.
func NewFinder(config *Config, logger *zap.Logger) *Finder {
	return &Finder{
		config: config,
		logger: logger,
	}
}

// Find uses concurrency to generate and test candidate numbers until it collects at least N prime palindromic numbers.
// Instead of canceling immediately upon receiving N results, we wait briefly to allow slower candidates (like 2) to be processed.
// Finally, we sort all received results and return the smallest N numbers and their sum.
func (f *Finder) Find(ctx context.Context) ([]int, int, error) {
	if err := f.config.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create a child context that we can cancel after a short wait.
	childCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	candidateCh := make(chan int, 100)
	resultCh := make(chan int, 100)

	var wg sync.WaitGroup
	workerCount := runtime.NumCPU()
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go f.worker(childCtx, candidateCh, resultCh, &wg)
	}

	go f.generateCandidates(childCtx, candidateCh)

	// Instead of canceling as soon as we receive N numbers,
	// wait a short period to allow all workers to process early candidates.
	waitDuration := 100 * time.Millisecond
	select {
	case <-time.After(waitDuration):
	case <-childCtx.Done():
	}

	// Cancel further candidate generation.
	cancelFunc()
	wg.Wait()
	close(resultCh)

	var numbers []int
	for num := range resultCh {
		numbers = append(numbers, num)
	}

	sort.Ints(numbers)
	if len(numbers) < f.config.N {
		return nil, 0, fmt.Errorf("only found %d prime palindromic numbers", len(numbers))
	}
	selected := numbers[:f.config.N]
	sum := 0
	for _, n := range selected {
		sum += n
	}
	return selected, sum, nil
}

func (f *Finder) worker(ctx context.Context, candidateCh <-chan int, resultCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case num, ok := <-candidateCh:
			if !ok {
				return
			}
			if isPrime(num) && isPalindrome(num) {
				select {
				case resultCh <- num:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (f *Finder) generateCandidates(ctx context.Context, candidateCh chan<- int) {
	defer close(candidateCh)
	for i := 2; ; i++ {
		select {
		case <-ctx.Done():
			return
		case candidateCh <- i:
		}
	}
}

func isPrime(num int) bool {
	if num < 2 {
		return false
	}
	if num == 2 {
		return true
	}
	if num%2 == 0 {
		return false
	}
	sqrtNum := int(math.Sqrt(float64(num)))
	for i := 3; i <= sqrtNum; i += 2 {
		if num%i == 0 {
			return false
		}
	}
	return true
}

func isPalindrome(num int) bool {
	str := fmt.Sprintf("%d", num)
	length := len(str)
	for i := 0; i < length/2; i++ {
		if str[i] != str[length-1-i] {
			return false
		}
	}
	return true
}
