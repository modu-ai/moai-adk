package v4manifest

import (
	"encoding/json"
	"strings"
	"testing"

	// Test-only import of the learning-subsystem Tier SSOT. The production
	// v4manifest package is deliberately separate from internal/harness (see
	// types.go package doc), but the learning-tier vocabulary MUST be derived
	// from harness.Tier.String() (REQ-HRR-002). This cross-check mechanically
	// enforces that derivation — it fails if either side drifts.
	"github.com/modu-ai/moai-adk/internal/harness"
)

// learningManifest returns the canonical valid fixture with a learning block
// attached. Tests mutate exactly one learning field per case.
func learningManifest() Manifest {
	m := validManifest()
	m.Learning = &LearningBlock{
		Enabled:           true,
		Tier:              LearningTierAutoUpdate,
		ConfidenceFloor:   0.70,
		MaxFindingsPerRun: 5,
	}
	return m
}

// TestManifestLearningBlockParsing verifies AC-HRR-001: a manifest carrying a
// learning block parses all 4 declared fields verbatim.
func TestManifestLearningBlockParsing(t *testing.T) {
	raw := `{
		"name": "moai-adk-dev",
		"domain": "moai-adk CLI template development",
		"source_request": "build a harness for moai-adk CLI template development",
		"patterns": ["Pipeline", "Producer-Reviewer"],
		"entry_command": "/harness:moai-adk-dev",
		"runner_workflow": "harness-moai-adk-dev-run.js",
		"specialists": [
			{
				"role": "template-neutrality-auditor",
				"primitive": "sub-agent",
				"isolation": "none",
				"effort": "xhigh",
				"model": "inherit"
			}
		],
		"sprint_contract": {
			"dimensions": ["neutrality", "coverage"],
			"thresholds": {"neutrality": 0, "coverage": 0.85}
		},
		"learning": {
			"enabled": true,
			"tier": "auto_update",
			"confidence_floor": 0.72,
			"max_findings_per_run": 8
		}
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("json.Unmarshal failed on manifest with learning block: %v", err)
	}

	if m.Learning == nil {
		t.Fatalf("Learning is nil after parse (AC-HRR-001 requires the block to be recognized)")
	}
	if !m.Learning.Enabled {
		t.Errorf("Learning.Enabled = false, want true")
	}
	if m.Learning.Tier != "auto_update" {
		t.Errorf("Learning.Tier = %q, want %q", m.Learning.Tier, "auto_update")
	}
	if m.Learning.ConfidenceFloor != 0.72 {
		t.Errorf("Learning.ConfidenceFloor = %v, want 0.72", m.Learning.ConfidenceFloor)
	}
	if m.Learning.MaxFindingsPerRun != 8 {
		t.Errorf("Learning.MaxFindingsPerRun = %d, want 8", m.Learning.MaxFindingsPerRun)
	}
}

// TestManifestLearningBlockLegacyNil verifies AC-HRR-001 / AC-HRR-010: a
// manifest WITHOUT a learning block parses successfully and leaves Learning ==
// nil. The legacy 8-field (+Schedule) baseline MUST remain valid (REQ-HRR-010).
func TestManifestLearningBlockLegacyNil(t *testing.T) {
	raw := `{
		"name": "moai-adk-dev",
		"domain": "moai-adk CLI template development",
		"source_request": "build a harness for moai-adk CLI template development",
		"patterns": ["Pipeline", "Producer-Reviewer"],
		"entry_command": "/harness:moai-adk-dev",
		"runner_workflow": "harness-moai-adk-dev-run.js",
		"specialists": [
			{
				"role": "template-neutrality-auditor",
				"primitive": "sub-agent",
				"isolation": "none",
				"effort": "xhigh",
				"model": "inherit"
			}
		],
		"sprint_contract": {
			"dimensions": ["neutrality", "coverage"],
			"thresholds": {"neutrality": 0, "coverage": 0.85}
		}
	}`

	var m Manifest
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("json.Unmarshal failed on legacy manifest (no learning block): %v", err)
	}
	if m.Learning != nil {
		t.Fatalf("Learning = %v, want nil for legacy manifest without learning block", m.Learning)
	}
	// Legacy manifest MUST still pass Validate (REQ-HRR-010 — no rejection).
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected legacy manifest without learning block (violates REQ-HRR-010): %v", err)
	}
}

// TestLearningTierVocabulary verifies AC-HRR-002: the tier vocabulary is derived
// from harness.Tier.String(). All 4 SSOT values are valid; any other value
// (including the parallel vocabulary PIPE-REPAIR removed, e.g.
// "recommendation"/"approval_required") MUST be rejected.
func TestLearningTierVocabulary(t *testing.T) {
	validTiers := []string{
		LearningTierObservation,
		LearningTierHeuristic,
		LearningTierRule,
		LearningTierAutoUpdate,
	}
	for _, tier := range validTiers {
		t.Run("valid/"+tier, func(t *testing.T) {
			m := learningManifest()
			m.Learning.Tier = tier
			if err := Validate(m); err != nil {
				t.Errorf("Validate rejected valid learning.tier %q: %v", tier, err)
			}
		})
	}

	invalidTiers := []string{
		"recommendation",    // parallel vocabulary PIPE-REPAIR removed (AP-1)
		"approval_required", // parallel vocabulary PIPE-REPAIR removed (AP-1)
		"auto",              // truncated
		"RULE",              // wrong case
		"",                  // empty is accepted as unset (EC-1 partial block policy)
	}
	for _, tier := range invalidTiers {
		t.Run("invalid/"+tier, func(t *testing.T) {
			m := learningManifest()
			m.Learning.Tier = tier
			// Empty tier is accepted (EC-1 partial-block policy: defaults
			// applied downstream in M2); only NON-EMPTY out-of-vocab values
			// are rejected.
			if tier == "" {
				if err := Validate(m); err != nil {
					t.Errorf("Validate rejected empty learning.tier (EC-1 should accept as unset): %v", err)
				}
				return
			}
			err := Validate(m)
			if err == nil {
				t.Errorf("Validate accepted invalid learning.tier %q (REQ-HRR-002 requires rejection of parallel vocabulary)", tier)
				return
			}
			if !strings.Contains(err.Error(), "learning.tier") {
				t.Errorf("Validate error for learning.tier %q does not name the field: %v", tier, err)
			}
		})
	}
}

// TestLearningTierVocabularyMatchesHarnessSSOT mechanically enforces
// REQ-HRR-002: the v4manifest learning-tier vocabulary MUST be derived from
// harness.Tier.String(), NOT a separate parallel vocabulary. The test imports
// the learning-subsystem SSOT and asserts the closed sets are identical.
//
// If harness.Tier.String() vocabulary changes, this test fails until the
// v4manifest constants are updated to match — preventing silent drift.
func TestLearningTierVocabularyMatchesHarnessSSOT(t *testing.T) {
	ssotTiers := map[string]bool{
		harness.TierObservation.String(): true,
		harness.TierHeuristic.String():   true,
		harness.TierRule.String():        true,
		harness.TierAutoUpdate.String():  true,
	}
	v4manifestTiers := map[string]bool{
		LearningTierObservation: true,
		LearningTierHeuristic:   true,
		LearningTierRule:        true,
		LearningTierAutoUpdate:  true,
	}

	// Every v4manifest tier MUST appear in the SSOT.
	for tier := range v4manifestTiers {
		if !ssotTiers[tier] {
			t.Errorf("v4manifest learning tier %q is NOT in harness.Tier.String() SSOT — parallel vocabulary (violates REQ-HRR-002 / AP-1)", tier)
		}
	}
	// Every SSOT tier MUST be recognized by v4manifest.
	for tier := range ssotTiers {
		if !v4manifestTiers[tier] {
			t.Errorf("harness.Tier.String() SSOT value %q is missing from v4manifest learning tier set — SSOT derivation incomplete (violates REQ-HRR-002)", tier)
		}
	}
	// validLearningTiers (the Validate-time map) MUST equal the constant set.
	for tier := range v4manifestTiers {
		if !validLearningTiers[tier] {
			t.Errorf("validLearningTiers map is missing %q — Validate will reject a SSOT-derived tier", tier)
		}
	}
	for tier := range validLearningTiers {
		if !v4manifestTiers[tier] {
			t.Errorf("validLearningTiers map contains %q which is NOT a SSOT tier — parallel vocabulary leak (violates REQ-HRR-002)", tier)
		}
	}
}

// TestValidate_LearningAbsentRegression verifies the additive-only guarantee
// for the learning block: a manifest WITHOUT a learning block (nil pointer)
// still passes Validate — every pre-existing valid manifest remains valid.
// This mirrors TestValidate_ScheduleAbsentRegression for the Schedule field.
func TestValidate_LearningAbsentRegression(t *testing.T) {
	m := validManifest()
	if m.Learning != nil {
		t.Fatalf("validManifest fixture unexpectedly carries a learning block")
	}
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected baseline valid manifest once Learning was added as a nil optional: %v", err)
	}
}

// TestValidate_LearningBlockValid verifies a fully-populated learning block
// on the canonical fixture passes Validate (AC-HRR-001 happy path).
func TestValidate_LearningBlockValid(t *testing.T) {
	m := learningManifest()
	if err := Validate(m); err != nil {
		t.Fatalf("Validate rejected valid manifest with learning block: %v", err)
	}
}
