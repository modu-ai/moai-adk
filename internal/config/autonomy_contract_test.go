package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadAutonomyFixture writes body as the workflow.yaml of a fresh temp project,
// loads it through the production Loader, and resolves workflow.autonomy.
// An empty body writes no workflow.yaml at all.
func loadAutonomyFixture(t *testing.T, body string) AutonomySettings {
	t.Helper()
	moaiDir := filepath.Join(t.TempDir(), ".moai")
	sections := filepath.Join(moaiDir, "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("write workflow.yaml: %v", err)
		}
	}
	cfg, err := NewLoader().Load(moaiDir)
	if err != nil {
		t.Fatalf("Load: %v (loading must always succeed)", err)
	}
	return ResolveAutonomy(cfg.Workflow)
}

// warningFor returns the first warning that names key, or "".
func warningFor(s AutonomySettings, key string) string {
	for _, w := range s.Warnings {
		if strings.Contains(w, key) {
			return w
		}
	}
	return ""
}

// TestAC_CONTRACT_019 covers the configuration sub-cases of AC-CONTRACT-019.
// The clause "a human-path sign still succeeds while the receipt path refuses
// with kickoff_decider_jev_sole" is the signer's behavior (M5) and is asserted
// there, not here.
func TestAC_CONTRACT_019(t *testing.T) {
	t.Run("defaults_without_autonomy_section", func(t *testing.T) {
		for name, body := range map[string]string{
			"no_workflow_file":       "",
			"workflow_without_block": "workflow:\n  default_mode: \"\"\n",
		} {
			t.Run(name, func(t *testing.T) {
				s := loadAutonomyFixture(t, body)
				if s.Mode != "guided" {
					t.Errorf("mode = %q, want guided", s.Mode)
				}
				if s.BatchSign {
					t.Error("batch_sign = true, want false")
				}
				if s.SecondReview != "required" {
					t.Errorf("second_review = %q, want required", s.SecondReview)
				}
				if s.PushDevelop {
					t.Error("push_develop = true, want false")
				}
				if s.BudgetDefault != (AutonomyBudget{Turns: 60, Operations: 40, AuditRetries: 2}) {
					t.Errorf("budget = %+v, want 60/40/2", s.BudgetDefault)
				}
				if s.Decider != "human" {
					t.Errorf("effective decider = %q, want human", s.Decider)
				}
				if s.JevMinConfidence != 0.50 {
					t.Errorf("jev_min_confidence = %v, want 0.50", s.JevMinConfidence)
				}
				if len(s.Warnings) != 0 || s.DeciderError != nil {
					t.Errorf("unexpected warnings %v / error %v", s.Warnings, s.DeciderError)
				}
			})
		}
	})

	t.Run("contract_without_decider_derives_llm", func(t *testing.T) {
		s := loadAutonomyFixture(t, "workflow:\n  autonomy:\n    mode: contract\n")
		if s.Mode != "contract" || s.Decider != "llm" {
			t.Errorf("mode/decider = %q/%q, want contract/llm", s.Mode, s.Decider)
		}
	})

	t.Run("contract_with_explicit_human_wins", func(t *testing.T) {
		s := loadAutonomyFixture(t, "workflow:\n  autonomy:\n    mode: contract\n    kickoff:\n      decider: human\n")
		if s.Decider != "human" {
			t.Errorf("decider = %q, want human (an explicit value wins)", s.Decider)
		}
	})

	fallbacks := []struct {
		name, body, key string
		check           func(AutonomySettings) bool
		want            string
	}{
		{"mode_yolo", "workflow:\n  autonomy:\n    mode: yolo\n", "workflow.autonomy.mode",
			func(s AutonomySettings) bool { return s.Mode == "guided" }, "guided"},
		{"second_review_sometimes", "workflow:\n  autonomy:\n    contract:\n      second_review: sometimes\n", "workflow.autonomy.contract.second_review",
			func(s AutonomySettings) bool { return s.SecondReview == "required" }, "required"},
		{"decider_oracle", "workflow:\n  autonomy:\n    kickoff:\n      decider: oracle\n", "workflow.autonomy.kickoff.decider",
			func(s AutonomySettings) bool { return s.Decider == "human" }, "human"},
		{"jev_min_confidence_1_5", "workflow:\n  autonomy:\n    kickoff:\n      jev_min_confidence: 1.5\n", "workflow.autonomy.kickoff.jev_min_confidence",
			func(s AutonomySettings) bool { return s.JevMinConfidence == 0.50 }, "0.50"},
	}
	for _, tc := range fallbacks {
		t.Run("fallback_"+tc.name, func(t *testing.T) {
			s := loadAutonomyFixture(t, tc.body)
			if !tc.check(s) {
				t.Errorf("value did not fall back to %s: %+v", tc.want, s)
			}
			if warningFor(s, tc.key) == "" {
				t.Errorf("no warning names %s; warnings = %v", tc.key, s.Warnings)
			}
			if s.DeciderError != nil {
				t.Errorf("unexpected configuration error: %v", s.DeciderError)
			}
		})
	}

	t.Run("valid_values_with_jev_disabled", func(t *testing.T) {
		s := loadAutonomyFixture(t, "workflow:\n"+
			"  jev:\n    enabled: false\n"+
			"  autonomy:\n    mode: contract\n"+
			"    contract:\n      second_review: advisory\n"+
			"    kickoff:\n      decider: llm+jev\n      jev_min_confidence: 0.6\n")
		if s.Mode != "contract" || s.SecondReview != "advisory" || s.Decider != "llm+jev" || s.JevMinConfidence != 0.6 {
			t.Errorf("values not returned as configured: %+v", s)
		}
		if s.JevEnabled {
			t.Error("JevEnabled = true, want false")
		}
		if s.DeciderError != nil || len(s.Warnings) != 0 {
			t.Errorf("unexpected error %v / warnings %v", s.DeciderError, s.Warnings)
		}
	})

	t.Run("decider_jev_is_a_configuration_error", func(t *testing.T) {
		s := loadAutonomyFixture(t, "workflow:\n  autonomy:\n    mode: contract\n    kickoff:\n      decider: jev\n")
		if s.DeciderError == nil {
			t.Fatal("no configuration error for decider: jev")
		}
		if !errors.Is(s.DeciderError, ErrKickoffDeciderJevSole) {
			t.Errorf("error %v is not ErrKickoffDeciderJevSole", s.DeciderError)
		}
		msg := s.DeciderError.Error()
		if !strings.Contains(msg, "workflow.autonomy.kickoff.decider") || !strings.Contains(msg, "never a sole decider") {
			t.Errorf("error text %q must name the key and state Jev is never a sole decider", msg)
		}
		if s.Decider == "human" {
			t.Error("decider fell back to human; decider: jev must not fall back")
		}
		if s.Decider != "jev" {
			t.Errorf("decider = %q, want jev (no fallback)", s.Decider)
		}
	})
}

func TestResolveAutonomy_ZeroValueWorkflowTakesDefaults(t *testing.T) {
	s := ResolveAutonomy(WorkflowConfig{})
	if s.Mode != "guided" || s.SecondReview != "required" || s.Decider != "human" || !s.DeciderDerived {
		t.Errorf("zero-value workflow did not resolve to defaults: %+v", s)
	}
	if s.JevMinConfidence != 0.50 || s.BudgetDefault != (AutonomyBudget{Turns: 60, Operations: 40, AuditRetries: 2}) {
		t.Errorf("numeric defaults not applied for absent keys: %+v", s)
	}
}

func TestResolveAutonomy_ExplicitZeroIsNotAbsent(t *testing.T) {
	zeroF, zeroI := 0.0, 0
	var wf WorkflowConfig
	wf.Autonomy.Kickoff.JevMinConfidence = &zeroF
	wf.Autonomy.Escalation.BudgetDefault = AutonomyBudgetConfig{Turns: &zeroI, Operations: &zeroI, AuditRetries: &zeroI}
	s := ResolveAutonomy(wf)
	if s.JevMinConfidence != 0 {
		t.Errorf("explicit jev_min_confidence 0 = %v, want 0 (in range, not absent)", s.JevMinConfidence)
	}
	if s.BudgetDefault != (AutonomyBudget{}) {
		t.Errorf("explicit zero budget = %+v, want 0/0/0", s.BudgetDefault)
	}
	if len(s.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", s.Warnings)
	}
}

func TestResolveAutonomy_JevMinConfidenceRange(t *testing.T) {
	for _, tc := range []struct {
		in   float64
		want float64
		warn bool
	}{
		{0, 0, false}, {1, 1, false}, {0.75, 0.75, false},
		{-0.1, 0.50, true}, {1.0001, 0.50, true},
	} {
		v := tc.in
		var wf WorkflowConfig
		wf.Autonomy.Kickoff.JevMinConfidence = &v
		s := ResolveAutonomy(wf)
		if s.JevMinConfidence != tc.want {
			t.Errorf("in %v: got %v, want %v", tc.in, s.JevMinConfidence, tc.want)
		}
		if got := warningFor(s, "workflow.autonomy.kickoff.jev_min_confidence") != ""; got != tc.warn {
			t.Errorf("in %v: warning emitted = %v, want %v", tc.in, got, tc.warn)
		}
	}
}

func TestResolveAutonomy_DeciderBranches(t *testing.T) {
	for _, tc := range []struct {
		mode, decider, want string
		derived, warn       bool
	}{
		{"", "", "human", true, false},
		{"guided", "", "human", true, false},
		{"contract", "", "llm", true, false},
		{"guided", "llm", "llm", false, false},
		{"guided", "llm+jev", "llm+jev", false, false},
		{"contract", "human", "human", false, false},
		{"contract", "oracle", "human", false, true},
		{"yolo", "", "human", true, false},
	} {
		var wf WorkflowConfig
		wf.Autonomy.Mode = tc.mode
		wf.Autonomy.Kickoff.Decider = tc.decider
		s := ResolveAutonomy(wf)
		if s.Decider != tc.want || s.DeciderDerived != tc.derived {
			t.Errorf("mode %q decider %q: got %q derived=%v, want %q derived=%v",
				tc.mode, tc.decider, s.Decider, s.DeciderDerived, tc.want, tc.derived)
		}
		if got := warningFor(s, "workflow.autonomy.kickoff.decider") != ""; got != tc.warn {
			t.Errorf("mode %q decider %q: decider warning = %v, want %v", tc.mode, tc.decider, got, tc.warn)
		}
	}
}

// TestAutonomy_CacheSchemaBumped guards the config cache hazard recorded on
// configCacheSchemaVersion: a cache written before WorkflowConfig gained
// Autonomy must not be served over a workflow.yaml that sets its keys.
func TestAutonomy_CacheSchemaBumped(t *testing.T) {
	if configCacheSchemaVersion < 7 {
		t.Errorf("configCacheSchemaVersion = %d, want >= 7 — Workflow gained Autonomy without a cache schema bump", configCacheSchemaVersion)
	}
}

func TestResolveAutonomy_SecondReviewSetAndPassThrough(t *testing.T) {
	for _, v := range []string{"required", "advisory", "off"} {
		var wf WorkflowConfig
		wf.Autonomy.Contract.SecondReview = v
		wf.Autonomy.Contract.BatchSign = true
		wf.Autonomy.Contract.PushDevelop = true
		wf.Jev.Enabled = true
		s := ResolveAutonomy(wf)
		if s.SecondReview != v || !s.BatchSign || !s.PushDevelop || !s.JevEnabled || len(s.Warnings) != 0 {
			t.Errorf("second_review %q: %+v", v, s)
		}
	}
}
