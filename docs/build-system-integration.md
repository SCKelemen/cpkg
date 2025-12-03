# Build System Integration

## Overview

cpkg is a source-only package manager. It does not compile code or manage build artifacts. Instead, cpkg provides a minimal contract: it resolves dependencies and makes source files available at known paths. Your build system (Make, CMake, Meson, etc.) is responsible for compiling and linking.

## Integration Contract

### Source File Location

After running `cpkg sync` and `cpkg vendor`, source files are available at predictable paths:

1. **Submodule paths**: Source files are in git submodules at `third_party/cpkg/<module-path>/<subdir>/`
2. **Vendor paths**: After `cpkg vendor`, files are symlinked or copied to `vendor/<module-path>/<subdir>/`

The `cpkg.lock.yaml` file contains a `sourcePath` field for each dependency, which points to the exact location of source files:

```yaml
dependencies:
  github.com/user/repo:
    version: v1.0.0
    sourcePath: third_party/cpkg/github.com/user/repo
  github.com/user/repo/submodule:
    version: v1.0.0
    subdir: submodule
    sourcePath: third_party/cpkg/github.com/user/repo/submodule/submodule
```

### Reading the Lockfile

Build systems can read `cpkg.lock.yaml` to discover source file locations:

```yaml
apiVersion: cpkg.ringil.dev/v0
kind: Lockfile
dependencies:
  github.com/user/repo:
    version: v1.0.0
    commit: abc123...
    sourcePath: third_party/cpkg/github.com/user/repo
```

The `sourcePath` field always points to the directory containing the actual source files, accounting for subdirectories when using multi-module support.

## Build System Examples

### CMake

```cmake
# Read cpkg.lock.yaml to get source paths
find_package(yaml-cpp REQUIRED)
include(yaml-cpp/yaml-cpp.cmake)

# Load lockfile
yaml-cpp::yaml-cpp lockfile
lockfile.load_file("cpkg.lock.yaml")

# Get dependencies
auto deps = lockfile["dependencies"];

# Add include directories and source files
foreach(dep IN LISTS deps)
    get_filename_component(dep_path ${deps[dep]["sourcePath"]} ABSOLUTE)
    target_include_directories(myapp PRIVATE ${dep_path})
    file(GLOB_RECURSE dep_sources "${dep_path}/*.c" "${dep_path}/*.h")
    target_sources(myapp PRIVATE ${dep_sources})
endforeach()
```

### Make

```make
# Parse cpkg.lock.yaml to extract source paths
# (using a simple script or yq/jq)

CPKG_LOCK := cpkg.lock.yaml
DEPS := $(shell grep -A 5 "sourcePath:" $(CPKG_LOCK) | grep "sourcePath:" | cut -d: -f2 | tr -d ' ')

# Add to include paths
CFLAGS += $(foreach dep,$(DEPS),-I$(dep))

# Add source files
SOURCES += $(foreach dep,$(DEPS),$(wildcard $(dep)/*.c))
```

### Using the Vendor Directory

If you use `cpkg vendor`, source files are in the `vendor/` directory:

```cmake
# Vendor directory structure
vendor/
  github.com/
    user/
      repo/
        source.c
        header.h

# In CMake
file(GLOB_RECURSE vendor_sources "vendor/**/*.c")
target_sources(myapp PRIVATE ${vendor_sources})
target_include_directories(myapp PRIVATE vendor)
```

## Workflow

1. **Development**: Use submodules (`cpkg sync`) for version control integration
2. **CI/Build**: Use vendor directory (`cpkg vendor`) for reproducible builds
3. **Build system**: Read `sourcePath` from lockfile or use vendor directory

## Best Practices

- **Use vendor for CI**: Run `cpkg vendor` in CI to avoid submodule initialization issues
- **Use symlinks by default**: Faster and uses less disk space (Unix-like systems)
- **Use copy for portability**: Use `cpkg vendor --copy` if symlinks cause issues
- **Read lockfile programmatically**: Parse `cpkg.lock.yaml` to discover dependencies automatically
- **Include paths**: Add `vendor/` or individual `sourcePath` directories to your compiler's include path

## Example: Full Integration

```bash
# In your build script
cpkg tidy    # Resolve dependencies
cpkg sync     # Initialize submodules (or skip for CI)
cpkg vendor   # Create vendor directory

# Your build system then uses vendor/ or sourcePath from lockfile
make build
```

The build system only needs to know:
- Source files are in `vendor/` (after `cpkg vendor`)
- Or read `sourcePath` from `cpkg.lock.yaml` for each dependency

cpkg handles the rest: dependency resolution, version locking, and file layout.

