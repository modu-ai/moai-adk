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

	cases := map[string]string{
		"critical finding": guardianPayload,
		"clean payload":    `{"content":"func add(a, b int) int { return a + b }","file_path":"a.go"}`,
		"empty content":    `{"file_path":"a.go"}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			in := &HookInput{ToolName: "Write", ToolInput: json.RawMessage(payload)}
			out, err := NewPostToolGuardianHandler().Handle(context.Background(), in)
			if err != nil {
				t.Fatalf("guardian handler must be fail-open, got error: %v", err)
			}
			if out == nil {
				return // silent pass carries nothing to assert
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
