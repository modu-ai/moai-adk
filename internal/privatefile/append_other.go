//go:build !unix && !windows

package privatefile

import (
	"fmt"
	"os"
)

func openAppendNative(string) (*os.File, error) {
	return nil, fmt.Errorf("private append: native no-follow open is unsupported on this platform")
}
