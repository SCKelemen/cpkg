package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/format"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/submodule"
)

var statusCmd = clix.NewCommand("status",
	clix.WithCommandShort("Show dependency status"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runStatus(ctx)
	}),
)

type statusOutput struct {
	Dependencies []statusDependency `json:"dependencies" yaml:"dependencies"`
}

type statusDependency struct {
	Module        string `json:"module" yaml:"module"`
	Constraint    string `json:"constraint" yaml:"constraint"`
	LockedVersion string `json:"locked_version" yaml:"locked_version"`
	LocalVersion  string `json:"local_version" yaml:"local_version"`
	Status        string `json:"status" yaml:"status"`
}

func runStatus(ctx *clix.Context) error {
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
	if err != nil {
		if GetFormat() != format.FormatText {
			// Return empty output for JSON/YAML when no lockfile
			output := statusOutput{Dependencies: []statusDependency{}}
			return format.Write(ctx.App.Out, GetFormat(), output)
		}
		fmt.Fprintf(ctx.App.Out, "No lockfile found. Run 'cpkg tidy' to create one.\n")
		return nil
	}

	outputFormat := GetFormat()
	deps := make([]statusDependency, 0, len(m.Dependencies))

	// Check each dependency
	for modulePath, dep := range m.Dependencies {
		constraint := dep.Version
		lockedVersion := ""
		localVersion := ""
		status := "NO_LOCK"

		if lockDep, exists := lock.Dependencies[modulePath]; exists {
			lockedVersion = lockDep.Version
			status = "OK"

			// Check local submodule state
			submodulePath := lockDep.Path
			if !filepath.IsAbs(submodulePath) {
				submodulePath = filepath.Join(filepath.Dir(manifestPath), submodulePath)
			}

			// Resolve symlinks (important for macOS)
			resolvedPath, err := filepath.EvalSymlinks(submodulePath)
			if err == nil {
				submodulePath = resolvedPath
			}

			// Get relative path for git submodule commands
			relPath := submodulePath
			if filepath.IsAbs(submodulePath) {
				realCwd, _ := filepath.EvalSymlinks(cwd)
				if rel, err := filepath.Rel(realCwd, submodulePath); err == nil {
					relPath = rel
				}
			}

			if submodule.SubmoduleExists(relPath) {
				currentCommit, err := submodule.GetSubmoduleCommit(submodulePath)
				if err == nil {
					lockedCommit := lockDep.Commit
					if currentCommit == lockedCommit {
						localVersion = lockedVersion
					} else {
						shortCommit := currentCommit
						if len(shortCommit) > 7 {
							shortCommit = shortCommit[:7]
						}
						localVersion = shortCommit
						status = "OUT_OF_SYNC"
					}

					// Check if dirty
					dirty, _ := submodule.IsSubmoduleDirty(submodulePath)
					if dirty {
						localVersion += " (dirty)"
						status = "DIRTY"
					}
				}
			} else {
				localVersion = "MISSING"
				status = "MISSING"
			}
		} else {
			localVersion = "MISSING"
		}

		deps = append(deps, statusDependency{
			Module:        modulePath,
			Constraint:    constraint,
			LockedVersion: lockedVersion,
			LocalVersion:  localVersion,
			Status:        status,
		})
	}

	// Output in requested format
	if outputFormat != format.FormatText {
		output := statusOutput{Dependencies: deps}
		return format.Write(ctx.App.Out, outputFormat, output)
	}

	// Text output
	fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %-15s\n", "MODULE", "CONSTRAINT", "LOCKED", "LOCAL", "STATUS")
	fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %-15s\n",
		"──────────────────────────────────────────────────",
		"───────────────", "───────────────", "───────────────", "───────────────")

	for _, dep := range deps {
		locked := dep.LockedVersion
		if locked == "" {
			locked = "NO_LOCK"
		}
		local := dep.LocalVersion
		if local == "" {
			local = "MISSING"
		}
		fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %-15s\n",
			dep.Module, dep.Constraint, locked, local, dep.Status)
	}

	return nil
}
