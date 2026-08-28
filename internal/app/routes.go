package app

import (
	"github.com/puppe1990/cais/pkg/cais"

	"github.com/puppe1990/github-projects-viewer-cais/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	_ = cfg
	home := handlers.NewHomeHandler(deps.Site, deps.Inertia, deps.Loader)
	r.Get("/", home.Index)
	r.Get("/u/{login}", cais.StringParam("login", home.Show))
	r.Get("/u/{login}/orgs/{org}", cais.StringParams("login", "org", home.ShowOrg))
	r.Get("/traffic/{owner}/{repo}", cais.StringParams("owner", "repo", home.RepoTraffic))
}
