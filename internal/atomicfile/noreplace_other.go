//go:build !linux && !darwin && !windows

package atomicfile

import "errors"

func renameNoReplace(string, string) error { return errors.ErrUnsupported }
