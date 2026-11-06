// codex_review_state_root_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632,
// REQ-MN-025): the receipt producer and the Codex Stop chain's receipt reader
// resolve the SAME root — the git toplevel of the session tree — so a session
// started in a subdirectory never reads "receipt absent" for a receipt that
// was written.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestCodexReviewStateRootAgreesWithProducer pins the seam: a receipt the
// producer recorded from a session in a SUBDIRECTORY lands under the git
// toplevel, and the root the Stop chain reader resolves for the same session
// is that same toplevel — the reader finds what the producer wrote.
func TestCodexReviewStateRootAgreesWithProducer(t *testing.T) {
	repo := t.TempDir()
	cardScopeGit(t, repo, "init", "-q", "-b", "main")
	writeCardFile(t, repo, "a.go", "package a\n")
	writeCardFile(t, repo, ".moai/state/keep.json", "{}\n")
	cardScopeGit(t, repo, "add", "-A")
	cardScopeGit(t, repo, "commit", "-q", "-m", "init")
	cardScopeGit(t, repo, "config", "user.email", "fx@example.com")
	cardScopeGit(t, repo, "config", "user.name", "fx")
	writeCardFile(t, repo, "changed.go", "package a\n\n// changed\n")

	// The session runs in a subdirectory of the repository.
	nested := filepath.Join(repo, "cmd", "sub")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	withCodexSessionArgs(t, codexSessionScript("clean change, no findings"))
	if _, err := produceCodexReviewReceipt(context.Background(), nested); err != nil {
		t.Fatalf("produceCodexReviewReceipt from the nested session tree: %v", err)
	}

	// The reader's root for the same session tree. git resolves the macOS
	// /var symlink to /private/var, so the comparison is on resolved paths —
	// the identity the producer and the reader share is what REQ-MN-025 pins.
	repoCanon, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	readerRoot := codexReviewStateRoot(nested)
	if readerRoot != repoCanon {
		t.Fatalf("codexReviewStateRoot(%q) = %q, want the git toplevel %q (REQ-MN-025: one root for producer and reader)", nested, readerRoot, repoCanon)
	}
	// The producer's receipt is findable under the reader's root.
	toplevel := filepath.Join(repo, ".moai", "state", "verify")
	entries, err := os.ReadDir(toplevel)
	if err != nil || len(entries) == 0 {
		t.Fatalf("no receipt snapshots under %s (err=%v) — the producer wrote somewhere else than the root the reader resolves", toplevel, err)
	}
}
