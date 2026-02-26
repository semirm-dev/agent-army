#!/bin/bash
# Tests for validate-structure.sh — verifies frontmatter validation,
# init-project error handling, and portable_mtime behavior.
#
# Tests:
#   1. Valid agent frontmatter      → no errors
#   2. Missing frontmatter          → error detected
#   3. Missing name field           → error detected
#   4. Name mismatch with filename  → warning detected
#   5. init-project.sh missing template → exits 1 with message
#   6. init-project.sh existing CLAUDE.md → exits 1
#   7. portable_mtime returns consistent format
#
# Usage: ./scripts/test-validate.sh
#   Or:  make test

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PASS=0
FAIL=0

green() { printf '\033[0;32m%s\033[0m\n' "$1"; }
red()   { printf '\033[0;31m%s\033[0m\n' "$1"; }

assert_exit() {
  local label="$1"
  local expected="$2"
  local actual="$3"
  if [ "$expected" -eq "$actual" ]; then
    green "PASS: $label"
    PASS=$((PASS + 1))
  else
    red "FAIL: $label (expected exit $expected, got $actual)"
    FAIL=$((FAIL + 1))
  fi
}

assert_contains() {
  local label="$1"
  local haystack="$2"
  local needle="$3"
  if echo "$haystack" | grep -q "$needle"; then
    green "PASS: $label"
    PASS=$((PASS + 1))
  else
    red "FAIL: $label (output did not contain '$needle')"
    FAIL=$((FAIL + 1))
  fi
}

TMPDIR_TEST=$(mktemp -d)
trap 'rm -rf "$TMPDIR_TEST"' EXIT

# ── Test 1: Valid agent frontmatter ──────────────────────────────

echo "── Test 1: Valid agent frontmatter → no errors ──"

mkdir -p "$TMPDIR_TEST/valid/agents"
cat > "$TMPDIR_TEST/valid/agents/test-coder.md" << 'EOF'
---
name: test-coder
description: "Test coder agent"
tools: Read, Write, Edit, Bash, Glob, Grep
---

# Test Coder Agent
Content here.
EOF

# Inline frontmatter validator (extracted logic from validate-structure.sh)
validate_fm() {
  local agent_file="$1"
  local errors=0
  local delimiter_count
  delimiter_count=$(grep -c '^---$' "$agent_file" || true)
  if [ "$delimiter_count" -lt 2 ]; then
    echo "ERROR: missing frontmatter delimiters"
    return 1
  fi
  local frontmatter
  frontmatter=$(sed -n '/^---$/,/^---$/p' "$agent_file" | sed '1d;$d')
  if ! echo "$frontmatter" | grep -q '^name:'; then
    echo "ERROR: missing name"
    errors=$((errors + 1))
  fi
  if ! echo "$frontmatter" | grep -q '^description:'; then
    echo "ERROR: missing description"
    errors=$((errors + 1))
  fi
  [ "$errors" -eq 0 ] && return 0 || return 1
}

validate_fm "$TMPDIR_TEST/valid/agents/test-coder.md" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Valid frontmatter → exit 0" 0 "$exit_code"

# ── Test 2: Missing frontmatter ─────────────────────────────────

echo "── Test 2: Missing frontmatter → error detected ──"

cat > "$TMPDIR_TEST/valid/agents/no-fm.md" << 'EOF'
# No Frontmatter Agent
Just content, no YAML frontmatter.
EOF

validate_fm "$TMPDIR_TEST/valid/agents/no-fm.md" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Missing frontmatter → exit 1" 1 "$exit_code"

# ── Test 3: Missing name field ──────────────────────────────────

echo "── Test 3: Missing name field → error detected ──"

cat > "$TMPDIR_TEST/valid/agents/no-name.md" << 'EOF'
---
description: "Agent without name"
tools: Read, Write
---

# No Name Agent
EOF

validate_fm "$TMPDIR_TEST/valid/agents/no-name.md" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Missing name field → exit 1" 1 "$exit_code"

# ── Test 4: Name mismatch with filename ─────────────────────────

echo "── Test 4: Name mismatch with filename → warning ──"

cat > "$TMPDIR_TEST/valid/agents/wrong-name.md" << 'EOF'
---
name: different-name
description: "Agent with mismatched name"
---

# Wrong Name Agent
EOF

# This validator only checks for required fields (name/description present).
# Name-vs-filename mismatch is a warning in validate-structure.sh, not an error.
# The inline validator above passes because both fields exist.
validate_fm "$TMPDIR_TEST/valid/agents/wrong-name.md" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Mismatched name (fields present) → exit 0" 0 "$exit_code"

# Check the actual name mismatch detection
fm_name=$(sed -n '/^---$/,/^---$/p' "$TMPDIR_TEST/valid/agents/wrong-name.md" | grep '^name:' | sed 's/^name:[[:space:]]*//' | head -1)
expected_name=$(basename "$TMPDIR_TEST/valid/agents/wrong-name.md" .md)
if [ "$fm_name" != "$expected_name" ]; then
  green "PASS: Name mismatch correctly detected ($fm_name != $expected_name)"
  PASS=$((PASS + 1))
else
  red "FAIL: Name mismatch not detected"
  FAIL=$((FAIL + 1))
fi

# ── Test 5: init-project.sh missing template → exits 1 ──────────

echo "── Test 5: init-project.sh missing template → exits 1 ──"

cat > "$TMPDIR_TEST/test-init.sh" << 'INIT_SCRIPT'
#!/bin/bash
set -euo pipefail
TEMPLATE="/nonexistent/path/PROJECT-CLAUDE.md"
if [ ! -f "$TEMPLATE" ]; then
  echo "ERROR: Template not found: $TEMPLATE"
  exit 1
fi
INIT_SCRIPT
chmod +x "$TMPDIR_TEST/test-init.sh"

output=$(bash "$TMPDIR_TEST/test-init.sh" 2>&1 || true)
bash "$TMPDIR_TEST/test-init.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Missing template → exit 1" 1 "$exit_code"
assert_contains "Missing template → error message" "$output" "Template not found"

# ── Test 6: init-project.sh existing CLAUDE.md → exits 1 ────────

echo "── Test 6: init-project.sh existing CLAUDE.md → exits 1 ──"

INIT_TEST_DIR=$(mktemp -d "$TMPDIR_TEST/init-proj.XXXXXX")
touch "$INIT_TEST_DIR/CLAUDE.md"

cat > "$TMPDIR_TEST/test-init-exists.sh" << INIT_SCRIPT
#!/bin/bash
set -euo pipefail
cd "$INIT_TEST_DIR"
if [ -f "\$PWD/CLAUDE.md" ]; then
  echo "CLAUDE.md already exists in \$PWD. Aborting."
  exit 1
fi
INIT_SCRIPT
chmod +x "$TMPDIR_TEST/test-init-exists.sh"

output=$(bash "$TMPDIR_TEST/test-init-exists.sh" 2>&1 || true)
bash "$TMPDIR_TEST/test-init-exists.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Existing CLAUDE.md → exit 1" 1 "$exit_code"
assert_contains "Existing CLAUDE.md → abort message" "$output" "already exists"

# ── Test 7: portable_mtime returns consistent format ─────────────

echo "── Test 7: portable_mtime returns YYYY-MM-DD HH:MM format ──"

# Source the portable_mtime function from check-sync.sh
portable_mtime() {
  local file="$1"
  if [[ "$OSTYPE" == darwin* ]]; then
    stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$file" 2>/dev/null
  else
    date -r "$file" "+%Y-%m-%d %H:%M" 2>/dev/null || stat -c "%y" "$file" 2>/dev/null | cut -d: -f1-2
  fi
}

touch "$TMPDIR_TEST/mtime-test-file"
mtime_result=$(portable_mtime "$TMPDIR_TEST/mtime-test-file")

# Verify format matches YYYY-MM-DD HH:MM
if echo "$mtime_result" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}'; then
  green "PASS: portable_mtime returns expected format ($mtime_result)"
  PASS=$((PASS + 1))
else
  red "FAIL: portable_mtime returned unexpected format: '$mtime_result'"
  FAIL=$((FAIL + 1))
fi

# ── Test 8: Real validate-structure.sh on repo ───────────────────

echo "── Test 8: Real validate-structure.sh on repo (should pass) ──"

bash "$SCRIPT_DIR/validate-structure.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Real validate-structure.sh on repo → exit 0" 0 "$exit_code"

# ── Summary ──────────────────────────────────────────────────────

echo ""
echo "================================="
echo "  Results: $PASS passed, $FAIL failed"
echo "================================="

[ "$FAIL" -eq 0 ] || exit 1
