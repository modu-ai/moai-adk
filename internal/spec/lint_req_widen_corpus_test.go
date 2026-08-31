package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// corpusScanEnv gates the corpus measurement harness. The scan walks the entire
// repository SPEC corpus and is a measurement instrument, not a regression test,
// so it is skipped unless explicitly requested. CI cost is therefore zero.
const corpusScanEnv = "MOAI_T362_CORPUS_SCAN"

// corpusMeasurementRelPath is where the machine-readable measurement is written,
// relative to the repository root.
const corpusMeasurementRelPath = ".moai/reports/t362/m1-corpus-measurement.txt"

// corpusSpecGlobRel is the glob (relative to the repository root) that defines
// the scanned set. Recorded verbatim in the report so the scanned set is
// attributable.
const corpusSpecGlobRel = ".moai/specs/SPEC-*/spec.md"

// The repository root is located by findRepoRoot (drift_doctrine_test.go), which
// walks up to the directory containing go.mod.

type corpusPair struct {
	file  string
	reqID string
}

// corpusFileDetail is the per-file row emitted for every spec.md carrying at
// least one wide REQ definition line.
type corpusFileDetail struct {
	spec      string
	narrow    int
	wide      int
	uncovered int
	tier      string
}

// TestCorpusREQWideningMeasurement measures the blast radius of the widened REQ
// pattern against the real corpus using the real Go parser.
//
// AC-CRS-001-002 (maps REQ-CRS-001-001, REQ-CRS-001-004).
//
// It writes a machine-readable report to .moai/reports/t362/m1-corpus-measurement.txt
// and asserts nothing about the numbers — it is an instrument, and the numbers are
// the deliverable.
func TestCorpusREQWideningMeasurement(t *testing.T) {
	if os.Getenv(corpusScanEnv) != "1" {
		t.Skipf("corpus scan skipped; set %s=1 to run", corpusScanEnv)
	}

	root := findRepoRoot(t)
	glob := filepath.Join(root, corpusSpecGlobRel)
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("glob %q: %v", glob, err)
	}
	sort.Strings(paths)

	var (
		totalFiles      = len(paths)
		narrowFiles     int
		wideFiles       int
		narrowLines     int
		wideLines       int
		narrowNotInWide []corpusPair
		simCoverageTot  int
		invalidREQID    int
		perSpecCoverage = map[string]int{}
		perFile         []corpusFileDetail
		tierOfNewFile   = map[string]int{}
		newlyCollecting int
		parseErrors     []string
	)

	for _, p := range paths {
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			rel = p
		}

		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%s: %v", rel, doc.ParseError))
			continue
		}

		narrow := parseREQs(doc.Body)
		wide := parseREQsWide(doc.Body)

		narrowLines += len(narrow)
		wideLines += len(wide)
		if len(narrow) > 0 {
			narrowFiles++
		}
		if len(wide) > 0 {
			wideFiles++
		}
		if len(wide) > 0 && len(narrow) == 0 {
			newlyCollecting++
			tier := strings.TrimSpace(doc.Frontmatter.Tier)
			if tier == "" {
				tier = "(absent)"
			}
			tierOfNewFile[tier]++
		}

		// Containment: every (file, reqID) in NARROW must appear in WIDE.
		wideIDs := make(map[string]bool, len(wide))
		for _, r := range wide {
			wideIDs[r.ID] = true
		}
		for _, r := range narrow {
			if !wideIDs[r.ID] {
				narrowNotInWide = append(narrowNotInWide, corpusPair{
					file:  fmt.Sprintf("%s:%d", rel, r.Line),
					reqID: r.ID,
				})
			}
		}

		// Simulated CoverageIncomplete using the REAL Go AC path: doc.Criteria
		// comes from parseSPECDoc (findACSectionStart / extractACLines), compared
		// against the WIDE REQ set exactly as CoverageRule.Check does.
		if len(wide) > 0 {
			covered := collectAllREQIDs(doc.Criteria)
			uncovered := 0
			for _, r := range wide {
				if !covered[r.ID] {
					uncovered++
				}
			}
			specDir := filepath.Base(filepath.Dir(p))
			if uncovered > 0 {
				perSpecCoverage[specDir] = uncovered
				simCoverageTot += uncovered
			}
			tier := strings.TrimSpace(doc.Frontmatter.Tier)
			if tier == "" {
				tier = "(absent)"
			}
			perFile = append(perFile, corpusFileDetail{
				spec:      specDir,
				narrow:    len(narrow),
				wide:      len(wide),
				uncovered: uncovered,
				tier:      tier,
			})
		}

		// InvalidREQID the WIDE set would produce against the PRE-M2 validation
		// pattern, frozen as reqIDPatternPreM2.
		//
		// It must NOT read the live reqIDPattern. M1's figure is a statement
		// about the pre-M2 world, and reading the live pattern silently rewrote
		// it from 825 to 6 the moment M2 shipped — mutating committed evidence
		// that M1's own prose cites, under an unchanged label. Same defect the
		// Gate-0 harness carried; frozen here for the same reason.
		for _, r := range wide {
			if !reqIDPatternPreM2.MatchString(r.ID) {
				invalidREQID++
			}
		}
	}

	// Top 10 SPECs by simulated CoverageIncomplete count.
	type kv struct {
		spec string
		n    int
	}
	var top []kv
	for s, n := range perSpecCoverage {
		top = append(top, kv{s, n})
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].n != top[j].n {
			return top[i].n > top[j].n
		}
		return top[i].spec < top[j].spec
	})

	var tiers []string
	for tier := range tierOfNewFile {
		tiers = append(tiers, tier)
	}
	sort.Strings(tiers)

	var b strings.Builder
	fmt.Fprintf(&b, "# SPEC-COVERAGE-RULE-SCOPE-001 M1 — corpus REQ widening measurement\n")
	fmt.Fprintf(&b, "# produced by: MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/... -run TestCorpusREQWideningMeasurement -v\n")
	fmt.Fprintf(&b, "scan_root=%s\n", root)
	fmt.Fprintf(&b, "scan_glob=%s\n", corpusSpecGlobRel)
	fmt.Fprintf(&b, "\n[1] total spec.md scanned\n")
	fmt.Fprintf(&b, "total_spec_md=%d\n", totalFiles)
	fmt.Fprintf(&b, "parse_errors=%d\n", len(parseErrors))
	for _, e := range parseErrors {
		fmt.Fprintf(&b, "parse_error_item=%s\n", e)
	}
	fmt.Fprintf(&b, "\n[2] files with >=1 REQ definition line\n")
	fmt.Fprintf(&b, "files_narrow=%d\n", narrowFiles)
	fmt.Fprintf(&b, "files_wide=%d\n", wideFiles)
	fmt.Fprintf(&b, "\n[3] REQ definition LINE count\n")
	fmt.Fprintf(&b, "lines_narrow=%d\n", narrowLines)
	fmt.Fprintf(&b, "lines_wide=%d\n", wideLines)
	fmt.Fprintf(&b, "\n[4] containment: (file, reqID) in NARROW absent from WIDE\n")
	fmt.Fprintf(&b, "narrow_not_in_wide=%d\n", len(narrowNotInWide))
	for _, p := range narrowNotInWide {
		fmt.Fprintf(&b, "narrow_not_in_wide_item=%s %s\n", p.file, p.reqID)
	}
	fmt.Fprintf(&b, "\n[5] simulated CoverageIncomplete (real Go AC path vs WIDE REQ set)\n")
	fmt.Fprintf(&b, "sim_coverage_incomplete_total=%d\n", simCoverageTot)
	fmt.Fprintf(&b, "sim_coverage_incomplete_specs=%d\n", len(perSpecCoverage))
	for i, e := range top {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&b, "sim_top=%4d  %s\n", e.n, e.spec)
	}
	fmt.Fprintf(&b, "\n[6] InvalidREQID the WIDE set would produce against the frozen PRE-M2 pattern\n")
	fmt.Fprintf(&b, "measured_against_pattern=%s\n", reqIDPatternPreM2.String())
	fmt.Fprintf(&b, "invalid_req_id_count=%d\n", invalidREQID)
	fmt.Fprintf(&b, "\n[7] tier breakdown of newly-collecting files (wide>0 AND narrow==0)\n")
	fmt.Fprintf(&b, "newly_collecting_files=%d\n", newlyCollecting)
	for _, tier := range tiers {
		fmt.Fprintf(&b, "newly_collecting_tier=%s %d\n", tier, tierOfNewFile[tier])
	}
	fmt.Fprintf(&b, "\n[8] per-file detail for every file with >=1 wide REQ line\n")
	fmt.Fprintf(&b, "# format: perfile=<spec-dir> narrow=<n> wide=<n> uncovered=<n> tier=<v>\n")
	for _, d := range perFile {
		fmt.Fprintf(&b, "perfile=%s narrow=%d wide=%d uncovered=%d tier=%s\n",
			d.spec, d.narrow, d.wide, d.uncovered, d.tier)
	}

	out := filepath.Join(root, corpusMeasurementRelPath)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(out), err)
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write %q: %v", out, err)
	}

	t.Logf("measurement written to %s\n%s", out, b.String())
}
