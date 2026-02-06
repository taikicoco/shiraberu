package prompt

import (
	"errors"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
	apperrors "github.com/taikicoco/shiraberu/internal/errors"
	"github.com/taikicoco/shiraberu/internal/period"
)

// MockIO はテスト用のIO実装
type MockIO struct {
	readLineResponses []string
	readLineIdx       int
	selectResponses   []int
	selectIdx         int
}

func (m *MockIO) ReadLine(label string, defaultVal string) (string, error) {
	if m.readLineIdx >= len(m.readLineResponses) {
		return defaultVal, nil
	}
	resp := m.readLineResponses[m.readLineIdx]
	m.readLineIdx++
	return resp, nil
}

func (m *MockIO) Select(label string, options []string, defaultIdx int) (int, error) {
	if m.selectIdx >= len(m.selectResponses) {
		return defaultIdx, nil
	}
	resp := m.selectResponses[m.selectIdx]
	m.selectIdx++
	return resp, nil
}

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
			want:  "20250115.md",
		},
		{
			name:  "same day html",
			start: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			ext:   ".html",
			want:  "20250115.html",
		},
		{
			name:  "date range markdown",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC),
			ext:   ".md",
			want:  "20250101-20250107.md",
		},
		{
			name:  "date range html",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			ext:   ".html",
			want:  "20250101-20250131.html",
		},
		{
			name:  "across months",
			start: time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			ext:   ".md",
			want:  "20241215-20250115.md",
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
	opts := Options{
		Org:        "test-org",
		StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		PeriodType: period.TypeMonth,
		Format:     "html",
		OutputPath: "/path/to/output.html",
	}

	if opts.Org != "test-org" {
		t.Errorf("Org: got %q, want %q", opts.Org, "test-org")
	}
	if opts.Format != "html" {
		t.Errorf("Format: got %q, want %q", opts.Format, "html")
	}
	if opts.PeriodType != period.TypeMonth {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeMonth)
	}
}

func TestPromptSelectMonth(t *testing.T) {
	today := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	mockIO := &MockIO{
		selectResponses: []int{0}, // Select first month (current)
	}
	r := NewRunner(mockIO)

	start, end := r.promptSelectMonth(today)

	if start.Day() != 1 {
		t.Errorf("Start day should be 1, got %d", start.Day())
	}

	expectedEnd := start.AddDate(0, 1, -1)
	if !end.Equal(expectedEnd) {
		t.Errorf("End date: got %v, want %v", end, expectedEnd)
	}
}

func TestRunner_Run_SingleDayBrowser(t *testing.T) {
	// Flow: org → username → period mode (single day) → select today → confirm → format (browser)
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",  // Organization
			"testuser", // Username
			"",        // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			0, // Output format: HTML (open in browser)
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Org != "my-org" {
		t.Errorf("Org: got %q, want %q", opts.Org, "my-org")
	}
	if opts.Username != "testuser" {
		t.Errorf("Username: got %q, want %q", opts.Username, "testuser")
	}
	if opts.Format != "browser" {
		t.Errorf("Format: got %q, want %q", opts.Format, "browser")
	}
	if opts.PeriodType != period.TypeCustom {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeCustom)
	}
}

func TestRunner_Run_DateRangeMarkdown(t *testing.T) {
	// Flow: org → username → period mode (date range) → this week → confirm → format (markdown)
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			0, // Select range: This week
			2, // Output format: Markdown
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "markdown"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Org != "my-org" {
		t.Errorf("Org: got %q, want %q", opts.Org, "my-org")
	}
	if opts.Format != "markdown" {
		t.Errorf("Format: got %q, want %q", opts.Format, "markdown")
	}
	if opts.PeriodType != period.TypeWeek {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeWeek)
	}
}

func TestRunner_Run_DefaultUsername(t *testing.T) {
	// When username is empty, should use defaultUsername
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org", // Organization
			"",       // Username (empty → use default)
			"",       // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Username != "default-user" {
		t.Errorf("Username: got %q, want %q", opts.Username, "default-user")
	}
}

func TestRunner_Run_EmptyOrgError(t *testing.T) {
	// When org is empty and no default, should return error
	mockIO := &MockIO{
		readLineResponses: []string{
			"", // Organization (empty)
		},
	}

	cfg := &config.Config{Org: "", Format: "browser"}
	r := NewRunner(mockIO)

	_, err := r.Run(cfg, "default-user")
	if !errors.Is(err, apperrors.ErrOrgRequired) {
		t.Errorf("expected ErrOrgRequired, got %v", err)
	}
}

func TestRunner_Run_OutputPathHTML(t *testing.T) {
	// When format is html and output_dir is set, should generate output path
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			1, // Output format: HTML
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "html", OutputDir: "/tmp/reports"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "html" {
		t.Errorf("Format: got %q, want %q", opts.Format, "html")
	}
	if opts.OutputPath == "" {
		t.Error("OutputPath should not be empty when OutputDir is set")
	}
}

func TestRunner_Run_OutputPathMarkdown(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			2, // Output format: Markdown
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "markdown", OutputDir: "/tmp/reports"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "markdown" {
		t.Errorf("Format: got %q, want %q", opts.Format, "markdown")
	}
	if opts.OutputPath == "" {
		t.Error("OutputPath should not be empty when OutputDir is set")
	}
}

func TestRunner_Run_BackFromPeriodToUsername(t *testing.T) {
	// Flow: org → username → period mode (back) → username → period mode → single day → format
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username (first time)
			"testuser", // Username (after back)
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			2, // Period type: Back
			0, // Period type: Single day (after returning from back)
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "browser" {
		t.Errorf("Format: got %q, want %q", opts.Format, "browser")
	}
}

func TestRunner_Run_BackFromFormatToPeriod(t *testing.T) {
	// Flow: org → username → period mode → single day → format (back) → period mode → single day → format
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (first, Enter = OK)
			"",         // confirmDateRange (second, Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			3, // Output format: Back
			0, // Period type: Single day (after back)
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "browser" {
		t.Errorf("Format: got %q, want %q", opts.Format, "browser")
	}
}

func TestRunner_Run_DateRangeLastWeek(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			1, // Select range: Last week
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PeriodType != period.TypeWeek {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeWeek)
	}
}

func TestRunner_Run_DateRangeLastMonth(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			3, // Select range: Last month
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PeriodType != period.TypeMonth {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeMonth)
	}
}

func TestRunner_Run_DateRangeSelectMonth(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			4, // Select range: Select month
			2, // Select month: 3rd option (2 months ago)
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PeriodType != period.TypeMonth {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeMonth)
	}
	if opts.StartDate.Day() != 1 {
		t.Errorf("Start day should be 1, got %d", opts.StartDate.Day())
	}
}

func TestRunner_Run_DateRangeEnterDates(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",      // Organization
			"testuser",    // Username
			"2025-01-01",  // Start date
			"2025-01-31",  // End date
			"",            // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			5, // Select range: Enter dates
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PeriodType != period.TypeCustom {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeCustom)
	}
	expectedStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !opts.StartDate.Equal(expectedStart) {
		t.Errorf("StartDate: got %v, want %v", opts.StartDate, expectedStart)
	}
	expectedEnd := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	if !opts.EndDate.Equal(expectedEnd) {
		t.Errorf("EndDate: got %v, want %v", opts.EndDate, expectedEnd)
	}
}

func TestRunner_Run_SingleDayYesterday(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			1, // Select date: Yesterday
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()
	yesterday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)
	if !opts.StartDate.Equal(yesterday) {
		t.Errorf("StartDate: got %v, want %v", opts.StartDate, yesterday)
	}
}

func TestRunner_Run_SingleDayEnterDate(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",     // Organization
			"testuser",   // Username
			"2025-03-15", // Enter date
			"",           // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			2, // Select date: Enter date
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	if !opts.StartDate.Equal(expected) {
		t.Errorf("StartDate: got %v, want %v", opts.StartDate, expected)
	}
}

func TestRunner_Run_SingleDayBack(t *testing.T) {
	// Flow: period mode → single day → back → period mode → single day → today → format
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			3, // Select date: Back
			0, // Period type: Single day (after back)
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "browser" {
		t.Errorf("Format: got %q, want %q", opts.Format, "browser")
	}
}

func TestRunner_Run_DateRangeBack(t *testing.T) {
	// Flow: period mode → date range → back → period mode → single day → today → format
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			6, // Select range: Back
			0, // Period type: Single day (after back)
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "browser" {
		t.Errorf("Format: got %q, want %q", opts.Format, "browser")
	}
}

func TestRunner_ConfirmDateRange_ChangeStart(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"s",          // Change start
			"2025-02-01", // New start date
			"",           // Confirm OK
		},
	}
	r := NewRunner(mockIO)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := r.confirmDateRange(start, end)

	expectedStart := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	if !newStart.Equal(expectedStart) {
		t.Errorf("Start: got %v, want %v", newStart, expectedStart)
	}
	if !newEnd.Equal(end) {
		t.Errorf("End should not change: got %v, want %v", newEnd, end)
	}
}

func TestRunner_ConfirmDateRange_ChangeEnd(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"e",          // Change end
			"2025-02-28", // New end date
			"",           // Confirm OK
		},
	}
	r := NewRunner(mockIO)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := r.confirmDateRange(start, end)

	if !newStart.Equal(start) {
		t.Errorf("Start should not change: got %v, want %v", newStart, start)
	}
	expectedEnd := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
	if !newEnd.Equal(expectedEnd) {
		t.Errorf("End: got %v, want %v", newEnd, expectedEnd)
	}
}

func TestRunner_ConfirmDateRange_DirectConfirm(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"", // Confirm OK
		},
	}
	r := NewRunner(mockIO)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	newStart, newEnd := r.confirmDateRange(start, end)

	if !newStart.Equal(start) {
		t.Errorf("Start should not change: got %v, want %v", newStart, start)
	}
	if !newEnd.Equal(end) {
		t.Errorf("End should not change: got %v, want %v", newEnd, end)
	}
}

func TestRunner_PromptDate_InvalidFormat(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"not-a-date", // Invalid date
		},
	}
	r := NewRunner(mockIO)

	defaultDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	result := r.promptDate("Date", defaultDate)

	if !result.Equal(defaultDate) {
		t.Errorf("Should return default date on invalid input: got %v, want %v", result, defaultDate)
	}
}

func TestRunner_PromptDate_ValidFormat(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"2025-06-15", // Valid date
		},
	}
	r := NewRunner(mockIO)

	defaultDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	result := r.promptDate("Date", defaultDate)

	expected := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestRunner_PromptText_Error(t *testing.T) {
	mockIO := &ErrorMockIO{}
	r := NewRunner(mockIO)

	result := r.promptText("Label", "default")
	if result != "default" {
		t.Errorf("Should return default on error: got %q, want %q", result, "default")
	}
}

func TestRunner_PromptSelect_Error(t *testing.T) {
	mockIO := &ErrorMockIO{}
	r := NewRunner(mockIO)

	result := r.promptSelect("Label", []string{"a", "b"}, 1)
	if result != 1 {
		t.Errorf("Should return default on error: got %d, want %d", result, 1)
	}
}

// ErrorMockIO はエラーを返すテスト用IO
type ErrorMockIO struct{}

func (m *ErrorMockIO) ReadLine(label string, defaultVal string) (string, error) {
	return "", errors.New("mock error")
}

func (m *ErrorMockIO) Select(label string, options []string, defaultIdx int) (int, error) {
	return 0, errors.New("mock error")
}

func TestRunner_Run_DefaultFormatIdx(t *testing.T) {
	// Test that default format index is correctly set from config
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			1, // Output format: HTML (index 1, which maps to "html")
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "html", OutputDir: ""}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Format != "html" {
		t.Errorf("Format: got %q, want %q", opts.Format, "html")
	}
}

func TestRunner_Run_BrowserNoOutputPath(t *testing.T) {
	// Browser format should not generate output path even if OutputDir is set
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			0, // Period type: Single day
			0, // Select date: Today
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser", OutputDir: "/tmp/reports"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.OutputPath != "" {
		t.Errorf("OutputPath should be empty for browser format: got %q", opts.OutputPath)
	}
}

func TestRunner_Run_DateRangeThisMonth(t *testing.T) {
	mockIO := &MockIO{
		readLineResponses: []string{
			"my-org",   // Organization
			"testuser", // Username
			"",         // confirmDateRange (Enter = OK)
		},
		selectResponses: []int{
			1, // Period type: Date range
			2, // Select range: This month
			0, // Output format: browser
		},
	}

	cfg := &config.Config{Org: "default-org", Format: "browser"}
	r := NewRunner(mockIO)

	opts, err := r.Run(cfg, "default-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.PeriodType != period.TypeMonth {
		t.Errorf("PeriodType: got %q, want %q", opts.PeriodType, period.TypeMonth)
	}
	if opts.StartDate.Day() != 1 {
		t.Errorf("StartDate should be first of month, got day %d", opts.StartDate.Day())
	}
}
