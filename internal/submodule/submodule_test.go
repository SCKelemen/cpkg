package submodule

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSubmoduleExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test: no .gitmodules
	if SubmoduleExists("some/path") {
		t.Error("expected false when .gitmodules doesn't exist")
	}

	// Create .gitmodules
	gitmodulesPath := filepath.Join(tmpDir, ".gitmodules")
	content := `[submodule "test/path"]
	path = test/path
	url = https://github.com/test/repo.git
`
	if err := os.WriteFile(gitmodulesPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .gitmodules: %v", err)
	}

	// Change to tmpDir to test
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Test: submodule exists
	if !SubmoduleExists("test/path") {
		t.Error("expected true when submodule exists in .gitmodules")
	}

	// Test: submodule doesn't exist
	if SubmoduleExists("other/path") {
		t.Error("expected false when submodule doesn't exist")
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "deep", "nested", "path", "file.txt")

	if err := EnsureDir(testPath); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

