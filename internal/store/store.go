package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/session"
	caissqlite "github.com/puppe1990/cais/pkg/cais/sqlite"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	_ "modernc.org/sqlite"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
)

type Store interface {
	Sessions() session.Store
	Ping() error
	DB() *sql.DB
	Close() error
	SaveProfile(profile catalog.Profile, fetchedAt time.Time) error
	LoadProfile(login string) (catalog.Profile, bool, error)
	ProfileFetchedAt(login string) (time.Time, error)
	SaveRepos(owner, source string, repos []catalog.Repo, fetchedAt time.Time) error
	LoadRepos(owner, source string) ([]catalog.Repo, error)
	SaveOrgs(userLogin string, orgs []catalog.Org) error
	LoadOrgs(userLogin string) ([]catalog.Org, error)
	SaveTraffic(owner, repo string, traffic catalog.Traffic, capturedAt time.Time) error
	LoadTraffic(owner, repo string) (catalog.Traffic, bool, error)
	TouchWatch(login string) error
	WatchedLogins() ([]string, error)
	CachedRepoRefs() ([]catalog.RepoRef, error)
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := caissqlite.Configure(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}
	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := session.EnsureSQLiteSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sessions schema: %w", err)
	}

	cfg := sqllog.ConfigForEnv(env)
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	return &SQLiteStore{db: sqllog.Wrap(db, cfg)}, nil
}

func (s *SQLiteStore) Sessions() session.Store {
	return session.NewSQLiteStore(s.db.Raw())
}

func (s *SQLiteStore) Ping() error {
	return s.db.Raw().Ping()
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
