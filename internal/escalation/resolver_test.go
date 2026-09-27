package escalation_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

// contractSettings loads the worktree's workflow.yaml through the production
// loader and resolves workflow.autonomy, as the hooks will.
func contractSettings(t *testing.T, w *escalationtest.Worktree) config.AutonomySettings {
	t.Helper()
	cfg, err := config.NewLoader().Load(w.Path(".moai"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return config.ResolveAutonomy(cfg.Workflow)
}

// resolveFrom runs the resolver from a subdirectory of the worktree.
func resolveFrom(t *testing.T, w *escalationtest.Worktree) escalation.Resolution {
	t.Helper()
	sub := w.Path("internal/deep/dir")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	s := contractSettings(t, w)
	res, err := escalation.Resolve(sub, escalation.LoadVerifyEnv(w.Root, s))
	if err != nil {
		t.Fatalf("Resolve(%s): %v", w.Card, err)
	}
	return res
}

// queueFixture writes a queue file for the worktree's project at the path the
// queue store uses; the resolver must never read it.
func queueFixture(t *testing.T, w *escalationtest.Worktree, specID string) string {
	t.Helper()
	p := w.Path(".moai/state/todo/backlog.json")
	body := `{"items":[{"id":"t9001","text":"fixture card","state":"picked"`
	if specID != "" {
		body += `,"spec_id":"` + specID + `"`
	}
	body += "}]}\n"
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// twoSpecTree builds a worktree named card holding signed-valid contracts for
// in-progress SPEC-A-001 (card t9001) and SPEC-B-001 (card t9002).
func twoSpecTree(t *testing.T, card string) *escalationtest.Worktree {
	t.Helper()
	w := escalationtest.NewWorktree(t, card)
	w.WriteMode("contract", false)
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Card: "t9001"})
	w.AddSpec("SPEC-B-001", escalationtest.SpecOptions{Card: "t9002"})
	return w
}

// AC-AE-002 (REQ-AE-002): the card id comes from the worktree directory name,
// the contract's card field links it to a SPEC, and neither the branch name
// nor the queue spec_id is an input.
func TestResolveContractByCardField(t *testing.T) {
	want := map[string]string{"t9001": "SPEC-A-001", "t9002": "SPEC-B-001"}
	for card, spec := range want {
		t.Run(card, func(t *testing.T) {
			w := twoSpecTree(t, card)
			queue := queueFixture(t, w, "")

			check := func(label string) {
				t.Helper()
				res := resolveFrom(t, w)
				if res.Card != card {
					t.Errorf("%s: card = %q, want %q", label, res.Card, card)
				}
				if !res.Armed || res.SpecID != spec {
					t.Errorf("%s: armed=%v spec=%q, want armed against %s (lines %+v)", label, res.Armed, res.SpecID, spec, res.Lines)
				}
				if res.Report == nil || res.Report.State != "signed-valid" {
					t.Errorf("%s: verify report not signed-valid: %+v", label, res.Report)
				}
				wantPath := w.Path(".moai/specs/" + spec + "/contract.yaml")
				if res.ContractPath != wantPath {
					t.Errorf("%s: contract path = %q, want %q", label, res.ContractPath, wantPath)
				}
				if len(res.Lines) != 0 {
					t.Errorf("%s: unexpected audit lines %+v", label, res.Lines)
				}
			}

			check("baseline")

			// A queue spec_id pointing at the other SPEC changes nothing.
			queueFixture(t, w, "SPEC-B-001")
			if card == "t9002" {
				queueFixture(t, w, "SPEC-A-001")
			}
			check("queue spec_id set")

			// A branch naming the other card changes nothing.
			w.SetBranch("WT-t9002")
			if card == "t9002" {
				w.SetBranch("WT-t9001")
			}
			check("branch renamed")

			// An unreadable queue file changes nothing: the resolver never
			// opens it. (Permission bits do not deny reads on Windows.)
			if runtime.GOOS != "windows" {
				if err := os.Chmod(queue, 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(queue, 0o644) })
				check("queue unreadable")
			}
		})
	}

	// Structural guard: no non-test file of the detector package imports the
	// queue store, so no resolution can read the queue.
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, imp := range af.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasSuffix(path, "/internal/kanban") {
				t.Errorf("%s imports the queue store %s", f, path)
			}
		}
	}
}

// AC-AE-003 (REQ-AE-002, REQ-AE-023): every cause of not-armed writes exactly
// one not-armed line naming it, two claimants add one warning naming both
// SPEC IDs, and no escalation record is written for any of them.
func TestResolverNotArmedCases(t *testing.T) {
	cases := []struct {
		card    string
		setup   func(w *escalationtest.Worktree)
		cause   string
		warning []string // substrings the one warning line must carry; nil = no warning
	}{
		{
			card: "t9003",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Card: "t9001"})
			},
			cause: escalation.CauseNoClaimant,
		},
		{
			card: "t9004",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
				w.AddSpec("SPEC-B-001", escalationtest.SpecOptions{})
			},
			cause:   escalation.CauseMultipleClaimants,
			warning: []string{"SPEC-A-001", "SPEC-B-001"},
		},
		{
			card: "t9005",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Status: "completed"})
			},
			cause: escalation.CauseTerminal,
		},
		{
			card: "t9006",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Status: `"completed"`})
			},
			cause: escalation.CauseTerminal,
		},
		{
			card: "t9007",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Status: "archived"})
			},
			cause: escalation.CauseTerminal,
		},
		{
			card: "t9008",
			setup: func(w *escalationtest.Worktree) {
				w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Unsigned: true})
			},
			cause: escalation.CauseUnsigned,
		},
	}
	for _, tc := range cases {
		t.Run(tc.card, func(t *testing.T) {
			w := escalationtest.NewWorktree(t, tc.card)
			w.WriteMode("contract", false)
			tc.setup(w)

			res := resolveFrom(t, w)
			if res.Armed {
				t.Fatalf("armed against %s, want not-armed (%s)", res.SpecID, tc.cause)
			}

			var notArmed, warnings []escalation.AuditLine
			for _, l := range res.Lines {
				switch l.Kind {
				case escalation.LineNotArmed:
					notArmed = append(notArmed, l)
				case escalation.LineWarning:
					warnings = append(warnings, l)
				default:
					t.Errorf("unexpected line kind %q: %+v", l.Kind, l)
				}
			}
			if len(notArmed) != 1 {
				t.Fatalf("not-armed lines = %d, want exactly 1: %+v", len(notArmed), res.Lines)
			}
			if notArmed[0].Cause != tc.cause || notArmed[0].Card != tc.card {
				t.Errorf("not-armed line = %+v, want cause %q card %q", notArmed[0], tc.cause, tc.card)
			}
			if tc.warning == nil {
				if len(warnings) != 0 {
					t.Errorf("warnings = %+v, want none", warnings)
				}
			} else {
				if len(warnings) != 1 {
					t.Fatalf("warnings = %d, want exactly 1: %+v", len(warnings), res.Lines)
				}
				for _, s := range tc.warning {
					if !strings.Contains(warnings[0].Detail, s) {
						t.Errorf("warning %q does not name %s", warnings[0].Detail, s)
					}
				}
				if !slices.Equal(warnings[0].Specs, []string{"SPEC-A-001", "SPEC-B-001"}) {
					t.Errorf("warning specs = %v, want both SPEC IDs", warnings[0].Specs)
				}
			}
			assertNoEscalationRecords(t, w)
		})
	}

	// A signed-invalid sole candidate: one not-armed line plus one warning
	// carrying the verify reason codes (REQ-AE-023).
	t.Run("signed-invalid", func(t *testing.T) {
		w := escalationtest.NewWorktree(t, "t9009")
		w.WriteMode("contract", false)
		w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{})
		// Change a byte outside the signature: the digest no longer matches.
		c := string(w.Read(".moai/specs/SPEC-A-001/contract.yaml"))
		w.Write(".moai/specs/SPEC-A-001/contract.yaml",
			strings.Replace(c, "Implement the fixture feature", "Implement another feature", 1))

		res := resolveFrom(t, w)
		if res.Armed {
			t.Fatal("armed against a signed-invalid contract")
		}
		if len(res.Lines) != 2 || res.Lines[0].Kind != escalation.LineNotArmed || res.Lines[1].Kind != escalation.LineWarning {
			t.Fatalf("lines = %+v, want [not-armed, warning]", res.Lines)
		}
		if res.Lines[0].Cause != escalation.CauseSignedInvalid {
			t.Errorf("cause = %q, want %q", res.Lines[0].Cause, escalation.CauseSignedInvalid)
		}
		if !slices.Contains(res.Lines[1].Reasons, "contract_digest_mismatch") {
			t.Errorf("warning reasons = %v, want contract_digest_mismatch", res.Lines[1].Reasons)
		}
		assertNoEscalationRecords(t, w)
	})

	// No worktree root at all: not-armed, cause named.
	t.Run("no-worktree", func(t *testing.T) {
		dir := t.TempDir()
		res, err := escalation.Resolve(dir, escalation.VerifyEnv{})
		if err != nil {
			t.Fatal(err)
		}
		if res.Armed || len(res.Lines) != 1 || res.Lines[0].Cause != escalation.CauseNoWorktree {
			t.Errorf("resolution = %+v, want one not-armed line with cause %q", res, escalation.CauseNoWorktree)
		}
	})
}

// assertNoEscalationRecords fails when any escalation record directory exists
// under the worktree.
func assertNoEscalationRecords(t *testing.T, w *escalationtest.Worktree) {
	t.Helper()
	matches, _ := filepath.Glob(w.Path(".moai/reports/*/escalation"))
	if len(matches) != 0 {
		t.Errorf("escalation records written: %v", matches)
	}
}
