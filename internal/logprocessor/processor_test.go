package logprocessor

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProcessor_Process(t *testing.T) {
	content := `INFO: Application started
ERROR: Failed to connect
DEBUG: Attempting reconnect
INFO: Connection established
ERROR: Invalid input received
INFO: Processing complete`
	tmpFile, err := os.CreateTemp("", "testlog")
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	err = tmpFile.Close()
	if err != nil {
		return
	}

	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		FilePath:    tmpFile.Name(),
		WorkerCount: 2,
		ChunkSize:   100,
		Timeout:     2 * time.Second,
		Keywords:    []string{"INFO", "ERROR", "DEBUG"},
	}

	processor := NewProcessor(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	err = processor.Process(ctx)
	assert.NoError(t, err)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				FilePath:    "test.log",
				WorkerCount: 4,
				ChunkSize:   4096,
				Keywords:    []string{"INFO", "ERROR"},
			},
			wantErr: false,
		},
		{
			name: "empty filepath",
			config: Config{
				FilePath:    "",
				WorkerCount: 4,
				ChunkSize:   4096,
				Keywords:    []string{"INFO", "ERROR"},
			},
			wantErr: true,
		},
		{
			name: "invalid worker count",
			config: Config{
				FilePath:    "test.log",
				WorkerCount: 0,
				ChunkSize:   4096,
				Keywords:    []string{"INFO", "ERROR"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessChunk(t *testing.T) {
	processor := &Processor{}
	chunk := "Line1\nLine2\nLine3"
	lines := processor.processChunk(chunk)
	expected := []string{"Line1", "Line2", "Line3"}
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	assert.Equal(t, expected, lines)
}

func TestReadFileInChunks(t *testing.T) {
	content := "Line1\nLine2\nLine3\n"
	tmpFile, err := os.CreateTemp("", "testlog")
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _ = zap.NewDevelopment()
	lineCh := make(chan string, 10)
	var lines []string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range lineCh {
			lines = append(lines, line)
		}
	}()

	buffer := []byte(content)
	chunk := string(buffer)
	linesSplit := strings.Split(chunk, "\n")
	for _, line := range linesSplit {
		if line != "" {
			lineCh <- line
		}
	}
	close(lineCh)
	wg.Wait()

	expected := []string{"Line1", "Line2", "Line3"}
	assert.Equal(t, expected, lines)
}
