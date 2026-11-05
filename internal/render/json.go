package render

import (
	"encoding/json"
	"io"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

const jsonDateLayout = "2006-01-02"

type activityReportJSON struct {
	GeneratedAt string         `json:"generatedAt"`
	Org         string         `json:"org"`
	Username    string         `json:"username"`
	Since       string         `json:"since"`
	Until       string         `json:"until"`
	Activities  []activityJSON `json:"activities"`
}

type activityJSON struct {
	Type        string          `json:"type"`
	OccurredAt  string          `json:"occurredAt"`
	URL         string          `json:"url"`
	ReviewState string          `json:"reviewState,omitempty"`
	Body        string          `json:"body,omitempty"`
	PullRequest pullRequestJSON `json:"pullRequest"`
}

type pullRequestJSON struct {
	Number       int     `json:"number"`
	Title        string  `json:"title"`
	URL          string  `json:"url"`
	Repository   string  `json:"repository"`
	State        string  `json:"state"`
	IsDraft      bool    `json:"isDraft"`
	Author       string  `json:"author"`
	BaseRef      string  `json:"baseRef"`
	Body         string  `json:"body"`
	CreatedAt    string  `json:"createdAt"`
	MergedAt     *string `json:"mergedAt"`
	ClosedAt     *string `json:"closedAt"`
	UpdatedAt    string  `json:"updatedAt"`
	Additions    int     `json:"additions"`
	Deletions    int     `json:"deletions"`
	ChangedFiles int     `json:"changedFiles"`
	CommentCount int     `json:"commentCount"`
}

// RenderJSON は関わりの一覧をJSONで書き出す。日時はJSTで出す
func RenderJSON(w io.Writer, report *pr.ActivityReport) error {
	output := activityReportJSON{
		GeneratedAt: formatJST(report.GeneratedAt),
		Org:         report.Org,
		Username:    report.Username,
		Since:       report.StartDate.Format(jsonDateLayout),
		Until:       report.EndDate.Format(jsonDateLayout),
		Activities:  make([]activityJSON, 0, len(report.Activities)),
	}

	for _, activity := range report.Activities {
		output.Activities = append(output.Activities, activityJSON{
			Type:        string(activity.Type),
			OccurredAt:  formatJST(activity.OccurredAt),
			URL:         activity.URL,
			ReviewState: activity.ReviewState,
			Body:        activity.Body,
			PullRequest: toPullRequestJSON(activity.PullRequest),
		})
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func toPullRequestJSON(detail github.PullRequestDetail) pullRequestJSON {
	return pullRequestJSON{
		Number:       detail.Number,
		Title:        detail.Title,
		URL:          detail.URL,
		Repository:   detail.Repository,
		State:        detail.State,
		IsDraft:      detail.IsDraft,
		Author:       detail.Author,
		BaseRef:      detail.BaseRef,
		Body:         detail.Body,
		CreatedAt:    formatJST(detail.CreatedAt),
		MergedAt:     formatOptionalJST(detail.MergedAt),
		ClosedAt:     formatOptionalJST(detail.ClosedAt),
		UpdatedAt:    formatJST(detail.UpdatedAt),
		Additions:    detail.Additions,
		Deletions:    detail.Deletions,
		ChangedFiles: detail.ChangedFiles,
		CommentCount: detail.CommentCount,
	}
}

func formatJST(value time.Time) string {
	return value.In(timezone.JST).Format(time.RFC3339)
}

func formatOptionalJST(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatJST(*value)
	return &formatted
}
