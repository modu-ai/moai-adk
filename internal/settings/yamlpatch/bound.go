package yamlpatch

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RootOptions preserves the generic seam's mode/empty-document choices.
// ParentPerm opts into creating missing parent directories beneath the root;
// zero retains the ordinary create-temp error for a missing parent.
type RootOptions struct {
	Mode       os.FileMode
	AllowEmpty bool
	ParentPerm os.FileMode
}

// PatchFileAtRoot patches a local relative path beneath an explicitly trusted
// project root. Root aliases are accepted, but linked child components are not.
// All file operations stay relative to the opened root, including publication.
//
// @MX:ANCHOR: [AUTO] project-bound YAML read, parent creation and atomic publication.
// @MX:REASON: settings, profile, config and CLI writers share this containment boundary.
func PatchFileAtRoot(root, relative string, edits []KeyEdit, options RootOptions) (retErr error) {
	if len(edits) == 0 {
		return nil
	}
	if !filepath.IsLocal(relative) || filepath.Clean(relative) == "." {
		return fmt.Errorf("yamlpatch: stat %s: target is not a local file path", relative)
	}
	relative = filepath.Clean(relative)
	r, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", relative, err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("yamlpatch: close root: %w", err))
		}
	}()
	if err := checkRootSectionPath(r, relative); err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", relative, err)
	}
	if options.ParentPerm != 0 {
		if err := r.MkdirAll(filepath.Dir(relative), options.ParentPerm); err != nil {
			return fmt.Errorf("yamlpatch: create parent: %w", err)
		}
	}
	if err := checkRootSectionPath(r, relative); err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", relative, err)
	}
	return patchFileWithIO(relative, edits, options.Mode, options.AllowEmpty, r.ReadFile,
		func(path string, data []byte, perm os.FileMode) error {
			return atomicWriteAtRoot(r, path, data, perm)
		})
}

func checkRootSectionPath(root *os.Root, relative string) error {
	parts := strings.Split(filepath.ToSlash(relative), "/")
	current := ""
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := root.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 || i < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("section component %s is linked, irregular or not a directory", current)
		}
	}
	return nil
}

func atomicWriteAtRoot(root *os.Root, path string, data []byte, perm os.FileMode) (retErr error) {
	if err := checkRootSectionPath(root, path); err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", path, err)
	}
	mode := defaultFilePerm
	info, err := root.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("yamlpatch: stat %s: %w", path, err)
	}
	if err == nil {
		mode = info.Mode().Perm()
	}
	if perm != 0 {
		mode = perm
	}
	var tmp *os.File
	var name string
	for range 100 {
		name = filepath.Join(filepath.Dir(path), ".yamlpatch-"+rand.Text()+".tmp")
		tmp, err = root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if !os.IsExist(err) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("yamlpatch: create temp: %w", err)
	}
	closed := false
	published := false
	defer func() {
		if !closed {
			if err := tmp.Close(); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("yamlpatch: close temp: %w", err))
			}
		}
		if !published {
			if err := root.Remove(name); err != nil && !os.IsNotExist(err) {
				retErr = errors.Join(retErr, fmt.Errorf("yamlpatch: remove temp: %w", err))
			}
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("yamlpatch: write temp: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		return fmt.Errorf("yamlpatch: chmod temp: %w", err)
	}
	closeErr := tmp.Close()
	closed = true
	if closeErr != nil {
		return fmt.Errorf("yamlpatch: close temp: %w", closeErr)
	}
	if err := checkRootSectionPath(root, path); err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", path, err)
	}
	if err := root.Rename(name, path); err != nil {
		return fmt.Errorf("yamlpatch: rename: %w", err)
	}
	published = true
	return nil
}
