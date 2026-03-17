# Implementation Plan: SPEC-001 — Spec Compliance Fixes

## Overview

A full audit of the codebase revealed five categories of spec violations: missing `var version` declarations,
swallowed errors, missing doc comments on all exported `lib/` symbols, and zero test coverage. This plan
addresses every finding in priority order, working in baby steps with one commit per logical change.

**Key principles applied:**
- TDD: failing test before implementation
- Baby steps: one commit per step
- lib/ isolation: core logic stays in lib/, no GUI imports

## Architecture Context

### Files to Modify
- `cli/main.go` — add `var version string` for ldflags injection
- `gui/main.go` — add `var version string` for ldflags injection
- `lib/watermark.go` — fix swallowed error (CameraModel); add doc comments on all exported symbols
- `gui/util.go` — fix swallowed Walk error

### New Files to Create
- `lib/watermark_test.go` — table-driven unit tests for all exported lib/ functions

---

## Implementation Steps

### Step 0: Create Feature Branch

```bash
git checkout main && git pull origin main
git checkout -b feature/SPEC-001-spec-compliance
```

---

### Step 1: Add `var version string` to CLI and GUI

**Files:** `cli/main.go`, `gui/main.go`

Add `var version string` at package level in both `main` packages so the ldflags injection
(`-X main.version=...`) has a receiver. Without this the injected value is silently discarded.

Expected additions:
```go
// cli/main.go — after the import block
var version string

// gui/main.go — after the import block
var version string
```

**Commit:** `fix(cli): declare version var for ldflags injection`
           `fix(gui): declare version var for ldflags injection`

---

### Step 2: Fix Swallowed Error — CameraModel

**File:** `lib/watermark.go:96`

Current:
```go
s.CameraModel, _ = model.StringVal()
```

Fix: capture the error and leave CameraModel empty on failure (the value is already inside an
`if err == nil` guard, so failing the StringVal conversion is a genuine edge case that should
not be silently ignored):
```go
if val, err := model.StringVal(); err == nil {
    s.CameraModel = strings.TrimSpace(val)
}
```

Remove the standalone `strings.TrimSpace` call on the next line (line 97) since it is now
inlined above.

**Commit:** `fix(lib): stop swallowing CameraModel StringVal error in ReadEXIF`

---

### Step 3: Fix Swallowed Walk Error in GUI

**File:** `gui/util.go:38`

Current:
```go
_ = filepath.Walk(p, func(...) error { ... })
```

Fix: log the error to stderr (GUI cannot return errors from collectFiles without a larger
refactor; logging is the least-surprise behaviour):
```go
if err := filepath.Walk(p, func(...) error { ... }); err != nil {
    fmt.Fprintf(os.Stderr, "walk %q: %v\n", p, err)
}
```

Add `"fmt"` to the import block if not already present.

**Commit:** `fix(gui): log Walk error instead of silently discarding it`

---

### Step 4: Add Doc Comments to All Exported lib/ Symbols

**File:** `lib/watermark.go`

Add Go doc comments immediately above each exported symbol:

| Symbol | Comment |
|--------|---------|
| `MaxWhatsAppDim` | Maximum pixel dimension (longest side) for WhatsApp HD output. |
| `JPEGQuality` | JPEG encoding quality used for all output files. |
| `PhotoSettings` | PhotoSettings holds the EXIF camera-settings fields extracted from a photo. |
| `ReadEXIF` | ReadEXIF reads EXIF metadata from the file at path and returns the camera settings and the raw Exif value (used for orientation). On any read or decode failure an empty PhotoSettings and nil Exif are returned so the caller can still process the image without metadata. |
| `DecodeImage` | DecodeImage decodes a JPEG, PNG, or ARW file into an image.Image. ARW files are decoded via three fallback strategies: embedded JPEG, dcraw, then ImageMagick. |
| `ApplyOrientation` | ApplyOrientation rotates img according to the EXIF orientation tag. It is a no-op when x is nil or the tag is absent. |
| `ResizeForWhatsApp` | ResizeForWhatsApp downsizes img so its longest side is at most MaxWhatsAppDim pixels, preserving the aspect ratio with CatmullRom resampling. Images already within the limit are returned unchanged. |
| `AddWatermark` | AddWatermark stamps a translucent pill-shaped overlay containing the camera settings from s onto img. The overlay opacity is controlled by the opacity parameter (0 = invisible, 1 = fully opaque). If s contains no data the image is returned unchanged. |
| `IsSupportedExt` | IsSupportedExt reports whether path has a supported image extension (arw, jpg, jpeg, png). |
| `CollectFiles` | CollectFiles expands the given paths (files or directories) into a flat list of supported image files. Unsupported files are skipped with a message to stderr. Directories are walked recursively. |
| `OutputPath` | OutputPath returns the output file path for inputPath under outDir, replacing the extension with .jpg. |
| `ProcessFile` | ProcessFile reads, orientates, resizes, watermarks, and writes inputPath as a JPEG to outDir. It is the single entry point for processing one image. |

**Commit:** `docs(lib): add doc comments to all exported symbols`

---

### Step 5: Write Failing Tests

**File:** `lib/watermark_test.go` (new file)

Write table-driven tests. Run `go test ./lib/` — tests must FAIL before Step 6 only if
they exercise unimplemented paths; for existing code they should immediately pass (TDD here
means writing the tests first and verifying they exercise real behaviour).

**Tests to include:**

#### `TestIsSupportedExt`
Table: `.jpg`, `.jpeg`, `.JPG`, `.arw`, `.ARW`, `.png`, `.PNG`, `.gif`, `.raw`, `""` → expected bool.

#### `TestOutputPath`
Table: `("photos/img.jpg", "out")→"out/img.jpg"`, `("raw/shot.arw", "out")→"out/shot.jpg"`,
`("img.png", ".")→"./img.jpg"`.

#### `TestResizeForWhatsApp_NoResize`
Image already ≤ 2560 on both sides → same pixel dimensions returned.

#### `TestResizeForWhatsApp_LandscapeResize`
Image 5120×2560 → 2560×1280.

#### `TestResizeForWhatsApp_PortraitResize`
Image 2560×5120 → 1280×2560.

#### `TestApplyOrientation_Nil`
Passing nil Exif → same image returned.

#### `TestCollectFiles_Empty`
Temp dir with no files → empty slice, no error.

#### `TestCollectFiles_WithFiles`
Temp dir with one `.jpg` and one `.txt` → only the jpg in the result.

#### `TestCollectFiles_Nested`
Temp dir with nested subdir containing a `.jpg` → file collected when walking.

#### `TestCollectFiles_UnsupportedFile`
Single `.bmp` path → empty result (skipped with message to stderr).

#### `TestReadEXIF_MissingFile`
Non-existent path → empty PhotoSettings, nil Exif (no panic).

#### `TestReadEXIF_NoExif`
Valid JPEG with no EXIF data → empty PhotoSettings, nil Exif.

#### `TestAddWatermark_EmptySettings`
Synthetic 100×100 image + empty PhotoSettings → returned image is unchanged (same bounds).

#### `TestAddWatermark_WithSettings`
Synthetic 800×600 image + populated PhotoSettings → returned image has same bounds, no error.

#### `TestDecodeImage_UnsupportedFormat`
Non-existent path → error returned.

#### `TestDecodeImage_ValidJPEG`
Temp JPEG file created programmatically → image decoded successfully.

#### `TestProcessFile_Integration`
Temp JPEG file + temp output dir → output JPEG created, no error.

**Commit:** `test(lib): add table-driven tests covering all exported functions`

---

### Step 6: Verify Coverage

```bash
go test ./lib/ -coverprofile=coverage.out
go tool cover -func=coverage.out  # target ≥ 80%
go vet ./...
```

If coverage is below 80%, add additional test cases for uncovered branches before committing.

---

### Step 7: Final Check

```bash
go vet ./...
go test ./...
```

Both must pass with zero errors.

---

## Definition of Done

- [x] `var version string` declared in cli/ and gui/
- [x] No swallowed errors in lib/ or gui/
- [x] All exported lib/ symbols have doc comments
- [x] `go test ./lib/` passes with ≥ 80% statement coverage
- [x] Table-driven tests used throughout
- [x] `go vet ./...` clean
- [x] Every commit follows Conventional Commits format
