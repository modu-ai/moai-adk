package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// t1869: a codex review/audit/task child must not inherit Orca pane-owned
// variables. Every value below is a fake set with t.Setenv; no real token or
// endpoint is read or written.

const t1869DumpEnv = "T1869_ENV_DUMP"

// Fake pane-owned variables (must be dropped) and kept variables (must survive).
var (
	t1869Dropped = []string{
		"ORCA_PANE_KEY", "ORCA_TAB_ID", "ORCA_WORKTREE_ID", "ORCA_AGENT_LAUNCH_TOKEN",
		"ORCA_AGENT_PANE", "ORCA_TERMINAL_HANDLE",
		"ORCA_AGENT_HOOK_ENDPOINT", "ORCA_AGENT_HOOK_ENV", "ORCA_AGENT_HOOK_PORT",
		"ORCA_AGENT_HOOK_TOKEN", "ORCA_AGENT_HOOK_TRANSPORT", "ORCA_AGENT_HOOK_VERSION",
		"ORCA_AGENT_HOOK_OPENCODE_TUI",
	}
	t1869Kept = []string{
		"ORCA_APP_VERSION", "ORCA_CLI_BIN_DIR", "ORCA_USER_DATA_PATH", "ORCA_IMAGE_PROTOCOL",
		"ORCA_OMP_FAKE", "ORCA_PI_SOURCE_AGENT_DIR", "ORCA_OPENCODE_AGENT",
		"T1869_SENTINEL",
	}
)

func setT1869FakeEnv(t *testing.T) {
	t.Helper()
	for _, k := range t1869Dropped {
		t.Setenv(k, "fake-"+k)
	}
	for _, k := range t1869Kept {
		t.Setenv(k, "fake-"+k)
	}
}

// t1869EnvMap turns a KEY=VALUE slice into a map.
func t1869EnvMap(env []string) map[string]string {
	m := map[string]string{}
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}

func assertT1869Env(t *testing.T, got map[string]string) {
	t.Helper()
	for _, k := range t1869Dropped {
		if _, ok := got[k]; ok {
			t.Errorf("pane-owned %s reached the codex child", k)
		}
	}
	for _, k := range t1869Kept {
		if got[k] != "fake-"+k {
			t.Errorf("%s must survive into the codex child, got %q", k, got[k])
		}
	}
}

// t1869FakeCodex writes a stand-in codex binary that dumps its environment to a
// file and exits 0 after printing stdout. It returns the binary and the dump path.
func t1869FakeCodex(t *testing.T, stdout string) (string, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stand-in binary")
	}
	dir := t.TempDir()
	dump := filepath.Join(dir, "env.txt")
	bin := filepath.Join(dir, "codex")
	script := "#!/bin/sh\nenv > \"$" + t1869DumpEnv + "\"\ncat >/dev/null 2>&1 </dev/null\nprintf '%s' '" + stdout + "'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(t1869DumpEnv, dump)
	return bin, dump
}

func readT1869Dump(t *testing.T, dump string) map[string]string {
	t.Helper()
	b, err := os.ReadFile(dump)
	if err != nil {
		t.Fatalf("the fake codex did not run: %v", err)
	}
	return t1869EnvMap(strings.Split(strings.TrimRight(string(b), "\n"), "\n"))
}

func TestCodexSingleShotRunnerDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "ok")
	if _, err := (realCodexRunner{}).run(context.Background(), bin, nil, ""); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertT1869Env(t, readT1869Dump(t, dump))
}

func TestCodexLoginStatusDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "Logged in")
	if _, _, _, err := defaultLoginStatusRunner(context.Background(), bin); err != nil {
		t.Fatalf("login status: %v", err)
	}
	assertT1869Env(t, readT1869Dump(t, dump))
}

func TestCodexAuditExecDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "")
	_, err := codexAuditExec(context.Background(), bin, nil, t.TempDir(), strings.NewReader(""), &bytes.Buffer{}, 30*time.Second)
	if err != nil {
		t.Fatalf("codexAuditExec: %v", err)
	}
	assertT1869Env(t, readT1869Dump(t, dump))
}

func TestCodexAuditMCPListDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "[]")
	if _, err := codexAuditMCPDisableArgs(context.Background(), bin, t.TempDir()); err != nil {
		t.Fatalf("codexAuditMCPDisableArgs: %v", err)
	}
	assertT1869Env(t, readT1869Dump(t, dump))
}

// unsetenvForTest removes key for the test and restores its previous value AND
// presence afterwards. t.Setenv(key, "") would leave it set to empty, which is a
// different state from unset.
func unsetenvForTest(t *testing.T, key string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// t1869 production semantics under test: an ordinary codex child inherits the
// parent's audit-leg marker as it is (nothing added, nothing removed); a leg
// child gets MOAI_CODEX_AUDIT_LEG=1 appended last, so it wins over any ambient value.
func TestCodexChildCmdDropsOrcaPaneEnvOrdinaryAndLeg(t *testing.T) {
	for _, amb := range []struct {
		name    string
		set     bool
		value   string
		ordWant string // marker value an ordinary child must show ("" = absent)
	}{
		{"ambient-unset", false, "", ""},
		{"ambient-1", true, "1", "1"},
		{"ambient-0", true, "0", "0"},
	} {
		t.Run(amb.name, func(t *testing.T) {
			unsetenvForTest(t, codexAuditLegEnv)
			if amb.set {
				t.Setenv(codexAuditLegEnv, amb.value)
			}
			setT1869FakeEnv(t)

			ord := t1869EnvMap(codexChildCmd(context.Background(), "codex", []string{"app-server"}).Env)
			assertT1869Env(t, ord)
			v, present := ord[codexAuditLegEnv]
			if amb.ordWant == "" && present {
				t.Errorf("an ordinary child with no ambient marker must not carry one, got %q", v)
			}
			if amb.ordWant != "" && v != amb.ordWant {
				t.Errorf("an ordinary child must inherit the ambient marker %q, got %q", amb.ordWant, v)
			}

			legEnv := codexChildCmd(withCodexAuditLeg(context.Background()), "codex", []string{"app-server"}).Env
			assertT1869Env(t, t1869EnvMap(legEnv))
			last := ""
			for _, e := range legEnv {
				if k, val, _ := strings.Cut(e, "="); k == codexAuditLegEnv {
					last = val
				}
			}
			if last != "1" {
				t.Errorf("a leg child must end with %s=1 (last entry wins), got %q", codexAuditLegEnv, last)
			}
		})
	}
}

// t1869Expected is the independent statement of the contract: env minus the
// six pane keys and the ORCA_AGENT_HOOK_ family, nothing else.
func t1869Expected(env []string) []string {
	var out []string
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		skip := strings.HasPrefix(k, "ORCA_AGENT_HOOK_")
		for _, d := range t1869Dropped {
			if k == d {
				skip = true
			}
		}
		if !skip {
			out = append(out, e)
		}
	}
	return out
}

func assertT1869SameSet(t *testing.T, got, want []string) {
	t.Helper()
	g, w := append([]string(nil), got...), append([]string(nil), want...)
	slices.Sort(g)
	slices.Sort(w)
	if !slices.Equal(g, w) {
		t.Fatalf("child env differs from parent env minus exactly the dropped keys:\n got %d entries, want %d\n only in got: %v\n only in want: %v",
			len(g), len(w), t1869Diff(g, w), t1869Diff(w, g))
	}
}

func t1869Diff(a, b []string) []string {
	var d []string
	for _, x := range a {
		if !slices.Contains(b, x) {
			k, _, _ := strings.Cut(x, "=")
			d = append(d, k) // key only: never print values
		}
	}
	return d
}

// An ordinary child's env is exactly the parent env minus the dropped keys.
func TestCodexChildCmdEnvIsParentMinusDroppedKeys(t *testing.T) {
	setT1869FakeEnv(t)
	got := codexChildCmd(context.Background(), "codex", []string{"app-server"}).Env
	assertT1869SameSet(t, got, t1869Expected(os.Environ()))
}

// The app-server spawn path (review gate, codex_task, MCP codex tools).
func TestCodexSessionRunnerStartDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := (realCodexSessionRunner{}).start(ctx, bin, nil)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	for {
		if _, ok := conn.recv(); !ok { // EOF: the fake exited after dumping its env
			break
		}
	}
	_ = conn.close()
	assertT1869Env(t, readT1869Dump(t, dump))
}

// The claude audit / participation child env builder (leader ruling, same cause
// family). Fixed synthetic input keeps it hermetic.
func TestScrubClaudeAuditEnvDropsOnlyOrcaPaneKeysAdditionally(t *testing.T) {
	in := []string{"PATH=/bin", "HOME=/h", "T1869_SENTINEL=s", "ORCA_APP_VERSION=1", "ANTHROPIC_API_KEY=k"}
	want := []string{"PATH=/bin", "HOME=/h", "T1869_SENTINEL=s", "ORCA_APP_VERSION=1"} // ANTHROPIC_* is the existing scrub
	for _, k := range t1869Dropped {
		in = append(in, k+"=v")
	}
	assertT1869SameSet(t, scrubClaudeAuditEnv(in), want)
}

func TestClaudeAuditChildDropsOrcaPaneEnv(t *testing.T) {
	setT1869FakeEnv(t)
	bin, dump := t1869FakeCodex(t, "")
	env := scrubClaudeAuditEnv(os.Environ())
	if _, _, err := (realClaudeRunner{}).RunAudit(context.Background(), bin, t.TempDir(), nil, env, nil); err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	assertT1869Env(t, readT1869Dump(t, dump))
}

func TestCodexChildCmdKeepsLaneIdentityDropOfLauncherEnv(t *testing.T) {
	// codexChildEnv (the interactive launcher env) must still keep the Orca
	// variables a lane session legitimately runs with: only the review path scrubs.
	setT1869FakeEnv(t)
	got := t1869EnvMap(codexChildEnv())
	for _, k := range t1869Dropped {
		if got[k] != "fake-"+k {
			t.Errorf("the launcher env must keep %s, got %q", k, got[k])
		}
	}
}

// The matcher semantics (t1869): exact, case-sensitive, anchored keys plus the
// ORCA_AGENT_HOOK_ prefix. Near-miss keys are decided, not left implicit.
func TestIsOrcaPaneOwnedEnvKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{"ORCA_PANE_KEY", true},
		{"ORCA_TAB_ID", true},
		{"ORCA_WORKTREE_ID", true},
		{"ORCA_AGENT_LAUNCH_TOKEN", true},
		{"ORCA_AGENT_PANE", true},
		{"ORCA_TERMINAL_HANDLE", true},
		{"ORCA_AGENT_HOOK_ENDPOINT", true},
		{"ORCA_AGENT_HOOK_OPENCODE_TUI", true},
		{"ORCA_AGENT_HOOK_", true}, // the bare prefix is still the hook family
		{"ORCA_PANE_KEYS", false},  // longer exact-key lookalike is kept
		{"XORCA_PANE_KEY", false},  // not anchored at the start
		{"orca_pane_key", false},   // case-sensitive
		{"ORCA_AGENT_HOOK", false}, // no trailing underscore: not the family
		{"ORCA_AGENT_PANES", false},
		{"ORCA_APP_VERSION", false},
		{"ORCA_CLI_BIN_DIR", false},
		{"ORCA_USER_DATA_PATH", false},
		{"ORCA_IMAGE_PROTOCOL", false},
		{"ORCA_OMP_MODE", false},
		{"ORCA_PI_SOURCE_AGENT_DIR", false},
		{"ORCA_OPENCODE_AGENT", false},
		{"PATH", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isOrcaPaneOwnedEnvKey(c.key); got != c.want {
			t.Errorf("isOrcaPaneOwnedEnvKey(%q) = %v, want %v", c.key, got, c.want)
		}
	}
}

func TestDropOrcaPaneEnvKeepsOrderAndValuesAndDoesNotMutateInput(t *testing.T) {
	in := []string{"A=1", "ORCA_PANE_KEY=x=y", "B=2=3", "ORCA_AGENT_HOOK_TOKEN=t", "ORCA_APP_VERSION=9", "NOEQUALS"}
	snapshot := append([]string(nil), in...)
	got := dropOrcaPaneEnv(in)
	want := []string{"A=1", "B=2=3", "ORCA_APP_VERSION=9", "NOEQUALS"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %q, want %q", got, want)
	}
	if strings.Join(in, "\x00") != strings.Join(snapshot, "\x00") {
		t.Fatalf("input mutated: %q", in)
	}
}

// codexAuditChildEnv composes both drops: lane identity AND pane ownership.
func TestCodexAuditChildEnvComposesLaneAndOrcaDrops(t *testing.T) {
	setT1869FakeEnv(t)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "fake-session")
	got := t1869EnvMap(codexAuditChildEnv())
	if _, ok := got["CLAUDE_CODE_SESSION_ID"]; ok {
		t.Error("the lane identity drop of codexChildEnv must still apply")
	}
	assertT1869Env(t, got)
}
