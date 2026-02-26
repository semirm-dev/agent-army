#!/bin/bash
# Validates the structural integrity of the agent-rules repository.
# Checks: agent files exist, rule files exist, agent triads complete,
# synced pairs registered, CLAUDE.md references valid.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq
CLAUDE_MD="$LIB_DIR/claude/CLAUDE.md"
AGENTS_DIR="$LIB_DIR/claude/agents"
RULES_DIR="$LIB_DIR/claude/rules"
CURSOR_DIR="$LIB_DIR/cursor"

CURSOR_AGENTS_DIR="$LIB_DIR/cursor/agents"

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
  # Skip docker (builder+reviewer+tester triad), arch (reviewer only), docs (writer only)
  if [ "$prefix" = "docker" ]; then
    HAS_BUILDER=false
    HAS_REVIEWER=false
    HAS_TESTER=false
    [ -f "$AGENTS_DIR/docker-builder.md" ] && HAS_BUILDER=true
    [ -f "$AGENTS_DIR/docker-reviewer.md" ] && HAS_REVIEWER=true
    [ -f "$AGENTS_DIR/docker-tester.md" ] && HAS_TESTER=true
    if $HAS_BUILDER && $HAS_REVIEWER && $HAS_TESTER; then
      ok "docker: builder + reviewer + tester present"
    else
      MISSING=""
      $HAS_BUILDER || MISSING="$MISSING builder"
      $HAS_REVIEWER || MISSING="$MISSING reviewer"
      $HAS_TESTER || MISSING="$MISSING tester"
      warn "docker: missing$MISSING"
    fi
    continue
  fi

  # arch-reviewer is a standalone agent (no triad)
  if [ "$prefix" = "arch" ]; then
    [ -f "$AGENTS_DIR/arch-reviewer.md" ] && ok "arch: reviewer present (standalone)" || warn "arch: reviewer missing"
    continue
  fi

  # docs-writer is a standalone agent (no triad)
  if [ "$prefix" = "docs" ]; then
    [ -f "$AGENTS_DIR/docs-writer.md" ] && ok "docs: writer present (standalone)" || warn "docs: writer missing"
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
while IFS='|' read -r _ rule_col cursor_col _; do
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
done < <(grep -E 'rules/.*cursor/' "$CLAUDE_MD")
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
if [ -d "$LIB_DIR/claude/skills" ]; then
  SKILL_COUNT=$(find "$LIB_DIR/claude/skills" -maxdepth 2 -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')
  ok "$SKILL_COUNT custom skill(s) found"
  for skill_dir in "$LIB_DIR/claude/skills"/*/; do
    [ -f "$skill_dir/SKILL.md" ] && ok "  $(basename "$skill_dir")/SKILL.md"
  done
else
  warn "No claude/skills/ directory found"
fi
echo ""

# 7. Check skill files referenced by agents
echo "--- Skill references in agents ---"
for agent_file in "$AGENTS_DIR"/*.md; do
  AGENT_NAME=$(basename "$agent_file")
  # Match skill names in backtick-quoted references (pattern built from config.json)
  SKILL_PATTERN=$(cfg '.custom_skills[].name' | paste -sd'|' -)
  AGENT_SKILLS=$(grep -oE "\`(${SKILL_PATTERN})\`" "$agent_file" 2>/dev/null | tr -d '`' | sort -u || true)
  for skill in $AGENT_SKILLS; do
    if [ -f "$LIB_DIR/claude/skills/${skill}/SKILL.md" ]; then
      ok "$AGENT_NAME → claude/skills/${skill}/SKILL.md"
    else
      error "$AGENT_NAME references skill '${skill}' but claude/skills/${skill}/SKILL.md doesn't exist"
    fi
  done
done
echo ""

# 8. Check CLAUDE.md sync pairs table matches config.json
echo "--- CLAUDE.md sync pairs table vs config.json ---"

# Extract rule filenames from CLAUDE.md table rows (lines containing both rules/ and cursor/)
CLAUDE_RULES=$(grep -E 'rules/.*cursor/' "$CLAUDE_MD" | grep -v '^\s*<!--' | grep -oE 'rules/[a-z-]+\.md' | sort)

# Extract rule filenames from config.json
CONFIG_RULES=$(cfg '.sync_pairs[].rule' | sed 's/^/rules\//' | sort)

if [ "$CLAUDE_RULES" = "$CONFIG_RULES" ]; then
  ok "CLAUDE.md table matches config.json sync_pairs ($(echo "$CONFIG_RULES" | wc -l | tr -d ' ') pairs)"
else
  # Find entries in config but missing from CLAUDE.md
  MISSING_FROM_CLAUDE=$(comm -23 <(echo "$CONFIG_RULES") <(echo "$CLAUDE_RULES"))
  # Find entries in CLAUDE.md but missing from config
  MISSING_FROM_CONFIG=$(comm -13 <(echo "$CONFIG_RULES") <(echo "$CLAUDE_RULES"))

  if [ -n "$MISSING_FROM_CLAUDE" ]; then
    error "config.json has sync pairs not in CLAUDE.md table: $MISSING_FROM_CLAUDE"
  fi
  if [ -n "$MISSING_FROM_CONFIG" ]; then
    error "CLAUDE.md table has sync pairs not in config.json: $MISSING_FROM_CONFIG"
  fi
fi
echo ""

# 9. Check CLAUDE.md Custom Skills list matches config.json custom_skills
echo "--- CLAUDE.md Custom Skills vs config.json ---"

# Extract skill names from CLAUDE.md Custom Skills section only (between "Custom Skills:" and "Plugins (superpowers):")
CLAUDE_SKILLS=$(sed -n '/\*\*Custom Skills:\*\*/,/\*\*Plugins (superpowers):\*\*/p' "$CLAUDE_MD" | grep -E '^\s+- `[a-z-]+` --' | grep -oE '`[a-z-]+`' | tr -d '`' | sort)

# Extract skill names from config.json
CONFIG_SKILLS=$(cfg '.custom_skills[].name' | sort)

if [ "$CLAUDE_SKILLS" = "$CONFIG_SKILLS" ]; then
  ok "CLAUDE.md Custom Skills matches config.json ($(echo "$CONFIG_SKILLS" | wc -l | tr -d ' ') skills)"
else
  MISSING_FROM_CLAUDE=$(comm -23 <(echo "$CONFIG_SKILLS") <(echo "$CLAUDE_SKILLS"))
  MISSING_FROM_CONFIG=$(comm -13 <(echo "$CONFIG_SKILLS") <(echo "$CLAUDE_SKILLS"))

  if [ -n "$MISSING_FROM_CLAUDE" ]; then
    for s in $MISSING_FROM_CLAUDE; do
      error "config.json has skill '$s' not listed in CLAUDE.md Custom Skills"
    done
  fi
  if [ -n "$MISSING_FROM_CONFIG" ]; then
    for s in $MISSING_FROM_CONFIG; do
      error "CLAUDE.md lists skill '$s' not in config.json custom_skills"
    done
  fi
fi
echo ""

# 10. Check cursor/agents/ parity with claude/agents/
echo "--- Cursor agent parity ---"
if [ -d "$CURSOR_AGENTS_DIR" ]; then
  CLAUDE_AGENT_COUNT=$(find "$AGENTS_DIR" -maxdepth 1 -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
  CURSOR_AGENT_COUNT=$(find "$CURSOR_AGENTS_DIR" -maxdepth 1 -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
  ok "cursor/agents/ has $CURSOR_AGENT_COUNT agents (claude/agents/ has $CLAUDE_AGENT_COUNT)"

  # Check each Claude agent has a Cursor counterpart
  for agent_file in "$AGENTS_DIR"/*.md; do
    AGENT_NAME=$(basename "$agent_file")
    if [ -f "$CURSOR_AGENTS_DIR/$AGENT_NAME" ]; then
      ok "cursor/agents/$AGENT_NAME"
    else
      error "claude/agents/$AGENT_NAME has no cursor counterpart at cursor/agents/$AGENT_NAME"
    fi
  done
else
  error "cursor/agents/ directory not found — Cursor subagents will not work"
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
