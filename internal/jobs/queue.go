package jobs

import (
	"context"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"
)

type SQLiteQueue struct {
	Jobs *caisjobs.Store
}

func (q SQLiteQueue) EnqueueTraffic(login string) error {
	_, err := caisjobs.Enqueue(context.Background(), q.Jobs, caisjobs.Options{
		Kind:    "SnapshotTraffic",
		Payload: trafficPayload{Login: login},
	})
	return err
}

func (q SQLiteQueue) EnqueueRefresh(login string) error {
	_, err := caisjobs.Enqueue(context.Background(), q.Jobs, caisjobs.Options{
		Kind:    "RefreshCatalog",
		Payload: refreshPayload{Login: login},
	})
	return err
}
