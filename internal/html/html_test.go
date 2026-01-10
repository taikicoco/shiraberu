package html

import (
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
)

func TestCalcSummary(t *testing.T) {
	report := &pr.Report{
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Opened: []github.PullRequest{
					{Additions: 100, Deletions: 50},
					{Additions: 200, Deletions: 30},
				},
				Draft: []github.PullRequest{
					{Additions: 10, Deletions: 5},
				},
				Merged: []github.PullRequest{
					{Additions: 500, Deletions: 100},
				},
				Reviewed: []github.PullRequest{
					{Additions: 0, Deletions: 0},
					{Additions: 0, Deletions: 0},
				},
			},
			{
				Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				Opened: []github.PullRequest{
					{Additions: 50, Deletions: 20},
				},
				Merged: []github.PullRequest{
					{Additions: 300, Deletions: 150},
				},
			},
		},
	}

	summary := calcSummary(report)

	if summary.OpenedCount != 3 {
		t.Errorf("OpenedCount: got %d, want 3", summary.OpenedCount)
	}
	if summary.DraftCount != 1 {
		t.Errorf("DraftCount: got %d, want 1", summary.DraftCount)
	}
	if summary.MergedCount != 2 {
		t.Errorf("MergedCount: got %d, want 2", summary.MergedCount)
	}
	if summary.ReviewedCount != 2 {
		t.Errorf("ReviewedCount: got %d, want 2", summary.ReviewedCount)
	}
	// Additions: 100+200+10+500+50+300 = 1160
	if summary.Additions != 1160 {
		t.Errorf("Additions: got %d, want 1160", summary.Additions)
	}
	// Deletions: 50+30+5+100+20+150 = 355
	if summary.Deletions != 355 {
		t.Errorf("Deletions: got %d, want 355", summary.Deletions)
	}
}

func TestCalcDailyStats(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 2, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{Additions: 100, Deletions: 50},
				},
				Merged: []github.PullRequest{
					{Additions: 200, Deletions: 30},
				},
			},
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{Additions: 50, Deletions: 20},
					{Additions: 30, Deletions: 10},
				},
			},
		},
	}

	stats := calcDailyStats(report)

	if len(stats) != 2 {
		t.Fatalf("len(stats): got %d, want 2", len(stats))
	}

	// Should be sorted by date ascending
	if stats[0].Date != "2025-01-01" {
		t.Errorf("stats[0].Date: got %s, want 2025-01-01", stats[0].Date)
	}
	if stats[1].Date != "2025-01-02" {
		t.Errorf("stats[1].Date: got %s, want 2025-01-02", stats[1].Date)
	}

	// Check first day stats
	if stats[0].OpenedCount != 2 {
		t.Errorf("stats[0].OpenedCount: got %d, want 2", stats[0].OpenedCount)
	}
	if stats[0].Additions != 80 {
		t.Errorf("stats[0].Additions: got %d, want 80", stats[0].Additions)
	}

	// Check second day stats
	if stats[1].OpenedCount != 1 {
		t.Errorf("stats[1].OpenedCount: got %d, want 1", stats[1].OpenedCount)
	}
	if stats[1].MergedCount != 1 {
		t.Errorf("stats[1].MergedCount: got %d, want 1", stats[1].MergedCount)
	}
	if stats[1].TotalPRs != 2 {
		t.Errorf("stats[1].TotalPRs: got %d, want 2", stats[1].TotalPRs)
	}
}

func TestCalcWeeklyStats(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 6, 0, 0, 0, 0, jst), // Week 2
				Opened: []github.PullRequest{
					{}, {},
				},
			},
			{
				Date: time.Date(2025, 1, 7, 0, 0, 0, 0, jst), // Week 2
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, jst), // Week 1
				Merged: []github.PullRequest{
					{}, {}, {},
				},
			},
		},
	}

	stats := calcWeeklyStats(report)

	if len(stats) != 2 {
		t.Fatalf("len(stats): got %d, want 2", len(stats))
	}

	// Check aggregation
	totalOpened := 0
	totalMerged := 0
	for _, s := range stats {
		totalOpened += s.OpenedCount
		totalMerged += s.MergedCount
	}

	if totalOpened != 3 {
		t.Errorf("totalOpened: got %d, want 3", totalOpened)
	}
	if totalMerged != 3 {
		t.Errorf("totalMerged: got %d, want 3", totalMerged)
	}
}

func TestCalcMonthlyStats(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	report := &pr.Report{
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 15, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{}, {},
				},
			},
			{
				Date: time.Date(2025, 1, 20, 0, 0, 0, 0, jst),
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2024, 12, 25, 0, 0, 0, 0, jst),
				Merged: []github.PullRequest{
					{}, {}, {},
				},
			},
		},
	}

	stats := calcMonthlyStats(report)

	if len(stats) != 2 {
		t.Fatalf("len(stats): got %d, want 2", len(stats))
	}

	// Should be sorted by month ascending
	if stats[0].Month != "12月" {
		t.Errorf("stats[0].Month: got %s, want 12月", stats[0].Month)
	}
	if stats[1].Month != "1月" {
		t.Errorf("stats[1].Month: got %s, want 1月", stats[1].Month)
	}

	// Check December
	if stats[0].MergedCount != 3 {
		t.Errorf("stats[0].MergedCount: got %d, want 3", stats[0].MergedCount)
	}

	// Check January (aggregated)
	if stats[1].OpenedCount != 3 {
		t.Errorf("stats[1].OpenedCount: got %d, want 3", stats[1].OpenedCount)
	}
}

func TestCalcRepoStats(t *testing.T) {
	report := &pr.Report{
		Days: []pr.DailyPRs{
			{
				Opened: []github.PullRequest{
					{Repository: "repo-a"},
					{Repository: "repo-a"},
					{Repository: "repo-b"},
				},
				Merged: []github.PullRequest{
					{Repository: "repo-a"},
				},
				Reviewed: []github.PullRequest{
					{Repository: "repo-c"},
					{Repository: "repo-b"},
				},
			},
		},
	}

	stats := calcRepoStats(report)

	if len(stats) != 3 {
		t.Fatalf("len(stats): got %d, want 3", len(stats))
	}

	// Should be sorted by count descending
	if stats[0].Repository != "repo-a" || stats[0].Count != 3 {
		t.Errorf("stats[0]: got %s=%d, want repo-a=3", stats[0].Repository, stats[0].Count)
	}
	if stats[1].Repository != "repo-b" || stats[1].Count != 2 {
		t.Errorf("stats[1]: got %s=%d, want repo-b=2", stats[1].Repository, stats[1].Count)
	}
	if stats[2].Repository != "repo-c" || stats[2].Count != 1 {
		t.Errorf("stats[2]: got %s=%d, want repo-c=1", stats[2].Repository, stats[2].Count)
	}
}

func TestCalcSummaryDiff(t *testing.T) {
	current := Summary{
		OpenedCount:   10,
		DraftCount:    3,
		MergedCount:   8,
		ReviewedCount: 5,
	}

	t.Run("with previous report", func(t *testing.T) {
		previousReport := &pr.Report{
			Days: []pr.DailyPRs{
				{
					Opened:   []github.PullRequest{{}, {}, {}, {}, {}, {}, {}},     // 7
					Draft:    []github.PullRequest{{}, {}, {}, {}},                 // 4
					Merged:   []github.PullRequest{{}, {}, {}, {}, {}, {}, {}, {}}, // 8
					Reviewed: []github.PullRequest{{}, {}, {}},                     // 3
				},
			},
		}

		diff := calcSummaryDiff(current, previousReport)

		if !diff.HasPrevious {
			t.Error("HasPrevious: got false, want true")
		}
		if diff.OpenedDiff != 3 { // 10 - 7
			t.Errorf("OpenedDiff: got %d, want 3", diff.OpenedDiff)
		}
		if diff.DraftDiff != -1 { // 3 - 4
			t.Errorf("DraftDiff: got %d, want -1", diff.DraftDiff)
		}
		if diff.MergedDiff != 0 { // 8 - 8
			t.Errorf("MergedDiff: got %d, want 0", diff.MergedDiff)
		}
		if diff.ReviewedDiff != 2 { // 5 - 3
			t.Errorf("ReviewedDiff: got %d, want 2", diff.ReviewedDiff)
		}
	})

	t.Run("without previous report", func(t *testing.T) {
		diff := calcSummaryDiff(current, nil)

		if diff.HasPrevious {
			t.Error("HasPrevious: got true, want false")
		}
	})
}

func TestCalcSummary_Empty(t *testing.T) {
	report := &pr.Report{
		Days: []pr.DailyPRs{},
	}

	summary := calcSummary(report)

	if summary.OpenedCount != 0 {
		t.Errorf("OpenedCount: got %d, want 0", summary.OpenedCount)
	}
	if summary.Additions != 0 {
		t.Errorf("Additions: got %d, want 0", summary.Additions)
	}
}

func TestCalcDailyStats_Empty(t *testing.T) {
	report := &pr.Report{
		Days: []pr.DailyPRs{},
	}

	stats := calcDailyStats(report)

	if len(stats) != 0 {
		t.Errorf("len(stats): got %d, want 0", len(stats))
	}
}
