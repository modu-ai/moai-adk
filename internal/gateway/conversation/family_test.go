package conversation

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFamilyNewAndExactResumeNeedsCompletedTranscript(t *testing.T) {
	m := newTestManager(t)
	f, err := m.New(context.Background(), NewRequest{CWD: filepath.Join(m.root, "project"), Project: "project"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Resume(context.Background(), f.UUID); err == nil {
		t.Fatal("incomplete family resumed")
	}
	transcript := filepath.Join(f.ConfigDir, "projects", "project", f.UUID+".jsonl")
	if err = os.MkdirAll(filepath.Dir(transcript), 0700); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"session_id":"` + f.UUID + `","project":"project","type":"assistant"}
{"session_id":"` + f.UUID + `","project":"project","type":"completed"}
`)
	if err = os.WriteFile(transcript, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err = m.Complete(context.Background(), f.UUID, transcript, 1); err != nil {
		t.Fatal(err)
	}
	r, err := m.Resume(context.Background(), f.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Args) != 2 || r.Args[0] != "--resume" || r.Args[1] != f.UUID {
		t.Fatalf("args=%v", r.Args)
	}
}

func TestContinueOrdersAndRejectsTies(t *testing.T) {
	m := newTestManager(t)
	a := mustCompleted(t, m, "a", 1)
	_, _ = m.New(context.Background(), NewRequest{CWD: filepath.Join(m.root, "b"), Project: "b"})
	got, err := m.Continue(context.Background(), "a")
	if err != nil || got.UUID != a.UUID {
		t.Fatalf("continue=%+v err=%v", got, err)
	}
	if _, err := m.Continue(context.Background(), "missing"); err == nil {
		t.Fatal("missing project accepted")
	}
}

func TestForkCopiesReceiptSnapshotButOwnsLease(t *testing.T) {
	m := newTestManager(t)
	p := mustCompleted(t, m, "p", 1)
	f, err := m.Fork(context.Background(), p.UUID)
	if err != nil {
		t.Fatal(err)
	}
	if f.UUID == p.UUID || f.FamilyID != p.FamilyID || len(f.Args) != 5 {
		t.Fatalf("fork=%+v", f)
	}
	l, err := m.AcquireLease(f.FamilyID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.AcquireLease(f.FamilyID); err == nil {
		t.Fatal("duplicate lease")
	}
	l.Release()
}

func TestSecureNamespaceMatrix(t *testing.T) {
	if got, err := ResolveSecureNamespace("", true, "source", true); err != nil || got != "" {
		t.Fatalf("explicit secure empty: %q %v", got, err)
	}
	if got, err := ResolveSecureNamespace("", false, "source", true); err != nil || got != "source" {
		t.Fatalf("source fallback: %q %v", got, err)
	}
	if got, err := ResolveSecureNamespace("", false, "", false); err != nil || got != "" {
		t.Fatalf("default namespace: %q %v", got, err)
	}
	if _, err := ResolveSecureNamespace("", false, "", true); err == nil {
		t.Fatal("explicit empty source must be rejected")
	}
	if got := SelectSecureNamespace("secure", "source"); got != "secure" {
		t.Fatal(got)
	}
	if got := SelectSecureNamespace("", "source"); got != "source" {
		t.Fatal(got)
	}
	if got := SelectSecureNamespace("", ""); got != "" {
		t.Fatal(got)
	}
	if err := ValidateSecureNamespace("", true); err == nil {
		t.Fatal("legacy validation should reject explicit empty source")
	}
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustCompleted(t *testing.T, m *Manager, project string, seq uint64) Descriptor {
	t.Helper()
	f, err := m.New(context.Background(), NewRequest{CWD: filepath.Join(m.root, project), Project: project})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.ConfigDir, "projects", project, f.UUID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"session_id":"` + f.UUID + `","project":"` + project + `","type":"completed"}
`)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Complete(context.Background(), f.UUID, path, seq); err != nil {
		t.Fatal(err)
	}
	return f
}
