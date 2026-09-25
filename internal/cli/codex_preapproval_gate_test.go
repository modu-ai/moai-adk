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
	"regexp"
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

type preApprovalStartupProof struct {
	Fixture  string
	Stdout   []byte
	Stderr   []byte
	Launches map[string]int
}

func preApprovalStartupValid(proof preApprovalStartupProof) bool {
	if proof.Launches["moai"] < 1 || (proof.Fixture == "car010" && proof.Launches["decoy"] < 1) {
		return false
	}
	if bytes.Contains(proof.Stderr, []byte("Error loading config")) ||
		bytes.Contains(proof.Stderr, []byte("Error: ")) ||
		bytes.Contains(proof.Stderr, []byte("unknown configuration field")) ||
		bytes.Contains(proof.Stderr, []byte("unknown variant")) ||
		bytes.Contains(proof.Stderr, []byte("Not inside a trusted directory")) {
		return false
	}
	threadStarts := 0
	for _, line := range bytes.Split(bytes.TrimSpace(proof.Stdout), []byte("\n")) {
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
			} `json:"item"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			return false
		}
		if event.Type == "thread.started" {
			threadStarts++
		}
		if strings.HasPrefix(event.Type, "item.") && event.Item.Type != "error" {
			return false
		}
	}
	return threadStarts == 1
}

// preApprovalGate stops before any LIVE call when exported inputs, the arm
// comparison, or the non-model startup check fail. The caller owns the actual
// Codex invocation; the gate cannot create it until every check succeeds.
func preApprovalGate(inputs []string, diffOutput string, startup preApprovalStartupProof, invoke func() error) []preApprovalLedgerRow {
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
	if !preApprovalStartupValid(startup) {
		return stop("startup check failed")
	}
	if err := invoke(); err != nil {
		return stop("LIVE launch failed: " + err.Error())
	}
	return []preApprovalLedgerRow{{Kind: "live"}}
}

var preApprovalHunk = regexp.MustCompile(`^[0-9]+a[0-9]+(,[0-9]+)?$`)

func preApprovalOnlyProjectConfigDiff(output string) bool {
	if !strings.HasSuffix(output, "\n") {
		return false
	}
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 4 && len(lines) != 5 {
		return false
	}
	if lines[0] != "diff -r inputs/control/project-config.toml inputs/treatment/project-config.toml" ||
		!preApprovalHunk.MatchString(lines[1]) {
		return false
	}
	if len(lines) == 5 {
		if lines[2] != "> " {
			return false
		}
		lines = append(lines[:2], lines[3:]...)
	}
	return lines[2] == "> [mcp_servers.moai.tools.codex_role_audit]" &&
		lines[3] == `> approval_mode = "approve"`
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
	validStartup := preApprovalStartupProof{Fixture: "disc-control", Stdout: []byte("{\"type\":\"thread.started\"}\n"), Launches: map[string]int{"moai": 1}}
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
		name, diff string
		remove     bool
	}{
		{"missing_input", validDiff, true},
		{"arm_diff_exceeds", "diff -r inputs/control/argv.txt inputs/treatment/argv.txt\n1a2\n> extra\n", false},
		{"startup_check_fails", validDiff, false},
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
			startup := validStartup
			if tc.name == "startup_check_fails" {
				out, err := exec.Command(fakeCodex, "exec", "--strict-config", "-c", `model_provider="dead"`).CombinedOutput()
				if err != nil {
					t.Fatal(err)
				}
				startup.Stderr = out
			}
			rows := preApprovalGate(inputs, tc.diff, startup, func() error {
				return exec.Command(fakeCodex, "exec", "--json", "LIVE").Run()
			})
			if tc.name == "arm_diff_exceeds" {
				diffRoot := t.TempDir()
				for _, arm := range []string{"control", "treatment"} {
					if err := os.MkdirAll(filepath.Join(diffRoot, "inputs", arm), 0o700); err != nil {
						t.Fatal(err)
					}
					project := []byte("base\n")
					binary := []byte{0, 1}
					if arm == "treatment" {
						project = append(project, []byte("\n[mcp_servers.moai.tools.codex_role_audit]\napproval_mode = \"approve\"\n")...)
						binary = []byte{0, 2}
					}
					if err := os.WriteFile(filepath.Join(diffRoot, "inputs", arm, "project-config.toml"), project, 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(diffRoot, "inputs", arm, "z.bin"), binary, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				diff := exec.Command("diff", "-r", "inputs/control", "inputs/treatment")
				diff.Dir = diffRoot
				actualDiff, diffErr := diff.CombinedOutput()
				if e, ok := diffErr.(*exec.ExitError); !ok || e.ExitCode() != 1 || !bytes.Contains(actualDiff, []byte("Binary files ")) {
					t.Fatalf("binary diff fixture failed: %v: %s", diffErr, actualDiff)
				}
				for _, mutation := range []string{
					string(actualDiff),
					validDiff + "Only in inputs/control: extra.txt\n",
				} {
					mutantRows := preApprovalGate(inputs, mutation, validStartup, func() error {
						return exec.Command(fakeCodex, "exec", "--json", "LIVE").Run()
					})
					if len(mutantRows) != 1 || mutantRows[0].Kind != "stop" {
						t.Fatalf("extra arm diff accepted: %q", mutation)
					}
				}
			}
			if tc.name == "startup_check_fails" {
				for _, mutation := range []preApprovalStartupProof{
					{},
					{Fixture: "disc-control", Stdout: []byte("arbitrary text"), Launches: map[string]int{"moai": 1}},
					{Fixture: "disc-control", Stdout: validStartup.Stdout},
					{Fixture: "disc-control", Stdout: []byte("{\"type\":\"thread.started\"}\n{\"type\":\"thread.started\"}\n"), Launches: map[string]int{"moai": 1}},
					{Fixture: "disc-control", Stdout: validStartup.Stdout, Stderr: []byte("Error: unknown variant `bogus`"), Launches: map[string]int{"moai": 1}},
					{Fixture: "disc-control", Stdout: []byte("{\"type\":\"thread.started\"}\n{\"type\":\"item.completed\",\"item\":{\"type\":\"command_execution\"}}\n"), Launches: map[string]int{"moai": 1}},
					{Fixture: "car010", Stdout: validStartup.Stdout, Launches: map[string]int{"moai": 1}},
				} {
					mutantRows := preApprovalGate(inputs, validDiff, mutation, func() error {
						return exec.Command(fakeCodex, "exec", "--json", "LIVE").Run()
					})
					if len(mutantRows) != 1 || mutantRows[0].Kind != "stop" {
						t.Fatalf("invalid startup proof accepted: %+v", mutation)
					}
				}
				positiveCalls := 0
				positive := preApprovalGate(inputs, validDiff, validStartup, func() error { positiveCalls++; return nil })
				if positiveCalls != 1 || len(positive) != 1 || positive[0].Kind != "live" {
					t.Fatalf("valid startup proof did not reach callback: rows=%v calls=%d", positive, positiveCalls)
				}
			}
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

func TestCodexPreApprovalArgvShape(t *testing.T) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		t.Skipf("CODEX_NOT_INSTALLED: %v", err)
	}
	root := filepath.Join(t.TempDir(), "fixture")
	prompt := preApprovalDiscriminatorPrompt(root)
	args := preApprovalDiscriminatorArgs(root, prompt)
	want := []string{"exec", "--strict-config", "-s", "workspace-write", "-c", `approval_policy="never"`, "-C", root, "--json", prompt}
	if len(args) != len(want) || !bytes.Equal(preApprovalArgvBytes(args), preApprovalArgvBytes(want)) {
		t.Fatalf("discriminator argv shape changed: %q", args)
	}
	if !preApprovalArgvMatches(root, []byte(prompt), preApprovalArgvBytes(args)) {
		t.Fatal("valid exported argv rejected")
	}
	mutations := []struct {
		name, root, prompt string
		args               []string
	}{
		{"unsupported_flag", root, prompt, append(append([]string{}, args[:1]...), append([]string{"--ask-for-approval", "never"}, args[1:]...)...)},
		{"missing_never", root, prompt, append(append([]string{}, args[:5]...), args[6:]...)},
		{"root_mismatch", filepath.Join(t.TempDir(), "other"), prompt, append([]string{}, args...)},
		{"prompt_mismatch", root, prompt + " changed", append([]string{}, args...)},
	}
	for _, mutation := range mutations {
		if preApprovalArgvMatches(mutation.root, []byte(mutation.prompt), preApprovalArgvBytes(mutation.args)) {
			t.Fatalf("invalid argv accepted: %s", mutation.name)
		}
	}
	bad := exec.Command(bin, "exec", "--ask-for-approval", "never", "--help")
	badOutput, badErr := bad.CombinedOutput()
	if badErr == nil || !bytes.Contains(badOutput, []byte("unexpected argument")) {
		t.Fatalf("unsupported exec approval spelling unexpectedly accepted: %v: %s", badErr, badOutput)
	}
	good := exec.Command(bin, "exec", "--strict-config", "-s", "workspace-write", "-c", `approval_policy="never"`, "--help")
	goodOutput, goodErr := good.CombinedOutput()
	if goodErr != nil || !bytes.Contains(goodOutput, []byte("Run Codex non-interactively")) {
		t.Fatalf("supported argv prefix rejected: %v: %s", goodErr, goodOutput)
	}
}
