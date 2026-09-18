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
	if err != nil && err != windows.ERROR_ALREADY_EXISTS {
		return true // unknown error, let it run
	}

	// CreateMutex returns a valid handle to the *existing* mutex when it
	// already exists. Older x/sys/windows wrappers only set err when
	// handle == 0, so the err check above can miss the already-exists case.
	// GetLastError reliably reports ERROR_ALREADY_EXISTS (183) in that case.
	alreadyExists := err == windows.ERROR_ALREADY_EXISTS ||
		windows.GetLastError() == uint32(windows.ERROR_ALREADY_EXISTS)

	if alreadyExists {
		windows.CloseHandle(handle)
		return false
	}

	// First instance: keep the handle open for the process lifetime.
	// (Intentional — the OS closes it on process exit.)
	return true
}
