package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

func TestPingItemFallsBackToGETOnHEADErrors(t *testing.T) {
	cfg := &config.Config{DatabasePath: filepath.Join(t.TempDir(), "ping.db")}
	database, err := appdb.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	ctx := context.Background()
	user := &models.User{ID: "ping-user", Username: "ping-user"}
	dashboard := &models.Dashboard{ID: "ping-dashboard", OwnerID: user.ID, Name: "Ping", Slug: "ping", Privacy: models.PrivacyPublic}
	group := &models.Group{ID: "ping-group", DashboardID: dashboard.ID, Title: "Ping"}
	for _, model := range []any{user, dashboard, group} {
		if _, err := database.NewInsert().Model(model).Exec(ctx); err != nil {
			t.Fatal(err)
		}
	}
	tests := []struct {
		name                  string
		headStatus, getStatus int
		wantGET, reachable    bool
	}{
		{"healthy HEAD", 200, 500, false, true},
		{"HEAD redirect", 302, 500, false, true},
		{"PBS bad request", 400, 200, true, true},
		{"HEAD unauthorized", 401, 200, true, true},
		{"HEAD forbidden", 403, 200, true, true},
		{"HEAD not found", 404, 200, true, true},
		{"method not allowed", 405, 206, true, true},
		{"HEAD server error", 500, 200, true, true},
		{"HEAD not implemented", 501, 200, true, true},
		{"GET unauthorized", 401, 401, true, false},
		{"GET server error", 400, 503, true, false},
		{"GET redirect", 400, 302, true, true},
	}
	h := NewGroupItemHandler(database)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				if r.URL.Path != "/" {
					t.Error("redirect was followed")
				}
				w.Header().Set("Location", "/redirect-target")
				if r.Method == http.MethodHead {
					w.WriteHeader(tc.headStatus)
					return
				}
				if r.Header.Get("Range") != "bytes=0-0" {
					t.Error("GET must request a limited range")
				}
				w.WriteHeader(tc.getStatus)
			}))
			defer server.Close()
			item := &models.Item{ID: tc.name, GroupID: group.ID, Title: tc.name, URL: server.URL, PingEnabled: true}
			if _, err := database.NewInsert().Model(item).Exec(ctx); err != nil {
				t.Fatal(err)
			}
			router := chi.NewRouter()
			router.Get("/items/{id}/ping", h.PingItem)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items/"+url.PathEscape(item.ID)+"/ping", nil))
			var result struct {
				Reachable bool `json:"reachable"`
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Reachable != tc.reachable {
				t.Errorf("reachable = %v, want %v", result.Reachable, tc.reachable)
			}
			wantMethods := []string{http.MethodHead}
			if tc.wantGET {
				wantMethods = append(wantMethods, http.MethodGet)
			}
			if !slices.Equal(methods, wantMethods) {
				t.Errorf("methods = %v, want %v", methods, wantMethods)
			}
		})
	}
}

func TestCloneGroupToDashboardCopiesItemsAndAppendsGroup(t *testing.T) {
	cfg := &config.Config{DatabasePath: filepath.Join(t.TempDir(), "clone.db")}
	database, err := appdb.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	ctx := context.Background()

	user := &models.User{ID: "user-1", Username: "admin", Role: models.RoleAdmin}
	if _, err := database.NewInsert().Model(user).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	sourceDashboard := &models.Dashboard{ID: "dashboard-1", OwnerID: user.ID, Name: "Source", Slug: "source"}
	targetDashboard := &models.Dashboard{ID: "dashboard-2", OwnerID: user.ID, Name: "Target", Slug: "target"}
	dashboards := []*models.Dashboard{sourceDashboard, targetDashboard}
	if _, err := database.NewInsert().Model(&dashboards).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	existing := &models.Group{ID: "existing", DashboardID: targetDashboard.ID, Title: "Existing", ItemSize: models.Size1x1, Position: 4}
	source := &models.Group{
		ID: "source-group", DashboardID: sourceDashboard.ID, Title: "Services", Description: "Internal",
		Icon: "mdi:server", IconDark: "mdi:server-outline", ItemSize: models.Size1x2, Position: 0,
	}
	groups := []*models.Group{existing, source}
	if _, err := database.NewInsert().Model(&groups).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	item := &models.Item{
		ID: "source-item", GroupID: source.ID, Title: "Status", Description: "Health", URL: "https://status.example.com",
		Icon: "mdi:heart", IconDark: "mdi:heart-outline", PingEnabled: true, PingOnlyDown: true,
		PingURL: "https://health.example.com", PingSkipTLS: true, Hotkey: "Ctrl+KeyH", HotkeyLabel: "р", Position: 3,
	}
	if _, err := database.NewInsert().Model(item).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	source.Items = []*models.Item{item}

	clone, err := NewGroupItemHandler(database).cloneGroupToDashboard(ctx, source, targetDashboard.ID)
	if err != nil {
		t.Fatal(err)
	}
	if clone.ID == source.ID || clone.DashboardID != targetDashboard.ID || clone.Title != source.Title || clone.Position != 5 || clone.ItemSize != source.ItemSize {
		t.Fatalf("unexpected cloned group: %#v", clone)
	}
	if len(clone.Items) != 1 {
		t.Fatalf("cloned items = %d", len(clone.Items))
	}
	clonedItem := clone.Items[0]
	if clonedItem.ID == item.ID || clonedItem.GroupID != clone.ID || clonedItem.Title != item.Title || clonedItem.PingURL != item.PingURL || !clonedItem.PingSkipTLS || clonedItem.Hotkey != item.Hotkey || clonedItem.HotkeyLabel != item.HotkeyLabel || clonedItem.Position != item.Position {
		t.Fatalf("unexpected cloned item: %#v", clonedItem)
	}
}

func TestPingClientsDoNotFollowRedirects(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect was followed")
	}))
	defer target.Close()
	redirector := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusFound))
	defer redirector.Close()
	h := NewGroupItemHandler(nil)
	for _, client := range []*http.Client{h.httpClient, h.insecureHTTPClient} {
		resp, err := client.Get(redirector.URL)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	}
}
