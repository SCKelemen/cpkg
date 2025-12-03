# Vendor Directory Explanation

## Overview

The `cpkg vendor` command creates a flat directory structure containing all dependency source files. This is useful for build systems that expect a simple directory layout, or for CI environments where git submodules can be problematic.

## Command

```bash
cpkg vendor [--symlink|--copy]
```

- `--symlink` (default on Unix): Create symlinks to submodule files
- `--copy`: Copy files instead of symlinking

## Directory Structure

### Before Vendor

```
project/
├── cpkg.yaml
├── lock.cpkg.yaml
└── third_party/
    └── cpkg/
        └── github.com/
            └── user/
                └── repo/          # Git submodule
                    ├── module1/
                    │   ├── file1.c
                    │   └── file1.h
                    └── module2/
                        ├── file2.c
                        └── file2.h
```

### After Vendor (with subdirectories)

```
project/
├── cpkg.yaml
├── lock.cpkg.yaml
├── third_party/
│   └── cpkg/
│       └── github.com/
│           └── user/
│               └── repo/          # Git submodule (unchanged)
│                   ├── module1/
│                   └── module2/
└── vendor/
    └── github.com/
        └── user/
            └── repo/
                ├── module1/      # Symlink or copy
                │   ├── file1.c
                │   └── file1.h
                └── module2/      # Symlink or copy
                    ├── file2.c
                    └── file2.h
```

## How It Works

1. **Read lockfile**: `cpkg vendor` reads `lock.cpkg.yaml` to find all dependencies
2. **Get source paths**: For each dependency, it uses the `sourcePath` field from the lockfile
3. **Create vendor structure**: Creates the directory structure `vendor/<module-path>/`
4. **Link or copy**: Either symlinks or copies files from the submodule to the vendor directory

### Example Lockfile Entry

```yaml
dependencies:
  github.com/user/repo/module1:
    version: v1.0.0
    path: third_party/cpkg/github.com/user/repo/module1
    subdir: module1
    sourcePath: third_party/cpkg/github.com/user/repo/module1/module1
```

The `sourcePath` field points to the actual source files. When vendoring:
- Source: `third_party/cpkg/github.com/user/repo/module1/module1/`
- Destination: `vendor/github.com/user/repo/module1/`

## Symlink vs Copy

### Symlink (Default on Unix)

**Advantages:**
- Fast (no file copying)
- Low disk usage
- Always reflects submodule state
- Easy to update

**Disadvantages:**
- Not portable (doesn't work on Windows without special permissions)
- Can break if submodules are removed

### Copy

**Advantages:**
- Portable across all platforms
- Independent of submodule state
- Works in all environments

**Disadvantages:**
- Slower (file copying)
- Uses more disk space
- Must re-run to update

## Use Cases

### Development

Use submodules directly:
```bash
cpkg sync
# Build system uses third_party/cpkg/... directly
```

### CI/CD

Use vendor directory for reproducible builds:
```bash
cpkg tidy
cpkg vendor --copy  # Copy for portability
# Build system uses vendor/... directory
```

### Build Systems

Some build systems work better with a flat vendor directory:
```bash
cpkg vendor
# CMake/Make can simply include vendor/ in search paths
```

## Integration with Build Systems

### CMake Example

```cmake
# After cpkg vendor
target_include_directories(myapp PRIVATE vendor)
file(GLOB_RECURSE vendor_sources "vendor/**/*.c")
target_sources(myapp PRIVATE ${vendor_sources})
```

### Make Example

```make
# After cpkg vendor
CFLAGS += -Ivendor
SOURCES += $(wildcard vendor/**/*.c)
```

## Best Practices

1. **Use symlinks for development**: Faster and uses less space
2. **Use copy for CI**: More reliable in CI environments
3. **Don't commit vendor/**: Add `vendor/` to `.gitignore` (it's generated)
4. **Regenerate as needed**: Run `cpkg vendor` after `cpkg sync` or `cpkg upgrade`

## Workflow

```bash
# Initial setup
cpkg init
cpkg add github.com/user/repo@^1.0.0
cpkg tidy
cpkg sync

# Before building
cpkg vendor

# Build
make build
```

The vendor directory provides a clean, flat structure that build systems can easily consume, while the submodules remain the source of truth for version control.


