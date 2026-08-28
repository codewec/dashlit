package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestSelfhstResultPrefersSVGAndThemeVariants(t *testing.T) {
	result := selfhstResult(selfhstIcon{
		Name: "Example", Reference: "example", SVG: "Yes", PNG: "Yes", Light: "Yes", Dark: "Yes",
	})
	if result.Icon != "selfhst-icon:example-dark.svg" {
		t.Fatalf("light-theme icon = %q", result.Icon)
	}
	if result.IconDark != "selfhst-icon:example-light.svg" {
		t.Fatalf("dark-theme icon = %q", result.IconDark)
	}
}

func TestSelfhstResultFallsBackToPNG(t *testing.T) {
	result := selfhstResult(selfhstIcon{Name: "Example", Reference: "example", SVG: "No", PNG: "Yes"})
	if result.Icon != "selfhst-icon:example.png" || result.IconDark != "selfhst-icon:example.png" {
		t.Fatalf("unexpected PNG result: %#v", result)
	}
}

func TestProxySelfhstServesCachedIcon(t *testing.T) {
	cacheDir := t.TempDir()
	path := filepath.Join(cacheDir, "selfhst", "convoy.svg")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`), 0600); err != nil {
		t.Fatal(err)
	}
	handler := NewIconHandler(nil, &config.Config{IconCacheDir: cacheDir})
	router := chi.NewRouter()
	router.Get("/api/icons/selfhst/*", handler.ProxySelfhst)
	request := httptest.NewRequest(http.MethodGet, "/api/icons/selfhst/convoy.svg", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("status = %d, content-type = %q, body = %s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
}

func TestSearchSelfhstMatchesNameReferenceAndTags(t *testing.T) {
	icons := []selfhstIcon{
		{Name: "Home Assistant", Reference: "home-assistant", Tags: "Automation", SVG: "Yes"},
		{Name: "Unrelated", Reference: "other", SVG: "Yes"},
	}
	for _, query := range []string{"assistant", "home-assistant", "automation"} {
		results := searchSelfhst(icons, query, 10)
		if len(results) != 1 || results[0].Name != "Home Assistant" {
			t.Fatalf("query %q returned %#v", query, results)
		}
	}
}
