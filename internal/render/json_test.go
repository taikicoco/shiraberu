package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

func TestRenderJSON(t *testing.T) {
	mergedAt := time.Date(2026, 5, 2, 1, 0, 0, 0, time.UTC)
	report := &pr.ActivityReport{
		GeneratedAt: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		StartDate:   time.Date(2026, 5, 1, 0, 0, 0, 0, timezone.JST),
		EndDate:     time.Date(2026, 5, 31, 0, 0, 0, 0, timezone.JST),
		Org:         "org",
		Username:    "me",
		Activities: []pr.Activity{
			{
				Type:        pr.ActivityReviewed,
				OccurredAt:  time.Date(2026, 5, 2, 15, 30, 0, 0, time.UTC),
				URL:         "https://github.com/org/repo/pull/1#r1",
				ReviewState: "APPROVED",
				Body:        "a <b> & c",
				PullRequest: github.PullRequestDetail{
					Number:     1,
					Title:      "Title",
					URL:        "https://github.com/org/repo/pull/1",
					Repository: "org/repo",
					State:      "merged",
					Author:     "someone",
					BaseRef:    "main",
					Body:       "pr body",
					CreatedAt:  time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					MergedAt:   &mergedAt,
					UpdatedAt:  time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
					Additions:  10,
				},
			},
			{
				Type:        pr.ActivityCreated,
				OccurredAt:  time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
				URL:         "https://github.com/org/repo/pull/2",
				PullRequest: github.PullRequestDetail{Number: 2},
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, report); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	var got activityReportJSON
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output should be valid JSON: %v\n%s", err, buf.String())
	}

	if got.Since != "2026-05-01" || got.Until != "2026-05-31" || got.Org != "org" || got.Username != "me" {
		t.Errorf("header: %+v", got)
	}
	if len(got.Activities) != 2 {
		t.Fatalf("activities: got %d, want 2", len(got.Activities))
	}

	reviewed := got.Activities[0]
	if reviewed.Type != "reviewed" || reviewed.ReviewState != "APPROVED" || reviewed.Body != "a <b> & c" {
		t.Errorf("reviewed: %+v", reviewed)
	}
	// 日時は JST で出す
	if reviewed.OccurredAt != "2026-05-03T00:30:00+09:00" {
		t.Errorf("OccurredAt: got %q", reviewed.OccurredAt)
	}
	if reviewed.PullRequest.MergedAt == nil || *reviewed.PullRequest.MergedAt != "2026-05-02T10:00:00+09:00" {
		t.Errorf("MergedAt: got %v", reviewed.PullRequest.MergedAt)
	}
	if reviewed.PullRequest.ClosedAt != nil {
		t.Errorf("ClosedAt should be null: got %v", *reviewed.PullRequest.ClosedAt)
	}

	// HTML をエスケープしない
	if !strings.Contains(buf.String(), "a <b> & c") {
		t.Errorf("body should not be HTML-escaped: %s", buf.String())
	}
	// 自分のレビュー・コメント以外では、空の項目を出さない
	if strings.Contains(strings.Split(buf.String(), `"type": "created"`)[1], "reviewState") {
		t.Errorf("created should not have reviewState: %s", buf.String())
	}
}

func TestRenderJSON_NoActivities(t *testing.T) {
	var buf bytes.Buffer
	report := &pr.ActivityReport{Org: "org", Username: "me"}
	if err := RenderJSON(&buf, report); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	// 0 件でも null ではなく空の配列にする
	if !strings.Contains(buf.String(), `"activities": []`) {
		t.Errorf("activities should be an empty array: %s", buf.String())
	}
}
