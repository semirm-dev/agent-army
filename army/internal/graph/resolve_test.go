package graph

import (
	"testing"
)

func TestResolveTransitive_Empty(t *testing.T) {
	got := ResolveTransitive(nil, func(string) []string { return nil })
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestResolveTransitive_Chain(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {},
	}
	got := ResolveTransitive([]string{"A"}, func(n string) []string { return deps[n] })
	want := []string{"A", "B", "C"}
	assertStrSlice(t, got, want)
}

func TestResolveTransitive_Diamond(t *testing.T) {
	deps := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
		"D": {},
	}
	got := ResolveTransitive([]string{"A"}, func(n string) []string { return deps[n] })
	want := []string{"A", "B", "C", "D"}
	assertStrSlice(t, got, want)
}

func TestResolveTransitive_Cycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	got := ResolveTransitive([]string{"A"}, func(n string) []string { return deps[n] })
	want := []string{"A", "B"}
	assertStrSlice(t, got, want)
}

func TestFindRedundant_None(t *testing.T) {
	deps := map[string][]string{
		"A": {},
		"B": {},
	}
	got := FindRedundant([]string{"A", "B"}, func(n string) []string { return deps[n] })
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestFindRedundant_OneRedundant(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {},
	}
	got := FindRedundant([]string{"A", "B"}, func(n string) []string { return deps[n] })
	if len(got) != 1 {
		t.Fatalf("got %d redundancies, want 1", len(got))
	}
	if got[0].Target != "B" || got[0].CoveredBy != "A" {
		t.Errorf("got target=%q coveredBy=%q, want B/A", got[0].Target, got[0].CoveredBy)
	}
}

func TestFindRedundant_Transitive(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {},
	}
	got := FindRedundant([]string{"A", "C"}, func(n string) []string { return deps[n] })
	if len(got) != 1 {
		t.Fatalf("got %d, want 1", len(got))
	}
	if got[0].Target != "C" {
		t.Errorf("target = %q, want C", got[0].Target)
	}
}

func assertStrSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
