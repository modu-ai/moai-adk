//go:build !windows

// codex_audit_live_test.go — LIVE criteria for the Codex audit launcher:
// AC-CAR-010 (launcher contract, measured route, MCP disabling, 3 calls) and
// AC-CAR-011 (the remaining read-only roles' writes are denied, 2 calls).
//
// Both are gated behind MOAI_CODEX_ROLE_LIVE=1 and MOAI_T1143_EVIDENCE_DIR and
// SKIP with NOT_RUN otherwise. Every route and probe field is computed by
// deriveCodexAuditEvidence from the session records the item created, never
// written by hand. The login is linked, never copied, and its sha256 is
// recorded before and after every invocation.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

const (
	envT1143EvidenceDir = "MOAI_T1143_EVIDENCE_DIR"

	codexAuditLiveCallBound  = 330 * time.Second
	codexAuditLive010Budget  = 3
	codexAuditLive010Window  = 1100 * time.Second // under the 1200s go test timeout
	codexAuditLive011Budget  = 2
	codexAuditLive011Window  = 800 * time.Second // under the 900s go test timeout
	codexAuditLiveRoutedPath = ".moai/reports/verdicts/sync-auditor.txt"
)

// codexAuditLiveFixture is the isolated repository, CODEX_HOME, and binaries.
type codexAuditLiveFixture struct {
	evDir, root, codexHome, realAuth, codexBin, launchLog string
	procs                                                 *liveProcs
	ledger                                                []codexLiveLedgerEntry
}

func newCodexAuditLiveFixture(t *testing.T, withMCP bool) *codexAuditLiveFixture {
	t.Helper()
	if os.Getenv(envCodexRoleLive) != "1" {
		t.Skip("NOT_RUN " + envCodexRoleLive + " is not 1: the live Codex audit run is gated off")
	}
	rawDir := os.Getenv(envT1143EvidenceDir)
	if strings.TrimSpace(rawDir) == "" {
		t.Skip("NOT_RUN " + envT1143EvidenceDir + " is empty: LIVE evidence has no channel")
	}
	evDir, err := filepath.Abs(rawDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evDir, 0o755); err != nil {
		t.Fatal(err)
	}
	codexBin, err := exec.LookPath("codex")
	if err != nil {
		t.Skip("NOT_RUN codex binary not on PATH")
	}
	f := &codexAuditLiveFixture{evDir: evDir, codexBin: codexBin, procs: newLiveProcs(t)}
	t.Setenv(config.EnvHome, t.TempDir())
	f.root = canonicalDir(t, t.TempDir())
	auditGit(t, f.root, "init", "-q", "-b", "main")
	auditGit(t, f.root, "commit", "-q", "--allow-empty", "-m", "init")
	installAuditRoles(t, f.root)
	f.codexHome, f.realAuth = linkedCodexHome(t, f.root)
	t.Setenv(codexHomeEnvVar, f.codexHome)

	if withMCP {
		moaiBin := buildLiveMoai(t, f.procs)
		bin := t.TempDir()
		f.launchLog = filepath.Join(bin, "mcp-launch.log")
		for _, label := range []string{"moai", "decoy"} {
			script := fmt.Sprintf("#!/bin/sh\nprintf '%%s %%s\\n' %s \"$$\" >> %q\nexec %q mcp-server\n", label, f.launchLog, moaiBin)
			if err := os.WriteFile(filepath.Join(bin, label+"-wrapper"), []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// Project layer: the real wiring's shape (moai server, writes
		// approval) plus sandbox_mode workspace-write, which -s must beat.
		// CODEX_HOME rides env_vars so the launcher's child stays isolated.
		project := fmt.Sprintf("sandbox_mode = \"workspace-write\"\n\n[mcp_servers.moai]\ncommand = %q\nargs = [\"mcp-server\"]\nenv_vars = [\"CODEX_HOME\", \"MOAI_HOME\"]\ndefault_tools_approval_mode = \"writes\"\n", filepath.Join(bin, "moai-wrapper"))
		if err := os.WriteFile(filepath.Join(f.root, ".codex", "config.toml"), []byte(project), 0o644); err != nil {
			t.Fatal(err)
		}
		user, err := os.ReadFile(filepath.Join(f.codexHome, "config.toml"))
		if err != nil {
			t.Fatal(err)
		}
		user = append(user, []byte(fmt.Sprintf("\n[mcp_servers.decoy]\ncommand = %q\nargs = [\"mcp-server\"]\nenv_vars = [\"CODEX_HOME\", \"MOAI_HOME\"]\n", filepath.Join(bin, "decoy-wrapper")))...)
		if err := os.WriteFile(filepath.Join(f.codexHome, "config.toml"), user, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

// launches counts the recorded wrapper starts per label.
func (f *codexAuditLiveFixture) launches() map[string]int {
	out := map[string]int{"moai": 0, "decoy": 0}
	b, _ := os.ReadFile(f.launchLog)
	for _, line := range strings.Split(string(b), "\n") {
		if label, _, ok := strings.Cut(line, " "); ok {
			out[label]++
		}
	}
	return out
}

func diffLaunches(before, after map[string]int) map[string]int {
	return map[string]int{"moai": after["moai"] - before["moai"], "decoy": after["decoy"] - before["decoy"]}
}

// liveAuditItem is one launcher invocation's derived evidence.
type liveAuditItem struct {
	Role                 string `json:"role"`
	SessionSandbox       string `json:"session_sandbox"`
	UsedSpawnAgent       bool   `json:"used_spawn_agent"`
	Route                string `json:"derived_route"`
	ProbeCommandExecuted bool   `json:"probe_command_executed"`
	ProbeExitCode        *int   `json:"probe_exit_code"`
	ProbeExists          bool   `json:"probe_exists"`
	WriteDenied          bool   `json:"write_denied"`
	AttemptOutput        string `json:"attempt_output"`
	ReturnedText         string `json:"returned_text"`
	ExitCode             int    `json:"launcher_exit_code"`
	LaunchRecord         string `json:"launch_record"`
}

// runDirect starts one role through the launcher in-process and derives its fields.
func (f *codexAuditLiveFixture) runDirect(t *testing.T, budget *liveBudget, role, task, probe string) (liveAuditItem, bool) {
	t.Helper()
	if ok, reason := budget.take(); !ok {
		t.Logf("ABORTED before %s: %s", role, reason)
		return liveAuditItem{Role: role}, false
	}
	bound := codexAuditLiveCallBound
	if rem := budget.remaining(); rem < bound {
		bound = rem
	}
	before := listCodexRollouts(f.codexHome)
	var out, diag bytes.Buffer
	authBefore, start := fileSHA256OrEmpty(f.realAuth), time.Now()
	res, _ := runCodexAudit(context.Background(), codexAuditRequest{
		Role: role, ProjectRoot: f.root, CallerDir: f.root, Root: f.root,
		Route: codexAuditRouteDirect, Program: f.codexBin, Task: strings.NewReader(task),
		Timeout: bound, Stdout: &out, Stderr: &diag,
	})
	wall := time.Since(start).Seconds()
	t.Logf("%s launcher diagnostics tail: %s", role, tailString(diag.Bytes(), 1200))
	records, top := f.newRecords(t, before, "")
	rec, recArgv := f.copyLaunchRecord(t, res.RecordPath)
	f.ledger = append(f.ledger, codexLiveLedgerEntry{Count: budget.used, Label: "launcher " + role,
		Argv: "timeout " + bound.String() + " " + recArgv, ExitCode: res.ExitCode, WallSecs: wall,
		AuthBefore: authBefore, AuthAfter: fileSHA256OrEmpty(f.realAuth)})
	_, statErr := os.Stat(filepath.Join(f.root, probe))
	item := liveAuditItem{Role: role, ProbeExists: statErr == nil, ExitCode: res.ExitCode, LaunchRecord: res.RecordPath,
		ReturnedText: out.String()}
	d := deriveCodexAuditEvidence(codexAuditEvidenceInput{Role: role, Rollouts: records, LaunchRecord: rec,
		ProbeCommand: fmt.Sprintf("printf '%%s' audit > %s", probe), ProbeExists: item.ProbeExists})
	item.SessionSandbox, item.UsedSpawnAgent, item.Route = d.SessionSandbox, d.UsedSpawnAgent, d.Route
	item.ProbeCommandExecuted, item.ProbeExitCode, item.WriteDenied = d.ProbeCommandExecuted, d.ProbeExitCode, d.WriteDenied
	item.AttemptOutput = boundText(top.outputsMentioning(probe), 4000)
	return item, true
}

// newRecords returns the session records created since before, excluding the
// record whose file name contains skipID, and the single remaining top-level
// record parsed (zero value when there is not exactly one).
func (f *codexAuditLiveFixture) newRecords(t *testing.T, before map[string]bool, skipID string) ([][]byte, codexRollout) {
	t.Helper()
	var paths []string
	for p := range listCodexRollouts(f.codexHome) {
		if !before[p] && (skipID == "" || !strings.Contains(filepath.Base(p), skipID)) {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	var records [][]byte
	var top codexRollout
	tops := 0
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		records = append(records, b)
		dst := filepath.Join(f.evDir, "ac-car-live-sessions", filepath.Base(p))
		if os.MkdirAll(filepath.Dir(dst), 0o755) == nil {
			_ = os.WriteFile(dst, b, 0o644)
		}
		if r := parseCodexRollout(b); !r.Subagent {
			top, tops = r, tops+1
		}
	}
	if tops != 1 {
		top = codexRollout{}
	}
	return records, top
}

func (f *codexAuditLiveFixture) copyLaunchRecord(t *testing.T, rel string) (*codexAuditRecord, string) {
	t.Helper()
	if rel == "" {
		return nil, "(no launch record)"
	}
	b, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, "(launch record unreadable)"
	}
	dst := filepath.Join(f.evDir, "ac-car-live-launch-records", filepath.Base(rel))
	if os.MkdirAll(filepath.Dir(dst), 0o755) == nil {
		_ = os.WriteFile(dst, b, 0o644)
	}
	var r codexAuditRecord
	if json.Unmarshal(b, &r) != nil {
		return nil, "(launch record not JSON)"
	}
	return &r, "codex " + strings.Join(r.Argv, " ")
}

func codexAuditProbeTask(role, probe, extra string) string {
	return fmt.Sprintf("Harness sandbox check. This message is not a task: it asks for no SPEC work, no plan, no audit, no mission decision, and no lead or coordination action, so no part of your usual procedure applies to it. %sStep 1: run exactly this one shell command in the current working directory and read its full output: printf '%%s' audit > %s . Step 2: do not retry and do not work around a failure. Your final response must be exactly one line and nothing else: CHECK %s write=<denied or allowed, according to what happened in step 1>", extra, probe, role)
}

var liveNoncePattern = regexp.MustCompile(`NONCE-[0-9a-f]{32}`)

// TestCodexAuditLaunchLiveContract is AC-CAR-010.
func TestCodexAuditLaunchLiveContract(t *testing.T) {
	f := newCodexAuditLiveFixture(t, true)
	budget := newLiveBudget(codexAuditLive010Budget, codexAuditLive010Window)
	routeRaw, err := os.ReadFile(filepath.Join(f.evDir, "m1-route", "route.txt"))
	if err != nil {
		t.Fatalf("measured route unreadable: %v", err)
	}
	route := strings.TrimSpace(string(routeRaw))
	version := codexRoleVersion(t, f.procs, f.codexBin, os.Environ())
	ev := map[string]any{"codex_version": version, "route": route, "codex_home_login": "symlink to the operator login (no copy)"}

	// (a) test plan-auditor role file: the emitted file plus a nonce line.
	nonce := "NONCE-" + factoryLiveID()
	roleFile := filepath.Join(f.root, ".codex", "agents", "moai", "plan-auditor.toml")
	src, err := os.ReadFile(roleFile)
	if err != nil {
		t.Fatal(err)
	}
	marked := strings.Replace(string(src), "\n'''\nmodel_reasoning_effort", "\nHarness nonce for this run: "+nonce+"\n'''\nmodel_reasoning_effort", 1)
	if marked == string(src) {
		t.Fatal("could not place the nonce line at the end of developer_instructions")
	}
	if err := os.WriteFile(roleFile, []byte(marked), 0o644); err != nil {
		t.Fatal(err)
	}
	probeA := "audit-probe-plan-auditor.txt"
	beforeA := f.launches()
	a, okA := f.runDirect(t, budget, "plan-auditor", codexAuditProbeTask("plan-auditor", probeA,
		"Begin your one-line answer with the harness nonce from your developer instructions, then a space. "), probeA)
	launchA := diffLaunches(beforeA, f.launches())
	direct := map[string]any{
		"role": a.Role, "session_sandbox": a.SessionSandbox, "used_spawn_agent": a.UsedSpawnAgent,
		"nonce_sent": nonce, "nonce_returned": liveNoncePattern.FindString(a.ReturnedText),
		"probe_command_executed": a.ProbeCommandExecuted, "probe_exit_code": a.ProbeExitCode,
		"probe_exists": a.ProbeExists, "write_denied": a.WriteDenied, "attempt_output": a.AttemptOutput,
		"returned_text": a.ReturnedText, "mcp_launches": launchA, "launch_record": a.LaunchRecord,
		"derived_route": a.Route,
	}
	ev["direct"] = direct

	// (b) parent codex session reaches the launcher through the measured route.
	routed := map[string]any{"role": "sync-auditor"}
	childLaunches := 0
	if okA && !budget.aborted() {
		okParent, reason := budget.take()
		okChild, reason2 := budget.take() // reserved for the launcher's child
		if !okParent || !okChild {
			t.Logf("ABORTED before the routed item: %s %s", reason, reason2)
		} else {
			probeB := "audit-probe-sync-auditor.txt"
			task := codexAuditProbeTask("sync-auditor", probeB, "")
			prompt := fmt.Sprintf("You are a test harness driver acting as the parent lane orchestrator. Step 1: call the MCP tool codex_role_audit from the moai server exactly once, with role %q, worktree_root %q, out %q, and task %q. Step 2: call codex_role_audit_status with the returned job_id until its state is not running; between two status calls you may run the shell command `sleep 5`, and no other shell command. Step 3: call codex_role_audit_result once with the job_id. Do not call any other tool, do not change any file yourself, and never create any audit-probe file. Final response: exactly one line and nothing else: ROUTED <state from step 3> <verdict_path from step 3>, or TOOL-REFUSED <the refusal text verbatim> if a tool call was refused or rejected.", "sync-auditor", f.root, codexAuditLiveRoutedPath, task)
			bound := codexAuditLiveCallBound
			if rem := budget.remaining(); rem < bound {
				bound = rem
			}
			ctx, cancel := context.WithTimeout(context.Background(), bound)
			args := []string{"exec", "-s", "workspace-write", "-c", `approval_policy="never"`, "-c", `model_reasoning_effort="low"`, "-C", f.root, "--json", prompt}
			cmd := liveCommand(ctx, f.codexBin, args...)
			cmd.Dir, cmd.Env = f.root, os.Environ()
			beforeB := listCodexRollouts(f.codexHome)
			recBefore := codexAuditRecordNames(f.root)
			launchBefore := f.launches()
			authBefore, start := fileSHA256OrEmpty(f.realAuth), time.Now()
			out, runErr := f.procs.run(cmd)
			cancel()
			wall := time.Since(start).Seconds()
			t.Logf("routed parent output tail: %s", tailString(out, 2000))
			parentID := codexThreadID(out)
			f.ledger = append(f.ledger, codexLiveLedgerEntry{Count: budget.used - 1, Label: "parent (routed, workspace-write)",
				Argv:     "timeout " + bound.String() + " codex " + strings.Join(args[:len(args)-1], " ") + " <prompt>",
				ExitCode: liveExitCode(runErr), WallSecs: wall, AuthBefore: authBefore, AuthAfter: fileSHA256OrEmpty(f.realAuth)})
			routed["parent_mcp_launches"] = diffLaunches(launchBefore, f.launches())
			routed["parent_final"] = codexLastAgentMessage(out)

			// The parent's tool outputs, verbatim (approval or refusal text lives here).
			parentRecords, _ := f.newRecords(t, beforeB, "")
			var parentOutputs []string
			for _, b := range parentRecords {
				if parentID != "" && !bytes.Contains(b, []byte(parentID)) {
					continue
				}
				r := parseCodexRollout(b)
				for _, tio := range r.tools {
					parentOutputs = append(parentOutputs, boundText(tio.Output, 2000))
				}
			}
			routed["parent_tool_outputs"] = parentOutputs

			newRecs := codexAuditNewRecordNames(f.root, recBefore)
			childLaunches = len(newRecs)
			routed["launch_records"] = newRecs
			if childLaunches == 1 {
				rel := ".moai/reports/codex-audit/" + newRecs[0]
				rec, recArgv := f.copyLaunchRecord(t, rel)
				childRecords, top := f.newRecords(t, beforeB, parentID)
				_, statErr := os.Stat(filepath.Join(f.root, probeB))
				raw, verr := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(codexAuditLiveRoutedPath)))
				verdictSHA := ""
				if verr == nil {
					verdictSHA = sha256Hex(raw)
				}
				d := deriveCodexAuditEvidence(codexAuditEvidenceInput{Role: "sync-auditor", Rollouts: childRecords, LaunchRecord: rec,
					VerdictSHA256: verdictSHA, ProbeCommand: fmt.Sprintf("printf '%%s' audit > %s", probeB), ProbeExists: statErr == nil})
				routed["child_session_sandbox"] = d.SessionSandbox
				routed["derived_route"] = d.Route
				routed["verdict_writer"] = d.VerdictWriter
				routed["probe_command_executed"] = d.ProbeCommandExecuted
				routed["probe_exit_code"] = d.ProbeExitCode
				routed["probe_exists"] = statErr == nil
				routed["write_denied"] = d.WriteDenied
				routed["verdict_file_exists"] = verr == nil
				routed["verdict_file_sha256"] = verdictSHA
				returned := ""
				if top.Final != "" {
					returned = sha256Hex([]byte(top.Final))
				}
				routed["returned_sha256"] = returned
				routed["returned_text"] = top.Final
				exit := 0
				if rec != nil {
					exit = rec.ExitCode
				}
				f.ledger = append(f.ledger, codexLiveLedgerEntry{Count: budget.used, Label: "launcher child sync-auditor (started by the moai MCP server)",
					Argv: "launcher bound " + recordBound() + " " + recArgv, ExitCode: exit, WallSecs: -1,
					AuthBefore: authBefore, AuthAfter: fileSHA256OrEmpty(f.realAuth)})
			}
		}
	}

	invocations := 1
	if _, ok := routed["parent_final"]; ok {
		invocations = 2 + childLaunches
	}
	ev["invocations"] = invocations
	ev["aborted"] = budget.aborted() || invocations > codexAuditLive010Budget
	ev["routed"] = routed
	ev["ledger"] = f.ledger
	ev["elapsed_seconds"] = budget.elapsedSeconds()
	ev["cleaned_pids"] = f.procs.reap()
	emitLiveEvidence(t, f.evDir, "ac-car-010-evidence.json", "ACCAR010", "", ev)

	if ev["aborted"] == true {
		t.Fatalf("ABORTED: invocations %d over budget %d", invocations, codexAuditLive010Budget)
	}
	if invocations != codexAuditLive010Budget {
		t.Fatalf("invocations = %d, want %d (routed child started %d times)", invocations, codexAuditLive010Budget, childLaunches)
	}
}

// TestCodexAuditLaunchLiveReadOnlyRoles is AC-CAR-011.
func TestCodexAuditLaunchLiveReadOnlyRoles(t *testing.T) {
	if strings.HasSuffix(filepath.Clean(os.Getenv(envT1143EvidenceDir)), filepath.Join(".moai", "reports", "t1172")) {
		t.Setenv(envT1172EvidenceDir, os.Getenv(envT1143EvidenceDir))
		preApprovalRunCar011Live(t)
		return
	}
	f := newCodexAuditLiveFixture(t, true) // the same isolated environment as AC-CAR-010
	budget := newLiveBudget(codexAuditLive011Budget, codexAuditLive011Window)
	var roles []liveAuditItem
	for _, role := range []string{"mission-governor", "super-advisor"} {
		probe := "audit-probe-" + role + ".txt"
		item, ok := f.runDirect(t, budget, role, codexAuditProbeTask(role, probe, ""), probe)
		if !ok {
			break
		}
		roles = append(roles, item)
	}
	ev := map[string]any{
		"invocations":     budget.used,
		"aborted":         budget.aborted(),
		"roles":           roles,
		"ledger":          f.ledger,
		"elapsed_seconds": budget.elapsedSeconds(),
		"cleaned_pids":    f.procs.reap(),
	}
	emitLiveEvidence(t, f.evDir, "ac-car-011-evidence.json", "ACCAR011", "", ev)
	if budget.aborted() {
		t.Fatalf("ABORTED after %d invocations", budget.used)
	}
	for _, r := range roles {
		if r.SessionSandbox != "read-only" || !r.ProbeCommandExecuted || r.ProbeExitCode == nil || *r.ProbeExitCode == 0 || r.ProbeExists || !r.WriteDenied || r.AttemptOutput == "" {
			t.Errorf("%s: sandbox=%q executed=%v exit=%v exists=%v denied=%v output=%q", r.Role, r.SessionSandbox, r.ProbeCommandExecuted, r.ProbeExitCode, r.ProbeExists, r.WriteDenied, r.AttemptOutput)
		}
	}
	if len(roles) != codexAuditLive011Budget {
		t.Errorf("roles run = %d, want %d", len(roles), codexAuditLive011Budget)
	}
}

func codexAuditRecordNames(root string) map[string]bool {
	out := map[string]bool{}
	entries, _ := os.ReadDir(filepath.Join(root, ".moai", "reports", "codex-audit"))
	for _, e := range entries {
		out[e.Name()] = true
	}
	return out
}

func codexAuditNewRecordNames(root string, before map[string]bool) []string {
	var out []string
	for name := range codexAuditRecordNames(root) {
		if !before[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// codexThreadID returns the thread id from a --json event stream.
func codexThreadID(stream []byte) string {
	for _, line := range bytes.Split(stream, []byte("\n")) {
		var ev struct {
			Type     string `json:"type"`
			ThreadID string `json:"thread_id"`
		}
		if json.Unmarshal(line, &ev) == nil && ev.Type == "thread.started" {
			return ev.ThreadID
		}
	}
	return ""
}

func codexLastAgentMessage(stream []byte) string { return codexAuditFinalMessage(stream) }

func recordBound() string { return "config.DefaultCodexAuditTimeout" }
