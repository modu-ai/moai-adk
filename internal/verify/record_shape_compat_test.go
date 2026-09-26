package verify

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// legacySnapshotJSON is a snapshot document in the shape written before the
// receipt fields existed (no config_digest / tool_version / verdict keys).
const legacySnapshotJSON = `{
  "key": "1111111111111111111111111111111111111111:abcdefabcdefabcd",
  "recorded_at": "2026-09-23T10:00:00Z",
  "checks": [
    {
      "check_id": "test",
      "command": "go test ./internal/verify/",
      "exit_code": 0,
      "recorded_at": "2026-09-23T10:00:00Z",
      "duration_ms": 1200
    }
  ]
}`

// TestSnapshotRecordShapeBackwardCompatible characterizes the reader contract
// the receipt extension must keep: a snapshot written before the extension
// still loads, still serves Source.Lookup, and re-saving an entry that sets
// none of the new fields writes none of their keys, so an older binary reading
// the file sees the shape it always saw.
func TestSnapshotRecordShapeBackwardCompatible(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	const key = "1111111111111111111111111111111111111111:abcdefabcdefabcd"
	path := SnapshotPath(root, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(legacySnapshotJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	snap, err := Load(root, key)
	if err != nil || snap == nil {
		t.Fatalf("legacy snapshot must load: snap=%v err=%v", snap, err)
	}
	if got := snap.FindCommand("go test ./internal/verify/"); got == nil || got.ExitCode != 0 {
		t.Fatalf("legacy entry not found by command: %+v", got)
	}

	src := &Source{
		ProjectRoot: root,
		Now:         func() time.Time { return time.Date(2026, 9, 23, 10, 1, 0, 0, time.UTC) },
		KeyFunc:     func(context.Context, string) (string, error) { return key, nil },
	}
	if exit, _, ok := src.Lookup(context.Background(), "go test ./internal/verify/"); !ok || exit != 0 {
		t.Fatalf("Source.Lookup over a legacy snapshot: exit=%d ok=%v", exit, ok)
	}

	if err := Save(root, snap); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{`"config_digest"`, `"tool_version"`, `"verdict"`} {
		if strings.Contains(string(data), k) {
			t.Errorf("re-saved legacy snapshot gained key %s:\n%s", k, data)
		}
	}
}
