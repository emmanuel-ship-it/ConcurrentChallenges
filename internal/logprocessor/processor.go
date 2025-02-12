package logprocessor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// Processor is responsible for processing the log file.
type Processor struct {
	config *Config
	logger *zap.Logger
}

// NewProcessor returns a new Processor given a configuration and logger.
func NewProcessor(config *Config, logger *zap.Logger) *Processor {
	return &Processor{
		config: config,
		logger: logger,
	}
}

// Process opens the log file, reads it in chunks, spawns worker goroutines to count keyword occurrences,
// and aggregates the results.
func (p *Processor) Process(ctx context.Context) error {
	if err := p.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	file, err := os.Open(p.config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	lineCh := make(chan string, 100)
	resultCh := make(chan map[string]int, p.config.WorkerCount)
	errCh := make(chan error, 1)

	// Start worker goroutines.
	var wg sync.WaitGroup
	for i := 0; i < p.config.WorkerCount; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, lineCh, resultCh)
	}

	// Start file reader.
	go p.readFile(ctx, file, lineCh, errCh)

	// Close the result channel when workers are done.
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return p.aggregateResults(ctx, resultCh, errCh)
}

func (p *Processor) worker(ctx context.Context, wg *sync.WaitGroup, lineCh <-chan string, resultCh chan<- map[string]int) {
	defer wg.Done()

	counts := make(map[string]int)
	for _, keyword := range p.config.Keywords {
		counts[strings.ToUpper(keyword)] = 0
	}

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-lineCh:
			if !ok {
				// No more lines; send counts and return.
				resultCh <- counts
				return
			}
			upperLine := strings.ToUpper(line)
			for keyword := range counts {
				counts[keyword] += strings.Count(upperLine, keyword)
			}
		}
	}
}

func (p *Processor) aggregateResults(ctx context.Context, resultCh <-chan map[string]int, errCh <-chan error) error {
	totalCounts := make(map[string]int)

	// Check for any file reading error.
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("file reading error: %w", err)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	// Aggregate results from all workers.
	for result := range resultCh {
		for keyword, count := range result {
			totalCounts[keyword] += count
		}
	}

	// Sort and output results.
	type wordCount struct {
		Keyword string
		Count   int
	}
	var results []wordCount
	for keyword, count := range totalCounts {
		results = append(results, wordCount{Keyword: keyword, Count: count})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	for _, r := range results {
		p.logger.Info("Keyword count", zap.String("keyword", r.Keyword), zap.Int("count", r.Count))
		fmt.Printf("%s: %d\n", r.Keyword, r.Count)
	}

	return nil
}

func (p *Processor) readFile(ctx context.Context, file *os.File, lineCh chan<- string, errCh chan<- error) {
	defer close(lineCh)

	reader := bufio.NewReader(file)
	buffer := make([]byte, p.config.ChunkSize)
	var leftover string

	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				errCh <- err
				return
			}

			chunk := string(buffer[:n])
			if leftover != "" {
				chunk = leftover + chunk
			}

			lines := p.processChunk(chunk)
			// Send all complete lines (all except possibly the last).
			for i := 0; i < len(lines)-1; i++ {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				case lineCh <- lines[i]:
					p.logger.Debug("Line processed", zap.String("line", lines[i]))
				}
			}

			if err == io.EOF {
				if lines[len(lines)-1] != "" {
					lineCh <- lines[len(lines)-1]
				}
				errCh <- nil
				return
			}

			leftover = lines[len(lines)-1]
		}
	}
}

func (p *Processor) processChunk(chunk string) []string {
	return strings.Split(chunk, "\n")
}
