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

var (
	vendorRoot    string
	vendorSymlink bool
	vendorCopy    bool
)

var vendorCmd = clix.NewCommand("vendor",
	clix.WithCommandShort("Copy or symlink resolved sources into vendor directory"),
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
	vendorCmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "symlink",
			Usage: "Create symlinks instead of copying files (faster, no duplication, default on Unix)",
		},
		Value: &vendorSymlink,
	})
	vendorCmd.Flags.BoolVar(clix.BoolVarOptions{
		FlagOptions: clix.FlagOptions{
			Name:  "copy",
			Usage: "Force copying files instead of symlinks (more compatible, uses more disk space)",
		},
		Value: &vendorCopy,
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

	// Determine whether to use symlinks
	// Default: use symlinks on Unix (macOS/Linux), copy on Windows
	// User can override with --symlink or --copy flags
	useSymlink := vendorSymlink || (!vendorCopy && os.PathSeparator == '/')

	count := 0
	for modulePath, dep := range lock.Dependencies {
		// Use SourcePath from lockfile (already computed as path + subdir)
		sourcePath := dep.SourcePath
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(projectRoot, sourcePath)
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

		// Remove existing destination if it exists
		if _, err := os.Lstat(destPath); err == nil {
			if err := os.RemoveAll(destPath); err != nil {
				return fmt.Errorf("failed to remove existing destination %s: %w", destPath, err)
			}
		}

		// Get relative paths for display
		relSource, _ := filepath.Rel(projectRoot, sourcePath)
		relDest, _ := filepath.Rel(projectRoot, destPath)
		if relSource == "" {
			relSource = sourcePath
		}
		if relDest == "" {
			relDest = destPath
		}

		if useSymlink {
			// Create symlink
			// Use relative path for portability
			linkTarget, err := filepath.Rel(filepath.Dir(destPath), sourcePath)
			if err != nil {
				// Fallback to absolute if relative fails
				linkTarget = sourcePath
			}
			if err := os.Symlink(linkTarget, destPath); err != nil {
				return fmt.Errorf("failed to create symlink for %s: %w", modulePath, err)
			}
			fmt.Fprintf(ctx.App.Out, "Symlinked %s @ %s\n", modulePath, dep.Version)
			fmt.Fprintf(ctx.App.Out, "  %s -> %s\n", relDest, relSource)
		} else {
			// Copy directory (simplified - in production you might want to use a proper copy library)
			if err := copyDir(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", modulePath, err)
			}
			fmt.Fprintf(ctx.App.Out, "Vendored %s @ %s\n", modulePath, dep.Version)
			fmt.Fprintf(ctx.App.Out, "  %s <- %s\n", relDest, relSource)
		}

		count++
	}

	action := "Vendored"
	if useSymlink {
		action = "Symlinked"
	}
	fmt.Fprintf(ctx.App.Out, "\n%s %d dependencies into %s/\n", action, count, vRoot)
	return nil
}

func copyDir(src, dst string) error {
	// Simple recursive copy using cp command
	// In production, you might want to use a proper Go file copy library
	cmd := exec.Command("cp", "-r", src, dst)
	return cmd.Run()
}
