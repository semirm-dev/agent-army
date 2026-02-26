#!/bin/bash
# bootstrap.sh — Interactive setup for agent-rules on a new device.
# Each step asks for confirmation before executing.
#
# Usage: ./scripts/bootstrap.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; }

step() {
  echo ""
  echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${BOLD}${CYAN}  $1${NC}"
  echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

ask() {
  local prompt="$1"
  local answer
  echo ""
  read -rp "$prompt [y/N] " answer
  [[ "$answer" =~ ^[Yy]$ ]]
}

echo ""
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}${CYAN}  agent-rules bootstrap${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Repo: $LIB_DIR"

# ── Step 1: Prerequisites ──────────────────────────────────────────

step "Step 1: Check prerequisites"

MISSING=()
command -v node   >/dev/null 2>&1 && ok "node $(node -v)"   || MISSING+=("node")
command -v npx    >/dev/null 2>&1 && ok "npx found"          || MISSING+=("npx")
command -v jq     >/dev/null 2>&1 && ok "jq $(jq --version)" || MISSING+=("jq")
command -v claude >/dev/null 2>&1 && ok "claude CLI found"   || MISSING+=("claude")
command -v rsync  >/dev/null 2>&1 && ok "rsync found"        || MISSING+=("rsync")

if [ ${#MISSING[@]} -gt 0 ]; then
  fail "Missing: ${MISSING[*]}"
  echo "Install missing tools before continuing."
  exit 1
fi

# ── Step 2: Sync rules ─────────────────────────────────────────────

step "Step 2: Sync rules"
if ask "Sync rules to ~/.claude/ and ~/.cursor/rules/?"; then
  bash "$LIB_DIR/scripts/rsync-rules.sh" claude
  bash "$LIB_DIR/scripts/rsync-rules.sh" cursor
  ok "Rules synced"
else
  warn "Skipped rule sync"
fi

# ── Step 3: Install Agent Skills ──────────────────────────────────

step "Step 3: Install Agent Skills"
SKILL_COUNT=$(cfg_raw '.npm_skills' | jq 'length')
echo "  Skills to install:"
while read -r skill_json; do
  name=$(echo "$skill_json" | jq -r '.name')
  echo "    - $name"
done < <(cfg_raw '.npm_skills[]')
if ask "Install $SKILL_COUNT skills?"; then
  while read -r skill_json; do
    name=$(echo "$skill_json" | jq -r '.name')
    repo=$(echo "$skill_json" | jq -r '.repo')
    if [ -L "$HOME/.claude/skills/$name" ] || [ -d "$HOME/.claude/skills/$name" ]; then
      ok "$name already installed, skipping"
    else
      echo "  Installing $name..."
      npx skills add "$repo" --skill "$name" -g -y || warn "Failed to install $name"
    fi
  done < <(cfg_raw '.npm_skills[]')
  ok "Skills installed"
else
  warn "Skipped skill installation"
fi

# ── Step 4: Install Claude Plugins ────────────────────────────────

step "Step 4: Install Claude Plugins"
echo "  Generating settings.json from config.json..."
bash "$LIB_DIR/scripts/generate-settings.sh"
SETTINGS_SRC="$LIB_DIR/claude/settings.json"
SETTINGS_DST="$HOME/.claude/settings.json"
echo "  Plugins enabled via settings.json:"
while read -r plugin; do
  echo "    - $plugin"
done < <(cfg '.plugins[]')
if ask "Deploy settings.json to $SETTINGS_DST?"; then
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

step "Step 5: Add shell aliases"
ALIAS_COUNT=$(cfg_raw '.shell_aliases' | jq 'length')
echo "  Aliases to add:"
while read -r alias_json; do
  name=$(echo "$alias_json" | jq -r '.name')
  desc=$(echo "$alias_json" | jq -r '.description')
  echo "    - $name — $desc"
done < <(cfg_raw '.shell_aliases[]')
if ask "Add $ALIAS_COUNT aliases to ~/.zshrc?"; then
  ZSHRC="$HOME/.zshrc"
  touch "$ZSHRC"
  while read -r alias_json; do
    alias_name=$(echo "$alias_json" | jq -r '.name')
    alias_script=$(echo "$alias_json" | jq -r '.script')
    alias_line="alias ${alias_name}='$LIB_DIR/${alias_script}'"
    if grep -qF "$alias_line" "$ZSHRC"; then
      ok "Already present: $alias_line"
    else
      echo "$alias_line" >> "$ZSHRC"
      ok "Added: $alias_line"
    fi
  done < <(cfg_raw '.shell_aliases[]')
else
  warn "Skipped alias setup"
fi

# ── Step 6: Verify ─────────────────────────────────────────────────

step "Step 6: Verify installation"

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
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BOLD}${GREEN}  Bootstrap complete${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BOLD}${YELLOW}  Next steps:${NC}"
echo -e "    1. source ~/.zshrc            (or open a new terminal to load aliases)"
echo -e "    2. cd <your-project> && claude (start using Claude Code with the new rules)"
echo ""
echo -e "${BOLD}${YELLOW}  Day-to-day after editing rules in this repo:${NC}"
echo -e "    sync-rules claude && sync-rules cursor"
