package resolver

import (
	"github.com/semir/agent-army/internal/model"
)

// ValidateAllRefs checks that every dependency reference points to an existing entity.
func ValidateAllRefs(
	skills []model.Skill,
	agents []model.Agent,
	plugins []string,
) []model.ValidationError {
	skillNames := makeSet(skills, func(s model.Skill) string { return s.Name })
	agentNames := makeSet(agents, func(a model.Agent) string { return a.Name })
	pluginNames := makeStringSet(plugins)

	var errors []model.ValidationError

	for _, s := range skills {
		errors = append(errors, checkRefs(s.Path, s.UsesSkills, "uses_skills", skillNames, "error")...)
	}

	for _, a := range agents {
		errors = append(errors, checkRefs(a.Path, a.UsesSkills, "uses_skills", skillNames, "error")...)
		errors = append(errors, checkRefs(a.Path, a.UsesPlugins, "uses_plugins", pluginNames, "warning")...)
		errors = append(errors, checkRefs(a.Path, a.DelegatesTo, "delegates_to", agentNames, "error")...)
	}

	return errors
}

func checkRefs(fileLabel string, refs []string, field string, validNames map[string]bool, severity string) []model.ValidationError {
	var errs []model.ValidationError
	for _, ref := range refs {
		if !validNames[ref] {
			errs = append(errs, model.ValidationError{
				FileLabel: fileLabel,
				Field:     field,
				Ref:       ref,
				Severity:  severity,
			})
		}
	}
	return errs
}

func makeSet[T any](items []T, key func(T) string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[key(item)] = true
	}
	return m
}

func makeStringSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}
