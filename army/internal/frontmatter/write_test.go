package frontmatter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteField_Replace(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.md")

	original := "---\nname: foo\nuses_rules: [old1, old2]\n---\n\n# Body\n"
	if err := os.WriteFile(fp, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteField(fp, "uses_rules", []string{"new1", "new2"}); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	got := string(content)

	want := "---\nname: foo\nuses_rules: [new1, new2]\n---\n\n# Body\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteField_Insert(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.md")

	original := "---\nname: foo\n---\n\n# Body\n"
	if err := os.WriteFile(fp, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteField(fp, "uses_rules", []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	got := string(content)

	want := "---\nname: foo\nuses_rules: [a, b]\n---\n\n# Body\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteField_Empty(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.md")

	original := "---\nname: foo\nuses_rules: [old]\n---\n"
	if err := os.WriteFile(fp, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := WriteField(fp, "uses_rules", nil); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(fp)
	got := string(content)

	want := "---\nname: foo\nuses_rules: []\n---\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatFieldLine(t *testing.T) {
	tests := []struct {
		field  string
		values []string
		want   string
	}{
		{"f", nil, "f: []"},
		{"f", []string{}, "f: []"},
		{"f", []string{"a"}, "f: [a]"},
		{"f", []string{"a", "b", "c"}, "f: [a, b, c]"},
	}

	for _, tt := range tests {
		got := FormatFieldLine(tt.field, tt.values)
		if got != tt.want {
			t.Errorf("FormatFieldLine(%q, %v) = %q, want %q", tt.field, tt.values, got, tt.want)
		}
	}
}
