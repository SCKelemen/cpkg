package submodule

import (
	"path/filepath"
)

// ResolveSymlinks resolves all symlinks in a path to get the real path.
// This is necessary on macOS where /tmp is a symlink to /private/tmp,
// and git submodules don't work with symlinked paths.
func ResolveSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}


