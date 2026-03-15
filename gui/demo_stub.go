//go:build !demo

package main

// demoWorkers is 0 in production builds (= auto worker count).
var demoWorkers int

// demoHook is a no-op in production builds.
func demoHook(_ func([]string)) {}
