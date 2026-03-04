package manifest

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OrderedMap preserves section order in the manifest.
type OrderedMap struct {
	Keys     []string
	Sections map[string][]Entry
}

// Entry preserves field order within a manifest entry.
type Entry struct {
	Keys   []string
	Values map[string]interface{}
}

// Add adds a string field.
func (e *Entry) Add(key, value string) {
	if e.Values == nil {
		e.Values = make(map[string]interface{})
	}
	e.Keys = append(e.Keys, key)
	e.Values[key] = value
}

// AddList adds a string list field.
func (e *Entry) AddList(key string, values []string) {
	if e.Values == nil {
		e.Values = make(map[string]interface{})
	}
	e.Keys = append(e.Keys, key)
	if values == nil {
		e.Values[key] = []string{}
	} else {
		e.Values[key] = values
	}
}

func formatValue(val interface{}) string {
	switch v := val.(type) {
	case []string:
		if len(v) == 0 {
			return "[]"
		}
		quoted := make([]string, len(v))
		for i, s := range v {
			quoted[i] = fmt.Sprintf("%q", s)
		}
		return "[" + strings.Join(quoted, ", ") + "]"
	default:
		encoded, _ := json.Marshal(val)
		return string(encoded)
	}
}

func formatEntry(e Entry) string {
	var pairs []string
	for _, key := range e.Keys {
		val := e.Values[key]
		pairs = append(pairs, fmt.Sprintf("%q: %s", key, formatValue(val)))
	}
	return "{ " + strings.Join(pairs, ", ") + " }"
}

func formatManifestJSON(m OrderedMap) string {
	var lines []string
	lines = append(lines, "{")

	for secIdx, section := range m.Keys {
		entries := m.Sections[section]
		isLast := secIdx == len(m.Keys)-1
		suffix := ""
		if !isLast {
			suffix = ","
		}

		lines = append(lines, fmt.Sprintf("  %q: [", section))
		for i, entry := range entries {
			comma := ""
			if i < len(entries)-1 {
				comma = ","
			}
			lines = append(lines, "    "+formatEntry(entry)+comma)
		}
		lines = append(lines, "  ]"+suffix)
	}

	lines = append(lines, "}")
	return strings.Join(lines, "\n") + "\n"
}
