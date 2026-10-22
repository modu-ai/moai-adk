package sqliteguard

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"modernc.org/sqlite"
)

// Open binds the supplied writable file URI to the prepared canonical path.
// Each physical connection receives its own same-object privacy pins.
func Open(path, dsn string) (*sql.DB, error) {
	canonical, boundDSN, err := canonicalBinding(path, dsn)
	if err != nil {
		return nil, err
	}
	base, err := sqlite.NewConnector(boundDSN)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(&connector{path: canonical, base: base, owners: processOwners, prepare: preparePins}), nil
}

func canonicalBinding(path, dsn string) (string, string, error) {
	if strings.TrimSpace(path) != path || !filepath.IsAbs(path) {
		return "", "", errors.New("SQLite privacy: expected absolute file path")
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme != "file" || u.Host != "" || u.Opaque != "" || u.Fragment != "" || u.User != nil {
		return "", "", errors.New("SQLite privacy: expected local file URI")
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", "", err
	}
	for k := range q {
		if k != "_pragma" && k != "_txlock" {
			return "", "", fmt.Errorf("SQLite privacy: unsupported URI parameter %q", k)
		}
	}
	native := filepath.FromSlash(u.Path)
	if len(native) > 2 && native[0] == '\\' && native[2] == ':' {
		native = native[1:]
	}
	if !strings.EqualFold(filepath.Clean(path), filepath.Clean(native)) {
		return "", "", errors.New("SQLite privacy: URI path does not name prepared file")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return "", "", err
	}
	canonical := filepath.Join(parent, filepath.Base(path))
	// Normalize the already-bound URI and guard path together, preserving
	// the caller's query. The runner's 8.3 aliases must name the same root.
	u.Path = filepath.ToSlash(canonical)
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	u.RawPath = ""
	return canonical, u.String(), nil
}

func preparePins(path string, acquired func(io.Closer)) error {
	if err := privatefile.CheckDirectory(filepath.Dir(path)); err != nil {
		return err
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := privatefile.OpenPinned(path+suffix, acquired); err != nil {
			return err
		}
	}
	// Retained handles forbid deletion of these objects. Do not unlink on
	// failure or replace any existing database/sidecar payload.
	return nil
}

// EnsureDirectory creates missing components privately and validates the
// requested existing leaf. Existing ancestors are never chmodded or taken
// over. A symlink/non-directory leaf is refused by native validation.
// @MX:ANCHOR: [AUTO] Create managed namespaces before SQLite or home-state payloads.
// @MX:REASON: Home/project layout, factory opening and profile lease opening share this boundary.
func EnsureDirectory(path string) error {
	if !filepath.IsAbs(path) || strings.TrimSpace(path) != path {
		return errors.New("SQLite privacy: expected absolute directory path")
	}
	info, err := os.Lstat(path)
	if err == nil {
		if !info.IsDir() {
			return errors.New("mkdir SQLite private namespace: parent is not a real directory")
		}
		// Do not reapply an inheritable DACL to an existing tree: Windows
		// propagation could alter existing children outside this operation.
		return privatefile.CheckDirectory(path)
	}
	if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(path)
	if parent == path {
		return err
	}
	if _, err := os.Lstat(parent); os.IsNotExist(err) {
		if err := EnsureDirectory(parent); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if err := privatefile.Mkdir(path); err != nil && !os.IsExist(err) {
		return err
	}
	return privatefile.CheckDirectory(path)
}

// SecureArtifact verifies the before-bytes policy installed by the physical
// connector; it never repairs a weak file after SQLite has already written.
func SecureArtifact(path string) error { return privatefile.Check(path) }
