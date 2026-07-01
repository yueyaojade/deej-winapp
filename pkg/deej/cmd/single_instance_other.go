//go:build !windows

package main

func ensureSingleInstance() bool {
	// Single-instance enforcement is Windows-specific (named mutex).
	// On other platforms, just proceed.
	return true
}
