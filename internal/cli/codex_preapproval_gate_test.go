//go:build !windows

package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	envCodexPreApprovalLive = "MOAI_CODEX_PREAPPROVAL_LIVE"
	envT1172EvidenceDir     = "MOAI_T1172_EVIDENCE_DIR"
	validDiff               = "diff -r inputs/control/project-config.toml inputs/treatment/project-config.toml\n7a8,9\n> [mcp_servers.moai.tools.codex_role_audit]\n> approval_mode = \"approve\"\n"
)

type preApprovalLedgerRow struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason,omitempty"`
}

// preApprovalGate stops before any LIVE call when exported inputs, the arm
// comparison, or the non-model startup check fail. The caller owns the actual
// Codex invocation; the gate cannot create it until every check succeeds.
func preApprovalGate(inputs []string, diffOutput, startupOutput string, invoke func() error) []preApprovalLedgerRow {
	stop := func(reason string) []preApprovalLedgerRow {
		return []preApprovalLedgerRow{{Kind: "stop", Reason: reason}}
	}
	for _, path := range inputs {
		if _, err := os.Stat(path); err != nil {
			return stop("missing input: " + path)
		}
	}
	if !preApprovalOnlyProjectConfigDiff(diffOutput) {
		return stop("arm diff exceeds project config table")
	}
	if strings.Contains(startupOutput, "Not inside a trusted directory") ||
		strings.Contains(startupOutput, `"type":"item.started"`) ||
		strings.Contains(startupOutput, `"type":"item.completed"`) {
		return stop("startup check failed")
	}
	if err := invoke(); err != nil {
		return stop("LIVE launch failed: " + err.Error())
	}
	return []preApprovalLedgerRow{{Kind: "live"}}
}

func preApprovalOnlyProjectConfigDiff(output string) bool {
	if !strings.HasPrefix(output, "diff -r inputs/control/project-config.toml inputs/treatment/project-config.toml\n") {
		return false
	}
	var additions []string
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n")[1:] {
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "Only in ") {
			return false
		}
		if strings.HasPrefix(line, "< ") || strings.HasPrefix(line, "! ") {
			return false
		}
		if strings.HasPrefix(line, "> ") && strings.TrimSpace(line) != ">" {
			additions = append(additions, strings.TrimPrefix(line, "> "))
		}
	}
	return len(additions) == 2 && additions[0] == "[mcp_servers.moai.tools.codex_role_audit]" &&
		additions[1] == `approval_mode = "approve"`
}

// The destructive clean is intentionally inaccessible until every assertion
// succeeds. The injected command lets the negative test prove zero git calls.
func preApprovalCleanFixture(root, tempRoot, sentinel string, worktrees []string, gitCommand func(...string) (string, error)) error {
	if root == "" || !filepath.IsAbs(root) || tempRoot == "" || !filepath.IsAbs(tempRoot) {
		return errors.New("fixture root must be an absolute path")
	}
	root, tempRoot = filepath.Clean(root), filepath.Clean(tempRoot)
	systemTemp, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		return fmt.Errorf("resolve temporary directory: %w", err)
	}
	within := func(path, parent string) bool {
		rel, err := filepath.Rel(parent, path)
		return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	if !within(root, tempRoot) || !within(root, systemTemp) {
		return errors.New("fixture root is outside temporary directory")
	}
	for _, tree := range worktrees {
		if root == filepath.Clean(tree) || within(root, filepath.Clean(tree)) {
			return errors.New("fixture root overlaps repository worktree")
		}
	}
	actual, err := os.ReadFile(filepath.Join(root, ".fixture-sentinel"))
	if err != nil || string(actual) != sentinel || sentinel == "" {
		return errors.New("fixture sentinel mismatch")
	}
	top, err := gitCommand("-C", root, "rev-parse", "--show-toplevel")
	if err != nil || filepath.Clean(strings.TrimSpace(top)) != root {
		return errors.New("fixture is not git top-level")
	}
	_, err = gitCommand("-C", root, "clean", "-ffdx")
	return err
}

func preApprovalSentinel(t *testing.T, root string) string {
	t.Helper()
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	token := hex.EncodeToString(buf)
	if err := os.WriteFile(filepath.Join(root, ".fixture-sentinel"), []byte(token), 0o600); err != nil {
		t.Fatal(err)
	}
	return token
}

func TestCodexPreApprovalGateRefusals(t *testing.T) {
	newInputs := func(t *testing.T) []string {
		t.Helper()
		root := t.TempDir()
		var paths []string
		for _, name := range []string{"codex-home-config.toml", "project-config.toml", "argv.txt", "prompt.txt", "root-state.txt"} {
			path := filepath.Join(root, name)
			if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
				t.Fatal(err)
			}
			paths = append(paths, path)
		}
		return paths
	}
	assertStopped := func(t *testing.T, rows []preApprovalLedgerRow, count int) {
		t.Helper()
		b, err := json.Marshal(rows)
		if err != nil || len(rows) != 1 || rows[0].Kind != "stop" || rows[0].Reason == "" || count != 0 {
			t.Fatalf("want stop row and zero LIVE launches; rows=%s count=%d err=%v", b, count, err)
		}
	}
	for _, tc := range []struct {
		name, diff, startup string
		remove              bool
	}{
		{"missing_input", validDiff, "startup safe", true},
		{"arm_diff_exceeds", "diff -r inputs/control/argv.txt inputs/treatment/argv.txt\n1a2\n> extra\n", "startup safe", false},
		{"startup_check_fails", validDiff, "Not inside a trusted directory", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputs := newInputs(t)
			fakeDir := t.TempDir()
			launchLog := filepath.Join(fakeDir, "launches")
			fakeCodex := filepath.Join(fakeDir, "codex")
			script := fmt.Sprintf("#!/bin/sh\ncase \" $* \" in\n  *--strict-config*) printf 'STARTUP\\n' >> %q; printf 'Not inside a trusted directory\\n' >&2 ;;\n  *) printf 'LIVE\\n' >> %q ;;\nesac\n", launchLog, launchLog)
			if err := os.WriteFile(fakeCodex, []byte(script), 0o700); err != nil {
				t.Fatal(err)
			}
			if tc.remove {
				if err := os.Remove(inputs[0]); err != nil {
					t.Fatal(err)
				}
			}
			startupOutput := tc.startup
			if tc.name == "startup_check_fails" {
				out, err := exec.Command(fakeCodex, "exec", "--strict-config", "-c", `model_provider="dead"`).CombinedOutput()
				if err != nil {
					t.Fatal(err)
				}
				startupOutput = string(out)
			}
			rows := preApprovalGate(inputs, tc.diff, startupOutput, func() error {
				return exec.Command(fakeCodex, "exec", "--json", "LIVE").Run()
			})
			log, err := os.ReadFile(launchLog)
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			count := bytes.Count(log, []byte("LIVE\n"))
			if tc.name == "startup_check_fails" && bytes.Count(log, []byte("STARTUP\n")) != 1 {
				t.Fatalf("fake Codex startup launch log=%q", log)
			}
			assertStopped(t, rows, count)
		})
	}
	t.Run("clean_refuses_non_fixture_root", func(t *testing.T) {
		temp := t.TempDir()
		valid := filepath.Join(temp, "fixture")
		if err := os.Mkdir(valid, 0o700); err != nil {
			t.Fatal(err)
		}
		token := preApprovalSentinel(t, valid)
		count := 0
		git := func(args ...string) (string, error) { count++; return valid, nil }
		for _, tc := range []struct{ root, token string }{
			{"", token}, {"relative", token}, {filepath.Dir(temp), token}, {valid, "wrong"},
		} {
			if err := preApprovalCleanFixture(tc.root, temp, tc.token, nil, git); err == nil {
				t.Fatalf("unsafe root accepted: %q", tc.root)
			}
		}
		if count != 0 {
			t.Fatalf("unsafe roots launched git %d times", count)
		}
	})
	t.Run("kills_only_recorded_pids", func(t *testing.T) {
		procs := newLiveProcs(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		t.Cleanup(cancel)
		recorded := liveCommand(ctx, "sleep", "4")
		if err := procs.start(recorded); err != nil {
			t.Fatal(err)
		}
		other := exec.CommandContext(ctx, "sleep", "4")
		if err := other.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = other.Process.Kill(); _ = other.Wait() })
		killed := procs.reap()
		if len(killed) != 1 || killed[0] != recorded.Process.Pid {
			t.Fatalf("cleaned pids=%v, want [%d]", killed, recorded.Process.Pid)
		}
		if err := other.Process.Signal(syscall.Signal(0)); err != nil {
			t.Fatalf("unrecorded process already dead: %v", err)
		}
	})
}
