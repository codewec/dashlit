# Releasing DashLit

This document describes the release process for maintainers. DashLit uses Conventional Commits, git-cliff, GitHub Releases, and multi-architecture images published to GHCR.

## Release model

- Development happens on the `main` branch.
- Stable versions use tags such as `v1.0.0`.
- Every push to `main` updates `ghcr.io/codewec/dashlit:dev`.
- Every release tag publishes both `ghcr.io/codewec/dashlit:main` and `ghcr.io/codewec/dashlit:<version>`.
- The `latest` image tag remains on the legacy generation and is intentionally not published by the current workflow.
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

## Create a release

### 1. Prepare the branch

Start from an up-to-date and clean `main` branch:

```bash
git switch main
git pull --ff-only origin main
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
cd ../docs
npm ci
npm run build
cd ..
```

### 3. Preview the release notes

git-cliff shows commits since the latest matching `v*` tag:

```bash
make changelog-preview
```

Review the output and choose the next semantic version.

### 4. Update `CHANGELOG.md`

Generate the changelog section for the selected version:

```bash
make release-changelog VERSION=v1.0.0
```

The command:

- validates the stable release tag format;
- refuses to add an existing version twice;
- collects commits since the latest version tag;
- prepends the new version section to `CHANGELOG.md`.

Review and commit the generated file:

```bash
git diff -- CHANGELOG.md
git add CHANGELOG.md
git commit -m "chore(release): prepare v1.0.0"
```

The `chore(release)` commit is excluded from generated release notes.

### 5. Push the release commit

```bash
git push origin main
```

This starts the container workflow and updates:

```text
ghcr.io/codewec/dashlit:dev
```

It also publishes the updated documentation when `CHANGELOG.md` changes. It does not create a GitHub Release yet.

### 6. Create and push the tag

Create the tag on the release commit that contains the updated changelog:

```bash
git tag -a v1.0.0 -m "DashLit v1.0.0"
git push origin v1.0.0
```

Pushing the tag starts two workflows:

1. The container workflow publishes `linux/amd64`, `linux/arm64`, and `linux/arm/v7` images:

   ```text
   ghcr.io/codewec/dashlit:main
   ghcr.io/codewec/dashlit:v1.0.0
   ```

2. The release workflow uses git-cliff to generate notes, builds standalone Linux binaries, and creates a GitHub Release with these assets:

   ```text
   dashlit_1.0.0_linux_amd64.tar.gz
   dashlit_1.0.0_linux_arm64.tar.gz
   dashlit_1.0.0_linux_armv7.tar.gz
   checksums.txt
   ```

The tag workflow promotes the release commit to `main` and publishes its immutable version tag. It does not publish or change `latest`.

### 7. Verify the release

```bash
docker pull ghcr.io/codewec/dashlit:v1.0.0
docker pull ghcr.io/codewec/dashlit:main
```

Confirm that both images report the expected version and architecture, the GitHub Release contains the expected notes, and the documentation shows the new changelog entry.

Download a standalone archive and verify it against `checksums.txt`. Each archive contains the `dashlit` binary and `LICENSE`.

## Next release

Choose the next version according to Semantic Versioning:

- patch for backward-compatible fixes, for example `v1.0.1`;
- minor for backward-compatible features, for example `v1.1.0`;
- major for breaking changes, for example `v2.0.0`.

## Important notes

- Never move or overwrite a published version tag. Create a new version instead.
- Do not publish or retarget `latest`; legacy users rely on it remaining on the previous generation.
- Create release tags only from commits already present on `main`.
- Back up persistent application data before testing an upgrade.
