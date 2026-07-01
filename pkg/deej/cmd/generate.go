//go:generate rsrc -manifest ../assets/deej.manifest -ico ../assets/logo.ico -arch amd64 -o rsrc_windows_amd64.syso

// Package main provides the deej Windows desktop client.
// Run `go generate` in this directory to regenerate the .syso resource file.
package main
