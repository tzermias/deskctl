package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestVerboseToggle(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)

	// Test verbose disabled (default)
	SetVerbose(false)
	Printf("test message")
	if buf.Len() > 0 {
		t.Error("Expected no output when verbose is disabled")
	}

	// Test verbose enabled
	buf.Reset()
	SetVerbose(true)
	Printf("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Error("Expected output when verbose is enabled")
	}
}

func TestPrintln(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)

	// Test verbose disabled
	SetVerbose(false)
	Println("test message")
	if buf.Len() > 0 {
		t.Error("Expected no output when verbose is disabled")
	}

	// Test verbose enabled
	buf.Reset()
	SetVerbose(true)
	Println("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Error("Expected output when verbose is enabled")
	}
}

func TestErrorAlwaysPrints(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetVerbose(false)

	// Test Error
	Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Expected error to print even when verbose is disabled")
	}

	// Test Errorf
	buf.Reset()
	Errorf("error %s", "formatted")
	if !strings.Contains(buf.String(), "error formatted") {
		t.Error("Expected formatted error to print even when verbose is disabled")
	}
}

func TestIsVerbose(t *testing.T) {
	SetVerbose(false)
	if IsVerbose() {
		t.Error("Expected IsVerbose to return false")
	}

	SetVerbose(true)
	if !IsVerbose() {
		t.Error("Expected IsVerbose to return true")
	}
}

func TestConcurrentAccess(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetVerbose(true)

	// Test concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			Printf("concurrent message")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 messages
	count := strings.Count(buf.String(), "concurrent message")
	if count != 10 {
		t.Errorf("Expected 10 messages, got %d", count)
	}
}
