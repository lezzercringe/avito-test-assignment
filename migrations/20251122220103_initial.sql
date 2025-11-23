-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
    name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS memberships (
    team_name VARCHAR(255) REFERENCES teams(name),
    user_id VARCHAR(255) REFERENCES users(id),
    PRIMARY KEY (team_name, user_id) -- implicit index on team_name
);

CREATE TABLE IF NOT EXISTS pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    original_team_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(id) NOT NULL,
    status VARCHAR(255) NOT NULL,
    merged_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS reviewers (
    pull_request_id VARCHAR(255) REFERENCES pull_requests(id),
    user_id VARCHAR(255) REFERENCES users(id),
    PRIMARY KEY (user_id, pull_request_id) -- implicit index on user_id
);


CREATE INDEX IF NOT EXISTS idx_memberships_user_id ON memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_reviewers_pull_request_id ON reviewers(pull_request_id);

-- +goose Down
DROP INDEX IF EXISTS idx_memberships_user_id;
DROP INDEX IF EXISTS idx_reviewers_pull_request_id;

DROP TABLE IF EXISTS reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
