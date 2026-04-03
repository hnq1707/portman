//go:build windows
// +build windows

package utils

import (
	"golang.org/x/sys/windows"
)

// IsAdmin checks if the current process has Administrator privileges on Windows.
func IsAdmin() bool {
	var sid *windows.SID

	// Although the documentation says this function is deprecated, it's still widely used
	// and reliable for this specific purpose.
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return member
}

// CanKill checks if the current process likely has permission to kill the given PID.
func CanKill(pid int) bool {
	// On Windows, if we are Admin, we can kill almost anything.
	// If not, we can only kill processes we own.
	if IsAdmin() {
		return true
	}
	
	// A simple check: try to open the process with PROCESS_TERMINATE access.
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return false
	}
	windows.CloseHandle(handle)
	return true
}

func IsPrivilegedPort(port int) bool {
	// Windows doesn't strictly have privileged ports < 1024 like Unix,
	// but some software might treat them as such.
	return false
}
