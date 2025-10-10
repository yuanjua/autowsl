# WinGet Publishing Setup

This document explains how to set up automated publishing to the Windows Package Manager (WinGet) repository.

## Overview

The project uses GitHub Actions to automatically publish releases to WinGet when a new release is created. The workflow uses the [vedantmgoyal9/winget-releaser](https://github.com/vedantmgoyal9/winget-releaser) action, which automates the process of creating and submitting pull requests to the [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) repository.

## Prerequisites

To enable WinGet publishing, you need to set up the following GitHub secrets:

### 1. WINGET_TOKEN

This is a Personal Access Token (PAT) with permissions to create pull requests in the winget-pkgs repository.

**Steps to create:**

1. Go to https://github.com/settings/tokens/new
2. Give it a descriptive name (e.g., "AutoWSL WinGet Publisher")
3. Set expiration as needed (recommend: 1 year)
4. Select the following scopes:
   - `public_repo` (for accessing public repositories)
5. Click "Generate token"
6. Copy the token immediately (you won't be able to see it again)

**Add to repository:**

1. Go to your repository's Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `WINGET_TOKEN`
4. Value: Paste the PAT you created
5. Click "Add secret"

### 2. WINGET_FORK_USER (Optional but Recommended)

This should be set to your GitHub username. The winget-releaser action will fork the winget-pkgs repository under this account if it doesn't already exist.

**Add to repository:**

1. Go to your repository's Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `WINGET_FORK_USER`
4. Value: Your GitHub username (e.g., `yuanjua`)
5. Click "Add secret"

If you don't set this, the action will use the repository owner's username by default.

## How It Works

### Trigger

The workflow is triggered in two ways:

1. **Automatically on Release**: When you publish a new release on GitHub (not a draft or pre-release)
2. **Manually**: Via the "Actions" tab in GitHub, you can trigger the workflow and specify a version

### What the Workflow Does

1. Detects all `.exe` installers from the release assets (both amd64 and arm64)
2. Automatically determines architecture based on filename
3. Creates or updates WinGet package manifests
4. Submits a pull request to microsoft/winget-pkgs
5. The PR will be reviewed by Microsoft's automated validation and human reviewers

### File Naming Convention

The workflow expects release assets to follow this naming pattern:
- `autowsl-windows-amd64.exe` - 64-bit x86 installer
- `autowsl-windows-arm64.exe` - ARM64 installer

This matches the current release workflow output.

## Package Identifier

The WinGet package identifier is: `yuanjua.autowsl`

This follows the WinGet naming convention of `Publisher.PackageName`.

## Testing

Before your first release, you can test the workflow:

1. Create a draft release with properly named executables
2. Go to Actions → Publish to WinGet → Run workflow
3. Enter the version tag (e.g., `v1.0.0`)
4. Check the workflow logs for any errors

## First-Time Submission

For your first submission to WinGet:

1. The PR will require manual review by Microsoft maintainers
2. They may request changes to the manifest
3. Future updates will be faster as the package is already in the registry
4. You may need to manually verify your identity as the package publisher

## Troubleshooting

### "Failed to create PR"
- Verify WINGET_TOKEN has the correct permissions
- Check if a fork already exists under WINGET_FORK_USER
- Ensure the release has properly named executable files

### "Validation failed"
- Check that the release tag follows semantic versioning (e.g., `v1.0.0`)
- Ensure executable files are actually Windows PE executables
- Review the action logs for specific validation errors

### "PR already exists"
- If a PR already exists for this version, the workflow will update it
- You may need to close the existing PR first if you want to recreate it

## Manual Override

If you need to manually publish to WinGet:

1. Fork https://github.com/microsoft/winget-pkgs
2. Create manifests in `manifests/y/yuanjua/autowsl/<version>/`
3. Follow the [WinGet manifest creation guide](https://github.com/microsoft/winget-pkgs/blob/master/AUTHORING_MANIFESTS.md)
4. Submit a PR to microsoft/winget-pkgs

## References

- [WinGet Documentation](https://docs.microsoft.com/en-us/windows/package-manager/)
- [winget-releaser Action](https://github.com/vedantmgoyal9/winget-releaser)
- [WinGet Package Repository](https://github.com/microsoft/winget-pkgs)
- [Manifest Authoring Guide](https://github.com/microsoft/winget-pkgs/blob/master/AUTHORING_MANIFESTS.md)
