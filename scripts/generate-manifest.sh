#!/usr/bin/env bash
# generate-manifest.sh — Generate manifest.json from rules/, skills/, and agents/ frontmatter.
# Idempotent: safe to re-run. Overwrites manifest.json each time.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"
AGENTS_DIR="$REPO_ROOT/agents"

# --- Helpers ---

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
  # Strip surrounding quotes
  val="${val#\"}" ; val="${val%\"}"
  val="${val#\'}" ; val="${val%\'}"
  printf '%s' "$val"
}

# Extract a YAML list (inline [a, b] or block - items) from frontmatter.
# Usage: extract_fm_list "key" "$frontmatter"
# Returns values one per line.
extract_fm_list() {
  local key="$1" fm="$2"

  # Check if the key exists at all
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

# Extract the first H1 heading from file content (after frontmatter).
# Usage: extract_h1 < file
extract_h1() {
  awk '
    /^---$/ { fm++; next }
    fm >= 2 && /^# / { sub(/^# /, ""); print; exit }
  '
}

# Join array items with a delimiter.
# Usage: join_by "," "${arr[@]}"
join_by() {
  local d="$1"; shift
  local first="$1"; shift || true
  printf '%s' "$first"
  for item in "$@"; do
    printf '%s%s' "$d" "$item"
  done
}

# Escape special JSON characters in a string.
json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

# --- Collect rules ---

declare -a rule_names=()
declare -a rule_scopes=()
declare -a rule_summaries=()
declare -a rule_paths=()
declare -a rule_languages=()
declare -a rule_uses_rules=()

while IFS= read -r file; do
  relpath="${file#"$RULES_DIR/"}"

  fm="$(get_frontmatter < "$file")"
  name="${relpath%.md}"
  scope="$(extract_fm_value "scope" "$fm")"
  summary="$(extract_h1 < "$file")"
  langs="$(extract_fm_list "languages" "$fm" | paste -sd ',' - || true)"
  uses_rules="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"

  [ -z "$scope" ] && scope="universal"

  rule_names+=("$name")
  rule_scopes+=("$scope")
  rule_summaries+=("$summary")
  rule_paths+=("rules/$relpath")
  rule_languages+=("$langs")
  rule_uses_rules+=("$uses_rules")
done < <(find "$RULES_DIR" -name '*.md' | sort)

# --- Transitive resolution helpers ---

# Look up uses_rules for a rule name by searching the parallel arrays.
_rule_deps() {
  local name="$1"
  for j in "${!rule_names[@]}"; do
    if [ "${rule_names[$j]}" = "$name" ]; then
      printf '%s' "${rule_uses_rules[$j]}"
      return
    fi
  done
}

# Transitively resolve uses_rules via BFS.
# Takes a comma-separated list of rule names, returns a deduplicated
# comma-separated list including all transitive dependencies.
resolve_uses_rules() {
  local input="$1"
  [ -z "$input" ] && return

  local visited=""
  local result=""
  local queue=""

  queue="$input"

  while [ -n "$queue" ]; do
    # Pop the first item from the queue
    local current
    current="$(echo "$queue" | cut -d',' -f1 | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    local rest
    rest="$(echo "$queue" | cut -d',' -f2- -s)"
    queue="$rest"

    [ -z "$current" ] && continue

    # Check if already visited (deduplicate)
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

    # Look up transitive deps for this rule
    local deps
    deps="$(_rule_deps "$current")"
    if [ -n "$deps" ] && [ -n "$queue" ]; then
      queue="${queue},${deps}"
    elif [ -n "$deps" ]; then
      queue="$deps"
    fi
  done

  printf '%s' "$result"
}

# Look up delegates_to for an agent name by searching the parallel arrays.
_agent_delegates() {
  local name="$1"
  for j in "${!agent_names[@]}"; do
    if [ "${agent_names[$j]}" = "$name" ]; then
      printf '%s' "${agent_delegates_to[$j]}"
      return
    fi
  done
}

# Transitively resolve delegates_to via BFS.
# Takes a comma-separated list of agent names, returns a deduplicated
# comma-separated list including all transitive delegations.
resolve_delegates_to() {
  local input="$1"
  [ -z "$input" ] && return

  local visited=""
  local result=""
  local queue=""

  queue="$input"

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
    deps="$(_agent_delegates "$current")"
    if [ -n "$deps" ] && [ -n "$queue" ]; then
      queue="${queue},${deps}"
    elif [ -n "$deps" ]; then
      queue="$deps"
    fi
  done

  printf '%s' "$result"
}

# Look up uses_rules for a skill name by searching the parallel arrays.
_skill_rules() {
  local name="$1"
  for j in "${!skill_names[@]}"; do
    if [ "${skill_names[$j]}" = "$name" ]; then
      printf '%s' "${skill_uses_rules[$j]}"
      return
    fi
  done
}

# --- Collect skills ---

declare -a skill_names=()
declare -a skill_summaries=()
declare -a skill_scopes=()
declare -a skill_languages=()
declare -a skill_uses_rules=()
declare -a skill_paths=()

while IFS= read -r file; do
  relpath="${file#"$SKILLS_DIR/"}"

  fm="$(get_frontmatter < "$file")"
  name="$(extract_fm_value "name" "$fm")"
  [ -z "$name" ] && name="${relpath%.md}"

  scope="$(extract_fm_value "scope" "$fm")"
  [ -z "$scope" ] && scope="universal"

  langs="$(extract_fm_list "languages" "$fm" | paste -sd ',' - || true)"
  uses="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"
  summary="$(extract_h1 < "$file")"

  skill_names+=("$name")
  skill_summaries+=("$summary")
  skill_scopes+=("$scope")
  skill_languages+=("$langs")
  skill_uses_rules+=("$uses")
  skill_paths+=("skills/$relpath")
done < <(find "$SKILLS_DIR" -name '*.md' | sort)

# --- Collect agents ---

declare -a agent_names=()
declare -a agent_roles=()
declare -a agent_scopes=()
declare -a agent_languages=()
declare -a agent_accesses=()
declare -a agent_uses_skills=()
declare -a agent_uses_rules=()
declare -a agent_uses_plugins=()
declare -a agent_delegates_to=()
declare -a agent_paths=()

while IFS= read -r file; do
  relpath="${file#"$AGENTS_DIR/"}"

  fm="$(get_frontmatter < "$file")"
  name="$(extract_fm_value "name" "$fm")"
  [ -z "$name" ] && name="${relpath%.md}"

  role="$(extract_fm_value "role" "$fm")"
  scope="$(extract_fm_value "scope" "$fm")"
  [ -z "$scope" ] && scope="universal"

  access="$(extract_fm_value "access" "$fm")"
  [ -z "$access" ] && access="read-write"

  langs="$(extract_fm_list "languages" "$fm" | paste -sd ',' - || true)"
  u_skills="$(extract_fm_list "uses_skills" "$fm" | paste -sd ',' - || true)"
  u_rules="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"
  u_plugins="$(extract_fm_list "uses_plugins" "$fm" | paste -sd ',' - || true)"
  delegates="$(extract_fm_list "delegates_to" "$fm" | paste -sd ',' - || true)"

  agent_names+=("$name")
  agent_roles+=("$role")
  agent_scopes+=("$scope")
  agent_languages+=("$langs")
  agent_accesses+=("$access")
  agent_uses_skills+=("$u_skills")
  agent_uses_rules+=("$u_rules")
  agent_uses_plugins+=("$u_plugins")
  agent_delegates_to+=("$delegates")
  agent_paths+=("agents/$relpath")
done < <(find "$AGENTS_DIR" -name '*.md' | sort)

# --- Helper: build a JSON array string from a comma-separated list ---
_csv_to_json_array() {
  local csv="$1"
  local json="["
  local first=1
  if [ -n "$csv" ]; then
    IFS=',' read -ra arr <<< "$csv"
    for item in "${arr[@]}"; do
      item="$(echo "$item" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      [ -z "$item" ] && continue
      [ "$first" -eq 0 ] && json+=", "
      json+="\"$(json_escape "$item")\""
      first=0
    done
  fi
  json+="]"
  printf '%s' "$json"
}

# --- Generate manifest.json ---

{
  echo '{'
  echo '  "rules": ['

  last_rule=$(( ${#rule_names[@]} - 1 ))
  for i in "${!rule_names[@]}"; do
    comma=","
    [ "$i" -eq "$last_rule" ] && comma=""

    langs="${rule_languages[$i]}"
    uses_rules="$(resolve_uses_rules "${rule_uses_rules[$i]}")"

    # Build optional JSON fields
    optional=""

    if [ "${rule_scopes[$i]}" = "language-specific" ] && [ -n "$langs" ]; then
      optional+=", \"languages\": $(_csv_to_json_array "$langs")"
    fi

    if [ -n "$uses_rules" ]; then
      optional+=", \"uses_rules\": $(_csv_to_json_array "$uses_rules")"
    fi

    echo "    { \"name\": \"$(json_escape "${rule_names[$i]}")\", \"scope\": \"$(json_escape "${rule_scopes[$i]}")\"${optional}, \"path\": \"$(json_escape "${rule_paths[$i]}")\" }${comma}"
  done

  echo '  ],'
  echo '  "skills": ['

  last_skill=$(( ${#skill_names[@]} - 1 ))
  for i in "${!skill_names[@]}"; do
    comma=","
    [ "$i" -eq "$last_skill" ] && comma=""

    langs="${skill_languages[$i]}"
    uses="$(resolve_uses_rules "${skill_uses_rules[$i]}")"

    # Build optional JSON fields
    optional=""

    if [ "${skill_scopes[$i]}" = "language-specific" ] && [ -n "$langs" ]; then
      optional+=", \"languages\": $(_csv_to_json_array "$langs")"
    fi

    echo "    { \"name\": \"$(json_escape "${skill_names[$i]}")\", \"scope\": \"$(json_escape "${skill_scopes[$i]}")\"${optional}, \"uses_rules\": $(_csv_to_json_array "$uses"), \"path\": \"$(json_escape "${skill_paths[$i]}")\" }${comma}"
  done

  echo '  ],'
  echo '  "agents": ['

  last_agent=$(( ${#agent_names[@]} - 1 ))
  for i in "${!agent_names[@]}"; do
    comma=","
    [ "$i" -eq "$last_agent" ] && comma=""

    langs="${agent_languages[$i]}"
    u_skills="${agent_uses_skills[$i]}"
    # Collect rules from agent's skills
    skill_rules_merged=""
    if [ -n "$u_skills" ]; then
      IFS=',' read -ra sk_arr <<< "$u_skills"
      for sk in "${sk_arr[@]}"; do
        sk="$(echo "$sk" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        sr="$(_skill_rules "$sk")"
        if [ -n "$sr" ]; then
          [ -n "$skill_rules_merged" ] && skill_rules_merged="${skill_rules_merged},${sr}" || skill_rules_merged="$sr"
        fi
      done
    fi

    # Merge agent's own rules with rules from skills, then resolve transitively
    combined_rules="${agent_uses_rules[$i]}"
    if [ -n "$skill_rules_merged" ]; then
      [ -n "$combined_rules" ] && combined_rules="${combined_rules},${skill_rules_merged}" || combined_rules="$skill_rules_merged"
    fi
    u_rules="$(resolve_uses_rules "$combined_rules")"
    u_plugins="${agent_uses_plugins[$i]}"
    delegates="$(resolve_delegates_to "${agent_delegates_to[$i]}")"

    # Build optional JSON fields
    optional=""

    if [ -n "$langs" ]; then
      optional+=", \"languages\": $(_csv_to_json_array "$langs")"
    fi

    echo "    { \"name\": \"$(json_escape "${agent_names[$i]}")\", \"role\": \"$(json_escape "${agent_roles[$i]}")\", \"scope\": \"$(json_escape "${agent_scopes[$i]}")\", \"access\": \"$(json_escape "${agent_accesses[$i]}")\"${optional}, \"uses_skills\": $(_csv_to_json_array "$u_skills"), \"uses_rules\": $(_csv_to_json_array "$u_rules"), \"uses_plugins\": $(_csv_to_json_array "$u_plugins"), \"delegates_to\": $(_csv_to_json_array "$delegates"), \"path\": \"$(json_escape "${agent_paths[$i]}")\" }${comma}"
  done

  echo '  ]'
  echo '}'
} > "$REPO_ROOT/manifest.json"

echo "Generated $REPO_ROOT/manifest.json"
echo "Done."
