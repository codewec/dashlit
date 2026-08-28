# Migrate from legacy DashLit

Legacy DashLit releases stored the dashboard in a `dashboard.json` file. The current release stores users, dashboards, groups, and items in SQLite and can import the old file during first-user setup.

## Before you begin

1. Stop the legacy container.
2. Back up the complete legacy data directory.
3. Confirm that `dashboard.json` is valid and preserve the original copy until migration is verified.
4. Start the new DashLit release with an empty database.

The automatic migration is offered only when there are no users in the database. This prevents an old file from unexpectedly changing an established installation.

## Place the file

Put `dashboard.json` in the same directory as the SQLite database.

For the standard container configuration, both resolve under `/data`:

```text
/data/
├── dashboard.json     # legacy input
└── bookmarks.db       # created by current DashLit
```

Example Compose mount using an existing host directory:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:beta
    volumes:
      - ./data:/data
```

## Replace the legacy Compose service

You can update the old `docker-compose.yml` in place instead of creating a separate deployment. Keep the existing data mount that contains `dashboard.json`, replace the legacy image and obsolete settings with the current service definition, and add a persistent `JWT_SECRET`.

For example, if the old host directory is `./data`, the relevant service can be replaced with:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:beta
    container_name: dashlit
    restart: unless-stopped
    ports:
      - '3000:8080'
    environment:
      JWT_SECRET: '${JWT_SECRET}'
    volumes:
      - ./data:/data
```

Create a `.env` file beside the Compose file with a long random `JWT_SECRET`. Before running `docker compose up -d`, save a copy of the original Compose file and the complete data directory. Do not replace the existing data mount with a new empty named volume, or the new container will not see `dashboard.json`.

If `DATABASE_PATH` is customized, place the JSON beside that file rather than directly under `DATA_DIR`.

## Run the migration

1. Open the DashLit login page.
2. Confirm that **Legacy version data found** appears below the login form.
3. Enable the import switch.
4. Create the first user with a password, or continue through OIDC.
5. DashLit creates a private dashboard named **Legacy dashboard**, owned by the first user and selected as their personal default.
6. Verify groups, links, URLs, descriptions, and icons.

The selection is carried through OIDC in a short-lived, HttpOnly cookie. No migration choice or dashboard data is stored in browser local storage.

<div class="screenshot-placeholder">
  <div><strong>Legacy migration placeholder</strong><br>Add a screenshot of the detection message and import switch.</div>
</div>

## What is migrated

The converter imports groups in their original order and links in each group's original order. It preserves group titles and descriptions plus link titles, descriptions, URLs, and icons.

Legacy presentation fields that have no equivalent in the current data model, such as per-icon colors, URL display modes, and link targets, are not imported.

## If the migration option does not appear

- Confirm that `dashboard.json` and `bookmarks.db` have the same parent directory.
- Confirm the file is readable by container UID/GID `10001`.
- Check server logs for an invalid or unsupported legacy JSON message.
- Confirm no account has already been created. Detection occurs only during server startup while the users table is empty.

If a user was created before the file was placed correctly, stop DashLit, restore the empty pre-registration database or begin with a new empty data directory, place the legacy file, and start again. Do not delete a populated database without a verified backup.

## After verification

Once the imported dashboard is verified and the new data directory is backed up, `dashboard.json` is no longer needed by the running installation. Keep an archival copy until you are confident the migration is complete.
