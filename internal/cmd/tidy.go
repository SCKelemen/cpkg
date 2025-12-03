package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var (
	tidyDepRoot string
	tidyCheck   bool
)

var tidyCmd = clix.NewCommand("tidy",
	clix.WithCommandShort("Resolve dependency graph and write lockfile"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runTidy(ctx)
	}),
)

func init() {
	tidyCmd.Flags = clix.NewFlagSet("tidy")
	tidyCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Override dependency root",
		},
		Value: &tidyDepRoot,
	})
	tidyCmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "check",
			Usage: "Check if lockfile would change without writing",
		},
		Value: &tidyCheck,
	})
}

func runTidy(ctx *clix.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	depRoot := tidyDepRoot
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = m.DepRoot
		}
	}

	// Load existing lockfile for comparison
	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	existingLock, _ := lockfile.Load(lockfilePath)

	// Resolve dependencies
	lock, err := resolveDependencies(m, depRoot)
	if err != nil {
		return err
	}

	var newDeps, updatedDeps, removedDeps []string

	// Track changes
	for modulePath, lockDep := range lock.Dependencies {
		if existingLock != nil {
			if existingDep, exists := existingLock.Dependencies[modulePath]; exists {
				if existingDep.Version != lockDep.Version {
					updatedDeps = append(updatedDeps, fmt.Sprintf("%s: %s → %s", modulePath, existingDep.Version, lockDep.Version))
				}
			} else {
				newDeps = append(newDeps, modulePath)
			}
		} else {
			newDeps = append(newDeps, modulePath)
		}
	}

	// Find removed dependencies
	if existingLock != nil {
		for modulePath := range existingLock.Dependencies {
			if _, exists := lock.Dependencies[modulePath]; !exists {
				removedDeps = append(removedDeps, modulePath)
			}
		}
	}

	// Print summary
	if len(newDeps) > 0 {
		for _, dep := range newDeps {
			fmt.Fprintf(ctx.App.Out, "+ %s @ %s\n", dep, lock.Dependencies[dep].Version)
		}
	}
	if len(updatedDeps) > 0 {
		for _, dep := range updatedDeps {
			fmt.Fprintf(ctx.App.Out, "~ %s\n", dep)
		}
	}
	if len(removedDeps) > 0 {
		for _, dep := range removedDeps {
			fmt.Fprintf(ctx.App.Out, "- %s\n", dep)
		}
	}

	// Check mode
	if tidyCheck {
		if existingLock == nil {
			return fmt.Errorf("lockfile would be created")
		}
		// Compare lockfiles (simplified - just check if dependencies match)
		if len(existingLock.Dependencies) != len(lock.Dependencies) {
			return fmt.Errorf("lockfile would change")
		}
		for module, dep := range lock.Dependencies {
			existingDep, exists := existingLock.Dependencies[module]
			if !exists || existingDep.Version != dep.Version || existingDep.Commit != dep.Commit {
				return fmt.Errorf("lockfile would change")
			}
		}
		return nil
	}

	// Write lockfile
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		return fmt.Errorf("failed to save lockfile: %w", err)
	}

	fmt.Fprintf(ctx.App.Out, "Lockfile written to %s\n", lockfile.LockfileName)
	return nil
}
