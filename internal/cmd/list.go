package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var listCmd = clix.NewCommand("list",
	clix.WithCommandShort("List all dependencies"),
	clix.WithCommandLong("List all dependencies from the manifest and their locked versions"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runList(ctx)
	}),
)

func runList(ctx *clix.Context) error {
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

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	hasLockfile := err == nil

	// Print header
	if hasLockfile {
		fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %s\n", "MODULE", "CONSTRAINT", "LOCKED", "STATUS")
		fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %s\n",
			"──────────────────────────────────────────────────",
			"───────────────", "───────────────", "─────")
	} else {
		fmt.Fprintf(ctx.App.Out, "%-50s %-15s\n", "MODULE", "CONSTRAINT")
		fmt.Fprintf(ctx.App.Out, "%-50s %-15s\n",
			"──────────────────────────────────────────────────",
			"───────────────")
		fmt.Fprintf(ctx.App.Err, "Warning: No lockfile found. Run 'cpkg tidy' to lock versions.\n\n")
	}

	// List dependencies
	for modulePath, dep := range m.Dependencies {
		constraint := dep.Version
		lockedVersion := "NOT_LOCKED"
		status := ""

		if hasLockfile {
			if lockDep, exists := lock.Dependencies[modulePath]; exists {
				lockedVersion = lockDep.Version
				status = "✓"
			} else {
				status = "⚠"
			}
		}

		if hasLockfile {
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %s\n",
				modulePath, constraint, lockedVersion, status)
		} else {
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s\n",
				modulePath, constraint)
		}
	}

	if len(m.Dependencies) == 0 {
		fmt.Fprintf(ctx.App.Out, "No dependencies.\n")
	}

	return nil
}

