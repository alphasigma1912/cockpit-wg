//go:build windows

package main

import "fmt"

// lockFileDescriptor is a no-op on Windows (not supported)
func lockFileDescriptor(_ int) error {
	return fmt.Errorf("file locking not supported on Windows")
}

// unlockFileDescriptor is a no-op on Windows (not supported)
func unlockFileDescriptor(_ int) error {
	return fmt.Errorf("file locking not supported on Windows")
}
