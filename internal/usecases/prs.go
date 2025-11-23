package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
)

type PullRequestService interface {
	Create(ctx context.Context, req CreateRequest) (*prs.PullRequestView, error)
	Merge(ctx context.Context, id string) (*prs.PullRequestView, error)
	ReassignReviewer(ctx context.Context, req ReassignReviewerRequest) (*ReassignReviewerResult, error)
}

type CreateRequest struct {
	ID       string
	AuthorID string
	Name     string
}

type ReassignReviewerRequest struct {
	PullRequestID    string
	UserIDToReassign string
}

type ReassignReviewerResult struct {
	PR           *prs.PullRequestView
	ReplacedByID string
}

type PickReviewersRequest struct {
	UserIDsToExclude []string
	WantCount        int
	Team             teams.Team
}

type ReviewerPicker interface {
	PickReviewersFromTeam(ctx context.Context, req PickReviewersRequest) ([]string, error)
}

var _ PullRequestService = &PullRequestServiceImpl{}

type PullRequestServiceImpl struct {
	rpicker  ReviewerPicker
	prRepo   prRepository
	teamRepo teamRepository
}

func (m *PullRequestServiceImpl) Create(ctx context.Context, req CreateRequest) (*prs.PullRequestView, error) {
	team, err := m.teamRepo.GetByMemberID(ctx, req.AuthorID)
	if err != nil {
		return nil, err
	}

	pr := prs.PullRequest{
		ID:               req.ID,
		Name:             req.Name,
		Status:           prs.StatusOpen,
		OriginalTeamName: team.Name,
		AuthorID:         req.AuthorID,
	}

	pickedIDs, err := m.rpicker.PickReviewersFromTeam(ctx, PickReviewersRequest{
		UserIDsToExclude: []string{req.AuthorID},
		Team:             *team,
		WantCount:        2,
	})
	if err != nil && !errors.Is(err, errorsx.ErrNoCandidate) {
		return nil, fmt.Errorf("picking reviewer: %w", err)
	}
	if !errors.Is(err, errorsx.ErrNoCandidate) {
		for _, id := range pickedIDs {
			if err := pr.AssignReviewer(id); err != nil {
				return nil, fmt.Errorf("assigning reviewer: %w", err)
			}
		}

	}

	if err := m.prRepo.Save(ctx, &pr); err != nil {
		return nil, fmt.Errorf("saving pr: %w", err)
	}

	return pr.ToView(), nil
}

func (m *PullRequestServiceImpl) ReassignReviewer(ctx context.Context, req ReassignReviewerRequest) (*ReassignReviewerResult, error) {
	pr, err := m.prRepo.GetByID(ctx, req.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("retreiving pull request with specified id: %w", err)
	}

	if err := pr.UnassignReviewer(req.UserIDToReassign); err != nil {
		return nil, fmt.Errorf("unassigning original reviewer: %w", err)
	}

	team, err := m.teamRepo.GetByName(ctx, pr.OriginalTeamName)
	if err != nil {
		return nil, fmt.Errorf("getting original pr team: %w", err)
	}

	pickedIDs, err := m.rpicker.PickReviewersFromTeam(ctx, PickReviewersRequest{
		UserIDsToExclude: []string{req.UserIDToReassign},
		WantCount:        1,
		Team:             *team,
	})
	if err != nil {
		return nil, fmt.Errorf("picking new reviewer: %w", err)
	}

	if err := pr.AssignReviewer(pickedIDs[0]); err != nil {
		return nil, fmt.Errorf("assigning new reviewer: %w", err)
	}

	if err := m.prRepo.Save(ctx, pr); err != nil {
		return nil, fmt.Errorf("saving pr: %w", err)
	}

	return &ReassignReviewerResult{
		PR:           pr.ToView(),
		ReplacedByID: pickedIDs[0],
	}, nil
}

func (m *PullRequestServiceImpl) Merge(ctx context.Context, id string) (*prs.PullRequestView, error) {
	pr, err := m.prRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("retrieving pr by id: %w", err)
	}

	pr.Merge()

	if err := m.prRepo.Save(ctx, pr); err != nil {
		return nil, fmt.Errorf("saving pr: %w", err)
	}

	return pr.ToView(), nil
}

func NewPullRequestService(rpicker ReviewerPicker, prRepo prRepository, teamRepo teamRepository) *PullRequestServiceImpl {
	return &PullRequestServiceImpl{
		rpicker:  rpicker,
		prRepo:   prRepo,
		teamRepo: teamRepo,
	}
}
