//go:build !unix && !windows

package privatefile

import (
	"errors"
	"os"
)

func checkSingleLinkNative(_ *os.File) error {
	return errors.New("private file: native link count unavailable")
}
