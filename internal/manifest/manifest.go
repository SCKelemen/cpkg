package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ManifestFileName = "cpkg.yaml"

type Manifest struct {
	APIVersion  string                 `yaml:"apiVersion"`
	Kind        string                 `yaml:"kind"`
	Module      string                 `yaml:"module"`
	Version     string                 `yaml:"version,omitempty"`
	DepRoot     string                 `yaml:"depRoot,omitempty"`
	Language    Language               `yaml:"language,omitempty"`
	Build       *Build                 `yaml:"build,omitempty"`
	Test        *Test                  `yaml:"test,omitempty"`
	Dependencies map[string]Dependency `yaml:"dependencies,omitempty"`
}

type Language struct {
	CStandard string `yaml:"cStandard,omitempty"`
	SKC       bool   `yaml:"skc,omitempty"`
}

type Build struct {
	Command []string            `yaml:"command,omitempty"`
	Targets map[string]BuildTarget `yaml:"targets,omitempty"`
}

type BuildTarget struct {
	Command []string `yaml:"command"`
}

type Test struct {
	Command []string `yaml:"command"`
}

type Dependency struct {
	Version string `yaml:"version"`
}

func DefaultDepRoot() string {
	return "third_party/cpkg"
}

func FindManifest(startDir string) (string, error) {
	dir := startDir
	for {
		manifestPath := filepath.Join(dir, ManifestFileName)
		if _, err := os.Stat(manifestPath); err == nil {
			return manifestPath, nil
		}
		if dir == "/" || dir == filepath.Dir(dir) {
			return "", fmt.Errorf("no %s found", ManifestFileName)
		}
		dir = filepath.Dir(dir)
	}
}

func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if m.DepRoot == "" {
		m.DepRoot = DefaultDepRoot()
	}

	return &m, nil
}

func Save(m *Manifest, path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}


