#!/usr/bin/env bash
set -euo pipefail

# Change to the directory where this script lives (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[rebuild] Pulling latest from origin/main..."
git pull origin main

echo "[rebuild] Stopping/removing containers..."
docker compose down

echo "[rebuild] Building images (pulling latest base layers)..."
docker compose build --pull

echo "[rebuild] Done. You can now run: docker compose run --rm clichat chat"



