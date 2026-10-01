// Card t529 — the worktree-isolation guard's refusal must classify as its own
// error category instead of falling through to the UnknownFailure catch-all.
//
// Provenance of the samples below. Three were captured from the guard itself,
// in worktree .claude/worktrees/t529, on 2026-09-12 — none is guessed:
//
//   - sampleGuardRefusalDashC  — main session, 10:53:59Z
//   - sampleGuardRefusalGitDir — background subagent, 11:00:13Z
//   - sampleGuardRefusalComplex — main session, 10:54:13Z
//
// The fourth shape (the working-directory-resolution variant) is the one card
// t529 was filed from. It is quoted from the card, NOT captured here, and the
// card quotes only its first clause — the continuation is unobserved. That
// asymmetry is load-bearing and drives the matcher's design; see
// TestClassifyError_GuardRefusal_AnchorOnly.
package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness"
)

const (
	// Measured — main session, cross-tree redirect via -C.
	sampleGuardRefusalDashC = "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529, but this command redirects git to the shared checkout via -C. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Run the equivalent from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529 without the redirect."

	// Measured — background subagent, cross-tree redirect via --git-dir. This
	// sample is why the card's premise ("the blockage is silent") is false: the
	// subagent's refused Bash reached PostToolUseFailure exactly as a main
	// session's does.
	sampleGuardRefusalGitDir = "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529, but this command redirects git to the shared checkout via --git-dir. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Run the equivalent from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529 without the redirect."

	// Measured — main session, compound command the guard could not verify.
	sampleGuardRefusalComplex = "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529, but this command is too complex to verify that it stays inside the worktree. Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Split it into plain, separate commands and run them from /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529."

	// Card-quoted, NOT measured here — the working-directory-resolution variant
	// that card t529 was filed from. Reproduced at the card's own truncation
	// point, deliberately: pretending to know the continuation would be an
	// unobserved claim.
	sampleGuardRefusalCwdCardQuoted = "This session is isolated in the worktree .claude/worktrees/t526, but this command's working directory resolved to the shared checkout (.claude/worktrees/t508)"
)

// TestClassifyError_GuardRefusal covers every guard refusal shape known to the
// card. Mutant killed: the WorktreeGuardRefusal category missing altogether, so
// a refusal lands in the UnknownFailure catch-all and cannot be selected by key
// — the state measured on 2026-09-12, where all 2730 historical refusal rows in
// .moai/lessons-inbox.jsonl carry event_key tool_failure:Bash:UnknownFailure.
func TestClassifyError_GuardRefusal(t *testing.T) {
	t.Parallel()

	samples := map[string]string{
		"redirect via -C":        sampleGuardRefusalDashC,
		"redirect via --git-dir": sampleGuardRefusalGitDir,
		"command too complex":    sampleGuardRefusalComplex,
		"cwd resolved (card)":    sampleGuardRefusalCwdCardQuoted,
	}

	h := &postToolUseFailureHandler{}
	for name, text := range samples {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := h.classifyError(&HookInput{
				ToolName:      "Bash",
				Error:         text,
				HookEventName: "PostToolUseFailure",
			})
			if got != WorktreeGuardRefusal {
				t.Errorf("classifyError() = %q, want %q", got, WorktreeGuardRefusal)
			}
		})
	}
}

// TestClassifyError_GuardRefusal_BeatsOOM pins the matcher's POSITION, not just
// its existence. Mutant killed: the guard branch placed anywhere after the OOM
// branch. The OOM matcher tests for the bare substring "137", and this
// repository names worktrees after card ids — a refusal raised inside
// .claude/worktrees/t137 therefore classifies as OOMKilled under that mutant.
// The failure is silent: an operator reading the key sees a memory kill that
// never happened.
func TestClassifyError_GuardRefusal_BeatsOOM(t *testing.T) {
	t.Parallel()

	text := strings.ReplaceAll(sampleGuardRefusalDashC, "t529", "t137")
	if !strings.Contains(text, "137") {
		t.Fatal("fixture lost its OOM-colliding substring; the mutant it kills is no longer reachable")
	}

	h := &postToolUseFailureHandler{}
	if got := h.classifyError(&HookInput{
		ToolName:      "Bash",
		Error:         text,
		HookEventName: "PostToolUseFailure",
	}); got != WorktreeGuardRefusal {
		t.Errorf("classifyError() = %q, want %q — the guard branch is running after the OOM branch", got, WorktreeGuardRefusal)
	}
}

// TestClassifyError_GuardRefusal_AnchorOnly pins WHICH token the matcher stands
// on, and is the recorded home of this detection's upstream dependency.
//
// The matcher anchors on "isolated in the worktree" ALONE. The tempting
// alternative — also requiring the "Refusing to run it" continuation — is
// rejected here on purpose: the card's own variant is quoted only to its first
// clause, so requiring the continuation would risk missing the exact scenario
// the card was filed from, and no capture exists to prove otherwise.
//
// Two mutants killed:
//   - the anchor widened to a generic token ("refusing to run"), which any
//     unrelated tool refusal can carry;
//   - the anchor dropped in favour of the continuation alone, which the
//     card-quoted variant may not contain.
func TestClassifyError_GuardRefusal_AnchorOnly(t *testing.T) {
	t.Parallel()

	h := &postToolUseFailureHandler{}

	// Carries the continuation but NOT the anchor — must not classify.
	generic := "Refusing to run it — the command was rejected by a policy check."
	if got := h.classifyError(&HookInput{
		ToolName:      "Bash",
		Error:         generic,
		HookEventName: "PostToolUseFailure",
	}); got == WorktreeGuardRefusal {
		t.Error("classifyError() classified a generic refusal as a worktree-guard refusal — the matcher is anchored on the wrong token")
	}

	// Carries the anchor and nothing else — must classify. This is the clause
	// the card-quoted variant is known to contain.
	anchorOnly := "This session is isolated in the worktree /tmp/wt, but this command's working directory resolved elsewhere"
	if got := h.classifyError(&HookInput{
		ToolName:      "Bash",
		Error:         anchorOnly,
		HookEventName: "PostToolUseFailure",
	}); got != WorktreeGuardRefusal {
		t.Errorf("classifyError() = %q, want %q — the anchor alone must be sufficient", got, WorktreeGuardRefusal)
	}
}

// TestFormatMessage_GuardRefusal_NamesTheGap covers the half of card t529 that
// the category alone does not reach. The hazard is not the refusal; it is an
// audit that degrades from measuring to source-reading and still reports PASS.
// The message an agent actually sees must therefore say that nothing ran and
// that a verification resting on it is a gap — not a pass.
//
// Mutant killed: a message that names the category but leaves the evidentiary
// consequence unstated, so a degraded auditor reads it as a retryable hiccup.
func TestFormatMessage_GuardRefusal_NamesTheGap(t *testing.T) {
	t.Parallel()

	h := &postToolUseFailureHandler{}
	msg := h.formatMessage(WorktreeGuardRefusal, &HookInput{
		ToolName:      "Bash",
		Error:         sampleGuardRefusalDashC,
		HookEventName: "PostToolUseFailure",
	})

	if !strings.HasPrefix(msg, string(WorktreeGuardRefusal)+":") {
		t.Errorf("formatMessage() = %q, want the category prefix", msg)
	}
	lower := strings.ToLower(msg)
	for _, want := range []string{"did not run", "gap", "not a pass"} {
		if !strings.Contains(lower, want) {
			t.Errorf("formatMessage() omits %q — the evidentiary consequence is unstated: %q", want, msg)
		}
	}
}

// ─── SPEC-SESSION-ANCHOR-ATTR-001 W1 — refusal attribution (REQ-SAA-001/002) ───
//
// The t1064 defect class: 4,358 WorktreeGuardRefusal rows carry only
// subject:Bash + context_hash — no session_id, no cwd, no tree path — so the
// 2026-09-29 refusals (t1337 vs t1339 vs legitimate merge-window refusals)
// cannot be separated from the log. These tests pin the attribution fields the
// failure observer must add to guard-refusal rows.

// TestGuardRefusalWorktreePath covers the tree-path extraction from the
// refusal text for every sample shape known to the card. The path is the
// segment the runtime quotes right after the anchor, up to the ", " separator
// or end of line. The card-quoted variant carries a relative path — extracted
// as quoted, not absolutized.
func TestGuardRefusalWorktreePath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		text string
		want string
	}{
		{"redirect via -C", sampleGuardRefusalDashC, "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529"},
		{"redirect via --git-dir", sampleGuardRefusalGitDir, "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529"},
		{"command too complex", sampleGuardRefusalComplex, "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529"},
		{"cwd resolved (card, relative path)", sampleGuardRefusalCwdCardQuoted, ".claude/worktrees/t526"},
		{"anchor only, no path quoted", "This session is isolated in the worktree", ""},
		{"anchor mid-line, path at line end", "prefix\nThis session is isolated in the worktree /tmp/wt\nsuffix", "/tmp/wt"},
		{"no anchor at all", "unrelated error text", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := guardRefusalWorktreePath(c.text); got != c.want {
				t.Errorf("guardRefusalWorktreePath() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestClassifyError_GuardRefusal_UnknownMarkers covers REQ-SAA-002: when a
// session identifier, the cwd, or the quoted tree path cannot be resolved at
// refusal-record time, the row carries explicit "unknown" markers and is still
// written — never dropped or deferred.
//
// The pure half asserts the marker substitution; the wire half asserts the
// shall-not-drop clause end to end through the PostToolUseFailure handler.
func TestClassifyError_GuardRefusal_UnknownMarkers(t *testing.T) {
	// Pure half: empty session id, empty cwd, and an anchor-only error text
	// (no path quoted) all resolve to explicit unknown markers.
	sid, cwd, treePath := guardRefusalAttribution("", "", "This session is isolated in the worktree")
	if sid != "unknown" || cwd != "unknown" || treePath != "unknown" {
		t.Errorf("guardRefusalAttribution() = (%q, %q, %q), want explicit unknown markers on all three", sid, cwd, treePath)
	}

	// Wire half: unresolvable session id — the row is still recorded.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	input := &HookInput{
		SessionID:     "", // unresolvable — REQ-SAA-002
		ToolName:      "Bash",
		Error:         sampleGuardRefusalCwdCardQuoted,
		HookEventName: "PostToolUseFailure",
	}
	h := NewPostToolUseFailureHandler()
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	events := readUsageLog(t, root)
	found := findEvent(events, harness.EventTypeToolFailure)
	if found == nil {
		t.Fatal("guard refusal row was dropped — REQ-SAA-002 forbids dropping an unattributable row")
	}
	if found.SessionID != "unknown" {
		t.Errorf("session_id = %q, want explicit \"unknown\" marker", found.SessionID)
	}
	// The cwd was resolvable here (os.Getwd fallback), so it must carry the
	// resolved value, not the marker — the marker is only for the unresolvable
	// case (asserted on the pure half above).
	if found.Cwd == "" || found.Cwd == "unknown" {
		t.Errorf("cwd = %q, want the resolved working directory", found.Cwd)
	}
	if got := found.WorktreePath; got != ".claude/worktrees/t526" {
		t.Errorf("worktree_path = %q, want the rejection-quoted path %q", got, ".claude/worktrees/t526")
	}
}

// TestRecordToolFailureEvent_GuardRefusalRowFields covers REQ-SAA-001: a
// WorktreeGuardRefusal row carries session_id, the resolved cwd, and the
// rejection-quoted worktree path in addition to the existing subject and
// context_hash fields. A non-guard failure row keeps the pre-existing shape
// (no attribution fields) — the requirement is scoped to guard refusals.
func TestRecordToolFailureEvent_GuardRefusalRowFields(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)

	guardInput := &HookInput{
		SessionID:     "sess-guard-001",
		ToolName:      "Bash",
		Error:         sampleGuardRefusalDashC,
		HookEventName: "PostToolUseFailure",
	}
	h := NewPostToolUseFailureHandler()
	if _, err := h.Handle(context.Background(), guardInput); err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	otherInput := &HookInput{
		SessionID:     "sess-other-001",
		ToolName:      "Bash",
		Error:         "exit status 1",
		HookEventName: "PostToolUseFailure",
	}
	if _, err := h.Handle(context.Background(), otherInput); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	events := readUsageLog(t, root)
	var guard, other *harness.Event
	for i := range events {
		switch {
		case events[i].ContextHash == string(WorktreeGuardRefusal):
			guard = &events[i]
		case events[i].ContextHash == string(ExitError):
			other = &events[i]
		}
	}
	if guard == nil {
		t.Fatalf("no WorktreeGuardRefusal row recorded; got %d events", len(events))
	}
	if guard.SessionID != "sess-guard-001" {
		t.Errorf("guard row session_id = %q, want %q", guard.SessionID, "sess-guard-001")
	}
	if guard.Cwd == "" || guard.Cwd == "unknown" {
		t.Errorf("guard row cwd = %q, want the resolved working directory", guard.Cwd)
	}
	if want := "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t529"; guard.WorktreePath != want {
		t.Errorf("guard row worktree_path = %q, want %q", guard.WorktreePath, want)
	}
	if guard.Subject != "Bash" || guard.ContextHash != string(WorktreeGuardRefusal) {
		t.Errorf("guard row subject/context_hash = %q/%q — the pre-existing fields must survive alongside the new ones", guard.Subject, guard.ContextHash)
	}

	if other == nil {
		t.Fatalf("no ExitError row recorded; got %d events", len(events))
	}
	if other.SessionID != "" || other.Cwd != "" || other.WorktreePath != "" {
		t.Errorf("non-guard row grew attribution fields (session_id=%q cwd=%q worktree_path=%q) — REQ-SAA-001 is scoped to guard refusals",
			other.SessionID, other.Cwd, other.WorktreePath)
	}
}
