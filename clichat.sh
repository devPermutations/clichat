#!/usr/bin/env bash
set -euo pipefail

# Change to the directory where this script lives (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Forward any extra args to the chat command
docker compose run --rm clichat chat "$@"


