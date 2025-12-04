package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/format"
	"github.com/SCKelemen/cpkg/internal/git"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/modulepath"
	"github.com/SCKelemen/cpkg/internal/semver"
)

var checkCmd = clix.NewCommand("check",
	clix.WithCommandShort("Check for newer versions of dependencies"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runCheck(ctx)
	}),
)

type checkOutput struct {
	Dependencies []checkDependency `json:"dependencies" yaml:"dependencies"`
	AllUpToDate  bool              `json:"all_up_to_date" yaml:"all_up_to_date"`
}

type checkDependency struct {
	Module     string `json:"module" yaml:"module"`
	Current    string `json:"current" yaml:"current"`
	Latest     string `json:"latest" yaml:"latest"`
	Constraint string `json:"constraint" yaml:"constraint"`
	Notes      string `json:"notes" yaml:"notes"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

func runCheck(ctx *clix.Context) error {
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

	outputFormat := GetFormat()
	deps := make([]checkDependency, 0, len(m.Dependencies))
	hasUpdates := false

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
			deps = append(deps, checkDependency{
				Module:     modulePath,
				Current:    currentVersion,
				Latest:     "ERROR",
				Constraint: constraint,
				Error:      fmt.Sprintf("invalid module path: %v", err),
			})
			continue
		}

		repoURL := git.ModulePathToRepoURL(mp.RepoURL)
		allTags, err := git.LsRemoteTags(repoURL)
		if err != nil {
			deps = append(deps, checkDependency{
				Module:     modulePath,
				Current:    currentVersion,
				Latest:     "ERROR",
				Constraint: constraint,
				Error:      fmt.Sprintf("failed to fetch tags: %v", err),
			})
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

		notes := ""
		if latestV.Compare(currentV) > 0 {
			hasUpdates = true
			// Determine update type
			if latestV.Major > currentV.Major {
				notes = "major available"
			} else if latestV.Minor > currentV.Minor {
				notes = "minor available"
			} else {
				notes = "patch available"
			}
		} else {
			notes = "up to date"
		}

		deps = append(deps, checkDependency{
			Module:     modulePath,
			Current:    currentVersion,
			Latest:     latestCompatible,
			Constraint: constraint,
			Notes:      notes,
		})
	}

	// Output in requested format
	if outputFormat != format.FormatText {
		output := checkOutput{
			Dependencies: deps,
			AllUpToDate:  !hasUpdates,
		}
		return format.Write(ctx.App.Out, outputFormat, output)
	}

	// Text output
	fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %s\n", "MODULE", "CURRENT", "LATEST", "CONSTRAINT", "NOTES")
	fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %s\n",
		"──────────────────────────────────────────────────",
		"───────────────", "───────────────", "───────────────", "─────")

	for _, dep := range deps {
		if dep.Error != "" {
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %s\n",
				dep.Module, dep.Current, dep.Latest, dep.Constraint, dep.Error)
		} else {
			fmt.Fprintf(ctx.App.Out, "%-50s %-15s %-15s %-15s %s\n",
				dep.Module, dep.Current, dep.Latest, dep.Constraint, dep.Notes)
		}
	}

	if !hasUpdates {
		fmt.Fprintf(ctx.App.Out, "\nAll dependencies are up to date.\n")
	}

	return nil
}
