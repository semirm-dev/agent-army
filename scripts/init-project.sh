#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE="$SCRIPT_DIR/../templates/PROJECT-CLAUDE.md"

if [ -f "$PWD/CLAUDE.md" ]; then
  echo "CLAUDE.md already exists in $PWD. Aborting."
  exit 1
fi

if [ ! -f "$TEMPLATE" ]; then
  echo "ERROR: Template not found: $TEMPLATE"
  echo "  Expected at: templates/PROJECT-CLAUDE.md (relative to repo root)"
  echo "  Run this script from the agent-army repository root, or check that the template file exists."
  exit 1
fi

cp "$TEMPLATE" "$PWD/CLAUDE.md"
echo "Created CLAUDE.md in $PWD. Edit it to match your project."
