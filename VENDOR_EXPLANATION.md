# How Vendor Works with Multi-Module Repos

## Your Understanding is Correct! ✅

Yes, exactly as you described:

1. **Submodules manage versioning** - Each module gets its own submodule (even from the same repo)
2. **Vendor copies only needed files** - From each submodule's subdirectory
3. **Output contains only subsets** - Just the relevant files for each package

## Visual Example

### Step 1: After `cpkg sync` (Submodules)

You have multiple submodules, each at different commits:

```
third_party/cpkg/github.com/user/firmware-lib/
├── intrusive_list/          # Submodule #1: entire repo at commit abc123
│   ├── .git (submodule)
│   ├── intrusive_list/       # ← The actual source files
│   │   ├── intrusive_list.h
│   │   └── intrusive_list.c
│   ├── span/                # ← Other modules (not needed)
│   ├── view/                # ← Other modules (not needed)
│   └── README.md
│
└── span/                    # Submodule #2: entire repo at commit def456
    ├── .git (submodule)
    ├── intrusive_list/      # ← Other modules (not needed)
    ├── span/                # ← The actual source files
    │   ├── span.h
    │   └── span.c
    ├── view/                # ← Other modules (not needed)
    └── README.md
```

**Note**: Yes, the same repo is cloned twice (once per module), each at a different commit.

### Step 2: After `cpkg vendor` (Vendored Files)

Only the relevant subdirectories are copied:

```
vendor/
└── github.com/
    └── user/
        └── firmware-lib/
            ├── intrusive_list/      # ← Only files from submodule #1's intrusive_list/
            │   ├── intrusive_list.h
            │   └── intrusive_list.c
            └── span/                # ← Only files from submodule #2's span/
                ├── span.h
                └── span.c
```

**Key points:**
- ✅ No duplicate repos in vendor/
- ✅ Only the needed subdirectories
- ✅ Each package's files come from its own submodule checkout (at the correct commit)

## How It Works in Code

### Lockfile Structure

```yaml
dependencies:
  github.com/user/firmware-lib/intrusive_list:
    version: v1.0.0
    commit: abc123...
    path: third_party/cpkg/github.com/user/firmware-lib/intrusive_list
    subdir: intrusive_list  # ← Tells vendor where to find files
    
  github.com/user/firmware-lib/span:
    version: v1.0.0
    commit: def456...
    path: third_party/cpkg/github.com/user/firmware-lib/span
    subdir: span  # ← Tells vendor where to find files
```

### Vendor Process

For each dependency:
1. **Source**: `path` + `subdir` = `third_party/cpkg/.../intrusive_list/intrusive_list/`
2. **Destination**: `vendor/github.com/user/firmware-lib/intrusive_list/`
3. **Copy**: Only the `intrusive_list/` subdirectory from that specific submodule checkout

## Why This Design?

### Submodules (for versioning)
- ✅ Each module can be at a different commit
- ✅ Git tracks the exact commit for each
- ✅ Reproducible builds

### Vendor (for building)
- ✅ Flat structure, easy for build systems
- ✅ Only needed files (no extra repos)
- ✅ No submodule complexity in build

## Trade-offs

**Pros:**
- ✅ Independent versioning works perfectly
- ✅ Vendor output is clean (only needed files)
- ✅ Build system doesn't need to understand submodules

**Cons:**
- ⚠️ Multiple submodules from same repo (more disk space)
- ⚠️ More git operations during sync

But for your use case (small C libraries), this is totally fine!

## Summary

Your understanding is **100% correct**:

> "We use submodules to manage versioning, but then we can use vendor to pull the files needed for building. So we might have the same repo cloned several times in the submodules, once for each package, but then when we call vendor, we only copy the relevant files for each package, from each package's copy of the repo, so the output file only contains the subsets from each repo tag."

Exactly! 🎯

