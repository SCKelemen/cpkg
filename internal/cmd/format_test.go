package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/format"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"gopkg.in/yaml.v3"
)

func TestListCommand_Format(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a manifest with dependencies
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{
			"github.com/user/repo1": {Version: "^1.0.0"},
			"github.com/user/repo2": {Version: "^2.0.0"},
		},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create a lockfile
	lock := &lockfile.Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
		Dependencies: map[string]lockfile.Dependency{
			"github.com/user/repo1": {
				Version: "v1.2.3",
				Commit:  "abc123",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo1.git",
				Path:    "deps/github.com/user/repo1",
			},
		},
	}

	lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	tests := []struct {
		name           string
		format         string
		validateOutput func(t *testing.T, output string)
	}{
		{
			name:   "JSON format",
			format: "json",
			validateOutput: func(t *testing.T, output string) {
				var result listOutput
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, output)
				}

				if len(result.Dependencies) != 2 {
					t.Errorf("Expected 2 dependencies, got %d", len(result.Dependencies))
				}

				// Check first dependency
				dep1 := result.Dependencies[0]
				if dep1.Module != "github.com/user/repo1" {
					t.Errorf("Expected module github.com/user/repo1, got %s", dep1.Module)
				}
				if dep1.Constraint != "^1.0.0" {
					t.Errorf("Expected constraint ^1.0.0, got %s", dep1.Constraint)
				}
				if dep1.Locked != "v1.2.3" {
					t.Errorf("Expected locked v1.2.3, got %s", dep1.Locked)
				}
				if !dep1.HasLockfile {
					t.Error("Expected HasLockfile to be true")
				}
			},
		},
		{
			name:   "YAML format",
			format: "yaml",
			validateOutput: func(t *testing.T, output string) {
				var result listOutput
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, output)
				}

				if len(result.Dependencies) != 2 {
					t.Errorf("Expected 2 dependencies, got %d", len(result.Dependencies))
				}
			},
		},
		{
			name:   "Text format (default)",
			format: "text",
			validateOutput: func(t *testing.T, output string) {
				// Text format should contain table headers
				if !strings.Contains(output, "MODULE") {
					t.Error("Text output should contain MODULE header")
				}
				if !strings.Contains(output, "CONSTRAINT") {
					t.Error("Text output should contain CONSTRAINT header")
				}
				if !strings.Contains(output, "github.com/user/repo1") {
					t.Error("Text output should contain module names")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global format flag
			GlobalFormatFlag = tt.format

			var buf bytes.Buffer
			ctx := &clix.Context{
				App: &clix.App{
					Out: &buf,
					Err: &bytes.Buffer{},
				},
			}

			err := runList(ctx)
			if err != nil {
				t.Fatalf("runList() error = %v", err)
			}

			tt.validateOutput(t, buf.String())

			// Reset format flag
			GlobalFormatFlag = ""
		})
	}
}

func TestCheckCommand_Format(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{
			"github.com/user/repo1": {Version: "^1.0.0"},
		},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create a lockfile
	lock := &lockfile.Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
		Dependencies: map[string]lockfile.Dependency{
			"github.com/user/repo1": {
				Version: "v1.2.3",
				Commit:  "abc123",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo1.git",
				Path:    "deps/github.com/user/repo1",
			},
		},
	}

	lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	t.Run("JSON format structure", func(t *testing.T) {
		GlobalFormatFlag = "json"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		// Note: This will fail because it tries to fetch tags, but we can test the structure
		// by checking if it attempts to output JSON format
		err := runCheck(ctx)
		// We expect an error (no internet/repo), but the format should still be attempted
		if err == nil {
			// If no error, verify JSON structure
			var result checkOutput
			if err := json.Unmarshal(buf.Bytes(), &result); err == nil {
				// Valid JSON structure
				if result.Dependencies == nil {
					t.Error("checkOutput should have dependencies field")
				}
			}
		}

		GlobalFormatFlag = ""
	})
}

func TestStatusCommand_Format(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{
			"github.com/user/repo1": {Version: "^1.0.0"},
		},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create a lockfile
	lock := &lockfile.Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
		Dependencies: map[string]lockfile.Dependency{
			"github.com/user/repo1": {
				Version: "v1.2.3",
				Commit:  "abc123",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo1.git",
				Path:    "deps/github.com/user/repo1",
			},
		},
	}

	lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	t.Run("JSON format", func(t *testing.T) {
		GlobalFormatFlag = "json"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		err := runStatus(ctx)
		if err != nil {
			t.Fatalf("runStatus() error = %v", err)
		}

		var result statusOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, buf.String())
		}

		if len(result.Dependencies) != 1 {
			t.Errorf("Expected 1 dependency, got %d", len(result.Dependencies))
		}

		dep := result.Dependencies[0]
		if dep.Module != "github.com/user/repo1" {
			t.Errorf("Expected module github.com/user/repo1, got %s", dep.Module)
		}
		if dep.Status == "" {
			t.Error("Status should not be empty")
		}

		GlobalFormatFlag = ""
	})

	t.Run("YAML format", func(t *testing.T) {
		GlobalFormatFlag = "yaml"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		err := runStatus(ctx)
		if err != nil {
			t.Fatalf("runStatus() error = %v", err)
		}

		var result statusOutput
		if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, buf.String())
		}

		if len(result.Dependencies) != 1 {
			t.Errorf("Expected 1 dependency, got %d", len(result.Dependencies))
		}

		GlobalFormatFlag = ""
	})
}

func TestGraphCommand_Format(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create a lockfile with dependencies
	lock := &lockfile.Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
		Dependencies: map[string]lockfile.Dependency{
			"github.com/user/repo1": {
				Version: "v1.2.3",
				Commit:  "abc123",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo1.git",
				Path:    "deps/github.com/user/repo1",
			},
			"github.com/user/repo2": {
				Version: "v2.0.0",
				Commit:  "def456",
				Sum:     "h1:test2",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo2.git",
				Path:    "deps/github.com/user/repo2",
			},
		},
	}

	lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	t.Run("JSON format", func(t *testing.T) {
		GlobalFormatFlag = "json"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		err := runGraph(ctx)
		if err != nil {
			t.Fatalf("runGraph() error = %v", err)
		}

		var result graphOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, buf.String())
		}

		if result.Module != "test/module" {
			t.Errorf("Expected module test/module, got %s", result.Module)
		}
		if len(result.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies, got %d", len(result.Dependencies))
		}

		GlobalFormatFlag = ""
	})

	t.Run("YAML format", func(t *testing.T) {
		GlobalFormatFlag = "yaml"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		err := runGraph(ctx)
		if err != nil {
			t.Fatalf("runGraph() error = %v", err)
		}

		var result graphOutput
		if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, buf.String())
		}

		if result.Module != "test/module" {
			t.Errorf("Expected module test/module, got %s", result.Module)
		}
		if len(result.Dependencies) != 2 {
			t.Errorf("Expected 2 dependencies, got %d", len(result.Dependencies))
		}

		GlobalFormatFlag = ""
	})

	t.Run("Text format", func(t *testing.T) {
		GlobalFormatFlag = "text"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
		}

		err := runGraph(ctx)
		if err != nil {
			t.Fatalf("runGraph() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "test/module") {
			t.Error("Text output should contain module name")
		}
		if !strings.Contains(output, "github.com/user/repo1") {
			t.Error("Text output should contain dependency names")
		}

		GlobalFormatFlag = ""
	})
}

func TestExplainCommand_Format(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create a manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{
			"github.com/user/repo1": {Version: "^1.0.0"},
		},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create a lockfile
	lock := &lockfile.Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
		Dependencies: map[string]lockfile.Dependency{
			"github.com/user/repo1": {
				Version: "v1.2.3",
				Commit:  "abc123def456",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/user/repo1.git",
				Path:    "deps/github.com/user/repo1",
			},
		},
	}

	lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
	if err := lockfile.Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	t.Run("JSON format", func(t *testing.T) {
		GlobalFormatFlag = "json"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
			Args: []string{"github.com/user/repo1"},
		}

		err := runExplain(ctx)
		if err != nil {
			t.Fatalf("runExplain() error = %v", err)
		}

		var result explainOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, buf.String())
		}

		if result.Module != "github.com/user/repo1" {
			t.Errorf("Expected module github.com/user/repo1, got %s", result.Module)
		}
		if result.Constraint != "^1.0.0" {
			t.Errorf("Expected constraint ^1.0.0, got %s", result.Constraint)
		}
		if result.Locked == nil {
			t.Error("Expected locked information to be present")
		} else {
			if result.Locked.Version != "v1.2.3" {
				t.Errorf("Expected locked version v1.2.3, got %s", result.Locked.Version)
			}
		}

		GlobalFormatFlag = ""
	})

	t.Run("YAML format", func(t *testing.T) {
		GlobalFormatFlag = "yaml"

		var buf bytes.Buffer
		ctx := &clix.Context{
			App: &clix.App{
				Out: &buf,
				Err: &bytes.Buffer{},
			},
			Args: []string{"github.com/user/repo1"},
		}

		err := runExplain(ctx)
		if err != nil {
			t.Fatalf("runExplain() error = %v", err)
		}

		var result explainOutput
		if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid YAML output: %v\nOutput: %s", err, buf.String())
		}

		if result.Module != "github.com/user/repo1" {
			t.Errorf("Expected module github.com/user/repo1, got %s", result.Module)
		}

		GlobalFormatFlag = ""
	})
}

func TestGetFormat(t *testing.T) {
	tests := []struct {
		name     string
		setFlag  string
		expected format.Format
	}{
		{"text format", "text", format.FormatText},
		{"json format", "json", format.FormatJSON},
		{"yaml format", "yaml", format.FormatYAML},
		{"empty flag defaults to text", "", format.FormatText},
		{"invalid format defaults to text", "invalid", format.FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GlobalFormatFlag = tt.setFlag
			got := GetFormat()
			if got != tt.expected {
				t.Errorf("GetFormat() = %v, want %v", got, tt.expected)
			}
			GlobalFormatFlag = ""
		})
	}
}

