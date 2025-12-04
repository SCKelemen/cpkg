# cpkg — Source-Only Module & Dependency Manager for C/SKC (v0)

A minimal, Go-implemented module and dependency manager for C (and SKC firmware) inspired by Go modules and Elm packages.

* **Source-only**: no prebuilt binaries, no ABI matrix.
* **Git-first**: modules are git repos, addressed by URL-ish paths.
* **Semver**: dependencies are versioned with semantic version tags.
* **Build-agnostic**: cpkg resolves and lays out source; your build system compiles it.
* **Submodule-native**: cpkg wraps git submodules to keep repos clean and reviewable.

This document describes the v0 specification: file formats and CLI behavior.

---

## 0. Integration Contract

**cpkg's job is not to teach the compiler about modules.** The compiler/linker don't know anything about "modules" – they just see `.c`/`.h` files.

Instead, **cpkg guarantees that "the right files" are present at a known path, at a known version.**

### The Minimal Contract

For each dependency module in `lock.cpkg.yaml`, `cpkg sync` guarantees:

* There is a directory at `<depRoot>/<module-path>/` (e.g., `third_party/cpkg/github.com/ringil/firmware-ds/span/`)
* That directory is a git checkout at the **exact commit** pinned in the lockfile
* No other version of those files exists in your tree unless you put it there

That's it. No magic compiler integration. Just:

> "If your build system looks in `third_party/cpkg/github.com/ringil/firmware-ds/span`, those files are definitely the right version."

From there, *any* build chain (CMake, Make, vendor IDE project files, Bazel, etc.) can just treat these directories like normal source trees.

### Multi-Module Repos

For modules with subpaths (e.g., `github.com/ringil/firmware-ds/intrusive-list`):

* The submodule is at `<depRoot>/<module-path>/` (entire repo checkout)
* The actual source files are in a subdirectory within that checkout
* The lockfile's `subdir` field indicates where the files are
* Your build system should use `path + subdir` to locate source files

Example:
```yaml
# lock.cpkg.yaml
dependencies:
  github.com/ringil/firmware-ds/intrusive-list:
    path: third_party/cpkg/github.com/ringil/firmware-ds/intrusive-list
    subdir: intrusive-list  # ← Files are in this subdirectory
```

Build system usage:
```cmake
# CMake example
set(INTRUSIVE_LIST_DIR 
    "${CPKG_DEP_ROOT}/github.com/ringil/firmware-ds/intrusive-list/intrusive-list")
add_library(ds_intrusive_list ${INTRUSIVE_LIST_DIR}/intrusive_list.c)
target_include_directories(ds_intrusive_list PUBLIC ${INTRUSIVE_LIST_DIR})
```

### Alternative: Use `cpkg vendor`

For a simpler flat structure without submodules:

```bash
cpkg vendor
```

This copies only the needed files to `vendor/`, making it easier for build systems to find them.

---

## 1. Goals

1. **Deterministic dependencies**

   * Every project has a manifest (`cpkg.yaml`) and a lockfile (`lock.cpkg.yaml`).
   * Lockfile pins each dependency to a concrete version + commit + checksum.

2. **Git-first, source-only**

   * v0 supports **only git-based modules**.
   * cpkg fetches source trees; it does not build binaries.

3. **Build-agnostic**

   * cpkg does **not** replace Make/CMake/Ninja/etc.
   * `cpkg build` and `cpkg test` are thin wrappers that:

     1. Resolve dependencies.
     2. Sync submodules.
     3. Run project-defined commands.

4. **Submodule-friendly**

   * Dependencies are laid out under a configurable root (default `third_party/cpkg/`).
   * cpkg manages `.gitmodules` and submodule SHAs for you.

5. **Internal-first**

   * Designed to start as an internal tool.
   * External deps can be forked; versions are tagged in internal forks.

---

## 2. Files

### 2.1 `cpkg.yaml` — Module Manifest

**File name:** `cpkg.yaml`

YAML manifest describing the module and its direct dependencies.

**Module Discovery**: cpkg uses Go-style module discovery. When you run `cpkg` commands, it automatically finds the nearest `cpkg.yaml` by walking up the directory tree from your current working directory. This means you can have multiple `cpkg.yaml` files in subdirectories, each treated as an independent module with its own `lock.cpkg.yaml` and dependency graph. Commands operate on only one module at a time (the one whose manifest is found).

**Note**: Unlike Go's workspace mode, cpkg does not support operating on multiple modules simultaneously in a single command invocation. Each module is completely independent.

```yaml
apiVersion: cpkg.ringil.dev/v0
kind: Module

module: github.com/ringil/device-fw
# Optional: only for reusable libraries. Apps/firmware may omit or use 0.0.0
version: 0.1.0

# Root directory for resolved dependencies within this repo.
# Default: third_party/cpkg
depRoot: third_party/cpkg

language:
  cStandard: c23       # or c17, etc
  skc: true            # follows SKC style/rules

# Optional per-target build wrappers for `cpkg build`.
build:
  # Default build command (used when no --target is passed)
  command: ["ninja", "-C", "build"]

  # Optional target-specific build commands
  targets:
    host:
      command: ["ninja", "-C", "build-host"]
    stm32f4:
      command: ["ninja", "-C", "build-stm32f4"]

# Optional test wrapper for `cpkg test`.
test:
  command: ["ctest", "--test-dir", "build-host"]

# Dependencies: module path -> semver constraint string.
# v0: git-only, https assumed unless overridden.
dependencies:
  github.com/ringil/wolfssl-fork:
    version: "^5.7.0"
  github.com/ringil/mbedtls-fork:
    version: "^3.5.0"
  git.internal/ringil/stm32-hal-skc:
    version: "^1.1.0"
  git.internal/ringil/stsafe-a110:
    version: "^1.0.0"
```

#### 2.1.1 Field semantics

* `apiVersion`: Schema version for the manifest (`cpkg.ringil.dev/v0`).
* `kind`: Always `Module` for v0.
* `module`: Module path (URL-ish). In v0, this must correspond to a git repo.
* `version`: Semantic version for this module itself (optional for applications).
* `depRoot`: Directory under which dependencies are laid out.

  * Default: `third_party/cpkg` if omitted.
* `language`:

  * `cStandard`: e.g., `c23`, `c17`.
  * `skc`: Boolean indicating SKC-style code; informational for tooling.
* `build`:

  * `command`: Default command to run for `cpkg build`.
  * `targets`: Map of target-name → `{ command: [...] }` for target-specific builds.
* `test`:

  * `command`: Command to run for `cpkg test`.
* `dependencies`:

  * Keys: module paths.
  * Values: objects with at least `version` (a semver range string).

### 2.2 `lock.cpkg.yaml` — Lockfile

**File name:** `lock.cpkg.yaml`

Generated by `cpkg tidy`. MUST NOT be edited manually.

```yaml
# lock.cpkg.yaml — generated, do not edit by hand.

apiVersion: cpkg.ringil.dev/v0
kind: Lockfile

module: github.com/ringil/device-fw
generatedBy: cpkg 0.1.0
generatedAt: 2025-12-03T12:34:56Z

depRoot: third_party/cmod

dependencies:
  github.com/ringil/wolfssl-fork:
    version: v5.7.3          # concrete semver tag
    commit:  1234567890abcdef1234567890abcdef12345678
    sum:     h1:base64-checksum-of-tree
    vcs:     git
    repoURL: https://github.com/ringil/wolfssl-fork.git
    path:    third_party/cpkg/github.com/ringil/wolfssl-fork

  github.com/ringil/mbedtls-fork:
    version: v3.5.2
    commit:  abcdef0123456789abcdef0123456789abcdef01
    sum:     h1:another-checksum
    vcs:     git
    repoURL: https://github.com/ringil/mbedtls-fork.git
    path:    third_party/cpkg/github.com/ringil/mbedtls-fork

  git.internal/ringil/stm32-hal-skc:
    version: v1.1.4
    commit:  99887766554433221100ffeeddccbbaa11223344
    sum:     h1:...
    vcs:     git
    repoURL: https://git.internal/ringil/stm32-hal-skc.git
    path:    third_party/cpkg/git.internal/ringil/stm32-hal-skc

  git.internal/ringil/stsafe-a110:
    version: v1.0.3
    commit:  aabbccddeeff00112233445566778899aabbccdd
    sum:     h1:...
    vcs:     git
    repoURL: https://git.internal/ringil/stsafe-a110.git
    path:    third_party/cpkg/git.internal/ringil/stsafe-a110
```

#### 2.2.1 Field semantics

* `apiVersion`: Schema version for the lockfile (`cpkg.ringil.dev/v0`).
* `kind`: Always `Lockfile` for v0.
* `module`: Must match `module` from `cpkg.yaml`.
* `generatedBy`: cpkg binary version.
* `generatedAt`: Timestamp of lock generation.
* `depRoot`: Derived from manifest; used for layout.
* `dependencies`:

  * Keys: module paths.
  * Values:

    * `version`: Concrete version tag (e.g., `v5.7.3`).
    * `commit`: Full commit SHA.
    * `sum`: Checksum of the tree/tag (Go-style `h1:` hash).
    * `vcs`: Version control system type; v0: always `git`.
    * `repoURL`: Fully resolved git URL.
    * `path`: Filesystem path where this dependency should live.

---

## 3. CLI Specification (v0)

Implementation: Go + `github.com/SCKelemen/clix` for ergonomic, pretty CLI output.

Binary name: `cpkg`

### 3.1 Global behavior

* Automatically finds the nearest `cpkg.yaml` by walking up the directory tree from the current working directory (Go-style module discovery).
* Exit status:

  * `0` on success.
  * Non-zero on error, with clear message.

#### 3.1.1 Global flags

* `--dep-root DIR`

  * Override `depRoot` from `cmod.yaml`/lockfile.
* `--verbose`, `-v`

  * More logging (debug info, git commands when useful).
* `--quiet`, `-q`

  * Minimal output.
* `--color {auto,always,never}`

  * Color handling; delegated to clix.

#### 3.1.2 Environment variables

* `CPKG_DEP_ROOT` — runtime override of `depRoot`.
* `CPKG_TARGET` — default target for `cpkg build`/`cpkg test` when `--target` not supplied.

---

### 3.2 Commands

#### 3.2.1 `cpkg init`

**Purpose:** Initialize a new module.

**Usage:**

```sh
cpkg init [--module <path>] [--dep-root <dir>]
```

**Behavior:**

* If `cpkg.yaml` exists → error (v0; later may support `--force`).

* Determine `module`:

  * If `--module` provided, use it.
  * Else, if in a git repo with `origin` remote, infer from remote URL (e.g., `github.com/user/repo`).
  * Else, error (module path required).

* Determine `depRoot`:

  * `--dep-root` flag, then `CPKG_DEP_ROOT`, else default `third_party/cpkg`.

* Write minimal `cpkg.yaml`:

  ```yaml
  apiVersion: cpkg.ringil.dev/v0
  kind: Module
  module: <resolved-module>
  depRoot: <resolved-depRoot>
  language:
    cStandard: c23
    skc: true
  dependencies: {}
  ```

* Pretty output (via clix) showing created file and inferred module.

---

#### 3.2.2 `cpkg add`

**Purpose:** Add or update a dependency constraint in `cpkg.yaml`.

**Usage:**

```sh
cpkg add <module[@version]>...
# examples:
cpkg add github.com/ringil/wolfssl-fork@^5.7.0
cpkg add git.internal/ringil/stm32-hal-skc@^1.1.0
```

**Behavior:**

* For each argument `<module[@version]>`:

  * Parse module path and version range.
  * If version omitted (v0): error or default to a policy (e.g., require explicit; v0 can be strict).
* Update `dependencies` in `cpkg.yaml`:

  * Add new entries or update existing version ranges.
* Do **NOT** modify `lock.cpkg.yaml`.
* Pretty-print a diff-style summary:

  * `+ github.com/ringil/wolfssl-fork @ ^5.7.0`
  * `~ github.com/ringil/stm32-hal-skc: ^1.0.0 → ^1.1.0`

---

#### 3.2.3 `cpkg tidy`

**Purpose:** Resolve dependency graph and write `lock.cpkg.yaml`.

**Usage:**

```sh
cpkg tidy [--dep-root DIR] [--check]
```

**Behavior:**

* Read `cpkg.yaml`.
* Resolve all `dependencies`:

  * v0: git-only resolution.
  * Map module path → repo URL:

    * Default: `https://<module>.git`.
    * Future: overrides via config.
  * `git ls-remote --tags` to list tags.
  * Pick highest compatible tag satisfying the semver constraint.
  * Get commit SHA for that tag.
  * Compute checksum (`sum`) of that tree/tag.
  * Compute `path` as `<depRoot>/<module>`.
* Construct in-memory lockfile object.
* If `--check`:

  * Compare to existing `lock.cpkg.yaml`.
  * Exit `0` if identical; non-zero if changes would be required.
  * Do not write the lockfile.
* Else:

  * Write `lock.cpkg.yaml` atomically.
* Pretty summary:

  * New deps, updated versions, removed deps.

---

#### 3.2.4 `cpkg sync`

**Purpose:** Make working tree (git submodules) match `lock.cpkg.yaml`.

**Usage:**

```sh
cpkg sync [--dep-root DIR]
```

**Behavior:**

* Ensure `lock.cpkg.yaml` exists; if not, run `cpkg tidy` first.
* For each dependency entry in lockfile:

  * Ensure `.gitmodules` has an entry for `path` with:

    ```ini
    [submodule "<path>"]
      path = <path>
      url  = <repoURL>
    ```

  * If not present, run `git submodule add <repoURL> <path>`.

  * If present but URL mismatched, run `git submodule set-url <path> <repoURL>`.

  * Run `git submodule init <path>`.

  * Run `git -C <path> fetch --tags`.

  * Run `git -C <path> checkout <commit>`.
* Optional (v0 may just warn): detect submodules under `depRoot` not in lockfile and warn or offer to remove.
* Do not commit; leave that to the user.
* Pretty output per dependency:

  * `✓ github.com/ringil/wolfssl-fork @ v5.7.3 (1234567)`

---

#### 3.2.5 `cpkg vendor`

**Purpose:** Copy resolved sources into a vendor directory for a fully self-contained tree (optional workflow).

**Usage:**

```sh
cpkg vendor [--vendor-root DIR]
```

**Behavior:**

* Resolve `vendorRoot`:

  * `--vendor-root`, env var (future), or default (`vendor` or `third_party/vendor`).
* Requires `lock.cpkg.yaml`.
* For each dependency:

  * Copy source tree from `path` (submodule path) into `vendorRoot/<module>`.
  * Alternatively: fetch directly from repo; v0 can prefer copying from submodule.
* Does not modify `.gitmodules`.
* Pretty summary:

  * `Vendored 4 dependencies into vendor/`.

---

#### 3.2.6 `cpkg status`

**Purpose:** Show dependency status at a glance.

**Usage:**

```sh
cpkg status
```

**Behavior:**

* Compare `cpkg.yaml` and `lock.cpkg.yaml`.
* For each dependency:

  * Show:

    * Module
    * Constraint (from manifest)
    * Locked version (from lockfile)
    * Local submodule state (missing / dirty / at locked commit)
    * Status: OK / NO_LOCK / OUT_OF_SYNC
* Pretty table output via clix.

Example:

```text
MODULE                                 CONSTRAINT   LOCKED   LOCAL     STATUS
github.com/ringil/wolfssl-fork         ^5.7.0       v5.7.3   v5.7.3    OK
git.internal/ringil/stm32-hal-skc      ^1.1.0       v1.1.4   v1.1.2    OUT_OF_SYNC
```

---

#### 3.2.7 `cpkg check`

**Purpose:** Check for newer versions of dependencies (without modifying files).

**Usage:**

```sh
cpkg check
```

**Behavior:**

* Read `cpkg.yaml` and `lock.cpkg.yaml`.
* For each dependency:

  * Fetch tags from repo.
  * Determine latest tag(s) that:

    * Satisfy the manifest constraint.
    * Or represent the next patch/minor/major.
* Print a table:

```text
MODULE                           CURRENT   LATEST   CONSTRAINT     NOTES
github.com/ringil/wolfssl-fork   v5.7.3    v5.7.5   ^5.7.0         patch available
```

* Future: `--json` flag for machine-readable output for a custom dependabot-like service.

---

#### 3.2.8 `cpkg build`

**Purpose:** Run the project's build command after ensuring dependencies are resolved and synced.

**Usage:**

```sh
cpkg build [--target <name>] [--dep-root DIR]
```

**Behavior:**

* Internally:

  1. `cpkg tidy` (resolve & update lockfile).
  2. `cpkg sync` (update submodules).
  3. Determine build command:

     * If `--target <name>` provided and `build.targets.<name>` exists in manifest, use that.
     * Else, use `build.command`.
     * If no command found → error.
  4. Exec the command as a child process.
* Environment variables for the build:

  * `CPKG_ROOT` → project root.
  * `CPKG_DEP_ROOT` → resolved depRoot.
  * `CPKG_TARGET` → target name (if provided).
* Pretty logs (steps, timings, exit status).

---

#### 3.2.9 `cpkg test`

**Purpose:** Run the project's test command after ensuring deps are resolved/synced and build is done.

**Usage:**

```sh
cpkg test [--target <name>] [--dep-root DIR]
```

**Behavior:**

* Equivalent to:

  1. `cpkg build [--target ...]`
  2. Run `test.command` from `cpkg.yaml`.
* Error if `test.command` is missing.
* Same env vars as `build`.

---

#### 3.2.10 `cpkg graph` (optional, nice-to-have)

**Purpose:** Display the dependency graph.

**Usage:**

```sh
cpkg graph
```

**Behavior:**

* Read `lock.cpkg.yaml`.
* Print a simple tree:

```text
github.com/ringil/device-fw
├─ github.com/ringil/wolfssl-fork v5.7.3
├─ github.com/ringil/mbedtls-fork v3.5.2
├─ git.internal/ringil/stm32-hal-skc v1.1.4
└─ git.internal/ringil/stsafe-a110 v1.0.3
```

* v0 assumes only direct dependencies; future versions can support transitive.

---

## 4. Implementation Notes (Non-normative)

* Implementation language: Go.
* CLI framework: `github.com/SCKelemen/clix`.
* Internal packages (suggested):

  * `internal/manifest` — parse & validate `cpkg.yaml`.
  * `internal/lockfile` — parse & write `lock.cpkg.yaml`.
  * `internal/semver` — semver range handling (or third-party lib).
  * `internal/git` — git operations (ls-remote, clone, fetch, checkout).
  * `internal/submodule` — .gitmodules management + submodule commands.
  * `internal/ui` — clix-based pretty printing.

v0 intentionally focuses on:

* Git-only modules.
* Direct dependencies.
* Submodule layout.

Future versions may add:

* S3/GCS-based sources with signed URLs.
* Transitive dependencies.
* Override files for repo URL remapping.
* JSON output for CI bots.
* Integration with a pkg.go.dev / Elm-like documentation index for C/SKC.
