package frontmatter

import (
	"regexp"
	"strings"
)

// Value represents a frontmatter value: either scalar string or string list.
type Value struct {
	Scalar string
	List   []string
	IsList bool
}

// Frontmatter maps field names to their values.
type Frontmatter map[string]Value

// StringVal returns the scalar value or defaultVal if missing/list/empty.
func (fm Frontmatter) StringVal(key, defaultVal string) string {
	v, ok := fm[key]
	if !ok || v.IsList {
		return defaultVal
	}
	if v.Scalar == "" {
		return defaultVal
	}
	return v.Scalar
}

// ListVal returns the list value, or nil if missing.
// A non-empty scalar is coerced to a single-element list.
func (fm Frontmatter) ListVal(key string) []string {
	v, ok := fm[key]
	if !ok {
		return nil
	}
	if v.IsList {
		return v.List
	}
	s := strings.TrimSpace(v.Scalar)
	if s == "" {
		return nil
	}
	return []string{s}
}

var fieldRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)`)
var blockItemRe = regexp.MustCompile(`^\s+-\s+(.*)`)

// ParseFrontmatter parses YAML frontmatter between --- markers.
func ParseFrontmatter(content string) Frontmatter {
	fmLines := extractFrontmatterLines(content)
	if len(fmLines) == 0 {
		return Frontmatter{}
	}
	return parseFMLines(fmLines)
}

// ExtractH1 returns the first # Heading line after frontmatter.
func ExtractH1(content string) string {
	pastFrontmatter := false
	fmCount := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimRight(line, " \t\r") == "---" {
			fmCount++
			if fmCount >= 2 {
				pastFrontmatter = true
			}
			continue
		}
		if !pastFrontmatter && fmCount == 0 {
			pastFrontmatter = true
		}
		if pastFrontmatter && strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

func extractFrontmatterLines(content string) []string {
	lines := strings.Split(content, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t\r") == "---" {
			if start == -1 {
				start = i + 1
			} else {
				return lines[start:i]
			}
		}
	}
	return nil
}

func parseFMLines(lines []string) Frontmatter {
	result := Frontmatter{}
	i := 0
	for i < len(lines) {
		line := lines[i]
		m := fieldRe.FindStringSubmatch(line)
		if m == nil {
			i++
			continue
		}

		key := m[1]
		rawValue := strings.TrimSpace(m[2])

		if strings.HasPrefix(rawValue, "[") {
			result[key] = Value{List: parseInlineList(rawValue), IsList: true}
			i++
			continue
		}

		if rawValue == "" {
			blockItems, consumed := parseBlockList(lines, i+1)
			if blockItems != nil {
				result[key] = Value{List: blockItems, IsList: true}
				i += 1 + consumed
				continue
			}
			result[key] = Value{Scalar: ""}
			i++
			continue
		}

		result[key] = Value{Scalar: stripQuotes(rawValue)}
		i++
	}
	return result
}

func parseInlineList(raw string) []string {
	inner := strings.Trim(raw, "[]")
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return []string{}
	}
	parts := strings.Split(inner, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, stripQuotes(p))
		}
	}
	return result
}

func parseBlockList(lines []string, start int) ([]string, int) {
	var items []string
	consumed := 0
	for i := start; i < len(lines); i++ {
		m := blockItemRe.FindStringSubmatch(lines[i])
		if m != nil {
			items = append(items, stripQuotes(strings.TrimSpace(m[1])))
			consumed++
		} else {
			break
		}
	}
	if consumed == 0 {
		return nil, 0
	}
	return items, consumed
}

func stripQuotes(val string) string {
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
			return val[1 : len(val)-1]
		}
	}
	return val
}
