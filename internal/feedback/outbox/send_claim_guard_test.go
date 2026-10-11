package outbox

// The claim-path guard (card-review finding, P2): the claim file's path was
// built by string concatenation — send-<id>.claim — so a tampered queue ID
// carrying a traversal (x/../../review-target) resolved OUTSIDE the
// user-scoped store, and an external file at the resolved path was treated
// as a stale lock and deleted. The ID must pass the f<digits> allowlist —
// the shape the drain mints (fmt.Sprintf("f%d", seq)) — before any path is
// constructed.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestItemSendClaimPathRejectsTraversal: a traversal-shaped ID is refused
// with an error, and the resolved path never leaves the store directory.
func TestItemSendClaimPathRejectsTraversal(t *testing.T) {
	for _, hostile := range []string{"x/../../review-target", "../evil", "f1/../../x", "", "f1x y", "g1"} {
		path, err := ItemSendClaimPath(hostile)
		if err == nil {
			t.Fatalf("ItemSendClaimPath(%q) returned %q without error — the f<digits> allowlist is not applied", hostile, path)
		}
	}
	if path, err := ItemSendClaimPath("f12"); err != nil || filepath.Base(path) != "send-f12.claim" {
		t.Fatalf("the legitimate id f12 was refused: path=%q err=%v", path, err)
	}
}

// TestClaimTraversalCannotTouchExternalFile: the end-to-end shape — a
// tampered queue item whose ID traverses must not create, claim, or delete
// a file outside the store, even one that looks like a stale lock.
func TestClaimTraversalCannotTouchExternalFile(t *testing.T) {
	spoolFixture(t) // establishes the temporary MOAI_HOME
	external := filepath.Join(os.Getenv("MOAI_HOME"), "state", "review-target.claim")
	if err := os.MkdirAll(filepath.Dir(external), 0o700); err != nil {
		t.Fatalf("mkdir external dir: %v", err)
	}
	if err := os.WriteFile(external, []byte("not-a-lock: preserved evidence\n"), 0o600); err != nil {
		t.Fatalf("seed external file: %v", err)
	}

	if _, err := ClaimItemSend(context.Background(), "x/../../review-target"); err == nil {
		t.Fatal("ClaimItemSend accepted a traversal-shaped id")
	}

	raw, err := os.ReadFile(external)
	if err != nil || string(raw) != "not-a-lock: preserved evidence\n" {
		t.Fatalf("the external file was disturbed by the traversal claim (err=%v, content=%q)", err, raw)
	}
}
