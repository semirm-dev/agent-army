#!/bin/bash
# lib.sh — Shared functions for agent-rules scripts.
# Source this file at the top of any script that reads config.json.
#
# Provides:
#   LIB_DIR       — Absolute path to repo root
#   CONFIG_FILE   — Absolute path to config.json
#   require_jq()  — Fail-fast if jq is not installed
#   cfg '.path'   — Read string value from config.json (one per line)
#   cfg_raw '.path' — Read raw JSON value from config.json

# Resolve repo root (works whether sourced from scripts/ or elsewhere)
if [ -z "${LIB_DIR:-}" ]; then
  LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

CONFIG_FILE="$LIB_DIR/config.json"

require_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq is required but not installed."
    echo "  macOS:  brew install jq"
    echo "  Linux:  sudo apt-get install jq"
    exit 1
  fi
}

# Read string values from config.json (one per line for arrays, single value for scalars)
cfg() {
  jq -r "$1" "$CONFIG_FILE"
}

# Read raw JSON from config.json (preserves structure)
cfg_raw() {
  jq -c "$1" "$CONFIG_FILE"
}
