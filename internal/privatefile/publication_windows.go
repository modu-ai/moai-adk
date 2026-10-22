package privatefile

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func openPublicationAnchorNative(path string) (*os.File, error) {
	return openExisting(path, windows.GENERIC_READ|windows.GENERIC_WRITE|windows.DELETE|windows.READ_CONTROL, false)
}

func publicationInfoNative(path string) (os.FileInfo, error) {
	// Capture the ID directly: Lstat defers loading it through an exclusive
	// pathname open, which conflicts with the retained RW/DELETE cleanup handle.
	f, err := openExisting(path, windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES, false)
	if err != nil {
		return nil, err
	}
	info, statErr := f.Stat()
	return info, errors.Join(statErr, f.Close())
}

func revokePublicationNative(f *os.File) error {
	// Clear bytes first, including any aliases of this object. Deletion is
	// requested on the captured handle, never on a replaceable pathname.
	truncateErr, syncErr := f.Truncate(0), f.Sync()
	remove := byte(1) // FILE_DISPOSITION_INFO.DeleteFile (BOOLEAN).
	deleteErr := windows.SetFileInformationByHandle(windows.Handle(f.Fd()), windows.FileDispositionInfo, &remove, 1)
	return errors.Join(truncateErr, syncErr, deleteErr)
}
