package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The launch defaults for factory sessions (card t1808, operator per-role
// defaults): the leader (-f) and lane (-l) of each launcher start from a
// model + effort pair that loads from config and stays changeable mid-session.

func TestSessionDefaultsShippedTable(t *testing.T) {
	t.Parallel()
	llm := NewDefaultLLMConfig()
	want := map[string]map[string]ModelEffort{
		SessionBackendCC:    {SessionRoleLeader: {"opus", "high"}, SessionRoleLane: {"sonnet", "medium"}},
		SessionBackendGLM:   {SessionRoleLeader: {"opus", "max"}, SessionRoleLane: {"sonnet", "high"}},
		SessionBackendCodex: {SessionRoleLeader: {CodexModelSol, "high"}, SessionRoleLane: {CodexModelSol, "medium"}},
	}
	for backend, roles := range want {
		for role, w := range roles {
			got, ok := llm.ResolveSessionDefault(backend, role)
			if !ok || got != w {
				t.Errorf("ResolveSessionDefault(%s, %s) = %+v (ok %v), want %+v", backend, role, got, ok, w)
			}
		}
	}
	if _, ok := llm.ResolveSessionDefault("cc", "worker"); ok {
		t.Error("an unknown role must not resolve")
	}
	if _, ok := llm.ResolveSessionDefault("gpt", SessionRoleLeader); ok {
		t.Error("an unknown backend must not resolve")
	}
}

func TestSessionDefaultsModeSelectsSource(t *testing.T) {
	t.Parallel()
	llm := NewDefaultLLMConfig()
	llm.Defaults.Sessions.CC.Lane = ModelEffort{Model: "opus", Effort: "low"}
	llm.Defaults.Sessions.Codex.Leader = ModelEffort{Effort: "xhigh"} // model left empty

	// default mode: the shipped table, whatever the file carries.
	llm.Defaults.Sessions.Mode = SessionDefaultsModeDefault
	if got, _ := llm.ResolveSessionDefault(SessionBackendCC, SessionRoleLane); got != (ModelEffort{"sonnet", "medium"}) {
		t.Errorf("default mode cc lane = %+v, want the shipped sonnet/medium", got)
	}
	// custom mode: the file's values, an empty field falling back per field.
	llm.Defaults.Sessions.Mode = SessionDefaultsModeCustom
	if got, _ := llm.ResolveSessionDefault(SessionBackendCC, SessionRoleLane); got != (ModelEffort{"opus", "low"}) {
		t.Errorf("custom mode cc lane = %+v, want opus/low from llm.yaml", got)
	}
	if got, _ := llm.ResolveSessionDefault(SessionBackendCodex, SessionRoleLeader); got != (ModelEffort{CodexModelSol, "xhigh"}) {
		t.Errorf("custom mode codex leader = %+v, want the shipped model with the custom effort", got)
	}
	// an empty mode reads as default.
	llm.Defaults.Sessions.Mode = ""
	if got, _ := llm.ResolveSessionDefault(SessionBackendCC, SessionRoleLane); got != (ModelEffort{"sonnet", "medium"}) {
		t.Errorf("empty mode cc lane = %+v, want the shipped default", got)
	}
}

func TestSessionDefaultsValidation(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		mutate func(*SessionDefaults)
		field  string // "" = valid
	}{
		"shipped defaults":        {func(*SessionDefaults) {}, ""},
		"unknown mode":            {func(s *SessionDefaults) { s.Mode = "auto" }, "llm.defaults.sessions.mode"},
		"cc lane effort max":      {func(s *SessionDefaults) { s.CC.Lane.Effort = "max" }, "llm.defaults.sessions.cc.lane.effort"},
		"codex leader ultra":      {func(s *SessionDefaults) { s.Codex.Leader.Effort = "ultra" }, "llm.defaults.sessions.codex.leader.effort"},
		"cc model off the set":    {func(s *SessionDefaults) { s.CC.Leader.Model = "gpt-x" }, "llm.defaults.sessions.cc.leader.model"},
		"glm effort off the set":  {func(s *SessionDefaults) { s.GLM.Lane.Effort = "medium" }, "llm.defaults.sessions.glm.lane.effort"},
		"glm leader max (exempt)": {func(s *SessionDefaults) { s.GLM.Leader.Effort = "max" }, ""},
		"cc xhigh":                {func(s *SessionDefaults) { s.CC.Leader.Effort = "xhigh" }, ""},
	}
	for name, tc := range cases {
		cfg := NewDefaultConfig()
		cfg.LLM.Defaults.Sessions.Mode = SessionDefaultsModeCustom
		tc.mutate(&cfg.LLM.Defaults.Sessions)
		errs := validateSessionDefaults(cfg)
		switch {
		case tc.field == "" && len(errs) != 0:
			t.Errorf("%s: want valid, got %+v", name, errs)
		case tc.field != "" && (len(errs) != 1 || errs[0].Field != tc.field):
			t.Errorf("%s: want one error on %s, got %+v", name, tc.field, errs)
		}
	}
}

// TestNoShippedDefaultEffortIsMaxOrUltra proves no default effort the config
// package ships carries max or ultra, and pins the exemption explicitly: the
// GLM surfaces speak the z.ai reasoning-state vocabulary, where max is the
// reasoning depth z.ai defaults to, and the `moai glm -f` launch maps to it.
// A GLM value is a launch/reasoning mapping, not a Claude or Codex effort
// default, so exactly these cells may carry max — and the test fails if the
// exempt set grows or a non-exempt default reaches max/ultra.
func TestNoShippedDefaultEffortIsMaxOrUltra(t *testing.T) {
	t.Parallel()
	llm := NewDefaultLLMConfig()
	audit := NewDefaultWorkflowConfig().Audit

	shipped := map[string]string{
		"claude audit pin effort": audit.Claude.Effort,
		"codex audit pin effort":  audit.Codex.Effort,
	}
	for _, backend := range []string{SessionBackendCC, SessionBackendCodex} {
		for _, role := range []string{SessionRoleLeader, SessionRoleLane} {
			me, _ := llm.ResolveSessionDefault(backend, role)
			shipped["session "+backend+" "+role] = me.Effort
		}
	}
	for name, v := range shipped {
		if v == "max" || v == "ultra" {
			t.Errorf("shipped default %s = %q; max/ultra are never a shipped default", name, v)
		}
	}

	glmLeader, _ := llm.ResolveSessionDefault(SessionBackendGLM, SessionRoleLeader)
	exempt := map[string]string{
		"session glm leader (launch mapping)": glmLeader.Effort,
		"glm tier effort high":                DefaultGLMEffortHigh,
		"glm tier effort fable":               DefaultGLMEffortFable,
		"glm audit pin effort":                DefaultGLMAuditEffort,
	}
	var maxCells []string
	for name, v := range exempt {
		if v == "max" {
			maxCells = append(maxCells, name)
		}
	}
	sort.Strings(maxCells)
	if len(maxCells) != len(exempt) {
		t.Errorf("exempt GLM cells carrying max = %v, want all %d (the exemption is scoped to exactly these)", maxCells, len(exempt))
	}
	glmLane, _ := llm.ResolveSessionDefault(SessionBackendGLM, SessionRoleLane)
	for name, v := range map[string]string{
		"session glm lane":       glmLane.Effort,
		"glm tier effort medium": DefaultGLMEffortMedium,
		"glm tier effort low":    DefaultGLMEffortLow,
	} {
		if v == "max" || v == "ultra" {
			t.Errorf("%s = %q is outside the max exemption", name, v)
		}
	}
}

// TestSessionDefaultsTemplateSymmetry pins the distributed llm.yaml
// llm.defaults.sessions block to the Go struct (keys both ways) and to the
// shipped default values.
func TestSessionDefaultsTemplateSymmetry(t *testing.T) {
	t.Parallel()
	_, file, _, _ := runtime.Caller(0)
	tmpl := filepath.Join(filepath.Dir(file), "..", "template", "templates", ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(tmpl)
	if err != nil {
		t.Fatalf("read template llm.yaml: %v", err)
	}
	var raw struct {
		LLM struct {
			Defaults struct {
				Sessions map[string]any `yaml:"sessions"`
			} `yaml:"defaults"`
		} `yaml:"llm"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse template llm.yaml: %v", err)
	}
	gotKeys := sessionSortedKeys(raw.LLM.Defaults.Sessions)
	wantKeys := sessionYAMLTags(reflect.TypeOf(SessionDefaults{}))
	if strings.Join(gotKeys, ",") != strings.Join(wantKeys, ",") {
		t.Errorf("template llm.defaults.sessions keys = %v, struct tags = %v", gotKeys, wantKeys)
	}

	var typed struct {
		LLM struct {
			Defaults LLMDefaults `yaml:"defaults"`
		} `yaml:"llm"`
	}
	if err := yaml.Unmarshal(data, &typed); err != nil {
		t.Fatalf("typed parse: %v", err)
	}
	tmplLLM := NewDefaultLLMConfig()
	tmplLLM.Defaults = typed.LLM.Defaults
	if tmplLLM.Defaults.Sessions.Mode != SessionDefaultsModeDefault {
		t.Errorf("template mode = %q, want %q", tmplLLM.Defaults.Sessions.Mode, SessionDefaultsModeDefault)
	}
	// The template cells restate the shipped table, so switching the mode to
	// custom without editing them changes nothing.
	tmplLLM.Defaults.Sessions.Mode = SessionDefaultsModeCustom
	def := NewDefaultLLMConfig()
	for _, backend := range []string{SessionBackendCC, SessionBackendGLM, SessionBackendCodex} {
		for _, role := range []string{SessionRoleLeader, SessionRoleLane} {
			got, _ := tmplLLM.ResolveSessionDefault(backend, role)
			want, _ := def.ResolveSessionDefault(backend, role)
			if got != want {
				t.Errorf("template %s.%s = %+v, want the shipped %+v", backend, role, got, want)
			}
		}
	}
}

func sessionSortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sessionYAMLTags(t reflect.Type) []string {
	var out []string
	for i := 0; i < t.NumField(); i++ {
		tag := strings.Split(t.Field(i).Tag.Get("yaml"), ",")[0]
		if tag != "" && tag != "-" {
			out = append(out, tag)
		}
	}
	sort.Strings(out)
	return out
}
