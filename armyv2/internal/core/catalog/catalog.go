package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Service provides access to the merged catalog of plugins, skills, and tech profiles.
type Service struct {
	catalog types.Catalog
}

// New creates a catalog service by loading the embedded catalog and merging
// it with any updated catalog found at ~/.armyv2/catalog.json.
func New() (*Service, error) {
	base, err := parseCatalog(embeddedCatalog)
	if err != nil {
		return nil, fmt.Errorf("parsing embedded catalog: %w", err)
	}

	updated, err := loadUpdatedCatalog()
	if err != nil {
		// Non-fatal: if we can't load the updated catalog, use embedded only.
		return &Service{catalog: base}, nil
	}

	merged := mergeCatalogs(base, updated)
	return &Service{catalog: merged}, nil
}

// NewFromBytes creates a catalog service from raw JSON bytes. Useful for testing.
func NewFromBytes(data []byte) (*Service, error) {
	if err := Validate(data); err != nil {
		return nil, fmt.Errorf("validating catalog: %w", err)
	}

	cat, err := parseCatalog(data)
	if err != nil {
		return nil, fmt.Errorf("parsing catalog: %w", err)
	}

	return &Service{catalog: cat}, nil
}

// FindPlugin returns the catalog plugin with the given name (case-insensitive).
func (s *Service) FindPlugin(name string) (types.CatalogPlugin, bool) {
	lower := strings.ToLower(name)
	for _, p := range s.catalog.Plugins {
		if strings.ToLower(p.Name) == lower {
			return p, true
		}
	}
	return types.CatalogPlugin{}, false
}

// FindSkill returns the catalog skill with the given name (case-insensitive).
func (s *Service) FindSkill(name string) (types.CatalogSkill, bool) {
	lower := strings.ToLower(name)
	for _, sk := range s.catalog.Skills {
		if strings.ToLower(sk.Name) == lower {
			return sk, true
		}
	}
	return types.CatalogSkill{}, false
}

// GetTechProfile returns the tech profile for the given technology name (case-insensitive).
func (s *Service) GetTechProfile(name string) (types.TechProfile, bool) {
	lower := strings.ToLower(name)
	for k, v := range s.catalog.TechProfiles {
		if strings.ToLower(k) == lower {
			return v, true
		}
	}
	return types.TechProfile{}, false
}

// AllPlugins returns all plugins in the catalog.
func (s *Service) AllPlugins() []types.CatalogPlugin {
	result := make([]types.CatalogPlugin, len(s.catalog.Plugins))
	copy(result, s.catalog.Plugins)
	return result
}

// AllSkills returns all skills in the catalog.
func (s *Service) AllSkills() []types.CatalogSkill {
	result := make([]types.CatalogSkill, len(s.catalog.Skills))
	copy(result, s.catalog.Skills)
	return result
}

// AllTechProfiles returns all tech profiles in the catalog.
func (s *Service) AllTechProfiles() map[string]types.TechProfile {
	result := make(map[string]types.TechProfile, len(s.catalog.TechProfiles))
	for k, v := range s.catalog.TechProfiles {
		result[k] = v
	}
	return result
}

// Validate checks that the given JSON bytes represent a valid catalog structure.
func Validate(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// version is required
	versionRaw, ok := raw["version"]
	if !ok {
		return fmt.Errorf("missing required field: version")
	}
	var version int
	if err := json.Unmarshal(versionRaw, &version); err != nil {
		return fmt.Errorf("field 'version' must be an integer: %w", err)
	}
	if version < 1 {
		return fmt.Errorf("field 'version' must be >= 1, got %d", version)
	}

	// plugins must be an array if present
	if pluginsRaw, ok := raw["plugins"]; ok {
		var plugins []json.RawMessage
		if err := json.Unmarshal(pluginsRaw, &plugins); err != nil {
			return fmt.Errorf("field 'plugins' must be an array: %w", err)
		}
		for i, pRaw := range plugins {
			if err := validateCatalogPlugin(pRaw, i); err != nil {
				return err
			}
		}
	}

	// skills must be an array if present
	if skillsRaw, ok := raw["skills"]; ok {
		var skills []json.RawMessage
		if err := json.Unmarshal(skillsRaw, &skills); err != nil {
			return fmt.Errorf("field 'skills' must be an array: %w", err)
		}
		for i, sRaw := range skills {
			if err := validateCatalogSkill(sRaw, i); err != nil {
				return err
			}
		}
	}

	// tech_profiles must be an object if present
	if tpRaw, ok := raw["tech_profiles"]; ok {
		var profiles map[string]json.RawMessage
		if err := json.Unmarshal(tpRaw, &profiles); err != nil {
			return fmt.Errorf("field 'tech_profiles' must be an object: %w", err)
		}
	}

	return nil
}

// validateCatalogPlugin checks that a plugin entry has the required 'name' field.
func validateCatalogPlugin(data json.RawMessage, index int) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("plugins[%d]: must be an object: %w", index, err)
	}
	if _, ok := fields["name"]; !ok {
		return fmt.Errorf("plugins[%d]: missing required field 'name'", index)
	}
	return nil
}

// validateCatalogSkill checks that a skill entry has the required 'name' field.
func validateCatalogSkill(data json.RawMessage, index int) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("skills[%d]: must be an object: %w", index, err)
	}
	if _, ok := fields["name"]; !ok {
		return fmt.Errorf("skills[%d]: missing required field 'name'", index)
	}
	return nil
}

// parseCatalog unmarshals JSON bytes into a Catalog struct.
func parseCatalog(data []byte) (types.Catalog, error) {
	var cat types.Catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return types.Catalog{}, fmt.Errorf("unmarshaling catalog: %w", err)
	}

	// Ensure nil slices/maps are initialized to empty values.
	if cat.Plugins == nil {
		cat.Plugins = []types.CatalogPlugin{}
	}
	if cat.Skills == nil {
		cat.Skills = []types.CatalogSkill{}
	}
	if cat.TechProfiles == nil {
		cat.TechProfiles = make(map[string]types.TechProfile)
	}

	return cat, nil
}

// loadUpdatedCatalog attempts to read ~/.armyv2/catalog.json.
func loadUpdatedCatalog() (types.Catalog, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return types.Catalog{}, fmt.Errorf("getting home directory: %w", err)
	}

	path := filepath.Join(home, ".armyv2", "catalog.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return types.Catalog{}, fmt.Errorf("reading %s: %w", path, err)
	}

	cat, err := parseCatalog(data)
	if err != nil {
		return types.Catalog{}, fmt.Errorf("parsing updated catalog: %w", err)
	}

	return cat, nil
}

// mergeCatalogs merges the base (embedded) catalog with an updated one.
// The updated catalog takes precedence if its version is >= the base version.
// When the updated catalog wins, its plugins and skills fully replace the base lists;
// tech profiles are merged at the key level (updated keys override base keys).
func mergeCatalogs(base, updated types.Catalog) types.Catalog {
	if updated.Version < base.Version {
		return base
	}

	merged := types.Catalog{
		Version:      updated.Version,
		UpdatedAt:    updated.UpdatedAt,
		Plugins:      updated.Plugins,
		Skills:       updated.Skills,
		TechProfiles: make(map[string]types.TechProfile),
	}

	// Start with base tech profiles.
	for k, v := range base.TechProfiles {
		merged.TechProfiles[k] = v
	}
	// Override with updated tech profiles.
	for k, v := range updated.TechProfiles {
		merged.TechProfiles[k] = v
	}

	// Ensure non-nil collections.
	if merged.Plugins == nil {
		merged.Plugins = []types.CatalogPlugin{}
	}
	if merged.Skills == nil {
		merged.Skills = []types.CatalogSkill{}
	}

	return merged
}
