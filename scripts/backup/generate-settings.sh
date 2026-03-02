#!/bin/bash
# generate-settings.sh — Generates claude/settings.json from config.json.
#
# Usage: ./scripts/generate-settings.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

OUTPUT="$LIB_DIR/claude/settings.json"

# Build enabledPlugins object: { "name@marketplace": true, ... }
PLUGINS_OBJ=$(cfg_raw '.plugins' | jq '[.[] | {key: "\(.name)@\(.marketplace)", value: true}] | from_entries')

# Build permissions allow array
PERMISSIONS_ALLOW=$(cfg_raw '.settings.permissions_allow')

# Build the full settings.json
jq -n \
  --argjson plugins "$PLUGINS_OBJ" \
  --argjson allow "$PERMISSIONS_ALLOW" \
  --arg defaultMode "$(cfg '.settings.defaultMode')" \
  --argjson skipDangerous "$(cfg '.settings.skipDangerousModePermissionPrompt')" \
  --arg statusType "$(cfg '.settings.statusLine.type')" \
  --arg statusCmd "$(cfg '.settings.statusLine.command')" \
  '{
    permissions: {
      allow: $allow,
      defaultMode: $defaultMode
    },
    statusLine: {
      type: $statusType,
      command: $statusCmd
    },
    enabledPlugins: $plugins,
    skipDangerousModePermissionPrompt: $skipDangerous,
    _generated: "DO NOT EDIT — generated from config.json by scripts/generate-settings.sh"
  }' > "${OUTPUT}.tmp" && mv "${OUTPUT}.tmp" "$OUTPUT"

echo "  Generated $OUTPUT"
