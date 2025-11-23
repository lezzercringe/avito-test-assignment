package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
	"github.com/lezzercringe/avito-test-assignment/mocks"
)

func setupUserTest(t *testing.T) (*usecases.UserServiceImpl, *mocks.MockteamRepository, *mocks.MockuserRepository, *mocks.MockprRepository, *mocks.MockReviewerPicker, *mocks.MockTxManager, *mocks.MockTxHandle) {
	ctrl := gomock.NewController(t)

	teamRepo := mocks.NewMockteamRepository(ctrl)
	userRepo := mocks.NewMockuserRepository(ctrl)
	prRepo := mocks.NewMockprRepository(ctrl)
	rpicker := mocks.NewMockReviewerPicker(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	txHandle := mocks.NewMockTxHandle(ctrl)

	service := usecases.NewUserService(txManager, userRepo, teamRepo, prRepo, rpicker)

	return service, teamRepo, userRepo, prRepo, rpicker, txManager, txHandle
}

func assertUserView(t *testing.T, result *usecases.UserView, expectedID, expectedName string, expectedActive bool) {
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, expectedName, result.Name)
	assert.Equal(t, expectedActive, result.IsActive)
}

func assertUserWithTeamView(t *testing.T, result *usecases.UserWithTeamView, expectedID, expectedName, expectedTeamName string, expectedActive bool) {
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, expectedName, result.Name)
	assert.Equal(t, expectedTeamName, result.TeamName)
	assert.Equal(t, expectedActive, result.IsActive)
}

func assertReviewedPRView(t *testing.T, result *usecases.ReviewedPullRequestView, expectedID, expectedName, expectedAuthorID string, expectedStatus prs.Status) {
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, expectedName, result.Name)
	assert.Equal(t, expectedAuthorID, result.AuthorID)
	assert.Equal(t, expectedStatus, result.Status)
}

func TestUserService_GetReview(t *testing.T) {
	service, _, _, prRepo, _, _, _ := setupUserTest(t)
	ctx := context.Background()

	t.Run("user with reviewed PRs", func(t *testing.T) {
		userID := "reviewer-1"

		prs := []*prs.PullRequest{
			{
				ID:       "pr-1",
				Name:     "Fix bug",
				Status:   "OPEN",
				AuthorID: "author-1",
			},
			{
				ID:       "pr-2",
				Name:     "Add feature",
				Status:   "MERGED",
				AuthorID: "author-2",
				MergedAt: time.Now(),
			},
		}

		prRepo.EXPECT().GetManyByReviewerID(ctx, userID).Return(prs, nil)

		result, err := service.GetReview(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assertReviewedPRView(t, result[0], "pr-1", "Fix bug", "author-1", "OPEN")
		assertReviewedPRView(t, result[1], "pr-2", "Add feature", "author-2", "MERGED")
	})

	t.Run("user with no reviewed PRs", func(t *testing.T) {
		userID := "reviewer-2"

		prRepo.EXPECT().GetManyByReviewerID(ctx, userID).Return([]*prs.PullRequest{}, nil)

		result, err := service.GetReview(ctx, userID)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("repo error", func(t *testing.T) {
		userID := "reviewer-3"

		prRepo.EXPECT().GetManyByReviewerID(ctx, userID).Return(nil, errors.New("repo error"))

		result, err := service.GetReview(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "repo error")
	})
}

func TestUserService_SetIsActive(t *testing.T) {
	service, teamRepo, userRepo, _, _, _, _ := setupUserTest(t)
	ctx := context.Background()

	t.Run("activate user", func(t *testing.T) {
		userID := "user-1"

		user := &users.User{
			ID:     userID,
			Name:   "John Doe",
			Active: false,
		}

		team := &teams.Team{
			Name: "team-alpha",
		}

		userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
		userRepo.EXPECT().Save(ctx, user).Return(nil)
		teamRepo.EXPECT().GetByMemberID(ctx, userID).Return(team, nil)

		result, err := service.SetIsActive(ctx, userID, true)

		require.NoError(t, err)
		assertUserWithTeamView(t, result, userID, "John Doe", "team-alpha", true)
		assert.True(t, user.Active)
	})

	t.Run("deactivate user", func(t *testing.T) {
		userID := "user-2"

		user := &users.User{
			ID:     userID,
			Name:   "Jane Smith",
			Active: true,
		}

		team := &teams.Team{
			Name: "team-beta",
		}

		userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
		userRepo.EXPECT().Save(ctx, user).Return(nil)
		teamRepo.EXPECT().GetByMemberID(ctx, userID).Return(team, nil)

		result, err := service.SetIsActive(ctx, userID, false)

		require.NoError(t, err)
		assertUserWithTeamView(t, result, userID, "Jane Smith", "team-beta", false)
		assert.False(t, user.Active)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent"

		userRepo.EXPECT().Get(ctx, userID).Return(nil, errors.New("user not found"))

		result, err := service.SetIsActive(ctx, userID, true)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("save error", func(t *testing.T) {
		userID := "user-3"

		user := &users.User{
			ID:     userID,
			Name:   "Bob Wilson",
			Active: false,
		}

		userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
		userRepo.EXPECT().Save(ctx, user).Return(errors.New("save error"))

		result, err := service.SetIsActive(ctx, userID, true)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "save error")
	})

	t.Run("team not found", func(t *testing.T) {
		userID := "user-4"

		user := &users.User{
			ID:     userID,
			Name:   "Alice Brown",
			Active: false,
		}

		userRepo.EXPECT().Get(ctx, userID).Return(user, nil)
		userRepo.EXPECT().Save(ctx, user).Return(nil)
		teamRepo.EXPECT().GetByMemberID(ctx, userID).Return(nil, errors.New("team not found"))

		result, err := service.SetIsActive(ctx, userID, true)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "getting team")
	})
}

func TestUserService_Deactivate(t *testing.T) {
	service, teamRepo, userRepo, prRepo, rpicker, txManager, txHandle := setupUserTest(t)
	ctx := context.Background()

	t.Run("deactivate single user with no PRs", func(t *testing.T) {
		userID := "user-1"

		user := &users.User{
			ID:     userID,
			Name:   "John Doe",
			Active: true,
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return([]*usecases.PRWithMatchedReviewers{}, nil)
		userRepo.EXPECT().GetMany(ctx, userID).Return([]*users.User{user}, nil)
		userRepo.EXPECT().SaveMany(ctx, user).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.Deactivate(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assertUserView(t, result[0], userID, "John Doe", false)
		assert.False(t, user.Active)
	})

	t.Run("deactivate multiple users", func(t *testing.T) {
		userIDs := []string{"user-1", "user-2"}

		users := []*users.User{
			{ID: "user-1", Name: "John Doe", Active: true},
			{ID: "user-2", Name: "Jane Smith", Active: true},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, "user-1", "user-2").Return([]*usecases.PRWithMatchedReviewers{}, nil)
		userRepo.EXPECT().GetMany(ctx, "user-1", "user-2").Return(users, nil)
		userRepo.EXPECT().SaveMany(ctx, users[0], users[1]).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.Deactivate(ctx, userIDs...)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assertUserView(t, result[0], "user-1", "John Doe", false)
		assertUserView(t, result[1], "user-2", "Jane Smith", false)
		assert.False(t, users[0].Active)
		assert.False(t, users[1].Active)
	})

	t.Run("deactivate user with PR reassignment", func(t *testing.T) {
		userID := "reviewer-1"

		user := &users.User{
			ID:     userID,
			Name:   "John Doe",
			Active: true,
		}

		pr := &prs.PullRequest{
			ID:               "pr-1",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{userID, "reviewer-2"},
		}

		team := teams.Team{
			Name:      "team-alpha",
			MemberIDs: []string{"author-1", "reviewer-1", "reviewer-2", "new-reviewer"},
		}

		prWithReviewers := []*usecases.PRWithMatchedReviewers{
			{
				PR:                 pr,
				MatchedReviewerIDs: []string{userID},
			},
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return(prWithReviewers, nil)
		teamRepo.EXPECT().GetManyByNames(ctx, "team-alpha").Return(map[string]teams.Team{"team-alpha": team}, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{userID, "author-1"},
			WantCount:        1,
			Team:             team,
		}).Return([]string{"new-reviewer"}, nil)
		prRepo.EXPECT().SaveMany(ctx, pr).Return(nil)
		userRepo.EXPECT().GetMany(ctx, userID).Return([]*users.User{user}, nil)
		userRepo.EXPECT().SaveMany(ctx, user).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(nil)

		result, err := service.Deactivate(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assertUserView(t, result[0], userID, "John Doe", false)
		assert.False(t, user.Active)
		assert.Contains(t, pr.ReviewerIDs, "new-reviewer")
		assert.NotContains(t, pr.ReviewerIDs, userID)
	})

	t.Run("tx manager error", func(t *testing.T) {
		userID := "user-1"

		txManager.EXPECT().WithTx(ctx).Return(nil, nil, errors.New("tx error"))

		result, err := service.Deactivate(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "starting tx")
	})

	t.Run("pr reassignment error", func(t *testing.T) {
		userID := "user-1"

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return(nil, errors.New("pr error"))

		result, err := service.Deactivate(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "reassigning reviewed prs")
	})

	t.Run("user retrieval error", func(t *testing.T) {
		userID := "user-1"

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return([]*usecases.PRWithMatchedReviewers{}, nil)
		userRepo.EXPECT().GetMany(ctx, userID).Return(nil, errors.New("user error"))

		result, err := service.Deactivate(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "retrieving users")
	})

	t.Run("user save error", func(t *testing.T) {
		userID := "user-1"

		user := &users.User{
			ID:     userID,
			Name:   "John Doe",
			Active: true,
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return([]*usecases.PRWithMatchedReviewers{}, nil)
		userRepo.EXPECT().GetMany(ctx, userID).Return([]*users.User{user}, nil)
		userRepo.EXPECT().SaveMany(ctx, user).Return(errors.New("save error"))

		result, err := service.Deactivate(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "saving users")
	})

	t.Run("commit error", func(t *testing.T) {
		userID := "user-1"

		user := &users.User{
			ID:     userID,
			Name:   "John Doe",
			Active: true,
		}

		txManager.EXPECT().WithTx(ctx).Return(ctx, txHandle, nil)
		txHandle.EXPECT().Rollback(ctx).Return(nil)
		prRepo.EXPECT().GetAllUnmergedWithAnyOfReviewers(ctx, userID).Return([]*usecases.PRWithMatchedReviewers{}, nil)
		userRepo.EXPECT().GetMany(ctx, userID).Return([]*users.User{user}, nil)
		userRepo.EXPECT().SaveMany(ctx, user).Return(nil)
		txHandle.EXPECT().Commit(ctx).Return(errors.New("commit error"))

		result, err := service.Deactivate(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "commiting tx")
	})
}
