# Codecov Setup Guide

This document provides step-by-step instructions for setting up code coverage reporting with Codecov.

## Overview

Codecov is a free code coverage service that:
- Collects coverage reports from your CI pipeline
- Displays coverage percentages and trends
- Provides PR comments showing coverage changes
- Allows setting coverage thresholds (e.g., fail if coverage drops > 5%)
- Works automatically with public GitHub repositories

## Step 1: Create a Codecov Account

1. Visit [codecov.io](https://codecov.io)
2. Click **Sign Up** in the top right
3. Select **GitHub** as your sign-up method
4. Authorize Codecov to access your GitHub account
   - Codecov requires access to public repositories (and private if you use Codecov pro)
   - You'll be redirected back to Codecov after authorization
5. Review the permissions and click **Authorize codecov**

## Step 2: Activate Repository

After signing up:

1. Log in to [codecov.io](https://codecov.io)
2. Go to **Dashboard** or click your profile → **Repositories**
3. Find **PresensiGo** in the list (may take a few minutes to appear)
4. Click on the repository to open its settings
5. The repository is now active and ready to receive coverage reports

## Step 3: Verify CI Integration

The CI workflow already includes coverage upload steps:

**Backend Coverage Upload:**
```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: ./backend/coverage.out
    flags: backend
    fail_ci_if_error: false
```

**Mobile Coverage Upload:**
```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: ./presensigo_mobile/coverage/lcov.info
    flags: mobile
    fail_ci_if_error: false
```

**How It Works:**
1. When the workflow runs, coverage reports are generated:
   - Backend: `backend/coverage.out` (Go coverage format)
   - Mobile: `presensigo_mobile/coverage/lcov.info` (LCOV format)
2. The `codecov/codecov-action@v3` step uploads these files to Codecov
3. Codecov parses the files and displays coverage metrics
4. `fail_ci_if_error: false` ensures upload failure doesn't block the PR

## Step 4: View Coverage Reports

After the first workflow run:

1. Log in to [codecov.io](https://codecov.io)
2. Click on **PresensiGo** repository
3. You'll see:
   - Overall coverage percentage
   - Coverage breakdown by backend/mobile
   - File-by-file coverage details
   - Coverage trend over time

On GitHub PRs:
1. Go to your pull request
2. Scroll to the bottom where checks are listed
3. Look for **codecov/presensigo** (or similar)
4. Click it to see:
   - Coverage change compared to base branch
   - Files with coverage changes
   - Line-by-line coverage details

## Step 5 (Optional): Set Coverage Thresholds

Coverage thresholds allow you to enforce minimum coverage standards.

### Method 1: codecov.yml Configuration

Create `codecov.yml` in the repository root:

```yaml
coverage:
  precision: 2
  round: down
  range: "70..100"

ignore:
  - "vendor"
  - "**/mocks"
  - "**/*_test.go"

flags:
  backend:
    paths:
      - backend/
    carryforward: true
  mobile:
    paths:
      - presensigo_mobile/
    carryforward: true

require:
  changes: false
  patch: false
  project:
    default:
      target: 70%
      threshold: 5%
    backend:
      target: 75%
      threshold: 5%
    mobile:
      target: 65%
      threshold: 5%
```

**Configuration Explained:**

- `coverage.range: "70..100"` — Acceptable coverage range (green: 70-100%, red: <70%)
- `ignore` — Paths to exclude from coverage calculations
- `flags` — Different coverage flags for backend and mobile
- `require.project` — Minimum coverage thresholds:
  - `target: 75%` — Fail if overall coverage drops below 75%
  - `threshold: 5%` — Fail if coverage drop is > 5% compared to base branch

### Method 2: Codecov Dashboard Settings

1. Log in to Codecov
2. Go to **PresensiGo** repository
3. Click **Settings** (gear icon)
4. Navigate to **Coverage** or **Requirements**
5. Set:
   - **Minimum coverage:** 70%
   - **Coverage threshold:** 5%
   - **Patch coverage:** 80%
6. Save settings

Codecov will now comment on PRs with:
- ✅ Coverage check passed
- ❌ Coverage check failed (with reason)

## Step 6: Add Coverage Badge to README

Add Codecov badge to `README.md`:

```markdown
## Status

[![CI Status](https://github.com/[OWNER]/PresensiGo/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/[OWNER]/PresensiGo/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/[OWNER]/PresensiGo/branch/main/graph/badge.svg?token=YOUR_TOKEN)](https://codecov.io/gh/[OWNER]/PresensiGo)
```

**Replace:**
- `[OWNER]` with your GitHub username or organization
- `YOUR_TOKEN` (optional) — Get from Codecov dashboard if you want a public badge for a private repo

The badge updates automatically whenever coverage changes.

## Troubleshooting

### Coverage Not Appearing on Codecov

**Symptoms:** Workflow runs but no coverage data on codecov.io

**Solutions:**
1. Check workflow logs for upload errors:
   - Go to GitHub Actions
   - Click the failed/recent workflow run
   - Look for "Upload coverage to Codecov" step
2. Verify coverage files are generated:
   ```bash
   cd backend && go test ./... -coverprofile=coverage.out
   ls -la coverage.out
   
   cd presensigo_mobile && flutter test --coverage
   ls -la coverage/lcov.info
   ```
3. Ensure `.gitignore` doesn't exclude coverage files (optional, not required in repo)
4. Check Codecov dashboard for errors:
   - Go to codecov.io/gh/[owner]/PresensiGo
   - Look for error messages or warnings

### "Coverage file not found" Error

**Cause:** Coverage file path is incorrect or file wasn't generated

**Fix:**
1. Verify file paths in workflow match actual output:
   - Backend: `./backend/coverage.out` (exact path)
   - Mobile: `./presensigo_mobile/coverage/lcov.info` (exact path)
2. Check step logs for coverage generation success
3. Run commands locally to verify output paths

### Codecov Action Rate Limiting

**Symptoms:** Upload fails with rate limit error

**Solution:**
- This is rare for public repos
- `fail_ci_if_error: false` ensures this doesn't block the PR
- Retry workflow run (GitHub Actions menu)

### Coverage Threshold Failing Unexpectedly

**Symptoms:** Coverage check fails when you expect it to pass

**Cause:** Codecov may be comparing against different base branch

**Solutions:**
1. Check codecov.yml or dashboard settings
2. Ensure branch protection rule uses correct base branch
3. Check PR against correct base branch (main vs develop)

## Reference

- [Codecov Documentation](https://docs.codecov.io)
- [codecov-action GitHub](https://github.com/codecov/codecov-action)
- [codecov.yml Schema](https://docs.codecov.io/docs/codecovyml-reference)

## Next Steps

1. ✅ Create Codecov account and activate repository
2. ✅ Verify first workflow run uploads coverage
3. ✅ Check Codecov dashboard for coverage report
4. ✅ (Optional) Configure codecov.yml with coverage thresholds
5. ✅ Add coverage badge to README
6. ✅ Share Codecov dashboard link with team
