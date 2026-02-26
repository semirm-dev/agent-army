#!/bin/bash
# generate-claude.sh — Regenerates auto-generated sections of CLAUDE.md from config.json.
# Replaces content between marker comment pairs, leaving the rest untouched.
#
# Marker pairs:
#   <!-- BEGIN:agent-definitions -->  / <!-- END:agent-definitions -->
#   <!-- BEGIN:custom-skills -->      / <!-- END:custom-skills -->
#   <!-- BEGIN:sync-pairs-table -->   / <!-- END:sync-pairs-table -->

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

CLAUDE_MD="$LIB_DIR/claude/CLAUDE.md"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# --- Generate agent definitions ---
jq -r '
  .agents[] |
  "  - **\(.group):** " +
  ([.roles[] | "`\(.name).md`" + (if .notes then " \(.notes)" else "" end)] | join(", "))
' "$CONFIG_FILE" > "$TMP_DIR/agents.txt"

# --- Generate custom skills ---
{
  jq -r '
    .custom_skills[] |
    "  - `\(.name)` -- Invoke when \(.invoke_when)."
  ' "$CONFIG_FILE"
  echo '  - _(Add languages: create `<lang>-coder.md`, `<lang>-reviewer.md`, `<lang>-tester.md`)_'
} > "$TMP_DIR/skills.txt"

# --- Generate sync pairs table ---
{
  echo '| Rule File | Synced With | Content |'
  echo '|-----------|-------------|---------|'
  jq -r '
    .sync_pairs[] |
    "| `rules/\(.rule)` | `cursor/\(.cursor)` | \(.description) |"
  ' "$CONFIG_FILE"
} > "$TMP_DIR/sync.txt"

# --- Replace content between markers ---
# Replaces all lines between BEGIN and END markers (exclusive) with content from a file.
replace_between_markers() {
  local input="$1"
  local begin_marker="$2"
  local end_marker="$3"
  local content_file="$4"
  local output="$5"

  awk -v begin="$begin_marker" -v end="$end_marker" -v cf="$content_file" '
    index($0, begin) {
      print
      while ((getline line < cf) > 0) print line
      skip = 1
      next
    }
    index($0, end) {
      skip = 0
    }
    !skip { print }
  ' "$input" > "$output"
}

# Chain replacements through temp files
cp "$CLAUDE_MD" "$TMP_DIR/step0.md"

replace_between_markers "$TMP_DIR/step0.md" \
  "<!-- BEGIN:agent-definitions -->" \
  "<!-- END:agent-definitions -->" \
  "$TMP_DIR/agents.txt" \
  "$TMP_DIR/step1.md"

replace_between_markers "$TMP_DIR/step1.md" \
  "<!-- BEGIN:custom-skills -->" \
  "<!-- END:custom-skills -->" \
  "$TMP_DIR/skills.txt" \
  "$TMP_DIR/step2.md"

replace_between_markers "$TMP_DIR/step2.md" \
  "<!-- BEGIN:sync-pairs-table -->" \
  "<!-- END:sync-pairs-table -->" \
  "$TMP_DIR/sync.txt" \
  "$TMP_DIR/step3.md"

cp "$TMP_DIR/step3.md" "$CLAUDE_MD"
echo "✓ CLAUDE.md sections regenerated from config.json"

# --- Also regenerate cursor/000-index.mdc if it has markers ---
CURSOR_INDEX="$LIB_DIR/cursor/000-index.mdc"
if grep -q "BEGIN:agent-definitions" "$CURSOR_INDEX" 2>/dev/null; then
  cp "$CURSOR_INDEX" "$TMP_DIR/cursor0.mdc"

  replace_between_markers "$TMP_DIR/cursor0.mdc" \
    "<!-- BEGIN:agent-definitions -->" \
    "<!-- END:agent-definitions -->" \
    "$TMP_DIR/agents.txt" \
    "$TMP_DIR/cursor1.mdc"

  replace_between_markers "$TMP_DIR/cursor1.mdc" \
    "<!-- BEGIN:custom-skills -->" \
    "<!-- END:custom-skills -->" \
    "$TMP_DIR/skills.txt" \
    "$TMP_DIR/cursor2.mdc"

  cp "$TMP_DIR/cursor2.mdc" "$CURSOR_INDEX"
  echo "✓ cursor/000-index.mdc sections regenerated from config.json"
fi
