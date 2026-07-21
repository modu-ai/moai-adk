package hook

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/handoff"
)

// --- helpers -----------------------------------------------------------------

func autoCfgProvider(guide bool) *mockConfigProvider {
	return &mockConfigProvider{cfg: &config.Config{Handoff: config.HandoffConfig{Mode: "auto", Guide: guide}}}
}

func manualCfgProvider() *mockConfigProvider {
	return &mockConfigProvider{cfg: &config.Config{Handoff: config.HandoffConfig{Mode: "manual"}}}
}

func mustSavePending(t *testing.T, pd string, rec *handoff.PendingRecord) {
	t.Helper()
	if err := handoff.SavePending(pd, rec); err != nil {
		t.Fatalf("SavePending: %v", err)
	}
}

func livePending(body string) *handoff.PendingRecord {
	return &handoff.PendingRecord{Body: body, ConversationLanguage: "en", SavedAt: time.Now()}
}

func stalePending(body string) *handoff.PendingRecord {
	return &handoff.PendingRecord{Body: body, ConversationLanguage: "en", SavedAt: time.Now().Add(-10 * 24 * time.Hour)}
}

func injectInput(source, projectDir string) *HookInput {
	return &HookInput{Source: source, CWD: projectDir, ProjectDir: projectDir, SessionID: "sess-a1b2c3d4"}
}

func additionalContextOf(out *HookOutput) string {
	if out == nil || out.HookSpecificOutput == nil {
		return ""
	}
	return out.HookSpecificOutput.AdditionalContext
}

func pendingExists(pd string) bool {
	_, err := os.Stat(handoff.PendingPath(pd))
	return err == nil
}

func consumedNames(t *testing.T, pd string) []string {
	t.Helper()
	entries, err := os.ReadDir(handoff.ConsumedDir(pd))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read consumed dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// captureStderr runs fn with os.Stderr redirected to a pipe and returns what was
// written. NOT parallel-safe (mutates global os.Stderr) — callers must not
// t.Parallel().
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = old
	data, _ := io.ReadAll(r)
	return string(data)
}

// --- AC-008: branch table (auto mode, 4 sources) -----------------------------

func TestBranchTable_AutoMode(t *testing.T) {
	t.Parallel()

	sources := []string{"startup", "resume", "clear", "compact"}
	for _, src := range sources {
		src := src
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			pd := t.TempDir()
			mustSavePending(t, pd, livePending("resume for "+src))

			h := NewHandoffInjectHandler(autoCfgProvider(false))
			out, err := h.Handle(context.Background(), injectInput(src, pd))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}

			if src == "clear" {
				if additionalContextOf(out) == "" {
					t.Error("clear: expected additionalContext injection, got empty")
				}
				if pendingExists(pd) {
					t.Error("clear: pending.json should be consumed (renamed away)")
				}
				if n := len(consumedNames(t, pd)); n != 1 {
					t.Errorf("clear: expected 1 consumed file, got %d", n)
				}
			} else {
				if additionalContextOf(out) != "" {
					t.Errorf("%s: expected no injection, got %q", src, additionalContextOf(out))
				}
				if !pendingExists(pd) {
					t.Errorf("%s: pending.json should be preserved (not consumed)", src)
				}
				if n := len(consumedNames(t, pd)); n != 0 {
					t.Errorf("%s: expected 0 consumed files, got %d", src, n)
				}
			}
		})
	}
}

// --- AC-009: manual mode pure no-op (4 sources + stale) ----------------------

func TestManualMode_NoOp(t *testing.T) {
	t.Parallel()

	for _, src := range []string{"startup", "resume", "clear", "compact"} {
		src := src
		t.Run(src, func(t *testing.T) {
			t.Parallel()
			pd := t.TempDir()
			mustSavePending(t, pd, livePending("resume"))
			before, err := os.ReadFile(handoff.PendingPath(pd))
			if err != nil {
				t.Fatalf("read pending before: %v", err)
			}

			h := NewHandoffInjectHandler(manualCfgProvider())
			out, err := h.Handle(context.Background(), injectInput(src, pd))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			if additionalContextOf(out) != "" {
				t.Errorf("%s: manual mode must not inject, got %q", src, additionalContextOf(out))
			}
			after, err := os.ReadFile(handoff.PendingPath(pd))
			if err != nil {
				t.Fatalf("read pending after: %v", err)
			}
			if string(before) != string(after) {
				t.Errorf("%s: manual mode must leave pending.json byte-unchanged", src)
			}
		})
	}
}

func TestManualMode_StalePendingPreserved(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, stalePending("stale resume"))
	before, err := os.ReadFile(handoff.PendingPath(pd))
	if err != nil {
		t.Fatalf("read before: %v", err)
	}

	// manual mode + stale (past TTL) + source clear → still a pure no-op.
	h := NewHandoffInjectHandler(manualCfgProvider())
	if _, err := h.Handle(context.Background(), injectInput("clear", pd)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	after, err := os.ReadFile(handoff.PendingPath(pd))
	if err != nil {
		t.Fatalf("read after: manual must not remove a stale pending: %v", err)
	}
	if string(before) != string(after) {
		t.Error("manual mode removed/rewrote a stale pending — REQ-009 pure no-op violated")
	}
}

// --- AC-010: non-clear source notice-only (guide true/false) -----------------
// NOT parallel: captures os.Stderr.

func TestNonClearSource_NoticeOnly(t *testing.T) {
	for _, src := range []string{"startup", "resume", "compact"} {
		for _, guide := range []bool{true, false} {
			src, guide := src, guide
			name := src
			if guide {
				name += "_guide"
			} else {
				name += "_noguide"
			}
			t.Run(name, func(t *testing.T) {
				pd := t.TempDir()
				mustSavePending(t, pd, livePending("resume")) // LIVE (non-stale)

				h := NewHandoffInjectHandler(autoCfgProvider(guide))
				var out *HookOutput
				stderr := captureStderr(t, func() {
					o, err := h.Handle(context.Background(), injectInput(src, pd))
					if err != nil {
						t.Fatalf("Handle: %v", err)
					}
					out = o
				})

				if additionalContextOf(out) != "" {
					t.Errorf("%s: notice-only must not inject", src)
				}
				if !pendingExists(pd) {
					t.Errorf("%s: notice-only must preserve pending.json", src)
				}
				hasHint := strings.Contains(stderr, "auto-resume")
				if guide && !hasHint {
					t.Errorf("%s guide=true: expected a stderr hint, got %q", src, stderr)
				}
				if !guide && hasHint {
					t.Errorf("%s guide=false: expected NO stderr hint, got %q", src, stderr)
				}
			})
		}
	}
}

// --- AC-011: degrade-to-guidance (no xhigh claim) ----------------------------

func TestDegradeToGuidance(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	rec := &handoff.PendingRecord{
		Body:                 "ultrathink. SPEC-X run. 실행: /moai run SPEC-X",
		ConversationLanguage: "ko",
		Directives:           handoff.Directives{Ultrathink: true},
		SavedAt:              time.Now(),
	}
	mustSavePending(t, pd, rec)

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	ctx := additionalContextOf(out)
	if ctx == "" {
		t.Fatal("expected injection")
	}
	if strings.Contains(ctx, "xhigh") {
		t.Errorf("additionalContext must NOT contain 'xhigh' (verification-claim-integrity): %q", ctx)
	}
	if !strings.Contains(ctx, "ultrathink") {
		t.Error("additionalContext should carry ultrathink restoration guidance")
	}
	if !strings.Contains(ctx, "입력") {
		t.Error("additionalContext should carry manual-input restoration guidance (ko '입력')")
	}
}

// --- AC-012: claim-then-inject + audit preserve ------------------------------

func TestClaimThenInject_AuditPreserved(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume body"))
	origBytes, err := os.ReadFile(handoff.PendingPath(pd))
	if err != nil {
		t.Fatalf("read pending: %v", err)
	}

	// A memory-like audit file elsewhere must survive untouched.
	memoryFile := filepath.Join(pd, "project_x.md")
	if err := os.WriteFile(memoryFile, []byte("audit"), 0o644); err != nil {
		t.Fatalf("write memory file: %v", err)
	}

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if additionalContextOf(out) == "" {
		t.Fatal("expected injection after claim")
	}
	if pendingExists(pd) {
		t.Error("pending.json should be absent after consume")
	}
	names := consumedNames(t, pd)
	if len(names) != 1 {
		t.Fatalf("expected 1 consumed file, got %d", len(names))
	}
	consumedBytes, err := os.ReadFile(filepath.Join(handoff.ConsumedDir(pd), names[0]))
	if err != nil {
		t.Fatalf("read consumed: %v", err)
	}
	if string(consumedBytes) != string(origBytes) {
		t.Error("consumed file content should equal the original pending (rename preserves bytes)")
	}
	if _, err := os.Stat(memoryFile); err != nil {
		t.Errorf("memory audit file must not be deleted: %v", err)
	}
}

// --- AC-013a: concurrent consume single winner (-race) -----------------------

func TestConcurrentConsume_SingleWinner(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume"))

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	var injected int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := h.Handle(context.Background(), injectInput("clear", pd))
			if err != nil {
				t.Errorf("Handle: %v", err)
				return
			}
			if additionalContextOf(out) != "" {
				atomic.AddInt32(&injected, 1)
			}
		}()
	}
	wg.Wait()

	if injected != 1 {
		t.Errorf("expected exactly 1 goroutine to inject, got %d", injected)
	}
	if n := len(consumedNames(t, pd)); n != 1 {
		t.Errorf("expected exactly 1 consumed file, got %d", n)
	}
	if pendingExists(pd) {
		t.Error("pending.json should be consumed exactly once")
	}
}

// --- AC-013b: rename failure fail-open (arbitrary errno) ----------------------
// NOT parallel: mutates the handoffRenameFunc package global.

func TestRenameFailure_FailOpen(t *testing.T) {
	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume"))

	orig := handoffRenameFunc
	defer func() { handoffRenameFunc = orig }()
	// Force a NON-ENOENT error to prove the handler does not depend on os.IsNotExist.
	handoffRenameFunc = func(_, _ string) error { return errors.New("forced rename failure (EACCES-like)") }

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle must be fail-open (return nil error), got: %v", err)
	}
	if additionalContextOf(out) != "" {
		t.Error("rename failure must skip injection")
	}
	if !pendingExists(pd) {
		t.Error("rename failure must leave pending.json in place")
	}
}

// TestConcurrentConsume_SingleWinner_RenameAlwaysSucceeds reproduces, on any
// platform, the Windows failure this claim path was fixed for: two concurrent
// claims both saw their rename succeed, so both injected (AC-013a observed
// injected=2 against a single consumed file — one file object moved twice).
//
// Stubbing rename to always succeed models exactly that, and pins the invariant
// that exclusivity comes from the claim gate rather than from the loser's
// rename happening to fail. Under the previous rename-as-claim design this
// fails with injected=2.
//
// NOT parallel: mutates the handoffRenameFunc package global.
func TestConcurrentConsume_SingleWinner_RenameAlwaysSucceeds(t *testing.T) {
	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume"))

	orig := handoffRenameFunc
	defer func() { handoffRenameFunc = orig }()
	// Model the observed Windows behaviour: the move still happens, but a caller
	// whose source is already gone is told it succeeded rather than failing.
	handoffRenameFunc = func(from, to string) error {
		_ = os.Rename(from, to)
		return nil
	}

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	var injected int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := h.Handle(context.Background(), injectInput("clear", pd))
			if err != nil {
				t.Errorf("Handle: %v", err)
				return
			}
			if additionalContextOf(out) != "" {
				atomic.AddInt32(&injected, 1)
			}
		}()
	}
	wg.Wait()

	if injected != 1 {
		t.Errorf("expected exactly 1 winner even when every rename succeeds, got %d", injected)
	}
}

// TestClaimGate_HeldGateBlocksClaim pins that a held gate makes the claim yield
// rather than steal it — the property that keeps the winner single.
//
// NOT parallel: mutates the handoffRenameFunc package global.
func TestClaimGate_HeldGateBlocksClaim(t *testing.T) {
	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume"))

	orig := handoffRenameFunc
	defer func() { handoffRenameFunc = orig }()
	handoffRenameFunc = func(from, to string) error { _ = os.Rename(from, to); return nil }

	if err := os.WriteFile(handoff.ClaimGatePath(pd), nil, 0o600); err != nil {
		t.Fatalf("seed gate: %v", err)
	}

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle must be fail-open, got: %v", err)
	}
	if additionalContextOf(out) != "" {
		t.Error("a held gate must block the claim")
	}
	if !pendingExists(pd) {
		t.Error("a blocked claim must leave pending.json in place")
	}
}

// TestClaimGate_ReleasedAfterClaim proves the gate does not leak on the success
// path: a second record saved afterwards is still claimable.
func TestClaimGate_ReleasedAfterClaim(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, livePending("first"))

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	if out, err := h.Handle(context.Background(), injectInput("clear", pd)); err != nil {
		t.Fatalf("Handle: %v", err)
	} else if additionalContextOf(out) == "" {
		t.Fatal("expected the first record to be injected")
	}
	if _, err := os.Stat(handoff.ClaimGatePath(pd)); !os.IsNotExist(err) {
		t.Errorf("gate must be released after a successful claim; stat err=%v", err)
	}

	mustSavePending(t, pd, livePending("second"))
	if out, err := h.Handle(context.Background(), injectInput("clear", pd)); err != nil {
		t.Fatalf("Handle: %v", err)
	} else if additionalContextOf(out) == "" {
		t.Error("expected the second record to be injected too")
	}
}

// TestSavePending_ClearsLeakedGate proves a gate orphaned by a killed consumer
// does not disable auto-resume forever: writing the next record clears it.
func TestSavePending_ClearsLeakedGate(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, livePending("resume"))
	if err := os.WriteFile(handoff.ClaimGatePath(pd), nil, 0o600); err != nil {
		t.Fatalf("seed gate: %v", err)
	}

	// Re-saving is the recovery point: the new record has never been claimed.
	mustSavePending(t, pd, livePending("resume"))
	if _, err := os.Stat(handoff.ClaimGatePath(pd)); !os.IsNotExist(err) {
		t.Fatalf("SavePending must clear a leaked gate; stat err=%v", err)
	}

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	if out, err := h.Handle(context.Background(), injectInput("clear", pd)); err != nil {
		t.Fatalf("Handle: %v", err)
	} else if additionalContextOf(out) == "" {
		t.Error("expected injection after the leaked gate was cleared")
	}
}

// --- AC-014: NULL session_id nonce filename shape ----------------------------

func TestNonceFallback_FilenameShape(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	// saved_by_session empty → crypto/rand 8-hex nonce.
	mustSavePending(t, pd, &handoff.PendingRecord{Body: "resume", SavedBySession: "", SavedAt: time.Now()})

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	if _, err := h.Handle(context.Background(), injectInput("clear", pd)); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	names := consumedNames(t, pd)
	if len(names) != 1 {
		t.Fatalf("expected 1 consumed file, got %d", len(names))
	}
	shape := regexp.MustCompile(`^\d+-[0-9a-f]{8}\.json$`)
	if !shape.MatchString(names[0]) {
		t.Errorf("consumed filename %q does not match ^\\d+-[0-9a-f]{8}\\.json$", names[0])
	}
}

// --- AC-015: i18n header -----------------------------------------------------

func TestInjectionHeader_I18n(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"ko": "[MoAI 자동 재개",
		"en": "[MoAI auto-resume",
		"ja": "[MoAI 自動再開",
		"zh": "[MoAI 自动恢复",
		"fr": "[MoAI auto-resume", // outside {ko,en,ja,zh} → English fallback
	}
	for lang, wantHeader := range cases {
		lang, wantHeader := lang, wantHeader
		t.Run(lang, func(t *testing.T) {
			t.Parallel()
			got := renderHandoffContext(&handoff.PendingRecord{Body: "b", ConversationLanguage: lang})
			if !strings.HasPrefix(got, wantHeader) {
				t.Errorf("lang %q: header prefix\n got: %q\nwant prefix: %q", lang, handoffFirstLine(got), wantHeader)
			}
		})
	}
}

func handoffFirstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// --- AC-016: subagent boundary (no user interaction) -------------------------

func TestNoUserInteraction(t *testing.T) {
	t.Parallel()

	for _, f := range []string{"handoff_inject.go", "handoff_inject_render.go"} {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if strings.Contains(line, "AskUserQuestion") || strings.Contains(line, "mcp__askuser") {
				t.Errorf("%s:%d references a user-interaction token (subagent boundary): %s", f, i+1, trimmed)
			}
		}
	}
}

// --- AC-017: fail-open on corrupt pending ------------------------------------

func TestFailOpen_CorruptPending(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	dir := filepath.Join(pd, ".moai", "state", "handoff")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(handoff.PendingPath(pd), []byte("{corrupt json"), 0o600); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}

	h := NewHandoffInjectHandler(autoCfgProvider(false))
	out, err := h.Handle(context.Background(), injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Handle must be fail-open, got: %v", err)
	}
	if additionalContextOf(out) != "" {
		t.Error("corrupt pending must not inject")
	}
	if !pendingExists(pd) {
		t.Error("corrupt pending must be preserved (not renamed)")
	}
	if n := len(consumedNames(t, pd)); n != 0 {
		t.Errorf("corrupt pending must not produce a consumed file, got %d", n)
	}
}

// --- AC-018: 3-handler coexist (registry accumulate-all) ---------------------

func TestThreeHandlerCoexist(t *testing.T) {
	t.Parallel()

	pd := t.TempDir()
	mustSavePending(t, pd, &handoff.PendingRecord{Body: "resume", ConversationLanguage: "en", SavedAt: time.Now()})

	cfg := &mockConfigProvider{cfg: &config.Config{Handoff: config.HandoffConfig{Mode: "auto"}}}
	reg := NewRegistry(cfg)
	reg.Register(NewSessionStartHandler(cfg))
	reg.Register(NewAutoUpdateHandler(func(_ context.Context) (*AutoUpdateResult, error) {
		return &AutoUpdateResult{Updated: false}, nil
	}))
	reg.Register(NewHandoffInjectHandler(cfg))

	out, err := reg.Dispatch(context.Background(), EventSessionStart, injectInput("clear", pd))
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	ctx := additionalContextOf(out)
	if !strings.Contains(ctx, "moai session attribution") {
		t.Errorf("merged additionalContext missing sessionStartHandler attribution: %q", ctx)
	}
	if !strings.Contains(ctx, "auto-resume") {
		t.Errorf("merged additionalContext missing handoff guidance (both handlers must survive the accumulate-all merge): %q", ctx)
	}
}

// --- AC-019: auto-mode stale TTL cleanup (N1 precedence) ---------------------
// NOT parallel: captures os.Stderr.

func TestStaleTTLCleanup_AutoOnly(t *testing.T) {
	// stale + auto + notice-only source + guide=true → remove, no inject, NO hint.
	t.Run("startup_guide_stale", func(t *testing.T) {
		pd := t.TempDir()
		mustSavePending(t, pd, stalePending("stale"))

		h := NewHandoffInjectHandler(autoCfgProvider(true))
		var out *HookOutput
		stderr := captureStderr(t, func() {
			o, err := h.Handle(context.Background(), injectInput("startup", pd))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			out = o
		})
		if additionalContextOf(out) != "" {
			t.Error("stale: must not inject")
		}
		if pendingExists(pd) {
			t.Error("stale auto-mode: pending.json must be removed")
		}
		if n := len(consumedNames(t, pd)); n != 0 {
			t.Errorf("stale cleanup must NOT produce a consumed file (not a real consume), got %d", n)
		}
		if strings.Contains(stderr, "auto-resume") {
			t.Errorf("N1: stale cleanup must suppress the notice hint even when guide=true, got %q", stderr)
		}
	})

	// stale + auto + clear → also removed, no inject (stale precedence over clear).
	t.Run("clear_stale", func(t *testing.T) {
		pd := t.TempDir()
		mustSavePending(t, pd, stalePending("stale"))

		h := NewHandoffInjectHandler(autoCfgProvider(false))
		out, err := h.Handle(context.Background(), injectInput("clear", pd))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if additionalContextOf(out) != "" {
			t.Error("stale clear: must not inject a stale record")
		}
		if pendingExists(pd) {
			t.Error("stale clear: pending.json must be removed")
		}
		if n := len(consumedNames(t, pd)); n != 0 {
			t.Errorf("stale clear cleanup must not create a consumed file, got %d", n)
		}
	})
}
