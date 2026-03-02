#!/usr/bin/env bash
# lib-deps.sh — Shared dependency library for uses_rules resolution.
# Source this file; do not execute directly.
#
# Provides:
#   get_frontmatter()          — extract YAML frontmatter from stdin
#   extract_fm_value()         — extract scalar value from frontmatter
#   extract_fm_list()          — extract list value from frontmatter
#   write_uses_rules()         — rewrite uses_rules line in a file
#   load_rule_deps()           — scan rules/*.md, populate _LIB_RULE_NAMES / _LIB_RULE_USES_RULES
#   _lib_rule_deps()           — lookup deps for a rule name
#   lib_resolve_uses_rules()   — BFS transitive resolution
#   find_redundant_rules()     — detect entries covered transitively by another entry

# Guard against direct execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "Error: lib-deps.sh must be sourced, not executed." >&2
  exit 1
fi

# Detect repo root relative to this file
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
_LIB_REPO_ROOT="$(cd "$_LIB_DIR/.." && pwd)"

# --- Frontmatter parsing ---

# Extract YAML frontmatter block (lines between the two --- markers, exclusive).
get_frontmatter() {
  awk '/^---$/ { count++; next } count == 1 { print } count >= 2 { exit }'
}

# Extract a single scalar value from frontmatter.
# Usage: extract_fm_value "key" "$frontmatter"
extract_fm_value() {
  local key="$1" fm="$2"
  local val
  val="$(echo "$fm" | awk -v k="$key" '$0 ~ "^"k":" { sub("^"k":[ \t]*", ""); print; exit }')"
  val="${val#\"}" ; val="${val%\"}"
  val="${val#\'}" ; val="${val%\'}"
  printf '%s' "$val"
}

# Extract a YAML list (inline [a, b] or block - items) from frontmatter.
# Usage: extract_fm_list "key" "$frontmatter"
# Returns values one per line.
extract_fm_list() {
  local key="$1" fm="$2"

  if ! echo "$fm" | grep -q "^${key}:"; then
    return
  fi

  local line
  line="$(echo "$fm" | awk -v k="$key" '$0 ~ "^"k":" { sub("^"k":[ \t]*", ""); print; exit }')"

  # Inline: [a, b, c]
  if [[ "$line" == \[* ]]; then
    echo "$line" | tr -d '[]' | tr ',' '\n' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' | sed '/^$/d'
    return
  fi

  # Block list: lines starting with "  - " after the key line
  local in_list=0
  while IFS= read -r l; do
    if [[ "$l" == "${key}:"* ]]; then
      in_list=1
      continue
    fi
    if [ "$in_list" -eq 1 ]; then
      if [[ "$l" =~ ^[[:space:]]*-[[:space:]] ]]; then
        echo "$l" | sed 's/^[[:space:]]*-[[:space:]]*//' | sed 's/^"//; s/"$//; s/^'"'"'//; s/'"'"'$//'
      else
        break
      fi
    fi
  done <<< "$fm"
}

# --- File writing ---

# Rewrite uses_rules line in a file.
# Usage: write_uses_rules <file> <comma-separated-rules>
# If rules is empty, removes the uses_rules line entirely.
write_uses_rules() {
  local file="$1" rules="$2"

  if [ -z "$rules" ]; then
    awk '
      /^---$/ { c++; print; next }
      c == 1 && /^uses_rules:/ { next }
      { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
  else
    local has_line
    has_line=$(awk '/^---$/{c++;next} c==1 && /^uses_rules:/{print "yes";exit}' "$file")

    if [ "$has_line" = "yes" ]; then
      awk -v newval="uses_rules: [$rules]" '
        /^---$/ { c++; print; next }
        c == 1 && /^uses_rules:/ { print newval; next }
        { print }
      ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    else
      awk -v newval="uses_rules: [$rules]" '
        /^---$/ { c++ }
        c == 2 { print newval; c = 3 }
        { print }
      ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    fi
  fi
}

# --- Rule dependency loading ---

declare -a _LIB_RULE_NAMES=()
declare -a _LIB_RULE_USES_RULES=()

# Scan rules/*.md and populate _LIB_RULE_NAMES / _LIB_RULE_USES_RULES arrays.
load_rule_deps() {
  _LIB_RULE_NAMES=()
  _LIB_RULE_USES_RULES=()

  local rules_dir="$_LIB_REPO_ROOT/rules"

  while IFS= read -r file; do
    local relpath="${file#"$rules_dir/"}"
    local name="${relpath%.md}"
    local fm
    fm="$(get_frontmatter < "$file")"
    local uses
    uses="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"

    _LIB_RULE_NAMES+=("$name")
    _LIB_RULE_USES_RULES+=("$uses")
  done < <(find "$rules_dir" -name '*.md' | sort)
}

# Lookup deps for a rule name from loaded arrays.
_lib_rule_deps() {
  local name="$1"
  for j in "${!_LIB_RULE_NAMES[@]}"; do
    if [ "${_LIB_RULE_NAMES[$j]}" = "$name" ]; then
      printf '%s' "${_LIB_RULE_USES_RULES[$j]}"
      return
    fi
  done
}

# --- Transitive resolution ---

# BFS transitive resolution of uses_rules.
# Takes a comma-separated list, returns deduplicated comma-separated list
# including all transitive dependencies.
lib_resolve_uses_rules() {
  local input="$1"
  [ -z "$input" ] && return

  local visited=""
  local result=""
  local queue="$input"

  while [ -n "$queue" ]; do
    local current
    current="$(echo "$queue" | cut -d',' -f1 | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    local rest
    rest="$(echo "$queue" | cut -d',' -f2- -s)"
    queue="$rest"

    [ -z "$current" ] && continue

    # Check if already visited
    local already=0
    if [ -n "$visited" ]; then
      IFS=',' read -ra vis_arr <<< "$visited"
      for v in "${vis_arr[@]}"; do
        if [ "$v" = "$current" ]; then
          already=1
          break
        fi
      done
    fi
    [ "$already" -eq 1 ] && continue

    # Mark visited and add to result
    if [ -n "$visited" ]; then
      visited="${visited},${current}"
    else
      visited="$current"
    fi
    if [ -n "$result" ]; then
      result="${result},${current}"
    else
      result="$current"
    fi

    # Enqueue transitive deps
    local deps
    deps="$(_lib_rule_deps "$current")"
    if [ -n "$deps" ] && [ -n "$queue" ]; then
      queue="${queue},${deps}"
    elif [ -n "$deps" ]; then
      queue="$deps"
    fi
  done

  printf '%s' "$result"
}

# --- Redundancy detection ---

# For each rule R in a comma-separated list, check if any *other* rule O
# in the same list has R in its transitive closure.
# Output: one "redundant|covered_by" line per redundant entry.
find_redundant_rules() {
  local csv="$1"
  [ -z "$csv" ] && return

  # Split into array
  local -a entries=()
  IFS=',' read -ra entries <<< "$csv"
  # Trim whitespace
  for i in "${!entries[@]}"; do
    entries[$i]="$(echo "${entries[$i]}" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  done

  # For each entry, compute the transitive closure of every *other* entry
  # and check if this entry appears in it.
  for i in "${!entries[@]}"; do
    local target="${entries[$i]}"

    for j in "${!entries[@]}"; do
      [ "$i" = "$j" ] && continue
      local other="${entries[$j]}"

      # Resolve transitive deps of the other entry
      local other_deps
      other_deps="$(lib_resolve_uses_rules "$other")"
      [ -z "$other_deps" ] && continue

      # Check if target is in other's transitive closure
      IFS=',' read -ra dep_arr <<< "$other_deps"
      for d in "${dep_arr[@]}"; do
        d="$(echo "$d" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        if [ "$d" = "$target" ]; then
          echo "${target}|${other}"
          break 2  # found one cover, move to next target
        fi
      done
    done
  done
}
