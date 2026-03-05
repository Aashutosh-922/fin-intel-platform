#!/usr/bin/env bash
set -euo pipefail

# Disk guard for local/dev deployments.
# Purpose: proactively reclaim Docker disk usage before Kafka crashes from ENOSPC.

THRESHOLD="${THRESHOLD:-85}" # percent
ROOT_MOUNT="${ROOT_MOUNT:-/}"

usage_pct() {
  df -P "$ROOT_MOUNT" | awk 'NR==2 {gsub("%","",$5); print $5}'
}

echo "[disk-guard] root mount: $ROOT_MOUNT"
echo "[disk-guard] threshold: ${THRESHOLD}%"

CURRENT="$(usage_pct)"
echo "[disk-guard] current usage: ${CURRENT}%"

if (( CURRENT < THRESHOLD )); then
  echo "[disk-guard] usage below threshold; no cleanup needed"
  exit 0
fi

echo "[disk-guard] usage above threshold; running safe Docker cleanup..."
docker builder prune -f
docker image prune -f
docker container prune -f

AFTER="$(usage_pct)"
echo "[disk-guard] usage after cleanup: ${AFTER}%"

if (( AFTER >= THRESHOLD )); then
  echo "[disk-guard] WARNING: usage still above threshold"
  echo "[disk-guard] consider: docker volume prune -f (destructive for unused volumes)"
fi

