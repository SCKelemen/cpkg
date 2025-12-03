# Multi-Module Support Design

## Overview

Support multiple independently versioned modules/packages from the same git repository. This enables monorepo-style organization where related components can be versioned and imported separately.

## Use Cases

1. **Data structures library**: A repo with multiple data structures (`intrusive_list`, `span`, `view`) that can be imported independently
2. **Protocol implementations**: A repo with multiple protocol clients (`brski`, `est`, `fdo`, `oadopt`) that are versioned separately

## Design

### Module Path Format

Module paths can include a subpath to specify a module within a repository:

```
github.com/user/repo/subpath
```

Examples:
- `github.com/user/firmware-lib/intrusive_list`
- `github.com/user/protocols/brski/client`
- `github.com/user/protocols/oadopt/client`

### Tag Naming Conventions

Two tag formats are supported:

1. **Prefix format** (recommended): `subpath/v1.0.0`
   - Example: `intrusive_list/v1.0.0`, `brski/v2.1.0`
   - Clear separation, easy to filter

2. **Suffix format**: `v1.0.0-subpath`
   - Example: `v1.0.0-intrusive_list`, `v2.1.0-brski`
   - Alternative for repos that prefer version-first

### Implementation Details

#### 1. Module Path Parsing

```go
type ModulePath struct {
    RepoURL string  // github.com/user/repo
    Subpath string  // intrusive_list (empty for root module)
    Full    string  // github.com/user/repo/intrusive_list
}

func ParseModulePath(path string) ModulePath {
    // Split on '/' and identify where repo ends and subpath begins
    // This is heuristic-based - we assume the repo part is at least 2 segments
    // (host/owner/repo) and everything after is the subpath
}
```

#### 2. Tag Filtering

When resolving dependencies, filter tags that match the subpath:

```go
func filterTagsForSubpath(tags []string, subpath string) []string {
    var filtered []string
    
    for _, tag := range tags {
        // Try prefix format: subpath/v1.0.0
        if strings.HasPrefix(tag, subpath+"/") {
            filtered = append(filtered, tag)
            continue
        }
        
        // Try suffix format: v1.0.0-subpath
        if strings.HasSuffix(tag, "-"+subpath) {
            filtered = append(filtered, tag)
            continue
        }
    }
    
    return filtered
}
```

#### 3. Tag Parsing

When parsing tags with subpaths, extract the version part:

```go
func extractVersionFromTag(tag, subpath string) (string, error) {
    // Prefix format: subpath/v1.0.0 -> v1.0.0
    if strings.HasPrefix(tag, subpath+"/") {
        return strings.TrimPrefix(tag, subpath+"/"), nil
    }
    
    // Suffix format: v1.0.0-subpath -> v1.0.0
    if strings.HasSuffix(tag, "-"+subpath) {
        return strings.TrimSuffix(tag, "-"+subpath), nil
    }
    
    return "", fmt.Errorf("tag does not match subpath format")
}
```

#### 4. Dependency Path

The dependency path in the lockfile includes the subpath:

```yaml
dependencies:
  github.com/user/repo/intrusive_list:
    version: v1.0.0
    path: third_party/cpkg/github.com/user/repo/intrusive_list
```

The submodule still points to the entire repo, but the path reflects the subpath.

#### 5. Submodule Handling

**Important**: Each module from the same repo gets its own submodule. This is necessary because git submodules can only point to a single commit, and independently versioned modules may need different commits.

For example:
- `github.com/user/repo/intrusive_list` → submodule at `third_party/cpkg/github.com/user/repo/intrusive_list` (commit A)
- `github.com/user/repo/span` → submodule at `third_party/cpkg/github.com/user/repo/span` (commit B)

Each submodule is a separate checkout of the same repository, allowing different modules to be at different commits.

## Example

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
  github.com/user/firmware-lib/span:
    version: v1.0.0
    commit: def456...
    repoURL: https://github.com/user/firmware-lib.git
    path: third_party/cpkg/github.com/user/firmware-lib/span
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
                └── firmware-lib/  # git submodule
                    ├── intrusive_list/
                    ├── span/
                    └── view/
```

## Lockfile vs Go's go.sum

**cpkg.lock.yaml** is more like **go.mod + go.sum combined**:
- It locks concrete versions (like go.mod)
- It includes checksums (like go.sum)
- It's a single file for simplicity

The lockfile serves both purposes:
- **Version locking**: Pins exact versions and commits
- **Integrity checking**: Includes checksums to verify dependency integrity

## Backward Compatibility

- Module paths without subpaths continue to work (root module)
- Tags without subpath prefixes continue to work
- Existing lockfiles remain valid

## Future Considerations

1. **Tag discovery**: Could support a `cpkg.tags` file in the repo root that maps subpaths to tag patterns
2. **Multiple tag formats**: Could support custom tag formats via configuration
3. **Shared dependencies**: Modules from the same repo could share a single submodule checkout (optimization)

