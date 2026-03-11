#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOC="$SCRIPT_DIR/../PLUGINS_AND_SKILLS.md"

if [ ! -f "$DOC" ]; then
  echo "Error: PLUGINS_AND_SKILLS.md not found at $DOC" >&2
  exit 1
fi

rm -f /tmp/sync_sh_error

echo "=== Installing Plugins ==="
# Extract plugin install commands: `/plugin install name@marketplace`
grep -oE '/plugin install [^ |`]+' "$DOC" | sed 's/`$//' | while read -r line; do
  cmd="claude ${line#/}"
  echo "→ $cmd"
  if ! $cmd < /dev/null; then
    echo "  ✗ Failed: $cmd" >&2
    touch /tmp/sync_sh_error
  fi
done

echo ""
echo "=== Installing Skills ==="
# Extract skill install commands from Skills tables.
# 1. Remove everything from "### Plugin-Provided Skills" onward
# 2. Extract `npx skills add ...` commands from backtick-delimited text
# Note: stdin is redirected to /dev/null for each command to prevent
# interactive tools from consuming the piped command list.
sed -n '1,/^### Plugin-Provided Skills/p' "$DOC" \
  | sed -n 's/.*`\(npx skills add [^`]*\)`.*/\1/p' \
  | grep -v '<' \
  | while read -r cmd; do
    full_cmd="$cmd -y"
    echo "→ $full_cmd"
    if ! $full_cmd < /dev/null; then
      echo "  ✗ Failed: $full_cmd" >&2
      touch /tmp/sync_sh_error
    fi
  done

if [ -f /tmp/sync_sh_error ]; then
  rm -f /tmp/sync_sh_error
  echo ""
  echo "Some installations failed. Check output above."
  exit 1
fi

echo ""
echo "Done. All plugins and skills installed."
