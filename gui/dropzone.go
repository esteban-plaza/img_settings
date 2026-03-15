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

	dz.label = canvas.NewText("Drop photos or folders here", color.White)
	dz.label.Alignment = fyne.TextAlignCenter
	dz.label.TextSize = 15

	dz.sub = canvas.NewText("or tap to choose", color.White)
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
		dz.label.Color = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xdd}
		dz.sub.Text = "or tap to choose"
		dz.sub.Color = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x60}
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

// ── Cloud icon widget ─────────────────────────────────────────────────────────

// cloudIcon is a small canvas.Raster that draws a stylised upload-cloud icon.
// It reads the parent DropZone state to choose colours.
type cloudIcon struct {
	widget.BaseWidget
	dz     *DropZone
	raster *canvas.Raster
}

func newCloudIcon(dz *DropZone) *cloudIcon {
	ci := &cloudIcon{dz: dz}
	ci.raster = canvas.NewRaster(ci.draw)
	ci.ExtendBaseWidget(ci)
	return ci
}

func (ci *cloudIcon) MinSize() fyne.Size { return fyne.NewSize(64, 52) }

func (ci *cloudIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ci.raster)
}

func (ci *cloudIcon) draw(w, h int) image.Image {
	dc := gg.NewContext(w, h)
	state := dzState(ci.dz.state.Load())

	cx, cy := float64(w)/2, float64(h)/2+2
	r := float64(h) * 0.32

	switch state {
	case dzHover:
		dc.SetRGBA(0.0, 0.44, 0.89, 0.90)
	case dzProcessing:
		dc.SetRGBA(0.55, 0.55, 0.60, 0.50)
	default:
		dc.SetRGBA(0.65, 0.65, 0.70, 0.70)
	}

	// Cloud shape
	dc.DrawEllipse(cx, cy, r*1.05, r*0.68)
	dc.DrawCircle(cx-r*0.65, cy+r*0.05, r*0.50)
	dc.DrawCircle(cx+r*0.60, cy+r*0.05, r*0.42)
	dc.Fill()

	// Arrow up — white, centered
	lw := r * 0.14
	dc.SetLineWidth(lw)
	aw := r * 0.38
	ay := cy - float64(h)*0.06
	if state == dzHover {
		dc.SetRGBA(1, 1, 1, 1.0)
	} else {
		dc.SetRGBA(1, 1, 1, 0.85)
	}
	dc.DrawLine(cx, ay, cx, ay+r*0.62)
	dc.Stroke()
	dc.MoveTo(cx-aw/2, ay+aw*0.55)
	dc.LineTo(cx, ay)
	dc.LineTo(cx+aw/2, ay+aw*0.55)
	dc.Stroke()

	return dc.Image()
}
