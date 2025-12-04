# cpkg Commands Reference

This document provides comprehensive documentation for all cpkg commands, including usage, flags, arguments, and examples.

## Table of Contents

- [Global Flags](#global-flags)
- [init](#init) - Initialize a new module
- [add](#add) - Add or update a dependency constraint
- [tidy](#tidy) - Resolve dependency graph and write lockfile
- [sync](#sync) - Sync git submodules to match lockfile
- [vendor](#vendor) - Copy or symlink resolved sources into vendor directory
- [upgrade](#upgrade) - Upgrade dependencies to latest compatible versions
- [check](#check) - Check for newer versions of dependencies
- [list](#list) - List all dependencies
- [status](#status) - Show dependency status
- [explain](#explain) - Explain a dependency in detail
- [build](#build) - Build the project
- [test](#test) - Run tests
- [graph](#graph) - Display the dependency graph
- [version](#version) - Show version information
- [help](#help) - Show help for commands

## Global Flags

These flags are available for all commands:

- `-h, --help` - Show help information
- `-v, --version` - Show version information (root command only)
- `--format {text,json,yaml}` - Output format for structured data (default: `text`)
  - `text` - Human-readable text output (default)
  - `json` - JSON output for machine parsing (e.g., with `jq`)
  - `yaml` - YAML output for machine parsing
- `--dep-root DIR` - Override dependency root directory (if supported by command)
- `--verbose, -v` - More logging (debug info, git commands when useful)
- `--quiet, -q` - Minimal output
- `--color {auto,always,never}` - Color handling

### Format Flag

The `--format` flag allows you to switch the output format of commands that produce structured data. This makes it easy to work with cpkg output in other tools like `jq`, `yq`, or custom scripts.

**Supported commands:**
- `cpkg list --format json`
- `cpkg check --format json`
- `cpkg status --format yaml`
- `cpkg explain <module> --format json`
- `cpkg graph --format json`

**Examples:**

```bash
# Get JSON output for parsing with jq
cpkg list --format json | jq '.dependencies[] | select(.status == "locked")'

# Get YAML output
cpkg check --format yaml > check-results.yaml

# Use in scripts
for module in $(cpkg list --format json | jq -r '.dependencies[].module'); do
  echo "Checking $module"
  cpkg explain "$module" --format json
done
```

### Environment Variables

- `CPKG_DEP_ROOT` - Runtime override of `depRoot`
- `CPKG_TARGET` - Default target for `cpkg build`/`cpkg test` when `--target` not supplied

---

## init

Initialize a new cpkg module in the current directory.

### Help Text

```
Initialize a new module

USAGE
  cpkg init [flags]

Initialize a new cpkg module in the current directory

FLAGS
  --module             Module path
  --dep-root           Dependency root directory
  -h, --help           Show help information
```

### Description

Creates a new `cpkg.yaml` manifest file in the current directory. The module path can be inferred from the git repository's remote URL, or explicitly specified with the `--module` flag.

### Flags

- `--module <path>` - Explicitly specify the module path (e.g., `github.com/user/repo`). If not provided, cpkg will attempt to infer it from the git repository's `origin` remote URL.
- `--dep-root <dir>` - Set the dependency root directory. Defaults to `third_party/cpkg` if not specified. Can also be set via `CPKG_DEP_ROOT` environment variable.

### Behavior

- If `cpkg.yaml` already exists, the command will error (v0 behavior; `--force` may be added in future versions).
- The created manifest includes:
  - `apiVersion: cpkg.ringil.dev/v0`
  - `kind: Module`
  - `module: <resolved-module>`
  - `depRoot: <resolved-depRoot>`
  - Default language settings (`cStandard: c23`, `skc: true`)
  - Empty dependencies map

### Examples

```bash
# Initialize with auto-detected module path
cpkg init

# Initialize with explicit module path
cpkg init --module github.com/user/myproject

# Initialize with custom dependency root
cpkg init --dep-root deps
```

---

## add

Add or update a dependency constraint in `cpkg.yaml`.

### Help Text

```
Add or update a dependency constraint

USAGE
  cpkg add [flags]

FLAGS
  -h, --help           Show help information
```

### Description

Adds one or more dependencies to the manifest file. Each dependency must include a version constraint (e.g., `^1.0.0`, `~2.1.0`). If a dependency already exists, it will be updated with the new version constraint.

### Arguments

The command accepts one or more module specifications in the format `module[@version]`:

- `module` - The module path (e.g., `github.com/user/repo` or `github.com/user/repo/subpath`)
- `@version` - Optional version constraint. If not provided, the command will error and require a version.

### Examples

```bash
# Add a single dependency
cpkg add github.com/user/repo@^1.0.0

# Add multiple dependencies at once
cpkg add github.com/user/repo1@^1.0.0 github.com/user/repo2@^2.0.0

# Add a module from a subpath (multi-module support)
cpkg add github.com/user/repo/intrusive_list@^1.0.0
cpkg add github.com/user/repo/span@^1.0.0

# Update an existing dependency
cpkg add github.com/user/repo@^2.0.0
```

### Output

The command shows which dependencies were added or updated:
- `+ module @ version` - New dependency added
- `~ module: old_version → new_version` - Existing dependency updated

---

## tidy

Resolve dependency graph and write lockfile.

### Help Text

```
Resolve dependency graph and write lockfile

USAGE
  cpkg tidy [flags]

FLAGS
  --dep-root           Override dependency root
  --check              Check if lockfile would change without writing
  -h, --help           Show help information
```

### Description

Resolves all dependencies specified in `cpkg.yaml`, determines the latest compatible versions that satisfy the constraints, and writes the resolved information to `lock.cpkg.yaml`. This command:

1. Reads `cpkg.yaml` to get dependency constraints
2. Fetches tags from each dependency's repository
3. Filters tags based on module subpath (for multi-module repos)
4. Selects the highest version that satisfies each constraint
5. Resolves commit SHAs and computes checksums
6. Writes the lockfile with exact versions, commits, and paths

### Flags

- `--dep-root <dir>` - Override the dependency root directory. Defaults to the value in `cpkg.yaml` or `CPKG_DEP_ROOT` environment variable.
- `--check` - Check if the lockfile would change without actually writing it. Useful for CI/CD to verify dependencies are up to date. Exits with non-zero status if changes would be made.

### Output

The command shows a summary of changes:
- `+ module @ version` - New dependency added
- `~ module: old_version → new_version` - Dependency version updated
- `- module` - Dependency removed

### Examples

```bash
# Resolve and lock dependencies
cpkg tidy

# Check if lockfile is up to date (CI/CD)
cpkg tidy --check

# Use custom dependency root
cpkg tidy --dep-root deps
```

### Notes

- Requires `cpkg.yaml` to exist
- Creates or updates `lock.cpkg.yaml` in the same directory as `cpkg.yaml`
- The lockfile pins exact versions, commits, and checksums for reproducible builds

---

## sync

Sync git submodules to match lockfile.

### Help Text

```
Sync git submodules to match lockfile

USAGE
  cpkg sync [flags]

FLAGS
  --dep-root           Override dependency root
  -h, --help           Show help information
```

### Description

Synchronizes git submodules to match the versions specified in `lock.cpkg.yaml`. This command:

1. Reads `lock.cpkg.yaml` to get locked dependency information
2. For each dependency:
   - Adds the submodule if it doesn't exist
   - Updates the submodule URL if it has changed
   - Initializes the submodule if needed
   - Fetches tags and commits
   - Checks out the exact commit specified in the lockfile

### Flags

- `--dep-root <dir>` - Override the dependency root directory. Note: paths in the lockfile are already resolved, so this flag may not have an effect in all cases.

### Output

The command shows the status of each dependency:
- `+ module` - New submodule added
- `~ module (URL updated)` - Submodule URL updated
- `✓ module @ version (commit)` - Submodule synced successfully

### Examples

```bash
# Sync all dependencies
cpkg sync

# Sync with custom dependency root
cpkg sync --dep-root deps
```

### Notes

- Requires `lock.cpkg.yaml` to exist (run `cpkg tidy` first if needed)
- Creates git submodules under the dependency root directory
- Each dependency gets its own submodule, even if multiple modules come from the same repository (for multi-module support)

---

## vendor

Copy or symlink resolved sources into vendor directory.

### Help Text

```
Copy or symlink resolved sources into vendor directory

USAGE
  cpkg vendor [flags]

FLAGS
  --vendor-root        Vendor root directory
  --symlink            Create symlinks instead of copying files (faster, no duplication, default on Unix)
  --copy               Force copying files instead of symlinks (more compatible, uses more disk space)
  -h, --help           Show help information
```

### Description

Creates a flat `vendor/` directory containing all dependency sources, either as symlinks or copies. This is useful for build systems that prefer a single vendor directory over git submodules.

### Flags

- `--vendor-root <dir>` - Specify the vendor root directory. Defaults to `vendor` if not specified.
- `--symlink` - Create symlinks instead of copying files. This is faster and doesn't duplicate files, but symlinks may not work in all environments (e.g., some Windows setups, certain build tools).
- `--copy` - Force copying files instead of symlinks. Uses more disk space but is more compatible across platforms and build tools.

### Default Behavior

- **Unix-like systems (Linux, macOS)**: Uses symlinks by default
- **Windows**: Uses copies by default
- You can override the default with `--symlink` or `--copy` flags

### Output

For each dependency:
- `Symlinked module @ version` or `Vendored module @ version`
- Shows source and destination paths
- Summary line showing total number of dependencies processed

### Examples

```bash
# Create vendor directory with symlinks (default on Unix)
cpkg vendor

# Force copying instead of symlinks
cpkg vendor --copy

# Use custom vendor directory
cpkg vendor --vendor-root third_party

# Explicitly use symlinks
cpkg vendor --symlink
```

### Notes

- Requires `lock.cpkg.yaml` to exist (run `cpkg tidy` first)
- Uses the `sourcePath` field from the lockfile to locate source files
- The vendor directory structure mirrors module paths (e.g., `vendor/github.com/user/repo/`)
- Existing vendor entries are removed before creating new ones

---

## upgrade

Upgrade dependencies to latest compatible versions.

### Help Text

```
Upgrade dependencies to latest compatible versions

USAGE
  cpkg upgrade [flags]

Upgrade dependencies to the latest versions that satisfy their constraints, then run tidy and sync

FLAGS
  --all                Upgrade all dependencies (even if no updates available)
  --dep-root           Override dependency root
  -h, --help           Show help information
```

### Description

Checks for newer versions of dependencies that satisfy the constraints in `cpkg.yaml`, updates the manifest if newer versions are found, then runs `cpkg tidy` and `cpkg sync` to update the lockfile and submodules.

### Flags

- `--all` - Upgrade all dependencies even if no updates are available. This will refresh all dependencies to their latest compatible versions.
- `--dep-root <dir>` - Override the dependency root directory.

### Behavior

1. Reads `cpkg.yaml` and `lock.cpkg.yaml`
2. For each dependency, fetches tags and finds the latest version that satisfies the constraint
3. If a newer version is found, updates the manifest
4. Runs `cpkg tidy` to update the lockfile
5. Runs `cpkg sync` to update submodules

### Output

Shows which dependencies are being upgraded:
- `Upgrading module: old_version → new_version` - Dependency upgraded
- `Refreshing module: version` - Dependency refreshed (when using `--all`)
- `All dependencies are up to date.` - No updates available

### Examples

```bash
# Upgrade dependencies with available updates
cpkg upgrade

# Upgrade all dependencies (refresh to latest compatible versions)
cpkg upgrade --all

# Upgrade with custom dependency root
cpkg upgrade --dep-root deps
```

### Notes

- Only upgrades within the constraints specified in `cpkg.yaml`
- Does not modify version constraints (e.g., `^1.0.0` stays `^1.0.0`)
- Automatically runs `tidy` and `sync` after upgrading

---

## check

Check for newer versions of dependencies (without modifying files).

### Help Text

```
Check for newer versions of dependencies

USAGE
  cpkg check [flags]

FLAGS
  -h, --help           Show help information
```

### Format Support

The `check` command supports the `--format` flag for JSON and YAML output:

```bash
# JSON output
cpkg check --format json

# YAML output
cpkg check --format yaml
```

**JSON/YAML Structure:**
```json
{
  "dependencies": [
    {
      "module": "github.com/user/repo",
      "current": "v1.2.3",
      "latest": "v1.2.5",
      "constraint": "^1.2.0",
      "notes": "patch available"
    }
  ],
  "all_up_to_date": false
}
```

### Description

Displays a table showing which dependencies have newer versions available, without modifying any files. This is useful for checking if updates are available before running `cpkg upgrade`.

### Output

Displays a table with the following columns:
- **MODULE** - The module path
- **CURRENT** - The currently locked version
- **LATEST** - The latest available version that satisfies the constraint
- **CONSTRAINT** - The version constraint from `cpkg.yaml`
- **NOTES** - Status information:
  - `patch available` - A patch version update is available
  - `minor available` - A minor version update is available
  - `major available` - A major version update is available (but may not satisfy constraint)
  - `up to date` - No updates available
  - `ERROR` - An error occurred (e.g., invalid module path, failed to fetch tags)

### Examples

```bash
# Check for available updates
cpkg check
```

### Example Output

```
MODULE                           CURRENT   LATEST     CONSTRAINT   NOTES
github.com/user/repo             v1.2.3    v1.2.5     ^1.2.0       patch available
github.com/user/other            v2.0.0    v2.1.0     ^2.0.0       minor available
github.com/user/stable           v1.0.0    v1.0.0     ^1.0.0       up to date

All dependencies are up to date.
```

### Notes

- Requires `cpkg.yaml` and `lock.cpkg.yaml` to exist
- Does not modify any files
- Fetches tags from remote repositories to determine latest versions
- Respects version constraints (only shows versions that satisfy constraints)

---

## list

List all dependencies from the manifest and their locked versions.

### Help Text

```
List all dependencies

USAGE
  cpkg list [flags]

List all dependencies from the manifest and their locked versions

FLAGS
  -h, --help           Show help information
```

### Format Support

The `list` command supports the `--format` flag for JSON and YAML output:

```bash
# JSON output
cpkg list --format json

# YAML output
cpkg list --format yaml
```

**JSON/YAML Structure:**
```json
{
  "dependencies": [
    {
      "module": "github.com/user/repo",
      "constraint": "^1.0.0",
      "locked": "v1.2.3",
      "status": "locked",
      "has_lockfile": true
    }
  ]
}
```

### Description

Displays a table of all dependencies listed in `cpkg.yaml`, showing their constraints and locked versions (if a lockfile exists).

### Output

If lockfile exists:
- **MODULE** - The module path
- **CONSTRAINT** - The version constraint from `cpkg.yaml`
- **LOCKED** - The locked version from `lock.cpkg.yaml`
- **STATUS** - Status indicator:
  - `✓` - Dependency is locked
  - `⚠` - Dependency is in manifest but not in lockfile

If no lockfile exists:
- Shows only **MODULE** and **CONSTRAINT** columns
- Displays a warning message

### Examples

```bash
# List all dependencies
cpkg list
```

### Example Output

```
MODULE                           CONSTRAINT   LOCKED       STATUS
github.com/user/repo             ^1.2.0       v1.2.3       ✓
github.com/user/other            ^2.0.0       v2.0.0       ✓
github.com/user/new              ^1.0.0       NOT_LOCKED  ⚠
```

### Notes

- Requires `cpkg.yaml` to exist
- Works without `lock.cpkg.yaml`, but shows limited information
- Run `cpkg tidy` to create/update the lockfile

---

## status

Show dependency status.

### Help Text

```
Show dependency status

USAGE
  cpkg status [flags]

FLAGS
  -h, --help           Show help information
```

### Format Support

The `status` command supports the `--format` flag for JSON and YAML output:

```bash
# JSON output
cpkg status --format json

# YAML output
cpkg status --format yaml
```

**JSON/YAML Structure:**
```json
{
  "dependencies": [
    {
      "module": "github.com/user/repo",
      "constraint": "^1.0.0",
      "locked_version": "v1.2.3",
      "local_version": "v1.2.3",
      "status": "OK"
    }
  ]
}
```

### Description

Displays detailed status information about each dependency, including whether the local git submodule matches the locked version and if there are any uncommitted changes.

### Output

A table with the following columns:
- **MODULE** - The module path
- **CONSTRAINT** - The version constraint from `cpkg.yaml`
- **LOCKED** - The locked version from `lock.cpkg.yaml`
- **LOCAL** - The local submodule state:
  - Version number if in sync
  - Commit SHA if out of sync
  - `(dirty)` suffix if there are uncommitted changes
  - `MISSING` if submodule is not initialized
- **STATUS** - Overall status:
  - `OK` - Submodule is in sync with lockfile
  - `OUT_OF_SYNC` - Submodule is at a different commit
  - `DIRTY` - Submodule has uncommitted changes
  - `MISSING` - Submodule is not initialized
  - `NO_LOCK` - Dependency is not in lockfile

### Examples

```bash
# Check dependency status
cpkg status
```

### Example Output

```
MODULE                           CONSTRAINT   LOCKED       LOCAL          STATUS
github.com/user/repo             ^1.2.0       v1.2.3       v1.2.3         OK
github.com/user/other            ^2.0.0       v2.0.0       a1b2c3d        OUT_OF_SYNC
github.com/user/modified         ^1.0.0       v1.0.0       v1.0.0 (dirty) DIRTY
github.com/user/missing          ^1.0.0       v1.0.0       MISSING        MISSING
```

### Notes

- Requires `cpkg.yaml` to exist
- Works without `lock.cpkg.yaml`, but shows limited information
- Checks actual git submodule state, not just lockfile contents
- Run `cpkg sync` to sync submodules to locked versions

---

## explain

Explain a dependency in detail.

### Help Text

```
Explain a dependency in detail

USAGE
  cpkg explain [flags]

Show detailed information about a specific dependency

FLAGS
  -h, --help           Show help information

ARGUMENTS
  module               required
```

### Format Support

The `explain` command supports the `--format` flag for JSON and YAML output:

```bash
# JSON output
cpkg explain github.com/user/repo --format json

# YAML output
cpkg explain github.com/user/repo --format yaml
```

**JSON/YAML Structure:**
```json
{
  "module": "github.com/user/repo",
  "constraint": "^1.0.0",
  "locked": {
    "version": "v1.2.3",
    "commit": "a1b2c3d4e5f6...",
    "sum": "h1:abc123...",
    "vcs": "git",
    "repo_url": "https://github.com/user/repo.git",
    "path": "third_party/cpkg/github.com/user/repo"
  },
  "local_state": {
    "submodule_exists": true,
    "current_commit": "a1b2c3d",
    "in_sync": true,
    "is_dirty": false
  }
}
```

### Description

Shows comprehensive information about a specific dependency, including its constraint, locked version, commit, repository URL, local submodule state, and more.

### Arguments

- `module` (required) - The module path to explain (must match exactly as it appears in `cpkg.yaml`)

### Output

Detailed information including:
- **Dependency**: Module path
- **Constraint**: Version constraint from manifest
- **Locked Information** (if lockfile exists):
  - Version
  - Commit SHA
  - Checksum
  - VCS type
  - Repository URL
  - Local path
- **Local State**:
  - Whether submodule exists
  - Current commit
  - Sync status (in sync, out of sync)
  - Working tree status (clean, dirty)

### Examples

```bash
# Explain a specific dependency
cpkg explain github.com/user/repo

# Explain a subpath module
cpkg explain github.com/user/repo/subpath
```

### Example Output

```
Dependency: github.com/user/repo
─────────────────────────────────────────────────────────────

Constraint: ^1.2.0

Locked Information:
  Version: v1.2.3
  Commit:  a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  Sum:     h1:abc123...
  VCS:     git
  Repo:    https://github.com/user/repo.git
  Path:    third_party/cpkg/github.com/user/repo

Local State:
  Submodule: exists
  Current commit: a1b2c3d
  Status: ✓ in sync
  Working tree: ✓ clean
```

### Notes

- Requires `cpkg.yaml` to exist
- The module path must match exactly as specified in the manifest
- Provides the most detailed view of a dependency's state

---

## build

Build the project.

### Help Text

```
Build the project

USAGE
  cpkg build [flags]

FLAGS
  --target             Build target name
  --dep-root           Override dependency root
  -h, --help           Show help information
```

### Description

Runs the project's build command after ensuring dependencies are resolved and synced. This command:

1. Runs `cpkg tidy` to resolve and lock dependencies
2. Runs `cpkg sync` to sync git submodules
3. Determines the build command from `cpkg.yaml`
4. Executes the build command with appropriate environment variables

### Flags

- `--target <name>` - Specify a build target name. If provided, cpkg will look for `build.targets.<name>.command` in the manifest. If not provided, uses `build.command` or the `CPKG_TARGET` environment variable.
- `--dep-root <dir>` - Override the dependency root directory.

### Environment Variables

The build command receives these environment variables:
- `CPKG_ROOT` - Project root directory (where `cpkg.yaml` is located)
- `CPKG_DEP_ROOT` - Dependency root directory
- `CPKG_TARGET` - Build target name (if specified)

### Examples

```bash
# Build using default build command
cpkg build

# Build a specific target
cpkg build --target release

# Build with custom dependency root
cpkg build --dep-root deps
```

### Notes

- Requires `cpkg.yaml` with a `build.command` or `build.targets` configuration
- Automatically runs `tidy` and `sync` before building
- The build command is executed in the project root directory
- Build command output is streamed to stdout/stderr

---

## test

Run tests.

### Help Text

```
Run tests

USAGE
  cpkg test [flags]

FLAGS
  --target             Test target name
  --dep-root           Override dependency root
  -h, --help           Show help information
```

### Description

Runs the project's test command. This command:

1. Runs `cpkg build` first (to ensure the project is built)
2. Executes the test command from `cpkg.yaml`

### Flags

- `--target <name>` - Specify a test target name. If provided, cpkg will look for `test.targets.<name>.command` in the manifest. If not provided, uses `test.command` or the `CPKG_TARGET` environment variable.
- `--dep-root <dir>` - Override the dependency root directory.

### Environment Variables

The test command receives the same environment variables as `build`:
- `CPKG_ROOT` - Project root directory
- `CPKG_DEP_ROOT` - Dependency root directory
- `CPKG_TARGET` - Test target name (if specified)

### Examples

```bash
# Run tests using default test command
cpkg test

# Run tests for a specific target
cpkg test --target integration

# Run tests with custom dependency root
cpkg test --dep-root deps
```

### Notes

- Requires `cpkg.yaml` with a `test.command` configuration
- Automatically runs `build` before testing
- The test command is executed in the project root directory
- Test command output is streamed to stdout/stderr

---

## graph

Display the dependency graph.

### Help Text

```
Display the dependency graph

USAGE
  cpkg graph [flags]

FLAGS
  -h, --help           Show help information
```

### Format Support

The `graph` command supports the `--format` flag for JSON and YAML output:

```bash
# JSON output
cpkg graph --format json

# YAML output
cpkg graph --format yaml
```

**JSON/YAML Structure:**
```json
{
  "module": "github.com/user/myproject",
  "dependencies": [
    {
      "module": "github.com/user/dep1",
      "version": "v1.2.3"
    },
    {
      "module": "github.com/user/dep2",
      "version": "v2.0.0"
    }
  ]
}
```

### Description

Displays a simple tree view of the dependency graph, showing the root module and all its dependencies with their locked versions.

### Output

A tree structure showing:
- Root module name
- Dependencies as children with version information

### Examples

```bash
# Display dependency graph
cpkg graph
```

### Example Output

```
github.com/user/myproject
├─ github.com/user/dep1 v1.2.3
├─ github.com/user/dep2 v2.0.0
└─ github.com/user/dep3 v1.0.0
```

### Notes

- Requires `lock.cpkg.yaml` to exist (run `cpkg tidy` first)
- Currently shows a flat list (transitive dependencies not yet supported)
- Dependencies are sorted alphabetically

---

## version

Show version information.

### Help Text

```
Show version information

USAGE
  cpkg version

FLAGS
  -h, --help           Show help information
```

### Description

Displays detailed version information about the cpkg binary, including:
- Version number (semantic version)
- Git commit hash (short)
- Build date
- Go version
- Operating system and architecture

### Examples

```bash
# Show version information
cpkg version

# Short version flag (also available on root command)
cpkg --version
```

### Example Output

```
Version: 2.0.0
Commit:  817d5f2
Date:    2025-12-04
Go:      go1.23.0
OS/Arch: darwin/arm64
```

### Notes

- Version information is injected at build time via ldflags
- For pre-built binaries, this information comes from the git tag
- For local builds, version defaults to "dev"

---

## help

Show help for commands.

### Help Text

```
Show help for commands

USAGE
  cpkg help [command]

FLAGS
  -h, --help           Show help information
```

### Description

Displays help information for cpkg commands. Can be used to get help for the root command or any specific command.

### Usage

```bash
# Show general help
cpkg help

# Show help for a specific command
cpkg help add
cpkg help tidy
cpkg help build
```

### Notes

- Equivalent to `cpkg --help` or `cpkg <command> --help`
- The help extension is provided by the clix library

---

## Command Workflow

A typical workflow using cpkg commands:

1. **Initialize** a new module:
   ```bash
   cpkg init
   ```

2. **Add** dependencies:
   ```bash
   cpkg add github.com/user/repo@^1.0.0
   ```

3. **Resolve and lock** dependencies:
   ```bash
   cpkg tidy
   ```

4. **Sync** git submodules:
   ```bash
   cpkg sync
   ```

5. **Check** for updates:
   ```bash
   cpkg check
   ```

6. **Upgrade** dependencies (if needed):
   ```bash
   cpkg upgrade
   ```

7. **Build** the project:
   ```bash
   cpkg build
   ```

8. **Test** the project:
   ```bash
   cpkg test
   ```

For more information, see the [README](../README.md) and other documentation in the `docs/` directory.

