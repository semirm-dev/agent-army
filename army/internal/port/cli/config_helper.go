package cli

import (
	"os"

	"github.com/smahovkic/agent-army/army/internal/core/config"
	"github.com/smahovkic/agent-army/army/internal/core/manifest"
)

// registerManifestMapping saves a cwd → manifestPath mapping in config.json.
// Failures are non-fatal and silently ignored (config is a convenience feature).
func registerManifestMapping(manifestPath string) {
	defaultPath, err := manifest.DefaultPath()
	if err != nil {
		return
	}

	cfg, err := config.Load()
	if err != nil {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	if err := config.Register(cfg, cwd, manifestPath, defaultPath); err != nil {
		return
	}

	_ = config.Save(cfg)
}
