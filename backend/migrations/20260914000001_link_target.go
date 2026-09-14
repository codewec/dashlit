package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE dashboards ADD COLUMN open_in_new_tab BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE groups ADD COLUMN open_in_new_tab BOOLEAN;
ALTER TABLE items ADD COLUMN open_in_new_tab BOOLEAN;
`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE items DROP COLUMN open_in_new_tab;
ALTER TABLE groups DROP COLUMN open_in_new_tab;
ALTER TABLE dashboards DROP COLUMN open_in_new_tab;
`)
		return err
	})
}
