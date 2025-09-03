#!/usr/bin/env bash
set -euo pipefail

# Change to the directory where this script lives (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Allocate a TTY only when running interactively
DOCKER_TTY=()
if [ -t 0 ] && [ -t 1 ]; then
  DOCKER_TTY=(-it)
fi

# Ensure image exists (build if missing)
if ! docker image inspect clichat:local >/dev/null 2>&1; then
  docker compose build
fi

# Run the chat client inside the container with proper TTY so TAB works
exec docker run --rm "${DOCKER_TTY[@]}" \
  --env-file .env \
  -v "$PWD":/app \
  -w /app \
  clichat:local chat "$@"


