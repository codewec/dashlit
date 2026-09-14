package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE users ADD COLUMN theme TEXT NOT NULL DEFAULT 'system';
ALTER TABLE users ADD COLUMN custom_theme TEXT NOT NULL DEFAULT '';
`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE users DROP COLUMN custom_theme;
ALTER TABLE users DROP COLUMN theme;
`)
		return err
	})
}
