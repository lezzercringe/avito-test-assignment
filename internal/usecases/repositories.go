package usecases

import (
	"context"

	"github.com/lezzercringe/avito-test-assignment/internal/prs"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

type teamRepository interface {
	GetByName(ctx context.Context, name string) (*teams.Team, error)
	GetManyByNames(ctx context.Context, names ...string) (map[string]teams.Team, error)
	GetByMemberID(ctx context.Context, memberID string) (*teams.Team, error)
	Save(ctx context.Context, t *teams.Team) error
}

type PRWithMatchedReviewers struct {
	PR                 *prs.PullRequest
	MatchedReviewerIDs []string
}

type prRepository interface {
	GetByID(ctx context.Context, id string) (*prs.PullRequest, error)
	GetManyByReviewerID(ctx context.Context, reviewerID string) ([]*prs.PullRequest, error)
	GetAllUnmergedWithAnyOfReviewers(ctx context.Context, reviewerIDs ...string) ([]*PRWithMatchedReviewers, error)
	Save(ctx context.Context, pr *prs.PullRequest) error
	SaveMany(ctx context.Context, el ...*prs.PullRequest) error
}

type userRepository interface {
	Save(ctx context.Context, u *users.User) error
	Get(ctx context.Context, id string) (*users.User, error)

	SaveMany(ctx context.Context, u ...*users.User) error
	GetMany(ctx context.Context, id ...string) ([]*users.User, error)
}
