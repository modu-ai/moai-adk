//go:build !windows

// migrate_agency_disk_unix.go: Unix/macOS disk space verification/check with/by/todirect.
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-011
package cli

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"syscall"
)

// availableDiskBytes return available disk space in bytes at the specified path filesystem.
// use syscall.Statfs to minimize platform dependency.
func availableDiskBytes(path string) (uint64, error) {
var stat syscall.Statfs_t
if err := syscall.Statfs(path, &stat); err != nil {
return 0, fmt.Errorf("statfs %s: %w", path, err)
}
// Bavail: number of blocks available to ordinary users (excluding reserved blocks
//nolint:gosec // convert to uint64: Bavail value is always positive
return stat.Bavail * uint64(stat.Bsize), nil
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
// exclude symbolic links from size calculation (also skipped during copying)
info, err := os.Lstat(path)
if err != nil {
return nil //nolint:nilerr // ignore if file cannot be accessed
}
if info.Mode()&os.ModeSymlink != 0 {
return nil
}
//nolint:gosec // int64 → uint64: file size is always positive
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
// size calculation failure is not critical — pass without warning
return nil
}

available, err := availableDiskBytes(sourcePath)
if err != nil {
// Statfs failure is not critical — pass without warning
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
