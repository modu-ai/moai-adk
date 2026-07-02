// Package harness — v4 harness reference-integrity smoke gate tests.
// SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-005 (AC-HEP-005a/005b/006, EC-4).
package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// doctorManifestJSON returns a schema-valid v4 manifest for the named harness.
func doctorManifestJSON(name string) string {
	return `{
  "name": "` + name + `",
  "domain": "test domain",
  "source_request": "test request",
  "patterns": ["Pipeline"],
  "specialists": [{"role":"tester","primitive":"sub-agent","isolation":"none","effort":"low","model":"inherit"}],
  "sprint_contract": {"dimensions":["correctness"],"thresholds":{"correctness":0.8}},
  "entry_command": "/harness:` + name + `",
  "runner_workflow": "harness-` + name + `-run.js"
}`
}

// doctorRunnerJS returns a minimal Runner referencing the given manifest path
// constant and specialist agent name.
func doctorRunnerJS(manifestPath, agentRef string) string {
	return `// harness-run.js
const MANIFEST_PATH = "` + manifestPath + `";
async function run({ agent }) {
  // delegates human-gated work to ` + agentRef + ` sub-agent.
  return { manifest: MANIFEST_PATH };
}
module.exports = { run, MANIFEST_PATH };
`
}

// buildHarnessFixture writes a full v4 harness (command + manifest + Runner) and
// optionally the specialist agent. manifestPathConst and agentRef are embedded
// verbatim into the Runner so tests can inject B5 defect variants. Reuses the
// package-level writeFile helper (v4lifecycle_test.go).
func buildHarnessFixture(t *testing.T, root, name, manifestPathConst, agentRef string, createAgent bool) {
	t.Helper()
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name+".md"), "# /harness:"+name+"\nThin wrapper.\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name, "manifest.json"), doctorManifestJSON(name))
	writeFile(t, filepath.Join(root, ".claude", "workflows", "harness-"+name+"-run.js"), doctorRunnerJS(manifestPathConst, agentRef))
	if createAgent {
		writeFile(t, filepath.Join(root, ".claude", "agents", "harness", agentRef+".md"), "# "+agentRef+"\nSpecialist.\n")
	}
}

// TestDoctor_ValidHarness_Passes verifies AC-HEP-005a: a fully-wired v4 harness
// (valid manifest + Runner whose MANIFEST_PATH resolves + existing specialist)
// yields 0 ERROR findings.
func TestDoctor_ValidHarness_Passes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	buildHarnessFixture(t, root, "goodh",
		".claude/commands/harness/goodh/manifest.json",
		"harness-goodh-specialist", true)

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.Harnesses != 1 {
		t.Errorf("harnesses = %d, want 1", report.Harnesses)
	}
	if report.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0; findings=%+v", report.ErrorCount, report.Findings)
	}
}

// TestDoctor_ZeroHarness_Graceful verifies AC-HEP-005b: an empty project (no
// .claude/commands/harness/) yields 0 harnesses, 0 findings, no error.
func TestDoctor_ZeroHarness_Graceful(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.Harnesses != 0 {
		t.Errorf("harnesses = %d, want 0", report.Harnesses)
	}
	if report.ErrorCount != 0 || len(report.Findings) != 0 {
		t.Errorf("expected 0 findings on empty project, got %+v", report.Findings)
	}
}

// TestDoctor_DefectClass_Detected verifies AC-HEP-006: the B5 defect class
// (Runner manifest-path const pointing at a non-existent path + non-existent
// specialist agent reference) yields >= 2 ERROR findings.
func TestDoctor_DefectClass_Detected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Defect 1: MANIFEST_PATH points at a path with no <name> subdir (non-existent).
	// Defect 2: agent ref harness-nonexistent-specialist has no agent file.
	buildHarnessFixture(t, root, "badh",
		".claude/commands/harness/manifest.json",
		"harness-nonexistent-specialist", false)

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	errCount := report.ErrorCount
	if errCount < 2 {
		t.Fatalf("error_count = %d, want >= 2 (wrong-manifest-path + unresolved-agent); findings=%+v", errCount, report.Findings)
	}
	// Confirm the two distinct defect axes are present.
	sawRunner, sawAgent := false, false
	for _, f := range report.Findings {
		if f.Severity != SeverityError {
			continue
		}
		if f.Axis == "runner" {
			sawRunner = true
		}
		if f.Axis == "agent" {
			sawAgent = true
		}
	}
	if !sawRunner {
		t.Errorf("expected a runner-axis ERROR (wrong MANIFEST_PATH); findings=%+v", report.Findings)
	}
	if !sawAgent {
		t.Errorf("expected an agent-axis ERROR (unresolved specialist); findings=%+v", report.Findings)
	}
}

// TestDoctor_ThinHarness_InfoNotError verifies EC-4 + Doctor severity policy: a
// command-only thin harness (manifest + Runner absent) is reported as an INFO
// note, NOT an ERROR — the global smoke does not false-fail on github/release.
func TestDoctor_ThinHarness_InfoNotError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Command file only — no manifest, no Runner.
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "thinh.md"), "# /harness:thinh\nThin.\n")

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Errorf("error_count = %d, want 0 (thin harness must not ERROR); findings=%+v", report.ErrorCount, report.Findings)
	}
	if report.InfoCount < 1 {
		t.Errorf("info_count = %d, want >= 1 (thin harness INFO note)", report.InfoCount)
	}
}

// TestDoctorCmd_ExitCodes verifies the cobra command surface: a valid harness
// returns nil (exit 0); a defect harness returns an error (non-0 exit).
func TestDoctorCmd_ExitCodes(t *testing.T) {
	t.Parallel()

	// Valid → nil error.
	rootOK := t.TempDir()
	buildHarnessFixture(t, rootOK, "goodh",
		".claude/commands/harness/goodh/manifest.json",
		"harness-goodh-specialist", true)
	cmd := NewHarnessDoctorCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--project-root", rootOK})
	if err := cmd.Execute(); err != nil {
		t.Errorf("valid harness doctor: err = %v, want nil (exit 0)", err)
	}

	// Defect → non-nil error (exit 1).
	rootBad := t.TempDir()
	buildHarnessFixture(t, rootBad, "badh",
		".claude/commands/harness/manifest.json",
		"harness-nonexistent-specialist", false)
	cmd2 := NewHarnessDoctorCmd()
	cmd2.SetOut(&bytes.Buffer{})
	cmd2.SetErr(&bytes.Buffer{})
	cmd2.SetArgs([]string{"--project-root", rootBad})
	if err := cmd2.Execute(); err == nil {
		t.Errorf("defect harness doctor: err = nil, want non-nil (exit 1)")
	}
}

// TestDoctor_MalformedManifest verifies the manifest-axis ERROR path when the
// manifest.json is present but not valid JSON.
func TestDoctor_MalformedManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "brokenh.md"), "# thin\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "brokenh", "manifest.json"), "{ not valid json ]")

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.ErrorCount < 1 {
		t.Fatalf("error_count = %d, want >= 1 (malformed manifest JSON)", report.ErrorCount)
	}
	sawManifest := false
	for _, f := range report.Findings {
		if f.Axis == "manifest" && f.Severity == SeverityError {
			sawManifest = true
		}
	}
	if !sawManifest {
		t.Errorf("expected a manifest-axis ERROR; findings=%+v", report.Findings)
	}
}

// TestDoctor_RunnerMissing verifies the runner-axis ERROR path when a valid
// manifest declares a Runner that does not exist on disk.
func TestDoctor_RunnerMissing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "norunner.md"), "# thin\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "norunner", "manifest.json"), doctorManifestJSON("norunner"))
	// No .claude/workflows/harness-norunner-run.js written.

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	sawRunner := false
	for _, f := range report.Findings {
		if f.Axis == "runner" && f.Severity == SeverityError {
			sawRunner = true
		}
	}
	if !sawRunner {
		t.Errorf("expected a runner-axis ERROR (Runner file missing); findings=%+v", report.Findings)
	}
}

// TestDoctorCmd_JSONOutput verifies the --json branch emits a parseable report.
func TestDoctorCmd_JSONOutput(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	buildHarnessFixture(t, root, "goodh",
		".claude/commands/harness/goodh/manifest.json",
		"harness-goodh-specialist", true)

	cmd := NewHarnessDoctorCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--json", "--project-root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var got DoctorReport
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout.String())
	}
	if got.Harnesses != 1 || got.ErrorCount != 0 {
		t.Errorf("report = %+v, want 1 harness / 0 errors", got)
	}
}

// TestDoctor_NoAskUserQuestion is the C-HRA-008 static guard for doctor.go: the
// smoke-gate source MUST NOT reference the deferred user-question tool.
func TestDoctor_NoAskUserQuestion(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("doctor.go")
	if err != nil {
		t.Fatalf("read doctor.go: %v", err)
	}
	for _, needle := range []string{"AskUserQuestion", "mcp__askuser"} {
		if bytes.Contains(data, []byte(needle)) {
			t.Errorf("doctor.go references %q — subagent-boundary violation (C-HRA-008)", needle)
		}
	}
}
