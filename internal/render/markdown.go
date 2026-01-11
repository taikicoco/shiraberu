package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
)

func formatPeriod(start, end time.Time) string {
	if start.Equal(end) {
		return start.Format("2006/01/02")
	}
	return start.Format("2006/01/02") + " 〜 " + end.Format("2006/01/02")
}

func RenderMarkdown(w io.Writer, report *pr.Report) error {
	periodLabel := formatPeriod(report.StartDate, report.EndDate)

	fmt.Fprintf(w, "# PR Log (%s)\n\n", periodLabel)
	fmt.Fprintf(w, "Organization: %s\n", report.Org)
	fmt.Fprintf(w, "Generated: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04"))

	if len(report.Days) == 0 {
		fmt.Fprintln(w, "No pull requests found.")
		return nil
	}

	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	for _, day := range report.Days {
		weekday := weekdays[day.Date.Weekday()]
		fmt.Fprintf(w, "## %s (%s)\n\n", day.Date.Format("2006-01-02"), weekday)

		if len(day.Opened) > 0 {
			fmt.Fprintln(w, "### Opened")
			for _, p := range day.Opened {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}

		if len(day.Draft) > 0 {
			fmt.Fprintln(w, "### Draft")
			for _, p := range day.Draft {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}

		if len(day.Merged) > 0 {
			fmt.Fprintln(w, "### Merged")
			for _, p := range day.Merged {
				writePRLine(w, p)
			}
			fmt.Fprintln(w)
		}

		if len(day.Reviewed) > 0 {
			fmt.Fprintln(w, "### Reviewed")
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
	return strings.ToUpper(s[:1]) + s[1:]
}
