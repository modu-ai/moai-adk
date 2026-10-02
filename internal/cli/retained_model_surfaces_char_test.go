package cli

// Characterisation tests for the model/effort surfaces that stay when subagents
// inherit the main session's model and effort: the main-session effort resolved
// from a preference profile, the GLM model alias mapping and session reasoning
// state, and the cross-model audit pin precedence with its backend-default
// fallback. They pin the behaviour observed before any removal so later
// milestones can prove it unchanged. Not parallel-safe: they swap package-level
// seams (profile.BaseDirOverride, launchEffortPrefsFn, projectDirResolver).

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
)

func TestCharacterize_LaunchEffortFromPreferenceProfile(t *testing.T) {
	base := t.TempDir()
	origBase := profile.BaseDirOverride
	profile.BaseDirOverride = base
	t.Cleanup(func() { profile.BaseDirOverride = origBase })
	origFn := launchEffortPrefsFn
	launchEffortPrefsFn = profile.ReadPreferences
	t.Cleanup(func() { launchEffortPrefsFn = origFn })

	cases := []struct {
		name        string
		prefs       profile.ProfilePreferences
		wantPayload map[string]any
		wantArgs    []string
	}{
		{"policy high", profile.ProfilePreferences{ModelPolicy: "high"}, map[string]any{"effortLevel": "high"}, nil},
		{"policy medium", profile.ProfilePreferences{ModelPolicy: "medium"}, map[string]any{"effortLevel": "medium"}, nil},
		{"policy low", profile.ProfilePreferences{ModelPolicy: "low"}, map[string]any{"effortLevel": "low"}, nil},
		{"explicit effort wins", profile.ProfilePreferences{EffortLevel: "xhigh", ModelPolicy: "low"}, map[string]any{"effortLevel": "xhigh"}, nil},
		{"max leaves the payload", profile.ProfilePreferences{EffortLevel: "max"}, nil, []string{"--effort", "max"}},
		{"unknown policy injects nothing", profile.ProfilePreferences{ModelPolicy: "bogus"}, nil, nil},
		{"empty injects nothing", profile.ProfilePreferences{}, nil, nil},
	}
	for i, tc := range cases {
		name := "char" + string(rune('a'+i))
		if err := profile.WritePreferences(name, tc.prefs); err != nil {
			t.Fatalf("%s: write preferences: %v", tc.name, err)
		}
		payload, args := applyLaunchEffort(nil, name)
		if !reflect.DeepEqual(payload, tc.wantPayload) {
			t.Errorf("%s: payload = %#v, want %#v", tc.name, payload, tc.wantPayload)
		}
		if !reflect.DeepEqual(args, tc.wantArgs) {
			t.Errorf("%s: args = %#v, want %#v", tc.name, args, tc.wantArgs)
		}
	}
}

func TestCharacterize_GLMAliasMapping(t *testing.T) {
	d := config.NewDefaultLLMConfig().GLM.Models
	cases := []struct {
		name                           string
		in                             config.GLMModels
		wantH, wantM, wantL, wantFable string
	}{
		{"empty falls back to defaults", config.GLMModels{}, d.High, d.Medium, d.Low, d.Fable},
		// The legacy opus/sonnet/haiku alias FIELDS are deleted (REQ-MMU-004,
		// DR-2) — the characterization now pins the removal behavior instead:
		// a removed or unknown id falls back to the tier default (never a
		// silent pass-through).
		{"removed ids fall back to defaults", config.GLMModels{High: "glm-4.7", Medium: "glm-5.1", Low: "glm-4.6", Fable: "glm-5.2"}, d.High, d.Medium, d.Low, d.Fable},
		{"offered ids pass through", config.GLMModels{High: "glm-5.3", Medium: "glm-5.3-flash", Low: "glm-5.3-flash", Fable: "glm-5.3"}, "glm-5.3", "glm-5.3-flash", "glm-5.3-flash", "glm-5.3"},
	}
	for _, tc := range cases {
		h, m, l, f := resolveGLMModels(tc.in)
		if h != tc.wantH || m != tc.wantM || l != tc.wantL || f != tc.wantFable {
			t.Errorf("%s: resolveGLMModels = (%q,%q,%q,%q), want (%q,%q,%q,%q)",
				tc.name, h, m, l, f, tc.wantH, tc.wantM, tc.wantL, tc.wantFable)
		}
	}
}

func TestCharacterize_GLMSessionReasoning(t *testing.T) {
	low := template.GLMStateLow
	maxS := template.GLMStateMax
	for effort, want := range map[string]string{
		"low": low, "medium": maxS, "high": maxS, "xhigh": maxS, "max": maxS, "bogus": maxS,
	} {
		if got := template.CollapseClaudeEffortToGLM(effort).Name; got != want {
			t.Errorf("CollapseClaudeEffortToGLM(%q) = %q, want %q", effort, got, want)
		}
		if got := template.CollapseClaudeEffortToGLMForModel(config.DefaultGLM53Flash, effort).Name; got != maxS {
			t.Errorf("CollapseClaudeEffortToGLMForModel(flash, %q) = %q, want %q", effort, got, maxS)
		}
	}
	if got := template.SessionGLMReasoningState().Name; got != maxS {
		t.Errorf("SessionGLMReasoningState = %q, want %q", got, maxS)
	}
	if got := template.SessionGLMReasoningStateForEffort("").Name; got != maxS {
		t.Errorf("SessionGLMReasoningStateForEffort(\"\") = %q, want %q", got, maxS)
	}
	if got := template.SessionGLMReasoningStateForEffort("low").Name; got != low {
		t.Errorf("SessionGLMReasoningStateForEffort(low) = %q, want %q", got, low)
	}
	if got := template.SessionGLMReasoningStateForModel("glm-5.3", "low").Name; got != low {
		t.Errorf("SessionGLMReasoningStateForModel(glm-5.3, low) = %q, want %q", got, low)
	}
	if got := template.SessionGLMReasoningStateForModel(config.DefaultGLM53Flash, "low").Name; got != maxS {
		t.Errorf("SessionGLMReasoningStateForModel(flash, low) = %q, want %q", got, maxS)
	}
}

// charAuditRoot returns a project root with no llm.yaml and, when pins is
// non-empty, a workflow.yaml carrying that workflow.audit body.
func charAuditRoot(t *testing.T, pins string) string {
	t.Helper()
	root := t.TempDir()
	if pins == "" {
		return root
	}
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte("workflow:\n  audit:\n"+pins), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCharacterize_AuditPinPrecedenceAndBackendDefault(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	t.Cleanup(func() { projectDirResolver = old })

	pinned := charAuditRoot(t,
		"    codex:\n      model: gpt-5-codex\n      effort: high\n"+
			"    glm:\n      model: glm-4.6\n      effort: low\n")
	projectDirResolver = func() string { return pinned }

	if got := resolveCodexAuditModelEffort(map[string]any{"cwd": pinned}); got != (config.ModelEffort{Model: "gpt-5-codex", Effort: "high"}) {
		t.Errorf("codex audit with pin = %+v, want the pin", got)
	}
	if got := resolveGLMAuditModelEffort(pinned); got != (config.ModelEffort{Model: "glm-4.6", Effort: "low"}) {
		t.Errorf("glm audit with pin = %+v, want the pin", got)
	}

	bare := charAuditRoot(t, "")
	projectDirResolver = func() string { return bare }

	if got := resolveCodexAuditModelEffort(map[string]any{"cwd": bare}); got != (config.ModelEffort{Model: codexAuditDefaultModel, Effort: codexAuditDefaultEffort}) {
		t.Errorf("codex audit without pin = %+v, want the {%s, %s} default pin (REQ-MMU-001 — the former zero value fell with superseded REQ-AMP-005)", got, codexAuditDefaultModel, codexAuditDefaultEffort)
	}
	if got := resolveCodexModelEffort(map[string]any{"cwd": bare}); got != (config.ModelEffort{}) {
		t.Errorf("codex task without pin = %+v, want the zero value", got)
	}
	if got := resolveGLMAuditModelEffort(bare); got != (config.ModelEffort{Model: glmAuditDefaultModel, Effort: glmAuditDefaultEffort}) {
		t.Errorf("glm audit without pin = %+v, want the {%s, %s} default pin (REQ-TIER-004 — the former empty-effort backend default fell with the operator pin flip)", got, glmAuditDefaultModel, glmAuditDefaultEffort)
	}
	if got := resolveGLMTaskModel(); got != config.DefaultGLMHigh {
		t.Errorf("glm task default = %q, want %q", got, config.DefaultGLMHigh)
	}
}
