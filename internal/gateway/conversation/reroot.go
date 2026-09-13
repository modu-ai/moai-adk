package conversation

import (
	"os"
	"path/filepath"
)

// TranscriptPath resolves the conversation record's native transcript pointer
// to its absolute path without the resume completion gate: a wedged
// transcript can be incomplete, and the launcher-side wedge recovery (card
// t700, SPEC-GATEWAY-WEDGE-REROOT-001) must still find the file the record
// points at. The record is only read — nothing is written and no receipt
// state is touched here; the recovery itself lives launcher-side
// (internal/cli) and is gateway-write-free by construction.
func (m *Manager) TranscriptPath(id string) (string, error) {
	m.mu.Lock()
	r, ok := m.entries[id]
	m.mu.Unlock()
	if !ok {
		return "", ErrMissing
	}
	if r.Transcript == "" {
		return "", ErrIncomplete
	}
	p := filepath.Join(r.ConfigDir, filepath.FromSlash(r.Transcript))
	if !within(r.ConfigDir, p) {
		return "", ErrInvalid
	}
	// Same escape guards as the resume path: every path component down to the
	// private config root must be a real directory, never a symlink.
	for part := p; ; part = filepath.Dir(part) {
		i, e := os.Lstat(part)
		if e != nil || i.Mode()&os.ModeSymlink != 0 {
			return "", ErrInvalid
		}
		if part == r.ConfigDir {
			break
		}
		if filepath.Dir(part) == part {
			return "", ErrInvalid
		}
	}
	return p, nil
}
