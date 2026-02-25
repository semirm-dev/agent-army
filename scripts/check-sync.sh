#!/bin/bash
# Checks content parity between claude/rules/*.md and Cursor .mdc files,
# plus core CLAUDE.md sections vs 000-index.mdc and 200-planning.mdc.
#
# Usage: ./scripts/check-sync.sh

set -euo pipefail

LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLAUDE="$LIB_DIR/claude/CLAUDE.md"
RULES_DIR="$LIB_DIR/claude/rules"
CURSOR_DIR="$LIB_DIR/cursor"

DRIFT_FOUND=0
DEPLOYED_DIR="$HOME/.claude"
CHECK_DEPLOYED=0

# Parse flags
for arg in "$@"; do
  case "$arg" in
    --deployed) CHECK_DEPLOYED=1 ;;
  esac
done

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

# Diff two sections extracted from files
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
    echo "  Between: $file_a"
    echo "  And:     $file_b"
    # Show file timestamps for sync direction hints
    if [ -f "$file_a" ] && [ -f "$file_b" ]; then
      local mod_a mod_b
      mod_a=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$file_a" 2>/dev/null || stat -c "%y" "$file_a" 2>/dev/null | cut -d. -f1)
      mod_b=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$file_b" 2>/dev/null || stat -c "%y" "$file_b" 2>/dev/null | cut -d. -f1)
      echo "  $file_a last modified: $mod_a"
      echo "  $file_b last modified: $mod_b"
    fi
    diff --unified=2 "$tmp_a" "$tmp_b" | head -40
    echo ""
    DRIFT_FOUND=1
  fi

  rm -f "$tmp_a" "$tmp_b"
}

# Diff entire pattern files (after stripping comments and headings)
diff_rule_file() {
  local label="$1"
  local rule_file="$2"
  local cursor_file="$3"
  local normalize="${4:-}"

  local tmp_a tmp_b
  tmp_a=$(mktemp)
  tmp_b=$(mktemp)

  # Strip sync comments, front matter, and normalize headings
  if [ -n "$normalize" ]; then
    grep -v '^<!-- Sync:' "$rule_file" | grep -v '^[[:space:]]*$' | sed 's/^#\{1,6\} //' | sed "$normalize" > "$tmp_a"
    sed -n '/^---$/,/^---$/!p' "$cursor_file" | grep -v '^<!-- Sync:' | grep -v '^[[:space:]]*$' | sed 's/^#\{1,6\} //' | sed "$normalize" > "$tmp_b"
  else
    grep -v '^<!-- Sync:' "$rule_file" | grep -v '^[[:space:]]*$' | sed 's/^#\{1,6\} //' > "$tmp_a"
    sed -n '/^---$/,/^---$/!p' "$cursor_file" | grep -v '^<!-- Sync:' | grep -v '^[[:space:]]*$' | sed 's/^#\{1,6\} //' > "$tmp_b"
  fi

  if ! diff -q "$tmp_a" "$tmp_b" > /dev/null 2>&1; then
    echo "DRIFT in [$label]"
    echo "  Between: $rule_file"
    echo "  And:     $cursor_file"
    # Show file timestamps for sync direction hints
    local mod_a mod_b
    mod_a=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$rule_file" 2>/dev/null || stat -c "%y" "$rule_file" 2>/dev/null | cut -d. -f1)
    mod_b=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$cursor_file" 2>/dev/null || stat -c "%y" "$cursor_file" 2>/dev/null | cut -d. -f1)
    echo "  $rule_file last modified: $mod_a"
    echo "  $cursor_file last modified: $mod_b"
    echo "  Hint: update the older file to match the newer one"
    diff --unified=2 "$tmp_a" "$tmp_b" | head -40
    echo ""
    DRIFT_FOUND=1
  fi

  rm -f "$tmp_a" "$tmp_b"
}

echo "=== Checking sync between rules and Cursor .mdc files ==="
echo ""

# 1. Go patterns: rules/go-patterns.md vs 100-golang.mdc
diff_rule_file \
  "Go Patterns" \
  "$RULES_DIR/go-patterns.md" \
  "$CURSOR_DIR/100-golang.mdc"

# 2. TypeScript patterns: rules/ts-patterns.md vs 101-typescript.mdc
diff_rule_file \
  "TypeScript Patterns" \
  "$RULES_DIR/ts-patterns.md" \
  "$CURSOR_DIR/101-typescript.mdc"

# 3. Python patterns: rules/py-patterns.md vs 102-python.mdc
diff_rule_file \
  "Python Patterns" \
  "$RULES_DIR/py-patterns.md" \
  "$CURSOR_DIR/102-python.mdc"

# 4. Git workflow: rules/git-workflow.md vs 300-git.mdc
diff_rule_file \
  "Git Workflow" \
  "$RULES_DIR/git-workflow.md" \
  "$CURSOR_DIR/300-git.mdc"

# 5. Safety section: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Deletion & Safety" \
  "$CLAUDE" "🛡️ Deletion & Safety" \
  "$CURSOR_DIR/000-index.mdc" "🛡️ Deletion & Safety" \
  "🤖" "🤖"

# 6. Communication: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Communication Style" \
  "$CLAUDE" "🛠️ Communication Style" \
  "$CURSOR_DIR/000-index.mdc" "🛠️ Communication Style" \
  "^---" "^$"

# 7. Planning: CLAUDE.md vs 200-planning.mdc
diff_sections \
  "Planning" \
  "$CLAUDE" "Agentic Implementation Plan" \
  "$CURSOR_DIR/200-planning.mdc" "Agentic Implementation Plan" \
  "^---" "ZZZZZ_SENTINEL_EOF" \
  "s/sub-agents/agents/g;s/agents\/cycles/agents/g"

# 8. React patterns: rules/react-patterns.md vs 103-react.mdc
diff_rule_file \
  "React Patterns" \
  "$RULES_DIR/react-patterns.md" \
  "$CURSOR_DIR/103-react.mdc"

# 9. API design: rules/api-design.md vs 400-api-design.mdc
diff_rule_file \
  "API Design" \
  "$RULES_DIR/api-design.md" \
  "$CURSOR_DIR/400-api-design.mdc"

# 10. Database: rules/database.md vs 401-database.mdc
diff_rule_file \
  "Database Patterns" \
  "$RULES_DIR/database.md" \
  "$CURSOR_DIR/401-database.mdc"

# Deployed vs repo comparison (--deployed flag)
if [ "$CHECK_DEPLOYED" -eq 1 ]; then
  echo ""
  echo "=== Checking repo vs deployed (~/.claude/) ==="
  echo ""

  for f in "$LIB_DIR/claude/CLAUDE.md" "$LIB_DIR/claude/settings.json"; do
    basename=$(basename "$f")
    deployed="$DEPLOYED_DIR/$basename"
    if [ -f "$deployed" ]; then
      if ! diff -q "$f" "$deployed" > /dev/null 2>&1; then
        echo "DRIFT in [deployed $basename]"
        echo "  Between: $f"
        echo "  And:     $deployed"
        diff --unified=2 "$f" "$deployed" | head -40
        echo ""
        DRIFT_FOUND=1
      fi
    else
      echo "MISSING: $deployed (not deployed yet)"
      DRIFT_FOUND=1
    fi
  done

  # Check agents
  for f in "$LIB_DIR/claude/agents/"*.md; do
    basename=$(basename "$f")
    deployed="$DEPLOYED_DIR/agents/$basename"
    if [ -f "$deployed" ]; then
      if ! diff -q "$f" "$deployed" > /dev/null 2>&1; then
        echo "DRIFT in [deployed agents/$basename]"
        diff --unified=2 "$f" "$deployed" | head -20
        echo ""
        DRIFT_FOUND=1
      fi
    fi
  done

  # Check rules
  for f in "$LIB_DIR/claude/rules/"*.md; do
    basename=$(basename "$f")
    deployed="$DEPLOYED_DIR/rules/$basename"
    if [ -f "$deployed" ]; then
      if ! diff -q "$f" "$deployed" > /dev/null 2>&1; then
        echo "DRIFT in [deployed rules/$basename]"
        diff --unified=2 "$f" "$deployed" | head -20
        echo ""
        DRIFT_FOUND=1
      fi
    fi
  done
fi

if [ "$DRIFT_FOUND" -eq 0 ]; then
  echo "All sections in sync."
else
  echo "=== Drift detected. Update the source files to restore parity. ==="
  exit 1
fi
