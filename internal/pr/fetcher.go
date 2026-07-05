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

type dateChunk struct {
	start time.Time
	end   time.Time
}

func splitDateRange(start, end time.Time) []dateChunk {
	var chunks []dateChunk
	current := start
	for current.Before(end) {
		// End of current month
		chunkEnd := time.Date(current.Year(), current.Month()+1, 0, 23, 59, 59, 0, current.Location())
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		chunks = append(chunks, dateChunk{start: current, end: chunkEnd})
		// Move to first day of next month
		current = time.Date(chunkEnd.Year(), chunkEnd.Month()+1, 1, 0, 0, 0, 0, chunkEnd.Location())
	}
	return chunks
}

func dedup(prs []github.PullRequest) []github.PullRequest {
	seen := make(map[string]bool)
	var result []github.PullRequest
	for _, pr := range prs {
		if pr.URL != "" && seen[pr.URL] {
			continue
		}
		seen[pr.URL] = true
		result = append(result, pr)
	}
	return result
}

func (f *Fetcher) Fetch(org, username string, startDate, endDate time.Time) (*Report, error) {
	startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, timezone.JST)
	endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, timezone.JST)

	chunks := splitDateRange(startTime, endTime)

	var allOpened, allMerged, allReviewed []github.PullRequest

	for _, chunk := range chunks {
		dateRange := chunk.start.Format(time.RFC3339) + ".." + chunk.end.Format(time.RFC3339)

		openedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:open", "created:"+dateRange)
		if err != nil {
			return nil, err
		}
		allOpened = append(allOpened, openedPRs...)

		mergedPRs, err := f.client.SearchPRs(org, "is:pr author:"+username+" is:merged", "merged:"+dateRange)
		if err != nil {
			return nil, err
		}
		allMerged = append(allMerged, mergedPRs...)

		reviewedPRs, err := f.client.SearchPRs(org, "is:pr reviewed-by:"+username+" -author:"+username, "updated:"+dateRange)
		if err != nil {
			return nil, err
		}
		allReviewed = append(allReviewed, reviewedPRs...)
	}

	days := groupByDate(dedup(allOpened), dedup(allMerged), dedup(allReviewed))

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
