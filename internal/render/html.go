package render

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"sort"

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
	// 期間内の全日を含むマップを作成
	dayMap := make(map[string]DayJSON)

	// まず期間内の全日を空で初期化
	for d := report.StartDate; !d.After(report.EndDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayMap[dateStr] = DayJSON{
			Date:     dateStr,
			Opened:   []PRJSON{},
			Draft:    []PRJSON{},
			Merged:   []PRJSON{},
			Reviewed: []PRJSON{},
		}
	}

	// PRがある日のデータを埋める
	for _, day := range report.Days {
		dateStr := day.Date.Format("2006-01-02")
		dayMap[dateStr] = DayJSON{
			Date:     dateStr,
			Opened:   convertPRsToJSON(day.Opened),
			Draft:    convertPRsToJSON(day.Draft),
			Merged:   convertPRsToJSON(day.Merged),
			Reviewed: convertPRsToJSON(day.Reviewed),
		}
	}

	// スライスに変換してソート
	days := make([]DayJSON, 0, len(dayMap))
	for _, day := range dayMap {
		days = append(days, day)
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

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
