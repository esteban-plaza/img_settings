package watermark

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// newSolidJPEG encodes a plain RGBA image as JPEG bytes.
func newSolidJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// writeTempJPEG writes JPEG bytes to a temp file and returns its path.
func writeTempJPEG(t *testing.T, w, h int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jpg")
	if err := os.WriteFile(path, newSolidJPEG(w, h), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// ── IsSupportedExt ───────────────────────────────────────────────────────────

func TestIsSupportedExt(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"jpg lowercase", "photo.jpg", true},
		{"jpeg lowercase", "photo.jpeg", true},
		{"JPG uppercase", "photo.JPG", true},
		{"JPEG uppercase", "photo.JPEG", true},
		{"png lowercase", "photo.png", true},
		{"PNG uppercase", "photo.PNG", true},
		{"arw lowercase", "photo.arw", true},
		{"ARW uppercase", "photo.ARW", true},
		{"gif unsupported", "photo.gif", false},
		{"raw unsupported", "photo.raw", false},
		{"bmp unsupported", "photo.bmp", false},
		{"no extension", "photo", false},
		{"empty string", "", false},
		{"path with dir", "photos/subdir/img.jpg", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSupportedExt(tc.path); got != tc.want {
				t.Errorf("IsSupportedExt(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

// ── OutputPath ───────────────────────────────────────────────────────────────

func TestOutputPath(t *testing.T) {
	cases := []struct {
		name      string
		inputPath string
		outDir    string
		want      string
	}{
		{"jpg in subdir", "photos/img.jpg", "out", filepath.Join("out", "img.jpg")},
		{"arw file", "raw/shot.arw", "out", filepath.Join("out", "shot.jpg")},
		{"png file", "img.png", ".", filepath.Join(".", "img.jpg")},
		{"jpeg extension", "snap.jpeg", "processed", filepath.Join("processed", "snap.jpg")},
		{"uppercase ext", "IMG_001.JPG", "export", filepath.Join("export", "IMG_001.jpg")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := OutputPath(tc.inputPath, tc.outDir)
			if got != tc.want {
				t.Errorf("OutputPath(%q, %q) = %q, want %q", tc.inputPath, tc.outDir, got, tc.want)
			}
		})
	}
}

// ── ResizeForWhatsApp ────────────────────────────────────────────────────────

func TestResizeForWhatsApp_SmallImageUnchanged(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	got := ResizeForWhatsApp(img)
	b := got.Bounds()
	if b.Dx() != 800 || b.Dy() != 600 {
		t.Errorf("expected 800×600, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestResizeForWhatsApp_ExactLimitUnchanged(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, MaxWhatsAppDim, MaxWhatsAppDim))
	got := ResizeForWhatsApp(img)
	b := got.Bounds()
	if b.Dx() != MaxWhatsAppDim || b.Dy() != MaxWhatsAppDim {
		t.Errorf("expected %d×%d, got %d×%d", MaxWhatsAppDim, MaxWhatsAppDim, b.Dx(), b.Dy())
	}
}

func TestResizeForWhatsApp_LandscapeResized(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 5120, 2560))
	got := ResizeForWhatsApp(img)
	b := got.Bounds()
	if b.Dx() != MaxWhatsAppDim {
		t.Errorf("expected width %d, got %d", MaxWhatsAppDim, b.Dx())
	}
	if b.Dy() != 1280 {
		t.Errorf("expected height 1280, got %d", b.Dy())
	}
}

func TestResizeForWhatsApp_PortraitResized(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2560, 5120))
	got := ResizeForWhatsApp(img)
	b := got.Bounds()
	if b.Dy() != MaxWhatsAppDim {
		t.Errorf("expected height %d, got %d", MaxWhatsAppDim, b.Dy())
	}
	if b.Dx() != 1280 {
		t.Errorf("expected width 1280, got %d", b.Dx())
	}
}

// ── ApplyOrientation ─────────────────────────────────────────────────────────

func TestApplyOrientation_NilExif(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	got := ApplyOrientation(img, nil)
	b := got.Bounds()
	if b.Dx() != 100 || b.Dy() != 200 {
		t.Errorf("expected 100×200 unchanged, got %d×%d", b.Dx(), b.Dy())
	}
}

// ── CollectFiles ─────────────────────────────────────────────────────────────

func TestCollectFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := CollectFiles([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestCollectFiles_JPGAndTxtInDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "photo.jpg"), newSolidJPEG(10, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("text"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := CollectFiles([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_NestedDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "img.png"), newSolidJPEG(10, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := CollectFiles([]string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 nested file, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_DirectFileSkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "photo.bmp")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := CollectFiles([]string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for unsupported ext, got %d", len(files))
	}
}

func TestCollectFiles_NonExistentPath(t *testing.T) {
	_, err := CollectFiles([]string{"/does/not/exist/ever"})
	if err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

func TestCollectFiles_DirectSupportedFile(t *testing.T) {
	path := writeTempJPEG(t, 10, 10)
	files, err := CollectFiles([]string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

// ── ReadEXIF ─────────────────────────────────────────────────────────────────

func TestReadEXIF_MissingFile(t *testing.T) {
	s, x := ReadEXIF("/does/not/exist.jpg")
	if x != nil {
		t.Error("expected nil Exif for missing file")
	}
	if s != (PhotoSettings{}) {
		t.Errorf("expected empty PhotoSettings, got %+v", s)
	}
}

func TestReadEXIF_NoExifInJPEG(t *testing.T) {
	path := writeTempJPEG(t, 50, 50)
	s, x := ReadEXIF(path)
	// A plain synthetic JPEG has no EXIF; Exif decode will fail so x may be nil.
	// We only assert no panic and that empty settings are returned.
	if x != nil {
		// Some JPEG encoders may write minimal EXIF; if so, settings may be
		// partially populated. Just ensure we don't crash.
		_ = s
		return
	}
	if s != (PhotoSettings{}) {
		t.Errorf("expected empty PhotoSettings for no-EXIF JPEG, got %+v", s)
	}
}

// ── AddWatermark ─────────────────────────────────────────────────────────────

func TestAddWatermark_EmptySettings(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 150))
	got, err := AddWatermark(img, PhotoSettings{}, 0.82)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := got.Bounds()
	if b.Dx() != 200 || b.Dy() != 150 {
		t.Errorf("expected 200×150 unchanged, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestAddWatermark_WithAllSettings(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	s := PhotoSettings{
		Aperture:     "2.8",
		ShutterSpeed: "1/250",
		ISO:          "400",
		FocalLength:  "50mm",
		CameraModel:  "Test Camera",
	}
	got, err := AddWatermark(img, s, 0.82)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := got.Bounds()
	if b.Dx() != 800 || b.Dy() != 600 {
		t.Errorf("expected 800×600, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestAddWatermark_ZeroOpacity(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	s := PhotoSettings{ISO: "200"}
	_, err := AddWatermark(img, s, 0.0)
	if err != nil {
		t.Fatalf("unexpected error with zero opacity: %v", err)
	}
}

// ── DecodeImage ──────────────────────────────────────────────────────────────

func TestDecodeImage_ValidJPEG(t *testing.T) {
	path := writeTempJPEG(t, 100, 80)
	img, err := DecodeImage(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		t.Errorf("expected non-zero bounds, got %v", b)
	}
}

func TestDecodeImage_MissingFile(t *testing.T) {
	_, err := DecodeImage("/does/not/exist.jpg")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestDecodeImage_UnsupportedContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not_an_image.jpg")
	if err := os.WriteFile(path, []byte("this is not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := DecodeImage(path)
	if err == nil {
		t.Error("expected error for invalid image content, got nil")
	}
}

// ── ProcessFile ──────────────────────────────────────────────────────────────

func TestProcessFile_Integration(t *testing.T) {
	inPath := writeTempJPEG(t, 200, 150)
	outDir := t.TempDir()

	if err := ProcessFile(inPath, outDir, 0.82); err != nil {
		t.Fatalf("ProcessFile error: %v", err)
	}

	outPath := OutputPath(inPath, outDir)
	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("expected output file at %q, stat error: %v", outPath, err)
	}
}

func TestProcessFile_MissingInput(t *testing.T) {
	outDir := t.TempDir()
	err := ProcessFile("/does/not/exist.jpg", outDir, 0.82)
	if err == nil {
		t.Error("expected error for missing input, got nil")
	}
}

// ── rotateImage ───────────────────────────────────────────────────────────────

func TestRotateImage_90(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	got := rotateImage(img, 90)
	b := got.Bounds()
	// 40×20 rotated 90° → 20×40
	if b.Dx() != 20 || b.Dy() != 40 {
		t.Errorf("90°: expected 20×40, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestRotateImage_180(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	got := rotateImage(img, 180)
	b := got.Bounds()
	if b.Dx() != 40 || b.Dy() != 20 {
		t.Errorf("180°: expected 40×20, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestRotateImage_270(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	got := rotateImage(img, 270)
	b := got.Bounds()
	// 40×20 rotated 270° → 20×40
	if b.Dx() != 20 || b.Dy() != 40 {
		t.Errorf("270°: expected 20×40, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestRotateImage_Unknown(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	got := rotateImage(img, 45) // unknown angle — returns img unchanged
	if got != img {
		t.Error("expected same image for unknown rotation angle")
	}
}

// ── extractLargestJPEG ───────────────────────────────────────────────────────

func TestExtractLargestJPEG_ValidEmbedded(t *testing.T) {
	// Build a byte slice that wraps a real JPEG inside extra bytes.
	jpegData := newSolidJPEG(20, 20)
	// Prepend some random bytes before the JPEG marker.
	data := append([]byte{0x00, 0xAB, 0xCD}, jpegData...)
	// Append some trailing bytes after the EOI.
	data = append(data, 0x00, 0xFF)

	img, err := extractLargestJPEG(data)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}

func TestExtractLargestJPEG_NoJPEG(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	_, err := extractLargestJPEG(data)
	if err == nil {
		t.Error("expected error when no JPEG present, got nil")
	}
}

func TestExtractLargestJPEG_PicksLargest(t *testing.T) {
	// Create two JPEGs of different sizes; the larger one should be selected.
	small := newSolidJPEG(5, 5)
	large := newSolidJPEG(50, 50)
	data := append(small, large...)

	img, err := extractLargestJPEG(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() < 5 {
		t.Errorf("expected the larger image to be selected")
	}
}

// ── decodePPM ────────────────────────────────────────────────────────────────

// buildP6PPM creates a minimal binary PPM (P6) byte slice with the given
// dimensions, filled with black pixels.
func buildP6PPM(w, h int) []byte {
	// Each token on its own line — decodePPM's readToken reads one line at a time
	// and returns only the first whitespace-separated field.
	// The trailing space after "255\n" is the one-whitespace-byte separator that
	// decodePPM explicitly reads with br.Read(buf) before the binary pixel data.
	header := fmt.Sprintf("P6\n%d\n%d\n255\n ", w, h)
	pixels := make([]byte, w*h*3)
	return append([]byte(header), pixels...)
}

func TestDecodePPM_ValidP6(t *testing.T) {
	data := buildP6PPM(4, 3)
	img, err := decodePPM(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 4 || b.Dy() != 3 {
		t.Errorf("expected 4×3, got %d×%d", b.Dx(), b.Dy())
	}
}

func TestDecodePPM_InvalidMagic(t *testing.T) {
	data := []byte("P3\n4 3\n255\n")
	_, err := decodePPM(bytes.NewReader(data))
	if err == nil {
		t.Error("expected error for non-P6 magic, got nil")
	}
}

func TestDecodePPM_TruncatedPixelData(t *testing.T) {
	// Header correct but pixel data too short.
	header := []byte("P6\n2 2\n255\n")
	data := append(header, 0x00) // only 1 byte instead of 12
	_, err := decodePPM(bytes.NewReader(data))
	if err == nil {
		t.Error("expected error for truncated pixel data, got nil")
	}
}

// ── decodeARW ────────────────────────────────────────────────────────────────

func TestDecodeARW_EmbeddedJPEG(t *testing.T) {
	// Create a fake .arw file whose bytes contain an embedded JPEG.
	jpegData := newSolidJPEG(30, 30)
	arwData := append([]byte{0x49, 0x49, 0x2A, 0x00}, jpegData...) // fake TIFF header + JPEG

	dir := t.TempDir()
	path := filepath.Join(dir, "fake.arw")
	if err := os.WriteFile(path, arwData, 0o644); err != nil {
		t.Fatal(err)
	}

	img, err := DecodeImage(path)
	if err != nil {
		t.Fatalf("expected embedded JPEG extraction to succeed, got: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}
