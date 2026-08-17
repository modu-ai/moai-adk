package cli

import (
	"os"
	"strings"
	"testing"
)

// glmTaskFamilyTools is the four-tool GLM delegation surface, paired with the
// read-only hint each MUST carry. Mirrors codexPhase2Tools.
//
// The hint is a capability statement, not a style choice: glm_job_status and
// glm_job_result only read a record, while glm_task starts a model call and
// writes a job record, and glm_job_cancel revokes a running job's context and
// records the cancellation.
var glmTaskFamilyTools = []struct {
	name         string
	readOnlyHint bool
	// jobIDArg reports whether this tool takes the job-id argument. The three
	// job-control tools do; glm_task takes a prompt instead.
	jobIDArg bool
}{
	{glmTaskToolName, false, false},
	{glmJobStatusToolName, true, true},
	{glmJobResultToolName, true, true},
	{glmJobCancelToolName, false, true},
}

// TestGLMJobTools_RegistrationShape is the binding check for the schema/hint
// half of registration. It reads the registered tool set and asserts, per
// tool, a declared input schema and the expected read-only hint — by VALUE,
// not presence (mcp.NewTool seeds ReadOnlyHint to a non-nil false and a
// non-nil Properties map, so nil/presence checks could not fail). Mirrors
// TestCodexJobTools_RegistrationShape.
func TestGLMJobTools_RegistrationShape(t *testing.T) {
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	registered := srv.ListTools()

	for _, want := range glmTaskFamilyTools {
		tool, ok := registered[want.name]
		if !ok || tool == nil {
			t.Errorf("tool %q is not registered", want.name)
			continue
		}

		// Declared JSON Schema: per-tool properties, not merely a non-empty
		// struct.
		if n := len(tool.Tool.InputSchema.Properties); n == 0 {
			t.Errorf("tool %q declares no input-schema properties", want.name)
		}
		if want.jobIDArg {
			if _, has := tool.Tool.InputSchema.Properties[glmJobIDArg]; !has {
				t.Errorf("tool %q does not declare the %q property", want.name, glmJobIDArg)
			}
		}

		// Read-only hint, asserted by value.
		hint := tool.Tool.Annotations.ReadOnlyHint
		if hint == nil {
			t.Errorf("tool %q: ReadOnlyHint is nil — mcp.NewTool always seeds it, so a nil here means the tool was not built with NewTool", want.name)
			continue
		}
		if *hint != want.readOnlyHint {
			t.Errorf("tool %q: ReadOnlyHint = %v, want %v", want.name, *hint, want.readOnlyHint)
		}
	}
}

// TestGLMTaskFamily_RegisteredIndependently asserts each of the four tool
// names appears at the registration site as its own quoted literal. Each name
// is asserted INDEPENDENTLY — a line-counting grep over an alternation counts
// matching LINES, so four lines naming only glm_task would clear a >= 4 gate
// with three tools missing. Mirrors TestCodexPhase2Tools_RegisteredIndependently.
func TestGLMTaskFamily_RegisteredIndependently(t *testing.T) {
	src, err := os.ReadFile("mcp_server.go")
	if err != nil {
		t.Fatalf("read mcp_server.go: %v", err)
	}
	body := string(src)

	for _, want := range glmTaskFamilyTools {
		if !strings.Contains(body, `"`+want.name+`"`) {
			t.Errorf("MISSING %s — not registered as a quoted literal in mcp_server.go", want.name)
		}
	}
}

// TestGLMTaskFamily_NoAskUserQuestion is the C-HRA-008 static guard for the
// sources this family added: the tools return structured results; an unknown
// job, a missing input, and an unavailable GLM are all results the orchestrator
// translates, never a prompt from inside a subagent. Mirrors
// TestCodexPhase2_NoAskUserQuestion via the package's
// assertNoAskUserQuestionInSource helper.
func TestGLMTaskFamily_NoAskUserQuestion(t *testing.T) {
	for _, f := range []string{
		"glm_task.go",
		"glm_job_control.go",
		"glm_jobs.go",
	} {
		assertNoAskUserQuestionInSource(t, f)
	}
}
