package html

import (
	"fmt"
	"io"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
)

func formatPeriod(start, end time.Time) string {
	if start.Equal(end) {
		return start.Format("2006-01-02")
	}
	return start.Format("2006-01-02") + " 〜 " + end.Format("2006-01-02")
}

func RenderMarkdown(w io.Writer, report *pr.Report) error {
	periodLabel := formatPeriod(report.StartDate, report.EndDate)

	fmt.Fprintf(w, "# PR Log (%s)\n\n", periodLabel)
	fmt.Fprintf(w, "Organization: %s\n", report.Org)
	fmt.Fprintf(w, "生成日時: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04"))

	if len(report.Days) == 0 {
		fmt.Fprintln(w, "該当するPRはありませんでした。")
		return nil
	}

	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}

	for _, day := range report.Days {
		weekday := weekdays[day.Date.Weekday()]
		fmt.Fprintf(w, "## %s (%s)\n\n", day.Date.Format("2006-01-02"), weekday)

		if len(day.Opened) > 0 {
			fmt.Fprintln(w, "### オープンしたPR")
			for _, p := range day.Opened {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}

		if len(day.Merged) > 0 {
			fmt.Fprintln(w, "### マージしたPR")
			for _, p := range day.Merged {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}

		if len(day.Reviewed) > 0 {
			fmt.Fprintln(w, "### レビューしたPR")
			for _, p := range day.Reviewed {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}

func writePRLine(w io.Writer, p github.PullRequest) {
	state := capitalize(p.State)
	_, _ = fmt.Fprintf(w, "- [%s](%s) - %s (%s)\n", p.Title, p.URL, p.Repository, state)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
