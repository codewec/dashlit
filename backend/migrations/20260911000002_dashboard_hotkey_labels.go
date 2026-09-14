package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE dashboards ADD COLUMN hotkey TEXT NOT NULL DEFAULT '';
ALTER TABLE dashboards ADD COLUMN hotkey_label TEXT NOT NULL DEFAULT '';
ALTER TABLE items ADD COLUMN hotkey_label TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX idx_dashboards_owner_hotkey
ON dashboards(owner_id, lower(trim(hotkey)))
WHERE trim(hotkey) <> '';
`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP INDEX IF EXISTS idx_dashboards_owner_hotkey;
ALTER TABLE items DROP COLUMN hotkey_label;
ALTER TABLE dashboards DROP COLUMN hotkey_label;
ALTER TABLE dashboards DROP COLUMN hotkey;
`)
		return err
	})
}
