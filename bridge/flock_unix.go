//go:build unix

package main

import "syscall"

// lockFileDescriptor locks the file descriptor using Unix flock
func lockFileDescriptor(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_EX)
}

// unlockFileDescriptor unlocks the file descriptor using Unix flock
func unlockFileDescriptor(fd int) error {
	return syscall.Flock(fd, syscall.LOCK_UN)
}
