package pr

import (
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

// ActivityType はPRへの関わり方を表す
type ActivityType string

const (
	ActivityCreated   ActivityType = "created"
	ActivityMerged    ActivityType = "merged"
	ActivityClosed    ActivityType = "closed"
	ActivityReviewed  ActivityType = "reviewed"
	ActivityCommented ActivityType = "commented"
)

// Activity はPRへの関わり1件を表す
type Activity struct {
	Type        ActivityType
	OccurredAt  time.Time
	URL         string
	ReviewState string
	Body        string
	PullRequest github.PullRequestDetail
}

// ActivityReport は期間内の関わりを古い順に持つ
type ActivityReport struct {
	GeneratedAt time.Time
	StartDate   time.Time
	EndDate     time.Time
	Org         string
	Username    string
	Activities  []Activity
}

// DetailSearcher はPRの詳細検索を抽象化するインターフェース
type DetailSearcher interface {
	SearchPRDetails(org, query, dateFilter, username string) ([]github.PullRequestDetail, error)
}

// ActivityFetcher は期間内の関わりを集める
type ActivityFetcher struct {
	client DetailSearcher
}

// NewActivityFetcher は ActivityFetcher を作成する
func NewActivityFetcher(client DetailSearcher) *ActivityFetcher {
	return &ActivityFetcher{client: client}
}

// Fetch は startDate から endDate まで（JST）の関わりを集める。
//
// 自分のPRは作成日・マージ日・閉じた日で引けるので、月ごとに分けて検索する。
// レビューとコメントは日付で検索できない。PRの更新日はレビューやコメントの日以降になるので、
// 開始日以降に更新されたPRを検索し、レビューとコメントの日時で期間内に絞る
func (f *ActivityFetcher) Fetch(org, username string, startDate, endDate time.Time) (*ActivityReport, error) {
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, timezone.JST)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, timezone.JST)

	authored := []struct {
		activityType ActivityType
		query        string
		qualifier    string
		occurredAt   func(github.PullRequestDetail) *time.Time
	}{
		{ActivityCreated, "is:pr author:" + username, "created", func(d github.PullRequestDetail) *time.Time { return &d.CreatedAt }},
		{ActivityMerged, "is:pr author:" + username + " is:merged", "merged", func(d github.PullRequestDetail) *time.Time { return d.MergedAt }},
		{ActivityClosed, "is:pr author:" + username + " is:closed is:unmerged", "closed", func(d github.PullRequestDetail) *time.Time { return d.ClosedAt }},
	}

	collected := newActivityCollector(start, end)

	for _, chunk := range splitDateRange(start, end) {
		dateRange := chunk.start.Format(time.RFC3339) + ".." + chunk.end.Format(time.RFC3339)
		for _, search := range authored {
			details, err := f.client.SearchPRDetails(org, search.query, search.qualifier+":"+dateRange, username)
			if err != nil {
				return nil, err
			}
			for _, detail := range details {
				if at := search.occurredAt(detail); at != nil {
					collected.add(Activity{Type: search.activityType, OccurredAt: *at, URL: detail.URL, PullRequest: detail})
				}
			}
		}
	}

	updatedSince := "updated:>=" + start.Format(time.RFC3339)

	reviewed, err := f.client.SearchPRDetails(org, "is:pr reviewed-by:"+username+" -author:"+username, updatedSince, username)
	if err != nil {
		return nil, err
	}
	for _, detail := range reviewed {
		for _, review := range detail.Reviews {
			collected.add(Activity{
				Type:        ActivityReviewed,
				OccurredAt:  review.SubmittedAt,
				URL:         review.URL,
				ReviewState: review.State,
				Body:        review.Body,
				PullRequest: detail,
			})
		}
	}

	commented, err := f.client.SearchPRDetails(org, "is:pr commenter:"+username+" -author:"+username, updatedSince, username)
	if err != nil {
		return nil, err
	}
	for _, detail := range commented {
		for _, comment := range detail.Comments {
			collected.add(Activity{
				Type:        ActivityCommented,
				OccurredAt:  comment.CreatedAt,
				URL:         comment.URL,
				Body:        comment.Body,
				PullRequest: detail,
			})
		}
	}

	return &ActivityReport{
		GeneratedAt: time.Now(),
		StartDate:   startDate,
		EndDate:     endDate,
		Org:         org,
		Username:    username,
		Activities:  collected.sorted(),
	}, nil
}

// activityCollector は期間内の関わりを重複なく集める
type activityCollector struct {
	start      time.Time
	end        time.Time
	seen       map[string]bool
	activities []Activity
}

func newActivityCollector(start, end time.Time) *activityCollector {
	return &activityCollector{start: start, end: end, seen: make(map[string]bool)}
}

func (c *activityCollector) add(activity Activity) {
	if activity.OccurredAt.Before(c.start) || activity.OccurredAt.After(c.end) {
		return
	}
	key := string(activity.Type) + " " + activity.URL + " " + activity.OccurredAt.UTC().Format(time.RFC3339)
	if c.seen[key] {
		return
	}
	c.seen[key] = true
	c.activities = append(c.activities, activity)
}

func (c *activityCollector) sorted() []Activity {
	sort.SliceStable(c.activities, func(i, j int) bool {
		return c.activities[i].OccurredAt.Before(c.activities[j].OccurredAt)
	})
	return c.activities
}
