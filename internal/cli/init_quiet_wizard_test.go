package cli

// Init execution tests for the quiet init wizard (SPEC-INIT-QUIET-WIZARD-001
// M2; AC-IQW-006, AC-IQW-007a, AC-IQW-008, AC-IQW-009, AC-IQW-011).
//
// Every run goes through the real runInit with the home-safety helper
// (prepareSafeInitHome) in place: the three home seams point under
// t.TempDir, the shell-config step runs into a counting spy, and the 8-item
// real-home fingerprint must be unchanged when the test ends. The helper sets
// environment variables, so no test in this file may run in parallel (Go
// testing panics if one tries).
//
// The interactive path is driven without a TTY through the isInteractiveStdin
// and runWizardFn seams. The injected wizard result carries ONLY the four kept
// answers plus the fixed seeds wizard.RunWithDefaults applies (design.md §6.1),
// so this file compiles both before and after the removed WizardResult fields
// are deleted.
//
// Projects are created through --root without --name, so the project name is
// left to the directory-name default (REQ-IQW-003) rather than taken from a
// positional path string.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/template"
)

// quietWizardProjectDirName is the project directory name every run uses, so
// two runs under different temp roots render the same project.name.
const quietWizardProjectDirName = "quiet-wizard-proj"

// quietWizardLLMYAML is the llm section file name (internal/defs has no
// constant for it).
const quietWizardLLMYAML = "llm.yaml"

// quietWizardComparedSections are the section files REQ-IQW-015 requires to be
// byte-identical between a four-answer interactive run and a flag-less
// non-interactive run.
var quietWizardComparedSections = []string{
	defs.WorkflowYAML,
	defs.ProjectYAML,
	defs.ReportYAML,
	defs.FeedbackYAML,
	quietWizardLLMYAML,
}

// quietWizardKeptAnswers returns the injected wizard result: the four kept
// answers plus the fixed defaults RunWithDefaults seeds, so opts match a
// production interactive run that accepted every default.
func quietWizardKeptAnswers(agentWiring string) *wizard.WizardResult {
	return &wizard.WizardResult{
		ConversationLang:          "en",
		UserName:                  "tester",
		AgentWiring:               agentWiring,
		AutonomyTier:              "semi-auto",
		LSPEnabled:                true,
		EnforceQuality:            true,
		CoverageExemptionsEnabled: false,
		DesignEnabled:             true,
		ClaudeDesignEnabled:       true,
	}
}

// quietWizardRunInit runs the real runInit for one project named
// quietWizardProjectDirName under parentDir, and returns the project directory
// and the captured stdout. A non-nil wiz selects the interactive path (the
// wizard seam returns wiz); nil selects --non-interactive. The home-safety
// helper is installed on t, so every run needs its own t — use a subtest for a
// second run in the same test.
func quietWizardRunInit(t *testing.T, parentDir string, wiz *wizard.WizardResult, flags map[string]string) (projectDir, stdout string) {
	t.Helper()
	prepareSafeInitHome(t)

	// No network update check on either path.
	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	// The wizard seam is swapped on both paths so the call count proves which
	// path runInit actually took: 1 on the interactive path, 0 otherwise.
	wizardCalls := 0
	origWizard := runWizardFn
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
		wizardCalls++
		if wiz == nil {
			return nil, errors.New("wizard seam reached on the non-interactive path")
		}
		return wiz, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	if wiz != nil {
		origInteractive := isInteractiveStdin
		isInteractiveStdin = func() bool { return true }
		t.Cleanup(func() { isInteractiveStdin = origInteractive })
	}

	projectDir = filepath.Join(parentDir, quietWizardProjectDirName)
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("root", projectDir); err != nil {
		t.Fatalf("set --root: %v", err)
	}
	if wiz == nil {
		if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
			t.Fatalf("set --non-interactive: %v", err)
		}
	}
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("runInit: %v (stderr: %s)", err, errBuf.String())
	}

	wantCalls := 0
	if wiz != nil {
		wantCalls = 1
	}
	if wizardCalls != wantCalls {
		t.Fatalf("wizard seam calls = %d, want %d: runInit did not take the intended path", wizardCalls, wantCalls)
	}
	return projectDir, out.String()
}

// quietWizardMismatch is one observer finding: a key whose observed value is
// not the expected default.
type quietWizardMismatch struct {
	key, want, got string
}

func (m quietWizardMismatch) String() string {
	return fmt.Sprintf("%s: want %q, got %q", m.key, m.want, m.got)
}

// quietWizardFormatMismatches renders findings one per line.
func quietWizardFormatMismatches(ms []quietWizardMismatch) string {
	lines := make([]string, 0, len(ms))
	for _, m := range ms {
		lines = append(lines, "  "+m.String())
	}
	return strings.Join(lines, "\n")
}

// quietWizardReadYAML decodes the YAML file at p into v.
func quietWizardReadYAML(p string, v any) error {
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// quietWizardPerformanceTierLine returns the first non-comment
// performance_tier line of an llm.yaml document, verbatim.
func quietWizardPerformanceTierLine(data []byte) (string, bool) {
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "performance_tier:") {
			return line, true
		}
	}
	return "", false
}

// quietWizardTemplatePerformanceTierLine reads the performance_tier line the
// shipped llm.yaml template carries.
func quietWizardTemplatePerformanceTierLine() (string, error) {
	embedFS, err := template.EmbeddedTemplates()
	if err != nil {
		return "", fmt.Errorf("embedded templates: %w", err)
	}
	data, err := fs.ReadFile(embedFS, path.Join(defs.MoAIDir, defs.SectionsSubdir, quietWizardLLMYAML))
	if err != nil {
		return "", fmt.Errorf("read template llm.yaml: %w", err)
	}
	line, ok := quietWizardPerformanceTierLine(data)
	if !ok {
		return "", errors.New("template llm.yaml has no performance_tier line")
	}
	return line, nil
}

// quietWizardObserveDefaults is the observer: it returns every key under
// projectDir whose value is not the default an all-defaults run must leave
// (acceptance.md AC-IQW-006 table). project/report/llm/feedback values are read
// from disk; workflow and feedback values are also read as the config loader
// resolves them. An unreadable input is itself a finding, never a skip.
func quietWizardObserveDefaults(projectDir, wantName string) []quietWizardMismatch {
	var found []quietWizardMismatch
	check := func(key, want, got string) {
		if got != want {
			found = append(found, quietWizardMismatch{key: key, want: want, got: got})
		}
	}
	fail := func(key string, err error) {
		found = append(found, quietWizardMismatch{key: key, want: "readable", got: err.Error()})
	}
	sections := filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir)

	var projectDoc struct {
		Project struct {
			Name string `yaml:"name"`
			Mode string `yaml:"mode"`
		} `yaml:"project"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.ProjectYAML), &projectDoc); err != nil {
		fail("project.yaml", err)
	} else {
		check("project.yaml project.name", wantName, projectDoc.Project.Name)
		check("project.yaml project.mode", "personal", projectDoc.Project.Mode)
	}

	var reportDoc struct {
		Report struct {
			Format string `yaml:"format"`
		} `yaml:"report"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.ReportYAML), &reportDoc); err != nil {
		fail("report.yaml", err)
	} else {
		check("report.yaml report.format", "html+md", reportDoc.Report.Format)
	}

	llmPath := filepath.Join(sections, quietWizardLLMYAML)
	var llmDoc struct {
		LLM struct {
			Profile string `yaml:"profile"`
		} `yaml:"llm"`
	}
	if err := quietWizardReadYAML(llmPath, &llmDoc); err != nil {
		fail("llm.yaml", err)
	} else {
		check("llm.yaml llm.profile", "medium", llmDoc.LLM.Profile)
	}
	if wantLine, err := quietWizardTemplatePerformanceTierLine(); err != nil {
		fail("template llm.yaml performance_tier", err)
	} else if data, err := os.ReadFile(llmPath); err != nil {
		fail("llm.yaml performance_tier", err)
	} else {
		gotLine, _ := quietWizardPerformanceTierLine(data)
		check("llm.yaml llm.performance_tier line", wantLine, gotLine)
	}

	var feedbackDoc struct {
		Feedback struct {
			AutoSubmit *bool `yaml:"auto_submit"`
		} `yaml:"feedback"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.FeedbackYAML), &feedbackDoc); err != nil {
		fail("feedback.yaml", err)
	} else if feedbackDoc.Feedback.AutoSubmit == nil {
		check("feedback.yaml feedback.auto_submit", "false", "absent")
	} else {
		check("feedback.yaml feedback.auto_submit", "false", strconv.FormatBool(*feedbackDoc.Feedback.AutoSubmit))
	}

	// Disk view of workflow.yaml: it must parse (the loader silently keeps
	// its defaults for an unparseable workflow section, which would make the
	// resolved checks below vacuous), and it must carry no todo.enabled key.
	var workflowDoc map[string]any
	if err := quietWizardReadYAML(filepath.Join(sections, defs.WorkflowYAML), &workflowDoc); err != nil {
		fail("workflow.yaml", err)
	} else if wf, ok := workflowDoc["workflow"].(map[string]any); !ok {
		fail("workflow.yaml workflow", errors.New("no workflow mapping"))
	} else {
		todo, _ := wf["todo"].(map[string]any)
		if v, present := todo["enabled"]; present {
			check("workflow.yaml workflow.todo.enabled (disk key)", "absent", fmt.Sprint(v))
		}
	}

	cfg, err := config.NewLoader().Load(filepath.Join(projectDir, defs.MoAIDir))
	if err != nil {
		fail("config loader", err)
	} else {
		check("workflow.worktree.auto_create", "false", strconv.FormatBool(cfg.Workflow.Worktree.AutoCreate))
		check("workflow.todo.enabled (resolved)", "true", strconv.FormatBool(cfg.TodoEnabled()))
		continuation, unmatched := cfg.ProjectContinuation()
		check("workflow.project.continuation", "card", continuation)
		check("workflow.project.continuation (unmatched value)", "", unmatched)
		check("workflow.audit.model", "claude", cfg.Workflow.Audit.Model)
		check("workflow.audit.gates.claude", "required", cfg.Workflow.Audit.Gates.Claude)
		check("workflow.audit.gates.codex", "required", cfg.Workflow.Audit.Gates.Codex)
		check("workflow.audit.gates.glm", "advisory", cfg.Workflow.Audit.Gates.GLM)
		check("workflow.codex.review_gate.enabled", "false", strconv.FormatBool(cfg.Workflow.Codex.ReviewGate.Enabled))
		check("feedback.auto_submit (resolved)", "false", strconv.FormatBool(cfg.Feedback.AutoSubmit))
	}

	var mcpDoc struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if data, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json")); err != nil {
		fail(".mcp.json", err)
	} else if err := json.Unmarshal(data, &mcpDoc); err != nil {
		fail(".mcp.json", err)
	} else if _, hasMoai := mcpDoc.MCPServers["moai"]; !hasMoai {
		check(".mcp.json mcpServers.moai", "present", "absent")
	}

	return found
}

// quietWizardReplaceLine rewrites the single line of the file at p that
// matches pattern (multi-line mode), failing the test unless exactly one line
// matches — a missing target would make the negative control vacuous.
func quietWizardReplaceLine(t *testing.T, p, pattern, replacement string) {
	t.Helper()
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	re := regexp.MustCompile(pattern)
	if n := len(re.FindAllIndex(data, -1)); n != 1 {
		t.Fatalf("%s: pattern %q matched %d lines, want exactly 1", p, pattern, n)
	}
	if err := os.WriteFile(p, re.ReplaceAll(data, []byte(replacement)), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

// quietWizardFirstDiff names the first differing line of two documents.
func quietWizardFirstDiff(a, b []byte) string {
	al := strings.Split(string(a), "\n")
	bl := strings.Split(string(b), "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return fmt.Sprintf("line %d: interactive %q, non-interactive %q", i+1, al[i], bl[i])
		}
	}
	return fmt.Sprintf("line counts differ: interactive %d, non-interactive %d", len(al), len(bl))
}

// TestRunInit_QuietWizardUnsetResolvesToDefaults asserts AC-IQW-006: an
// interactive init that received only the four kept answers leaves every
// removed key at the value an all-defaults run resolves to.
func TestRunInit_QuietWizardUnsetResolvesToDefaults(t *testing.T) {
	projectDir, _ := quietWizardRunInit(t, t.TempDir(), quietWizardKeptAnswers("claude"), nil)

	if found := quietWizardObserveDefaults(projectDir, quietWizardProjectDirName); len(found) != 0 {
		t.Errorf("removed keys did not resolve to their defaults (%d findings):\n%s", len(found), quietWizardFormatMismatches(found))
	}
}

// TestRunInit_QuietWizardObserverDetectsNonDefault asserts AC-IQW-007a, the
// in-test negative control: on a copy of a generated project whose
// project.mode and worktree.auto_create are flipped to non-default values, the
// same observer reports exactly those two keys, while the untouched original
// reports none.
func TestRunInit_QuietWizardObserverDetectsNonDefault(t *testing.T) {
	projectDir, _ := quietWizardRunInit(t, t.TempDir(), quietWizardKeptAnswers("claude"), nil)

	if found := quietWizardObserveDefaults(projectDir, quietWizardProjectDirName); len(found) != 0 {
		t.Fatalf("original project must have 0 findings before the control is meaningful (%d):\n%s", len(found), quietWizardFormatMismatches(found))
	}

	copyDir := filepath.Join(t.TempDir(), "observer-copy")
	srcConfig := filepath.Join(projectDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.CopyFS(filepath.Join(copyDir, defs.MoAIDir, defs.ConfigSubdir), os.DirFS(srcConfig)); err != nil {
		t.Fatalf("copy .moai/config: %v", err)
	}
	mcpData, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(copyDir, ".mcp.json"), mcpData, 0o644); err != nil {
		t.Fatalf("copy .mcp.json: %v", err)
	}

	copySections := filepath.Join(copyDir, defs.MoAIDir, defs.SectionsSubdir)
	quietWizardReplaceLine(t, filepath.Join(copySections, defs.ProjectYAML), `(?m)^([ \t]*)mode: personal$`, "${1}mode: team")
	quietWizardReplaceLine(t, filepath.Join(copySections, defs.WorkflowYAML), `(?m)^([ \t]*)auto_create: false$`, "${1}auto_create: true")

	found := quietWizardObserveDefaults(copyDir, quietWizardProjectDirName)
	keys := make([]string, 0, len(found))
	for _, m := range found {
		keys = append(keys, m.key)
	}
	slices.Sort(keys)
	want := []string{"project.yaml project.mode", "workflow.worktree.auto_create"}
	if !slices.Equal(keys, want) {
		t.Errorf("observer findings on the mutated copy = %q, want exactly %q\n%s", keys, want, quietWizardFormatMismatches(found))
	}
}

// TestRunInit_QuietWizardSectionFilesMatchNonInteractive asserts AC-IQW-008:
// the section files a four-answer interactive run leaves are byte-identical to
// those of a flag-less non-interactive run in a directory of the same name.
// The interactive workflow.yaml carries no workflow.audit model/gates keys and
// no workflow.todo key, while the template's audit.codex / audit.glm pins stay.
func TestRunInit_QuietWizardSectionFilesMatchNonInteractive(t *testing.T) {
	interactiveRoot := t.TempDir()
	nonInteractiveRoot := t.TempDir()
	var interactiveDir, nonInteractiveDir string

	t.Run("interactive", func(t *testing.T) {
		interactiveDir, _ = quietWizardRunInit(t, interactiveRoot, quietWizardKeptAnswers("claude"), nil)
	})
	t.Run("non_interactive", func(t *testing.T) {
		nonInteractiveDir, _ = quietWizardRunInit(t, nonInteractiveRoot, nil, nil)
	})
	if interactiveDir == "" || nonInteractiveDir == "" {
		t.Fatal("an init run did not complete; the section files cannot be compared")
	}

	for _, name := range quietWizardComparedSections {
		a, errA := os.ReadFile(filepath.Join(interactiveDir, defs.MoAIDir, defs.SectionsSubdir, name))
		b, errB := os.ReadFile(filepath.Join(nonInteractiveDir, defs.MoAIDir, defs.SectionsSubdir, name))
		if errA != nil || errB != nil {
			t.Errorf("%s: read interactive=%v non-interactive=%v", name, errA, errB)
			continue
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s differs between the interactive and the non-interactive run; first difference: %s", name, quietWizardFirstDiff(a, b))
		}
	}

	var workflowDoc map[string]any
	if err := quietWizardReadYAML(filepath.Join(interactiveDir, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML), &workflowDoc); err != nil {
		t.Fatalf("parse interactive workflow.yaml: %v", err)
	}
	wf, ok := workflowDoc["workflow"].(map[string]any)
	if !ok {
		t.Fatal("interactive workflow.yaml has no workflow mapping")
	}
	if _, present := wf["todo"]; present {
		t.Errorf("interactive workflow.yaml carries a workflow.todo key: %v", wf["todo"])
	}
	audit, ok := wf["audit"].(map[string]any)
	if !ok {
		t.Fatal("interactive workflow.yaml has no workflow.audit mapping (the template ships the codex/glm pins there)")
	}
	for _, k := range []string{"model", "gates"} {
		if _, present := audit[k]; present {
			t.Errorf("interactive workflow.yaml carries workflow.audit.%s: %v", k, audit[k])
		}
	}
	for _, k := range []string{"codex", "glm"} {
		if _, present := audit[k]; !present {
			t.Errorf("interactive workflow.yaml lost the template pin workflow.audit.%s", k)
		}
	}
}

// TestRunInit_QuietWizardProvisionsMCPByDefault asserts AC-IQW-009: with no
// MCP field in the injected result, the interactive path runs the .mcp.json
// ensure-entry call by default — announced for claude and both — and the codex
// harness rule still skips it.
func TestRunInit_QuietWizardProvisionsMCPByDefault(t *testing.T) {
	cases := []struct {
		wiring           string
		wantAnnouncement bool
	}{
		{wiring: "claude", wantAnnouncement: true},
		{wiring: "codex", wantAnnouncement: false},
		{wiring: "both", wantAnnouncement: true},
	}
	for _, tc := range cases {
		t.Run(tc.wiring, func(t *testing.T) {
			projectDir, stdout := quietWizardRunInit(t, t.TempDir(), quietWizardKeptAnswers(tc.wiring), nil)

			if got := strings.Contains(stdout, mcpProvisionAnnouncement); got != tc.wantAnnouncement {
				t.Errorf("harness %s: provisioning announcement present = %t, want %t\nstdout:\n%s", tc.wiring, got, tc.wantAnnouncement, stdout)
			}
			if tc.wiring == "codex" {
				// Reachability: the codex selection reached the wiring consumer,
				// so the absent announcement is the harness rule, not a lost answer.
				assertCodexArtifacts(t, projectDir, true)
			}
			if tc.wantAnnouncement {
				var mcpDoc struct {
					MCPServers map[string]json.RawMessage `json:"mcpServers"`
				}
				data, err := os.ReadFile(filepath.Join(projectDir, ".mcp.json"))
				if err != nil {
					t.Fatalf("read .mcp.json: %v", err)
				}
				if err := json.Unmarshal(data, &mcpDoc); err != nil {
					t.Fatalf("parse .mcp.json: %v", err)
				}
				if _, ok := mcpDoc.MCPServers["moai"]; !ok {
					t.Errorf("harness %s: .mcp.json has no mcpServers.moai entry", tc.wiring)
				}
			}
		})
	}
}

// TestRunInit_QuietWizardFlagsStillPersist asserts AC-IQW-011: on the
// interactive path, --project-mode team and --worktree-auto-create=true are
// still written.
func TestRunInit_QuietWizardFlagsStillPersist(t *testing.T) {
	projectDir, _ := quietWizardRunInit(t, t.TempDir(), quietWizardKeptAnswers("claude"), map[string]string{
		"project-mode":         "team",
		"worktree-auto-create": "true",
	})
	sections := filepath.Join(projectDir, defs.MoAIDir, defs.SectionsSubdir)

	var projectDoc struct {
		Project struct {
			Mode string `yaml:"mode"`
		} `yaml:"project"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.ProjectYAML), &projectDoc); err != nil {
		t.Fatalf("parse project.yaml: %v", err)
	}
	if projectDoc.Project.Mode != "team" {
		t.Errorf("project.yaml project.mode = %q, want %q (--project-mode team)", projectDoc.Project.Mode, "team")
	}

	var workflowDoc struct {
		Workflow struct {
			Worktree struct {
				AutoCreate *bool `yaml:"auto_create"`
			} `yaml:"worktree"`
		} `yaml:"workflow"`
	}
	if err := quietWizardReadYAML(filepath.Join(sections, defs.WorkflowYAML), &workflowDoc); err != nil {
		t.Fatalf("parse workflow.yaml: %v", err)
	}
	switch ac := workflowDoc.Workflow.Worktree.AutoCreate; {
	case ac == nil:
		t.Errorf("workflow.yaml workflow.worktree.auto_create is absent, want true (--worktree-auto-create=true)")
	case !*ac:
		t.Errorf("workflow.yaml workflow.worktree.auto_create = false, want true (--worktree-auto-create=true)")
	}
}
