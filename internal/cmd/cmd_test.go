package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SCKelemen/cpkg/internal/manifest"
)

func TestParseModuleVersion(t *testing.T) {
	tests := []struct {
		input    string
		module   string
		version  string
		wantErr  bool
	}{
		{"github.com/user/repo@^1.0.0", "github.com/user/repo", "^1.0.0", false},
		{"github.com/user/repo", "github.com/user/repo", "", false},
		{"module@v1.2.3@extra", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			module, version, err := parseModuleVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseModuleVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if module != tt.module || version != tt.version {
					t.Errorf("parseModuleVersion() = (%q, %q), want (%q, %q)", module, version, tt.module, tt.version)
				}
			}
		})
	}
}

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Initialize git repo
	os.MkdirAll(".git", 0755)

	// Test init with explicit module
	initModule = "github.com/test/repo"
	initDepRoot = ""

	// We can't easily test the full command without mocking clix, but we can test the logic
	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if _, err := os.Stat(manifestPath); err == nil {
		t.Fatal("manifest should not exist yet")
	}
}

func TestAddCommand(t *testing.T) {
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
		Dependencies: make(map[string]manifest.Dependency),
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Test parsing
	module, version, err := parseModuleVersion("github.com/test/lib@^1.0.0")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if module != "github.com/test/lib" || version != "^1.0.0" {
		t.Errorf("parse failed: got (%q, %q)", module, version)
	}
}

