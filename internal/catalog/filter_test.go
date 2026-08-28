package catalog

import "testing"

func sampleRepos() []Repo {
	return []Repo{
		{Name: "alpha", Description: "first", Language: "Go", Stars: 10, Forks: 1, UpdatedAt: "2026-08-01T00:00:00Z"},
		{Name: "beta", Description: "second site", Language: "Python", Stars: 3, Forks: 8, Homepage: "https://example.com", Fork: true, UpdatedAt: "2026-08-20T00:00:00Z"},
		{Name: "gamma", Description: "tooling", Language: "Go", Stars: 30, Forks: 2, UpdatedAt: "2026-07-01T00:00:00Z"},
	}
}

func TestApply_MinStars(t *testing.T) {
	got := Apply(sampleRepos(), Filters{MinStars: 10, IncludeForks: true, SortBy: "stars", SortOrder: "desc"})
	if len(got) != 2 || got[0].Name != "gamma" || got[1].Name != "alpha" {
		t.Fatalf("got %#v", names(got))
	}
}

func TestApply_LanguageAndQuery(t *testing.T) {
	got := Apply(sampleRepos(), Filters{Language: "Go", Query: "tool", IncludeForks: true, SortBy: "name", SortOrder: "asc"})
	if len(got) != 1 || got[0].Name != "gamma" {
		t.Fatalf("got %#v", names(got))
	}
}

func TestApply_HidesForksAndRequiresHomepage(t *testing.T) {
	got := Apply(sampleRepos(), Filters{IncludeForks: false, HasHomepage: true, SortBy: "stars", SortOrder: "desc"})
	if len(got) != 0 {
		t.Fatalf("got %#v", names(got))
	}
	got = Apply(sampleRepos(), Filters{IncludeForks: true, HasHomepage: true, SortBy: "stars", SortOrder: "desc"})
	if len(got) != 1 || got[0].Name != "beta" {
		t.Fatalf("got %#v", names(got))
	}
}

func TestApply_SortUpdatedAsc(t *testing.T) {
	got := Apply(sampleRepos(), Filters{IncludeForks: true, SortBy: "updated", SortOrder: "asc"})
	if names(got)[0] != "gamma" || names(got)[2] != "beta" {
		t.Fatalf("got %#v", names(got))
	}
}

func names(repos []Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Name
	}
	return out
}
