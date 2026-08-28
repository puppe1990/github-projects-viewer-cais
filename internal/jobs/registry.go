package jobs

import (
	"database/sql"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
)

// RegisterAll wires app and framework jobs into the worker registry.
func RegisterAll(reg *caisjobs.Registry, db *sql.DB, s store.Store, loader *catalog.Loader) {
	_ = s
	reg.Register(caisjobs.KindPruneSessions, caisjobs.PruneSessionsHandler(db))
	reg.Register(caisjobs.KindPruneFinished, caisjobs.PruneFinishedHandler(db))
	reg.Register("SnapshotTraffic", PerformSnapshotTraffic(loader))
	reg.Register("RefreshCatalog", PerformRefreshCatalog(loader))
	// cais:jobs-register
}
