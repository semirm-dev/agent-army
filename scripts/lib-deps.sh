#!/usr/bin/env bash
# lib-deps.sh — Shared dependency library for dependency resolution.
# Source this file; do not execute directly.
#
# Provides:
#   get_frontmatter()          — extract YAML frontmatter from stdin
#   extract_fm_value()         — extract scalar value from frontmatter
#   extract_fm_list()          — extract list value from frontmatter
#   write_uses_rules()         — rewrite uses_rules line in a file
#   write_delegates_to()       — rewrite delegates_to line in a file
#   load_rule_deps()           — scan rules/*.md, populate _LIB_RULE_NAMES / _LIB_RULE_USES_RULES
#   load_skill_names()         — scan skills/*.md, populate _LIB_SKILL_NAMES
#   load_agent_data()          — scan agents/*.md, populate _LIB_AGENT_NAMES / _LIB_AGENT_DELEGATES_TO
#   load_known_plugins()       — parse config.json, populate _LIB_KNOWN_PLUGINS
#   _lib_rule_deps()           — lookup deps for a rule name
#   _lib_agent_delegates()     — lookup delegates for an agent name
#   lib_resolve_uses_rules()   — BFS transitive resolution for uses_rules
#   lib_resolve_delegates_to() — BFS transitive resolution for delegates_to
#   find_redundant_rules()     — detect uses_rules entries covered transitively
#   find_redundant_delegates() — detect delegates_to entries covered transitively
#   validate_refs_exist()      — check that refs exist in a valid-names list

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

# Rewrite delegates_to line in a file.
# Usage: write_delegates_to <file> <comma-separated-agents>
# If agents is empty, sets delegates_to: [].
write_delegates_to() {
  local file="$1" agents="$2"

  local new_val
  if [ -z "$agents" ]; then
    new_val="delegates_to: []"
  else
    new_val="delegates_to: [$agents]"
  fi

  local has_line
  has_line=$(awk '/^---$/{c++;next} c==1 && /^delegates_to:/{print "yes";exit}' "$file")

  if [ "$has_line" = "yes" ]; then
    awk -v newval="$new_val" '
      /^---$/ { c++; print; next }
      c == 1 && /^delegates_to:/ { print newval; next }
      { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
  else
    awk -v newval="$new_val" '
      /^---$/ { c++ }
      c == 2 { print newval; c = 3 }
      { print }
    ' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
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

# --- Skill name loading ---

declare -a _LIB_SKILL_NAMES=()

# Scan skills/**/*.md and populate _LIB_SKILL_NAMES array.
load_skill_names() {
  _LIB_SKILL_NAMES=()

  local skills_dir="$_LIB_REPO_ROOT/skills"

  while IFS= read -r file; do
    local relpath="${file#"$skills_dir/"}"
    local name="${relpath%.md}"
    _LIB_SKILL_NAMES+=("$name")
  done < <(find "$skills_dir" -name '*.md' | sort)
}

# --- Agent data loading ---

declare -a _LIB_AGENT_NAMES=()
declare -a _LIB_AGENT_DELEGATES_TO=()

# Scan agents/**/*.md and populate _LIB_AGENT_NAMES / _LIB_AGENT_DELEGATES_TO arrays.
load_agent_data() {
  _LIB_AGENT_NAMES=()
  _LIB_AGENT_DELEGATES_TO=()

  local agents_dir="$_LIB_REPO_ROOT/agents"

  while IFS= read -r file; do
    local relpath="${file#"$agents_dir/"}"
    local name="${relpath%.md}"
    local fm
    fm="$(get_frontmatter < "$file")"
    local delegates
    delegates="$(extract_fm_list "delegates_to" "$fm" | paste -sd ',' - || true)"

    _LIB_AGENT_NAMES+=("$name")
    _LIB_AGENT_DELEGATES_TO+=("$delegates")
  done < <(find "$agents_dir" -name '*.md' | sort)
}

# Lookup delegates for an agent name from loaded arrays.
_lib_agent_delegates() {
  local name="$1"
  for j in "${!_LIB_AGENT_NAMES[@]}"; do
    if [ "${_LIB_AGENT_NAMES[$j]}" = "$name" ]; then
      printf '%s' "${_LIB_AGENT_DELEGATES_TO[$j]}"
      return
    fi
  done
}

# --- Known plugins loading ---

declare -a _LIB_KNOWN_PLUGINS=()

# Parse config.json public_plugins[].name via python3, populate _LIB_KNOWN_PLUGINS.
load_known_plugins() {
  _LIB_KNOWN_PLUGINS=()

  local config_file="$_LIB_REPO_ROOT/config.json"
  [ -f "$config_file" ] || return

  local names
  names="$(python3 -c "
import json, sys
with open('$config_file') as f:
    cfg = json.load(f)
for p in cfg.get('public_plugins', []):
    print(p['name'])
" 2>/dev/null || true)"

  while IFS= read -r name; do
    [ -z "$name" ] && continue
    _LIB_KNOWN_PLUGINS+=("$name")
  done <<< "$names"
}

# --- delegates_to transitive resolution ---

# BFS transitive resolution of delegates_to.
# Takes a comma-separated list, returns deduplicated comma-separated list
# including all transitive delegations.
lib_resolve_delegates_to() {
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

    local deps
    deps="$(_lib_agent_delegates "$current")"
    if [ -n "$deps" ] && [ -n "$queue" ]; then
      queue="${queue},${deps}"
    elif [ -n "$deps" ]; then
      queue="$deps"
    fi
  done

  printf '%s' "$result"
}

# --- delegates_to redundancy detection ---

# For each agent A in a comma-separated list, check if any *other* agent O
# in the same list has A in its transitive closure.
# Output: one "redundant|covered_by" line per redundant entry.
find_redundant_delegates() {
  local csv="$1"
  [ -z "$csv" ] && return

  local -a entries=()
  IFS=',' read -ra entries <<< "$csv"
  for i in "${!entries[@]}"; do
    entries[$i]="$(echo "${entries[$i]}" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  done

  for i in "${!entries[@]}"; do
    local target="${entries[$i]}"

    for j in "${!entries[@]}"; do
      [ "$i" = "$j" ] && continue
      local other="${entries[$j]}"

      local other_deps
      other_deps="$(lib_resolve_delegates_to "$other")"
      [ -z "$other_deps" ] && continue

      IFS=',' read -ra dep_arr <<< "$other_deps"
      for d in "${dep_arr[@]}"; do
        d="$(echo "$d" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        if [ "$d" = "$target" ]; then
          echo "${target}|${other}"
          break 2
        fi
      done
    done
  done
}

# --- Reference validation ---

# Check that each ref in a comma-separated list exists in a valid-names array.
# Usage: validate_refs_exist "ref1,ref2" "valid1 valid2 valid3" "context_label"
# Prints one line per missing ref: "ref_name"
validate_refs_exist() {
  local refs_csv="$1" valid_names="$2"
  [ -z "$refs_csv" ] && return

  IFS=',' read -ra refs <<< "$refs_csv"
  for ref in "${refs[@]}"; do
    ref="$(echo "$ref" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    [ -z "$ref" ] && continue

    local found=0
    for valid in $valid_names; do
      if [ "$ref" = "$valid" ]; then
        found=1
        break
      fi
    done
    if [ "$found" -eq 0 ]; then
      echo "$ref"
    fi
  done
}
