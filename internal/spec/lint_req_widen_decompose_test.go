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

// decomposeReportRelPath is where the Gate-0 decomposition is written, relative
// to the repository root.
const decomposeReportRelPath = ".moai/reports/t362/m2-gate0-decomposition.txt"

// Mechanical misread predicates.
//
// A "misread" is an extraction the widened pattern collects whose SOURCE LINE is
// not a REQ definition at all — a mapping row, a cross-reference list item, or a
// bullet that merely mentions a REQ. Each predicate is decidable on the line and
// its enclosing heading; none is an eyeball judgment.
//
// P1 exists in two readings. P1-raw over-fires (every corpus hit is a definition
// citing another REQ) and P1-corrected supersedes it. BOTH are reported and
// P1-raw's hits are enumerated in full: an adjudication that discards a
// predicate is only checkable if its input survives, and a report that emits
// only the corrected count reads identically whether the adjudication was sound
// or wrong. The misread verdict is the union of P1-corrected, P2 and P3.

// reqIDPatternPreM2 is the Gate-0 population definition, FROZEN as a literal.
//
// It is the validation pattern as it stood before SPEC-COVERAGE-RULE-SCOPE-001
// M2 widened `reqIDPattern`. Freezing it is not stylistic: the harness
// originally defined its rejected population as `!reqIDPattern.MatchString(...)`
// against the LIVE pattern, so when M2 shipped, the whole decomposition silently
// re-based from 825 rejections in three shape classes to 6 in one — same
// section headings, same file, different subject. A reader comparing this
// artifact against the prose that cites it would have found the 825, the three
// shape classes, and the 22 misread candidates simply gone, with nothing saying
// they had ever been measured.
//
// That is a measurement decided against a moving definition, which is the
// defect class this SPEC exists to document, reproduced inside its own evidence
// instrument. The literal below cannot move.
var reqIDPatternPreM2 = regexp.MustCompile(`^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`)

var (
	// P1-raw: the captured text names one or more FURTHER REQ tokens.
	//
	// The intent was to catch a mapping or index row. It does not: a REQ
	// definition may legitimately cite another REQ in its body, and all 22
	// corpus hits are exactly that. Retained and reported because the
	// adjudication that discarded it is only checkable if its input survives.
	reqTokenAnywhere = regexp.MustCompile(`REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+`)

	// P1-corrected narrows P1-raw mechanically rather than by reading. A
	// genuine mapping row carries REQ tokens and essentially nothing else; a
	// definition carries a requirement sentence around them. Strip every REQ
	// token plus punctuation and whitespace from the captured text: a mapping
	// row leaves nothing behind, a definition leaves its prose.
	//
	// This is what "corrected P1" means in the report — a stated predicate a
	// reader can disagree with, not a private judgment call.
	nonSubstantiveChars = regexp.MustCompile(`[\s\p{P}\p{S}]+`)

	// P3: the nearest preceding markdown heading names a section that by
	// construction does not hold requirement definitions.
	nonReqHeadingPattern = regexp.MustCompile(`(?i)(cross[- ]?reference|교차\s*참조|references?\b|gaps?\b|미검증|exclusions?\b|out of scope|범위 밖|history|이력|배경|background|anti-?pattern|반패턴|risk|위험|가정|glossary|용어|appendix|부록|측정|measurement|evidence|증거|checklist|체크리스트)`)

	headingPattern = regexp.MustCompile(`^\s{0,3}#{1,6}\s+(.*)$`)
)

type rejectedSample struct {
	loc  string
	line string
	id   string
}

type shapeClass struct {
	segs      int    // total '-'-separated segment count of the ID
	domain    string // "alpha" | "alnum"
	domainLen int    // number of segments between REQ and the numeric tail
	tailSegs  int    // number of trailing all-digit segments
	tailWidth string // digit widths of the tail, e.g. "3" or "3.3"
}

func (c shapeClass) key() string {
	return fmt.Sprintf("segs=%d domain=%s domainSegs=%d tailSegs=%d tailWidth=%s",
		c.segs, c.domain, c.domainLen, c.tailSegs, c.tailWidth)
}

func classifyREQID(id string) shapeClass {
	toks := strings.Split(id, "-")
	// toks[0] == "REQ" for anything the wide pattern produced.
	body := toks[1:]
	tail := 0
	for i := len(body) - 1; i >= 0; i-- {
		if isAllDigits(body[i]) {
			tail++
		} else {
			break
		}
	}
	var widths []string
	for i := len(body) - tail; i < len(body); i++ {
		widths = append(widths, fmt.Sprint(len(body[i])))
	}
	domainSegs := body[:len(body)-tail]
	alphabet := "alpha"
	for _, s := range domainSegs {
		if strings.ContainsAny(s, "0123456789") {
			alphabet = "alnum"
			break
		}
	}
	return shapeClass{
		segs:      len(toks),
		domain:    alphabet,
		domainLen: len(domainSegs),
		tailSegs:  tail,
		tailWidth: strings.Join(widths, "."),
	}
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// nearestHeading returns the nearest preceding markdown heading text for a
// 1-based line index, or "" when none precedes it.
func nearestHeading(lines []string, lineNo int) string {
	for i := lineNo - 2; i >= 0; i-- {
		if i >= len(lines) {
			continue
		}
		if m := headingPattern.FindStringSubmatch(lines[i]); m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func pctOf(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) * 100 / float64(d)
}

// reqIDPatternProposed is the M2 candidate validation pattern: strictly NARROWER
// than the widened extraction, so InvalidREQIDRule keeps a non-empty rejection
// class and cannot pass vacuously.
//
// Extraction accepts REQ-<[A-Z0-9]+ segments>-<digits>. This rejects, among
// others: a purely numeric domain segment (REQ-123-001), a domain of three or
// more segments (REQ-A-B-C-001), and a numeric tail that is not exactly one or
// two groups of three digits (REQ-ABC-1, REQ-ABC-0001, REQ-ABC-001-001-001).
var reqIDPatternProposed = regexp.MustCompile(`^REQ-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)?-\d{3}(?:-\d{3})?$`)

// reqIDPatternVacuousMutant is the option-(ii) mutant: validation aligned
// EXACTLY to reqLineWidePattern's capture group. It is measured against the live
// corpus alongside the shipped pattern so the claim "InvalidREQIDRule is not
// vacuous" is an observation rather than an inference from the shipped pattern's
// shape. A rule whose corpus firing count is identical under both is a rule
// whose rejection class no real document reaches.
var reqIDPatternVacuousMutant = regexp.MustCompile(`^REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+$`)

// measureWiringBlastRadius simulates EVERY doc.REQs consumer under the narrow
// and the wide extraction. CoverageRule is not the only consumer:
// EARSModalityRule emits ModalityMalformed at SeverityError, and
// REQIDUniquenessRule emits InvalidREQID / DuplicateREQID at SeverityError.
// None of those three codes is in eraDemotableCodes, so none is demoted on a
// grandfather-era SPEC.
func measureWiringBlastRadius(t *testing.T, paths []string, root string) string {
	t.Helper()

	type counts struct {
		modality   int
		legacy     int
		invalidNo  int // InvalidREQID under the shipped reqIDPattern (post-M2)
		invalidPre int // InvalidREQID under the frozen pre-M2 pattern
		invalidPr  int // InvalidREQID under reqIDPatternProposed
		invalidMu  int // InvalidREQID under reqIDPatternVacuousMutant (option (ii))
		dupCur     int // DuplicateREQID reachable under the CURRENT reqIDPattern
		dupPr      int // DuplicateREQID reachable under reqIDPatternProposed
		coverage   int
	}
	var narrowC, wideC counts
	var modalitySamples []rejectedSample
	var dupSamples []rejectedSample
	var invalidPrSamples []rejectedSample

	tally := func(doc *SPECDoc, reqs []REQEntry, c *counts, rel string, collect bool) {
		skip := map[string]bool{}
		for _, code := range doc.LintSkip {
			skip[code] = true
		}
		seenCur := map[string]int{}
		seenPr := map[string]int{}
		for _, r := range reqs {
			if isModalityMalformed(r.Text) && !skip["ModalityMalformed"] {
				c.modality++
				if collect && len(modalitySamples) < 20 {
					modalitySamples = append(modalitySamples,
						rejectedSample{loc: fmt.Sprintf("%s:%d", rel, r.Line), id: r.ID, line: r.Text})
				}
			}
			if isLegacyEARSPattern(r.Text) && !skip["LegacyEARSKeyword"] {
				c.legacy++
			}
			if !reqIDPattern.MatchString(r.ID) {
				if !skip["InvalidREQID"] {
					c.invalidNo++
				}
			} else if _, dup := seenCur[r.ID]; dup {
				if !skip["DuplicateREQID"] {
					c.dupCur++
				}
			} else {
				seenCur[r.ID] = r.Line
			}
			if !reqIDPatternVacuousMutant.MatchString(r.ID) && !skip["InvalidREQID"] {
				c.invalidMu++
			}
			if !reqIDPatternPreM2.MatchString(r.ID) && !skip["InvalidREQID"] {
				c.invalidPre++
			}
			if !reqIDPatternProposed.MatchString(r.ID) {
				if !skip["InvalidREQID"] {
					c.invalidPr++
					if collect && len(invalidPrSamples) < 30 {
						invalidPrSamples = append(invalidPrSamples,
							rejectedSample{loc: fmt.Sprintf("%s:%d", rel, r.Line), id: r.ID, line: r.Text})
					}
				}
			} else if first, dup := seenPr[r.ID]; dup {
				if !skip["DuplicateREQID"] {
					c.dupPr++
					if collect && len(dupSamples) < 20 {
						dupSamples = append(dupSamples, rejectedSample{
							loc:  fmt.Sprintf("%s:%d", rel, r.Line),
							id:   r.ID,
							line: fmt.Sprintf("duplicate of line %d", first),
						})
					}
				}
			} else {
				seenPr[r.ID] = r.Line
			}
		}
		if len(reqs) > 0 && !skip["CoverageIncomplete"] {
			covered := collectAllREQIDs(doc.Criteria)
			for _, r := range reqs {
				if !covered[r.ID] {
					c.coverage++
				}
			}
		}
	}

	for _, p := range paths {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		tally(doc, parseREQs(doc.Body), &narrowC, rel, false)
		tally(doc, parseREQsWide(doc.Body), &wideC, rel, true)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[F] FULL wiring blast radius — every doc.REQs consumer, narrow vs wide\n")
	fmt.Fprintf(&b, "# CoverageRule is NOT the only consumer. ModalityMalformed / InvalidREQID /\n")
	fmt.Fprintf(&b, "# DuplicateREQID / CoverageIncomplete are all SeverityError and NONE is in\n")
	fmt.Fprintf(&b, "# eraDemotableCodes, so none is demoted on a grandfather-era SPEC.\n")
	fmt.Fprintf(&b, "# lint.skip is honored; era demotion is not (it does not reach these codes).\n")
	fmt.Fprintf(&b, "blast_ModalityMalformed_error narrow=%d wide=%d\n", narrowC.modality, wideC.modality)
	fmt.Fprintf(&b, "blast_LegacyEARSKeyword_warning narrow=%d wide=%d\n", narrowC.legacy, wideC.legacy)
	fmt.Fprintf(&b, "# Each InvalidREQID row names the pattern literal it measured, because\n")
	fmt.Fprintf(&b, "# \"current\" moved across the M2 boundary: the same label read 825 before M2\n")
	fmt.Fprintf(&b, "# and 6 after. A label that does not carry its own attribution is not one.\n")
	fmt.Fprintf(&b, "blast_InvalidREQID_shippedPattern_postM2 narrow=%d wide=%d  pattern=%s\n",
		narrowC.invalidNo, wideC.invalidNo, reqIDPattern.String())
	fmt.Fprintf(&b, "blast_InvalidREQID_preM2Pattern narrow=%d wide=%d  pattern=%s\n",
		narrowC.invalidPre, wideC.invalidPre, reqIDPatternPreM2.String())
	fmt.Fprintf(&b, "blast_InvalidREQID_proposedPattern narrow=%d wide=%d  pattern=%s\n",
		narrowC.invalidPr, wideC.invalidPr, reqIDPatternProposed.String())
	fmt.Fprintf(&b, "\n# Corpus-level mutant probe (option (ii) vacuity check). The mutant aligns\n")
	fmt.Fprintf(&b, "# validation EXACTLY to the extraction. A shipped-vs-mutant delta of 0 would\n")
	fmt.Fprintf(&b, "# mean the shipped rejection class is unreachable in practice — i.e. the rule\n")
	fmt.Fprintf(&b, "# is vacuous on real documents whatever its regexp says.\n")
	fmt.Fprintf(&b, "blast_InvalidREQID_vacuousMutant narrow=%d wide=%d  pattern=%s\n",
		narrowC.invalidMu, wideC.invalidMu, reqIDPatternVacuousMutant.String())
	fmt.Fprintf(&b, "mutant_probe_delta_wide=%d\n", wideC.invalidPr-wideC.invalidMu)
	fmt.Fprintf(&b, "blast_DuplicateREQID_error_currentPattern narrow=%d wide=%d\n", narrowC.dupCur, wideC.dupCur)
	fmt.Fprintf(&b, "blast_DuplicateREQID_error_proposedPattern narrow=%d wide=%d\n", narrowC.dupPr, wideC.dupPr)
	fmt.Fprintf(&b, "blast_CoverageIncomplete_error narrow=%d wide=%d\n", narrowC.coverage, wideC.coverage)
	for _, e := range modalitySamples {
		fmt.Fprintf(&b, "modality_sample=%s %s\n   | %.160s\n", e.loc, e.id, e.line)
	}
	for _, e := range dupSamples {
		fmt.Fprintf(&b, "duplicate_sample=%s %s (%s)\n", e.loc, e.id, e.line)
	}
	for _, e := range invalidPrSamples {
		fmt.Fprintf(&b, "invalid_proposed_sample=%s %s\n   | %.160s\n", e.loc, e.id, e.line)
	}
	return b.String()
}

// TestCorpusRejectedREQIDDecomposition decomposes every ID the WIDE extraction
// collects that the CURRENT reqIDPattern rejects, plus the residual population
// that is narrow-VALID in shape yet NOT collected by the narrow LINE pattern.
//
// Gate 0 for M2. Like the M1 harness it asserts nothing about the numbers — it
// is an instrument, and the numbers are the deliverable.
func TestCorpusRejectedREQIDDecomposition(t *testing.T) {
	if os.Getenv(corpusScanEnv) != "1" {
		t.Skipf("corpus scan skipped; set %s=1 to run", corpusScanEnv)
	}

	root := findRepoRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, corpusSpecGlobRel))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	sort.Strings(paths)

	var (
		wideTotal      int
		rejected       int
		accepted       int
		shapeCount     = map[string]int{}
		shapeSamples   = map[string][]rejectedSample{}
		p1Count        int
		p1CorrCount    int
		p2Count        int
		p3Count        int
		unionCount     int
		p1Samples      []rejectedSample
		p1CorrSamples  []rejectedSample
		p2Samples      []rejectedSample
		p3Samples      []rejectedSample
		narrowValidNot []rejectedSample
		wideNotNarrow  int
		shippedReject  int
	)

	for _, p := range paths {
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			rel = p
		}
		doc := parseSPECDoc(p)
		if doc.ParseError != nil {
			continue
		}
		lines := strings.Split(doc.Body, "\n")
		wide := parseREQsWide(doc.Body)
		narrow := parseREQs(doc.Body)

		narrowAt := map[string]bool{}
		for _, r := range narrow {
			narrowAt[fmt.Sprintf("%d|%s", r.Line, r.ID)] = true
		}

		for _, r := range wide {
			wideTotal++
			src := ""
			if r.Line-1 >= 0 && r.Line-1 < len(lines) {
				src = strings.TrimRight(lines[r.Line-1], " \t\r")
			}
			s := rejectedSample{loc: fmt.Sprintf("%s:%d", rel, r.Line), line: src, id: r.ID}

			if !reqIDPattern.MatchString(r.ID) {
				shippedReject++
			}

			if !reqIDPatternPreM2.MatchString(r.ID) {
				rejected++
				k := classifyREQID(r.ID).key()
				shapeCount[k]++
				if len(shapeSamples[k]) < 3 {
					shapeSamples[k] = append(shapeSamples[k], s)
				}

				hit := false
				if n := len(reqTokenAnywhere.FindAllString(r.Text, -1)); n >= 1 {
					p1Count++
					hit = true
					if len(p1Samples) < 60 {
						p1Samples = append(p1Samples, s)
					}
					// P1-corrected: strip every REQ token, then every
					// punctuation/space rune. A mapping row leaves nothing.
					residue := nonSubstantiveChars.ReplaceAllString(
						reqTokenAnywhere.ReplaceAllString(r.Text, ""), "")
					if residue == "" {
						p1CorrCount++
						if len(p1CorrSamples) < 60 {
							p1CorrSamples = append(p1CorrSamples, s)
						}
					}
				}
				if strings.TrimSpace(r.Text) == "" {
					p2Count++
					hit = true
					if len(p2Samples) < 60 {
						p2Samples = append(p2Samples, s)
					}
				}
				if h := nearestHeading(lines, r.Line); h != "" && nonReqHeadingPattern.MatchString(h) {
					p3Count++
					hit = true
					if len(p3Samples) < 60 {
						p3Samples = append(p3Samples, s)
					}
				}
				if hit {
					unionCount++
				}
			} else {
				accepted++
				if !narrowAt[fmt.Sprintf("%d|%s", r.Line, r.ID)] {
					wideNotNarrow++
					if len(narrowValidNot) < 40 {
						narrowValidNot = append(narrowValidNot, s)
					}
				}
			}
		}
	}

	type kv struct {
		k string
		n int
	}
	var shapes []kv
	for k, n := range shapeCount {
		shapes = append(shapes, kv{k, n})
	}
	sort.Slice(shapes, func(i, j int) bool {
		if shapes[i].n != shapes[j].n {
			return shapes[i].n > shapes[j].n
		}
		return shapes[i].k < shapes[j].k
	})

	var b strings.Builder
	fmt.Fprintf(&b, "# SPEC-COVERAGE-RULE-SCOPE-001 M2 Gate 0 — rejected-REQ-ID decomposition\n")
	fmt.Fprintf(&b, "# produced by: MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/... -run TestCorpusRejectedREQIDDecomposition -v\n")
	fmt.Fprintf(&b, "scan_glob=%s\n", corpusSpecGlobRel)
	fmt.Fprintf(&b, "\n# POPULATION DEFINITION — frozen, and deliberately not the live pattern.\n")
	fmt.Fprintf(&b, "# Sections [A]-[E] decompose the ids the widened extraction collects that the\n")
	fmt.Fprintf(&b, "# PRE-M2 validation pattern rejected. That pattern is frozen as a literal here:\n")
	fmt.Fprintf(&b, "population_pattern_preM2=%s\n", reqIDPatternPreM2.String())
	fmt.Fprintf(&b, "# An earlier revision defined this population against the LIVE reqIDPattern, so\n")
	fmt.Fprintf(&b, "# shipping M2 silently re-based every count below from 825 to 6 under unchanged\n")
	fmt.Fprintf(&b, "# headings. The post-M2 residual is reported separately, at the end of [D].\n")
	fmt.Fprintf(&b, "shipped_pattern_now=%s\n", reqIDPattern.String())

	fmt.Fprintf(&b, "\n[A] totals (population: rejected by the frozen PRE-M2 pattern)\n")
	fmt.Fprintf(&b, "wide_extractions_total=%d\n", wideTotal)
	fmt.Fprintf(&b, "rejected_by_preM2Pattern=%d\n", rejected)
	fmt.Fprintf(&b, "accepted_by_preM2Pattern=%d\n", accepted)

	fmt.Fprintf(&b, "\n[B] shape histogram of REJECTED ids (frozen PRE-M2 population)\n")
	fmt.Fprintf(&b, "# key: segs=<total '-' segments>  domain=<alpha|alnum>  domainSegs=<n>  tailSegs=<n>  tailWidth=<digit widths>\n")
	for _, s := range shapes {
		fmt.Fprintf(&b, "shape=%-70s count=%d\n", s.k, s.n)
	}

	fmt.Fprintf(&b, "\n[C] up to 3 verbatim source lines per shape class\n")
	for _, s := range shapes {
		fmt.Fprintf(&b, "\n-- shape: %s (n=%d)\n", s.k, s.n)
		for _, e := range shapeSamples[s.k] {
			fmt.Fprintf(&b, "   %s\n   | %s\n", e.loc, e.line)
		}
	}

	fmt.Fprintf(&b, "\n[D] misread predicates (each decidable on the source line)\n")
	fmt.Fprintf(&b, "# A misread is an extraction whose SOURCE LINE is not a REQ definition at all.\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# P1-raw       = captured text names >=1 FURTHER REQ token.\n")
	fmt.Fprintf(&b, "#                Intended to catch a mapping/index row. It does NOT: a definition\n")
	fmt.Fprintf(&b, "#                may legitimately cite another REQ in its body. Reported, and every\n")
	fmt.Fprintf(&b, "#                hit enumerated below, so the narrowing that discarded it is checkable.\n")
	fmt.Fprintf(&b, "# P1-corrected = of the P1-raw hits, those whose captured text is NOTHING BUT REQ\n")
	fmt.Fprintf(&b, "#                tokens once punctuation and whitespace are stripped — a real mapping\n")
	fmt.Fprintf(&b, "#                row leaves an empty residue, a definition leaves its prose.\n")
	fmt.Fprintf(&b, "# P2           = captured text empty after trim (no requirement body)\n")
	fmt.Fprintf(&b, "# P3           = nearest preceding heading matches a non-requirements section pattern\n")
	fmt.Fprintf(&b, "#\n")
	fmt.Fprintf(&b, "# The misread verdict is the UNION of P1-corrected, P2 and P3. P1-raw is a\n")
	fmt.Fprintf(&b, "# superseded candidate and is NOT counted in it.\n")
	fmt.Fprintf(&b, "misread_p1_raw_extra_req_tokens=%d\n", p1Count)
	fmt.Fprintf(&b, "misread_p1_corrected_tokens_only=%d\n", p1CorrCount)
	fmt.Fprintf(&b, "misread_p2_empty_text=%d\n", p2Count)
	fmt.Fprintf(&b, "misread_p3_non_req_heading=%d\n", p3Count)
	fmt.Fprintf(&b, "misread_verdict_union_corrected=%d\n", p1CorrCount+p2Count+p3Count)
	fmt.Fprintf(&b, "misread_union_raw_upper_bound=%d\n", unionCount)
	fmt.Fprintf(&b, "misread_verdict_pct_of_rejected=%.1f\n", pctOf(p1CorrCount+p2Count+p3Count, rejected))

	fmt.Fprintf(&b, "\n-- every P1-raw hit, enumerated (n=%d)\n", p1Count)
	for _, e := range p1Samples {
		fmt.Fprintf(&b, "p1_raw=%s %s\n   | %s\n", e.loc, e.id, e.line)
	}
	fmt.Fprintf(&b, "\n-- every P1-corrected hit, enumerated (n=%d)\n", p1CorrCount)
	for _, e := range p1CorrSamples {
		fmt.Fprintf(&b, "p1_corrected=%s %s\n   | %s\n", e.loc, e.id, e.line)
	}
	fmt.Fprintf(&b, "\n-- post-M2 residual, reported here so it is not mistaken for the population above\n")
	fmt.Fprintf(&b, "rejected_by_shippedPattern_postM2=%d\n", shippedReject)
	for _, e := range p2Samples {
		fmt.Fprintf(&b, "misread_sample_P2=%s\n   | %s\n", e.loc, e.line)
	}
	for _, e := range p3Samples {
		fmt.Fprintf(&b, "misread_sample_P3=%s\n   | %s\n", e.loc, e.line)
	}

	fmt.Fprintf(&b, "\n[E] narrow-shape-VALID ids the NARROW LINE pattern did NOT collect\n")
	fmt.Fprintf(&b, "wide_accepted_not_collected_by_narrow_line=%d\n", wideNotNarrow)
	for _, e := range narrowValidNot {
		fmt.Fprintf(&b, "residual=%s %s\n   | %s\n", e.loc, e.id, e.line)
	}

	fmt.Fprintf(&b, "\n%s", measureWiringBlastRadius(t, paths, root))

	out := filepath.Join(root, decomposeReportRelPath)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Logf("decomposition written to %s", out)
}
