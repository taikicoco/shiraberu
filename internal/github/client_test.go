package github

import (
	"os"
	"os/exec"
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

// fakeExecCommand returns a helper process command for mocking
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess is a helper process for mocking exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(1)
	}

	// Mock responses based on command
	switch {
	case args[0] == "gh" && len(args) > 1 && args[1] == "api":
		if len(args) > 2 && args[2] == "user" {
			// Mock gh api user response
			_, _ = os.Stdout.WriteString("testuser\n")
			os.Exit(0)
		}
		if len(args) > 2 && args[2] == "graphql" {
			// Mock gh api graphql response
			response := `{
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
			_, _ = os.Stdout.WriteString(response)
			os.Exit(0)
		}
	}
	os.Exit(1)
}

func TestNewClient(t *testing.T) {
	// Save and restore original execCommand
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = fakeExecCommand

	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	if client.Username() != "testuser" {
		t.Errorf("Username: got %q, want %q", client.Username(), "testuser")
	}
}

func TestClient_SearchPRs(t *testing.T) {
	// Save and restore original execCommand
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = fakeExecCommand

	client := &Client{username: "testuser"}
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

func TestGetUsername(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = fakeExecCommand

	username, err := getUsername()
	if err != nil {
		t.Fatalf("getUsername() failed: %v", err)
	}
	if username != "testuser" {
		t.Errorf("username: got %q, want %q", username, "testuser")
	}
}
