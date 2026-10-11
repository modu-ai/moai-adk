package cli

// codex_review_entry_delivery_test.go — SPEC-REVIEW-GATE-REPAIR-001 M2
// (REQ-RG-001, REQ-RG-002; AC-RG-001, AC-RG-002, AC-RG-003). The UserPromptSubmit
// entry hook never blocks a prompt. A fresh FAIL for the current review key is
// delivered as additional context exactly once per key, and the delivery record
// keeps later prompts of the same key from carrying the detail again. The
// assertions read the observable hook output and the verify state root only;
// they do not name the delivery record's file, so its format stays a run-phase
// choice.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// entryDeliveryDetail is the finding text the fixtures preserve with the fail
// receipt. The tests count its occurrences in the delivered context.
const entryDeliveryDetail = "blocking finding on shared file."

// entryDeliveryRecordFail records a fresh FAIL receipt for the fixture's current
// tree key and preserves the finding text as the detail file, as the background
// producer does at record time.
func entryDeliveryRecordFail(t *testing.T, root, finding string) {
	t.Helper()
	recordEntryReceipt(t, root, codexReviewVerdictFail, 1, "")
	codexReviewPreserveFindings(root, mustEntryState(t, root), ReviewOutput{
		Verdict: "fail",
		Summary: "reviewer failed the turn",
		Findings: []Finding{{
			Severity: "P1", Title: "delayed finding",
			Body: finding, File: "main.go", Line: 3,
		}},
	})
}

// entryDeliveryMoveTree writes an untracked file so the tree key moves, the way
// an edit between two turns does.
func entryDeliveryMoveTree(t *testing.T, root, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte("moved\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// entryDeliveryEntry runs the entry hook over the fixture root and fails the
// test on error. Every entry-hook path is fail-open, so an error is a defect.
func entryDeliveryEntry(t *testing.T, root string) *hook.HookOutput {
	t.Helper()
	out, err := HandleCodexReviewEntry(&hook.HookInput{CWD: root}, true, root)
	if err != nil {
		t.Fatalf("entry hook must not error; got %v", err)
	}
	if out == nil {
		t.Fatal("entry hook must return an output, got nil")
	}
	return out
}

// entryDeliveryContext returns the additional context the entry hook delivered,
// or "" when it delivered none.
func entryDeliveryContext(out *hook.HookOutput) string {
	if out == nil || out.HookSpecificOutput == nil {
		return ""
	}
	return out.HookSpecificOutput.AdditionalContext
}

// entryDeliveryState maps every regular file under the verify state root to its
// content. A delivery record shows up as a new or changed entry wherever the
// implementation writes it.
func entryDeliveryState(t *testing.T, root string) map[string]string {
	t.Helper()
	base := filepath.Join(root, ".moai", "state", "verify")
	files := map[string]string{}
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return nil
		}
		files[rel] = string(b)
		return nil
	})
	return files
}

// entryDeliveryChanged reports whether the state root gained or changed a file.
func entryDeliveryChanged(before, after map[string]string) bool {
	if len(before) != len(after) {
		return true
	}
	for name, content := range after {
		if prev, ok := before[name]; !ok || prev != content {
			return true
		}
	}
	return false
}

// TestReviewEntry_FreshFailNeverBlocks — AC-RG-001 (REQ-RG-001): a fresh FAIL
// receipt for the current key does not block the prompt at turn entry.
func TestReviewEntry_FreshFailNeverBlocks(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	entryDeliveryRecordFail(t, root, entryDeliveryDetail)

	out := entryDeliveryEntry(t, root)
	if out.Decision == hook.DecisionBlock {
		t.Fatalf("AC-RG-001: a fresh FAIL receipt must not block the prompt; got decision=block with reason=%q", out.Reason)
	}
}

// TestReviewEntry_FirstPromptDeliversDetailOnce — AC-RG-002 (REQ-RG-002): the
// first prompt for a fresh FAIL key carries the finding detail exactly once as
// additional context, and the delivery leaves a record under the verify state
// root.
func TestReviewEntry_FirstPromptDeliversDetailOnce(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	entryDeliveryRecordFail(t, root, entryDeliveryDetail)
	before := entryDeliveryState(t, root)

	out := entryDeliveryEntry(t, root)
	ctx := entryDeliveryContext(out)
	if n := strings.Count(ctx, entryDeliveryDetail); n != 1 {
		t.Fatalf("AC-RG-002: the first prompt must carry the finding detail exactly once as additional context; got %d occurrence(s) in additionalContext=%q", n, ctx)
	}
	if !entryDeliveryChanged(before, entryDeliveryState(t, root)) {
		t.Fatal("AC-RG-002: the delivery must leave a per-key delivery record under the verify state root; the state root is unchanged")
	}
	if out.Decision == hook.DecisionBlock {
		t.Fatal("AC-RG-002: the delivery must not block the prompt")
	}
}

// TestReviewEntry_RepeatPromptOmitsDetail — AC-RG-003 (REQ-RG-002): once a key
// has a delivery record, the next prompt of the same key does not carry the
// detail again, in additional context or in any other output field.
func TestReviewEntry_RepeatPromptOmitsDetail(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	entryDeliveryRecordFail(t, root, entryDeliveryDetail)
	_ = entryDeliveryEntry(t, root) // the first prompt: the delivery

	// Premise (not a behavior assertion): the receipt is still fresh for the
	// same key after the first prompt. Otherwise the silence below could be a
	// moved key instead of a suppressed repeat, and the test would pass vacuously.
	state := mustEntryState(t, root)
	if chk := verify.CheckReceipt(verify.LoadReceipt(root, state), state, time.Now(), codexReviewCacheTTL); !chk.Run {
		t.Fatalf("premise: the fail receipt must stay fresh for the same key after the first prompt; %s", chk.Reason)
	}

	out := entryDeliveryEntry(t, root)
	if strings.Contains(out.Reason, entryDeliveryDetail) || strings.Contains(entryDeliveryContext(out), entryDeliveryDetail) {
		t.Fatalf("AC-RG-003: the second prompt of a delivered key must not repeat the detail; reason=%q additionalContext=%q", out.Reason, entryDeliveryContext(out))
	}
	if out.Decision == hook.DecisionBlock {
		t.Fatal("AC-RG-003: the second prompt of a delivered key must not be blocked")
	}
}

// TestReviewEntry_NewKeyDeliversAgain — REQ-RG-002: a changed review key has its
// own delivery record, so the detail of the new key is delivered once even after
// the old key was delivered.
func TestReviewEntry_NewKeyDeliversAgain(t *testing.T) {
	const firstDetail = "first key finding."
	const secondDetail = "second key finding."
	root := cacheTestRoot(t)
	entrySeams(t)
	entryDeliveryRecordFail(t, root, firstDetail)
	_ = entryDeliveryEntry(t, root) // delivers the first key

	entryDeliveryMoveTree(t, root, "moved.txt")
	entryDeliveryRecordFail(t, root, secondDetail)

	out := entryDeliveryEntry(t, root)
	ctx := entryDeliveryContext(out)
	if n := strings.Count(ctx, secondDetail); n != 1 {
		t.Fatalf("REQ-RG-002: a changed review key must deliver its own detail exactly once as additional context; got %d occurrence(s) in additionalContext=%q", n, ctx)
	}
	if strings.Contains(ctx, firstDetail) {
		t.Fatalf("REQ-RG-002: the old key's detail must not ride along with the new key; additionalContext=%q", ctx)
	}
}

// TestReviewEntry_InconclusiveAllowsWithoutBlock — REQ-RG-001 covers every
// receipt state, fresh inconclusive included. Guard: the entry hook allows a
// fresh inconclusive receipt. This guard passes before M2 as well; it is not a
// RED test.
func TestReviewEntry_InconclusiveAllowsWithoutBlock(t *testing.T) {
	root := cacheTestRoot(t)
	entrySeams(t)
	recordEntryReceipt(t, root, codexReviewVerdictInconclusive, 2, "")

	out := entryDeliveryEntry(t, root)
	if out.Decision == hook.DecisionBlock {
		t.Fatalf("REQ-RG-001: a fresh inconclusive receipt must not block the prompt; got reason=%q", out.Reason)
	}
}
