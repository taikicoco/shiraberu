package pr

import (
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

// PRSearcher はPR検索機能を抽象化するインターフェース
type PRSearcher interface {
	Username() string
	SearchPRs(org string, query string, dateFilter string) ([]github.PullRequest, error)
}

type DailyPRs struct {
	Date     time.Time
	Opened   []github.PullRequest
	Draft    []github.PullRequest
	Merged   []github.PullRequest
	Reviewed []github.PullRequest
}

type Report struct {
	GeneratedAt time.Time
	StartDate   time.Time
	EndDate     time.Time
	Org         string
	Username    string
	Days        []DailyPRs
}

type Fetcher struct {
	client PRSearcher
}

func NewFetcher(client PRSearcher) *Fetcher {
	return &Fetcher{client: client}
}

func (f *Fetcher) Fetch(org, username string, startDate, endDate time.Time) (*Report, error) {

	// JST timezone for date range
	startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, timezone.JST)
	endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, timezone.JST)

	dateRange := startTime.Format(time.RFC3339) + ".." + endTime.Format(time.RFC3339)

	// Opened PRs
	openedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:open", "created:"+dateRange)
	if err != nil {
		return nil, err
	}

	// Merged PRs
	mergedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:merged", "merged:"+dateRange)
	if err != nil {
		return nil, err
	}

	// Reviewed PRs
	reviewedPRs, err := f.client.SearchPRs(org, "is:pr reviewed-by:"+username+" -author:"+username, "updated:"+dateRange)
	if err != nil {
		return nil, err
	}

	days := groupByDate(openedPRs, mergedPRs, reviewedPRs)

	return &Report{
		GeneratedAt: time.Now(),
		StartDate:   startDate,
		EndDate:     endDate,
		Org:         org,
		Username:    username,
		Days:        days,
	}, nil
}

func groupByDate(opened, merged, reviewed []github.PullRequest) []DailyPRs {
	dateMap := make(map[string]*DailyPRs)

	addPR := func(pr github.PullRequest, category string) {
		date := pr.CreatedAt
		if category == "merged" && pr.MergedAt != nil {
			date = *pr.MergedAt
		}
		if category == "reviewed" {
			date = pr.UpdatedAt
		}

		// UTCからJSTに変換してからグループ化
		dateJST := date.In(timezone.JST)
		dateStr := dateJST.Format("2006-01-02")
		if _, ok := dateMap[dateStr]; !ok {
			dateMap[dateStr] = &DailyPRs{
				Date: time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, timezone.JST),
			}
		}

		switch category {
		case "opened":
			dateMap[dateStr].Opened = append(dateMap[dateStr].Opened, pr)
		case "draft":
			dateMap[dateStr].Draft = append(dateMap[dateStr].Draft, pr)
		case "merged":
			dateMap[dateStr].Merged = append(dateMap[dateStr].Merged, pr)
		case "reviewed":
			dateMap[dateStr].Reviewed = append(dateMap[dateStr].Reviewed, pr)
		}
	}

	for _, pr := range opened {
		if pr.IsDraft {
			addPR(pr, "draft")
		} else {
			addPR(pr, "opened")
		}
	}
	for _, pr := range merged {
		addPR(pr, "merged")
	}
	for _, pr := range reviewed {
		addPR(pr, "reviewed")
	}

	days := make([]DailyPRs, 0, len(dateMap))
	for _, d := range dateMap {
		days = append(days, *d)
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.After(days[j].Date)
	})

	return days
}
