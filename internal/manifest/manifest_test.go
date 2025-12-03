package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Test: no manifest exists
	_, err := FindManifest(tmpDir)
	if err == nil {
		t.Error("expected error when manifest doesn't exist")
	}

	// Test: manifest exists
	manifestPath := filepath.Join(tmpDir, ManifestFileName)
	m := &Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
	}
	if err := Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	found, err := FindManifest(tmpDir)
	if err != nil {
		t.Fatalf("failed to find manifest: %v", err)
	}
	if found != manifestPath {
		t.Errorf("expected %s, got %s", manifestPath, found)
	}
}

func TestLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	m := &Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]Dependency{
			"github.com/test/lib": {
				Version: "^1.0.0",
			},
		},
	}

	if err := Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Module != m.Module {
		t.Errorf("expected module %s, got %s", m.Module, loaded.Module)
	}
	if loaded.DepRoot != m.DepRoot {
		t.Errorf("expected depRoot %s, got %s", m.DepRoot, loaded.DepRoot)
	}
	if len(loaded.Dependencies) != len(m.Dependencies) {
		t.Errorf("expected %d dependencies, got %d", len(m.Dependencies), len(loaded.Dependencies))
	}
}

func TestDefaultDepRoot(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	m := &Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		// DepRoot not set
	}

	if err := Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.DepRoot != DefaultDepRoot() {
		t.Errorf("expected default depRoot %s, got %s", DefaultDepRoot(), loaded.DepRoot)
	}
}

func TestFindManifestInParent(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ManifestFileName)

	m := &Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
	}
	if err := Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Should find manifest in parent
	found, err := FindManifest(subDir)
	if err != nil {
		t.Fatalf("failed to find manifest: %v", err)
	}
	if found != manifestPath {
		t.Errorf("expected %s, got %s", manifestPath, found)
	}
}

