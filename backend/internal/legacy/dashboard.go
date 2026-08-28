package legacy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/internal/models"
)

const dashboardFilename = "dashboard.json"

type dashboardFile struct {
	Version string        `json:"version"`
	Groups  []legacyGroup `json:"groups"`
}

type legacyGroup struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Items       []legacyItem `json:"items"`
}

type legacyItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
}

// Migrator keeps the legacy data discovered at startup in memory. This is
// intentional: dashboard.json is inspected only while the users table is empty.
type Migrator struct {
	db      *bun.DB
	payload *dashboardFile
}

func NewMigrator(ctx context.Context, db *bun.DB, cfg *config.Config) (*Migrator, error) {
	m := &Migrator{db: db}
	count, err := db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil || count != 0 {
		return m, err
	}

	path := filepath.Join(filepath.Dir(cfg.DatabasePath), dashboardFilename)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return m, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read legacy dashboard %q: %w", path, err)
	}
	var payload dashboardFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse legacy dashboard %q: %w", path, err)
	}
	if payload.Version == "" || payload.Groups == nil {
		return nil, fmt.Errorf("parse legacy dashboard %q: unsupported format", path)
	}
	m.payload = &payload
	return m, nil
}

func (m *Migrator) Available(ctx context.Context) bool {
	if m == nil || m.payload == nil {
		return false
	}
	count, err := m.db.NewSelect().Model((*models.User)(nil)).Count(ctx)
	return err == nil && count == 0
}

// ImportForFirstUser creates one private default dashboard. The checks inside
// the transaction make a stale or forged UI choice harmless.
func (m *Migrator) ImportForFirstUser(ctx context.Context, user *models.User) error {
	if m == nil || m.payload == nil {
		return errors.New("legacy dashboard is not available")
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	users, err := tx.NewSelect().Model((*models.User)(nil)).Where("id = ?", user.ID).Count(ctx)
	if err != nil || users != 1 {
		return errors.New("legacy dashboard can only be imported for the first user")
	}
	totalUsers, err := tx.NewSelect().Model((*models.User)(nil)).Count(ctx)
	if err != nil || totalUsers != 1 {
		return errors.New("legacy dashboard can only be imported for the first user")
	}
	dashboardCount, err := tx.NewSelect().Model((*models.Dashboard)(nil)).Count(ctx)
	if err != nil || dashboardCount != 0 {
		return errors.New("legacy dashboard can only be imported into an empty database")
	}

	dashboard := &models.Dashboard{
		ID: uuid.NewString(), OwnerID: user.ID, Name: "Legacy dashboard", Slug: "legacy-dashboard",
		Layout: models.LayoutRows, Width: models.WidthDefault, Privacy: models.PrivacyPrivate, IsDefault: true,
	}
	if _, err := tx.NewInsert().Model(dashboard).Exec(ctx); err != nil {
		return err
	}
	for groupPosition, sourceGroup := range m.payload.Groups {
		group := &models.Group{
			ID: uuid.NewString(), DashboardID: dashboard.ID, Title: sourceGroup.Title,
			Description: sourceGroup.Description, ItemSize: models.Size1x1, Position: groupPosition,
		}
		if _, err := tx.NewInsert().Model(group).Exec(ctx); err != nil {
			return err
		}
		for itemPosition, sourceItem := range sourceGroup.Items {
			icon := strings.TrimSpace(sourceItem.Icon)
			if icon == "" {
				icon = "mdi:link"
			}
			item := &models.Item{
				ID: uuid.NewString(), GroupID: group.ID, Title: sourceItem.Title,
				Description: sourceItem.Description, URL: sourceItem.URL, Icon: icon, Position: itemPosition,
			}
			if _, err := tx.NewInsert().Model(item).Exec(ctx); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
