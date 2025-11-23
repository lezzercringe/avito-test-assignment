package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres/generated"
	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/usecases"
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

func (r *PRRepository) GetManyByReviewerID(ctx context.Context, reviewerID string) (result []*prs.PullRequest, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	generatedPRs, err := queries.GetManyPullRequestsByReviewerID(ctx, reviewerID)
	if err != nil {
		return nil, err
	}

	result = make([]*prs.PullRequest, len(generatedPRs))
	for i, pr := range generatedPRs {
		reviewerIDs, err := queries.GetPRReviewers(ctx, pr.ID)
		if err != nil {
			return nil, err
		}

		var mergedAt time.Time
		if pr.MergedAt.Valid {
			mergedAt = pr.MergedAt.Time
		}

		result[i] = &prs.PullRequest{
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

	err = queries.DeleteAllReviewersForPRs(ctx, []string{pr.ID})
	if err != nil {
		return err
	}

	if len(pr.ReviewerIDs) > 0 {
		prIDs := make([]string, len(pr.ReviewerIDs))
		for i := range pr.ReviewerIDs {
			prIDs[i] = pr.ID
		}

		err = queries.SaveManyReviewers(ctx, generated.SaveManyReviewersParams{
			Column1: prIDs,
			Column2: pr.ReviewerIDs,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PRRepository) GetAllUnmergedWithAnyOfReviewers(ctx context.Context, reviewerIDs ...string) (result []*usecases.PRWithMatchedReviewers, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	if len(reviewerIDs) == 0 {
		return []*usecases.PRWithMatchedReviewers{}, nil
	}

	rows, err := queries.GetPRsWithAnyReviewers(ctx, reviewerIDs)
	if err != nil {
		return nil, err
	}

	result = make([]*usecases.PRWithMatchedReviewers, len(rows))
	for i, row := range rows {
		var mergedAt time.Time
		if row.MergedAt.Valid {
			mergedAt = row.MergedAt.Time
		}

		result[i] = &usecases.PRWithMatchedReviewers{
			PR: &prs.PullRequest{
				ID:               row.ID,
				Name:             row.Name,
				Status:           prs.Status(row.Status),
				OriginalTeamName: row.OriginalTeamName,
				AuthorID:         row.AuthorID,
				ReviewerIDs:      row.MatchedReviewerIds,
				MergedAt:         mergedAt,
			},
			MatchedReviewerIDs: row.MatchedReviewerIds,
		}
	}

	return result, nil
}

func (r *PRRepository) SaveMany(ctx context.Context, prs ...*prs.PullRequest) (err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	if len(prs) == 0 {
		return nil
	}

	ids := make([]string, len(prs))
	names := make([]string, len(prs))
	originalTeamNames := make([]string, len(prs))
	authorIDs := make([]string, len(prs))
	statuses := make([]string, len(prs))
	mergedAts := make([]time.Time, len(prs))

	for i, pr := range prs {
		ids[i] = pr.ID
		names[i] = pr.Name
		originalTeamNames[i] = pr.OriginalTeamName
		authorIDs[i] = pr.AuthorID
		statuses[i] = string(pr.Status)

		if !pr.MergedAt.IsZero() {
			mergedAts[i] = pr.MergedAt
		} else {
			mergedAts[i] = time.Time{}
		}
	}

	err = queries.SaveManyPullRequests(ctx, generated.SaveManyPullRequestsParams{
		Column1: ids,
		Column2: names,
		Column3: originalTeamNames,
		Column4: authorIDs,
		Column5: statuses,
		Column6: mergedAts,
	})
	if err != nil {
		return err
	}

	prIDsToDelete := make([]string, len(prs))
	for i, pr := range prs {
		prIDsToDelete[i] = pr.ID
	}

	err = queries.DeleteAllReviewersForPRs(ctx, prIDsToDelete)
	if err != nil {
		return err
	}

	var allPRIDs []string
	var allReviewerIDs []string

	for _, pr := range prs {
		if len(pr.ReviewerIDs) > 0 {
			for _, reviewerID := range pr.ReviewerIDs {
				allPRIDs = append(allPRIDs, pr.ID)
				allReviewerIDs = append(allReviewerIDs, reviewerID)
			}
		}
	}

	if len(allPRIDs) > 0 {
		err = queries.SaveManyReviewers(ctx, generated.SaveManyReviewersParams{
			Column1: allPRIDs,
			Column2: allReviewerIDs,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
