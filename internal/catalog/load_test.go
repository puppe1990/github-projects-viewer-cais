package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

type fakeGitHub struct {
	user       githubapi.User
	repos      []githubapi.Repo
	orgs       []githubapi.Org
	orgRepos   []githubapi.Repo
	traffic    githubapi.Traffic
	userErr    error
	reposErr   error
	trafficErr error
	token      bool
	users      int
}

func (f *fakeGitHub) HasToken() bool { return f.token }

func (f *fakeGitHub) Me(context.Context) (githubapi.User, error) {
	return f.user, f.userErr
}

func (f *fakeGitHub) User(context.Context, string) (githubapi.User, error) {
	f.users++
	return f.user, f.userErr
}
func (f *fakeGitHub) UserRepos(context.Context, string) ([]githubapi.Repo, error) {
	return f.repos, f.reposErr
}
func (f *fakeGitHub) UserOrgs(context.Context, string) ([]githubapi.Org, error) {
	return f.orgs, nil
}
func (f *fakeGitHub) OrgRepos(context.Context, string) ([]githubapi.Repo, error) {
	return f.orgRepos, nil
}
func (f *fakeGitHub) Traffic(context.Context, string, string) (githubapi.Traffic, error) {
	return f.traffic, f.trafficErr
}

type fakeQueue struct {
	traffic []string
	refresh []string
}

func (q *fakeQueue) EnqueueTraffic(login string) error {
	q.traffic = append(q.traffic, login)
	return nil
}
func (q *fakeQueue) EnqueueRefresh(login string) error {
	q.refresh = append(q.refresh, login)
	return nil
}

type memCache struct {
	profiles map[string]Profile
	fetched  map[string]time.Time
	repos    map[string][]Repo
	orgs     map[string][]Org
	traffic  map[string]Traffic
	watches  []string
}

func newMemCache() *memCache {
	return &memCache{
		profiles: map[string]Profile{},
		fetched:  map[string]time.Time{},
		repos:    map[string][]Repo{},
		orgs:     map[string][]Org{},
		traffic:  map[string]Traffic{},
	}
}

func keyLogin(login string) string { return strings.ToLower(strings.TrimSpace(login)) }

func (m *memCache) SaveProfile(profile Profile, fetchedAt time.Time) error {
	k := keyLogin(profile.Login)
	m.profiles[k] = profile
	m.fetched[k] = fetchedAt
	return nil
}
func (m *memCache) LoadProfile(login string) (Profile, bool, error) {
	p, ok := m.profiles[keyLogin(login)]
	return p, ok, nil
}
func (m *memCache) ProfileFetchedAt(login string) (time.Time, error) {
	return m.fetched[keyLogin(login)], nil
}
func (m *memCache) SaveRepos(owner, source string, repos []Repo, _ time.Time) error {
	m.repos[keyLogin(owner)+"|"+source] = repos
	return nil
}
func (m *memCache) LoadRepos(owner, source string) ([]Repo, error) {
	return m.repos[keyLogin(owner)+"|"+source], nil
}
func (m *memCache) SaveOrgs(userLogin string, orgs []Org) error {
	m.orgs[keyLogin(userLogin)] = orgs
	return nil
}
func (m *memCache) LoadOrgs(userLogin string) ([]Org, error) {
	return m.orgs[keyLogin(userLogin)], nil
}
func (m *memCache) SaveTraffic(owner, repo string, traffic Traffic, _ time.Time) error {
	m.traffic[keyLogin(owner)+"/"+repo] = traffic
	return nil
}
func (m *memCache) LoadTraffic(owner, repo string) (Traffic, bool, error) {
	tr, ok := m.traffic[keyLogin(owner)+"/"+repo]
	return tr, ok, nil
}
func (m *memCache) TouchWatch(login string) error {
	m.watches = append(m.watches, keyLogin(login))
	return nil
}
func (m *memCache) WatchedLogins() ([]string, error) { return m.watches, nil }
func (m *memCache) CachedRepoRefs() ([]RepoRef, error) {
	return nil, nil
}

func testLoader(t *testing.T, gh *fakeGitHub, q *fakeQueue) (*Loader, *memCache) {
	t.Helper()
	cache := newMemCache()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	return &Loader{GitHub: gh, Cache: cache, Queue: q, Now: func() time.Time { return now }, FreshFor: 30 * time.Minute}, cache
}

func TestLoader_User_FetchesAndCaches(t *testing.T) {
	gh := &fakeGitHub{
		user:  githubapi.User{Login: "octocat", Name: "The Octocat", PublicRepos: 8},
		repos: []githubapi.Repo{{Name: "hello-world", OwnerLogin: "octocat", Stars: 10}},
		orgs:  []githubapi.Org{{Login: "github"}},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})

	snap, err := loader.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Profile.Name != "The Octocat" || len(snap.Repos) != 1 || len(snap.Orgs) != 1 {
		t.Fatalf("snapshot = %+v", snap)
	}
	if snap.Source != SourceUser || snap.SourceLogin != "octocat" {
		t.Fatalf("source = %s %s", snap.Source, snap.SourceLogin)
	}

	gh.users = 0
	gh.user.Name = "changed"
	snap, err = loader.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if gh.users != 0 {
		t.Fatal("expected cache hit")
	}
	if snap.Profile.Name != "The Octocat" {
		t.Fatalf("cached name = %s", snap.Profile.Name)
	}
}

func TestLoader_User_FallsBackToCache(t *testing.T) {
	gh := &fakeGitHub{
		user:  githubapi.User{Login: "octocat", Name: "The Octocat"},
		repos: []githubapi.Repo{{Name: "hello-world", OwnerLogin: "octocat"}},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	if _, err := loader.User(context.Background(), "octocat"); err != nil {
		t.Fatal(err)
	}

	loader.Now = func() time.Time { return time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC) }
	gh.userErr = errors.New("network down")
	snap, err := loader.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Profile.Name != "The Octocat" {
		t.Fatalf("stale = %+v", snap.Profile)
	}
}

func TestLoader_User_NotFound(t *testing.T) {
	gh := &fakeGitHub{userErr: githubapi.ErrNotFound}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	_, err := loader.User(context.Background(), "missing")
	if err != githubapi.ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestLoader_User_SnapshotsTrafficWhenToken(t *testing.T) {
	gh := &fakeGitHub{
		token:   true,
		user:    githubapi.User{Login: "puppe1990", Name: "Matheus"},
		repos:   []githubapi.Repo{{Name: "cais", OwnerLogin: "puppe1990"}},
		traffic: githubapi.Traffic{Views: 17, ViewUniques: 3, Clones: 4, CloneUniques: 2, Available: true},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})

	snap, err := loader.User(context.Background(), "puppe1990")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Repos) != 1 {
		t.Fatalf("repos = %+v", snap.Repos)
	}
	tr := snap.Repos[0].Traffic
	if !tr.Available || tr.Views != 17 || tr.Clones != 4 {
		t.Fatalf("expected live traffic on first load, got %+v", tr)
	}
}

func TestLoader_AttachTraffic(t *testing.T) {
	gh := &fakeGitHub{
		user:  githubapi.User{Login: "octocat"},
		repos: []githubapi.Repo{{Name: "hello-world", OwnerLogin: "octocat"}},
	}
	loader, s := testLoader(t, gh, &fakeQueue{})
	if err := s.SaveTraffic("octocat", "hello-world", Traffic{Views: 99, Available: true}, time.Now()); err != nil {
		t.Fatal(err)
	}
	snap, err := loader.User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Repos[0].Traffic.Available || snap.Repos[0].Traffic.Views != 99 {
		t.Fatalf("traffic = %+v", snap.Repos[0].Traffic)
	}
}

func TestLoader_SnapshotTraffic_SkipsForbidden(t *testing.T) {
	gh := &fakeGitHub{trafficErr: githubapi.ErrForbidden}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	if err := loader.SnapshotTraffic(context.Background(), "octocat", "hello-world"); err != nil {
		t.Fatal(err)
	}
}

func TestLoader_SnapshotTraffic_Persists(t *testing.T) {
	gh := &fakeGitHub{traffic: githubapi.Traffic{Views: 5, Clones: 2, Available: true}}
	loader, s := testLoader(t, gh, &fakeQueue{})
	if err := loader.SnapshotTraffic(context.Background(), "octocat", "hello-world"); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadTraffic("octocat", "hello-world")
	if err != nil || !ok || got.Views != 5 || got.Clones != 2 {
		t.Fatalf("ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestLoader_User_SkipsCacheWhenAuthenticatedSelf(t *testing.T) {
	gh := &fakeGitHub{
		token: true,
		user:  githubapi.User{Login: "puppe1990", Name: "Matheus"},
		repos: []githubapi.Repo{{Name: "cais", OwnerLogin: "puppe1990"}},
		orgs:  []githubapi.Org{{Login: "purchasestore"}},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	if _, err := loader.User(context.Background(), "puppe1990"); err != nil {
		t.Fatal(err)
	}
	gh.orgs = []githubapi.Org{{Login: "purchasestore"}, {Login: "hidden-org"}}
	gh.users = 0
	snap, err := loader.User(context.Background(), "puppe1990")
	if err != nil {
		t.Fatal(err)
	}
	if gh.users != 1 {
		t.Fatalf("users = %d, want refetch for authenticated self", gh.users)
	}
	if len(snap.Orgs) != 2 {
		t.Fatalf("orgs = %+v", snap.Orgs)
	}
}

func TestLoader_All_CombinesPersonalAndOrgRepos(t *testing.T) {
	gh := &fakeGitHub{
		token:    true,
		user:     githubapi.User{Login: "puppe1990", Name: "Matheus"},
		repos:    []githubapi.Repo{{Name: "cais", FullName: "puppe1990/cais", OwnerLogin: "puppe1990"}},
		orgs:     []githubapi.Org{{Login: "hidden-org"}},
		orgRepos: []githubapi.Repo{{Name: "private-app", FullName: "hidden-org/private-app", OwnerLogin: "hidden-org", Private: true}},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	snap, err := loader.All(context.Background(), "puppe1990")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Source != SourceAll {
		t.Fatalf("source = %s", snap.Source)
	}
	if len(snap.Repos) != 2 {
		t.Fatalf("repos = %+v", snap.Repos)
	}
}

func TestLoader_Org(t *testing.T) {
	gh := &fakeGitHub{
		user:     githubapi.User{Login: "octocat", Name: "The Octocat"},
		orgRepos: []githubapi.Repo{{Name: "linguist", OwnerLogin: "github", Stars: 20}},
		orgs:     []githubapi.Org{{Login: "github"}},
	}
	loader, _ := testLoader(t, gh, &fakeQueue{})
	if _, err := loader.User(context.Background(), "octocat"); err != nil {
		t.Fatal(err)
	}
	snap, err := loader.Org(context.Background(), "octocat", "github")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Source != SourceOrg || snap.SourceLogin != "github" || len(snap.Repos) != 1 {
		t.Fatalf("org snap = %+v", snap)
	}
}
