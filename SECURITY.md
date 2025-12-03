# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

Please report (suspected) security vulnerabilities to **[security@kelemen.com](mailto:security@kelemen.com)**. You will receive a response within 72 hours. If the issue is confirmed, we will release a patch as soon as possible depending on complexity but historically within a few days.

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)

We will strive to reward and acknowledge security contributions.

## Security Considerations

cpkg manages git submodules and interacts with external repositories. Security considerations:

- **Git operations**: cpkg uses `git` commands to clone and manage repositories. Ensure you trust the repositories you're adding as dependencies.
- **Lockfile integrity**: The lockfile includes checksums to verify dependency integrity.
- **Submodule security**: Git submodules are managed through standard git commands, inheriting git's security model.

## Best Practices

- Review dependencies before adding them
- Keep dependencies up to date: `cpkg outdated` and `cpkg upgrade`
- Verify lockfile checksums match expected values
- Use version constraints to limit dependency updates

