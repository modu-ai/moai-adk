package cli

import (
	"os"
	"strings"
	"testing"
)

// codexPhase2Tools is the four-tool surface SPEC-CODEX-PHASE2-001 registers,
// paired with the read-only hint each MUST carry (AC-CX2-016).
//
// The hint is a capability statement, not a style choice: codex_job_status and
// codex_job_result only read a record, while codex_task starts a turn (and may
// write the working tree when the project opted in) and codex_job_cancel
// interrupts a turn and may terminate a process.
var codexPhase2Tools = []struct {
	name         string
	readOnlyHint bool
	// jobIDArg reports whether this tool takes the job-id argument. The three
	// job-control tools do; codex_task takes a prompt instead.
	jobIDArg bool
}{
	{codexTaskToolName, false, false},
	{codexJobStatusToolName, true, true},
	{codexJobResultToolName, true, true},
	{codexJobCancelToolName, false, true},
}

// TestCodexJobTools_RegistrationShape is the binding check for the schema/hint
// half of REQ-CX2-013 (AC-CX2-016). It reads the registered tool set and
// asserts, per tool, a declared input schema and the expected read-only hint.
//
// Both assertions are stated against the VALUE, not the presence, of the field.
// mcp.NewTool seeds every tool with Annotations.ReadOnlyHint = ToBoolPtr(false)
// and a non-nil Properties map (mcp-go v0.57.0 mcp/tools.go:846-861), so the
// pointer is never nil and the schema struct is never absent. A nil-check would
// false-fail a correctly-registered write tool, and a mere "schema is present"
// check could not fail at all — hence len(Properties) > 0 and an equality test
// on the dereferenced hint.
//
// A grep cannot stand in for this test: it cannot distinguish a per-tool
// annotation from one annotation anywhere in the file.
func TestCodexJobTools_RegistrationShape(t *testing.T) {
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	registered := srv.ListTools()

	for _, want := range codexPhase2Tools {
		tool, ok := registered[want.name]
		if !ok || tool == nil {
			t.Errorf("tool %q is not registered (REQ-CX2-013)", want.name)
			continue
		}

		// Declared JSON Schema: per-tool properties, not merely a non-empty
		// struct.
		if n := len(tool.Tool.InputSchema.Properties); n == 0 {
			t.Errorf("tool %q declares no input-schema properties (REQ-CX2-013)", want.name)
		}
		if want.jobIDArg {
			if _, has := tool.Tool.InputSchema.Properties[codexJobIDArg]; !has {
				t.Errorf("tool %q does not declare the %q property (REQ-CX2-013)", want.name, codexJobIDArg)
			}
		}

		// Read-only hint, asserted by value.
		hint := tool.Tool.Annotations.ReadOnlyHint
		if hint == nil {
			t.Errorf("tool %q: ReadOnlyHint is nil — mcp.NewTool always seeds it, so a nil here means the tool was not built with NewTool", want.name)
			continue
		}
		if *hint != want.readOnlyHint {
			t.Errorf("tool %q: ReadOnlyHint = %v, want %v (AC-CX2-016)", want.name, *hint, want.readOnlyHint)
		}
	}
}

// TestCodexPhase2Tools_RegisteredIndependently asserts each of the four tool
// names appears at the registration site as its own quoted literal, matching
// the shape "codex_audit" / "codex_setup" already use.
//
// Each name is asserted INDEPENDENTLY, one assertion per name. A line-counting
// `grep -c` over an alternation is deliberately NOT used: it counts matching
// LINES, so four lines naming only codex_task would clear a `>= 4` gate with
// three tools missing (AC-CX2-016).
//
// The quoted literal and the handler-side constant cannot drift apart:
// TestCodexJobTools_RegistrationShape above looks the registered tool up BY the
// constant, so a literal that stopped matching its constant would fail there.
func TestCodexPhase2Tools_RegisteredIndependently(t *testing.T) {
	src, err := os.ReadFile("mcp_server.go")
	if err != nil {
		t.Fatalf("read mcp_server.go: %v", err)
	}
	body := string(src)

	for _, want := range codexPhase2Tools {
		if !strings.Contains(body, `"`+want.name+`"`) {
			t.Errorf("MISSING %s — not registered as a quoted literal in mcp_server.go (REQ-CX2-013)", want.name)
		}
	}
}

// TestCodexPhase2_NoAskUserQuestion is the REQ-CX2-014 / C-HRA-008 static guard
// for the sources this SPEC added. The tools return structured results; an
// unknown job, a missing input, and a refused write are all results the
// orchestrator translates, never a prompt from inside a subagent.
//
// Shape mirrors internal/cli/worktree/new_test.go::TestNew_NoAskUserQuestion via
// the package's assertNoAskUserQuestionInSource helper (comment lines excluded,
// so documenting the constraint is not a false positive).
func TestCodexPhase2_NoAskUserQuestion(t *testing.T) {
	for _, f := range []string{
		"codex_task.go",
		"codex_job_control.go",
		"codex_jobs.go",
		"mcp_codex.go",
	} {
		assertNoAskUserQuestionInSource(t, f)
	}
}
