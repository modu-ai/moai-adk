package config

// audit_plan_test.go — tests for the audit plan resolver
// (SPEC-AUDIT-MODEL-CONVERGE-001 M2; AC-ACV-001, -002, -003 resolver part,
// -004). Every table sweep carries a positive control that proves it ran: a sweep
// over an empty table would pass vacuously.
//
// @MX:SPEC: SPEC-AUDIT-MODEL-CONVERGE-001

import (
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// auditPlanTokens are the five inputs of REQ-ACV-001: the empty token and the
// closed set. The empty token is "key absent", a distinct input from `claude`.
var auditPlanTokens = []string{"", AuditModelClaude, AuditModelCodex, AuditModelGLM, AuditModelMulti}

// auditPlanBackendOrder is the stable order of AuditPlan.Backends.
var auditPlanBackendOrder = []string{"claude", "codex", "glm"}

// auditPlanWantRow is one row of design.md §D.1: the gate the token assigns to
// each backend, in claude / codex / glm order, and whether that cell is the
// operator's choice (source config.model) or the distributed default.
type auditPlanWantRow struct {
	gates    [3]string
	explicit [3]bool
}

// auditPlanTable is design.md §D.1 written out independently of the resolver.
var auditPlanTable = map[string]auditPlanWantRow{
	"": {
		gates:    [3]string{AuditGateRequired, AuditGateRequired, AuditGateAdvisory},
		explicit: [3]bool{false, false, false},
	},
	AuditModelClaude: { // decision D7': Claude explicit, codex and glm left at the default
		gates:    [3]string{AuditGateRequired, AuditGateRequired, AuditGateAdvisory},
		explicit: [3]bool{true, false, false},
	},
	AuditModelCodex: {
		gates:    [3]string{AuditGateOff, AuditGateRequired, AuditGateOff},
		explicit: [3]bool{true, true, true},
	},
	AuditModelGLM: {
		gates:    [3]string{AuditGateOff, AuditGateOff, AuditGateRequired},
		explicit: [3]bool{true, true, true},
	},
	AuditModelMulti: {
		gates:    [3]string{AuditGateRequired, AuditGateRequired, AuditGateAdvisory},
		explicit: [3]bool{true, true, true},
	},
}

func tokenLabel(token string) string {
	if token == "" {
		return "empty"
	}
	return token
}

func setGate(g *AuditGates, i int, v string) {
	switch i {
	case 0:
		g.Claude = v
	case 1:
		g.Codex = v
	case 2:
		g.GLM = v
	}
}

// assertBackendOrder checks the plan names claude, codex, glm exactly once each,
// in that order, and returns false when it does not.
func assertBackendOrder(t *testing.T, plan AuditPlan) bool {
	t.Helper()
	if len(plan.Backends) != len(auditPlanBackendOrder) {
		t.Errorf("plan has %d backend entries, want %d: %+v", len(plan.Backends), len(auditPlanBackendOrder), plan.Backends)
		return false
	}
	for i, want := range auditPlanBackendOrder {
		if plan.Backends[i].Backend != want {
			t.Errorf("Backends[%d].Backend = %q, want %q", i, plan.Backends[i].Backend, want)
			return false
		}
	}
	return true
}

// TestResolveAuditPlan_TokenTable pins every row of design.md §D.1 (AC-ACV-001).
func TestResolveAuditPlan_TokenTable(t *testing.T) {
	ran := 0
	for _, token := range auditPlanTokens {
		t.Run(tokenLabel(token), func(t *testing.T) {
			ran++
			plan, err := ResolveAuditPlan(AuditConfig{Model: token}, AuditGates{})
			if err != nil {
				t.Fatalf("ResolveAuditPlan(%q): unexpected error: %v", token, err)
			}
			if !assertBackendOrder(t, plan) {
				return
			}
			want := auditPlanTable[token]
			for i, e := range plan.Backends {
				if e.Gate != want.gates[i] {
					t.Errorf("%s gate = %q, want %q", e.Backend, e.Gate, want.gates[i])
				}
				if e.Explicit != want.explicit[i] {
					t.Errorf("%s explicit = %v, want %v", e.Backend, e.Explicit, want.explicit[i])
				}
				wantSource := AuditPlanSourceDefault
				if want.explicit[i] {
					wantSource = AuditPlanSourceConfigModel
				}
				if e.Source != wantSource {
					t.Errorf("%s source = %q, want %q", e.Backend, e.Source, wantSource)
				}
				if !isClosedAuditGate(e.Gate) {
					t.Errorf("%s gate %q is outside off|advisory|required", e.Backend, e.Gate)
				}
			}
			wantModelSource := "config"
			if token == "" {
				wantModelSource = "default"
			}
			if plan.Model != token || plan.ModelSource() != wantModelSource {
				t.Errorf("Model/ModelSource = %q/%q, want %q/%q", plan.Model, plan.ModelSource(), token, wantModelSource)
			}
		})
	}
	if ran != len(auditPlanTokens) || len(auditPlanTokens) != 5 {
		t.Fatalf("positive control: swept %d of %d tokens (want the 5 of REQ-ACV-001)", ran, len(auditPlanTokens))
	}
}

func isClosedAuditGate(g string) bool {
	for _, v := range ValidAuditGates() {
		if g == v {
			return true
		}
	}
	return false
}

// TestResolveAuditPlan_Precedence sweeps every token x every backend x every
// combination of {caller gate present/absent, configured gate present/absent}
// over all three gate values (AC-ACV-002): the ladder is argument >
// config.gates > config.model > default, and the deciding source is recorded.
func TestResolveAuditPlan_Precedence(t *testing.T) {
	gateValues := ValidAuditGates()
	choices := append([]string{""}, gateValues...) // "" = not supplied
	cases := 0
	for _, token := range auditPlanTokens {
		for bi := range auditPlanBackendOrder {
			for _, argVal := range choices {
				for _, cfgVal := range choices {
					cases++
					var args, cfgGates AuditGates
					setGate(&args, bi, argVal)
					setGate(&cfgGates, bi, cfgVal)
					plan, err := ResolveAuditPlan(AuditConfig{Model: token, Gates: cfgGates}, args)
					if err != nil {
						t.Fatalf("token=%q backend=%d arg=%q cfg=%q: %v", token, bi, argVal, cfgVal, err)
					}
					if !assertBackendOrder(t, plan) {
						return
					}
					got := plan.Backends[bi]

					wantGate := auditPlanTable[token].gates[bi]
					wantSource := AuditPlanSourceDefault
					wantExplicit := false
					if auditPlanTable[token].explicit[bi] {
						wantSource, wantExplicit = AuditPlanSourceConfigModel, true
					}
					if cfgVal != "" {
						wantGate, wantSource, wantExplicit = cfgVal, AuditPlanSourceConfigGates, true
					}
					if argVal != "" {
						// A caller-supplied gate wins and is NOT explicit (design.md §D.2).
						wantGate, wantSource, wantExplicit = argVal, AuditPlanSourceArgument, false
					}
					if got.Gate != wantGate || got.Source != wantSource || got.Explicit != wantExplicit {
						t.Errorf("token=%q backend=%s arg=%q cfg=%q: got {gate:%s source:%s explicit:%v}, want {gate:%s source:%s explicit:%v}",
							token, got.Backend, argVal, cfgVal, got.Gate, got.Source, got.Explicit, wantGate, wantSource, wantExplicit)
					}
					// A gate set for one backend must leave the other two on their token cell.
					for oi := range auditPlanBackendOrder {
						if oi == bi {
							continue
						}
						if other := plan.Backends[oi]; other.Gate != auditPlanTable[token].gates[oi] {
							t.Errorf("token=%q: setting backend %d moved backend %s to %q", token, bi, other.Backend, other.Gate)
						}
					}
				}
			}
		}
	}
	// 5 tokens x 3 backends x 4 argument choices x 4 config choices.
	if want := 5 * 3 * 4 * 4; cases != want {
		t.Fatalf("positive control: swept %d cases, want %d", cases, want)
	}
}

// TestResolveAuditPlan_PrecedenceExample is the worked example of AC-ACV-002:
// multi + audit.gates.glm: off + caller gate codex: advisory.
func TestResolveAuditPlan_PrecedenceExample(t *testing.T) {
	audit := AuditConfig{Model: AuditModelMulti, Gates: AuditGates{GLM: AuditGateOff}}
	plan, err := ResolveAuditPlan(audit, AuditGates{Codex: AuditGateAdvisory})
	if err != nil {
		t.Fatalf("ResolveAuditPlan: %v", err)
	}
	want := []AuditPlanEntry{
		{Backend: "claude", Gate: AuditGateRequired, Source: AuditPlanSourceConfigModel, Explicit: true},
		{Backend: "codex", Gate: AuditGateAdvisory, Source: AuditPlanSourceArgument, Explicit: false},
		{Backend: "glm", Gate: AuditGateOff, Source: AuditPlanSourceConfigGates, Explicit: true},
	}
	if !reflect.DeepEqual(plan.Backends, want) {
		t.Fatalf("with the caller gate: got %+v, want %+v", plan.Backends, want)
	}

	plan, err = ResolveAuditPlan(audit, AuditGates{})
	if err != nil {
		t.Fatalf("ResolveAuditPlan (no caller gate): %v", err)
	}
	if !assertBackendOrder(t, plan) {
		return
	}
	if got := plan.Backends[1]; got.Gate != AuditGateRequired || got.Source != AuditPlanSourceConfigModel {
		t.Errorf("without the caller gate codex = %+v, want required from config.model", got)
	}

	plan, err = ResolveAuditPlan(AuditConfig{}, AuditGates{})
	if err != nil {
		t.Fatalf("ResolveAuditPlan (empty): %v", err)
	}
	if !assertBackendOrder(t, plan) {
		return
	}
	for _, e := range plan.Backends {
		if e.Source != AuditPlanSourceDefault || e.Explicit {
			t.Errorf("empty token, no gates: %s = %+v, want source default and not explicit", e.Backend, e)
		}
	}
}

// TestResolveAuditPlan_ExplicitGates checks the accessor REQ-ACV-005 names: only
// entries that are explicit appear, every other field stays empty (the "not
// configured" reading the fail-closed readers key on).
func TestResolveAuditPlan_ExplicitGates(t *testing.T) {
	cases := []struct {
		name  string
		audit AuditConfig
		args  AuditGates
		want  AuditGates
	}{
		{"empty token", AuditConfig{}, AuditGates{}, AuditGates{}},
		{"pins only", AuditConfig{Codex: ModelEffort{Model: "x", Effort: "high"}}, AuditGates{}, AuditGates{}},
		{"claude token", AuditConfig{Model: AuditModelClaude}, AuditGates{}, AuditGates{Claude: AuditGateRequired}},
		{"codex token", AuditConfig{Model: AuditModelCodex}, AuditGates{}, AuditGates{Claude: AuditGateOff, Codex: AuditGateRequired, GLM: AuditGateOff}},
		{"glm token", AuditConfig{Model: AuditModelGLM}, AuditGates{}, AuditGates{Claude: AuditGateOff, Codex: AuditGateOff, GLM: AuditGateRequired}},
		{"multi token", AuditConfig{Model: AuditModelMulti}, AuditGates{}, AuditGates{Claude: AuditGateRequired, Codex: AuditGateRequired, GLM: AuditGateAdvisory}},
		{"gates only", AuditConfig{Gates: AuditGates{Codex: AuditGateRequired}}, AuditGates{}, AuditGates{Codex: AuditGateRequired}},
		{"explicit off is explicit", AuditConfig{Model: AuditModelMulti, Gates: AuditGates{Codex: AuditGateOff}}, AuditGates{},
			AuditGates{Claude: AuditGateRequired, Codex: AuditGateOff, GLM: AuditGateAdvisory}},
		{"argument is not explicit", AuditConfig{}, AuditGates{Codex: AuditGateRequired}, AuditGates{}},
		{"argument over multi drops that entry", AuditConfig{Model: AuditModelMulti}, AuditGates{Codex: AuditGateAdvisory},
			AuditGates{Claude: AuditGateRequired, GLM: AuditGateAdvisory}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := ResolveAuditPlan(tc.audit, tc.args)
			if err != nil {
				t.Fatalf("ResolveAuditPlan: %v", err)
			}
			if got := plan.ExplicitGates(); got != tc.want {
				t.Errorf("ExplicitGates() = %+v, want %+v", got, tc.want)
			}
		})
	}
	if len(cases) != 10 {
		t.Fatalf("positive control: %d cases, want 10", len(cases))
	}
}

// TestResolveAuditPlan_FromConfig covers the plan_source predicate: true exactly
// when some backend's gate came from the tree's configuration.
func TestResolveAuditPlan_FromConfig(t *testing.T) {
	cases := []struct {
		name  string
		audit AuditConfig
		args  AuditGates
		want  bool
	}{
		{"nothing configured", AuditConfig{}, AuditGates{}, false},
		{"arguments only", AuditConfig{}, AuditGates{Codex: AuditGateRequired}, false},
		{"model token", AuditConfig{Model: AuditModelMulti}, AuditGates{}, true},
		{"gates block", AuditConfig{Gates: AuditGates{GLM: AuditGateOff}}, AuditGates{}, true},
		{"config fully overridden by arguments", AuditConfig{Model: AuditModelClaude}, AuditGates{Claude: AuditGateOff}, false},
	}
	for _, tc := range cases {
		plan, err := ResolveAuditPlan(tc.audit, tc.args)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := plan.FromConfig(); got != tc.want {
			t.Errorf("%s: FromConfig() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestResolveAuditPlan_RejectsUnknown pins REQ-ACV-003 and AC-ACV-003's resolver
// part: an invalid token or gate returns an error naming the key and value and
// listing the accepted values, and never a default plan.
func TestResolveAuditPlan_RejectsUnknown(t *testing.T) {
	cases := []struct {
		name    string
		audit   AuditConfig
		args    AuditGates
		wantErr string
	}{
		{"unknown token", AuditConfig{Model: "grok"}, AuditGates{},
			`audit_model "grok" unknown (want one of claude|codex|glm|multi)`},
		{"token is case sensitive", AuditConfig{Model: "Multi"}, AuditGates{},
			`audit_model "Multi" unknown (want one of claude|codex|glm|multi)`},
		{"invalid configured claude gate", AuditConfig{Gates: AuditGates{Claude: "Required"}}, AuditGates{},
			`audit.gates.claude "Required" unknown (want one of off|advisory|required)`},
		{"invalid configured codex gate", AuditConfig{Model: AuditModelMulti, Gates: AuditGates{Codex: "requird"}}, AuditGates{},
			`audit.gates.codex "requird" unknown (want one of off|advisory|required)`},
		{"invalid configured glm gate", AuditConfig{Gates: AuditGates{GLM: "on"}}, AuditGates{},
			`audit.gates.glm "on" unknown (want one of off|advisory|required)`},
		{"invalid supplied gate", AuditConfig{}, AuditGates{Codex: "requird"},
			`gates.codex "requird" unknown (want one of off|advisory|required)`},
		{"token error wins over gate error", AuditConfig{Model: "grok", Gates: AuditGates{Codex: "requird"}}, AuditGates{},
			`audit_model "grok" unknown (want one of claude|codex|glm|multi)`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := ResolveAuditPlan(tc.audit, tc.args)
			if err == nil {
				t.Fatalf("ResolveAuditPlan returned no error, plan %+v", plan)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tc.wantErr)
			}
			if len(plan.Backends) != 0 || plan.Model != "" {
				t.Errorf("an error must not carry a plan (a default would hide it): %+v", plan)
			}
		})
	}
	if len(cases) != 7 {
		t.Fatalf("positive control: %d cases, want 7", len(cases))
	}
}

// TestResolveAuditPlan_TrimsWhitespace: surrounding whitespace on a token or a
// gate is trimmed and a whitespace-only value reads as unset (AC-ACV-003).
func TestResolveAuditPlan_TrimsWhitespace(t *testing.T) {
	plan, err := ResolveAuditPlan(AuditConfig{Model: " multi ", Gates: AuditGates{Codex: " off "}}, AuditGates{GLM: "\tadvisory\n"})
	if err != nil {
		t.Fatalf("ResolveAuditPlan: %v", err)
	}
	if !assertBackendOrder(t, plan) {
		return
	}
	if plan.Model != AuditModelMulti {
		t.Errorf("Model = %q, want %q (trimmed)", plan.Model, AuditModelMulti)
	}
	if got := plan.Backends[1]; got.Gate != AuditGateOff || got.Source != AuditPlanSourceConfigGates {
		t.Errorf("codex = %+v, want off from config.gates", got)
	}
	if got := plan.Backends[2]; got.Gate != AuditGateAdvisory || got.Source != AuditPlanSourceArgument {
		t.Errorf("glm = %+v, want advisory from argument", got)
	}

	blank, err := ResolveAuditPlan(AuditConfig{Model: "  ", Gates: AuditGates{Codex: " "}}, AuditGates{GLM: "\t"})
	if err != nil {
		t.Fatalf("ResolveAuditPlan (blank): %v", err)
	}
	if !assertBackendOrder(t, blank) {
		return
	}
	for _, e := range blank.Backends {
		if e.Source != AuditPlanSourceDefault {
			t.Errorf("blank values: %s source = %q, want default", e.Backend, e.Source)
		}
	}
	if blank.Model != "" {
		t.Errorf("blank token: Model = %q, want empty", blank.Model)
	}
}

// TestResolveAuditPlan_SourceIsPure is the mechanical form of AC-ACV-004's first
// two clauses: the resolver file is in package config, imports no file, clock,
// network or process package, and names no default-merged constructor or
// internal/cli.
func TestResolveAuditPlan_SourceIsPure(t *testing.T) {
	const path = "audit_plan.go"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(raw)
	// Positive control: the file is the resolver, not an empty or unrelated file.
	if !strings.Contains(src, "func ResolveAuditPlan(") || !strings.HasPrefix(src, "package config") {
		t.Fatalf("%s is not the resolver source (no ResolveAuditPlan / package config)", path)
	}

	file, err := parser.ParseFile(token.NewFileSet(), path, raw, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(file.Imports) == 0 {
		t.Fatal("positive control: the import scan saw no imports")
	}
	forbidden := map[string]bool{"os": true, "time": true, "net": true, "io": true, "path/filepath": true, "os/exec": true}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			t.Fatalf("unquote import %s: %v", imp.Path.Value, err)
		}
		if forbidden[p] {
			t.Errorf("%s imports %q: the resolver is a pure function (REQ-ACV-004)", path, p)
		}
		if strings.Contains(p, "internal/cli") {
			t.Errorf("%s imports %q: internal/config cannot depend on internal/cli", path, p)
		}
	}
	for _, banned := range []string{"NewDefaultWorkflowConfig", "NewDefaultConfig", "internal/cli"} {
		if strings.Contains(src, banned) {
			t.Errorf("%s mentions %q: the resolver must never be fed from a default-merged configuration", path, banned)
		}
	}
}
