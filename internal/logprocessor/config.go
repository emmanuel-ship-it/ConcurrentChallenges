package logprocessor

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
    FilePath    string
    WorkerCount int
    ChunkSize   int
    Timeout     time.Duration
    Keywords    []string
}

func NewConfig() *Config {
    filePath := flag.String("file", "log.txt", "Path to the log file")
    workerCount := flag.Int("workers", 4, "Number of worker goroutines")
    chunkSize := flag.Int("chunkSize", 4096, "Chunk size in bytes for file reading")
    timeout := flag.Duration("timeout", 10*time.Second, "Timeout for processing the log file")
    keywordsArg := flag.String("keywords", "INFO,ERROR,DEBUG", "Comma-separated list of keywords to count")
    flag.Parse()

    keywords := strings.Split(*keywordsArg, ",")
    for i := range keywords {
        keywords[i] = strings.TrimSpace(keywords[i])
    }

    return &Config{
        FilePath:    *filePath,
        WorkerCount: *workerCount,
        ChunkSize:   *chunkSize,
        Timeout:     *timeout,
        Keywords:    keywords,
    }
}

func (c *Config) Validate() error {
    if c.FilePath == "" {
        return fmt.Errorf("file path cannot be empty")
    }
    if c.WorkerCount <= 0 {
        return fmt.Errorf("worker count must be positive, got %d", c.WorkerCount)
    }
    if c.ChunkSize <= 0 {
        return fmt.Errorf("chunk size must be positive, got %d", c.ChunkSize)
    }
    if len(c.Keywords) == 0 {
        return fmt.Errorf("at least one keyword must be specified")
    }
    return nil
}
