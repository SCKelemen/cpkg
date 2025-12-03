package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetGitRemoteURL(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository or no origin remote: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func InferModulePath(dir string) (string, error) {
	url, err := GetGitRemoteURL(dir)
	if err != nil {
		return "", err
	}

	// Handle various git URL formats
	// https://github.com/user/repo.git -> github.com/user/repo
	// git@github.com:user/repo.git -> github.com/user/repo
	// https://git.internal/user/repo.git -> git.internal/user/repo

	url = strings.TrimSuffix(url, ".git")
	
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "git@") {
		// git@github.com:user/repo -> github.com/user/repo
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
	}

	return url, nil
}

func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil || filepath.IsAbs(gitDir)
}

