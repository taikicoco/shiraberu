package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/timezone"
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
	// Additions: only merged PRs (500+300 = 800)
	if summary.Additions != 800 {
		t.Errorf("Additions: got %d, want 800", summary.Additions)
	}
	// Deletions: only merged PRs (100+150 = 250)
	if summary.Deletions != 250 {
		t.Errorf("Deletions: got %d, want 250", summary.Deletions)
	}
}

func TestCalcDailyStats(t *testing.T) {
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 3, 0, 0, 0, 0, timezone.JST), // 3日間の期間
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{Additions: 100, Deletions: 50},
				},
				Merged: []github.PullRequest{
					{Additions: 200, Deletions: 30},
				},
			},
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{Additions: 50, Deletions: 20},
					{Additions: 30, Deletions: 10},
				},
			},
			// 1/3 はPRなし（0で表示される）
		},
	}

	stats := calcDailyStats(report)

	// 期間内の全日（1/1, 1/2, 1/3）が含まれる
	if len(stats) != 3 {
		t.Fatalf("len(stats): got %d, want 3", len(stats))
	}

	// Should be sorted by date ascending
	if stats[0].Date != "2025-01-01" {
		t.Errorf("stats[0].Date: got %s, want 2025-01-01", stats[0].Date)
	}
	if stats[1].Date != "2025-01-02" {
		t.Errorf("stats[1].Date: got %s, want 2025-01-02", stats[1].Date)
	}
	if stats[2].Date != "2025-01-03" {
		t.Errorf("stats[2].Date: got %s, want 2025-01-03", stats[2].Date)
	}

	// Check first day stats (1/1)
	if stats[0].OpenedCount != 2 {
		t.Errorf("stats[0].OpenedCount: got %d, want 2", stats[0].OpenedCount)
	}
	if stats[0].Additions != 80 { // 50 + 30
		t.Errorf("stats[0].Additions: got %d, want 80", stats[0].Additions)
	}
	if stats[0].Deletions != 30 { // 20 + 10
		t.Errorf("stats[0].Deletions: got %d, want 30", stats[0].Deletions)
	}

	// Check second day stats (1/2)
	if stats[1].OpenedCount != 1 {
		t.Errorf("stats[1].OpenedCount: got %d, want 1", stats[1].OpenedCount)
	}
	if stats[1].MergedCount != 1 {
		t.Errorf("stats[1].MergedCount: got %d, want 1", stats[1].MergedCount)
	}
	if stats[1].TotalPRs != 2 {
		t.Errorf("stats[1].TotalPRs: got %d, want 2", stats[1].TotalPRs)
	}

	// Check third day stats (1/3) - PRなしの日
	if stats[2].OpenedCount != 0 {
		t.Errorf("stats[2].OpenedCount: got %d, want 0", stats[2].OpenedCount)
	}
	if stats[2].TotalPRs != 0 {
		t.Errorf("stats[2].TotalPRs: got %d, want 0", stats[2].TotalPRs)
	}
}

func TestCalcWeeklyStats(t *testing.T) {
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 7, 0, 0, 0, 0, timezone.JST),
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 6, 0, 0, 0, 0, timezone.JST), // Week 2 (Monday)
				Opened: []github.PullRequest{
					{}, {},
				},
			},
			{
				Date: time.Date(2025, 1, 7, 0, 0, 0, 0, timezone.JST), // Week 2 (Tuesday)
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST), // Week 1 (Wednesday)
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

	// Should be sorted by week ascending
	// Week 1: 2024-12-30 (Mon) 〜 2025-01-05 (Sun)
	// Week 2: 2025-01-06 (Mon) 〜 2025-01-12 (Sun)
	if stats[0].Week != "12/30 〜 1/5" {
		t.Errorf("stats[0].Week: got %s, want 12/30 〜 1/5", stats[0].Week)
	}
	if stats[1].Week != "1/6 〜 1/12" {
		t.Errorf("stats[1].Week: got %s, want 1/6 〜 1/12", stats[1].Week)
	}

	// Check Week 1 stats (contains only merged PRs)
	if stats[0].MergedCount != 3 {
		t.Errorf("stats[0].MergedCount: got %d, want 3", stats[0].MergedCount)
	}
	if stats[0].OpenedCount != 0 {
		t.Errorf("stats[0].OpenedCount: got %d, want 0", stats[0].OpenedCount)
	}

	// Check Week 2 stats (aggregated from two days: 1/6 and 1/7)
	if stats[1].OpenedCount != 3 { // 2 + 1
		t.Errorf("stats[1].OpenedCount: got %d, want 3", stats[1].OpenedCount)
	}
	if stats[1].MergedCount != 0 {
		t.Errorf("stats[1].MergedCount: got %d, want 0", stats[1].MergedCount)
	}

	// Check StartDate/EndDate
	if stats[0].StartDate != "2024-12-30" {
		t.Errorf("stats[0].StartDate: got %s, want 2024-12-30", stats[0].StartDate)
	}
	if stats[0].EndDate != "2025-01-05" {
		t.Errorf("stats[0].EndDate: got %s, want 2025-01-05", stats[0].EndDate)
	}
	if stats[1].StartDate != "2025-01-06" {
		t.Errorf("stats[1].StartDate: got %s, want 2025-01-06", stats[1].StartDate)
	}
	if stats[1].EndDate != "2025-01-12" {
		t.Errorf("stats[1].EndDate: got %s, want 2025-01-12", stats[1].EndDate)
	}
}

func TestCalcMonthlyStats(t *testing.T) {
	report := &pr.Report{
		StartDate: time.Date(2024, 12, 25, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, timezone.JST),
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{}, {},
				},
			},
			{
				Date: time.Date(2025, 1, 20, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2024, 12, 25, 0, 0, 0, 0, timezone.JST),
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
	if stats[0].Month != "Dec 2024" {
		t.Errorf("stats[0].Month: got %s, want Dec 2024", stats[0].Month)
	}
	if stats[1].Month != "Jan 2025" {
		t.Errorf("stats[1].Month: got %s, want Jan 2025", stats[1].Month)
	}

	// Check December
	if stats[0].MergedCount != 3 {
		t.Errorf("stats[0].MergedCount: got %d, want 3", stats[0].MergedCount)
	}

	// Check January (aggregated)
	if stats[1].OpenedCount != 3 {
		t.Errorf("stats[1].OpenedCount: got %d, want 3", stats[1].OpenedCount)
	}

	// Check StartDate/EndDate
	if stats[0].StartDate != "2024-12-01" {
		t.Errorf("stats[0].StartDate: got %s, want 2024-12-01", stats[0].StartDate)
	}
	if stats[0].EndDate != "2024-12-31" {
		t.Errorf("stats[0].EndDate: got %s, want 2024-12-31", stats[0].EndDate)
	}
	if stats[1].StartDate != "2025-01-01" {
		t.Errorf("stats[1].StartDate: got %s, want 2025-01-01", stats[1].StartDate)
	}
	if stats[1].EndDate != "2025-01-31" {
		t.Errorf("stats[1].EndDate: got %s, want 2025-01-31", stats[1].EndDate)
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
	// StartDate > EndDate の場合は0件
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	stats := calcDailyStats(report)

	if len(stats) != 0 {
		t.Errorf("len(stats): got %d, want 0", len(stats))
	}
}

func TestCalcDailyStats_SingleDay(t *testing.T) {
	// 同じ日の場合は1件（PRなしでも0として表示）
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	stats := calcDailyStats(report)

	if len(stats) != 1 {
		t.Fatalf("len(stats): got %d, want 1", len(stats))
	}
	if stats[0].Date != "2025-01-01" {
		t.Errorf("stats[0].Date: got %s, want 2025-01-01", stats[0].Date)
	}
	if stats[0].TotalPRs != 0 {
		t.Errorf("stats[0].TotalPRs: got %d, want 0", stats[0].TotalPRs)
	}
}

func TestFormatPeriod(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  string
	}{
		{
			name:  "same day",
			start: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  "2025/01/15",
		},
		{
			name:  "different days",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			want:  "2025/01/01 〜 2025/01/31",
		},
		{
			name:  "across months",
			start: time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			want:  "2024/12/15 〜 2025/01/15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPeriod(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("formatPeriod(%v, %v): got %q, want %q", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"open", "Open"},
		{"merged", "Merged"},
		{"closed", "Closed"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q): got %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderHTML(t *testing.T) {
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, timezone.JST),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
		Org:         "test-org",
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 10, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{
						Title:      "Add feature",
						URL:        "https://github.com/test/repo/pull/1",
						Repository: "test/repo",
						Additions:  100,
						Deletions:  50,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderHTML(&buf, report, nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	html := buf.String()

	// Check that essential elements are present
	checks := []string{
		"test-org",
		"2025-01-01",
		"Add feature",
		"test/repo",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("HTML output should contain %q", check)
		}
	}
}

func TestRenderHTML_Empty(t *testing.T) {
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, timezone.JST),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
		Org:         "test-org",
		Days:        []pr.DailyPRs{},
	}

	var buf bytes.Buffer
	err := RenderHTML(&buf, report, nil)
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	// Should still render without error for empty report
	if buf.Len() == 0 {
		t.Error("HTML output should not be empty")
	}
}

func TestRenderMarkdown(t *testing.T) {
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, timezone.JST),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
		Org:         "test-org",
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 10, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{
						Title:      "Add feature",
						URL:        "https://github.com/test/repo/pull/1",
						Repository: "test/repo",
						State:      "open",
					},
				},
				Merged: []github.PullRequest{
					{
						Title:      "Fix bug",
						URL:        "https://github.com/test/repo/pull/2",
						Repository: "test/repo",
						State:      "merged",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderMarkdown(&buf, report)
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}

	markdown := buf.String()

	checks := []struct {
		text string
		desc string
	}{
		{"# PR Log", "header"},
		{"test-org", "org name"},
		{"2025-01-10", "date"},
		{"Add feature", "opened PR title"},
		{"Fix bug", "merged PR title"},
		{"### Opened", "opened section"},
		{"### Merged", "merged section"},
	}

	for _, c := range checks {
		if !strings.Contains(markdown, c.text) {
			t.Errorf("Markdown output should contain %s (%q)", c.desc, c.text)
		}
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	report := &pr.Report{
		GeneratedAt: time.Date(2025, 1, 15, 10, 30, 0, 0, timezone.JST),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:     time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
		Org:         "test-org",
		Days:        []pr.DailyPRs{},
	}

	var buf bytes.Buffer
	err := RenderMarkdown(&buf, report)
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}

	markdown := buf.String()

	if !strings.Contains(markdown, "No pull requests found") {
		t.Error("Empty report should show no PR message")
	}
}

func TestCalcWeeklyStats_Empty(t *testing.T) {
	// StartDate > EndDate の場合は0件
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	stats := calcWeeklyStats(report)

	if len(stats) != 0 {
		t.Errorf("len(stats): got %d, want 0", len(stats))
	}
}

func TestCalcWeeklyStats_WithGap(t *testing.T) {
	// PRがない週も含まれることを確認
	// Week 1: 2024-12-30 〜 2025-01-05
	// Week 2: 2025-01-06 〜 2025-01-12 (PRなし)
	// Week 3: 2025-01-13 〜 2025-01-19
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST),
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST), // Week 1
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2025, 1, 15, 0, 0, 0, 0, timezone.JST), // Week 3
				Merged: []github.PullRequest{
					{},
				},
			},
			// Week 2 (1/6-1/12) はPRなし
		},
	}

	stats := calcWeeklyStats(report)

	// 3週間すべて含まれる
	if len(stats) != 3 {
		t.Fatalf("len(stats): got %d, want 3", len(stats))
	}

	// Week 2 (PRなし週) のカウントは0
	if stats[1].OpenedCount != 0 {
		t.Errorf("stats[1].OpenedCount: got %d, want 0", stats[1].OpenedCount)
	}
	if stats[1].MergedCount != 0 {
		t.Errorf("stats[1].MergedCount: got %d, want 0", stats[1].MergedCount)
	}
}

func TestCalcMonthlyStats_Empty(t *testing.T) {
	// StartDate > EndDate の場合は0件
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	stats := calcMonthlyStats(report)

	if len(stats) != 0 {
		t.Errorf("len(stats): got %d, want 0", len(stats))
	}
}

func TestCalcMonthlyStats_WithGap(t *testing.T) {
	// PRがない月も含まれることを確認
	// Dec 2024, Jan 2025 (PRなし), Feb 2025
	report := &pr.Report{
		StartDate: time.Date(2024, 12, 15, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 2, 15, 0, 0, 0, 0, timezone.JST),
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2024, 12, 20, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{},
				},
			},
			{
				Date: time.Date(2025, 2, 10, 0, 0, 0, 0, timezone.JST),
				Merged: []github.PullRequest{
					{},
				},
			},
			// Jan 2025 はPRなし
		},
	}

	stats := calcMonthlyStats(report)

	// 3ヶ月すべて含まれる
	if len(stats) != 3 {
		t.Fatalf("len(stats): got %d, want 3", len(stats))
	}

	// 月順にソートされている
	if stats[0].Month != "Dec 2024" {
		t.Errorf("stats[0].Month: got %s, want Dec 2024", stats[0].Month)
	}
	if stats[1].Month != "Jan 2025" {
		t.Errorf("stats[1].Month: got %s, want Jan 2025", stats[1].Month)
	}
	if stats[2].Month != "Feb 2025" {
		t.Errorf("stats[2].Month: got %s, want Feb 2025", stats[2].Month)
	}

	// Jan 2025 (PRなし月) のカウントは0
	if stats[1].OpenedCount != 0 {
		t.Errorf("stats[1].OpenedCount: got %d, want 0", stats[1].OpenedCount)
	}
	if stats[1].MergedCount != 0 {
		t.Errorf("stats[1].MergedCount: got %d, want 0", stats[1].MergedCount)
	}
}

func TestCalcRepoStats_Empty(t *testing.T) {
	report := &pr.Report{
		Days: []pr.DailyPRs{},
	}

	stats := calcRepoStats(report)

	if len(stats) != 0 {
		t.Errorf("len(stats): got %d, want 0", len(stats))
	}
}

func TestConvertToDaysJSON(t *testing.T) {
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 3, 0, 0, 0, 0, timezone.JST), // 3日間の期間
		Days: []pr.DailyPRs{
			{
				Date: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
				Opened: []github.PullRequest{
					{Title: "PR1", Repository: "test/repo"},
				},
			},
			// 1/1 と 1/3 はPRなし
		},
	}

	days := convertToDaysJSON(report)

	// 期間内の全日（1/1, 1/2, 1/3）が含まれる
	if len(days) != 3 {
		t.Fatalf("len(days): got %d, want 3", len(days))
	}

	// Should be sorted by date ascending
	if days[0].Date != "2025-01-01" {
		t.Errorf("days[0].Date: got %s, want 2025-01-01", days[0].Date)
	}
	if days[1].Date != "2025-01-02" {
		t.Errorf("days[1].Date: got %s, want 2025-01-02", days[1].Date)
	}
	if days[2].Date != "2025-01-03" {
		t.Errorf("days[2].Date: got %s, want 2025-01-03", days[2].Date)
	}

	// Check day without PRs (1/1) - should have empty slices
	if len(days[0].Opened) != 0 {
		t.Errorf("days[0].Opened: got %d, want 0", len(days[0].Opened))
	}
	if days[0].Opened == nil {
		t.Error("days[0].Opened should be empty slice, not nil")
	}

	// Check day with PRs (1/2)
	if len(days[1].Opened) != 1 {
		t.Errorf("days[1].Opened: got %d, want 1", len(days[1].Opened))
	}
	if days[1].Opened[0].Title != "PR1" {
		t.Errorf("days[1].Opened[0].Title: got %s, want PR1", days[1].Opened[0].Title)
	}

	// Check day without PRs (1/3) - should have empty slices
	if len(days[2].Opened) != 0 {
		t.Errorf("days[2].Opened: got %d, want 0", len(days[2].Opened))
	}
}

func TestConvertToDaysJSON_Empty(t *testing.T) {
	// StartDate > EndDate の場合は0件
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 2, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	days := convertToDaysJSON(report)

	if len(days) != 0 {
		t.Errorf("len(days): got %d, want 0", len(days))
	}
}

func TestConvertToDaysJSON_SingleDay(t *testing.T) {
	// 同じ日の場合は1件
	report := &pr.Report{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, timezone.JST),
		Days:      []pr.DailyPRs{},
	}

	days := convertToDaysJSON(report)

	if len(days) != 1 {
		t.Fatalf("len(days): got %d, want 1", len(days))
	}
	if days[0].Date != "2025-01-01" {
		t.Errorf("days[0].Date: got %s, want 2025-01-01", days[0].Date)
	}
}
