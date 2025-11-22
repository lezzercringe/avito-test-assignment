package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres/generated"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
)

type PRRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) *PRRepository {
	return &PRRepository{pool: pool}
}

func (r *PRRepository) getQueries(ctx context.Context) *generated.Queries {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return generated.New(tx)
	}
	return generated.New(r.pool)
}

func (r *PRRepository) GetByID(ctx context.Context, id string) (pr *prs.PullRequest, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	generatedPR, err := queries.GetPullRequestByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reviewerIDs, err := queries.GetPRReviewers(ctx, id)
	if err != nil {
		return nil, err
	}

	var mergedAt time.Time
	if generatedPR.MergedAt.Valid {
		mergedAt = generatedPR.MergedAt.Time
	}

	pr = &prs.PullRequest{
		ID:               generatedPR.ID,
		Name:             generatedPR.Name,
		Status:           prs.Status(generatedPR.Status),
		OriginalTeamName: generatedPR.OriginalTeamName,
		AuthorID:         generatedPR.AuthorID,
		ReviewerIDs:      reviewerIDs,
		MergedAt:         mergedAt,
	}

	return pr, nil
}

func (r *PRRepository) GetManyByReviewerID(ctx context.Context, reviewerID string) (result []prs.PullRequest, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	generatedPRs, err := queries.GetManyPullRequestsByReviewerID(ctx, reviewerID)
	if err != nil {
		return nil, err
	}

	result = make([]prs.PullRequest, len(generatedPRs))
	for i, pr := range generatedPRs {
		reviewerIDs, err := queries.GetPRReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}

		var mergedAt time.Time
		if pr.MergedAt.Valid {
			mergedAt = pr.MergedAt.Time
		}

		result[i] = prs.PullRequest{
			ID:               pr.ID,
			Name:             pr.Name,
			Status:           prs.Status(pr.Status),
			OriginalTeamName: pr.OriginalTeamName,
			AuthorID:         pr.AuthorID,
			ReviewerIDs:      reviewerIDs,
			MergedAt:         mergedAt,
		}
	}

	return result, nil
}

func (r *PRRepository) Save(ctx context.Context, pr *prs.PullRequest) (err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	var mergedAt pgtype.Timestamptz
	if !pr.MergedAt.IsZero() {
		mergedAt = pgtype.Timestamptz{Time: pr.MergedAt, Valid: true}
	}

	err = queries.SavePullRequest(ctx, generated.SavePullRequestParams{
		ID:               pr.ID,
		Name:             pr.Name,
		OriginalTeamName: pr.OriginalTeamName,
		AuthorID:         pr.AuthorID,
		Status:           string(pr.Status),
		MergedAt:         mergedAt,
	})
	if err != nil {
		return err
	}

	if len(pr.ReviewerIDs) > 0 {
		err = queries.SaveReviewers(ctx, generated.SaveReviewersParams{
			PullRequestID: pr.ID,
			Column2:       pr.ReviewerIDs,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
