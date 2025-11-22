package usecases

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"

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

func (r *RandomReviewerPicker) PickReviewerFromTeam(ctx context.Context, req PickReviewerRequest) (string, error) {
	pickableIDs := filter(req.Team.MemberIDs, func(id string) bool { return !slices.Contains(req.UserIDsToExclude, id) })

	coworkers, err := r.userRepo.GetMany(ctx, pickableIDs...)
	if err != nil {
		return "", fmt.Errorf("retrieving candidates: %w", err)
	}

	active := filter(coworkers, func(u users.User) bool { return u.Active })
	ix := rand.IntN(len(active))
	return active[ix].ID, nil
}

func NewRandomReviewerPicker(userRepo userRepository) *RandomReviewerPicker {
	return &RandomReviewerPicker{userRepo: userRepo}
}
