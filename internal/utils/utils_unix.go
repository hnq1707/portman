//go:build !windows
// +build !windows

package utils

import (
	"os"
	"syscall"
)

// IsAdmin checks if the process is running as root on Unix/Linux/macOS.
func IsAdmin() bool {
	return os.Geteuid() == 0
}

// CanKill checks if we likely have permission to kill a process.
func CanKill(pid int) bool {
	if IsAdmin() {
		return true
	}
	
	// On Unix, try sending signal 0 (check permission)
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func IsPrivilegedPort(port int) bool {
	return port < 1024
}
