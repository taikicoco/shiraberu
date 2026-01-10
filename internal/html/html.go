package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"sort"

	"github.com/taikicoco/shiraberu/internal/pr"
)

//go:embed templates/*.html
var templateFS embed.FS

var htmlTemplate *template.Template

func init() {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	var err error
	htmlTemplate, err = template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
}

type Summary struct {
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
	Additions     int
	Deletions     int
}

// SummaryDiff は前期間との差分
type SummaryDiff struct {
	OpenedDiff   int
	DraftDiff    int
	MergedDiff   int
	ReviewedDiff int
	HasPrevious  bool // 前期間データがあるかどうか
}

// DailyStat はグラフ用の日別統計データ
type DailyStat struct {
	Date          string // "2006-01-02" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
	Additions     int
	Deletions     int
	TotalPRs      int // 日別詳細のサマリー表示用
}

// WeeklyStat は週別統計データ
type WeeklyStat struct {
	Week          string // "1/1週" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
}

// MonthlyStat は月別統計データ
type MonthlyStat struct {
	Month         string // "2006-01" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
}

// RepoStat はリポジトリ別の統計データ
type RepoStat struct {
	Repository string
	Count      int
}

type HTMLData struct {
	Report       *pr.Report
	Summary      Summary
	SummaryDiff  SummaryDiff
	DailyStats   []DailyStat
	WeeklyStats  []WeeklyStat
	MonthlyStats []MonthlyStat
	RepoStats    []RepoStat
	Weekdays     []string
	PeriodLabel  string
}

func RenderHTML(w io.Writer, report *pr.Report, previousReport *pr.Report) error {
	summary := calcSummary(report)
	dailyStats := calcDailyStats(report)
	weeklyStats := calcWeeklyStats(report)
	monthlyStats := calcMonthlyStats(report)
	repoStats := calcRepoStats(report)
	summaryDiff := calcSummaryDiff(summary, previousReport)

	data := HTMLData{
		Report:       report,
		Summary:      summary,
		SummaryDiff:  summaryDiff,
		DailyStats:   dailyStats,
		WeeklyStats:  weeklyStats,
		MonthlyStats: monthlyStats,
		RepoStats:    repoStats,
		Weekdays:     []string{"日", "月", "火", "水", "木", "金", "土"},
		PeriodLabel:  formatPeriod(report.StartDate, report.EndDate),
	}
	return htmlTemplate.ExecuteTemplate(w, "report.html", data)
}

func calcSummary(report *pr.Report) Summary {
	var s Summary
	for _, day := range report.Days {
		s.OpenedCount += len(day.Opened)
		s.DraftCount += len(day.Draft)
		s.MergedCount += len(day.Merged)
		s.ReviewedCount += len(day.Reviewed)
		for _, p := range day.Opened {
			s.Additions += p.Additions
			s.Deletions += p.Deletions
		}
		for _, p := range day.Draft {
			s.Additions += p.Additions
			s.Deletions += p.Deletions
		}
		for _, p := range day.Merged {
			s.Additions += p.Additions
			s.Deletions += p.Deletions
		}
	}
	return s
}

func calcDailyStats(report *pr.Report) []DailyStat {
	stats := make([]DailyStat, 0, len(report.Days))
	for _, day := range report.Days {
		var additions, deletions int
		for _, p := range day.Opened {
			additions += p.Additions
			deletions += p.Deletions
		}
		for _, p := range day.Draft {
			additions += p.Additions
			deletions += p.Deletions
		}
		for _, p := range day.Merged {
			additions += p.Additions
			deletions += p.Deletions
		}
		totalPRs := len(day.Opened) + len(day.Draft) + len(day.Merged) + len(day.Reviewed)
		stats = append(stats, DailyStat{
			Date:          day.Date.Format("2006-01-02"),
			OpenedCount:   len(day.Opened),
			DraftCount:    len(day.Draft),
			MergedCount:   len(day.Merged),
			ReviewedCount: len(day.Reviewed),
			Additions:     additions,
			Deletions:     deletions,
			TotalPRs:      totalPRs,
		})
	}

	// グラフ用に日付昇順（左=過去、右=最新）にソート
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Date < stats[j].Date
	})

	return stats
}

func calcRepoStats(report *pr.Report) []RepoStat {
	repoCount := make(map[string]int)
	for _, day := range report.Days {
		for _, p := range day.Opened {
			repoCount[p.Repository]++
		}
		for _, p := range day.Draft {
			repoCount[p.Repository]++
		}
		for _, p := range day.Merged {
			repoCount[p.Repository]++
		}
		for _, p := range day.Reviewed {
			repoCount[p.Repository]++
		}
	}

	stats := make([]RepoStat, 0, len(repoCount))
	for repo, count := range repoCount {
		stats = append(stats, RepoStat{
			Repository: repo,
			Count:      count,
		})
	}

	// 件数の多い順にソート
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

func calcWeeklyStats(report *pr.Report) []WeeklyStat {
	type weekData struct {
		stat      *WeeklyStat
		startDate string // for label generation
	}
	weekMap := make(map[string]*weekData)

	for _, day := range report.Days {
		year, week := day.Date.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		if _, ok := weekMap[weekKey]; !ok {
			// Calculate week start date (Monday) for label
			weekStart := day.Date.AddDate(0, 0, -int(day.Date.Weekday())+1)
			if day.Date.Weekday() == 0 { // Sunday
				weekStart = day.Date.AddDate(0, 0, -6)
			}
			weekLabel := fmt.Sprintf("%d/%d週", weekStart.Month(), weekStart.Day())
			weekMap[weekKey] = &weekData{
				stat:      &WeeklyStat{Week: weekLabel},
				startDate: weekKey,
			}
		}

		weekMap[weekKey].stat.OpenedCount += len(day.Opened)
		weekMap[weekKey].stat.DraftCount += len(day.Draft)
		weekMap[weekKey].stat.MergedCount += len(day.Merged)
		weekMap[weekKey].stat.ReviewedCount += len(day.Reviewed)
	}

	stats := make([]WeeklyStat, 0, len(weekMap))
	keys := make([]string, 0, len(weekMap))
	for k := range weekMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		stats = append(stats, *weekMap[k].stat)
	}

	return stats
}

func calcMonthlyStats(report *pr.Report) []MonthlyStat {
	monthMap := make(map[string]*MonthlyStat)

	for _, day := range report.Days {
		monthKey := day.Date.Format("2006-01")
		monthLabel := fmt.Sprintf("%d月", day.Date.Month())

		if _, ok := monthMap[monthKey]; !ok {
			monthMap[monthKey] = &MonthlyStat{Month: monthLabel}
		}

		monthMap[monthKey].OpenedCount += len(day.Opened)
		monthMap[monthKey].DraftCount += len(day.Draft)
		monthMap[monthKey].MergedCount += len(day.Merged)
		monthMap[monthKey].ReviewedCount += len(day.Reviewed)
	}

	stats := make([]MonthlyStat, 0, len(monthMap))
	keys := make([]string, 0, len(monthMap))
	for k := range monthMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		stats = append(stats, *monthMap[k])
	}

	return stats
}

func calcSummaryDiff(current Summary, previousReport *pr.Report) SummaryDiff {
	if previousReport == nil {
		return SummaryDiff{HasPrevious: false}
	}

	prev := calcSummary(previousReport)
	return SummaryDiff{
		OpenedDiff:   current.OpenedCount - prev.OpenedCount,
		DraftDiff:    current.DraftCount - prev.DraftCount,
		MergedDiff:   current.MergedCount - prev.MergedCount,
		ReviewedDiff: current.ReviewedCount - prev.ReviewedCount,
		HasPrevious:  true,
	}
}
