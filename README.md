# cpkg

Source-only package manager for C source code.

## Overview

cpkg is a minimal, Go-implemented module and dependency manager for C (and SKC firmware) inspired by Go modules and Elm packages.

* **Source-only**: no prebuilt binaries, no ABI matrix.
* **Git-first**: modules are git repos, addressed by URL-ish paths.
* **Semver**: dependencies are versioned with semantic version tags.
* **Build-agnostic**: cpkg resolves and lays out source; your build system compiles it.
* **Submodule-native**: cpkg wraps git submodules to keep repos clean and reviewable.
* **Incremental adoption**: Use arbitrary subdirectories from any repo without upstream changes.
* **Version flexibility**: Different subdirectories can use different versions from the same repo.

## Installation

```bash
# Install from source (recommended)
go install github.com/SCKelemen/cpkg@latest

# Or download a pre-built binary from the [Releases](https://github.com/SCKelemen/cpkg/releases) page
```

## Quick Start

```bash
# Initialize a new module
cpkg init

# Add a dependency
cpkg add github.com/user/repo@^1.0.0

# Resolve and lock dependencies
cpkg tidy

# Sync git submodules
cpkg sync

# Build the project
cpkg build
```

## Dependency Management

### Adding Dependencies

```bash
# Add a new dependency with a version constraint
cpkg add github.com/user/repo@^1.0.0

# Add multiple dependencies at once
cpkg add github.com/user/repo1@^1.0.0 github.com/user/repo2@^2.0.0

# Add a module from a subpath (multi-module support)
cpkg add github.com/user/repo/intrusive_list@^1.0.0
cpkg add github.com/user/repo/span@^1.0.0
```

### Multi-Module Support

cpkg supports multiple modules from the same repository, with two modes:

#### 1. Versioned Submodules (For Repos You Control)

If you control the repository, you can create tags for each submodule:

**Tag Naming Conventions:**

1. **Prefix format** (recommended): `subpath/v1.0.0`
   ```bash
   git tag intrusive_list/v1.0.0
   git tag span/v1.0.0
   ```

2. **Suffix format**: `v1.0.0-subpath`
   ```bash
   git tag v1.0.0-intrusive_list
   git tag v1.0.0-span
   ```

**Example:**

```yaml
# cpkg.yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: "^1.0.0"
  github.com/user/firmware-lib/span:
    version: "^1.0.0"
```

#### 2. Arbitrary Subdirectories (For Any Repo) - Incremental Adoption

You can use **any subdirectory** from **any repository**, even if it doesn't have cpkg-specific tags or even know about cpkg. This enables **incremental adoption** - you can start using parts of a library without requiring the upstream repo to support cpkg.

cpkg will use the root repository tags and point to the subdirectory:

```yaml
# cpkg.yaml
dependencies:
  github.com/Mbed-TLS/mbedtls/library:  # Just the library source files
    version: "^3.6.0"
  github.com/Mbed-TLS/mbedtls/include:  # Just the headers
    version: "^3.6.0"
```

**Different versions for different subdirectories:**

You can even use different versions of different subdirectories from the same repo:

```yaml
# cpkg.yaml
dependencies:
  github.com/Mbed-TLS/mbedtls/library:  # Use v3.6.4
    version: "3.6.4"
  github.com/Mbed-TLS/mbedtls/include:  # Use v3.6.5
    version: "3.6.5"
```

This works because cpkg:
- Uses the root repo tags (e.g., `v3.6.4`, `v3.6.5`)
- Points to the subdirectory (e.g., `library/` or `include/`)
- Creates **separate submodules** for each, allowing different commits
- Each submodule can be at a different commit, even from the same repo

**How it works with different commits:**

Each module from the same repo gets its own git submodule. This allows independently versioned modules to be at different commits:

- `intrusive_list/v1.0.0` → commit `abc123` → submodule at `third_party/cpkg/.../intrusive_list`
- `span/v1.0.0` → commit `def456` → submodule at `third_party/cpkg/.../span`

This is necessary because git submodules can only point to a single commit. See [Multiple Commits](docs/multi-module-commits.md) for details.

### Checking for Updates

```bash
# Check which dependencies have updates available
cpkg outdated
```

This will show a table of dependencies with their current version, latest compatible version, and update type (major/minor/patch).

### Upgrading Dependencies

```bash
# Upgrade all dependencies to latest compatible versions
cpkg upgrade

# Upgrade all dependencies (even if already up to date)
cpkg upgrade --all
```

The `upgrade` command will:
1. Check each dependency for newer compatible versions
2. Update the lockfile with new versions
3. Sync git submodules to the new commits

### Viewing Dependencies

```bash
# List all dependencies
cpkg list

# Get detailed information about a specific dependency
cpkg explain github.com/user/repo

# View dependency graph
cpkg graph
```

### Manual Workflow

If you prefer to manually update dependencies:

```bash
# 1. Check what's outdated
cpkg outdated

# 2. Update the version constraint in cpkg.yaml (optional)
cpkg add github.com/user/repo@^2.0.0

# 3. Resolve and lock new versions
cpkg tidy

# 4. Sync submodules to locked versions
cpkg sync
```

## Automated Dependency Updates

### Using the GitHub Action

cpkg provides a reusable GitHub Action for automated dependency updates. You can use it in two ways:

#### Option 1: Using the action from this repository

Add this to your `.github/workflows/dependencies.yml`:

```yaml
name: Update Dependencies

on:
  schedule:
    # Run weekly on Monday at 00:00 UTC
    - cron: '0 0 * * 1'
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Update dependencies
        uses: github.com/SCKelemen/cpkg/.github/actions/cpkg-upgrade@main
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          pr-title: "chore: update dependencies"
          pr-body: |
            Automated dependency update.
            
            This PR was created by the cpkg upgrade workflow.
```

#### Option 2: Using a dedicated action repository (recommended)

If you prefer to use a dedicated action repository (e.g., `github.com/SCKelemen/cpkg-upgrade-action`), you can reference it directly:

```yaml
- name: Update dependencies
  uses: github.com/SCKelemen/cpkg-upgrade-action@v1
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
```

The action will:
- Check for outdated dependencies using `cpkg outdated`
- Upgrade them to latest compatible versions using `cpkg upgrade`
- Create a pull request with the changes (only if there are actual updates)

### Customizing the Action

You can customize the PR title, body, and branch name:

```yaml
- name: Update dependencies
  uses: github.com/SCKelemen/cpkg/.github/actions/cpkg-upgrade@main
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
    pr-title: "chore(deps): upgrade dependencies"
    pr-body: "Automated dependency updates via cpkg"
    branch: update-deps
    commit-message: "chore: update dependencies"
```

## Lockfile

The `cpkg.lock.yaml` file is similar to **Go's `go.mod` + `go.sum` combined**:
- **Version locking**: Pins exact versions and commits (like `go.mod`)
- **Integrity checking**: Includes checksums to verify dependency integrity (like `go.sum`)
- **Subdirectory tracking**: For multi-module repos, includes a `subdir` field indicating where source files are located

It's a single file for simplicity, serving both purposes.

### Example Lockfile Entry

```yaml
dependencies:
  github.com/user/repo/intrusive_list:
    version: v1.0.0
    commit: abc123...
    repoURL: https://github.com/user/repo.git
    path: third_party/cpkg/github.com/user/repo/intrusive_list
    subdir: intrusive_list  # ← Source files are in this subdirectory
```

## Build System Integration

cpkg manages dependencies and provides source files, but **does not control the compiler or linker**. Your build system is responsible for compiling and linking.

### Finding Source Files

The lockfile includes a `sourcePath` field that points directly to where the source files are:

```yaml
dependencies:
  github.com/user/repo/intrusive_list:
    path: third_party/cpkg/github.com/user/repo/intrusive_list  # Submodule path
    subdir: intrusive_list  # Subdirectory within repo
    sourcePath: third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list  # ← Use this!
```

Your build system should:

1. Read `cpkg.lock.yaml`
2. Use the `sourcePath` field directly (no computation needed)
3. Add include paths and compile

### Using `cpkg vendor`

For a simpler flat structure, use `cpkg vendor`:

```bash
# Default: creates symlinks on Unix (macOS/Linux), copies on Windows
cpkg vendor

# Force symlinks (faster, no duplication)
cpkg vendor --symlink

# Force copying (more compatible, uses more disk space)
cpkg vendor --copy
```

**Symlinks vs Copying:**
- **Symlinks** (default on Unix): No disk duplication, always in sync with submodules, faster. Best for development on macOS/Linux.
- **Copies** (default on Windows): More compatible with all build systems, works offline after vendoring.
- **Copying** (default): Works everywhere, build systems always handle it correctly, but uses more disk space.

Both create a clean `vendor/` directory structure that's easy for build systems to use.

See [Build System Integration](docs/build-system-integration.md) for detailed examples with CMake, Make, etc.

## Documentation

See [cpkg.md](./cpkg.md) for the full specification.

See [MULTI_MODULE_DESIGN.md](./MULTI_MODULE_DESIGN.md) for details on multi-module support.

## Documentation

- [Build System Integration](docs/build-system-integration.md) - How to integrate cpkg with your build system
- [Multi-Module Support](docs/multi-module-design.md) - Using multiple modules from the same repository
- [Multiple Commits](docs/multi-module-commits.md) - How cpkg handles different commits for modules from the same repo
- [Vendor Directory](docs/vendor-explanation.md) - Understanding the vendor command and directory structure

## License

MIT

