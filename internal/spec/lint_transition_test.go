package spec

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// transitionFixture describes a temporary git repository built for one
// status-transition case: a SPEC whose history performs exactly one
// `status:` transition (or, for the singleCommit shape, none at all).
type transitionFixture struct {
	// from is the status written by the first commit. Ignored when singleCommit.
	from string
	// to is the status written by the second commit (or the only commit when
	// singleCommit is set).
	to string
	// trailer, when non-empty, is the `Authored-By-Agent:` value on the
	// transition commit. Empty means the commit carries no trailer at all.
	trailer string
	// modernEra writes a progress.md carrying the V3R6 H-4 signals, so the
	// SPEC is not grandfather-protected (era.go ClassifyEra).
	modernEra bool
	// singleCommit lands spec.md in one commit, producing the `(none) → X`
	// shape the extractor yields for an added `status:` line with no removed
	// counterpart (spec.md §A.5 D5).
	singleCommit bool
	// uncommitted leaves spec.md untracked, so git history records no
	// transition at all.
	uncommitted bool
	// noGit skips `git init` entirely.
	noGit bool
	// outOfScopeHeading writes a conforming `### Out of Scope —` heading so the
	// fixture does not also trip MissingExclusions.
	outOfScopeHeading bool
}

// builtFixture is what buildTransitionFixture hands back to a case.
type builtFixture struct {
	repo       string
	specDir    string
	specPath   string
	transition string // SHA of the commit that performed the transition
}

const transitionFixtureID = "SPEC-FIXTURE-TRANSITION-001"

func transitionSpecBody(status string, conformingOutOfScope bool) string {
	outOfScope := "## Out of Scope\n\n- nothing beyond the fixture\n"
	if conformingOutOfScope {
		outOfScope = "### Out of Scope — everything not named above\n\n- nothing beyond the fixture\n"
	}
	return "---\n" +
		"id: " + transitionFixtureID + "\n" +
		"title: \"fixture - status transition\"\n" +
		"version: \"0.1.0\"\n" +
		"status: " + status + "\n" +
		"created: 2026-08-31\n" +
		"updated: 2026-08-31\n" +
		"author: manager-spec\n" +
		"priority: P2\n" +
		"phase: \"v3.2.0\"\n" +
		"module: \"fixture\"\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"fixture\"\n" +
		"tier: S\n" +
		"---\n\n## HISTORY\n\n- **2026-08-31** - fixture.\n\n" + outOfScope
}

func buildTransitionFixture(t *testing.T, f transitionFixture) builtFixture {
	t.Helper()

	repo := t.TempDir()
	specDir := filepath.Join(repo, ".moai", "specs", transitionFixtureID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(specDir, "spec.md")

	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	write := func(status string) {
		t.Helper()
		if err := os.WriteFile(specPath, []byte(transitionSpecBody(status, f.outOfScopeHeading)), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if f.modernEra {
		progress := "# Progress\n\n" +
			"## §E.2 Run-phase Evidence\n\n- evidence\n\n" +
			"## §E.4 Sync-phase Audit-Ready Signal\n\n" +
			"sync_commit_sha: \"a1b2c3d4e5f6\"\n"
		if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progress), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if f.noGit {
		write(f.to)
		return builtFixture{repo: repo, specDir: specDir, specPath: specPath}
	}

	git("init", "-q", "-b", "main")

	if f.uncommitted {
		// A tracked placeholder so the repo has at least one commit, then the
		// SPEC left untracked: git history records no `status:` line at all.
		if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("placeholder\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		git("add", "README.md")
		git("commit", "-q", "-m", "chore: placeholder")
		write(f.to)
		return builtFixture{repo: repo, specDir: specDir, specPath: specPath}
	}

	if f.singleCommit {
		write(f.to)
		git("add", ".")
		git("commit", "-q", "-m", "feat("+transitionFixtureID+"): plan-phase fixture")
		return builtFixture{repo: repo, specDir: specDir, specPath: specPath, transition: git("rev-parse", "HEAD")}
	}

	write(f.from)
	git("add", ".")
	git("commit", "-q", "-m", "feat("+transitionFixtureID+"): plan-phase fixture\n\nAuthored-By-Agent: manager-spec")

	write(f.to)
	git("add", ".")
	msg := "chore(" + transitionFixtureID + "): move status"
	if f.trailer != "" {
		msg += "\n\nAuthored-By-Agent: " + f.trailer
	}
	git("commit", "-q", "-m", msg)

	return builtFixture{repo: repo, specDir: specDir, specPath: specPath, transition: git("rev-parse", "HEAD")}
}

// lintFixture runs the linter over a built fixture from inside its repository.
func lintFixture(t *testing.T, b builtFixture, strict bool) *Report {
	t.Helper()
	t.Chdir(b.repo)
	linter := NewLinter(LinterOptions{BaseDir: b.specDir, Strict: strict})
	report, err := linter.Lint([]string{b.specPath})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	return report
}

func findingsWithCode(report *Report, code string) []Finding {
	var out []Finding
	for _, f := range report.Findings {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

// TestStatusTransitionValidityRule is the bidirectional regression set for
// SPEC-STATUS-TRANSITION-VALIDITY-001 (card t376). Every case below runs in
// ONE execution, so an all-silent run is distinguishable from a harness that
// stopped working: the wantInvalid cases are the live control (AC-STV-010).
//
// Covers AC-STV-001..AC-STV-009, AC-STV-014, AC-STV-015, AC-STV-017.
func TestStatusTransitionValidityRule(t *testing.T) {
	cases := []struct {
		name    string
		fixture transitionFixture
		// wantInvalid: a StatusTransitionInvalid finding must be emitted.
		wantInvalid bool
		// wantToken: when non-empty, a StatusTokenUnrecognized finding naming
		// this token must be emitted, and StatusTransitionInvalid must NOT be.
		wantToken string
		// ac names the acceptance criterion the case discharges.
		ac string
	}{
		// ---- MUST FIRE (the live control of AC-STV-010) ----
		{
			name:        "draft_to_completed_is_caught",
			fixture:     transitionFixture{from: "draft", to: "completed", trailer: "manager-develop"},
			wantInvalid: true,
			ac:          "AC-STV-001",
		},
		{
			name:        "completed_to_draft_is_caught",
			fixture:     transitionFixture{from: "completed", to: "draft", trailer: "manager-spec"},
			wantInvalid: true,
			ac:          "AC-STV-002",
		},
		{
			name:        "completed_to_implemented_reversal_is_caught",
			fixture:     transitionFixture{from: "completed", to: "implemented", trailer: "manager-docs"},
			wantInvalid: true,
			ac:          "AC-STV-014",
		},
		{
			// The property that separates this rule from OwnershipTransitionRule,
			// whose measured behavior on a trailer-less commit is a silent skip.
			name:        "draft_to_completed_without_trailer_is_caught",
			fixture:     transitionFixture{from: "draft", to: "completed"},
			wantInvalid: true,
			ac:          "AC-STV-003",
		},

		// ---- MUST STAY SILENT (the canonical set of spec.md §A.7) ----
		{
			name:    "implemented_to_completed_right_owner_passes",
			fixture: transitionFixture{from: "implemented", to: "completed", trailer: "manager-docs"},
			ac:      "AC-STV-004",
		},
		{
			name:    "draft_to_in_progress_passes",
			fixture: transitionFixture{from: "draft", to: "in-progress", trailer: "manager-develop"},
			ac:      "AC-STV-005",
		},
		{
			name:    "in_progress_to_implemented_passes",
			fixture: transitionFixture{from: "in-progress", to: "implemented", trailer: "manager-docs"},
			ac:      "AC-STV-006",
		},
		{
			name:    "completed_to_in_progress_amendment_passes",
			fixture: transitionFixture{from: "completed", to: "in-progress", trailer: "manager-develop"},
			ac:      "AC-STV-007",
		},
		{
			// 217 census occurrences — the single-sync-commit close (§A.5 D4).
			name:    "in_progress_to_completed_single_sync_close_passes",
			fixture: transitionFixture{from: "in-progress", to: "completed", trailer: "manager-docs"},
			ac:      "AC-STV-007a",
		},
		{
			// 50 census occurrences — adopted from lint_ownership.go (§A.5 D1).
			name:    "draft_to_implemented_passes",
			fixture: transitionFixture{from: "draft", to: "implemented", trailer: "manager-develop"},
			ac:      "AC-STV-007a",
		},
		{
			name:    "planned_on_the_left_is_tolerated",
			fixture: transitionFixture{from: "planned", to: "completed", trailer: "manager-docs"},
			ac:      "AC-STV-008",
		},
		{
			name:    "planned_on_the_right_is_tolerated",
			fixture: transitionFixture{from: "in-progress", to: "planned", trailer: "manager-spec"},
			ac:      "AC-STV-008",
		},
		{
			name:    "terminal_superseded_target_passes",
			fixture: transitionFixture{from: "draft", to: "superseded", trailer: "manager-spec"},
			ac:      "AC-STV-005",
		},
		{
			name:    "terminal_archived_target_passes",
			fixture: transitionFixture{from: "in-progress", to: "archived", trailer: "manager-docs"},
			ac:      "AC-STV-005",
		},
		{
			name:    "terminal_rejected_target_passes",
			fixture: transitionFixture{from: "draft", to: "rejected", trailer: "manager-docs"},
			ac:      "AC-STV-005",
		},
		{
			// Module convention (internal/spec/CLAUDE.md): a new rule must not
			// false-flag a sibling SPEC closed before the rule shipped.
			name:    "closed_sibling_spec_is_not_false_flagged",
			fixture: transitionFixture{from: "in-progress", to: "implemented", trailer: "manager-docs", modernEra: true},
			ac:      "AC-STV-006",
		},

		// ---- SKIPPED SHAPES ----
		{
			// 136 of 713 census records carry the `(none)` shape, 104 of them
			// targeting completed (§A.5 D5).
			name:    "none_to_completed_is_skipped",
			fixture: transitionFixture{to: "completed", singleCommit: true},
			ac:      "AC-STV-017",
		},
		{
			name:    "none_to_draft_is_skipped",
			fixture: transitionFixture{to: "draft", singleCommit: true},
			ac:      "AC-STV-017",
		},
		{
			name:    "no_transition_in_history_is_silent",
			fixture: transitionFixture{to: "draft", uncommitted: true},
			ac:      "AC-STV-009",
		},
		{
			name:    "non_git_directory_is_silent",
			fixture: transitionFixture{to: "completed", noGit: true},
			ac:      "AC-STV-009",
		},

		// ---- UNRECOGNIZED TOKENS (real values from the corpus census) ----
		{
			name:      "synced_token_fires_its_own_code",
			fixture:   transitionFixture{from: "synced", to: "completed", trailer: "manager-docs"},
			wantToken: "synced",
			ac:        "AC-STV-015",
		},
		{
			name:      "approved_token_fires_its_own_code",
			fixture:   transitionFixture{from: "approved", to: "completed", trailer: "manager-docs"},
			wantToken: "approved",
			ac:        "AC-STV-015",
		},
		{
			name:      "cancelled_token_fires_its_own_code",
			fixture:   transitionFixture{from: "cancelled", to: "rejected", trailer: "manager-docs"},
			wantToken: "cancelled",
			ac:        "AC-STV-015",
		},
		{
			name:      "case_variant_token_fires_its_own_code",
			fixture:   transitionFixture{from: "Completed", to: "completed", trailer: "manager-docs"},
			wantToken: "Completed",
			ac:        "AC-STV-015",
		},
	}

	// AC-STV-010: the live control. At least one case in THIS run must have
	// asserted a finding that actually fired.
	firedAtLeastOnce := false

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := buildTransitionFixture(t, tc.fixture)
			report := lintFixture(t, b, false)

			invalid := findingsWithCode(report, "StatusTransitionInvalid")
			token := findingsWithCode(report, "StatusTokenUnrecognized")

			switch {
			case tc.wantInvalid:
				if len(invalid) != 1 {
					t.Fatalf("%s: want exactly 1 StatusTransitionInvalid, got %d: %+v", tc.ac, len(invalid), invalid)
				}
				msg := invalid[0].Message
				for _, want := range []string{tc.fixture.from, tc.fixture.to, b.transition} {
					if !strings.Contains(msg, want) {
						t.Errorf("%s: message %q does not name %q", tc.ac, msg, want)
					}
				}
				if invalid[0].Severity != SeverityWarning {
					t.Errorf("%s: want warning severity, got %v", tc.ac, invalid[0].Severity)
				}
				if len(token) != 0 {
					t.Errorf("%s: unexpected StatusTokenUnrecognized: %+v", tc.ac, token)
				}
				firedAtLeastOnce = true

			case tc.wantToken != "":
				if len(token) != 1 {
					t.Fatalf("%s: want exactly 1 StatusTokenUnrecognized, got %d: %+v", tc.ac, len(token), token)
				}
				if !strings.Contains(token[0].Message, tc.wantToken) {
					t.Errorf("%s: message %q does not name token %q", tc.ac, token[0].Message, tc.wantToken)
				}
				// The load-bearing half: the new code must not become a second
				// spelling of the first.
				if len(invalid) != 0 {
					t.Errorf("%s: token case also emitted StatusTransitionInvalid: %+v", tc.ac, invalid)
				}
				firedAtLeastOnce = true

			default:
				if len(invalid) != 0 {
					t.Errorf("%s: want no StatusTransitionInvalid, got %+v", tc.ac, invalid)
				}
				if len(token) != 0 {
					t.Errorf("%s: want no StatusTokenUnrecognized, got %+v", tc.ac, token)
				}
			}

			// AC-STV-009: neither new code is ever emitted at error severity.
			for _, f := range append(invalid, token...) {
				if f.Severity == SeverityError {
					t.Errorf("%s: %s emitted at error severity: %+v", tc.ac, f.Code, f)
				}
			}
		})
	}

	// AC-STV-010: a run in which every assertion is "no finding" cannot
	// distinguish a correct implementation from a rule that was never wired in.
	if !firedAtLeastOnce {
		t.Fatal("AC-STV-010 live control: no case in this run produced a finding — " +
			"the harness, not the corpus, is what this failure is about")
	}
}

// TestStatusTransitionTrailerIndependence discharges AC-STV-003 directly: two
// repositories identical except for the `Authored-By-Agent:` trailer on the
// transition commit must yield findings that differ in no field other than the
// commit SHA and the file path.
func TestStatusTransitionTrailerIndependence(t *testing.T) {
	get := func(trailer string) (Finding, builtFixture) {
		t.Helper()
		b := buildTransitionFixture(t, transitionFixture{from: "draft", to: "completed", trailer: trailer})
		report := lintFixture(t, b, false)
		fs := findingsWithCode(report, "StatusTransitionInvalid")
		if len(fs) != 1 {
			t.Fatalf("trailer=%q: want exactly 1 StatusTransitionInvalid, got %d: %+v", trailer, len(fs), fs)
		}
		return fs[0], b
	}

	withTrailer, bWith := get("manager-develop")
	without, bWithout := get("")

	if withTrailer.Severity != without.Severity {
		t.Errorf("AC-STV-003: severity differs: %v vs %v", withTrailer.Severity, without.Severity)
	}
	if withTrailer.Advisory != without.Advisory {
		t.Errorf("AC-STV-003: advisory differs: %v vs %v", withTrailer.Advisory, without.Advisory)
	}
	if withTrailer.Code != without.Code {
		t.Errorf("AC-STV-003: code differs: %q vs %q", withTrailer.Code, without.Code)
	}
	if withTrailer.Line != without.Line {
		t.Errorf("AC-STV-003: line differs: %d vs %d", withTrailer.Line, without.Line)
	}

	// Normalize the two permitted differences — the commit SHA and the path —
	// and require the messages to be identical thereafter.
	normalize := func(f Finding, b builtFixture) string {
		m := strings.ReplaceAll(f.Message, b.transition, "<SHA>")
		return strings.ReplaceAll(m, b.specPath, "<PATH>")
	}
	if a, c := normalize(withTrailer, bWith), normalize(without, bWithout); a != c {
		t.Errorf("AC-STV-003: messages differ beyond SHA and path:\n with trailer: %s\n without:      %s", a, c)
	}
}

// TestStatusTransitionFindingGates discharges AC-STV-018: on a modern-era SPEC
// whose frontmatter status is not terminal, the finding carries advisory=false
// and HasErrors() reports true under --strict.
func TestStatusTransitionFindingGates(t *testing.T) {
	// completed → implemented: invalid, and `implemented` is NOT in
	// terminalStatusEnum, so applyEraDemotion's second disjunct is false too.
	b := buildTransitionFixture(t, transitionFixture{
		from:              "completed",
		to:                "implemented",
		trailer:           "manager-docs",
		modernEra:         true,
		outOfScopeHeading: true,
	})
	report := lintFixture(t, b, true)

	fs := findingsWithCode(report, "StatusTransitionInvalid")
	if len(fs) != 1 {
		t.Fatalf("AC-STV-018: want exactly 1 StatusTransitionInvalid, got %d: %+v", len(fs), fs)
	}
	if fs[0].Advisory {
		t.Errorf("AC-STV-018: finding is advisory — it reports but cannot gate: %+v", fs[0])
	}
	if !report.HasErrors() {
		t.Errorf("AC-STV-018: HasErrors() is false under --strict despite a non-advisory warning")
	}
}

// TestStatusTransitionCheckOrder pins the check order down rather than letting
// it fall out of statement sequence (plan.md M1). The order is:
//
//	(none) skip  →  token recognition  →  pair validity
//
// The quote-wrapped `(none) → "in-progress"` row of the corpus census is the
// document whose reported code depends entirely on this order.
func TestStatusTransitionCheckOrder(t *testing.T) {
	t.Run("none_skip_precedes_token_check", func(t *testing.T) {
		// An unrecognized token reached only through the `(none)` shape must be
		// skipped, not reported: the extractor cannot tell creation from
		// truncation, so the pair is not evidence of anything.
		b := buildTransitionFixture(t, transitionFixture{to: "synced", singleCommit: true})
		report := lintFixture(t, b, false)
		if got := findingsWithCode(report, "StatusTokenUnrecognized"); len(got) != 0 {
			t.Errorf("(none) skip must precede the token check, got: %+v", got)
		}
		if got := findingsWithCode(report, "StatusTransitionInvalid"); len(got) != 0 {
			t.Errorf("(none) skip must precede the pair check, got: %+v", got)
		}
	})

	t.Run("token_check_precedes_pair_check", func(t *testing.T) {
		// `cancelled → draft` is both an unrecognized token AND an
		// out-of-set pair. The token check runs first, so exactly one finding
		// is emitted and it names the token.
		b := buildTransitionFixture(t, transitionFixture{from: "cancelled", to: "draft", trailer: "manager-spec"})
		report := lintFixture(t, b, false)
		if got := findingsWithCode(report, "StatusTokenUnrecognized"); len(got) != 1 {
			t.Errorf("want 1 StatusTokenUnrecognized, got %d: %+v", len(got), got)
		}
		if got := findingsWithCode(report, "StatusTransitionInvalid"); len(got) != 0 {
			t.Errorf("token check must precede the pair check, got: %+v", got)
		}
	})
}

// TestStatusTransitionRuleGuards covers the rule's skip guards at the unit
// seam, using the same injected-lookup hook the sibling ownership rule's tests
// use. Each case is a shape the corpus actually contains.
func TestStatusTransitionRuleGuards(t *testing.T) {
	rule := &StatusTransitionValidityRule{}

	t.Run("closed_sibling_with_no_detected_transition", func(t *testing.T) {
		// Module convention (internal/spec/CLAUDE.md): a new rule must not
		// false-flag a sibling SPEC closed before the rule shipped. Mirrors
		// TestOwnershipTransitionRule_NoTransition.
		defer withFakeOwnershipLookup(t, nil, nil)()
		doc := &SPECDoc{
			Path:        ".moai/specs/SPEC-CLOSED-001/spec.md",
			Frontmatter: SPECFrontmatter{ID: "SPEC-CLOSED-001", Status: "completed"},
		}
		if got := rule.Check(doc, nil); len(got) != 0 {
			t.Errorf("want zero findings for a closed SPEC with no detected transition, got %+v", got)
		}
	})

	t.Run("git_unreachable_is_silent", func(t *testing.T) {
		// REQ-STV-007. OwnershipTransitionRule already reports unreachable git
		// as an Info finding; this rule must not report the same fact twice.
		defer withFakeOwnershipLookup(t, nil, errors.New("git unreachable"))()
		doc := &SPECDoc{
			Path:        ".moai/specs/SPEC-NOGIT-001/spec.md",
			Frontmatter: SPECFrontmatter{ID: "SPEC-NOGIT-001", Status: "draft"},
		}
		if got := rule.Check(doc, nil); len(got) != 0 {
			t.Errorf("want zero findings when git is unreachable, got %+v", got)
		}
	})

	t.Run("unreadable_frontmatter_id_defers_to_frontmatter_rule", func(t *testing.T) {
		// Measured on the live corpus: SPEC-V3R6-LINK-FIX-001 carries the
		// rejected snake_case alias `spec_id:` instead of `id:`, so the YAML
		// decoder yields an empty ID. Its git history DOES record
		// `draft → completed`, which is why the corpus census counts 50 of that
		// edge while the linter reports 49 — this document is the difference.
		//
		// Skipping it is the correct behavior, not a gap to close here:
		// FrontmatterInvalid already reports the missing `id`, and naming a
		// transition on a SPEC whose identity cannot be read would report a
		// second fact resting on a broken premise.
		defer withFakeOwnershipLookup(t, &ownershipTransitionRecord{
			PreviousStatus: "draft",
			CurrentStatus:  "completed",
			CommitSHA:      "deadbeef",
		}, nil)()
		doc := &SPECDoc{
			Path:        ".moai/specs/SPEC-ALIASED-ID-001/spec.md",
			Frontmatter: SPECFrontmatter{ID: "", Status: "completed"},
		}
		if got := rule.Check(doc, nil); len(got) != 0 {
			t.Errorf("want zero findings when the frontmatter ID is unreadable, got %+v", got)
		}
	})
}

// TestIsCanonicalStatusTransition exercises the edge set directly, so a change
// to the table is visible as a test change rather than only as a corpus delta.
func TestIsCanonicalStatusTransition(t *testing.T) {
	tests := []struct {
		prev, curr string
		want       bool
		why        string
	}{
		// canonical (spec.md §A.7)
		{"draft", "in-progress", true, "matrix row 2"},
		{"in-progress", "implemented", true, "matrix row 3"},
		{"implemented", "completed", true, "matrix row 3"},
		{"in-progress", "completed", true, "single-sync-commit close (D4)"},
		{"completed", "in-progress", true, "declared amendment (matrix row 7)"},
		{"draft", "implemented", true, "adopted from lint_ownership.go (D1)"},
		// terminal targets — * → X (matrix rows 4-6)
		{"draft", "superseded", true, "* → superseded"},
		{"completed", "archived", true, "* → archived"},
		{"in-progress", "rejected", true, "* → rejected"},
		// planned — legacy-optional, no active-flow owner
		{"planned", "completed", true, "planned tolerated on the left"},
		{"draft", "planned", true, "planned tolerated on the right"},
		// invalid
		{"draft", "completed", false, "skips both run and sync"},
		{"completed", "draft", false, "reversal"},
		{"completed", "implemented", false, "reversal, 48 census occurrences"},
		{"implemented", "in-progress", false, "reversal"},
		{"implemented", "draft", false, "reversal"},
		{"draft", "draft", false, "self-edge is not a transition"},
	}

	for _, tt := range tests {
		t.Run(tt.prev+"_to_"+tt.curr, func(t *testing.T) {
			if got := isCanonicalStatusTransition(tt.prev, tt.curr); got != tt.want {
				t.Errorf("isCanonicalStatusTransition(%q, %q) = %v, want %v (%s)",
					tt.prev, tt.curr, got, tt.want, tt.why)
			}
		})
	}
}

// TestStatusTransitionRuleIsObservationOnly discharges the first half of
// AC-STV-012: a lint run leaves the working tree byte-identical.
func TestStatusTransitionRuleIsObservationOnly(t *testing.T) {
	b := buildTransitionFixture(t, transitionFixture{from: "draft", to: "completed", trailer: "manager-develop"})

	porcelain := func() string {
		t.Helper()
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = b.repo
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("git status: %v", err)
		}
		return string(out)
	}

	before := porcelain()
	report := lintFixture(t, b, false)
	if len(findingsWithCode(report, "StatusTransitionInvalid")) != 1 {
		t.Fatal("AC-STV-012: control case did not fire — the run under test may not have exercised the rule")
	}
	if after := porcelain(); after != before {
		t.Errorf("AC-STV-012: lint run mutated the tree:\nbefore: %q\nafter:  %q", before, after)
	}
}
