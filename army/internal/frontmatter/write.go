package frontmatter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// WriteField replaces or inserts a frontmatter field with values.
// Format: field: [val1, val2] (or field: [] for empty).
// Writes atomically via .tmp file and os.Rename.
func WriteField(filePath, field string, values []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	newLine := FormatFieldLine(field, values)
	updated := replaceOrInsertField(string(content), field, newLine)

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("rename %s: %w", filePath, err)
	}
	return nil
}

// FormatFieldLine builds a frontmatter line like field: [val1, val2].
func FormatFieldLine(field string, values []string) string {
	if len(values) == 0 {
		return field + ": []"
	}
	return field + ": [" + strings.Join(values, ", ") + "]"
}

func replaceOrInsertField(content, field, newLine string) string {
	lines := splitKeepEnds(content)
	fmDashCount := 0
	fieldReplaced := false
	var result []string

	fieldPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(field) + `:`)

	for _, line := range lines {
		stripped := strings.TrimRight(line, "\n\r")

		if stripped == "---" {
			fmDashCount++
			if fmDashCount == 2 && !fieldReplaced {
				result = append(result, newLine+"\n")
				fieldReplaced = true
			}
			result = append(result, line)
			continue
		}

		if fmDashCount == 1 && !fieldReplaced {
			if fieldPattern.MatchString(stripped) {
				result = append(result, newLine+"\n")
				fieldReplaced = true
				continue
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "")
}

func splitKeepEnds(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i+1])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
