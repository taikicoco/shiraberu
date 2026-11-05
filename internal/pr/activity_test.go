package pr

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

// fakeDetailSearcher はクエリに含まれる語でレスポンスを返す
type fakeDetailSearcher struct {
	responses map[string][]github.PullRequestDetail
	err       error
	queries   []string
}

func (f *fakeDetailSearcher) SearchPRDetails(_, query, dateFilter, _ string) ([]github.PullRequestDetail, error) {
	f.queries = append(f.queries, query+" "+dateFilter)
	if f.err != nil {
		return nil, f.err
	}
	for pattern, details := range f.responses {
		if strings.Contains(query+" "+dateFilter, pattern) {
			return details, nil
		}
	}
	return nil, nil
}

func jst(month time.Month, day, hour int) time.Time {
	return time.Date(2026, month, day, hour, 0, 0, 0, timezone.JST)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func TestActivityFetcher_Fetch(t *testing.T) {
	const prURL = "https://github.com/org/repo/pull/"

	searcher := &fakeDetailSearcher{responses: map[string][]github.PullRequestDetail{
		"created:": {{URL: prURL + "1", CreatedAt: jst(5, 2, 10)}},
		"merged:":  {{URL: prURL + "1", CreatedAt: jst(5, 2, 10), MergedAt: timePtr(jst(5, 3, 10))}},
		"closed:":  {{URL: prURL + "2", CreatedAt: jst(5, 1, 10), ClosedAt: timePtr(jst(5, 4, 10))}},
		"reviewed-by:": {{
			URL: prURL + "3",
			Reviews: []github.Review{
				{State: "APPROVED", SubmittedAt: jst(5, 5, 10), URL: prURL + "3#r1", Body: "lgtm"},
				{State: "COMMENTED", SubmittedAt: jst(4, 30, 10), URL: prURL + "3#r0"},
			},
		}},
		"commenter:": {{
			URL: prURL + "4",
			Comments: []github.Comment{
				{CreatedAt: jst(5, 1, 9), URL: prURL + "4#c1", Body: "first"},
				{CreatedAt: jst(6, 1, 9), URL: prURL + "4#c2", Body: "after the range"},
			},
		}},
	}}

	report, err := NewActivityFetcher(searcher).Fetch("org", "me", jst(5, 1, 0), jst(5, 31, 0))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	want := []struct {
		activityType ActivityType
		url          string
	}{
		{ActivityCommented, prURL + "4#c1"},
		{ActivityCreated, prURL + "1"},
		{ActivityMerged, prURL + "1"},
		{ActivityClosed, prURL + "2"},
		{ActivityReviewed, prURL + "3#r1"},
	}
	if len(report.Activities) != len(want) {
		t.Fatalf("activities: got %d, want %d: %+v", len(report.Activities), len(want), report.Activities)
	}
	for i, w := range want {
		got := report.Activities[i]
		if got.Type != w.activityType || got.URL != w.url {
			t.Errorf("activities[%d]: got %s %s, want %s %s", i, got.Type, got.URL, w.activityType, w.url)
		}
	}

	reviewed := report.Activities[4]
	if reviewed.ReviewState != "APPROVED" || reviewed.Body != "lgtm" {
		t.Errorf("reviewed: %+v", reviewed)
	}
}

func TestActivityFetcher_Fetch_Queries(t *testing.T) {
	searcher := &fakeDetailSearcher{}

	// 5 月と 6 月にまたがるので、自分の PR の検索は 2 回ずつになる
	if _, err := NewActivityFetcher(searcher).Fetch("org", "me", jst(5, 31, 0), jst(6, 1, 0)); err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	joined := strings.Join(searcher.queries, "\n")
	for _, want := range []string{
		"is:pr author:me created:2026-05-31T00:00:00+09:00..2026-05-31T23:59:59+09:00",
		"is:pr author:me created:2026-06-01T00:00:00+09:00..2026-06-01T23:59:59+09:00",
		"is:pr author:me is:merged merged:",
		"is:pr author:me is:closed is:unmerged closed:",
		"is:pr reviewed-by:me -author:me updated:>=2026-05-31T00:00:00+09:00",
		"is:pr commenter:me -author:me updated:>=2026-05-31T00:00:00+09:00",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("queries should contain %q:\n%s", want, joined)
		}
	}
	if len(searcher.queries) != 8 {
		t.Errorf("queries: got %d, want 8", len(searcher.queries))
	}
}

func TestActivityFetcher_Fetch_Deduplicates(t *testing.T) {
	detail := github.PullRequestDetail{
		URL:     "https://github.com/org/repo/pull/1",
		Reviews: []github.Review{{State: "APPROVED", SubmittedAt: jst(5, 2, 10), URL: "https://github.com/org/repo/pull/1#r1"}},
	}
	searcher := &fakeDetailSearcher{responses: map[string][]github.PullRequestDetail{
		"reviewed-by:": {detail, detail},
	}}

	report, err := NewActivityFetcher(searcher).Fetch("org", "me", jst(5, 1, 0), jst(5, 31, 0))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(report.Activities) != 1 {
		t.Errorf("activities: got %d, want 1", len(report.Activities))
	}
}

func TestActivityFetcher_Fetch_Error(t *testing.T) {
	wantErr := errors.New("search failed")
	searcher := &fakeDetailSearcher{err: wantErr}

	if _, err := NewActivityFetcher(searcher).Fetch("org", "me", jst(5, 1, 0), jst(5, 31, 0)); !errors.Is(err, wantErr) {
		t.Errorf("err: got %v, want %v", err, wantErr)
	}
}
