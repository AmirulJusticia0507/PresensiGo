# Requirements Document

## Introduction

PresensiGo currently lacks automated continuous integration and continuous deployment (CI/CD) checks. Code changes are not automatically tested, linted, or validated before merge. This creates technical debt and increases the risk of regressions.

This specification establishes a GitHub Actions workflow that runs on every push and pull request, performing automated checks for both backend (Go) and mobile (Flutter). The workflow ensures code quality, catches bugs early, and provides visibility into build status via badges.

## Requirements

### R1: GitHub Actions Workflow

**Requirement:** Create a GitHub Actions workflow file that triggers on PR/push events and runs all checks.

- Workflow file: `.github/workflows/ci.yml`
- Triggers: `on: [push, pull_request]`
- Runs for all branches; can optionally skip tags
- Provides clear pass/fail status to PR reviewers
- Fails the PR if any check fails; prevents merge until fixed

**Rationale:** Automated checks catch issues before code review, saving time and reducing defects in main branch.

### R2: Backend Checks

**Requirement:** Run Go-specific code quality and test checks on the backend.

- `go test ./...` — Run all unit tests; fail if any test fails
- `go vet ./...` — Run Go's static analyzer; fail if issues found
- `go fmt -l` — Check code formatting; fail if code is not formatted correctly
- `golint` (or `golangci-lint`) — Run linter; fail if issues found

**Rationale:** These checks are standard Go practices. Tests verify functionality, vet catches common bugs, fmt ensures consistency, lint enforces style guidelines.

### R3: Mobile Checks

**Requirement:** Run Flutter-specific code quality and test checks on the mobile app.

- `flutter analyze` — Analyze Dart code for issues; fail if issues found
- `flutter test` — Run all widget/unit tests; fail if any test fails
- Optional: `flutter build apk --analyze` — Verify APK builds successfully

**Rationale:** Flutter analyze catches code quality issues early. Tests verify mobile functionality. Build verification ensures no compile-time errors.

### R4: Code Coverage Reporting

**Requirement:** Generate and report code coverage metrics for backend and mobile (optional but recommended).

- Backend: Generate coverage report via `go test -cover ./...`
- Mobile: Generate coverage report via `flutter test --coverage`
- Upload coverage to code coverage service (e.g., Codecov, Coveralls)
- Display coverage percentage in PR or as artifact

**Rationale:** Coverage metrics help track testing quality and identify under-tested areas.

### R5: Fail PR If Any Check Fails

**Requirement:** Configure GitHub to require all workflow checks to pass before PR merge.

- Set workflow as required status check in branch protection rules
- PR cannot be merged if workflow is not green
- Developers must fix failures and re-push to update PR status

**Rationale:** Enforcing passing checks prevents broken code from reaching main branch.

### R6: Status Badge in README

**Requirement:** Add a status badge to README that displays current CI workflow status.

- Badge shows "passing" (green) or "failing" (red)
- Badge links to latest workflow run
- Demonstrates project health to users/contributors

**Rationale:** Status badge provides quick visibility into project health. Shows contributors that CI is active and tests matter.

### R7: Document CI Process

**Requirement:** Add documentation in README or CONTRIBUTING.md explaining the CI pipeline.

- Explain when checks run (push, PR)
- List what each check verifies
- Explain how to fix common failures (format with `go fmt`, run tests locally, etc.)
- Explain how to view CI results (GitHub Actions tab, PR checks)

**Rationale:** Documentation helps developers understand expectations and quickly fix failures.

## Success Criteria

1. ✅ `.github/workflows/ci.yml` exists and is syntactically valid
2. ✅ Workflow triggers on push and PR to any branch
3. ✅ Backend job runs `go test ./...`, `go vet ./...`, `go fmt -l`, and `golint`
4. ✅ Mobile job runs `flutter analyze` and `flutter test`
5. ✅ Workflow fails if any check fails; status shown in PR
6. ✅ Code coverage report generated and accessible (via artifact or external service)
7. ✅ Status badge added to README showing pass/fail and linking to latest run
8. ✅ CONTRIBUTING.md or README documents the CI process
9. ✅ Branch protection rule requires CI workflow to pass before merge
10. ✅ Workflow tested end-to-end: intentional failure confirms checks work

## Glossary

- **GitHub Actions:** GitHub's built-in CI/CD service; runs workflows triggered by events (push, PR, schedule)
- **Workflow:** A YAML file in `.github/workflows/` that defines steps to run
- **Job:** A set of steps that run on a single runner; can run in parallel with other jobs
- **Step:** A single command or action executed within a job
- **Status Check:** A required or optional check that must pass before PR merge (e.g., "CI / Build")
- **Coverage:** Percentage of code lines/branches executed by tests; higher = better testing
- **Artifact:** Output files from a workflow run (e.g., coverage reports, build outputs)
