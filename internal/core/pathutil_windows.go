//go:build windows

package core

import (
	"syscall"
	"unsafe"
)

// GetLongPathName converts Windows 8.3 short paths to full Unicode paths.
// This resolves paths like "C:\Users\John~1" back to "C:\Users\JohnDoe".
func GetLongPathName(shortPath string) (string, error) {
	// Convert to UTF-16
	ptr, err := syscall.UTF16PtrFromString(shortPath)
	if err != nil {
		return "", err
	}

	// GetLongPathNameW returns required buffer size
	n, _, err := procGetLongPathNameW.Call(
		uintptr(unsafe.Pointer(ptr)),
		0,
		0,
	)
	if err != nil && err != syscall.ERROR_INSUFFICIENT_BUFFER {
		return "", err
	}

	if n == 0 {
		return shortPath, nil // No conversion needed
	}

	// Allocate buffer and get actual long path
	buf := make([]uint16, n)
	n, _, err = procGetLongPathNameW.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(n),
	)
	if err != nil {
		return "", err
	}

	return syscall.UTF16ToString(buf), nil
}

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetLongPathNameW = modkernel32.NewProc("GetLongPathNameW")
)
