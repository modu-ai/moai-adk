package hook

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
	"github.com/modu-ai/moai-adk/internal/config"
)

// swgHandlerWithConfig builds a preToolHandler wired to a config with the
// guard's deny layer set to enabled, rooted at projectDir.
func swgHandlerWithConfig(enabled bool, projectDir string) *preToolHandler {
	cfg := config.NewDefaultConfig()
	cfg.Workflow.SubagentWriteGuard.Enabled = enabled
	return &preToolHandler{
		cfg:        &auditConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}
}

// Tests for the subagent destructive-write guard
// (SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001). Every test names its acceptance
// criterion. The guard decides solely from the parsed payload fields
// (REQ-SWG-001); a deny requires positive evidence on all four predicate
// conditions (REQ-SWG-004), evaluated in the stated order.

// swgSetupTrackedRepo creates a temp git repository with one tracked file of
// the given byte size and returns the repo path and the file's absolute path.
func swgSetupTrackedRepo(t *testing.T, existingBytes int) (repo, filePath string) {
	t.Helper()
	requireGit(t)
	repo = t.TempDir()
	gitInitRepo(t, repo)
	filePath = filepath.Join(repo, "tracked.txt")
	content := strings.Repeat("a", existingBytes)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write tracked file: %v", err)
	}
	mustRunGit(t, repo, "add", "tracked.txt")
	mustRunGit(t, repo, "commit", "-m", "add tracked file")
	return repo, filePath
}

// swgPayload builds a Write-tool HookInput carrying the given structured
// fields, the form measured in evidence-probe-hook-payloads.jsonl record 2.
func swgPayload(agentID, agentType, filePath, content string) *HookInput {
	toolInput := map[string]any{
		"file_path": filePath,
		"content":   content,
	}
	raw, err := json.Marshal(toolInput)
	if err != nil {
		panic(err)
	}
	return &HookInput{
		ToolName:  "Write",
		ToolInput: raw,
		AgentID:   agentID,
		AgentType: agentType,
		CWD:       "",
	}
}

// AC-SWG-001a — a destructive out-of-scope write is refused: a subagent Write
// replacing a tracked 20000-byte file with 300 bytes (1.5%, inside SWG-T2 and
// above SWG-T1) must yield a deny whose reason begins with the sentinel and
// names the target path and both byte sizes.
func TestSubagentWriteGuardDestructiveWriteDenied(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)

	if ev.Decision != "deny" {
		t.Fatalf("decision = %q, want deny (reason %q)", ev.Decision, ev.Reason)
	}
	if !strings.HasPrefix(ev.Reason, "SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION:") {
		t.Fatalf("reason %q does not begin with the sentinel", ev.Reason)
	}
	if !strings.Contains(ev.Reason, filePath) {
		t.Fatalf("reason %q does not name the target path", ev.Reason)
	}
	if !strings.Contains(ev.Reason, "20000") || !strings.Contains(ev.Reason, "300") {
		t.Fatalf("reason %q does not name both byte sizes", ev.Reason)
	}
}

// AC-SWG-001b — a normal write still passes: the identical payload with
// 20600-byte content (a growth, outside SWG-T2) reaches the allow
// fall-through and emits no deny.
func TestSubagentWriteGuardNormalWriteAllowed(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 20600))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)

	if ev.Decision != "allow" {
		t.Fatalf("decision = %q, want allow (reason %q)", ev.Decision, ev.Reason)
	}
}

// AC-SWG-009 — the size floor holds: a rewrite of a tracked 1500-byte file
// (below SWG-T1) with a 10-byte stub passes even though the ratio would hold.
func TestSubagentWriteGuardSizeFloorHolds(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 1500)

	input := swgPayload("agent-abc123", "general-purpose", filePath, "stub")
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)

	if ev.Decision != "allow" {
		t.Fatalf("decision = %q, want allow below the size floor (reason %q)", ev.Decision, ev.Reason)
	}
}

// AC-SWG-005 — both main-session forms pass and are not evaluated at all:
// a payload with an empty agent_id is outside the guard's subject whether
// agent_type is empty (plain) or carries an agent name (claude --agent).
func TestSubagentWriteGuardMainSessionNotEvaluated(t *testing.T) {
	_, filePath := swgSetupTrackedRepo(t, 20000)
	nonGit := t.TempDir() // deliberately NOT the repo: even a git-valid target must not matter

	t.Run("PlainMainSession", func(t *testing.T) {
		input := swgPayload("", "", filePath, strings.Repeat("x", 100))
		input.CWD = nonGit
		ev := evaluateSubagentWrite(input, true)
		if ev.Decision != "" {
			t.Fatalf("decision = %q, want no evaluation at all for a plain main session", ev.Decision)
		}
	})

	t.Run("AgentFlagMainSession", func(t *testing.T) {
		input := swgPayload("", "manager-git", filePath, strings.Repeat("x", 100))
		input.CWD = nonGit
		ev := evaluateSubagentWrite(input, true)
		if ev.Decision != "" {
			t.Fatalf("decision = %q, want no evaluation for a claude --agent main session (agent_type must not be the discriminant)", ev.Decision)
		}
	})
}

// AC-SWG-006 — fail-open on uncertainty: an unparseable payload, an absent
// content field, and a target outside any git repository each allow the write.
func TestSubagentWriteGuardFailOpenOnUncertainty(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)
	nonGit := t.TempDir()

	t.Run("UnparseablePayload", func(t *testing.T) {
		input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 100))
		input.ToolInput = []byte("{not json")
		input.CWD = repo
		ev := evaluateSubagentWrite(input, true)
		if ev.Decision != "fail-open" {
			t.Fatalf("decision = %q, want fail-open (reason %q)", ev.Decision, ev.Reason)
		}
		if ev.Reason == "" {
			t.Fatalf("fail-open row must carry its reason")
		}
	})

	t.Run("ContentFieldAbsent", func(t *testing.T) {
		raw, err := json.Marshal(map[string]any{"file_path": filePath})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		input := swgPayload("agent-abc123", "general-purpose", filePath, "ignored")
		input.ToolInput = raw
		input.CWD = repo
		ev := evaluateSubagentWrite(input, true)
		if ev.Decision != "fail-open" {
			t.Fatalf("decision = %q, want fail-open when content is absent (reason %q)", ev.Decision, ev.Reason)
		}
	})

	t.Run("TargetOutsideGitRepository", func(t *testing.T) {
		outside := filepath.Join(nonGit, "tracked.txt")
		if err := os.WriteFile(outside, []byte(strings.Repeat("a", 20000)), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		input := swgPayload("agent-abc123", "general-purpose", outside, strings.Repeat("x", 100))
		input.CWD = nonGit
		ev := evaluateSubagentWrite(input, true)
		if ev.Decision != "fail-open" {
			t.Fatalf("decision = %q, want fail-open when the target is not in a git repository (reason %q)", ev.Decision, ev.Reason)
		}
	})
}

// AC-SWG-003 — an untracked target passes: a file that exists in the working
// tree but is NOT tracked at HEAD is allowed even when every size condition
// holds. This is the criterion that keeps a SPEC-authoring subagent working
// on artifacts it created earlier in its own run (REQ-SWG-010).
func TestSubagentWriteGuardUntrackedTargetPasses(t *testing.T) {
	repo, _ := swgSetupTrackedRepo(t, 20000)
	untracked := filepath.Join(repo, "fresh.txt")
	if err := os.WriteFile(untracked, []byte(strings.Repeat("a", 20000)), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	input := swgPayload("agent-abc123", "general-purpose", untracked, strings.Repeat("x", 100))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != "allow" {
		t.Fatalf("decision = %q, want allow for an untracked target (reason %q)", ev.Decision, ev.Reason)
	}
}

// AC-SWG-004 — an absolute file_path expressed through a symlinked directory
// is resolved and matched, not prefix-compared. The precedent's Defect A is a
// relative `strings.HasPrefix` against an absolute payload path; the mutant
// form of this criterion replaces resolution with that prefix match and must
// go red.
func TestSubagentWriteGuardSymlinkedAbsolutePathMatches(t *testing.T) {
	repo, _ := swgSetupTrackedRepo(t, 20000)
	linkDir := filepath.Join(filepath.Dir(repo), "swg-link")
	if err := os.Symlink(repo, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(linkDir) })
	linkedPath := filepath.Join(linkDir, "tracked.txt")

	input := swgPayload("agent-abc123", "general-purpose", linkedPath, strings.Repeat("x", 300))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != "deny" {
		t.Fatalf("decision = %q, want deny for a symlinked absolute path (reason %q)", ev.Decision, ev.Reason)
	}
}

// REQ-SWG-004's evaluation order is the deliverable, not a comment about it:
// a payload failing the ratio condition must reach NO git invocation at all.
// The positive control proves the counting mechanism itself is alive — a
// counter that never counts would make the zero assertion vacuous.
func TestSubagentWriteGuardNoGitInvocationWhenRatioFails(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)

	orig := gitcore.ExecCommand
	calls := 0
	gitcore.ExecCommand = func(name string, args ...string) *exec.Cmd {
		calls++
		return exec.Command(name, args...)
	}
	t.Cleanup(func() { gitcore.ExecCommand = orig })

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 20600))
	input.CWD = repo
	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != "allow" {
		t.Fatalf("decision = %q, want allow for a growing write (reason %q)", ev.Decision, ev.Reason)
	}
	if calls != 0 {
		t.Fatalf("git invoked %d time(s) for a payload failing the ratio condition; want 0 (REQ-SWG-004 order)", calls)
	}

	denyInput := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	denyInput.CWD = repo
	ev = evaluateSubagentWrite(denyInput, true)
	if ev.Decision != "deny" {
		t.Fatalf("positive control: decision = %q, want deny (reason %q)", ev.Decision, ev.Reason)
	}
	if calls == 0 {
		t.Fatalf("positive control: the git counter observed no invocations on the deny path — the counting mechanism is dead")
	}
}

// BenchmarkSubagentWriteGuardDenyPath measures the guard's added latency on
// the deny path (both git subprocesses included) against the 5s hook budget.
// The measurement is recorded in the run-phase evidence; the benchmark asserts
// nothing about the absolute value.
func BenchmarkSubagentWriteGuardDenyPath(b *testing.B) {
	requireGit := func() {
		if _, err := exec.LookPath("git"); err != nil {
			b.Skip("git not available")
		}
	}
	requireGit()
	repo := b.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			b.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	filePath := filepath.Join(repo, "tracked.txt")
	if err := os.WriteFile(filePath, []byte(strings.Repeat("a", 20000)), 0o644); err != nil {
		b.Fatalf("write: %v", err)
	}
	run("add", "tracked.txt")
	run("-c", "user.email=t@local", "-c", "user.name=t", "commit", "-m", "seed")

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	input.CWD = repo

	b.ResetTimer()
	start := time.Now()
	for i := 0; i < b.N; i++ {
		if ev := evaluateSubagentWrite(input, true); ev.Decision != "deny" {
			b.Fatalf("decision = %q, want deny", ev.Decision)
		}
	}
	elapsed := time.Since(start)
	b.ReportMetric(float64(elapsed.Microseconds())/float64(b.N), "µs/op")
}

// M4 wiring — the enabled deny layer returns the deny reason from the handler
// method, so the PreToolUse call site can refuse the write.
func TestSubagentWriteGuardHandlerEnabledDenies(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)
	h := swgHandlerWithConfig(true, repo)

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	input.CWD = repo

	reason := h.checkSubagentDestructiveWrite(input)
	if !strings.HasPrefix(reason, "SUBAGENT_DESTRUCTIVE_WRITE_VIOLATION:") {
		t.Fatalf("reason = %q, want a deny reason with the sentinel", reason)
	}
}

// AC-SWG-002 — the disabled path emits no deny. Paired with AC-SWG-012
// (M5): a guard that went fully inert when disabled would pass this while
// violating the family contract — the audit row test joins the pair.
func TestSubagentWriteGuardHandlerDisabledAllows(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)
	h := swgHandlerWithConfig(false, repo)

	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	input.CWD = repo

	if reason := h.checkSubagentDestructiveWrite(input); reason != "" {
		t.Fatalf("reason = %q, want no deny with the deny layer disabled", reason)
	}
}

// nil and empty ConfigProvider must read as disabled — the family shape.
func TestSubagentWriteGuardHandlerNilConfigFailClosed(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)

	t.Run("NilProvider", func(t *testing.T) {
		h := &preToolHandler{cfg: nil, policy: DefaultSecurityPolicy(), projectDir: repo}
		input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
		input.CWD = repo
		if reason := h.checkSubagentDestructiveWrite(input); reason != "" {
			t.Fatalf("reason = %q, want no deny with a nil ConfigProvider", reason)
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		h := &preToolHandler{cfg: &auditConfigProvider{cfg: nil}, policy: DefaultSecurityPolicy(), projectDir: repo}
		input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
		input.CWD = repo
		if reason := h.checkSubagentDestructiveWrite(input); reason != "" {
			t.Fatalf("reason = %q, want no deny with a nil config", reason)
		}
	})
}
