# Development Guide — img-settings

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | `go version` to verify |
| Xcode Command Line Tools | latest | macOS only; needed for CGO |
| `mingw-w64` | any | macOS only; needed for Windows cross-compile (`brew install mingw-w64`) |
| `fyne` CLI | latest | Optional; only needed for `.app` bundle (`make fyne-setup`) |

---

## Getting Started

```bash
# 1. Clone the repository
git clone https://github.com/esteban-plaza/img-settings.git
cd img-settings

# 2. Download Go modules
go mod download

# 3. Quick dev build (native platform, CGO enabled)
make dev
# → dist/img-settings-gui-dev
```

---

## Build Targets

```bash
make dev          # Fast native dev build — use during development
make macos        # macOS arm64 + amd64 binaries → dist/
make windows      # Windows amd64 .exe → dist/  (requires mingw-w64)
make cli          # CLI binary (CGO=0) → dist/img-settings-cli
make all          # macos + windows + cli
make bundle-macos # macOS .app bundle (requires fyne CLI: make fyne-setup)
make clean        # Remove dist/ and gui/*.syso
```

All outputs land in `dist/`. The version string is derived from the most recent git tag via `git describe --tags --always --dirty`.

### Linux headless compile check (CI only)

```bash
sudo apt-get install -y libgl1-mesa-dev xorg-dev
CGO_ENABLED=1 go build -o /tmp/img-settings-gui ./gui/
```

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run only lib tests (pure Go, no display needed)
go test ./lib/

# Run with coverage
go test ./lib/ -coverprofile=coverage.out
go tool cover -func=coverage.out

# Open coverage report in browser
go tool cover -html=coverage.out

# Run a specific test
go test ./lib/ -run TestReadEXIF

# Verbose output
go test -v ./lib/
```

### Coverage thresholds

| Package | Target |
|---------|--------|
| `lib/`  | 80%    |
| `cli/`  | 60%    |
| `gui/`  | best-effort |

---

## Static Analysis

```bash
# Vet (mandatory — same check as CI)
go vet ./...

# Format (run before every commit)
gofmt -w .

# Or with goimports (preferred)
goimports -w .
```

---

## CLI Usage

```bash
# Build the CLI
make cli

# Basic usage
./dist/img-settings-cli photo.jpg

# Batch folder
./dist/img-settings-cli -out processed/ photos/

# With custom opacity
./dist/img-settings-cli -opacity 0.6 photo.jpg

# Flags
#   -out <dir>       Output folder (default: "out")
#   -opacity <0-1>   Watermark opacity (default: 0.82)
```

---

## GUI Usage

```bash
# Build and run
make dev
./dist/img-settings-gui-dev
```

Drag-and-drop a photo or folder onto the window. Adjust opacity and output directory in the toolbar. Processed files appear in the results view with success/error status.

---

## Project Structure (quick reference)

```
lib/           # Core logic — EXIF, image decode, watermark, file helpers
               # Pure Go (CGO=0 safe for lib + cli packages)
gui/           # Fyne GUI — CGO required
cli/           # CLI — CGO=0 single static binary
.github/
  workflows/
    ci.yml       # go vet + build checks (push & PR)
    release.yml  # semantic-release + cross-platform binaries
docs/          # User guides (EN + ES)
scripts/       # Screenshot automation, sample generation
dist/          # Build outputs (git-ignored)
```

---

## Git Workflow

### Creating a Feature Branch

```bash
# 1. Start from latest main
git checkout main
git pull origin main

# 2. Create feature branch
git checkout -b feature/<short-description>
# Examples:
#   feature/cli-output-flag
#   fix/arw-orientation
#   chore/update-deps

# 3. Work in baby steps — one commit per logical change
# 4. Push and open a PR against main
git push -u origin feature/<short-description>
```

### Commit Message Format (Conventional Commits)

```
<type>(<scope>): <subject>
```

| Type | Scope examples | When to use |
|------|---------------|-------------|
| `feat` | `lib`, `gui`, `cli` | New user-visible feature |
| `fix` | `lib`, `gui`, `cli` | Bug fix |
| `docs` | `docs`, `readme` | Documentation only |
| `chore` | `ci`, `deps`, `build` | Tooling / CI change |
| `refactor` | `lib`, `gui` | Code restructure, no behaviour change |
| `test` | `lib`, `cli` | Adding or fixing tests |
| `perf` | `lib` | Performance improvement |

Examples:
```
feat(cli): add --quality flag for JPEG output
fix(lib): correct EXIF orientation for tag value 8
chore(ci): cache Go modules between CI runs
test(lib): add table-driven tests for OutputPath
docs: add development guide to ai-specs
```

### Merging

- All PRs require CI to pass before merge.
- Merge to `main` triggers semantic-release if the commit history includes a releasable change (`feat` or `fix`).

---

## Release Process

Releases are fully automated via semantic-release and GitHub Actions.

1. Merge a PR with a `feat` or `fix` commit to `main`.
2. `release.yml` runs semantic-release, which:
   - Determines the next semver version from commit messages.
   - Creates a GitHub Release with auto-generated release notes.
   - Triggers parallel build jobs for macOS (arm64/amd64) and Windows (amd64).
   - Uploads all binaries as release assets.
3. No manual version bumping — the commit message type drives the version.

| Commit type | Version bump |
|-------------|-------------|
| `fix` | Patch (1.0.0 → 1.0.1) |
| `feat` | Minor (1.0.0 → 1.1.0) |
| `BREAKING CHANGE` footer | Major (1.0.0 → 2.0.0) |

---

## Definition of Done Checklist

Before merging any pull request, verify:

**Development**
- [ ] Feature or fix is implemented following the Go standards in `go-standards.mdc`
- [ ] No new `interface{}` usage without justification
- [ ] Error paths are handled and wrapped with context
- [ ] No `panic` in `lib/` code
- [ ] `lib/` does not import `gui/` or `cli/`

**Testing**
- [ ] New behaviour is covered by unit tests
- [ ] Table-driven tests used where multiple inputs are tested
- [ ] `go test ./lib/` passes with ≥ 80% statement coverage
- [ ] `go test ./cli/` passes with ≥ 60% statement coverage
- [ ] No tests skipped without a documented reason

**Code Quality**
- [ ] `go vet ./...` passes with zero warnings
- [ ] `gofmt -l .` outputs nothing (no unformatted files)
- [ ] Exported symbols have doc comments
- [ ] No dead code or commented-out blocks committed

**CI**
- [ ] All CI checks pass (go vet + lib/cli build + GUI headless build)
- [ ] No new build warnings

**Documentation**
- [ ] `docs/USER_GUIDE.md` updated if user-facing behaviour changed
- [ ] `docs/USER_GUIDE.es.md` updated to match English guide
- [ ] `README.md` updated if setup or usage changed
- [ ] Commit messages follow Conventional Commits format

---

## Troubleshooting

### `CGO_ENABLED=1` build fails on Linux

Install the required OpenGL and X11 headers:
```bash
sudo apt-get install -y libgl1-mesa-dev xorg-dev
```

### `x86_64-w64-mingw32-gcc` not found

Install mingw-w64 (macOS):
```bash
brew install mingw-w64
```

### ARW files not processed

img-settings uses three fallback strategies for Sony RAW files:
1. Embedded JPEG extraction (no tools needed)
2. `dcraw` — `brew install dcraw`
3. ImageMagick — `brew install imagemagick`

If all three fail, the file is skipped with an error message.

### macOS quarantine warning ("unidentified developer")

Remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine /path/to/img-settings
```

### Fyne window does not open

Ensure CGO is enabled and you are running on a platform with a display:
```bash
CGO_ENABLED=1 go run ./gui/
```
