package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/format"
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

type listOutput struct {
	Dependencies []listDependency `json:"dependencies" yaml:"dependencies"`
}

type listDependency struct {
	Module      string `json:"module" yaml:"module"`
	Constraint  string `json:"constraint" yaml:"constraint"`
	Locked     string `json:"locked,omitempty" yaml:"locked,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty"`
	HasLockfile bool   `json:"has_lockfile" yaml:"has_lockfile"`
}

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

	outputFormat := GetFormat()
	deps := make([]listDependency, 0, len(m.Dependencies))

	// Collect dependency information
	for modulePath, dep := range m.Dependencies {
		constraint := dep.Version
		lockedVersion := ""
		status := ""

		if hasLockfile {
			if lockDep, exists := lock.Dependencies[modulePath]; exists {
				lockedVersion = lockDep.Version
				status = "locked"
			} else {
				status = "not_locked"
			}
		}

		deps = append(deps, listDependency{
			Module:      modulePath,
			Constraint:  constraint,
			Locked:     lockedVersion,
			Status:      status,
			HasLockfile: hasLockfile,
		})
	}

	// Output in requested format
	if outputFormat != format.FormatText {
		output := listOutput{Dependencies: deps}
		return format.Write(ctx.App.Out, outputFormat, output)
	}

	// Text output
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

	for _, dep := range deps {
		if hasLockfile {
			statusSymbol := "✓"
			if dep.Status == "not_locked" {
				statusSymbol = "⚠"
			}
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %s\n",
				dep.Module, dep.Constraint, dep.Locked, statusSymbol)
		} else {
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s\n",
				dep.Module, dep.Constraint)
		}
	}

	if len(deps) == 0 {
		fmt.Fprintf(ctx.App.Out, "No dependencies.\n")
	}

	return nil
}


