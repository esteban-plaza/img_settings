package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// showFolderPicker opens the Fyne folder picker dialog and calls cb with
// the selected paths (a single-element slice containing the chosen folder).
func showFolderPicker(w fyne.Window, cb func([]string)) {
	dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
		if err != nil || lu == nil {
			return
		}
		cb([]string{uriToPath(lu)})
	}, w)
}
