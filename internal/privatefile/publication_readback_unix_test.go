//go:build !windows

package privatefile

import "os"

func publicationReadbackForTest(path string) ([]byte, error) {
	return os.ReadFile(path)
}
