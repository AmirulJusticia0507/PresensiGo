# Implementation Plan

## Overview

This milestone establishes continuous integration and deployment infrastructure for the PresensiGo project. It automates testing and quality checks for both backend and mobile components, integrates code coverage reporting, and documents the CI/CD process for the team. The workflow ensures code quality gates are met before merge and provides visibility through status badges and comprehensive CI logs.

## Task Dependency Graph

Task 1 creates the workflow foundation. Tasks 2 and 3 are parallel (backend and mobile checks). Tasks 4-7 are sequential (coverage setup, documentation, testing, badges).

```json
{
  "waves": [
    { "wave": 1, "tasks": ["Task 1"] },
    { "wave": 2, "tasks": ["Task 2", "Task 3"] },
    { "wave": 3, "tasks": ["Task 4"] },
    { "wave": 4, "tasks": ["Task 5"] },
    { "wave": 5, "tasks": ["Task 6"] },
    { "wave": 6, "tasks": ["Task 7"] }
  ]
}
```

## Tasks

- [ ] 1. Create GitHub Actions workflow file
  - Create `.github/workflows/ci.yml`
  - Configure `on: [push, pull_request]` triggers
  - Define matrix for backend and mobile jobs running in parallel
  - Ensure workflow syntax is valid (can test with GitHub Actions validator)
  - Workflow should not require any secrets initially (build is public)

- [ ] 2. Set up backend checks
  - Backend job: check out code and set up Go 1.21
  - Run `go test ./...` with verbose flag
  - Run `go vet ./...` and fail on issues
  - Run `go fmt -l` to check formatting (fail if files need formatting)
  - Run `golangci-lint` using golangci/golangci-lint-action
  - Each step should have clear name describing what it does
  - Workflow should fail if any check fails

- [ ] 3. Set up mobile checks
  - Mobile job: check out code and set up Flutter 3.13.x
  - Run `flutter pub get` to fetch dependencies
  - Run `flutter analyze --no-fatal-infos` and fail on high-severity issues
  - Run `flutter test --coverage` to run all widget/unit tests
  - Each step should have clear name describing what it does
  - Workflow should fail if any check fails

- [ ] 4. Configure code coverage
  - Backend: Generate coverage with `go test ./... -coverprofile=coverage.out`
  - Mobile: Ensure `flutter test --coverage` generates `coverage/lcov.info`
  - Add codecov/codecov-action@v3 step to upload both coverage reports
  - Set `fail_ci_if_error: false` so coverage upload failure doesn't block merge
  - Create Codecov account (free for public repos) and verify integration works
  - Optional: Set coverage thresholds in Codecov dashboard (e.g., fail if coverage drops > 5%)

- [ ] 5. Add status badge to README
  - Add GitHub Actions status badge to README: `[![CI Status](https://github.com/.../badge.svg)](https://...)`
  - Add Codecov coverage badge: `[![codecov](https://codecov.io/gh/.../badge.svg)](https://codecov.io/gh/...)`
  - Badges should link to workflow runs and codecov dashboard respectively
  - Badges should automatically update when workflow runs

- [ ] 6. Document CI process
  - Create or update `CONTRIBUTING.md` file
  - Document when checks run (push to any branch, PRs to main/develop)
  - List each check and what it verifies (go test → functionality, go vet → bugs, go fmt → style, etc.)
  - Provide instructions for fixing common failures (e.g., run `go fmt ./...` to fix format issues)
  - Explain how to view CI results (GitHub Actions tab, PR checks)
  - Include command to run all checks locally before pushing

- [ ] 7. Test workflow end-to-end
  - Push a commit with an intentional test failure (e.g., fail a test or syntax error)
  - Verify workflow runs and shows red status in PR
  - Verify specific step that failed is highlighted in workflow output
  - Fix the intentional failure and push commit
  - Verify workflow re-runs and shows green status
  - Confirm merge is allowed only after workflow passes
  - Verify status badge updated to reflect latest status

## Notes

- Backend and mobile CI checks run in parallel within the workflow matrix for efficiency
- All checks must be deterministic and not depend on external service availability (except Codecov)
- Documentation should include troubleshooting steps for developers running checks locally
- End-to-end testing validates both success and failure scenarios to ensure workflow reliability
