# Homebrew Tap Setup Guide

This document outlines the setup required to enable Homebrew tap functionality for mcp-server-dump.

## Prerequisites

### 1. Create Homebrew Tap Repository

Create a new GitHub repository named `homebrew-tap` under the `spandigital` organization:

```bash
# Repository: spandigital/homebrew-tap
# Description: Homebrew tap for SPAN Digital tools
# Public repository
```

### 2. GitHub Personal Access Token

GoReleaser requires a Personal Access Token (PAT) to write to the homebrew-tap repository.

#### Create PAT:
1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Set expiration as needed
4. Required scopes:
   - `public_repo` (for public repositories)
   - `write:packages` (if publishing packages)

#### Add to Repository Secrets:
1. Go to `spandigital/mcp-server-dump` → Settings → Secrets and variables → Actions
2. Add new repository secret:
   - Name: `TAP_GITHUB_TOKEN`
   - Value: Your generated PAT

## How It Works

1. **Release Trigger**: When a new tag (e.g., `v1.15.0`) is pushed to the main repository
2. **GoReleaser Action**: The GitHub Actions workflow runs GoReleaser
3. **Formula Generation**: GoReleaser generates a Homebrew formula based on the release assets
4. **Tap Update**: Using `TAP_GITHUB_TOKEN`, GoReleaser pushes the formula to `spandigital/homebrew-tap`
5. **User Installation**: Users can then install via `brew install spandigital/homebrew-tap/mcp-server-dump`

## Repository Structure

After the first release, the `homebrew-tap` repository will contain:

```
spandigital/homebrew-tap/
├── Formula/
│   └── mcp-server-dump.rb
└── README.md
```

## Testing

After setup, test with a new release:

1. Create and push a new tag: `git tag v1.15.0 && git push origin v1.15.0`
2. Check GitHub Actions for successful release
3. Verify formula appears in `spandigital/homebrew-tap`
4. Test installation: `brew install spandigital/homebrew-tap/mcp-server-dump`

## Troubleshooting

### Common Issues:

1. **"Resource not accessible by integration"**
   - Ensure `TAP_GITHUB_TOKEN` is set in repository secrets
   - Verify token has `public_repo` scope
   - Check that `spandigital/homebrew-tap` repository exists

2. **Formula generation fails**
   - Verify GoReleaser configuration syntax
   - Check that release assets are properly generated
   - Ensure binary names match configuration

3. **Permission denied**
   - Verify PAT hasn't expired
   - Check token has write access to `homebrew-tap` repository

### Debug Steps:

1. Check GitHub Actions logs for detailed error messages
2. Verify GoReleaser configuration: `goreleaser check`
3. Test locally: `goreleaser release --snapshot --clean`