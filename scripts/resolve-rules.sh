#!/usr/bin/env bash
# resolve-rules.sh — Detect and remove redundant uses_rules entries across all files.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"

source "$REPO_ROOT/scripts/lib-deps.sh"

load_rule_deps

# --- Scan all files for redundancies ---

declare -a files_with_redundancies=()
declare -a file_paths=()
declare -a file_before=()
declare -a file_after=()
declare -a file_removed=()

scan_dir() {
  local dir="$1" prefix="$2"

  while IFS= read -r file; do
    local relpath="${file#"$dir/"}"
    local fm
    fm="$(get_frontmatter < "$file")"
    local uses
    uses="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"

    [ -z "$uses" ] && continue

    local redundancies
    redundancies="$(find_redundant_rules "$uses")"
    [ -z "$redundancies" ] && continue

    # Collect redundant rule names
    local -a redundant_names=()
    local -a reasons=()
    while IFS='|' read -r rname covered; do
      redundant_names+=("$rname")
      reasons+=("$rname (covered by $covered)")
    done <<< "$redundancies"

    # Build cleaned list
    IFS=',' read -ra orig_arr <<< "$uses"
    local -a cleaned=()
    for entry in "${orig_arr[@]}"; do
      entry="$(echo "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      local is_redundant=0
      for rn in "${redundant_names[@]}"; do
        if [ "$entry" = "$rn" ]; then
          is_redundant=1
          break
        fi
      done
      [ "$is_redundant" -eq 0 ] && cleaned+=("$entry")
    done

    local cleaned_csv=""
    for c in "${cleaned[@]}"; do
      if [ -z "$cleaned_csv" ]; then
        cleaned_csv="$c"
      else
        cleaned_csv="$cleaned_csv, $c"
      fi
    done

    files_with_redundancies+=("${prefix}${relpath}")
    file_paths+=("$file")
    file_before+=("$uses")
    file_after+=("$cleaned_csv")
    file_removed+=("$(printf '%s\n' "${reasons[@]}" | paste -sd ';' -)")
  done < <(find "$dir" -name '*.md' | sort)
}

scan_dir "$RULES_DIR" "rules/"
scan_dir "$SKILLS_DIR" "skills/"

# --- Report ---

if [ "${#files_with_redundancies[@]}" -eq 0 ]; then
  echo "No redundant uses_rules entries found."
  exit 0
fi

echo "=== Redundant uses_rules Report ==="
echo ""

total_removed=0
for i in "${!files_with_redundancies[@]}"; do
  echo "File: ${files_with_redundancies[$i]}"
  echo "  Before: [${file_before[$i]}]"
  echo "  After:  [${file_after[$i]}]"
  echo "  Removed:"
  IFS=';' read -ra reason_arr <<< "${file_removed[$i]}"
  for reason in "${reason_arr[@]}"; do
    echo "    - $reason"
    total_removed=$((total_removed + 1))
  done
  echo ""
done

echo "Total: $total_removed redundant entries across ${#files_with_redundancies[@]} files."
echo ""

# --- Confirm ---

printf "Remove these redundant entries? [y/N] "
read -r confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Aborted. No files changed."
  exit 0
fi

# --- Apply changes ---

echo ""
for i in "${!file_paths[@]}"; do
  write_uses_rules "${file_paths[$i]}" "${file_after[$i]}"
  echo "Updated ${files_with_redundancies[$i]}"
done

# --- Regenerate manifest ---

echo ""
echo "Regenerating manifest.json..."
bash "$REPO_ROOT/scripts/generate-manifest.sh"
