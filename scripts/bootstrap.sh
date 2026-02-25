#!/bin/bash
# bootstrap.sh — Interactive setup for agent-rules on a new device.
# Each step asks for confirmation before executing.
#
# Usage: ./scripts/bootstrap.sh

set -euo pipefail

LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; }

ask() {
  local prompt="$1"
  local answer
  echo ""
  read -rp "$prompt [y/N] " answer
  [[ "$answer" =~ ^[Yy]$ ]]
}

echo "================================="
echo "  agent-rules bootstrap"
echo "================================="
echo ""
echo "Repo: $LIB_DIR"
echo ""

# ── Step 1: Prerequisites ──────────────────────────────────────────

echo "── Step 1: Check prerequisites ──"

MISSING=()
command -v node   >/dev/null 2>&1 && ok "node $(node -v)"   || MISSING+=("node")
command -v npx    >/dev/null 2>&1 && ok "npx found"          || MISSING+=("npx")
command -v claude >/dev/null 2>&1 && ok "claude CLI found"   || MISSING+=("claude")
command -v rsync  >/dev/null 2>&1 && ok "rsync found"        || MISSING+=("rsync")

if [ ${#MISSING[@]} -gt 0 ]; then
  fail "Missing: ${MISSING[*]}"
  echo "Install missing tools before continuing."
  exit 1
fi

# ── Step 2: Sync rules ─────────────────────────────────────────────

if ask "Step 2: Sync rules to ~/.claude/ and ~/.cursor/rules/?"; then
  bash "$LIB_DIR/scripts/rsync-rules.sh" claude
  bash "$LIB_DIR/scripts/rsync-rules.sh" cursor
  ok "Rules synced"
else
  warn "Skipped rule sync"
fi

# ── Step 3: Install npm skills ─────────────────────────────────────

echo "  Skills to install:"
echo "    - golang-pro"
echo "    - browser-use"
echo "    - database-schema-designer"
echo "    - skill-creator"
echo "    - find-skills"
if ask "Step 3: Install npm skills (5 skills)?"; then
  SKILLS=(
    "https://github.com/jeffallan/claude-skills --skill golang-pro"
    "https://github.com/browser-use/browser-use --skill browser-use"
    "https://github.com/softaworks/agent-toolkit --skill database-schema-designer"
    "https://github.com/anthropics/skills --skill skill-creator"
    "https://github.com/anthropics/skills --skill find-skills"
  )
  for skill_cmd in "${SKILLS[@]}"; do
    skill_name="${skill_cmd##*--skill }"
    if [ -L "$HOME/.claude/skills/$skill_name" ] || [ -d "$HOME/.claude/skills/$skill_name" ]; then
      ok "$skill_name already installed, skipping"
    else
      echo "  Installing $skill_name..."
      npx skills add "$skill_cmd" || warn "Failed to install $skill_name"
    fi
  done
  ok "Skills installed"
else
  warn "Skipped skill installation"
fi

# ── Step 4: Deploy settings.json (enables plugins) ────────────────

SETTINGS_SRC="$LIB_DIR/claude/settings.json"
SETTINGS_DST="$HOME/.claude/settings.json"

echo "  Plugins enabled via settings.json:"
echo "    - context7"
echo "    - frontend-design"
echo "    - code-review"
echo "    - superpowers"
echo "    - security-guidance"
echo "    - code-simplifier"
if ask "Step 4: Deploy settings.json to $SETTINGS_DST?"; then
  if [ -f "$SETTINGS_DST" ]; then
    echo "  Current diff:"
    diff --unified=3 "$SETTINGS_DST" "$SETTINGS_SRC" || true
    if ask "  Overwrite existing settings.json?"; then
      cp "$SETTINGS_SRC" "$SETTINGS_DST"
      ok "Settings deployed"
    else
      warn "Kept existing settings.json"
    fi
  else
    mkdir -p "$(dirname "$SETTINGS_DST")"
    cp "$SETTINGS_SRC" "$SETTINGS_DST"
    ok "Settings deployed"
  fi
else
  warn "Skipped settings deployment"
fi

# ── Step 5: Add shell aliases ──────────────────────────────────────

if ask "Step 5: Add aliases to ~/.zshrc?"; then
  ALIASES=(
    "alias sync-rules='$LIB_DIR/scripts/rsync-rules.sh'"
    "alias check-sync='$LIB_DIR/scripts/check-sync.sh'"
  )
  ZSHRC="$HOME/.zshrc"
  touch "$ZSHRC"
  for alias_line in "${ALIASES[@]}"; do
    if grep -qF "$alias_line" "$ZSHRC"; then
      ok "Already present: $alias_line"
    else
      echo "$alias_line" >> "$ZSHRC"
      ok "Added: $alias_line"
    fi
  done
else
  warn "Skipped alias setup"
fi

# ── Step 6: Verify ─────────────────────────────────────────────────

echo ""
echo "── Verification ──"

echo "Skills installed:"
ls "$HOME/.claude/skills/" 2>/dev/null || warn "No skills directory"

echo ""
echo "Agents available:"
ls "$HOME/.claude/agents/" 2>/dev/null || warn "No agents directory"

echo ""
if [ -x "$LIB_DIR/scripts/check-sync.sh" ]; then
  echo "Running check-sync..."
  bash "$LIB_DIR/scripts/check-sync.sh" || warn "Sync drift detected (see above)"
else
  warn "check-sync.sh not found or not executable"
fi

echo ""
echo "================================="
echo "  Bootstrap complete"
echo "================================="
