package cli

import (
	"context"
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

// LIVE gate and budget for AC-DHR-012 and AC-DHR-023 (lead decision 4):
// twelve role loads plus two audit-role write attempts, one codex exec
// process each, fourteen in total.
const (
	envCodexRoleLive        = "MOAI_CODEX_ROLE_LIVE"
	codexRoleLiveBudget     = 14
	codexRoleLiveWindow     = 1650 * time.Second // stays under the 1800s go test timeout
	codexRoleLiveCallBound  = 330 * time.Second
	codexRolePositiveRole   = "manager-docs"
	codexRolePositiveFile   = "pc-probe.txt"
	codexRoleMalformedToken = "Ignoring malformed agent role definition"
)

var codexRoleNoncePattern = regexp.MustCompile(`NONCE ([0-9a-f]{32})`)

type codexRoleLoad struct {
	Name                 string `json:"name"`
	NonceSent            string `json:"nonce_sent"`
	NonceReturned        string `json:"nonce_returned"`
	SubagentRoleObserved string `json:"subagent_role_observed"`
	SubagentSandbox      string `json:"subagent_sandbox"`
	ParentExitError      string `json:"parent_exit_error,omitempty"`
}

type codexWriteAttempt struct {
	Role            string `json:"role"`
	AttemptOutput   string `json:"attempt_output"`
	Denied          bool   `json:"denied"`
	ProbeExists     bool   `json:"probe_exists"`
	SubagentSandbox string `json:"subagent_sandbox"`
	ParentExitError string `json:"parent_exit_error,omitempty"`
}

type codexAuditVerdict struct {
	Role                  string `json:"role"`
	Nonce                 string `json:"nonce"`
	WriteDenied           bool   `json:"write_denied"`
	VerdictFileExists     bool   `json:"verdict_file_exists"`
	ReturnedSHA256        string `json:"returned_sha256"`
	VerdictFileSHA256     string `json:"verdict_file_sha256"`
	ReturnedContainsNonce bool   `json:"returned_contains_nonce"`
	ReturnedText          string `json:"returned_text"`
}

// TestCodexRoleLiveLoadAndReadOnly is AC-DHR-012 (REQ-DHR-014) and the
// evidence source of AC-DHR-023 (REQ-DHR-015). It runs a real codex binary in
// an isolated repository, CODEX_HOME and MOAI_HOME: every emitted role is
// spawned once by name and must echo its own nonce from its own session
// record; a workspace-write role writes a positive-control file; then the two
// audit roles are asked to write a probe file under their read-only sandbox,
// and the parent writes each returned verdict to a file.
func TestCodexRoleLiveLoadAndReadOnly(t *testing.T) {
	if os.Getenv(envCodexRoleLive) != "1" {
		t.Skip("NOT_RUN " + envCodexRoleLive + " is not 1: the live Codex role run is gated off")
	}
	evDir := liveEvidenceDir(t)
	codexBin, err := exec.LookPath("codex")
	if err != nil {
		t.Skip("NOT_RUN codex binary not on PATH")
	}
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	roleDir := filepath.Join(repoRoot, "internal", "template", "templates", ".codex", "agents", "moai")
	roles := codexRoleNames(t, roleDir)

	// Cleanup is registered here, before the first spawned process.
	procs := newLiveProcs(t)
	budget := newLiveBudget(codexRoleLiveBudget, codexRoleLiveWindow)
	t.Setenv(config.EnvHome, t.TempDir())

	root := canonicalDir(t, t.TempDir())
	codexHome, authHash := isolatedCodexHome(t, root)
	if err := os.MkdirAll(filepath.Join(root, ".codex", "agents", "moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, r := range roles {
		b, err := os.ReadFile(filepath.Join(roleDir, r+".toml"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".codex", "agents", "moai", r+".toml"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitCtx, gitCancel := context.WithTimeout(context.Background(), time.Minute)
	defer gitCancel()
	gitInit := liveCommand(gitCtx, "git", "init", "-q")
	gitInit.Dir = root
	if out, err := procs.run(gitInit); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	moaiBin := buildLiveMoai(t, procs)
	env := codexRoleEnv(codexHome, filepath.Dir(moaiBin))
	version := codexRoleVersion(t, procs, codexBin, env)

	call := func(label, prompt string) (rollouts []codexRollout, exitErr string, ok bool) {
		if okTake, reason := budget.take(); !okTake {
			t.Logf("ABORTED before %s: %s", label, reason)
			return nil, "", false
		}
		bound := codexRoleLiveCallBound
		if rem := budget.remaining(); rem < bound {
			bound = rem
		}
		ctx, cancel := context.WithTimeout(context.Background(), bound)
		defer cancel()
		before := listCodexRollouts(codexHome)
		cmd := liveCommand(ctx, codexBin, "exec", "--skip-git-repo-check",
			"-s", "workspace-write",
			"-c", `approval_policy="never"`,
			"-c", `model_reasoning_effort="low"`,
			"-C", root, "--json", prompt)
		cmd.Dir, cmd.Env = root, env
		out, runErr := procs.run(cmd)
		if runErr != nil {
			exitErr = runErr.Error()
		}
		codexRoleScanMalformed(t, label, out)
		t.Logf("%s codex output tail: %s", label, tailString(out, 1200))
		return newCodexRollouts(before, listCodexRollouts(codexHome)), exitErr, true
	}

	ev := map[string]any{
		"codex_version":       version,
		"budget":              codexRoleLiveBudget,
		"codex_home_isolated": true,
		"moai_home_isolated":  true,
		"sandbox_flags":       "codex exec -s workspace-write -c approval_policy=\"never\" (parent); role sandbox from the role file",
	}
	var loads []codexRoleLoad
	positive := map[string]any{"role": codexRolePositiveRole, "path": codexRolePositiveFile, "exists": false, "sha256": ""}
	var attempts []codexWriteAttempt
	var audits []codexAuditVerdict
	systemic := ""

	for i, role := range roles {
		nonce := factoryLiveID()
		msg := fmt.Sprintf("This is a load probe. Do not read or change any file and do not run any command. Your final response must be exactly one line and nothing else: NONCE %s", nonce)
		if role == codexRolePositiveRole {
			msg = fmt.Sprintf("This is a write probe. Run exactly this one shell command in the current working directory: printf '%%s' '%s' > %s . Do not change any other file. Your final response must be exactly one line and nothing else: NONCE %s", nonce, codexRolePositiveFile, nonce)
		}
		prompt := fmt.Sprintf("You are a test harness driver. Use the spawn_agent tool exactly once with agent_type set to %q and this exact message for the subagent: %q. Wait for that subagent to finish (use wait_agent). Then reply with the subagent's final message verbatim and nothing else. Do not run any command yourself and do not change any file yourself.", role, msg)
		rollouts, exitErr, ok := call("role "+role, prompt)
		if !ok {
			break
		}
		load := codexRoleLoad{Name: role, NonceSent: nonce, ParentExitError: exitErr}
		if sub, found := codexSubagentFor(rollouts, role); found {
			load.SubagentRoleObserved, load.SubagentSandbox = sub.Role, sub.Sandbox
			if m := codexRoleNoncePattern.FindStringSubmatch(sub.Final); m != nil {
				load.NonceReturned = m[1]
			}
		}
		loads = append(loads, load)
		if role == codexRolePositiveRole {
			if b, err := os.ReadFile(filepath.Join(root, codexRolePositiveFile)); err == nil {
				positive["exists"], positive["sha256"] = true, sha256Hex(b)
			}
		}
		if i == 0 && (exitErr != "" || load.SubagentRoleObserved == "") {
			systemic = fmt.Sprintf("first role call produced no named subagent (exit=%q); stopping to spend no further calls", exitErr)
			break
		}
	}

	if systemic == "" && !budget.aborted() {
		for _, role := range []string{"plan-auditor", "sync-auditor"} {
			nonce := factoryLiveID()
			probe := "audit-probe-" + role + ".txt"
			msg := fmt.Sprintf("Audit write probe. Step 1: run exactly this one shell command in the current working directory and read its full output: printf '%%s' audit > %s . Step 2: do not retry and do not work around a failure. Your final response must be exactly one line and nothing else: VERDICT %s nonce=%s write=<denied or allowed, according to what happened in step 1>", probe, role, nonce)
			prompt := fmt.Sprintf("You are a test harness driver acting as the parent lane orchestrator. Use the spawn_agent tool exactly once with agent_type set to %q and this exact message for the subagent: %q. Wait for it to finish (use wait_agent). Then write the subagent's final message text, exactly as returned and with nothing added, to the file verdicts/%s.txt (create the verdicts directory). You must never create any audit-probe file yourself. Finally reply DONE.", role, msg, role)
			rollouts, exitErr, ok := call("audit "+role, prompt)
			if !ok {
				break
			}
			a := codexWriteAttempt{Role: role, ParentExitError: exitErr}
			_, statErr := os.Stat(filepath.Join(root, probe))
			a.ProbeExists = statErr == nil
			v := codexAuditVerdict{Role: role, Nonce: nonce}
			if sub, found := codexSubagentFor(rollouts, role); found {
				a.SubagentSandbox = sub.Sandbox
				a.AttemptOutput = boundText(sub.outputsMentioning(probe), 4000)
				v.ReturnedText = sub.Final
			}
			a.Denied = !a.ProbeExists && a.AttemptOutput != "" && liveDenialPattern.MatchString(a.AttemptOutput)
			attempts = append(attempts, a)
			v.WriteDenied = a.Denied
			if b, err := os.ReadFile(filepath.Join(root, "verdicts", role+".txt")); err == nil {
				v.VerdictFileExists = true
				v.VerdictFileSHA256 = sha256Hex([]byte(strings.TrimRight(string(b), " \t\r\n")))
			}
			if v.ReturnedText != "" {
				v.ReturnedSHA256 = sha256Hex([]byte(strings.TrimRight(v.ReturnedText, " \t\r\n")))
				v.ReturnedContainsNonce = strings.Contains(v.ReturnedText, nonce)
			}
			audits = append(audits, v)
		}
	}

	cleaned := procs.reap()
	sort.Slice(loads, func(i, j int) bool { return loads[i].Name < loads[j].Name })
	ev["invocations"] = budget.used
	ev["elapsed_seconds"] = budget.elapsedSeconds()
	ev["roles"] = loads
	ev["positive_control"] = positive
	ev["write_attempts"] = attempts
	ev["aborted"] = budget.aborted()
	ev["systemic_stop"] = systemic
	ev["cleaned_pids"] = cleaned
	ev["codex_auth_rewritten"] = codexAuthChanged(codexHome, authHash)
	codexRoleExportSessions(t, codexHome, filepath.Join(evDir, "ac012-sessions"))
	emitLiveEvidence(t, evDir, "ac012-evidence.json", "AC012", "", ev)

	// AC-DHR-023: a missing returned text is recorded as not_run in the
	// evidence file only; nothing is printed that the AC-DHR-012 judge reads.
	ev023 := map[string]any{"audits": audits, "normalization": "trailing whitespace trimmed from both texts before hashing"}
	notRun := len(audits) != 2
	for _, a := range audits {
		if a.ReturnedText == "" {
			notRun = true
		}
	}
	if notRun {
		ev023["not_run"] = true
	}
	emitLiveEvidence(t, evDir, "ac023-evidence.json", "AC023", "", ev023)

	if ev["codex_auth_rewritten"] == true {
		t.Log("warning: codex rewrote the copied login during the run; the operator login may need a fresh codex login")
	}
	if budget.aborted() {
		t.Fatalf("ABORTED after %d invocations", budget.used)
	}
	if systemic != "" {
		t.Fatal(systemic)
	}
	codexRoleAssert(t, roles, loads, positive, attempts, budget.used)
}

// codexRoleAssert enforces the AC-DHR-012 conditions the evidence records.
func codexRoleAssert(t *testing.T, roles []string, loads []codexRoleLoad, positive map[string]any, attempts []codexWriteAttempt, used int) {
	t.Helper()
	if used != codexRoleLiveBudget {
		t.Errorf("invocations = %d, want %d", used, codexRoleLiveBudget)
	}
	if len(loads) != len(roles) {
		t.Errorf("role loads = %d, want %d", len(loads), len(roles))
	}
	seen := map[string]bool{}
	for _, l := range loads {
		if l.NonceReturned == "" || l.NonceReturned != l.NonceSent || seen[l.NonceSent] {
			t.Errorf("role %s: sent %q returned %q (observed role %q)", l.Name, l.NonceSent, l.NonceReturned, l.SubagentRoleObserved)
		}
		seen[l.NonceSent] = true
	}
	if positive["exists"] != true {
		t.Errorf("positive control %s was not written by the workspace-write role", codexRolePositiveFile)
	}
	if len(attempts) != 2 {
		t.Errorf("write attempts = %d, want 2", len(attempts))
	}
	for _, a := range attempts {
		if !a.Denied || a.ProbeExists || a.AttemptOutput == "" {
			t.Errorf("audit role %s: denied=%v probe_exists=%v output=%q sandbox=%q", a.Role, a.Denied, a.ProbeExists, a.AttemptOutput, a.SubagentSandbox)
		}
	}
}

func codexRoleNames(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.toml"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("no emitted role files under %s: %v", dir, err)
	}
	var names []string
	for _, m := range matches {
		names = append(names, strings.TrimSuffix(filepath.Base(m), ".toml"))
	}
	sort.Strings(names)
	return names
}

func codexRoleEnv(codexHome, moaiDir string) []string {
	env := os.Environ()
	env = replaceEnvValue(env, "CODEX_HOME", codexHome)
	env = replaceEnvValue(env, "PATH", moaiDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return factoryLiveWithoutAttribution(env)
}

func codexRoleVersion(t *testing.T, procs *liveProcs, bin string, env []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := liveCommand(ctx, bin, "--version")
	cmd.Env = env
	out, err := procs.run(cmd)
	if err != nil {
		t.Fatalf("codex --version: %v: %s", err, out)
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "codex-cli ")
}

// codexSubagentFor picks the subagent session that ran under the named role.
func codexSubagentFor(rollouts []codexRollout, role string) (codexRollout, bool) {
	var hit codexRollout
	found := false
	for _, r := range rollouts {
		if r.Subagent && r.Role == role {
			if !found || (hit.Final == "" && r.Final != "") {
				hit, found = r, true
			}
		}
	}
	return hit, found
}

func codexRoleScanMalformed(t *testing.T, label string, out []byte) {
	t.Helper()
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, codexRoleMalformedToken) {
			t.Logf("%s: %s", label, strings.TrimSpace(line))
		}
	}
}

// codexRoleExportSessions copies the isolated session records next to the
// evidence so the extraction can be re-read after the temp home is removed.
// Only sessions/ is copied, never auth.json.
func codexRoleExportSessions(t *testing.T, codexHome, dst string) {
	t.Helper()
	_ = os.RemoveAll(dst)
	src := filepath.Join(codexHome, "sessions")
	_ = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		target := filepath.Join(dst, rel)
		if os.MkdirAll(filepath.Dir(target), 0o755) == nil {
			_ = os.WriteFile(target, b, 0o644)
		}
		return nil
	})
}

func boundText(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
