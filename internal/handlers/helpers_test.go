package handlers

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais/meta"

	"github.com/puppe1990/github-projects-viewer-cais/internal/store"
)

func testSite() meta.Site {
	return meta.Site{AppName: "github-projects-viewer-cais", AppURL: "https://cais.example.com"}
}

func setupTestStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}
