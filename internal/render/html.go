package render

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

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
