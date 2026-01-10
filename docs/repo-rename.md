# GitHub Repository Rename Guide

This document provides instructions for renaming the GitHub repository from `VideoTranscript.app` to `omnitranscripts`.

## When to Rename

The codebase has been updated to use `omnitranscripts` as the module name and all internal references have been updated. The repository rename is optional but recommended to align the GitHub URL with the project name.

## Pre-Rename Checklist

Before renaming the repository:

- [ ] Ensure all code changes from the rename are merged to main
- [ ] Notify any active contributors or CI/CD systems
- [ ] Update any external documentation that references the repo URL
- [ ] Note any GitHub Actions workflows that reference the repo name

## Rename Steps

### 1. Rename the Repository

1. Go to your repository on GitHub
2. Navigate to **Settings** > **General**
3. Under "Repository name", change from `VideoTranscript.app` to `omnitranscripts`
4. Click **Rename**

GitHub will automatically set up redirects from the old URL to the new URL.

### 2. Update Local Clone

After renaming, update your local remote URL:

```bash
# Check current remote
git remote -v

# Update remote URL
git remote set-url origin https://github.com/wilmoore/omnitranscripts.git

# Verify the change
git remote -v
```

### 3. Update Go Module Path (Optional)

If you want to use the canonical Go module path:

1. Update `go.mod`:
```diff
- module omnitranscripts
+ module github.com/wilmoore/omnitranscripts
```

2. Update all import paths in the codebase to use `github.com/wilmoore/omnitranscripts/...`

This step is optional for local development but recommended if you plan to publish the module for public consumption.

### 4. Update External References

After renaming, update any references in:

- [ ] README badge URLs (already updated in code)
- [ ] CI/CD configuration files
- [ ] External documentation or wikis
- [ ] Package manager registries (if published)
- [ ] Any webhook URLs that include the repo name

## Redirect Behavior

GitHub automatically redirects old URLs to new URLs:

- `github.com/wilmoore/VideoTranscript.app` → `github.com/wilmoore/omnitranscripts`
- All issues, PRs, and commits maintain their links
- Forks will continue to work but show the new upstream name

## Rollback

If you need to rollback, simply rename the repository back to `VideoTranscript.app` using the same steps. GitHub will re-establish the redirect in the opposite direction.

## Notes

- GitHub redirects are permanent until another repo claims the old name
- Contributors with local clones should update their remote URLs
- The redirect only works for web URLs; `go get` with the old path may not work after the redirect expires (usually 100 days)
