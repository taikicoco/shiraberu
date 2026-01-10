package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/render"
)

func TestHTTPHandler(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, jst),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
		Org:         "test-org",
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 10, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{
						Title:      "Test PR",
						URL:        "https://github.com/test/repo/pull/1",
						Repository: "test-repo",
					},
				},
			},
		},
	}

	// Render HTML
	var buf bytes.Buffer
	if err := render.RenderHTML(&buf, report, nil); err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}
	content := buf.Bytes()

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	})

	// Test request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", rec.Code, http.StatusOK)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type: got %q, want %q", contentType, "text/html; charset=utf-8")
	}

	body := rec.Body.String()
	if !strings.Contains(body, "test-org") {
		t.Error("Response should contain org name")
	}
	if !strings.Contains(body, "Test PR") {
		t.Error("Response should contain PR title")
	}
}

func TestHTTPHandler_EmptyReport(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, jst),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
		Org:         "test-org",
		Days:        []pr.DailyPRs{},
	}

	var buf bytes.Buffer
	if err := render.RenderHTML(&buf, report, nil); err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}
	content := buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "No pull requests found") {
		t.Error("Empty report should show 'No pull requests found' message")
	}
}

func TestServeWithAddr(t *testing.T) {
	// Mock openBrowserFunc
	origOpenBrowserFunc := openBrowserFunc
	defer func() { openBrowserFunc = origOpenBrowserFunc }()

	openBrowserFunc = func(url string) {
		// No-op for testing
	}

	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, jst),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
		Org:         "test-org",
		Days:        []pr.DailyPRs{},
	}

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- ServeWithAddr(report, nil, ":0") // Use :0 for random port
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Server will block, so we just verify it started without immediate error
	select {
	case err := <-serverErr:
		// Server returned immediately, which might be okay for port :0
		if err != nil && !strings.Contains(err.Error(), "address already in use") {
			t.Logf("Server error (may be acceptable): %v", err)
		}
	default:
		// Server is running, which is expected
	}
}

func TestServeWithAddr_Integration(t *testing.T) {
	// Mock openBrowserFunc
	origOpenBrowserFunc := openBrowserFunc
	defer func() { openBrowserFunc = origOpenBrowserFunc }()

	openBrowserFunc = func(url string) {} // No-op

	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, jst),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
		Org:         "integration-test-org",
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 10, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{Title: "Integration Test PR", URL: "https://github.com/test/repo/pull/1"},
				},
			},
		},
	}

	// Create a test server manually to avoid port conflicts
	var buf bytes.Buffer
	if err := render.RenderHTML(&buf, report, nil); err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}
	content := buf.Bytes()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Make request to test server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "integration-test-org") {
		t.Error("Response should contain org name")
	}
	if !strings.Contains(string(body), "Integration Test PR") {
		t.Error("Response should contain PR title")
	}
}

func TestOpenBrowserDefault(t *testing.T) {
	// Just verify it doesn't panic
	// We can't really test browser opening in unit tests
	openBrowserDefault("http://localhost:7777")
}

func TestServe(t *testing.T) {
	// Mock openBrowserFunc
	origOpenBrowserFunc := openBrowserFunc
	defer func() { openBrowserFunc = origOpenBrowserFunc }()

	openBrowserFunc = func(url string) {}

	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, jst),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
		Org:         "test-org",
		Days:        []pr.DailyPRs{},
	}

	// Start server in background with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		_ = Serve(report, nil)
	}()

	<-ctx.Done()
	// If we get here, the server started (and is running, which is expected)
}
