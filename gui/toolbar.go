package main

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// settingsBar is the floating bottom toolbar.
type settingsBar struct {
	container fyne.CanvasObject

	recurse    bool
	opacity    float64
	outDir     string // empty = auto-derive at process time
	outDirBase string // the base derived from the last drop

	opacityLabel *canvas.Text
	outDirLabel  *canvas.Text
	window       fyne.Window
}

func newSettingsBar(w fyne.Window) *settingsBar {
	tb := &settingsBar{
		opacity: 0.82,
		window:  w,
	}

	// ── Subfolders ────────────────────────────────────────────────────────────
	subCheck := widget.NewCheck("Subfolders", func(v bool) {
		tb.recurse = v
	})
	subCheck.Checked = false

	// ── Opacity ───────────────────────────────────────────────────────────────
	tb.opacityLabel = canvas.NewText(fmt.Sprintf("%.0f%%", tb.opacity*100), nil)
	tb.opacityLabel.TextSize = 12

	opacitySlider := widget.NewSlider(0, 1)
	opacitySlider.Value = tb.opacity
	opacitySlider.Step = 0.01
	opacitySlider.OnChanged = func(v float64) {
		tb.opacity = v
		tb.opacityLabel.Text = fmt.Sprintf("%.0f%%", v*100)
		tb.opacityLabel.Refresh()
	}
	opacitySlider.SetValue(tb.opacity)

	opacitySection := container.NewHBox(
		labelText("Opacity", 12),
		container.NewGridWrap(fyne.NewSize(110, 28), opacitySlider),
		tb.opacityLabel,
	)

	// ── Dest folder ───────────────────────────────────────────────────────────
	tb.outDirLabel = canvas.NewText("auto (./out)", nil)
	tb.outDirLabel.TextSize = 12

	folderBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil || lu == nil {
				return
			}
			tb.outDir = uriToPath(lu)
			tb.outDirLabel.Text = filepath.Base(tb.outDir)
			tb.outDirLabel.Refresh()
		}, w)
	})
	folderBtn.Importance = widget.LowImportance

	clearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		tb.outDir = ""
		if tb.outDirBase != "" {
			tb.outDirLabel.Text = "→ " + filepath.Base(tb.outDirBase)
		} else {
			tb.outDirLabel.Text = "auto (./out)"
		}
		tb.outDirLabel.Refresh()
	})
	clearBtn.Importance = widget.LowImportance

	destSection := container.NewHBox(
		labelText("Output", 12),
		folderBtn,
		tb.outDirLabel,
		clearBtn,
	)

	// ── Separator line ────────────────────────────────────────────────────────
	sep := canvas.NewLine(nil)
	sep.StrokeWidth = 0.5

	// ── Layout ────────────────────────────────────────────────────────────────
	row := container.NewHBox(
		subCheck,
		spacer(16),
		opacitySection,
		spacer(16),
		destSection,
	)

	tb.container = container.NewBorder(sep, nil, nil, nil,
		container.NewCenter(container.NewPadded(row)),
	)

	return tb
}

// options returns the current settings as an Options struct.
// outDirBase is used to derive the default output directory.
func (tb *settingsBar) options() Options {
	out := tb.outDir
	if out == "" && tb.outDirBase != "" {
		out = filepath.Join(tb.outDirBase, "out")
	}
	return Options{
		Recurse: tb.recurse,
		Opacity: tb.opacity,
		OutDir:  out,
	}
}

// setDropBase updates the derived output dir label when files are dropped.
func (tb *settingsBar) setDropBase(base string) {
	tb.outDirBase = base
	if tb.outDir == "" {
		tb.outDirLabel.Text = "→ " + filepath.Base(filepath.Join(base, "out"))
		tb.outDirLabel.Refresh()
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func labelText(s string, size float32) *canvas.Text {
	t := canvas.NewText(s, nil)
	t.TextSize = size
	return t
}

func spacer(w float32) fyne.CanvasObject {
	r := canvas.NewRectangle(nil)
	r.SetMinSize(fyne.NewSize(w, 1))
	return r
}
