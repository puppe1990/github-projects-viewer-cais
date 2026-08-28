package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
)

func canon(login string) string {
	return strings.ToLower(strings.TrimSpace(login))
}

func marshalDays(days []catalog.TrafficPoint) ([]byte, error) {
	if days == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(days)
}

func (s *SQLiteStore) SaveProfile(profile catalog.Profile, fetchedAt time.Time) error {
	payload, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO github_profiles (login, payload, fetched_at) VALUES (?, ?, ?)
		 ON CONFLICT(login) DO UPDATE SET payload = excluded.payload, fetched_at = excluded.fetched_at`,
		canon(profile.Login), string(payload), fetchedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LoadProfile(login string) (catalog.Profile, bool, error) {
	var payload string
	err := s.db.QueryRow(`SELECT payload FROM github_profiles WHERE login = ?`, canon(login)).Scan(&payload)
	if err == sql.ErrNoRows {
		return catalog.Profile{}, false, nil
	}
	if err != nil {
		return catalog.Profile{}, false, fmt.Errorf("load profile: %w", err)
	}
	var profile catalog.Profile
	if err := json.Unmarshal([]byte(payload), &profile); err != nil {
		return catalog.Profile{}, false, fmt.Errorf("decode profile: %w", err)
	}
	return profile, true, nil
}

func (s *SQLiteStore) ProfileFetchedAt(login string) (time.Time, error) {
	var raw string
	err := s.db.QueryRow(`SELECT fetched_at FROM github_profiles WHERE login = ?`, canon(login)).Scan(&raw)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("profile fetched_at: %w", err)
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse fetched_at: %w", err)
	}
	return t, nil
}

func (s *SQLiteStore) SaveRepos(owner, source string, repos []catalog.Repo, fetchedAt time.Time) error {
	ownerKey := canon(owner)
	tx, err := s.db.Raw().Begin()
	if err != nil {
		return fmt.Errorf("save repos begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM github_repos WHERE owner_login = ? AND source = ?`, ownerKey, source); err != nil {
		return fmt.Errorf("clear repos: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO github_repos (owner_login, name, source, payload, fetched_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare repos: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	stamp := fetchedAt.UTC().Format(time.RFC3339)
	for _, repo := range repos {
		payload, err := json.Marshal(repo)
		if err != nil {
			return fmt.Errorf("marshal repo: %w", err)
		}
		if _, err := stmt.Exec(ownerKey, repo.Name, source, string(payload), stamp); err != nil {
			return fmt.Errorf("insert repo %s: %w", repo.Name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save repos commit: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LoadRepos(owner, source string) ([]catalog.Repo, error) {
	rows, err := s.db.Query(
		`SELECT payload FROM github_repos WHERE owner_login = ? AND source = ? ORDER BY name`,
		canon(owner), source,
	)
	if err != nil {
		return nil, fmt.Errorf("load repos: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var repos []catalog.Repo
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan repo: %w", err)
		}
		var repo catalog.Repo
		if err := json.Unmarshal([]byte(payload), &repo); err != nil {
			return nil, fmt.Errorf("decode repo: %w", err)
		}
		repos = append(repos, repo)
	}
	return repos, rows.Err()
}

func (s *SQLiteStore) SaveOrgs(userLogin string, orgs []catalog.Org) error {
	userKey := canon(userLogin)
	tx, err := s.db.Raw().Begin()
	if err != nil {
		return fmt.Errorf("save orgs begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM github_orgs WHERE user_login = ?`, userKey); err != nil {
		return fmt.Errorf("clear orgs: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO github_orgs (user_login, org_login, payload) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare orgs: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	for _, org := range orgs {
		payload, err := json.Marshal(org)
		if err != nil {
			return fmt.Errorf("marshal org: %w", err)
		}
		if _, err := stmt.Exec(userKey, canon(org.Login), string(payload)); err != nil {
			return fmt.Errorf("insert org %s: %w", org.Login, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save orgs commit: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LoadOrgs(userLogin string) ([]catalog.Org, error) {
	rows, err := s.db.Query(`SELECT payload FROM github_orgs WHERE user_login = ? ORDER BY org_login`, canon(userLogin))
	if err != nil {
		return nil, fmt.Errorf("load orgs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var orgs []catalog.Org
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan org: %w", err)
		}
		var org catalog.Org
		if err := json.Unmarshal([]byte(payload), &org); err != nil {
			return nil, fmt.Errorf("decode org: %w", err)
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (s *SQLiteStore) SaveTraffic(owner, repo string, traffic catalog.Traffic, capturedAt time.Time) error {
	viewDays, err := marshalDays(traffic.ViewsByDay)
	if err != nil {
		return fmt.Errorf("marshal view days: %w", err)
	}
	cloneDays, err := marshalDays(traffic.ClonesByDay)
	if err != nil {
		return fmt.Errorf("marshal clone days: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO traffic_windows (owner_login, repo_name, views, view_uniques, clones, clone_uniques, captured_at, view_days, clone_days)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(owner_login, repo_name) DO UPDATE SET
		   views = excluded.views,
		   view_uniques = excluded.view_uniques,
		   clones = excluded.clones,
		   clone_uniques = excluded.clone_uniques,
		   captured_at = excluded.captured_at,
		   view_days = excluded.view_days,
		   clone_days = excluded.clone_days`,
		canon(owner), repo, traffic.Views, traffic.ViewUniques, traffic.Clones, traffic.CloneUniques,
		capturedAt.UTC().Format(time.RFC3339), string(viewDays), string(cloneDays),
	)
	if err != nil {
		return fmt.Errorf("save traffic: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LoadTraffic(owner, repo string) (catalog.Traffic, bool, error) {
	var tr catalog.Traffic
	var viewDays, cloneDays string
	err := s.db.QueryRow(
		`SELECT views, view_uniques, clones, clone_uniques, view_days, clone_days FROM traffic_windows WHERE owner_login = ? AND repo_name = ?`,
		canon(owner), repo,
	).Scan(&tr.Views, &tr.ViewUniques, &tr.Clones, &tr.CloneUniques, &viewDays, &cloneDays)
	if err == sql.ErrNoRows {
		return catalog.Traffic{}, false, nil
	}
	if err != nil {
		return catalog.Traffic{}, false, fmt.Errorf("load traffic: %w", err)
	}
	if viewDays != "" {
		if err := json.Unmarshal([]byte(viewDays), &tr.ViewsByDay); err != nil {
			return catalog.Traffic{}, false, fmt.Errorf("decode view days: %w", err)
		}
	}
	if cloneDays != "" {
		if err := json.Unmarshal([]byte(cloneDays), &tr.ClonesByDay); err != nil {
			return catalog.Traffic{}, false, fmt.Errorf("decode clone days: %w", err)
		}
	}
	tr.Available = true
	return tr, true, nil
}

func (s *SQLiteStore) TouchWatch(login string) error {
	_, err := s.db.Exec(
		`INSERT INTO catalog_watches (login, last_viewed_at) VALUES (?, ?)
		 ON CONFLICT(login) DO UPDATE SET last_viewed_at = excluded.last_viewed_at`,
		canon(login), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("touch watch: %w", err)
	}
	return nil
}

func (s *SQLiteStore) WatchedLogins() ([]string, error) {
	rows, err := s.db.Query(`SELECT login FROM catalog_watches ORDER BY last_viewed_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("watched logins: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var logins []string
	for rows.Next() {
		var login string
		if err := rows.Scan(&login); err != nil {
			return nil, err
		}
		logins = append(logins, login)
	}
	return logins, rows.Err()
}

func (s *SQLiteStore) CachedRepoRefs() ([]catalog.RepoRef, error) {
	rows, err := s.db.Query(`SELECT DISTINCT owner_login, name FROM github_repos ORDER BY owner_login, name`)
	if err != nil {
		return nil, fmt.Errorf("cached repo refs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var refs []catalog.RepoRef
	for rows.Next() {
		var ref catalog.RepoRef
		if err := rows.Scan(&ref.Owner, &ref.Name); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}
