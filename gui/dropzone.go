package main

import (
	"image"
	"image/color"
	"math"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/fogleman/gg"
)

// dzState encodes the visual state of the drop zone.
type dzState int32

const (
	dzIdle       dzState = 0
	dzHover      dzState = 1
	dzProcessing dzState = 2
)

// DropZone is a custom widget: a dashed-border rectangular area that
// accepts file drops and shows contextual visual feedback.
type DropZone struct {
	widget.BaseWidget

	state atomic.Int32

	border *canvas.Raster // draws the dashed border + tinted bg
	label  *canvas.Text
	sub    *canvas.Text

	// OnTapped is called when the user clicks the drop zone.
	OnTapped func()
}

func newDropZone() *DropZone {
	dz := &DropZone{}

	dz.border = canvas.NewRaster(dz.drawBorder)

	dz.label = canvas.NewText("Drop photos or folders here", color.Black)
	dz.label.Alignment = fyne.TextAlignCenter
	dz.label.TextSize = 15

	dz.sub = canvas.NewText("or click to browse", color.Black)
	dz.sub.Alignment = fyne.TextAlignCenter
	dz.sub.TextSize = 12

	dz.ExtendBaseWidget(dz)
	return dz
}

// Tapped implements fyne.Tappable.
func (dz *DropZone) Tapped(_ *fyne.PointEvent) {
	if dz.OnTapped != nil {
		dz.OnTapped()
	}
}

// SetState updates the visual state and triggers a repaint.
func (dz *DropZone) SetState(s dzState) {
	dz.state.Store(int32(s))
	dz.border.Refresh()
	dz.updateLabels()
}

func (dz *DropZone) updateLabels() {
	switch dzState(dz.state.Load()) {
	case dzProcessing:
		dz.label.Text = "Processing…"
		dz.label.Color = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xb0}
		dz.sub.Text = ""
	case dzHover:
		dz.label.Text = "Release to process"
		dz.label.Color = color.NRGBA{R: 0x00, G: 0x71, B: 0xe3, A: 0xff}
		dz.sub.Text = ""
	default:
		dz.label.Text = "Drop photos or folders here"
		dz.label.Color = color.NRGBA{R: 0x3A, G: 0x3A, B: 0x3C, A: 0xff}
		dz.sub.Text = "or click to browse"
		dz.sub.Color = color.NRGBA{R: 0x8E, G: 0x8E, B: 0x93, A: 0xff}
	}
	dz.label.Refresh()
	dz.sub.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (dz *DropZone) CreateRenderer() fyne.WidgetRenderer {
	dz.updateLabels()

	// Stack: raster background behind the centered label vbox.
	content := container.NewStack(
		dz.border,
		container.NewCenter(
			container.NewVBox(
				newCloudIcon(dz),
				dz.label,
				dz.sub,
			),
		),
	)
	return widget.NewSimpleRenderer(content)
}

// drawBorder is the raster draw function — called every repaint.
func (dz *DropZone) drawBorder(w, h int) image.Image {
	dc := gg.NewContext(w, h)
	state := dzState(dz.state.Load())

	pad := 18.0
	fw := float64(w) - pad*2
	fh := float64(h) - pad*2
	corner := math.Min(fw, fh) * 0.04

	// Subtle tinted background
	switch state {
	case dzHover:
		dc.SetRGBA(0.0, 0.44, 0.89, 0.07)
		dc.DrawRoundedRectangle(pad, pad, fw, fh, corner)
		dc.Fill()
	}

	// Dashed border
	switch state {
	case dzHover:
		dc.SetRGBA(0.0, 0.44, 0.89, 0.80)
	case dzProcessing:
		dc.SetRGBA(0.55, 0.55, 0.60, 0.35)
	default:
		dc.SetRGBA(0.55, 0.55, 0.60, 0.45)
	}
	dc.SetLineWidth(1.5)
	dc.SetDash(9, 6)
	dc.DrawRoundedRectangle(pad, pad, fw, fh, corner)
	dc.Stroke()

	return dc.Image()
}

// ── Photo icon widget ─────────────────────────────────────────────────────────

// photoIcon draws a minimal image-frame icon (rounded rect + mountain + sun).
type photoIcon struct {
	widget.BaseWidget
	dz     *DropZone
	raster *canvas.Raster
}

func newCloudIcon(dz *DropZone) *photoIcon {
	ci := &photoIcon{dz: dz}
	ci.raster = canvas.NewRaster(ci.draw)
	ci.ExtendBaseWidget(ci)
	return ci
}

func (ci *photoIcon) MinSize() fyne.Size { return fyne.NewSize(72, 58) }

func (ci *photoIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ci.raster)
}

func (ci *photoIcon) draw(w, h int) image.Image {
	dc := gg.NewContext(w, h)
	state := dzState(ci.dz.state.Load())

	var sr, sg, sb, sa float64
	switch state {
	case dzHover:
		sr, sg, sb, sa = 0.0, 0.44, 0.89, 1.0
	default:
		sr, sg, sb, sa = 0.56, 0.56, 0.58, 1.0
	}

	fw, fh := float64(w), float64(h)
	lw := fw * 0.048

	// Photo frame: landscape rect centred in the icon area
	fx := fw * 0.06
	fy := fh * 0.10
	frW := fw * 0.88
	frH := fh * 0.80
	rc := fh * 0.10

	// very light fill
	dc.SetRGBA(sr, sg, sb, 0.08)
	dc.DrawRoundedRectangle(fx, fy, frW, frH, rc)
	dc.Fill()

	// frame border
	dc.SetLineWidth(lw)
	dc.SetRGBA(sr, sg, sb, sa)
	dc.DrawRoundedRectangle(fx, fy, frW, frH, rc)
	dc.Stroke()

	// Sun — small filled circle, upper-left inside frame
	sunR := fh * 0.09
	dc.DrawCircle(fx+frW*0.22, fy+frH*0.30, sunR)
	dc.Fill()

	// Mountain — single peak, stroke only
	// clip to inside the frame so lines don't overflow
	dc.DrawRoundedRectangle(fx+lw, fy+lw, frW-lw*2, frH-lw*2, rc)
	dc.Clip()

	mountainBaseY := fy + frH - lw*0.5
	dc.MoveTo(fx, mountainBaseY)
	dc.LineTo(fx+frW*0.38, fy+frH*0.42)
	dc.LineTo(fx+frW*0.60, fy+frH*0.62)
	dc.LineTo(fx+frW*0.80, fy+frH*0.38)
	dc.LineTo(fx+frW, mountainBaseY)
	dc.SetLineWidth(lw)
	dc.SetRGBA(sr, sg, sb, sa*0.50)
	dc.Stroke()

	return dc.Image()
}
