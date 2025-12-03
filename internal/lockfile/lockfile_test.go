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

func TestFindLockfileInParent(t *testing.T) {
	tmpDir := t.TempDir()
	lockfilePath := filepath.Join(tmpDir, LockfileName)

	lock := &Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/module",
	}
	if err := Save(lock, lockfilePath); err != nil {
		t.Fatalf("failed to save lockfile: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Should find lockfile in parent
	found, err := FindLockfile(subDir)
	if err != nil {
		t.Fatalf("failed to find lockfile: %v", err)
	}
	if found != lockfilePath {
		t.Errorf("expected %s, got %s", lockfilePath, found)
	}
}

func TestFindLockfileNestedModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root lockfile
	rootLockfilePath := filepath.Join(tmpDir, LockfileName)
	rootLock := &Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/root",
	}
	if err := Save(rootLock, rootLockfilePath); err != nil {
		t.Fatalf("failed to save root lockfile: %v", err)
	}

	// Create nested subdirectory with its own lockfile
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested subdir: %v", err)
	}

	nestedLockfilePath := filepath.Join(filepath.Dir(subDir), LockfileName)
	nestedLock := &Lockfile{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Lockfile",
		Module:     "test/nested",
	}
	if err := Save(nestedLock, nestedLockfilePath); err != nil {
		t.Fatalf("failed to save nested lockfile: %v", err)
	}

	// Should find the nearest lockfile (nested one), not the root one
	found, err := FindLockfile(subDir)
	if err != nil {
		t.Fatalf("failed to find lockfile: %v", err)
	}
	if found != nestedLockfilePath {
		t.Errorf("expected nested lockfile %s, got %s", nestedLockfilePath, found)
	}

	// From a directory between root and nested, should find nested (nearest)
	betweenDir := filepath.Join(tmpDir, "subdir")
	found, err = FindLockfile(betweenDir)
	if err != nil {
		t.Fatalf("failed to find lockfile: %v", err)
	}
	if found != nestedLockfilePath {
		t.Errorf("expected nested lockfile %s, got %s", nestedLockfilePath, found)
	}
}


