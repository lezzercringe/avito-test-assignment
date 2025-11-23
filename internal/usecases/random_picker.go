package usecases

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

var _ ReviewerPicker = &RandomReviewerPicker{}

type RandomReviewerPicker struct {
	userRepo userRepository
}

func filter[T any](xs []T, filterFn func(T) bool) []T {
	filtered := make([]T, 0, len(xs))
	for _, el := range xs {
		if filterFn(el) {
			filtered = append(filtered, el)
		}
	}
	return filtered
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (r *RandomReviewerPicker) PickReviewersFromTeam(ctx context.Context, req PickReviewersRequest) ([]string, error) {
	includedIDs := filter(
		req.Team.MemberIDs,
		func(id string) bool { return !slices.Contains(req.UserIDsToExclude, id) },
	)

	if len(includedIDs) == 0 {
		return nil, errorsx.ErrNoCandidate
	}

	included, err := r.userRepo.GetMany(ctx, includedIDs...)
	if err != nil {
		return nil, fmt.Errorf("retrieving candidates: %w", err)
	}

	active := filter(included, func(u *users.User) bool { return u.Active })
	if len(active) == 0 {
		return nil, errorsx.ErrNoCandidate
	}

	countToPick := minInt(req.WantCount, len(active))
	pickedIDs := make([]string, 0, countToPick)

	for range countToPick {
		ix := rand.IntN(len(active))
		pickedIDs = append(pickedIDs, active[ix].ID)
		active = append(active[:ix], active[ix+1:]...)
	}

	return pickedIDs, nil
}

func NewRandomReviewerPicker(userRepo userRepository) *RandomReviewerPicker {
	return &RandomReviewerPicker{userRepo: userRepo}
}
