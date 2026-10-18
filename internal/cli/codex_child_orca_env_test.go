package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
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

func TestCodexChildCmdDropsOrcaPaneEnvOrdinaryAndLeg(t *testing.T) {
	setT1869FakeEnv(t)
	ordinary := codexChildCmd(context.Background(), "codex", []string{"app-server"})
	got := t1869EnvMap(ordinary.Env)
	assertT1869Env(t, got)
	if _, ok := got[codexAuditLegEnv]; ok {
		t.Fatalf("an ordinary child must not carry the audit-leg marker")
	}
	leg := codexChildCmd(withCodexAuditLeg(context.Background()), "codex", []string{"app-server"})
	lgot := t1869EnvMap(leg.Env)
	assertT1869Env(t, lgot)
	if lgot[codexAuditLegEnv] != "1" {
		t.Fatalf("the audit-leg marker must survive, got %q", lgot[codexAuditLegEnv])
	}
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
