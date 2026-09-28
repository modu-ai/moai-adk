package codexwiring

// seed_test.go — audit F4 (card t1273, sync-audit-opus.md): SeedHooksIfMissing
// had 0.0% in-package coverage and the package sat below the 85% profile
// threshold. The three branches that decide the function's behavior:
//
//   - seed-when-absent: the file is created with RenderHooks(nil)'s bytes.
//   - existing-file byte identity: any existing hooks.json — user-owned,
//     moai-owned, or mixed — is left untouched (the idempotency contract).
//   - stat error: a failure that is NOT ErrNotExist is caller-visible
//     (fail-open posture is the caller's decision, REQ-HN-007).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSeedHooksIfMissing_SeedsAbsentFile verifies the primary branch: a root
// without wiring gets .codex/hooks.json created with the rendered MoAI-owned
// table, and the call reports that it seeded.
func TestSeedHooksIfMissing_SeedsAbsentFile(t *testing.T) {
	root := t.TempDir()

	seeded, err := SeedHooksIfMissing(root)
	if err != nil {
		t.Fatalf("SeedHooksIfMissing: %v", err)
	}
	if !seeded {
		t.Error("SeedHooksIfMissing reported false on an absent file; want true")
	}

	data, err := os.ReadFile(filepath.Join(root, HooksRelPath))
	if err != nil {
		t.Fatalf("seeded %s must exist: %v", HooksRelPath, err)
	}
	want, err := RenderHooks(nil)
	if err != nil {
		t.Fatalf("RenderHooks(nil): %v", err)
	}
	if string(data) != string(want) {
		t.Error("seeded bytes differ from RenderHooks(nil) — a second hooks.json builder would drift")
	}
	if !strings.Contains(string(data), `"moai hook `) {
		t.Errorf("seeded hooks.json must carry moai hook commands; got %s", data)
	}
}

// TestSeedHooksIfMissing_ExistingFileByteIdentity verifies the idempotency
// contract: an existing hooks.json is never rewritten, whatever it contains.
func TestSeedHooksIfMissing_ExistingFileByteIdentity(t *testing.T) {
	root := t.TempDir()
	hooksPath := filepath.Join(root, HooksRelPath)
	userDoc := `{"description":"user desc","hooks":{"UserEvent":[{"hooks":[{"type":"command","command":"user-tool"}]}]}}`
	if err := os.MkdirAll(filepath.Dir(hooksPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(hooksPath, []byte(userDoc), 0o600); err != nil {
		t.Fatalf("write user doc: %v", err)
	}

	seeded, err := SeedHooksIfMissing(root)
	if err != nil {
		t.Fatalf("SeedHooksIfMissing: %v", err)
	}
	if seeded {
		t.Error("SeedHooksIfMissing reported true for an existing file; want false (no rewrite)")
	}

	after, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(after) != userDoc {
		t.Errorf("existing hooks.json must stay byte-identical; got %s", after)
	}
}

// TestSeedHooksIfMissing_StatErrorIsCallerVisible verifies the stat-error
// branch: a failure that is not ErrNotExist (here, .codex occupied by a
// regular file, so the hooks.json stat fails with ENOTDIR) returns an error
// naming the path instead of silently seeding or silently skipping.
func TestSeedHooksIfMissing_StatErrorIsCallerVisible(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".codex"), []byte("occupied"), 0o644); err != nil {
		t.Fatalf("occupy .codex: %v", err)
	}

	seeded, err := SeedHooksIfMissing(root)
	if err == nil {
		t.Fatal("stat error over .codex/hooks.json must surface, not be swallowed")
	}
	if seeded {
		t.Error("a failed seed must not report that it seeded")
	}
	if !strings.Contains(err.Error(), HooksRelPath) {
		t.Errorf("error must name the refused path %s; got %v", HooksRelPath, err)
	}
	if _, statErr := os.Stat(filepath.Join(root, HooksRelPath)); statErr == nil {
		t.Error("a failed seed must not leave a hooks.json behind")
	}
}
