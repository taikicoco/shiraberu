package html

import (
	"embed"
	"html/template"
	"io"

	"github.com/taikicoco/shiraberu/internal/pr"
)

//go:embed templates/*.html
var templateFS embed.FS

var htmlTemplate *template.Template

func init() {
	var err error
	htmlTemplate, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
}

type Summary struct {
	OpenedCount  int
	DraftCount   int
	MergedCount  int
	ReviewedCount int
	Additions    int
	Deletions    int
}

type HTMLData struct {
	Report      *pr.Report
	Summary     Summary
	Weekdays    []string
	PeriodLabel string
}

func RenderHTML(w io.Writer, report *pr.Report) error {
	summary := calcSummary(report)
	data := HTMLData{
		Report:      report,
		Summary:     summary,
		Weekdays:    []string{"日", "月", "火", "水", "木", "金", "土"},
		PeriodLabel: formatPeriod(report.StartDate, report.EndDate),
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
