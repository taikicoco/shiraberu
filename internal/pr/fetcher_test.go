package pr

import (
	"strings"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

func TestGroupByDate(t *testing.T) {
	// Create test PRs with different dates
	opened := []github.PullRequest{
		{
			Title:     "PR 1",
			CreatedAt: time.Date(2025, 1, 10, 10, 0, 0, 0, timezone.JST),
		},
		{
			Title:     "PR 2",
			CreatedAt: time.Date(2025, 1, 10, 15, 0, 0, 0, timezone.JST),
		},
		{
			Title:     "PR 3 (draft)",
			CreatedAt: time.Date(2025, 1, 11, 9, 0, 0, 0, timezone.JST),
			IsDraft:   true,
		},
	}

	mergedAt := time.Date(2025, 1, 10, 18, 0, 0, 0, timezone.JST)
	merged := []github.PullRequest{
		{
			Title:    "Merged PR",
			MergedAt: &mergedAt,
		},
	}

	reviewed := []github.PullRequest{
		{
			Title:     "Reviewed PR",
			UpdatedAt: time.Date(2025, 1, 11, 14, 0, 0, 0, timezone.JST),
		},
	}

	days := groupByDate(opened, merged, reviewed)

	// Should have 2 days
	if len(days) != 2 {
		t.Fatalf("len(days): got %d, want 2", len(days))
	}

	// Days should be sorted descending (newest first)
	if days[0].Date.Before(days[1].Date) {
		t.Error("days should be sorted descending")
	}

	// Check 2025-01-11 (first in list, newer)
	day11 := days[0]
	if day11.Date.Day() != 11 {
		t.Errorf("first day should be 11, got %d", day11.Date.Day())
	}
	if len(day11.Draft) != 1 {
		t.Errorf("day11 Draft count: got %d, want 1", len(day11.Draft))
	}
	if len(day11.Reviewed) != 1 {
		t.Errorf("day11 Reviewed count: got %d, want 1", len(day11.Reviewed))
	}

	// Check 2025-01-10 (second in list, older)
	day10 := days[1]
	if day10.Date.Day() != 10 {
		t.Errorf("second day should be 10, got %d", day10.Date.Day())
	}
	if len(day10.Opened) != 2 {
		t.Errorf("day10 Opened count: got %d, want 2", len(day10.Opened))
	}
	if len(day10.Merged) != 1 {
		t.Errorf("day10 Merged count: got %d, want 1", len(day10.Merged))
	}
}

func TestGroupByDate_Empty(t *testing.T) {
	days := groupByDate(nil, nil, nil)
	if len(days) != 0 {
		t.Errorf("len(days): got %d, want 0", len(days))
	}
}

func TestGroupByDate_UTCToJST(t *testing.T) {
	// UTC 2025-01-10 20:00 = JST 2025-01-11 05:00
	utc := time.UTC
	opened := []github.PullRequest{
		{
			Title:     "UTC PR",
			CreatedAt: time.Date(2025, 1, 10, 20, 0, 0, 0, utc),
		},
	}

	days := groupByDate(opened, nil, nil)

	if len(days) != 1 {
		t.Fatalf("len(days): got %d, want 1", len(days))
	}

	// Should be grouped as JST 2025-01-11
	if days[0].Date.Day() != 11 {
		t.Errorf("day should be 11 (JST), got %d", days[0].Date.Day())
	}
}

// MockPRSearcher is a mock implementation of github.PRSearcher
type MockPRSearcher struct {
	username  string
	openedPRs []github.PullRequest
	mergedPRs []github.PullRequest
	reviewPRs []github.PullRequest
	err       error
}

func (m *MockPRSearcher) Username() string {
	return m.username
}

func (m *MockPRSearcher) SearchPRs(org string, query string, dateFilter string) ([]github.PullRequest, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Return different PRs based on query
	if strings.Contains(query, "is:open") {
		return m.openedPRs, nil
	}
	if strings.Contains(query, "is:merged") {
		return m.mergedPRs, nil
	}
	if strings.Contains(query, "reviewed-by:") {
		return m.reviewPRs, nil
	}

	return nil, nil
}

func TestNewFetcher(t *testing.T) {
	mock := &MockPRSearcher{username: "testuser"}
	fetcher := NewFetcher(mock)
	if fetcher == nil {
		t.Error("NewFetcher returned nil")
	}
}

func TestFetcher_Fetch(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 10, 0, 0, 0, timezone.JST)
	mergedAt := time.Date(2025, 1, 10, 18, 0, 0, 0, timezone.JST)
	updatedAt := time.Date(2025, 1, 11, 14, 0, 0, 0, timezone.JST)

	mock := &MockPRSearcher{
		username: "testuser",
		openedPRs: []github.PullRequest{
			{
				Title:     "Opened PR",
				URL:       "https://github.com/test/repo/pull/1",
				CreatedAt: createdAt,
			},
		},
		mergedPRs: []github.PullRequest{
			{
				Title:    "Merged PR",
				URL:      "https://github.com/test/repo/pull/2",
				MergedAt: &mergedAt,
			},
		},
		reviewPRs: []github.PullRequest{
			{
				Title:     "Reviewed PR",
				URL:       "https://github.com/test/repo/pull/3",
				UpdatedAt: updatedAt,
			},
		},
	}

	fetcher := NewFetcher(mock)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, timezone.JST)

	report, err := fetcher.Fetch("test-org", "testuser", startDate, endDate)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	if report.Org != "test-org" {
		t.Errorf("Org: got %q, want %q", report.Org, "test-org")
	}
	if !report.StartDate.Equal(startDate) {
		t.Errorf("StartDate: got %v, want %v", report.StartDate, startDate)
	}
	if !report.EndDate.Equal(endDate) {
		t.Errorf("EndDate: got %v, want %v", report.EndDate, endDate)
	}
	if len(report.Days) == 0 {
		t.Error("Days should not be empty")
	}

	// Check that PRs were grouped
	totalOpened := 0
	totalMerged := 0
	totalReviewed := 0
	for _, day := range report.Days {
		totalOpened += len(day.Opened)
		totalMerged += len(day.Merged)
		totalReviewed += len(day.Reviewed)
	}

	if totalOpened != 1 {
		t.Errorf("totalOpened: got %d, want 1", totalOpened)
	}
	if totalMerged != 1 {
		t.Errorf("totalMerged: got %d, want 1", totalMerged)
	}
	if totalReviewed != 1 {
		t.Errorf("totalReviewed: got %d, want 1", totalReviewed)
	}
}

func TestFetcher_Fetch_Error(t *testing.T) {
	mock := &MockPRSearcher{
		username: "testuser",
		err:      errMock,
	}

	fetcher := NewFetcher(mock)

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := fetcher.Fetch("test-org", "testuser", startDate, endDate)
	if err == nil {
		t.Error("Fetch() should return error")
	}
}

func TestSplitDateRange_SingleMonth(t *testing.T) {
	start := time.Date(2025, 1, 10, 0, 0, 0, 0, timezone.JST)
	end := time.Date(2025, 1, 25, 23, 59, 59, 0, timezone.JST)

	chunks := splitDateRange(start, end)

	if len(chunks) != 1 {
		t.Fatalf("len(chunks): got %d, want 1", len(chunks))
	}
	if !chunks[0].start.Equal(start) {
		t.Errorf("start: got %v, want %v", chunks[0].start, start)
	}
	if !chunks[0].end.Equal(end) {
		t.Errorf("end: got %v, want %v", chunks[0].end, end)
	}
}

func TestSplitDateRange_MultipleMonths(t *testing.T) {
	start := time.Date(2025, 4, 1, 0, 0, 0, 0, timezone.JST)
	end := time.Date(2025, 6, 15, 23, 59, 59, 0, timezone.JST)

	chunks := splitDateRange(start, end)

	if len(chunks) != 3 {
		t.Fatalf("len(chunks): got %d, want 3", len(chunks))
	}

	// April chunk: 04-01 ~ 04-30
	if chunks[0].start.Month() != 4 || chunks[0].start.Day() != 1 {
		t.Errorf("chunk[0] start: got %v", chunks[0].start)
	}
	if chunks[0].end.Month() != 4 || chunks[0].end.Day() != 30 {
		t.Errorf("chunk[0] end: got %v", chunks[0].end)
	}

	// May chunk: 05-01 ~ 05-31
	if chunks[1].start.Month() != 5 || chunks[1].start.Day() != 1 {
		t.Errorf("chunk[1] start: got %v", chunks[1].start)
	}
	if chunks[1].end.Month() != 5 || chunks[1].end.Day() != 31 {
		t.Errorf("chunk[1] end: got %v", chunks[1].end)
	}

	// June chunk: 06-01 ~ 06-15
	if chunks[2].start.Month() != 6 || chunks[2].start.Day() != 1 {
		t.Errorf("chunk[2] start: got %v", chunks[2].start)
	}
	if chunks[2].end.Month() != 6 || chunks[2].end.Day() != 15 {
		t.Errorf("chunk[2] end: got %v", chunks[2].end)
	}
}

func TestSplitDateRange_YearBoundary(t *testing.T) {
	start := time.Date(2025, 12, 1, 0, 0, 0, 0, timezone.JST)
	end := time.Date(2026, 1, 31, 23, 59, 59, 0, timezone.JST)

	chunks := splitDateRange(start, end)

	if len(chunks) != 2 {
		t.Fatalf("len(chunks): got %d, want 2", len(chunks))
	}
	if chunks[0].end.Month() != 12 || chunks[0].end.Day() != 31 {
		t.Errorf("chunk[0] end: got %v", chunks[0].end)
	}
	if chunks[1].start.Year() != 2026 || chunks[1].start.Month() != 1 {
		t.Errorf("chunk[1] start: got %v", chunks[1].start)
	}
}

func TestDedup(t *testing.T) {
	prs := []github.PullRequest{
		{URL: "https://github.com/test/repo/pull/1", Title: "PR 1"},
		{URL: "https://github.com/test/repo/pull/2", Title: "PR 2"},
		{URL: "https://github.com/test/repo/pull/1", Title: "PR 1 dup"},
	}

	result := dedup(prs)
	if len(result) != 2 {
		t.Fatalf("len(result): got %d, want 2", len(result))
	}
	if result[0].Title != "PR 1" {
		t.Errorf("first PR title: got %q, want %q", result[0].Title, "PR 1")
	}
	if result[1].Title != "PR 2" {
		t.Errorf("second PR title: got %q, want %q", result[1].Title, "PR 2")
	}
}

var errMock = &mockError{}

type mockError struct{}

func (e *mockError) Error() string {
	return "mock error"
}
