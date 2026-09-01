package handlers

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

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
		PingURL: "https://health.example.com", PingSkipTLS: true, Position: 3,
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
	if clonedItem.ID == item.ID || clonedItem.GroupID != clone.ID || clonedItem.Title != item.Title || clonedItem.PingURL != item.PingURL || !clonedItem.PingSkipTLS || clonedItem.Position != item.Position {
		t.Fatalf("unexpected cloned item: %#v", clonedItem)
	}
}
