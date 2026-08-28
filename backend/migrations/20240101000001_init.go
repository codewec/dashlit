package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT,
  role TEXT NOT NULL DEFAULT 'user',
  oidc_subject TEXT,
  oidc_issuer TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oidc_identity
ON users(oidc_issuer, oidc_subject)
WHERE oidc_issuer IS NOT NULL AND oidc_subject IS NOT NULL;

CREATE TABLE IF NOT EXISTS dashboards (
  id TEXT PRIMARY KEY,
  owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  icon TEXT NOT NULL DEFAULT '',
  icon_dark TEXT NOT NULL DEFAULT '',
  layout TEXT NOT NULL DEFAULT 'rows',
  width TEXT NOT NULL DEFAULT 'default',
  privacy TEXT NOT NULL DEFAULT 'private',
  clean_mode INTEGER NOT NULL DEFAULT 0,
  is_main INTEGER NOT NULL DEFAULT 0,
  is_default INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboards_one_main ON dashboards(is_main) WHERE is_main = 1;
CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboards_owner_default ON dashboards(owner_id) WHERE is_default = 1;

CREATE TABLE IF NOT EXISTS groups (
  id TEXT PRIMARY KEY,
  dashboard_id TEXT NOT NULL REFERENCES dashboards(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  icon TEXT NOT NULL DEFAULT '',
  icon_dark TEXT NOT NULL DEFAULT '',
  item_size TEXT NOT NULL DEFAULT '1x1',
  position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_groups_dashboard ON groups(dashboard_id, position);

CREATE TABLE IF NOT EXISTS items (
  id TEXT PRIMARY KEY,
  group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL,
  icon TEXT NOT NULL,
  icon_dark TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_items_group ON items(group_id, position);

CREATE TABLE IF NOT EXISTS uploaded_icons (
  id TEXT PRIMARY KEY,
  filename TEXT NOT NULL,
  mime TEXT NOT NULL,
  owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
`)
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.ExecContext(ctx, `
DROP INDEX IF EXISTS idx_users_oidc_identity;
DROP TABLE IF EXISTS uploaded_icons;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS dashboards;
DROP TABLE IF EXISTS users;
`)
		return err
	})
}
