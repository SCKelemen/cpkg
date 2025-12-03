package modulepath

import (
	"fmt"
	"strings"
)

// ModulePath represents a parsed module path with repo URL and optional subpath
type ModulePath struct {
	RepoURL string // The repository URL (e.g., github.com/user/repo)
	Subpath string // The subpath within the repo (e.g., intrusive_list)
	Full    string // The full module path
}

// ParseModulePath parses a module path into repo URL and subpath.
// Examples:
//   - github.com/user/repo -> repo: github.com/user/repo, subpath: ""
//   - github.com/user/repo/subpath -> repo: github.com/user/repo, subpath: subpath
//   - github.com/user/repo/path/to/module -> repo: github.com/user/repo, subpath: path/to/module
func ParseModulePath(path string) (ModulePath, error) {
	mp := ModulePath{
		Full: path,
	}

	// Handle direct URLs
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		// For direct URLs, we need to extract the repo part
		// This is more complex - for now, we'll assume the entire URL is the repo
		// and there's no subpath support for direct URLs
		mp.RepoURL = path
		mp.Subpath = ""
		return mp, nil
	}

	// Split by '/'
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return mp, fmt.Errorf("invalid module path: must have at least host/owner/repo")
	}

	// The repo part is at least: host/owner/repo (3 parts minimum)
	// Everything after is the subpath
	// However, we need to be smart about this:
	// - github.com/user/repo -> repo is 3 parts
	// - git.internal/user/repo -> repo is 3 parts
	// - github.com/user/repo/subpath -> repo is 3 parts, subpath is 1 part
	// - github.com/user/repo/path/to/module -> repo is 3 parts, subpath is 3 parts

	// Heuristic: if we have 3+ parts, the first 3 are the repo
	// If we have more, the rest is the subpath
	if len(parts) >= 3 {
		mp.RepoURL = strings.Join(parts[:3], "/")
		if len(parts) > 3 {
			mp.Subpath = strings.Join(parts[3:], "/")
		}
	} else {
		// Less than 3 parts - treat entire path as repo
		mp.RepoURL = path
		mp.Subpath = ""
	}

	return mp, nil
}

// FilterTagsForSubpath filters tags that match the subpath.
// Supports two formats:
// 1. Prefix format: subpath/v1.0.0
// 2. Suffix format: v1.0.0-subpath
func FilterTagsForSubpath(tags []string, subpath string) []string {
	if subpath == "" {
		// No subpath - return all tags (but filter out tags with subpaths)
		var filtered []string
		for _, tag := range tags {
			// Skip tags that look like they have subpaths
			if strings.Contains(tag, "/") && !strings.HasPrefix(tag, "v") {
				// Likely a subpath tag, skip it
				continue
			}
			if strings.Contains(tag, "-") && strings.HasPrefix(tag, "v") {
				// Could be a suffix-format subpath tag, skip it
				// But we need to be careful - pre-release versions also have dashes
				// For now, we'll include it and let semver parsing handle it
			}
			filtered = append(filtered, tag)
		}
		return filtered
	}

	var filtered []string
	for _, tag := range tags {
		// Try prefix format: subpath/v1.0.0
		if strings.HasPrefix(tag, subpath+"/") {
			filtered = append(filtered, tag)
			continue
		}

		// Try suffix format: v1.0.0-subpath
		// But be careful - we need to match exactly, not just any dash
		// v1.0.0-subpath should match, but v1.0.0-alpha should not (unless subpath is "alpha")
		if strings.HasSuffix(tag, "-"+subpath) {
			filtered = append(filtered, tag)
			continue
		}
	}

	return filtered
}

// ExtractVersionFromTag extracts the version part from a tag that includes a subpath.
// Returns the version string (e.g., "v1.0.0") and the original tag if no subpath.
func ExtractVersionFromTag(tag, subpath string) (string, error) {
	if subpath == "" {
		// No subpath - return tag as-is
		return tag, nil
	}

	// Try prefix format: subpath/v1.0.0 -> v1.0.0
	if strings.HasPrefix(tag, subpath+"/") {
		version := strings.TrimPrefix(tag, subpath+"/")
		return version, nil
	}

	// Try suffix format: v1.0.0-subpath -> v1.0.0
	if strings.HasSuffix(tag, "-"+subpath) {
		version := strings.TrimSuffix(tag, "-"+subpath)
		return version, nil
	}

	// Tag doesn't match subpath format - this shouldn't happen if filtering worked correctly
	return "", fmt.Errorf("tag %s does not match subpath format for %s", tag, subpath)
}

