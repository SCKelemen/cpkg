package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/git"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var (
	initModule  string
	initDepRoot string
)

var initCmd = clix.NewCommand("init",
	clix.WithCommandShort("Initialize a new module"),
	clix.WithCommandLong("Initialize a new cpkg module in the current directory"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runInit(ctx)
	}),
)

func init() {
	initCmd.Flags = clix.NewFlagSet("init")
	initCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "module",
			Usage: "Module path",
		},
		Value: &initModule,
	})
	initCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Dependency root directory",
		},
		Value: &initDepRoot,
	})
}

func runInit(ctx *clix.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath := filepath.Join(cwd, manifest.ManifestFileName)
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("%s already exists", manifest.ManifestFileName)
	}

	// Determine module path
	module := initModule
	if module == "" {
		if git.IsGitRepo(cwd) {
			inferred, err := git.InferModulePath(cwd)
			if err == nil {
				module = inferred
			}
		}
		if module == "" {
			return fmt.Errorf("module path required (use --module or ensure git repo with origin remote)")
		}
	}

	// Determine dep root
	depRoot := initDepRoot
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = manifest.DefaultDepRoot()
		}
	}

	// Create manifest
	m := &manifest.Manifest{
		APIVersion: "cpkg.ringil.dev/v0",
		Kind:       "Module",
		Module:     module,
		DepRoot:    depRoot,
		Language: manifest.Language{
			CStandard: "c23",
			SKC:       true,
		},
		Dependencies: make(map[string]manifest.Dependency),
	}

	if err := manifest.Save(m, manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	fmt.Fprintf(ctx.App.Out, "Created %s\n", manifest.ManifestFileName)
	fmt.Fprintf(ctx.App.Out, "Module: %s\n", module)
	fmt.Fprintf(ctx.App.Out, "DepRoot: %s\n", depRoot)

	return nil
}
