package migrations

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

func NewMigrator(db *bun.DB) *migrate.Migrator {
	return migrate.NewMigrator(db, Migrations)
}
