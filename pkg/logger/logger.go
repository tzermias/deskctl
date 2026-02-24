package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	mu      sync.RWMutex
	verbose bool
	logger  *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

// SetVerbose enables or disables verbose logging
func SetVerbose(v bool) {
	mu.Lock()
	defer mu.Unlock()
	verbose = v
}

// IsVerbose returns whether verbose logging is enabled
func IsVerbose() bool {
	mu.RLock()
	defer mu.RUnlock()
	return verbose
}

// SetOutput sets the output destination for the logger
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	logger.SetOutput(w)
}

// Printf prints a log message only if verbose mode is enabled
func Printf(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if verbose {
		logger.Printf(format, v...)
	}
}

// Println prints a log message only if verbose mode is enabled
func Println(v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if verbose {
		logger.Println(v...)
	}
}

// Error always prints error messages regardless of verbose mode
func Error(v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	logger.Println(v...)
}

// Errorf always prints error messages regardless of verbose mode
func Errorf(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	logger.Printf(format, v...)
}
