package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

type UserWithTeamView struct {
	ID, Name string
	IsActive bool
	TeamName string
}

type UserView struct {
	ID, Name string
	IsActive bool
}

type ReviewedPullRequestView struct {
	ID, Name string
	Status   prs.Status
	AuthorID string
}

type UserService interface {
	SetIsActive(ctx context.Context, id string, isActive bool) (*UserWithTeamView, error)
	GetReview(ctx context.Context, id string) ([]*ReviewedPullRequestView, error)
	Deactivate(ctx context.Context, ids ...string) ([]*UserView, error)
}

var _ UserService = &UserServiceImpl{}

type UserServiceImpl struct {
	teamRepo  teamRepository
	txManager TxManager
	userRepo  userRepository
	prRepo    prRepository
	rpicker   ReviewerPicker
}

func userIntoView(user *users.User) *UserView {
	return &UserView{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: user.Active,
	}
}

func usersIntoViews(xs []*users.User) []*UserView {
	views := make([]*UserView, len(xs))
	for i, u := range xs {
		views[i] = userIntoView(u)

	}
	return views
}

func (s *UserServiceImpl) GetReview(ctx context.Context, id string) ([]*ReviewedPullRequestView, error) {
	prs, err := s.prRepo.GetManyByReviewerID(ctx, id)
	if err != nil {
		return nil, err
	}

	views := make([]*ReviewedPullRequestView, 0, len(prs))
	for _, pr := range prs {
		views = append(views, &ReviewedPullRequestView{
			ID:       pr.ID,
			Name:     pr.Name,
			Status:   pr.Status,
			AuthorID: pr.AuthorID,
		})
	}

	return views, nil
}

func (s *UserServiceImpl) SetIsActive(ctx context.Context, id string, isActive bool) (*UserWithTeamView, error) {
	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Active = isActive

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	team, err := s.teamRepo.GetByMemberID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}

	return &UserWithTeamView{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: isActive,
		TeamName: team.Name,
	}, nil
}

func collectUniqueTeamNames(pullRequests []*PRWithMatchedReviewers) []string {
	set := make(map[string]struct{})
	names := make([]string, 0, len(pullRequests))

	for _, entry := range pullRequests {
		if _, ok := set[entry.PR.OriginalTeamName]; !ok {
			set[entry.PR.OriginalTeamName] = struct{}{}
			names = append(names, entry.PR.OriginalTeamName)
		}
	}

	return names
}

// reassignAllPRsWithReviewers finds all the PR's assigned to set of users, removes those users from their reviewers list.
//
// WARN: method must be called within a transaction.
func (s *UserServiceImpl) reassignAllPRsWithReviewers(ctx context.Context, reviewerIDs ...string) error {
	result, err := s.prRepo.GetAllUnmergedWithAnyOfReviewers(ctx, reviewerIDs...)
	if err != nil {
		return fmt.Errorf("retrieving prs with reviewers: %w", err)
	}

	if len(result) == 0 {
		return nil
	}

	teams, err := s.teamRepo.GetManyByNames(ctx, collectUniqueTeamNames(result)...)
	if err != nil {
		return fmt.Errorf("retrieving teams: %w", err)
	}

	pullRequests := make([]*prs.PullRequest, len(result))
	for i, entry := range result {
		for _, matchedID := range entry.MatchedReviewerIDs {
			if err := entry.PR.UnassignReviewer(matchedID); err != nil {
				return fmt.Errorf("unassigning reviewer: %w", err)
			}
		}

		newReviewers, err := s.rpicker.PickReviewersFromTeam(ctx, PickReviewersRequest{
			UserIDsToExclude: append(slices.Clone(entry.MatchedReviewerIDs), entry.PR.AuthorID),
			WantCount:        len(entry.MatchedReviewerIDs),
			Team:             teams[entry.PR.OriginalTeamName],
		})
		if err != nil && !errors.Is(err, errorsx.ErrNoCandidate) {
			return fmt.Errorf("picking new reviewers: %w", err)
		}

		for _, reviewer := range newReviewers {
			if err := entry.PR.AssignReviewer(reviewer); err != nil {
				return fmt.Errorf("assigning new reviewers: %w", err)
			}
		}

		pullRequests[i] = entry.PR
	}

	return s.prRepo.SaveMany(ctx, pullRequests...)
}

func (s *UserServiceImpl) Deactivate(ctx context.Context, idsToDeactivate ...string) ([]*UserView, error) {
	ctx, txHandle, err := s.txManager.WithTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting tx: %w", err)
	}

	defer txHandle.Rollback(ctx)

	if err := s.reassignAllPRsWithReviewers(ctx, idsToDeactivate...); err != nil {
		return nil, fmt.Errorf("reassigning reviewed prs: %w", err)
	}

	users, err := s.userRepo.GetMany(ctx, idsToDeactivate...)
	if err != nil {
		return nil, fmt.Errorf("retrieving users: %w", err)
	}

	for _, u := range users {
		u.Active = false
	}

	if err := s.userRepo.SaveMany(ctx, users...); err != nil {
		return nil, fmt.Errorf("saving users: %w", err)
	}

	if err := txHandle.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commiting tx: %w", err)
	}

	return usersIntoViews(users), nil
}

func NewUserService(
	txManager TxManager,
	userRepo userRepository,
	teamRepo teamRepository,
	prRepo prRepository,
	rpicker ReviewerPicker,
) *UserServiceImpl {
	return &UserServiceImpl{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		prRepo:    prRepo,
		txManager: txManager,
		rpicker:   rpicker,
	}
}
