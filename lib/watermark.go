package watermark

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	exiflib "github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

const (
	MaxWhatsAppDim = 2560
	JPEGQuality    = 92
)

// ── EXIF settings ────────────────────────────────────────────────────────────

type PhotoSettings struct {
	Aperture     string
	ShutterSpeed string
	ISO          string
	FocalLength  string
	CameraModel  string
}

func ReadEXIF(path string) (PhotoSettings, *exiflib.Exif) {
	f, err := os.Open(path)
	if err != nil {
		return PhotoSettings{}, nil
	}
	defer f.Close()

	x, err := exiflib.Decode(f)
	if err != nil {
		return PhotoSettings{}, nil
	}

	s := PhotoSettings{}

	if fn, err := x.Get(exiflib.FNumber); err == nil {
		if num, denom, err := fn.Rat2(0); err == nil && denom != 0 {
			s.Aperture = fmt.Sprintf("%.1f", float64(num)/float64(denom))
			s.Aperture = strings.TrimSuffix(s.Aperture, ".0")
		}
	}

	if et, err := x.Get(exiflib.ExposureTime); err == nil {
		if num, denom, err := et.Rat2(0); err == nil && denom != 0 {
			val := float64(num) / float64(denom)
			if val >= 1 {
				if val == math.Floor(val) {
					s.ShutterSpeed = fmt.Sprintf("%.0fs", val)
				} else {
					s.ShutterSpeed = fmt.Sprintf("%.1fs", val)
				}
			} else {
				s.ShutterSpeed = fmt.Sprintf("1/%d", int(math.Round(1.0/val)))
			}
		}
	}

	if iso, err := x.Get(exiflib.ISOSpeedRatings); err == nil {
		if val, err := iso.Int(0); err == nil {
			s.ISO = strconv.Itoa(val)
		}
	}

	if fl, err := x.Get(exiflib.FocalLength); err == nil {
		if num, denom, err := fl.Rat2(0); err == nil && denom != 0 {
			fval := float64(num) / float64(denom)
			if fval == math.Floor(fval) {
				s.FocalLength = fmt.Sprintf("%.0fmm", fval)
			} else {
				s.FocalLength = fmt.Sprintf("%.1fmm", fval)
			}
		}
	}

	if model, err := x.Get(exiflib.Model); err == nil {
		if val, err := model.StringVal(); err == nil {
			s.CameraModel = strings.TrimSpace(val)
		}
	}

	return s, x
}

// ── Image decoding ────────────────────────────────────────────────────────────

func DecodeImage(path string) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".arw" {
		return decodeARW(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func decodeARW(path string) (image.Image, error) {
	// 1. Try to extract the largest JPEG embedded in the file (most Sony ARW contain full-res preview)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if img, err := extractLargestJPEG(data); err == nil {
		return img, nil
	}

	// 2. Try dcraw piped to stdout (PPM)
	if _, err := exec.LookPath("dcraw"); err == nil {
		cmd := exec.Command("dcraw", "-c", "-w", "-q", "3", path)
		out, err := cmd.Output()
		if err == nil {
			if img, err := decodePPM(bytes.NewReader(out)); err == nil {
				return img, nil
			}
		}
	}

	// 3. Try ImageMagick (convert or magick)
	for _, tool := range []string{"convert", "magick"} {
		if _, err := exec.LookPath(tool); err == nil {
			cmd := exec.Command(tool, path, "ppm:-")
			out, err := cmd.Output()
			if err == nil {
				if img, err := decodePPM(bytes.NewReader(out)); err == nil {
					return img, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("cannot decode ARW file %q — install dcraw or ImageMagick", filepath.Base(path))
}

// extractLargestJPEG scans raw bytes for the largest embedded JPEG.
func extractLargestJPEG(data []byte) (image.Image, error) {
	type span struct{ start, end int }
	var spans []span

	for i := 0; i < len(data)-3; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD8 && data[i+2] == 0xFF {
			for j := i + 4; j < len(data)-1; j++ {
				if data[j] == 0xFF && data[j+1] == 0xD9 {
					spans = append(spans, span{i, j + 2})
					break
				}
			}
		}
	}
	if len(spans) == 0 {
		return nil, fmt.Errorf("no JPEG found")
	}
	best := spans[0]
	for _, s := range spans[1:] {
		if s.end-s.start > best.end-best.start {
			best = s
		}
	}
	return jpeg.Decode(bytes.NewReader(data[best.start:best.end]))
}

// decodePPM decodes a binary PPM (P6) stream.
func decodePPM(r io.Reader) (image.Image, error) {
	br := bufio.NewReader(r)
	readToken := func() (string, error) {
		for {
			line, err := br.ReadString('\n')
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				if err != nil {
					return "", err
				}
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0], nil
			}
			if err != nil {
				return "", err
			}
		}
	}

	magic, err := readToken()
	if err != nil || magic != "P6" {
		return nil, fmt.Errorf("not a P6 PPM")
	}
	ws, _ := readToken()
	hs, _ := readToken()
	ms, _ := readToken()
	w, _ := strconv.Atoi(ws)
	h, _ := strconv.Atoi(hs)
	maxval, _ := strconv.Atoi(ms)
	if w == 0 || h == 0 || maxval == 0 {
		return nil, fmt.Errorf("invalid PPM header")
	}

	// After the header there is exactly one whitespace byte before pixel data
	buf := make([]byte, 1)
	br.Read(buf)

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	pix := make([]byte, w*h*3)
	if _, err := io.ReadFull(br, pix); err != nil {
		return nil, err
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 3
			img.SetNRGBA(x, y, color.NRGBA{R: pix[i], G: pix[i+1], B: pix[i+2], A: 255})
		}
	}
	return img, nil
}

// ── Orientation ───────────────────────────────────────────────────────────────

func ApplyOrientation(img image.Image, x *exiflib.Exif) image.Image {
	if x == nil {
		return img
	}
	tag, err := x.Get(exiflib.Orientation)
	if err != nil {
		return img
	}
	val, err := tag.Int(0)
	if err != nil {
		return img
	}
	switch val {
	case 3:
		return rotateImage(img, 180)
	case 6:
		return rotateImage(img, 90)
	case 8:
		return rotateImage(img, 270)
	}
	return img
}

func rotateImage(img image.Image, deg int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	switch deg {
	case 90:
		dst := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(h-1-y, x, img.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 180:
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(w-1-x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	case 270:
		dst := image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
			}
		}
		return dst
	}
	return img
}

// ── Resize ───────────────────────────────────────────────────────────────────

func ResizeForWhatsApp(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= MaxWhatsAppDim && h <= MaxWhatsAppDim {
		return img
	}
	var nw, nh int
	if w >= h {
		nw = MaxWhatsAppDim
		nh = int(math.Round(float64(h) * float64(MaxWhatsAppDim) / float64(w)))
	} else {
		nh = MaxWhatsAppDim
		nw = int(math.Round(float64(w) * float64(MaxWhatsAppDim) / float64(h)))
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}

// ── Watermark ─────────────────────────────────────────────────────────────────

type wmItem struct {
	icon string
	text string
	tw   float64 // measured text width
}

func buildItems(s PhotoSettings) []wmItem {
	var items []wmItem
	if s.CameraModel != "" {
		items = append(items, wmItem{icon: "camera", text: s.CameraModel})
	}
	if s.Aperture != "" {
		items = append(items, wmItem{icon: "aperture", text: "f/" + s.Aperture})
	}
	if s.ShutterSpeed != "" {
		items = append(items, wmItem{icon: "shutter", text: s.ShutterSpeed})
	}
	if s.ISO != "" {
		items = append(items, wmItem{icon: "iso", text: "ISO " + s.ISO})
	}
	if s.FocalLength != "" {
		items = append(items, wmItem{icon: "focal", text: s.FocalLength})
	}
	return items
}

func AddWatermark(img image.Image, s PhotoSettings, opacity float64) (image.Image, error) {
	items := buildItems(s)
	if len(items) == 0 {
		// No EXIF data — return image as-is
		return img, nil
	}

	b := img.Bounds()
	imgW := float64(b.Dx())
	imgH := float64(b.Dy())

	// Font size proportional to image height, clamped
	fontSize := math.Round(imgH / 52.0)
	fontSize = math.Max(fontSize, 16)
	fontSize = math.Min(fontSize, 56)

	margin := math.Max(imgH/55.0, 14)
	maxBoxW := imgW - 2*margin

	// Shrink font until all items fit within the image width
	var face font.Face
	var iconSize, padV, padH, textIconGap, itemGap, boxW, boxH float64
	const sepW = 1.0
	dcMeasure := gg.NewContext(1, 1)

	for {
		var err error
		face, err = loadFont(fontSize)
		if err != nil {
			return nil, err
		}

		iconSize = fontSize * 1.25
		padV = fontSize * 0.65
		padH = fontSize * 0.90
		textIconGap = fontSize * 0.32
		itemGap = fontSize * 0.85

		dcMeasure.SetFontFace(face)
		for i := range items {
			items[i].tw, _ = dcMeasure.MeasureString(items[i].text)
		}

		boxW = padH * 2
		for i, it := range items {
			boxW += iconSize + textIconGap + it.tw
			if i < len(items)-1 {
				boxW += itemGap + sepW
			}
		}
		boxH = iconSize + padV*2

		if boxW <= maxBoxW || fontSize <= 10 {
			break
		}
		fontSize--
	}

	// Watermark top-left origin
	ox := (imgW - boxW) / 2
	oy := imgH - boxH - margin

	// Draw on a separate layer at 82% opacity
	layer := gg.NewContext(int(imgW), int(imgH))
	layer.SetFontFace(face)

	// Background pill
	layer.SetRGBA(0.06, 0.06, 0.08, opacity)
	layer.DrawRoundedRectangle(ox, oy, boxW, boxH, boxH/2)
	layer.Fill()

	// Items
	cx := ox + padH
	cy := oy + boxH/2.0

	for i, it := range items {
		drawIcon(layer, it.icon, cx+iconSize/2, cy, iconSize)
		cx += iconSize + textIconGap

		layer.SetRGBA(1, 1, 1, 0.95)
		layer.DrawStringAnchored(it.text, cx+it.tw/2, cy+fontSize*0.04, 0.5, 0.38)
		cx += it.tw

		if i < len(items)-1 {
			cx += itemGap * 0.45
			layer.SetRGBA(1, 1, 1, 0.22)
			layer.SetLineWidth(sepW)
			layer.DrawLine(cx, oy+boxH*0.2, cx, oy+boxH*0.8)
			layer.Stroke()
			cx += sepW + itemGap*0.55
		}
	}

	// Composite layer over original image using 1.0 alpha (opacity is in RGBA above)
	dc := gg.NewContextForImage(img)
	dc.DrawImage(layer.Image(), 0, 0)
	return dc.Image(), nil
}

// ── Icons ─────────────────────────────────────────────────────────────────────

func drawIcon(dc *gg.Context, kind string, cx, cy, size float64) {
	r := size * 0.42
	lw := size * 0.085
	dc.SetLineWidth(lw)

	switch kind {

	case "aperture":
		// Outer circle
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.DrawCircle(cx, cy, r)
		dc.Stroke()
		// 6 aperture blades: lines from inner radius to outer, rotated
		dc.SetLineWidth(lw * 0.85)
		for i := 0; i < 6; i++ {
			a1 := float64(i)*math.Pi/3.0 - math.Pi/6
			a2 := a1 + math.Pi/3.0
			x1 := cx + r*0.30*math.Cos(a1)
			y1 := cy + r*0.30*math.Sin(a1)
			x2 := cx + r*0.88*math.Cos(a2)
			y2 := cy + r*0.88*math.Sin(a2)
			dc.DrawLine(x1, y1, x2, y2)
			dc.Stroke()
		}
		// Centre dot
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.DrawCircle(cx, cy, r*0.18)
		dc.Fill()

	case "shutter":
		// Outer circle
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.DrawCircle(cx, cy, r)
		dc.Stroke()
		// 4 tick marks
		dc.SetLineWidth(lw * 0.75)
		for i := 0; i < 4; i++ {
			a := float64(i)*math.Pi/2.0 - math.Pi/2
			dc.DrawLine(
				cx+r*0.72*math.Cos(a), cy+r*0.72*math.Sin(a),
				cx+r*0.95*math.Cos(a), cy+r*0.95*math.Sin(a),
			)
			dc.Stroke()
		}
		// Hour hand (~10 o'clock)
		dc.SetLineWidth(lw)
		dc.DrawLine(cx, cy, cx+r*0.42*math.Cos(-2.25), cy+r*0.42*math.Sin(-2.25))
		dc.Stroke()
		// Minute hand (12 o'clock)
		dc.DrawLine(cx, cy, cx, cy-r*0.72)
		dc.Stroke()
		// Centre dot
		dc.DrawCircle(cx, cy, lw*0.7)
		dc.Fill()

	case "iso":
		// Chip rectangle
		rw := r * 1.45
		rh := r * 1.20
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.DrawRoundedRectangle(cx-rw/2, cy-rh/2, rw, rh, lw*1.2)
		dc.Stroke()
		// Inner 2×2 grid
		dc.SetLineWidth(lw * 0.65)
		dc.DrawLine(cx, cy-rh/2+lw*1.8, cx, cy+rh/2-lw*1.8)
		dc.Stroke()
		dc.DrawLine(cx-rw/2+lw*1.8, cy, cx+rw/2-lw*1.8, cy)
		dc.Stroke()
		// Tiny connection pins on left/right
		pinLen := rw * 0.18
		for _, py := range []float64{cy - rh*0.22, cy + rh*0.22} {
			dc.DrawLine(cx-rw/2-pinLen, py, cx-rw/2, py)
			dc.Stroke()
			dc.DrawLine(cx+rw/2, py, cx+rw/2+pinLen, py)
			dc.Stroke()
		}

	case "focal":
		// Outer lens barrel (ellipse slightly wider)
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.DrawEllipse(cx, cy, r*0.95, r)
		dc.Stroke()
		// Inner element circle
		dc.DrawCircle(cx, cy, r*0.52)
		dc.Stroke()
		// Innermost element
		dc.DrawCircle(cx, cy, r*0.22)
		dc.Stroke()
		// Small lens glare
		dc.SetRGBA(1, 1, 1, 0.55)
		dc.DrawEllipse(cx-r*0.22, cy-r*0.30, r*0.13, r*0.07)
		dc.Fill()

	case "camera":
		dc.SetRGBA(1, 1, 1, 0.92)
		dc.SetLineWidth(lw)

		// Camera body — main rectangle
		bw := r * 2.0
		bh := r * 1.30
		bx := cx - bw/2
		by := cy - bh/2 + r*0.10
		dc.DrawRoundedRectangle(bx, by, bw, bh, lw*1.4)
		dc.Stroke()

		// Viewfinder bump on top-left
		vw := bw * 0.38
		vh := bh * 0.30
		dc.DrawRoundedRectangle(bx+bw*0.12, by-vh+lw*0.5, vw, vh, lw*0.8)
		dc.Stroke()

		// Lens — outer ring
		dc.DrawCircle(cx+bw*0.05, cy+r*0.12, r*0.50)
		dc.Stroke()
		// Lens — inner ring
		dc.DrawCircle(cx+bw*0.05, cy+r*0.12, r*0.28)
		dc.Stroke()

		// Shutter button — small circle top-right
		dc.SetRGBA(1, 1, 1, 0.70)
		dc.DrawCircle(bx+bw*0.78, by-vh*0.10, lw*1.1)
		dc.Fill()
	}
}

// ── Font ─────────────────────────────────────────────────────────────────────

func loadFont(size float64) (font.Face, error) {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("parsing embedded font: %w", err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("creating font face: %w", err)
	}
	return face, nil
}

// ── File helpers ──────────────────────────────────────────────────────────────

func IsSupportedExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".arw" || ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func CollectFiles(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", arg, err)
		}
		if info.IsDir() {
			err := filepath.Walk(arg, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() && IsSupportedExt(p) {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if !IsSupportedExt(arg) {
				fmt.Fprintf(os.Stderr, "skip: %s (unsupported format)\n", arg)
				continue
			}
			files = append(files, arg)
		}
	}
	return files, nil
}

func OutputPath(inputPath, outDir string) string {
	base := filepath.Base(inputPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(outDir, stem+".jpg")
}

// ── Process ───────────────────────────────────────────────────────────────────

func ProcessFile(inputPath, outDir string, opacity float64) error {
	settings, exifData := ReadEXIF(inputPath)

	img, err := DecodeImage(inputPath)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	img = ApplyOrientation(img, exifData)
	img = ResizeForWhatsApp(img)

	img, err = AddWatermark(img, settings, opacity)
	if err != nil {
		return fmt.Errorf("watermark: %w", err)
	}

	out := OutputPath(inputPath, outDir)
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return fmt.Errorf("encode JPEG: %w", err)
	}
	return nil
}
