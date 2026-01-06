package pr

import (
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
)

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
	Days        []DailyPRs
}

type Fetcher struct {
	client *github.Client
}

func NewFetcher(client *github.Client) *Fetcher {
	return &Fetcher{client: client}
}

func (f *Fetcher) Fetch(org string, startDate, endDate time.Time) (*Report, error) {
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	username := f.client.Username()

	dateRange := startStr + ".." + endStr

	// オープンしたPR: 作成日で絞り込み
	openedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:open", "created:"+dateRange)
	if err != nil {
		return nil, err
	}

	// マージしたPR: マージ日で絞り込み
	mergedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:merged", "merged:"+dateRange)
	if err != nil {
		return nil, err
	}

	// レビューしたPR: 更新日で絞り込み（レビュー日は直接取れないため）
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

		dateStr := date.Format("2006-01-02")
		if _, ok := dateMap[dateStr]; !ok {
			dateMap[dateStr] = &DailyPRs{
				Date: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()),
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
