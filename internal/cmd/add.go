package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var addCmd = clix.NewCommand("add",
	clix.WithCommandShort("Add or update a dependency constraint"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runAdd(ctx)
	}),
)

func runAdd(ctx *clix.Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("no modules specified")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	if m.Dependencies == nil {
		m.Dependencies = make(map[string]manifest.Dependency)
	}

	// Parse each module@version argument
	for _, arg := range ctx.Args {
		module, version, err := parseModuleVersion(arg)
		if err != nil {
			return fmt.Errorf("invalid module specification %q: %w", arg, err)
		}

		if version == "" {
			return fmt.Errorf("version required for %s (e.g., %s@^1.0.0)", module, module)
		}

		oldDep, exists := m.Dependencies[module]
		if exists {
			fmt.Fprintf(ctx.App.Out, "~ %s: %s → %s\n", module, oldDep.Version, version)
		} else {
			fmt.Fprintf(ctx.App.Out, "+ %s @ %s\n", module, version)
		}

		m.Dependencies[module] = manifest.Dependency{
			Version: version,
		}
	}

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	return nil
}

func parseModuleVersion(spec string) (module, version string, err error) {
	parts := strings.Split(spec, "@")
	if len(parts) == 1 {
		return parts[0], "", nil
	}
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("invalid format: expected module@version")
}
