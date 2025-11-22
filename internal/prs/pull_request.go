package prs

import (
	"slices"
	"time"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
)

const maxReviewers = 2

type Status string

const (
	StatusOpen   Status = "OPEN"
	StatusMerged Status = "MERGED"
)

type PullRequest struct {
	ID               string
	Name             string
	Status           Status
	OriginalTeamName string
	AuthorID         string
	ReviewerIDs      []string
	MergedAt         time.Time
}

type PullRequestView struct {
	ID          string
	Name        string
	AuthorID    string
	Status      string
	ReviewerIDs []string
	MergedAt    time.Time
}

func (p *PullRequest) ToView() *PullRequestView {
	return &PullRequestView{
		ID:          p.ID,
		Name:        p.Name,
		AuthorID:    p.AuthorID,
		Status:      string(p.Status),
		ReviewerIDs: p.ReviewerIDs,
		MergedAt:    p.MergedAt,
	}
}

func (p *PullRequest) Merge() {
	if p.Status == StatusMerged {
		return
	}
	p.Status = StatusMerged
	p.MergedAt = time.Now()
}

func (p *PullRequest) AssignReviewer(userID string) error {
	if p.Status == StatusMerged {
		return errorsx.ErrModifyMergedPR
	}

	if len(p.ReviewerIDs) >= maxReviewers {
		return errorsx.ErrToManyReviewers
	}

	if slices.Contains(p.ReviewerIDs, userID) {
		return errorsx.ErrCandidateIsAlreadyReviewer
	}

	p.ReviewerIDs = append(p.ReviewerIDs, userID)
	return nil
}

func (p *PullRequest) UnassignReviewer(userID string) error {
	if p.Status == StatusMerged {
		return errorsx.ErrModifyMergedPR
	}

	ix := slices.Index(p.ReviewerIDs, userID)
	if ix == -1 {
		return errorsx.ErrNotPreviouslyAssigned
	}

	p.ReviewerIDs = append(p.ReviewerIDs[:ix], p.ReviewerIDs[ix+1:]...)
	return nil
}
