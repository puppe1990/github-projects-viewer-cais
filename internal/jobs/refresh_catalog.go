package jobs

import (
	"context"
	"encoding/json"

	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
)

type refreshPayload struct {
	Login string `json:"login"`
}

func PerformRefreshCatalog(loader *catalog.Loader) caisjobs.Handler {
	return func(ctx context.Context, payload []byte) error {
		var p refreshPayload
		_ = json.Unmarshal(payload, &p)
		if p.Login != "" {
			return skipRateLimited(loader.Refresh(ctx, p.Login))
		}
		logins, err := loader.Cache.WatchedLogins()
		if err != nil {
			return err
		}
		for _, login := range logins {
			if err := skipRateLimited(loader.Refresh(ctx, login)); err != nil {
				return err
			}
		}
		return nil
	}
}
