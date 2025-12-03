package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	Build string
}

var versionRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([\w\.-]+))?(?:\+([\w\.-]+))?$`)

func Parse(version string) (*Version, error) {
	version = strings.TrimPrefix(version, "v")
	matches := versionRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	v := &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	if matches[4] != "" {
		v.Pre = matches[4]
	}
	if matches[5] != "" {
		v.Build = matches[5]
	}

	return v, nil
}

func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch - other.Patch
	}
	if v.Pre != "" && other.Pre == "" {
		return -1
	}
	if v.Pre == "" && other.Pre != "" {
		return 1
	}
	if v.Pre != "" && other.Pre != "" {
		return strings.Compare(v.Pre, other.Pre)
	}
	return 0
}

func (v *Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

func (v *Version) Satisfies(constraint string) (bool, error) {
	if constraint == "" {
		return true, nil
	}

	constraint = strings.TrimSpace(constraint)

	// Handle caret (^) constraint
	if strings.HasPrefix(constraint, "^") {
		baseVersion, err := Parse(strings.TrimPrefix(constraint, "^"))
		if err != nil {
			return false, err
		}
		return satisfiesCaret(v, baseVersion), nil
	}

	// Handle tilde (~) constraint
	if strings.HasPrefix(constraint, "~") {
		baseVersion, err := Parse(strings.TrimPrefix(constraint, "~"))
		if err != nil {
			return false, err
		}
		return satisfiesTilde(v, baseVersion), nil
	}

	// Handle exact match
	if strings.HasPrefix(constraint, "=") || !strings.ContainsAny(constraint, "^~><") {
		exactVersion, err := Parse(strings.TrimPrefix(constraint, "="))
		if err != nil {
			return false, err
		}
		return v.Compare(exactVersion) == 0, nil
	}

	return false, fmt.Errorf("unsupported constraint: %s", constraint)
}

func satisfiesCaret(v, base *Version) bool {
	// ^1.2.3 allows >=1.2.3 <2.0.0
	// Also allow pre-releases if base version matches (e.g., 1.2.3-alpha satisfies ^1.2.3)
	if base.Major > 0 {
		if v.Major != base.Major {
			return false
		}
		// If base version matches exactly (including pre-release), allow it
		if v.Major == base.Major && v.Minor == base.Minor && v.Patch == base.Patch {
			return true
		}
		return v.Compare(base) >= 0
	}
	// ^0.2.3 allows >=0.2.3 <0.3.0
	if base.Minor > 0 {
		if v.Major != base.Major || v.Minor != base.Minor {
			return false
		}
		// Allow pre-releases of the base version
		if v.Patch == base.Patch {
			return true
		}
		return v.Compare(base) >= 0
	}
	// ^0.0.3 allows >=0.0.3 <0.0.4
	return v.Major == base.Major && v.Minor == base.Minor && v.Patch == base.Patch
}

func satisfiesTilde(v, base *Version) bool {
	// ~1.2.3 allows >=1.2.3 <1.3.0
	return v.Major == base.Major && v.Minor == base.Minor && v.Compare(base) >= 0
}
