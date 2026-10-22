package privatefile

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

func secureExistingNative(path string, before os.FileInfo, directory bool) (retErr error) {
	// O_PATH can hold an owner object even when its current mode denies reads.
	// NOFOLLOW keeps a replacement symlink from selecting its target.
	fd, err := unix.Open(path, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return &os.PathError{Op: "open private object", Path: path, Err: err}
	}
	f := os.NewFile(uintptr(fd), path)
	defer func() { retErr = errors.Join(retErr, f.Close()) }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, info) || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("private object changed after inspection: %s", path)
	}
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	err = unix.Fchmodat(fd, "", uint32(mode), unix.AT_EMPTY_PATH)
	if errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EOPNOTSUPP) {
		// Older Linux kernels lack fchmodat2. A procfs descriptor link still
		// names this held object, even if its original pathname is replaced.
		return chmodProcHandle(f, mode)
	}
	return err
}

func chmodProcHandle(f *os.File, mode os.FileMode) error {
	var fs unix.Statfs_t
	if err := unix.Statfs("/proc/self/fd", &fs); err != nil {
		return err
	}
	if fs.Type != unix.PROC_SUPER_MAGIC {
		return fmt.Errorf("private permission repair requires a procfs descriptor namespace")
	}
	return os.Chmod("/proc/self/fd/"+strconv.FormatUint(uint64(f.Fd()), 10), mode)
}
