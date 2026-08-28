package db

import (
	"context"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
)

// RunSeeds populates demo data. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	jobStore := caisjobs.NewStore(s.DB())
	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		return err
	}

	if err := jobStore.UpsertRecurring(context.Background(), caisjobs.RecurringOptions{
		Kind: "SnapshotTraffic",
		Cron: "0 * * * *",
	}); err != nil {
		return err
	}

	if err := jobStore.UpsertRecurring(context.Background(), caisjobs.RecurringOptions{
		Kind: "RefreshCatalog",
		Cron: "15 * * * *",
	}); err != nil {
		return err
	}
	// cais:recurring-seeds
	// cais:seeds
	return nil
}
