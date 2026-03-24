package cli

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/army/internal/installer"
	"github.com/smahovkic/agent-army/army/internal/runner"
	"github.com/smahovkic/agent-army/army/internal/state"
	"github.com/smahovkic/agent-army/army/internal/core/catalog"
	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/orchestrator"
	"github.com/smahovkic/agent-army/army/internal/core/types"
)

// commandRunner is satisfied by runner.RealRunner and runner.DryRunner.
type commandRunner interface {
	Run(cmd string, args ...string) (string, error)
}

// deps bundles all resolved dependencies for commands.
type deps struct {
	catalog      *catalog.Service
	manifest     *types.Manifest
	manifestPath string
	orchestrator *orchestrator.Orchestrator
	state        *state.Reader
}

func resolveDeps() (*deps, error) {
	cat, err := catalog.New()
	if err != nil {
		return nil, fmt.Errorf("loading catalog: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}
	manifestPath, err := manifest.ResolveFromDir(cwd)
	if err != nil {
		return nil, fmt.Errorf("determining manifest path: %w", err)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading manifest: %w", err)
	}

	var r commandRunner
	if globalFlags.DryRun {
		r = runner.NewDry()
	} else {
		r = runner.NewReal(nil)
	}

	pi := installer.NewPlugin(r)
	si := installer.NewSkill(r)
	sr := state.New()
	orch := orchestrator.New(pi, si, sr, os.Stdout)

	return &deps{
		catalog:      cat,
		manifest:     m,
		manifestPath: manifestPath,
		orchestrator: orch,
		state:        sr,
	}, nil
}
