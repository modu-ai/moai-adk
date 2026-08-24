package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// guardianPayload is a Write tool_input whose content trips the guardian's
// critical-severity hardcoded-secret class.
const guardianPayload = `{"content":"api_key = \"sk-live-abcdef0123456789\"","file_path":"a.py"}`

// guardianStandaloneEnvelope wraps guardianPayload in the PostToolUse envelope
// HandleSecurityScan reads from stdin.
const guardianStandaloneEnvelope = `{"tool_name":"Write","tool_input":` + guardianPayload + `}`

func guardianInput() *HookInput {
	return &HookInput{
		ToolName:  "Write",
		ToolInput: json.RawMessage(guardianPayload),
		SessionID: "test-session",
	}
}

// standaloneAdditionalContext runs HandleSecurityScan directly and returns the
// additionalContext string it emits.
func standaloneAdditionalContext(t *testing.T) string {
	t.Helper()
	var out bytes.Buffer
	if err := security.HandleSecurityScan(nil, strings.NewReader(guardianStandaloneEnvelope), &out, t.TempDir()); err != nil {
		t.Fatalf("HandleSecurityScan returned error: %v", err)
	}
	var m struct {
		HookSpecificOutput struct {
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &m); err != nil {
		t.Fatalf("standalone output is not valid JSON: %q (%v)", out.String(), err)
	}
	if m.HookSpecificOutput.AdditionalContext == "" {
		t.Fatalf("standalone handler produced no additionalContext for %q", guardianStandaloneEnvelope)
	}
	return m.HookSpecificOutput.AdditionalContext
}

// TestPostToolGuardianMergeKeepsBothAdvisories — AC-SSS-012.
// Both the post-tool handler's systemMessage and the guardian's
// additionalContext survive a single dispatch, and a preceding handler's block
// decision does not prevent the guardian scan from being reached (spec §A.3).
func TestPostToolGuardianMergeKeepsBothAdvisories(t *testing.T) {
	t.Parallel()

	const postToolText = "post-tool advisory: 1 lint error"

	t.Run("both advisories survive the merge", func(t *testing.T) {
		reg := NewRegistry(&mockConfigProvider{cfg: newTestConfig()})
		reg.Register(&mockHandler{
			event:  EventPostToolUse,
			output: &HookOutput{SystemMessage: postToolText},
		})
		reg.Register(NewPostToolGuardianHandler())

		out, err := reg.Dispatch(context.Background(), EventPostToolUse, guardianInput())
		if err != nil {
			t.Fatalf("Dispatch error: %v", err)
		}
		if out.SystemMessage != postToolText {
			t.Errorf("systemMessage lost or overwritten: got %q, want %q", out.SystemMessage, postToolText)
		}
		if out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext == "" {
			t.Fatalf("guardian additionalContext dropped: %+v", out.HookSpecificOutput)
		}
		ac := out.HookSpecificOutput.AdditionalContext
		if !strings.Contains(ac, "hardcoded-secret") {
			t.Errorf("additionalContext does not carry the guardian finding: %q", ac)
		}
		if strings.Contains(ac, postToolText) {
			t.Errorf("additionalContext contains the post-tool text: %q", ac)
		}
		if strings.Contains(out.SystemMessage, ac) {
			t.Errorf("systemMessage contains the guardian text: %q", out.SystemMessage)
		}
	})

	t.Run("a preceding block does not short-circuit the guardian", func(t *testing.T) {
		reg := NewRegistry(&mockConfigProvider{cfg: newTestConfig()})
		reg.Register(&mockHandler{
			event:  EventPostToolUse,
			output: &HookOutput{Decision: DecisionBlock, Reason: "blocked by an earlier handler"},
		})
		reg.Register(NewPostToolGuardianHandler())

		out, err := reg.Dispatch(context.Background(), EventPostToolUse, guardianInput())
		if err != nil {
			t.Fatalf("Dispatch error: %v", err)
		}
		if out.Decision != DecisionBlock {
			t.Errorf("the preceding handler's block decision must survive: %+v", out)
		}
		if out.HookSpecificOutput == nil || out.HookSpecificOutput.AdditionalContext == "" {
			t.Fatalf("guardian scan was short-circuited by the preceding block: %+v", out)
		}
	})
}

// TestMergedGuardianTextMatchesStandalone — AC-SSS-013. The merged handler's
// additionalContext is byte-identical to what HandleSecurityScan emits.
func TestMergedGuardianTextMatchesStandalone(t *testing.T) {
	t.Parallel()

	want := standaloneAdditionalContext(t)

	out, err := NewPostToolGuardianHandler().Handle(context.Background(), guardianInput())
	if err != nil {
		t.Fatalf("guardian handler returned error: %v", err)
	}
	if out == nil || out.HookSpecificOutput == nil {
		t.Fatalf("guardian handler produced no output for a finding-bearing payload")
	}
	if got := out.HookSpecificOutput.AdditionalContext; got != want {
		t.Errorf("advisory text changed by the merge:\n got: %q\nwant: %q", got, want)
	}
}

// TestMergedGuardianNeverBlocks — AC-SSS-014. The guardian introduces no
// decision, no permissionDecision, no continue:false, and no non-zero exit.
func TestMergedGuardianNeverBlocks(t *testing.T) {
	t.Parallel()

	// The table covers every branch of the handler's tool-name switch, since an
	// untested tool name is an untested content-extraction path: Edit reaches
	// content through new_string, MultiEdit concatenates every edits[].new_string,
	// and an unsupported name must stay silent even on a finding-bearing payload.
	const secretLine = `api_key = \"sk-live-abcdef0123456789\"`
	cases := map[string]struct {
		tool string
		// payload is the tool_input blob handed to the handler.
		payload string
		// wantFinding requires the guardian advisory to carry the finding.
		// wantSilent requires the handler to produce no output at all.
		wantFinding bool
		wantSilent  bool
	}{
		"critical finding": {tool: "Write", payload: guardianPayload, wantFinding: true},
		"clean payload":    {tool: "Write", payload: `{"content":"func add(a, b int) int { return a + b }","file_path":"a.go"}`, wantSilent: true},
		"empty content":    {tool: "Write", payload: `{"file_path":"a.go"}`, wantSilent: true},
		"edit new_string":  {tool: "Edit", payload: `{"file_path":"a.py","new_string":"` + secretLine + `"}`, wantFinding: true},
		// The finding sits in the SECOND edit: reaching it proves the whole
		// batch is scanned, not just the first entry.
		"multiedit batch":  {tool: "MultiEdit", payload: `{"file_path":"a.py","edits":[{"new_string":"x = 1"},{"new_string":"` + secretLine + `"}]}`, wantFinding: true},
		"unsupported tool": {tool: "Read", payload: guardianPayload, wantSilent: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			in := &HookInput{ToolName: tc.tool, ToolInput: json.RawMessage(tc.payload)}
			out, err := NewPostToolGuardianHandler().Handle(context.Background(), in)
			if err != nil {
				t.Fatalf("guardian handler must be fail-open, got error: %v", err)
			}
			if tc.wantSilent && out != nil {
				t.Fatalf("handler must stay silent for %s, got %+v", tc.tool, out)
			}
			if out == nil {
				if tc.wantFinding {
					t.Fatalf("%s payload carries a finding but the handler produced no output", tc.tool)
				}
				return // silent pass carries nothing further to assert
			}
			if tc.wantFinding {
				if out.HookSpecificOutput == nil {
					t.Fatalf("%s finding dropped: no hookSpecificOutput", tc.tool)
				}
				if ac := out.HookSpecificOutput.AdditionalContext; !strings.Contains(ac, "hardcoded-secret") {
					t.Errorf("%s advisory does not carry the finding: %q", tc.tool, ac)
				}
			}
			if out.Decision != "" {
				t.Errorf("guardian must not set a decision, got %q", out.Decision)
			}
			if out.Continue != nil {
				t.Errorf("guardian must not set continue, got %v", *out.Continue)
			}
			if out.ExitCode != 0 {
				t.Errorf("guardian must not request a non-zero exit, got %d", out.ExitCode)
			}
			if out.HookSpecificOutput != nil {
				if out.HookSpecificOutput.PermissionDecision != "" {
					t.Errorf("guardian must not set permissionDecision, got %q", out.HookSpecificOutput.PermissionDecision)
				}
				if out.HookSpecificOutput.Decision != nil {
					t.Errorf("guardian must not set a decision object, got %+v", out.HookSpecificOutput.Decision)
				}
			}
		})
	}
}

// TestGuardianHonorsDispatchContext — the registry gives every handler a
// deadline (context.WithTimeout in Dispatch), so an already-expired context
// must yield the same silent nil as an unreadable payload rather than a scan
// that outruns the hook budget.
func TestGuardianHonorsDispatchContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out, err := NewPostToolGuardianHandler().Handle(ctx, guardianInput())
	if err != nil {
		t.Fatalf("an expired context must stay fail-open, got error: %v", err)
	}
	if out != nil {
		t.Errorf("handler scanned past an expired deadline: %+v", out)
	}

	// Control: the same payload on a live context still produces the advisory,
	// so the guard above is not silencing every scan.
	live, err := NewPostToolGuardianHandler().Handle(context.Background(), guardianInput())
	if err != nil || live == nil || live.HookSpecificOutput == nil {
		t.Fatalf("control scan on a live context produced no advisory: out=%+v err=%v", live, err)
	}
}
