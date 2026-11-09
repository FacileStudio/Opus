-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS github_id TEXT;
CREATE INDEX IF NOT EXISTS idx_users_github_id ON users (github_id) WHERE github_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS project_github_repos (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, repo_full_name)
);
CREATE INDEX IF NOT EXISTS idx_project_github_repos_name ON project_github_repos (repo_full_name);

CREATE TABLE IF NOT EXISTS task_pull_requests (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    repo_full_name TEXT NOT NULL,
    pr_number INT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    state TEXT NOT NULL,
    head_branch TEXT NOT NULL DEFAULT '',
    base_branch TEXT NOT NULL DEFAULT '',
    author_github_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, repo_full_name, pr_number)
);
CREATE INDEX IF NOT EXISTS idx_task_pull_requests_task ON task_pull_requests (task_id);

-- +goose Down
DROP TABLE IF EXISTS task_pull_requests;
DROP TABLE IF EXISTS project_github_repos;
DROP INDEX IF EXISTS idx_users_github_id;
ALTER TABLE users DROP COLUMN IF EXISTS github_id;
