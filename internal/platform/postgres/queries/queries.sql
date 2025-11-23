-- TEAMS

-- name: GetTeamByName :one
SELECT name FROM teams
WHERE name = $1;

-- name: GetTeamByMemberID :one
SELECT name FROM teams t
JOIN memberships m ON m.team_name = t.name 
WHERE m.user_id = $1;

-- name: SaveTeam :exec
INSERT INTO teams (name) VALUES ($1)
ON CONFLICT (name)
DO UPDATE SET name = EXCLUDED.name;

-- name: SaveMembership :exec
INSERT INTO memberships (team_name, user_id) VALUES ($1, $2);

-- USERS

-- name: GetUserByID :one
SELECT id, name, active FROM users
WHERE id = $1;

-- name: GetManyUsersByIDs :many
SELECT id, name, active FROM users
WHERE id = ANY($1::varchar[]);

-- name: SaveUser :exec
INSERT INTO users (id, name, active) VALUES ($1, $2, $3)
ON CONFLICT (id) 
DO UPDATE SET name = excluded.name, active = excluded.active;

-- name: SaveManyUsers :exec
INSERT INTO users (id, name, active)
SELECT UNNEST($1::varchar[]), UNNEST($2::varchar[]), UNNEST($3::boolean[])
ON CONFLICT (id)
DO UPDATE SET
    name = EXCLUDED.name,
    active = EXCLUDED.active;


-- PRs

-- name: GetPullRequestByID :one
SELECT id, name, original_team_name, author_id, status, merged_at FROM pull_requests
WHERE id = $1;

-- name: GetManyPullRequestsByReviewerID :many
SELECT id, name, original_team_name, author_id, status, merged_at FROM pull_requests pr
JOIN reviewers r ON pr.id = r.pull_request_id
WHERE r.user_id = $1;

-- name: SavePullRequest :exec
INSERT INTO pull_requests(id, name, original_team_name, author_id, status, merged_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    original_team_name = EXCLUDED.original_team_name,
    author_id = EXCLUDED.author_id,
    status = EXCLUDED.status,
    merged_at = EXCLUDED.merged_at;

-- name: GetTeamMembers :many
SELECT user_id FROM memberships WHERE team_name = $1;

-- name: GetPRReviewers :many  
SELECT user_id FROM reviewers WHERE pull_request_id = $1;

-- name: SaveReviewers :exec
INSERT INTO reviewers (pull_request_id, user_id) 
SELECT $1, UNNEST($2::varchar[])
ON CONFLICT (user_id, pull_request_id) DO NOTHING;

-- name: DeleteReviewers :exec
DELETE FROM reviewers WHERE pull_request_id = $1 AND user_id = ANY($2::varchar[]);

-- name: GetManyTeamsByNames :many
SELECT name FROM teams
WHERE name = ANY($1::varchar[]);

-- name: GetPRsWithAnyReviewers :many
SELECT
    pr.id,
    pr.name,
    pr.original_team_name,
    pr.author_id,
    pr.status,
    pr.merged_at,
    ARRAY_AGG(r.user_id)::varchar[] AS matched_reviewer_ids
FROM pull_requests pr
JOIN reviewers r
  ON pr.id = r.pull_request_id
WHERE r.user_id = ANY($1::varchar[])
  -- AND pr.status = "OPEN"
GROUP BY pr.id, pr.name, pr.original_team_name, pr.author_id, pr.status, pr.merged_at
ORDER BY pr.id;

-- name: SaveManyPullRequests :exec
INSERT INTO pull_requests(id, name, original_team_name, author_id, status, merged_at)
SELECT UNNEST($1::varchar[]), UNNEST($2::varchar[]), UNNEST($3::varchar[]), UNNEST($4::varchar[]), UNNEST($5::varchar[]), UNNEST($6::timestamptz[])
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    original_team_name = EXCLUDED.original_team_name,
    author_id = EXCLUDED.author_id,
    status = EXCLUDED.status,
    merged_at = EXCLUDED.merged_at;

-- name: DeleteAllReviewersForPRs :exec
DELETE FROM reviewers WHERE pull_request_id = ANY($1::varchar[]);

-- name: SaveManyReviewers :exec
INSERT INTO reviewers (pull_request_id, user_id)
SELECT UNNEST($1::varchar[]), UNNEST($2::varchar[])
ON CONFLICT (user_id, pull_request_id) DO NOTHING;
