package catalog

import (
	"context"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

type GitHub interface {
	HasToken() bool
	Me(ctx context.Context) (githubapi.User, error)
	User(ctx context.Context, login string) (githubapi.User, error)
	UserRepos(ctx context.Context, login string) ([]githubapi.Repo, error)
	UserOrgs(ctx context.Context, login string) ([]githubapi.Org, error)
	OrgRepos(ctx context.Context, org string) ([]githubapi.Repo, error)
	Traffic(ctx context.Context, owner, repo string) (githubapi.Traffic, error)
}

type Cache interface {
	SaveProfile(profile Profile, fetchedAt time.Time) error
	LoadProfile(login string) (Profile, bool, error)
	ProfileFetchedAt(login string) (time.Time, error)
	SaveRepos(owner, source string, repos []Repo, fetchedAt time.Time) error
	LoadRepos(owner, source string) ([]Repo, error)
	SaveOrgs(userLogin string, orgs []Org) error
	LoadOrgs(userLogin string) ([]Org, error)
	SaveTraffic(owner, repo string, traffic Traffic, capturedAt time.Time) error
	LoadTraffic(owner, repo string) (Traffic, bool, error)
	TouchWatch(login string) error
	WatchedLogins() ([]string, error)
	CachedRepoRefs() ([]RepoRef, error)
}

type Queue interface {
	EnqueueTraffic(login string) error
	EnqueueRefresh(login string) error
}
