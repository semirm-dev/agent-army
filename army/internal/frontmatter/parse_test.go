package frontmatter

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantKeys []string
	}{
		{"empty content", "", nil},
		{"no frontmatter", "# Hello\nworld", nil},
		{"scalar values", "---\nname: foo\nscope: universal\n---\n", []string{"name", "scope"}},
		{"inline list", "---\nlanguages: [go, python]\n---\n", []string{"languages"}},
		{"empty list", "---\nuses_rules: []\n---\n", []string{"uses_rules"}},
		{"block list", "---\nlanguages:\n  - go\n  - python\n---\n", []string{"languages"}},
		{"quoted value", "---\ndescription: \"has: colon\"\n---\n", []string{"description"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := ParseFrontmatter(tt.content)
			if tt.wantKeys == nil {
				if len(fm) != 0 {
					t.Errorf("expected empty frontmatter, got %d keys", len(fm))
				}
				return
			}
			for _, key := range tt.wantKeys {
				if _, ok := fm[key]; !ok {
					t.Errorf("missing key %q", key)
				}
			}
		})
	}
}

func TestParseFrontmatter_Values(t *testing.T) {
	content := "---\nname: foo\nscope: universal\nlanguages: [go, python]\nuses_rules: []\ndescription: \"has: colon\"\n---\n"
	fm := ParseFrontmatter(content)

	if got := fm.StringVal("name", ""); got != "foo" {
		t.Errorf("name = %q, want %q", got, "foo")
	}
	if got := fm.StringVal("scope", ""); got != "universal" {
		t.Errorf("scope = %q, want %q", got, "universal")
	}
	if got := fm.StringVal("description", ""); got != "has: colon" {
		t.Errorf("description = %q, want %q", got, "has: colon")
	}

	langs := fm.ListVal("languages")
	if len(langs) != 2 || langs[0] != "go" || langs[1] != "python" {
		t.Errorf("languages = %v, want [go python]", langs)
	}

	rules := fm.ListVal("uses_rules")
	if len(rules) != 0 {
		t.Errorf("uses_rules = %v, want empty", rules)
	}
}

func TestParseFrontmatter_BlockList(t *testing.T) {
	content := "---\nlanguages:\n  - go\n  - python\n  - typescript\n---\n"
	fm := ParseFrontmatter(content)

	langs := fm.ListVal("languages")
	if len(langs) != 3 {
		t.Fatalf("languages len = %d, want 3", len(langs))
	}
	expected := []string{"go", "python", "typescript"}
	for i, want := range expected {
		if langs[i] != want {
			t.Errorf("languages[%d] = %q, want %q", i, langs[i], want)
		}
	}
}

func TestParseFrontmatter_QuotedListItems(t *testing.T) {
	content := "---\nlanguages: ['go', \"python\"]\n---\n"
	fm := ParseFrontmatter(content)

	langs := fm.ListVal("languages")
	if len(langs) != 2 || langs[0] != "go" || langs[1] != "python" {
		t.Errorf("languages = %v, want [go python]", langs)
	}
}

func TestExtractH1(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"with frontmatter", "---\nname: foo\n---\n\n# My Title\n\nbody", "My Title"},
		{"without frontmatter", "# Direct Title\n\nbody", "Direct Title"},
		{"no h1", "---\nname: foo\n---\n\nno heading here", ""},
		{"multiple h1", "---\nname: foo\n---\n\n# First\n\n# Second", "First"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractH1(tt.content)
			if got != tt.want {
				t.Errorf("ExtractH1() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringVal(t *testing.T) {
	fm := Frontmatter{
		"name":  {Scalar: "foo"},
		"empty": {Scalar: ""},
		"list":  {List: []string{"a"}, IsList: true},
	}

	if got := fm.StringVal("name", "def"); got != "foo" {
		t.Errorf("name = %q, want foo", got)
	}
	if got := fm.StringVal("empty", "def"); got != "def" {
		t.Errorf("empty = %q, want def", got)
	}
	if got := fm.StringVal("list", "def"); got != "def" {
		t.Errorf("list = %q, want def", got)
	}
	if got := fm.StringVal("missing", "def"); got != "def" {
		t.Errorf("missing = %q, want def", got)
	}
}

func TestListVal(t *testing.T) {
	fm := Frontmatter{
		"langs":  {List: []string{"go", "py"}, IsList: true},
		"empty":  {List: []string{}, IsList: true},
		"scalar": {Scalar: "single"},
		"blank":  {Scalar: ""},
	}

	langs := fm.ListVal("langs")
	if len(langs) != 2 {
		t.Errorf("langs len = %d, want 2", len(langs))
	}

	empty := fm.ListVal("empty")
	if len(empty) != 0 {
		t.Errorf("empty len = %d, want 0", len(empty))
	}

	scalar := fm.ListVal("scalar")
	if len(scalar) != 1 || scalar[0] != "single" {
		t.Errorf("scalar = %v, want [single]", scalar)
	}

	blank := fm.ListVal("blank")
	if blank != nil {
		t.Errorf("blank = %v, want nil", blank)
	}

	missing := fm.ListVal("missing")
	if missing != nil {
		t.Errorf("missing = %v, want nil", missing)
	}
}
