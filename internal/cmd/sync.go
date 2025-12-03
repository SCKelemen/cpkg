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

var syncDepRoot string

var syncCmd = clix.NewCommand("sync",
	clix.WithCommandShort("Sync git submodules to match lockfile"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runSync(ctx)
	}),
)

func init() {
	syncCmd.Flags = clix.NewFlagSet("sync")
	syncCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Override dependency root",
		},
		Value: &syncDepRoot,
	})
}

func runSync(ctx *clix.Context) error {
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
		// If lockfile doesn't exist, run tidy first
		fmt.Fprintf(ctx.App.Out, "Lockfile not found, running tidy...\n")
		// We can't call tidy directly, so return error
		return fmt.Errorf("lockfile not found, run 'cpkg tidy' first")
	}

	// Note: depRoot is determined from lockfile, paths are already set in lock.Dependencies
	// syncDepRoot is not used here as paths come from lockfile

	// Resolve symlinks in working directory (important for macOS where /tmp is a symlink)
	realCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		realCwd = cwd // Fallback to original if resolution fails
	}
	realManifestPath, err := filepath.EvalSymlinks(manifestPath)
	if err != nil {
		realManifestPath = manifestPath // Fallback to original if resolution fails
	}

	// Sync each dependency
	for modulePath, dep := range lock.Dependencies {
		path := dep.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(realManifestPath), path)
		}

		// Resolve symlinks in the path (important for macOS)
		resolvedPath, err := submodule.ResolveSymlinks(path)
		if err == nil {
			path = resolvedPath
		}
		// If resolution fails, use original path (might not exist yet)

		// Check if submodule exists (use relative path from git root)
		relPath, err := filepath.Rel(realCwd, path)
		if err != nil {
			relPath = path // Fallback to absolute if relative fails
		}
		exists := submodule.SubmoduleExists(relPath)
		currentURL, _ := submodule.GetSubmoduleURL(relPath)

		if !exists {
			// Add submodule (use relative path for git submodule add)
			if err := submodule.EnsureDir(path); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", modulePath, err)
			}
			if err := submodule.AddSubmodule(dep.RepoURL, relPath); err != nil {
				return fmt.Errorf("failed to add submodule %s: %w", modulePath, err)
			}
			fmt.Fprintf(ctx.App.Out, "+ %s\n", modulePath)
		} else if currentURL != dep.RepoURL {
			// Update submodule URL
			if err := submodule.SetSubmoduleURL(relPath, dep.RepoURL); err != nil {
				return fmt.Errorf("failed to update submodule URL for %s: %w", modulePath, err)
			}
			fmt.Fprintf(ctx.App.Out, "~ %s (URL updated)\n", modulePath)
		}

		// Initialize if needed
		if err := submodule.InitSubmodule(relPath); err != nil {
			// Ignore error if already initialized
		}

		// Fetch tags and commits (use absolute path for git -C)
		if err := submodule.FetchTags(path); err != nil {
			return fmt.Errorf("failed to fetch tags for %s: %w", modulePath, err)
		}
		if err := submodule.FetchCommit(path); err != nil {
			return fmt.Errorf("failed to fetch commits for %s: %w", modulePath, err)
		}

		// Checkout the locked commit (use absolute path for git -C)
		if err := submodule.Checkout(path, dep.Commit); err != nil {
			return fmt.Errorf("failed to checkout %s@%s: %w", modulePath, dep.Version, err)
		}

		// Get current commit for display (use absolute path for git -C)
		currentCommit, _ := submodule.GetSubmoduleCommit(path)
		shortCommit := currentCommit
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}
		fmt.Fprintf(ctx.App.Out, "✓ %s @ %s (%s)\n", modulePath, dep.Version, shortCommit)
	}

	return nil
}
