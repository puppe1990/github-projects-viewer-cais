package catalog

import "github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"

type Profile = githubapi.User
type Org = githubapi.Org
type Traffic = githubapi.Traffic
type TrafficPoint = githubapi.TrafficPoint

type Repo struct {
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Description string  `json:"description"`
	Language    string  `json:"language"`
	HTMLURL     string  `json:"html_url"`
	Homepage    string  `json:"homepage"`
	UpdatedAt   string  `json:"updated_at"`
	OwnerLogin  string  `json:"owner_login"`
	Stars       int     `json:"stargazers_count"`
	Forks       int     `json:"forks_count"`
	Fork        bool    `json:"fork"`
	Archived    bool    `json:"archived"`
	Traffic     Traffic `json:"traffic"`
}

type RepoRef struct {
	Owner string
	Name  string
}

type Snapshot struct {
	Profile     Profile `json:"profile"`
	Orgs        []Org   `json:"orgs"`
	Repos       []Repo  `json:"repos"`
	Source      string  `json:"source"`
	SourceLogin string  `json:"source_login"`
}

const (
	SourceUser = "user"
	SourceOrg  = "org"
)

func ReposFromGitHub(in []githubapi.Repo) []Repo {
	out := make([]Repo, len(in))
	for i, r := range in {
		out[i] = Repo{
			Name:        r.Name,
			FullName:    r.FullName,
			Description: r.Description,
			Language:    r.Language,
			HTMLURL:     r.HTMLURL,
			Homepage:    r.Homepage,
			UpdatedAt:   r.UpdatedAt,
			OwnerLogin:  r.OwnerLogin,
			Stars:       r.Stars,
			Forks:       r.Forks,
			Fork:        r.Fork,
			Archived:    r.Archived,
		}
	}
	return out
}
