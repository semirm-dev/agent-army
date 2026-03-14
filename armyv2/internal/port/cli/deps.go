package cli

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/armyv2/internal/adapter/plugin"
	"github.com/smahovkic/agent-army/armyv2/internal/adapter/runner"
	"github.com/smahovkic/agent-army/armyv2/internal/adapter/skill"
	"github.com/smahovkic/agent-army/armyv2/internal/adapter/system"
	"github.com/smahovkic/agent-army/armyv2/internal/core/catalog"
	"github.com/smahovkic/agent-army/armyv2/internal/core/manifest"
	"github.com/smahovkic/agent-army/armyv2/internal/core/orchestrator"
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// deps bundles all resolved dependencies for commands.
type deps struct {
	catalog      *catalog.Service
	manifest     *types.Manifest
	manifestPath string
	orchestrator *orchestrator.Orchestrator
	system       *system.Reader
	runner       runner.Runner
}

func resolveDeps() (*deps, error) {
	cat, err := catalog.New()
	if err != nil {
		return nil, fmt.Errorf("loading catalog: %w", err)
	}

	manifestPath := globalFlags.ManifestPath
	if manifestPath == "" {
		manifestPath, err = manifest.DefaultPath()
		if err != nil {
			return nil, fmt.Errorf("determining manifest path: %w", err)
		}
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	var r runner.Runner
	if globalFlags.DryRun {
		r = runner.NewDry()
	} else {
		r = runner.NewReal(nil)
	}

	pluginAdapter := plugin.New(r)
	skillAdapter := skill.New(r)
	sysReader := system.New()
	orch := orchestrator.New(pluginAdapter, skillAdapter, sysReader, os.Stdout)

	return &deps{
		catalog:      cat,
		manifest:     m,
		manifestPath: manifestPath,
		orchestrator: orch,
		system:       sysReader,
		runner:       r,
	}, nil
}
