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
		SessionBackendGLM:   {SessionRoleLeader: {"opus", "high"}, SessionRoleLane: {"sonnet", "medium"}},
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
		"glm effort max":          {func(s *SessionDefaults) { s.GLM.Leader.Effort = "max" }, "llm.defaults.sessions.glm.leader.effort"},
		"glm effort Claude vocab": {func(s *SessionDefaults) { s.GLM.Lane.Effort = "medium" }, ""},
		"cc xhigh":                {func(s *SessionDefaults) { s.CC.Leader.Effort = "xhigh" }, ""},
		"codex model newline":     {func(s *SessionDefaults) { s.Codex.Lane.Model = "gpt\nx" }, "llm.defaults.sessions.codex.lane.model"},
		"codex model id":          {func(s *SessionDefaults) { s.Codex.Lane.Model = "gpt-6.1-sol" }, ""},
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
// package ships carries max or ultra — every session default (glm included:
// llm.defaults.sessions stores Claude vocabulary) and the Claude/Codex audit
// pins — and pins the exemption explicitly: the GLM reasoning-state
// vocabulary (z.ai low|high|max), where max is the depth z.ai defaults to.
// Exactly these GLM cells may carry max; the test fails if the exempt set
// grows or a non-exempt default reaches max/ultra.
func TestNoShippedDefaultEffortIsMaxOrUltra(t *testing.T) {
	t.Parallel()
	llm := NewDefaultLLMConfig()
	audit := NewDefaultWorkflowConfig().Audit

	shipped := map[string]string{
		"claude audit pin effort": audit.Claude.Effort,
		"codex audit pin effort":  audit.Codex.Effort,
	}
	for _, backend := range []string{SessionBackendCC, SessionBackendGLM, SessionBackendCodex} {
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

	exempt := map[string]string{
		"glm tier effort high": DefaultGLMEffortHigh,
		"glm tier effort low":  DefaultGLMEffortLow,
		"glm audit pin effort": DefaultGLMAuditEffort,
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
	for name, v := range map[string]string{
		"glm tier effort medium": DefaultGLMEffortMedium,
		"glm tier effort fable":  DefaultGLMEffortFable,
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

// TestSessionDefaultsResolveRejectsInvalid: a custom cell that
// validateSessionDefaults refuses never leaves the resolver — the shipped
// value for that field is returned instead (fail-open for the launchers).
func TestSessionDefaultsResolveRejectsInvalid(t *testing.T) {
	t.Parallel()
	llm := NewDefaultLLMConfig()
	llm.Defaults.Sessions.Mode = SessionDefaultsModeCustom
	llm.Defaults.Sessions.CC.Lane = ModelEffort{Model: "opus; x", Effort: "ultra"}
	llm.Defaults.Sessions.Codex.Lane = ModelEffort{Model: "gpt\nx", Effort: `high" model="evil`}
	llm.Defaults.Sessions.GLM.Leader = ModelEffort{Model: "opus", Effort: "max"}
	for _, tc := range []struct {
		backend, role string
		want          ModelEffort
	}{
		{SessionBackendCC, SessionRoleLane, ModelEffort{"sonnet", "medium"}},
		{SessionBackendCodex, SessionRoleLane, ModelEffort{CodexModelSol, "medium"}},
		{SessionBackendGLM, SessionRoleLeader, ModelEffort{"opus", "high"}},
	} {
		if got, ok := llm.ResolveSessionDefault(tc.backend, tc.role); !ok || got != tc.want {
			t.Errorf("%s %s = %+v (ok %v), want the shipped %+v", tc.backend, tc.role, got, ok, tc.want)
		}
	}
}

// TestSessionDefaultsModelSurroundingWhitespace: a session model carrying
// leading or trailing whitespace or control characters is normalized, not
// rejected — config.Validate accepts it and the resolver uses the trimmed
// value while reporting exactly one entry naming the key and the raw value,
// so the launcher can warn. Interior control characters stay rejected on both
// paths with the shipped fallback.
func TestSessionDefaultsModelSurroundingWhitespace(t *testing.T) {
	t.Parallel()
	const key = "llm.defaults.sessions.codex.lane.model"
	for name, tc := range map[string]struct {
		raw, want string
		accepted  bool
	}{
		"surrounding newline and tab": {"\ngpt-custom\t", "gpt-custom", true},
		"surrounding control char":    {"\x01gpt-custom ", "gpt-custom", true},
		"interior newline":            {" gpt\nx ", CodexModelSol, false},
	} {
		cfg := NewDefaultConfig()
		cfg.LLM.Defaults.Sessions.Mode = SessionDefaultsModeCustom
		cfg.LLM.Defaults.Sessions.Codex.Lane.Model = tc.raw
		errs := validateSessionDefaults(cfg)
		if tc.accepted && len(errs) != 0 {
			t.Errorf("%s: Validate = %+v, want accepted", name, errs)
		}
		if !tc.accepted && (len(errs) != 1 || errs[0].Field != key) {
			t.Errorf("%s: Validate = %+v, want one error on %s", name, errs, key)
		}
		got, notes, ok := cfg.LLM.ResolveSessionDefaultChecked(SessionBackendCodex, SessionRoleLane)
		if !ok || got.Model != tc.want {
			t.Errorf("%s: resolved model = %q (ok %v), want %q", name, got.Model, ok, tc.want)
		}
		if len(notes) != 1 || notes[0].Key != key || notes[0].Value != tc.raw || notes[0].Fallback != tc.want {
			t.Errorf("%s: resolver notes = %+v, want one entry {Key %s, Value %q, Fallback %q}", name, notes, key, tc.raw, tc.want)
		}
	}
}
