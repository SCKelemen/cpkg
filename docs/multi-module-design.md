# Multi-Module Support Design

## Overview

cpkg supports multiple independently versioned modules/packages from the same git repository. This enables monorepo-style organization where related components can be versioned and imported separately.

## Use Cases

1. **Data structures library**: A repository with multiple data structures (`intrusive_list`, `span`, `view`) that can be imported independently
2. **Protocol implementations**: A repository with multiple protocol clients (`brski`, `est`, `fdo`, `oadopt`) that are versioned separately
3. **Incremental adoption**: Use arbitrary subdirectories from any repository without requiring upstream support
4. **Version flexibility**: Use different versions of different subdirectories from the same repository

## Module Path Format

Module paths can include a subpath to specify a module within a repository:

```
github.com/user/repo/subpath
```

Examples:
- `github.com/user/firmware-lib/intrusive_list`
- `github.com/user/protocols/brski/client`
- `github.com/user/protocols/oadopt/client`

## Tag Naming Conventions

### For Repositories You Control

If you control the repository, you can create tags for each submodule. Two tag formats are supported:

1. **Prefix format** (recommended): `subpath/v1.0.0`
   - Example: `intrusive_list/v1.0.0`, `brski/v2.1.0`
   - Clear separation, easy to filter

2. **Suffix format**: `v1.0.0-subpath`
   - Example: `v1.0.0-intrusive_list`, `v2.1.0-brski`
   - Alternative for repositories that prefer version-first

### Fallback for Arbitrary Repositories

If a subpath doesn't have its own tags, cpkg falls back to using the root repository tags. This allows you to use any subdirectory from any repository without requiring upstream changes:

- Module: `github.com/Mbed-TLS/mbedtls/library`
- No tags like `library/v3.6.0` found
- Falls back to root tags: `v3.6.0`, `v3.6.5`, etc.
- Uses the selected tag and points to the `library/` subdirectory

## Implementation Details

### Module Path Parsing

Module paths are parsed to extract the repository URL and optional subpath:

```go
type ModulePath struct {
    RepoURL string  // github.com/user/repo
    Subpath string  // intrusive_list (empty for root module)
    Full    string  // github.com/user/repo/intrusive_list
}
```

The parser uses a heuristic: if a path has 3+ segments, the first 3 are assumed to be the repository (host/owner/repo), and everything after is the subpath.

### Tag Filtering

When resolving dependencies, cpkg filters tags that match the subpath:

1. **Prefix format**: Tags starting with `subpath/`
2. **Suffix format**: Tags ending with `-subpath`
3. **Fallback**: If no subpath tags found, use root repository tags

### Version Extraction

When parsing tags with subpaths, the version part is extracted:

- Prefix format: `subpath/v1.0.0` → `v1.0.0`
- Suffix format: `v1.0.0-subpath` → `v1.0.0`

### Dependency Path

The dependency path in the lockfile includes the subpath:

```yaml
dependencies:
  github.com/user/repo/intrusive_list:
    version: v1.0.0
    path: third_party/cpkg/github.com/user/repo/intrusive_list
    subdir: intrusive_list
    sourcePath: third_party/cpkg/github.com/user/repo/intrusive_list/intrusive_list
```

The submodule points to the entire repository, but the path reflects the subpath, and `sourcePath` points to the actual source files.

### Submodule Handling

**Important**: Each module from the same repository gets its own submodule. This is necessary because git submodules can only point to a single commit, and independently versioned modules may need different commits.

For example:
- `github.com/user/repo/intrusive_list` → submodule at `third_party/cpkg/github.com/user/repo/intrusive_list` (commit A)
- `github.com/user/repo/span` → submodule at `third_party/cpkg/github.com/user/repo/span` (commit B)

Each submodule is a separate checkout of the same repository, allowing different modules to be at different commits.

**This works even for arbitrary subdirectories:**

```yaml
dependencies:
  github.com/Mbed-TLS/mbedtls/library:
    version: "3.6.4"  # Points to commit c765c83
  github.com/Mbed-TLS/mbedtls/include:
    version: "3.6.5"  # Points to commit e185d7f
```

Both submodules are separate checkouts of the same repository, but at different commits. The `sourcePath` in the lockfile points to the subdirectory within each checkout.

## Complete Example

### Repository Structure

```
github.com/user/firmware-lib/
├── intrusive_list/
│   ├── intrusive_list.h
│   └── intrusive_list.c
├── span/
│   ├── span.h
│   └── span.c
└── view/
    ├── view.h
    └── view.c
```

### Git Tags

```bash
git tag intrusive_list/v1.0.0
git tag intrusive_list/v1.1.0
git tag span/v1.0.0
git tag view/v1.0.0
```

### cpkg.yaml

```yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: "^1.0.0"
  github.com/user/firmware-lib/span:
    version: "^1.0.0"
  github.com/user/firmware-lib/view:
    version: "^1.0.0"
```

### cpkg.lock.yaml

```yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: v1.1.0
    commit: abc123...
    repoURL: https://github.com/user/firmware-lib.git
    path: third_party/cpkg/github.com/user/firmware-lib/intrusive_list
    subdir: intrusive_list
    sourcePath: third_party/cpkg/github.com/user/firmware-lib/intrusive_list/intrusive_list
  github.com/user/firmware-lib/span:
    version: v1.0.0
    commit: def456...
    repoURL: https://github.com/user/firmware-lib.git
    path: third_party/cpkg/github.com/user/firmware-lib/span
    subdir: span
    sourcePath: third_party/cpkg/github.com/user/firmware-lib/span/span
```

### File System Layout

```
project/
├── cpkg.yaml
├── cpkg.lock.yaml
└── third_party/
    └── cpkg/
        └── github.com/
            └── user/
                └── firmware-lib/
                    ├── intrusive_list/  # Submodule at commit abc123
                    │   └── intrusive_list/
                    │       ├── intrusive_list.h
                    │       └── intrusive_list.c
                    └── span/            # Submodule at commit def456
                        └── span/
                            ├── span.h
                            └── span.c
```

## Lockfile Semantics

**cpkg.lock.yaml** combines the roles of **go.mod** and **go.sum**:

- **Version locking**: Pins exact versions and commits (like `go.mod`)
- **Integrity checking**: Includes checksums to verify dependency integrity (like `go.sum`)
- **Single file**: Simpler than maintaining separate files

The lockfile serves both purposes:
- **Version locking**: Pins exact versions and commits
- **Integrity checking**: Includes checksums to verify dependency integrity

## Backward Compatibility

- Module paths without subpaths continue to work (root module)
- Tags without subpath prefixes continue to work
- Existing lockfiles remain valid

## Future Considerations

1. **Tag discovery**: Could support a `cpkg.tags` file in the repository root that maps subpaths to tag patterns
2. **Multiple tag formats**: Could support custom tag formats via configuration
3. **Shared dependencies**: Modules from the same repository could share a single submodule checkout (optimization)

