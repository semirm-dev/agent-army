package model

// ValidationError for a missing dependency reference.
type ValidationError struct {
	FileLabel string
	Field     string
	Ref       string
	Severity  string // "error" or "warning"
}

// Fix proposes a change to a frontmatter field.
type Fix struct {
	Label    string
	Field    string
	FilePath string
	Before   []string
	After    []string
	Reasons  []string
}

// Redundancy records that Target is covered by CoveredBy transitively.
type Redundancy struct {
	Target    string
	CoveredBy string
}
