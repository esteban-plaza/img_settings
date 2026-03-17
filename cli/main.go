package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	watermark "github.com/esteban-plaza/img-settings/lib"
)

// version is injected at build time via -ldflags "-X main.version=<tag>".
var version string

func main() {
	outDir := flag.String("out", "out", "output directory (relative to cwd)")
	opacityFlag := flag.Float64("opacity", 0.82, "watermark opacity (0.0 = invisible, 1.0 = fully opaque)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `img_settings — add camera-settings watermark to photos

Usage:
  img_settings [flags] <file|folder> [file|folder ...]

Flags:
  -out <dir>       output folder (default: "out")
  -opacity <0-1>   watermark opacity (default: 0.82)

Supported input formats: ARW, JPG, JPEG, PNG
Output format: always JPG (WhatsApp HD quality, max 2560px)

For ARW files, install dcraw or ImageMagick if the embedded JPEG
preview is not sufficient.

`)
		flag.PrintDefaults()
	}
	flag.Parse()

	opacity := *opacityFlag
	if opacity < 0 {
		opacity = 0
	} else if opacity > 1 {
		opacity = 1
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output dir %q: %v\n", *outDir, err)
		os.Exit(1)
	}

	files, err := watermark.CollectFiles(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("no supported files found (ARW/JPG/PNG)")
		os.Exit(0)
	}

	workers := runtime.NumCPU() * 2
	fmt.Printf("processing %d file(s) → %s/  [%d workers]\n\n", len(files), *outDir, workers)

	type result struct {
		name string
		size string
		err  error
	}

	jobs := make(chan string, len(files))
	results := make(chan result, len(files))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				err := watermark.ProcessFile(f, *outDir, opacity)
				r := result{name: filepath.Base(f)}
				if err != nil {
					r.err = err
				} else {
					out := watermark.OutputPath(f, *outDir)
					if fi, err := os.Stat(out); err == nil {
						r.size = fmt.Sprintf("%.1f MB", float64(fi.Size())/1e6)
					}
				}
				results <- r
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	ok, fail := 0, 0
	for r := range results {
		fmt.Printf("  %-40s  ", r.name)
		if r.err != nil {
			fmt.Printf("FAIL: %v\n", r.err)
			fail++
		} else {
			fmt.Printf("OK  (%s)\n", r.size)
			ok++
		}
	}

	fmt.Printf("\ndone: %d ok, %d failed\n", ok, fail)
	if fail > 0 {
		os.Exit(1)
	}
}
