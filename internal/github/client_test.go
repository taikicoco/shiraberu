package github

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeState(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MERGED", "merged"},
		{"OPEN", "open"},
		{"CLOSED", "closed"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeState(tt.input)
			if got != tt.want {
				t.Errorf("normalizeState(%q): got %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClient_Username(t *testing.T) {
	c := &Client{username: "testuser"}
	if got := c.Username(); got != "testuser" {
		t.Errorf("Username(): got %q, want %q", got, "testuser")
	}
}

// MockExecutor はテスト用のCommandExecutor実装
type MockExecutor struct {
	responses map[string][]byte
	errors    map[string]error
}

// NewMockExecutor は新しいMockExecutorを作成する
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

// SetResponse は指定コマンドへのレスポンスを設定する
func (m *MockExecutor) SetResponse(cmdPattern string, response []byte) {
	m.responses[cmdPattern] = response
}

// SetError は指定コマンドへのエラーを設定する
func (m *MockExecutor) SetError(cmdPattern string, err error) {
	m.errors[cmdPattern] = err
}

// Execute はモックされたコマンド実行
func (m *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(args, " ")

	for pattern, err := range m.errors {
		if strings.Contains(cmd, pattern) {
			return nil, err
		}
	}

	for pattern, response := range m.responses {
		if strings.Contains(cmd, pattern) {
			return response, nil
		}
	}

	return nil, errors.New("no mock response for: " + cmd)
}

func TestNewClient(t *testing.T) {
	mock := NewMockExecutor()
	mock.SetResponse("gh api user", []byte("testuser\n"))

	client, err := NewClient(WithExecutor(mock))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	if client.Username() != "testuser" {
		t.Errorf("Username: got %q, want %q", client.Username(), "testuser")
	}
}

func TestNewClient_EmptyUsername(t *testing.T) {
	mock := NewMockExecutor()
	mock.SetResponse("gh api user", []byte("\n"))

	_, err := NewClient(WithExecutor(mock))
	if err == nil {
		t.Fatal("NewClient() should fail with empty username")
	}
	if !strings.Contains(err.Error(), "empty username") {
		t.Errorf("Error should mention empty username, got: %v", err)
	}
}

func TestNewClient_APIError(t *testing.T) {
	mock := NewMockExecutor()
	mock.SetError("gh api user", errors.New("API error"))

	_, err := NewClient(WithExecutor(mock))
	if err == nil {
		t.Fatal("NewClient() should fail on API error")
	}
}

func TestClient_SearchPRs(t *testing.T) {
	graphQLResponse := `{
		"data": {
			"search": {
				"pageInfo": {
					"hasNextPage": false,
					"endCursor": ""
				},
				"nodes": [
					{
						"title": "Test PR",
						"url": "https://github.com/test/repo/pull/1",
						"state": "OPEN",
						"isDraft": false,
						"createdAt": "2025-01-10T10:00:00Z",
						"mergedAt": "",
						"updatedAt": "2025-01-10T12:00:00Z",
						"additions": 100,
						"deletions": 50,
						"changedFiles": 5,
						"comments": {"totalCount": 3},
						"repository": {"name": "test-repo"}
					}
				]
			}
		}
	}`

	mock := NewMockExecutor()
	mock.SetResponse("graphql", []byte(graphQLResponse))

	client := &Client{username: "testuser", executor: mock}
	prs, err := client.SearchPRs("test-org", "is:pr", "created:2025-01-01..2025-01-31")
	if err != nil {
		t.Fatalf("SearchPRs() failed: %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("len(prs): got %d, want 1", len(prs))
	}

	pr := prs[0]
	if pr.Title != "Test PR" {
		t.Errorf("Title: got %q, want %q", pr.Title, "Test PR")
	}
	if pr.Repository != "test-repo" {
		t.Errorf("Repository: got %q, want %q", pr.Repository, "test-repo")
	}
	if pr.State != "open" {
		t.Errorf("State: got %q, want %q", pr.State, "open")
	}
	if pr.Additions != 100 {
		t.Errorf("Additions: got %d, want %d", pr.Additions, 100)
	}
	if pr.Deletions != 50 {
		t.Errorf("Deletions: got %d, want %d", pr.Deletions, 50)
	}
	if pr.Comments != 3 {
		t.Errorf("Comments: got %d, want %d", pr.Comments, 3)
	}
}

func TestClient_SearchPRs_Pagination(t *testing.T) {
	firstResponse := `{
		"data": {
			"search": {
				"pageInfo": {
					"hasNextPage": true,
					"endCursor": "cursor123"
				},
				"nodes": [
					{
						"title": "PR 1",
						"url": "https://github.com/test/repo/pull/1",
						"state": "OPEN",
						"isDraft": false,
						"createdAt": "2025-01-10T10:00:00Z",
						"updatedAt": "2025-01-10T12:00:00Z",
						"additions": 10,
						"deletions": 5,
						"changedFiles": 1,
						"comments": {"totalCount": 1},
						"repository": {"name": "test-repo"}
					}
				]
			}
		}
	}`
	secondResponse := `{
		"data": {
			"search": {
				"pageInfo": {
					"hasNextPage": false,
					"endCursor": ""
				},
				"nodes": [
					{
						"title": "PR 2",
						"url": "https://github.com/test/repo/pull/2",
						"state": "MERGED",
						"isDraft": false,
						"createdAt": "2025-01-11T10:00:00Z",
						"mergedAt": "2025-01-12T10:00:00Z",
						"updatedAt": "2025-01-12T12:00:00Z",
						"additions": 20,
						"deletions": 10,
						"changedFiles": 2,
						"comments": {"totalCount": 2},
						"repository": {"name": "test-repo"}
					}
				]
			}
		}
	}`

	callCount := 0
	mock := &PaginationMockExecutor{
		responses: [][]byte{[]byte(firstResponse), []byte(secondResponse)},
		callCount: &callCount,
	}

	client := &Client{username: "testuser", executor: mock}
	prs, err := client.SearchPRs("test-org", "is:pr", "created:2025-01-01..2025-01-31")
	if err != nil {
		t.Fatalf("SearchPRs() failed: %v", err)
	}

	if len(prs) != 2 {
		t.Fatalf("len(prs): got %d, want 2", len(prs))
	}
	if prs[0].Title != "PR 1" {
		t.Errorf("prs[0].Title: got %q, want %q", prs[0].Title, "PR 1")
	}
	if prs[1].Title != "PR 2" {
		t.Errorf("prs[1].Title: got %q, want %q", prs[1].Title, "PR 2")
	}
}

// PaginationMockExecutor はページネーションテスト用のモック
type PaginationMockExecutor struct {
	responses [][]byte
	callCount *int
}

func (m *PaginationMockExecutor) Execute(name string, args ...string) ([]byte, error) {
	idx := *m.callCount
	if idx >= len(m.responses) {
		return nil, errors.New("no more mock responses")
	}
	*m.callCount++
	return m.responses[idx], nil
}
