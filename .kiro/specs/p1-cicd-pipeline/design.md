# Design Document

## Overview

This design establishes a GitHub Actions CI/CD pipeline that runs automated checks on both backend (Go) and mobile (Flutter) codebases. The pipeline is triggered on every push and pull request, runs in parallel where possible, and reports results back to the PR.

Key features:
- **Multi-job workflow:** Backend and mobile jobs run in parallel
- **Consistent checks:** Same checks run on all branches and PRs
- **Fast feedback:** Developers see results within minutes
- **Coverage tracking:** Code coverage metrics collected and reported
- **Required checks:** PR cannot be merged without passing all checks

## Architecture

```
GitHub Push/PR Event
    ↓
.github/workflows/ci.yml triggered
    ↓
    ├─ Backend Job (Linux runner)
    │   ├─ Checkout code
    │   ├─ Set up Go
    │   ├─ Run go test ./...
    │   ├─ Run go vet ./...
    │   ├─ Run go fmt -l
    │   ├─ Run golangci-lint
    │   ├─ Generate coverage report
    │   └─ Upload coverage to Codecov
    │
    └─ Mobile Job (Linux runner with Flutter)
        ├─ Checkout code
        ├─ Set up Flutter
        ├─ Run flutter analyze
        ├─ Run flutter test
        ├─ Generate coverage report
        └─ Upload coverage to Codecov
    ↓
Report Results
    ├─ Show status in PR checks
    ├─ Update status badge
    └─ Fail PR if any check fails
```

## Components and Interfaces

### 1. GitHub Actions Workflow File

**File:** `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main, develop, 'fix/**', 'feat/**']
  pull_request:
    branches: [main, develop]

jobs:
  backend:
    name: Backend (Go)
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.21
          cache: true
      
      - name: Run tests
        run: cd backend && go test ./... -v
      
      - name: Run vet
        run: cd backend && go vet ./...
      
      - name: Check formatting
        run: cd backend && go fmt -l . && [ -z "$(go fmt -l .)" ] || exit 1
      
      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          working-directory: backend
          version: latest
      
      - name: Generate coverage
        run: cd backend && go test ./... -coverprofile=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
          flags: backend
          fail_ci_if_error: false

  mobile:
    name: Mobile (Flutter)
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Flutter
        uses: subosito/flutter-action@v2
        with:
          flutter-version: 3.13.x
          cache: true
      
      - name: Get dependencies
        run: cd presensigo_mobile && flutter pub get
      
      - name: Run analyze
        run: cd presensigo_mobile && flutter analyze --no-fatal-infos
      
      - name: Run tests
        run: cd presensigo_mobile && flutter test --coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./presensigo_mobile/coverage/lcov.info
          flags: mobile
          fail_ci_if_error: false

  status:
    name: Status Check
    runs-on: ubuntu-latest
    needs: [backend, mobile]
    if: always()
    
    steps:
      - name: Check status
        run: |
          if [ "${{ needs.backend.result }}" != "success" ] || [ "${{ needs.mobile.result }}" != "success" ]; then
            exit 1
          fi
```

**Key Configuration:**
- Triggers on push to main/develop/feature branches and on PRs
- Uses ubuntu-latest runner (free tier on public repos)
- Backend job: Go 1.21, runs in `backend/` directory
- Mobile job: Flutter 3.13.x, runs in `presensigo_mobile/` directory
- Coverage uploaded to Codecov (free for public repos)
- Status job ensures both backend and mobile pass before PR approval

### 2. Go Checks Configuration

**Backend Test Discovery:**
- Go automatically discovers `*_test.go` files
- `go test ./...` runs all tests in all packages
- Coverage generated with `-coverprofile=coverage.out`

**Linter Configuration:**
File: `.golangci.yml` (optional, defines custom linter rules)

```yaml
linters:
  enable:
    - vet
    - fmt
    - goimports
    - govet
    - errcheck
    - staticcheck

issues:
  exclude-rules:
    # Exclude test files from some checks
    - path: _test\.go$
      linters:
        - errcheck
```

### 3. Flutter Checks Configuration

**Mobile Test Discovery:**
- Flutter automatically discovers tests in `test/` directory
- Tests must be `*_test.dart` files

**Analyze Configuration:**
File: `presensigo_mobile/analysis_options.yaml` (optional, defines linter rules)

```yaml
linter:
  rules:
    - avoid_empty_else
    - avoid_print
    - avoid_relative_lib_imports
    - avoid_returning_null_for_future
    - avoid_slow_async_io
    - cancel_subscriptions
    - close_sinks
    - comment_references
    - control_flow_in_finally
    - empty_statements
    - hash_and_equals
    - invariant_booleans
    - iterable_contains_unrelated_type
    - list_remove_unrelated_type
    - literal_only_boolean_expressions
    - no_adjacent_strings_in_list
    - no_duplicate_case_values
    - prefer_void_to_null
    - throw_in_finally
    - unnecessary_statements
    - unrelated_type_equality_checks
```

### 4. Code Coverage Reporting

**Backend Coverage:**
```bash
cd backend
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html  # Optional: local view
```

**Mobile Coverage:**
```bash
cd presensigo_mobile
flutter test --coverage
# Coverage output: coverage/lcov.info
```

**Codecov Integration:**
- Codecov automatically parses coverage files
- Displays coverage report in PR with pass/fail based on coverage thresholds
- Links to coverage dashboard: codecov.io/gh/[owner]/[repo]

### 5. Status Badge

**Add to README.md:**

```markdown
## Status

[![CI Status](https://github.com/[owner]/PresensiGo/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/[owner]/PresensiGo/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/[owner]/PresensiGo/branch/main/graph/badge.svg)](https://codecov.io/gh/[owner]/PresensiGo)
```

The badge updates automatically whenever the workflow runs. Clicking it links to the latest workflow run.

### 6. Branch Protection Rules

**Configuration in GitHub Repository Settings:**

1. Go to **Settings → Branches → Branch protection rules**
2. Create rule for `main` branch:
   - Require status checks: `Backend (Go)`, `Mobile (Flutter)`
   - Dismiss stale reviews when new commits pushed
   - Require branches to be up to date before merge
   - Include administrators in restrictions

This prevents merging until CI passes.

### 7. Documentation

**File:** `CONTRIBUTING.md` or README section

```markdown
## Continuous Integration

### Automated Checks

All pull requests are automatically tested by GitHub Actions. The workflow runs:

#### Backend Checks
- `go test ./...` — Runs all unit tests
- `go vet ./...` — Static analysis
- `go fmt -l` — Code formatting check
- `golangci-lint` — Additional linting

#### Mobile Checks
- `flutter analyze` — Code analysis
- `flutter test` — Widget and unit tests

### Fixing CI Failures

**Test Failures:**
```bash
cd backend && go test ./... -v  # Run locally to debug
```

**Format Issues:**
```bash
cd backend && go fmt ./...  # Auto-format all files
```

**Lint Issues:**
Review the error message and fix the code. Common issues:
- Unused imports: remove or use `_` blank import
- Unused variables: remove or prefix with `_`
- Missing error checks: add `if err != nil` handling

**Dart Analysis Issues:**
```bash
cd presensigo_mobile && flutter analyze  # Run locally
# Fix issues reported or update analysis_options.yaml if needed
```

### Running Checks Locally

Before pushing, run checks locally to avoid PR rejection:

```bash
# Backend
cd backend && go test ./... && go vet ./... && go fmt ./... && golangci-lint run

# Mobile
cd presensigo_mobile && flutter analyze && flutter test
```

### Viewing Results

1. Go to **GitHub Actions** tab in the repository
2. Click the workflow run name
3. Click individual job to see detailed output
4. Scroll down to see which checks failed and error details

Alternatively, PR status appears at the bottom of the PR page with a link to the workflow run.
```

## Data Models

### Workflow Execution Summary

```
Workflow Run
├─ Backend Job
│  ├─ Checkout
│  ├─ Setup Go
│  ├─ go test ./...          → Pass/Fail
│  ├─ go vet ./...           → Pass/Fail
│  ├─ go fmt -l              → Pass/Fail
│  ├─ golangci-lint          → Pass/Fail
│  └─ Coverage Report
│
├─ Mobile Job
│  ├─ Checkout
│  ├─ Setup Flutter
│  ├─ flutter pub get
│  ├─ flutter analyze        → Pass/Fail
│  ├─ flutter test           → Pass/Fail
│  └─ Coverage Report
│
└─ Status Job (depends on both)
   └─ Overall: Pass/Fail
```

### Error Classification

| Check | Failure Cause | Fix |
|---|---|---|
| `go test` | Test assertion fails | Fix the code or test |
| `go vet` | Possible bug detected | Review vet output and fix |
| `go fmt` | Code not formatted | Run `go fmt ./...` |
| `golangci-lint` | Code quality issue | Fix lint error or update config |
| `flutter analyze` | Dart analysis issue | Fix code or update analysis_options.yaml |
| `flutter test` | Test fails or exception | Fix code or test |

## Correctness Properties

**Property 1: Workflow Runs on Every Push and PR**
- **Description:** CI workflow is triggered automatically for every push and pull request
- **Test:** Push commit to branch → workflow runs; open PR → workflow runs

**Property 2: All Checks Must Pass Before PR Merge**
- **Description:** If any check fails, PR shows red status and merge is blocked
- **Test:** Intentionally fail a test, push → PR merge button disabled until test passes

**Property 3: Status Badge Reflects Current Status**
- **Description:** Badge shows green (passing) or red (failing) based on latest workflow run
- **Test:** All checks pass → badge green; introduce failing test → badge red

**Property 4: Coverage Report Generated and Accessible**
- **Description:** Coverage metrics are calculated and viewable in Codecov
- **Test:** Run workflow → codecov.io shows coverage percentage and trend

**Property 5: Developers Can Fix Issues Locally Using Same Commands**
- **Description:** Running `go test ./...` locally produces same results as CI
- **Test:** Run go test locally, get same pass/fail as CI

## Error Handling

### Transient Failures (Network, Timeout)

If a step fails due to network timeout or temporary unavailability:
- Workflow automatically retries once (default GitHub Actions behavior)
- If retry succeeds, PR status updates to green
- If retry fails, PR status remains red; developer must investigate

### Missing Dependencies

If `go mod download` or `flutter pub get` fails:
- Usually indicates a network issue or malformed `go.mod`/`pubspec.yaml`
- Run locally to debug: `go mod tidy` or `flutter pub get`
- Check dependency versions and fix

### Environment Mismatch

If tests pass locally but fail in CI (or vice versa):
- CI uses a clean Ubuntu environment; local machine may have extra tools installed
- Try running tests in Docker: `docker run -it ubuntu:22.04 bash` and install Go/Flutter
- Or use GitHub's `runner` labels to select a specific environment

### Secrets and Credentials

- Never commit API keys, tokens, or passwords to the repository
- If CI needs credentials (e.g., for code coverage upload), add them as GitHub Secrets
- Reference in workflow as `${{ secrets.SECRET_NAME }}`
