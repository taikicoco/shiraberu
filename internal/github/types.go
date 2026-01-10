package github

import "time"

type PullRequest struct {
	Title        string
	URL          string
	Repository   string
	State        string
	IsDraft      bool
	CreatedAt    time.Time
	MergedAt     *time.Time
	UpdatedAt    time.Time
	Additions    int
	Deletions    int
	ChangedFiles int
	Comments     int
}
