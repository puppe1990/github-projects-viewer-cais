package githubapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_Me_UsesUserEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"login": "puppe1990", "name": "Matheus"})
	}))
	t.Cleanup(srv.Close)

	user, err := New(srv.URL, "secret").Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Login != "puppe1990" {
		t.Fatalf("login = %s", user.Login)
	}
}

func TestClient_User_ParsesProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/octocat" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Accept = %s", r.Header.Get("Accept"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"login":        "octocat",
			"name":         "The Octocat",
			"bio":          "GitHub mascot",
			"avatar_url":   "https://avatars.githubusercontent.com/u/1",
			"html_url":     "https://github.com/octocat",
			"location":     "San Francisco",
			"company":      "@github",
			"blog":         "https://github.blog",
			"public_repos": 8,
			"followers":    4000,
			"following":    9,
		})
	}))
	t.Cleanup(srv.Close)

	user, err := New(srv.URL, "").User(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if user.Login != "octocat" || user.Name != "The Octocat" || user.PublicRepos != 8 {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestClient_User_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := New(srv.URL, "").User(context.Background(), "missing")
	if err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestClient_SendsBearerToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q", got)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_, _ = New(srv.URL, "secret").User(context.Background(), "x")
}

func TestClient_UserRepos_UsesAuthenticatedOwnerPath(t *testing.T) {
	var sawUserRepos, sawPublicRepos bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/user":
			_ = json.NewEncoder(w).Encode(map[string]any{"login": "puppe1990"})
		case r.URL.Path == "/user/repos":
			sawUserRepos = true
			if r.URL.Query().Get("affiliation") != "owner" {
				t.Errorf("affiliation = %s", r.URL.Query().Get("affiliation"))
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"name": "secret", "private": true, "owner": map[string]any{"login": "puppe1990"}, "stargazers_count": 0},
			})
		case strings.HasPrefix(r.URL.Path, "/users/"):
			sawPublicRepos = true
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	repos, err := New(srv.URL, "tok").UserRepos(context.Background(), "puppe1990")
	if err != nil {
		t.Fatal(err)
	}
	if !sawUserRepos || sawPublicRepos {
		t.Fatalf("authenticated=%v public=%v", sawUserRepos, sawPublicRepos)
	}
	if len(repos) != 1 || !repos[0].Private {
		t.Fatalf("repos = %+v", repos)
	}
}

func TestClient_UserOrgs_UsesAuthenticatedOrgs(t *testing.T) {
	var sawUserOrgs, sawPublicOrgs bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			_ = json.NewEncoder(w).Encode(map[string]any{"login": "puppe1990"})
		case "/user/orgs":
			sawUserOrgs = true
			_ = json.NewEncoder(w).Encode([]map[string]any{{"login": "hidden-org", "avatar_url": "https://example.com/o.png"}})
		case "/users/puppe1990/orgs":
			sawPublicOrgs = true
			_ = json.NewEncoder(w).Encode([]any{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	orgs, err := New(srv.URL, "tok").UserOrgs(context.Background(), "puppe1990")
	if err != nil {
		t.Fatal(err)
	}
	if !sawUserOrgs || sawPublicOrgs {
		t.Fatalf("authenticated=%v public=%v", sawUserOrgs, sawPublicOrgs)
	}
	if len(orgs) != 1 || orgs[0].Login != "hidden-org" {
		t.Fatalf("orgs = %+v", orgs)
	}
}

func TestClient_UserRepos_Paginates(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", `<`+srv.URL+r.URL.Path+`?page=2>; rel="next"`)
			_ = json.NewEncoder(w).Encode([]map[string]any{
				repoJSON("hello-world", 10),
			})
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			repoJSON("spoon-knife", 1),
		})
	}))
	t.Cleanup(srv.Close)

	repos, err := New(srv.URL, "").UserRepos(context.Background(), "octocat")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("len = %d, want 2", len(repos))
	}
	if repos[0].Name != "hello-world" || repos[0].Stars != 10 || repos[0].OwnerLogin != "octocat" {
		t.Fatalf("repo[0] = %+v", repos[0])
	}
}

func TestClient_Traffic_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Must have push access"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := New(srv.URL, "tok").Traffic(context.Background(), "octocat", "hello-world")
	if err != ErrForbidden {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestClient_Traffic_SumsViewsAndClones(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/traffic/views"):
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 1250, "uniques": 487})
		case strings.HasSuffix(r.URL.Path, "/traffic/clones"):
			_ = json.NewEncoder(w).Encode(map[string]any{"count": 340, "uniques": 128})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	tr, err := New(srv.URL, "tok").Traffic(context.Background(), "octocat", "hello-world")
	if err != nil {
		t.Fatal(err)
	}
	if !tr.Available || tr.Views != 1250 || tr.ViewUniques != 487 || tr.Clones != 340 || tr.CloneUniques != 128 {
		t.Fatalf("traffic = %+v", tr)
	}
}

func TestClient_Traffic_IncludesDailyPoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/traffic/views"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count":   10,
				"uniques": 4,
				"views":   []map[string]any{{"timestamp": "2026-08-27T00:00:00Z", "count": 7, "uniques": 3}},
			})
		case strings.HasSuffix(r.URL.Path, "/traffic/clones"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count":   5,
				"uniques": 2,
				"clones":  []map[string]any{{"timestamp": "2026-08-27T00:00:00Z", "count": 5, "uniques": 2}},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	tr, err := New(srv.URL, "tok").Traffic(context.Background(), "puppe1990", "cais")
	if err != nil {
		t.Fatal(err)
	}
	if len(tr.ViewsByDay) != 1 || tr.ViewsByDay[0].Count != 7 {
		t.Fatalf("views by day = %+v", tr.ViewsByDay)
	}
	if len(tr.ClonesByDay) != 1 || tr.ClonesByDay[0].Count != 5 {
		t.Fatalf("clones by day = %+v", tr.ClonesByDay)
	}
}

func TestClient_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	_, err := New(srv.URL, "").User(context.Background(), "octocat")
	if err != ErrRateLimited {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}

func repoJSON(name string, stars int) map[string]any {
	return map[string]any{
		"name":             name,
		"full_name":        "octocat/" + name,
		"description":      name + " desc",
		"language":         "Go",
		"html_url":         "https://github.com/octocat/" + name,
		"homepage":         "https://example.com",
		"updated_at":       "2026-08-01T00:00:00Z",
		"stargazers_count": stars,
		"forks_count":      2,
		"fork":             false,
		"archived":         false,
		"owner":            map[string]any{"login": "octocat"},
	}
}
