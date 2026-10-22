//go:build !windows

package graph

import (
	"errors"
	"os"
)

func privateGTDTestFileInfo(path string) (os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, statErr := f.Stat()
	return info, errors.Join(statErr, f.Close())
}
