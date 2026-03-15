package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	watermark "github.com/esteban-plaza/img_settings/lib"
)

// Options holds the settings from the toolbar.
type Options struct {
	Recurse bool
	Opacity float64
	OutDir  string // empty = derive from first dropped path
}

// jobResult is the result of processing a single file.
type jobResult struct {
	Name string
	Size string // e.g. "2.3 MB"
	Err  error
}

// processFiles runs the watermark pipeline concurrently over files,
// calling onResult (on a background goroutine) after each file finishes.
// It blocks until all files are processed.
func processFiles(files []string, opts Options, onResult func(jobResult)) {
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		onResult(jobResult{Name: "setup", Err: fmt.Errorf("create output dir: %w", err)})
		return
	}

	workers := runtime.NumCPU() * 2
	jobs := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				err := watermark.ProcessFile(f, opts.OutDir, opts.Opacity)
				r := jobResult{Name: filepath.Base(f)}
				if err != nil {
					r.Err = err
				} else {
					out := watermark.OutputPath(f, opts.OutDir)
					r.Size = humanSize(out)
				}
				onResult(r)
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
}
