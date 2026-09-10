package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

type trafficPayload struct {
	Login string `json:"login"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func PerformSnapshotTraffic(loader *catalog.Loader) caisjobs.Handler {
	return func(ctx context.Context, payload []byte) error {
		var p trafficPayload
		_ = json.Unmarshal(payload, &p)
		return skipRateLimited(snapshotTraffic(ctx, loader, p))
	}
}

func snapshotTraffic(ctx context.Context, loader *catalog.Loader, p trafficPayload) error {
	if p.Owner != "" && p.Repo != "" {
		return loader.SnapshotTraffic(ctx, p.Owner, p.Repo)
	}
	if p.Login != "" {
		return loader.SnapshotLogin(ctx, p.Login)
	}
	logins, err := loader.Cache.WatchedLogins()
	if err != nil {
		return err
	}
	var first error
	for _, login := range logins {
		if err := loader.SnapshotLogin(ctx, login); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// skipRateLimited ends a rate-limited run without marking it failed. The worker
// backs off seconds while GitHub's window resets in up to an hour, so retrying
// now only buries the job in the failed pile; the next scheduled run refreshes it.
func skipRateLimited(err error) error {
	if !errors.Is(err, githubapi.ErrRateLimited) {
		return err
	}
	log.Printf("jobs skipped: %v; the hourly SnapshotTraffic cron retries later", err)
	return nil
}
