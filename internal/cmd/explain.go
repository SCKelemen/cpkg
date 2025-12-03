package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/submodule"
)

var explainCmd = clix.NewCommand("explain",
	clix.WithCommandShort("Explain a dependency in detail"),
	clix.WithCommandLong("Show detailed information about a specific dependency"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runExplain(ctx)
	}),
	clix.WithCommandArguments(
		clix.NewArgument(
			clix.WithArgName("module"),
			clix.WithArgRequired(),
		),
	),
)

func runExplain(ctx *clix.Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("module path required")
	}

	modulePath := ctx.Args[0]

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

	dep, exists := m.Dependencies[modulePath]
	if !exists {
		return fmt.Errorf("dependency %s not found in manifest", modulePath)
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	hasLockfile := err == nil

	// Print dependency information
	fmt.Fprintf(ctx.App.Out, "Dependency: %s\n", modulePath)
	fmt.Fprintf(ctx.App.Out, "─────────────────────────────────────────────────────────────\n\n")

	fmt.Fprintf(ctx.App.Out, "Constraint: %s\n", dep.Version)

	if hasLockfile {
		if lockDep, exists := lock.Dependencies[modulePath]; exists {
			fmt.Fprintf(ctx.App.Out, "\nLocked Information:\n")
			fmt.Fprintf(ctx.App.Out, "  Version: %s\n", lockDep.Version)
			fmt.Fprintf(ctx.App.Out, "  Commit:  %s\n", lockDep.Commit)
			fmt.Fprintf(ctx.App.Out, "  Sum:     %s\n", lockDep.Sum)
			fmt.Fprintf(ctx.App.Out, "  VCS:     %s\n", lockDep.VCS)
			fmt.Fprintf(ctx.App.Out, "  Repo:    %s\n", lockDep.RepoURL)
			fmt.Fprintf(ctx.App.Out, "  Path:    %s\n", lockDep.Path)

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

			fmt.Fprintf(ctx.App.Out, "\nLocal State:\n")
			if submodule.SubmoduleExists(relPath) {
				currentCommit, err := submodule.GetSubmoduleCommit(submodulePath)
				if err == nil {
					shortCommit := currentCommit
					if len(shortCommit) > 7 {
						shortCommit = shortCommit[:7]
					}
					fmt.Fprintf(ctx.App.Out, "  Submodule: exists\n")
					fmt.Fprintf(ctx.App.Out, "  Current commit: %s\n", shortCommit)

					if currentCommit == lockDep.Commit {
						fmt.Fprintf(ctx.App.Out, "  Status: ✓ in sync\n")
					} else {
						fmt.Fprintf(ctx.App.Out, "  Status: ⚠ out of sync (locked: %s)\n", lockDep.Commit[:7])
					}

					dirty, _ := submodule.IsSubmoduleDirty(submodulePath)
					if dirty {
						fmt.Fprintf(ctx.App.Out, "  Working tree: ⚠ dirty (has uncommitted changes)\n")
					} else {
						fmt.Fprintf(ctx.App.Out, "  Working tree: ✓ clean\n")
					}
				} else {
					fmt.Fprintf(ctx.App.Out, "  Submodule: exists but cannot read commit\n")
				}
			} else {
				fmt.Fprintf(ctx.App.Out, "  Submodule: ⚠ not initialized\n")
				fmt.Fprintf(ctx.App.Out, "  Run 'cpkg sync' to initialize\n")
			}
		} else {
			fmt.Fprintf(ctx.App.Out, "\nLocked: ⚠ not locked\n")
			fmt.Fprintf(ctx.App.Out, "Run 'cpkg tidy' to lock this dependency\n")
		}
	} else {
		fmt.Fprintf(ctx.App.Out, "\nLocked: ⚠ no lockfile found\n")
		fmt.Fprintf(ctx.App.Out, "Run 'cpkg tidy' to create lockfile\n")
	}

	return nil
}

