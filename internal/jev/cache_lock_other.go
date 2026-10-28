//go:build !unix

package jev

// cache_lock_other.go — the non-unix counterpart of cache_lock_unix.go: the
// alias-map update runs WITHOUT a cross-process lock here (no portable flock
// — Windows file locking needs LockFileEx and a different unlock model). The
// serialized update is the unix behavior this package ships and tests; on
// other platforms the update stays atomic-per-file (writeModelsMap) but a
// concurrent model switch can still lose a Prior append, as documented in
// updateAliasMap. Single-user local caches make that window practically
// unreachable there; revisit with LockFileEx if a multi-process Windows
// deployment ever needs it.

import "os"

// lockAliasMap is the no-op stub for non-unix platforms.
func lockAliasMap(dir string) (*os.File, error) {
	return nil, nil
}

// unlockAliasMap is the no-op stub for non-unix platforms.
func unlockAliasMap(f *os.File) {}
