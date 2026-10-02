# Contributing to PresensiGo

Thank you for your interest in contributing to PresensiGo! This guide will help you understand our development workflow, CI/CD process, and how to ensure your contributions meet our quality standards.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/PresensiGo.git`
3. Create a feature branch: `git checkout -b feat/your-feature`
4. Make your changes
5. Run tests locally (see "Running Checks Locally" below)
6. Push to your fork and open a Pull Request

## Continuous Integration (CI) Process

Our CI/CD pipeline automatically runs checks on every push and pull request. Understanding this process helps you fix issues quickly and avoid rejected PRs.

### When CI Runs

The CI workflow is **automatically triggered** on:
- **Every push** to any branch (`main`, `develop`, `fix/**`, `feat/**`)
- **Every pull request** targeting `main` or `develop` branches

CI runs for all branches and PRs. For PRs targeting `main` or `develop`, the checks are **required** to pass before merging.

### CI Jobs

The workflow runs two main jobs in **parallel** for efficiency:

#### Backend (Go) Checks
Validates the Go API codebase:

1. **`go test ./...`** — Runs all unit and integration tests
   - Verifies functionality of all packages
   - Fails if any test fails or panics
   - Output: Pass/Fail status

2. **`go vet ./...`** — Static analysis to detect potential bugs
   - Detects suspicious code patterns (e.g., unreachable code, incorrect printf usage)
   - Fails if vet identifies issues
   - Output: Warning/Error messages

3. **`go fmt -l`** — Ensures code style consistency
   - Checks if code follows Go formatting standards
   - Fails if any file needs formatting
   - Output: List of files that need formatting

4. **`golangci-lint`** — Comprehensive linting suite
   - Runs multiple linters (vet, fmt, goimports, errcheck, staticcheck, etc.)
   - Detects code quality issues and potential bugs
   - Fails if issues are found
   - Output: Lint errors with file/line information

5. **Code Coverage** — Generates coverage metrics
   - Runs tests with coverage tracking
   - Uploads to [Codecov](https://codecov.io) for tracking trends
   - Does not block merge if upload fails

#### Mobile (Flutter) Checks
Validates the Flutter mobile app:

1. **`flutter analyze`** — Code analysis for Dart/Flutter
   - Analyzes code for errors and style violations
   - Fails on high-severity issues (warnings are ignored with `--no-fatal-infos`)
   - Output: Analysis errors/warnings

2. **`flutter test`** — Runs all widget and unit tests
   - Executes tests in the `test/` directory
   - Generates coverage report (`coverage/lcov.info`)
   - Fails if any test fails or crashes
   - Output: Pass/Fail per test

3. **Code Coverage** — Generates coverage metrics
   - Tracks test coverage percentage
   - Uploads to [Codecov](https://codecov.io)
   - Does not block merge if upload fails

### Overall Status

The workflow passes **only if both backend and mobile jobs succeed**. If either job fails, the entire workflow is marked as failed, and the PR cannot be merged.

## Viewing CI Results

### GitHub Actions Tab

1. Go to the **Actions** tab in the repository
2. Click the workflow run you want to inspect
3. Click the job name (Backend or Mobile) to expand details
4. Scroll through the logs to find the failed step

### PR Checks Section

1. Open your Pull Request
2. Scroll to the bottom to see the **Checks** section
3. Each check shows as ✅ (pass) or ❌ (fail)
4. Click "Details" next to a failed check to see the error

### Direct Links

- **Latest workflow runs:** https://github.com/AmirulJusticia0507/PresensiGo/actions/workflows/ci.yml
- **Branch status:** Check the workflow badge at the top of README.md

## Fixing Common CI Failures

### Backend (Go) Failures

#### Test Failure: `go test ./...`

**Error:** Tests are failing in CI but may pass locally

**Steps to fix:**
```bash
cd backend
go test ./... -v  # Run locally to see which test failed
# Read the error message and fix the code
go test ./... -v  # Re-run to verify
```

**Common causes:**
- Assertion failed (e.g., expected value doesn't match actual)
- Race condition (use `-race` flag: `go test ./... -race`)
- Missing mocked dependency
- Timeout (increase timeout if necessary)

#### Vet Error: `go vet ./...`

**Error:** Static analysis detected suspicious code

**Steps to fix:**
```bash
cd backend
go vet ./...  # See the specific errors
# Fix issues reported by vet
go vet ./...  # Verify fix
```

**Common issues:**
- Unused variables: remove them or prefix with `_`
- Unreachable code: remove dead branches
- Incorrect function signatures: check function parameters/return types
- Printf format mismatch: ensure format string matches arguments

#### Format Error: `go fmt -l`

**Error:** Code is not formatted according to Go standards

**Steps to fix:**
```bash
cd backend
go fmt ./...  # This auto-formats all Go files
# Verify fix
go fmt -l ./...  # Should show no files
```

**Note:** `go fmt` modifies files in-place. After running it, commit the formatting changes.

#### Lint Error: `golangci-lint`

**Error:** Linting rules detected code quality issues

**Steps to fix:**
```bash
cd backend
golangci-lint run  # See all lint errors locally
# Fix issues reported (examples below)
# For unused imports, remove or use _
# For unused variables, remove or prefix with _
# For missing error checks, add if err != nil { ... }
golangci-lint run  # Verify fix
```

**Common lint issues:**
- **Unused imports:** Remove the import or use it (if needed for side effects, use `import _ "package"`)
- **Unused variables:** Delete or use them (prefix with `_` if intentionally unused)
- **Error not checked:** Add `if err != nil` handling
- **Unhandled type assertion:** Add `, ok := value.(Type)` check
- **Ineffectual assignment:** Remove or use the variable

### Mobile (Flutter) Failures

#### Analyze Error: `flutter analyze`

**Error:** Dart code has analysis issues

**Steps to fix:**
```bash
cd presensigo_mobile
flutter analyze  # Run locally to see errors
# Fix issues reported
flutter analyze  # Verify fix
```

**Common issues:**
- Unused imports: remove or use with `_`
- Unused local variables: delete or prefix with `_`
- Missing override annotations: add `@override` for overridden methods
- Type mismatches: check parameter/return types
- Null safety issues: add `!` null assertion or `?` nullable check

**Updating analysis_options.yaml:**

If an error is intentional or your team wants to allow it, you can configure the linter rules in `presensigo_mobile/analysis_options.yaml`:

```yaml
linter:
  rules:
    - avoid_print  # Enable this rule
    # - specific_rule  # Disable by commenting out
```

#### Test Failure: `flutter test`

**Error:** Unit or widget tests are failing

**Steps to fix:**
```bash
cd presensigo_mobile
flutter test  # Run tests locally
# Read the error output to identify the failing test
# Fix the code or test based on the error
flutter test --verbose  # Re-run with detailed output
```

**Common causes:**
- Widget build error: check widget parameters and state management
- Assertion failure: expected value doesn't match actual
- Async/await issue: ensure futures are properly awaited
- Mock setup error: verify mock objects are properly configured
- Null pointer exception: check null safety and assertions

## Running All Checks Locally

Before pushing your changes, run all CI checks locally to catch issues early. This saves time and prevents rejected PRs.

### Backend Checks

```bash
cd backend

# Run all checks in sequence
go test ./... -v          # Tests
go vet ./...              # Static analysis
go fmt ./...              # Format code
golangci-lint run         # Linting
```

**One-liner to run all checks:**
```bash
cd backend && go test ./... -v && go vet ./... && go fmt ./... && golangci-lint run && echo "All backend checks passed!"
```

### Mobile Checks

```bash
cd presensigo_mobile

# Run all checks in sequence
flutter pub get           # Get dependencies
flutter analyze           # Code analysis
flutter test              # Run tests
```

**One-liner to run all checks:**
```bash
cd presensigo_mobile && flutter pub get && flutter analyze && flutter test && echo "All mobile checks passed!"
```

### Running Everything (Recommended Before Push)

```bash
# Backend
cd backend && go test ./... -v && go vet ./... && go fmt ./... && golangci-lint run

# Mobile
cd presensigo_mobile && flutter pub get && flutter analyze && flutter test

# If both succeed, safe to push!
echo "✓ All checks passed! Ready to push."
```

Or create a script in your repo root:
```bash
#!/bin/bash
echo "Running backend checks..."
cd backend && go test ./... -v && go vet ./... && go fmt ./... && golangci-lint run || exit 1
echo "Backend checks passed!"

echo "Running mobile checks..."
cd ../presensigo_mobile && flutter pub get && flutter analyze && flutter test || exit 1
echo "Mobile checks passed!"

echo "✓ All checks passed! Ready to push."
```

## Code Coverage

### What is Code Coverage?

Code coverage measures what percentage of your code is executed by tests. Higher coverage means more of your code is tested.

### Viewing Coverage Reports

1. **Local:** After running tests
   ```bash
   cd backend
   go tool cover -html=coverage.out -o coverage.html
   # Open coverage.html in a browser
   ```

2. **On GitHub:** Coverage reports are uploaded to [Codecov](https://codecov.io/gh/AmirulJusticia0507/PresensiGo)
   - View overall coverage: https://codecov.io/gh/AmirulJusticia0507/PresensiGo
   - View per-PR coverage: Check the PR checks section

### Coverage Goals

While coverage isn't a hard requirement to merge, we aim for:
- **Backend:** > 70% coverage for critical paths (auth, attendance, geofencing)
- **Mobile:** > 60% coverage for core features

## Writing Tests

### Backend (Go)

Place test files in the same directory as the code being tested with `_test.go` suffix.

**Example:** For `backend/internal/usecase/auth_usecase.go`, create `backend/internal/usecase/auth_usecase_test.go`

```go
package usecase

import (
    "testing"
)

func TestAuthenticateUser(t *testing.T) {
    // Arrange
    email := "user@example.com"
    password := "password123"
    
    // Act
    result, err := AuthenticateUser(email, password)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result == nil {
        t.Fatal("expected user, got nil")
    }
}
```

### Mobile (Flutter)

Place test files in the `test/` directory with `_test.dart` suffix.

**Example:** For `lib/features/auth/login_screen.dart`, create `test/features/auth/login_screen_test.dart`

```dart
import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_mobile/features/auth/login_screen.dart';

void main() {
  testWidgets('Login button is visible', (WidgetTester tester) async {
    await tester.pumpWidget(const MyApp());
    
    expect(find.byType(ElevatedButton), findsOneWidget);
  });
}
```

## PR Guidelines

1. **Create a feature branch** from `develop` or `main`
2. **Run all checks locally** before pushing
3. **Keep PRs focused** — one feature or fix per PR
4. **Write descriptive commit messages** — explain what and why
5. **Link related issues** — use "Closes #123" in PR description
6. **Wait for CI to pass** — don't merge until all checks are green
7. **Request review** from maintainers

### PR Title Format

```
feat: add user profile endpoint
fix: correct geofencing validation
docs: update API documentation
test: add attendance usecase tests
chore: update dependencies
```

### PR Description Template

```markdown
## What does this PR do?
Brief description of changes.

## Why?
Motivation and context for the changes.

## Testing
How to test the changes locally.

## Screenshots (if UI changes)
Optional screenshots of UI changes.

## Checklist
- [ ] Tests pass locally
- [ ] Code is formatted
- [ ] Linting passes
- [ ] No breaking changes (or documented)
```

## Troubleshooting

### "CI failed but tests pass locally"

**Possible causes:**
- Environment difference (Go version, Flutter version)
- Race conditions (add `-race` flag: `go test ./... -race`)
- Missing test files in git

**Solution:**
1. Check Go version: `go version` (should be 1.21+)
2. Check Flutter version: `flutter --version` (should be 3.13.x)
3. Ensure all test files are committed: `git add .` and `git status`
4. Run tests with race detector: `go test ./... -race`

### "I accidentally pushed to main"

1. Create a new branch from your commit: `git branch recovery-branch`
2. Reset main: `git reset --hard origin/main` (with maintainer approval)
3. Cherry-pick commits if needed: `git cherry-pick <commit-hash>`

### "CI is stuck or taking too long"

Check the Actions tab to see if a job is running. If stuck:
1. Cancel the workflow manually from the Actions tab
2. Commit a small change to trigger a new run

### "Coverage upload failed"

This is non-critical and doesn't block merges. The workflow continues. If you see repeated failures, check:
1. Codecov configuration is correct
2. Coverage files are being generated

## Useful Links

- **CI Workflow:** `.github/workflows/ci.yml`
- **Actions Tab:** https://github.com/AmirulJusticia0507/PresensiGo/actions
- **Codecov Dashboard:** https://codecov.io/gh/AmirulJusticia0507/PresensiGo
- **Go Documentation:** https://golang.org/doc
- **Flutter Documentation:** https://flutter.dev/docs

## Questions?

- **Open an issue** for bugs or feature requests
- **Discuss in PRs** for code review feedback
- **Check existing issues** for answers to common questions

Thank you for contributing to PresensiGo! 🎉
