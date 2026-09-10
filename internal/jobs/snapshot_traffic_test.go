package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
)

type stubGitHub struct {
	traffic    githubapi.Traffic
	trafficErr error
	user       githubapi.User
	userErr    error
	repos      []githubapi.Repo
	userCalls  int
}

func (s *stubGitHub) HasToken() bool { return true }
func (s *stubGitHub) Me(context.Context) (githubapi.User, error) {
	return s.user, nil
}
func (s *stubGitHub) User(context.Context, string) (githubapi.User, error) {
	s.userCalls++
	return s.user, s.userErr
}
func (s *stubGitHub) UserRepos(context.Context, string) ([]githubapi.Repo, error) {
	return s.repos, nil
}
func (s *stubGitHub) UserOrgs(context.Context, string) ([]githubapi.Org, error) {
	return nil, nil
}
func (s *stubGitHub) OrgRepos(context.Context, string) ([]githubapi.Repo, error) {
	return nil, nil
}
func (s *stubGitHub) Traffic(context.Context, string, string) (githubapi.Traffic, error) {
	return s.traffic, s.trafficErr
}

func testJobLoader(t *testing.T, gh catalog.GitHub) (*catalog.Loader, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return &catalog.Loader{
		GitHub:   gh,
		Cache:    s,
		Now:      func() time.Time { return time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC) },
		FreshFor: time.Hour,
	}, s
}

func TestPerformSnapshotTraffic_SavesWindow(t *testing.T) {
	loader, s := testJobLoader(t, &stubGitHub{
		traffic: githubapi.Traffic{Views: 42, Clones: 7, Available: true},
	})
	h := PerformSnapshotTraffic(loader)
	if err := h(context.Background(), []byte(`{"owner":"octocat","repo":"hello-world"}`)); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadTraffic("octocat", "hello-world")
	if err != nil || !ok || got.Views != 42 {
		t.Fatalf("ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestPerformSnapshotTraffic_LoginUsesCachedRepos(t *testing.T) {
	loader, s := testJobLoader(t, &stubGitHub{
		traffic: githubapi.Traffic{Views: 3, Available: true},
	})
	if err := s.SaveRepos("octocat", catalog.SourceUser, []catalog.Repo{{Name: "hello-world", OwnerLogin: "octocat"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	h := PerformSnapshotTraffic(loader)
	if err := h(context.Background(), []byte(`{"login":"octocat"}`)); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadTraffic("octocat", "hello-world")
	if err != nil || !ok || got.Views != 3 {
		t.Fatalf("ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestPerformSnapshotTraffic_RateLimitedIsSkipped(t *testing.T) {
	loader, s := testJobLoader(t, &stubGitHub{trafficErr: githubapi.ErrRateLimited})
	if err := s.SaveRepos("octocat", catalog.SourceUser, []catalog.Repo{{Name: "hello-world", OwnerLogin: "octocat"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	h := PerformSnapshotTraffic(loader)
	if err := h(context.Background(), []byte(`{"login":"octocat"}`)); err != nil {
		t.Fatalf("rate limit must not fail the job: %v", err)
	}
}

func TestPerformRefreshCatalog_FetchesUser(t *testing.T) {
	loader, s := testJobLoader(t, &stubGitHub{
		user:  githubapi.User{Login: "octocat", Name: "The Octocat"},
		repos: []githubapi.Repo{{Name: "hello-world", OwnerLogin: "octocat", Stars: 9}},
	})
	h := PerformRefreshCatalog(loader)
	if err := h(context.Background(), []byte(`{"login":"octocat"}`)); err != nil {
		t.Fatal(err)
	}
	profile, ok, err := s.LoadProfile("octocat")
	if err != nil || !ok || profile.Name != "The Octocat" {
		t.Fatalf("ok=%v err=%v profile=%+v", ok, err, profile)
	}
}

func TestPerformRefreshCatalog_StopsOnFirstError(t *testing.T) {
	gh := &stubGitHub{userErr: errors.New("github: HTTP 502")}
	loader, s := testJobLoader(t, gh)
	for _, login := range []string{"alpha", "beta", "gamma"} {
		if err := s.TouchWatch(login); err != nil {
			t.Fatal(err)
		}
	}
	h := PerformRefreshCatalog(loader)
	if err := h(context.Background(), []byte(`{}`)); err == nil {
		t.Fatal("expected the fetch error to surface")
	}
	if gh.userCalls != 1 {
		t.Fatalf("user calls = %d, want 1: a failing login must not make the job walk the whole watch list", gh.userCalls)
	}
}

func TestPerformRefreshCatalog_RateLimitedIsSkipped(t *testing.T) {
	loader, s := testJobLoader(t, &stubGitHub{userErr: githubapi.ErrRateLimited})
	if err := s.TouchWatch("octocat"); err != nil {
		t.Fatal(err)
	}
	h := PerformRefreshCatalog(loader)
	if err := h(context.Background(), []byte(`{}`)); err != nil {
		t.Fatalf("rate limit must not fail the job: %v", err)
	}
}
