package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
)

var vendorRoot string

var vendorCmd = clix.NewCommand("vendor",
	clix.WithCommandShort("Copy resolved sources into vendor directory"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runVendor(ctx)
	}),
)

func init() {
	vendorCmd.Flags = clix.NewFlagSet("vendor")
	vendorCmd.Flags.StringVar(clix.StringVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "vendor-root",
			Usage: "Vendor root directory",
		},
		Value: &vendorRoot,
	})
}

func runVendor(ctx *clix.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath, err := manifest.FindManifest(cwd)
	if err != nil {
		return fmt.Errorf("no %s found: %w", manifest.ManifestFileName, err)
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	if err != nil {
		return fmt.Errorf("no lockfile found, run 'cpkg tidy' first: %w", err)
	}

	// Determine vendor root
	vRoot := vendorRoot
	if vRoot == "" {
		vRoot = "vendor"
	}

	projectRoot := filepath.Dir(manifestPath)
	vendorDir := filepath.Join(projectRoot, vRoot)

	// Create vendor directory
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		return fmt.Errorf("failed to create vendor directory: %w", err)
	}

	count := 0
	for modulePath, dep := range lock.Dependencies {
		sourcePath := dep.Path
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(projectRoot, sourcePath)
		}

		// If there's a subdir, the actual source files are in that subdirectory
		// within the submodule checkout
		if dep.Subdir != "" {
			sourcePath = filepath.Join(sourcePath, dep.Subdir)
		}

		// Check if source exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Fprintf(ctx.App.Err, "Warning: source path %s does not exist, skipping\n", sourcePath)
			continue
		}

		destPath := filepath.Join(vendorDir, modulePath)

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create vendor directory for %s: %w", modulePath, err)
		}

		// Copy directory (simplified - in production you might want to use a proper copy library)
		if err := copyDir(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", modulePath, err)
		}

		count++
		fmt.Fprintf(ctx.App.Out, "Vendored %s @ %s\n", modulePath, dep.Version)
	}

	fmt.Fprintf(ctx.App.Out, "\nVendored %d dependencies into %s/\n", count, vRoot)
	return nil
}

func copyDir(src, dst string) error {
	// Simple recursive copy using cp command
	// In production, you might want to use a proper Go file copy library
	cmd := exec.Command("cp", "-r", src, dst)
	return cmd.Run()
}
