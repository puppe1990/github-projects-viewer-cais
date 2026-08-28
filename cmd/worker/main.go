package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/puppe1990/cais/pkg/cais"
	caisjobs "github.com/puppe1990/cais/pkg/cais/jobs"

	"github.com/puppe1990/github-projects-viewer-cais/internal/catalog"
	"github.com/puppe1990/github-projects-viewer-cais/internal/githubapi"
	appjobs "github.com/puppe1990/github-projects-viewer-cais/internal/jobs"
	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	queues := flag.String("queues", "default", "comma-separated queue names")
	concurrency := flag.Int("concurrency", 2, "worker goroutines")
	flag.Parse()

	cfg := cais.Load()
	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		return err
	}

	gh := githubapi.New("", os.Getenv("GITHUB_TOKEN"))
	loader := &catalog.Loader{GitHub: gh, Cache: s}
	reg := caisjobs.NewRegistry()
	appjobs.RegisterAll(reg, s.DB(), s, loader)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := caisjobs.NewWorker(caisjobs.WorkerConfig{
		Store:       caisjobs.NewStore(s.DB()),
		Registry:    reg,
		Queues:      splitQueues(*queues),
		Concurrency: *concurrency,
	})
	log.Printf("=> Worker started (queues=%s, concurrency=%d)", *queues, *concurrency)
	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		return err
	}
	return nil
}

func splitQueues(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return []string{caisjobs.DefaultQueue}
	}
	return out
}
