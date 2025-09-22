# GitHub CI/CD Setup Guide

This guide explains how to set up your repository for automated releases with GitHub Actions.

## 🚀 Quick Setup

1. **Create a GitHub repository** and push your code:
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git branch -M main
   git remote add origin https://github.com/yourusername/labelstudio-to-yolo.git
   git push -u origin main
   ```

2. **The GitHub Actions workflows are already configured!** They will automatically:
   - Run tests on every push/PR
   - Build releases when you create tags

## 📦 Creating Releases

### Automatic Releases (Recommended)

1. **Create and push a tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Run the full test suite
   - Build binaries for all platforms:
     - Linux (x64, ARM64)
     - Windows (x64)
     - macOS (Intel, Apple Silicon)
   - Create a GitHub release with:
     - Binaries for all platforms
     - SHA256 checksums
     - Auto-generated release notes

### Manual Releases

1. Go to your GitHub repository
2. Click "Releases" → "Create a new release"
3. Choose a tag (or create a new one like `v1.0.0`)
4. GitHub Actions will build and attach binaries automatically

## 🔧 Workflow Details

### CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `master` branch
- Pull requests to `main` or `master`

**What it does:**
- ✅ Runs tests on Go 1.21+
- ✅ Checks code formatting
- ✅ Runs `go vet` for static analysis
- ✅ Builds for multiple platforms
- ✅ Uploads build artifacts
- ✅ Generates test coverage reports

### Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Push tags starting with `v*` (e.g., `v1.0.0`, `v2.1.3`)
- Manual release creation

**What it builds:**
- `labelstudio-to-yolo_linux_amd64` - Linux x64
- `labelstudio-to-yolo_linux_arm64` - Linux ARM64
- `labelstudio-to-yolo_windows_amd64.exe` - Windows x64
- `labelstudio-to-yolo_darwin_amd64` - macOS Intel
- `labelstudio-to-yolo_darwin_arm64` - macOS Apple Silicon

**Each binary includes:**
- Version information
- Build timestamp
- Git commit hash
- SHA256 checksum file

## 🏷️ Version Tagging Strategy

**Recommended format:** `v<major>.<minor>.<patch>`

Examples:
- `v1.0.0` - Initial release
- `v1.0.1` - Bug fix
- `v1.1.0` - New features
- `v2.0.0` - Breaking changes

```bash
# Patch release (bug fixes)
git tag v1.0.1
git push origin v1.0.1

# Minor release (new features)
git tag v1.1.0
git push origin v1.1.0

# Major release (breaking changes)
git tag v2.0.0
git push origin v2.0.0
```

## 🔒 Permissions

The workflows use the built-in `GITHUB_TOKEN` which automatically has the necessary permissions to:
- Read repository contents
- Create releases
- Upload assets

**No additional setup required!**

## 📊 Monitoring Builds

1. **View workflow runs:** Go to your repository → "Actions" tab
2. **Check build status:** Look for green checkmarks or red X's
3. **Debug failures:** Click on failed workflows to see logs
4. **Download artifacts:** Available for 7 days from CI builds

## 🛠️ Customization

### Update Repository URLs

In the following files, replace `yourusername/labelstudio-to-yolo` with your actual repository:

- `README.md` - Badge URLs and links
- `CHANGELOG.md` - Release comparison links

### Modify Build Targets

To add or remove build targets, edit `.github/workflows/release.yml`:

```yaml
strategy:
  matrix:
    include:
      # Add new platforms here
      - goos: freebsd
        goarch: amd64
        suffix: _freebsd_amd64
```

### Change Trigger Conditions

To build on different events, modify the `on:` section in the workflow files:

```yaml
on:
  push:
    branches: [ main, develop ]  # Add more branches
    tags: [ 'v*', 'release-*' ]  # Different tag patterns
```

## 🧪 Testing the Setup

1. **Test CI:** Push any commit to main
   ```bash
   git commit --allow-empty -m "Test CI"
   git push origin main
   ```

2. **Test Release:** Create a test tag
   ```bash
   git tag v0.1.0-test
   git push origin v0.1.0-test
   ```

3. **Verify:** Check the "Actions" and "Releases" tabs on GitHub

## 🚨 Troubleshooting

### Build Failures

**Common issues:**
- **Test failures:** Fix failing tests before tagging
- **Go version:** Ensure Go 1.21+ compatibility
- **Import paths:** Use `go mod tidy` to fix dependencies

### Permission Errors

**If uploads fail:**
- Check repository settings → Actions → General
- Ensure "Read and write permissions" are enabled for `GITHUB_TOKEN`

### Missing Releases

**If releases aren't created:**
- Verify tag format starts with `v` (e.g., `v1.0.0`)
- Check workflow logs in the Actions tab
- Ensure the tag pushed to the main branch

## 📈 Success Metrics

After setup, you'll have:
- ✅ **Automated testing** on every commit
- ✅ **Cross-platform builds** for 5 architectures
- ✅ **Secure releases** with checksums
- ✅ **Professional presentation** with badges and documentation
- ✅ **Zero-maintenance** releases once configured

---

Your repository is now ready for professional-grade CI/CD! 🎉
