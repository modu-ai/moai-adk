//go:build !unix && !windows

package cli

import (
	"fmt"
	"os"
)

func openCodexLocalFile(path string) (*os.File, error) {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return nil, err
	}
	return nil, fmt.Errorf("race-resistant local instruction reading unsupported on this platform")
}
