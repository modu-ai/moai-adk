package ledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Snapshot writes a consistent copy of the ledger to dst: VACUUM INTO a temporary file in
// the same directory, then a rename over dst, so a reader never sees half a snapshot. The
// file is private to its owner.
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
	if err := os.Rename(name, dst); err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	return nil
}
