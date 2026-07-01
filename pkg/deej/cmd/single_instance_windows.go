//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

func ensureSingleInstance() bool {
	name, err := windows.UTF16PtrFromString("Global\\deej-winapp")
	if err != nil {
		return true // can't create name, let it run
	}

	handle, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		if err == windows.ERROR_ALREADY_EXISTS {
			if handle != 0 {
				windows.CloseHandle(handle)
			}
			return false
		}
		return true // unknown error, let it run
	}

	// Mutex created successfully (first instance).
	// Handle stays open — Windows auto-releases it when the process exits.
	return true
}
