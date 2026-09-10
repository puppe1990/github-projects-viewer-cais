package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

// deadClientWriter reproduces a client that goes away mid-response: net/http
// records the implicit 200 when the first body chunk is written, then every
// write fails because the connection is gone.
type deadClientWriter struct {
	header   http.Header
	statuses []int
}

func newDeadClientWriter() *deadClientWriter {
	return &deadClientWriter{header: http.Header{}}
}

func (w *deadClientWriter) Header() http.Header { return w.header }

func (w *deadClientWriter) WriteHeader(code int) {
	w.statuses = append(w.statuses, code)
}

func (w *deadClientWriter) Write(p []byte) (int, error) {
	if len(w.statuses) == 0 {
		w.statuses = append(w.statuses, http.StatusOK)
	}
	return 0, errors.New("write: connection reset by peer")
}

type homeGitHub struct {
	user     githubapi.User
	repos    []githubapi.Repo
	orgs     []githubapi.Org
	orgRepos []githubapi.Repo
	userErr  error
	token    bool
	traffic  githubapi.Traffic
}

func (h homeGitHub) HasToken() bool { return h.token }
func (h homeGitHub) Me(context.Context) (githubapi.User, error) {
	if h.user.Login == "" {
		return githubapi.User{}, h.userErr
	}
	return h.user, h.userErr
}
func (h homeGitHub) User(context.Context, string) (githubapi.User, error) {
	return h.user, h.userErr
}
func (h homeGitHub) UserRepos(context.Context, string) ([]githubapi.Repo, error) {
	return h.repos, nil
}
func (h homeGitHub) UserOrgs(context.Context, string) ([]githubapi.Org, error) {
	return h.orgs, nil
}
func (h homeGitHub) OrgRepos(context.Context, string) ([]githubapi.Repo, error) {
	return h.orgRepos, nil
}
func (h homeGitHub) Traffic(context.Context, string, string) (githubapi.Traffic, error) {
	if h.traffic.Available {
		return h.traffic, nil
	}
	return githubapi.Traffic{}, githubapi.ErrForbidden
}

func newHomeHandler(t *testing.T, gh catalog.GitHub) *HomeHandler {
	t.Helper()
	s := setupTestStore(t)
	loader := &catalog.Loader{
		GitHub:   gh,
		Cache:    s,
		Now:      func() time.Time { return time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC) },
		FreshFor: time.Hour,
	}
	return NewHomeHandler(testSite(), setupTestInertia(t), loader)
}

func TestHomeHandler_IndexRedirectsToAuthenticatedUser(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{
		token: true,
		user:  githubapi.User{Login: "puppe1990", Name: "Matheus"},
		repos: []githubapi.Repo{{Name: "cais", OwnerLogin: "puppe1990"}},
	})
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusSeeOther)
	}
	if loc := rr.Header().Get("Location"); loc != "/u/puppe1990/all" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestHomeHandler_Returns200(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{})
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHomeHandler_InertiaComponent(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{})
	rr := httptest.NewRecorder()
	h.Index(rr, inertiaRequest(http.MethodGet, "/", nil))
	assertInertiaComponent(t, rr, "Home")
}

func TestHomeHandler_ShowUserCatalog(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{
		user:  githubapi.User{Login: "octocat", Name: "The Octocat", PublicRepos: 8},
		repos: []githubapi.Repo{{Name: "hello-world", Stars: 10, OwnerLogin: "octocat"}},
		orgs:  []githubapi.Org{{Login: "github"}},
	})
	rr := httptest.NewRecorder()
	h.Show(rr, inertiaRequest(http.MethodGet, "/u/octocat", nil), "octocat")
	assertInertiaComponent(t, rr, "Home")
	if assertInertiaProp(t, rr, "populated") != true {
		t.Fatal("expected populated")
	}
	profile := assertInertiaProp(t, rr, "profile").(map[string]any)
	if profile["login"] != "octocat" || profile["name"] != "The Octocat" {
		t.Fatalf("profile = %#v", profile)
	}
	repos := assertInertiaProp(t, rr, "repos").([]any)
	if len(repos) != 1 {
		t.Fatalf("repos = %#v", repos)
	}
}

func TestHomeHandler_ShowNotFound(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{userErr: githubapi.ErrNotFound})
	rr := httptest.NewRecorder()
	h.Show(rr, inertiaRequest(http.MethodGet, "/u/missing", nil), "missing")
	msg := assertInertiaProp(t, rr, "error").(string)
	if msg == "" {
		t.Fatal("expected error message")
	}
	if assertInertiaProp(t, rr, "populated") != false {
		t.Fatal("populated should be false")
	}
}

func TestHomeHandler_ShowAll(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{
		token:    true,
		user:     githubapi.User{Login: "puppe1990", Name: "Matheus"},
		repos:    []githubapi.Repo{{Name: "cais", FullName: "puppe1990/cais", OwnerLogin: "puppe1990"}},
		orgs:     []githubapi.Org{{Login: "hidden-org"}},
		orgRepos: []githubapi.Repo{{Name: "private-app", FullName: "hidden-org/private-app", OwnerLogin: "hidden-org", Private: true}},
	})
	rr := httptest.NewRecorder()
	h.ShowAll(rr, inertiaRequest(http.MethodGet, "/u/puppe1990/all", nil), "puppe1990")
	source := assertInertiaProp(t, rr, "source").(map[string]any)
	if source["type"] != catalog.SourceAll {
		t.Fatalf("source = %#v", source)
	}
	repos := assertInertiaProp(t, rr, "repos").([]any)
	if len(repos) != 2 {
		t.Fatalf("repos = %#v", repos)
	}
}

func TestHomeHandler_ShowOrg(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{
		user:     githubapi.User{Login: "octocat", Name: "The Octocat"},
		orgs:     []githubapi.Org{{Login: "github"}},
		orgRepos: []githubapi.Repo{{Name: "linguist", OwnerLogin: "github", Stars: 4}},
	})
	rr := httptest.NewRecorder()
	h.ShowOrg(rr, inertiaRequest(http.MethodGet, "/u/octocat/orgs/github", nil), "octocat", "github")
	source := assertInertiaProp(t, rr, "source").(map[string]any)
	if source["type"] != catalog.SourceOrg || source["login"] != "github" {
		t.Fatalf("source = %#v", source)
	}
	repos := assertInertiaProp(t, rr, "repos").([]any)
	if len(repos) != 1 {
		t.Fatalf("repos = %#v", repos)
	}
}

func TestHomeHandler_RepoTrafficJSON(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{
		token: true,
		traffic: githubapi.Traffic{
			Views: 17, ViewUniques: 3, Clones: 342, CloneUniques: 104, Available: true,
			ViewsByDay: []githubapi.TrafficPoint{{Timestamp: "2026-08-27T00:00:00Z", Count: 7, Uniques: 3}},
		},
	})
	rr := httptest.NewRecorder()
	h.RepoTraffic(rr, httptest.NewRequest(http.MethodGet, "/traffic/puppe1990/cais", nil), "puppe1990", "cais")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"views":17`) {
		t.Fatalf("body = %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"views_by_day"`) {
		t.Fatalf("missing daily series: %s", rr.Body.String())
	}
}

func TestHomeHandler_ContentType(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{})
	rr := httptest.NewRecorder()
	h.Index(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rr.Header().Get("Content-Type"); got == "" {
		t.Errorf("Content-Type = %q", got)
	}
}

func TestHomeHandler_AbortedClientDoesNotRewriteHeader(t *testing.T) {
	h := newHomeHandler(t, homeGitHub{})
	w := newDeadClientWriter()
	h.Index(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if len(w.statuses) != 1 || w.statuses[0] != http.StatusOK {
		t.Fatalf("statuses = %v, want [200]: an aborted client already has a response in flight, so appending an error page only corrupts it and logs a phantom 500", w.statuses)
	}
}
