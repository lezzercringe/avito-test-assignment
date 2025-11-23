package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
	"github.com/lezzercringe/avito-test-assignment/mocks"
)

func setupPRTest(t *testing.T) (*usecases.PullRequestServiceImpl, *mocks.MockReviewerPicker, *mocks.MockprRepository, *mocks.MockteamRepository) {
	ctrl := gomock.NewController(t)

	rpicker := mocks.NewMockReviewerPicker(ctrl)
	prRepo := mocks.NewMockprRepository(ctrl)
	teamRepo := mocks.NewMockteamRepository(ctrl)

	service := usecases.NewPullRequestService(rpicker, prRepo, teamRepo)

	return service, rpicker, prRepo, teamRepo
}

func assertPRView(t *testing.T, result *prs.PullRequestView, expectedID, expectedName, expectedAuthorID string, expectedStatus string, expectedReviewerIDs []string) {
	assert.Equal(t, expectedID, result.ID)
	assert.Equal(t, expectedName, result.Name)
	assert.Equal(t, expectedAuthorID, result.AuthorID)
	assert.Equal(t, expectedStatus, result.Status)
	assert.Equal(t, expectedReviewerIDs, result.ReviewerIDs)
}

func assertReassignResult(t *testing.T, result *usecases.ReassignReviewerResult, expectedPR *prs.PullRequestView, expectedReplacedByID string) {
	assertPRView(t, result.PR, expectedPR.ID, expectedPR.Name, expectedPR.AuthorID, expectedPR.Status, expectedPR.ReviewerIDs)
	assert.Equal(t, expectedReplacedByID, result.ReplacedByID)
}

func TestPullRequestService_Create_Ok(t *testing.T) {
	service, rpicker, prRepo, teamRepo := setupPRTest(t)
	ctx := context.Background()

	t.Run("create PR with reviewers found", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-123",
			AuthorID: "author-1",
			Name:     "Fix bug",
		}

		team := &teams.Team{
			Name:      "team-alpha",
			MemberIDs: []string{"author-1", "user-2", "user-3"},
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.AuthorID},
			Team:             *team,
			WantCount:        2,
		}).Return([]string{"user-2", "user-3"}, nil)
		prRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

		result, err := service.Create(ctx, req)

		require.NoError(t, err)
		assertPRView(t, result, req.ID, req.Name, req.AuthorID, "OPEN", []string{"user-2", "user-3"})
	})

	t.Run("create PR with no reviewers found", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-456",
			AuthorID: "author-2",
			Name:     "Add feature",
		}

		team := &teams.Team{
			Name:      "team-beta",
			MemberIDs: []string{"author-2"},
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.AuthorID},
			Team:             *team,
			WantCount:        2,
		}).Return(nil, errorsx.ErrNoCandidate)
		prRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

		result, err := service.Create(ctx, req)

		require.NoError(t, err)
		assertPRView(t, result, req.ID, req.Name, req.AuthorID, "OPEN", nil)
	})

	t.Run("create PR with partial reviewers", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-789",
			AuthorID: "author-3",
			Name:     "Refactor code",
		}

		team := &teams.Team{
			Name:      "team-gamma",
			MemberIDs: []string{"author-3", "user-4"},
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.AuthorID},
			Team:             *team,
			WantCount:        2,
		}).Return([]string{"user-4"}, nil)
		prRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

		result, err := service.Create(ctx, req)

		require.NoError(t, err)
		assertPRView(t, result, req.ID, req.Name, req.AuthorID, "OPEN", []string{"user-4"})
	})
}

func TestPullRequestService_Create_Error(t *testing.T) {
	service, rpicker, prRepo, teamRepo := setupPRTest(t)
	ctx := context.Background()

	t.Run("team not found", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-123",
			AuthorID: "author-1",
			Name:     "Fix bug",
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(nil, errors.New("team not found"))

		result, err := service.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "team not found")
	})

	t.Run("reviewer picker error", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-456",
			AuthorID: "author-2",
			Name:     "Add feature",
		}

		team := &teams.Team{
			Name:      "team-beta",
			MemberIDs: []string{"author-2", "user-1"},
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.AuthorID},
			Team:             *team,
			WantCount:        2,
		}).Return(nil, errors.New("picker error"))

		result, err := service.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "picking reviewer")
	})

	t.Run("pr save error", func(t *testing.T) {
		req := usecases.CreateRequest{
			ID:       "pr-789",
			AuthorID: "author-3",
			Name:     "Refactor code",
		}

		team := &teams.Team{
			Name:      "team-gamma",
			MemberIDs: []string{"author-3", "user-4"},
		}

		teamRepo.EXPECT().GetByMemberID(ctx, req.AuthorID).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.AuthorID},
			Team:             *team,
			WantCount:        2,
		}).Return([]string{"user-4"}, nil)
		prRepo.EXPECT().Save(ctx, gomock.Any()).Return(errors.New("save error"))

		result, err := service.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "saving pr")
	})
}

func TestPullRequestService_ReassignReviewer_Ok(t *testing.T) {
	service, rpicker, prRepo, teamRepo := setupPRTest(t)
	ctx := context.Background()

	t.Run("successful reviewer reassignment", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-123",
			UserIDToReassign: "old-reviewer",
		}

		pr := &prs.PullRequest{
			ID:               "pr-123",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"old-reviewer", "another-reviewer"},
		}

		team := &teams.Team{
			Name:      "team-alpha",
			MemberIDs: []string{"author-1", "old-reviewer", "new-reviewer", "another-reviewer"},
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(pr, nil)
		teamRepo.EXPECT().GetByName(ctx, pr.OriginalTeamName).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.UserIDToReassign},
			Team:             *team,
			WantCount:        1,
		}).Return([]string{"new-reviewer"}, nil)
		prRepo.EXPECT().Save(ctx, pr).Return(nil)

		result, err := service.ReassignReviewer(ctx, req)

		require.NoError(t, err)
		expectedPR := &prs.PullRequestView{
			ID:          "pr-123",
			Name:        "Fix bug",
			AuthorID:    "author-1",
			Status:      "OPEN",
			ReviewerIDs: []string{"another-reviewer", "new-reviewer"},
		}
		assertReassignResult(t, result, expectedPR, "new-reviewer")
	})
}

func TestPullRequestService_ReassignReviewer_Error(t *testing.T) {
	service, rpicker, prRepo, teamRepo := setupPRTest(t)
	ctx := context.Background()

	t.Run("pr not found", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-999",
			UserIDToReassign: "old-reviewer",
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(nil, errors.New("pr not found"))

		result, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "retreiving pull request")
	})

	t.Run("unassign reviewer error", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-123",
			UserIDToReassign: "non-reviewer",
		}

		pr := &prs.PullRequest{
			ID:               "pr-123",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(pr, nil)

		result, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unassigning original reviewer")
	})

	t.Run("team not found", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-123",
			UserIDToReassign: "reviewer-1",
		}

		pr := &prs.PullRequest{
			ID:               "pr-123",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-missing",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(pr, nil)
		teamRepo.EXPECT().GetByName(ctx, pr.OriginalTeamName).Return(nil, errors.New("team not found"))

		result, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "getting original pr team")
	})

	t.Run("picker error", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-123",
			UserIDToReassign: "reviewer-1",
		}

		pr := &prs.PullRequest{
			ID:               "pr-123",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		}

		team := &teams.Team{
			Name:      "team-alpha",
			MemberIDs: []string{"author-1", "reviewer-1"},
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(pr, nil)
		teamRepo.EXPECT().GetByName(ctx, pr.OriginalTeamName).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.UserIDToReassign},
			Team:             *team,
			WantCount:        1,
		}).Return(nil, errors.New("picker error"))

		result, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "picking new reviewer")
	})

	t.Run("save error", func(t *testing.T) {
		req := usecases.ReassignReviewerRequest{
			PullRequestID:    "pr-123",
			UserIDToReassign: "reviewer-1",
		}

		pr := &prs.PullRequest{
			ID:               "pr-123",
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		}

		team := &teams.Team{
			Name:      "team-alpha",
			MemberIDs: []string{"author-1", "reviewer-1", "new-reviewer"},
		}

		prRepo.EXPECT().GetByID(ctx, req.PullRequestID).Return(pr, nil)
		teamRepo.EXPECT().GetByName(ctx, pr.OriginalTeamName).Return(team, nil)
		rpicker.EXPECT().PickReviewersFromTeam(ctx, usecases.PickReviewersRequest{
			UserIDsToExclude: []string{req.UserIDToReassign},
			Team:             *team,
			WantCount:        1,
		}).Return([]string{"new-reviewer"}, nil)
		prRepo.EXPECT().Save(ctx, pr).Return(errors.New("save error"))

		result, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "saving pr")
	})
}

func TestPullRequestService_Merge_Ok(t *testing.T) {
	service, _, prRepo, _ := setupPRTest(t)
	ctx := context.Background()

	t.Run("merge open PR", func(t *testing.T) {
		prID := "pr-123"

		pr := &prs.PullRequest{
			ID:               prID,
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		}

		prRepo.EXPECT().GetByID(ctx, prID).Return(pr, nil)
		prRepo.EXPECT().Save(ctx, pr).Return(nil)

		result, err := service.Merge(ctx, prID)

		require.NoError(t, err)
		assertPRView(t, result, prID, "Fix bug", "author-1", "MERGED", []string{"reviewer-1", "reviewer-2"})
		assert.True(t, !result.MergedAt.IsZero())
	})

	t.Run("merge already merged PR", func(t *testing.T) {
		prID := "pr-456"
		mergedAt := time.Now().Add(-time.Hour)

		pr := &prs.PullRequest{
			ID:               prID,
			Name:             "Add feature",
			Status:           prs.StatusMerged,
			OriginalTeamName: "team-beta",
			AuthorID:         "author-2",
			ReviewerIDs:      []string{"reviewer-3"},
			MergedAt:         mergedAt,
		}

		prRepo.EXPECT().GetByID(ctx, prID).Return(pr, nil)
		prRepo.EXPECT().Save(ctx, pr).Return(nil)

		result, err := service.Merge(ctx, prID)

		require.NoError(t, err)
		assertPRView(t, result, prID, "Add feature", "author-2", "MERGED", []string{"reviewer-3"})
		assert.Equal(t, mergedAt, result.MergedAt)
	})
}

func TestPullRequestService_Merge_Error(t *testing.T) {
	service, _, prRepo, _ := setupPRTest(t)
	ctx := context.Background()

	t.Run("pr not found", func(t *testing.T) {
		prID := "pr-999"

		prRepo.EXPECT().GetByID(ctx, prID).Return(nil, errors.New("pr not found"))

		result, err := service.Merge(ctx, prID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "retrieving pr by id")
	})

	t.Run("save error", func(t *testing.T) {
		prID := "pr-123"

		pr := &prs.PullRequest{
			ID:               prID,
			Name:             "Fix bug",
			Status:           prs.StatusOpen,
			OriginalTeamName: "team-alpha",
			AuthorID:         "author-1",
			ReviewerIDs:      []string{"reviewer-1"},
		}

		prRepo.EXPECT().GetByID(ctx, prID).Return(pr, nil)
		prRepo.EXPECT().Save(ctx, pr).Return(errors.New("save error"))

		result, err := service.Merge(ctx, prID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "saving pr")
	})
}
