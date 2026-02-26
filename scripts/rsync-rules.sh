#!/bin/bash
set -euo pipefail
# 1. Master Library Path
LIB_DIR="$(cd "$(dirname "$0")/.." && pwd)"

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
  EXCLUDES=(
    --exclude settings.json
    --exclude installed_plugins.json
    --exclude skills/
    --exclude plugins/
    --exclude projects/
    --exclude todos/
  )
fi

rsync -av ${EXCLUDES[@]+"${EXCLUDES[@]}"} "$LIB_DIR/$FOLDER/" "$TARGET_DIR/" \
  || { echo "rsync failed"; exit 1; }

echo "🎉 Done. Rules are now physically mirrored (fixes Cursor indexing bugs)."
