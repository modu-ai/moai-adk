package constitution

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeOversight is a test double for the HumanOversight interface that
// auto-approves every proposal. It lets non-dry-run Execute tests run without
// blocking on os.Stdin (NewHumanOversight hardcodes os.Stdin as its reader).
type fakeOversight struct{}

func (fakeOversight) Approve(*AmendmentProposal, bool) (bool, error) { return true, nil }

// rejectingOversight is a test double that always rejects (returns false, nil),
// letting Execute tests exercise the "user rejected the amendment" return path.
type rejectingOversight struct{}

func (rejectingOversight) Approve(*AmendmentProposal, bool) (bool, error) { return false, nil }

// erroringOversight is a test double that returns an error from Approve,
// letting Execute tests exercise the Layer 5 error-return path.
type erroringOversight struct{}

func (erroringOversight) Approve(*AmendmentProposal, bool) (bool, error) {
	return false, errors.New("oversight subsystem unavailable")
}

// writeTestRegistry writes a zone-registry.md fixture containing one Frozen and
// one Evolvable rule under projectDir so that LoadRegistry can parse it during
// Execute integration tests. All paths stay under the caller's t.TempDir()
// (CLAUDE.local.md §6 isolation). The referenced rule files do not need to
// exist on disk: the loader marks missing files orphan (a warning, not fatal).
func writeTestRegistry(t *testing.T, projectDir string) {
	t.Helper()
	registryDir := filepath.Join(projectDir, ".claude", "rules", "moai", "core")
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	content := "# Test Registry\n\n" +
		"```yaml\n" +
		"- id: CONST-V3R2-001\n" +
		"  zone: Frozen\n" +
		"  file: dummy.md\n" +
		"  anchor: \"#frozen\"\n" +
		"  clause: \"TRUST 5 framework\"\n" +
		"  canary_gate: true\n" +
		"- id: CONST-V3R2-003\n" +
		"  zone: Evolvable\n" +
		"  file: dummy.md\n" +
		"  anchor: \"#evolvable\"\n" +
		"  clause: \"Never use time predictions.\"\n" +
		"  canary_gate: false\n" +
		"- id: CONST-V3R2-004\n" +
		"  zone: Evolvable\n" +
		"  file: dummy.md\n" +
		"  anchor: \"#canary\"\n" +
		"  clause: \"Use canary evaluation.\"\n" +
		"  canary_gate: true\n" +
		"```\n"
	registryPath := filepath.Join(registryDir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

func TestNewPipeline(t *testing.T) {
	p := NewPipeline()
	if p == nil {
		t.Fatal("NewPipeline returned nil")
	}
	if p.FrozenGuard == nil || p.Canary == nil || p.ContradictionDetector == nil ||
		p.RateLimiter == nil || p.HumanOversight == nil {
		t.Fatal("NewPipeline: one or more layers not initialized")
	}
}

func TestPipeline_Execute_DryRun_Success(t *testing.T) {
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-003",
		Before: "Never use time predictions.",
		After:  "Never use time predictions in plans.",
	}
	log, err := p.Execute(proposal, dir, true)
	if err != nil {
		t.Fatalf("Execute dry-run failed: %v", err)
	}
	if log == nil {
		t.Fatal("Execute dry-run returned nil log")
	}
	if log.RuleID != proposal.RuleID {
		t.Errorf("log.RuleID = %q, want %q", log.RuleID, proposal.RuleID)
	}
	if log.ZoneBefore != ZoneEvolvable {
		t.Errorf("log.ZoneBefore = %v, want Evolvable", log.ZoneBefore)
	}
	if !proposal.Approved {
		t.Error("Execute did not mark proposal Approved")
	}
}

func TestPipeline_Execute_RuleNotFound(t *testing.T) {
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-999",
		Before: "a",
		After:  "b",
	}
	_, err := p.Execute(proposal, dir, true)
	if err == nil {
		t.Fatal("Execute: expected rule-not-found error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Execute err = %v, want 'not found'", err)
	}
}

func TestPipeline_Execute_FrozenGuard_Rejects(t *testing.T) {
	// A Frozen-zone rule amendment without Evidence is rejected at Layer 1.
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-001",
		Before: "TRUST 5 framework",
		After:  "TRUST 4 framework",
	}
	_, err := p.Execute(proposal, dir, true)
	if err == nil {
		t.Fatal("Execute: expected FrozenGuard rejection")
	}
	if !strings.Contains(err.Error(), "FrozenGuard") {
		t.Errorf("Execute err = %v, want 'FrozenGuard'", err)
	}
}

func TestPipeline_Execute_NonDryRun_AmendmentStubError(t *testing.T) {
	// Non-dry-run Execute reaches applyAmendment, which calls updateSourceFile.
	// updateSourceFile is a documented stub that always errors, so Execute
	// returns the wrapped amendment error. This characterizes the current
	// (stub) behavior without modifying production code (C-3).
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	p.HumanOversight = fakeOversight{}                     // avoid stdin block
	p.LockFilePath = filepath.Join(dir, ".amendment.lock") // pin lock to TempDir (EC-5)
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-003",
		Before: "Never use time predictions.",
		After:  "Never use time predictions in plans.",
	}
	_, err := p.Execute(proposal, dir, false)
	if err == nil {
		t.Fatal("Execute: expected amendment error (updateSourceFile stub)")
	}
	if !strings.Contains(err.Error(), "amendment application error") {
		t.Errorf("Execute err = %v, want 'amendment application error'", err)
	}
	// The deferred releaseLock must have removed the lock file despite the error.
	if _, statErr := os.Stat(p.LockFilePath); !os.IsNotExist(statErr) {
		t.Error("Execute non-dry-run: lock file not released after error")
	}
}

func TestPipeline_Execute_CanaryUnavailable_Continues(t *testing.T) {
	// A rule with canary_gate=true triggers Canary.Evaluate. With no
	// .moai/specs/ directory in the TempDir, Canary returns ErrCanaryUnavailable
	// — which Execute treats as non-fatal and continues to the remaining layers.
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-004",
		Before: "Use canary evaluation.",
		After:  "Use canary evaluation on every amendment.",
	}
	log, err := p.Execute(proposal, dir, true)
	if err != nil {
		t.Fatalf("Execute canary-unavailable should continue: %v", err)
	}
	if log == nil {
		t.Fatal("Execute returned nil log")
	}
	if log.CanaryVerdict != "unavailable" {
		t.Errorf("CanaryVerdict = %q, want unavailable", log.CanaryVerdict)
	}
}

func TestPipeline_Execute_UserRejected(t *testing.T) {
	// When HumanOversight returns (false, nil), Execute returns a
	// "user rejected the amendment" error (Layer 5 rejection).
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	p.HumanOversight = rejectingOversight{}
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-003",
		Before: "Never use time predictions.",
		After:  "Never use time predictions in plans.",
	}
	_, err := p.Execute(proposal, dir, true)
	if err == nil {
		t.Fatal("Execute: expected user-rejected error")
	}
	if !strings.Contains(err.Error(), "user rejected") {
		t.Errorf("Execute err = %v, want 'user rejected'", err)
	}
}

func TestPipeline_Execute_HumanOversightError(t *testing.T) {
	// When HumanOversight returns a non-nil error, Execute wraps it as a
	// Layer 5 failure (distinct from the (false, nil) rejection path).
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	p.HumanOversight = erroringOversight{}
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-003",
		Before: "Never use time predictions.",
		After:  "Never use time predictions in plans.",
	}
	_, err := p.Execute(proposal, dir, true)
	if err == nil {
		t.Fatal("Execute: expected HumanOversight error")
	}
	if !strings.Contains(err.Error(), "HumanOversight") {
		t.Errorf("Execute err = %v, want 'HumanOversight'", err)
	}
}

func TestPipeline_Execute_LockBusy(t *testing.T) {
	// Non-dry-run Execute acquires the single-writer lock; if the lock already
	// exists, acquireLock returns ErrAmendmentInProgress and Execute propagates
	// it (the registry is never loaded). Lock path is pinned to TempDir (EC-5).
	dir := t.TempDir()
	writeTestRegistry(t, dir)
	p := NewPipeline()
	p.LockFilePath = filepath.Join(dir, ".amendment.lock")
	if err := os.WriteFile(p.LockFilePath, []byte("busy"), 0o644); err != nil {
		t.Fatal(err)
	}
	proposal := &AmendmentProposal{
		RuleID: "CONST-V3R2-003",
		Before: "Never use time predictions.",
		After:  "Never use time predictions in plans.",
	}
	_, err := p.Execute(proposal, dir, false)
	if err == nil {
		t.Fatal("Execute: expected lock-busy error")
	}
	if _, ok := err.(*ErrAmendmentInProgress); !ok {
		t.Errorf("Execute err type = %T, want *ErrAmendmentInProgress", err)
	}
}

func TestPipeline_createLogEntry(t *testing.T) {
	p := NewPipeline()
	tests := []struct {
		name              string
		proposal          *AmendmentProposal
		wantCanaryVerdict string
		wantConflictCount int
	}{
		{
			name:              "nil canary result",
			proposal:          &AmendmentProposal{RuleID: "CONST-V3R2-003"},
			wantCanaryVerdict: "skipped",
			wantConflictCount: 0,
		},
		{
			name: "canary unavailable",
			proposal: &AmendmentProposal{
				RuleID:       "R",
				CanaryResult: &CanaryResult{Available: false, Reason: "no specs"},
			},
			wantCanaryVerdict: "unavailable",
			wantConflictCount: 0,
		},
		{
			name: "canary passed",
			proposal: &AmendmentProposal{
				RuleID:       "R",
				CanaryResult: &CanaryResult{Available: true, Passed: true},
			},
			wantCanaryVerdict: "passed",
			wantConflictCount: 0,
		},
		{
			name: "canary rejected",
			proposal: &AmendmentProposal{
				RuleID:       "R",
				CanaryResult: &CanaryResult{Available: true, Passed: false},
			},
			wantCanaryVerdict: "rejected",
			wantConflictCount: 0,
		},
		{
			name: "with contradictions",
			proposal: &AmendmentProposal{
				RuleID: "R",
				Contradicts: &ContradictionResult{
					Conflicts: []ConflictDetail{
						{ConflictingRuleID: "X", Description: "conflict one", IsBlocking: true},
					},
				},
			},
			wantCanaryVerdict: "skipped",
			wantConflictCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := p.createLogEntry(tt.proposal, ZoneEvolvable)
			if log.CanaryVerdict != tt.wantCanaryVerdict {
				t.Errorf("CanaryVerdict = %q, want %q", log.CanaryVerdict, tt.wantCanaryVerdict)
			}
			if len(log.Contradictions) != tt.wantConflictCount {
				t.Errorf("Contradictions len = %d, want %d", len(log.Contradictions), tt.wantConflictCount)
			}
			if log.RuleID != tt.proposal.RuleID {
				t.Errorf("RuleID = %q, want %q", log.RuleID, tt.proposal.RuleID)
			}
		})
	}
}

func TestPipeline_applyAmendment_StubError(t *testing.T) {
	// updateSourceFile is a documented stub returning "not yet implemented",
	// so applyAmendment's only reachable path is the source-file-update error
	// wrap. The registry/evolution-log update lines that follow are unreachable
	// until the stub is implemented; this test characterizes the reachable path.
	p := NewPipeline()
	dir := t.TempDir()
	proposal := &AmendmentProposal{RuleID: "CONST-V3R2-003", Before: "a", After: "b"}
	rule := Rule{
		ID:     "CONST-V3R2-003",
		Zone:   ZoneEvolvable,
		File:   filepath.Join(dir, "dummy.md"),
		Anchor: "#section",
		Clause: "a",
	}
	registryPath := filepath.Join(dir, "zone-registry.md")
	err := p.applyAmendment(proposal, rule, dir, registryPath)
	if err == nil {
		t.Fatal("applyAmendment: expected error (updateSourceFile stub)")
	}
	if !strings.Contains(err.Error(), "source file update error") {
		t.Errorf("applyAmendment err = %v, want 'source file update error'", err)
	}
}

func TestPipeline_acquireLock_DryRunNoOp(t *testing.T) {
	p := NewPipeline()
	if err := p.acquireLock(true); err != nil {
		t.Errorf("acquireLock(dryRun=true) err = %v", err)
	}
	if p.LockFilePath != "" {
		t.Errorf("acquireLock(dryRun=true) set LockFilePath = %q, want empty", p.LockFilePath)
	}
}

func TestPipeline_acquireLock_CreatesAndReleases(t *testing.T) {
	p := NewPipeline()
	p.LockFilePath = filepath.Join(t.TempDir(), "test.lock")
	if err := p.acquireLock(false); err != nil {
		t.Fatalf("acquireLock: %v", err)
	}
	if _, err := os.Stat(p.LockFilePath); err != nil {
		t.Errorf("lock file not created: %v", err)
	}
	p.releaseLock()
	if _, err := os.Stat(p.LockFilePath); !os.IsNotExist(err) {
		t.Error("releaseLock did not remove lock file")
	}
}

func TestPipeline_acquireLock_AlreadyInProgress(t *testing.T) {
	p := NewPipeline()
	lockPath := filepath.Join(t.TempDir(), "existing.lock")
	if err := os.WriteFile(lockPath, []byte("preexisting"), 0o644); err != nil {
		t.Fatal(err)
	}
	p.LockFilePath = lockPath
	err := p.acquireLock(false)
	if err == nil {
		t.Fatal("acquireLock: expected ErrAmendmentInProgress")
	}
	if _, ok := err.(*ErrAmendmentInProgress); !ok {
		t.Errorf("acquireLock err type = %T, want *ErrAmendmentInProgress", err)
	}
}

// updateSourceFile and updateRegistryClause are documented stubs (TODO: not yet
// implemented). These characterization tests pin the current error behavior
// without modifying production code (C-3). When the stubs gain real
// implementations, these tests should be updated to assert success-path behavior.
func TestUpdateSourceFile_StubError(t *testing.T) {
	err := updateSourceFile("dummy.go", "#anchor", "new clause")
	if err == nil {
		t.Fatal("updateSourceFile: expected not-implemented error")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("updateSourceFile err = %v, want 'not yet implemented'", err)
	}
}

func TestUpdateRegistryClause_StubError(t *testing.T) {
	err := updateRegistryClause("zone-registry.md", "CONST-V3R2-001", "new clause")
	if err == nil {
		t.Fatal("updateRegistryClause: expected not-implemented error")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("updateRegistryClause err = %v, want 'not yet implemented'", err)
	}
}

// TestValidationError_Error covers the ValidationError.Error() stringer (used
// for fatal validation failures). Co-located here to lift package coverage.
func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{SentinelKey: "DUPLICATE_ID", Message: "dup found"}
	got := ve.Error()
	if !strings.Contains(got, "DUPLICATE_ID") || !strings.Contains(got, "dup found") {
		t.Errorf("ValidationError.Error() = %q", got)
	}
}

// TestAsValidationError covers the nil-input, matching, and non-matching branches.
func TestAsValidationError(t *testing.T) {
	var target *ValidationError
	if AsValidationError(nil, &target) {
		t.Error("AsValidationError(nil) should return false")
	}
	ve := &ValidationError{SentinelKey: "X", Message: "m"}
	if !AsValidationError(ve, &target) {
		t.Error("AsValidationError(ve) should return true")
	}
	if target != ve {
		t.Error("target not assigned to ve")
	}
	if AsValidationError(errors.New("other error"), &target) {
		t.Error("AsValidationError(non-ValidationError) should return false")
	}
}
