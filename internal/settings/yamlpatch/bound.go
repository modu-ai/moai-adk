package yamlpatch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
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
func PatchFileAtRoot(root, relative string, edits []KeyEdit, options RootOptions) error {
	if len(edits) == 0 {
		return nil
	}
	return withSectionRoot(root, relative, options, func(r *os.Root, path string) error {
		return patchFileWithIO(path, edits, options.Mode, options.AllowEmpty, r.ReadFile,
			func(path string, data []byte, perm os.FileMode) error {
				return atomicWriteAtRoot(r, path, data, perm)
			})
	})
}

// WriteFileAtRoot atomically publishes an already serialized section beneath
// the trusted project root, with the same path and mode policy as rooted edits.
// It preserves data exactly; AllowEmpty applies only to PatchFileAtRoot.
// A zero Mode preserves an existing file's mode or uses 0644 for a new file.
func WriteFileAtRoot(root, relative string, data []byte, options RootOptions) error {
	return withSectionRoot(root, relative, options, func(r *os.Root, path string) error {
		return atomicWriteAtRoot(r, path, data, options.Mode)
	})
}

func withSectionRoot(root, relative string, options RootOptions, operation func(*os.Root, string) error) (retErr error) {
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
	return operation(r, relative)
}

func checkRootSectionPath(root *os.Root, relative string) error {
	return checkRootPath(root, relative, false)
}

func checkRootPath(root *os.Root, relative string, directoryTarget bool) error {
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
		leaf := i == len(parts)-1
		if !leaf && !info.IsDir() || leaf && !info.Mode().IsRegular() && (!directoryTarget || !info.IsDir()) || info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
			return fmt.Errorf("section component %s is not a regular file or real parent directory", current)
		}
	}
	return nil
}

func atomicWriteAtRoot(root *os.Root, path string, data []byte, perm os.FileMode) error {
	// Public rooted operations already require a regular leaf before reading.
	// Keep the lower publication helper's directory-target rename/error contract,
	// so a failed installation still exercises and verifies temporary cleanup.
	if err := checkRootPath(root, path, true); err != nil {
		return fmt.Errorf("yamlpatch: stat %s: %w", path, err)
	}
	err := atomicfile.Write(path, data, defaultFilePerm, atomicfile.RootOptions{Root: root, ForceMode: perm,
		BeforeRename: func() error { return checkRootPath(root, path, true) }})
	if err == nil {
		return nil
	}
	var failure *atomicfile.WriteError
	if errors.As(err, &failure) {
		return fmt.Errorf("yamlpatch: %s: %w", failure.Op, err)
	}
	return fmt.Errorf("yamlpatch: write: %w", err)
}
