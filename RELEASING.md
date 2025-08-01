# Release Process

This document describes the release process for the tokenizer project.

## Prerequisites

Before creating a release, ensure you have:

1. **GoReleaser** installed: `go install github.com/goreleaser/goreleaser@latest`
2. **GitHub CLI** installed: `brew install gh` (or see [installation docs](https://cli.github.com/))
3. Write access to the repository
4. A GitHub personal access token with `repo` scope

## Release Steps

### 1. Prepare the Release

1. Ensure all changes are committed and pushed to the main branch
2. Run tests to ensure everything is working:
   ```bash
   make test-all
   ```
3. Update documentation if needed:
   ```bash
   make generate
   ```

### 2. Create a Release Tag

Create and push a new version tag:

```bash
# For a new patch release (e.g., v1.0.0 -> v1.0.1)
make tag VERSION=v1.0.1

# For a new minor release (e.g., v1.0.0 -> v1.1.0)
make tag VERSION=v1.1.0

# For a new major release (e.g., v1.0.0 -> v2.0.0)
make tag VERSION=v2.0.0
```

Push the tag to GitHub:
```bash
git push origin v1.0.1  # Replace with your version
```

### 3. Create the Release

The GitHub Actions workflow will automatically:
1. Run all tests
2. Build binaries for multiple platforms
3. Create a GitHub release with:
   - Automated changelog based on commit messages
   - Binary downloads for all platforms
   - Checksums file

### 4. Verify the Release

1. Check the [Releases page](https://github.com/agentstation/tokenizer/releases) to ensure the release was created
2. Download and test a binary to ensure it works correctly:
   ```bash
   # Download the binary for your platform
   curl -L https://github.com/agentstation/tokenizer/releases/download/v1.0.1/tokenizer_1.0.1_darwin_arm64.tar.gz | tar xz
   
   # Test it
   ./tokenizer version
   ```

### 5. Post-Release

1. Update any documentation that references the version
2. Announce the release (if applicable)
3. Update any dependent projects

## Manual Release (if needed)

If the automated release fails, you can create a release manually:

```bash
# Test the release process locally
make release-snapshot

# Create the actual release
GITHUB_TOKEN=your_token_here make release
```

## Versioning Guidelines

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0) - Incompatible API changes
- **MINOR** version (0.X.0) - New functionality, backwards compatible
- **PATCH** version (0.0.X) - Bug fixes, backwards compatible

## Commit Message Format

For better changelogs, use conventional commit messages:

- `feat:` - New features (triggers MINOR version bump)
- `fix:` - Bug fixes (triggers PATCH version bump)
- `perf:` - Performance improvements
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks
- `test:` - Test additions/changes

Breaking changes should include `!` after the type or `BREAKING CHANGE:` in the footer.

Example:
```
feat!: remove deprecated API endpoints

BREAKING CHANGE: The /api/v1/encode endpoint has been removed.
Use /api/v2/encode instead.
```

## Troubleshooting

### Release Workflow Fails

1. Check the [Actions tab](https://github.com/agentstation/tokenizer/actions) for error details
2. Ensure your tag follows the `v*` pattern (e.g., `v1.0.0`)
3. Verify GitHub token permissions

### GoReleaser Errors

1. Run `goreleaser check` to validate configuration
2. Test locally with `make release-snapshot`
3. Check GoReleaser logs for specific errors

### Binary Not Working

1. Ensure CGO is disabled in the build (check `.goreleaser.yaml`)
2. Test on the target platform
3. Check for missing dependencies

## Rollback Procedure

If a release has critical issues:

1. Delete the release from GitHub (keep the tag)
2. Fix the issues
3. Create a new patch release with the fixes

Never delete or modify existing tags that have been published.