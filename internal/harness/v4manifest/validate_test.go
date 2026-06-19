package v4manifest

import (
	"strings"
	"testing"
)

// validManifest is the canonical valid manifest fixture used as the baseline
// for all validation tests. Every field is set to a valid value per design §C.
// Invalid-variant tests clone this fixture and mutate exactly one field, then
// assert Validate rejects the variant (AC-HV4-006a).
func validManifest() Manifest {
	return Manifest{
		Name:           "moai-adk-dev",
		Domain:         "moai-adk CLI template development",
		SourceRequest:  "build a harness for moai-adk CLI template development",
		Patterns:       []string{PatternPipeline, PatternProducerReviewer},
		EntryCommand:   "/harness:moai-adk-dev",
		RunnerWorkflow: "harness-moai-adk-dev-run.js",
		Specialists: []Specialist{
			{
				Role:      "template-neutrality-auditor",
				Primitive: PrimitiveSubAgent,
				Isolation: IsolationNone,
				Effort:    EffortXhigh,
				Model:     ModelInherit,
			},
		},
		SprintContract: SprintContract{
			Dimensions: []string{"neutrality", "coverage"},
			Thresholds: map[string]interface{}{
				"neutrality": 0,
				"coverage":   0.85,
			},
		},
	}
}

// TestValidate_AcceptsValidManifest verifies the canonical valid fixture
// passes validation (AC-HV4-006a happy path).
func TestValidate_AcceptsValidManifest(t *testing.T) {
	m := validManifest()
	if err := Validate(m); err != nil {
		t.Fatalf("Validate on valid manifest returned error: %v", err)
	}
}

// TestValidate_RejectsEachMissingTopLevelField verifies each of the 8
// top-level fields is required (AC-HV4-006a). Each sub-test mutates exactly
// one field to empty and asserts rejection.
func TestValidate_RejectsEachMissingTopLevelField(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Manifest)
		field  string // substring expected in the error message
	}{
		{"missing_name", func(m *Manifest) { m.Name = "" }, "name"},
		{"missing_domain", func(m *Manifest) { m.Domain = "" }, "domain"},
		{"missing_source_request", func(m *Manifest) { m.SourceRequest = "" }, "source_request"},
		{"missing_patterns", func(m *Manifest) { m.Patterns = nil }, "patterns"},
		{"missing_entry_command", func(m *Manifest) { m.EntryCommand = "" }, "entry_command"},
		{"missing_runner_workflow", func(m *Manifest) { m.RunnerWorkflow = "" }, "runner_workflow"},
		{"missing_specialists", func(m *Manifest) { m.Specialists = nil }, "specialists"},
		{"missing_sprint_dimensions", func(m *Manifest) { m.SprintContract.Dimensions = nil }, "dimensions"},
		{"nil_sprint_thresholds", func(m *Manifest) { m.SprintContract.Thresholds = nil }, "thresholds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validManifest()
			tc.mutate(&m)
			err := Validate(m)
			if err == nil {
				t.Fatalf("Validate accepted manifest with %s (expected rejection)", tc.name)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("Validate error %q does not mention field %q", err.Error(), tc.field)
			}
		})
	}
}

// TestValidate_RejectsInvalidPrimitive verifies each specialist.primitive
// MUST be exactly one of the 5 primitives (AC-HV4-005a + design §C.2).
func TestValidate_RejectsInvalidPrimitive(t *testing.T) {
	cases := []struct {
		name      string
		primitive string
		wantValid bool
	}{
		{"sub_agent", PrimitiveSubAgent, true},
		{"dynamic_workflow", PrimitiveDynamicWorkflow, true},
		{"worktree", PrimitiveWorktree, true},
		{"goal", PrimitiveGoal, true},
		{"adversarial_fan_out", PrimitiveAdversarialFanOut, true},
		{"free_text", "some-invented-primitive", false},
		{"empty", "", false},
		{"agent_lowercase_typo", "sub_agent", false}, // must be "sub-agent" (hyphen)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validManifest()
			m.Specialists[0].Primitive = tc.primitive
			err := Validate(m)
			if tc.wantValid && err != nil {
				t.Fatalf("Validate rejected valid primitive %q: %v", tc.primitive, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("Validate accepted invalid primitive %q (expected rejection)", tc.primitive)
			}
		})
	}
}

// TestValidate_RejectsInvalidIsolation verifies specialist.isolation MUST be
// none|worktree (AC-HV4-005a + design §C.2).
func TestValidate_RejectsInvalidIsolation(t *testing.T) {
	cases := []struct {
		isolation string
		wantValid bool
	}{
		{IsolationNone, true},
		{IsolationWorktree, true},
		{"container", false},
		{"", false},
		{"worktrees", false}, // plural typo
	}
	for _, tc := range cases {
		t.Run(tc.isolation, func(t *testing.T) {
			m := validManifest()
			m.Specialists[0].Isolation = tc.isolation
			err := Validate(m)
			if tc.wantValid && err != nil {
				t.Fatalf("Validate rejected valid isolation %q: %v", tc.isolation, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("Validate accepted invalid isolation %q", tc.isolation)
			}
		})
	}
}

// TestValidate_RejectsInvalidEffort verifies specialist.effort MUST be
// low|medium|high|xhigh|max (AC-HV4-005a + design §C.2).
func TestValidate_RejectsInvalidEffort(t *testing.T) {
	cases := []struct {
		effort   string
		wantValid bool
	}{
		{EffortLow, true},
		{EffortMedium, true},
		{EffortHigh, true},
		{EffortXhigh, true},
		{EffortMax, true},
		{"ultra", false},
		{"", false},
		{"HIGH", false}, // case-sensitive
	}
	for _, tc := range cases {
		t.Run(tc.effort, func(t *testing.T) {
			m := validManifest()
			m.Specialists[0].Effort = tc.effort
			err := Validate(m)
			if tc.wantValid && err != nil {
				t.Fatalf("Validate rejected valid effort %q: %v", tc.effort, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("Validate accepted invalid effort %q", tc.effort)
			}
		})
	}
}

// TestValidate_RejectsInvalidModel verifies specialist.model MUST be
// inherit|haiku|sonnet|opus (AC-HV4-005a + design §C.2).
func TestValidate_RejectsInvalidModel(t *testing.T) {
	cases := []struct {
		model    string
		wantValid bool
	}{
		{ModelInherit, true},
		{ModelHaiku, true},
		{ModelSonnet, true},
		{ModelOpus, true},
		{"gpt-4", false},
		{"", false},
		{"Sonnet", false}, // case-sensitive
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			m := validManifest()
			m.Specialists[0].Model = tc.model
			err := Validate(m)
			if tc.wantValid && err != nil {
				t.Fatalf("Validate rejected valid model %q: %v", tc.model, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("Validate accepted invalid model %q", tc.model)
			}
		})
	}
}

// TestValidate_RejectsInvalidPattern verifies patterns[] entries MUST be from
// the 6-pattern catalog (AC-HV4-004b + design §C.2).
func TestValidate_RejectsInvalidPattern(t *testing.T) {
	cases := []struct {
		pattern  string
		wantValid bool
	}{
		{PatternPipeline, true},
		{PatternFanOutFanIn, true},
		{PatternExpertPool, true},
		{PatternProducerReviewer, true},
		{PatternSupervisor, true},
		{PatternHierarchicalDelegation, true},
		{"Custom-Pattern", false},
		{"pipeline-lowercase", false}, // case-sensitive
	}
	for _, tc := range cases {
		t.Run(tc.pattern, func(t *testing.T) {
			m := validManifest()
			m.Patterns = []string{tc.pattern}
			err := Validate(m)
			if tc.wantValid && err != nil {
				t.Fatalf("Validate rejected valid pattern %q: %v", tc.pattern, err)
			}
			if !tc.wantValid && err == nil {
				t.Fatalf("Validate accepted invalid pattern %q", tc.pattern)
			}
		})
	}
}

// TestValidate_RequiresAllFiveSpecialistFields verifies a specialist missing
// ANY of the 5 sub-fields (role/primitive/isolation/effort/model) is rejected.
// AC-HV4-005a: 0 specialists missing any of the 5 fields.
func TestValidate_RequiresAllFiveSpecialistFields(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Specialist)
		field  string
	}{
		{"missing_role", func(s *Specialist) { s.Role = "" }, "role"},
		{"missing_primitive", func(s *Specialist) { s.Primitive = "" }, "primitive"},
		{"missing_isolation", func(s *Specialist) { s.Isolation = "" }, "isolation"},
		{"missing_effort", func(s *Specialist) { s.Effort = "" }, "effort"},
		{"missing_model", func(s *Specialist) { s.Model = "" }, "model"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validManifest()
			tc.mutate(&m.Specialists[0])
			err := Validate(m)
			if err == nil {
				t.Fatalf("Validate accepted specialist with %s (expected rejection)", tc.name)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("Validate error %q does not mention field %q", err.Error(), tc.field)
			}
		})
	}
}

// TestValidate_AcceptsMultipleSpecialists verifies a manifest with 2+
// specialists (each valid) passes — the common multi-specialist harness case.
func TestValidate_AcceptsMultipleSpecialists(t *testing.T) {
	m := validManifest()
	m.Specialists = append(m.Specialists, Specialist{
		Role:      "coverage-gap-finder",
		Primitive: PrimitiveDynamicWorkflow,
		Isolation: IsolationWorktree,
		Effort:    EffortHigh,
		Model:     ModelSonnet,
	})
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected 2-specialist manifest: %v", err)
	}
}
