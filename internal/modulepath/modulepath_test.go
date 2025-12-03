package modulepath

import (
	"testing"
)

func TestParseModulePath(t *testing.T) {
	tests := []struct {
		input    string
		repoURL  string
		subpath  string
		hasError bool
	}{
		{"github.com/user/repo", "github.com/user/repo", "", false},
		{"github.com/user/repo/subpath", "github.com/user/repo", "subpath", false},
		{"github.com/user/repo/path/to/module", "github.com/user/repo", "path/to/module", false},
		{"git.internal/user/repo", "git.internal/user/repo", "", false},
		{"git.internal/user/repo/subpath", "git.internal/user/repo", "subpath", false},
		{"https://github.com/user/repo.git", "https://github.com/user/repo.git", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mp, err := ParseModulePath(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if mp.RepoURL != tt.repoURL {
				t.Errorf("repoURL: got %q, want %q", mp.RepoURL, tt.repoURL)
			}
			if mp.Subpath != tt.subpath {
				t.Errorf("subpath: got %q, want %q", mp.Subpath, tt.subpath)
			}
		})
	}
}

func TestFilterTagsForSubpath(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		subpath string
		want    []string
	}{
		{
			name:    "no subpath",
			tags:    []string{"v1.0.0", "v1.1.0", "v2.0.0"},
			subpath: "",
			want:    []string{"v1.0.0", "v1.1.0", "v2.0.0"},
		},
		{
			name:    "prefix format",
			tags:    []string{"intrusive_list/v1.0.0", "intrusive_list/v1.1.0", "span/v1.0.0", "v1.0.0"},
			subpath: "intrusive_list",
			want:    []string{"intrusive_list/v1.0.0", "intrusive_list/v1.1.0"},
		},
		{
			name:    "suffix format",
			tags:    []string{"v1.0.0-intrusive_list", "v1.1.0-intrusive_list", "v1.0.0-span", "v1.0.0"},
			subpath: "intrusive_list",
			want:    []string{"v1.0.0-intrusive_list", "v1.1.0-intrusive_list"},
		},
		{
			name:    "mixed formats",
			tags:    []string{"intrusive_list/v1.0.0", "v1.1.0-intrusive_list", "span/v1.0.0"},
			subpath: "intrusive_list",
			want:    []string{"intrusive_list/v1.0.0", "v1.1.0-intrusive_list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterTagsForSubpath(tt.tags, tt.subpath)
			if len(got) != len(tt.want) {
				t.Errorf("got %d tags, want %d", len(got), len(tt.want))
				return
			}
			for i, tag := range got {
				if tag != tt.want[i] {
					t.Errorf("tag[%d]: got %q, want %q", i, tag, tt.want[i])
				}
			}
		})
	}
}

func TestExtractVersionFromTag(t *testing.T) {
	tests := []struct {
		tag     string
		subpath string
		want    string
		hasError bool
	}{
		{"v1.0.0", "", "v1.0.0", false},
		{"intrusive_list/v1.0.0", "intrusive_list", "v1.0.0", false},
		{"v1.0.0-intrusive_list", "intrusive_list", "v1.0.0", false},
		{"intrusive_list/v1.1.0", "intrusive_list", "v1.1.0", false},
		{"v1.1.0-intrusive_list", "intrusive_list", "v1.1.0", false},
		{"intrusive_list/v1.0.0", "span", "", true}, // Wrong subpath
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got, err := ExtractVersionFromTag(tt.tag, tt.subpath)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

