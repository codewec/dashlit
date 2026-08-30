#!/usr/bin/env bash

set -Eeuo pipefail

readonly REPOSITORY="codewec/dashlit"
readonly INSTALL_DIR="/opt/dashlit"
readonly DATA_DIR="/var/lib/dashlit"
readonly CONFIG_DIR="/etc/dashlit"
readonly CONFIG_FILE="${CONFIG_DIR}/dashlit.env"
readonly SERVICE_FILE="/etc/systemd/system/dashlit.service"
readonly UPDATE_COMMAND="/usr/local/bin/dashlit-update"
readonly RELEASE_VERSION="${DASHLIT_VERSION:-latest}"
readonly LISTEN_ADDR="${DASHLIT_ADDR:-:8080}"
readonly DOWNLOAD_BASE="https://github.com/${REPOSITORY}/releases"

temporary_directory=""

log() {
  printf '[DashLit] %s\n' "$*"
}

fail() {
  printf '[DashLit] Error: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n "$temporary_directory" && -d "$temporary_directory" ]]; then
    rm -rf -- "$temporary_directory"
  fi
}

trap cleanup EXIT

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  fail "run this script as root"
fi

if ! command -v systemctl >/dev/null 2>&1; then
  fail "systemd is required"
fi

install_dependencies() {
  local missing=()
  local command

  for command in curl tar sha256sum openssl; do
    command -v "$command" >/dev/null 2>&1 || missing+=("$command")
  done

  if [[ ${#missing[@]} -eq 0 ]]; then
    return
  fi

  log "Installing required tools: ${missing[*]}"
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl tar coreutils openssl
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y ca-certificates curl tar coreutils openssl
  elif command -v yum >/dev/null 2>&1; then
    yum install -y ca-certificates curl tar coreutils openssl
  elif command -v apk >/dev/null 2>&1; then
    apk add --no-cache ca-certificates curl tar coreutils openssl
  else
    fail "install curl, tar, sha256sum, and openssl, then run the script again"
  fi
}

resolve_architecture() {
  case "$(uname -m)" in
    x86_64 | amd64) printf 'amd64\n' ;;
    aarch64 | arm64) printf 'arm64\n' ;;
    armv7l | armv7) printf 'armv7\n' ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

resolve_release_tag() {
  local effective_url

  if [[ "$RELEASE_VERSION" != "latest" ]]; then
    if [[ "$RELEASE_VERSION" == v* ]]; then
      printf '%s\n' "$RELEASE_VERSION"
    else
      printf 'v%s\n' "$RELEASE_VERSION"
    fi
    return
  fi

  effective_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' "${DOWNLOAD_BASE}/latest")"
  [[ "$effective_url" == */tag/* ]] || fail "could not determine the latest release"
  printf '%s\n' "${effective_url##*/}"
}

install_dependencies

architecture="$(resolve_architecture)"
release_tag="$(resolve_release_tag)"
version="${release_tag#v}"
asset="dashlit_${version}_linux_${architecture}.tar.gz"
release_url="${DOWNLOAD_BASE}/download/${release_tag}"
temporary_directory="$(mktemp -d /tmp/dashlit-install.XXXXXX)"

log "Downloading DashLit ${release_tag} for linux/${architecture}"
curl -fsSL --retry 3 -o "${temporary_directory}/${asset}" "${release_url}/${asset}"
curl -fsSL --retry 3 -o "${temporary_directory}/checksums.txt" "${release_url}/checksums.txt"

expected_checksum="$(awk -v asset="$asset" '$2 == asset || $2 == ("./" asset) { print $1; exit }' "${temporary_directory}/checksums.txt")"
[[ "$expected_checksum" =~ ^[0-9a-fA-F]{64}$ ]] || fail "checksum for ${asset} is missing from checksums.txt"
printf '%s  %s\n' "$expected_checksum" "${temporary_directory}/${asset}" | sha256sum --check --status || fail "checksum verification failed"

mkdir -p "${temporary_directory}/release"
tar -xzf "${temporary_directory}/${asset}" -C "${temporary_directory}/release"
[[ -f "${temporary_directory}/release/dashlit" ]] || fail "release archive does not contain the dashlit binary"

if ! getent group dashlit >/dev/null 2>&1; then
  groupadd --system dashlit
fi

if ! id dashlit >/dev/null 2>&1; then
  useradd --system --gid dashlit --home-dir "$DATA_DIR" --shell "$(command -v nologin || printf '/usr/sbin/nologin')" dashlit
else
  usermod --append --groups dashlit dashlit
fi

install -d -m 0755 "$INSTALL_DIR" "$CONFIG_DIR"
install -d -o dashlit -g dashlit -m 0750 "$DATA_DIR"

if [[ ! -f "$CONFIG_FILE" ]]; then
  jwt_secret="$(openssl rand -hex 32)"
  cat <<EOF >"$CONFIG_FILE"
ADDR=${LISTEN_ADDR}
DATA_DIR=${DATA_DIR}
JWT_SECRET=${jwt_secret}
EOF
  chown root:dashlit "$CONFIG_FILE"
  chmod 0640 "$CONFIG_FILE"
  log "Created ${CONFIG_FILE}"
fi

previous_binary=""
if [[ -x "${INSTALL_DIR}/dashlit" ]]; then
  previous_binary="${temporary_directory}/dashlit.previous"
  cp -a "${INSTALL_DIR}/dashlit" "$previous_binary"
  systemctl stop dashlit 2>/dev/null || true
fi

install -m 0755 "${temporary_directory}/release/dashlit" "${INSTALL_DIR}/dashlit"

cat <<EOF >"$SERVICE_FILE"
[Unit]
Description=DashLit dashboard
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=dashlit
Group=dashlit
WorkingDirectory=${INSTALL_DIR}
EnvironmentFile=${CONFIG_FILE}
ExecStart=${INSTALL_DIR}/dashlit
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=${DATA_DIR}
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

cat <<'EOF' >"$UPDATE_COMMAND"
#!/usr/bin/env bash

set -Eeuo pipefail

readonly INSTALL_DIR="/opt/dashlit"
readonly DOWNLOAD_BASE="https://github.com/codewec/dashlit/releases"

temporary_directory=""

log() {
  printf '[DashLit] %s\n' "$*"
}

fail() {
  printf '[DashLit] Error: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n "$temporary_directory" && -d "$temporary_directory" ]]; then
    rm -rf -- "$temporary_directory"
  fi
}

trap cleanup EXIT

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  fail "run dashlit-update as root"
fi

case "$(uname -m)" in
  x86_64 | amd64) architecture="amd64" ;;
  aarch64 | arm64) architecture="arm64" ;;
  armv7l | armv7) architecture="armv7" ;;
  *) fail "unsupported architecture: $(uname -m)" ;;
esac

effective_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' "${DOWNLOAD_BASE}/latest")"
[[ "$effective_url" == */tag/* ]] || fail "could not determine the latest release"
release_tag="${effective_url##*/}"

if [[ -f "${INSTALL_DIR}/VERSION" ]] && [[ "$(<"${INSTALL_DIR}/VERSION")" == "$release_tag" ]]; then
  log "DashLit ${release_tag} is already installed"
  exit 0
fi

version="${release_tag#v}"
asset="dashlit_${version}_linux_${architecture}.tar.gz"
release_url="${DOWNLOAD_BASE}/download/${release_tag}"
temporary_directory="$(mktemp -d /tmp/dashlit-update.XXXXXX)"

log "Downloading DashLit ${release_tag} for linux/${architecture}"
curl -fsSL --retry 3 -o "${temporary_directory}/${asset}" "${release_url}/${asset}"
curl -fsSL --retry 3 -o "${temporary_directory}/checksums.txt" "${release_url}/checksums.txt"

expected_checksum="$(awk -v asset="$asset" '$2 == asset || $2 == ("./" asset) { print $1; exit }' "${temporary_directory}/checksums.txt")"
[[ "$expected_checksum" =~ ^[0-9a-fA-F]{64}$ ]] || fail "checksum for ${asset} is missing from checksums.txt"
printf '%s  %s\n' "$expected_checksum" "${temporary_directory}/${asset}" | sha256sum --check --status || fail "checksum verification failed"

mkdir -p "${temporary_directory}/release"
tar -xzf "${temporary_directory}/${asset}" -C "${temporary_directory}/release"
[[ -f "${temporary_directory}/release/dashlit" ]] || fail "release archive does not contain the dashlit binary"
[[ -x "${INSTALL_DIR}/dashlit" ]] || fail "DashLit is not installed in ${INSTALL_DIR}"

cp -a "${INSTALL_DIR}/dashlit" "${temporary_directory}/dashlit.previous"
systemctl stop dashlit
install -m 0755 "${temporary_directory}/release/dashlit" "${INSTALL_DIR}/dashlit"

service_started=false
if systemctl start dashlit; then
  sleep 2
  if systemctl is-active --quiet dashlit; then
    service_started=true
  fi
fi

if [[ "$service_started" != true ]]; then
  log "The new version failed to start; restoring the previous binary"
  install -m 0755 "${temporary_directory}/dashlit.previous" "${INSTALL_DIR}/dashlit"
  systemctl restart dashlit || true
  fail "the update failed; inspect logs with: journalctl -u dashlit"
fi

printf '%s\n' "$release_tag" >"${INSTALL_DIR}/VERSION"
log "DashLit was updated to ${release_tag}"
EOF
chmod 0755 "$UPDATE_COMMAND"

systemctl daemon-reload
systemctl enable dashlit >/dev/null

service_started=false
if systemctl restart dashlit; then
  sleep 2
  if systemctl is-active --quiet dashlit; then
    service_started=true
  fi
fi

if [[ "$service_started" != true ]]; then
  if [[ -n "$previous_binary" && -f "$previous_binary" ]]; then
    log "The new version failed to start; restoring the previous binary"
    install -m 0755 "$previous_binary" "${INSTALL_DIR}/dashlit"
    systemctl restart dashlit || true
  fi
  fail "DashLit failed to start; inspect logs with: journalctl -u dashlit"
fi

printf '%s\n' "$release_tag" >"${INSTALL_DIR}/VERSION"
log "DashLit ${release_tag} is running on ${LISTEN_ADDR}"
log "Configuration: ${CONFIG_FILE}"
log "Data: ${DATA_DIR}"
log "Update command: dashlit-update"
