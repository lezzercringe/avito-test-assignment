package errorsx

import "errors"

// general errors
var (
	ErrAlreadyExists = errors.New("entity already exists")
	ErrNotFound      = errors.New("required entity was not found")
)

// pr-specific errors
var (
	ErrModifyMergedPR             = errors.New("cannot modify merged pull request")
	ErrNotPreviouslyAssigned      = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate                = errors.New("no active replacement candidate in team")
	ErrToManyReviewers            = errors.New("TODO")
	ErrCandidateIsAlreadyReviewer = errors.New("TODO")
)

// team-specific errors
var (
	ErrTeamName        = errors.New("TODO")
	ErrDuplicateMember = errors.New("TODO")
)

// user-specific errors
var (
	ErrUserName = errors.New("TODO")
	ErrUserID   = errors.New("TODO")
)
