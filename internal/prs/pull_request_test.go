package prs_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
)

func TestPullRequest_ToView(t *testing.T) {
	pr := &prs.PullRequest{
		ID:               "pr-123",
		Name:             "Fix bug",
		Status:           prs.StatusOpen,
		OriginalTeamName: "team-alpha",
		AuthorID:         "author-1",
		ReviewerIDs:      []string{"reviewer-1", "reviewer-2"},
		MergedAt:         time.Time{},
	}

	view := pr.ToView()

	assert.Equal(t, pr.ID, view.ID)
	assert.Equal(t, pr.Name, view.Name)
	assert.Equal(t, pr.AuthorID, view.AuthorID)
	assert.Equal(t, string(pr.Status), view.Status)
	assert.Equal(t, pr.ReviewerIDs, view.ReviewerIDs)
	assert.Equal(t, pr.MergedAt, view.MergedAt)
}

func TestPullRequest_Merge(t *testing.T) {
	t.Run("merge open PR", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:     "pr-123",
			Name:   "Fix bug",
			Status: prs.StatusOpen,
		}

		pr.Merge()

		assert.Equal(t, prs.StatusMerged, pr.Status)
		assert.False(t, pr.MergedAt.IsZero())
	})

	t.Run("idempotent merge - already merged PR", func(t *testing.T) {
		originalMergedAt := time.Now().Add(-time.Hour)
		pr := &prs.PullRequest{
			ID:       "pr-456",
			Name:     "Add feature",
			Status:   prs.StatusMerged,
			MergedAt: originalMergedAt,
		}

		pr.Merge()

		assert.Equal(t, prs.StatusMerged, pr.Status)
		assert.Equal(t, originalMergedAt, pr.MergedAt)
	})
}

func TestPullRequest_AssignReviewer(t *testing.T) {
	t.Run("assign reviewer to open PR", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusOpen,
			ReviewerIDs: []string{"reviewer-1"},
		}

		err := pr.AssignReviewer("reviewer-2")

		require.NoError(t, err)
		assert.Contains(t, pr.ReviewerIDs, "reviewer-2")
		assert.Len(t, pr.ReviewerIDs, 2)
	})

	t.Run("cannot assign to merged PR", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusMerged,
			ReviewerIDs: []string{"reviewer-1"},
		}

		err := pr.AssignReviewer("reviewer-2")

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrModifyMergedPR, err)
		assert.Len(t, pr.ReviewerIDs, 1)
	})

	t.Run("cannot assign already assigned reviewer", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusOpen,
			ReviewerIDs: []string{"reviewer-1"},
		}

		err := pr.AssignReviewer("reviewer-1")

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrCandidateIsAlreadyReviewer, err)
		assert.Len(t, pr.ReviewerIDs, 1)
	})

	t.Run("cannot assign more than max reviewers", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusOpen,
			ReviewerIDs: []string{"reviewer-1", "reviewer-2"},
		}

		err := pr.AssignReviewer("reviewer-3")

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrToManyReviewers, err)
		assert.Len(t, pr.ReviewerIDs, 2)
	})
}

func TestPullRequest_UnassignReviewer(t *testing.T) {
	t.Run("unassign reviewer from open PR", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusOpen,
			ReviewerIDs: []string{"reviewer-1", "reviewer-2"},
		}

		err := pr.UnassignReviewer("reviewer-1")

		require.NoError(t, err)
		assert.NotContains(t, pr.ReviewerIDs, "reviewer-1")
		assert.Contains(t, pr.ReviewerIDs, "reviewer-2")
		assert.Len(t, pr.ReviewerIDs, 1)
	})

	t.Run("cannot unassign from merged PR", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusMerged,
			ReviewerIDs: []string{"reviewer-1", "reviewer-2"},
		}

		err := pr.UnassignReviewer("reviewer-1")

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrModifyMergedPR, err)
		assert.Len(t, pr.ReviewerIDs, 2)
	})

	t.Run("cannot unassign not assigned reviewer", func(t *testing.T) {
		pr := &prs.PullRequest{
			ID:          "pr-123",
			Status:      prs.StatusOpen,
			ReviewerIDs: []string{"reviewer-1", "reviewer-2"},
		}

		err := pr.UnassignReviewer("reviewer-3")

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrNotPreviouslyAssigned, err)
		assert.Len(t, pr.ReviewerIDs, 2)
	})
}
