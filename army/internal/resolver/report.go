package resolver

import (
	"fmt"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

var fieldLocation = map[string]string{
	"uses_rules":   "spec/rules/",
	"uses_skills":  "spec/skills/",
	"uses_plugins": "config.json public_plugins",
	"delegates_to": "spec/agents/",
}

// FormatReport formats a human-readable validation report.
func FormatReport(errors []model.ValidationError, fixes []model.Fix) string {
	var realErrors, warnings []model.ValidationError
	for _, e := range errors {
		if e.Severity == "error" {
			realErrors = append(realErrors, e)
		} else {
			warnings = append(warnings, e)
		}
	}

	if len(realErrors) == 0 && len(warnings) == 0 && len(fixes) == 0 {
		return "All dependency references are valid. No redundancies found."
	}

	var lines []string
	lines = append(lines, "=== Dependency Validation Report ===", "")

	if len(realErrors) > 0 {
		lines = append(lines, "--- Errors (must fix manually) ---", "")
		for _, err := range realErrors {
			loc := fieldLocation[err.Field]
			if loc == "" {
				loc = "unknown"
			}
			lines = append(lines, fmt.Sprintf("  [ERROR] %s", err.FileLabel))
			lines = append(lines, fmt.Sprintf("    %s: %q not found in %s", err.Field, err.Ref, loc))
		}
		lines = append(lines, "")
	}

	if len(warnings) > 0 {
		lines = append(lines, "--- Warnings ---", "")
		for _, w := range warnings {
			loc := fieldLocation[w.Field]
			if loc == "" {
				loc = "unknown"
			}
			lines = append(lines, fmt.Sprintf("  [WARN] %s", w.FileLabel))
			lines = append(lines, fmt.Sprintf("    %s: %q not found in %s", w.Field, w.Ref, loc))
		}
		lines = append(lines, "")
	}

	if len(fixes) > 0 {
		lines = append(lines, "--- Redundancies (auto-fixable) ---", "")
		for _, fix := range fixes {
			lines = append(lines, fmt.Sprintf("  [FIX] %s", fix.Label))
			for _, reason := range fix.Reasons {
				lines = append(lines, fmt.Sprintf("    %s: %s", fix.Field, reason))
			}
			lines = append(lines, fmt.Sprintf("    Before: [%s]", strings.Join(fix.Before, ", ")))
			lines = append(lines, fmt.Sprintf("    After:  [%s]", strings.Join(fix.After, ", ")))
			lines = append(lines, "")
		}
	}

	lines = append(lines, fmt.Sprintf("Summary: %d error(s), %d warning(s), %d fixable redundanc(ies) across files.",
		len(realErrors), len(warnings), len(fixes)))
	lines = append(lines, "")

	if len(realErrors) > 0 {
		lines = append(lines, "Fix errors above before auto-fixing redundancies.")
	}

	return strings.Join(lines, "\n")
}
