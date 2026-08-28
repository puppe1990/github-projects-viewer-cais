package store

import (
	"testing"
	"time"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
)

func TestCatalog_SaveAndLoadProfileReposOrgs(t *testing.T) {
	s := newTestStore(t)
	profile := catalog.Profile{Login: "octocat", Name: "The Octocat", PublicRepos: 8}
	if err := s.SaveProfile(profile, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadProfile("Octocat")
	if err != nil || !ok {
		t.Fatalf("load profile ok=%v err=%v", ok, err)
	}
	if got.Name != "The Octocat" || got.PublicRepos != 8 {
		t.Fatalf("profile = %+v", got)
	}

	repos := []catalog.Repo{{Name: "hello-world", OwnerLogin: "octocat", Stars: 10, Language: "Go"}}
	if err := s.SaveRepos("octocat", catalog.SourceUser, repos, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadRepos("octocat", catalog.SourceUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].Stars != 10 {
		t.Fatalf("repos = %+v", loaded)
	}

	orgs := []catalog.Org{{Login: "github", Description: "Where the world builds software"}}
	if err := s.SaveOrgs("octocat", orgs); err != nil {
		t.Fatal(err)
	}
	gotOrgs, err := s.LoadOrgs("octocat")
	if err != nil {
		t.Fatal(err)
	}
	if len(gotOrgs) != 1 || gotOrgs[0].Login != "github" {
		t.Fatalf("orgs = %+v", gotOrgs)
	}
}

func TestCatalog_TrafficAndWatches(t *testing.T) {
	s := newTestStore(t)
	tr := catalog.Traffic{Views: 12, ViewUniques: 4, Clones: 3, CloneUniques: 2, Available: true}
	if err := s.SaveTraffic("octocat", "hello-world", tr, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadTraffic("octocat", "hello-world")
	if err != nil || !ok || got.Views != 12 || !got.Available {
		t.Fatalf("traffic ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestCatalog_TrafficStoresDailySeries(t *testing.T) {
	s := newTestStore(t)
	tr := catalog.Traffic{
		Views: 10, ViewUniques: 4, Clones: 5, CloneUniques: 2, Available: true,
		ViewsByDay:  []catalog.TrafficPoint{{Timestamp: "2026-08-27T00:00:00Z", Count: 7, Uniques: 3}},
		ClonesByDay: []catalog.TrafficPoint{{Timestamp: "2026-08-27T00:00:00Z", Count: 5, Uniques: 2}},
	}
	if err := s.SaveTraffic("puppe1990", "cais", tr, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LoadTraffic("puppe1990", "cais")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if len(got.ViewsByDay) != 1 || got.ViewsByDay[0].Count != 7 {
		t.Fatalf("views by day = %+v", got.ViewsByDay)
	}
	if len(got.ClonesByDay) != 1 || got.ClonesByDay[0].Count != 5 {
		t.Fatalf("clones by day = %+v", got.ClonesByDay)
	}

	if err := s.TouchWatch("octocat"); err != nil {
		t.Fatal(err)
	}
	logins, err := s.WatchedLogins()
	if err != nil {
		t.Fatal(err)
	}
	if len(logins) != 1 || logins[0] != "octocat" {
		t.Fatalf("watches = %#v", logins)
	}

	if err := s.SaveRepos("octocat", catalog.SourceUser, []catalog.Repo{{Name: "hello-world", OwnerLogin: "octocat"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	refs, err := s.CachedRepoRefs()
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].Owner != "octocat" || refs[0].Name != "hello-world" {
		t.Fatalf("refs = %+v", refs)
	}
}

func TestCatalog_ProfileFetchedAt(t *testing.T) {
	s := newTestStore(t)
	when := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	if err := s.SaveProfile(catalog.Profile{Login: "octocat"}, when); err != nil {
		t.Fatal(err)
	}
	got, err := s.ProfileFetchedAt("octocat")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(when) {
		t.Fatalf("fetched_at = %s want %s", got, when)
	}
}
