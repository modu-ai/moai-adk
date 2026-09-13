package hook

// slot_lease_guard_test.go — the PreToolUse slot-lease guard (card t607, M1 RED).
//
// A deny needs positive evidence on EVERY term of one compound condition:
// guard enabled ∧ a configured pattern matches outside quotes ∧ the holder is a
// different session ∧ the holder is live ∧ the declared bound has not elapsed.
// AC-RSL-011 gives each term its own row where only that term is false, and
// each allow row asserts its OWN audit reason, so a mutant that deletes one
// check and leaks through a different allow path fails on the reason rather
// than passing on the bare allow.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

const (
	slotGuardResource = "demo"
	slotGuardCommand  = "heavy-suite run --all"
	slotAdvisoryTag   = "[moai:slot-lease] advisory:"
)

// scrubSlotGuardEnv pins the environment named by the SPEC's isolation clause.
//
// The session-pid override variable is deliberately NOT touched here: the
// guard never resolves an owner pid (it only reads the pid a record carries),
// and internal/cli's TestSessionPIDStamp_NotSetFromHooks forbids any hook
// source, test files included, from naming that variable at all.
func scrubSlotGuardEnv(t *testing.T, root string) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
}

// slotGuardRepo is a primary git checkout, so hook-root normalization has a
// git common dir to resolve and resolves to the checkout itself.
func slotGuardRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	parent := t.TempDir()
	repo := filepath.Join(parent, "primary")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gitInitRepo(t, repo)
	scrubSlotGuardEnv(t, repo)
	return repo
}

func slotGuardInput(t *testing.T, sessionID, command string) *HookInput {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"command": command})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	return &HookInput{
		SessionID:     sessionID,
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     raw,
	}
}

func slotGuardConfig(enabled bool) config.SlotLeaseConfig {
	return config.SlotLeaseConfig{
		Enabled:            enabled,
		DefaultMaxDuration: "30m",
		Resources: map[string]config.SlotLeaseResourceConfig{
			slotGuardResource: {Commands: []string{`\bheavy-suite\b`}},
		},
	}
}

// liveForeignLease is held by s-1 with a live owner (this test process) and an
// unelapsed bound.
func liveForeignLease() kanban.SlotLease {
	now := time.Now().UTC()
	return kanban.SlotLease{
		Resource:    slotGuardResource,
		SessionID:   "s-1",
		SessionName: "lane-s-1",
		PID:         os.Getpid(),
		PIDSource:   kanban.PIDSourceSessionOwner,
		Command:     "heavy-suite",
		AcquiredAt:  now.Add(-time.Minute).Format(time.RFC3339),
		MaxDuration: time.Hour.String(),
		ExpiresAt:   now.Add(time.Hour).Format(time.RFC3339),
	}
}

func slotGuardRecordPath(root, resource string) string {
	return filepath.Join(root, ".moai", "state", "slot-leases", resource+".json")
}

func seedSlotGuardLease(t *testing.T, root string, lease kanban.SlotLease) {
	t.Helper()
	path := slotGuardRecordPath(root, lease.Resource)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("seed mkdir: %v", err)
	}
	data, err := json.MarshalIndent(&lease, "", "  ")
	if err != nil {
		t.Fatalf("seed marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}
}

func slotGuardAuditPath(root string) string {
	return filepath.Join(root, ".moai", "logs", "slot-lease-audit.jsonl")
}

// slotGuardAudit returns the audit log's entries (nil when the log is absent).
func slotGuardAudit(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(slotGuardAuditPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("audit line is not JSON (%v): %q", err, line)
		}
		out = append(out, m)
	}
	return out
}

func lastSlotGuardEvent(t *testing.T, root string) (event, reason string, ok bool) {
	t.Helper()
	entries := slotGuardAudit(t, root)
	if len(entries) == 0 {
		return "", "", false
	}
	last := entries[len(entries)-1]
	event, _ = last["event"].(string)
	reason, _ = last["reason"].(string)
	return event, reason, true
}

func exitedChildPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("running the dummy child: %v", err)
	}
	pid := cmd.Process.Pid
	if kanban.FactoryProcessAlive(pid) {
		t.Skipf("pid %d of an exited child reads live (reused); the stale row is not exercisable", pid)
	}
	return pid
}

// AC-RSL-011 — the deny compound condition, one row per term.
func TestSlotLeaseGuard_DenyMatrix(t *testing.T) {
	type row struct {
		name      string
		enabled   bool
		command   string
		caller    string
		seed      func(t *testing.T, root string)
		resources map[string]config.SlotLeaseResourceConfig
		wantDeny  bool
		wantEvent string // "" = no audit line may be added
		denyNames string
	}
	seedLive := func(t *testing.T, root string) { seedSlotGuardLease(t, root, liveForeignLease()) }
	rows := []row{
		{name: "base", enabled: true, command: slotGuardCommand, caller: "s-2", seed: seedLive, wantDeny: true, wantEvent: "guard-deny", denyNames: slotGuardResource},
		{name: "n-enabled", enabled: false, command: slotGuardCommand, caller: "s-2", seed: seedLive},
		{name: "n-match", enabled: true, command: "echo hello", caller: "s-2", seed: seedLive},
		{name: "n-quoted", enabled: true, command: "echo 'heavy-suite run --all'", caller: "s-2", seed: seedLive},
		{name: "n-self", enabled: true, command: slotGuardCommand, caller: "s-1", seed: seedLive, wantEvent: "allow-self"},
		{name: "n-alive", enabled: true, command: slotGuardCommand, caller: "s-2", seed: func(t *testing.T, root string) {
			lease := liveForeignLease()
			lease.PID = exitedChildPID(t)
			seedSlotGuardLease(t, root, lease)
		}, wantEvent: "allow-stale"},
		{name: "n-bound", enabled: true, command: slotGuardCommand, caller: "s-2", seed: func(t *testing.T, root string) {
			lease := liveForeignLease()
			now := time.Now().UTC()
			lease.AcquiredAt = now.Add(-2 * time.Hour).Format(time.RFC3339)
			lease.ExpiresAt = now.Add(-time.Hour).Format(time.RFC3339)
			seedSlotGuardLease(t, root, lease)
		}, wantEvent: "allow-expired"},
		{name: "n-held", enabled: true, command: slotGuardCommand, caller: "s-2", seed: func(*testing.T, string) {}, wantEvent: "allow-unheld"},
		{
			// Two resources match; only "demo" satisfies the deny condition.
			// "alpha" sorts first and is free, so a mutant that judges only the
			// first attributed resource allows and fails this row.
			name: "multi", enabled: true, command: "bench-all && " + slotGuardCommand, caller: "s-2", seed: seedLive,
			resources: map[string]config.SlotLeaseResourceConfig{
				"alpha":           {Commands: []string{`\bbench-all\b`}},
				slotGuardResource: {Commands: []string{`\bheavy-suite\b`}},
			},
			wantDeny: true, wantEvent: "guard-deny", denyNames: slotGuardResource,
		},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			root := slotGuardRepo(t)
			tc.seed(t, root)
			cfg := slotGuardConfig(tc.enabled)
			if tc.resources != nil {
				cfg.Resources = tc.resources
			}
			linesBefore := len(slotGuardAudit(t, root))
			var advisory bytes.Buffer

			decision, reason := checkSlotLease(slotGuardInput(t, tc.caller, tc.command), root, cfg, &advisory)

			if tc.wantDeny {
				if decision != DecisionDeny {
					t.Fatalf("decision = %q, want deny", decision)
				}
				if !strings.HasPrefix(reason, slotLeaseViolationPrefix) {
					t.Errorf("reason lacks the %q prefix: %s", slotLeaseViolationPrefix, reason)
				}
				for _, want := range []string{tc.denyNames, "lane-s-1", "moai slot"} {
					if !strings.Contains(reason, want) {
						t.Errorf("reason does not name %q (resource / holder / how to wait, release, or take over): %s", want, reason)
					}
				}
			} else if decision == DecisionDeny {
				t.Fatalf("decision = deny (%s), want allow", reason)
			}

			if tc.wantEvent == "" {
				if after := len(slotGuardAudit(t, root)); after != linesBefore {
					t.Errorf("audit lines %d -> %d, want no line for a call the guard does not attribute to a resource", linesBefore, after)
				}
				return
			}
			event, _, ok := lastSlotGuardEvent(t, root)
			if !ok {
				t.Fatalf("no audit line written, want last event %q", tc.wantEvent)
			}
			if event != tc.wantEvent {
				t.Errorf("last audit event = %q, want %q", event, tc.wantEvent)
			}
		})
	}
}

// AC-RSL-012 — once enabled has been read as true, every uncertainty allows,
// writes an advisory, and (where a root exists) a fail-open audit line.
func TestSlotLeaseGuard_FailOpen(t *testing.T) {
	type row struct {
		name     string
		caller   string
		cfg      func() config.SlotLeaseConfig
		seed     func(t *testing.T, root string)
		noRoot   bool
		auditReq bool
	}
	seedLive := func(t *testing.T, root string) { seedSlotGuardLease(t, root, liveForeignLease()) }
	rows := []row{
		{name: "corrupt_record", caller: "s-2", cfg: func() config.SlotLeaseConfig { return slotGuardConfig(true) }, seed: func(t *testing.T, root string) {
			path := slotGuardRecordPath(root, slotGuardResource)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
		}, auditReq: true},
		{name: "missing_session_id", caller: "", cfg: func() config.SlotLeaseConfig { return slotGuardConfig(true) }, seed: seedLive, auditReq: true},
		{name: "uncompilable_pattern", caller: "s-2", cfg: func() config.SlotLeaseConfig {
			c := slotGuardConfig(true)
			c.Resources = map[string]config.SlotLeaseResourceConfig{slotGuardResource: {Commands: []string{"("}}}
			return c
		}, seed: seedLive, auditReq: true},
		{name: "missing_project_root", caller: "s-2", cfg: func() config.SlotLeaseConfig { return slotGuardConfig(true) }, seed: func(*testing.T, string) {}, noRoot: true},
		{name: "uninterpretable_resource_entry", caller: "s-2", cfg: func() config.SlotLeaseConfig {
			c := slotGuardConfig(true)
			c.Resources = map[string]config.SlotLeaseResourceConfig{slotGuardResource: {Invalid: "commands is not a list of pattern strings"}}
			return c
		}, seed: seedLive, auditReq: true},
	}
	for _, tc := range rows {
		t.Run(tc.name, func(t *testing.T) {
			root := slotGuardRepo(t)
			tc.seed(t, root)
			hookRoot := root
			if tc.noRoot {
				hookRoot = ""
			}
			var advisory bytes.Buffer
			decision, reason := checkSlotLease(slotGuardInput(t, tc.caller, slotGuardCommand), hookRoot, tc.cfg(), &advisory)
			if decision == DecisionDeny {
				t.Fatalf("denied on uncertainty (%s) — the guard must fail open", reason)
			}
			if !strings.Contains(advisory.String(), slotAdvisoryTag) {
				t.Errorf("no %q line on the advisory stream; a silent allow makes the guard look enforcing while it is not. got: %q", slotAdvisoryTag, advisory.String())
			}
			if !tc.auditReq {
				return
			}
			event, auditReason, ok := lastSlotGuardEvent(t, root)
			if !ok || event != "fail-open" || auditReason == "" {
				t.Errorf("last audit = (%q, %q, present=%v), want a fail-open line carrying its reason", event, auditReason, ok)
			}
		})
	}
}

// AC-RSL-010 — disabled (key absent, explicit false, or no configuration at
// all) reads no record and says nothing. Driven through the real handler so
// the call-site gate is what is tested; each disabled row has a positive
// control proving the same fixture DOES deny / advise once enabled.
func TestSlotLeaseGuard_DisabledNeverReadsNorDenies(t *testing.T) {
	resources := map[string]config.SlotLeaseResourceConfig{
		slotGuardResource: {Commands: []string{`\bheavy-suite\b`}},
	}
	cfgWith := func(enabled bool) *config.Config {
		c := config.NewDefaultConfig()
		c.Workflow.SlotLease.Enabled = enabled
		c.Workflow.SlotLease.Resources = resources
		return c
	}
	states := []struct {
		name     string
		provider ConfigProvider
	}{
		{"key_absent", &mockConfigProvider{cfg: func() *config.Config {
			c := config.NewDefaultConfig()
			c.Workflow.SlotLease.Resources = resources // enabled never set
			return c
		}()}},
		{"explicit_false", &mockConfigProvider{cfg: cfgWith(false)}},
		{"config_unavailable", &mockConfigProvider{cfg: nil}},
	}
	records := []struct {
		name string
		seed func(t *testing.T, root string)
	}{
		{"live_foreign_holder", func(t *testing.T, root string) { seedSlotGuardLease(t, root, liveForeignLease()) }},
		{"corrupt_record", func(t *testing.T, root string) {
			path := slotGuardRecordPath(root, slotGuardResource)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
		}},
	}

	handle := func(t *testing.T, provider ConfigProvider, root string) (decision, stderr string) {
		t.Helper()
		h := &preToolHandler{cfg: provider, policy: DefaultSecurityPolicy(), projectDir: root}
		var out *HookOutput
		stderr = captureStderr(t, func() {
			var err error
			out, err = h.Handle(context.Background(), slotGuardInput(t, "s-2", slotGuardCommand))
			if err != nil {
				t.Errorf("Handle: %v", err)
			}
		})
		return decisionOf(out), stderr
	}

	for _, st := range states {
		for _, rec := range records {
			t.Run(st.name+"/"+rec.name, func(t *testing.T) {
				root := slotGuardRepo(t)
				rec.seed(t, root)
				decision, stderr := handle(t, st.provider, root)
				if decision == DecisionDeny {
					t.Errorf("disabled guard denied")
				}
				if strings.Contains(stderr, "[moai:slot-lease]") {
					t.Errorf("disabled guard wrote an advisory — the disabled path read the record or treated missing config as uncertainty: %q", stderr)
				}
				if _, err := os.Stat(slotGuardAuditPath(root)); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("disabled guard created the audit log (stat err %v)", err)
				}
			})
		}
	}

	// Positive controls: the same fixtures, enabled, reach the guard.
	t.Run("control_enabled/live_foreign_holder_denies", func(t *testing.T) {
		root := slotGuardRepo(t)
		seedSlotGuardLease(t, root, liveForeignLease())
		if decision, _ := handle(t, &mockConfigProvider{cfg: cfgWith(true)}, root); decision != DecisionDeny {
			t.Fatalf("enabled guard did not deny a live foreign holder through the handler (decision %q) — the disabled rows above assert nothing until this passes", decision)
		}
	})
	t.Run("control_enabled/corrupt_record_advises", func(t *testing.T) {
		root := slotGuardRepo(t)
		records[1].seed(t, root)
		if _, stderr := handle(t, &mockConfigProvider{cfg: cfgWith(true)}, root); !strings.Contains(stderr, slotAdvisoryTag) {
			t.Fatalf("enabled guard gave no advisory on a corrupt record through the handler — the stderr capture in the disabled rows cannot be interpreted. stderr: %q", stderr)
		}
	})
}

// AC-RSL-016 — the guard normalizes the hook root to the primary checkout with
// the SAME resolver the CLI uses, so a hook root inside a linked worktree
// still reads the record kept in the primary.
func TestSlotLeaseGuard_NormalizesWorktreeRootToPrimary(t *testing.T) {
	requireGit(t)
	parent := t.TempDir()
	primary := filepath.Join(parent, "primary")
	wt := filepath.Join(parent, "wt")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gitInitRepo(t, primary)
	mustRunGit(t, primary, "worktree", "add", wt, "-b", "wt-branch")
	scrubSlotGuardEnv(t, primary)
	seedSlotGuardLease(t, primary, liveForeignLease())
	if _, err := os.Stat(slotGuardRecordPath(wt, slotGuardResource)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("precondition: the worktree must hold no record of its own (stat err %v)", err)
	}
	cfg := slotGuardConfig(true)

	t.Run("wt-deny", func(t *testing.T) {
		var advisory bytes.Buffer
		decision, reason := checkSlotLease(slotGuardInput(t, "s-2", slotGuardCommand), wt, cfg, &advisory)
		if decision != DecisionDeny {
			t.Fatalf("hook root in a linked worktree: decision = %q, want deny — the guard read the worktree's empty state instead of the primary's record", decision)
		}
		if !strings.HasPrefix(reason, slotLeaseViolationPrefix) || !strings.Contains(reason, "lane-s-1") {
			t.Errorf("reason = %q, want the sentinel and the holder", reason)
		}
	})

	t.Run("primary-deny", func(t *testing.T) {
		var advisory bytes.Buffer
		if decision, _ := checkSlotLease(slotGuardInput(t, "s-2", slotGuardCommand), primary, cfg, &advisory); decision != DecisionDeny {
			t.Fatalf("hook root at the primary: decision = %q, want deny", decision)
		}
	})

	t.Run("no-git", func(t *testing.T) {
		plain := t.TempDir()
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(plain))
		var advisory bytes.Buffer
		decision, reason := checkSlotLease(slotGuardInput(t, "s-2", slotGuardCommand), plain, cfg, &advisory)
		if decision == DecisionDeny {
			t.Fatalf("unresolvable root denied (%s) — must fail open", reason)
		}
		if !strings.Contains(advisory.String(), slotAdvisoryTag) {
			t.Errorf("no advisory for an unresolvable root; got %q", advisory.String())
		}
		// N4 (plan-audit iter2): with no primary to resolve, the fail-open
		// line lands under the hook root as given.
		event, auditReason, ok := lastSlotGuardEvent(t, plain)
		if !ok || event != "fail-open" || auditReason == "" {
			t.Errorf("last audit under the unnormalized hook root = (%q, %q, present=%v), want a fail-open line", event, auditReason, ok)
		}
	})
}
