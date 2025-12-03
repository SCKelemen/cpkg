# Multiple Commits for Modules from Same Repo

## How It Works

When you have multiple modules from the same repository that need different git commits, cpkg creates **separate submodules** for each module path.

### Example Scenario

You have a repository with two independently versioned modules:

```yaml
# cpkg.yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: "^1.0.0"  # Tag: intrusive_list/v1.0.0 → commit abc123
  github.com/user/firmware-lib/span:
    version: "^1.0.0"  # Tag: span/v1.0.0 → commit def456
```

### What Happens

1. **Resolution (`cpkg tidy`)**: 
   - Resolves `intrusive_list/v1.0.0` → commit `abc123`
   - Resolves `span/v1.0.0` → commit `def456`
   - Creates separate entries in the lockfile

2. **Sync (`cpkg sync`)**:
   - Creates submodule at `third_party/cpkg/github.com/user/firmware-lib/intrusive_list`
     - Points to repo: `https://github.com/user/firmware-lib.git`
     - Checked out to commit: `abc123`
   - Creates submodule at `third_party/cpkg/github.com/user/firmware-lib/span`
     - Points to repo: `https://github.com/user/firmware-lib.git`
     - Checked out to commit: `def456`

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

### Why Separate Submodules?

Git submodules can only point to a **single commit** per submodule. If you tried to use one submodule for both modules:
- Checking out `abc123` would give you `intrusive_list` at the right version
- But then checking out `def456` would overwrite it, losing the correct `intrusive_list` version

By using separate submodules, each module can be at its own commit independently.

### Trade-offs

**Pros:**
- ✅ Each module can be at a different commit
- ✅ Independent versioning works correctly
- ✅ No conflicts between module versions

**Cons:**
- ⚠️ Multiple submodules from the same repo (more entries in `.gitmodules`)
- ⚠️ More disk space (each submodule is a full clone)
- ⚠️ More git operations during sync

### Alternative Approaches (Future Ideas)

These approaches could be explored in future releases to optimize submodule usage:

1. **Git Worktrees**: Use git worktrees to have multiple checkouts from one submodule
   - **Pros**: Single submodule entry, multiple checkouts at different commits
   - **Cons**: More complex to manage, requires git 2.5+, worktrees are less familiar to users
   - **Status**: Not implemented - current approach (separate submodules) is simpler and more explicit

2. **Sparse Checkout**: Checkout only needed paths from a submodule
   - **Pros**: Smaller disk usage, faster checkouts
   - **Cons**: Doesn't solve the different commits problem, adds complexity
   - **Status**: Not implemented - could be useful optimization for large repos with single-commit modules

3. **Accept Limitation**: Require all modules from same repo to be at same commit
   - Simplest, but loses independent versioning

The current approach (separate submodules) is the most straightforward and reliable solution.

