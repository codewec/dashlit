package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/models"
)

var errHotkeyConflict = errors.New("hotkey conflicts with another shortcut")

type hotkeyConflict struct {
	hotkey string
}

func (err hotkeyConflict) Error() string {
	return "Hotkey “" + err.hotkey + "” is already in use."
}

func (hotkeyConflict) Unwrap() error {
	return errHotkeyConflict
}

func hotkeyConflictError(hotkey, label string) error {
	parts := strings.Split(normalizeHotkey(hotkey), "+")
	if len(parts) > 0 {
		key := strings.TrimSpace(parts[len(parts)-1])
		if label != "" {
			parts[len(parts)-1] = label
		} else if strings.HasPrefix(key, "Digit") {
			parts[len(parts)-1] = strings.TrimPrefix(key, "Digit")
		} else if strings.HasPrefix(key, "Key") {
			parts[len(parts)-1] = strings.TrimPrefix(key, "Key")
		}
	}
	return hotkeyConflict{hotkey: strings.Join(parts, " + ")}
}

func normalizeHotkey(hotkey string) string {
	return strings.TrimSpace(hotkey)
}

func canonicalHotkey(hotkey string) string {
	return strings.ToLower(normalizeHotkey(hotkey))
}

func dashboardHotkeyConflicts(ctx context.Context, db *bun.DB, ownerID, hotkey, excludeDashboardID string) (bool, error) {
	canonical := canonicalHotkey(hotkey)
	if canonical == "" {
		return false, nil
	}

	dashboards := db.NewSelect().Model((*models.Dashboard)(nil)).
		Where("d.owner_id = ?", ownerID).
		Where("LOWER(TRIM(d.hotkey)) = ?", canonical)
	if excludeDashboardID != "" {
		dashboards = dashboards.Where("d.id <> ?", excludeDashboardID)
	}
	exists, err := dashboards.Exists(ctx)
	if err != nil || exists {
		return exists, err
	}

	return db.NewSelect().Model((*models.Item)(nil)).
		Join("JOIN groups AS g ON g.id = i.group_id").
		Join("JOIN dashboards AS d ON d.id = g.dashboard_id").
		Where("d.owner_id = ?", ownerID).
		Where("LOWER(TRIM(i.hotkey)) = ?", canonical).
		Exists(ctx)
}

func itemHotkeyConflicts(ctx context.Context, db *bun.DB, ownerID, hotkey string) (bool, error) {
	canonical := canonicalHotkey(hotkey)
	if canonical == "" {
		return false, nil
	}
	return db.NewSelect().Model((*models.Dashboard)(nil)).
		Where("d.owner_id = ?", ownerID).
		Where("LOWER(TRIM(d.hotkey)) = ?", canonical).
		Exists(ctx)
}
