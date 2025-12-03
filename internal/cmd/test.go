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
	testTarget  string
	testDepRoot string
)

var testCmd = clix.NewCommand("test",
	clix.WithCommandShort("Run tests"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runTest(ctx)
	}),
)

func init() {
	testCmd.Flags = clix.NewFlagSet("test")
	testCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "target",
			Usage: "Test target name",
		},
		Value: &testTarget,
	})
	testCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "dep-root",
			Usage: "Override dependency root",
		},
		Value: &testDepRoot,
	})
}

func runTest(ctx *clix.Context) error {
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

	// Run build first
	if err := runBuild(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Check for test command
	if m.Test == nil || len(m.Test.Command) == 0 {
		return fmt.Errorf("no test command configured in %s", manifest.ManifestFileName)
	}

	// Set environment variables
	depRoot := testDepRoot
	if depRoot == "" {
		if envDepRoot := os.Getenv("CPKG_DEP_ROOT"); envDepRoot != "" {
			depRoot = envDepRoot
		} else {
			depRoot = m.DepRoot
		}
	}

	target := testTarget
	if target == "" {
		target = os.Getenv("CPKG_TARGET")
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("CPKG_ROOT=%s", filepath.Dir(manifestPath)))
	env = append(env, fmt.Sprintf("CPKG_DEP_ROOT=%s", depRoot))
	if target != "" {
		env = append(env, fmt.Sprintf("CPKG_TARGET=%s", target))
	}

	// Execute test command
	fmt.Fprintf(ctx.App.Out, "Running tests...\n")
	cmd := exec.Command(m.Test.Command[0], m.Test.Command[1:]...)
	cmd.Dir = filepath.Dir(manifestPath)
	cmd.Env = env
	cmd.Stdout = ctx.App.Out
	cmd.Stderr = ctx.App.Err

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	fmt.Fprintf(ctx.App.Out, "Tests completed successfully.\n")
	return nil
}
