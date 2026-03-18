package orchestrator

import (
	"fmt"
	"io"
	"strings"

	"github.com/smahovkic/agent-army/army/internal/core/diff"
	"github.com/smahovkic/agent-army/army/internal/core/types"
)

// PluginInstaller handles plugin install/remove operations.
type PluginInstaller interface {
	Install(name string) error
	Remove(name string) error
}

// SkillInstaller handles skill install/remove operations.
type SkillInstaller interface {
	Install(name, source string) error
	Remove(name string) error
}

// SystemReader reads installed state from the filesystem.
type SystemReader interface {
	InstalledPlugins() ([]types.InstalledPlugin, error)
	InstalledSkills() ([]types.InstalledSkill, error)
}

// Orchestrator coordinates install/remove/sync operations.
type Orchestrator struct {
	plugins PluginInstaller
	skills  SkillInstaller
	system  SystemReader
	out     io.Writer
}

// New creates an Orchestrator with the given adapters.
func New(plugins PluginInstaller, skills SkillInstaller, system SystemReader, out io.Writer) *Orchestrator {
	return &Orchestrator{
		plugins: plugins,
		skills:  skills,
		system:  system,
		out:     out,
	}
}

// Result tracks the outcome of a batch operation.
type Result struct {
	Succeeded int
	Failed    int
	Errors    []error
}

// PlanActions produces the list of actions needed to reconcile manifest with installed state.
func (o *Orchestrator) PlanActions(manifest *types.Manifest) ([]types.Action, error) {
	installedPlugins, err := o.system.InstalledPlugins()
	if err != nil {
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}
	installedSkills, err := o.system.InstalledSkills()
	if err != nil {
		return nil, fmt.Errorf("reading installed skills: %w", err)
	}

	d := diff.Compare(manifest, installedPlugins, installedSkills)

	var actions []types.Action

	for _, p := range d.MissingPlugins {
		actions = append(actions, types.Action{
			Type:        "install",
			ItemType:    "plugin",
			Name:        p.Name,
			Source:      p.Marketplace,
			Destination: p.Destination,
		})
	}

	for _, p := range d.ExtraPlugins {
		actions = append(actions, types.Action{
			Type:     "remove",
			ItemType: "plugin",
			Name:     p.Name,
			Source:   p.Marketplace,
		})
	}

	for _, s := range d.MissingSkills {
		actions = append(actions, types.Action{
			Type:        "install",
			ItemType:    "skill",
			Name:        s.Name,
			Source:      s.Source,
			Destination: s.Destination,
		})
	}

	for _, s := range d.ExtraSkills {
		actions = append(actions, types.Action{
			Type:     "remove",
			ItemType: "skill",
			Name:     s.Name,
			Source:   s.Source,
		})
	}

	return actions, nil
}

// PlanClear produces remove actions for all manifest items that are currently installed.
func (o *Orchestrator) PlanClear(manifest *types.Manifest) ([]types.Action, error) {
	installedPlugins, err := o.system.InstalledPlugins()
	if err != nil {
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}
	installedSkills, err := o.system.InstalledSkills()
	if err != nil {
		return nil, fmt.Errorf("reading installed skills: %w", err)
	}

	pluginSet := make(map[string]bool, len(installedPlugins))
	for _, p := range installedPlugins {
		pluginSet[strings.ToLower(p.Name)] = true
	}
	skillSet := make(map[string]bool, len(installedSkills))
	for _, s := range installedSkills {
		skillSet[strings.ToLower(s.Name)] = true
	}

	var actions []types.Action

	for _, mp := range manifest.Plugins {
		if pluginSet[strings.ToLower(mp.Name)] {
			actions = append(actions, types.Action{
				Type:     "remove",
				ItemType: "plugin",
				Name:     mp.Name,
				Source:   mp.Marketplace,
			})
		}
	}

	for _, ms := range manifest.Skills {
		if skillSet[strings.ToLower(ms.Name)] {
			actions = append(actions, types.Action{
				Type:     "remove",
				ItemType: "skill",
				Name:     ms.Name,
				Source:   ms.Source,
			})
		}
	}

	return actions, nil
}

// PlanFullClear produces remove actions for every installed plugin and skill.
func (o *Orchestrator) PlanFullClear() ([]types.Action, error) {
	installedPlugins, err := o.system.InstalledPlugins()
	if err != nil {
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}
	installedSkills, err := o.system.InstalledSkills()
	if err != nil {
		return nil, fmt.Errorf("reading installed skills: %w", err)
	}

	var actions []types.Action

	for _, p := range installedPlugins {
		actions = append(actions, types.Action{
			Type:     "remove",
			ItemType: "plugin",
			Name:     p.Name,
			Source:   p.Marketplace,
		})
	}

	for _, s := range installedSkills {
		actions = append(actions, types.Action{
			Type:     "remove",
			ItemType: "skill",
			Name:     s.Name,
			Source:   s.Source,
		})
	}

	return actions, nil
}

// Execute runs the given actions sequentially to avoid races on the external
// state files managed by the claude CLI. Continues on failure and reports all errors.
func (o *Orchestrator) Execute(actions []types.Action) Result {
	var pluginActions, skillActions []types.Action
	for _, a := range actions {
		switch a.ItemType {
		case "plugin":
			pluginActions = append(pluginActions, a)
		case "skill":
			skillActions = append(skillActions, a)
		}
	}

	result := Result{}

	// Execute plugin actions sequentially to avoid races on installed_plugins.json
	for _, a := range pluginActions {
		if err := o.executeAction(a); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err)
		} else {
			result.Succeeded++
		}
	}

	// Execute skill actions sequentially
	for _, a := range skillActions {
		if err := o.executeAction(a); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, err)
		} else {
			result.Succeeded++
		}
	}

	return result
}

// InstallItems installs the given plugins and skills directly (used by setup wizard).
func (o *Orchestrator) InstallItems(plugins []types.ManifestPlugin, skills []types.ManifestSkill) Result {
	var actions []types.Action

	for _, p := range plugins {
		actions = append(actions, types.Action{
			Type:        "install",
			ItemType:    "plugin",
			Name:        p.Name,
			Source:      p.Marketplace,
			Destination: p.Destination,
		})
	}

	for _, s := range skills {
		actions = append(actions, types.Action{
			Type:        "install",
			ItemType:    "skill",
			Name:        s.Name,
			Source:      s.Source,
			Destination: s.Destination,
		})
	}

	return o.Execute(actions)
}

func (o *Orchestrator) executeAction(a types.Action) error {
	switch {
	case a.ItemType == "plugin" && a.Type == "install":
		fmt.Fprintf(o.out, "Installing plugin %s...\n", a.Name)
		if err := o.plugins.Install(a.Name); err != nil {
			fmt.Fprintf(o.out, "Failed to install plugin %s: %v\n", a.Name, err)
			return fmt.Errorf("install plugin %s: %w", a.Name, err)
		}
		fmt.Fprintf(o.out, "Installed plugin %s\n", a.Name)

	case a.ItemType == "plugin" && a.Type == "remove":
		fmt.Fprintf(o.out, "Removing plugin %s...\n", a.Name)
		if err := o.plugins.Remove(a.Name); err != nil {
			fmt.Fprintf(o.out, "Failed to remove plugin %s: %v\n", a.Name, err)
			return fmt.Errorf("remove plugin %s: %w", a.Name, err)
		}
		fmt.Fprintf(o.out, "Removed plugin %s\n", a.Name)

	case a.ItemType == "skill" && a.Type == "install":
		fmt.Fprintf(o.out, "Installing skill %s...\n", a.Name)
		if err := o.skills.Install(a.Name, a.Source); err != nil {
			fmt.Fprintf(o.out, "Failed to install skill %s: %v\n", a.Name, err)
			return fmt.Errorf("install skill %s: %w", a.Name, err)
		}
		fmt.Fprintf(o.out, "Installed skill %s\n", a.Name)

	case a.ItemType == "skill" && a.Type == "remove":
		fmt.Fprintf(o.out, "Removing skill %s...\n", a.Name)
		if err := o.skills.Remove(a.Name); err != nil {
			fmt.Fprintf(o.out, "Failed to remove skill %s: %v\n", a.Name, err)
			return fmt.Errorf("remove skill %s: %w", a.Name, err)
		}
		fmt.Fprintf(o.out, "Removed skill %s\n", a.Name)
	}

	return nil
}
