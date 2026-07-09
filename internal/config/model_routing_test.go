package config

// model_routing_test.go — SPEC-TOKEN-ROUTING-001 tests (AC-TR-001..004, AC-TR-007, AC-TR-012)
// + SPEC-AGENT-ARCH-V2-001 M3 tests (AC-AA2-008, AC-AA2-009 — RouteModelFor 3-arg +
// model_routing_profiles 3×12 matrix).
//
// Covers: round-trip load (flat legacy + 3-tier profiles), happy-path
// RouteModelFor (3-arg), closed-set validation (model/effort/perfTier),
// fallback on absent pair, input tier/phase/perfTier validation, shared
// closed-set with workflow_agents, unavailable-model advisory surface, and the
// 36-entry 3×12 golden matrix (design.md §D.5).

import (
	"os"
	"path/filepath"
	"testing"
)

// fullRoutingYAML is a 12-entry (3 Tier x 4 Phase) legacy flat model_routing
// block used by the SPEC-TOKEN-ROUTING-001 round-trip / validation tests.
const fullRoutingYAML = `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: sonnet, effort: medium }
        S-run: { model: sonnet, effort: high }
        S-sync: { model: sonnet, effort: low }
        S-mx: { model: sonnet, effort: low }
        M-plan: { model: sonnet, effort: medium }
        M-run: { model: sonnet, effort: xhigh }
        M-sync: { model: sonnet, effort: medium }
        M-mx: { model: sonnet, effort: low }
        L-plan: { model: opus, effort: high }
        L-run: { model: opus, effort: xhigh }
        L-sync: { model: sonnet, effort: high }
        L-mx: { model: sonnet, effort: medium }
`

// fullProfilesYAML is the No-Haiku 3-tier model_routing_profiles block
// (SPEC-AGENT-ARCH-V2-001 M3, design.md §D.5) — 3 perfTiers × 12 (Tier×Phase)
// cells = 36 entries. Each perfTier (max/medium/low) carries S/M/L ×
// plan/run/sync/mx.
const fullProfilesYAML = `workflow:
    default_mode: ""
    model_routing_profiles:
        max:
            S-plan: { model: opus, effort: high }
            S-run: { model: sonnet, effort: xhigh }
            S-sync: { model: sonnet, effort: medium }
            S-mx: { model: sonnet, effort: low }
            M-plan: { model: opus, effort: xhigh }
            M-run: { model: sonnet, effort: xhigh }
            M-sync: { model: sonnet, effort: medium }
            M-mx: { model: sonnet, effort: low }
            L-plan: { model: opus, effort: xhigh }
            L-run: { model: sonnet, effort: xhigh }
            L-sync: { model: sonnet, effort: high }
            L-mx: { model: sonnet, effort: medium }
        medium:
            S-plan: { model: sonnet, effort: medium }
            S-run: { model: sonnet, effort: high }
            S-sync: { model: sonnet, effort: low }
            S-mx: { model: sonnet, effort: low }
            M-plan: { model: sonnet, effort: medium }
            M-run: { model: sonnet, effort: xhigh }
            M-sync: { model: sonnet, effort: medium }
            M-mx: { model: sonnet, effort: low }
            L-plan: { model: opus, effort: high }
            L-run: { model: sonnet, effort: xhigh }
            L-sync: { model: sonnet, effort: high }
            L-mx: { model: sonnet, effort: medium }
        low:
            S-plan: { model: sonnet, effort: medium }
            S-run: { model: sonnet, effort: high }
            S-sync: { model: sonnet, effort: low }
            S-mx: { model: sonnet, effort: low }
            M-plan: { model: sonnet, effort: medium }
            M-run: { model: sonnet, effort: high }
            M-sync: { model: sonnet, effort: low }
            M-mx: { model: sonnet, effort: low }
            L-plan: { model: sonnet, effort: xhigh }
            L-run: { model: sonnet, effort: xhigh }
            L-sync: { model: sonnet, effort: medium }
            L-mx: { model: sonnet, effort: low }
`

// seedWorkflowWithModelRouting writes the given workflow.yaml content to a
// temp project root and returns the root path.
func seedWorkflowWithModelRouting(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// ---------------------------------------------------------------------------
// Legacy flat ModelRouting tests (SPEC-TOKEN-ROUTING-001)
// ---------------------------------------------------------------------------

// TestModelRoutingRoundTripLoad verifies the 12-entry flat block loads into the
// typed map with correct model/effort values (AC-TR-001).
func TestModelRoutingRoundTripLoad(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullRoutingYAML)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	mr := cfg.Workflow.ModelRouting
	if len(mr) != 12 {
		t.Fatalf("ModelRouting length = %d, want 12", len(mr))
	}
	if got := mr["S-sync"]; got.Model != "sonnet" || got.Effort != "low" {
		t.Errorf("S-sync = %+v, want {sonnet low}", got)
	}
	if got := mr["L-run"]; got.Model != "opus" || got.Effort != "xhigh" {
		t.Errorf("L-run = %+v, want {opus xhigh}", got)
	}
}

func TestValidateModelRoutingClosedSetModel(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: gpt-4, effort: low }
`)

	_, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw (loader does not validate): %v", err)
	}

	raw, rerr := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	verr := ValidateModelRoutingFromYAML(raw)
	if verr == nil {
		t.Fatalf("expected validation error for model gpt-4, got nil")
	}
}

func TestValidateModelRoutingClosedSetEffort(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: sonnet, effort: ultra }
`)

	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if verr := ValidateModelRoutingFromYAML(raw); verr == nil {
		t.Fatalf("expected validation error for effort ultra, got nil")
	}
}

func TestValidateModelRoutingValid(t *testing.T) {
	t.Parallel()
	if verr := ValidateModelRoutingFromYAML([]byte(fullRoutingYAML)); verr != nil {
		t.Fatalf("expected nil for valid 12-entry block, got %v", verr)
	}
}

func TestSharedClosedSetWithWorkflowAgents(t *testing.T) {
	t.Parallel()
	workflowAgentsEfforts := map[string]bool{
		"low": true, "medium": true, "high": true, "xhigh": true,
	}
	for effort := range workflowAgentsEfforts {
		if !validRoutingEfforts[effort] {
			t.Errorf("effort %q drift (REQ-TR-007)", effort)
		}
	}
	for effort := range validRoutingEfforts {
		if effort == "max" {
			continue
		}
		if !workflowAgentsEfforts[effort] {
			t.Errorf("effort %q drift (REQ-TR-007)", effort)
		}
	}
}

func TestValidateModelRoutingGlmAllowed(t *testing.T) {
	t.Parallel()
	if !validRoutingModels["glm"] {
		t.Errorf("glm must be in the model_routing closed model set (REQ-TR-012)")
	}
}

func TestLoadModelRouting(t *testing.T) {
	t.Parallel()

	writeTempYAML := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "workflow.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	tests := []struct {
		name    string
		prepare func(t *testing.T) string
		wantErr bool
		check   func(t *testing.T, m map[string]ModelRoutingEntry)
	}{
		{
			name:    "happy path — valid model_routing block loads typed entries",
			prepare: func(t *testing.T) string { return writeTempYAML(t, fullRoutingYAML) },
			check: func(t *testing.T, m map[string]ModelRoutingEntry) {
				if len(m) != 12 {
					t.Fatalf("ModelRouting length = %d, want 12", len(m))
				}
			},
		},
		{
			name:    "file not found — absence returns nil, nil",
			prepare: func(t *testing.T) string { return filepath.Join(t.TempDir(), "does-not-exist.yaml") },
			check: func(t *testing.T, m map[string]ModelRoutingEntry) {
				if m != nil {
					t.Errorf("absent-file ModelRouting = %+v, want nil", m)
				}
			},
		},
		{
			name:    "malformed YAML — unmarshal failure returns non-nil error",
			prepare: func(t *testing.T) string { return writeTempYAML(t, "workflow:\n    model_routing: { model: sonnet\n") },
			wantErr: true,
		},
		{
			name:    "absent block — valid YAML without model_routing returns nil, nil",
			prepare: func(t *testing.T) string { return writeTempYAML(t, "workflow:\n    default_mode: \"\"\n") },
			check: func(t *testing.T, m map[string]ModelRoutingEntry) {
				if m != nil {
					t.Errorf("absent-block ModelRouting = %+v, want nil", m)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.prepare(t)
			m, err := LoadModelRouting(path)
			if tt.wantErr && err == nil {
				t.Fatalf("LoadModelRouting(%q) error = nil, want non-nil", path)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("LoadModelRouting(%q) error = %v, want nil", path, err)
			}
			if tt.check != nil {
				tt.check(t, m)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// M3 tests — RouteModelFor 3-arg + model_routing_profiles (SPEC-AGENT-ARCH-V2-001)
// ---------------------------------------------------------------------------

// TestRouteModelForHappyPath3Arg verifies RouteModelFor(specTier, phase,
// perfTier) returns the declared profiles entry with FallbackApplied=false
// (AC-AA2-008).
func TestRouteModelForHappyPath3Arg(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullProfilesYAML)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	// max tier: L-plan = opus/xhigh (design.md §D.5).
	entry, err := cfg.RouteModelFor("L", "plan", "max")
	if err != nil {
		t.Fatalf("RouteModelFor(L,plan,max): %v", err)
	}
	if entry.Model != "opus" || entry.Effort != "xhigh" {
		t.Errorf("RouteModelFor(L,plan,max) = %+v, want {opus xhigh}", entry)
	}
	if entry.FallbackApplied {
		t.Errorf("RouteModelFor(L,plan,max) FallbackApplied = true, want false")
	}

	// medium tier: S-run = sonnet/high.
	entry, err = cfg.RouteModelFor("S", "run", "medium")
	if err != nil {
		t.Fatalf("RouteModelFor(S,run,medium): %v", err)
	}
	if entry.Model != "sonnet" || entry.Effort != "high" {
		t.Errorf("RouteModelFor(S,run,medium) = %+v, want {sonnet high}", entry)
	}
}

// TestRouteModelForFallback3Arg verifies an absent perfTier or (specTier,
// phase) pair returns the documented default with FallbackApplied=true
// (AC-AA2-008).
func TestRouteModelForFallback3Arg(t *testing.T) {
	t.Parallel()
	// Only max tier, only S-plan — every other lookup falls back.
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing_profiles:
        max:
            S-plan: { model: opus, effort: high }
`)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	// Declared pair: no fallback.
	entry, err := cfg.RouteModelFor("S", "plan", "max")
	if err != nil {
		t.Fatalf("RouteModelFor(S,plan,max): %v", err)
	}
	if entry.FallbackApplied {
		t.Errorf("RouteModelFor(S,plan,max) FallbackApplied = true, want false (declared)")
	}

	// Absent (specTier, phase) within declared perfTier: fallback.
	entry, err = cfg.RouteModelFor("L", "mx", "max")
	if err != nil {
		t.Fatalf("RouteModelFor(L,mx,max): %v", err)
	}
	if !entry.FallbackApplied {
		t.Errorf("RouteModelFor(L,mx,max) FallbackApplied = false, want true (absent pair)")
	}

	// Absent perfTier: fallback.
	entry, err = cfg.RouteModelFor("S", "plan", "low")
	if err != nil {
		t.Fatalf("RouteModelFor(S,plan,low) on absent perfTier: %v", err)
	}
	if !entry.FallbackApplied {
		t.Errorf("RouteModelFor(S,plan,low) FallbackApplied = false, want true (absent perfTier)")
	}
}

// TestRouteModelForNilProfiles verifies a nil profiles map (block absent)
// yields the fallback for every lookup without error.
func TestRouteModelForNilProfiles(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, "workflow:\n    default_mode: \"\"\n")

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if cfg.Workflow.ModelRoutingProfiles != nil {
		t.Fatalf("expected nil ModelRoutingProfiles for absent block, got %+v", cfg.Workflow.ModelRoutingProfiles)
	}

	for _, tc := range []struct{ tier, phase, perf string }{
		{"S", "plan", "max"}, {"M", "run", "medium"}, {"L", "sync", "low"},
	} {
		entry, err := cfg.RouteModelFor(tc.tier, tc.phase, tc.perf)
		if err != nil {
			t.Errorf("RouteModelFor(%s,%s,%s) on nil profiles: %v", tc.tier, tc.phase, tc.perf, err)
		}
		if !entry.FallbackApplied {
			t.Errorf("RouteModelFor(%s,%s,%s) FallbackApplied = false, want true", tc.tier, tc.phase, tc.perf)
		}
	}
}

// TestRouteModelForInvalidInput3Arg verifies tier/phase/perfTier outside the
// closed set return an error (AC-AA2-008 input validation).
func TestRouteModelForInvalidInput3Arg(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullProfilesYAML)
	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	if _, err := cfg.RouteModelFor("X", "run", "max"); err == nil {
		t.Errorf("RouteModelFor(X,run,max): expected error for invalid tier, got nil")
	}
	if _, err := cfg.RouteModelFor("S", "deploy", "max"); err == nil {
		t.Errorf("RouteModelFor(S,deploy,max): expected error for invalid phase, got nil")
	}
	if _, err := cfg.RouteModelFor("S", "run", "ultra"); err == nil {
		t.Errorf("RouteModelFor(S,run,ultra): expected error for invalid perfTier, got nil")
	}
}

// TestRouteModelForUnavailableModel3Arg verifies the accessor returns a glm
// entry as-is for the orchestrator's env-aware advisory (REQ-TR-012).
func TestRouteModelForUnavailableModel3Arg(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing_profiles:
        max:
            S-sync: { model: glm, effort: low }
`)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	entry, err := cfg.RouteModelFor("S", "sync", "max")
	if err != nil {
		t.Fatalf("RouteModelFor(S,sync,max) with glm model: %v", err)
	}
	if entry.Model != "glm" {
		t.Errorf("RouteModelFor(S,sync,max) Model = %q, want glm", entry.Model)
	}
}

// TestRouteModelFor3x12Matrix is the golden test that exercises all 36 entries
// of the model_routing_profiles block against the design.md §D.5 expected
// values (AC-AA2-009). Every cell MUST be present and match {model, effort}.
func TestRouteModelFor3x12Matrix(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullProfilesYAML)
	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
}

	// Expected {model, effort} per (perfTier, tier, phase) — design.md §D.5.
	expected := map[string]map[string]ModelRoutingEntry{
		"max": {
			"S-plan": {"opus", "high", false}, "S-run": {"sonnet", "xhigh", false},
			"S-sync": {"sonnet", "medium", false}, "S-mx": {"sonnet", "low", false},
			"M-plan": {"opus", "xhigh", false}, "M-run": {"sonnet", "xhigh", false},
			"M-sync": {"sonnet", "medium", false}, "M-mx": {"sonnet", "low", false},
			"L-plan": {"opus", "xhigh", false}, "L-run": {"sonnet", "xhigh", false},
			"L-sync": {"sonnet", "high", false}, "L-mx": {"sonnet", "medium", false},
		},
		"medium": {
			"S-plan": {"sonnet", "medium", false}, "S-run": {"sonnet", "high", false},
			"S-sync": {"sonnet", "low", false}, "S-mx": {"sonnet", "low", false},
			"M-plan": {"sonnet", "medium", false}, "M-run": {"sonnet", "xhigh", false},
			"M-sync": {"sonnet", "medium", false}, "M-mx": {"sonnet", "low", false},
			"L-plan": {"opus", "high", false}, "L-run": {"sonnet", "xhigh", false},
			"L-sync": {"sonnet", "high", false}, "L-mx": {"sonnet", "medium", false},
		},
		"low": {
			"S-plan": {"sonnet", "medium", false}, "S-run": {"sonnet", "high", false},
			"S-sync": {"sonnet", "low", false}, "S-mx": {"sonnet", "low", false},
			"M-plan": {"sonnet", "medium", false}, "M-run": {"sonnet", "high", false},
			"M-sync": {"sonnet", "low", false}, "M-mx": {"sonnet", "low", false},
			"L-plan": {"sonnet", "xhigh", false}, "L-run": {"sonnet", "xhigh", false},
			"L-sync": {"sonnet", "medium", false}, "L-mx": {"sonnet", "low", false},
		},
	}

	tiers := []string{"S", "M", "L"}
	phases := []string{"plan", "run", "sync", "mx"}
	perfTiers := []string{"max", "medium", "low"}
	count := 0
	for _, perf := range perfTiers {
		for _, tier := range tiers {
			for _, phase := range phases {
				key := tier + "-" + phase
				want := expected[perf][key]
				got, err := cfg.RouteModelFor(tier, phase, perf)
				if err != nil {
					t.Fatalf("RouteModelFor(%s,%s,%s): %v", tier, phase, perf, err)
				}
				if got.Model != want.Model || got.Effort != want.Effort {
					t.Errorf("RouteModelFor(%s,%s,%s) = {%s %s}, want {%s %s}",
						tier, phase, perf, got.Model, got.Effort, want.Model, want.Effort)
				}
				if got.FallbackApplied {
					t.Errorf("RouteModelFor(%s,%s,%s) FallbackApplied = true, want false", tier, phase, perf)
				}
				count++
			}
		}
	}
	if count != 36 {
		t.Fatalf("3×12 matrix exercised %d cells, want 36", count)
	}
}

// TestValidateModelRoutingProfiles verifies the 3-tier profiles block validates
// cleanly (happy path) and rejects an out-of-closed-set value.
func TestValidateModelRoutingProfiles(t *testing.T) {
	t.Parallel()
	// Happy path: full 36-entry block validates.
	root := seedWorkflowWithModelRouting(t, fullProfilesYAML)
	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if verr := cfg.ValidateModelRoutingProfiles(); verr != nil {
		t.Fatalf("ValidateModelRoutingProfiles on valid block: %v", verr)
	}

	// Rejection: invalid perfTier key.
	badRoot := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing_profiles:
        ultra:
            S-plan: { model: sonnet, effort: low }
`)
	badCfg, err := NewConfigManager().LoadRaw(badRoot)
	if err != nil {
		t.Fatalf("LoadRaw (bad): %v", err)
	}
	if verr := badCfg.ValidateModelRoutingProfiles(); verr == nil {
		t.Fatalf("expected validation error for invalid perfTier ultra, got nil")
	}
}
