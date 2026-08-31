.PHONY: frontend backend build run dev-backend dev-frontend seed clean git-cliff-install changelog changelog-preview release-changelog

DATA_DIR ?= ./data
export DATA_DIR

BUILD_VERSION ?= dev
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf unknown)

GIT_CLIFF_VERSION ?= 2.13.0
GIT_CLIFF_BIN := $(CURDIR)/bin/git-cliff

frontend:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	rm -rf backend/cmd/server/static
	mkdir -p backend/cmd/server/static
	cp -r frontend/dist/* backend/cmd/server/static/

backend:
	cd backend && go build -ldflags="-X main.version=$(BUILD_VERSION) -X main.commit=$(BUILD_COMMIT)" -o ../app ./cmd/server

build: frontend backend

run: build
	./app

dev-backend:
	cd backend && DEV_MODE=1 go run ./cmd/server

dev-frontend:
	cd frontend && pnpm run dev

seed:
	cd backend && go run ./cmd/seed

clean:
	rm -rf frontend/dist app data/*.db
	mkdir -p backend/cmd/server/static
	find backend/cmd/server/static -mindepth 1 ! -path backend/cmd/server/static/.gitkeep -delete
	touch backend/cmd/server/static/.gitkeep

$(GIT_CLIFF_BIN):
	@set -eu; \
	os="$$(uname -s)"; \
	arch="$$(uname -m)"; \
	case "$$arch" in \
		x86_64|amd64) arch="x86_64" ;; \
		aarch64|arm64) arch="aarch64" ;; \
		*) echo "Unsupported architecture: $$arch" >&2; exit 1 ;; \
	esac; \
	case "$$os" in \
		Linux) target="unknown-linux-musl" ;; \
		Darwin) target="apple-darwin" ;; \
		*) echo "Unsupported operating system: $$os" >&2; exit 1 ;; \
	esac; \
	archive="git-cliff-$(GIT_CLIFF_VERSION)-$$arch-$$target.tar.gz"; \
	url="https://github.com/orhun/git-cliff/releases/download/v$(GIT_CLIFF_VERSION)/$$archive"; \
	tmp_dir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp_dir"' 0 1 2 3 15; \
	echo "Downloading git-cliff v$(GIT_CLIFF_VERSION) for $$arch-$$target..."; \
	curl --fail --location --silent --show-error "$$url" --output "$$tmp_dir/$$archive"; \
	tar -xzf "$$tmp_dir/$$archive" -C "$$tmp_dir"; \
	mkdir -p "$(dir $(GIT_CLIFF_BIN))"; \
	install -m 0755 "$$tmp_dir/git-cliff-$(GIT_CLIFF_VERSION)/git-cliff" "$(GIT_CLIFF_BIN)"

git-cliff-install: $(GIT_CLIFF_BIN)
	@"$(GIT_CLIFF_BIN)" --version

changelog: $(GIT_CLIFF_BIN)
	@"$(GIT_CLIFF_BIN)" -o CHANGELOG.md

changelog-preview: $(GIT_CLIFF_BIN)
	@"$(GIT_CLIFF_BIN)" --unreleased --strip header

release-changelog: $(GIT_CLIFF_BIN)
	@version='$(VERSION)'; \
	if [ -z "$$version" ]; then \
		echo "VERSION is required (example: make release-changelog VERSION=v1.0.0)" >&2; \
		exit 1; \
	fi; \
	if ! printf '%s\n' "$$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "VERSION must be a release tag such as v1.0.0" >&2; \
		exit 1; \
	fi; \
	if grep -Fq "## [$${version#v}]" CHANGELOG.md; then \
		echo "CHANGELOG.md already contains $$version" >&2; \
		exit 1; \
	fi; \
	"$(GIT_CLIFF_BIN)" --unreleased --tag "$$version" --prepend CHANGELOG.md; \
	echo "CHANGELOG.md updated for $$version"
