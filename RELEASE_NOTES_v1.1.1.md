# cpkg v1.1.1 Release Notes

## Overview

cpkg v1.1.1 is a patch release that improves CI/CD workflows, adds release notes automation, and fixes various workflow issues. This release focuses on improving the release process and developer experience.

## Installation

### From Source (Recommended)

```bash
go install github.com/SCKelemen/cpkg@v1.1.1
```

### Pre-built Binaries

Download pre-built binaries for your platform from the [Releases](https://github.com/SCKelemen/cpkg/releases) page:

- **Linux**: `cpkg-linux-amd64`
- **macOS**: `cpkg-darwin-amd64`

After downloading, make the binary executable and move it to your PATH:

```bash
# Linux/macOS
chmod +x cpkg-linux-amd64  # or cpkg-darwin-amd64
sudo mv cpkg-linux-amd64 /usr/local/bin/cpkg  # or cpkg-darwin-amd64
```

## What's New

### Improvements

#### 1. Automated Release Notes

Release notes are now automatically included in GitHub release descriptions. When you create a release tag, the workflow will:
- Look for `RELEASE_NOTES_v{VERSION}.md` in the repository
- Include it in the GitHub release description
- Fall back to auto-generated notes if the file doesn't exist

This makes release pages much more informative and user-friendly.

#### 2. CI/CD Enhancements

- **Removed Windows support**: Simplified CI/CD by removing Windows builds for faster, more reliable CI runs
- **Updated Go versions**: Now testing with Go 1.21, 1.23, and 1.25
- **Better error handling**: Improved Go installation verification and race detector fallback
- **GOPROXY configuration**: Explicit GOPROXY setup to prevent dependency download issues

#### 3. Workflow Fixes

- Fixed Windows compatibility issues in workflows (before Windows removal)
- Added Go cache cleanup to reduce toolchain extraction warnings
- Improved binary verification steps
- Better handling of Go standard library access issues

## Bug Fixes

- Fixed GOPROXY configuration issues that could cause dependency download failures
- Fixed race detector failures by making it optional with fallback
- Fixed Go toolchain extraction warnings in CI logs

## Technical Details

### CI/CD Changes

- **Platforms**: Linux and macOS only (Windows removed for simplicity)
- **Go versions tested**: 1.21, 1.23, 1.25
- **Release automation**: Release notes are automatically included from `RELEASE_NOTES_v{VERSION}.md` files

### Release Process

The release workflow now:
1. Extracts version from git tag
2. Looks for corresponding `RELEASE_NOTES_v{VERSION}.md` file
3. Includes it in the GitHub release description
4. Falls back to basic notes if file not found

## Migration Guide

### From v1.1.0 to v1.1.1

No breaking changes. This is a patch release with improvements and bug fixes.

**No action required** - you can upgrade directly:

```bash
go install github.com/SCKelemen/cpkg@v1.1.1
```

Or if using pre-built binaries, download the new version from the [Releases](https://github.com/SCKelemen/cpkg/releases) page.

## Requirements

- Go 1.21 or later
- Git (for submodule management)
- A C build system (Make, CMake, Ninja, etc.)

## Full Changelog

For a complete list of changes, see the [GitHub compare view](https://github.com/SCKelemen/cpkg/compare/v1.1.0...v1.1.1).

