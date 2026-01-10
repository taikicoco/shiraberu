package demo

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
)

var repos = []string{
	"api-server",
	"web-frontend",
	"mobile-app",
	"infra-terraform",
	"shared-libs",
}

var prTitles = []string{
	"feat: Add user authentication",
	"fix: Resolve memory leak in cache",
	"refactor: Extract common utilities",
	"docs: Update API documentation",
	"chore: Bump dependencies",
	"feat: Implement dark mode",
	"fix: Handle edge case in parser",
	"feat: Add export to CSV feature",
	"refactor: Migrate to new database",
	"fix: Correct timezone handling",
	"feat: Add notification system",
	"chore: Update CI/CD pipeline",
	"feat: Implement search functionality",
	"fix: Fix pagination bug",
	"refactor: Improve error handling",
}

// Demo generation constants
const (
	weekendActivityRate = 0.3  // Probability of activity on weekends
	draftPRRate         = 0.3  // Probability of generating a draft PR (1 - 0.7)
	maxOpenedPRs        = 4    // Max opened PRs per day (0-3)
	maxMergedPRs        = 5    // Max merged PRs per day (0-4)
	maxReviewedPRs      = 4    // Max reviewed PRs per day (0-3)
	maxAdditions        = 500  // Max line additions
	minAdditions        = 10   // Min line additions
	maxDeletions        = 200  // Max line deletions
	minDeletions        = 5    // Min line deletions
	maxChangedFiles     = 20   // Max changed files
	maxComments         = 10   // Max comments
)

// GenerateReport generates demo data for the given date range
func GenerateReport(startDate, endDate time.Time) (*pr.Report, *pr.Report) {
	jst := time.FixedZone("JST", 9*60*60)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate current period
	report := generatePeriodReport(startDate, endDate, jst, r, "demo-org")

	// Generate previous period for comparison
	duration := endDate.Sub(startDate)
	prevEndDate := startDate.AddDate(0, 0, -1)
	prevStartDate := prevEndDate.Add(-duration)
	previousReport := generatePeriodReport(prevStartDate, prevEndDate, jst, r, "demo-org")

	return report, previousReport
}

func generatePeriodReport(startDate, endDate time.Time, jst *time.Location, r *rand.Rand, org string) *pr.Report {
	days := []pr.DailyPRs{}

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		date := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, jst)

		// Skip some days randomly (weekends have less activity)
		weekday := date.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			if r.Float32() > weekendActivityRate {
				continue
			}
		}

		day := pr.DailyPRs{Date: date}

		// Generate opened PRs
		openedCount := r.Intn(maxOpenedPRs)
		for i := 0; i < openedCount; i++ {
			day.Opened = append(day.Opened, generatePR(r, date, "open"))
		}

		// Generate draft PRs
		if r.Float32() < draftPRRate {
			day.Draft = append(day.Draft, generatePR(r, date, "draft"))
		}

		// Generate merged PRs
		mergedCount := r.Intn(maxMergedPRs)
		for i := 0; i < mergedCount; i++ {
			p := generatePR(r, date, "merged")
			mergedAt := date.Add(time.Duration(r.Intn(24)) * time.Hour)
			p.MergedAt = &mergedAt
			day.Merged = append(day.Merged, p)
		}

		// Generate reviewed PRs
		reviewedCount := r.Intn(maxReviewedPRs)
		for i := 0; i < reviewedCount; i++ {
			day.Reviewed = append(day.Reviewed, generatePR(r, date, "open"))
		}

		// Only add day if it has any PRs
		if len(day.Opened) > 0 || len(day.Draft) > 0 || len(day.Merged) > 0 || len(day.Reviewed) > 0 {
			days = append(days, day)
		}
	}

	// Sort by date descending (newest first) to match fetcher.go behavior
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.After(days[j].Date)
	})

	return &pr.Report{
		GeneratedAt: time.Now(),
		StartDate:   startDate,
		EndDate:     endDate,
		Org:         org,
		Days:        days,
	}
}

func generatePR(r *rand.Rand, date time.Time, state string) github.PullRequest {
	repo := repos[r.Intn(len(repos))]
	title := prTitles[r.Intn(len(prTitles))]

	additions := r.Intn(maxAdditions) + minAdditions
	deletions := r.Intn(maxDeletions) + minDeletions

	return github.PullRequest{
		Title:        title,
		URL:          "https://github.com/demo-org/" + repo + "/pull/" + randomPRNumber(r),
		Repository:   repo,
		State:        state,
		IsDraft:      state == "draft",
		CreatedAt:    date,
		UpdatedAt:    date.Add(time.Duration(r.Intn(48)) * time.Hour),
		Additions:    additions,
		Deletions:    deletions,
		ChangedFiles: r.Intn(maxChangedFiles) + 1,
		Comments:     r.Intn(maxComments),
	}
}

func randomPRNumber(r *rand.Rand) string {
	n := r.Intn(9000) + 1000
	return fmt.Sprintf("%d", n)
}
