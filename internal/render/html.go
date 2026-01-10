package render

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"sort"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
)

//go:embed templates/*.html
var templateFS embed.FS

var htmlTemplate *template.Template

func init() {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"json": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
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
	Week          string // "1/1 〜 1/7" 形式
	StartDate     string // "2006-01-02" 形式
	EndDate       string // "2006-01-02" 形式
	OpenedCount   int
	DraftCount    int
	MergedCount   int
	ReviewedCount int
}

// MonthlyStat は月別統計データ
type MonthlyStat struct {
	Month         string // "Jan 2006" 形式
	StartDate     string // "2006-01-02" 形式
	EndDate       string // "2006-01-02" 形式
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

// DayJSON はJavaScript用の日別データ
type DayJSON struct {
	Date     string   `json:"date"`
	Opened   []PRJSON `json:"opened"`
	Draft    []PRJSON `json:"draft"`
	Merged   []PRJSON `json:"merged"`
	Reviewed []PRJSON `json:"reviewed"`
}

// PRJSON はJavaScript用のPRデータ
type PRJSON struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	Repository string `json:"repository"`
	State      string `json:"state"`
	IsDraft    bool   `json:"isDraft"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	Comments   int    `json:"comments"`
}

type HTMLData struct {
	Report            *pr.Report
	Summary           Summary
	SummaryDiff       SummaryDiff
	DailyStats        []DailyStat
	WeeklyStats       []WeeklyStat
	MonthlyStats      []MonthlyStat
	RepoStats         []RepoStat
	Weekdays          []string
	PeriodLabel       string
	DaysJSON          []DayJSON
	OriginalStartDate string
	OriginalEndDate   string
}

func RenderHTML(w io.Writer, report *pr.Report, previousReport *pr.Report) error {
	summary := calcSummary(report)
	dailyStats := calcDailyStats(report)
	weeklyStats := calcWeeklyStats(report)
	monthlyStats := calcMonthlyStats(report)
	repoStats := calcRepoStats(report)
	summaryDiff := calcSummaryDiff(summary, previousReport)
	daysJSON := convertToDaysJSON(report)

	data := HTMLData{
		Report:            report,
		Summary:           summary,
		SummaryDiff:       summaryDiff,
		DailyStats:        dailyStats,
		WeeklyStats:       weeklyStats,
		MonthlyStats:      monthlyStats,
		RepoStats:         repoStats,
		Weekdays:          []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		PeriodLabel:       formatPeriod(report.StartDate, report.EndDate),
		DaysJSON:          daysJSON,
		OriginalStartDate: report.StartDate.Format("2006-01-02"),
		OriginalEndDate:   report.EndDate.Format("2006-01-02"),
	}
	return htmlTemplate.ExecuteTemplate(w, "report.html", data)
}

func convertToDaysJSON(report *pr.Report) []DayJSON {
	var days []DayJSON
	for _, day := range report.Days {
		d := DayJSON{
			Date:     day.Date.Format("2006-01-02"),
			Opened:   convertPRsToJSON(day.Opened),
			Draft:    convertPRsToJSON(day.Draft),
			Merged:   convertPRsToJSON(day.Merged),
			Reviewed: convertPRsToJSON(day.Reviewed),
		}
		days = append(days, d)
	}
	return days
}

func convertPRsToJSON(prs []github.PullRequest) []PRJSON {
	result := make([]PRJSON, 0, len(prs))
	for _, p := range prs {
		result = append(result, PRJSON{
			Title:      p.Title,
			URL:        p.URL,
			Repository: p.Repository,
			State:      p.State,
			IsDraft:    p.IsDraft,
			Additions:  p.Additions,
			Deletions:  p.Deletions,
			Comments:   p.Comments,
		})
	}
	return result
}

func calcSummary(report *pr.Report) Summary {
	var s Summary
	for _, day := range report.Days {
		s.OpenedCount += len(day.Opened)
		s.DraftCount += len(day.Draft)
		s.MergedCount += len(day.Merged)
		s.ReviewedCount += len(day.Reviewed)
		// Additions/Deletions are only counted for merged PRs
		for _, p := range day.Merged {
			s.Additions += p.Additions
			s.Deletions += p.Deletions
		}
	}
	return s
}

func calcDailyStats(report *pr.Report) []DailyStat {
	// 期間内の全日を含むマップを作成
	dayMap := make(map[string]*DailyStat)

	// まず期間内の全日を0で初期化
	for d := report.StartDate; !d.After(report.EndDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayMap[dateStr] = &DailyStat{
			Date: dateStr,
		}
	}

	// PRがある日のデータを埋める
	for _, day := range report.Days {
		dateStr := day.Date.Format("2006-01-02")
		stat, ok := dayMap[dateStr]
		if !ok {
			stat = &DailyStat{Date: dateStr}
			dayMap[dateStr] = stat
		}

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

		stat.OpenedCount = len(day.Opened)
		stat.DraftCount = len(day.Draft)
		stat.MergedCount = len(day.Merged)
		stat.ReviewedCount = len(day.Reviewed)
		stat.Additions = additions
		stat.Deletions = deletions
		stat.TotalPRs = len(day.Opened) + len(day.Draft) + len(day.Merged) + len(day.Reviewed)
	}

	// スライスに変換
	stats := make([]DailyStat, 0, len(dayMap))
	for _, stat := range dayMap {
		stats = append(stats, *stat)
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
		stat    *WeeklyStat
		weekKey string
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
			weekEnd := weekStart.AddDate(0, 0, 6) // Sunday
			weekLabel := fmt.Sprintf("%d/%d 〜 %d/%d", weekStart.Month(), weekStart.Day(), weekEnd.Month(), weekEnd.Day())
			weekMap[weekKey] = &weekData{
				stat: &WeeklyStat{
					Week:      weekLabel,
					StartDate: weekStart.Format("2006-01-02"),
					EndDate:   weekEnd.Format("2006-01-02"),
				},
				weekKey: weekKey,
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
		monthLabel := day.Date.Format("Jan 2006")

		if _, ok := monthMap[monthKey]; !ok {
			// Calculate month start and end dates
			monthStart := time.Date(day.Date.Year(), day.Date.Month(), 1, 0, 0, 0, 0, day.Date.Location())
			monthEnd := monthStart.AddDate(0, 1, -1) // Last day of month
			monthMap[monthKey] = &MonthlyStat{
				Month:     monthLabel,
				StartDate: monthStart.Format("2006-01-02"),
				EndDate:   monthEnd.Format("2006-01-02"),
			}
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
