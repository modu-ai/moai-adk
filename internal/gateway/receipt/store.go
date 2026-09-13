package receipt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var ErrUnsupported = errors.New("private receipt persistence unsupported on this platform")
var ErrState = errors.New("invalid or inaccessible private receipt state")

type Store struct {
	mu        sync.RWMutex
	root      *os.Root
	dir, uuid string
	identity  os.FileInfo
	closed    bool
}

// OpenStore uses only an already authorized, existing private absolute root.
// create must be selected by the launcher for a new conversation; resume never
// initializes missing state. Windows stays closed until its durability gate.
func OpenStore(ctx context.Context, dir, authorizedUUID string, create bool) (*Store, error) {
	if e := platformSupported(); e != nil {
		return nil, e
	}
	if _, e := New(authorizedUUID); e != nil {
		return nil, e
	}
	if !filepath.IsAbs(dir) || filepath.Clean(dir) != dir {
		return nil, ErrState
	}
	for p := dir; ; p = filepath.Dir(p) {
		i, e := os.Lstat(p)
		if e != nil || i.Mode()&os.ModeSymlink != 0 {
			return nil, ErrState
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	identity, e := os.Lstat(dir)
	if e != nil || !private(identity, true) {
		return nil, ErrState
	}
	root, e := os.OpenRoot(dir)
	if e != nil {
		return nil, ErrState
	}
	s := &Store{root: root, dir: dir, uuid: authorizedUUID, identity: identity}
	e = s.transaction(ctx, create, func(guard func() error) error {
		if create {
			if _, e := root.Lstat("manifest.json"); !errors.Is(e, os.ErrNotExist) {
				return ErrState
			}
			m, _ := New(authorizedUUID)
			return s.write(ctx, m, guard, nil)
		}
		_, e := s.read()
		return e
	})
	if e != nil {
		_ = root.Close()
		return nil, e
	}
	return s, nil
}
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.root.Close()
}
func (s *Store) rootValid() error {
	i, e := os.Lstat(s.dir)
	if e != nil || !private(i, true) || !os.SameFile(s.identity, i) {
		return ErrState
	}
	i, e = s.root.Stat(".")
	if e != nil || !os.SameFile(s.identity, i) {
		return ErrState
	}
	return nil
}
func (s *Store) file(name string, flags int) (*os.File, error) {
	before, e := s.root.Lstat(name)
	if e != nil && !errors.Is(e, os.ErrNotExist) {
		return nil, ErrState
	}
	if e == nil && !private(before, false) {
		return nil, ErrState
	}
	f, e := s.root.OpenFile(name, flags|safeOpenFlags(), 0600)
	if e != nil {
		return nil, ErrState
	}
	after, e := f.Stat()
	path, e2 := s.root.Lstat(name)
	if e != nil || e2 != nil || !private(after, false) || !private(path, false) || !os.SameFile(after, path) || (before != nil && !os.SameFile(before, after)) {
		_ = f.Close()
		return nil, ErrState
	}
	return f, nil
}

// @MX:WARN: [AUTO] Process lock serializes manifest updates across Store instances.
// @MX:REASON: Separate descriptors avoid lost updates; cancellation bounds waits.
func (s *Store) transaction(ctx context.Context, createLock bool, action func(func() error) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrState
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := s.rootValid(); e != nil {
		return e
	}
	flags := os.O_RDWR
	if createLock {
		flags |= os.O_CREATE
	}
	f, e := s.file(".lock", flags)
	if e != nil {
		return e
	}
	defer func() { _ = f.Close() }() // lock release is governed by unlock + process exit
	if e = lock(ctx, f); e != nil {
		return e
	}
	defer unlock(f)
	guard := func() error {
		current, e := s.root.Lstat(".lock")
		opened, e2 := f.Stat()
		if e != nil || e2 != nil || !private(current, false) || !private(opened, false) || !os.SameFile(current, opened) {
			return ErrState
		}
		return s.rootValid()
	}
	if e = guard(); e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = action(guard); e != nil {
		return e
	}
	return guard()
}
func (s *Store) Snapshot(ctx context.Context) (*Manifest, error) {
	var m *Manifest
	e := s.transaction(ctx, false, func(guard func() error) error { var e error; m, e = s.read(); return e })
	return m, e
}
func (s *Store) Publish(ctx context.Context, c Candidate) error {
	return s.transaction(ctx, false, func(guard func() error) error {
		m, e := s.read()
		if e != nil {
			return e
		}
		baseline, e := m.Marshal()
		if e != nil {
			return e
		}
		expected := Hash(baseline)
		previous := m.generation
		if e = m.Publish(c); e != nil {
			return e
		}
		if m.generation == previous {
			return nil
		}
		if e = ctx.Err(); e != nil {
			return e
		}
		return s.write(ctx, m, guard, &expected)
	})
}

// Rebase persists the public-history reset after a verified compaction. The
// once-per-epoch and digest contrast live in the caller's RebaseLedger; the
// store only applies the durable reset.
func (s *Store) Rebase(ctx context.Context) error {
	return s.transaction(ctx, false, func(guard func() error) error {
		m, e := s.read()
		if e != nil {
			return e
		}
		baseline, e := m.Marshal()
		if e != nil {
			return e
		}
		expected := Hash(baseline)
		if e = m.Rebase(); e != nil {
			return e
		}
		if e = ctx.Err(); e != nil {
			return e
		}
		return s.write(ctx, m, guard, &expected)
	})
}
func (s *Store) read() (*Manifest, error) {
	f, e := s.file("manifest.json", os.O_RDONLY)
	if e != nil {
		return nil, e
	}
	defer func() { _ = f.Close() }() // read-only manifest source; no write-back to lose
	raw, e := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if e != nil {
		return nil, ErrState
	}
	return Parse(raw, s.uuid)
}
func (s *Store) write(ctx context.Context, m *Manifest, guard func() error, expected *Digest) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	raw, e := m.Marshal()
	if e != nil {
		return e
	}
	var nonce [16]byte
	if _, e = rand.Read(nonce[:]); e != nil {
		return ErrState
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = guard(); e != nil {
		return e
	}
	name := ".manifest-" + hex.EncodeToString(nonce[:])
	f, e := s.file(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY)
	if e != nil {
		return e
	}
	defer func() { _ = s.root.Remove(name) }() // temp manifest cleanup; the published copy owns the state
	_, e = f.Write(raw)
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		return ErrState
	}
	// Check the locked inode and the snapshot we read, immediately before
	// commit. A replaced lock must not let a stale transaction erase a newer one.
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = guard(); e != nil {
		return e
	}
	if expected == nil {
		if _, e = s.root.Lstat("manifest.json"); !errors.Is(e, os.ErrNotExist) {
			return ErrState
		}
	} else {
		current, err := s.read()
		if err != nil {
			return err
		}
		baseline, err := current.Marshal()
		if err != nil || Hash(baseline) != *expected {
			return ErrState
		}
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = guard(); e != nil {
		return e
	}
	// The final filesystem guard can itself outlast cancellation.
	if e = ctx.Err(); e != nil {
		return e
	}
	if e = s.root.Rename(name, "manifest.json"); e != nil {
		return ErrState
	}
	directory, e := s.root.Open(".")
	if e != nil {
		return ErrState
	}
	e = directory.Sync()
	_ = directory.Close()
	if e != nil {
		return ErrState
	}
	if e = guard(); e != nil {
		return e
	}
	published, e := s.read()
	if e != nil {
		return e
	}
	readback, e := published.Marshal()
	if e != nil || Hash(raw) != Hash(readback) {
		return ErrState
	}
	return guard()
}
