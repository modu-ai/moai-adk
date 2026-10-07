package hook

// session_start_memory_budget_test.go — the AC-MFB-012 cells for the
// SessionStart MEMORY.md budget advisory (SPEC-MEMORY-FOLD-BUDGET-001
// follow-up card): the four-source loop, the silence cells, the join bound,
// the sandbox containment, the recorder positive control, and the join-bound
// constant's value/ceiling pin (audit finding D23's join-bound half).
//
// The assertions read the SERIALIZED handler output, never the Data map: a
// line placed in the json:"-" Data map would reach nobody, and asserting on
// the serialized bytes is what closes that mutant.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// memoryBudgetPathRecorder is the path-recording read seam TestMain installs
// over memoryBudgetReadFile (audit finding D5). It records EVERY path it is
// asked for, opens only the ones beneath the sandbox root, and answers
// os.ErrNotExist for the rest — so a handler test that steers the advisory
// at the real store leaves a recorded path (the post-run containment
// assertion fails naming it) without the real store ever being opened.
type memoryBudgetPathRecorder struct {
	mu   sync.Mutex
	seen []string
	root string
}

// memoryBudgetRecorder is the recorder TestMain installed, if any. Tests use
// memoryBudgetRecorderSnapshot to read it; production paths never touch it.
var memoryBudgetRecorder *memoryBudgetPathRecorder

// installMemoryBudgetReadRecorder wraps memoryBudgetReadFile with the
// recording containment seam. Called from TestMain, once per binary.
func installMemoryBudgetReadRecorder(root string) {
	memoryBudgetRecorder = &memoryBudgetPathRecorder{root: root}
	memoryBudgetReadFile = memoryBudgetRecorder.read
}

func (r *memoryBudgetPathRecorder) read(path string) ([]byte, error) {
	r.mu.Lock()
	r.seen = append(r.seen, path)
	r.mu.Unlock()
	if !memoryBudgetUnderRoot(path, r.root) {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(path)
}

// snapshot returns a copy of every path recorded so far.
func (r *memoryBudgetPathRecorder) snapshot() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.seen...)
}

// memoryBudgetRecorderPathsOutside returns every recorded path that does not
// lie beneath root — the TestMain post-run containment assertion's input.
func memoryBudgetRecorderPathsOutside(root string) []string {
	var offenders []string
	for _, p := range memoryBudgetRecorder.snapshot() {
		if !memoryBudgetUnderRoot(p, root) {
			offenders = append(offenders, p)
		}
	}
	return offenders
}

// memoryBudgetUnderRoot reports whether path is root itself or lies beneath
// it, compared by the same literal strings the derivation built both from
// (the sandbox root is passed around verbatim; no symlink resolution is
// involved on either side).
func memoryBudgetUnderRoot(path, root string) bool {
	if root == "" {
		return false
	}
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(filepath.Separator))
}

// memoryBudgetTestSlug computes the project slug with its OWN literal
// mapping rather than calling projectSlug: this is the independent
// expectation of OD-9's drift guard, so a future edit to either the hook's
// projectSlug or the CLI's memoryProjectSlug that changes the mapping breaks
// this test loudly instead of silently following the drift.
func memoryBudgetTestSlug(abs string) string {
	clean := filepath.Clean(abs)
	var b strings.Builder
	for _, r := range clean {
		switch r {
		case '/', '\\', '.', ':':
			b.WriteRune('-')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// memoryBudgetTestRoot makes this test's isolated profile root as a subdir
// of the TestMain home sandbox — the only place the recording seam opens
// files — and registers its removal. A t.TempDir() root would sit outside
// the sandbox and the seam would refuse its reads by design.
func memoryBudgetTestRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(os.Getenv(moaiHomeSandboxEnv), t.Name())
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("make isolated profile root: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	return root
}

// memoryBudgetPlantStore writes a MEMORY.md of indexLen bytes (one line, so
// the byte axis is the one that fires) into store and returns the store dir.
func memoryBudgetPlantStore(t *testing.T, store string, indexLen int) string {
	t.Helper()
	if err := os.MkdirAll(store, 0o755); err != nil {
		t.Fatalf("make store: %v", err)
	}
	body := strings.Repeat("a", indexLen-1) + "\n"
	if err := os.WriteFile(filepath.Join(store, "MEMORY.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	return store
}

// memoryBudgetAtWarnFloor returns the byte length at the warn threshold
// exactly: len*100 == warnPercent*byteCap, the integer test's boundary.
func memoryBudgetAtWarnFloor() int {
	return config.DefaultMemoryIndexByteCap * config.DefaultMemoryIndexWarnPercent / 100
}

// memoryBudgetLineWithPrefix returns the single advisory line from a
// serialized HookOutput, failing the test when the count is not exactly one.
func memoryBudgetLineWithPrefix(t *testing.T, blob []byte) string {
	t.Helper()
	if n := strings.Count(string(blob), memoryBudgetPrefix); n != 1 {
		t.Fatalf("serialized output carries %d %q occurrences, want exactly 1 (one line, one channel):\n%s", n, memoryBudgetPrefix, blob)
	}
	var parsed struct {
		HookSpecificOutput *struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(blob, &parsed); err != nil {
		t.Fatalf("parse serialized output: %v", err)
	}
	if parsed.HookSpecificOutput == nil {
		t.Fatalf("serialized output has no hookSpecificOutput block:\n%s", blob)
	}
	for _, line := range strings.Split(parsed.HookSpecificOutput.AdditionalContext, "\n") {
		if strings.Contains(line, memoryBudgetPrefix) {
			return line
		}
	}
	t.Fatalf("prefix counted in the serialized bytes but not found in additionalContext:\n%s", blob)
	return ""
}

// TestSessionStartMemoryBudget_FourSources is AC-MFB-012 (a) and (d): every
// session source gets exactly one budget line through additionalContext and
// no other channel, the line names the store path / larger percentage /
// keyed measure / doctor pointer, and the derived store equals the
// independently computed profile-key path beneath the temporary root.
func TestSessionStartMemoryBudget_FourSources(t *testing.T) {
	// Not parallel: t.Setenv mutates process-wide state.
	clearFactoryEnv(t)
	t.Setenv("ANTHROPIC_BASE_URL", "")

	root := memoryBudgetTestRoot(t)
	cfgDir := filepath.Join(root, "profile-config")
	projectDir := filepath.Join(root, "project")
	wantStore := filepath.Join(cfgDir, "projects", memoryBudgetTestSlug(projectDir), "memory")
	memoryBudgetPlantStore(t, wantStore, memoryBudgetAtWarnFloor()+1)

	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root) // os.UserHomeDir reads USERPROFILE on Windows
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)

	if wantPct := (memoryBudgetAtWarnFloor() + 1) * 100 / config.DefaultMemoryIndexByteCap; wantPct < config.DefaultMemoryIndexWarnPercent {
		t.Fatalf("planted index (%d bytes) sits below the warn threshold", memoryBudgetAtWarnFloor()+1)
	}

	for _, source := range []string{"startup", "resume", "clear", "compact"} {
		t.Run(source, func(t *testing.T) {
			h := NewSessionStartHandler(&mockConfigProvider{cfg: newTestConfig()})
			out, err := h.Handle(context.Background(), &HookInput{
				SessionID:     "sess-mem-budget-" + source,
				CWD:           projectDir,
				ProjectDir:    projectDir,
				HookEventName: "SessionStart",
				Source:        source,
			})
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			blob, err := json.Marshal(out)
			if err != nil {
				t.Fatalf("marshal output: %v", err)
			}
			line := memoryBudgetLineWithPrefix(t, blob)
			for _, want := range []string{wantStore, fmt.Sprintf("%d%%", (memoryBudgetAtWarnFloor()+1)*100/config.DefaultMemoryIndexByteCap), "bytes", "moai memory doctor"} {
				if !strings.Contains(line, want) {
					t.Errorf("line %q does not name %q", line, want)
				}
			}
			// (d): the derived store lies beneath the temporary root and
			// equals the independently computed profile-key path.
			if !memoryBudgetUnderRoot(wantStore, root) {
				t.Errorf("store %q is not beneath the isolated root %q", wantStore, root)
			}
		})
	}
}

// TestSessionStartMemoryBudget_BelowAbsentUnreadableKillswitch is
// AC-MFB-012 (b): a store one byte below the threshold, an absent store, an
// unreadable MEMORY.md, and MOAI_MEMORY_AUDIT=0 each add nothing and the
// handler returns no error.
func TestSessionStartMemoryBudget_BelowAbsentUnreadableKillswitch(t *testing.T) {
	// Not parallel: t.Setenv and the seam swap mutate process-wide state.
	clearFactoryEnv(t)
	t.Setenv("ANTHROPIC_BASE_URL", "")

	runHandle := func(t *testing.T) string {
		t.Helper()
		root := memoryBudgetTestRoot(t)
		h := NewSessionStartHandler(&mockConfigProvider{cfg: newTestConfig()})
		out, err := h.Handle(context.Background(), &HookInput{
			SessionID:     "sess-mem-budget-silence",
			CWD:           filepath.Join(root, "project"),
			ProjectDir:    filepath.Join(root, "project"),
			HookEventName: "SessionStart",
			Source:        "startup",
		})
		if err != nil {
			t.Fatalf("Handle returned an error: %v", err)
		}
		blob, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("marshal output: %v", err)
		}
		return string(blob)
	}

	t.Run("one byte below the threshold", func(t *testing.T) {
		root := memoryBudgetTestRoot(t)
		projectDir := filepath.Join(root, "project")
		store := filepath.Join(root, ".claude", "projects", memoryBudgetTestSlug(projectDir), "memory")
		memoryBudgetPlantStore(t, store, memoryBudgetAtWarnFloor()-1)
		t.Setenv("HOME", root)
		t.Setenv("USERPROFILE", root)
		t.Setenv(config.EnvClaudeConfigDir, "")
		if got := runHandle(t); strings.Contains(got, memoryBudgetPrefix) {
			t.Errorf("below-threshold store produced a budget line:\n%s", got)
		}
	})

	t.Run("absent store", func(t *testing.T) {
		root := memoryBudgetTestRoot(t)
		t.Setenv("HOME", root)
		t.Setenv("USERPROFILE", root)
		t.Setenv(config.EnvClaudeConfigDir, "")
		if err := os.MkdirAll(filepath.Join(root, "project"), 0o755); err != nil {
			t.Fatalf("make project dir: %v", err)
		}
		if got := runHandle(t); strings.Contains(got, memoryBudgetPrefix) {
			t.Errorf("absent store produced a budget line:\n%s", got)
		}
	})

	t.Run("unreadable MEMORY.md", func(t *testing.T) {
		// The read error is injected through the advisory's own read seam:
		// the production contract is "on any read error it adds nothing", and a
		// seam-level error exercises that branch deterministically on every
		// platform (chmod-based unreadability is not enforceable on Windows).
		root := memoryBudgetTestRoot(t)
		t.Setenv("HOME", root)
		t.Setenv("USERPROFILE", root)
		t.Setenv(config.EnvClaudeConfigDir, "")
		if err := os.MkdirAll(filepath.Join(root, "project"), 0o755); err != nil {
			t.Fatalf("make project dir: %v", err)
		}
		orig := memoryBudgetReadFile
		memoryBudgetReadFile = func(path string) ([]byte, error) {
			return nil, os.ErrPermission
		}
		t.Cleanup(func() { memoryBudgetReadFile = orig })
		if got := runHandle(t); strings.Contains(got, memoryBudgetPrefix) {
			t.Errorf("unreadable MEMORY.md produced a budget line:\n%s", got)
		}
	})

	t.Run("kill switch", func(t *testing.T) {
		root := memoryBudgetTestRoot(t)
		projectDir := filepath.Join(root, "project")
		store := filepath.Join(root, ".claude", "projects", memoryBudgetTestSlug(projectDir), "memory")
		memoryBudgetPlantStore(t, store, memoryBudgetAtWarnFloor()+1)
		t.Setenv("HOME", root)
		t.Setenv("USERPROFILE", root)
		t.Setenv(config.EnvClaudeConfigDir, "")
		t.Setenv(config.EnvMemoryAudit, "0")
		if got := runHandle(t); strings.Contains(got, memoryBudgetPrefix) {
			t.Errorf("MOAI_MEMORY_AUDIT=0 produced a budget line:\n%s", got)
		}
	})
}

// TestSessionStartMemoryBudget_JoinBound is AC-MFB-012 (c): a read seam
// that blocks past the join bound, with the bound overridden to 50 ms, adds
// nothing and returns in under 250 ms (bound plus 200 ms slack).
//
// The advisory is invoked directly on its async path rather than through
// Handle: the bound is the advisory's contribution, and a whole-Handle
// timing budget would be spent on unrelated stages (settings chain, drift,
// factory notices) that have nothing to do with the join bound under test.
func TestSessionStartMemoryBudget_JoinBound(t *testing.T) {
	// Not parallel: the seam and bound swaps mutate package state.
	root := memoryBudgetTestRoot(t)
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("make project dir: %v", err)
	}
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv(config.EnvClaudeConfigDir, "")

	release := make(chan struct{})
	origSeam := memoryBudgetReadFile
	memoryBudgetReadFile = func(path string) ([]byte, error) {
		<-release // blocks until the cleanup-guaranteed teardown release
		return nil, os.ErrNotExist
	}
	origBound := memoryBudgetJoinBound
	memoryBudgetJoinBound = 50 * time.Millisecond
	t.Cleanup(func() {
		memoryBudgetReadFile = origSeam
		memoryBudgetJoinBound = origBound
		close(release) // lets the abandoned reader exit so the binary leaks no goroutine
	})

	start := time.Now()
	line := memoryBudgetAdvisory(context.Background(), projectDir, true)
	elapsed := time.Since(start)
	if line != "" {
		t.Errorf("blocked read produced a budget line: %q", line)
	}
	if limit := 50*time.Millisecond + 200*time.Millisecond; elapsed >= limit {
		t.Errorf("advisory returned in %v, want under %v (50ms bound + 200ms slack)", elapsed, limit)
	}
}

// TestSessionStartMemoryBudget_StaysUnderTempHome is AC-MFB-012 (d)'s
// containment half on the default key: the derived store lies beneath the
// temporary root, the line names it, and every path the read seam recorded
// during this test lies beneath the TestMain sandbox root.
func TestSessionStartMemoryBudget_StaysUnderTempHome(t *testing.T) {
	// Not parallel: t.Setenv mutates process-wide state.
	root := memoryBudgetTestRoot(t)
	projectDir := filepath.Join(root, "project")
	wantStore := filepath.Join(root, ".claude", "projects", memoryBudgetTestSlug(projectDir), "memory")
	memoryBudgetPlantStore(t, wantStore, memoryBudgetAtWarnFloor()+1)

	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv(config.EnvClaudeConfigDir, "")

	before := len(memoryBudgetRecorder.snapshot())
	line := memoryBudgetAdvisory(context.Background(), projectDir, false)
	if !strings.Contains(line, wantStore) {
		t.Fatalf("default-key line %q does not name the independently computed store %q", line, wantStore)
	}
	if !memoryBudgetUnderRoot(wantStore, root) {
		t.Errorf("store %q is not beneath the isolated root %q", wantStore, root)
	}
	sandbox := os.Getenv(moaiHomeSandboxEnv)
	for _, p := range memoryBudgetRecorder.snapshot()[before:] {
		if !memoryBudgetUnderRoot(p, sandbox) {
			t.Errorf("read seam recorded a path outside the sandbox root: %q", p)
		}
	}
}

// TestSessionStartMemoryBudget_RecorderSeesAdvisoryRead is AC-MFB-012 (e),
// the positive control: with a planted store the path-recording read seam
// logs exactly that path — proving the recorder is wired to the advisory's
// read, so the containment assertion guards the real thing.
func TestSessionStartMemoryBudget_RecorderSeesAdvisoryRead(t *testing.T) {
	// Not parallel: t.Setenv mutates process-wide state.
	root := memoryBudgetTestRoot(t)
	cfgDir := filepath.Join(root, "profile-config")
	projectDir := filepath.Join(root, "project")
	store := memoryBudgetPlantStore(t, filepath.Join(cfgDir, "projects", memoryBudgetTestSlug(projectDir), "memory"), memoryBudgetAtWarnFloor()+1)

	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)

	before := len(memoryBudgetRecorder.snapshot())
	if line := memoryBudgetAdvisory(context.Background(), projectDir, false); line == "" {
		t.Fatal("planted over-threshold store produced no line")
	}
	delta := memoryBudgetRecorder.snapshot()[before:]
	want := filepath.Join(store, "MEMORY.md")
	if len(delta) != 1 || delta[0] != want {
		t.Errorf("recorder logged %v, want exactly [%q]", delta, want)
	}
}

// TestSessionStartMemoryBudget_JoinBoundBelowHookTimeout pins the D23
// join-bound half: the constant carries its OWN value (not a reuse of
// DefaultHookAsyncJoinTimeout or binaryLagJoinBound) and sits strictly below
// the 5s per-event hook timeout policy (internal/hook/CLAUDE.md).
func TestSessionStartMemoryBudget_JoinBoundBelowHookTimeout(t *testing.T) {
	t.Parallel()
	if memoryBudgetJoinBound != 250*time.Millisecond {
		t.Errorf("memoryBudgetJoinBound = %v, want its own value 250ms", memoryBudgetJoinBound)
	}
	if config.DefaultMemoryBudgetJoinBound != 250*time.Millisecond {
		t.Errorf("config.DefaultMemoryBudgetJoinBound = %v, want 250ms", config.DefaultMemoryBudgetJoinBound)
	}
	const hookPolicyTimeout = 5 * time.Second // internal/hook/CLAUDE.md: ≤5s per hook event
	if memoryBudgetJoinBound >= hookPolicyTimeout {
		t.Errorf("memoryBudgetJoinBound = %v is not strictly below the %v hook timeout policy",
			memoryBudgetJoinBound, hookPolicyTimeout)
	}
}
