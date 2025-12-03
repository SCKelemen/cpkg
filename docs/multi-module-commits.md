# Multiple Commits for Modules from the Same Repository

## Problem

When multiple modules from the same repository need different git commits (e.g., different versions), git submodules present a challenge: each submodule can only point to a single commit.

## Solution

cpkg creates **separate submodules** for each module path, even when they come from the same repository. This allows each module to be at its own commit independently.

## Example

Consider a repository with two independently versioned modules:

```yaml
# cpkg.yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: "^1.0.0"  # Tag: intrusive_list/v1.0.0 → commit abc123
  github.com/user/firmware-lib/span:
    version: "^1.0.0"  # Tag: span/v1.0.0 → commit def456
```

### Resolution Process

1. **`cpkg tidy`**: Resolves versions and commits
   - `intrusive_list/v1.0.0` → commit `abc123`
   - `span/v1.0.0` → commit `def456`
   - Creates separate entries in the lockfile

2. **`cpkg sync`**: Creates separate submodules
   - Submodule at `third_party/cpkg/github.com/user/firmware-lib/intrusive_list`
     - Repository: `https://github.com/user/firmware-lib.git`
     - Commit: `abc123`
   - Submodule at `third_party/cpkg/github.com/user/firmware-lib/span`
     - Repository: `https://github.com/user/firmware-lib.git`
     - Commit: `def456`

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
                    │   ├── intrusive_list.h
                    │   └── intrusive_list.c
                    └── span/            # Submodule at commit def456
                        ├── span.h
                        └── span.c
```

### .gitmodules

```ini
[submodule "third_party/cpkg/github.com/user/firmware-lib/intrusive_list"]
    path = third_party/cpkg/github.com/user/firmware-lib/intrusive_list
    url = https://github.com/user/firmware-lib.git

[submodule "third_party/cpkg/github.com/user/firmware-lib/span"]
    path = third_party/cpkg/github.com/user/firmware-lib/span
    url = https://github.com/user/firmware-lib.git
```

## Why Separate Submodules?

Git submodules can only point to a **single commit** per submodule entry. If you tried to use one submodule for both modules:

- Checking out `abc123` would give you `intrusive_list` at the correct version
- Checking out `def456` would overwrite it, losing the correct `intrusive_list` version

By using separate submodules, each module can be at its own commit independently.

## Trade-offs

### Advantages

- ✅ Each module can be at a different commit
- ✅ Independent versioning works correctly
- ✅ No conflicts between module versions
- ✅ Clear separation in `.gitmodules`

### Disadvantages

- ⚠️ Multiple submodule entries for the same repository (more entries in `.gitmodules`)
- ⚠️ More disk space (each submodule is a full clone)
- ⚠️ More git operations during sync

## Alternative Approaches (Future Considerations)

These approaches could be explored in future releases to optimize submodule usage:

### Git Worktrees

Use git worktrees to have multiple checkouts from one submodule.

- **Pros**: Single submodule entry, multiple checkouts at different commits
- **Cons**: More complex to manage, requires git 2.5+, worktrees are less familiar to users
- **Status**: Not implemented - current approach (separate submodules) is simpler and more explicit

### Sparse Checkout

Checkout only needed paths from a submodule.

- **Pros**: Smaller disk usage, faster checkouts
- **Cons**: Doesn't solve the different commits problem, adds complexity
- **Status**: Not implemented - could be useful optimization for large repos with single-commit modules

## Conclusion

The current approach (separate submodules) is the most straightforward and reliable solution. It provides clear separation, works with standard git tools, and ensures each module can be independently versioned.

