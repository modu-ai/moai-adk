package ledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// syncPath makes a file or a directory durable. A variable so a case can see it called.
var syncPath = func(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close() //nolint:errcheck // the sync error is the one to report
		return err
	}
	return f.Close()
}

// Snapshot writes a consistent copy of the ledger to dst: VACUUM INTO a temporary file in
// the same directory, then a rename over dst, so a reader never sees half a snapshot. The
// file is private to its owner. The temporary file is synced before the rename and the
// directory after it, so a crash leaves the old snapshot or the whole new one.
func (s *Store) Snapshot(ctx context.Context, dst string) error {
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(dst)+".tmp-*")
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	name := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(name) //nolint:errcheck // gone after the rename; a leftover is removed here
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, name); err != nil {
		return fmt.Errorf("ledger: snapshot: %w", err)
	}
	if err := os.Chmod(name, 0o600); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if err := syncPath(name); err != nil {
		return fmt.Errorf("ledger: sync snapshot: %w", err)
	}
	if err := os.Rename(name, dst); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if err := syncPath(dir); err != nil {
		return fmt.Errorf("ledger: sync snapshot directory: %w", err)
	}
	return nil
}
