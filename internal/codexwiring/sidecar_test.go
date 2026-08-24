package codexwiring

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestSidecarRoundTrip verifies the trust sidecar write/read contract: what
// Wire records is exactly what LoadSidecar returns (REQ-CW-008 — doctor's
// divergence verdict and the regeneration comparison read this baseline).
func TestSidecarRoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := writeSidecar(root, []byte("hooks-bytes"), []byte("config-bytes")); err != nil {
		t.Fatalf("writeSidecar: %v", err)
	}
	doc, present, err := LoadSidecar(root)
	if err != nil {
		t.Fatalf("LoadSidecar: %v", err)
	}
	if !present {
		t.Fatal("LoadSidecar reports absent sidecar right after a write")
	}
	if doc.HooksSHA256 == "" || doc.ConfigSHA256 == "" {
		t.Errorf("sidecar hashes empty after write: %+v", doc)
	}
	sum := sha256.Sum256([]byte("hooks-bytes"))
	if doc.HooksSHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("hooks_sha256 = %q, want sha256 of the recorded hooks content", doc.HooksSHA256)
	}
}

// TestSidecarAbsentIsNotError verifies LoadSidecar treats a missing sidecar as
// "no baseline" (present=false, nil error) — doctor surfaces that state, it
// does not fail on it.
func TestSidecarAbsentIsNotError(t *testing.T) {
	root := t.TempDir()
	_, present, err := LoadSidecar(root)
	if err != nil {
		t.Fatalf("missing sidecar must not error: %v", err)
	}
	if present {
		t.Error("absent sidecar reported as present")
	}
	if _, statErr := os.Stat(filepath.Join(root, SidecarPath)); statErr == nil {
		t.Error("LoadSidecar created the sidecar — reads must not write")
	}
}

// TestWireWritesSidecarOnlyOnHooksWrite anchors the sidecar semantics: an
// unchanged regeneration (no hooks write) leaves the sidecar untouched.
func TestWireWritesSidecarOnlyOnHooksWrite(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	sidecarPath := filepath.Join(root, SidecarPath)
	before, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatal(err)
	}
	var out, warn bytes.Buffer
	if _, err := Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire(unchanged): %v", err)
	}
	after, err := os.ReadFile(sidecarPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("unchanged regeneration rewrote the sidecar:\nbefore %q\nafter  %q", before, after)
	}
}

// TestWireReestablishesMissingSidecar verifies the self-healing baseline: a
// lost sidecar (e.g. a force-reinit resets .moai/state while hooks.json is
// unchanged, so no hooks write happens) is re-recorded on the next pass when
// the on-disk hooks still byte-match the render — otherwise doctor's
// divergence detection silently degrades to "no baseline".
func TestWireReestablishesMissingSidecar(t *testing.T) {
	root, _, _, _ := wireFresh(t)
	sidecarPath := filepath.Join(root, SidecarPath)
	if err := os.Remove(sidecarPath); err != nil {
		t.Fatal(err)
	}

	var out, warn bytes.Buffer
	res, err := Wire(root, &out, &warn)
	if err != nil {
		t.Fatalf("Wire(sidecar-missing): %v", err)
	}
	if res.HooksWritten {
		t.Errorf("unchanged hooks must not be rewritten (only the sidecar re-establishes)")
	}
	doc, present, err := LoadSidecar(root)
	if err != nil || !present {
		t.Fatalf("sidecar not re-established (present=%v err=%v)", present, err)
	}
	raw, rerr := os.ReadFile(filepath.Join(root, HooksRelPath))
	if rerr != nil {
		t.Fatal(rerr)
	}
	sum := sha256.Sum256(raw)
	if doc.HooksSHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("re-established sidecar hash does not match the on-disk hooks.json")
	}
}
