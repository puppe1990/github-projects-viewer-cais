package catalog

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

type Loader struct {
	GitHub   GitHub
	Cache    Cache
	Queue    Queue
	Now      func() time.Time
	FreshFor time.Duration
}

func (l *Loader) now() time.Time {
	if l.Now != nil {
		return l.Now()
	}
	return time.Now().UTC()
}

func (l *Loader) freshFor() time.Duration {
	if l.FreshFor > 0 {
		return l.FreshFor
	}
	return 30 * time.Minute
}

func (l *Loader) User(ctx context.Context, login string) (Snapshot, error) {
	if err := l.Cache.TouchWatch(login); err != nil {
		return Snapshot{}, err
	}
	if snap, ok, err := l.cachedUser(login); err != nil {
		return Snapshot{}, err
	} else if ok {
		return l.withTraffic(ctx, login, snap)
	}

	snap, err := l.fetchUser(ctx, login)
	if err == nil {
		return l.withTraffic(ctx, login, snap)
	}
	if errors.Is(err, githubapi.ErrNotFound) {
		return Snapshot{}, err
	}
	if snap, ok, cacheErr := l.cachedUserIgnoreFresh(login); cacheErr != nil {
		return Snapshot{}, cacheErr
	} else if ok {
		return snap, nil
	}
	return Snapshot{}, err
}

func (l *Loader) Org(ctx context.Context, userLogin, orgLogin string) (Snapshot, error) {
	userSnap, err := l.User(ctx, userLogin)
	if err != nil {
		return Snapshot{}, err
	}

	repos, err := l.GitHub.OrgRepos(ctx, orgLogin)
	if err != nil {
		cached, cacheErr := l.Cache.LoadRepos(orgLogin, SourceOrg)
		if cacheErr != nil {
			return Snapshot{}, cacheErr
		}
		if len(cached) > 0 {
			snap, snapErr := l.orgSnapshot(userSnap, orgLogin, cached)
			if snapErr != nil {
				return Snapshot{}, snapErr
			}
			return l.withTraffic(ctx, orgLogin, snap)
		}
		return Snapshot{}, err
	}
	mapped := ReposFromGitHub(repos)
	if err := l.Cache.SaveRepos(orgLogin, SourceOrg, mapped, l.now()); err != nil {
		return Snapshot{}, err
	}
	snap, err := l.orgSnapshot(userSnap, orgLogin, mapped)
	if err != nil {
		return Snapshot{}, err
	}
	return l.withTraffic(ctx, orgLogin, snap)
}

func (l *Loader) Refresh(ctx context.Context, login string) error {
	_, err := l.fetchUser(ctx, login)
	if err != nil {
		return err
	}
	if l.GitHub.HasToken() && l.Queue != nil {
		return l.Queue.EnqueueTraffic(login)
	}
	return nil
}

func (l *Loader) SnapshotTraffic(ctx context.Context, owner, repo string) error {
	tr, err := l.GitHub.Traffic(ctx, owner, repo)
	if errors.Is(err, githubapi.ErrForbidden) || errors.Is(err, githubapi.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return l.Cache.SaveTraffic(owner, repo, tr, l.now())
}

func (l *Loader) SnapshotLogin(ctx context.Context, login string) error {
	repos, err := l.Cache.LoadRepos(login, SourceUser)
	if err != nil {
		return err
	}
	orgs, err := l.Cache.LoadOrgs(login)
	if err != nil {
		return err
	}
	var first error
	for _, repo := range repos {
		if err := l.SnapshotTraffic(ctx, repo.OwnerLogin, repo.Name); err != nil && first == nil {
			first = err
		}
	}
	for _, org := range orgs {
		orgRepos, err := l.Cache.LoadRepos(org.Login, SourceOrg)
		if err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		for _, repo := range orgRepos {
			if err := l.SnapshotTraffic(ctx, repo.OwnerLogin, repo.Name); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}

func (l *Loader) fetchUser(ctx context.Context, login string) (Snapshot, error) {
	user, err := l.GitHub.User(ctx, login)
	if err != nil {
		return Snapshot{}, err
	}
	repos, err := l.GitHub.UserRepos(ctx, login)
	if err != nil {
		return Snapshot{}, err
	}
	orgs, err := l.GitHub.UserOrgs(ctx, login)
	if err != nil {
		orgs = nil
	}
	mapped := ReposFromGitHub(repos)
	now := l.now()
	if err := l.Cache.SaveProfile(user, now); err != nil {
		return Snapshot{}, err
	}
	if err := l.Cache.SaveRepos(user.Login, SourceUser, mapped, now); err != nil {
		return Snapshot{}, err
	}
	if err := l.Cache.SaveOrgs(user.Login, orgs); err != nil {
		return Snapshot{}, err
	}
	return l.userSnapshot(user, orgs, mapped)
}

func (l *Loader) cachedUser(login string) (Snapshot, bool, error) {
	fetched, err := l.Cache.ProfileFetchedAt(login)
	if err != nil {
		return Snapshot{}, false, err
	}
	if fetched.IsZero() || l.now().Sub(fetched) >= l.freshFor() {
		return Snapshot{}, false, nil
	}
	return l.cachedUserIgnoreFresh(login)
}

func (l *Loader) cachedUserIgnoreFresh(login string) (Snapshot, bool, error) {
	profile, ok, err := l.Cache.LoadProfile(login)
	if err != nil || !ok {
		return Snapshot{}, ok, err
	}
	repos, err := l.Cache.LoadRepos(profile.Login, SourceUser)
	if err != nil {
		return Snapshot{}, false, err
	}
	orgs, err := l.Cache.LoadOrgs(profile.Login)
	if err != nil {
		return Snapshot{}, false, err
	}
	snap, err := l.userSnapshot(profile, orgs, repos)
	return snap, true, err
}

func (l *Loader) userSnapshot(profile Profile, orgs []Org, repos []Repo) (Snapshot, error) {
	repos, err := l.attachTraffic(repos)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Profile:     profile,
		Orgs:        orgs,
		Repos:       repos,
		Source:      SourceUser,
		SourceLogin: profile.Login,
	}, nil
}

func (l *Loader) orgSnapshot(userSnap Snapshot, orgLogin string, repos []Repo) (Snapshot, error) {
	repos, err := l.attachTraffic(repos)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		Profile:     userSnap.Profile,
		Orgs:        userSnap.Orgs,
		Repos:       repos,
		Source:      SourceOrg,
		SourceLogin: orgLogin,
	}, nil
}

func (l *Loader) withTraffic(ctx context.Context, login string, snap Snapshot) (Snapshot, error) {
	if l.GitHub.HasToken() {
		if err := l.fillMissingTraffic(ctx, snap.Repos); err != nil {
			return Snapshot{}, err
		}
		if l.Queue != nil {
			_ = l.Queue.EnqueueTraffic(login)
		}
	}
	repos, err := l.attachTraffic(snap.Repos)
	if err != nil {
		return Snapshot{}, err
	}
	snap.Repos = repos
	return snap, nil
}

func (l *Loader) fillMissingTraffic(ctx context.Context, repos []Repo) error {
	var missing []Repo
	for _, repo := range repos {
		if repo.OwnerLogin == "" || repo.Name == "" {
			continue
		}
		_, ok, err := l.Cache.LoadTraffic(repo.OwnerLogin, repo.Name)
		if err != nil {
			return err
		}
		if !ok {
			missing = append(missing, repo)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	workers := 6
	if len(missing) < workers {
		workers = len(missing)
	}
	jobs := make(chan Repo)
	errCh := make(chan error, len(missing))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				if err := l.SnapshotTraffic(ctx, repo.OwnerLogin, repo.Name); err != nil {
					errCh <- err
				}
			}
		}()
	}
	for _, repo := range missing {
		jobs <- repo
	}
	close(jobs)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) attachTraffic(repos []Repo) ([]Repo, error) {
	out := make([]Repo, len(repos))
	copy(out, repos)
	for i, repo := range out {
		owner := repo.OwnerLogin
		if owner == "" {
			continue
		}
		tr, ok, err := l.Cache.LoadTraffic(owner, repo.Name)
		if err != nil {
			return nil, fmt.Errorf("attach traffic %s/%s: %w", owner, repo.Name, err)
		}
		if ok {
			out[i].Traffic = tr
		}
	}
	return out, nil
}
