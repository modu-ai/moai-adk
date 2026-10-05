// zz_t747_anchor_probe_test.go — t747 / SPEC-AC-ANCHOR-SCOPE-001 corpus probe,
// promoted from .moai/reports/t747/probe/anchor_scope_probe_test.go into the
// tree as a COMMITTED two-column test (plan M1, the zz_t528 pattern of
// zz_t528_anchor_probe_test.go).
//
// RE-DERIVATION
//
//	T747_PROBE_OUT=<abs dir> go test ./internal/spec/ -run TestT747AnchorScope -v -count=1 -timeout 600s
//
// (T747_PROBE_OUT is optional; the default is a per-run t.TempDir() directory —
// this probe writes NOTHING into the repository tree, SPEC-HARNESS-EVIDENCE-WRITE-001.)
//
// TWO COLUMNS, ONE RUN. t747BaseACSectionStart below is a FROZEN copy of the
// PRE-repair findACSectionStart (vocabulary-only anchoring, first-empty fallback).
// findACSectionStart itself is the live column. Before the repair the two agree
// (that agreement is what makes the live column a valid re-measurement of the
// plan-phase before-numbers); after it, the frozen column keeps re-deriving the
// 14 / 9 / 1240 / 165 before-figures in the SAME run as the after-figures, so a
// comparison never crosses two runs or two trees (REQ-ACAS-005/006). Never edit
// t747BaseACSectionStart to track parser.go — it is the baseline anchor, not a
// mirror.
//
// DEFECT MEMBERSHIP is re-derived in-run from the FROZEN column (baseline
// narrow-miss ∪ baseline empty-anchor among declaration-bearing files), which
// is the authoritative control-set derivation (spec.md §E note); when the
// plan-phase frozen lists under .moai/reports/t747/probe/ are readable they are
// cross-checked against the derivation and any mismatch is logged, never
// self-healed.
//
// DISCRIMINATOR. declRe (package var, zz_t528_anchor_probe_test.go) IS
// discriminator B — verbatim, byte-frozen. The FROZEN baseline column keeps the
// plan-phase scan shape (break on any "##"-prefixed line) so the before-numbers
// stay comparable with 1240/14/9; the LIVE column is measured on the parse's
// own region shape (anchor-level break, as extractACLines reads it), which is
// the region AC-747-001/002 define. Both shapes are derived in-run; see
// t747InSectionDecls.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// t747ColonFormRe — the probe's own copy of the AC-shaped list form (bullet +
// AC-…: colon; separator REQUIRED). Used only to classify residual defect files
// (does the file carry a colon-form declaration anywhere?). The production
// region qualifier (parser.go, M2) is pattern-identical; this copy exists so
// the probe compiles and measures identically on the PRE-repair tree, where the
// production qualifier does not exist yet.
var t747ColonFormRe = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}AC-[A-Za-z0-9.-]*[A-Za-z0-9]\*{0,2}\s*(?:\([^()]*\)\s*)?\*{0,2}\s*[:—–]\s*`)

// t747BaseACSectionStart — FROZEN copy of findACSectionStart as of the pre-repair
// tree (8665f80b3 measurement baseline). Vocabulary-only anchoring with the
// first-empty fallback. Do not update this to track parser.go.
func t747BaseACSectionStart(lines []string) int {
	first := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if markdownHeadingLevel(trimmed) < 2 || !isACSectionHeading(trimmed) {
			continue
		}
		if first < 0 {
			first = i + 1
		}
		if len(extractACLines(lines, i+1, false)) > 0 {
			return i + 1
		}
	}
	return first
}

// t747InSectionDecls returns the declRe declaration lines inside the anchored
// region of lines starting at start (1-based line number; <=0 means no anchor).
// Two region shapes exist, and BOTH are load-bearing:
//
//   - strict (any "##"-prefixed line breaks) — the plan-phase frozen shape; the
//     baseline column and the 1240/14/9 comparability contract are measured
//     with it. It cannot see declarations nested under "###" subheadings.
//   - parse (break at a heading of the anchor's own level or higher) — the
//     region extractACLines actually reads; the live column's headline and the
//     REPAIRED disposition are measured with it, because AC-747-001/002 define
//     the region as the parse's region.
//
// Each entry is "lineNo:trimmed text".
func t747InSectionDecls(lines []string, start int, parseShape bool) []string {
	if start <= 0 || start > len(lines) {
		return nil
	}
	anchorLevel := markdownHeadingLevel(strings.TrimSpace(lines[start-1]))
	var out []string
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if parseShape {
			if level := markdownHeadingLevel(trimmed); level > 0 && level <= anchorLevel {
				break
			}
		} else if strings.HasPrefix(trimmed, "##") {
			break
		}
		if declRe.MatchString(lines[i]) {
			out = append(out, fmt.Sprintf("%d:%s", i+1, trimmed))
		}
	}
	return out
}

func t747AnchorDesc(lines []string, start int) string {
	if start <= 0 || start > len(lines) {
		return "NONE"
	}
	level := markdownHeadingLevel(strings.TrimSpace(lines[start-1]))
	return fmt.Sprintf("L%d:%s", level, strings.TrimSpace(lines[start-1]))
}

func TestT747AnchorScope(t *testing.T) {
	root := "../../.moai/specs"
	out := os.Getenv("T747_PROBE_OUT")
	if out == "" {
		out = t.TempDir()
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	type fileResult struct {
		path                 string
		lines                []string
		anywhere             int
		baseStart, liveStart int
		// baseIn — strict shape on the FROZEN anchor (frozen comparability).
		// liveInStrict — strict shape on the LIVE anchor (anchor-movement view).
		// liveIn — parse shape on the LIVE anchor (the parse's region: the
		// headline after-column and REPAIRED are measured here).
		baseIn, liveInStrict, liveIn []string
		baseInParse                  []string
	}
	var results []*fileResult

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		lines := strings.Split(string(b), "\n")
		fr := &fileResult{path: p, lines: lines}
		for _, line := range lines {
			if declRe.FindStringSubmatch(line) != nil {
				fr.anywhere++
			}
		}
		fr.baseStart = t747BaseACSectionStart(lines)
		fr.liveStart = findACSectionStart(lines)
		fr.baseIn = t747InSectionDecls(lines, fr.baseStart, false)
		fr.baseInParse = t747InSectionDecls(lines, fr.baseStart, true)
		fr.liveInStrict = t747InSectionDecls(lines, fr.liveStart, false)
		fr.liveIn = t747InSectionDecls(lines, fr.liveStart, true)
		results = append(results, fr)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].path < results[j].path })

	// Headline counters — before (frozen column) and after (live column), one run.
	// The after-column is measured on the parse's region (t747InSectionDecls
	// parse shape); the strict-shape after-total is reported alongside for
	// strict comparability with the frozen 1240.
	denominator := len(results)
	beforeNarrow, afterNarrow := 0, 0
	beforeEmpty, afterEmpty := 0, 0
	beforeInTotal, afterInTotal, afterInStrictTotal := 0, 0, 0
	declAnywhereTotal := 0
	var fileList []string

	// Defect membership, baseline-derived (frozen column), per spec.md §E.
	baseNarrow := map[string]bool{}
	baseEmpty := map[string]bool{}
	for _, fr := range results {
		fileList = append(fileList, fr.path)
		declAnywhereTotal += fr.anywhere
		beforeInTotal += len(fr.baseIn)
		afterInTotal += len(fr.liveIn)
		afterInStrictTotal += len(fr.liveInStrict)
		if fr.anywhere == 0 {
			continue
		}
		if fr.baseStart <= 0 {
			beforeNarrow++
			baseNarrow[fr.path] = true
		} else if len(fr.baseIn) == 0 {
			beforeEmpty++
			baseEmpty[fr.path] = true
		}
		if fr.liveStart <= 0 {
			afterNarrow++
		} else if len(fr.liveIn) == 0 {
			afterEmpty++
		}
	}

	// Control set: declaration-bearing files OUTSIDE the baseline defect union.
	// Two comparison instruments for AC-747-003, both must be delta-free (or
	// justified): strict shape (frozen comparability — pure anchor movement)
	// and parse shape (the parse-level no-regression instrument).
	controlTotal := 0
	var controlDeltas, controlParseDeltas []string
	for _, fr := range results {
		if fr.anywhere == 0 || baseNarrow[fr.path] || baseEmpty[fr.path] {
			continue
		}
		controlTotal++
		if strings.Join(fr.baseIn, "\n") != strings.Join(fr.liveInStrict, "\n") {
			controlDeltas = append(controlDeltas,
				fmt.Sprintf("%s\tbaseline=%d entries\tlive=%d entries", fr.path, len(fr.baseIn), len(fr.liveInStrict)))
		}
		if strings.Join(fr.baseInParse, "\n") != strings.Join(fr.liveIn, "\n") {
			controlParseDeltas = append(controlParseDeltas,
				fmt.Sprintf("%s\tbaseline=%d entries\tlive=%d entries", fr.path, len(fr.baseInParse), len(fr.liveIn)))
		}
	}

	// Prose bound (AC-747-004): every newly in-section declaration must sit in a
	// defect-set file. Compared on the parse shape both sides (same shape), so
	// a control file with an unchanged anchor contributes no false entry.
	var newlyIncluded []string
	for _, fr := range results {
		baseSet := map[string]bool{}
		for _, e := range fr.baseInParse {
			baseSet[e] = true
		}
		for _, e := range fr.liveIn {
			if !baseSet[e] {
				class := "control"
				if baseNarrow[fr.path] {
					class = "narrow-defect"
				} else if baseEmpty[fr.path] {
					class = "empty-defect"
				}
				newlyIncluded = append(newlyIncluded, fmt.Sprintf("%s\t%s\t%s", class, fr.path, e))
			}
		}
	}

	// Per-file disposition for every baseline defect file (AC-747-001/002):
	// repaired = the live anchor's PARSE region holds >=1 declaration line;
	// otherwise the classification code names what the file's declarations are.
	var dispositions []string
	repairedNarrow, repairedEmpty := 0, 0
	for _, fr := range results {
		class := ""
		if baseNarrow[fr.path] {
			class = "narrow"
		} else if baseEmpty[fr.path] {
			class = "empty"
		} else {
			continue
		}
		code := ""
		switch {
		case len(fr.liveIn) > 0:
			if class == "narrow" {
				repairedNarrow++
			} else {
				repairedEmpty++
			}
			code = "REPAIRED"
		case fr.liveStart > 0 && len(extractACLines(fr.lines, fr.liveStart, false)) > 0:
			// Anchored at a genuinely parseable (non-bullet) AC section; the
			// declRe-based empty-anchor metric cannot see those lines.
			code = "ANCHOR-PARSEABLE-NONBULLET"
		case fr.liveStart > 0:
			code = "ANCHORED-REGION-NO-DECLRE"
		default:
			// Still no anchor: scan for a colon-form bullet anywhere to say why.
			hasColonForm := false
			for _, line := range fr.lines {
				if t747ColonFormRe.MatchString(line) {
					hasColonForm = true
					break
				}
			}
			if hasColonForm {
				code = "RESIDUAL-COLON-FORM-UNANCHORED"
			} else {
				code = "PROSE-SHAPED-NO-COLON-FORM"
			}
		}
		dispositions = append(dispositions, fmt.Sprintf("%s\t%s\tbase=%s\tlive=%s\tbaseIn=%d\tliveInParse=%d\t%s",
			class, fr.path, t747AnchorDesc(fr.lines, fr.baseStart), t747AnchorDesc(fr.lines, fr.liveStart),
			len(fr.baseIn), len(fr.liveIn), code))
	}

	// Cross-check the in-run defect derivation against the plan-phase frozen
	// lists (read-only, logged, never self-healed).
	t747CrossCheckFrozen(t, "anchor-missed.txt", baseNarrow, out)
	t747CrossCheckFrozen(t, "empty-anchor.txt", baseEmpty, out)

	// Artifacts (output dir only — never the repository tree).
	t528Write(t, out, "filelist.txt", fileList)
	t528Write(t, out, "defect-disposition.txt", dispositions)
	t528Write(t, out, "control-delta.txt", controlDeltas)
	t528Write(t, out, "control-parse-delta.txt", controlParseDeltas)
	t528Write(t, out, "newly-included.txt", newlyIncluded)

	t.Logf("OUTDIR = %s", out)
	t.Logf("DENOMINATOR spec.md read = %d (list: %s/filelist.txt)", denominator, out)
	t.Logf("NARROW axis (decls, no anchor): before=%d after=%d repaired=%d", beforeNarrow, afterNarrow, repairedNarrow)
	t.Logf("EMPTY axis (anchored, 0 in-section decls): before=%d after=%d repaired=%d", beforeEmpty, afterEmpty, repairedEmpty)
	t.Logf("declarations WHOLE FILE = %d", declAnywhereTotal)
	t.Logf("declarations IN-SECTION: before(frozen-strict)=%d after(parse-region)=%d after(strict-shape)=%d", beforeInTotal, afterInTotal, afterInStrictTotal)
	t.Logf("CONTROL set (decl-bearing, non-defect) = %d, strict-shape deltas = %d, parse-shape deltas = %d", controlTotal, len(controlDeltas), len(controlParseDeltas))
	t.Logf("NEWLY-INCLUDED declarations = %d (list: %s/newly-included.txt)", len(newlyIncluded), out)
	t.Logf("DEFECT dispositions = %d (list: %s/defect-disposition.txt)", len(dispositions), out)
}

// t747CrossCheckFrozen compares an in-run derived defect set against the
// plan-phase frozen list (paths under ../../.moai/reports/t747/probe/), logging
// match/mismatch. The frozen files are never written.
func t747CrossCheckFrozen(t *testing.T, name string, derived map[string]bool, outDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../.moai/reports/t747/probe", name))
	if err != nil {
		t.Logf("FROZEN-CROSSCHECK %s: frozen list unreadable (%v) — in-run derivation stands, no comparison", name, err)
		return
	}
	frozen := map[string]bool{}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path := strings.Fields(line)[0]
		frozen[path] = true
		count++
	}
	match := len(frozen) == len(derived)
	if match {
		for p := range derived {
			if !frozen[p] {
				match = false
				break
			}
		}
	}
	if match {
		t.Logf("FROZEN-CROSSCHECK %s: MATCH (%d files; detail written to %s)", name, count, outDir)
	} else {
		t.Logf("FROZEN-CROSSCHECK %s: MISMATCH — frozen=%d derived=%d (inspect before trusting either column)", name, count, len(derived))
		for p := range derived {
			if !frozen[p] {
				t.Logf("  derived-only: %s", p)
			}
		}
		for p := range frozen {
			if !derived[p] {
				t.Logf("  frozen-only:  %s", p)
			}
		}
	}
}
