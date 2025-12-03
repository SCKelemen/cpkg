package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
		wantErr  bool
	}{
		{"1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"v1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"1.2.3-alpha", &Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"}, false},
		{"v5.8.0-stable", &Version{Major: 5, Minor: 8, Patch: 0, Pre: "stable"}, false},
		{"1.2.3+build", &Version{Major: 1, Minor: 2, Patch: 3, Build: "build"}, false},
		{"invalid", nil, true},
		{"1.2", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.expected.Major || got.Minor != tt.expected.Minor || got.Patch != tt.expected.Patch {
					t.Errorf("Parse() = %+v, want %+v", got, tt.expected)
				}
				if got.Pre != tt.expected.Pre {
					t.Errorf("Parse() Pre = %q, want %q", got.Pre, tt.expected.Pre)
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.3.0", "1.2.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"5.8.0-stable", "5.8.0", -1}, // pre-release is less than release
		{"5.8.0", "5.8.0-stable", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			v1, _ := Parse(tt.v1)
			v2, _ := Parse(tt.v2)
			if got := v1.Compare(v2); got != tt.expected {
				t.Errorf("Compare() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		version  string
		constraint string
		expected bool
	}{
		{"1.2.3", "^1.0.0", true},
		{"1.2.3", "^1.2.0", true},
		{"2.0.0", "^1.0.0", false},
		{"1.2.3", "~1.2.0", true},
		{"1.3.0", "~1.2.0", false},
		{"1.2.3", "1.2.3", true},
		{"1.2.4", "1.2.3", false},
		{"1.2.3", "", true}, // empty constraint always satisfies
		{"5.8.0-stable", "^5.8.0", true}, // pre-release should satisfy if base version matches
		{"5.8.4-stable", "^5.8.0", true},
		{"5.9.0", "^5.8.0", true},
		{"6.0.0", "^5.8.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version+" "+tt.constraint, func(t *testing.T) {
			v, _ := Parse(tt.version)
			got, err := v.Satisfies(tt.constraint)
			if err != nil {
				t.Errorf("Satisfies() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("Satisfies() = %v, want %v", got, tt.expected)
			}
		})
	}
}
