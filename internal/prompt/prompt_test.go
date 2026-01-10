package prompt

import (
	"testing"
	"time"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		ext   string
		want  string
	}{
		{
			name:  "same day markdown",
			start: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			ext:   ".md",
			want:  "2025-01-15.md",
		},
		{
			name:  "same day html",
			start: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			ext:   ".html",
			want:  "2025-01-15.html",
		},
		{
			name:  "date range markdown",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC),
			ext:   ".md",
			want:  "2025-01-01_2025-01-07.md",
		},
		{
			name:  "date range html",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			ext:   ".html",
			want:  "2025-01-01_2025-01-31.html",
		},
		{
			name:  "across months",
			start: time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			ext:   ".md",
			want:  "2024-12-15_2025-01-15.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFilename(tt.start, tt.end, tt.ext)
			if got != tt.want {
				t.Errorf("generateFilename(%v, %v, %q): got %q, want %q",
					tt.start.Format("2006-01-02"), tt.end.Format("2006-01-02"), tt.ext, got, tt.want)
			}
		})
	}
}

func TestOptions_Struct(t *testing.T) {
	// Test that Options struct can be created with expected fields
	opts := Options{
		Org:        "test-org",
		StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		PeriodType: PeriodTypeMonth,
		Format:     "html",
		OutputPath: "/path/to/output.html",
	}

	if opts.Org != "test-org" {
		t.Errorf("Org: got %q, want %q", opts.Org, "test-org")
	}
	if opts.Format != "html" {
		t.Errorf("Format: got %q, want %q", opts.Format, "html")
	}
	if opts.PeriodType != PeriodTypeMonth {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, PeriodTypeMonth)
	}
}

func TestPeriodTypeConstants(t *testing.T) {
	if PeriodTypeWeek != "week" {
		t.Errorf("PeriodTypeWeek: got %q, want %q", PeriodTypeWeek, "week")
	}
	if PeriodTypeMonth != "month" {
		t.Errorf("PeriodTypeMonth: got %q, want %q", PeriodTypeMonth, "month")
	}
	if PeriodTypeCustom != "custom" {
		t.Errorf("PeriodTypeCustom: got %q, want %q", PeriodTypeCustom, "custom")
	}
}

func TestPromptSelectMonth(t *testing.T) {
	today := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	// Test that promptSelectMonth returns valid date range for a month
	// Note: We can't fully test the interactive selection, but we can test the month calculation
	start, end := promptSelectMonth(today)

	// Start should be the first day of a month
	if start.Day() != 1 {
		t.Errorf("Start day should be 1, got %d", start.Day())
	}

	// End should be the last day of the same month
	expectedEnd := start.AddDate(0, 1, -1)
	if !end.Equal(expectedEnd) {
		t.Errorf("End date: got %v, want %v", end, expectedEnd)
	}
}
