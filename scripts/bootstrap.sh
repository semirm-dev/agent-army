#!/bin/bash
# bootstrap.sh — Interactive setup for agent-army on a new device.
# Each step asks for confirmation before executing.
#
# Usage: ./scripts/bootstrap.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

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
echo -e "${BOLD}${CYAN}  agent-army bootstrap${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Repo: $LIB_DIR"

# ── Step 1: Prerequisites ──────────────────────────────────────────

step "Step 1: Check prerequisites"

MISSING=()
command -v node   >/dev/null 2>&1 && ok "node $(node -v)"         || { fail "node";   MISSING+=("node"); }
command -v npx    >/dev/null 2>&1 && ok "npx $(npx --version)"    || { fail "npx";    MISSING+=("npx"); }
command -v jq     >/dev/null 2>&1 && ok "jq $(jq --version)"      || { fail "jq";     MISSING+=("jq"); }
command -v claude >/dev/null 2>&1 && ok "claude"                   || { fail "claude"; MISSING+=("claude"); }
command -v rsync  >/dev/null 2>&1 && ok "rsync"                    || { fail "rsync";  MISSING+=("rsync"); }

if [ ${#MISSING[@]} -gt 0 ]; then
  echo ""
  warn "Missing tools: ${MISSING[*]}"

  IS_MAC=false
  [[ "$(uname)" == "Darwin" ]] && IS_MAC=true

  install_cmd_for() {
    local tool="$1"
    case "$tool" in
      node)   $IS_MAC && echo "brew install node"  || echo "sudo apt-get install -y nodejs" ;;
      npx)    echo "npm install -g npm" ;;
      jq)     $IS_MAC && echo "brew install jq"    || echo "sudo apt-get install -y jq" ;;
      claude) echo "npm install -g @anthropic-ai/claude-code" ;;
      rsync)  $IS_MAC && echo "brew install rsync" || echo "sudo apt-get install -y rsync" ;;
      *)      echo "" ;;
    esac
  }

  STILL_MISSING=()
  for tool in "${MISSING[@]}"; do
    # npx ships with node — skip if node is also being installed
    if [[ "$tool" == "npx" ]] && printf '%s\n' "${MISSING[@]}" | grep -qx "node"; then
      echo "  (npx will be installed with node)"
      continue
    fi

    cmd=$(install_cmd_for "$tool")
    if [ -z "$cmd" ]; then
      STILL_MISSING+=("$tool")
      continue
    fi

    if ask "  Install $tool? ($cmd)"; then
      if eval "$cmd"; then
        ok "$tool installed"
      else
        fail "Failed to install $tool"
        STILL_MISSING+=("$tool")
      fi
    else
      STILL_MISSING+=("$tool")
    fi
  done

  if [ ${#STILL_MISSING[@]} -gt 0 ]; then
    fail "Still missing: ${STILL_MISSING[*]}"
    echo "Install missing tools before continuing."
    exit 1
  fi
  ok "All prerequisites satisfied"
fi

require_jq

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
    if [ -d "$HOME/.agents/skills/$name" ] || [ -d "$HOME/.claude/skills/$name" ]; then
      ok "$name already installed, skipping"
    else
      echo "  Installing $name..."
      npx --yes skills add "$repo" --skill "$name" -y -g < /dev/null || warn "Failed to install $name"
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
while read -r plugin_json; do
  pname=$(echo "$plugin_json" | jq -r '.name')
  pmkt=$(echo "$plugin_json" | jq -r '.marketplace')
  echo "    - ${pname}@${pmkt}"
done < <(cfg_raw '.plugins[]')
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

if ask "Install plugins via claude CLI?"; then
  bash "$LIB_DIR/scripts/sync-plugins.sh"
  ok "Plugins installed"
else
  warn "Skipped plugin installation (run 'bash scripts/sync-plugins.sh' later)"
fi

# ── Step 5: Verify ─────────────────────────────────────────────────

step "Step 5: Verify installation"

echo "Skills installed (npm / global):"
ls "$HOME/.agents/skills/" 2>/dev/null || warn "No .agents/skills directory"
echo ""
echo "Skills installed (custom / claude):"
ls "$HOME/.claude/skills/" 2>/dev/null || warn "No .claude/skills directory"

echo ""
echo "Agents available:"
ls "$HOME/.claude/agents/" 2>/dev/null || warn "No agents directory"

echo ""
echo "Plugins configured:"
if command -v claude >/dev/null 2>&1; then
  plugin_list=$(claude plugin list --scope user 2>/dev/null || true)
  if [ -n "$plugin_list" ]; then
    echo "$plugin_list"
  else
    while read -r pjson; do
      pname=$(echo "$pjson" | jq -r '.name')
      pmkt=$(echo "$pjson" | jq -r '.marketplace')
      echo "  - ${pname}@${pmkt}"
    done < <(cfg_raw '.plugins[]')
    echo "  (listed from config.json — 'claude plugin list' returned nothing)"
  fi
else
  warn "claude CLI not found — cannot list plugins"
fi

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
echo -e "    cd <your-project> && claude    (start using Claude Code with the new rules)"
echo ""
echo -e "${BOLD}${YELLOW}  Day-to-day after editing rules in this repo:${NC}"
echo -e "    make bootstrap"
