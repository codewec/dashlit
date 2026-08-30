#!/usr/bin/env bash

set -Eeuo pipefail

readonly INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/codewec/dashlit/main/scripts/install.sh"

log() {
  printf '[DashLit LXC] %s\n' "$*"
}

fail() {
  printf '[DashLit LXC] Error: %s\n' "$*" >&2
  exit 1
}

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  fail "run this script as root on a Proxmox VE host"
fi

for command in pct pvesh pvesm pveam; do
  command -v "$command" >/dev/null 2>&1 || fail "${command} was not found; run this script on a Proxmox VE host"
done

ctid="${DASHLIT_CTID:-$(pvesh get /cluster/nextid)}"
hostname="${DASHLIT_HOSTNAME:-dashlit}"
cores="${DASHLIT_CORES:-1}"
memory="${DASHLIT_MEMORY:-512}"
disk="${DASHLIT_DISK:-4}"
bridge="${DASHLIT_BRIDGE:-vmbr0}"
ip_config="${DASHLIT_IP_CONFIG:-dhcp}"
storage="${DASHLIT_STORAGE:-$(pvesm status -content rootdir | awk 'NR > 1 && $3 == "active" { print $1; exit }')}"
template_storage="${DASHLIT_TEMPLATE_STORAGE:-$(pvesm status -content vztmpl | awk 'NR > 1 && $3 == "active" { print $1; exit }')}"

[[ "$ctid" =~ ^[0-9]+$ ]] || fail "invalid container ID: ${ctid}"
[[ "$hostname" =~ ^[a-zA-Z0-9][a-zA-Z0-9.-]*$ ]] || fail "invalid hostname: ${hostname}"
[[ "$cores" =~ ^[1-9][0-9]*$ ]] || fail "DASHLIT_CORES must be a positive integer"
[[ "$memory" =~ ^[1-9][0-9]*$ ]] || fail "DASHLIT_MEMORY must be a positive integer"
[[ "$disk" =~ ^[1-9][0-9]*$ ]] || fail "DASHLIT_DISK must be a positive integer"
[[ -n "$storage" ]] || fail "no active storage with rootdir content was found"
[[ -n "$template_storage" ]] || fail "no active storage with vztmpl content was found"

if pct status "$ctid" >/dev/null 2>&1; then
  fail "container ${ctid} already exists"
fi

if ! ip link show "$bridge" >/dev/null 2>&1; then
  fail "network bridge ${bridge} does not exist"
fi

case "$(uname -m)" in
  x86_64 | amd64) template_architecture="amd64" ;;
  aarch64 | arm64) template_architecture="arm64" ;;
  *) fail "unsupported Proxmox host architecture: $(uname -m)" ;;
esac

template="$(
  pveam list "$template_storage" |
    awk -v architecture="$template_architecture" '$1 ~ ("/debian-13-standard_.*_" architecture "\\.tar\\.(zst|gz)$") { sub(".*/", "", $1); print $1 }' |
    sort -V |
    tail -n 1
)"

if [[ -n "$template" ]]; then
  log "Using existing template ${template_storage}:vztmpl/${template}"
else
  template="$(
    pveam available --section system |
      awk -v architecture="$template_architecture" '$2 ~ ("^debian-13-standard_.*_" architecture "\\.tar\\.(zst|gz)$") { print $2 }' |
      sort -V |
      tail -n 1
  )"
  [[ -n "$template" ]] || fail "a Debian 13 ${template_architecture} LXC template was not found"
  log "Downloading ${template}"
  pveam download "$template_storage" "$template"
fi

log "Creating unprivileged container ${ctid} (${hostname})"
pct create "$ctid" "${template_storage}:vztmpl/${template}" \
  --hostname "$hostname" \
  --cores "$cores" \
  --memory "$memory" \
  --swap 0 \
  --rootfs "${storage}:${disk}" \
  --net0 "name=eth0,bridge=${bridge},ip=${ip_config}" \
  --unprivileged 1 \
  --features nesting=1 \
  --onboot 1

pct start "$ctid"

log "Waiting for the container network"
container_ip=""
for _ in {1..60}; do
  container_ip="$(pct exec "$ctid" -- hostname -I 2>/dev/null | awk '{ print $1 }' || true)"
  if [[ -n "$container_ip" ]]; then
    break
  fi
  sleep 2
done
[[ -n "$container_ip" ]] || fail "container ${ctid} did not obtain an IP address"

log "Installing prerequisites in container ${ctid}"
pct exec "$ctid" -- bash -c 'apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl'

log "Installing DashLit"
pct exec "$ctid" -- env DASHLIT_ADDR=:80 bash -c "curl -fsSL '${INSTALL_SCRIPT_URL}' | bash"

log "Enabling passwordless root login on the Proxmox console"
pct exec "$ctid" -- bash -c "install -d -m 0755 /etc/systemd/system/container-getty@1.service.d && printf '%s\n' '[Service]' 'ExecStart=' 'ExecStart=-/sbin/agetty --autologin root --noclear --keep-baud 115200,38400,9600 \$TERM' > /etc/systemd/system/container-getty@1.service.d/autologin.conf && systemctl daemon-reload && systemctl restart container-getty@1.service"

log "Installation complete"
printf 'DashLit: http://%s\n' "$container_ip"
printf 'Container: %s\n' "$ctid"
printf 'Console: open it in Proxmox or run pct enter %s (no password required)\n' "$ctid"
printf 'Update inside the container: dashlit-update\n'
