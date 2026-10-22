package privatefile

import (
	"path/filepath"
	"strings"
)

// nativePath matches Go's long-path handling at the native Windows boundary.
// Keep short paths, device paths and already extended names unchanged.
func nativePath(path string) string {
	if len(path) < 248 || strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\\.\`) {
		return path
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if strings.HasPrefix(absolute, `\\`) {
		return `\\?\UNC\` + absolute[2:]
	}
	return `\\?\` + absolute
}
