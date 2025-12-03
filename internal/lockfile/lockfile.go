package lockfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const LockfileName = "cpkg.lock.yaml"

type Lockfile struct {
	APIVersion   string                `yaml:"apiVersion"`
	Kind         string                `yaml:"kind"`
	Module       string                `yaml:"module"`
	GeneratedBy  string                `yaml:"generatedBy"`
	GeneratedAt  string                `yaml:"generatedAt"`
	DepRoot      string                `yaml:"depRoot"`
	Dependencies map[string]Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Version    string `yaml:"version"`
	Commit     string `yaml:"commit"`
	Sum        string `yaml:"sum"`
	VCS        string `yaml:"vcs"`
	RepoURL    string `yaml:"repoURL"`
	Path       string `yaml:"path"`             // Submodule path (entire repo checkout)
	Subdir     string `yaml:"subdir,omitempty"` // Subdirectory within the repo (e.g., "intrusive_list", "span")
	SourcePath string `yaml:"sourcePath"`       // Actual path to source files (path + subdir if subdir exists)
}

func FindLockfile(startDir string) (string, error) {
	dir := startDir
	for {
		lockfilePath := filepath.Join(dir, LockfileName)
		if _, err := os.Stat(lockfilePath); err == nil {
			return lockfilePath, nil
		}
		if dir == "/" || dir == filepath.Dir(dir) {
			return "", fmt.Errorf("no %s found", LockfileName)
		}
		dir = filepath.Dir(dir)
	}
}

func Load(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var l Lockfile
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("failed to parse lockfile: %w", err)
	}

	return &l, nil
}

func Save(l *Lockfile, path string) error {
	l.GeneratedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal lockfile: %w", err)
	}

	// Write atomically
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename lockfile: %w", err)
	}

	return nil
}
