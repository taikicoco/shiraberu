package github

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/taikicoco/shiraberu/internal/errors"
)

// scriptedExecutor は呼び出しごとに決まったレスポンスを順番に返す
type scriptedExecutor struct {
	responses []string
	calls     [][]string
}

func (e *scriptedExecutor) Execute(_ string, args ...string) ([]byte, error) {
	e.calls = append(e.calls, args)
	response := e.responses[len(e.calls)-1]
	return []byte(response), nil
}

const firstDetailPage = `{"data":{"search":{"issueCount":2,"pageInfo":{"hasNextPage":true,"endCursor":"next"},"nodes":[
  {"number":1,"title":"First","url":"https://github.com/org/repo/pull/1","state":"MERGED","isDraft":false,"body":"pr body",
   "baseRefName":"main","createdAt":"2026-05-01T01:00:00Z","mergedAt":"2026-05-02T01:00:00Z","closedAt":"2026-05-02T01:00:00Z",
   "updatedAt":"2026-05-03T01:00:00Z","additions":10,"deletions":2,"changedFiles":3,"author":{"login":"someone"},
   "repository":{"nameWithOwner":"org/repo"},
   "reviews":{"nodes":[
     {"state":"APPROVED","submittedAt":"2026-05-02T00:00:00Z","url":"https://github.com/org/repo/pull/1#r1","body":"lgtm"},
     {"state":"PENDING","submittedAt":null,"url":"https://github.com/org/repo/pull/1#r2","body":"draft"}
   ]},
   "comments":{"totalCount":2,"nodes":[
     {"createdAt":"2026-05-01T02:00:00Z","url":"https://github.com/org/repo/pull/1#c1","body":"mine","author":{"login":"ME"}},
     {"createdAt":"2026-05-01T03:00:00Z","url":"https://github.com/org/repo/pull/1#c2","body":"theirs","author":{"login":"someone"}}
   ]}}
]}}}`

const secondDetailPage = `{"data":{"search":{"issueCount":2,"pageInfo":{"hasNextPage":false,"endCursor":""},"nodes":[
  {"number":2,"title":"Second","url":"https://github.com/org/repo/pull/2","state":"OPEN","isDraft":true,"body":"",
   "baseRefName":"feature","createdAt":"2026-05-04T01:00:00Z","mergedAt":null,"closedAt":null,
   "updatedAt":"2026-05-04T01:00:00Z","additions":0,"deletions":0,"changedFiles":0,"author":null,
   "repository":{"nameWithOwner":"org/repo"},"reviews":{"nodes":[]},"comments":{"totalCount":0,"nodes":[]}}
]}}}`

func TestSearchPRDetails(t *testing.T) {
	executor := &scriptedExecutor{responses: []string{firstDetailPage, secondDetailPage}}
	client := &Client{executor: executor}

	details, err := client.SearchPRDetails("org", "is:pr author:me", "created:2026-05-01..2026-05-31", "me")
	if err != nil {
		t.Fatalf("SearchPRDetails: %v", err)
	}

	if len(executor.calls) != 2 {
		t.Fatalf("calls: got %d, want 2", len(executor.calls))
	}
	firstCall := strings.Join(executor.calls[0], " ")
	for _, want := range []string{"q=is:pr author:me org:org created:2026-05-01..2026-05-31", "login=me"} {
		if !strings.Contains(firstCall, want) {
			t.Errorf("first call should contain %q: %s", want, firstCall)
		}
	}
	if !strings.Contains(strings.Join(executor.calls[1], " "), "cursor=next") {
		t.Errorf("second call should use the cursor: %v", executor.calls[1])
	}

	if len(details) != 2 {
		t.Fatalf("details: got %d, want 2", len(details))
	}

	first := details[0]
	if first.Number != 1 || first.Repository != "org/repo" || first.State != "merged" {
		t.Errorf("first detail: %+v", first)
	}
	if first.Author != "someone" || first.BaseRef != "main" || first.Body != "pr body" {
		t.Errorf("first detail: %+v", first)
	}
	if first.MergedAt == nil || first.ClosedAt == nil {
		t.Errorf("merged and closed times should be set: %+v", first)
	}
	if first.CommentCount != 2 {
		t.Errorf("CommentCount: got %d, want 2", first.CommentCount)
	}

	// 提出前のレビューは含めない
	if len(first.Reviews) != 1 || first.Reviews[0].State != "APPROVED" {
		t.Errorf("Reviews: %+v", first.Reviews)
	}
	// 自分のコメントだけを残す。ログイン名の大文字小文字は区別しない
	if len(first.Comments) != 1 || first.Comments[0].Body != "mine" {
		t.Errorf("Comments: %+v", first.Comments)
	}

	second := details[1]
	if second.MergedAt != nil || second.ClosedAt != nil || second.Author != "" || !second.IsDraft {
		t.Errorf("second detail: %+v", second)
	}
}

func TestSearchPRDetails_TooManyResults(t *testing.T) {
	executor := &scriptedExecutor{responses: []string{
		`{"data":{"search":{"issueCount":1001,"pageInfo":{"hasNextPage":true,"endCursor":"next"},"nodes":[]}}}`,
	}}
	client := &Client{executor: executor}

	_, err := client.SearchPRDetails("org", "is:pr", "updated:>=2026-05-01", "me")
	if !errors.Is(err, apperrors.ErrTooManyResults) {
		t.Errorf("err: got %v, want ErrTooManyResults", err)
	}
}
