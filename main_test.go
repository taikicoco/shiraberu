package main

import (
	"errors"
	"testing"

	apperrors "github.com/taikicoco/shiraberu/internal/errors"
)

// Tests for calcPreviousPeriod have been moved to internal/period/period_test.go

func TestParseBatchOptions(t *testing.T) {
	opts, err := parseBatchOptions("2026-05-01", "2026-05-31", "org", "json", "out.json")
	if err != nil {
		t.Fatalf("parseBatchOptions: %v", err)
	}
	if opts.Org != "org" || opts.Output != "out.json" {
		t.Errorf("opts: %+v", opts)
	}
	if got := opts.StartDate.Format(dateLayout); got != "2026-05-01" {
		t.Errorf("StartDate: got %s", got)
	}
	if got := opts.EndDate.Format(dateLayout); got != "2026-05-31" {
		t.Errorf("EndDate: got %s", got)
	}
}

func TestParseBatchOptions_UntilDefaultsToSince(t *testing.T) {
	opts, err := parseBatchOptions("2026-05-01", "", "org", "json", "")
	if err != nil {
		t.Fatalf("parseBatchOptions: %v", err)
	}
	if !opts.StartDate.Equal(opts.EndDate) {
		t.Errorf("EndDate should equal StartDate: %v, %v", opts.StartDate, opts.EndDate)
	}
}

func TestParseBatchOptions_Errors(t *testing.T) {
	tests := []struct {
		name    string
		since   string
		until   string
		org     string
		format  string
		wantErr error
	}{
		{"no since", "", "", "org", "json", apperrors.ErrSinceRequired},
		{"no org", "2026-05-01", "", "", "json", apperrors.ErrOrgRequired},
		{"unsupported format", "2026-05-01", "", "org", "markdown", apperrors.ErrUnsupportedFormat},
		{"invalid since", "2026/05/01", "", "org", "json", apperrors.ErrInvalidDate},
		{"invalid until", "2026-05-01", "tomorrow", "org", "json", apperrors.ErrInvalidDate},
		{"since after until", "2026-05-02", "2026-05-01", "org", "json", apperrors.ErrInvalidDateRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseBatchOptions(tt.since, tt.until, tt.org, tt.format, "")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err: got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
