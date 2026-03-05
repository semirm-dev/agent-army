package bootstrap

import (
	"os"
	"path/filepath"
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

func TestTargetDirSuffix(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{TargetClaude, ".claude"},
		{TargetCursor, ".cursor"},
		{TargetGemini, ".gemini"},
		{TargetAntigravity, ".agent"},
		{"unknown", ".claude"},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			if got := targetDirSuffix(tt.target); got != tt.want {
				t.Errorf("targetDirSuffix(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestTargetGlobalDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir: %v", err)
	}

	tests := []struct {
		target string
		want   string
	}{
		{TargetClaude, filepath.Join(home, ".claude")},
		{TargetCursor, filepath.Join(home, ".cursor")},
		{TargetGemini, filepath.Join(home, ".gemini")},
		{TargetAntigravity, filepath.Join(home, ".gemini", "antigravity")},
		{"unknown", filepath.Join(home, ".claude")},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			if got := targetGlobalDir(tt.target); got != tt.want {
				t.Errorf("targetGlobalDir(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}
