package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/taikicoco/shiraberu/internal/errors"
)

// searchResultLimit はGitHubの検索で取得できる件数の上限
const searchResultLimit = 1000

// Review は指定ユーザーが提出したレビュー
type Review struct {
	State       string
	SubmittedAt time.Time
	URL         string
	Body        string
}

// Comment は指定ユーザーが書いた会話のコメント
type Comment struct {
	CreatedAt time.Time
	URL       string
	Body      string
}

// PullRequestDetail はJSON出力に使う詳しいPRの情報。
// Reviews と Comments は検索時に指定したユーザーのものだけを持つ
type PullRequestDetail struct {
	Number       int
	Title        string
	URL          string
	Repository   string
	State        string
	IsDraft      bool
	Author       string
	BaseRef      string
	Body         string
	CreatedAt    time.Time
	MergedAt     *time.Time
	ClosedAt     *time.Time
	UpdatedAt    time.Time
	Additions    int
	Deletions    int
	ChangedFiles int
	CommentCount int
	Reviews      []Review
	Comments     []Comment
}

// detailSearchQuery はPRの詳細と、指定ユーザーのレビュー・コメントを取得するクエリ。
// comments には作成者での絞り込みがないので、受け取ったあとで絞る
const detailSearchQuery = `
query($q: String!, $login: String!, $cursor: String) {
  search(query: $q, type: ISSUE, first: 100, after: $cursor) {
    issueCount
    pageInfo {
      hasNextPage
      endCursor
    }
    nodes {
      ... on PullRequest {
        number
        title
        url
        state
        isDraft
        body
        baseRefName
        createdAt
        mergedAt
        closedAt
        updatedAt
        additions
        deletions
        changedFiles
        author {
          login
        }
        repository {
          nameWithOwner
        }
        reviews(first: 100, author: $login) {
          nodes {
            state
            submittedAt
            url
            body
          }
        }
        comments(last: 100) {
          totalCount
          nodes {
            createdAt
            url
            body
            author {
              login
            }
          }
        }
      }
    }
  }
}
`

type detailResponse struct {
	Data struct {
		Search struct {
			IssueCount int `json:"issueCount"`
			PageInfo   struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []detailNode `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

type actorNode struct {
	Login string `json:"login"`
}

type detailNode struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	State        string    `json:"state"`
	IsDraft      bool      `json:"isDraft"`
	Body         string    `json:"body"`
	BaseRefName  string    `json:"baseRefName"`
	CreatedAt    string    `json:"createdAt"`
	MergedAt     string    `json:"mergedAt"`
	ClosedAt     string    `json:"closedAt"`
	UpdatedAt    string    `json:"updatedAt"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	ChangedFiles int       `json:"changedFiles"`
	Author       actorNode `json:"author"`
	Repository   struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Reviews struct {
		Nodes []struct {
			State       string `json:"state"`
			SubmittedAt string `json:"submittedAt"`
			URL         string `json:"url"`
			Body        string `json:"body"`
		} `json:"nodes"`
	} `json:"reviews"`
	Comments struct {
		TotalCount int `json:"totalCount"`
		Nodes      []struct {
			CreatedAt string    `json:"createdAt"`
			URL       string    `json:"url"`
			Body      string    `json:"body"`
			Author    actorNode `json:"author"`
		} `json:"nodes"`
	} `json:"comments"`
}

// SearchPRDetails はPRを検索し、詳細と username のレビュー・コメントを返す。
// 検索結果がGitHubの上限を超えると、黙って欠けないようにエラーを返す
func (c *Client) SearchPRDetails(org, query, dateFilter, username string) ([]PullRequestDetail, error) {
	q := fmt.Sprintf("%s org:%s %s", query, org, dateFilter)

	var details []PullRequestDetail
	var cursor string

	for {
		args := []string{"api", "graphql",
			"-f", "q=" + q,
			"-f", "login=" + username,
			"-f", "query=" + detailSearchQuery,
		}
		if cursor != "" {
			args = append(args, "-f", "cursor="+cursor)
		}

		out, err := c.executor.Execute("gh", args...)
		if err != nil {
			return nil, fmt.Errorf("gh api graphql failed: %w", err)
		}

		var resp detailResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if count := resp.Data.Search.IssueCount; count > searchResultLimit {
			return nil, fmt.Errorf("%w: %d results for %q", apperrors.ErrTooManyResults, count, q)
		}

		for _, node := range resp.Data.Search.Nodes {
			if node.URL != "" {
				details = append(details, toDetail(node, username))
			}
		}

		if !resp.Data.Search.PageInfo.HasNextPage {
			return details, nil
		}
		cursor = resp.Data.Search.PageInfo.EndCursor
	}
}

func toDetail(node detailNode, username string) PullRequestDetail {
	detail := PullRequestDetail{
		Number:       node.Number,
		Title:        node.Title,
		URL:          node.URL,
		Repository:   node.Repository.NameWithOwner,
		State:        normalizeState(node.State),
		IsDraft:      node.IsDraft,
		Author:       node.Author.Login,
		BaseRef:      node.BaseRefName,
		Body:         node.Body,
		CreatedAt:    parseTime(node.CreatedAt),
		MergedAt:     parseOptionalTime(node.MergedAt),
		ClosedAt:     parseOptionalTime(node.ClosedAt),
		UpdatedAt:    parseTime(node.UpdatedAt),
		Additions:    node.Additions,
		Deletions:    node.Deletions,
		ChangedFiles: node.ChangedFiles,
		CommentCount: node.Comments.TotalCount,
	}

	for _, review := range node.Reviews.Nodes {
		// 提出前のレビューには submittedAt がない
		if submittedAt := parseOptionalTime(review.SubmittedAt); submittedAt != nil {
			detail.Reviews = append(detail.Reviews, Review{
				State:       review.State,
				SubmittedAt: *submittedAt,
				URL:         review.URL,
				Body:        review.Body,
			})
		}
	}

	for _, comment := range node.Comments.Nodes {
		if strings.EqualFold(comment.Author.Login, username) {
			detail.Comments = append(detail.Comments, Comment{
				CreatedAt: parseTime(comment.CreatedAt),
				URL:       comment.URL,
				Body:      comment.Body,
			})
		}
	}

	return detail
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func parseOptionalTime(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}
