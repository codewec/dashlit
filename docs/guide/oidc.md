# OIDC authentication

DashLit can authenticate users through an OpenID Connect provider such as Pocket ID. Password authentication may remain enabled as a fallback or be disabled after OIDC is working.

## Provider setup

Create an OIDC client in the provider and configure this callback URL:

```text
https://dash.example.com/api/auth/oidc/callback
```

Then provide the issuer, client ID, client secret, and matching redirect URL to DashLit:

```dotenv
OIDC_ISSUER=https://id.example.com
OIDC_CLIENT_ID=dashlit
OIDC_CLIENT_SECRET=provider-issued-secret
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
OIDC_BUTTON_TITLE=Sign in with Pocket ID
```

Restart the container and verify OIDC login before disabling password access.

## Account creation policy

With `DISABLE_OIDC_REGISTRATION=false`, an unknown OIDC identity creates a DashLit user. If the database is empty, that user becomes the first administrator just like a password-registered account.

Set `DISABLE_OIDC_REGISTRATION=true` when accounts should be provisioned in DashLit before users may sign in.

## Account linking

By default, DashLit links an incoming OIDC identity to a password account with the same normalized username. The existing account retains its password and gains OIDC login.

Set `DISABLE_OIDC_USER_MERGE=true` to prevent automatic linking. A conflicting username then receives a numeric suffix such as `alex-2`.

## Disabling password login

After a successful OIDC test, set:

```dotenv
DISABLE_PASSWORD_LOGIN=true
```

This setting is applied only while OIDC is fully configured. If the issuer or client ID is missing, DashLit retains password login as a recovery path.

## Troubleshooting

- Confirm the provider issuer URL is exact and reachable from the container.
- Ensure the registered callback matches `OIDC_REDIRECT_URL`, including scheme, host, and path.
- Forward the original host and protocol through the reverse proxy.
- Check that the server clock is synchronized; token validation depends on correct time.
- Temporarily retain password login until the complete redirect flow succeeds.
