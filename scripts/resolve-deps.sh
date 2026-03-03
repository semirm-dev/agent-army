#!/usr/bin/env bash
# TODO: AI_DELETION_REVIEW — Replaced by src/agent_army/resolver.py + cli.py
# resolve-deps.sh — Validate all dependency references and remove redundancies.
# Checks uses_rules, uses_skills, uses_plugins, delegates_to across rules/, skills/, agents/.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RULES_DIR="$REPO_ROOT/rules"
SKILLS_DIR="$REPO_ROOT/skills"
AGENTS_DIR="$REPO_ROOT/agents"

source "$REPO_ROOT/scripts/lib-deps.sh"

# =====================================================================
# Phase 1: Load all data
# =====================================================================

load_rule_deps
load_skill_names
load_agent_data
load_known_plugins

# Build space-separated valid-name strings for validate_refs_exist
rule_names_str="${_LIB_RULE_NAMES[*]}"
skill_names_str="${_LIB_SKILL_NAMES[*]}"
agent_names_str="${_LIB_AGENT_NAMES[*]}"
plugin_names_str="${_LIB_KNOWN_PLUGINS[*]}"

# =====================================================================
# Phase 2: Validate existence of all references
# =====================================================================

declare -a errors=()
declare -a warnings=()

check_uses_rules_exist() {
  local file="$1" label="$2"
  local fm
  fm="$(get_frontmatter < "$file")"
  local uses
  uses="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"
  [ -z "$uses" ] && return

  local missing
  missing="$(validate_refs_exist "$uses" "$rule_names_str")"
  while IFS= read -r ref; do
    [ -z "$ref" ] && continue
    errors+=("  [ERROR] $label")
    errors+=("    uses_rules: \"$ref\" not found in rules/")
  done <<< "$missing"
}

check_uses_skills_exist() {
  local file="$1" label="$2"
  local fm
  fm="$(get_frontmatter < "$file")"
  local uses
  uses="$(extract_fm_list "uses_skills" "$fm" | paste -sd ',' - || true)"
  [ -z "$uses" ] && return

  local missing
  missing="$(validate_refs_exist "$uses" "$skill_names_str")"
  while IFS= read -r ref; do
    [ -z "$ref" ] && continue
    errors+=("  [ERROR] $label")
    errors+=("    uses_skills: \"$ref\" not found in skills/")
  done <<< "$missing"
}

check_uses_plugins_exist() {
  local file="$1" label="$2"
  local fm
  fm="$(get_frontmatter < "$file")"
  local uses
  uses="$(extract_fm_list "uses_plugins" "$fm" | paste -sd ',' - || true)"
  [ -z "$uses" ] && return

  local missing
  missing="$(validate_refs_exist "$uses" "$plugin_names_str")"
  while IFS= read -r ref; do
    [ -z "$ref" ] && continue
    warnings+=("  [WARN] $label")
    warnings+=("    uses_plugins: \"$ref\" not found in config.json public_plugins")
  done <<< "$missing"
}

check_delegates_to_exist() {
  local file="$1" label="$2"
  local fm
  fm="$(get_frontmatter < "$file")"
  local delegates
  delegates="$(extract_fm_list "delegates_to" "$fm" | paste -sd ',' - || true)"
  [ -z "$delegates" ] && return

  local missing
  missing="$(validate_refs_exist "$delegates" "$agent_names_str")"
  while IFS= read -r ref; do
    [ -z "$ref" ] && continue
    errors+=("  [ERROR] $label")
    errors+=("    delegates_to: \"$ref\" not found in agents/")
  done <<< "$missing"
}

# Validate rules/*.md — uses_rules
while IFS= read -r file; do
  relpath="${file#"$RULES_DIR/"}"
  check_uses_rules_exist "$file" "rules/$relpath"
done < <(find "$RULES_DIR" -name '*.md' | sort)

# Validate skills/*.md — uses_rules
while IFS= read -r file; do
  relpath="${file#"$SKILLS_DIR/"}"
  check_uses_rules_exist "$file" "skills/$relpath"
done < <(find "$SKILLS_DIR" -name '*.md' | sort)

# Validate agents/*.md — uses_rules, uses_skills, uses_plugins, delegates_to
while IFS= read -r file; do
  relpath="${file#"$AGENTS_DIR/"}"
  label="agents/$relpath"
  check_uses_rules_exist "$file" "$label"
  check_uses_skills_exist "$file" "$label"
  check_uses_plugins_exist "$file" "$label"
  check_delegates_to_exist "$file" "$label"
done < <(find "$AGENTS_DIR" -name '*.md' | sort)

# =====================================================================
# Phase 3: Detect redundancies
# =====================================================================

declare -a fix_labels=()
declare -a fix_field=()
declare -a fix_paths=()
declare -a fix_before=()
declare -a fix_after=()
declare -a fix_removed=()

scan_uses_rules_redundancies() {
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

    local -a redundant_names=()
    local -a reasons=()
    while IFS='|' read -r rname covered; do
      redundant_names+=("$rname")
      reasons+=("\"$rname\" covered by \"$covered\"")
    done <<< "$redundancies"

    IFS=',' read -ra orig_arr <<< "$uses"
    local -a cleaned=()
    for entry in "${orig_arr[@]}"; do
      entry="$(echo "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      local is_redundant=0
      for rn in "${redundant_names[@]}"; do
        [ "$entry" = "$rn" ] && is_redundant=1 && break
      done
      [ "$is_redundant" -eq 0 ] && cleaned+=("$entry")
    done

    local cleaned_csv=""
    for c in "${cleaned[@]}"; do
      if [ -z "$cleaned_csv" ]; then cleaned_csv="$c"; else cleaned_csv="$cleaned_csv, $c"; fi
    done

    fix_labels+=("${prefix}${relpath}")
    fix_field+=("uses_rules")
    fix_paths+=("$file")
    fix_before+=("$uses")
    fix_after+=("$cleaned_csv")
    fix_removed+=("$(printf '%s\n' "${reasons[@]}" | paste -sd ';' -)")
  done < <(find "$dir" -name '*.md' | sort)
}

scan_delegates_to_redundancies() {
  while IFS= read -r file; do
    local relpath="${file#"$AGENTS_DIR/"}"
    local fm
    fm="$(get_frontmatter < "$file")"
    local delegates
    delegates="$(extract_fm_list "delegates_to" "$fm" | paste -sd ',' - || true)"
    [ -z "$delegates" ] && continue

    local redundancies
    redundancies="$(find_redundant_delegates "$delegates")"
    [ -z "$redundancies" ] && continue

    local -a redundant_names=()
    local -a reasons=()
    while IFS='|' read -r rname covered; do
      redundant_names+=("$rname")
      reasons+=("\"$rname\" covered by \"$covered\"")
    done <<< "$redundancies"

    IFS=',' read -ra orig_arr <<< "$delegates"
    local -a cleaned=()
    for entry in "${orig_arr[@]}"; do
      entry="$(echo "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      local is_redundant=0
      for rn in "${redundant_names[@]}"; do
        [ "$entry" = "$rn" ] && is_redundant=1 && break
      done
      [ "$is_redundant" -eq 0 ] && cleaned+=("$entry")
    done

    local cleaned_csv=""
    for c in "${cleaned[@]}"; do
      if [ -z "$cleaned_csv" ]; then cleaned_csv="$c"; else cleaned_csv="$cleaned_csv, $c"; fi
    done

    fix_labels+=("agents/${relpath}")
    fix_field+=("delegates_to")
    fix_paths+=("$file")
    fix_before+=("$delegates")
    fix_after+=("$cleaned_csv")
    fix_removed+=("$(printf '%s\n' "${reasons[@]}" | paste -sd ';' -)")
  done < <(find "$AGENTS_DIR" -name '*.md' | sort)
}

scan_agent_skill_rule_redundancies() {
  while IFS= read -r file; do
    local relpath="${file#"$AGENTS_DIR/"}"
    local fm
    fm="$(get_frontmatter < "$file")"
    local uses_rules
    uses_rules="$(extract_fm_list "uses_rules" "$fm" | paste -sd ',' - || true)"
    [ -z "$uses_rules" ] && continue

    local uses_skills
    uses_skills="$(extract_fm_list "uses_skills" "$fm" | paste -sd ',' - || true)"
    [ -z "$uses_skills" ] && continue

    local redundancies
    redundancies="$(find_rules_redundant_via_skills "$uses_rules" "$uses_skills")"
    [ -z "$redundancies" ] && continue

    local -a redundant_names=()
    local -a reasons=()
    while IFS='|' read -r rname covered; do
      redundant_names+=("$rname")
      reasons+=("\"$rname\" covered by $covered")
    done <<< "$redundancies"

    IFS=',' read -ra orig_arr <<< "$uses_rules"
    local -a cleaned=()
    for entry in "${orig_arr[@]}"; do
      entry="$(echo "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
      local is_redundant=0
      for rn in "${redundant_names[@]}"; do
        [ "$entry" = "$rn" ] && is_redundant=1 && break
      done
      [ "$is_redundant" -eq 0 ] && cleaned+=("$entry")
    done

    local cleaned_csv=""
    for c in ${cleaned[@]+"${cleaned[@]}"}; do
      if [ -z "$cleaned_csv" ]; then cleaned_csv="$c"; else cleaned_csv="$cleaned_csv, $c"; fi
    done

    # Check if this file+field already has a fix entry from rule-to-rule scan
    local already_fixed=0
    for i in ${!fix_paths[@]+"${!fix_paths[@]}"}; do
      if [ "${fix_paths[$i]}" = "$file" ] && [ "${fix_field[$i]}" = "uses_rules" ]; then
        # Merge: apply skill-transitive removals on top of existing cleaned result
        local existing_after="${fix_after[$i]}"
        local merged_after=""
        IFS=',' read -ra ea_arr <<< "$existing_after"
        for entry in "${ea_arr[@]}"; do
          entry="$(echo "$entry" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
          local is_redundant=0
          for rn in "${redundant_names[@]}"; do
            [ "$entry" = "$rn" ] && is_redundant=1 && break
          done
          if [ "$is_redundant" -eq 0 ]; then
            if [ -z "$merged_after" ]; then merged_after="$entry"; else merged_after="$merged_after, $entry"; fi
          fi
        done
        fix_after[$i]="$merged_after"
        fix_removed[$i]="${fix_removed[$i]};$(printf '%s\n' "${reasons[@]}" | paste -sd ';' -)"
        already_fixed=1
        break
      fi
    done

    if [ "$already_fixed" -eq 0 ]; then
      fix_labels+=("agents/${relpath}")
      fix_field+=("uses_rules")
      fix_paths+=("$file")
      fix_before+=("$uses_rules")
      fix_after+=("$cleaned_csv")
      fix_removed+=("$(printf '%s\n' "${reasons[@]}" | paste -sd ';' -)")
    fi
  done < <(find "$AGENTS_DIR" -name '*.md' | sort)
}

scan_uses_rules_redundancies "$RULES_DIR" "rules/"
scan_uses_rules_redundancies "$SKILLS_DIR" "skills/"
scan_uses_rules_redundancies "$AGENTS_DIR" "agents/"
scan_agent_skill_rule_redundancies
scan_delegates_to_redundancies

# =====================================================================
# Phase 4: Report + auto-fix
# =====================================================================

error_count=${#errors[@]}
warning_count=${#warnings[@]}
fix_count=${#fix_labels[@]}

# Count unique error files (every 2 lines = 1 error)
error_file_count=$((error_count / 2))
warning_file_count=$((warning_count / 2))

if [ "$error_count" -eq 0 ] && [ "$warning_count" -eq 0 ] && [ "$fix_count" -eq 0 ]; then
  echo "All dependency references are valid. No redundancies found."
  exit 0
fi

echo "=== Dependency Validation Report ==="
echo ""

# Print errors
if [ "$error_count" -gt 0 ]; then
  echo "--- Errors (must fix manually) ---"
  echo ""
  for line in "${errors[@]}"; do
    echo "$line"
  done
  echo ""
fi

# Print warnings
if [ "$warning_count" -gt 0 ]; then
  echo "--- Warnings ---"
  echo ""
  for line in "${warnings[@]}"; do
    echo "$line"
  done
  echo ""
fi

# Print fixable redundancies
if [ "$fix_count" -gt 0 ]; then
  echo "--- Redundancies (auto-fixable) ---"
  echo ""
  for i in "${!fix_labels[@]}"; do
    echo "  [FIX] ${fix_labels[$i]}"
    IFS=';' read -ra reason_arr <<< "${fix_removed[$i]}"
    for reason in "${reason_arr[@]}"; do
      echo "    ${fix_field[$i]}: $reason"
    done
    echo "    Before: [${fix_before[$i]}]"
    echo "    After:  [${fix_after[$i]}]"
    echo ""
  done
fi

echo "Summary: $error_file_count error(s), $warning_file_count warning(s), $fix_count fixable redundanc(ies) across files."
echo ""

# Errors block auto-fix
if [ "$error_count" -gt 0 ]; then
  echo "Fix errors above before auto-fixing redundancies."
  exit 1
fi

# Nothing to fix
if [ "$fix_count" -eq 0 ]; then
  exit 0
fi

# Confirm auto-fix
printf "Remove redundant entries? [y/N] "
read -r confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Aborted. No files changed."
  exit 0
fi

# Apply fixes
echo ""
for i in "${!fix_paths[@]}"; do
  if [ "${fix_field[$i]}" = "uses_rules" ]; then
    write_uses_rules "${fix_paths[$i]}" "${fix_after[$i]}"
  elif [ "${fix_field[$i]}" = "delegates_to" ]; then
    write_delegates_to "${fix_paths[$i]}" "${fix_after[$i]}"
  fi
  echo "Updated ${fix_labels[$i]}"
done

# Regenerate manifest
echo ""
echo "Regenerating manifest.json..."
bash "$REPO_ROOT/scripts/generate-manifest.sh"
