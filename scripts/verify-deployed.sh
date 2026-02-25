#!/bin/bash
# Verifies that the deployed state (~/.claude/, ~/.cursor/rules/) matches the repo.
# Run after `make deploy` to confirm everything is in sync.

set -euo pipefail

LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CLAUDE_TARGET="$HOME/.claude"
CURSOR_TARGET="$HOME/.cursor/rules"

ERRORS=0

error() { echo "  ✗ $1"; ERRORS=$((ERRORS + 1)); }
ok()    { echo "  ✓ $1"; }

echo "=== Verifying Deployed State ==="
echo ""

# 1. Check Claude agents
echo "--- Claude agents (~/.claude/agents/) ---"
for agent_file in "$LIB_DIR/claude/agents"/*.md; do
  BASENAME=$(basename "$agent_file")
  DEPLOYED="$CLAUDE_TARGET/agents/$BASENAME"
  if [ -f "$DEPLOYED" ]; then
    if diff -q "$agent_file" "$DEPLOYED" >/dev/null 2>&1; then
      ok "$BASENAME matches"
    else
      error "$BASENAME exists but content differs"
    fi
  else
    error "$BASENAME not deployed"
  fi
done
echo ""

# 2. Check Claude rules
echo "--- Claude rules (~/.claude/rules/) ---"
for rule_file in "$LIB_DIR/claude/rules"/*.md; do
  BASENAME=$(basename "$rule_file")
  DEPLOYED="$CLAUDE_TARGET/rules/$BASENAME"
  if [ -f "$DEPLOYED" ]; then
    if diff -q "$rule_file" "$DEPLOYED" >/dev/null 2>&1; then
      ok "$BASENAME matches"
    else
      error "$BASENAME exists but content differs"
    fi
  else
    error "$BASENAME not deployed"
  fi
done
echo ""

# 3. Check CLAUDE.md
echo "--- CLAUDE.md ---"
if [ -f "$CLAUDE_TARGET/CLAUDE.md" ]; then
  if diff -q "$LIB_DIR/claude/CLAUDE.md" "$CLAUDE_TARGET/CLAUDE.md" >/dev/null 2>&1; then
    ok "CLAUDE.md matches"
  else
    error "CLAUDE.md content differs"
  fi
else
  error "CLAUDE.md not deployed"
fi
echo ""

# 4. Check Cursor rules
echo "--- Cursor rules (~/.cursor/rules/) ---"
for cursor_file in "$LIB_DIR/cursor"/*.mdc; do
  BASENAME=$(basename "$cursor_file")
  DEPLOYED="$CURSOR_TARGET/$BASENAME"
  if [ -f "$DEPLOYED" ]; then
    if diff -q "$cursor_file" "$DEPLOYED" >/dev/null 2>&1; then
      ok "$BASENAME matches"
    else
      error "$BASENAME exists but content differs"
    fi
  else
    error "$BASENAME not deployed"
  fi
done
echo ""

# 5. Check for stale deployed files not in repo
echo "--- Stale deployed files ---"
if [ -d "$CLAUDE_TARGET/agents" ]; then
  for deployed_file in "$CLAUDE_TARGET/agents"/*.md; do
    [ -f "$deployed_file" ] || continue
    BASENAME=$(basename "$deployed_file")
    if [ ! -f "$LIB_DIR/claude/agents/$BASENAME" ]; then
      error "Stale agent: ~/.claude/agents/$BASENAME (not in repo)"
    fi
  done
fi
if [ -d "$CLAUDE_TARGET/rules" ]; then
  for deployed_file in "$CLAUDE_TARGET/rules"/*.md; do
    [ -f "$deployed_file" ] || continue
    BASENAME=$(basename "$deployed_file")
    if [ ! -f "$LIB_DIR/claude/rules/$BASENAME" ]; then
      error "Stale rule: ~/.claude/rules/$BASENAME (not in repo)"
    fi
  done
fi
ok "Stale file check complete"
echo ""

# Summary
echo "=== Summary ==="
if [ "$ERRORS" -gt 0 ]; then
  echo "  $ERRORS difference(s) found. Run 'make deploy' to fix."
  exit 1
fi

echo "  Deployed state matches repo. All clean."
exit 0
