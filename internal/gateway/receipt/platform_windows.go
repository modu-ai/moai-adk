//go:build windows

package receipt

import (
	"context"
	"os"
)

func platformSupported() error             { return ErrUnsupported }
func safeOpenFlags() int                   { return 0 }
func private(os.FileInfo, bool) bool       { return false }
func lock(context.Context, *os.File) error { return ErrUnsupported }
func unlock(*os.File)                      {}
