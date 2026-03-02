#!/bin/bash
# sync-plugins.sh — Register marketplaces and install Claude Code plugins from config.json.
#
# Idempotent: skips marketplaces already registered and plugins already installed.
# Requires: claude CLI, jq
#
# Usage: ./scripts/sync-plugins.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

if ! command -v claude >/dev/null 2>&1; then
  echo "ERROR: claude CLI not found in PATH"
  exit 1
fi

REGISTERED=$(claude plugin marketplace list 2>/dev/null || true)

echo "🔌 Syncing Claude Code plugins from config.json..."

# Phase 1: Register any marketplaces that have a 'source' field
while read -r plugin_json; do
  source_repo=$(echo "$plugin_json" | jq -r '.source // empty')
  marketplace=$(echo "$plugin_json" | jq -r '.marketplace')

  if [ -z "$source_repo" ]; then
    continue
  fi

  if echo "$REGISTERED" | grep -q "$marketplace"; then
    echo "  ✓ marketplace $marketplace already registered"
  else
    echo "  + adding marketplace $marketplace ($source_repo)..."
    if claude plugin marketplace add "$source_repo" 2>&1; then
      echo "  ✓ $marketplace added"
    else
      echo "  ✗ failed to add $marketplace — install its plugins manually"
    fi
  fi
done < <(cfg_raw '.plugins[]')

# Phase 2: Install each plugin
while read -r plugin_json; do
  name=$(echo "$plugin_json" | jq -r '.name')
  marketplace=$(echo "$plugin_json" | jq -r '.marketplace')
  qualified="${name}@${marketplace}"

  if claude plugin list --scope user 2>/dev/null | grep -q "$qualified"; then
    echo "  ✓ $qualified already installed"
  else
    echo "  + installing $qualified..."
    if claude plugin install "$qualified" --scope user 2>&1; then
      echo "  ✓ $qualified installed"
    else
      echo "  ✗ failed to install $qualified"
    fi
  fi
done < <(cfg_raw '.plugins[]')

echo "🎉 Plugin sync complete."
