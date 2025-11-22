package usecases

import (
	"context"
	"fmt"

	"github.com/lezzercringe/avito-test-assignment/internal/prs"
)

type UserView struct {
	ID, Name string
	IsActive bool
	TeamName string
}

type ReviewedPullRequestView struct {
	ID, Name string
	Status   prs.Status
	AuthorID string
}

type UserService interface {
	SetIsActive(ctx context.Context, id string, isActive bool) (*UserView, error)
	GetReview(ctx context.Context, id string) ([]ReviewedPullRequestView, error)
}

var _ UserService = &UserServiceImpl{}

type UserServiceImpl struct {
	teamRepo teamRepository
	userRepo userRepository
	prRepo   prRepository
}

func (s *UserServiceImpl) GetReview(ctx context.Context, id string) ([]ReviewedPullRequestView, error) {
	prs, err := s.prRepo.GetManyByReviewerID(ctx, id)
	if err != nil {
		return nil, err
	}

	views := make([]ReviewedPullRequestView, 0, len(prs))
	for _, pr := range prs {
		views = append(views, ReviewedPullRequestView{
			ID:       pr.ID,
			Name:     pr.Name,
			Status:   pr.Status,
			AuthorID: pr.AuthorID,
		})
	}

	return views, nil
}

func (s *UserServiceImpl) SetIsActive(ctx context.Context, id string, isActive bool) (*UserView, error) {
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

	return &UserView{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: isActive,
		TeamName: team.Name,
	}, nil
}

func NewUserService(userRepo userRepository, teamRepo teamRepository, prRepo prRepository) *UserServiceImpl {
	return &UserServiceImpl{
		teamRepo: teamRepo,
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}
