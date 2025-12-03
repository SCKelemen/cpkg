package lockfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, LockfileName)

	lock := &Lockfile{
		APIVersion:  "cpkg.ringil.dev/v0",
		Kind:        "Lockfile",
		Module:      "test/module",
		GeneratedBy: "cpkg 0.1.0",
		DepRoot:     "deps",
		Dependencies: map[string]Dependency{
			"github.com/test/lib": {
				Version: "v1.0.0",
				Commit:  "abc123",
				Sum:     "h1:test",
				VCS:     "git",
				RepoURL: "https://github.com/test/lib.git",
				Path:    "deps/github.com/test/lib",
			},
		},
	}

	if err := Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		t.Fatal("lockfile was not created")
	}

	loaded, err := Load(lockfilePath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Module != lock.Module {
		t.Errorf("expected module %s, got %s", lock.Module, loaded.Module)
	}
	if len(loaded.Dependencies) != len(lock.Dependencies) {
		t.Errorf("expected %d dependencies, got %d", len(lock.Dependencies), len(loaded.Dependencies))
	}
}

func TestFindLockfile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test: no lockfile exists
	_, err := FindLockfile(tmpDir)
	if err == nil {
		t.Error("expected error when lockfile doesn't exist")
	}

	// Test: lockfile exists
	lockfilePath := filepath.Join(tmpDir, LockfileName)
	lock := &Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
	}
	if err := Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	found, err := FindLockfile(tmpDir)
	if err != nil {
		t.Fatalf("failed to find lockfile: %v", err)
	}
	if found != lockfilePath {
		t.Errorf("expected %s, got %s", lockfilePath, found)
	}
}

