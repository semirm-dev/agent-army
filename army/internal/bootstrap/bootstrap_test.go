package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTargetDirSuffix(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{TargetClaude, ".claude"},
		{TargetCursor, ".cursor"},
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
