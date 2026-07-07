package config

// model_routing_test.go — SPEC-TOKEN-ROUTING-001 tests (AC-TR-001..004, AC-TR-007, AC-TR-012).
// Covers: round-trip load, happy-path RouteModelFor, closed-set validation
// (model/effort), fallback on absent pair, input tier/phase validation,
// shared closed-set with workflow_agents, unavailable-model advisory surface.

import (
	"os"
	"path/filepath"
	"testing"
)

// fullRoutingYAML is a 12-entry (3 Tier x 4 Phase) model_routing block used
// across happy-path tests.
const fullRoutingYAML = `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: sonnet, effort: medium }
        S-run: { model: sonnet, effort: high }
        S-sync: { model: haiku, effort: low }
        S-mx: { model: haiku, effort: low }
        M-plan: { model: sonnet, effort: medium }
        M-run: { model: sonnet, effort: xhigh }
        M-sync: { model: sonnet, effort: medium }
        M-mx: { model: haiku, effort: low }
        L-plan: { model: opus, effort: high }
        L-run: { model: opus, effort: xhigh }
        L-sync: { model: sonnet, effort: high }
        L-mx: { model: sonnet, effort: medium }
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

// TestModelRoutingRoundTripLoad verifies the 12-entry block loads into the
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
	if got := mr["S-sync"]; got.Model != "haiku" || got.Effort != "low" {
		t.Errorf("S-sync = %+v, want {haiku low}", got)
	}
	if got := mr["L-run"]; got.Model != "opus" || got.Effort != "xhigh" {
		t.Errorf("L-run = %+v, want {opus xhigh}", got)
	}
}

// TestRouteModelForHappyPath verifies RouteModelFor returns the declared entry
// with FallbackApplied=false (AC-TR-003).
func TestRouteModelForHappyPath(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullRoutingYAML)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	entry, err := cfg.RouteModelFor("S", "sync")
	if err != nil {
		t.Fatalf("RouteModelFor(S,sync): %v", err)
	}
	if entry.Model != "haiku" || entry.Effort != "low" {
		t.Errorf("RouteModelFor(S,sync) = %+v, want {haiku low}", entry)
	}
	if entry.FallbackApplied {
		t.Errorf("RouteModelFor(S,sync) FallbackApplied = true, want false (pair is declared)")
	}

	entry, err = cfg.RouteModelFor("L", "run")
	if err != nil {
		t.Fatalf("RouteModelFor(L,run): %v", err)
	}
	if entry.Model != "opus" || entry.Effort != "xhigh" {
		t.Errorf("RouteModelFor(L,run) = %+v, want {opus xhigh}", entry)
	}
	if entry.FallbackApplied {
		t.Errorf("RouteModelFor(L,run) FallbackApplied = true, want false")
	}
}

// TestRouteModelForFallback verifies that an absent (Tier, Phase) pair returns
// the documented default with FallbackApplied=true (AC-TR-002).
func TestRouteModelForFallback(t *testing.T) {
	t.Parallel()
	// Only S entries — (L, mx) is absent.
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: sonnet, effort: medium }
`)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	entry, err := cfg.RouteModelFor("L", "mx")
	if err != nil {
		t.Fatalf("RouteModelFor(L,mx) on absent pair: unexpected error %v", err)
	}
	if !entry.FallbackApplied {
		t.Errorf("RouteModelFor(L,mx) on absent pair: FallbackApplied = false, want true")
	}
	if entry.Model != defaultRoutingEntry.Model {
		t.Errorf("RouteModelFor(L,mx) fallback Model = %q, want %q", entry.Model, defaultRoutingEntry.Model)
	}
	if entry.Effort != defaultRoutingEntry.Effort {
		t.Errorf("RouteModelFor(L,mx) fallback Effort = %q, want %q", entry.Effort, defaultRoutingEntry.Effort)
	}
}

// TestRouteModelForNilMap verifies a nil map (block absent) yields the
// fallback for every lookup without error.
func TestRouteModelForNilMap(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, "workflow:\n    default_mode: \"\"\n")

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}
	if cfg.Workflow.ModelRouting != nil {
		t.Fatalf("expected nil ModelRouting for absent block, got %+v", cfg.Workflow.ModelRouting)
	}

	for _, tc := range []struct{ tier, phase string }{
		{"S", "plan"}, {"M", "run"}, {"L", "sync"},
	} {
		entry, err := cfg.RouteModelFor(tc.tier, tc.phase)
		if err != nil {
			t.Errorf("RouteModelFor(%s,%s) on nil map: %v", tc.tier, tc.phase, err)
		}
		if !entry.FallbackApplied {
			t.Errorf("RouteModelFor(%s,%s) on nil map: FallbackApplied = false, want true", tc.tier, tc.phase)
		}
	}
}

// TestRouteModelForInvalidInput verifies tier/phase outside the closed set
// return an error (acceptance.md §D.2 edge case — input validation).
func TestRouteModelForInvalidInput(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, fullRoutingYAML)
	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	if _, err := cfg.RouteModelFor("X", "run"); err == nil {
		t.Errorf("RouteModelFor(X,run): expected error for invalid tier, got nil")
	}
	if _, err := cfg.RouteModelFor("S", "deploy"); err == nil {
		t.Errorf("RouteModelFor(S,deploy): expected error for invalid phase, got nil")
	}
}

// TestValidateModelRoutingClosedSetModel verifies a non-closed-set model value
// is rejected (AC-TR-004, AC-TR-007).
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

	// ValidateModelRoutingFromYAML does closed-set validation.
	raw, rerr := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	verr := ValidateModelRoutingFromYAML(raw)
	if verr == nil {
		t.Fatalf("expected validation error for model gpt-4, got nil")
	}
}

// TestValidateModelRoutingClosedSetEffort verifies a non-closed-set effort
// value is rejected (AC-TR-004, AC-TR-007).
func TestValidateModelRoutingClosedSetEffort(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing:
        S-plan: { model: haiku, effort: ultra }
`)

	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if verr := ValidateModelRoutingFromYAML(raw); verr == nil {
		t.Fatalf("expected validation error for effort ultra, got nil")
	}
}

// TestValidateModelRoutingValid verifies the full 12-entry block validates
// cleanly (happy path for REQ-TR-007).
func TestValidateModelRoutingValid(t *testing.T) {
	t.Parallel()
	if verr := ValidateModelRoutingFromYAML([]byte(fullRoutingYAML)); verr != nil {
		t.Fatalf("expected nil for valid 12-entry block, got %v", verr)
	}
}

// TestSharedClosedSetWithWorkflowAgents verifies the model_routing effort set
// matches the workflow_agents effort set (REQ-TR-007 — the two maps' effort
// sets cannot drift apart).
func TestSharedClosedSetWithWorkflowAgents(t *testing.T) {
	t.Parallel()
	// The canonical workflow_agents effort values (from workflow.yaml).
	workflowAgentsEfforts := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"xhigh":  true,
	}

	// Every effort value workflow_agents uses MUST be valid in model_routing.
	for effort := range workflowAgentsEfforts {
		if !validRoutingEfforts[effort] {
			t.Errorf("effort %q is valid for workflow_agents but NOT for model_routing — drift (REQ-TR-007 violation)", effort)
		}
	}

	// Symmetrically, the model_routing effort set minus "max" (model_routing's
	// extension) must be a subset of the workflow_agents effort set. "max" is a
	// legitimate model_routing extension for the most expensive Tier/Phase.
	for effort := range validRoutingEfforts {
		if effort == "max" {
			continue
		}
		if !workflowAgentsEfforts[effort] {
			t.Errorf("effort %q is valid for model_routing but absent from workflow_agents — drift (REQ-TR-007)", effort)
		}
	}
}

// TestValidateModelRoutingGlmAllowed verifies glm is in the closed model set
// (REQ-TR-012 — glm is the deployment-neutrality extension; advisory, not
// blocking, when GLM env is absent).
func TestValidateModelRoutingGlmAllowed(t *testing.T) {
	t.Parallel()
	if !validRoutingModels["glm"] {
		t.Errorf("glm must be in the model_routing closed model set (REQ-TR-012)")
	}
}

// TestRouteModelForUnavailableModel verifies the accessor surface for a glm
// entry when GLM env is absent. REQ-TR-012 is a SHOULD: the loader surfaces
// an advisory rather than blocking. The advisory responsibility lives at the
// orchestrator layer (env detection); the loader only ensures glm is a valid
// value and is returned as-is. This test confirms the entry is returned
// without error, with the model value preserved for the orchestrator's
// env-aware advisory.
func TestRouteModelForUnavailableModel(t *testing.T) {
	t.Parallel()
	root := seedWorkflowWithModelRouting(t, `workflow:
    default_mode: ""
    model_routing:
        S-sync: { model: glm, effort: low }
`)

	cfg, err := NewConfigManager().LoadRaw(root)
	if err != nil {
		t.Fatalf("LoadRaw: %v", err)
	}

	entry, err := cfg.RouteModelFor("S", "sync")
	if err != nil {
		t.Fatalf("RouteModelFor(S,sync) with glm model: %v", err)
	}
	if entry.Model != "glm" {
		t.Errorf("RouteModelFor(S,sync) Model = %q, want glm (orchestrator decides advisory)", entry.Model)
	}
	// The orchestrator (not the loader) checks GLM env presence and surfaces
	// the advisory. The loader's contract here is: return the entry, don't
	// block on glm.
}

// TestLoadModelRouting exercises LoadModelRouting directly across the four
// contract paths: happy-path load, file-not-found (absence == fallback, not
// failure), malformed YAML (parse error), and absent model_routing block
// (valid YAML, no routing == fallback). Backfills the sync-auditor D3 coverage
// gap — LoadModelRouting was the only model_routing loader function with 0%
// direct coverage.
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
				if sync := m["S-sync"]; sync.Model != "haiku" || sync.Effort != "low" || sync.FallbackApplied {
					t.Errorf("S-sync = %+v, want {Model:haiku Effort:low FallbackApplied:false}", sync)
				}
				if lrun := m["L-run"]; lrun.Model != "opus" || lrun.Effort != "xhigh" || lrun.FallbackApplied {
					t.Errorf("L-run = %+v, want {Model:opus Effort:xhigh FallbackApplied:false}", lrun)
				}
			},
		},
		{
			name:    "file not found — absence returns nil, nil (fallback, not failure)",
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
			name:    "absent block — valid YAML without model_routing key returns nil, nil",
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
