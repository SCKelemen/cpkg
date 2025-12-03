package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/git"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/modulepath"
	"github.com/SCKelemen/cpkg/internal/semver"
)

var (
	upgradeAll    bool
	upgradeDepRoot string
)

var upgradeCmd = clix.NewCommand("upgrade",
	clix.WithCommandShort("Upgrade dependencies to latest compatible versions"),
	clix.WithCommandLong("Upgrade dependencies to the latest versions that satisfy their constraints, then run tidy and sync"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runUpgrade(ctx)
	}),
)

func init() {
	upgradeCmd.Flags = clix.NewFlagSet("upgrade")
	upgradeCmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "all",
			Usage: "Upgrade all dependencies (even if no updates available)",
		},
		Value: &upgradeAll,
	})
	upgradeCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Override dependency root",
		},
		Value: &upgradeDepRoot,
	})
}

func runUpgrade(ctx *clix.Context) error {
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
		return fmt.Errorf("no lockfile found, run 'cpkg tidy' first: %w", err)
	}

	upgraded := false
	updates := make(map[string]string) // module -> new version

	// Check each dependency for updates
	for modulePath, dep := range m.Dependencies {
		lockDep, exists := lock.Dependencies[modulePath]
		if !exists {
			continue
		}

		currentVersion := lockDep.Version
		constraint := dep.Version

		// Parse module path to extract repo URL and subpath
		mp, err := modulepath.ParseModulePath(modulePath)
		if err != nil {
			fmt.Fprintf(ctx.App.Err, "Warning: invalid module path %s: %v\n", modulePath, err)
			continue
		}

		repoURL := git.ModulePathToRepoURL(mp.RepoURL)
		allTags, err := git.LsRemoteTags(repoURL)
		if err != nil {
			fmt.Fprintf(ctx.App.Err, "Warning: failed to fetch tags for %s: %v\n", modulePath, err)
			continue
		}

		// Filter tags for this subpath
		tags := modulepath.FilterTagsForSubpath(allTags, mp.Subpath)

		// Extract version from tags
		var versionTags []string
		for _, tag := range tags {
			version, err := modulepath.ExtractVersionFromTag(tag, mp.Subpath)
			if err != nil {
				continue
			}
			versionTags = append(versionTags, version)
		}

		// Find latest compatible version
		latestCompatible, err := findCompatibleVersion(versionTags, constraint)
		if err != nil {
			fmt.Fprintf(ctx.App.Err, "Warning: no compatible version found for %s: %v\n", modulePath, err)
			continue
		}

		// Parse versions for comparison
		currentV, err := semver.Parse(currentVersion)
		if err != nil {
			continue
		}

		latestV, err := semver.Parse(latestCompatible)
		if err != nil {
			continue
		}

		if latestV.Compare(currentV) > 0 {
			updates[modulePath] = latestCompatible
			upgraded = true
			fmt.Fprintf(ctx.App.Out, "Upgrading %s: %s → %s\n", modulePath, currentVersion, latestCompatible)
		} else if upgradeAll {
			updates[modulePath] = latestCompatible
			fmt.Fprintf(ctx.App.Out, "Refreshing %s: %s\n", modulePath, latestCompatible)
		}
	}

	if !upgraded && !upgradeAll {
		fmt.Fprintf(ctx.App.Out, "All dependencies are up to date.\n")
		return nil
	}

	// Run tidy to update lockfile
	fmt.Fprintf(ctx.App.Out, "\nResolving dependencies...\n")
	depRoot := upgradeDepRoot
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = m.DepRoot
		}
	}

	if err := runTidyInternal(cwd, depRoot, false); err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Run sync to update submodules
	fmt.Fprintf(ctx.App.Out, "Syncing submodules...\n")
	if err := runSyncInternal(cwd, depRoot); err != nil {
		return fmt.Errorf("failed to sync submodules: %w", err)
	}

	fmt.Fprintf(ctx.App.Out, "Upgrade complete.\n")
	return nil
}

