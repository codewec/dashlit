# Releasing DashLit

This document describes the beta release process for maintainers. DashLit uses Conventional Commits, git-cliff, GitHub prereleases, and multi-architecture images published to GHCR.

## Release model

- Development happens on the `beta` branch.
- Beta versions use tags such as `v1.0.0-beta.1`.
- Every push to `beta` updates `ghcr.io/codewec/dashlit:beta`.
- Every prerelease tag publishes both `:<version>` and `:beta`.
- The `latest` image tag is intentionally not published during beta.
- Git tags are the source of truth for release versions.

## Prerequisites

You need Git, Make, curl, and tar. The Makefile downloads the pinned git-cliff binary to `./bin/git-cliff` automatically. You can install it ahead of time with:

```bash
make git-cliff-install
```

Use Conventional Commit messages while developing, for example:

```text
feat(front): add dashboard search
fix(back): validate imported item URLs
docs: update Docker instructions
```

## Create a beta release

### 1. Prepare the branch

Start from an up-to-date and clean `beta` branch:

```bash
git switch beta
git pull --ff-only origin beta
git status
```

Commit or discard any unrelated local changes before continuing.

### 2. Run the checks

```bash
cd backend
go test ./...
cd ../frontend
pnpm install --frozen-lockfile
pnpm exec svelte-check
pnpm run build
cd ..
```

### 3. Preview the release notes

git-cliff shows commits since the latest matching `v*` tag:

```bash
make changelog-preview
```

Review the output before choosing the version.

### 4. Update `CHANGELOG.md`

Choose the next prerelease version and generate its changelog section:

```bash
make release-changelog VERSION=v1.0.0-beta.1
```

The command:

- validates the prerelease tag format;
- refuses to add an existing version twice;
- collects commits since the latest version tag;
- prepends the new version section to `CHANGELOG.md`.

Review and commit the generated file:

```bash
git diff -- CHANGELOG.md
git add CHANGELOG.md
git commit -m "chore(release): prepare v1.0.0-beta.1"
```

The `chore(release)` commit is excluded from generated release notes.

### 5. Push the release commit

```bash
git push origin beta
```

This starts the container workflow and updates:

```text
ghcr.io/codewec/dashlit:beta
```

It does not create a GitHub Release.

### 6. Create and push the tag

Create the tag on the release commit that contains the updated changelog:

```bash
git tag -a v1.0.0-beta.1 -m "DashLit v1.0.0-beta.1"
git push origin v1.0.0-beta.1
```

Pushing the tag starts two workflows:

1. The container workflow publishes `linux/amd64` and `linux/arm64` images:

   ```text
   ghcr.io/codewec/dashlit:v1.0.0-beta.1
   ghcr.io/codewec/dashlit:beta
   ```

2. The release workflow uses git-cliff to generate notes and creates a GitHub prerelease.

Follow both workflows on the repository's Actions page. The release and container jobs run independently, so the GitHub prerelease may appear shortly before the image finishes publishing.

### 7. Verify the release

```bash
docker pull ghcr.io/codewec/dashlit:v1.0.0-beta.1
docker pull ghcr.io/codewec/dashlit:beta
```

Check that the GitHub Release is marked as a prerelease and contains the expected notes.

## Next beta

For the next beta of the same base version, increment the prerelease number:

```text
v1.0.0-beta.2
v1.0.0-beta.3
```

Use a new base version when appropriate, for example `v1.1.0-beta.1`.

## Important notes

- Never move or overwrite a published version tag. Create a new prerelease version instead.
- Do not create tags without the prerelease suffix while the project is in beta.
- Do not publish `latest` manually during the beta phase.
- Back up persistent application data before testing an upgrade.
