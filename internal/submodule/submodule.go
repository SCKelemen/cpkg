package submodule

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AddSubmodule(repoURL, path string) error {
	cmd := exec.Command("git", "submodule", "add", repoURL, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add submodule: %w\noutput: %s", err, string(output))
	}
	return nil
}

func SetSubmoduleURL(path, repoURL string) error {
	cmd := exec.Command("git", "submodule", "set-url", path, repoURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set submodule URL: %w\noutput: %s", err, string(output))
	}
	return nil
}

func InitSubmodule(path string) error {
	cmd := exec.Command("git", "submodule", "init", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to init submodule: %w\noutput: %s", err, string(output))
	}
	return nil
}

func FetchTags(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "--tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w\noutput: %s", err, string(output))
	}
	return nil
}

func FetchCommit(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "origin")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch commits: %w\noutput: %s", err, string(output))
	}
	return nil
}

func Checkout(path, commit string) error {
	cmd := exec.Command("git", "-C", path, "checkout", commit)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout commit: %w\noutput: %s", err, string(output))
	}
	return nil
}

func GetSubmoduleURL(path string) (string, error) {
	cmd := exec.Command("git", "config", "--file", ".gitmodules", "--get", fmt.Sprintf("submodule.%s.url", path))
	output, err := cmd.Output()
	if err != nil {
		return "", nil // Submodule not in .gitmodules
	}
	return strings.TrimSpace(string(output)), nil
}

func SubmoduleExists(path string) bool {
	gitmodulesPath := ".gitmodules"
	if _, err := os.Stat(gitmodulesPath); os.IsNotExist(err) {
		return false
	}

	cmd := exec.Command("git", "config", "--file", gitmodulesPath, "--get", fmt.Sprintf("submodule.%s.url", path))
	err := cmd.Run()
	return err == nil
}

func GetSubmoduleCommit(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get submodule commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func IsSubmoduleDirty(path string) (bool, error) {
	cmd := exec.Command("git", "-C", path, "diff", "--quiet")
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
		return true, nil
	}
	return false, err
}

func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}
