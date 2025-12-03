package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var graphCmd = clix.NewCommand("graph",
	clix.WithCommandShort("Display the dependency graph"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runGraph(ctx)
	}),
)

func runGraph(ctx *clix.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	if err != nil {
		return fmt.Errorf("no lockfile found, run 'cpkg tidy' first: %w", err)
	}

	// Print dependency tree
	fmt.Fprintf(ctx.App.Out, "%s\n", lock.Module)

	// Get sorted list of dependencies
	deps := make([]string, 0, len(lock.Dependencies))
	for module := range lock.Dependencies {
		deps = append(deps, module)
	}

	// Simple alphabetical sort for now
	for i := 0; i < len(deps); i++ {
		for j := i + 1; j < len(deps); j++ {
			if deps[i] > deps[j] {
				deps[i], deps[j] = deps[j], deps[i]
			}
		}
	}

	// Print dependencies
	for i, module := range deps {
		dep := lock.Dependencies[module]
		prefix := "├─"
		if i == len(deps)-1 {
			prefix = "└─"
		}
		fmt.Fprintf(ctx.App.Out, "%s %s %s\n", prefix, module, dep.Version)
	}

	return nil
}
