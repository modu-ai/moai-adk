package atomicfile

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func renameNoReplace(oldpath, newpath string) error {
	old, err := noReplaceWindowsPath(oldpath)
	if err != nil {
		return err
	}
	next, err := noReplaceWindowsPath(newpath)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(old, next, 0)
}

func noReplaceWindowsPath(path string) (*uint16, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	path = filepath.Clean(path)
	if !strings.HasPrefix(path, `\\?\`) && !strings.HasPrefix(path, `\\.\`) {
		if strings.HasPrefix(path, `\\`) {
			path = `\\?\UNC\` + strings.TrimPrefix(path, `\\`)
		} else {
			path = `\\?\` + path
		}
	}
	return windows.UTF16PtrFromString(path)
}
