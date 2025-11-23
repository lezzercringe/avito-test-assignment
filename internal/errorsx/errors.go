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
	ErrToManyReviewers            = errors.New("too many reviewers per pr")
	ErrCandidateIsAlreadyReviewer = errors.New("candidate is already a reviewer")
)

// team-specific errors
var (
	ErrTeamName        = errors.New("invalid team name")
	ErrDuplicateMember = errors.New("duplicate team member")
)

// user-specific errors
var (
	ErrUserName = errors.New("invalid user name")
	ErrUserID   = errors.New("invalid user id")
)
