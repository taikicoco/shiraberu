package spinner

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinner_StartStop(t *testing.T) {
	var buf bytes.Buffer
	s := &Spinner{
		message: "Loading...",
		writer:  &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	s.Start()
	time.Sleep(100 * time.Millisecond)
	s.Stop()

	// Should have written something
	if buf.Len() == 0 {
		t.Error("Spinner should have written output")
	}
}

func TestSpinner_Success(t *testing.T) {
	var buf bytes.Buffer
	s := &Spinner{
		message: "Loading...",
		writer:  &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Success("Done!")

	output := buf.String()
	if !strings.Contains(output, "✓ Done!") {
		t.Errorf("Output should contain success message, got: %q", output)
	}
}

func TestSpinner_Fail(t *testing.T) {
	var buf bytes.Buffer
	s := &Spinner{
		message: "Loading...",
		writer:  &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Fail("Error!")

	output := buf.String()
	if !strings.Contains(output, "✗ Error!") {
		t.Errorf("Output should contain fail message, got: %q", output)
	}
}

func TestSpinner_DoubleStart(t *testing.T) {
	var buf bytes.Buffer
	s := &Spinner{
		message: "Loading...",
		writer:  &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	s.Start()
	s.Start() // Should not panic or create multiple goroutines
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestSpinner_DoubleStop(t *testing.T) {
	var buf bytes.Buffer
	s := &Spinner{
		message: "Loading...",
		writer:  &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}

	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	s.Stop() // Should not panic
}

func TestNew(t *testing.T) {
	s := New("Test message")
	if s.message != "Test message" {
		t.Errorf("Message: got %q, want %q", s.message, "Test message")
	}
	if s.writer == nil {
		t.Error("Writer should not be nil")
	}
}
