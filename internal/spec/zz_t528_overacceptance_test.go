// zz_t528_overacceptance_test.go — t528 / SPEC-AC-COLLECTOR-ANCHOR-001,
// REQ-ACA-001-012 / AC-ACA-001-014 / M6.
//
// The over-acceptance axis. It is the one axis the mutant set cannot draw:
// every mutant removes capability, and over-acceptance happens while the
// capability WORKS. A widened parser that reads a prose bullet as a declaration
// is green under every other test in this card.
//
// Two guards, and the second exists because the first is a sample.
package spec

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// t528NonDeclarationBullets — REAL bullets that sit INSIDE an AC section of a
// real spec.md and are NOT declarations. Every line is verbatim from the corpus
// with its source; none is invented, because an invented line being rejected
// says nothing about what the corpus contains.
//
// Selection: of the 1502 bullets living inside AC sections corpus-wide, 335 are
// not declarations under discriminator B. These are the ones that mention an
// "AC-" token anyway — the hardest cases, and the ones a widened anchor is most
// likely to mistake for declarations.
var t528NonDeclarationBullets = []struct{ name, source, text string }{
	{
		name:   "milestone_line_naming_an_ac_range",
		source: ".moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md:456",
		text:   "- M1: config field + validation + fable enum (AC-MTP-001..005b, 024)",
	},
	{
		name:   "milestone_line_naming_an_ac_range_2",
		source: ".moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md:457",
		text:   "- M2: profile structure + matrices fidelity + unified apply pass (AC-MTP-006..013)",
	},
	{
		name:   "global_scope_line_naming_an_ac",
		source: ".moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md:463",
		text:   "- Global: full test suite, vet/lint, coverage, spec-lint delta (AC-MTP-026)",
	},
	{
		name:   "tier_grouping_line_listing_acs",
		source: ".moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md:295",
		text:   "- **T1 (Pre-Push)**: AC-CIAUT-001 (lint block), AC-CIAUT-002 (16-language detect), AC-CIAUT-003 (`--no-verify` log)",
	},
	{
		name:   "prose_note_mentioning_an_anchor",
		source: ".moai/specs/SPEC-DIVECC-COMPACTION-LAYER-NAMING-001/spec.md:150",
		text:   "- Non-vacuity note: this anchor pattern was confirmed to return **0 matches** against the current tree.",
	},
	{
		name:   "bare_ears_clause_bullet",
		source: ".moai/specs/SPEC-SEMAP-001/spec.md:110",
		text:   "- Then a contract violation report is generated",
	},
	{
		name:   "verification_command_bullet",
		source: ".moai/specs/SPEC-DIVECC-COMPACTION-LAYER-NAMING-001/spec.md:156",
		text:   "- Verification: `diff <local> <mirror>; echo \"exit=$?\"` → `exit=0`.",
	},
	{
		name:   "given_prelude_bullet",
		source: ".moai/specs/SPEC-V3R5-CORE-SLIM-001/spec.md:127",
		text:   "- Given: Track B preload addition is applied to 4 expert agents",
	},
}

// TestT528NonDeclarationBulletsStayRejected — the pinned regression guard.
func TestT528NonDeclarationBulletsStayRejected(t *testing.T) {
	for _, line := range t528NonDeclarationBullets {
		t.Run(line.name, func(t *testing.T) {
			if p := parseSingleACLine(line.text); p != nil {
				t.Errorf("non-declaration bullet was read as a declaration\n  source: %s\n  line:   %s\n  id:     %s",
					line.source, line.text, p.id)
			}
		})
	}
}

// TestT528NonDeclarationBulletsCorpusSweep — the whole population, not a sample.
//
// The pinned list above is eight lines chosen by hand, so on its own it proves
// only that eight lines are refused. This sweeps every non-declaration bullet
// the corpus actually contains inside an AC section.
//
// It carries its own NON-VACUITY control: the roster has to be non-empty and of
// the expected order, because a missing or truncated roster file would make this
// test pass while checking nothing — which is exactly how this card's earlier
// zero-valued findings went wrong.
func TestT528NonDeclarationBulletsCorpusSweep(t *testing.T) {
	const rosterPath = "../../.moai/reports/t528/probe/nondecl-bullets.txt"

	f, err := os.Open(rosterPath)
	if err != nil {
		t.Fatalf("roster missing — the sweep would otherwise pass while checking nothing: %v", err)
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)

	checked := 0
	var accepted []string
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		// roster entries are "path:lineno:text"
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		checked++
		if p := parseSingleACLine(parts[2]); p != nil {
			accepted = append(accepted, line+"   => collected as "+p.id)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	// NON-VACUITY: an empty or near-empty roster means the generator broke, not
	// that the corpus is clean.
	if checked < 300 {
		t.Fatalf("roster holds only %d entries; the measured corpus population is 335. "+
			"A shrunken roster makes this sweep vacuous — regenerate it before trusting a pass.", checked)
	}

	if len(accepted) > 0 {
		for _, a := range accepted {
			t.Errorf("over-acceptance: %s", a)
		}
	}
	t.Logf("non-declaration bullets swept = %d, read as declarations = %d", checked, len(accepted))
}
