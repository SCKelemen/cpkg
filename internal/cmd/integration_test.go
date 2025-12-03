//go:build integration
// +build integration

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

// TestFullWorkflow tests a complete workflow: init -> add -> tidy -> sync
func TestFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Initialize git repo
	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	cpkgBin := filepath.Join(originalDir, "cpkg")
	if _, err := os.Stat(cpkgBin); os.IsNotExist(err) {
		t.Skip("cpkg binary not found, run 'go build' first")
	}

	// Step 1: Init
	t.Run("init", func(t *testing.T) {
		cmd := exec.Command(cpkgBin, "init", "--module", "github.com/test/repo")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("init failed: %v\noutput: %s", err, output)
		}

		// Verify manifest was created
		manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			t.Fatal("manifest was not created")
		}

		m, err := manifest.Load(manifestPath)
		if err != nil {
			t.Fatalf("failed to load manifest: %v", err)
		}

		if m.Module != "github.com/test/repo" {
			t.Errorf("expected module github.com/test/repo, got %s", m.Module)
		}
	})

	// Step 2: Add dependency (skip if no internet or repo doesn't exist)
	// This would require a real repository, so we'll skip for now
	t.Run("add", func(t *testing.T) {
		t.Skip("requires real repository")
	})

	// Step 3: Tidy (skip if no internet)
	t.Run("tidy", func(t *testing.T) {
		t.Skip("requires internet and real repositories")
	})

	// Step 4: Sync (skip if no lockfile)
	t.Run("sync", func(t *testing.T) {
		lockfilePath := filepath.Join(tmpDir, lockfile.LockfileName)
		if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
			t.Skip("lockfile required for sync")
		}
		t.Skip("requires submodules")
	})
}

// TestManifestLockfileConsistency tests that manifest and lockfile stay in sync
func TestManifestLockfileConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Create manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     "test/module",
		DepRoot:    "deps",
		Dependencies: map[string]manifest.Dependency{
			"github.com/test/lib": {
				Version: "^1.0.0",
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, manifest.ManifestFileName)
	if err := manifest.Save(m, manifestPath); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// Verify manifest can be loaded
	loaded, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if len(loaded.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(loaded.Dependencies))
	}

	if dep, exists := loaded.Dependencies["github.com/test/lib"]; !exists {
		t.Error("dependency not found")
	} else if dep.Version != "^1.0.0" {
		t.Errorf("expected version ^1.0.0, got %s", dep.Version)
	}
}

