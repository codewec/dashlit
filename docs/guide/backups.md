# Backups and upgrades

DashLit uses SQLite and stores uploaded assets on disk. A useful backup contains both the database and uploaded icons.

## What to preserve

```text
bookmarks.db   required
icons/         required when custom icons were uploaded
icon-cache/    optional; DashLit can rebuild it
```

If a legacy `dashboard.json` is still present, preserve it until migration has been verified.

## Safe backup

The simplest consistent backup is taken while the container is stopped:

```bash
docker compose stop dashlit
# Copy or snapshot the persistent data volume here.
docker compose start dashlit
```

For backups without downtime, use SQLite's online backup tooling or snapshot storage that guarantees filesystem consistency. Copying an actively written SQLite file alone may produce an inconsistent backup.

## Upgrade procedure

1. Read the [changelog](/changelog).
2. Back up the persistent data.
3. Pull the selected image tag.
4. Recreate the container.
5. Sign in and verify dashboards, authentication, and uploaded icons.

```bash
docker compose pull
docker compose up -d
docker compose logs --tail=100 dashlit
```

Pin versioned tags when rollback predictability matters. A rollback that crosses a database migration boundary should restore the matching pre-upgrade backup rather than reuse a database modified by a newer release.

## Restore

Stop DashLit before replacing its data. Restore `bookmarks.db` and `icons/` together, verify ownership and permissions, then start the container and inspect its logs.
