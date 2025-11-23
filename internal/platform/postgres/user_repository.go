package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lezzercringe/avito-test-assignment/internal/platform/postgres/generated"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) getQueries(ctx context.Context) *generated.Queries {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return generated.New(tx)
	}
	return generated.New(r.pool)
}

func (r *UserRepository) Save(ctx context.Context, u *users.User) (err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	err = queries.SaveUser(ctx, generated.SaveUserParams{
		ID:     u.ID,
		Name:   u.Name,
		Active: u.Active,
	})
	return err
}

func (r *UserRepository) Get(ctx context.Context, id string) (user *users.User, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	generatedUser, err := queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user = &users.User{
		ID:     generatedUser.ID,
		Name:   generatedUser.Name,
		Active: generatedUser.Active,
	}

	return user, nil
}

func (r *UserRepository) SaveMany(ctx context.Context, u ...*users.User) (err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	if len(u) == 0 {
		return nil
	}

	ids := make([]string, len(u))
	names := make([]string, len(u))
	actives := make([]bool, len(u))

	for i, user := range u {
		ids[i] = user.ID
		names[i] = user.Name
		actives[i] = user.Active
	}

	err = queries.SaveManyUsers(ctx, generated.SaveManyUsersParams{
		Column1: ids,
		Column2: names,
		Column3: actives,
	})
	return err
}

func (r *UserRepository) GetMany(ctx context.Context, id ...string) (result []*users.User, err error) {
	defer func() {
		err = mapError(err)
	}()

	queries := r.getQueries(ctx)

	if len(id) == 0 {
		return []*users.User{}, nil
	}

	generatedUsers, err := queries.GetManyUsersByIDs(ctx, id)
	if err != nil {
		return nil, err
	}

	result = make([]*users.User, len(generatedUsers))
	for i, user := range generatedUsers {
		result[i] = &users.User{
			ID:     user.ID,
			Name:   user.Name,
			Active: user.Active,
		}
	}

	return result, nil
}
