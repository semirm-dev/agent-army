#!/usr/bin/env bash
# edit-uses-rules.sh — Interactive CLI for adding/removing uses_rules entries.
# Rewrites the uses_rules line in-place, then auto-regenerates manifest.json.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"
MANIFEST="$REPO_ROOT/manifest.json"

# --- Helpers ---

# Read valid rule names from manifest.json.
get_valid_rules() {
  python3 -c "
import json, sys
d = json.load(open('$MANIFEST'))
for r in d['rules']:
    print(r['name'])
"
}

# Extract current uses_rules from a file's frontmatter (inline style only).
# Returns one rule per line.
get_current_uses_rules() {
  local file="$1"
  awk '
    /^---$/ { c++; next }
    c == 1 && /^uses_rules:/ {
      sub(/^uses_rules:[ \t]*/, "")
      if ($0 ~ /^\[/) {
        gsub(/[\[\]]/, "")
        n = split($0, arr, ",")
        for (i = 1; i <= n; i++) {
          gsub(/^[ \t]+|[ \t]+$/, "", arr[i])
          if (arr[i] != "") print arr[i]
        }
      }
      next
    }
    c >= 2 { exit }
  ' "$file"
}

# Rewrite uses_rules line in a file.
# Usage: write_uses_rules <file> <comma-separated-rules>
# If rules is empty, removes the uses_rules line entirely.
write_uses_rules() {
  local file="$1" rules="$2"

  if [ -z "$rules" ]; then
    # Remove the uses_rules line
    awk '
      /^---$/ { c++; print; next }
      c == 1 && /^uses_rules:/ { next }
      { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
  else
    # Check if uses_rules line exists
    local has_line
    has_line=$(awk '/^---$/{c++;next} c==1 && /^uses_rules:/{print "yes";exit}' "$file")

    if [ "$has_line" = "yes" ]; then
      # Replace existing line
      awk -v newval="uses_rules: [$rules]" '
        /^---$/ { c++; print; next }
        c == 1 && /^uses_rules:/ { print newval; next }
        { print }
      ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    else
      # Insert before closing ---
      awk -v newval="uses_rules: [$rules]" '
        /^---$/ { c++ }
        c == 2 { print newval; c = 3 }
        { print }
      ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    fi
  fi
}

# Prompt user to select one item from a numbered list.
# Usage: select_one "prompt" "${items[@]}"
# Sets REPLY to the selected value.
select_one() {
  local prompt="$1"; shift
  local items=("$@")
  local i

  echo ""
  for i in "${!items[@]}"; do
    printf "  %d) %s\n" "$((i + 1))" "${items[$i]}"
  done
  echo ""

  while true; do
    printf "%s " "$prompt"
    read -r choice
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#items[@]}" ]; then
      REPLY="${items[$((choice - 1))]}"
      return
    fi
    echo "Invalid choice. Enter a number between 1 and ${#items[@]}."
  done
}

# Prompt user to select multiple items from a numbered list.
# Usage: select_multi "prompt" "${items[@]}"
# Sets MULTI_REPLY array to selected values.
select_multi() {
  local prompt="$1"; shift
  local items=("$@")
  local i

  echo ""
  for i in "${!items[@]}"; do
    printf "  %d) %s\n" "$((i + 1))" "${items[$i]}"
  done
  echo ""

  while true; do
    printf "%s (comma-separated, e.g. 1,3,5): " "$prompt"
    read -r choices
    MULTI_REPLY=()
    local valid=1
    IFS=',' read -ra nums <<< "$choices"
    for n in "${nums[@]}"; do
      n="$(echo "$n" | tr -d ' ')"
      if [[ "$n" =~ ^[0-9]+$ ]] && [ "$n" -ge 1 ] && [ "$n" -le "${#items[@]}" ]; then
        MULTI_REPLY+=("${items[$((n - 1))]}")
      else
        echo "Invalid number: $n"
        valid=0
        break
      fi
    done
    [ "$valid" -eq 1 ] && [ "${#MULTI_REPLY[@]}" -gt 0 ] && return
  done
}

# --- Main flow ---

echo "=== Edit uses_rules ==="

# 1. Choose type
select_one "Choose type:" "rule" "skill"
TYPE="$REPLY"

if [ "$TYPE" = "rule" ]; then
  SEARCH_DIR="$RULES_DIR"
  PREFIX="rules/"
else
  SEARCH_DIR="$SKILLS_DIR"
  PREFIX="skills/"
fi

# 2. Choose file
files=()
while IFS= read -r _f; do
  files+=("$_f")
done < <(find "$SEARCH_DIR" -name '*.md' | sort)
if [ "${#files[@]}" -eq 0 ]; then
  echo "No files found in $SEARCH_DIR"
  exit 1
fi

display_names=()
for f in "${files[@]}"; do
  display_names+=("${f#"$SEARCH_DIR/"}")
done

select_one "Choose file:" "${display_names[@]}"
CHOSEN_DISPLAY="$REPLY"
CHOSEN_FILE="$SEARCH_DIR/$CHOSEN_DISPLAY"

# 3. Show current uses_rules
echo ""
echo "File: ${PREFIX}${CHOSEN_DISPLAY}"
current=()
while IFS= read -r _r; do
  [ -n "$_r" ] && current+=("$_r")
done < <(get_current_uses_rules "$CHOSEN_FILE")
if [ "${#current[@]}" -eq 0 ]; then
  echo "Current uses_rules: (none)"
else
  echo "Current uses_rules: [$(IFS=', '; echo "${current[*]}")]"
fi

# 4. Choose action
select_one "Action:" "add" "remove"
ACTION="$REPLY"

if [ "$ACTION" = "add" ]; then
  # Build list of valid rules not already present
  all_rules=()
  while IFS= read -r _r; do
    [ -n "$_r" ] && all_rules+=("$_r")
  done < <(get_valid_rules)
  available=()
  for r in "${all_rules[@]}"; do
    local_match=0
    for c in "${current[@]}"; do
      if [ "$r" = "$c" ]; then
        local_match=1
        break
      fi
    done
    [ "$local_match" -eq 0 ] && available+=("$r")
  done

  if [ "${#available[@]}" -eq 0 ]; then
    echo "All rules are already present. Nothing to add."
    exit 0
  fi

  select_multi "Select rules to add" "${available[@]}"

  # Compute new list
  new_rules=("${current[@]}" "${MULTI_REPLY[@]}")

elif [ "$ACTION" = "remove" ]; then
  if [ "${#current[@]}" -eq 0 ]; then
    echo "No uses_rules to remove."
    exit 0
  fi

  select_multi "Select rules to remove" "${current[@]}"

  # Compute new list (current minus selected)
  new_rules=()
  for c in "${current[@]}"; do
    keep=1
    for r in "${MULTI_REPLY[@]}"; do
      if [ "$c" = "$r" ]; then
        keep=0
        break
      fi
    done
    [ "$keep" -eq 1 ] && new_rules+=("$c")
  done
fi

# Build comma-separated string
new_csv=""
for r in "${new_rules[@]}"; do
  if [ -z "$new_csv" ]; then
    new_csv="$r"
  else
    new_csv="$new_csv, $r"
  fi
done

# 5. Show diff preview
echo ""
echo "--- Change Preview ---"
if [ "${#current[@]}" -eq 0 ]; then
  echo "  Before: (none)"
else
  echo "  Before: [$(IFS=', '; echo "${current[*]}")]"
fi
if [ -z "$new_csv" ]; then
  echo "  After:  (none — line will be removed)"
else
  echo "  After:  [$new_csv]"
fi
echo ""

# 6. Confirm
printf "Apply this change? [y/N] "
read -r confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# 7. Write file
write_uses_rules "$CHOSEN_FILE" "$new_csv"
echo "Updated ${PREFIX}${CHOSEN_DISPLAY}"

# 8. Regenerate manifest
echo ""
echo "Regenerating manifest.json..."
bash "$REPO_ROOT/scripts/generate-manifest.sh"
