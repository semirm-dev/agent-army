#!/bin/bash
# Checks content parity between CLAUDE.md and Cursor .mdc files.
# Extracts shared sections and diffs them, flagging any drift.
#
# Usage: ./scripts/check-sync.sh

set -euo pipefail

LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLAUDE="$LIB_DIR/claude/CLAUDE.md"
CURSOR_DIR="$LIB_DIR/cursor"

DRIFT_FOUND=0

# Extract lines between two patterns (exclusive of the marker lines themselves)
extract_section() {
  local file="$1"
  local start_pattern="$2"
  local end_pattern="$3"
  # Print lines between start and end patterns (non-inclusive).
  # Strips leading '#' from heading markers so different heading levels don't cause false drift.
  awk "/$start_pattern/{found=1; next} /$end_pattern/{found=0} found" "$file" \
    | sed 's/^#\{1,6\} //'
}

diff_sections() {
  local label="$1"
  local file_a="$2"
  local section_a="$3"
  local file_b="$4"
  local section_b="$5"
  local end_a="${6:-^$}"
  local end_b="${7:-^$}"
  local normalize="${8:-}"  # optional sed expression to normalize platform-specific terms

  local tmp_a tmp_b
  tmp_a=$(mktemp)
  tmp_b=$(mktemp)

  if [ -n "$normalize" ]; then
    extract_section "$file_a" "$section_a" "$end_a" | grep -v '^[[:space:]]*$' | sed "$normalize" > "$tmp_a"
    extract_section "$file_b" "$section_b" "$end_b" | grep -v '^[[:space:]]*$' | sed "$normalize" > "$tmp_b"
  else
    extract_section "$file_a" "$section_a" "$end_a" | grep -v '^[[:space:]]*$' > "$tmp_a"
    extract_section "$file_b" "$section_b" "$end_b" | grep -v '^[[:space:]]*$' > "$tmp_b"
  fi

  if ! diff -q "$tmp_a" "$tmp_b" > /dev/null 2>&1; then
    echo "DRIFT in [$label]"
    echo "  Between: $file_a  ($section_a)"
    echo "  And:     $file_b  ($section_b)"
    diff --unified=2 "$tmp_a" "$tmp_b" | head -40
    echo ""
    DRIFT_FOUND=1
  fi

  rm -f "$tmp_a" "$tmp_b"
}

echo "=== Checking sync between CLAUDE.md and Cursor .mdc files ==="
echo ""

# 1. Coding Patterns: CLAUDE.md vs 100-golang.mdc
diff_sections \
  "Coding Patterns" \
  "$CLAUDE" "💻 Coding Patterns" \
  "$CURSOR_DIR/100-golang.mdc" "💻 Coding Patterns" \
  "🧪 Testing" "🧪 Testing"

# 2. Testing & Quality: CLAUDE.md vs 100-golang.mdc
diff_sections \
  "Testing & Quality" \
  "$CLAUDE" "🧪 Testing & Quality" \
  "$CURSOR_DIR/100-golang.mdc" "🧪 Testing & Quality" \
  "^$" "^$"

# 3. Safety section: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Deletion & Safety" \
  "$CLAUDE" "🛡️ Deletion & Safety" \
  "$CURSOR_DIR/000-index.mdc" "🛡️ Deletion & Safety" \
  "🤖" "🤖"

# 4. Communication: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Communication Style" \
  "$CLAUDE" "🛠️ Communication Style" \
  "$CURSOR_DIR/000-index.mdc" "🛠️ Communication Style" \
  "^---" "^$"

# 5. Planning: CLAUDE.md vs 200-planning.mdc
# Compare the full planning section. Use 💻 as the end marker for CLAUDE.md
# (next section after planning) and a sentinel for 200-planning.mdc (runs to EOF).
# Normalize known platform terms: "sub-agents" (Claude) vs "agents/cycles" (Cursor).
diff_sections \
  "Planning" \
  "$CLAUDE" "Agentic Implementation Plan" \
  "$CURSOR_DIR/200-planning.mdc" "Agentic Implementation Plan" \
  "💻 Coding Patterns" "ZZZZZ_SENTINEL_EOF" \
  "s/sub-agents/agents/g;s/agents\/cycles/agents/g"

if [ "$DRIFT_FOUND" -eq 0 ]; then
  echo "All sections in sync."
else
  echo "=== Drift detected. Update the source files to restore parity. ==="
  exit 1
fi
