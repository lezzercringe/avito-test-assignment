package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
	"github.com/lezzercringe/avito-test-assignment/mocks"
)

func setupTest(t *testing.T) (*usecases.TeamServiceImpl, *mocks.MockteamRepository, *mocks.MockuserRepository, *mocks.MockTxManager, *mocks.MockTxHandle) {
	ctrl := gomock.NewController(t)

	teamRepo := mocks.NewMockteamRepository(ctrl)
	userRepo := mocks.NewMockuserRepository(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	txHandle := mocks.NewMockTxHandle(ctrl)

	service := usecases.NewTeamService(txManager, teamRepo, userRepo)

	return service, teamRepo, userRepo, txManager, txHandle
}

func assertTeamView(t *testing.T, result *usecases.TeamView, expectedName string, expectedMembers []usecases.TeamMemberView) {
	assert.Equal(t, expectedName, result.Name)
	assert.Len(t, result.Members, len(expectedMembers))
	for i, expected := range expectedMembers {
		assert.Equal(t, expected.ID, result.Members[i].ID)
		assert.Equal(t, expected.Name, result.Members[i].Name)
		assert.Equal(t, expected.Active, result.Members[i].Active)
	}
}

func TestTeamService_AddTeam_Ok(t *testing.T) {
	service, teamRepo, userRepo, txManager, txHandle := setupTest(t)
	ctx := context.Background()

	t.Run("valid team with members", func(t *testing.T) {
		req := usecases.TeamView{
			Name: "test-team",
			Members: []usecases.TeamMemberView{
				{ID: "user1", Name: "User One", Active: true},
				{ID: "user2", Name: "User Two", Active: false},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx, gomock.Any()).Return(nil)
		teamRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.AddTeam(ctx, req)

		require.NoError(t, err)
		assertTeamView(t, result, req.Name, req.Members)
	})

	t.Run("valid team with no members", func(t *testing.T) {
		req := usecases.TeamView{
			Name:    "empty-team",
			Members: []usecases.TeamMemberView{},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx).Return(nil)
		teamRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.AddTeam(ctx, req)

		require.NoError(t, err)
		assertTeamView(t, result, req.Name, req.Members)
	})

	t.Run("valid team with single member", func(t *testing.T) {
		req := usecases.TeamView{
			Name: "single-member-team",
			Members: []usecases.TeamMemberView{
				{ID: "user1", Name: "Single User", Active: true},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx, gomock.Any()).Return(nil)
		teamRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.AddTeam(ctx, req)

		require.NoError(t, err)
		assertTeamView(t, result, req.Name, req.Members)
	})
}

func TestTeamService_AddTeam_Error(t *testing.T) {
	service, teamRepo, userRepo, txManager, txHandle := setupTest(t)
	ctx := context.Background()

	t.Run("tx manager error", func(t *testing.T) {
		req := usecases.TeamView{Name: "test-team"}

		txManager.EXPECT().WithTx(ctx).Return(nil, nil, errors.New("tx error"))

		result, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "starting tx")
	})

	t.Run("user repo save error", func(t *testing.T) {
		req := usecases.TeamView{
			Name: "test-team",
			Members: []usecases.TeamMemberView{
				{ID: "user1", Name: "User One", Active: true},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx, gomock.Any()).Return(errors.New("save error"))

		result, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "save error")
	})

	t.Run("team repo save error", func(t *testing.T) {
		req := usecases.TeamView{
			Name: "test-team",
			Members: []usecases.TeamMemberView{
				{ID: "user1", Name: "User One", Active: true},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx, gomock.Any()).Return(nil)
		teamRepo.EXPECT().Save(ctx, gomock.Any()).Return(errors.New("team save error"))

		result, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "saving team")
	})

	t.Run("commit error", func(t *testing.T) {
		req := usecases.TeamView{
			Name: "test-team",
			Members: []usecases.TeamMemberView{
				{ID: "user1", Name: "User One", Active: true},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		userRepo.EXPECT().SaveMany(ctx, gomock.Any()).Return(nil)
		teamRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(errors.New("commit error"))

		result, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "commiting tx")
	})
}

func TestTeamService_GetTeam_Ok(t *testing.T) {
	service, teamRepo, userRepo, _, _ := setupTest(t)
	ctx := context.Background()

	t.Run("existing team with members", func(t *testing.T) {
		teamName := "test-team"
		memberIDs := []string{"user1", "user2"}

		teamRepo.EXPECT().GetByName(ctx, teamName).Return(&teams.Team{
			Name:      teamName,
			MemberIDs: memberIDs,
		}, nil)

		userRepo.EXPECT().GetMany(ctx, "user1", "user2").Return([]*users.User{
			{ID: "user1", Name: "User One", Active: true},
			{ID: "user2", Name: "User Two", Active: false},
		}, nil)

		expectedMembers := []usecases.TeamMemberView{
			{ID: "user1", Name: "User One", Active: true},
			{ID: "user2", Name: "User Two", Active: false},
		}

		result, err := service.GetTeam(ctx, teamName)

		require.NoError(t, err)
		assertTeamView(t, result, teamName, expectedMembers)
	})

	t.Run("existing team with no members", func(t *testing.T) {
		teamName := "empty-team"

		teamRepo.EXPECT().GetByName(ctx, teamName).Return(&teams.Team{
			Name:      teamName,
			MemberIDs: []string{},
		}, nil)

		userRepo.EXPECT().GetMany(ctx).Return([]*users.User{}, nil)

		expectedMembers := []usecases.TeamMemberView{}

		result, err := service.GetTeam(ctx, teamName)

		require.NoError(t, err)
		assertTeamView(t, result, teamName, expectedMembers)
	})

	t.Run("existing team with mixed active/inactive members", func(t *testing.T) {
		teamName := "mixed-team"
		memberIDs := []string{"user1", "user2", "user3"}

		teamRepo.EXPECT().GetByName(ctx, teamName).Return(&teams.Team{
			Name:      teamName,
			MemberIDs: memberIDs,
		}, nil)

		userRepo.EXPECT().GetMany(ctx, "user1", "user2", "user3").Return([]*users.User{
			{ID: "user1", Name: "Active User", Active: true},
			{ID: "user2", Name: "Inactive User", Active: false},
			{ID: "user3", Name: "Another Active", Active: true},
		}, nil)

		expectedMembers := []usecases.TeamMemberView{
			{ID: "user1", Name: "Active User", Active: true},
			{ID: "user2", Name: "Inactive User", Active: false},
			{ID: "user3", Name: "Another Active", Active: true},
		}

		result, err := service.GetTeam(ctx, teamName)

		require.NoError(t, err)
		assertTeamView(t, result, teamName, expectedMembers)
	})
}

func TestTeamService_GetTeam_Error(t *testing.T) {
	service, teamRepo, userRepo, _, _ := setupTest(t)
	ctx := context.Background()

	t.Run("team not found", func(t *testing.T) {
		teamName := "nonexistent-team"

		teamRepo.EXPECT().GetByName(ctx, teamName).Return(nil, errors.New("team not found"))

		result, err := service.GetTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "team not found")
	})

	t.Run("user repo error", func(t *testing.T) {
		teamName := "test-team"
		memberIDs := []string{"user1"}

		teamRepo.EXPECT().GetByName(ctx, teamName).Return(&teams.Team{
			Name:      teamName,
			MemberIDs: memberIDs,
		}, nil)

		userRepo.EXPECT().GetMany(ctx, "user1").Return(nil, errors.New("user fetch error"))

		result, err := service.GetTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user fetch error")
	})
}
