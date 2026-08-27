package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/bookmarks-dashboard/backend/internal/config"
	"github.com/bookmarks-dashboard/backend/migrations"
)

func Connect(cfg *config.Config) (*bun.DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, cfg.DatabasePath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	if cfg.DevMode {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func runMigrations(db *bun.DB) error {
	migrator := migrations.NewMigrator(db)
	ctx := context.Background()
	if err := migrator.Init(ctx); err != nil {
		return err
	}
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		return nil
	}
	fmt.Printf("migrated to %s\n", group)
	return nil
}
