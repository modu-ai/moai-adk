package cli

// integration_lock_owner_liveness_test.go — the cross-process reproduction of
// the integration-lock owner-liveness defect (card t298,
// SPEC-INTEGRATION-LOCK-LIVENESS-001 M1).
//
// Why this file exists at all: the pre-existing coverage in
// integration_lock_cli_test.go drives the cobra command IN-PROCESS, so the pid
// AcquireIntegrationLock records (os.Getpid(), integration_lock.go:174-176) is
// the still-alive `go test` process. Every liveness assertion there is
// therefore satisfied by accident, and the field shape — a lane's acquire CLI
// exits, and its recorded pid dies with it — cannot be expressed. These tests
// exec the real built binary as a CHILD, let it exit, and then assert from the
// still-live parent. That is the only shape in which the defect is visible.
//
// The child binary is built with `go build -o`, never `go run`: the
// intermediate `go` process is not in the ancestry walk's wrapper-name set
// (internal/session/session_pid.go) and exits early, which would break the
// walk and produce another accidental green.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// buildMoaiBinary builds cmd/moai into this test's own temp dir and returns
// the binary path. Built per test rather than once per package so cleanup is
// t.TempDir()'s job; Go's build cache makes the repeat builds link-only.
func buildMoaiBinary(t *testing.T) string {
	t.Helper()

	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	modFile := strings.TrimSpace(string(gomod))
	if modFile == "" || modFile == os.DevNull {
		t.Fatalf("go env GOMOD did not name a module file (got %q)", modFile)
	}
	moduleRoot := filepath.Dir(modFile)

	bin := filepath.Join(t.TempDir(), "moai")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", bin, "./cmd/moai")
	build.Dir = moduleRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build -o %s ./cmd/moai (in %s): %v\n%s", bin, moduleRoot, err, out)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("built binary is not present at %s: %v", bin, err)
	}
	return bin
}

// runIntegrationChild execs the built binary as a child process and returns
// its combined output plus the run error.
//
// The env pinning contract mirrors runIntegration (integration_lock_cli_test.go:26-43):
// CLAUDE_PROJECT_DIR points the lock record at the throwaway root, and
// GIT_CEILING_DIRECTORIES stops the git fallback from walking out of it into
// the developer's real checkout. CLAUDE_CODE_SESSION_ID and MOAI_SESSION_PID
// are scrubbed from the inherited environment so each case states its own
// resolution inputs; extraEnv puts back only what that case wants.
func runIntegrationChild(t *testing.T, bin, root string, extraEnv []string, args ...string) (string, error) {
	t.Helper()

	scrubbed := []string{
		"CLAUDE_PROJECT_DIR=",
		"GIT_CEILING_DIRECTORIES=",
		config.EnvClaudeCodeSessionID + "=",
		config.EnvMoaiSessionPID + "=",
	}
	env := make([]string, 0, len(os.Environ())+len(extraEnv)+2)
	for _, kv := range os.Environ() {
		drop := false
		for _, prefix := range scrubbed {
			if strings.HasPrefix(kv, prefix) {
				drop = true
				break
			}
		}
		if !drop {
			env = append(env, kv)
		}
	}
	env = append(env,
		"CLAUDE_PROJECT_DIR="+root,
		"GIT_CEILING_DIRECTORIES="+filepath.Dir(root),
	)
	env = append(env, extraEnv...)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, append([]string{"integration"}, args...)...)
	cmd.Env = env
	// Run from the throwaway root so os.Getwd() and any git probe inside the
	// child cannot reach the real repository.
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// readLockRecordRaw reads the on-disk record as a generic map.
//
// Generic on purpose: AC-INL-001 clause (b) asserts a field the baseline
// IntegrationLock struct does not carry, and decoding into the struct would
// silently drop it rather than fail.
func readLockRecordRaw(t *testing.T, root string) map[string]any {
	t.Helper()
	path := filepath.Join(root, ".moai", "state", kanban.IntegrationLockFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no lock record at %s after the child acquire: %v", path, err)
	}
	var rec map[string]any
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("lock record at %s is not valid JSON (%v): %s", path, err, data)
	}
	return rec
}

// parentStatus runs `integration status --json` from the still-live parent
// test process against the same root the child wrote.
func parentStatus(t *testing.T, root string) (held, stale bool, lock kanban.IntegrationLock, raw string) {
	t.Helper()
	out, err := runIntegration(t, root, "status", "--json")
	if err != nil {
		t.Fatalf("parent status: %v (%s)", err, out)
	}
	var status struct {
		Held  bool                   `json:"held"`
		Stale bool                   `json:"stale"`
		Lock  kanban.IntegrationLock `json:"lock"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("status --json is not valid JSON (%v): %s", err, out)
	}
	return status.Held, status.Stale, status.Lock, out
}

// TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits is
// AC-INL-001: a window acquired by a child CLI must stay HELD while the
// session that ran it is alive.
//
// RED on the baseline tree for the stated reason: AcquireIntegrationLock
// records the acquire CLI's own pid, which dies when that child exits, so the
// parent reads the window as reclaimable and the owner-anchor marker
// (pid_source) is absent from the record entirely.
func TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits(t *testing.T) {
	bin := buildMoaiBinary(t)
	root := t.TempDir()

	// No MOAI_SESSION_PID: this case exercises the ancestry walk, whose
	// nearest non-wrapper ancestor is this still-live test process.
	out, err := runIntegrationChild(t, bin, root, nil, "acquire", "--session", "sess-owner-a", "--name", "lane-owner-a")
	if err != nil {
		t.Fatalf("child acquire failed: %v\n%s", err, out)
	}
	// The child has exited by the time CombinedOutput returns; its pid is dead.

	held, stale, lock, rawStatus := parentStatus(t, root)

	// Clause (a): the window is held by a live owner, not reclaimable.
	if !held {
		t.Errorf("window reads free after the child acquire: %s", rawStatus)
	}
	if stale {
		t.Errorf("window reads reclaimable while the owning session (this test process, pid %d) is alive; "+
			"recorded pid %d is the exited acquire CLI's: %s", os.Getpid(), lock.PID, rawStatus)
	}
	if lock.PID != os.Getpid() {
		t.Errorf("recorded pid = %d, want the owning session's pid %d", lock.PID, os.Getpid())
	}

	// Clause (b): the record carries the owner-anchor marker.
	rec := readLockRecordRaw(t, root)
	if got, _ := rec["pid_source"].(string); got != "session-owner" {
		t.Errorf("record pid_source = %q, want \"session-owner\"; record = %v", got, rec)
	}
}

// TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits is
// AC-INL-002: the explicit MOAI_SESSION_PID stamp must be the recorded owner.
//
// This variant needs no ancestry walk, so it is the platform-neutral one and
// doubles as the Windows-behavior proxy. RED on the baseline tree because the
// acquire path ignores MOAI_SESSION_PID entirely and records the child's pid.
func TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits(t *testing.T) {
	bin := buildMoaiBinary(t)
	root := t.TempDir()

	parentPID := os.Getpid()
	env := []string{fmt.Sprintf("%s=%d", config.EnvMoaiSessionPID, parentPID)}
	out, err := runIntegrationChild(t, bin, root, env, "acquire", "--session", "sess-owner-b", "--name", "lane-owner-b")
	if err != nil {
		t.Fatalf("child acquire failed: %v\n%s", err, out)
	}

	held, stale, lock, rawStatus := parentStatus(t, root)

	if !held {
		t.Errorf("window reads free after the child acquire: %s", rawStatus)
	}
	if stale {
		t.Errorf("window reads reclaimable while the stamped owner (pid %d) is alive: %s", parentPID, rawStatus)
	}
	if lock.PID != parentPID {
		t.Errorf("recorded pid = %d, want the stamped MOAI_SESSION_PID %d", lock.PID, parentPID)
	}
}

// TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder is AC-INL-012: a
// bare acquire must not transfer a LIVE holder's window.
//
// This is the 2026-08-27 field harm made mechanical. RED on the baseline tree
// because child A's recorded pid is its own and dead, so acquire takes the
// `case current.Stale():` reclaim arm (integration_lock.go:163): child B
// succeeds and reports `displaced: sess-a` while lane A is still working.
func TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder(t *testing.T) {
	bin := buildMoaiBinary(t)
	root := t.TempDir()

	outA, err := runIntegrationChild(t, bin, root, nil, "acquire", "--session", "sess-a", "--name", "lane-a")
	if err != nil {
		t.Fatalf("child A acquire failed: %v\n%s", err, outA)
	}

	outB, errB := runIntegrationChild(t, bin, root, nil, "acquire", "--session", "sess-b", "--name", "lane-b")

	// Clause (a): the second acquire is refused, naming the holder, and the
	// record still belongs to A.
	if errB == nil {
		t.Errorf("bare acquire by sess-b took the window from live holder sess-a (exit 0): %s", outB)
	}
	if !strings.Contains(outB, "sess-a") && !strings.Contains(outB, "lane-a") {
		t.Errorf("refusal does not name the holder sess-a/lane-a: %s", outB)
	}
	rec := readLockRecordRaw(t, root)
	if got, _ := rec["session_id"].(string); got != "sess-a" {
		t.Errorf("on-disk session_id = %q, want \"sess-a\" (the window must not have been transferred); record = %v", got, rec)
	}

	// Clause (b): a refusal produces no displacement bookkeeping.
	if strings.Contains(outB, "displaced") {
		t.Errorf("refused acquire reported a displacement; displaced-holder bookkeeping belongs to --force and stale-reclaim only: %s", outB)
	}
}
