// +build darwin

package util

import "fmt"

func getCurrentWindowProcessNames() ([]string, error) {
	return nil, fmt.Errorf("GetCurrentWindowProcessNames is not implemented on macOS")
}
