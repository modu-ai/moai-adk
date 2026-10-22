//go:build !windows

package privatefile

import (
	"errors"
	"fmt"
	"os"
)

func openPublicationAnchorNative(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, 0)
}

func publicationInfoNative(path string) (os.FileInfo, error) { return os.Lstat(path) }

func revokePublicationNative(f *os.File) error {
	// Unix has no portable conditional unlink by file descriptor. Clearing the
	// captured inode preserves a concurrent writer's replacement at its name.
	return errors.Join(f.Truncate(0), f.Sync(), fmt.Errorf("private file: cleanup kept the captured inode; pathname was not removed"))
}
