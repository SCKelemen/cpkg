package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SCKelemen/clix"
	"github.com/SCKelemen/cpkg/internal/format"
	"github.com/SCKelemen/cpkg/internal/lockfile"
	"github.com/SCKelemen/cpkg/internal/manifest"
	"github.com/SCKelemen/cpkg/internal/submodule"
)

type explainOutput struct {
	Module     string              `json:"module" yaml:"module"`
	Constraint string              `json:"constraint" yaml:"constraint"`
	Locked     *explainLocked      `json:"locked,omitempty" yaml:"locked,omitempty"`
	LocalState *explainLocalState  `json:"local_state,omitempty" yaml:"local_state,omitempty"`
}

type explainLocked struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	Sum     string `json:"sum" yaml:"sum"`
	VCS     string `json:"vcs" yaml:"vcs"`
	RepoURL string `json:"repo_url" yaml:"repo_url"`
	Path    string `json:"path" yaml:"path"`
}

type explainLocalState struct {
	SubmoduleExists bool   `json:"submodule_exists" yaml:"submodule_exists"`
	CurrentCommit   string `json:"current_commit,omitempty" yaml:"current_commit,omitempty"`
	InSync          bool   `json:"in_sync" yaml:"in_sync"`
	IsDirty         bool   `json:"is_dirty" yaml:"is_dirty"`
}

var explainCmd = clix.NewCommand("explain",
	clix.WithCommandShort("Explain a dependency in detail"),
	clix.WithCommandLong("Show detailed information about a specific dependency"),
	clix.WithCommandRun(func(ctx *clix.Context) error {
		return runExplain(ctx)
	}),
	clix.WithCommandArguments(
		clix.NewArgument(
			clix.WithArgName("module"),
			clix.WithArgRequired(),
		),
	),
)

func runExplain(ctx *clix.Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("module path required")
	}

	modulePath := ctx.Args[0]

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

	dep, exists := m.Dependencies[modulePath]
	if !exists {
		return fmt.Errorf("dependency %s not found in manifest", modulePath)
	}

	lockfilePath := filepath.Join(filepath.Dir(manifestPath), lockfile.LockfileName)
	lock, err := lockfile.Load(lockfilePath)
	hasLockfile := err == nil

	outputFormat := GetFormat()
	output := explainOutput{
		Module:     modulePath,
		Constraint: dep.Version,
	}

	if hasLockfile {
		if lockDep, exists := lock.Dependencies[modulePath]; exists {
			output.Locked = &explainLocked{
				Version: lockDep.Version,
				Commit:  lockDep.Commit,
				Sum:     lockDep.Sum,
				VCS:     lockDep.VCS,
				RepoURL: lockDep.RepoURL,
				Path:    lockDep.Path,
			}

			// Check local submodule state
			submodulePath := lockDep.Path
			if !filepath.IsAbs(submodulePath) {
				submodulePath = filepath.Join(filepath.Dir(manifestPath), submodulePath)
			}

			// Resolve symlinks (important for macOS)
			resolvedPath, err := filepath.EvalSymlinks(submodulePath)
			if err == nil {
				submodulePath = resolvedPath
			}

			// Get relative path for git submodule commands
			relPath := submodulePath
			if filepath.IsAbs(submodulePath) {
				realCwd, _ := filepath.EvalSymlinks(cwd)
				if rel, err := filepath.Rel(realCwd, submodulePath); err == nil {
					relPath = rel
				}
			}

			localState := &explainLocalState{
				SubmoduleExists: submodule.SubmoduleExists(relPath),
			}

			if localState.SubmoduleExists {
				currentCommit, err := submodule.GetSubmoduleCommit(submodulePath)
				if err == nil {
					localState.CurrentCommit = currentCommit
					localState.InSync = currentCommit == lockDep.Commit
					dirty, _ := submodule.IsSubmoduleDirty(submodulePath)
					localState.IsDirty = dirty
				}
			}

			output.LocalState = localState
		}
	}

	// Output in requested format
	if outputFormat != format.FormatText {
		return format.Write(ctx.App.Out, outputFormat, output)
	}

	// Text output
	fmt.Fprintf(ctx.App.Out, "Dependency: %s\n", modulePath)
	fmt.Fprintf(ctx.App.Out, "─────────────────────────────────────────────────────────────\n\n")
	fmt.Fprintf(ctx.App.Out, "Constraint: %s\n", dep.Version)

	if hasLockfile {
		if output.Locked != nil {
			fmt.Fprintf(ctx.App.Out, "\nLocked Information:\n")
			fmt.Fprintf(ctx.App.Out, "  Version: %s\n", output.Locked.Version)
			fmt.Fprintf(ctx.App.Out, "  Commit:  %s\n", output.Locked.Commit)
			fmt.Fprintf(ctx.App.Out, "  Sum:     %s\n", output.Locked.Sum)
			fmt.Fprintf(ctx.App.Out, "  VCS:     %s\n", output.Locked.VCS)
			fmt.Fprintf(ctx.App.Out, "  Repo:    %s\n", output.Locked.RepoURL)
			fmt.Fprintf(ctx.App.Out, "  Path:    %s\n", output.Locked.Path)

			if output.LocalState != nil {
				fmt.Fprintf(ctx.App.Out, "\nLocal State:\n")
				if output.LocalState.SubmoduleExists {
					shortCommit := output.LocalState.CurrentCommit
					if len(shortCommit) > 7 {
						shortCommit = shortCommit[:7]
					}
					fmt.Fprintf(ctx.App.Out, "  Submodule: exists\n")
					fmt.Fprintf(ctx.App.Out, "  Current commit: %s\n", shortCommit)

					if output.LocalState.InSync {
						fmt.Fprintf(ctx.App.Out, "  Status: ✓ in sync\n")
					} else {
						fmt.Fprintf(ctx.App.Out, "  Status: ⚠ out of sync (locked: %s)\n", output.Locked.Commit[:7])
					}

					if output.LocalState.IsDirty {
						fmt.Fprintf(ctx.App.Out, "  Working tree: ⚠ dirty (has uncommitted changes)\n")
					} else {
						fmt.Fprintf(ctx.App.Out, "  Working tree: ✓ clean\n")
					}
				} else {
					fmt.Fprintf(ctx.App.Out, "  Submodule: ⚠ not initialized\n")
					fmt.Fprintf(ctx.App.Out, "  Run 'cpkg sync' to initialize\n")
				}
			}
		} else {
			fmt.Fprintf(ctx.App.Out, "\nLocked: ⚠ not locked\n")
			fmt.Fprintf(ctx.App.Out, "Run 'cpkg tidy' to lock this dependency\n")
		}
	} else {
		fmt.Fprintf(ctx.App.Out, "\nLocked: ⚠ no lockfile found\n")
		fmt.Fprintf(ctx.App.Out, "Run 'cpkg tidy' to create lockfile\n")
	}

	return nil
}

