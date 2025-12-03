package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/submodule"
)

// runTidyInternal is an internal version of tidy that can be called from other commands
func runTidyInternal(cwd, depRootOverride string, check bool) error {
	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	depRoot := depRootOverride
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = m.DepRoot
		}
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := resolveDependencies(m, depRoot)
	if err != nil {
		return err
	}

	if check {
		existingLock, _ := lockfile.Load(lockfilePath)
		if existingLock == nil {
			return fmt.Errorf("lockfile would be created")
		}
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

	return lockfile.Save(lock, lockfilePath)
}

// runSyncInternal is an internal version of sync that can be called from other commands
func runSyncInternal(cwd, depRootOverride string) error {
	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	if err != nil {
		return fmt.Errorf("lockfile not found, run 'cpkg tidy' first: %w", err)
	}

	// Note: depRoot is determined from lockfile, paths are already set in lock.Dependencies
	// depRootOverride is not used here as paths come from lockfile

	// Resolve symlinks in working directory (important for macOS where /tmp is a symlink)
	realCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		realCwd = cwd // Fallback to original if resolution fails
	}
	realManifestPath, err := filepath.EvalSymlinks(manifestPath)
	if err != nil {
		realManifestPath = manifestPath // Fallback to original if resolution fails
	}

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

		// Use relative path for git submodule commands
		relPath, err := filepath.Rel(realCwd, path)
		if err != nil {
			relPath = path // Fallback to absolute if relative fails
		}

		exists := submodule.SubmoduleExists(relPath)
		currentURL, _ := submodule.GetSubmoduleURL(relPath)

		if !exists {
			if err := submodule.EnsureDir(path); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", modulePath, err)
			}
			if err := submodule.AddSubmodule(dep.RepoURL, relPath); err != nil {
				return fmt.Errorf("failed to add submodule %s: %w", modulePath, err)
			}
		} else if currentURL != dep.RepoURL {
			if err := submodule.SetSubmoduleURL(relPath, dep.RepoURL); err != nil {
				return fmt.Errorf("failed to update submodule URL for %s: %w", modulePath, err)
			}
		}

		if err := submodule.InitSubmodule(relPath); err != nil {
			// Ignore error if already initialized
		}

		if err := submodule.FetchTags(path); err != nil {
			return fmt.Errorf("failed to fetch tags for %s: %w", modulePath, err)
		}
		if err := submodule.FetchCommit(path); err != nil {
			return fmt.Errorf("failed to fetch commits for %s: %w", modulePath, err)
		}

		if err := submodule.Checkout(path, dep.Commit); err != nil {
			return fmt.Errorf("failed to checkout %s@%s: %w", modulePath, dep.Version, err)
		}
	}

	return nil
}
