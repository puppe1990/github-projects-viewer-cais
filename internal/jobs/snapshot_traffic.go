package jobs

import (
	"context"
	"encoding/json"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
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
}
