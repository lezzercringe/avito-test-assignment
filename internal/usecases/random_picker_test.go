package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
	"github.com/lezzercringe/avito-test-assignment/mocks"
)

func setupPickerTest(t *testing.T) (*usecases.RandomReviewerPicker, *mocks.MockuserRepository) {
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockuserRepository(ctrl)
	picker := usecases.NewRandomReviewerPicker(userRepo)
	return picker, userRepo
}

func assertPickedReviewers(t *testing.T, result []string, expectedCount int, allowedIDs []string) {
	assert.Len(t, result, expectedCount)
	for _, id := range result {
		assert.Contains(t, allowedIDs, id, "picked reviewer %s not in allowed set", id)
	}
	seen := make(map[string]bool)
	for _, id := range result {
		assert.False(t, seen[id], "duplicate reviewer ID: %s", id)
		seen[id] = true
	}
}

func TestRandomReviewerPicker_PickReviewersFromTeam(t *testing.T) {
	picker, userRepo := setupPickerTest(t)
	ctx := context.Background()

	t.Run("picks reviewers from team excluding specified users", func(t *testing.T) {
		team := &teams.Team{
			Name:      "test-team",
			MemberIDs: []string{"user1", "user2", "user3", "user4"},
		}
		req := usecases.PickReviewersRequest{
			UserIDsToExclude: []string{"user2"},
			Team:             *team,
			WantCount:        2,
		}

		userRepo.EXPECT().GetMany(ctx, "user1", "user3", "user4").Return([]*users.User{
			{ID: "user1", Name: "User 1", Active: true},
			{ID: "user3", Name: "User 3", Active: true},
			{ID: "user4", Name: "User 4", Active: false},
		}, nil)

		result, err := picker.PickReviewersFromTeam(ctx, req)

		require.NoError(t, err)
		assertPickedReviewers(t, result, 2, []string{"user1", "user3"})
	})

	t.Run("returns error when all team members are excluded", func(t *testing.T) {
		team := &teams.Team{
			Name:      "test-team",
			MemberIDs: []string{"user1", "user2"},
		}
		req := usecases.PickReviewersRequest{
			UserIDsToExclude: []string{"user1", "user2"},
			Team:             *team,
			WantCount:        1,
		}

		result, err := picker.PickReviewersFromTeam(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrNoCandidate, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when all retrieved users are inactive", func(t *testing.T) {
		team := &teams.Team{
			Name:      "test-team",
			MemberIDs: []string{"user1", "user2"},
		}
		req := usecases.PickReviewersRequest{
			UserIDsToExclude: []string{},
			Team:             *team,
			WantCount:        1,
		}

		userRepo.EXPECT().GetMany(ctx, "user1", "user2").Return([]*users.User{
			{ID: "user1", Name: "User 1", Active: false},
			{ID: "user2", Name: "User 2", Active: false},
		}, nil)

		result, err := picker.PickReviewersFromTeam(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrNoCandidate, err)
		assert.Nil(t, result)
	})

	t.Run("ensures no duplicate reviewers in selection", func(t *testing.T) {
		team := &teams.Team{
			Name:      "test-team",
			MemberIDs: []string{"user1", "user2", "user3", "user4", "user5"},
		}
		req := usecases.PickReviewersRequest{
			UserIDsToExclude: []string{},
			Team:             *team,
			WantCount:        3,
		}

		userRepo.EXPECT().GetMany(ctx, "user1", "user2", "user3", "user4", "user5").Return([]*users.User{
			{ID: "user1", Name: "User 1", Active: true},
			{ID: "user2", Name: "User 2", Active: true},
			{ID: "user3", Name: "User 3", Active: true},
			{ID: "user4", Name: "User 4", Active: true},
			{ID: "user5", Name: "User 5", Active: true},
		}, nil)

		result, err := picker.PickReviewersFromTeam(ctx, req)

		require.NoError(t, err)
		assertPickedReviewers(t, result, 3, []string{"user1", "user2", "user3", "user4", "user5"})
	})

	t.Run("propagates repository error when user retrieval fails", func(t *testing.T) {
		team := &teams.Team{
			Name:      "test-team",
			MemberIDs: []string{"user1", "user2"},
		}
		req := usecases.PickReviewersRequest{
			UserIDsToExclude: []string{},
			Team:             *team,
			WantCount:        1,
		}

		expectedErr := errors.New("database connection failed")
		userRepo.EXPECT().GetMany(ctx, "user1", "user2").Return(nil, expectedErr)

		result, err := picker.PickReviewersFromTeam(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "retrieving candidates")
		assert.Contains(t, err.Error(), expectedErr.Error())
		assert.Nil(t, result)
	})
}
