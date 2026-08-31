package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE items ADD COLUMN ping_url TEXT NOT NULL DEFAULT '';
ALTER TABLE items ADD COLUMN ping_skip_tls INTEGER NOT NULL DEFAULT 0;
`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
ALTER TABLE items DROP COLUMN ping_skip_tls;
ALTER TABLE items DROP COLUMN ping_url;
`)
		return err
	})
}
