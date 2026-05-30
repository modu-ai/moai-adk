package spec

import (
	"fmt"
	"strings"
	"testing"
)

// TestOwnershipTransitionRule_Pass exercises 7 canonical transitions where the commit
// subject prefix correctly matches the expected owner per the Status Transition Ownership Matrix.
// AC-AAT-005 + REQ-AAT-008 + REQ-AAT-009.
func TestOwnershipTransitionRule_Pass(t *testing.T) {
	tests := []struct {
		name            string
		prevStatus      string
		currStatus      string
		commitSubject   string
		authoredByAgent string
		fmStatus        string
	}{
		{
			name:            "none_to_draft",
			prevStatus:      "",
			currStatus:      "draft",
			commitSubject:   "feat(SPEC-FOO-001): plan-phase artifacts (Tier M Section A-E, 4 artifacts)",
			authoredByAgent: "manager-spec",
			fmStatus:        "draft",
		},
		{
			name:            "draft_to_in_progress",
			prevStatus:      "draft",
			currStatus:      "in-progress",
			commitSubject:   "fix(SPEC-FOO-001): M1 implement core handler",
			authoredByAgent: "manager-develop",
			fmStatus:        "in-progress",
		},
		{
			name:            "in_progress_to_implemented",
			prevStatus:      "in-progress",
			currStatus:      "implemented",
			commitSubject:   "docs(SPEC-FOO-001): sync-phase artifacts",
			authoredByAgent: "manager-docs",
			fmStatus:        "implemented",
		},
		{
			name:            "implemented_to_completed",
			prevStatus:      "implemented",
			currStatus:      "completed",
			commitSubject:   "chore(SPEC-FOO-001): Mx-phase audit-ready signal + 4-phase close",
			authoredByAgent: "orchestrator-direct",
			fmStatus:        "completed",
		},
		{
			name:            "any_to_superseded",
			prevStatus:      "implemented",
			currStatus:      "superseded",
			commitSubject:   "feat(SPEC-BAR-002): supersedes SPEC-FOO-001 with new design",
			authoredByAgent: "manager-spec",
			fmStatus:        "superseded",
		},
		{
			name:            "any_to_archived",
			prevStatus:      "implemented",
			currStatus:      "archived",
			commitSubject:   "chore(specs): archive SPEC-FOO-001 deprecated",
			authoredByAgent: "orchestrator-direct",
			fmStatus:        "archived",
		},
		{
			name:            "any_to_rejected",
			prevStatus:      "draft",
			currStatus:      "rejected",
			commitSubject:   "chore(SPEC-FOO-001): rejected per orchestrator decision",
			authoredByAgent: "orchestrator-direct",
			fmStatus:        "rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
				PreviousStatus:  tt.prevStatus,
				CurrentStatus:   tt.currStatus,
				CommitSubject:   tt.commitSubject,
				CommitSHA:       "abc1234fixture",
				AuthoredByAgent: tt.authoredByAgent,
			}, nil)
			defer restore()

			doc := &SPECDoc{
				Path: ".moai/specs/SPEC-FOO-001/spec.md",
				Frontmatter: SPECFrontmatter{
					ID:     "SPEC-FOO-001",
					Status: tt.fmStatus,
				},
			}

			rule := &OwnershipTransitionRule{}
			findings := rule.Check(doc, nil)
			for _, f := range findings {
				if f.Code == "OwnershipTransitionInvalid" {
					t.Errorf("expected no OwnershipTransitionInvalid finding for canonical transition %q→%q with subject %q; got: %s",
						tt.prevStatus, tt.currStatus, tt.commitSubject, f.Message)
				}
				if f.Code == "OwnershipTransitionUnreachable" {
					t.Errorf("unexpected OwnershipTransitionUnreachable in PASS case: %s", f.Message)
				}
			}
		})
	}
}

// TestOwnershipTransitionRule_Fail exercises 5 ownership-mismatch scenarios
// where the rule MUST emit `OwnershipTransitionInvalid` warning.
// AC-AAT-006 + REQ-AAT-008.
//
// M4 (AC-LSG-004) migration: the WHO signal is now the `Authored-By-Agent:` trailer
// (authoredByAgent field) rather than commit-subject prefix classification.
// Each fixture carries a trailer naming a transition-relevant actor that mismatches
// the matrix-canonical owner.
func TestOwnershipTransitionRule_Fail(t *testing.T) {
	tests := []struct {
		name            string
		prevStatus      string
		currStatus      string
		commitSubject   string
		authoredByAgent string
		fmStatus        string
	}{
		{
			// Forbidden crossing: manager-docs is not allowed to perform draft→in-progress
			// (that's manager-develop's transition per ARR-001 matrix).
			name:            "fail_manager_docs_modifying_run_phase_transition",
			prevStatus:      "draft",
			currStatus:      "in-progress",
			commitSubject:   "docs(SPEC-FOO-001): forbidden — manager-docs editing run-phase",
			authoredByAgent: "manager-docs",
			fmStatus:        "in-progress",
		},
		{
			// Format mismatch: implemented→completed expects orchestrator/manager-docs;
			// manager-develop trailer indicates run-phase agent took over Mx scope.
			name:            "fail_format_mismatch_develop_for_completed",
			prevStatus:      "implemented",
			currStatus:      "completed",
			commitSubject:   "fix(SPEC-FOO-001): M5 mistakenly close-out",
			authoredByAgent: "manager-develop",
			fmStatus:        "completed",
		},
		{
			// in-progress→implemented expects manager-docs; an orchestrator-direct trailer
			// mismatches (one-way alias: manager-docs-expected does NOT accept orchestrator).
			name:            "fail_orchestrator_used_for_sync_phase",
			prevStatus:      "in-progress",
			currStatus:      "implemented",
			commitSubject:   "chore(SPEC-FOO-001): wrong-actor sync attempt",
			authoredByAgent: "orchestrator-direct",
			fmStatus:        "implemented",
		},
		{
			// (none)→draft expects manager-spec but the trailer names manager-develop.
			name:            "fail_manager_develop_creating_spec",
			prevStatus:      "",
			currStatus:      "draft",
			commitSubject:   "fix(SPEC-FOO-001): wrong owner created spec",
			authoredByAgent: "manager-develop",
			fmStatus:        "draft",
		},
		{
			// (any)→superseded expects manager-spec, but the trailer names orchestrator-direct —
			// supersedes is the SPEC author's act, not an admin transition.
			name:            "fail_orchestrator_used_for_supersede",
			prevStatus:      "implemented",
			currStatus:      "superseded",
			commitSubject:   "chore(specs): wrong owner supersedes SPEC-FOO-001",
			authoredByAgent: "orchestrator-direct",
			fmStatus:        "superseded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
				PreviousStatus:  tt.prevStatus,
				CurrentStatus:   tt.currStatus,
				CommitSubject:   tt.commitSubject,
				CommitSHA:       "deadbeeffixture",
				AuthoredByAgent: tt.authoredByAgent,
			}, nil)
			defer restore()

			doc := &SPECDoc{
				Path: ".moai/specs/SPEC-FOO-001/spec.md",
				Frontmatter: SPECFrontmatter{
					ID:     "SPEC-FOO-001",
					Status: tt.fmStatus,
				},
			}

			rule := &OwnershipTransitionRule{}
			findings := rule.Check(doc, nil)

			var hasInvalid bool
			for _, f := range findings {
				if f.Code == "OwnershipTransitionInvalid" {
					hasInvalid = true
					if f.Severity != SeverityWarning {
						t.Errorf("expected severity %s, got %s", SeverityWarning, f.Severity)
					}
				}
			}
			if !hasInvalid {
				t.Errorf("expected OwnershipTransitionInvalid finding for case %q (subject=%q, transition=%q→%q), got none",
					tt.name, tt.commitSubject, tt.prevStatus, tt.currStatus)
			}
		})
	}
}

// TestOwnershipTransitionRule_UnreachableGit verifies graceful degradation when git is unreachable
// (non-git environment or tmpdir without git history). REQ-AAT-010 + REQ-AAT-011 + AC-AAT-007.
//
// Behavior: emit OwnershipTransitionUnreachable Info finding, do NOT panic, do NOT block.
func TestOwnershipTransitionRule_UnreachableGit(t *testing.T) {
	restore := withFakeOwnershipLookup(t, nil, fmt.Errorf("git unreachable: not a git repository"))
	defer restore()

	doc := &SPECDoc{
		Path: "/tmp/non-git-fixture/SPEC-FOO-001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-FOO-001",
			Status: "draft",
		},
	}

	rule := &OwnershipTransitionRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 1 {
		t.Fatalf("expected exactly 1 finding (Info), got %d: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Code != "OwnershipTransitionUnreachable" {
		t.Errorf("expected code OwnershipTransitionUnreachable, got %s", f.Code)
	}
	if f.Severity != SeverityInfo {
		t.Errorf("expected Info severity (no error/warning escalation), got %s", f.Severity)
	}
}

// TestOwnershipTransitionRule_NoTransition verifies that SPECs without a status-transition history
// (newly created file with no prior commit, or status unchanged across all commits) produce no findings.
// This guards against false-positives on stable closed SPECs (e.g., the ARR-001 self-evaluation case
// noted in plan.md §6 + M2 verification step 5).
func TestOwnershipTransitionRule_NoTransition(t *testing.T) {
	restore := withFakeOwnershipLookup(t, nil, nil) // (nil, nil) = graceful no-op (no transition found)
	defer restore()

	doc := &SPECDoc{
		Path: ".moai/specs/SPEC-CLOSED-001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-CLOSED-001",
			Status: "completed",
		},
	}

	rule := &OwnershipTransitionRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("expected zero findings for closed SPEC with no detected transition, got %d: %+v",
			len(findings), findings)
	}
}

// TestOwnershipTransitionRule_EmptyFrontmatter verifies the rule defers to FrontmatterSchemaRule
// when id/status are missing — no findings of its own.
func TestOwnershipTransitionRule_EmptyFrontmatter(t *testing.T) {
	doc := &SPECDoc{
		Path: ".moai/specs/SPEC-FOO-001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "",
			Status: "",
		},
	}
	rule := &OwnershipTransitionRule{}
	findings := rule.Check(doc, nil)
	if len(findings) != 0 {
		t.Errorf("expected zero findings for empty frontmatter, got %d", len(findings))
	}
}

// TestExpectedOwnerForTransition tests the matrix function directly for completeness.
func TestExpectedOwnerForTransition(t *testing.T) {
	tests := []struct {
		prev, curr string
		want       expectedOwnerKind
	}{
		{"", "draft", ownerManagerSpec},
		{"draft", "in-progress", ownerManagerDevelop},
		{"planned", "in-progress", ownerManagerDevelop},
		{"planned", "implemented", ownerManagerDevelop},
		{"draft", "implemented", ownerManagerDevelop},
		{"in-progress", "implemented", ownerManagerDocs},
		{"implemented", "completed", ownerOrchestrator},
		{"implemented", "superseded", ownerManagerSpec},
		{"draft", "archived", ownerOrchestrator},
		{"draft", "rejected", ownerOrchestrator},
		// unmapped (역행 / unknown)
		{"completed", "draft", ownerNone},
		{"draft", "completed", ownerNone},
		{"", "completed", ownerNone},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_to_%s", emptyOrValue(tt.prev), tt.curr), func(t *testing.T) {
			got := expectedOwnerForTransition(tt.prev, tt.curr)
			if got != tt.want {
				t.Errorf("expectedOwnerForTransition(%q, %q) = %v, want %v", tt.prev, tt.curr, got, tt.want)
			}
		})
	}
}

// TestCommitOwnerKind tests the subject classifier directly to lock down the prefix→owner mapping.
func TestCommitOwnerKind(t *testing.T) {
	tests := []struct {
		subject string
		want    expectedOwnerKind
	}{
		{"feat(SPEC-FOO-001): plan-phase artifacts", ownerManagerSpec},
		{"feat(SPEC-BAR-002): supersedes SPEC-FOO-001", ownerManagerSpec},
		{"feat(SPEC-FOO-001): M1 first run-phase commit", ownerManagerDevelop},
		{"fix(SPEC-FOO-001): M2 bugfix", ownerManagerDevelop},
		{"refactor(SPEC-FOO-001): M3 cleanup", ownerManagerDevelop},
		{"perf(SPEC-FOO-001): M4 optimization", ownerManagerDevelop},
		{"test(SPEC-FOO-001): M5 add fixtures", ownerManagerDevelop},
		{"docs(SPEC-FOO-001): sync-phase artifacts", ownerManagerDocs},
		{"chore(SPEC-FOO-001): Mx-phase audit signal + close", ownerOrchestrator},
		{"chore(specs): archive SPEC-FOO-001", ownerOrchestrator},
		{"chore(SPEC-FOO-001): rejected", ownerOrchestrator},
		{"plan(spec): SPEC-FOO-001 ...", ownerManagerSpec},
		// unknown classifications
		{"random unstructured subject", ownerNone},
		{"WIP: SPEC-FOO-001 work in progress", ownerNone},
	}
	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := commitOwnerKind(tt.subject)
			if got != tt.want {
				t.Errorf("commitOwnerKind(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

// TestOwnerMatches verifies the orchestrator/manager-docs alias allowance.
func TestOwnerMatches(t *testing.T) {
	if !ownerMatches(ownerManagerSpec, ownerManagerSpec) {
		t.Error("self-match should be true")
	}
	if !ownerMatches(ownerOrchestrator, ownerManagerDocs) {
		t.Error("orchestrator expected MUST accept manager-docs actual (matrix alias)")
	}
	if ownerMatches(ownerManagerDocs, ownerOrchestrator) {
		t.Error("manager-docs expected MUST NOT accept orchestrator actual (one-way alias)")
	}
	if ownerMatches(ownerManagerDevelop, ownerManagerSpec) {
		t.Error("manager-develop expected MUST NOT accept manager-spec actual")
	}
}

// TestExtractStatusDelta verifies that frontmatter status: diff lines are parsed correctly
// from `git log -p` output. Covers (a) addition-only (new SPEC), (b) modify (status transition),
// (c) no status line (unrelated diff).
func TestExtractStatusDelta(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		wantOld  string
		wantNew  string
		wantFind bool
	}{
		{
			name: "addition_only_new_spec",
			diff: `diff --git a/.moai/specs/SPEC-FOO/spec.md b/.moai/specs/SPEC-FOO/spec.md
new file mode 100644
+++ b/.moai/specs/SPEC-FOO/spec.md
@@ -0,0 +1,5 @@
+---
+id: SPEC-FOO-001
+status: draft
+---
`,
			wantOld:  "",
			wantNew:  "draft",
			wantFind: true,
		},
		{
			name: "modify_transition",
			diff: `diff --git a/.moai/specs/SPEC-FOO/spec.md b/.moai/specs/SPEC-FOO/spec.md
--- a/.moai/specs/SPEC-FOO/spec.md
+++ b/.moai/specs/SPEC-FOO/spec.md
@@ -1,5 +1,5 @@
-status: draft
+status: in-progress
`,
			wantOld:  "draft",
			wantNew:  "in-progress",
			wantFind: true,
		},
		{
			name:     "no_status_line",
			diff:     "diff --git a/README.md b/README.md\n--- a/README.md\n+++ b/README.md\n@@ -1,1 +1,1 @@\n-old text\n+new text\n",
			wantOld:  "",
			wantNew:  "",
			wantFind: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldS, newS, found := extractStatusDelta(tt.diff)
			if found != tt.wantFind {
				t.Errorf("extractStatusDelta found = %v, want %v", found, tt.wantFind)
			}
			if oldS != tt.wantOld {
				t.Errorf("oldStatus = %q, want %q", oldS, tt.wantOld)
			}
			if newS != tt.wantNew {
				t.Errorf("newStatus = %q, want %q", newS, tt.wantNew)
			}
		})
	}
}

// TestSplitGitLogCommits verifies multi-commit parsing of git log --format=%H%x00%s%x00%b%x00 -p output.
// The body field (%b) carries the Authored-By-Agent trailer parsed by parseAuthoredByAgent (M4).
func TestSplitGitLogCommits(t *testing.T) {
	out := "abc1234\x00feat(SPEC-FOO-001): plan-phase\x00\x00\ndiff --git a/spec.md b/spec.md\n+status: draft\n" +
		"def5678\x00fix(SPEC-FOO-001): M1 implement\x00Authored-By-Agent: manager-develop\n\x00\ndiff --git a/spec.md b/spec.md\n-status: draft\n+status: in-progress\n"

	commits := splitGitLogCommits(out)
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].hash != "abc1234" {
		t.Errorf("commit 0 hash = %q, want abc1234", commits[0].hash)
	}
	if commits[0].subject != "feat(SPEC-FOO-001): plan-phase" {
		t.Errorf("commit 0 subject mismatch: %q", commits[0].subject)
	}
	if commits[0].body != "" {
		t.Errorf("commit 0 body = %q, want empty", commits[0].body)
	}
	if commits[1].hash != "def5678" {
		t.Errorf("commit 1 hash = %q, want def5678", commits[1].hash)
	}
	if got := parseAuthoredByAgent(commits[1].body); got != "manager-develop" {
		t.Errorf("commit 1 Authored-By-Agent = %q, want manager-develop", got)
	}
}

// TestParseAuthoredByAgent verifies trailer extraction from commit body (M4 AC-LSG-004).
func TestParseAuthoredByAgent(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"trailer_present_develop", "Some change description\n\nAuthored-By-Agent: manager-develop\n🗿 MoAI <email@mo.ai.kr>", "manager-develop"},
		{"trailer_present_docs", "Authored-By-Agent: manager-docs", "manager-docs"},
		{"trailer_mixed_case_normalized", "Authored-By-Agent: Manager-Develop", "manager-develop"},
		{"trailer_leading_whitespace", "   Authored-By-Agent: orchestrator-direct", "orchestrator-direct"},
		{"trailer_absent_legacy", "Just a regular commit body\nwith no trailer", ""},
		{"empty_body", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseAuthoredByAgent(tt.body); got != tt.want {
				t.Errorf("parseAuthoredByAgent(%q) = %q, want %q", tt.body, got, tt.want)
			}
		})
	}
}

// withFakeOwnershipLookup temporarily replaces the package-level git lookup hook with a fake
// returning the given record + error tuple. The returned restore function MUST be deferred
// (it reverts the hook to the production lookupOwnershipTransitionFromGit).
func withFakeOwnershipLookup(t *testing.T, rec *ownershipTransitionRecord, err error) func() {
	t.Helper()
	prev := getOwnershipTransitionRunner
	getOwnershipTransitionRunner = func(specPath, specID string) (*ownershipTransitionRecord, error) {
		return rec, err
	}
	return func() {
		getOwnershipTransitionRunner = prev
	}
}

// TestOwnershipTransitionDetected exercises the M4 EXTEND feature (AC-LSG-004 / D5 resolution).
//
// The Authored-By-Agent commit-body trailer is the mechanical signal for WHO performed a
// status transition. When a commit subject matches `^(feat|fix|chore)\(SPEC-...\):.*$`
// AND its body contains `Authored-By-Agent: manager-develop` AND the diff shows the
// frontmatter transition `in-progress → implemented`, the rule emits exactly ONE
// OwnershipTransitionInvalid finding referencing the commit SHA — because the matrix
// assigns `in-progress → implemented` to manager-docs, NOT manager-develop.
//
// Negative fixture: same diff WITHOUT the trailer → zero findings (silent SKIP).
func TestOwnershipTransitionDetected(t *testing.T) {
	const commitSHA = "deadbeef1234"

	t.Run("trailer_present_invalid_transition_emits_one_finding", func(t *testing.T) {
		restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
			PreviousStatus:  "in-progress",
			CurrentStatus:   "implemented",
			CommitSubject:   "feat(SPEC-FOO-001): M5 close-out implementation",
			CommitSHA:       commitSHA,
			AuthoredByAgent: "manager-develop",
		}, nil)
		defer restore()

		doc := &SPECDoc{
			Path: ".moai/specs/SPEC-FOO-001/spec.md",
			Frontmatter: SPECFrontmatter{
				ID:     "SPEC-FOO-001",
				Status: "implemented",
			},
		}

		rule := &OwnershipTransitionRule{}
		findings := rule.Check(doc, nil)

		var invalidCount int
		for _, f := range findings {
			if f.Code == "OwnershipTransitionInvalid" {
				invalidCount++
				if f.Severity != SeverityWarning {
					t.Errorf("expected Warning severity, got %s", f.Severity)
				}
				if !strings.Contains(f.Message, commitSHA) {
					t.Errorf("expected finding message to reference commit SHA %q, got: %s", commitSHA, f.Message)
				}
			}
		}
		if invalidCount != 1 {
			t.Errorf("expected exactly 1 OwnershipTransitionInvalid finding, got %d: %+v", invalidCount, findings)
		}
	})

	t.Run("trailer_absent_silent_skip", func(t *testing.T) {
		restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
			PreviousStatus:  "in-progress",
			CurrentStatus:   "implemented",
			CommitSubject:   "feat(SPEC-FOO-001): M5 close-out implementation",
			CommitSHA:       commitSHA,
			AuthoredByAgent: "", // legacy / non-MoAI commit — no trailer
		}, nil)
		defer restore()

		doc := &SPECDoc{
			Path: ".moai/specs/SPEC-FOO-001/spec.md",
			Frontmatter: SPECFrontmatter{
				ID:     "SPEC-FOO-001",
				Status: "implemented",
			},
		}

		rule := &OwnershipTransitionRule{}
		findings := rule.Check(doc, nil)

		for _, f := range findings {
			if f.Code == "OwnershipTransitionInvalid" {
				t.Errorf("expected silent SKIP (zero OwnershipTransitionInvalid) for trailer-less commit, got: %s", f.Message)
			}
		}
	})

	t.Run("trailer_present_valid_owner_no_finding", func(t *testing.T) {
		// manager-docs IS the canonical owner of in-progress → implemented; trailer + matrix agree → no finding.
		restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
			PreviousStatus:  "in-progress",
			CurrentStatus:   "implemented",
			CommitSubject:   "docs(SPEC-FOO-001): sync-phase artifacts",
			CommitSHA:       commitSHA,
			AuthoredByAgent: "manager-docs",
		}, nil)
		defer restore()

		doc := &SPECDoc{
			Path: ".moai/specs/SPEC-FOO-001/spec.md",
			Frontmatter: SPECFrontmatter{
				ID:     "SPEC-FOO-001",
				Status: "implemented",
			},
		}

		rule := &OwnershipTransitionRule{}
		findings := rule.Check(doc, nil)

		for _, f := range findings {
			if f.Code == "OwnershipTransitionInvalid" {
				t.Errorf("expected no finding when trailer agent matches matrix owner, got: %s", f.Message)
			}
		}
	})
}

// TestSkipOptOut exercises the AC-LSG-012 lint.skip opt-out path.
//
// When the SPEC frontmatter contains `lint.skip: [OwnershipTransitionInvalid]`, no
// OwnershipTransitionInvalid finding is emitted; instead an OwnershipTransitionSkipped
// informational (Info severity) finding is emitted.
func TestSkipOptOut(t *testing.T) {
	const commitSHA = "cafef00d5678"

	// Commit signature that would otherwise be INVALID (manager-develop authoring
	// in-progress → implemented, owned by manager-docs).
	restore := withFakeOwnershipLookup(t, &ownershipTransitionRecord{
		PreviousStatus:  "in-progress",
		CurrentStatus:   "implemented",
		CommitSubject:   "feat(SPEC-FOO-001): M5 close-out implementation",
		CommitSHA:       commitSHA,
		AuthoredByAgent: "manager-develop",
	}, nil)
	defer restore()

	doc := &SPECDoc{
		Path: ".moai/specs/SPEC-FOO-001/spec.md",
		Frontmatter: SPECFrontmatter{
			ID:     "SPEC-FOO-001",
			Status: "implemented",
			LintConfig: struct {
				Skip []string `yaml:"skip"`
			}{Skip: []string{"OwnershipTransitionInvalid"}},
		},
	}

	rule := &OwnershipTransitionRule{}
	findings := rule.Check(doc, nil)

	var hasSkipped bool
	for _, f := range findings {
		if f.Code == "OwnershipTransitionInvalid" {
			t.Errorf("expected no OwnershipTransitionInvalid finding when lint.skip opt-out present, got: %s", f.Message)
		}
		if f.Code == "OwnershipTransitionSkipped" {
			hasSkipped = true
			if f.Severity != SeverityInfo {
				t.Errorf("expected Info severity for OwnershipTransitionSkipped, got %s", f.Severity)
			}
		}
	}
	if !hasSkipped {
		t.Errorf("expected an OwnershipTransitionSkipped informational finding when lint.skip opt-out present, got none: %+v", findings)
	}
}
