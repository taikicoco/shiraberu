package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	message string
	writer  io.Writer
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	running bool
}

func New(message string) *Spinner {
	return &Spinner{
		message: message,
		writer:  os.Stdout,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Fprintf(s.writer, "\r\033[K")
				close(s.done)
				return
			case <-ticker.C:
				fmt.Fprintf(s.writer, "\r%s %s", frames[i%len(frames)], s.message)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
}

func (s *Spinner) Success(message string) {
	s.Stop()
	fmt.Fprintf(s.writer, "✓ %s\n", message)
}

func (s *Spinner) Fail(message string) {
	s.Stop()
	fmt.Fprintf(s.writer, "✗ %s\n", message)
}
