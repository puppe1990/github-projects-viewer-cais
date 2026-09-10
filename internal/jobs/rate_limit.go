package jobs

import (
	"errors"
	"log"

	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
)

// skipRateLimited ends a rate-limited run without marking it failed. The worker
// backs off seconds while GitHub's window resets in up to an hour, so retrying
// now only buries the job in the failed pile; the next scheduled run refreshes it.
func skipRateLimited(err error) error {
	if !errors.Is(err, githubapi.ErrRateLimited) {
		return err
	}
	log.Printf("jobs skipped: %v; the hourly cron retries later", err)
	return nil
}
