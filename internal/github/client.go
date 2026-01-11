package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apperrors "github.com/taikicoco/shiraberu/internal/errors"
)

// CommandExecutor はコマンド実行を抽象化するインターフェース
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
}

// DefaultExecutor は実際のコマンドを実行するデフォルト実装
type DefaultExecutor struct{}

// Execute はシェルコマンドを実行し、標準出力を返す
func (e *DefaultExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command failed: %w\n%s", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return out, nil
}

type Client struct {
	username string
	executor CommandExecutor
}

// ClientOption はClientの設定オプション
type ClientOption func(*Client)

// WithExecutor はCommandExecutorを設定するオプション
func WithExecutor(e CommandExecutor) ClientOption {
	return func(c *Client) {
		c.executor = e
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		executor: &DefaultExecutor{},
	}
	for _, opt := range opts {
		opt(c)
	}

	username, err := c.getUsername()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub username: %w", err)
	}
	c.username = username
	return c, nil
}

func (c *Client) Username() string {
	return c.username
}

func (c *Client) getUsername() (string, error) {
	out, err := c.executor.Execute("gh", "api", "user", "--jq", ".login")
	if err != nil {
		return "", err
	}
	username := strings.TrimSpace(string(out))
	if username == "" {
		return "", apperrors.ErrEmptyUsername
	}
	return username, nil
}

// searchQuery is the GraphQL query for searching PRs.
// Uses 100 results per page for pagination.
const searchQuery = `
query($q: String!, $cursor: String) {
  search(query: $q, type: ISSUE, first: 100, after: $cursor) {
    pageInfo {
      hasNextPage
      endCursor
    }
    nodes {
      ... on PullRequest {
        title
        url
        state
        isDraft
        createdAt
        mergedAt
        updatedAt
        additions
        deletions
        changedFiles
        comments {
          totalCount
        }
        repository {
          name
        }
      }
    }
  }
}
`

type graphQLResponse struct {
	Data struct {
		Search struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []struct {
				Title        string `json:"title"`
				URL          string `json:"url"`
				State        string `json:"state"`
				IsDraft      bool   `json:"isDraft"`
				CreatedAt    string `json:"createdAt"`
				MergedAt     string `json:"mergedAt"`
				UpdatedAt    string `json:"updatedAt"`
				Additions    int    `json:"additions"`
				Deletions    int    `json:"deletions"`
				ChangedFiles int    `json:"changedFiles"`
				Comments     struct {
					TotalCount int `json:"totalCount"`
				} `json:"comments"`
				Repository struct {
					Name string `json:"name"`
				} `json:"repository"`
			} `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

func (c *Client) SearchPRs(org string, query string, dateFilter string) ([]PullRequest, error) {
	q := fmt.Sprintf("%s org:%s %s", query, org, dateFilter)

	var allPRs []PullRequest
	var cursor string

	for {
		args := []string{"api", "graphql",
			"-f", "q=" + q,
			"-f", "query=" + searchQuery,
		}
		if cursor != "" {
			args = append(args, "-f", "cursor="+cursor)
		}

		out, err := c.executor.Execute("gh", args...)
		if err != nil {
			return nil, fmt.Errorf("gh api graphql failed: %w", err)
		}

		var resp graphQLResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, node := range resp.Data.Search.Nodes {
			if node.URL == "" {
				continue
			}

			pr := PullRequest{
				Title:        node.Title,
				URL:          node.URL,
				Repository:   node.Repository.Name,
				State:        normalizeState(node.State),
				IsDraft:      node.IsDraft,
				Additions:    node.Additions,
				Deletions:    node.Deletions,
				ChangedFiles: node.ChangedFiles,
				Comments:     node.Comments.TotalCount,
			}

			if t, err := time.Parse(time.RFC3339, node.CreatedAt); err == nil {
				pr.CreatedAt = t
			}
			if t, err := time.Parse(time.RFC3339, node.UpdatedAt); err == nil {
				pr.UpdatedAt = t
			}
			if node.MergedAt != "" {
				if t, err := time.Parse(time.RFC3339, node.MergedAt); err == nil {
					pr.MergedAt = &t
				}
			}

			allPRs = append(allPRs, pr)
		}

		if !resp.Data.Search.PageInfo.HasNextPage {
			break
		}
		cursor = resp.Data.Search.PageInfo.EndCursor
	}

	return allPRs, nil
}

func normalizeState(state string) string {
	switch state {
	case "MERGED":
		return "merged"
	case "OPEN":
		return "open"
	case "CLOSED":
		return "closed"
	default:
		return state
	}
}
