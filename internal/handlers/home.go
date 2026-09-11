package handlers

import (
	"encoding/json"
	"errors"
	"log"
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
	tracked := &trackedWriter{ResponseWriter: w}
	if err := h.inertia.Render(tracked, r, "Home", props); err != nil {
		// Inertia writes the template straight to the response, so a client that
		// aborts mid-page leaves the status committed. Rewriting it as 500 would
		// log a phantom server error and append text to the partial HTML.
		if tracked.started {
			log.Printf("home: render %s: %v", r.URL.Path, err)
			return
		}
		http.Error(tracked, err.Error(), http.StatusInternalServerError)
	}
}

// trackedWriter records whether anything was written, which is the only way to
// tell a render failure that can still become a 500 from one that cannot.
type trackedWriter struct {
	http.ResponseWriter
	started bool
}

func (w *trackedWriter) WriteHeader(code int) {
	w.started = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *trackedWriter) Write(p []byte) (int, error) {
	w.started = true
	return w.ResponseWriter.Write(p)
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
