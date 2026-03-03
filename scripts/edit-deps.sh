#!/usr/bin/env bash
# TODO: AI_DELETION_REVIEW — Replaced by src/agent_army/editor.py + cli.py
# edit-deps.sh — Interactive CLI for adding/removing dependency entries.
# Supports uses_rules, uses_skills, uses_plugins, delegates_to across
# rules, skills, and agents. Rewrites YAML frontmatter in-place, then
# auto-regenerates manifest.json.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"
AGENTS_DIR="$REPO_ROOT/agents"

# Source shared dependency library
source "$REPO_ROOT/scripts/lib-deps.sh"

# --- Helpers ---

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

# Load valid values for a given field into VALID_VALUES array.
load_valid_values() {
  local field="$1"
  VALID_VALUES=()

  case "$field" in
    uses_rules)
      load_rule_deps
      VALID_VALUES=("${_LIB_RULE_NAMES[@]}")
      ;;
    uses_skills)
      load_skill_names
      VALID_VALUES=("${_LIB_SKILL_NAMES[@]}")
      ;;
    uses_plugins)
      load_known_plugins
      VALID_VALUES=("${_LIB_KNOWN_PLUGINS[@]}")
      ;;
    delegates_to)
      load_agent_data
      VALID_VALUES=("${_LIB_AGENT_NAMES[@]}")
      ;;
  esac
}

# Run redundancy check for transitive fields.
# Only uses_rules and delegates_to have transitive resolution.
check_redundancy() {
  local field="$1" csv="$2"
  [ -z "$csv" ] && return

  case "$field" in
    uses_rules)
      load_rule_deps
      local redundancies
      redundancies="$(find_redundant_rules "$csv")"
      if [ -n "$redundancies" ]; then
        echo ""
        echo "Warning: Redundant entries detected (rule-to-rule):"
        while IFS='|' read -r redundant covered_by; do
          echo "  - \"$redundant\" is already included transitively by \"$covered_by\""
        done <<< "$redundancies"
      fi

      # Skill-transitive check: only for agents
      if [ "$ENTITY_TYPE" = "agent" ]; then
        load_skill_names
        local agent_fm
        agent_fm="$(get_frontmatter < "$CHOSEN_FILE")"
        local agent_skills
        agent_skills="$(extract_fm_list "uses_skills" "$agent_fm" | paste -sd ',' - || true)"
        if [ -n "$agent_skills" ]; then
          local skill_redundancies
          skill_redundancies="$(find_rules_redundant_via_skills "$csv" "$agent_skills")"
          if [ -n "$skill_redundancies" ]; then
            echo ""
            echo "Warning: Redundant entries detected (covered via skills):"
            while IFS='|' read -r redundant covered_by; do
              echo "  - \"$redundant\" is already included transitively by $covered_by"
            done <<< "$skill_redundancies"
          fi
        fi
      fi
      ;;
    delegates_to)
      load_agent_data
      local redundancies
      redundancies="$(find_redundant_delegates "$csv")"
      if [ -n "$redundancies" ]; then
        echo ""
        echo "Warning: Redundant entries detected:"
        while IFS='|' read -r redundant covered_by; do
          echo "  - \"$redundant\" is already included transitively by \"$covered_by\""
        done <<< "$redundancies"
      fi
      ;;
  esac
}

# Write the new value to the file using the appropriate writer.
write_field() {
  local field="$1" file="$2" csv="$3"

  case "$field" in
    uses_rules)    write_uses_rules "$file" "$csv" ;;
    uses_skills)   write_uses_skills "$file" "$csv" ;;
    uses_plugins)  write_uses_plugins "$file" "$csv" ;;
    delegates_to)  write_delegates_to "$file" "$csv" ;;
  esac
}

# --- Main flow ---

echo "=== Edit Dependencies ==="

# 1. Choose entity type
select_one "Choose entity type:" "rule" "skill" "agent"
ENTITY_TYPE="$REPLY"

case "$ENTITY_TYPE" in
  rule)  SEARCH_DIR="$RULES_DIR";  PREFIX="rules/"  ;;
  skill) SEARCH_DIR="$SKILLS_DIR"; PREFIX="skills/" ;;
  agent) SEARCH_DIR="$AGENTS_DIR"; PREFIX="agents/" ;;
esac

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

# 3. Choose field — auto-select if only one option
case "$ENTITY_TYPE" in
  rule|skill)
    FIELD="uses_rules"
    echo ""
    echo "Field: uses_rules (auto-selected)"
    ;;
  agent)
    select_one "Choose field:" "uses_rules" "uses_skills" "uses_plugins" "delegates_to"
    FIELD="$REPLY"
    ;;
esac

# 4. Show current values
echo ""
echo "File: ${PREFIX}${CHOSEN_DISPLAY}"
fm="$(get_frontmatter < "$CHOSEN_FILE")"
current=()
while IFS= read -r _v; do
  [ -n "$_v" ] && current+=("$_v")
done < <(extract_fm_list "$FIELD" "$fm")

if [ "${#current[@]}" -eq 0 ]; then
  echo "Current ${FIELD}: (none)"
else
  echo "Current ${FIELD}: [$(printf '%s, ' "${current[@]}" | sed 's/, $//')]"
fi

# 5. Choose action
select_one "Action:" "add" "remove"
ACTION="$REPLY"

if [ "$ACTION" = "add" ]; then
  # Load valid values for this field
  load_valid_values "$FIELD"

  # Filter out already-present values
  available=()
  for v in "${VALID_VALUES[@]}"; do
    local_match=0
    for c in ${current[@]+"${current[@]}"}; do
      if [ "$v" = "$c" ]; then
        local_match=1
        break
      fi
    done
    [ "$local_match" -eq 0 ] && available+=("$v")
  done

  if [ "${#available[@]}" -eq 0 ]; then
    echo "All values are already present. Nothing to add."
    exit 0
  fi

  select_multi "Select entries to add" "${available[@]}"
  new_values=(${current[@]+"${current[@]}"} "${MULTI_REPLY[@]}")

elif [ "$ACTION" = "remove" ]; then
  if [ "${#current[@]}" -eq 0 ]; then
    echo "No ${FIELD} entries to remove."
    exit 0
  fi

  select_multi "Select entries to remove" "${current[@]}"

  # Compute new list (current minus selected)
  new_values=()
  for c in "${current[@]}"; do
    keep=1
    for r in "${MULTI_REPLY[@]}"; do
      if [ "$c" = "$r" ]; then
        keep=0
        break
      fi
    done
    [ "$keep" -eq 1 ] && new_values+=("$c")
  done
fi

# Build comma-separated string
new_csv=""
for v in ${new_values[@]+"${new_values[@]}"}; do
  if [ -z "$new_csv" ]; then
    new_csv="$v"
  else
    new_csv="$new_csv, $v"
  fi
done

# 6. Redundancy check (only for transitive fields)
check_redundancy "$FIELD" "$new_csv"

# 7. Show diff preview
echo ""
echo "--- Change Preview ---"
if [ "${#current[@]}" -eq 0 ]; then
  echo "  Before: (none)"
else
  echo "  Before: [$(printf '%s, ' "${current[@]}" | sed 's/, $//')]"
fi
if [ -z "$new_csv" ]; then
  echo "  After:  (none — field will be cleared)"
else
  echo "  After:  [$new_csv]"
fi
echo ""

# 8. Confirm
printf "Apply this change? [y/N] "
read -r confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

# 9. Write file
write_field "$FIELD" "$CHOSEN_FILE" "$new_csv"
echo "Updated ${PREFIX}${CHOSEN_DISPLAY}"

# 10. Regenerate manifest
echo ""
echo "Regenerating manifest.json..."
bash "$REPO_ROOT/scripts/generate-manifest.sh"
