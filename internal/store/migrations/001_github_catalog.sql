-- migration: github_catalog
-- up
CREATE TABLE github_profiles (
    login TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    fetched_at TEXT NOT NULL
);

CREATE TABLE github_repos (
    owner_login TEXT NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    payload TEXT NOT NULL,
    fetched_at TEXT NOT NULL,
    PRIMARY KEY (owner_login, name, source)
);

CREATE TABLE github_orgs (
    user_login TEXT NOT NULL,
    org_login TEXT NOT NULL,
    payload TEXT NOT NULL,
    PRIMARY KEY (user_login, org_login)
);

CREATE TABLE traffic_windows (
    owner_login TEXT NOT NULL,
    repo_name TEXT NOT NULL,
    views INTEGER NOT NULL,
    view_uniques INTEGER NOT NULL,
    clones INTEGER NOT NULL,
    clone_uniques INTEGER NOT NULL,
    captured_at TEXT NOT NULL,
    PRIMARY KEY (owner_login, repo_name)
);

CREATE TABLE catalog_watches (
    login TEXT PRIMARY KEY,
    last_viewed_at TEXT NOT NULL
);

-- down
DROP TABLE IF EXISTS catalog_watches;
DROP TABLE IF EXISTS traffic_windows;
DROP TABLE IF EXISTS github_orgs;
DROP TABLE IF EXISTS github_repos;
DROP TABLE IF EXISTS github_profiles;
