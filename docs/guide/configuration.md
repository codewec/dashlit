# Configuration

DashLit reads process environment variables and an optional `.env` file. Process variables take precedence. Container deployments should normally pass settings through Compose or the container runtime.

## Basic settings

Most container installations only need a secure `JWT_SECRET`. Docker already provides the correct internal listen address and data paths.

| Variable | Default | Purpose |
| --- | --- | --- |
| `JWT_SECRET` | Development value | Secret used to sign sessions. Always replace it in production. |
| `INITIAL_ADMIN_USERNAME` | Empty | Username used to create the first administrator automatically. |
| `INITIAL_ADMIN_PASSWORD` | Empty | Password used to create the first administrator automatically. |
| `DEV_MODE` | `false` | Enables verbose database diagnostics for development. |
| `UPDATE_CHECK_ENABLED` | `true` | Checks GitHub Releases for a newer stable DashLit version. |

Release builds contain their Git tag and commit. The installed version appears in the footer and can be inspected in a native installation with `dashlit --version`. DashLit checks GitHub only for administrator sessions and caches the result for 12 hours; regular users receive no available-release information. A failed check does not affect startup or normal operation. Set `UPDATE_CHECK_ENABLED=false` if the instance must not make this outbound request.

## Initial administrator

For unattended deployments, set both bootstrap variables before the first start:

```dotenv
INITIAL_ADMIN_USERNAME=admin
INITIAL_ADMIN_PASSWORD=replace-with-a-secure-password
```

DashLit creates this user as an administrator only when the users table is empty. The password must contain at least six characters. If only one variable is set on an empty database, startup stops with a configuration error. As soon as any user exists, both values are completely ignored: changing them cannot rename the administrator or reset its password. Remove them from Compose or `.env` after provisioning to avoid retaining the plaintext password.

This bootstrap is independent of `DISABLE_PASSWORD_REGISTRATION`, allowing registration to remain disabled in an automatically provisioned installation.

## Authentication settings

| Variable | Default | Purpose |
| --- | --- | --- |
| `OIDC_ISSUER` | Empty | OIDC issuer URL. OIDC is disabled when empty. |
| `OIDC_CLIENT_ID` | Empty | Client ID registered with the provider. |
| `OIDC_CLIENT_SECRET` | Empty | Client secret, when required by the provider. |
| `OIDC_REDIRECT_URL` | Local callback | Public URL ending in `/api/auth/oidc/callback`. |
| `OIDC_BUTTON_TITLE` | `Sign in with OIDC` | Text shown on the login button. |
| `OIDC_INSECURE_SKIP_TLS_VERIFY` | `false` | Disables TLS certificate verification for all OIDC requests. Unsafe; use only for a trusted private provider. |
| `DISABLE_PASSWORD_REGISTRATION` | `false` | Stops the creation of new password accounts. |
| `DISABLE_OIDC_REGISTRATION` | `false` | Stops unknown OIDC identities from creating accounts. |
| `DISABLE_OIDC_USER_MERGE` | `false` | Keeps matching password and OIDC identities separate. |
| `DISABLE_PASSWORD_LOGIN` | `false` | Disables password login after OIDC is fully configured. |

`OIDC_ISSUER` and `OIDC_CLIENT_ID` must either both be set or both be empty. Password login remains available as a recovery path until OIDC is fully configured, even when `DISABLE_PASSWORD_LOGIN=true`.

## Example production environment

```dotenv
JWT_SECRET=use-a-long-random-value
DEV_MODE=false

OIDC_ISSUER=https://id.example.com
OIDC_CLIENT_ID=dashlit
OIDC_CLIENT_SECRET=provider-issued-secret
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
OIDC_BUTTON_TITLE=Sign in with Pocket ID
OIDC_INSECURE_SKIP_TLS_VERIFY=false

DISABLE_PASSWORD_REGISTRATION=true
DISABLE_OIDC_REGISTRATION=false
DISABLE_OIDC_USER_MERGE=false
DISABLE_PASSWORD_LOGIN=true
```

## Advanced storage and network settings

These settings describe internal application paths and normally should not be added to a standard container deployment:

| Variable | Default | Purpose |
| --- | --- | --- |
| `ADDR` | `:8080` | Address and port listened on inside the process. |
| `DATA_DIR` | `./data` | Database, uploaded-icon, and cache directory. The image sets this to `/data`. |
| `DATABASE_PATH` | `$DATA_DIR/bookmarks.db` | Overrides the SQLite database location. |

Override them only for a custom runtime, filesystem layout, or source installation. When overriding `DATABASE_PATH`, keep its parent directory writable. Legacy migration looks for `dashboard.json` in that same parent directory.

## Container permissions

The published image runs as the unprivileged user with UID and GID `10001`. Named Docker volumes are initialized automatically. For a bind mount, the simplest option is to allow writing for everyone:

```bash
mkdir -p ./data
sudo chmod -R 777 ./data
```

For more restrictive permissions, assign the directory to the container's numeric UID. Note that your regular host user may then lose write access:

```bash
sudo chown -R 10001:10001 ./data
sudo chmod -R 750 ./data
```

Do not enable `DEV_MODE` in production unless you are actively diagnosing a problem; it produces verbose database logging.

## Testing the update notification locally

The two override variables below are intended only for development and are not needed in production:

```bash
DASHLIT_VERSION_OVERRIDE=v1.0.0 \
UPDATE_CHECK_LATEST_VERSION=v9.9.9 \
docker compose -f docker-compose-local.yml up --build
```

This makes the local instance behave as version `v1.0.0` with `v9.9.9` available, without contacting GitHub. Open a regular dashboard to check the footer, then enable clean mode to check the burger-menu chip and release item.
