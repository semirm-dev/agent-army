#!/bin/bash
# Tests for check-sync.sh — verifies drift detection works correctly.
#
# Creates temp files with known content, runs check-sync logic, asserts results.
#
# Tests:
#   1. Identical sections         → no drift (exit 0)
#   2. Extra content in cursor    → drift detected (exit 1)
#   3. Heading level differences  → ignored, no drift (exit 0)
#   4. Rule vs cursor file drift  → drift detected (exit 1)
#   5. Real check-sync on repo    → passes (exit 0)
#   6. Shared bullet drift        → drift detected (exit 1)
#   7. Empty extraction resilience → no crash (exit 0)
#
# Usage: ./scripts/test-check-sync.sh
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

TMPDIR_TEST=$(mktemp -d)
trap 'rm -rf "$TMPDIR_TEST"' EXIT

# ── Test 1: Identical sections → no drift ──────────────────────────

echo "── Test 1: No drift (identical sections) ──"

cat > "$TMPDIR_TEST/claude.md" <<'EOF'
# Header

## 🛡️ Deletion & Safety (Hard Constraints)
- Rule A
- Rule B

## 🤖 Next Section
EOF

cat > "$TMPDIR_TEST/cursor.mdc" <<'EOF'
---
globs: "**/*"
---

### 🛡️ Deletion & Safety (Hard Constraints)
- Rule A
- Rule B

### 🤖 Next Section
EOF

# Build a minimal check-sync that tests these two files
cat > "$TMPDIR_TEST/check.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

extract_section() {
  local file="\$1" start_pattern="\$2" end_pattern="\$3"
  awk "/\$start_pattern/{found=1; next} /\$end_pattern/{found=0} found" "\$file" \\
    | sed 's/^#\\{1,6\\} //'
}

diff_sections() {
  local label="\$1" file_a="\$2" section_a="\$3" file_b="\$4" section_b="\$5"
  local end_a="\${6:-^\$}" end_b="\${7:-^\$}"
  local tmp_a=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX") tmp_b=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX")
  extract_section "\$file_a" "\$section_a" "\$end_a" | grep -v '^[[:space:]]*\$' > "\$tmp_a"
  extract_section "\$file_b" "\$section_b" "\$end_b" | grep -v '^[[:space:]]*\$' > "\$tmp_b"
  if ! diff -q "\$tmp_a" "\$tmp_b" > /dev/null 2>&1; then
    DRIFT_FOUND=1
  fi
  rm -f "\$tmp_a" "\$tmp_b"
}

diff_sections "Safety" \\
  "$TMPDIR_TEST/claude.md" "🛡️ Deletion & Safety" \\
  "$TMPDIR_TEST/cursor.mdc" "🛡️ Deletion & Safety" \\
  "🤖" "🤖"

exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check.sh"

bash "$TMPDIR_TEST/check.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Identical sections → exit 0 (no drift)" 0 "$exit_code"

# ── Test 2: Drifted sections → drift detected ─────────────────────

echo "── Test 2: Drift detected (different content) ──"

cat > "$TMPDIR_TEST/cursor_drifted.mdc" <<'EOF'
---
globs: "**/*"
---

### 🛡️ Deletion & Safety (Hard Constraints)
- Rule A
- Rule B
- Rule C EXTRA

### 🤖 Next Section
EOF

cat > "$TMPDIR_TEST/check_drift.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

extract_section() {
  local file="\$1" start_pattern="\$2" end_pattern="\$3"
  awk "/\$start_pattern/{found=1; next} /\$end_pattern/{found=0} found" "\$file" \\
    | sed 's/^#\\{1,6\\} //'
}

diff_sections() {
  local label="\$1" file_a="\$2" section_a="\$3" file_b="\$4" section_b="\$5"
  local end_a="\${6:-^\$}" end_b="\${7:-^\$}"
  local tmp_a=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX") tmp_b=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX")
  extract_section "\$file_a" "\$section_a" "\$end_a" | grep -v '^[[:space:]]*\$' > "\$tmp_a"
  extract_section "\$file_b" "\$section_b" "\$end_b" | grep -v '^[[:space:]]*\$' > "\$tmp_b"
  if ! diff -q "\$tmp_a" "\$tmp_b" > /dev/null 2>&1; then
    DRIFT_FOUND=1
  fi
  rm -f "\$tmp_a" "\$tmp_b"
}

diff_sections "Safety" \\
  "$TMPDIR_TEST/claude.md" "🛡️ Deletion & Safety" \\
  "$TMPDIR_TEST/cursor_drifted.mdc" "🛡️ Deletion & Safety" \\
  "🤖" "🤖"

exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check_drift.sh"

bash "$TMPDIR_TEST/check_drift.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Drifted sections → exit 1 (drift found)" 1 "$exit_code"

# ── Test 3: Heading level differences are ignored ──────────────────

echo "── Test 3: Different heading levels → no drift ──"

cat > "$TMPDIR_TEST/claude_h2.md" <<'EOF'
## 🛡️ Safety
- Rule A

## End
EOF

cat > "$TMPDIR_TEST/cursor_h3.mdc" <<'EOF'
### 🛡️ Safety
- Rule A

### End
EOF

cat > "$TMPDIR_TEST/check_headings.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

extract_section() {
  local file="\$1" start_pattern="\$2" end_pattern="\$3"
  awk "/\$start_pattern/{found=1; next} /\$end_pattern/{found=0} found" "\$file" \\
    | sed 's/^#\\{1,6\\} //'
}

diff_sections() {
  local label="\$1" file_a="\$2" section_a="\$3" file_b="\$4" section_b="\$5"
  local end_a="\${6:-^\$}" end_b="\${7:-^\$}"
  local tmp_a=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX") tmp_b=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX")
  extract_section "\$file_a" "\$section_a" "\$end_a" | grep -v '^[[:space:]]*\$' > "\$tmp_a"
  extract_section "\$file_b" "\$section_b" "\$end_b" | grep -v '^[[:space:]]*\$' > "\$tmp_b"
  if ! diff -q "\$tmp_a" "\$tmp_b" > /dev/null 2>&1; then
    DRIFT_FOUND=1
  fi
  rm -f "\$tmp_a" "\$tmp_b"
}

diff_sections "Safety" \\
  "$TMPDIR_TEST/claude_h2.md" "🛡️ Safety" \\
  "$TMPDIR_TEST/cursor_h3.mdc" "🛡️ Safety" \\
  "End" "End"

exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check_headings.sh"

bash "$TMPDIR_TEST/check_headings.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Different heading levels → exit 0 (no drift)" 0 "$exit_code"

# ── Test 4: Rule file vs cursor file parity ────────────────────────

echo "── Test 4: Rule file drift detection ──"

cat > "$TMPDIR_TEST/rule.md" <<'EOF'
<!-- Sync: Must stay in sync with cursor/test.mdc -->

# 🐍 Python Patterns
- Rule A
- Rule B
EOF

cat > "$TMPDIR_TEST/cursor_rule.mdc" <<'EOF'
---
globs: "**/*.py"
---

<!-- Sync: Must stay in sync with claude/rules/py-patterns.md -->

## 🐍 Python Patterns
- Rule A
- Rule B
- Rule C EXTRA
EOF

cat > "$TMPDIR_TEST/check_rule.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

diff_rule_file() {
  local label="\$1" rule_file="\$2" cursor_file="\$3"
  local tmp_a=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX") tmp_b=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX")
  grep -v '^<!-- Sync:' "\$rule_file" | grep -v '^[[:space:]]*\$' | sed 's/^#\\{1,6\\} //' > "\$tmp_a"
  sed -n '/^---\$/,/^---\$/!p' "\$cursor_file" | grep -v '^<!-- Sync:' | grep -v '^[[:space:]]*\$' | sed 's/^#\\{1,6\\} //' > "\$tmp_b"
  if ! diff -q "\$tmp_a" "\$tmp_b" > /dev/null 2>&1; then
    DRIFT_FOUND=1
  fi
  rm -f "\$tmp_a" "\$tmp_b"
}

diff_rule_file "Python" \\
  "$TMPDIR_TEST/rule.md" \\
  "$TMPDIR_TEST/cursor_rule.mdc"

exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check_rule.sh"

bash "$TMPDIR_TEST/check_rule.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Rule file drift detected → exit 1" 1 "$exit_code"

# ── Test 5: Real check-sync.sh against repo ───────────────────────

echo "── Test 5: Real check-sync.sh on repo (should pass) ──"

bash "$SCRIPT_DIR/check-sync.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Real check-sync on repo → exit 0" 0 "$exit_code"

# ── Test 6: Shared bullet drift detection ─────────────────────────

echo "── Test 6: Shared bullet drift detection ──"

cat > "$TMPDIR_TEST/claude_bullets.md" <<'EOF'
## 🤖 Multi-Agent Management (The Manager Workflow)
- **Role:** You act as a **Lead Product Architect**.
- **Parallelism:** Split work into parallel subagents.
- **Agent Definitions:** (claude-only content here)
- **Verification:** Run the build command.
EOF

cat > "$TMPDIR_TEST/cursor_bullets.mdc" <<'EOF'
## 🤖 Multi-Agent Management (The Manager Workflow)
- **Role:** You act as a **Lead Technical Architect**.
- **Parallelism:** Split work into parallel subagents.
- **Verification:** Run the build command.
EOF

cat > "$TMPDIR_TEST/check_bullets.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

diff_shared_bullets() {
  local label="\$1" file_a="\$2" file_b="\$3"
  shift 3
  for bullet in "\$@"; do
    local line_a line_b
    line_a=\$(grep "^- \*\*\${bullet}:\*\*" "\$file_a" | head -1 || true)
    line_b=\$(grep "^- \*\*\${bullet}:\*\*" "\$file_b" | head -1 || true)
    if [ -z "\$line_a" ] && [ -z "\$line_b" ]; then
      continue
    fi
    if [ "\$line_a" != "\$line_b" ]; then
      DRIFT_FOUND=1
    fi
  done
}

diff_shared_bullets "Multi-Agent" \\
  "$TMPDIR_TEST/claude_bullets.md" "$TMPDIR_TEST/cursor_bullets.mdc" \\
  "Role" "Parallelism" "Verification"

exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check_bullets.sh"

bash "$TMPDIR_TEST/check_bullets.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Shared bullet drift detected → exit 1" 1 "$exit_code"

# ── Test 7: Empty extraction resilience ───────────────────────────

echo "── Test 7: Empty extraction → no crash ──"

cat > "$TMPDIR_TEST/check_empty.sh" <<SCRIPT
#!/bin/bash
set -euo pipefail
DRIFT_FOUND=0

extract_section() {
  local file="\$1" start_pattern="\$2" end_pattern="\$3"
  awk "/\$start_pattern/{found=1; next} /\$end_pattern/{found=0} found" "\$file" \\
    | sed 's/^#\\{1,6\\} //'
}

tmp_a=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX") tmp_b=\$(mktemp "$TMPDIR_TEST/inner.XXXXXX")

extract_section "$TMPDIR_TEST/claude.md" "NONEXISTENT_PATTERN" "ALSO_NONEXISTENT" \
  | { grep -v '^[[:space:]]*\$' || true; } > "\$tmp_a"
extract_section "$TMPDIR_TEST/cursor.mdc" "NONEXISTENT_PATTERN" "ALSO_NONEXISTENT" \
  | { grep -v '^[[:space:]]*\$' || true; } > "\$tmp_b"

if ! diff -q "\$tmp_a" "\$tmp_b" > /dev/null 2>&1; then
  DRIFT_FOUND=1
fi

rm -f "\$tmp_a" "\$tmp_b"
exit \$DRIFT_FOUND
SCRIPT
chmod +x "$TMPDIR_TEST/check_empty.sh"

bash "$TMPDIR_TEST/check_empty.sh" > /dev/null 2>&1 && exit_code=0 || exit_code=$?
assert_exit "Empty extraction → exit 0 (no crash)" 0 "$exit_code"

# ── Summary ────────────────────────────────────────────────────────

echo ""
echo "================================="
echo "  Results: $PASS passed, $FAIL failed"
echo "================================="

[ "$FAIL" -eq 0 ] || exit 1
