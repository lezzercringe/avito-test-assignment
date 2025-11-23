package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres/generated"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) getQueries(ctx context.Context) *generated.Queries {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return generated.New(tx)
	}
	return generated.New(r.pool)
}

func (r *TeamRepository) GetByName(ctx context.Context, name string) (team *teams.Team, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	teamName, err := queries.GetTeamByName(ctx, name)
	if err != nil {
		return nil, err
	}

	memberIDs, err := queries.GetTeamMembers(ctx, teamName)
	if err != nil {
		return nil, err
	}

	memberIDsStr := make([]string, len(memberIDs))
	copy(memberIDsStr, memberIDs)

	team = &teams.Team{
		Name:      teamName,
		MemberIDs: memberIDsStr,
	}

	return team, nil
}

func (r *TeamRepository) GetByMemberID(ctx context.Context, memberID string) (team *teams.Team, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	teamName, err := queries.GetTeamByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	memberIDs, err := queries.GetTeamMembers(ctx, teamName)
	if err != nil {
		return nil, err
	}

	memberIDsStr := make([]string, len(memberIDs))
	copy(memberIDsStr, memberIDs)

	team = &teams.Team{
		Name:      teamName,
		MemberIDs: memberIDsStr,
	}

	return team, nil
}

func (r *TeamRepository) Save(ctx context.Context, t *teams.Team) (err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	err = queries.SaveTeam(ctx, t.Name)
	if err != nil {
		return err
	}

	for _, memberID := range t.MemberIDs {
		err = queries.SaveMembership(ctx, generated.SaveMembershipParams{
			TeamName: t.Name,
			UserID:   memberID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamRepository) GetManyByNames(ctx context.Context, names ...string) (result map[string]teams.Team, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	if len(names) == 0 {
		return map[string]teams.Team{}, nil
	}

	teamNames, err := queries.GetManyTeamsByNames(ctx, names)
	if err != nil {
		return nil, err
	}

	result = make(map[string]teams.Team, len(teamNames))
	for _, teamName := range teamNames {
		memberIDs, err := queries.GetTeamMembers(ctx, teamName)
		if err != nil {
			return nil, err
		}

		memberIDsStr := make([]string, len(memberIDs))
		copy(memberIDsStr, memberIDs)

		result[teamName] = teams.Team{
			Name:      teamName,
			MemberIDs: memberIDsStr,
		}
	}

	return result, nil
}
