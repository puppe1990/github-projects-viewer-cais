package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/puppe1990/cais/pkg/cais/flash"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

type HomeHandler struct {
	site    meta.Site
	inertia *inertia.Inertia
	loader  *catalog.Loader
}

func NewHomeHandler(site meta.Site, i *inertia.Inertia, loader *catalog.Loader) *HomeHandler {
	return &HomeHandler{site: site, inertia: i, loader: loader}
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	if h.loader != nil && h.loader.GitHub != nil && h.loader.GitHub.HasToken() {
		me, err := h.loader.GitHub.Me(r.Context())
		if err == nil && me.Login != "" {
			http.Redirect(w, r, "/u/"+me.Login+"/all", http.StatusSeeOther)
			return
		}
	}
	h.render(w, r, catalog.Snapshot{}, "", "", false)
}

func (h *HomeHandler) Show(w http.ResponseWriter, r *http.Request, login string) {
	login = normalizeLogin(login)
	if login == "" {
		h.render(w, r, catalog.Snapshot{}, "", "Type a GitHub username to open their catalog.", false)
		return
	}
	snap, err := h.loader.User(r.Context(), login)
	if err != nil {
		h.render(w, r, catalog.Snapshot{}, login, catalogError(err), false)
		return
	}
	h.render(w, r, snap, login, "", true)
}

func (h *HomeHandler) ShowAll(w http.ResponseWriter, r *http.Request, login string) {
	login = normalizeLogin(login)
	if login == "" {
		h.render(w, r, catalog.Snapshot{}, "", "Type a GitHub username to open their catalog.", false)
		return
	}
	snap, err := h.loader.All(r.Context(), login)
	if err != nil {
		h.render(w, r, catalog.Snapshot{}, login, catalogError(err), false)
		return
	}
	h.render(w, r, snap, login, "", true)
}

func (h *HomeHandler) ShowOrg(w http.ResponseWriter, r *http.Request, login, org string) {
	login = normalizeLogin(login)
	org = normalizeLogin(org)
	if login == "" || org == "" {
		h.render(w, r, catalog.Snapshot{}, login, "Type a GitHub username to open their catalog.", false)
		return
	}
	snap, err := h.loader.Org(r.Context(), login, org)
	if err != nil {
		h.render(w, r, catalog.Snapshot{}, login, catalogError(err), false)
		return
	}
	h.render(w, r, snap, login, "", true)
}

func (h *HomeHandler) render(w http.ResponseWriter, r *http.Request, snap catalog.Snapshot, lookup, errMsg string, populated bool) {
	props := inertia.Props{
		"title":     "Projects Viewer",
		"site":      meta.ForRequest(h.site, r),
		"profile":   snap.Profile,
		"orgs":      snap.Orgs,
		"repos":     snap.Repos,
		"source":    map[string]string{"type": snap.Source, "login": snap.SourceLogin},
		"error":     errMsg,
		"populated": populated,
		"lookup":    lookup,
	}
	if snap.Orgs == nil {
		props["orgs"] = []catalog.Org{}
	}
	if snap.Repos == nil {
		props["repos"] = []catalog.Repo{}
	}
	if msg, ok := flash.MessageFromRequest(r); ok {
		props["flash"] = inertia.Flash{msg.Kind: msg.Message}
	}
	if err := h.inertia.Render(w, r, "Home", props); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HomeHandler) RepoTraffic(w http.ResponseWriter, r *http.Request, owner, repo string) {
	owner = normalizeLogin(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" || h.loader == nil {
		http.Error(w, "Could not load traffic for this repository.", http.StatusNotFound)
		return
	}
	if err := h.loader.SnapshotTraffic(r.Context(), owner, repo); err != nil {
		http.Error(w, catalogError(err), http.StatusBadGateway)
		return
	}
	tr, ok, err := h.loader.Cache.LoadTraffic(owner, repo)
	if err != nil || !ok {
		http.Error(w, "No traffic for this repository.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tr)
}

func normalizeLogin(login string) string {
	return strings.TrimPrefix(strings.TrimSpace(login), "@")
}

func catalogError(err error) string {
	switch {
	case errors.Is(err, githubapi.ErrNotFound):
		return "No GitHub user with that username. Check the spelling and try again."
	case errors.Is(err, githubapi.ErrRateLimited):
		return "GitHub rate limit hit. Wait a few minutes, or search again later."
	default:
		return "Could not load that profile."
	}
}
