package types

// CatalogPlugin represents a plugin in the catalog.
type CatalogPlugin struct {
	Name        string   `json:"name"`
	Marketplace string   `json:"marketplace"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// CatalogSkill represents a skill in the catalog.
type CatalogSkill struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// TechProfile maps a technology to its detection markers and recommended items.
type TechProfile struct {
	Detect  []string `json:"detect"`
	Plugins []string `json:"plugins"`
	Skills  []string `json:"skills"`
}

// Catalog is the full registry of available plugins, skills, and tech profiles.
type Catalog struct {
	Version      int                    `json:"version"`
	UpdatedAt    string                 `json:"updated_at"`
	Plugins      []CatalogPlugin        `json:"plugins"`
	Skills       []CatalogSkill         `json:"skills"`
	TechProfiles map[string]TechProfile `json:"tech_profiles"`
}

// ManifestPlugin represents a plugin in the user's manifest.
type ManifestPlugin struct {
	Name        string   `json:"name"`
	Marketplace string   `json:"marketplace"`
	Tags        []string `json:"tags"`
	Destination string   `json:"destination"` // "user" or "project"
}

// ManifestSkill represents a skill in the user's manifest.
type ManifestSkill struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Tags        []string `json:"tags"`
	Destination string   `json:"destination"` // "user" or "project"
}

// Manifest is the user's personal selection of plugins and skills.
type Manifest struct {
	Version int              `json:"version"`
	Plugins []ManifestPlugin `json:"plugins"`
	Skills  []ManifestSkill  `json:"skills"`
}

// InstalledPlugin represents a plugin found on the system.
type InstalledPlugin struct {
	Name        string
	Marketplace string
	Version     string
	Scope       string
}

// InstalledSkill represents a skill found on the system.
type InstalledSkill struct {
	Name      string
	Source    string
	SourceURL string
}

// DiffResult represents the comparison between manifest and installed state.
type DiffResult struct {
	MissingPlugins []ManifestPlugin  // In manifest but not installed
	ExtraPlugins   []InstalledPlugin // Installed but not in manifest
	MissingSkills  []ManifestSkill   // In manifest but not installed
	ExtraSkills    []InstalledSkill  // Installed but not in manifest
}

// DoctorIssue represents a health check finding.
type DoctorIssue struct {
	Severity    string // "error", "warning", "info"
	Category    string // "orphan", "drift", "missing", "broken"
	Description string
	Item        string // affected plugin/skill name
}

// Action represents an install or remove action for the orchestrator.
type Action struct {
	Type        string // "install" or "remove"
	ItemType    string // "plugin" or "skill"
	Name        string
	Source      string // marketplace for plugins, source repo for skills
	Destination string // "user" or "project"
}
