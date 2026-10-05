package handoff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestClaimPending_ImportsLegacyOnce(t *testing.T) {
	t.Parallel()
	pd := t.TempDir()
	legacy := filepath.Join(handoffStateDir(pd), "pending.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(&PendingRecord{SchemaVersion: 1, SavedAt: time.Now(), Body: "legacy resume"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	rec, id, present, err := ClaimPending(pd, "claim-a")
	if err != nil || !present || rec == nil || id == 0 {
		t.Fatalf("legacy claim: rec=%v id=%d present=%v err=%v", rec, id, present, err)
	}
	if rec.Body != "legacy resume" {
		t.Fatalf("body=%q", rec.Body)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy file must be retired after import: %v", err)
	}
	if err := FinishClaim(pd, id, true, "test", "claim-a"); err != nil {
		t.Fatal(err)
	}
	if _, _, present, err := ClaimPending(pd, "claim-b"); err != nil || present {
		t.Fatalf("legacy must not be re-imported: present=%v err=%v", present, err)
	}
}

// writeDecoyPendingMD creates a session-handoff/pending.md decoy and returns its
// path + initial modtime, so tests can assert the reverse-handoff flow never
// touches the SessionEnd flow's file (path-isolation guard).
func writeDecoyPendingMD(t *testing.T, projectDir string) (string, time.Time) {
	t.Helper()
	dir := filepath.Join(projectDir, ".moai", "state", "session-handoff")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session-handoff: %v", err)
	}
	p := filepath.Join(dir, "pending.md")
	if err := os.WriteFile(p, []byte("---\nsprint: s\nspec: x\nstatus: c\nindex_line: l\n---\n## Next Session Entry Point\n```text\nx\n```\n"), 0o644); err != nil {
		t.Fatalf("write decoy pending.md: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat decoy: %v", err)
	}
	return p, info.ModTime()
}

// TestSavePending_WritesJSONNotMarkdown verifies save writes factory.db and
// does NOT create/modify
// session-handoff/pending.md.
func TestSavePending_WritesJSONNotMarkdown(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	decoyPath, decoyMTime := writeDecoyPendingMD(t, projectDir)

	rec := &PendingRecord{Body: "resume body"}
	if err := SavePending(projectDir, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}

	parsed, present, err := ReadPending(projectDir)
	if err != nil || !present {
		t.Fatalf("read pending row: present=%v err=%v", present, err)
	}
	if parsed.Body != "resume body" {
		t.Errorf("body: got %q, want %q", parsed.Body, "resume body")
	}

	// session-handoff/pending.md untouched (mtime unchanged).
	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy after save: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Errorf("session-handoff/pending.md was modified by save (mtime changed) — path isolation violated")
	}
}

// TestSavePending_Schema verifies REQ-006: the persisted record carries at least
// schema_version, body (verbatim), directives, conversation_language, saved_at.
func TestSavePending_Schema(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	before := time.Now().Add(-time.Second)

	rec := &PendingRecord{
		Body:                 "verbatim\nbody\nwith\nnewlines",
		ConversationLanguage: "ko",
		Directives:           Directives{Ultrathink: true},
	}
	if err := SavePending(projectDir, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}

	parsed, present, err := ReadPending(projectDir)
	if err != nil || !present {
		t.Fatalf("read pending row: present=%v err=%v", present, err)
	}
	if parsed.SchemaVersion != PendingSchemaVersion {
		t.Errorf("schema_version: got %d, want %d", parsed.SchemaVersion, PendingSchemaVersion)
	}
	if parsed.Body != "verbatim\nbody\nwith\nnewlines" {
		t.Errorf("body not verbatim: got %q", parsed.Body)
	}
	if !parsed.Directives.Ultrathink {
		t.Errorf("directives.ultrathink: got false, want true")
	}
	if parsed.ConversationLanguage != "ko" {
		t.Errorf("conversation_language: got %q, want %q", parsed.ConversationLanguage, "ko")
	}
	if parsed.SavedAt.Before(before) {
		t.Errorf("saved_at not auto-populated: got %v", parsed.SavedAt)
	}
}

// TestClearPending verifies REQ-007: clear removes handoff/pending.json and does
// NOT touch session-handoff/pending.md. An absent file is a no-op.
func TestClearPending(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	decoyPath, decoyMTime := writeDecoyPendingMD(t, projectDir)

	if err := SavePending(projectDir, &PendingRecord{Body: "b"}); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
	if _, present, err := ReadPending(projectDir); err != nil || !present {
		t.Fatalf("pending row should exist before clear: present=%v err=%v", present, err)
	}

	if err := ClearPending(projectDir); err != nil {
		t.Fatalf("ClearPending: %v", err)
	}
	if _, present, err := ReadPending(projectDir); err != nil || present {
		t.Errorf("pending row should be cleared: present=%v err=%v", present, err)
	}

	// Decoy untouched.
	info, err := os.Stat(decoyPath)
	if err != nil {
		t.Fatalf("stat decoy after clear: %v", err)
	}
	if !info.ModTime().Equal(decoyMTime) {
		t.Errorf("session-handoff/pending.md was modified by clear — path isolation violated")
	}

	// Second clear (absent file) is a no-op.
	if err := ClearPending(projectDir); err != nil {
		t.Errorf("ClearPending on absent file should be a no-op, got: %v", err)
	}
}

// TestReadPending_States verifies the three ReadPending states: absent, valid,
// corrupt (the present bool drives the SessionStart handler's no-op vs warn).
func TestReadPending_States(t *testing.T) {
	t.Parallel()

	t.Run("absent", func(t *testing.T) {
		t.Parallel()
		rec, present, err := ReadPending(t.TempDir())
		if rec != nil || present || err != nil {
			t.Errorf("absent: got (%v, %v, %v), want (nil, false, nil)", rec, present, err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		pd := t.TempDir()
		if err := SavePending(pd, &PendingRecord{Body: "b", ConversationLanguage: "en"}); err != nil {
			t.Fatalf("SavePending: %v", err)
		}
		rec, present, err := ReadPending(pd)
		if err != nil || !present || rec == nil {
			t.Fatalf("valid: got (%v, %v, %v), want (rec, true, nil)", rec, present, err)
		}
		if rec.Body != "b" {
			t.Errorf("body: got %q, want %q", rec.Body, "b")
		}
	})

	t.Run("corrupt", func(t *testing.T) {
		t.Parallel()
		pd := t.TempDir()
		path := PendingPath(pd)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("not a sqlite database"), 0o600); err != nil {
			t.Fatalf("write corrupt: %v", err)
		}
		rec, present, err := ReadPending(pd)
		if err == nil || !present || rec != nil {
			t.Errorf("corrupt: got (%v, %v, %v), want (nil, true, err)", rec, present, err)
		}
	})
}

// TestSavePending_MkdirFails covers the MkdirAll error branch by placing a
// regular file where the state directory would be created.
func TestSavePending_MkdirFails(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	factoryDir := filepath.Dir(PendingPath(pd))
	if err := os.MkdirAll(filepath.Dir(factoryDir), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.WriteFile(factoryDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write factory blocker: %v", err)
	}
	if err := SavePending(pd, &PendingRecord{Body: "b"}); err == nil {
		t.Error("SavePending should fail when the state dir cannot be created")
	}
}

// TestClearPending_NonENOENTError covers the non-ENOENT remove-error branch by
// making pending.json a non-empty directory (os.Remove returns ENOTEMPTY).
func TestClearPending_NonENOENTError(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	// Make pending.json a directory containing a child so Remove fails non-ENOENT.
	pjDir := filepath.Join(handoffStateDir(pd), "pending.json")
	if err := os.MkdirAll(pjDir, 0o755); err != nil {
		t.Fatalf("mkdir pending dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pjDir, "child"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write child: %v", err)
	}
	if err := ClearPending(pd); err == nil {
		t.Error("ClearPending should surface a non-ENOENT remove error")
	}
}

// TestReadPending_NonENOENTError covers the non-ENOENT read-error branch by
// making pending.json a directory (ReadFile returns EISDIR, not ENOENT).
func TestReadPending_NonENOENTError(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	legacy := filepath.Join(handoffStateDir(pd), "pending.json")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatalf("mkdir pending dir: %v", err)
	}
	rec, present, err := ReadPending(pd)
	if err == nil {
		t.Error("ReadPending should surface a non-ENOENT read error")
	}
	if rec != nil || present {
		t.Errorf("non-ENOENT read: got (%v, %v), want (nil, false)", rec, present)
	}
}

// TestSavePending_NilRecord verifies the nil-record guard returns an error and
// writes nothing.
func TestSavePending_NilRecord(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	if err := SavePending(pd, nil); err == nil {
		t.Error("SavePending(nil) should return an error")
	}
	if _, err := os.Stat(PendingPath(pd)); !os.IsNotExist(err) {
		t.Error("SavePending(nil) must not create factory.db")
	}
}

// TestSavePending_FilePerm verifies factory.db is written 0o600.
func TestSavePending_FilePerm(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := SavePending(projectDir, &PendingRecord{Body: "b"}); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
	info, err := os.Stat(PendingPath(projectDir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if runtime.GOOS == "windows" {
		// Windows has no POSIX mode bits: Go synthesizes 0666/0444 from the
		// read-only attribute, so a 0600 comparison is not meaningful there.
		// The 0600 discipline stays asserted on unix.
		t.Skip("POSIX file mode bits are not represented on Windows; 0600 assertion covered on unix")
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("factory.db perm: got %o, want 0600", perm)
	}
}
