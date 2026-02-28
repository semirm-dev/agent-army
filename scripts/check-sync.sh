#!/bin/bash
# Checks content parity between claude/rules/*.md and Cursor .mdc files,
# plus core CLAUDE.md sections vs 000-index.mdc and 200-planning.mdc.
#
# Usage: ./scripts/check-sync.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

CLAUDE="$LIB_DIR/claude/CLAUDE.md"
RULES_DIR="$LIB_DIR/claude/rules"
CURSOR_DIR="$LIB_DIR/cursor"

DRIFT_FOUND=0
TMPDIR_SYNC=$(mktemp -d)
trap 'rm -rf "$TMPDIR_SYNC"' EXIT
DEPLOYED_DIR="$HOME/.claude"
CHECK_DEPLOYED=0

# Portable modification time (works on both BSD/macOS and GNU/Linux)
portable_mtime() {
  local file="$1"
  if [[ "$OSTYPE" == darwin* ]]; then
    stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$file" 2>/dev/null
  else
    date -r "$file" "+%Y-%m-%d %H:%M" 2>/dev/null || stat -c "%y" "$file" 2>/dev/null | cut -d: -f1-2
  fi
}

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
  awk -v begin="$start_pattern" -v finish="$end_pattern" \
    '$0 ~ begin{found=1; next} $0 ~ finish{found=0} found' "$file" \
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
  tmp_a=$(mktemp "$TMPDIR_SYNC/tmp.XXXXXX")
  tmp_b=$(mktemp "$TMPDIR_SYNC/tmp.XXXXXX")

  if [ -n "$normalize" ]; then
    extract_section "$file_a" "$section_a" "$end_a" | { grep -v '^[[:space:]]*$' || true; } | sed "$normalize" > "$tmp_a"
    extract_section "$file_b" "$section_b" "$end_b" | { grep -v '^[[:space:]]*$' || true; } | sed "$normalize" > "$tmp_b"
  else
    extract_section "$file_a" "$section_a" "$end_a" | { grep -v '^[[:space:]]*$' || true; } > "$tmp_a"
    extract_section "$file_b" "$section_b" "$end_b" | { grep -v '^[[:space:]]*$' || true; } > "$tmp_b"
  fi

  if ! diff -q "$tmp_a" "$tmp_b" > /dev/null 2>&1; then
    echo "DRIFT in [$label]"
    echo "  Between: $file_a"
    echo "  And:     $file_b"
    # Show file timestamps for sync direction hints
    if [ -f "$file_a" ] && [ -f "$file_b" ]; then
      local mod_a mod_b
      mod_a=$(portable_mtime "$file_a")
      mod_b=$(portable_mtime "$file_b")
      echo "  $file_a last modified: $mod_a"
      echo "  $file_b last modified: $mod_b"
    fi
    diff --unified=2 "$tmp_a" "$tmp_b" | head -40
    echo ""
    DRIFT_FOUND=1
  fi
}

# Diff entire pattern files (after stripping comments and headings)
diff_rule_file() {
  local label="$1"
  local rule_file="$2"
  local cursor_file="$3"
  local normalize="${4:-}"

  local tmp_a tmp_b
  tmp_a=$(mktemp "$TMPDIR_SYNC/tmp.XXXXXX")
  tmp_b=$(mktemp "$TMPDIR_SYNC/tmp.XXXXXX")

  # Strip sync comments, front matter, and normalize headings
  if [ -n "$normalize" ]; then
    grep -v '^<!-- Sync:' "$rule_file" | { grep -v '^[[:space:]]*$' || true; } | sed 's/^#\{1,6\} //' | sed "$normalize" > "$tmp_a"
    sed -n '/^---$/,/^---$/!p' "$cursor_file" | grep -v '^<!-- Sync:' | { grep -v '^[[:space:]]*$' || true; } | sed 's/^#\{1,6\} //' | sed "$normalize" > "$tmp_b"
  else
    grep -v '^<!-- Sync:' "$rule_file" | { grep -v '^[[:space:]]*$' || true; } | sed 's/^#\{1,6\} //' > "$tmp_a"
    sed -n '/^---$/,/^---$/!p' "$cursor_file" | grep -v '^<!-- Sync:' | { grep -v '^[[:space:]]*$' || true; } | sed 's/^#\{1,6\} //' > "$tmp_b"
  fi

  if ! diff -q "$tmp_a" "$tmp_b" > /dev/null 2>&1; then
    echo "DRIFT in [$label]"
    echo "  Between: $rule_file"
    echo "  And:     $cursor_file"
    # Show file timestamps for sync direction hints
    local mod_a mod_b
    mod_a=$(portable_mtime "$rule_file")
    mod_b=$(portable_mtime "$cursor_file")
    echo "  $rule_file last modified: $mod_a"
    echo "  $cursor_file last modified: $mod_b"
    echo "  Hint: update the older file to match the newer one"
    diff --unified=2 "$tmp_a" "$tmp_b" | head -40
    echo ""
    DRIFT_FOUND=1
  fi
}

echo "=== Checking sync between rules and Cursor .mdc files ==="
echo ""

# Rule ↔ cursor file pairs (driven by config.json)
while read -r pair; do
  label=$(echo "$pair" | jq -r '.label')
  rule=$(echo "$pair" | jq -r '.rule')
  cursor=$(echo "$pair" | jq -r '.cursor')
  diff_rule_file "$label" "$RULES_DIR/$rule" "$CURSOR_DIR/$cursor"
done < <(cfg_raw '.sync_pairs[]')

# Special section diffs (unique extraction patterns, not config-driven)

# Safety section: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Deletion & Safety" \
  "$CLAUDE" "🛡️ Deletion & Safety" \
  "$CURSOR_DIR/000-index.mdc" "🛡️ Deletion & Safety" \
  "🤖" "🤖"

# Communication + Conflict Resolution: CLAUDE.md vs 000-index.mdc
diff_sections \
  "Communication & Conflict Resolution" \
  "$CLAUDE" "🛠️ Communication Style" \
  "$CURSOR_DIR/000-index.mdc" "🛠️ Communication Style" \
  "^---" "ZZZZZ_SENTINEL_EOF"

# Planning: CLAUDE.md vs 200-planning.mdc
diff_sections \
  "Planning" \
  "$CLAUDE" "Agentic Implementation Plan" \
  "$CURSOR_DIR/200-planning.mdc" "Agentic Implementation Plan" \
  "^---" "ZZZZZ_SENTINEL_EOF" \
  "s/sub-agents/agents/g;s/agents\/cycles/agents/g"

# Multi-Agent Management: only shared bullets are compared (the sections
# are intentionally different — CLAUDE.md has full agent/plugin lists,
# cursor has a simplified version).
diff_shared_bullets() {
  local label="$1" file_a="$2" file_b="$3"
  shift 3
  for bullet in "$@"; do
    local line_a line_b
    line_a=$(grep "^- \*\*${bullet}:\*\*" "$file_a" | head -1 || true)
    line_b=$(grep "^- \*\*${bullet}:\*\*" "$file_b" | head -1 || true)
    if [ -z "$line_a" ] && [ -z "$line_b" ]; then
      continue
    fi
    if [ "$line_a" != "$line_b" ]; then
      echo "DRIFT in [$label — $bullet]"
      echo "  $file_a: ${line_a:-(missing)}"
      echo "  $file_b: ${line_b:-(missing)}"
      DRIFT_FOUND=1
    fi
  done
}

diff_shared_bullets "Multi-Agent (shared bullets)" \
  "$CLAUDE" "$CURSOR_DIR/000-index.mdc" \
  "Role" "Parallelism" "Verification"

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
    else
      echo "MISSING: $deployed (not deployed yet)"
      DRIFT_FOUND=1
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
    else
      echo "MISSING: $deployed (not deployed yet)"
      DRIFT_FOUND=1
    fi
  done
fi

# Agent frontmatter parity: name + description must match between platforms
echo ""
echo "=== Checking agent frontmatter parity (claude ↔ cursor) ==="
echo ""

CLAUDE_AGENTS_DIR="$LIB_DIR/claude/agents"
CURSOR_AGENTS_DIR="$LIB_DIR/cursor/agents"

extract_frontmatter_field() {
  local file="$1" field="$2"
  sed -n '/^---$/,/^---$/p' "$file" | grep "^${field}:" | sed "s/^${field}:[[:space:]]*//" | sed 's/^"//;s/"$//'
}

if [ -d "$CURSOR_AGENTS_DIR" ]; then
  for claude_agent in "$CLAUDE_AGENTS_DIR"/*.md; do
    agent_name=$(basename "$claude_agent")
    cursor_agent="$CURSOR_AGENTS_DIR/$agent_name"

    if [ ! -f "$cursor_agent" ]; then
      continue
    fi

    claude_name=$(extract_frontmatter_field "$claude_agent" "name")
    cursor_name=$(extract_frontmatter_field "$cursor_agent" "name")
    claude_desc=$(extract_frontmatter_field "$claude_agent" "description")
    cursor_desc=$(extract_frontmatter_field "$cursor_agent" "description")

    if [ "$claude_name" != "$cursor_name" ]; then
      echo "DRIFT in [agent $agent_name — name]"
      echo "  claude: $claude_name"
      echo "  cursor: $cursor_name"
      DRIFT_FOUND=1
    fi

    if [ "$claude_desc" != "$cursor_desc" ]; then
      echo "DRIFT in [agent $agent_name — description]"
      echo "  claude: $claude_desc"
      echo "  cursor: $cursor_desc"
      DRIFT_FOUND=1
    fi
  done
fi

# AGENT-GUIDE.md skills matrix vs agent frontmatter (warning-level, non-blocking)
echo ""
echo "=== Checking AGENT-GUIDE.md skills matrix vs agent frontmatter (warnings only) ==="
echo ""

AGENT_GUIDE="$LIB_DIR/docs/AGENT-GUIDE.md"
WARN_COUNT=0

if [ -f "$AGENT_GUIDE" ]; then
  # Build list of known custom skill names from config.json
  CUSTOM_SKILLS_LIST=$(mktemp "$TMPDIR_SYNC/custom_skills.XXXXXX")
  cfg '.custom_skills[].name' > "$CUSTOM_SKILLS_LIST"

  # Cursor built-in subagent types (no .md files, skip in agent file checks)
  CURSOR_BUILTINS="explore generalPurpose shell browser-use code-reviewer code-simplifier docs-researcher"

  # Parse the skills matrix table from AGENT-GUIDE.md (not the built-in agents table)
  # Format: | agent-name | custom-skills | external-skills | subagents |
  GUIDE_MATRIX=$(mktemp "$TMPDIR_SYNC/guide_matrix.XXXXXX")
  awk '/^\| Agent \| Custom Skills/,/^$/' "$AGENT_GUIDE" \
    | grep -v '^| Agent ' \
    | grep -v '^|---' \
    | grep '|' \
    > "$GUIDE_MATRIX" || true

  while IFS='|' read -r _ agent_col skills_col plugins_col _; do
    # Trim whitespace
    guide_agent=$(echo "$agent_col" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    guide_skills_raw=$(echo "$skills_col" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    # Skip empty lines
    [ -z "$guide_agent" ] && continue

    # Skip Cursor built-in types
    if echo "$CURSOR_BUILTINS" | grep -qw "$guide_agent"; then
      continue
    fi

    # Resolve agent file (prefer cursor/ for skill path refs, fall back to claude/)
    cursor_agent_file="$CURSOR_AGENTS_DIR/${guide_agent}.md"
    claude_agent_file="$CLAUDE_AGENTS_DIR/${guide_agent}.md"
    if [ -f "$cursor_agent_file" ]; then
      agent_file="$cursor_agent_file"
    elif [ -f "$claude_agent_file" ]; then
      agent_file="$claude_agent_file"
    else
      echo "WARNING [AGENT-GUIDE skills] Agent '$guide_agent' listed in guide but no file found"
      WARN_COUNT=$((WARN_COUNT + 1))
      continue
    fi

    # Extract skills referenced in agent body (after YAML frontmatter)
    BODY_SKILLS=$(mktemp "$TMPDIR_SYNC/body_skills.XXXXXX")
    awk 'BEGIN{n=0} /^---$/{n++; next} n>=2' "$agent_file" \
      | grep -oE '~/\.(cursor|claude)/skills/[a-z-]+/SKILL\.md' \
      | sed 's|.*/skills/||;s|/SKILL\.md||' \
      | sort -u > "$BODY_SKILLS" || true

    # Filter body skills to only custom skills (exclude npm skills like golang-pro)
    BODY_CUSTOM=$(mktemp "$TMPDIR_SYNC/body_custom.XXXXXX")
    while read -r skill; do
      if grep -qx "$skill" "$CUSTOM_SKILLS_LIST"; then
        echo "$skill"
      fi
    done < "$BODY_SKILLS" | sort > "$BODY_CUSTOM"

    # Parse guide custom skills column (comma-separated, or em-dash/dash for none)
    # The guide uses Unicode em dash (U+2014) to mean "no skills"
    GUIDE_CUSTOM=$(mktemp "$TMPDIR_SYNC/guide_custom.XXXXXX")
    guide_skills_normalized=$(echo "$guide_skills_raw" | sed $'s/\xe2\x80\x94//g' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    if [ -n "$guide_skills_normalized" ]; then
      echo "$guide_skills_normalized" | tr ',' '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | sort > "$GUIDE_CUSTOM"
    else
      : > "$GUIDE_CUSTOM"
    fi

    # Compare
    if ! diff -q "$GUIDE_CUSTOM" "$BODY_CUSTOM" > /dev/null 2>&1; then
      guide_list=$(tr '\n' ',' < "$GUIDE_CUSTOM" | sed 's/,$//')
      body_list=$(tr '\n' ',' < "$BODY_CUSTOM" | sed 's/,$//')
      [ -z "$guide_list" ] && guide_list="none"
      [ -z "$body_list" ] && body_list="none"
      echo "WARNING [AGENT-GUIDE skills — $guide_agent]"
      echo "  guide says:      $guide_list"
      echo "  agent body says: $body_list"
      WARN_COUNT=$((WARN_COUNT + 1))
    fi
  done < "$GUIDE_MATRIX"

  # Also check for agents that exist as files but are missing from the guide
  for agent_file in "$CLAUDE_AGENTS_DIR"/*.md; do
    agent_basename=$(basename "$agent_file" .md)
    # Skip Cursor built-in types (they don't have .md files and aren't in the matrix)
    if echo "$CURSOR_BUILTINS" | grep -qw "$agent_basename"; then
      continue
    fi
    if ! grep -q "| ${agent_basename} " "$GUIDE_MATRIX" 2>/dev/null; then
      echo "WARNING [AGENT-GUIDE coverage] Agent '$agent_basename' exists in claude/agents/ but is not listed in AGENT-GUIDE.md matrix"
      WARN_COUNT=$((WARN_COUNT + 1))
    fi
  done

  if [ "$WARN_COUNT" -eq 0 ]; then
    echo "AGENT-GUIDE.md skills matrix matches agent frontmatter."
  else
    echo ""
    echo "Found $WARN_COUNT warning(s) in AGENT-GUIDE.md skills matrix check."
    echo "These are non-blocking warnings. Update docs/AGENT-GUIDE.md or agent frontmatter to resolve."
  fi
else
  echo "SKIP: docs/AGENT-GUIDE.md not found."
fi

if [ "$DRIFT_FOUND" -eq 0 ]; then
  echo ""
  echo "All sections in sync."
else
  echo ""
  echo "=== Drift detected. Update the source files to restore parity. ==="
  exit 1
fi
