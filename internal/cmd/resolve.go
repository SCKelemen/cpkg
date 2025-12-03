package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SCKelemen/cpkg/internal/git"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/modulepath"
	"github.com/SCKelemen/cpkg/internal/semver"
)

func resolveDependencies(m *manifest.Manifest, depRoot string) (*lockfile.Lockfile, error) {
	lock := &lockfile.Lockfile{
		APIVersion:  "cpkg.ringil.dev/v0",
		Kind:        "Lockfile",
		Module:      m.Module,
		GeneratedBy: "cpkg 0.1.0",
		DepRoot:     depRoot,
		Dependencies: make(map[string]lockfile.Dependency),
	}

	for modulePath, dep := range m.Dependencies {
		// Parse module path to extract repo URL and subpath
		mp, err := modulepath.ParseModulePath(modulePath)
		if err != nil {
			return nil, fmt.Errorf("invalid module path %s: %w", modulePath, err)
		}

		repoURL := git.ModulePathToRepoURL(mp.RepoURL)

		// Fetch tags
		allTags, err := git.LsRemoteTags(repoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags for %s: %w", modulePath, err)
		}

		// Filter tags for this subpath
		tags := modulepath.FilterTagsForSubpath(allTags, mp.Subpath)

		// Extract version from tags (remove subpath prefix/suffix)
		var versionTags []string
		for _, tag := range tags {
			version, parseErr := modulepath.ExtractVersionFromTag(tag, mp.Subpath)
			if parseErr != nil {
				continue // Skip tags that don't match format
			}
			versionTags = append(versionTags, version)
		}

		// If no subpath-specific tags found, fall back to root repo tags
		// This allows using arbitrary subdirectories from any repo
		var selectedVersion string
		var selectedTag string
		
		if len(versionTags) == 0 && mp.Subpath != "" {
			// No subpath tags - use root repo tags and point to subdirectory
			var rootVersionTags []string
			for _, tag := range allTags {
				// Filter out tags that look like they have subpaths
				if strings.Contains(tag, "/") && !strings.HasPrefix(tag, "v") {
					continue // Likely a subpath tag
				}
				if _, parseErr := semver.Parse(tag); parseErr == nil {
					rootVersionTags = append(rootVersionTags, tag)
				}
			}
			
			var resolveErr error
			selectedVersion, resolveErr = findCompatibleVersion(rootVersionTags, dep.Version)
			if resolveErr != nil {
				return nil, fmt.Errorf("no compatible version found for %s (constraint: %s): %w", modulePath, dep.Version, resolveErr)
			}
			selectedTag = selectedVersion // Root tag, no subpath prefix
		} else {
			// Find compatible version from subpath tags
			var resolveErr error
			selectedVersion, resolveErr = findCompatibleVersion(versionTags, dep.Version)
			if resolveErr != nil {
				return nil, fmt.Errorf("no compatible version found for %s (constraint: %s): %w", modulePath, dep.Version, resolveErr)
			}

			// Map back to the original tag format
			selectedTag = selectedVersion
			for _, tag := range tags {
				version, parseErr := modulepath.ExtractVersionFromTag(tag, mp.Subpath)
				if parseErr == nil && version == selectedVersion {
					selectedTag = tag
					break
				}
			}
		}

		// Get commit for tag
		commit, err := git.GetCommitForTag(repoURL, selectedTag)
		if err != nil {
			return nil, fmt.Errorf("failed to get commit for %s@%s: %w", modulePath, selectedTag, err)
		}

		// Compute checksum
		sum, err := git.ComputeTreeHash(repoURL, commit)
		if err != nil {
			return nil, fmt.Errorf("failed to compute checksum for %s: %w", modulePath, err)
		}

		path := filepath.Join(depRoot, modulePath)
		
		// Compute the actual source path (where the .c/.h files are)
		sourcePath := path
		if mp.Subpath != "" {
			sourcePath = filepath.Join(path, mp.Subpath)
		}

		lockDep := lockfile.Dependency{
			Version:    selectedVersion, // Store the version part (without subpath)
			Commit:     commit,
			Sum:        sum,
			VCS:        "git",
			RepoURL:    repoURL,
			Path:       path,        // Submodule path (entire repo checkout)
			Subdir:     mp.Subpath,  // Store the subdirectory within the repo
			SourcePath: sourcePath,  // Actual path to source files
		}

		lock.Dependencies[modulePath] = lockDep
	}

	return lock, nil
}

func findCompatibleVersion(tags []string, constraint string) (string, error) {
	var compatibleVersions []*semver.Version

	for _, tag := range tags {
		// Remove 'v' prefix if present for parsing
		v, err := semver.Parse(tag)
		if err != nil {
			continue // Skip invalid versions
		}

		satisfies, err := v.Satisfies(constraint)
		if err != nil {
			continue
		}
		if satisfies {
			compatibleVersions = append(compatibleVersions, v)
		}
	}

	if len(compatibleVersions) == 0 {
		return "", fmt.Errorf("no compatible version found")
	}

	// Sort descending (highest first)
	sort.Slice(compatibleVersions, func(i, j int) bool {
		return compatibleVersions[i].Compare(compatibleVersions[j]) > 0
	})

	// Return the highest compatible version with 'v' prefix
	selected := compatibleVersions[0]
	selectedStr := selected.String()
	// Find the original tag that matches
	for _, tag := range tags {
		v, err := semver.Parse(tag)
		if err == nil && v.Compare(selected) == 0 {
			return tag, nil
		}
	}
	return "v" + selectedStr, nil
}

