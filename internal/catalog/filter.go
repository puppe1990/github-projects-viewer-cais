package catalog

import (
	"sort"
	"strings"
	"time"
)

type Filters struct {
	Language     string
	MinStars     int
	Query        string
	IncludeForks bool
	HasHomepage  bool
	SortBy       string
	SortOrder    string
}

func Apply(repos []Repo, f Filters) []Repo {
	query := strings.ToLower(strings.TrimSpace(f.Query))
	language := strings.TrimSpace(f.Language)
	out := make([]Repo, 0, len(repos))
	for _, repo := range repos {
		if (repo.Stars) < f.MinStars {
			continue
		}
		repoLang := repo.Language
		if repoLang == "" {
			repoLang = "Not specified"
		}
		if language != "" && repoLang != language {
			continue
		}
		if query != "" {
			blob := strings.ToLower(repo.Name + " " + repo.Description)
			if !strings.Contains(blob, query) {
				continue
			}
		}
		if !f.IncludeForks && repo.Fork {
			continue
		}
		if f.HasHomepage && strings.TrimSpace(repo.Homepage) == "" {
			continue
		}
		out = append(out, repo)
	}
	sortRepos(out, f.SortBy, f.SortOrder)
	return out
}

func sortRepos(repos []Repo, sortBy, sortOrder string) {
	asc := sortOrder == "asc"
	sort.SliceStable(repos, func(i, j int) bool {
		less := repoLess(repos[i], repos[j], sortBy)
		if asc {
			return less
		}
		return repoLess(repos[j], repos[i], sortBy)
	})
}

func repoLess(a, b Repo, sortBy string) bool {
	switch sortBy {
	case "name":
		return a.Name < b.Name
	case "updated":
		return parseTime(a.UpdatedAt).Before(parseTime(b.UpdatedAt))
	case "forks":
		return a.Forks < b.Forks
	default:
		return a.Stars < b.Stars
	}
}

func parseTime(iso string) time.Time {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return time.Time{}
	}
	return t
}
