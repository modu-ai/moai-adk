// Package cli — `moai verify audit-plan` (SPEC-AUDIT-MODEL-CONVERGE-001 M4:
// AC-ACV-003 verb part, AC-ACV-005, AC-ACV-011, AC-ACV-012 verb part,
// AC-ACV-015).
//
// Every test drives the verb through the `verify` group's cobra tree over temp
// trees, with the backendCall seam stubbed to count calls: what is asserted is
// the JSON the verb prints, the exit code, and what it did not do.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// runAuditPlanCmd runs `moai verify audit-plan <args>` through a fresh verify
// group and returns stdout, stderr and the process exit code the error chain
// maps to (0 on success, -1 for an error that carries no exit code).
func runAuditPlanCmd(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := newVerifyCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"audit-plan"}, args...))
	err := cmd.Execute()
	if err != nil {
		code = -1
		if c, ok := ResolveExitCode(err); ok {
			code = c
		}
	}
	return out.String(), errBuf.String(), code
}

// auditPlanTree makes a temp project root whose workflow.yaml holds yaml; an
// empty yaml leaves the tree with no workflow.yaml at all.
func auditPlanTree(t *testing.T, yaml string) string {
	t.Helper()
	dir := t.TempDir()
	if yaml != "" {
		path := filepath.Join(dir, ".moai", "config", "sections", "workflow.yaml")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// auditPlanPinsOnlyYAML is the shape the distributed template ships: an audit
// block that carries only the three backend pins.
const auditPlanPinsOnlyYAML = "workflow:\n  audit:\n    claude:\n      model: claude-opus-5-5\n      effort: high\n" +
	"    codex:\n      model: gpt-6.1-sol\n      effort: high\n    glm:\n      model: glm-5.3\n      effort: high\n"

// planEntry mirrors one backends[] member of the verb's output.
type planEntry struct {
	Backend  string `json:"backend"`
	Gate     string `json:"gate"`
	Source   string `json:"source"`
	Explicit bool   `json:"explicit"`
}

// planOut mirrors the verb's JSON object. The two tri-state members are typed
// any because the unreadable state reports the string "unknown" there.
type planOut struct {
	ProjectRoot        string      `json:"project_root"`
	ConfigRoot         string      `json:"config_root"`
	ConfigStatus       string      `json:"config_status"`
	Model              string      `json:"model"`
	ModelSource        string      `json:"model_source"`
	Backends           []planEntry `json:"backends"`
	CrossModelActive   any         `json:"cross_model_active"`
	CrossModelRequired any         `json:"cross_model_required"`
	EnforcedRequired   []string    `json:"enforced_required"`
	Note               string      `json:"note"`
	ConvergenceCheck   *struct {
		OK     bool     `json:"ok"`
		Unmet  []string `json:"unmet"`
		Reason string   `json:"reason"`
	} `json:"convergence_check"`
}

// decodePlanOut decodes stdout as exactly one JSON object.
func decodePlanOut(t *testing.T, stdout string) planOut {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	var p planOut
	if err := dec.Decode(&p); err != nil {
		t.Fatalf("stdout is not a JSON object: %v\n%s", err, stdout)
	}
	if dec.More() {
		t.Fatalf("stdout carries more than one JSON value:\n%s", stdout)
	}
	return p
}

func TestAuditPlanCmd_RegisteredUnderVerify(t *testing.T) {
	var found bool
	for _, c := range newVerifyCmd().Commands() {
		if c.Name() == "audit-plan" {
			found = true
		}
	}
	if !found {
		t.Fatal("audit-plan must be registered under the verify group (verifyExtraCommands)")
	}
}

// TestAuditPlanCmd_HelpNamesTheVerbAndTheRootObligation: the verb's own usage
// names `moai verify audit-plan` (compare ledger E3, where the group's help is
// printed instead) and keeps the worktree-root obligation in the help text.
func TestAuditPlanCmd_HelpNamesTheVerbAndTheRootObligation(t *testing.T) {
	stdout, _, code := runAuditPlanCmd(t, "--help")
	if code != 0 {
		t.Fatalf("--help exit code = %d", code)
	}
	for _, want := range []string{"audit-plan", "--result", "--project-root", "toplevel"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help must mention %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "Shared diagnostic snapshot contract") {
		t.Errorf("help is the verify GROUP help, not the verb's usage:\n%s", stdout)
	}
}

// AC-ACV-011 / AC-ACV-003 (verb part): the plan for every token and every
// configuration state, with the D7' row for an explicit `claude`.
func TestAuditPlanCmd_PrintsPlan(t *testing.T) {
	type want struct {
		status, model, modelSource string
		entries                    [3]planEntry
		active, required           bool
		enforced                   []string
	}
	cases := []struct {
		name string
		yaml string
		want want
	}{
		{"no workflow.yaml", "", want{
			"absent", "", "default",
			[3]planEntry{{"claude", "required", "default", false}, {"codex", "required", "default", false}, {"glm", "advisory", "default", false}},
			false, false, []string{},
		}},
		{"model multi", planWorkflowYAML("multi", nil), want{
			"ok", "multi", "config",
			[3]planEntry{{"claude", "required", "config.model", true}, {"codex", "required", "config.model", true}, {"glm", "advisory", "config.model", true}},
			true, true, []string{"claude", "codex"},
		}},
		{"model claude is the D7' row", planWorkflowYAML("claude", nil), want{
			"ok", "claude", "config",
			[3]planEntry{{"claude", "required", "config.model", true}, {"codex", "required", "default", false}, {"glm", "advisory", "default", false}},
			false, false, []string{"claude"},
		}},
		{"model codex", planWorkflowYAML("codex", nil), want{
			"ok", "codex", "config",
			[3]planEntry{{"claude", "off", "config.model", true}, {"codex", "required", "config.model", true}, {"glm", "off", "config.model", true}},
			true, true, []string{"codex"},
		}},
		{"model glm", planWorkflowYAML("glm", nil), want{
			"ok", "glm", "config",
			[3]planEntry{{"claude", "off", "config.model", true}, {"codex", "off", "config.model", true}, {"glm", "required", "config.model", true}},
			true, true, []string{"glm"},
		}},
		{"model codex with gates.glm advisory", planWorkflowYAML("codex", map[string]string{"glm": "advisory"}), want{
			"ok", "codex", "config",
			[3]planEntry{{"claude", "off", "config.model", true}, {"codex", "required", "config.model", true}, {"glm", "advisory", "config.gates", true}},
			true, true, []string{"codex"},
		}},
		{"explicit gates without a model", planWorkflowYAML("", map[string]string{"codex": "required"}), want{
			"ok", "", "default",
			[3]planEntry{{"claude", "required", "default", false}, {"codex", "required", "config.gates", true}, {"glm", "advisory", "default", false}},
			true, true, []string{"codex"},
		}},
		{"explicit claude-only spelling", planWorkflowYAML("claude", map[string]string{"codex": "off", "glm": "off"}), want{
			"ok", "claude", "config",
			[3]planEntry{{"claude", "required", "config.model", true}, {"codex", "off", "config.gates", true}, {"glm", "off", "config.gates", true}},
			false, false, []string{"claude"},
		}},
		{"token is trimmed", "workflow:\n  audit:\n    model: \" multi \"\n", want{
			"ok", "multi", "config",
			[3]planEntry{{"claude", "required", "config.model", true}, {"codex", "required", "config.model", true}, {"glm", "advisory", "config.model", true}},
			true, true, []string{"claude", "codex"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := auditPlanTree(t, tc.yaml)
			stdout, stderr, code := runAuditPlanCmd(t, "--project-root", root)
			if code != 0 || stderr != "" {
				t.Fatalf("exit=%d stderr=%q", code, stderr)
			}
			got := decodePlanOut(t, stdout)
			if got.ProjectRoot != root {
				t.Errorf("project_root = %q, want %q", got.ProjectRoot, root)
			}
			if got.ConfigStatus != tc.want.status || got.Model != tc.want.model || got.ModelSource != tc.want.modelSource {
				t.Errorf("status/model/source = %q/%q/%q, want %q/%q/%q", got.ConfigStatus, got.Model, got.ModelSource,
					tc.want.status, tc.want.model, tc.want.modelSource)
			}
			if !reflect.DeepEqual(got.Backends, tc.want.entries[:]) {
				t.Errorf("backends = %+v\nwant      %+v", got.Backends, tc.want.entries)
			}
			if got.CrossModelActive != tc.want.active || got.CrossModelRequired != tc.want.required {
				t.Errorf("cross_model_active/required = %v/%v, want %v/%v", got.CrossModelActive, got.CrossModelRequired,
					tc.want.active, tc.want.required)
			}
			if !reflect.DeepEqual(got.EnforcedRequired, tc.want.enforced) {
				t.Errorf("enforced_required = %#v, want %#v", got.EnforcedRequired, tc.want.enforced)
			}
			if got.ConvergenceCheck != nil {
				t.Errorf("convergence_check must be absent without --result, got %+v", got.ConvergenceCheck)
			}
		})
	}

	t.Run("config-orphaned worktree reads its primary checkout", func(t *testing.T) {
		fx := newUntrackedFixture(t, planWorkflowYAML("multi", nil))
		if _, err := os.Stat(filepath.Join(fx.W, ".moai")); !os.IsNotExist(err) {
			t.Fatalf("premise: the worktree has no .moai (stat err=%v)", err)
		}
		stdout, stderr, code := runAuditPlanCmd(t, "--project-root", fx.W)
		if code != 0 || stderr != "" {
			t.Fatalf("exit=%d stderr=%q", code, stderr)
		}
		got := decodePlanOut(t, stdout)
		if got.ConfigStatus != "ok" || got.Model != "multi" || got.CrossModelActive != true {
			t.Errorf("the primary's model: multi must be the plan, got status=%q model=%q active=%v", got.ConfigStatus, got.Model, got.CrossModelActive)
		}
		if got.ProjectRoot != fx.W || got.ConfigRoot != fx.P {
			t.Errorf("project_root/config_root = %q/%q, want %q/%q", got.ProjectRoot, got.ConfigRoot, fx.W, fx.P)
		}
	})
}

// AC-ACV-008 (verb side) / AC-ACV-011: an audit block that carries only the
// backend pins is the distributed default, never a `claude` token.
func TestAuditPlanCmd_PinsOnlyIsDefault(t *testing.T) {
	root := auditPlanTree(t, auditPlanPinsOnlyYAML)
	stdout, stderr, code := runAuditPlanCmd(t, "--project-root", root)
	if code != 0 || stderr != "" {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	got := decodePlanOut(t, stdout)
	if got.ConfigStatus != "absent" || got.Model != "" || got.ModelSource != "default" {
		t.Errorf("pins-only status/model/source = %q/%q/%q, want absent//default", got.ConfigStatus, got.Model, got.ModelSource)
	}
	for _, e := range got.Backends {
		if e.Explicit || e.Source != "default" {
			t.Errorf("pins-only entry %+v must be a non-explicit default", e)
		}
	}
	if got.CrossModelActive != false || got.CrossModelRequired != false || len(got.EnforcedRequired) != 0 {
		t.Errorf("pins-only active/required/enforced = %v/%v/%v, want false/false/[]", got.CrossModelActive, got.CrossModelRequired, got.EnforcedRequired)
	}
}

// AC-ACV-003 / AC-ACV-011: an invalid token or gate exits 1 with the
// REQ-ACV-003 text on stderr and nothing on stdout; it is never defaulted.
func TestAuditPlanCmd_InvalidConfigExitsOne(t *testing.T) {
	cases := []struct {
		name, yaml, want string
	}{
		{"unknown token", planWorkflowYAML("grok", nil), `audit-plan: audit_model "grok" unknown (want one of claude|codex|glm|multi)`},
		{"token matching is case-sensitive", planWorkflowYAML("Multi", nil), `audit-plan: audit_model "Multi" unknown (want one of claude|codex|glm|multi)`},
		{"unknown gate", planWorkflowYAML("", map[string]string{"codex": "requird"}), `audit-plan: audit.gates.codex "requird" unknown (want one of off|advisory|required)`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := auditPlanTree(t, tc.yaml)
			stdout, stderr, code := runAuditPlanCmd(t, "--project-root", root)
			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if stdout != "" {
				t.Errorf("stdout must be empty on invalid configuration, got %q", stdout)
			}
			if !strings.Contains(stderr, tc.want) {
				t.Errorf("stderr = %q\nwant it to contain %q", stderr, tc.want)
			}
		})
	}
}

// AC-ACV-012 (verb part): a workflow.yaml that exists but cannot be read is a
// distinct, non-passing state — not the default plan and not an error exit.
func TestAuditPlanCmd_UnreadableConfig(t *testing.T) {
	assertUnreadable := func(t *testing.T, root, wantNote string) {
		t.Helper()
		stdout, stderr, code := runAuditPlanCmd(t, "--project-root", root)
		if code != 0 || stderr != "" {
			t.Fatalf("exit=%d stderr=%q (the state is reported on stdout, exit 0)", code, stderr)
		}
		got := decodePlanOut(t, stdout)
		if got.ConfigStatus != "unreadable" {
			t.Fatalf("config_status = %q, want unreadable\n%s", got.ConfigStatus, stdout)
		}
		if got.CrossModelActive != "unknown" || got.CrossModelRequired != "unknown" {
			t.Errorf("cross_model_active/required = %v/%v, want the string unknown", got.CrossModelActive, got.CrossModelRequired)
		}
		if got.Backends == nil || len(got.Backends) != 0 {
			t.Errorf("backends = %#v, want an empty array (no default plan is printed)", got.Backends)
		}
		if !strings.Contains(got.Note, wantNote) {
			t.Errorf("note = %q, want it to name %q", got.Note, wantNote)
		}
	}

	t.Run("corrupt yaml", func(t *testing.T) {
		root := auditPlanTree(t, "workflow: [unclosed\n")
		assertUnreadable(t, root, "workflow.yaml")
		stdout, _, _ := runAuditPlanCmd(t, "--project-root", root)
		if !strings.Contains(decodePlanOut(t, stdout).Note, "parse") {
			t.Errorf("the note must name the parse cause: %s", stdout)
		}
	})
	t.Run("a directory where the file should be", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"), 0o755); err != nil {
			t.Fatal(err)
		}
		assertUnreadable(t, root, "workflow.yaml")
	})
	t.Run("orphaned worktree whose primary cannot be identified", func(t *testing.T) {
		fx := newUntrackedFixture(t, planWorkflowYAML("multi", nil))
		admin := filepath.Join(fx.P, ".git", "worktrees", "W")
		if err := os.Remove(filepath.Join(admin, "HEAD")); err != nil {
			t.Fatal(err)
		}
		assertUnreadable(t, fx.W, "primary")
	})
	t.Run("the unreadable plan is never a pass for --result", func(t *testing.T) {
		root := auditPlanTree(t, "workflow: [unclosed\n")
		stdout, _, code := runAuditPlanCmd(t, "--project-root", root, "--result", `{"overall_verdict":"pass"}`)
		if code != 0 {
			t.Fatalf("exit code = %d", code)
		}
		got := decodePlanOut(t, stdout)
		if got.ConvergenceCheck == nil || got.ConvergenceCheck.OK {
			t.Errorf("convergence_check must be present and not ok for an unreadable plan: %+v", got.ConvergenceCheck)
		}
	})
}

// AC-ACV-011: --project-root names the tree read; with no flag the group's
// default ($CLAUDE_PROJECT_DIR) names it — the trap of design.md §D.4.
func TestAuditPlanCmd_ProjectRoot(t *testing.T) {
	envTree := auditPlanTree(t, planWorkflowYAML("multi", nil))
	flagTree := auditPlanTree(t, planWorkflowYAML("codex", nil))

	t.Setenv(config.EnvClaudeProjectDir, envTree)

	stdout, _, code := runAuditPlanCmd(t, "--project-root", flagTree)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if got := decodePlanOut(t, stdout); got.Model != "codex" || got.ProjectRoot != flagTree {
		t.Errorf("--project-root must win over the environment: model=%q project_root=%q", got.Model, got.ProjectRoot)
	}

	stdout, _, code = runAuditPlanCmd(t)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if got := decodePlanOut(t, stdout); got.Model != "multi" || got.ProjectRoot != envTree {
		t.Errorf("with no flag the CLAUDE_PROJECT_DIR tree is read: model=%q project_root=%q", got.Model, got.ProjectRoot)
	}

	t.Run("a root that is not a directory is a system error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "no-such-tree")
		stdout, stderr, code := runAuditPlanCmd(t, "--project-root", missing)
		if code != 2 || stdout != "" || !strings.HasPrefix(stderr, "audit-plan:") {
			t.Errorf("exit=%d stdout=%q stderr=%q, want exit 2, empty stdout, an audit-plan: line", code, stdout, stderr)
		}
	})
}

// AC-ACV-011: the output contract — stdout is one JSON object carrying
// config_status, or stderr carries an audit-plan: line, and nothing else.
func TestAuditPlanCmd_OutputContract(t *testing.T) {
	good := auditPlanTree(t, planWorkflowYAML("multi", nil))
	runs := []struct {
		name string
		args []string
		root string
	}{
		{"plan", nil, good},
		{"absent", nil, auditPlanTree(t, "")},
		{"unreadable", nil, auditPlanTree(t, "workflow: [unclosed\n")},
		{"invalid configuration", nil, auditPlanTree(t, planWorkflowYAML("grok", nil))},
		{"checker ok", []string{"--result", `{"plan_source":"config","per_backend_verdicts":[]}`}, good},
		{"checker rejects a non-object", []string{"--result", `[1]`}, good},
		{"missing root", nil, filepath.Join(t.TempDir(), "gone")},
	}
	for _, r := range runs {
		t.Run(r.name, func(t *testing.T) {
			stdout, stderr, code := runAuditPlanCmd(t, append([]string{"--project-root", r.root}, r.args...)...)
			if code == 0 {
				if stderr != "" {
					t.Errorf("a successful run writes nothing to stderr, got %q", stderr)
				}
				var m map[string]any
				if err := json.Unmarshal([]byte(stdout), &m); err != nil {
					t.Fatalf("stdout is not one JSON object: %v\n%s", err, stdout)
				}
				if _, ok := m["config_status"]; !ok {
					t.Errorf("stdout object carries no config_status:\n%s", stdout)
				}
				return
			}
			if stdout != "" {
				t.Errorf("an error run writes nothing to stdout, got %q", stdout)
			}
			lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
			if len(lines) != 1 || !strings.HasPrefix(lines[0], "audit-plan:") {
				t.Errorf("stderr must be exactly one audit-plan: line, got %q", stderr)
			}
		})
	}
}

// moaiFiles returns every regular file under root/.moai keyed by relative
// path, with its content, so a before/after comparison catches a create and a
// change alike. Directories are deliberately not listed: the CLI's own start-up
// directory creation is a pre-existing property of every command (design §D.5).
func moaiFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	base := filepath.Join(root, ".moai")
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			b, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			rel, _ := filepath.Rel(base, path)
			out[rel] = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// AC-ACV-011: the verb runs no audit backend and writes no regular file under
// the audited tree's .moai — across every fixture shape, with and without
// --result.
func TestAuditPlanCmd_WritesNoFiles(t *testing.T) {
	var calls atomic.Int64
	withBackendCall(t, func(_ context.Context, _, _, _, _ string) ReviewOutput {
		calls.Add(1)
		return ReviewOutput{Verdict: "pass", Findings: []Finding{}, NextSteps: []string{}}
	})

	fx := newUntrackedFixture(t, planWorkflowYAML("multi", nil))
	trees := map[string][]string{
		"no yaml":     {auditPlanTree(t, "")},
		"pins only":   {auditPlanTree(t, auditPlanPinsOnlyYAML)},
		"multi":       {auditPlanTree(t, planWorkflowYAML("multi", nil))},
		"claude":      {auditPlanTree(t, planWorkflowYAML("claude", nil))},
		"codex":       {auditPlanTree(t, planWorkflowYAML("codex", map[string]string{"glm": "advisory"}))},
		"invalid":     {auditPlanTree(t, planWorkflowYAML("grok", nil))},
		"unreadable":  {auditPlanTree(t, "workflow: [unclosed\n")},
		"orphaned wt": {fx.W, fx.P},
	}
	for name, roots := range trees {
		t.Run(name, func(t *testing.T) {
			before := map[string]map[string]string{}
			for _, r := range roots {
				before[r] = moaiFiles(t, r)
			}
			runAuditPlanCmd(t, "--project-root", roots[0])
			runAuditPlanCmd(t, "--project-root", roots[0], "--result", `{"plan_source":"config","per_backend_verdicts":[]}`)
			for _, r := range roots {
				if after := moaiFiles(t, r); !reflect.DeepEqual(before[r], after) {
					t.Errorf("a regular file under %s/.moai was created or changed:\nbefore %v\nafter  %v", r, keys(before[r]), keys(after))
				}
			}
		})
	}
	if n := calls.Load(); n != 0 {
		t.Errorf("the backendCall seam recorded %d calls, want 0", n)
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// --- the checker (REQ-ACV-016) ----------------------------------------------

// auditEntry renders one per_backend_verdicts member; verdict and gate are raw
// JSON text, so a test can pass a wrong-typed or an absent value.
func auditEntry(backend, gate, verdict string) string {
	parts := []string{fmt.Sprintf(`"backend":%q`, backend)}
	if gate != "" {
		parts = append(parts, `"gate":`+gate)
	}
	if verdict != "" {
		parts = append(parts, `"verdict":`+verdict)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// auditResultJSON renders an audit_multi result digest.
func auditResultJSON(planSource string, entries ...string) string {
	ps := ""
	if planSource != "" {
		ps = fmt.Sprintf(`"plan_source":%q,`, planSource)
	}
	return `{"overall_verdict":"pass","gate_unmet":"",` + ps + `"per_backend_verdicts":[` + strings.Join(entries, ",") + `]}`
}

const (
	jReq   = `"required"`
	jPass  = `"pass"`
	jFail  = `"fail"`
	jIncon = `"inconclusive"`
)

// checkerCase is one --result fixture against a model: multi tree (claude and
// codex enforced-required).
type checkerCase struct {
	name   string
	result string
	wantOK bool
	unmet  []string
}

// multiChecker returns the nine acceptance fixtures (a)-(i) plus the D4
// fixtures (normalisation, wrong JSON types, the answered-`fail` control).
func multiChecker() []checkerCase {
	claudeOK := auditEntry("claude", jReq, jPass)
	cs := []checkerCase{
		{"a_old_server_shape", auditResultJSON("", claudeOK, auditEntry("codex", jReq, jIncon)), false, []string{"codex"}},
		{"b_pass_without_plan_source", auditResultJSON("", claudeOK, auditEntry("codex", jReq, jPass)), false, []string{}},
		{"c_no_codex_entry", auditResultJSON("config", claudeOK), false, []string{"codex"}},
		{"d_codex_gate_advisory", auditResultJSON("config", claudeOK, auditEntry("codex", `"advisory"`, jPass)), false, []string{"codex"}},
		{"e_all_required_pass_with_config_source", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, jPass)), true, []string{}},
		{"i_verdict_member_absent", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, "")), false, []string{"codex"}},
		{"i_verdict_empty", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `""`)), false, []string{"codex"}},
		{"i_verdict_pas", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `"pas"`)), false, []string{"codex"}},
		// D4: normalisation is not applied.
		{"n_verdict_upper", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `"PASS"`)), false, []string{"codex"}},
		{"n_verdict_padded", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `" pass "`)), false, []string{"codex"}},
		{"n_verdict_capitalised", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `"Fail"`)), false, []string{"codex"}},
		{"n_gate_upper", auditResultJSON("config", claudeOK, auditEntry("codex", `"REQUIRED"`, jPass)), false, []string{"codex"}},
		{"n_gate_padded", auditResultJSON("config", claudeOK, auditEntry("codex", `" required"`, jPass)), false, []string{"codex"}},
		// D4: a wrong JSON type is never a verdict.
		{"t_verdict_number", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `1`)), false, []string{"codex"}},
		{"t_verdict_bool", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `true`)), false, []string{"codex"}},
		{"t_verdict_object", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `{"v":"pass"}`)), false, []string{"codex"}},
		{"t_verdict_array", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `["pass"]`)), false, []string{"codex"}},
		{"t_verdict_null", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, `null`)), false, []string{"codex"}},
		{"t_gate_bool", auditResultJSON("config", claudeOK, auditEntry("codex", `true`, jPass)), false, []string{"codex"}},
		{"t_gate_null", auditResultJSON("config", claudeOK, auditEntry("codex", `null`, jPass)), false, []string{"codex"}},
		{"t_backend_not_a_string", auditResultJSON("config", claudeOK, `{"backend":7,"gate":"required","verdict":"pass"}`), false, []string{"codex"}},
		{"t_entries_not_an_array", `{"plan_source":"config","per_backend_verdicts":"pass"}`, false, []string{"claude", "codex"}},
		{"t_plan_source_not_a_string", `{"plan_source":1,"per_backend_verdicts":[` + claudeOK + `,` + auditEntry("codex", jReq, jPass) + `]}`, false, []string{}},
		{"t_plan_source_other_value", auditResultJSON("env", claudeOK, auditEntry("codex", jReq, jPass)), false, []string{}},
		// D4 positive control: an answered `fail` is an answered gate. REQ-ACV-008
		// makes a gate unmet only when the backend is left without a verdict; a
		// backend that answered `fail` fails the verdict through the convergence
		// result itself, not through this check.
		{"p_fail_is_answered", auditResultJSON("config", auditEntry("claude", jReq, jFail), auditEntry("codex", jReq, jFail)), true, []string{}},
		{"p_extra_members_ignored", `{"overall_verdict":"pass","extra":{"x":1},"plan_source":"config","per_backend_verdicts":[` +
			`{"backend":"claude","gate":"required","verdict":"pass","summary":"s","findings":[],"next_steps":[]},` +
			`{"backend":"codex","gate":"required","verdict":"pass","summary":"s","findings":[],"next_steps":[]},` +
			`{"backend":"glm","gate":"advisory","verdict":"inconclusive"}]}`, true, []string{}},
		{"p_duplicate_conflicting_entries", auditResultJSON("config", claudeOK, auditEntry("codex", jReq, jPass), auditEntry("codex", jReq, jIncon)), false, []string{"codex"}},
	}
	return cs
}

func multiPlan(t *testing.T) config.AuditPlan {
	t.Helper()
	plan, err := config.ResolveAuditPlan(config.AuditConfig{Model: "multi"}, config.AuditGates{})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

// AC-ACV-015: the nine acceptance fixtures (a)-(i), through the verb.
func TestAuditPlanCmd_ResultCheck(t *testing.T) {
	multi := auditPlanTree(t, planWorkflowYAML("multi", nil))
	none := auditPlanTree(t, "")

	runCheck := func(t *testing.T, root, result string) planOut {
		t.Helper()
		stdout, stderr, code := runAuditPlanCmd(t, "--project-root", root, "--result", result)
		if code != 0 || stderr != "" {
			t.Fatalf("exit=%d stderr=%q", code, stderr)
		}
		got := decodePlanOut(t, stdout)
		if got.ConvergenceCheck == nil {
			t.Fatalf("convergence_check missing:\n%s", stdout)
		}
		return got
	}
	claudeOK := auditEntry("claude", jReq, jPass)

	t.Run("a_old_server_shaped_result", func(t *testing.T) {
		got := runCheck(t, multi, auditResultJSON("", claudeOK, auditEntry("codex", jReq, jIncon)))
		if got.ConvergenceCheck.OK || !reflect.DeepEqual(got.ConvergenceCheck.Unmet, []string{"codex"}) {
			t.Errorf("want ok=false unmet=[codex], got %+v", got.ConvergenceCheck)
		}
	})
	t.Run("b_pass_but_no_plan_source", func(t *testing.T) {
		got := runCheck(t, multi, auditResultJSON("", claudeOK, auditEntry("codex", jReq, jPass)))
		if got.ConvergenceCheck.OK || !strings.Contains(got.ConvergenceCheck.Reason, "plan_source") {
			t.Errorf("want ok=false with a reason naming plan_source, got %+v", got.ConvergenceCheck)
		}
	})
	t.Run("c_no_codex_entry", func(t *testing.T) {
		got := runCheck(t, multi, auditResultJSON("config", claudeOK))
		if got.ConvergenceCheck.OK || !reflect.DeepEqual(got.ConvergenceCheck.Unmet, []string{"codex"}) {
			t.Errorf("want ok=false unmet=[codex], got %+v", got.ConvergenceCheck)
		}
	})
	t.Run("d_codex_pass_but_gate_advisory", func(t *testing.T) {
		got := runCheck(t, multi, auditResultJSON("config", claudeOK, auditEntry("codex", `"advisory"`, jPass)))
		if got.ConvergenceCheck.OK || !reflect.DeepEqual(got.ConvergenceCheck.Unmet, []string{"codex"}) {
			t.Errorf("want ok=false unmet=[codex], got %+v", got.ConvergenceCheck)
		}
	})
	t.Run("e_all_required_pass_with_config_source", func(t *testing.T) {
		got := runCheck(t, multi, auditResultJSON("config", claudeOK, auditEntry("codex", jReq, jPass)))
		if !got.ConvergenceCheck.OK || len(got.ConvergenceCheck.Unmet) != 0 {
			t.Errorf("want ok=true, got %+v", got.ConvergenceCheck)
		}
	})
	t.Run("f_not_a_json_object", func(t *testing.T) {
		for _, bad := range []string{`[1]`, `"x"`, `7`, `null`, `{`, `{"plan_source":"config"`, ``} {
			stdout, stderr, code := runAuditPlanCmd(t, "--project-root", multi, "--result", bad)
			if code != 1 || stdout != "" || !strings.HasPrefix(stderr, "audit-plan:") {
				t.Errorf("--result %q: exit=%d stdout=%q stderr=%q, want exit 1, empty stdout, an audit-plan: line", bad, code, stdout, stderr)
			}
		}
	})
	t.Run("g_no_audit_configuration", func(t *testing.T) {
		for _, result := range []string{`{}`, auditResultJSON("", claudeOK, auditEntry("codex", jReq, jIncon))} {
			got := runCheck(t, none, result)
			if !got.ConvergenceCheck.OK || len(got.ConvergenceCheck.Unmet) != 0 {
				t.Errorf("a default plan has nothing enforced; want ok=true for %s, got %+v", result, got.ConvergenceCheck)
			}
		}
	})
	t.Run("h_digest_form_equals_full_result", func(t *testing.T) {
		pairs := []struct{ digest, full string }{
			{
				auditResultJSON("", claudeOK, auditEntry("codex", jReq, jIncon)),
				`{"per_backend_verdicts":[{"backend":"claude","gate":"required","verdict":"pass","summary":"ok","findings":[],"next_steps":[],"source":"in-session"},` +
					`{"backend":"codex","gate":"required","verdict":"inconclusive","summary":"binary missing","findings":[],"next_steps":["x"]}],` +
					`"overall_verdict":"pass","disagreement_flag":false,"fail_open_backends":["codex"],"gate_unmet":"","residual_risk_note":"","audit_receipt":""}`,
			},
			{
				auditResultJSON("config", claudeOK, auditEntry("codex", jReq, jPass)),
				`{"per_backend_verdicts":[{"backend":"claude","gate":"required","verdict":"pass","summary":"ok","findings":[],"next_steps":[]},` +
					`{"backend":"codex","gate":"required","verdict":"pass","summary":"ok","findings":[],"next_steps":[]}],` +
					`"overall_verdict":"pass","disagreement_flag":false,"fail_open_backends":[],"gate_unmet":"","plan_source":"config"}`,
			},
		}
		for i, p := range pairs {
			d, f := runCheck(t, multi, p.digest), runCheck(t, multi, p.full)
			if !reflect.DeepEqual(d.ConvergenceCheck, f.ConvergenceCheck) {
				t.Errorf("pair %d: digest %+v != full %+v", i, d.ConvergenceCheck, f.ConvergenceCheck)
			}
		}
	})
	t.Run("i_required_entries_with_a_malformed_verdict", func(t *testing.T) {
		for _, entry := range []string{
			`{"backend":"codex","gate":"required"}`,
			`{"backend":"codex","gate":"required","verdict":""}`,
			`{"backend":"codex","gate":"required","verdict":"pas"}`,
		} {
			got := runCheck(t, multi, auditResultJSON("config", claudeOK, entry))
			if got.ConvergenceCheck.OK || !reflect.DeepEqual(got.ConvergenceCheck.Unmet, []string{"codex"}) {
				t.Errorf("entry %s: want ok=false unmet=[codex], got %+v", entry, got.ConvergenceCheck)
			}
		}
	})
}

// D4: the fixtures that kill the predicate mutants (normalisation, wrong JSON
// types, the answered-`fail` control), through the verb.
func TestAuditPlanCmd_ResultCheck_Hardening(t *testing.T) {
	multi := auditPlanTree(t, planWorkflowYAML("multi", nil))
	for _, tc := range multiChecker() {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runAuditPlanCmd(t, "--project-root", multi, "--result", tc.result)
			if code != 0 || stderr != "" {
				t.Fatalf("exit=%d stderr=%q", code, stderr)
			}
			got := decodePlanOut(t, stdout).ConvergenceCheck
			if got == nil || got.OK != tc.wantOK {
				t.Fatalf("ok = %+v, want %v", got, tc.wantOK)
			}
			if !tc.wantOK && !reflect.DeepEqual(got.Unmet, tc.unmet) {
				t.Errorf("unmet = %#v, want %#v", got.Unmet, tc.unmet)
			}
			if !tc.wantOK && got.Reason == "" {
				t.Errorf("a failed check must carry a named reason")
			}
		})
	}
}

// D4: the mutants are mechanical, not claimed. Each variant of the entry
// predicate runs over the same fixtures; the real predicate agrees with every
// expectation, and each mutant disagrees with at least one.
func TestAuditPlanCmd_ResultCheck_PredicateMutantsAreKilled(t *testing.T) {
	plan := multiPlan(t)
	fixtures := multiChecker()

	lower := func(v any) (string, bool) {
		s, ok := v.(string)
		return strings.ToLower(strings.TrimSpace(s)), ok
	}
	mutants := []struct {
		name      string
		pred      auditEntryPredicate
		mustDieOn func(fixture string) bool
	}{
		{"literal_verdict_not_inconclusive", func(v, g any) bool { return v != "inconclusive" && g == "required" },
			func(f string) bool { return strings.HasPrefix(f, "i_") }},
		{"pass_only", func(v, g any) bool { return v == "pass" && g == "required" },
			func(f string) bool { return f == "p_fail_is_answered" }},
		{"case_and_space_folded", func(v, g any) bool {
			s, ok := lower(v)
			gs, gok := lower(g)
			return ok && gok && (s == "pass" || s == "fail") && gs == "required"
		}, func(f string) bool { return strings.HasPrefix(f, "n_") }},
		{"non_string_accepted", func(v, g any) bool {
			if s, ok := v.(string); ok {
				return (s == "pass" || s == "fail") && g == "required"
			}
			return g == "required"
		}, func(f string) bool {
			return strings.HasPrefix(f, "t_verdict_") || strings.HasPrefix(f, "i_verdict_member")
		}},
		{"any_nonempty_verdict", func(v, g any) bool { s, ok := v.(string); return ok && s != "" && g == "required" },
			func(f string) bool {
				return f == "i_verdict_pas" || f == "n_verdict_upper" || f == "n_verdict_padded" || f == "n_verdict_capitalised"
			}},
		{"gate_ignored", func(v, g any) bool { return v == "pass" || v == "fail" },
			func(f string) bool {
				return f == "d_codex_gate_advisory" || strings.HasPrefix(f, "n_gate_") || strings.HasPrefix(f, "t_gate_")
			}},
	}

	run := func(pred auditEntryPredicate, fixture checkerCase) bool {
		got, err := checkAuditResult(plan, fixture.result, pred)
		if err != nil {
			t.Fatalf("fixture %s: %v", fixture.name, err)
		}
		return got.OK
	}

	for _, f := range fixtures {
		if got := run(auditEntryAnswered, f); got != f.wantOK {
			t.Errorf("real predicate on %s: ok=%v, want %v", f.name, got, f.wantOK)
		}
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			var killers []string
			for _, f := range fixtures {
				if run(m.pred, f) != f.wantOK {
					killers = append(killers, f.name)
				}
			}
			if len(killers) == 0 {
				t.Fatalf("mutant %s survives every fixture", m.name)
			}
			t.Logf("mutant %s is killed by: %s", m.name, strings.Join(killers, ", "))
			var expected bool
			for _, k := range killers {
				if m.mustDieOn(k) {
					expected = true
				}
			}
			if !expected {
				t.Errorf("mutant %s is killed only by %v, none of the fixtures written to kill it", m.name, killers)
			}
			if m.name == "literal_verdict_not_inconclusive" {
				// The acceptance fixtures (a)-(e) are the first five rows; the mutant
				// survives all of them, which is why fixture (i) exists.
				for _, f := range fixtures[:5] {
					if run(m.pred, f) != f.wantOK {
						t.Errorf("fixture %s (a)-(e) must not kill the literal-predicate mutant", f.name)
					}
				}
			}
		})
	}
}

// --- structure guards ------------------------------------------------------

func readAuditPlanSource(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("audit_plan_cmd.go")
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// TestAuditPlanCmd_NoAskUserQuestion is the subagent-boundary static guard: the
// verb is non-interactive (it returns exit codes and structured output only).
func TestAuditPlanCmd_NoAskUserQuestion(t *testing.T) {
	src := readAuditPlanSource(t)
	if strings.Contains(src, "AskUserQuestion") || strings.Contains(src, "mcp__askuser__") {
		t.Fatal("audit-plan must remain non-interactive")
	}
}

// AC-ACV-015's grep as a test: the checker is pure — no store read, no session
// id, no reference to the persisted convergence file.
func TestAuditPlanCmd_ChecksNoStore(t *testing.T) {
	src := readAuditPlanSource(t)
	for _, banned := range []string{"audit-multi", "loadConvergenceResult", "check-session"} {
		if strings.Contains(src, banned) {
			t.Errorf("audit_plan_cmd.go must not mention %q", banned)
		}
	}
}

// AC-ACV-004 / AC-ACV-012 greps as a test: the verb reads raw, through the
// loader that reports a failure — never through the merged defaults or the
// reader that turns an error into an absent configuration.
func TestAuditPlanCmd_RawReadThroughTheReportingLoader(t *testing.T) {
	src := readAuditPlanSource(t)
	if !strings.Contains(src, "loadWorkflowAuditSection") {
		t.Error("the verb must read through loadWorkflowAuditSection")
	}
	for _, banned := range []string{"workflowAuditPins", "NewDefaultWorkflowConfig", "NewDefaultConfig"} {
		if strings.Contains(src, banned) {
			t.Errorf("audit_plan_cmd.go must not reference %q", banned)
		}
	}
}

// AC-ACV-005: the verb is one of the non-test consumers of the resolver.
func TestAuditPlanCmd_ConsumesTheResolver(t *testing.T) {
	if !strings.Contains(readAuditPlanSource(t), "ResolveAuditPlan(") {
		t.Error("audit_plan_cmd.go must call config.ResolveAuditPlan")
	}
}
