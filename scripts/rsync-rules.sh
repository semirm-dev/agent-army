#!/bin/bash
set -euo pipefail

# 1. Master Library Path & shared config
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib.sh"
require_jq

# 2. Check for folder argument
if [ -z "$1" ]; then
  echo "Usage: rsync-rules.sh <folder>"
  echo "  folder: cursor, claude"
  echo ""
  echo "Examples:"
  echo "  rsync-rules.sh cursor   # Syncs cursor/ -> ~/.cursor/rules/"
  echo "  rsync-rules.sh claude   # Syncs claude/ -> ~/.claude/ (includes agents/)"
  exit 1
fi

FOLDER="$1"

# 3. Smart target mapping
case "$FOLDER" in
  cursor) TARGET_DIR="$HOME/.cursor/rules" ;;
  claude) TARGET_DIR="$HOME/.claude" ;;
  *)
    echo "Unknown folder: $FOLDER"
    echo "Valid folders: cursor, claude"
    exit 1
    ;;
esac

# 4. Create target directory
mkdir -p "$TARGET_DIR"

echo "🔄 Mirroring $FOLDER/ to: $TARGET_DIR"

# 5. Mirror all files (exclude user-managed files that shouldn't be overwritten)
EXCLUDES=()
if [ "$FOLDER" = "claude" ]; then
  while read -r exclude; do
    EXCLUDES+=(--exclude "$exclude")
  done < <(cfg '.rsync_excludes[]')
fi
# Cursor agents live in ~/.cursor/agents/, not under rules/ — exclude from rules rsync
if [ "$FOLDER" = "cursor" ]; then
  EXCLUDES+=(--exclude "agents/")
fi

rsync -av ${EXCLUDES[@]+"${EXCLUDES[@]}"} "$LIB_DIR/$FOLDER/" "$TARGET_DIR/" \
  || { echo "rsync failed"; exit 1; }

# 6. Deploy shared skills to Cursor (source of truth is claude/skills/)
if [ "$FOLDER" = "cursor" ]; then
  CURSOR_SKILLS_DIR="$HOME/.cursor/skills"
  mkdir -p "$CURSOR_SKILLS_DIR"
  echo "🔄 Syncing custom skills to: $CURSOR_SKILLS_DIR"
  rsync -av "$LIB_DIR/claude/skills/" "$CURSOR_SKILLS_DIR/" \
    || { echo "skills rsync failed"; exit 1; }
fi

# 7. Deploy Cursor-native agents + plugin agents
if [ "$FOLDER" = "cursor" ]; then
  CURSOR_AGENTS_DIR="$HOME/.cursor/agents"
  mkdir -p "$CURSOR_AGENTS_DIR"

  # Custom agents (source of truth: cursor/agents/)
  if [ -d "$LIB_DIR/cursor/agents" ]; then
    echo "🔄 Syncing Cursor agents to: $CURSOR_AGENTS_DIR"
    rsync -av "$LIB_DIR/cursor/agents/" "$CURSOR_AGENTS_DIR/" \
      || { echo "agents rsync failed"; exit 1; }
  fi

  # Plugin agents from Claude plugin cache (third-party, self-updating)
  PLUGIN_CACHE="$HOME/.claude/plugins/cache"
  if [ -d "$PLUGIN_CACHE" ]; then
    echo "🔄 Syncing plugin agents to: $CURSOR_AGENTS_DIR"
    find "$PLUGIN_CACHE" -path "*/agents/*.md" -exec cp -v {} "$CURSOR_AGENTS_DIR/" \;
  fi
fi

echo "🎉 Done. Rules are now physically mirrored (fixes Cursor indexing bugs)."
