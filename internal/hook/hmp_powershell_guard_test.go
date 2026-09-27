package hook

// SPEC-HOOK-MATCHER-POWERSHELL-001 (card t1224) — guard and evidence behavior
// for the PowerShell tool.
//
// Every test pairs a PowerShell payload with the Bash payload carrying the same
// command text, and asserts the PowerShell outcome equals the Bash outcome. A
// PowerShell-only assertion would pass if both paths were broken the same way.
//
// Isolation (REQ-HMP-014): each TestHMP* function calls hmpIsolateHome, which
// points MOAI_HOME and CLAUDE_PROJECT_DIR at per-test directories, and none
// calls t.Parallel (t.Setenv forbids it).

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// hmpInput builds a PreToolUse payload for tool with the given command.
func hmpInput(t *testing.T, tool, sessionID, cwd, command string) *HookInput {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"command": command, "description": "probe"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return &HookInput{
		SessionID:     sessionID,
		HookEventName: "PreToolUse",
		ToolName:      tool,
		AgentType:     "manager-develop",
		CWD:           cwd,
		ToolInput:     raw,
	}
}

// hmpCfg returns the default config with the named guards enabled.
func hmpCfg(branchGuard, integrationLock bool, slot config.SlotLeaseConfig) *config.Config {
	cfg := config.NewDefaultConfig()
	cfg.Workflow.BranchGuard.Enabled = branchGuard
	cfg.Workflow.IntegrationLock.Enabled = integrationLock
	cfg.Workflow.SlotLease = slot
	return cfg
}

func hmpHandler(cfg *config.Config, projectDir string) *preToolHandler {
	return &preToolHandler{
		cfg:        &mockConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	}
}

func hmpHandle(t *testing.T, h *preToolHandler, in *HookInput) (decision, reason string) {
	t.Helper()
	out, err := h.Handle(context.Background(), in)
	if err != nil {
		t.Fatalf("Handle(%s, %q): %v", in.ToolName, string(in.ToolInput), err)
	}
	return decisionOf(out), reasonOf(out)
}

// TestHMPShellPredicate pins REQ-HMP-004's closed set.
func TestHMPShellPredicate(t *testing.T) {
	hmpIsolateHome(t)
	for name, want := range map[string]bool{
		"Bash": true, "PowerShell": true,
		"powershell": false, "bash": false, "Write": false, "Edit": false, "": false,
	} {
		if got := IsShellTool(name); got != want {
			t.Errorf("IsShellTool(%q) = %v, want %v", name, got, want)
		}
	}
}

// TestHMPDenyAskNoVerifyParity pins AC-HMP-005 (REQ-HMP-006).
func TestHMPDenyAskNoVerifyParity(t *testing.T) {
	hmpIsolateHome(t)
	dir := t.TempDir()
	cases := []struct {
		command string
		want    string
	}{
		{"terraform destroy", DecisionDeny},
		{`Remove-Item -Recurse -Force C:\`, DecisionDeny},
		{"git commit --no-verify -m x", DecisionDeny},
		{"git push --force origin feature/x", DecisionAsk},
		{"git reset --hard HEAD~1", DecisionAsk},
	}
	h := hmpHandler(hmpCfg(false, false, config.SlotLeaseConfig{}), dir)
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			bashDecision, bashReason := hmpHandle(t, h, hmpInput(t, "Bash", "s-1", dir, tc.command))
			if bashDecision != tc.want {
				t.Fatalf("premise: Bash %q decided %q, want %q", tc.command, bashDecision, tc.want)
			}
			psDecision, psReason := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", dir, tc.command))
			if psDecision != bashDecision || psReason != bashReason {
				t.Errorf("PowerShell %q = (%q, %q); Bash = (%q, %q)", tc.command, psDecision, psReason, bashDecision, bashReason)
			}
		})
	}
}

// TestHMPBranchGuardParity pins AC-HMP-006 (REQ-HMP-007).
func TestHMPBranchGuardParity(t *testing.T) {
	hmpIsolateHome(t)
	t.Setenv(branchGuardExemptEnv, "")
	repo := newBranchGuardRepoFixture(t)
	h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)

	bashDecision, bashReason := hmpHandle(t, h, hmpInput(t, "Bash", "s-1", repo, "git switch -c probe"))
	if bashDecision != DecisionDeny {
		t.Fatalf("premise: Bash branch switch in the primary checkout decided %q", bashDecision)
	}
	psDecision, psReason := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, "git switch -c probe"))
	if psDecision != DecisionDeny || !strings.HasPrefix(psReason, branchGuardViolationPrefix+":") {
		t.Fatalf("PowerShell branch switch = (%q, %q); want deny with %s:", psDecision, psReason, branchGuardViolationPrefix)
	}
	if psReason != bashReason {
		t.Errorf("reason differs: PowerShell %q, Bash %q", psReason, bashReason)
	}

	primary, wt := worktreeFixture(t)
	hw := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), primary)
	if d, r := hmpHandle(t, hw, hmpInput(t, "PowerShell", "s-1", wt, "git switch -c probe")); d == DecisionDeny {
		t.Errorf("PowerShell branch switch inside a linked worktree was denied: %q", r)
	}
}

// hmpSeedForeignLock records a live lock held by another session.
func hmpSeedForeignLock(t *testing.T, root string) {
	t.Helper()
	seedLock(t, root, kanban.IntegrationLock{
		SessionID:   "sess-holder",
		SessionName: "lane-9",
		PID:         os.Getpid(),
		Branch:      "release/v9.9.9",
		AcquiredAt:  "2026-01-01T00:00:00Z",
	})
}

// TestHMPIntegrationLockParity pins the first half of AC-HMP-007.
func TestHMPIntegrationLockParity(t *testing.T) {
	hmpIsolateHome(t)
	root := t.TempDir()
	hmpSeedForeignLock(t, root)
	h := hmpHandler(hmpCfg(false, true, config.SlotLeaseConfig{}), root)
	cmd := "git merge --no-ff WT-x"
	bashDecision, bashReason := hmpHandle(t, h, hmpInput(t, "Bash", "sess-other", root, cmd))
	if bashDecision != DecisionDeny {
		t.Fatalf("premise: Bash merge under a foreign live hold decided %q", bashDecision)
	}
	psDecision, psReason := hmpHandle(t, h, hmpInput(t, "PowerShell", "sess-other", root, cmd))
	if psDecision != bashDecision || psReason != bashReason {
		t.Errorf("PowerShell = (%q, %q); Bash = (%q, %q)", psDecision, psReason, bashDecision, bashReason)
	}
}

// TestHMPSlotLeaseParity pins the second half of AC-HMP-007.
func TestHMPSlotLeaseParity(t *testing.T) {
	hmpIsolateHome(t)
	root := slotGuardRepo(t)
	seedSlotGuardLease(t, root, liveForeignLease())
	slot := config.SlotLeaseConfig{
		Enabled:            true,
		DefaultMaxDuration: "30m",
		Resources: map[string]config.SlotLeaseResourceConfig{
			slotGuardResource: {Commands: []string{`\bgo\s+test\b`}},
		},
	}
	h := hmpHandler(hmpCfg(false, false, slot), root)
	cmd := "go test ./internal/cli/..."
	var bashDecision, bashReason, psDecision, psReason string
	captureStderr(t, func() {
		bashDecision, bashReason = hmpHandle(t, h, hmpInput(t, "Bash", "s-2", root, cmd))
		psDecision, psReason = hmpHandle(t, h, hmpInput(t, "PowerShell", "s-2", root, cmd))
	})
	if bashDecision != DecisionDeny {
		t.Fatalf("premise: Bash heavy command under a foreign live lease decided %q", bashDecision)
	}
	if psDecision != bashDecision || psReason != bashReason {
		t.Errorf("PowerShell = (%q, %q); Bash = (%q, %q)", psDecision, psReason, bashDecision, bashReason)
	}
}

// hmpDefectiveInputs are the wrongly-typed tool_input values of AC-HMP-008.
var hmpDefectiveInputs = []string{`"x"`, `[1]`, `null`, `{}`, `{"command": 42}`}

// hmpComparable renders an output with the echoed tool name masked, so two
// outputs that differ only in which tool they describe compare equal.
func hmpComparable(t *testing.T, out *HookOutput, tool string) string {
	t.Helper()
	if out == nil {
		return "<nil>"
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal output: %v", err)
	}
	return strings.ReplaceAll(string(data), `"tool_name":"`+tool+`"`, `"tool_name":"<tool>"`)
}

// TestHMPDefectiveToolInput pins the hook half of AC-HMP-008 (REQ-HMP-009).
func TestHMPDefectiveToolInput(t *testing.T) {
	hmpIsolateHome(t)
	root := t.TempDir()
	slot := config.SlotLeaseConfig{Enabled: true, DefaultMaxDuration: "30m",
		Resources: map[string]config.SlotLeaseResourceConfig{"demo": {Commands: []string{`.`}}}}
	pre := hmpHandler(hmpCfg(true, true, slot), root)
	post := NewPostToolHandler()
	for _, raw := range hmpDefectiveInputs {
		t.Run(raw, func(t *testing.T) {
			outputs := map[string][2]string{}
			for _, tool := range []string{"Bash", "PowerShell"} {
				in := &HookInput{SessionID: "s-1", HookEventName: "PreToolUse", ToolName: tool, CWD: root, ToolInput: json.RawMessage(raw)}
				preOut, err := pre.Handle(context.Background(), in)
				if err != nil {
					t.Fatalf("pre-tool %s: %v", tool, err)
				}
				pin := *in
				pin.HookEventName = "PostToolUse"
				postOut, err := post.Handle(context.Background(), &pin)
				if err != nil {
					t.Fatalf("post-tool %s: %v", tool, err)
				}
				outputs[tool] = [2]string{hmpComparable(t, preOut, tool), hmpComparable(t, postOut, tool)}
			}
			t.Logf("Bash control: pre=%s post=%s", outputs["Bash"][0], outputs["Bash"][1])
			if outputs["PowerShell"] != outputs["Bash"] {
				t.Errorf("PowerShell outputs %q differ from Bash %q", outputs["PowerShell"], outputs["Bash"])
			}
		})
	}
}

// hmpGoTestResponse is a Bash-shaped passing go test response (stdout only,
// no exit_code — the shape Claude Code delivers).
const hmpGoTestResponse = `{"stdout":"ok  \tgithub.com/x/y\t0.42s\n","stderr":"","interrupted":false,"isImage":false}`

func hmpPostInput(tool, command, response string) *HookInput {
	raw, _ := json.Marshal(map[string]any{"command": command})
	return &HookInput{
		SessionID:     "sess-ev",
		HookEventName: "PostToolUse",
		ToolName:      tool,
		ToolInput:     raw,
		ToolResponse:  json.RawMessage(response),
	}
}

// TestHMPEvidenceParity pins AC-HMP-010 (REQ-HMP-011) on the pure record path.
func TestHMPEvidenceParity(t *testing.T) {
	hmpIsolateHome(t)
	bashRec, bashOK := buildEvidenceRecord(hmpPostInput("Bash", "go test ./x/...", hmpGoTestResponse))
	if !bashOK || !bashRec.IsTestPass {
		t.Fatalf("premise: Bash go test record ok=%v pass=%v", bashOK, bashRec.IsTestPass)
	}
	psRec, psOK := buildEvidenceRecord(hmpPostInput("PowerShell", "go test ./x/...", hmpGoTestResponse))
	if psOK != bashOK || psRec.IsTestPass != bashRec.IsTestPass || psRec.IsTestFail != bashRec.IsTestFail ||
		psRec.Outcome != bashRec.Outcome || psRec.IsZeroExecution != bashRec.IsZeroExecution {
		t.Errorf("PowerShell record (ok=%v %+v) differs from Bash (ok=%v %+v)", psOK, psRec, bashOK, bashRec)
	}

	for _, shape := range []string{`{"result": 7}`, `[1, 2]`, `42`} {
		rec, ok := buildEvidenceRecord(hmpPostInput("PowerShell", "go test ./x/...", shape))
		if ok && rec.IsTestPass {
			t.Errorf("unrecognized response %s recorded a pass: %+v", shape, rec)
		}
	}

	zero := `{"stdout":"?   \tgithub.com/x/y\t[no test files]\n","stderr":""}`
	bashAdv := maybeZeroExecutionAdvisory(hmpPostInput("Bash", "go test ./x/...", zero), "")
	if !strings.Contains(bashAdv, zeroExecutionSentinel) {
		t.Fatalf("premise: Bash zero-execution advisory absent: %q", bashAdv)
	}
	if psAdv := maybeZeroExecutionAdvisory(hmpPostInput("PowerShell", "go test ./x/...", zero), ""); psAdv != bashAdv {
		t.Errorf("PowerShell zero-execution advisory %q differs from Bash %q", psAdv, bashAdv)
	}
}

// hmpTelemetryLines counts today's telemetry records under root.
func hmpTelemetryLines(t *testing.T, root string) int {
	t.Helper()
	day := time.Now().UTC().Format("2006-01-02")
	data, err := os.ReadFile(filepath.Join(root, ".moai", "evolution", "telemetry", "usage-"+day+".jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return 0
	}
	if err != nil {
		t.Fatalf("read telemetry: %v", err)
	}
	return strings.Count(strings.TrimSpace(string(data)), "\n") + 1
}

// TestHMPEvidenceWritePaths pins sites 5 and 9: the post-tool handler and
// LogBashEvidence write one record for a PowerShell test run, as for Bash.
func TestHMPEvidenceWritePaths(t *testing.T) {
	hmpIsolateHome(t)
	for _, tool := range []string{"Bash", "PowerShell"} {
		t.Run(tool, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			t.Setenv(config.EnvClaudeProjectDir, root)
			LogBashEvidence(hmpPostInput(tool, "go test ./x/...", hmpGoTestResponse))
			if got := hmpTelemetryLines(t, root); got != 1 {
				t.Errorf("LogBashEvidence(%s) wrote %d records, want 1", tool, got)
			}
			if _, err := NewPostToolHandler().Handle(context.Background(), hmpPostInput(tool, "go test ./x/...", hmpGoTestResponse)); err != nil {
				t.Fatalf("post-tool: %v", err)
			}
			if got := hmpTelemetryLines(t, root); got != 2 {
				t.Errorf("post-tool handler (%s) left %d records, want 2", tool, got)
			}
		})
	}
}

// hmpB64 encodes a PowerShell -EncodedCommand payload (UTF-16LE, base64).
func hmpB64(s string) string {
	b := make([]byte, 0, len(s)*2)
	for _, r := range s {
		b = append(b, byte(r), byte(r>>8))
	}
	return base64.StdEncoding.EncodeToString(b)
}

// hmpAuditLines returns the lines of a guard audit log containing needle.
func hmpAuditLines(t *testing.T, path, needle string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, needle) {
			out = append(out, line)
		}
	}
	return out
}

// TestHMPIndirectionClassification is the M0 measurement pinned as a test:
// which PowerShell indirection forms the existing parsers already classify.
func TestHMPIndirectionClassification(t *testing.T) {
	hmpIsolateHome(t)
	enc := hmpB64("git switch x")
	cases := []struct {
		command          string
		branchClassified bool
		mergeClassified  bool
	}{
		{"& git switch x", true, false},
		{"& git merge --no-ff WT-x", true, true},
		{`iex "git switch x"`, false, false},
		{`Invoke-Expression 'git merge --no-ff WT-x'`, false, false},
		{"Start-Process git -ArgumentList 'switch','x'", false, false},
		{"Start-Process git -ArgumentList 'merge','WT-x'", false, false},
		{"pwsh -EncodedCommand " + enc, false, false},
	}
	for _, tc := range cases {
		_, branch := matchBranchStateCommand(tc.command)
		merge := integrationMergePattern.MatchString(substituteQuotedArguments(tc.command))
		t.Logf("%-50q branch-guard=%v integration-lock=%v", tc.command, branch, merge)
		if branch != tc.branchClassified || merge != tc.mergeClassified {
			t.Errorf("%q: branch=%v merge=%v, want %v %v", tc.command, branch, merge, tc.branchClassified, tc.mergeClassified)
		}
	}
}

// TestHMPPowerShellIndirectionDetector pins the REQ-HMP-010 construct set.
func TestHMPPowerShellIndirectionDetector(t *testing.T) {
	hmpIsolateHome(t)
	enc := hmpB64("Write-Output ok")
	unclassified := []string{
		`iex "git switch probe"`,
		`IEX 'git merge --no-ff WT-x'`,
		`Invoke-Expression "git switch probe"`,
		"Start-Process git -ArgumentList 'switch','probe'",
		"start-process -FilePath git -ArgumentList merge",
		"pwsh -EncodedCommand " + enc,
		"pwsh -encodedcommand " + enc,
		"pwsh -EnCoDeDcOmMaNd " + enc,
		"pwsh -e " + enc,
		"pwsh -E " + enc,
		"pwsh -ec " + enc,
		"pwsh -en " + enc,
		"pwsh -enc " + enc,
		"pwsh -enco " + enc,
		"pwsh -encodedc " + enc,
		"pwsh --EncodedCommand " + enc,
		"pwsh /EncodedCommand " + enc,
		"pwsh.exe -NoProfile -enc " + enc,
		"powershell -enc " + enc,
		"powershell.exe -NonInteractive -e " + enc,
		"Get-Date; pwsh -enc " + enc,
	}
	for _, c := range unclassified {
		if got := powerShellIndirection(c); got == "" {
			t.Errorf("powerShellIndirection(%q) = \"\", want a construct", c)
		}
	}
	classifiedOrPlain := []string{
		"git switch probe",
		"& git switch probe",
		`iex "Get-Date"`,
		"Start-Process notepad",
		"pwsh -Command Get-Date",
		"pwsh -ExecutionPolicy Bypass -File x.ps1",
		"pwsh -NoProfile -File run.ps1",
		"git log -e",
		`Write-Output "pwsh -enc abc"`,
	}
	for _, c := range classifiedOrPlain {
		if got := powerShellIndirection(c); got != "" {
			t.Errorf("powerShellIndirection(%q) = %q, want \"\"", c, got)
		}
	}
}

// TestHMPUnclassifiedBranchGuard pins the branch-guard half of AC-HMP-009.
func TestHMPUnclassifiedBranchGuard(t *testing.T) {
	hmpIsolateHome(t)
	t.Setenv(branchGuardExemptEnv, "")
	enc := hmpB64("git status")
	cases := []string{
		`iex "git switch probe"`,
		"Start-Process git -ArgumentList 'switch','probe'",
		"pwsh -EncodedCommand " + enc,
		"powershell -enc " + enc,
	}
	for _, cmd := range cases {
		t.Run(cmd, func(t *testing.T) {
			repo := newBranchGuardRepoFixture(t)
			h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
			logPath := filepath.Join(repo, branchGuardAuditRelPath)
			if d, r := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, cmd)); d == DecisionDeny {
				t.Fatalf("unclassifiable construct denied: %q", r)
			}
			lines := hmpAuditLines(t, logPath, "powershell-unclassified")
			if len(lines) != 1 {
				t.Fatalf("want exactly 1 powershell-unclassified line in %s, got %d: %q", branchGuardAuditRelPath, len(lines), lines)
			}
			if !strings.Contains(hmpReasonField(lines[0]), "unclassifiable") {
				t.Errorf("audit line reason lacks the word unclassifiable: %q", lines[0])
			}
			if strings.Contains(lines[0], "Write-Output") || strings.Contains(lines[0], "git status\"") {
				t.Errorf("audit line suggests the payload was decoded: %q", lines[0])
			}
		})
	}

	t.Run("Bash eval control keeps its outcome", func(t *testing.T) {
		repo := newBranchGuardRepoFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
		if d, r := hmpHandle(t, h, hmpInput(t, "Bash", "s-1", repo, `eval "git switch probe"`)); d == DecisionDeny {
			t.Fatalf("Bash eval control denied: %q", r)
		}
		if lines := hmpAuditLines(t, filepath.Join(repo, branchGuardAuditRelPath), ""); len(strings.Join(lines, "")) != 0 {
			t.Errorf("Bash eval control wrote an audit line: %q", lines)
		}
	})

	t.Run("call operator stays classified", func(t *testing.T) {
		repo := newBranchGuardRepoFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
		if d, _ := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, "& git switch probe")); d != DecisionDeny {
			t.Errorf("`& git switch probe` = %q, want deny", d)
		}
	})

	t.Run("linked worktree writes no line", func(t *testing.T) {
		primary, wt := worktreeFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), primary)
		if d, _ := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", wt, `iex "git switch probe"`)); d == DecisionDeny {
			t.Errorf("worktree call denied")
		}
		if lines := hmpAuditLines(t, filepath.Join(primary, branchGuardAuditRelPath), "powershell-unclassified"); len(lines) != 0 {
			t.Errorf("worktree call wrote %d unclassified lines", len(lines))
		}
	})

	t.Run("deny list still applies behind the construct", func(t *testing.T) {
		repo := newBranchGuardRepoFixture(t)
		h := hmpHandler(hmpCfg(true, false, config.SlotLeaseConfig{}), repo)
		if d, _ := hmpHandle(t, h, hmpInput(t, "PowerShell", "s-1", repo, "iex terraform destroy")); d != DecisionDeny {
			t.Errorf("`iex terraform destroy` = %q, want deny from the destructive list", d)
		}
	})
}

// hmpReasonField extracts the reason="…" field of an audit line.
func hmpReasonField(line string) string {
	_, after, ok := strings.Cut(line, "reason=")
	if !ok {
		return ""
	}
	return after
}

// TestHMPUnclassifiedIntegrationLock pins the integration-lock half of
// AC-HMP-009.
func TestHMPUnclassifiedIntegrationLock(t *testing.T) {
	hmpIsolateHome(t)
	enc := hmpB64("git status")
	for _, cmd := range []string{`iex "git merge --no-ff WT-x"`, "pwsh -enc " + enc} {
		t.Run(cmd, func(t *testing.T) {
			root := t.TempDir()
			hmpSeedForeignLock(t, root)
			h := hmpHandler(hmpCfg(false, true, config.SlotLeaseConfig{}), root)
			if d, r := hmpHandle(t, h, hmpInput(t, "PowerShell", "sess-other", root, cmd)); d == DecisionDeny {
				t.Fatalf("unclassifiable construct denied: %q", r)
			}
			lines := hmpAuditLines(t, filepath.Join(root, integrationLockAuditRelPath), "powershell-unclassified")
			if len(lines) != 1 {
				t.Fatalf("want exactly 1 powershell-unclassified line, got %d: %q", len(lines), lines)
			}
			if !strings.Contains(hmpReasonField(lines[0]), "unclassifiable") {
				t.Errorf("audit line reason lacks the word unclassifiable: %q", lines[0])
			}
		})
	}

	t.Run("no hold writes no line", func(t *testing.T) {
		root := t.TempDir()
		h := hmpHandler(hmpCfg(false, true, config.SlotLeaseConfig{}), root)
		hmpHandle(t, h, hmpInput(t, "PowerShell", "sess-other", root, `iex "git merge --no-ff WT-x"`))
		if lines := hmpAuditLines(t, filepath.Join(root, integrationLockAuditRelPath), "powershell-unclassified"); len(lines) != 0 {
			t.Errorf("an unheld window wrote %d lines", len(lines))
		}
	})

	t.Run("Bash eval control keeps its outcome", func(t *testing.T) {
		root := t.TempDir()
		hmpSeedForeignLock(t, root)
		h := hmpHandler(hmpCfg(false, true, config.SlotLeaseConfig{}), root)
		if d, _ := hmpHandle(t, h, hmpInput(t, "Bash", "sess-other", root, `eval "git merge --no-ff WT-x"`)); d == DecisionDeny {
			t.Errorf("Bash eval control denied")
		}
		if _, err := os.Stat(filepath.Join(root, integrationLockAuditRelPath)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Bash eval control created the integration-lock audit log (err=%v)", err)
		}
	})
}
