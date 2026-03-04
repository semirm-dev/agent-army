package bootstrap

import (
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestCursorShortName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"go/patterns", "golang"},
		{"go/testing", "golang-testing"},
		{"typescript/patterns", "typescript"},
		{"security", "security"},
	}
	for _, tt := range tests {
		got := cursorShortName(model.Rule{Name: tt.name})
		if got != tt.want {
			t.Errorf("cursorShortName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
