package git

import (
	"testing"
)

func TestModulePathToRepoURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com/user/repo", "https://github.com/user/repo.git"},
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git"},
		{"http://example.com/repo.git", "http://example.com/repo.git"},
		{"git@github.com:user/repo.git", "git@github.com:user/repo.git"},
		{"git.internal/user/repo", "https://git.internal/user/repo.git"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ModulePathToRepoURL(tt.input)
			if got != tt.expected {
				t.Errorf("ModulePathToRepoURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInferModulePath(t *testing.T) {
	// This test requires a git repo, so we'll skip it in unit tests
	// Integration tests can cover this
	t.Skip("requires git repository")
}

