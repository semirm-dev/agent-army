package bootstrap

import (
	"sort"

	"github.com/semir/agent-army/internal/model"
)

// CursorAssignment holds a number and short name for a Cursor rule file.
type CursorAssignment struct {
	Number    int
	ShortName string
}

var cursorCategories = map[string]int{
	"language":       100,
	"git":            300,
	"api-db":         400,
	"infrastructure": 500,
}

var categoryPatterns = []struct {
	prefix   string
	category string
}{
	{"go/", "language"},
	{"python/", "language"},
	{"typescript/", "language"},
	{"react/", "language"},
	{"git", "git"},
	{"api-design", "api-db"},
	{"database", "api-db"},
}

func assignCursorNumbers(rules []model.Rule) []CursorAssignment {
	type indexed struct {
		origIdx int
		rule    model.Rule
	}

	categorized := make(map[string][]indexed)
	for i, r := range rules {
		cat := categorizeRule(r)
		categorized[cat] = append(categorized[cat], indexed{i, r})
	}

	type result struct {
		origIdx    int
		assignment CursorAssignment
	}
	var results []result

	catOrder := []string{"language", "git", "api-db", "infrastructure"}
	for _, cat := range catOrder {
		start := cursorCategories[cat]
		entries := categorized[cat]
		for offset, entry := range entries {
			results = append(results, result{
				origIdx: entry.origIdx,
				assignment: CursorAssignment{
					Number:    start + offset,
					ShortName: cursorShortName(entry.rule),
				},
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].origIdx < results[j].origIdx
	})

	assignments := make([]CursorAssignment, len(results))
	for i, r := range results {
		assignments[i] = r.assignment
	}
	return assignments
}

func categorizeRule(rule model.Rule) string {
	for _, p := range categoryPatterns {
		if len(rule.Name) >= len(p.prefix) && rule.Name[:len(p.prefix)] == p.prefix {
			return p.category
		}
		trimmed := p.prefix
		if trimmed[len(trimmed)-1] == '/' {
			trimmed = trimmed[:len(trimmed)-1]
		}
		if rule.Name == trimmed {
			return p.category
		}
	}
	return "infrastructure"
}

func cursorShortName(rule model.Rule) string {
	if name, ok := cursorLangNames[rule.Name]; ok {
		return name
	}
	return flattenName(rule.Name)
}
