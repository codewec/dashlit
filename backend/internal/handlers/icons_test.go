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
	results := selfhstResults(selfhstIcon{
		Name: "Example", Reference: "example", SVG: "Yes", PNG: "Yes", Light: "Yes", Dark: "Yes",
	})
	if len(results) != 2 {
		t.Fatalf("result count = %d", len(results))
	}
	if results[0].Icon != "selfhst-icon:example.svg" || results[0].IconDark != "" || results[0].Variant != "color" {
		t.Fatalf("unexpected color result: %#v", results[0])
	}
	if results[1].Icon != "selfhst-icon:example-dark.svg" || results[1].IconDark != "selfhst-icon:example-light.svg" || results[1].Variant != "monochrome" {
		t.Fatalf("unexpected monochrome result: %#v", results[1])
	}
}

func TestSelfhstResultFallsBackToPNG(t *testing.T) {
	results := selfhstResults(selfhstIcon{Name: "Example", Reference: "example", SVG: "No", PNG: "Yes"})
	if len(results) != 1 || results[0].Icon != "selfhst-icon:example.png" || results[0].IconDark != "" || results[0].Variant != "color" {
		t.Fatalf("unexpected PNG results: %#v", results)
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
