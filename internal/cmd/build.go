package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var (
	buildTarget  string
	buildDepRoot string
)

var buildCmd = clix.NewCommand("build",
	clix.WithCommandShort("Build the project"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runBuild(ctx)
	}),
)

func init() {
	buildCmd.Flags = clix.NewFlagSet("build")
	buildCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "target",
			Usage: "Build target name",
		},
		Value: &buildTarget,
	})
	buildCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Override dependency root",
		},
		Value: &buildDepRoot,
	})
}

func runBuild(ctx *clix.Context) error {
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

	// Run tidy first
	fmt.Fprintf(ctx.App.Out, "Resolving dependencies...\n")
	if err := runTidyInternal(cwd, buildDepRoot, false); err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Run sync
	fmt.Fprintf(ctx.App.Out, "Syncing submodules...\n")
	if err := runSyncInternal(cwd, buildDepRoot); err != nil {
		return fmt.Errorf("failed to sync submodules: %w", err)
	}

	// Determine build command
	var buildCmd []string
	target := buildTarget
	if target == "" {
		target = os.Getenv("CPKG_TARGET")
	}

	if target != "" && m.Build != nil && m.Build.Targets != nil {
		if targetConfig, exists := m.Build.Targets[target]; exists {
			buildCmd = targetConfig.Command
		}
	}

	if buildCmd == nil && m.Build != nil {
		buildCmd = m.Build.Command
	}

	if len(buildCmd) == 0 {
		return fmt.Errorf("no build command configured in %s", manifest.ManifestFileName)
	}

	// Set environment variables
	depRoot := buildDepRoot
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = m.DepRoot
		}
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("CPKG_ROOT=%s", filepath.Dir(manifestPath)))
	env = append(env, fmt.Sprintf("CPKG_DEP_ROOT=%s", depRoot))
	if target != "" {
		env = append(env, fmt.Sprintf("CPKG_TARGET=%s", target))
	}

	// Execute build command
	fmt.Fprintf(ctx.App.Out, "Running build command...\n")
	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Dir = filepath.Dir(manifestPath)
	cmd.Env = env
	cmd.Stdout = ctx.App.Out
	cmd.Stderr = ctx.App.Err

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Fprintf(ctx.App.Out, "Build completed successfully.\n")
	return nil
}
