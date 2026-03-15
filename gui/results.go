package main

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// resultView holds references to the progress bar and the items list.
type resultView struct {
	container fyne.CanvasObject
	progress  *widget.ProgressBar
	list      *fyne.Container // VBox of result rows
	scroll    *container.Scroll
}

// newResultView creates the results container (progress bar + scrollable list).
func newResultView() *resultView {
	rv := &resultView{}

	rv.progress = widget.NewProgressBar()
	rv.progress.TextFormatter = func() string {
		return fmt.Sprintf("%.0f%%", rv.progress.Value*100)
	}

	rv.list = container.NewVBox()
	rv.scroll = container.NewVScroll(rv.list)

	rv.container = container.NewBorder(
		container.NewPadded(rv.progress),
		nil, nil, nil,
		rv.scroll,
	)
	return rv
}

// appendResult adds a result row to the list.
func (rv *resultView) appendResult(r jobResult) {
	var icon fyne.Resource
	var nameColor color.Color

	if r.Err != nil {
		icon = theme.ErrorIcon()
		nameColor = color.NRGBA{R: 0xff, G: 0x3b, B: 0x30, A: 0xff} // Apple red
	} else {
		icon = theme.ConfirmIcon()
		nameColor = nil // default foreground
	}

	nameLabel := canvas.NewText(r.Name, nameColor)
	nameLabel.TextSize = 12

	var right fyne.CanvasObject
	if r.Err != nil {
		errLabel := canvas.NewText(r.Err.Error(), color.NRGBA{R: 0xff, G: 0x3b, B: 0x30, A: 0xcc})
		errLabel.TextSize = 11
		right = errLabel
	} else {
		sizeLabel := canvas.NewText(r.Size, color.NRGBA{R: 0x8e, G: 0x8e, B: 0x93, A: 0xff})
		sizeLabel.TextSize = 11
		right = sizeLabel
	}

	row := container.NewBorder(nil, nil,
		container.NewHBox(widget.NewIcon(icon), nameLabel),
		right,
	)

	rv.list.Add(row)
	rv.list.Refresh()
	rv.scroll.ScrollToBottom()
}

// setProgress updates the progress bar (0.0–1.0).
func (rv *resultView) setProgress(v float64) {
	rv.progress.SetValue(v)
}

// doneActions appends a bottom action bar with "Reveal" and "Reset" buttons.
func (rv *resultView) doneActions(w fyne.Window, outDir string, onReset func()) {
	revealLabel := "Reveal in Finder"
	if runtime.GOOS == "windows" {
		revealLabel = "Open in Explorer"
	}
	revealBtn := widget.NewButtonWithIcon(revealLabel, theme.FolderOpenIcon(), func() {
		revealInFinder(outDir)
	})
	revealBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButtonWithIcon("Process more", theme.ContentAddIcon(), func() {
		onReset()
	})
	resetBtn.Importance = widget.LowImportance

	actions := container.NewCenter(container.NewHBox(revealBtn, resetBtn))

	// Re-wrap container to include the action bar at the bottom.
	wrapped := container.NewBorder(
		container.NewPadded(rv.progress),
		container.NewPadded(actions),
		nil, nil,
		rv.scroll,
	)
	rv.container = wrapped
}

// revealInFinder opens the given directory in the OS file browser.
func revealInFinder(dir string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", dir).Start()
	case "windows":
		exec.Command("explorer", dir).Start()
	default:
		exec.Command("xdg-open", dir).Start()
	}
}
