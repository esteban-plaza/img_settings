package main

import (
	_ "embed"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

//go:embed assets/AppIcon.png
var iconBytes []byte

func main() {
	a := app.New()
	a.Settings().SetTheme(appleTheme{})

	icon := fyne.NewStaticResource("AppIcon.png", iconBytes)
	a.SetIcon(icon)

	w := a.NewWindow("img_settings")
	w.Resize(fyne.NewSize(720, 520))
	w.SetMaster()

	dz := newDropZone()
	tb := newSettingsBar(w)

	// mainStack holds the current "main" view (drop zone or results).
	mainStack := container.NewStack(dz)

	// root layout: main area with toolbar pinned at bottom.
	root := container.NewBorder(nil, tb.container, nil, nil, mainStack)
	w.SetContent(root)

	// showDropZone resets the UI back to the drop zone.
	showDropZone := func() {
		dz.SetState(dzIdle)
		mainStack.Objects = []fyne.CanvasObject{dz}
		mainStack.Refresh()
	}

	// onDrop is the shared handler for both window-level drops and folder picker.
	onDrop := func(paths []string) {
		if len(paths) == 0 {
			return
		}

		// Derive output base from first path.
		base := paths[0]
		if !isDir(base) {
			base = filepath.Dir(base)
		}
		tb.setDropBase(base)

		// Collect files (applying recurse option).
		opts := tb.options()
		if opts.OutDir == "" {
			opts.OutDir = filepath.Join(base, "out")
		}

		files := collectFiles(paths, opts.Recurse)
		if len(files) == 0 {
			return
		}

		// Switch to processing state.
		dz.SetState(dzProcessing)

		total := len(files)
		done := 0
		rv := newResultView()

		// Swap main view to results.
		mainStack.Objects = []fyne.CanvasObject{rv.container}
		mainStack.Refresh()

		go func() {
			processFiles(files, opts, func(r jobResult) {
				fyne.Do(func() {
					done++
					rv.setProgress(float64(done) / float64(total))
					rv.appendResult(r)

					if done == total {
						rv.doneActions(w, opts.OutDir, showDropZone)
						mainStack.Objects = []fyne.CanvasObject{rv.container}
						mainStack.Refresh()
					}
				})
			})
		}()
	}

	// Tapping the drop zone opens a folder picker as fallback.
	dz.OnTapped = func() {
		if dzState(dz.state.Load()) == dzProcessing {
			return
		}
		showFolderPicker(w, onDrop)
	}

	// Window-level drop: the only way to receive file drops in Fyne v2.
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		var paths []string
		for _, u := range uris {
			paths = append(paths, uriToPath(u))
		}
		onDrop(paths)
	})

	w.ShowAndRun()
}

