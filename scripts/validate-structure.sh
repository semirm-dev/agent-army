#!/bin/bash
# Validates the structural integrity of the agent-rules repository.
# Checks: agent files exist, rule files exist, agent triads complete,
# synced pairs registered, CLAUDE.md references valid.

set -euo pipefail

LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLAUDE_MD="$LIB_DIR/claude/CLAUDE.md"
AGENTS_DIR="$LIB_DIR/claude/agents"
RULES_DIR="$LIB_DIR/claude/rules"
CURSOR_DIR="$LIB_DIR/cursor"

ERRORS=0
WARNINGS=0

error() { echo "  ✗ ERROR: $1"; ERRORS=$((ERRORS + 1)); }
warn()  { echo "  ⚠ WARNING: $1"; WARNINGS=$((WARNINGS + 1)); }
ok()    { echo "  ✓ $1"; }

echo "=== Structural Validation ==="
echo ""

# 1. Check all agent files referenced in CLAUDE.md exist
echo "--- Agent files referenced in CLAUDE.md ---"
# Extract agent filenames from backtick-quoted references like `go-coder.md`
REFERENCED_AGENTS=$(grep -oE '`[a-z]+-[a-z]+\.md`' "$CLAUDE_MD" | tr -d '`' | sort -u)
for agent in $REFERENCED_AGENTS; do
  if [ -f "$AGENTS_DIR/$agent" ]; then
    ok "$agent exists"
  else
    error "$agent referenced in CLAUDE.md but not found in claude/agents/"
  fi
done
echo ""

# 2. Check all rule files referenced in CLAUDE.md exist
echo "--- Rule files referenced in CLAUDE.md ---"
REFERENCED_RULES=$(grep -oE 'rules/[a-z-]+\.md' "$CLAUDE_MD" | sort -u)
for rule in $REFERENCED_RULES; do
  if [ -f "$LIB_DIR/claude/$rule" ]; then
    ok "$rule exists"
  else
    error "$rule referenced in CLAUDE.md but not found"
  fi
done
echo ""

# 3. Check agent triads (coder/reviewer/tester) are complete
echo "--- Agent triad completeness ---"
# Extract language prefixes from agent filenames
PREFIXES=$(ls "$AGENTS_DIR"/*.md 2>/dev/null | xargs -I{} basename {} .md | sed 's/-[a-z]*$//' | sort -u)
for prefix in $PREFIXES; do
  # Skip docker (has builder+reviewer, not coder/reviewer/tester)
  if [ "$prefix" = "docker" ]; then
    HAS_BUILDER=false
    HAS_REVIEWER=false
    [ -f "$AGENTS_DIR/docker-builder.md" ] && HAS_BUILDER=true
    [ -f "$AGENTS_DIR/docker-reviewer.md" ] && HAS_REVIEWER=true
    if $HAS_BUILDER && $HAS_REVIEWER; then
      ok "docker: builder + reviewer present"
    else
      warn "docker: missing builder or reviewer"
    fi
    continue
  fi

  HAS_CODER=false
  HAS_REVIEWER=false
  HAS_TESTER=false
  [ -f "$AGENTS_DIR/${prefix}-coder.md" ] && HAS_CODER=true
  [ -f "$AGENTS_DIR/${prefix}-reviewer.md" ] && HAS_REVIEWER=true
  [ -f "$AGENTS_DIR/${prefix}-tester.md" ] && HAS_TESTER=true

  if $HAS_CODER && $HAS_REVIEWER && $HAS_TESTER; then
    ok "$prefix: coder + reviewer + tester present"
  else
    MISSING=""
    $HAS_CODER || MISSING="$MISSING coder"
    $HAS_REVIEWER || MISSING="$MISSING reviewer"
    $HAS_TESTER || MISSING="$MISSING tester"
    warn "$prefix: missing$MISSING"
  fi
done
echo ""

# 4. Check synced rule/cursor pairs
echo "--- Synced rule ↔ cursor pairs ---"
# Parse the CLAUDE.md table for synced pairs (rules/X.md → cursor/Y.mdc)
grep -E 'rules/.*cursor/' "$CLAUDE_MD" | while IFS='|' read -r _ rule_col cursor_col _; do
  rule=$(echo "$rule_col" | grep -oE 'rules/[a-z-]+\.md' || true)
  cursor=$(echo "$cursor_col" | grep -oE 'cursor/[0-9a-z-]+\.mdc' || true)
  if [ -n "$rule" ] && [ -n "$cursor" ]; then
    if [ -f "$LIB_DIR/claude/$rule" ] && [ -f "$LIB_DIR/$cursor" ]; then
      ok "$rule ↔ $cursor both exist"
    else
      [ ! -f "$LIB_DIR/claude/$rule" ] && error "$rule missing"
      [ ! -f "$LIB_DIR/$cursor" ] && error "$cursor missing"
    fi
  fi
done
echo ""

# 5. Check every rule file referenced by agent prompts exists
echo "--- Rule files referenced by agents ---"
for agent_file in "$AGENTS_DIR"/*.md; do
  AGENT_NAME=$(basename "$agent_file")
  AGENT_RULES=$(grep -oE 'rules/[a-z-]+\.md' "$agent_file" 2>/dev/null | sort -u || true)
  for rule in $AGENT_RULES; do
    RULE_PATH="$LIB_DIR/claude/$rule"
    # Handle ~/.claude/rules/ references
    if echo "$rule" | grep -q '^\~'; then
      RULE_PATH="$HOME/.claude/$rule"
    fi
    # Normalize: agents reference ~/.claude/rules/X.md, check claude/rules/X.md in repo
    RULE_BASENAME=$(basename "$rule")
    if [ -f "$RULES_DIR/$RULE_BASENAME" ]; then
      ok "$AGENT_NAME → rules/$RULE_BASENAME"
    else
      error "$AGENT_NAME references rules/$RULE_BASENAME which doesn't exist"
    fi
  done
done
echo ""

# 6. Check skills directory
echo "--- Skills ---"
if [ -d "$LIB_DIR/skills" ]; then
  SKILL_COUNT=$(ls "$LIB_DIR/skills"/*.md 2>/dev/null | wc -l | tr -d ' ')
  ok "$SKILL_COUNT custom skill(s) found"
  for skill_file in "$LIB_DIR/skills"/*.md; do
    ok "  $(basename "$skill_file")"
  done
else
  warn "No skills/ directory found"
fi
echo ""

# Summary
echo "=== Summary ==="
echo "  Errors:   $ERRORS"
echo "  Warnings: $WARNINGS"

if [ "$ERRORS" -gt 0 ]; then
  echo ""
  echo "Validation FAILED with $ERRORS error(s)."
  exit 1
fi

if [ "$WARNINGS" -gt 0 ]; then
  echo ""
  echo "Validation PASSED with $WARNINGS warning(s)."
  exit 0
fi

echo ""
echo "Validation PASSED. All checks clean."
exit 0
