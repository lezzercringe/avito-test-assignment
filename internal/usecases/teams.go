package usecases

import (
	"context"
	"fmt"

	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

type TeamMemberView struct {
	ID     string
	Name   string
	Active bool
}

type TeamView struct {
	Name    string
	Members []TeamMemberView
}

type TeamService interface {
	GetTeam(ctx context.Context, name string) (*TeamView, error)
	AddTeam(ctx context.Context, req TeamView) (*TeamView, error)
}

var _ TeamService = &TeamServiceImpl{}

type TeamServiceImpl struct {
	teamRepo  teamRepository
	userRepo  userRepository
	txManager TxManager
}

func (s *TeamServiceImpl) AddTeam(ctx context.Context, req TeamView) (*TeamView, error) {
	var memberIDs []string
	for _, member := range req.Members {
		memberIDs = append(memberIDs, member.ID)
	}

	team, err := teams.New(req.Name, memberIDs)
	if err != nil {
		return nil, err
	}

	ctx, tx, err := s.txManager.WithTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting tx: %w", err)
	}

	defer tx.Rollback(ctx)

	usersToUpsert := make([]*users.User, 0, len(req.Members))
	for _, member := range req.Members {
		user, err := users.New(member.ID, member.Name, member.Active)
		if err != nil {
			return nil, err
		}
		usersToUpsert = append(usersToUpsert, user)
	}

	if err := s.userRepo.SaveMany(ctx, usersToUpsert...); err != nil {
		return nil, err
	}

	if err := s.teamRepo.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("saving team: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commiting tx: %w", err)
	}

	return &req, nil
}

func (s *TeamServiceImpl) GetTeam(ctx context.Context, name string) (*TeamView, error) {
	team, err := s.teamRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	members, err := s.userRepo.GetMany(ctx, team.MemberIDs...)
	if err != nil {
		return nil, err
	}

	views := make([]TeamMemberView, 0, len(members))
	for _, m := range members {
		views = append(views, TeamMemberView{
			ID:     m.ID,
			Name:   m.Name,
			Active: m.Active,
		})
	}

	return &TeamView{
		Name:    name,
		Members: views,
	}, nil
}

func NewTeamService(txManager TxManager, teamRepo teamRepository, userRepo userRepository) *TeamServiceImpl {
	return &TeamServiceImpl{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		txManager: txManager,
	}
}
