package graph

import (
	"github.com/semir/agent-army/internal/model"
)

// ResolveTransitive walks the dependency graph breadth-first starting from seeds.
// Returns deduplicated list in BFS discovery order.
func ResolveTransitive(seeds []string, getDeps func(string) []string) []string {
	if len(seeds) == 0 {
		return nil
	}

	visited := make(map[string]bool)
	var result []string
	queue := make([]string, len(seeds))
	copy(queue, seeds)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		result = append(result, current)
		queue = append(queue, getDeps(current)...)
	}

	return result
}

// FindRedundant finds entries that are transitively covered by other entries.
// Returns at most one Redundancy per target (the first covering entry found).
func FindRedundant(entries []string, getDeps func(string) []string) []model.Redundancy {
	if len(entries) == 0 {
		return nil
	}

	var redundancies []model.Redundancy

	for i, target := range entries {
		for j, other := range entries {
			if i == j {
				continue
			}
			closure := ResolveTransitive([]string{other}, getDeps)
			if contains(closure, target) {
				redundancies = append(redundancies, model.Redundancy{
					Target:    target,
					CoveredBy: other,
				})
				break
			}
		}
	}

	return redundancies
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
