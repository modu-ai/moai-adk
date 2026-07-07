package constitution

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// TestNewCanary verifies the constructor returns a non-nil Canary.
func TestNewCanary(t *testing.T) {
	c := NewCanary()
	if c == nil {
		t.Fatal("NewCanary returned nil")
	}
}

// mkCompletedSpec creates a SPEC directory with a progress.md file under specsDir.
func mkCompletedSpec(t *testing.T, specsDir, name string) {
	t.Helper()
	dir := filepath.Join(specsDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte("done"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCanary_Evaluate_NoSpecsDir verifies CanaryUnavailable when specs dir is absent.
func TestCanary_Evaluate_NoSpecsDir(t *testing.T) {
	c := NewCanary()
	root := t.TempDir()
	result, err := c.Evaluate(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "MUST do X",
		After:  "MUST do X",
	}, root)
	if err == nil {
		t.Error("Evaluate on absent specs dir should return CanaryUnavailable error")
	}
	if result == nil || result.Available {
		t.Errorf("result should be non-nil and Available=false; got %+v", result)
	}
}

// TestCanary_Evaluate_InsufficientSpecs verifies CanaryUnavailable with < 3 specs.
func TestCanary_Evaluate_InsufficientSpecs(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, ".moai", "specs")
	for _, name := range []string{"SPEC-AAA001", "SPEC-AAA002"} {
		mkCompletedSpec(t, specsDir, name)
	}
	c := NewCanary()
	_, err := c.Evaluate(&AmendmentProposal{Before: "x", After: "x"}, root)
	if err == nil {
		t.Error("Evaluate with < 3 specs should return CanaryUnavailable")
	}
}

// TestCanary_Evaluate_Passed verifies the success path with ≥ 3 specs and low impact.
func TestCanary_Evaluate_Passed(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, ".moai", "specs")
	for _, name := range []string{"SPEC-AAA001", "SPEC-AAA002", "SPEC-AAA003"} {
		mkCompletedSpec(t, specsDir, name)
	}
	c := NewCanary()
	result, err := c.Evaluate(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "MUST do X",
		After:  "MUST do X",
	}, root)
	if err != nil {
		t.Fatalf("Evaluate should pass with low impact; got err %v", err)
	}
	if !result.Available {
		t.Error("result should be Available with ≥ 3 specs")
	}
	if !result.Passed {
		t.Errorf("result should Pass with neutral change; MaxDrop=%.2f", result.MaxDrop)
	}
}

// TestCanary_Evaluate_Rejected verifies CanaryRejected when score drop exceeds threshold.
func TestCanary_Evaluate_Rejected(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, ".moai", "specs")
	for _, name := range []string{"SPEC-AAA001", "SPEC-AAA002", "SPEC-AAA003"} {
		mkCompletedSpec(t, specsDir, name)
	}
	c := NewCanary()
	// Before is short; After is much longer + prohibition words → score drop > 0.10.
	_, err := c.Evaluate(&AmendmentProposal{
		RuleID: "CONST-V3R2-099",
		Before: "abc",
		After:  "MUST NOT do something with many words to ensure a large score drop here",
	}, root)
	if err == nil {
		t.Fatal("Evaluate should return CanaryRejected for large score drop")
	}
}

// TestCanary_estimateScoreImpact table-drives the score-impact heuristic.
func TestCanary_estimateScoreImpact(t *testing.T) {
	c := &canary{completedSpecPattern: regexp.MustCompile(`^SPEC-[A-Z0-9]+$`)}
	tests := []struct {
		name    string
		before  string
		after   string
		wantMin float64
		wantMax float64
	}{
		{"neutral same length", "MUST do X", "MUST do X", 0.99, 1.0},
		{"longer stricter", "abc", "MUST NOT do many more words here for length", 0.80, 0.95},
		{"shorter relaxed", "MUST do a lot of mandatory things here", "x", 1.0, 1.02},
		{"lose MUST adds MAY", "MUST required", "MAY optional", 1.0, 1.06},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := c.estimateScoreImpact(&AmendmentProposal{Before: tt.before, After: tt.after})
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("estimateScoreImpact(%q→%q) = %.3f; want [%.2f, %.2f]", tt.before, tt.after, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestCanary_sortMostRecent verifies descending modtime sort + limit.
func TestCanary_sortMostRecent(t *testing.T) {
	c := &canary{completedSpecPattern: regexp.MustCompile(`^SPEC-[A-Z0-9]+$`)}
	root := t.TempDir()
	specsDir := filepath.Join(root, ".moai", "specs")
	names := []string{"SPEC-OLD-001", "SPEC-NEW-001", "SPEC-MID-001"}
	base := time.Now().Add(-time.Hour)
	for i, name := range names {
		dir := filepath.Join(specsDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "progress.md")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	got := c.sortMostRecent(names, specsDir, 2)
	if len(got) != 2 {
		t.Fatalf("sortMostRecent limit=2: got %d, want 2", len(got))
	}
	// Most recent (highest mtime = last in creation order) should come first.
	if got[0] != "SPEC-MID-001" {
		t.Errorf("sortMostRecent[0]: got %s, want SPEC-MID-001 (most recent)", got[0])
	}
}

// TestParseScoreFromProgress verifies score extraction from progress.md content.
func TestParseScoreFromProgress(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "progress.md")
	if err := os.WriteFile(path, []byte("... sync-auditor Score: 0.87 ...\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	score, err := parseScoreFromProgress(path)
	if err != nil {
		t.Fatalf("parseScoreFromProgress: %v", err)
	}
	if score < 0.86 || score > 0.88 {
		t.Errorf("score: got %.2f, want ~0.87", score)
	}
}

// TestParseScoreFromProgress_NoMatch verifies the default fallback when no Score line exists.
func TestParseScoreFromProgress_NoMatch(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "progress.md")
	if err := os.WriteFile(path, []byte("no score line here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	score, err := parseScoreFromProgress(path)
	if err != nil {
		t.Fatalf("parseScoreFromProgress no-match: %v", err)
	}
	if score != 0.8 {
		t.Errorf("default score on no match: got %.2f, want 0.8", score)
	}
}

// --- Additional constitution coverage (layer constructors + error types) ---
// These cover trivial constructors and error .Error() methods that were at 0%
// and drag the package average below the 85% floor. All tests-only (C-3).

func TestNewFrozenGuard_Check(t *testing.T) {
	g := NewFrozenGuard()
	if g == nil {
		t.Fatal("NewFrozenGuard returned nil")
	}
	// Evolvable zone always passes.
	if err := g.Check(&AmendmentProposal{RuleID: "CONST-V3R2-099"}, ZoneEvolvable); err != nil {
		t.Errorf("Check on Evolvable should pass; got %v", err)
	}
	// Frozen without Evidence → error.
	if err := g.Check(&AmendmentProposal{RuleID: "CONST-V3R2-099"}, ZoneFrozen); err == nil {
		t.Error("Check on Frozen without Evidence should error")
	}
	// Frozen with Evidence → allowed.
	if err := g.Check(&AmendmentProposal{RuleID: "CONST-V3R2-099", Evidence: "demotion reason"}, ZoneFrozen); err != nil {
		t.Errorf("Check on Frozen with Evidence should pass; got %v", err)
	}
}

func TestNewRateLimiter(t *testing.T) {
	l := NewRateLimiter()
	if l == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	// Admit with no evolution log → passes (no history to limit against).
	if err := l.Admit(&AmendmentProposal{RuleID: "CONST-V3R2-099"}, filepath.Join(t.TempDir(), "evolution-log.md")); err != nil {
		t.Errorf("Admit with no log: %v", err)
	}
}

func TestNewHumanOversight(t *testing.T) {
	h := NewHumanOversight()
	if h == nil {
		t.Fatal("NewHumanOversight returned nil")
	}
	// Dry-run mode always approves without prompting.
 approved, err := h.Approve(&AmendmentProposal{RuleID: "CONST-V3R2-099", Before: "a", After: "b"}, true)
	if err != nil {
		t.Errorf("Approve dry-run: %v", err)
	}
	if !approved {
		t.Error("Approve dry-run should return true")
	}
}

func TestAmendmentErrorStrings(t *testing.T) {
	// Exercise every error .Error() method for coverage.
	cases := []struct {
		name string
		err  error
	}{
		{"ErrFrozenAmendment", &ErrFrozenAmendment{RuleID: "R", Reason: "x"}},
		{"ErrCanaryUnavailable", &ErrCanaryUnavailable{RequiredCount: 3, ActualCount: 1}},
		{"ErrCanaryRejected", &ErrCanaryRejected{RuleID: "R", ScoreDrop: 0.5, Threshold: 0.1, AffectedSpecs: []string{"S"}}},
		{"ErrContradictionDetected", &ErrContradictionDetected{NewRuleID: "R", ConflictingIDs: []string{"C"}, Conflicts: []string{"d"}}},
		{"ErrRateLimitExceeded", &ErrRateLimitExceeded{MaxPerWeek: 3, CooldownHours: 24}},
		{"ErrAmendmentInProgress", &ErrAmendmentInProgress{LockFilePath: "/tmp/lock"}},
		{"ErrRolledBack", &ErrRolledBack{RuleID: "R", CooldownDays: 30}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := c.err.Error()
			if s == "" {
				t.Errorf("%s.Error() returned empty string", c.name)
			}
		})
	}
}
