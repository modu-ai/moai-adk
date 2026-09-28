package kanban

// record_readall_test.go — ReadAll is the bulk reader the doctor factory
// section (SPEC-ROLE-NAMING-CODE-001 M3) consumes. Covered here, in the
// package that owns the record format, so the sweep's skip rules are pinned
// at the source rather than inferred from a consumer's coverage.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAllReturnsRecordsAndSkipsMalformed(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(RuntimeStateDirForRoot(root), name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("s1.json", `{"session_id":"s1","role":"leader"}`+"\n")
	write("s2.json", `{"session_id":"s2","role":"lane"}`+"\n")
	write("broken.json", `{not json`)
	write("notes.txt", "not a record")
	if err := os.MkdirAll(filepath.Join(RuntimeStateDirForRoot(root), "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	records, err := ReadAll(root)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("ReadAll returned %d records, want 2 (malformed file and non-.json skipped)", len(records))
	}
	roles := map[string]bool{}
	for _, r := range records {
		roles[r.Role] = true
	}
	if !roles["leader"] || !roles["lane"] {
		t.Errorf("ReadAll records = %+v, want the leader and lane records", records)
	}
}

func TestReadAllAbsentStateDirIsEmptyNotError(t *testing.T) {
	records, err := ReadAll(t.TempDir())
	if err != nil {
		t.Fatalf("ReadAll on an absent state dir: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("ReadAll returned %d records, want 0", len(records))
	}
	if records == nil {
		t.Error("ReadAll returned a nil slice; want an empty slice the caller can range safely")
	}
}
