package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// The seam allows deterministic replacement after open; validation and reading
// must continue to use that same descriptor, never reopen its pathname.
var codexOpenLocalFileFn = openCodexLocalFile

func readCodexLocalInstruction(root, name string) ([]byte, error) {
	path := filepath.Join(root, name)
	f, err := codexOpenLocalFileFn(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		// Diagnostic only: Lstat never grants access or feeds a later read.
		if info, statErr := os.Lstat(path); statErr == nil && !info.Mode().IsRegular() {
			return nil, &codexPathGuardError{Rel: name, Reason: "not a regular file (" + codexModeName(info.Mode()) + ")"}
		}
		return nil, &codexPathGuardError{Rel: name, Reason: "open: " + err.Error()}
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, &codexPathGuardError{Rel: name, Reason: "fstat: " + err.Error()}
	}
	if !info.Mode().IsRegular() {
		return nil, &codexPathGuardError{Rel: name, Reason: "not a regular file (" + codexModeName(info.Mode()) + ")"}
	}
	body, err := io.ReadAll(f)
	if err != nil {
		return nil, &codexPathGuardError{Rel: name, Reason: "read: " + err.Error()}
	}
	return body, nil
}
