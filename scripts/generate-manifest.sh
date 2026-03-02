#!/usr/bin/env bash
# generate-manifest.sh — Generate manifest.json from rules/ and skills/ frontmatter.
# Idempotent: safe to re-run. Overwrites manifest.json each time.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"

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

# --- Generate manifest.json ---

{
  echo '{'
  echo '  "rules": ['

  last_rule=$(( ${#rule_names[@]} - 1 ))
  for i in "${!rule_names[@]}"; do
    comma=","
    [ "$i" -eq "$last_rule" ] && comma=""

    langs="${rule_languages[$i]}"
    uses_rules="${rule_uses_rules[$i]}"

    # Build optional JSON fields
    optional=""

    if [ "${rule_scopes[$i]}" = "language-specific" ] && [ -n "$langs" ]; then
      lang_json="["
      first=1
      IFS=',' read -ra lang_arr <<< "$langs"
      for l in "${lang_arr[@]}"; do
        l="$(echo "$l" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        [ -z "$l" ] && continue
        [ "$first" -eq 0 ] && lang_json+=", "
        lang_json+="\"$(json_escape "$l")\""
        first=0
      done
      lang_json+="]"
      optional+=", \"languages\": ${lang_json}"
    fi

    if [ -n "$uses_rules" ]; then
      ur_json="["
      first=1
      IFS=',' read -ra ur_arr <<< "$uses_rules"
      for u in "${ur_arr[@]}"; do
        u="$(echo "$u" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        [ -z "$u" ] && continue
        [ "$first" -eq 0 ] && ur_json+=", "
        ur_json+="\"$(json_escape "$u")\""
        first=0
      done
      ur_json+="]"
      optional+=", \"uses_rules\": ${ur_json}"
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
      lang_json="["
      first=1
      IFS=',' read -ra lang_arr <<< "$langs"
      for l in "${lang_arr[@]}"; do
        l="$(echo "$l" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        [ -z "$l" ] && continue
        [ "$first" -eq 0 ] && lang_json+=", "
        lang_json+="\"$(json_escape "$l")\""
        first=0
      done
      lang_json+="]"
      optional+=", \"languages\": ${lang_json}"
    fi

    uses_json="["
    first=1
    if [ -n "$uses" ]; then
      IFS=',' read -ra uses_arr <<< "$uses"
      for u in "${uses_arr[@]}"; do
        u="$(echo "$u" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
        [ -z "$u" ] && continue
        [ "$first" -eq 0 ] && uses_json+=", "
        uses_json+="\"$(json_escape "$u")\""
        first=0
      done
    fi
    uses_json+="]"

    echo "    { \"name\": \"$(json_escape "${skill_names[$i]}")\", \"scope\": \"$(json_escape "${skill_scopes[$i]}")\"${optional}, \"uses_rules\": ${uses_json}, \"path\": \"$(json_escape "${skill_paths[$i]}")\" }${comma}"
  done

  echo '  ]'
  echo '}'
} > "$REPO_ROOT/manifest.json"

echo "Generated $REPO_ROOT/manifest.json"
echo "Done."
