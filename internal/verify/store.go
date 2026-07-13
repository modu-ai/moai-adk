package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SnapshotDir is the snapshot artifact directory under a project root —
// gitignored runtime state in the established evidence-persistence namespace.
const SnapshotDir = ".moai/state/verify/snapshots"

// headPrefixLen bounds the HEAD-SHA portion of the snapshot filename.
const headPrefixLen = 12

// SnapshotPath returns the on-disk path for a key's snapshot file. The
// filename is a sanitized key prefix (the raw key carries ':', which is not a
// valid filename character on Windows).
func SnapshotPath(projectRoot, key string) string {
	name := key
	if head, digest, ok := strings.Cut(key, ":"); ok {
		if len(head) > headPrefixLen {
			head = head[:headPrefixLen]
		}
		name = head + "-" + digest
	}
	name = strings.ReplaceAll(name, ":", "-")
	return filepath.Join(projectRoot, SnapshotDir, name+".json")
}

// Load reads the snapshot for the given key. A missing file returns
// (nil, nil) — absence is the plain re-execution path, never an error. A
// stored-key mismatch (filename-collision defense) is also treated as absent.
func Load(projectRoot, key string) (*Snapshot, error) {
	path := SnapshotPath(projectRoot, key)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("snapshot load %s: %w", path, err)
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("snapshot parse %s: %w", path, err)
	}
	if s.Key != key {
		return nil, nil
	}
	return &s, nil
}

// Save writes the snapshot atomically: a temp file is written in the target
// directory, then renamed into place — never a partial in-place write a
// concurrent reader could observe mid-write. Identical-key last-writer-wins
// is benign (same tree ⇒ equivalent results).
func Save(projectRoot string, s *Snapshot) error {
	dir := filepath.Join(projectRoot, SnapshotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("snapshot mkdir %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot marshal: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("snapshot tmp create: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("snapshot tmp write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("snapshot tmp close: %w", err)
	}
	finalPath := SnapshotPath(projectRoot, s.Key)
	if err := os.Rename(tmpName, finalPath); err != nil {
		return fmt.Errorf("snapshot rename %s: %w", finalPath, err)
	}
	cleanup = false
	return nil
}

// RecordCheck loads (or creates) the snapshot for key, upserts the check entry
// — an entry with the identical command is replaced (fresher result for the
// same tree), a distinct command is appended — advances the snapshot-level
// RecordedAt to the entry's capture time, and saves atomically.
func RecordCheck(projectRoot, key string, entry CheckEntry) (*Snapshot, error) {
	s, err := Load(projectRoot, key)
	if err != nil {
		return nil, err
	}
	if s == nil {
		s = &Snapshot{Key: key}
	}
	if existing := s.FindCommand(entry.Command); existing != nil {
		*existing = entry
	} else {
		s.Checks = append(s.Checks, entry)
	}
	if entry.RecordedAt.After(s.RecordedAt) {
		s.RecordedAt = entry.RecordedAt
	}
	if err := Save(projectRoot, s); err != nil {
		return nil, err
	}
	return s, nil
}
