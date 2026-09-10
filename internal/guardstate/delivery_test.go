package guardstate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// specTablePath is the normative state table's home. The table is the artifact
// (REQ-GSM-006); this test reads it rather than restating it, so a row added,
// reworded, or re-decided there is re-read here instead of drifting.
const specTablePath = ".moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md"

// m1FieldProbe links one M1 dependency, as the table's `Flipped by` column
// NAMES it, to a mechanical check that the field exists in the M1 schema AND is
// actually read by the parser.
//
// The probe is a round trip rather than a reflective field lookup: a field
// deleted from the struct is silently ignored by the YAML decoder, so a
// presence-only check would pass on a schema that no longer carries it. The
// round trip fails in that case, which is the point.
type m1FieldProbe struct {
	// marker is the phrase the table's cell uses to name the dependency.
	marker string
	// field is the schema field it refers to, for the failure message.
	field string
	// carried reports whether the parsed entry carries the value the fixture
	// declared for this field.
	carried func(Entry) bool
}

const probeFixture = `
entries:
  - kind: policy-rule
    locator: .github/workflows/release.yml
    events: [push]
    window: 90d
    measure: verdict-rendered
    expected_when: release-cycle
`

func m1FieldProbes() []m1FieldProbe {
	return []m1FieldProbe{
		{
			marker:  "`kind` field",
			field:   "Entry.Kind",
			carried: func(e Entry) bool { return e.Kind == "policy-rule" },
		},
		{
			marker:  "window field",
			field:   "Entry.Window",
			carried: func(e Entry) bool { return e.Window == "90d" },
		},
		{
			marker:  "measured-quantity value",
			field:   "Entry.Measure",
			carried: func(e Entry) bool { return e.Measure == MeasureVerdictRendered },
		},
		{
			marker:  "release-cycle-conditional field",
			field:   "Entry.ExpectedWhen",
			carried: func(e Entry) bool { return e.ExpectedWhen == "release-cycle" && e.IsConditional() },
		},
	}
}

// stateTableRows returns the `Flipped by` cell of every numbered row of the
// normative table, keyed by row number.
func stateTableRows(t *testing.T) map[string]string {
	t.Helper()
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, specTablePath))
	if err != nil {
		t.Fatalf("read %s: %v", specTablePath, err)
	}
	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "| ") {
			continue
		}
		cells := strings.Split(strings.Trim(trimmed, "|"), "|")
		if len(cells) != 6 {
			continue
		}
		num := strings.TrimSpace(cells[0])
		if len(num) != 1 || num[0] < '1' || num[0] > '9' {
			continue
		}
		rows[num] = cells[5]
	}
	return rows
}

// AC-GSM-007 clause (c) — a schema read, requiring no evaluator. FOR EVERY ROW
// of the state table — not only rows that happen to declare a dependency — the
// M1 fields that row's decision depends on are present in the M1 schema.
//
// Mutant this kills: an M1 schema shipped without one of the fields the table's
// rows decide on. That mutant does NOT leave a cell empty — it makes the cell
// produce the WRONG VALUE. Row 5 without its conditional field silently becomes
// row 6, and every correctly-quiet release-only subject is reported as an
// anomaly on every sweep. Neither the row-totality proof (AC-GSM-007 a) nor the
// value-reachability proof (AC-GSM-008) can see that, which is why this clause
// is a separate assertion read against the schema.
func TestStateTable_EveryRowsM1DependencyIsInTheSchema(t *testing.T) {
	rows := stateTableRows(t)
	const wantRows = 8
	if len(rows) != wantRows {
		t.Fatalf("read %d rows from the %s table, want %d — a partial read would check a partial table and pass", len(rows), specTablePath, wantRows)
	}

	m, err := ParseManifest([]byte(probeFixture))
	if err != nil {
		t.Fatalf("probe fixture rejected by the M1 schema: %v", err)
	}
	entry := m.Entries[0]

	probes := m1FieldProbes()
	matchedAnywhere := map[string]bool{}

	for num, cell := range rows {
		for _, p := range probes {
			if !strings.Contains(cell, p.marker) {
				continue
			}
			matchedAnywhere[p.marker] = true
			if !p.carried(entry) {
				t.Errorf("row %s declares the %s and the M1 schema does not carry it (%s): the row would produce the wrong value, not an empty cell", num, p.marker, p.field)
			}
		}
	}

	// Anti-vacuity, in both directions. A marker that matches no row means the
	// table's wording moved and this check silently stopped reading it; a row
	// set that names no dependency at all means the cells were emptied.
	if len(matchedAnywhere) == 0 {
		t.Fatalf("no row named any M1 dependency; the `Flipped by` column is not being read")
	}
	for _, p := range probes {
		if !matchedAnywhere[p.marker] {
			t.Errorf("no row names %q — either the table's wording changed or the dependency was dropped from it; this check cannot verify what it cannot find", p.marker)
		}
	}
}
