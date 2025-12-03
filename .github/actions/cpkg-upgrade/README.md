# cpkg Upgrade Action

A reusable GitHub Action for automatically upgrading cpkg dependencies and creating pull requests.

## Usage

### Using from this repository

Add this to your `.github/workflows/dependencies.yml`:

```yaml
name: Update Dependencies

on:
  schedule:
    # Run weekly on Monday at 00:00 UTC
    - cron: '0 0 * * 1'
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true

      - name: Update dependencies
        uses: github.com/SCKelemen/cpkg/.github/actions/cpkg-upgrade@main
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
```

### Using from a dedicated action repository

If this action is published in a separate repository (e.g., `github.com/SCKelemen/cpkg-upgrade-action`), you can reference it directly:

```yaml
- name: Update dependencies
  uses: github.com/SCKelemen/cpkg-upgrade-action@v1
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `token` | GitHub token for creating pull requests | Yes | - |
| `pr-title` | Title for the pull request | No | `chore: update dependencies` |
| `pr-body` | Body for the pull request | No | `Automated dependency update.\n\nThis PR was created by the cpkg upgrade workflow.` |
| `branch` | Branch name for the pull request | No | `cpkg-upgrade-dependencies` |
| `commit-message` | Commit message for the changes | No | `chore: update dependencies` |
| `go-version` | Go version to use | No | `1.21` |

## Example

```yaml
- name: Update dependencies
  uses: github.com/SCKelemen/cpkg/.github/actions/cpkg-upgrade@main
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
    pr-title: "chore(deps): upgrade dependencies"
    pr-body: |
      Automated dependency updates via cpkg.
      
      Run `cpkg outdated` to see what changed.
    branch: update-deps
    commit-message: "chore: update dependencies"
    go-version: "1.21"
```

## How It Works

1. Installs cpkg using `go install`
2. Checks for outdated dependencies using `cpkg outdated`
3. If updates are available, runs `cpkg upgrade` to upgrade all dependencies
4. Creates a pull request with the changes

The action will only create a PR if there are actual changes to commit.

