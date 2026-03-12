package resolver

import (
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestFormatReport_AllValid(t *testing.T) {
	got := FormatReport(nil, nil)
	want := "All dependency references are valid. No redundancies found."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatReport_WithErrors(t *testing.T) {
	errs := []model.ValidationError{
		{FileLabel: "spec/skills/api.md", Field: "uses_skills", Ref: "missing", Severity: "error"},
	}
	got := FormatReport(errs, nil)

	if !strings.Contains(got, "[ERROR] spec/skills/api.md") {
		t.Error("missing error line")
	}
	if !strings.Contains(got, `"missing" not found in spec/skills/`) {
		t.Error("missing ref detail")
	}
	if !strings.Contains(got, "1 error(s)") {
		t.Error("missing summary")
	}
}

func TestFormatReport_WithFixes(t *testing.T) {
	fixes := []model.Fix{
		{
			Label:    "spec/skills/c.md",
			Field:    "uses_skills",
			FilePath: "spec/skills/c.md",
			Before:   []string{"A", "B"},
			After:    []string{"A"},
			Reasons:  []string{`"B" covered by "A"`},
		},
	}
	got := FormatReport(nil, fixes)

	if !strings.Contains(got, "[FIX] spec/skills/c.md") {
		t.Error("missing fix line")
	}
	if !strings.Contains(got, "Before: [A, B]") {
		t.Error("missing before")
	}
	if !strings.Contains(got, "After:  [A]") {
		t.Error("missing after")
	}
}
