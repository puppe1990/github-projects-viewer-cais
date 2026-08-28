package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/puppe1990/cais/pkg/cais"
	"github.com/puppe1990/cais/pkg/cais/boot"
	"github.com/puppe1990/cais/pkg/cais/meta"
	inertia "github.com/romsar/gonertia/v3"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/app"
	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	appdb "github.com/puppe1990/github-projects-viewer-cais/internal/db"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
	appi18n "github.com/puppe1990/github-projects-viewer-cais/internal/i18n"
	appjobs "github.com/puppe1990/github-projects-viewer-cais/internal/jobs"
	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
	"github.com/puppe1990/github-projects-viewer-cais/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "github-projects-viewer-cais",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	messages := appi18n.NewCatalog(cfg.Locale)
	templatesDir, err := cais.ResolveWebDir("templates", cfg.TemplatesDir)
	if err != nil {
		templatesDir = ""
	}
	renderer, err := cais.NewRendererForEnv(cfg, tmplFS, templatesDir, messages)
	if err != nil {
		return nil, fmt.Errorf("renderer: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := cais.ResolveWebDir("static", cfg.StaticDir)
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	inertiaI, err := inertia.NewFromFileFS(tmplFS, "app.html")
	if err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("inertia root: %w", err)
	}

	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("jobs schema: %w", err)
	}
	if err := appdb.RunSeeds(s); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seeds: %w", err)
	}
	gh := githubapi.New("", os.Getenv("GITHUB_TOKEN"))
	loader := &catalog.Loader{
		GitHub: gh,
		Cache:  s,
		Queue:  appjobs.SQLiteQueue{Jobs: caisjobs.NewStore(s.DB())},
	}

	return app.New(cfg, app.Deps{
		Renderer:  renderer,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("github-projects-viewer-cais", cfg.AppURL),
		Catalog:   messages,
		Inertia:   inertiaI,
		Loader:    loader,
	})
}
