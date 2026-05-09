package worktree

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveSnapshot serializes the snapshot to JSON and writes to path.
// Parent directories are created with 0755; file written with 0644.
func SaveSnapshot(snap *Snapshot, path string) error {
	if snap == nil {
		return fmt.Errorf("nil snapshot")
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write snapshot %s: %w", path, err)
	}
	return nil
}

// LoadSnapshot reads a JSON-serialized snapshot from path.
func LoadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot %s: %w", path, err)
	}
	snap := &Snapshot{}
	if err := json.Unmarshal(data, snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return snap, nil
}

// SnapshotPath returns the canonical path for a snapshot ID under the project root.
func SnapshotPath(projectRoot, snapshotID string) string {
	return filepath.Join(projectRoot, fmt.Sprintf(SnapshotPathTemplate, snapshotID))
}
