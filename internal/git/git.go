package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func LsRemoteTags(repoURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", repoURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote tags: %w", err)
	}

	var tags []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		if strings.HasPrefix(ref, "refs/tags/") {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			// Remove ^{} suffix from annotated tags
			tag = strings.TrimSuffix(tag, "^{}")
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func GetCommitForTag(repoURL, tag string) (string, error) {
	cmd := exec.Command("git", "ls-remote", repoURL, tag)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit for tag: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no commit found for tag %s", tag)
	}

	parts := strings.Fields(lines[0])
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid output from git ls-remote")
	}

	return parts[0], nil
}

func ComputeTreeHash(repoURL, commit string) (string, error) {
	// For v0, we use a simplified checksum based on the commit SHA
	// In a full implementation, we'd want to compute a proper tree hash
	// by cloning the repo and computing a hash of the tree contents
	// For now, we use the commit SHA as the checksum (truncated for readability)
	if len(commit) < 16 {
		return "", fmt.Errorf("invalid commit SHA: %s", commit)
	}
	return "h1:" + commit[:16], nil
}

func ModulePathToRepoURL(modulePath string) string {
	// Default: https://<module>.git
	if strings.HasPrefix(modulePath, "http://") || strings.HasPrefix(modulePath, "https://") || strings.HasPrefix(modulePath, "git@") {
		return modulePath
	}
	return "https://" + modulePath + ".git"
}

// ExtractRepoURLFromModulePath extracts the repository URL from a module path,
// handling subpaths. For example:
//   - github.com/user/repo -> github.com/user/repo
//   - github.com/user/repo/subpath -> github.com/user/repo
func ExtractRepoURLFromModulePath(modulePath string) string {
	// Use the modulepath package to parse
	// For now, simple heuristic: take first 3 parts
	parts := strings.Split(modulePath, "/")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], "/")
	}
	return modulePath
}
