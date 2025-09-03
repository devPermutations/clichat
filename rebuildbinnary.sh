#!/usr/bin/env bash
set -euo pipefail

# Rebuilds the clichat binary after checking for upstream updates.
# Upstream: https://github.com/devPermutations/clichat

# Move to project root (directory of this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

REPO_URL="https://github.com/devPermutations/clichat"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "[error] Not a git repository. Clone $REPO_URL first."
  exit 1
fi

BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"

echo "[rebuild] Checking remote $REPO_URL ($BRANCH)..."
REMOTE_HEAD="$(git ls-remote "$REPO_URL" "refs/heads/$BRANCH" | awk '{print $1}')"
LOCAL_HEAD="$(git rev-parse HEAD)"

if [[ -n "${REMOTE_HEAD:-}" && "$REMOTE_HEAD" != "$LOCAL_HEAD" ]]; then
  echo "[rebuild] Updates found. Pulling latest..."
  if git remote get-url origin >/dev/null 2>&1; then
    git pull --ff-only origin "$BRANCH"
  else
    git fetch "$REPO_URL" "$BRANCH"
    git merge --ff-only FETCH_HEAD
  fi
else
  echo "[rebuild] Already up to date."
fi

echo "[rebuild] Building binary..."
mkdir -p bin
OUT="bin/clichat"
GOOS="$(go env GOOS)"
if [[ "$GOOS" == "windows" ]]; then
  OUT="bin/clichat.exe"
fi
go build -trimpath -ldflags "-s -w" -o "$OUT" ./cmd/clichat

echo "[rebuild] Built $OUT"
echo "[rebuild] Run with:"
if [[ "$GOOS" == "windows" ]]; then
  echo "  $OUT chat"
else
  echo "  ./$OUT chat"
fi


