// Package atomicfile provides a shared atomic, mode-preserving file-write
// helper. It is the single choke point for writes into .moai/config/** and CLI
// persistence paths so that every future write is atomic (crash leaves either
// the complete new or the complete prior content) and preserves the
// destination's pre-existing permission bits.
//
// Existing-file modes are read via os.Stat; new-file modes come from the
// caller-supplied defaultMode. Staging stays owner-only until publication.
package atomicfile

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// tempPattern is the glob prefix the helper uses for its in-directory temp
// files. It is exported so tests and the regression guard can detect orphans.
const tempPattern = ".moai-config-*.tmp"

// Write writes data to path atomically using a temp file in the destination's
// directory followed by os.Rename, then applies the correct permission bits
// before the rename so the result file never inherits os.CreateTemp's 0600
// default.
//
// Mode handling:
//   - When path already exists, the result file keeps the destination's
//     pre-existing mode (read via os.Stat), so the write neither narrows nor
//     widens the mode.
//   - When path does not exist, the result file is created at defaultMode (the
//     caller-supplied default; use defs.FilePerm for regular config, 0o600 for
//     caller-declared secret files).
//
// On any error before the rename completes, the temp file is removed so no
// orphan survives the failed call.
//
// @MX:ANCHOR: [AUTO] single atomic, mode-preserving write choke point for config/CLI persistence
// @MX:REASON: REQ-CAW-001..006 — every config/settings writer in internal/config + internal/cli routes through this; reversing its signature forces every call site to change.
// @MX:WARN: [AUTO] Generic and rooted writes share multiple failure and cleanup stages.
// @MX:REASON: Policy selection, mode resolution, payload IO, close, validation and publication must all preserve atomicity and propagate errors.
func Write(path string, data []byte, defaultMode os.FileMode, rooted ...RootOptions) (retErr error) {
	io, options, err := writerIO(rooted)
	if err != nil {
		return err
	}
	if options.Root != nil && (!filepath.IsLocal(path) || filepath.Clean(path) == ".") {
		return writeFailure("stat", "stat destination "+path, fmt.Errorf("target is not a local file path"))
	}
	dir := filepath.Dir(path)
	mode := defaultMode
	if info, err := io.stat(path); err == nil {
		if options.Root != nil && !info.Mode().IsRegular() && !info.IsDir() {
			return writeFailure("stat", "stat destination "+path, fmt.Errorf("target is not a regular file"))
		}
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return writeFailure("stat", "stat destination "+path, err)
	}
	if options.ForceMode != 0 {
		mode = options.ForceMode
	}
	tmp, tmpName, err := io.createTemp(dir, defaultMode)
	if err != nil {
		return writeFailure("create temp", "create temp file in "+dir, err)
	}
	closed, published := false, false
	defer func() {
		if !closed {
			if err := tmp.Close(); err != nil {
				retErr = errors.Join(retErr, writeFailure("close temp", "close temp file", err))
			}
		}
		if !published {
			if err := io.remove(tmpName); err != nil && !os.IsNotExist(err) {
				retErr = errors.Join(retErr, writeFailure("remove temp", "remove temp file", err))
			}
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return writeFailure("write temp", "write temp file", err)
	}
	if options.BeforeRename != nil {
		if err := options.BeforeRename(); err != nil {
			return writeFailure("stat", "validate destination "+path, err)
		}
	}
	// Broaden permissions only after the payload and destination checks finish.
	// Chmod the held object, then close it before the atomic rename.
	if err := tmp.Chmod(mode); err != nil {
		return writeFailure("chmod temp", fmt.Sprintf("chmod temp file to %v", mode), err)
	}
	closeErr := tmp.Close()
	closed = true
	if closeErr != nil {
		return writeFailure("close temp", "close temp file", closeErr)
	}
	if err := io.rename(tmpName, path); err != nil {
		return writeFailure("rename", "rename temp to "+path, err)
	}
	published = true
	return nil
}

// RootOptions binds the ordinary Write operation to an already opened root.
// Supplying a nil Root is an error, never permission to fall back to plain paths.
// ForceMode is optional: zero preserves the existing mode. BeforeRename can
// revalidate a caller's stricter component policy before atomic publication.
type RootOptions struct {
	Root         *os.Root
	ForceMode    os.FileMode
	BeforeRename func() error
}

// WriteError preserves the failing stage and wraps its original filesystem
// error. Callers may retain their own error prefix without parsing error text.
type WriteError struct {
	Op      string
	Err     error
	context string
}

func (e *WriteError) Error() string { return fmt.Sprintf("%s: %v", e.context, e.Err) }
func (e *WriteError) Unwrap() error { return e.Err }
func writeFailure(op, context string, err error) error {
	return &WriteError{Op: op, Err: err, context: context}
}

type writeOperations struct {
	stat       func(string) (os.FileInfo, error)
	createTemp func(string, os.FileMode) (*os.File, string, error)
	remove     func(string) error
	rename     func(string, string) error
}

func writerIO(rooted []RootOptions) (writeOperations, RootOptions, error) {
	if len(rooted) == 0 {
		return writeOperations{stat: os.Stat, remove: os.Remove, rename: os.Rename,
			createTemp: func(dir string, _ os.FileMode) (*os.File, string, error) {
				f, err := os.CreateTemp(dir, tempPattern)
				if err != nil {
					return nil, "", err
				}
				return f, f.Name(), nil
			}}, RootOptions{}, nil
	}
	if len(rooted) != 1 || rooted[0].Root == nil {
		return writeOperations{}, RootOptions{}, writeFailure("stat", "stat destination", fmt.Errorf("exactly one non-nil root option is required"))
	}
	options := rooted[0]
	root := options.Root
	return writeOperations{stat: root.Lstat, remove: root.Remove, rename: root.Rename,
		createTemp: func(dir string, _ os.FileMode) (*os.File, string, error) {
			for range 100 {
				name := filepath.Join(dir, strings.Replace(tempPattern, "*", rand.Text(), 1))
				f, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, defs.PrivateFilePerm)
				if os.IsExist(err) {
					continue
				}
				return f, name, err
			}
			return nil, "", fs.ErrExist
		}}, options, nil
}
