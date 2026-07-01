// +build darwin

// This file provides stubs so the package compiles on macOS for development.
// Windows audio session management is in session_finder_windows.go.

package deej

import (
	"fmt"

	"go.uber.org/zap"
)

type sessionFinder struct {
	logger *zap.SugaredLogger
}

func newSessionFinder(logger *zap.SugaredLogger) (*sessionFinder, error) {
	logger = logger.Named("session-finder")
	logger.Warn("Session finder not implemented on this platform")

	return &sessionFinder{logger: logger}, nil
}

func (sf *sessionFinder) GetAllSessions() ([]Session, error) {
	return nil, fmt.Errorf("GetAllSessions not implemented on this platform")
}

func (sf *sessionFinder) Release() error {
	return nil
}
