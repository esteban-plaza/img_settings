//go:build demo

package main

import (
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
)

// demoWorkers holds the worker count override parsed from --demo-workers.
var demoWorkers int

// demoHook reads flags from os.Args:
//
//	--demo <folder>          auto-trigger processing on startup
//	--demo-workers <n>       override worker count (default: 1 when demo active)
//
// Only compiled when built with -tags demo.
func demoHook(onDrop func([]string)) {
	args := os.Args[1:]
	var demoPath string
	workers := 1 // default to single-threaded for visible progress bar

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--demo":
			if i+1 < len(args) {
				demoPath = args[i+1]
				i++
			}
		case "--demo-workers":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					workers = n
				}
				i++
			}
		}
	}

	if demoPath == "" {
		return
	}

	demoWorkers = workers
	go func() {
		time.Sleep(2000 * time.Millisecond)
		fyne.Do(func() { onDrop([]string{demoPath}) })
	}()
}
