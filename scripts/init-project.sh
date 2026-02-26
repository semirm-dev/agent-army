#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE="$SCRIPT_DIR/../templates/PROJECT-CLAUDE.md"

if [ -f "$PWD/CLAUDE.md" ]; then
  echo "CLAUDE.md already exists in $PWD. Aborting."
  exit 1
fi

if [ ! -f "$TEMPLATE" ]; then
  echo "Template not found: $TEMPLATE"
  exit 1
fi

cp "$TEMPLATE" "$PWD/CLAUDE.md"
echo "Created CLAUDE.md in $PWD. Edit it to match your project."
