// Package cli — the config-error shape × expected-outcome matrix for
// `loadCeilingConfig` (card t1520, R3-P2-1/R3-P2-2; SPEC-AUDIT-CEILING-002
// follow-ups from card-review r2 CR2-P2-1 / r3 R3-P2-1·R3-P2-2). The Go test
// table below IS the matrix: each row carries a harness.yaml fixture written
// verbatim from the row and the expected outcome per the row.
//
// Config-error matrix (card t1520, FINAL v1.0). The authoritative table is
// .moai/reports/t1520/config-error-matrix.md; this header is its condensed
// form so the committed tree carries the table.
//
// Field states:
//
//	ceilings: absent | valid-full | valid-partial (present values valid) |
//	          partial-invalid (SOME keys decode, one fails — yaml.v3 fills the
//	          survivors and returns a TypeError, e.g. {S: 1, M: bad}; the
//	          typed map is left NON-nil, so nil-ness alone misses it —
//	          CR-P2-1) |
//	          malformed (decode into map[string]int fails outright, e.g. [bad])
//	policy:   absent | readable (on_final_hit string — unknown names are
//	          REQ-ACR-003 evaluation hold-arm concerns, not config errors) |
//	          unreadable (non-string on_final_hit, non-int auto_delta_rounds)
//	file:     not-found | io-error | strict-reject
//
// Matrix (file × ceilings × policy → expected loadCeilingConfig result):
//
//	M1  not-found     —             —          → full defaults + default policy, nil err
//	M2  ok            valid-full    readable   → pass-through, nil err
//	M3  ok            valid-partial readable   → seeded merge S + default M/L, nil err
//	                                           (internal/config/loader.go seeded-merge
//	                                           contract — unchanged by this card)
//	M4  strict-reject valid-partial unreadable → policy→"" + PARTIAL MAP MERGED OVER THE
//	                                           DEFAULTS (R3-P2-2 fix; was bare map[S:1])
//	                                           — evaluation and --record reach the hold arm
//	M5  strict-reject malformed     readable   → strict error propagates (the loose pass
//	                                           reports policyUnreadable=false)
//	M6  strict-reject malformed     unreadable → strict error propagates EVEN THOUGH the
//	                                           policy is unreadable (R3-P2-1 fix; recovery
//	                                           covers only the policy field) and the error
//	                                           names the ceiling decode — the fallback
//	                                           previously swallowed it into default
//	                                           ceilings and a zero-error --record
//	M7  strict-reject absent        unreadable → policy→"" + full defaults, nil err
//	M8  strict-reject absent        readable   → strict error propagates
//	M9  io-error      —             —          → strict error propagates (the loose read's
//	                                           readErr reports not-unreadable)
//	M10 ok            valid-partial unreadable → impossible combination: the loose pass
//	                                           calls a policy unreadable only when the
//	                                           strict decode of the same fields rejects
//	                                           too (both decodes read the same yaml type
//	                                           tables) — witnessed directly below, not
//	                                           exercised through loadCeilingConfig
//	M11 ok            neg/zero values + unknown tier key, readable → pass-through
//	                                           verbatim (no value validation — follow-up
//	                                           candidate, out of this card's scope)
//	M12 ok            —             readable-other (unknown policy name) → the string
//	                                           passes through; its hold-arm disposition is
//	                                           evaluation-level, not config-level
//	M13 strict-reject partial-invalid unreadable → strict error propagates and names
//	                                           the ceiling field (CR-P2-1 fix; the
//	                                           partial decode's survivors must not be
//	                                           adopted as a valid override — that
//	                                           absorbed the bad key into a default-
//	                                           substituted, zero-error --record)
//	M14 strict-reject partial-invalid readable   → strict error propagates (the loose
//	                                           pass reports policyUnreadable=false)
//
// Regression seals: a joint ceiling+policy decode failure (M6) re-reddens if
// the fallback again swallows the ceiling error; a partial-ceilings
// unreadable-policy load (M4) re-reddens if the defaults merge is dropped; a
// partially-decodable ceilings field (M13) re-reddens if the loose pass
// judges malformed-ness by map nil-ness alone.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

// matrixOverlay writes the given harness.yaml body into the temp project's
// config sections directory (the path loadCeilingConfig reads).
func matrixOverlay(t *testing.T, project, body string) {
	t.Helper()
	sections := filepath.Join(project, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sections, "harness.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ceilingMatrixRow is one matrix row: the fixture harness.yaml body written
// verbatim from the row, plus the expected loadCeilingConfig triple.
type ceilingMatrixRow struct {
	row         string // matrix row id + shape
	yaml        string // fixture harness.yaml body ("" = no harness.yaml at all)
	asDir       bool   // place a directory at the harness.yaml path (io-error arm)
	wantMap     map[string]int
	wantPolicy  string
	wantErr     bool
	errNames    string // substring the error must carry ("" = no substring check)
	witnessOnly bool   // M10: eliminated combination — the witness test owns the row, the generic body skips it
	focus       string // matrix §2 verification focus
}

func ceilingMatrixRows() []ceilingMatrixRow {
	return []ceilingMatrixRow{
		{
			row:        "M1 not-found / — / —",
			yaml:       "", // no harness.yaml written
			wantMap:    map[string]int{"S": 1, "M": 2, "L": 3},
			wantPolicy: "hold-and-split",
			focus:      "absent file is a defaults project, not an error (existing behavior pinned)",
		},
		{
			row: "M2 ok / valid-full / readable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 2
    M: 4
    L: 6
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold
`,
			wantMap:    map[string]int{"S": 2, "M": 4, "L": 6},
			wantPolicy: "hold",
			focus:      "a fully readable config passes through verbatim (existing behavior pinned)",
		},
		{
			row: "M3 ok / valid-partial / readable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 5
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split
`,
			wantMap:    map[string]int{"S": 5, "M": 2, "L": 3},
			wantPolicy: "hold-and-split",
			focus:      "the strict loader's seeded merge (loader.go) keeps default M/L under a partial key — unchanged by this card",
		},
		{
			row: "M4 strict-reject / valid-partial / unreadable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: []
`,
			wantMap:    map[string]int{"S": 1, "M": 2, "L": 3},
			wantPolicy: "",
			focus:      "R3-P2-2: the loose partial map merges over the shipped defaults (was bare map[S:1]) — evaluation and --record reach the hold arm",
		},
		{
			row: "M5 strict-reject / malformed / readable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings: [bad]
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split
`,
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			errNames:   config.ErrInvalidYAML.Error(),
			focus:      "a readable policy means the strict failure lives elsewhere — stay strict (existing behavior pinned)",
		},
		{
			row: "M6 strict-reject / malformed / unreadable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings: [bad]
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: []
`,
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			errNames:   "plan_audit_tier_ceilings",
			focus:      "R3-P2-1: the ceiling decode failure propagates even when the policy is unreadable (recovery covers only the policy field); the error must name the ceiling field so the record abstention is auditable",
		},
		{
			row: "M7 strict-reject / absent / unreadable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: []
`,
			wantMap:    map[string]int{"S": 1, "M": 2, "L": 3},
			wantPolicy: "",
			focus:      "absent ceilings key under policy recovery resolves the full defaults (existing behavior pinned)",
		},
		{
			row: "M8 strict-reject / absent / readable",
			yaml: `harness:
  evaluator:
    memory_scope: interactive
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split
`,
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			focus:      "a readable policy means the strict failure (memory_scope validation here) stays fatal (existing behavior pinned)",
		},
		{
			row:        "M9 io-error / — / —",
			asDir:      true, // a directory at the harness.yaml path: ReadFile fails, not ErrConfigNotFound
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			focus:      "a read failure is not a recovery arm (existing behavior pinned)",
		},
		{
			row: "M10 ok / valid-partial / unreadable (eliminated)",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
    M: 2
    L: 3
  plan_audit_ceiling_policy:
    auto_delta_rounds: {}
    on_final_hit: hold-and-split
`,
			focus:       "the combination cannot occur: the loose pass calls a policy unreadable only when the strict decode of the same fields rejects too (same yaml type tables) — this row is a direct witness, not a loadCeilingConfig case",
			witnessOnly: true,
		},
		{
			row: "M11 ok / zero·negative values + unknown tier key / readable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 0
    M: -2
    XL: 9
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split
`,
			wantMap:    map[string]int{"S": 0, "M": -2, "XL": 9, "L": 3},
			wantPolicy: "hold-and-split",
			focus:      "no ceilings value validation exists — values pass through verbatim (the seeded merge keeps the absent L key at its default; follow-up candidate, out of this card's scope per matrix §M11)",
		},
		{
			row: "M12 ok / — / readable-other (unknown policy name)",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
    M: 2
    L: 3
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: split-only
`,
			wantMap:    map[string]int{"S": 1, "M": 2, "L": 3},
			wantPolicy: "split-only",
			focus:      "an unknown policy NAME is not a config error — the string passes through; the hold-arm reading is evaluation-level (existing behavior pinned)",
		},
		{
			row: "M13 strict-reject / partial-invalid / unreadable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
    M: bad
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: []
`,
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			errNames:   "plan_audit_tier_ceilings",
			focus:      "CR-P2-1: a partially decodable ceilings field is still a defective ceilings field — yaml.v3 fills the survivors and returns a TypeError, so the typed map is non-nil; adopting the fragment absorbed the bad key into default-substituted ceilings and a zero-error --record",
		},
		{
			row: "M14 strict-reject / partial-invalid / readable",
			yaml: `harness:
  evaluator:
    memory_scope: per_iteration
  plan_audit_tier_ceilings:
    S: 1
    M: bad
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split
`,
			wantMap:    nil,
			wantPolicy: "",
			wantErr:    true,
			focus:      "same partial decode with a readable policy — the strict failure stays fatal (existing behavior pinned; guards the readable arm against adopting the fragment too)",
		},
	}
}

// TestSpecCeilingConfigMatrix — the matrix as a table: one case per row,
// function level. Each case asserts the exact (ceilings, onFinalHit, err)
// triple loadCeilingConfig returns for the row's fixture.
func TestSpecCeilingConfigMatrix(t *testing.T) {
	for _, tc := range ceilingMatrixRows() {
		t.Run(tc.row, func(t *testing.T) {
			if tc.witnessOnly {
				t.Skip("eliminated combination (see TestSpecCeilingConfigMatrixM10Witness)")
			}
			project := t.TempDir()
			switch {
			case tc.asDir:
				dir := filepath.Join(project, ".moai", "config", "sections", "harness.yaml")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			case tc.yaml == "":
				// not-found arm: write nothing.
			default:
				matrixOverlay(t, project, tc.yaml)
			}

			ceilings, onFinalHit, err := loadCeilingConfig(project)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("want an error, got ceilings=%v policy=%q and nil error (focus: %s)", ceilings, onFinalHit, tc.focus)
				}
				if tc.errNames != "" && !strings.Contains(err.Error(), tc.errNames) {
					t.Errorf("error %q does not mention %q (focus: %s)", err, tc.errNames, tc.focus)
				}
			} else if err != nil {
				t.Fatalf("want no error, got: %v (focus: %s)", err, tc.focus)
			}
			if !tc.wantErr {
				if !reflect.DeepEqual(ceilings, tc.wantMap) {
					t.Errorf("ceilings = %v, want %v (focus: %s)", ceilings, tc.wantMap, tc.focus)
				}
				if onFinalHit != tc.wantPolicy {
					t.Errorf("onFinalHit = %q, want %q (focus: %s)", onFinalHit, tc.wantPolicy, tc.focus)
				}
			}
		})
	}
}

// TestSpecCeilingConfigMatrixM10Witness — M10's elimination evidence: for the
// same fixture file, the strict loader REJECTS what the loose pass calls an
// unreadable policy. Both decodes read the same yaml type tables, so "the
// strict load passed AND the loose pass reported an unreadable policy" is
// structurally impossible — the recovery arm never races a working strict
// decode on the policy fields.
func TestSpecCeilingConfigMatrixM10Witness(t *testing.T) {
	for _, tc := range ceilingMatrixRows() {
		if !strings.HasPrefix(tc.row, "M10") {
			continue
		}
		project := t.TempDir()
		matrixOverlay(t, project, tc.yaml)
		path := filepath.Join(project, ".moai", "config", "sections", "harness.yaml")

		if _, strictErr := config.LoadHarnessConfig(path); strictErr == nil {
			t.Error("the strict loader accepted a policy fixture the loose pass calls unreadable — the M10 elimination no longer holds")
		}
		_, ceilingsMalformed, policyUnreadable := looseReadCeilingPolicy(path)
		if ceilingsMalformed {
			t.Error("the loose pass reported the ceilings field malformed — the M10 witness fixture must leave the ceilings decodable so the elimination isolates the policy arm")
		}
		if !policyUnreadable {
			t.Error("the loose pass did not report the unreadable policy (auto_delta_rounds as a map) — the M10 witness fixture no longer exercises the unreadable arm")
		}
	}
}

// TestSpecCeilingConfigMatrixVerb — CLI level for the rows the matrix marks
// observable at the verb: M4 and M7 reach --record (hold arm, record
// written), M6 and M13 record nothing and the error output names the ceiling
// field (the record abstention must be auditable).
func TestSpecCeilingConfigMatrixVerb(t *testing.T) {
	rows := ceilingMatrixRows()
	byRow := func(prefix string) ceilingMatrixRow {
		for _, r := range rows {
			if strings.HasPrefix(r.row, prefix) {
				return r
			}
		}
		panic("matrix row missing: " + prefix)
	}

	t.Run("M4 record reaches the hold arm with merged ceilings", func(t *testing.T) {
		project := ceilingVerbProject(t)
		matrixOverlay(t, project, byRow("M4").yaml)
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		t.Chdir(project)

		cmd := newSpecCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("M4 must reach the hold arm and record: %v\noutput: %s", err, out.String())
		}

		raw, err := os.ReadFile(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json"))
		if err != nil {
			t.Fatalf("no record written under the partial-ceilings unreadable-policy overlay: %v\noutput: %s", err, out.String())
		}
		var got runtime.CeilingOutcome
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("record is not the CeilingOutcome JSON: %v", err)
		}
		if got.Disposition != "hold" {
			t.Errorf("disposition = %q, want hold (the unreadable-policy fail-closed arm)", got.Disposition)
		}
		if got.Ceiling != 2 {
			t.Errorf("ceiling = %d, want 2 (the merged default M ceiling — a bare map[S:1] resolves no M and no L)", got.Ceiling)
		}
		if got.SplitProposalRef != "" {
			t.Errorf("unreadable-policy hold carries a split-proposal reference %q, want none", got.SplitProposalRef)
		}
	})

	t.Run("M6 no record and the error names the ceiling decode", func(t *testing.T) {
		project := ceilingVerbProject(t)
		matrixOverlay(t, project, byRow("M6").yaml)
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		t.Chdir(project)

		cmd := newSpecCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
		err := cmd.Execute()
		if err == nil {
			t.Fatalf("M6 must fail the verb instead of recording under default ceilings (the swallowed-error defect)\noutput: %s", out.String())
		}
		if !strings.Contains(err.Error(), "plan_audit_tier_ceilings") {
			t.Errorf("error %q does not name the ceiling field — the record abstention must be auditable", err)
		}
		if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")); !os.IsNotExist(statErr) {
			t.Errorf("a record was written despite the propagated ceiling error (stat err = %v)", statErr)
		}
	})

	t.Run("M13 no record under a partially decodable ceilings field", func(t *testing.T) {
		project := ceilingVerbProject(t)
		matrixOverlay(t, project, byRow("M13").yaml)
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		t.Chdir(project)

		cmd := newSpecCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
		err := cmd.Execute()
		if err == nil {
			t.Fatalf("M13 must fail the verb — the surviving fragment must not be adopted as a valid override (CR-P2-1: the defect recorded an outcome under default-substituted ceilings)\noutput: %s", out.String())
		}
		if !strings.Contains(err.Error(), "plan_audit_tier_ceilings") {
			t.Errorf("error %q does not name the ceiling field — the record abstention must be auditable", err)
		}
		if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")); !os.IsNotExist(statErr) {
			t.Errorf("a record was written despite the propagated ceiling error (stat err = %v)", statErr)
		}
	})

	t.Run("M7 record reaches the hold arm with full default ceilings", func(t *testing.T) {
		project := ceilingVerbProject(t)
		matrixOverlay(t, project, byRow("M7").yaml)
		t.Setenv("CLAUDE_PROJECT_DIR", "")
		t.Chdir(project)

		cmd := newSpecCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("M7 must reach the hold arm and record: %v\noutput: %s", err, out.String())
		}

		raw, err := os.ReadFile(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json"))
		if err != nil {
			t.Fatalf("no record written under the absent-ceilings unreadable-policy overlay: %v\noutput: %s", err, out.String())
		}
		var got runtime.CeilingOutcome
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("record is not the CeilingOutcome JSON: %v", err)
		}
		if got.Disposition != "hold" {
			t.Errorf("disposition = %q, want hold (the unreadable-policy fail-closed arm)", got.Disposition)
		}
		if got.Ceiling != 2 {
			t.Errorf("ceiling = %d, want 2 (the shipped default M ceiling — the key is absent entirely)", got.Ceiling)
		}
	})
}
