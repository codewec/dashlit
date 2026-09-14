package handlers

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/bookmarks-dashboard/backend/internal/config"
	appdb "github.com/bookmarks-dashboard/backend/internal/db"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

func TestHotkeyConflictsAreScopedByOwnerAndType(t *testing.T) {
	database, err := appdb.Connect(&config.Config{DatabasePath: filepath.Join(t.TempDir(), "hotkeys.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	ctx := context.Background()

	users := []*models.User{
		{ID: "owner-1", Username: "one", Role: models.RoleUser},
		{ID: "owner-2", Username: "two", Role: models.RoleUser},
	}
	if _, err := database.NewInsert().Model(&users).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	dashboards := []*models.Dashboard{
		{ID: "dash-1", OwnerID: "owner-1", Name: "One", Slug: "one", Hotkey: "Ctrl+Digit1"},
		{ID: "dash-2", OwnerID: "owner-1", Name: "Two", Slug: "two"},
		{ID: "dash-3", OwnerID: "owner-2", Name: "Other", Slug: "other", Hotkey: "Ctrl+Digit1"},
	}
	if _, err := database.NewInsert().Model(&dashboards).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	group := &models.Group{ID: "group-1", DashboardID: "dash-2", Title: "Links", ItemSize: models.Size1x1}
	if _, err := database.NewInsert().Model(group).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	item := &models.Item{ID: "item-1", GroupID: group.ID, Title: "Link", URL: "https://example.com", Icon: "mdi:link", Hotkey: "Alt+KeyL"}
	if _, err := database.NewInsert().Model(item).Exec(ctx); err != nil {
		t.Fatal(err)
	}

	assertConflict := func(name string, got bool, err error, want bool) {
		t.Helper()
		if err != nil || got != want {
			t.Fatalf("%s: conflict=%v err=%v, want %v", name, got, err, want)
		}
	}

	got, err := dashboardHotkeyConflicts(ctx, database, "owner-1", "Ctrl+Digit1", "")
	assertConflict("dashboard vs dashboard", got, err, true)
	got, err = dashboardHotkeyConflicts(ctx, database, "owner-1", "Ctrl+Digit1", "dash-1")
	assertConflict("exclude edited dashboard", got, err, false)
	got, err = dashboardHotkeyConflicts(ctx, database, "owner-1", "Alt+KeyL", "")
	assertConflict("dashboard vs item", got, err, true)
	got, err = itemHotkeyConflicts(ctx, database, "owner-1", "Ctrl+Digit1")
	assertConflict("item vs dashboard", got, err, true)
	got, err = itemHotkeyConflicts(ctx, database, "owner-1", "Alt+KeyL")
	assertConflict("item vs item remains allowed", got, err, false)
	got, err = dashboardHotkeyConflicts(ctx, database, "owner-2", "Alt+KeyL", "")
	assertConflict("different owner", got, err, false)
}

func TestHotkeyConflictErrorIncludesDisplayLabel(t *testing.T) {
	err := hotkeyConflictError(" Ctrl+Digit1 ", "1")
	if !errors.Is(err, errHotkeyConflict) {
		t.Fatal("conflict error must preserve its sentinel")
	}
	if got, want := err.Error(), "Hotkey “Ctrl + 1” is already in use."; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if got, want := hotkeyConflictError("Ctrl+Digit2", "").Error(), "Hotkey “Ctrl + 2” is already in use."; got != want {
		t.Fatalf("legacy error = %q, want %q", got, want)
	}
}

func TestHideForeignDashboardHotkeys(t *testing.T) {
	owner := &models.User{ID: "owner-1"}
	own := &models.Dashboard{OwnerID: owner.ID, Hotkey: "Ctrl+KeyO", HotkeyLabel: "o"}
	foreign := &models.Dashboard{OwnerID: "owner-2", Hotkey: "Ctrl+KeyF", HotkeyLabel: "f"}

	hideForeignDashboardHotkeys(owner, own, foreign)
	if own.Hotkey == "" || own.HotkeyLabel == "" {
		t.Fatal("owner hotkey was hidden")
	}
	if foreign.Hotkey != "" || foreign.HotkeyLabel != "" {
		t.Fatal("foreign hotkey was exposed")
	}

	public := &models.Dashboard{OwnerID: owner.ID, Hotkey: "Alt+KeyP", HotkeyLabel: "p"}
	hideForeignDashboardHotkeys(nil, public)
	if public.Hotkey != "" || public.HotkeyLabel != "" {
		t.Fatal("dashboard hotkey was exposed to an anonymous user")
	}
}
