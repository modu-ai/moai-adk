//go:build windows

// migrate_agency_disk_windows.go: Windows disk space verification/check with/by/todirect.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-011
package cli

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"syscall"
"unsafe"
)

// availableDiskBytes return available disk space in bytes at the specified path filesystem.
// use GetDiskFreeSpaceExW Win32 API.
func availableDiskBytes(path string) (uint64, error) {
kernel32 := syscall.MustLoadDLL("kernel32.dll")
getDiskFreeSpaceEx := kernel32.MustFindProc("GetDiskFreeSpaceExW")

pathPtr, err := syscall.UTF16PtrFromString(path)
if err != nil {
return 0, fmt.Errorf("utf16ptr %s: %w", path, err)
}

var freeBytesAvailable uint64
var totalNumberOfBytes uint64
var totalNumberOfFreeBytes uint64

ret, _, callErr := getDiskFreeSpaceEx.Call(
uintptr(unsafe.Pointer(pathPtr)),
uintptr(unsafe.Pointer(&freeBytesAvailable)),
uintptr(unsafe.Pointer(&totalNumberOfBytes)),
uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
)
if ret == 0 {
return 0, fmt.Errorf("GetDiskFreeSpaceExW: %w", callErr)
}
return freeBytesAvailable, nil
}

// dirSizeBytes return total size in bytes of all files under the specified directory.
// skip symbolic links.
//
// @MX:NOTE: [AUTO] helper for migration pre-check — used only for disk space calculation
func dirSizeBytes(root string) (uint64, error) {
var total uint64
err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
if walkErr != nil {
return walkErr
}
if d.IsDir() {
return nil
}
info, err := os.Lstat(path)
if err != nil {
return nil //nolint:nilerr
}
if info.Mode()&os.ModeSymlink != 0 {
return nil
}
//nolint:gosec
total += uint64(info.Size())
return nil
})
return total, err
}

// checkDiskSpaceFn function variable for injecting checkDiskSpace in tests.
//
// @MX:ANCHOR: [AUTO] entry point for disk space pre-check — supports test mocking
// @MX:REASON: [AUTO] called from runFull, tests (plural), future resume path; fan_in >= 3
var checkDiskSpaceFn = checkDiskSpace

// checkDiskSpace 2x the total size of files under sourcePath
// validate that sourcePath filesystem has enough available space.
// if insufficient space, return *MigrateError with ErrMigrateDiskFull code.
//
// REQ-MIGRATE-011: require at least 2x .agency/ size.
func checkDiskSpace(sourcePath string) error {
sourceSize, err := dirSizeBytes(sourcePath)
if err != nil {
return nil
}

available, err := availableDiskBytes(sourcePath)
if err != nil {
return nil
}

required := sourceSize * 2
if available < required {
return &MigrateError{
Code: ErrMigrateDiskFull,
Message: fmt.Sprintf(
"insufficient disk space: need %d bytes (source size %d × 2), available %d bytes",
required, sourceSize, available,
),
}
}
return nil
}
