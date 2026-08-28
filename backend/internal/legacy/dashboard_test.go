package legacy_test

import (
	"context"
	"github.com/uptrace/bun"
	"os"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/auth"
	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/legacy"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

func setup(t *testing.T, contents string) (*config.Config, *bun.DB, *auth.Service) {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{DatabasePath: filepath.Join(dir, "bookmarks.db"), JWTSecret: "test-secret"}
	if err := os.WriteFile(filepath.Join(dir, "dashboard.json"), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	database, err := appdb.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return cfg, database, auth.NewService(database, cfg)
}

func TestImportsLegacyDashboardForFirstUser(t *testing.T) {
	ctx := context.Background()
	cfg, database, authService := setup(t, `{
		"version":"0.0.6",
		"groups":[{"title":"Services","description":"Internal","items":[
			{"title":"Status","description":"Health","url":"https://status.example.com","icon":"mdi:heart-pulse"}
		]}]
	}`)
	migrator, err := legacy.NewMigrator(ctx, database, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !migrator.Available(ctx) {
		t.Fatal("legacy dashboard should be available while the database is empty")
	}
	user, err := authService.Register(ctx, "admin", "password")
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.ImportForFirstUser(ctx, user); err != nil {
		t.Fatal(err)
	}

	dashboard := new(models.Dashboard)
	if err := database.NewSelect().Model(dashboard).Relation("Groups").Relation("Groups.Items").Scan(ctx); err != nil {
		t.Fatal(err)
	}
	if dashboard.OwnerID != user.ID || !dashboard.IsDefault || len(dashboard.Groups) != 1 || len(dashboard.Groups[0].Items) != 1 {
		t.Fatalf("unexpected imported dashboard: %#v", dashboard)
	}
	if got := dashboard.Groups[0].Items[0].URL; got != "https://status.example.com" {
		t.Fatalf("imported URL = %q", got)
	}
}

func TestDoesNotInspectLegacyFileWhenUsersExist(t *testing.T) {
	ctx := context.Background()
	cfg, database, authService := setup(t, `{not valid json`)
	if _, err := authService.Register(ctx, "admin", "password"); err != nil {
		t.Fatal(err)
	}
	migrator, err := legacy.NewMigrator(ctx, database, cfg)
	if err != nil {
		t.Fatalf("legacy file was inspected for a non-empty database: %v", err)
	}
	if migrator.Available(ctx) {
		t.Fatal("legacy dashboard must not be available after a user exists")
	}
}
