package demo

import (
	"math/rand"
	"testing"
	"time"
)

func TestGenerateReport(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	report, previousReport := GenerateReport(startDate, endDate)

	// Check current report
	if report == nil {
		t.Fatal("report is nil")
	}
	if report.Org != "demo-org" {
		t.Errorf("Org: got %q, want %q", report.Org, "demo-org")
	}
	if !report.StartDate.Equal(startDate) {
		t.Errorf("StartDate: got %v, want %v", report.StartDate, startDate)
	}
	if !report.EndDate.Equal(endDate) {
		t.Errorf("EndDate: got %v, want %v", report.EndDate, endDate)
	}

	// Check previous report exists
	if previousReport == nil {
		t.Fatal("previousReport is nil")
	}
	if previousReport.Org != "demo-org" {
		t.Errorf("previousReport.Org: got %q, want %q", previousReport.Org, "demo-org")
	}
}

func TestGeneratePR(t *testing.T) {
	r := rand.New(rand.NewSource(42)) // Fixed seed for reproducibility
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		state   string
		isDraft bool
	}{
		{"open", false},
		{"merged", false},
		{"draft", true},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			pr := generatePR(r, date, tt.state)

			if pr.Title == "" {
				t.Error("Title should not be empty")
			}
			if pr.URL == "" {
				t.Error("URL should not be empty")
			}
			if pr.Repository == "" {
				t.Error("Repository should not be empty")
			}
			if pr.State != tt.state {
				t.Errorf("State: got %q, want %q", pr.State, tt.state)
			}
			if pr.IsDraft != tt.isDraft {
				t.Errorf("IsDraft: got %v, want %v", pr.IsDraft, tt.isDraft)
			}
			if pr.Additions < minAdditions || pr.Additions >= maxAdditions+minAdditions {
				t.Errorf("Additions out of range: %d", pr.Additions)
			}
			if pr.Deletions < minDeletions || pr.Deletions >= maxDeletions+minDeletions {
				t.Errorf("Deletions out of range: %d", pr.Deletions)
			}
			if pr.ChangedFiles < 1 || pr.ChangedFiles > maxChangedFiles {
				t.Errorf("ChangedFiles out of range: %d", pr.ChangedFiles)
			}
		})
	}
}

func TestRandomPRNumber(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	for i := 0; i < 100; i++ {
		num := randomPRNumber(r)
		if len(num) != 4 {
			t.Errorf("PR number should be 4 digits, got %q", num)
		}
	}
}

func TestGenerateReport_DaysSortedDescending(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC)

	report, _ := GenerateReport(startDate, endDate)

	if len(report.Days) < 2 {
		t.Skip("Not enough days generated to test sorting")
	}

	// Check days are sorted descending
	for i := 0; i < len(report.Days)-1; i++ {
		if report.Days[i].Date.Before(report.Days[i+1].Date) {
			t.Errorf("Days not sorted descending: %v before %v",
				report.Days[i].Date, report.Days[i+1].Date)
		}
	}
}

func TestGeneratePR_URLFormat(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	pr := generatePR(r, date, "open")

	// URL should start with expected prefix
	expectedPrefix := "https://github.com/demo-org/"
	if len(pr.URL) < len(expectedPrefix) {
		t.Errorf("URL too short: %s", pr.URL)
	}
	if pr.URL[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("URL should start with %q, got %q", expectedPrefix, pr.URL)
	}
}
