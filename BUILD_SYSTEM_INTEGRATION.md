# Build System Integration

## How cpkg Provides Source Files to Your Build System

cpkg manages dependencies and provides source files, but **does not control the compiler or linker**. Your build system (Make, CMake, Ninja, etc.) is responsible for compiling and linking.

## File Layout

### Submodule Layout

When you have a module from a repository, cpkg creates a git submodule:

```
third_party/cpkg/github.com/user/repo/
├── .git (submodule metadata)
├── intrusive_list/
│   ├── intrusive_list.h
│   └── intrusive_list.c
├── span/
│   ├── span.h
│   └── span.c
└── README.md
```

### Multi-Module Repos

For modules with subpaths (e.g., `github.com/user/repo/intrusive_list`):

1. **Submodule path**: `third_party/cpkg/github.com/user/repo/intrusive_list`
   - This is the entire repo checked out to a specific commit
   - Contains all files from the repo at that commit

2. **Source files location**: `third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list/`
   - The actual source files are in the subdirectory matching the module name
   - The lockfile includes a `subdir` field indicating this

### Lockfile Structure

```yaml
dependencies:
  github.com/user/repo/intrusive_list:
    version: v1.0.0
    commit: abc123...
    repoURL: https://github.com/user/repo.git
    path: third_party/cpkg/github.com/user/repo/intrusive_list  # Submodule path
    subdir: intrusive_list  # Subdirectory within the repo
    sourcePath: third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list  # ← Actual source files location
```

## How Your Build System Finds Files

### Option 1: Use the `sourcePath` Field (Recommended)

The lockfile includes a `sourcePath` field that points directly to where the source files are:

```python
# Python example (for CMake or build scripts)
import yaml

with open('cpkg.lock.yaml') as f:
    lock = yaml.safe_load(f)

for module, dep in lock['dependencies'].items():
    source_path = dep['sourcePath']  # Already computed: path + subdir
    # Add source_path to your build system
    print(f"Module {module}: sources in {source_path}")
```

**Note**: The `sourcePath` field is automatically computed by cpkg, so you don't need to manually combine `path` and `subdir`.

### Option 2: Use Environment Variables

cpkg sets `CPKG_DEP_ROOT` environment variable:

```bash
# In your build script
DEP_ROOT="${CPKG_DEP_ROOT:-third_party/cpkg}"

# For a module github.com/user/repo/intrusive_list
# Files are at: $DEP_ROOT/github.com/user/repo/intrusive_list/intrusive_list/
```

### Option 3: Use `cpkg vendor`

The `cpkg vendor` command creates a flat structure with only the needed files:

```bash
# Copy files (default)
cpkg vendor

# Or create symlinks (faster, no duplication)
cpkg vendor --symlink
```

This creates:
```
vendor/
└── github.com/
    └── user/
        └── repo/
            └── intrusive_list/
                ├── intrusive_list.h
                └── intrusive_list.c
```

**Symlinks vs Copying:**
- **Symlinks**: No disk duplication, always in sync with submodules, faster. Best for Unix/Linux/macOS.
- **Copying**: Works everywhere, more reliable with all build systems, uses more disk space.

Then your build system can use `vendor/` as a simple include path.

## CMake Example

```cmake
# Read cpkg.lock.yaml
file(READ "${CMAKE_SOURCE_DIR}/cpkg.lock.yaml" LOCKFILE_CONTENT)
string(REGEX MATCH "depRoot: ([^\n]+)" _ ${LOCKFILE_CONTENT})
set(CPKG_DEP_ROOT ${CMAKE_MATCH_1})
if(NOT CPKG_DEP_ROOT)
    set(CPKG_DEP_ROOT "third_party/cpkg")
endif()

# For each dependency, add source files
# Read sourcePath from lockfile (or use path + subdir)
# Example: sourcePath = "third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list"
set(INTRUSIVE_LIST_DIR "${CMAKE_SOURCE_DIR}/third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list")
add_library(intrusive_list STATIC
    ${INTRUSIVE_LIST_DIR}/intrusive_list.c
)
target_include_directories(intrusive_list PUBLIC ${INTRUSIVE_LIST_DIR})
```

## Makefile Example

```makefile
# Read CPKG_DEP_ROOT from environment or use default
CPKG_DEP_ROOT ?= third_party/cpkg

# Module paths
INTRUSIVE_LIST_DIR = $(CPKG_DEP_ROOT)/github.com/user/repo/intrusive_list/intrusive_list
SPAN_DIR = $(CPKG_DEP_ROOT)/github.com/user/repo/span/span

# Compile
intrusive_list.o: $(INTRUSIVE_LIST_DIR)/intrusive_list.c
	$(CC) -I$(INTRUSIVE_LIST_DIR) -c $< -o $@

span.o: $(SPAN_DIR)/span.c
	$(CC) -I$(SPAN_DIR) -c $< -o $@
```

## Key Points

1. **cpkg manages dependencies**: It resolves versions, checks out commits, and manages submodules.

2. **Your build system compiles**: You're responsible for:
   - Finding source files (use `sourcePath` from lockfile - it's already computed)
   - Adding include paths
   - Compiling and linking

3. **Subdirectories matter**: For multi-module repos, the actual source files are in a subdirectory within the submodule checkout.

4. **Use `cpkg vendor` for simplicity**: If you want a flat file structure without submodules, use `cpkg vendor` to copy files to a `vendor/` directory.

## Environment Variables

cpkg sets these environment variables when running build/test commands:

- `CPKG_ROOT`: Root directory of your project (where `cpkg.yaml` is)
- `CPKG_DEP_ROOT`: Dependency root directory (default: `third_party/cpkg`)
- `CPKG_TARGET`: Build target (if specified)

Your build system can use these to locate dependencies.

