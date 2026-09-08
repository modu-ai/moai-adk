// zz_t528_anchor_probe_test.go — t528 / SPEC-AC-COLLECTOR-ANCHOR-001 corpus
// measurement probe, promoted from .moai/reports/t528/probe/ac_anchor_probe_test.go
// into the tree as a COMMITTED test (REQ-ACA-001-014).
//
// RE-DERIVATION
//
//	go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s
//
// DISCRIMINATOR. declRe below IS discriminator B — the canonical one. It is a
// verbatim copy of the plan-phase artifact's declRe: a bullet is REQUIRED, dots
// are allowed, and an id may end in a letter. Do not edit it; the numbers in
// the report are only comparable while it is byte-identical to the artifact.
//
// TWO ACCEPTANCE COLUMNS, ONE PROCESS. baselineRe is a FROZEN copy of the
// pre-widening parseSingleACLine anchor; parseSingleACLine itself is the live
// column. Before the widening the two agree (that agreement is what makes the
// live column a valid re-measurement of the same quantity); after it, the
// frozen column keeps re-deriving the 216 baseline in the SAME run as the new
// figure, so a comparison never crosses two runs or two trees.
//
// DENOMINATOR. The corpus is self-modifying — authoring this card's own SPEC
// added one spec.md. The probe therefore writes the exact file list it read,
// so the denominator is fixed by an artifact rather than by a later `find`.
//
// OUTPUT DIRECTORY. Set T528_PROBE_OUT to redirect. The default is a run-scoped
// directory, NOT the plan-phase artifact directory: the pinned before-image
// (probe/positive-needle.txt, probe/filelist.txt) must never be overwritten by
// a post-widening run, or the no-regression control becomes a comparison of the
// widened parser against itself.
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

// declRe — an AC DECLARATION line (discriminator B). A bullet is REQUIRED, so a
// prose sentence mentioning an AC id at column 0 is excluded.
var declRe = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)

// baselineRe — FROZEN copy of what parseSingleACLine accepted BEFORE the t528
// widening, applied after its own strings.TrimLeft(trimmed, "- *"). Never
// update this to track parser.go: it is the baseline anchor, not a mirror.
var baselineRe = regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)

// numericTailRe — the plan §B.1 decision: last segment numeric, optional
// .a / .a.i sub-id suffix preserved.
var numericTailRe = regexp.MustCompile(`^AC-(?:[A-Za-z0-9]+-)*[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?$`)

func t528ProbeOutDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("T528_PROBE_OUT")
	if dir == "" {
		dir = "../../.moai/reports/t528/probe/out"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func t528Write(t *testing.T, dir, name string, lines []string) {
	t.Helper()
	body := ""
	if len(lines) > 0 {
		body = strings.Join(lines, "\n") + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// t528TreeShape renders one root and its children as ID(child,child) so a
// comparison sees the tree shape, not only the root id set.
func t528TreeShape(ac Acceptance) string {
	if len(ac.Children) == 0 {
		return ac.ID
	}
	var kids []string
	for _, c := range ac.Children {
		kids = append(kids, t528TreeShape(c))
	}
	return ac.ID + "(" + strings.Join(kids, ",") + ")"
}

func TestT528Anchor(t *testing.T) {
	root := "../../.moai/specs"
	out := t528ProbeOutDir(t)

	var readFiles []string
	var declTotal, acceptedLive, rejectedLive, acceptedBase int
	var tailCovered, tailUncovered int
	uncoveredShapes := map[string]int{}
	sepShape := map[string]int{}
	liveFiles := map[string]bool{}
	baseFiles := map[string]bool{}
	rejectedFiles := map[string]bool{}
	var newlyAccepted []string // accepted by the live parser, rejected by the frozen baseline
	var stillRejected []string
	var rootAC, reqMap []string

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		readFiles = append(readFiles, p)

		// Per-file parse result — the M0 before-image / M3 after-image basis.
		criteria, _ := ParseAcceptanceCriteria(string(b), false)
		var shapes []string
		for _, c := range criteria {
			shapes = append(shapes, t528TreeShape(c))
			if len(c.RequirementIDs) > 0 {
				reqMap = append(reqMap, fmt.Sprintf("%s\t%s\t%s", p, c.ID, strings.Join(c.RequirementIDs, ",")))
			}
			for _, ch := range c.Children {
				if len(ch.RequirementIDs) > 0 {
					reqMap = append(reqMap, fmt.Sprintf("%s\t%s\t%s", p, ch.ID, strings.Join(ch.RequirementIDs, ",")))
				}
			}
		}
		rootAC = append(rootAC, fmt.Sprintf("%s\t%s", p, strings.Join(shapes, ";")))

		lines := strings.Split(string(b), "\n")
		start := findACSectionStart(lines)
		if start < 0 {
			return nil
		}
		for i := start; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "##") {
				break
			}
			m := declRe.FindStringSubmatch(lines[i])
			if m == nil {
				continue
			}
			declTotal++
			id, rest := m[1], strings.TrimSpace(m[2])

			sep := "«none/other»"
			for _, c := range []string{":", "—", "–", "(", "-"} {
				if strings.HasPrefix(rest, c) {
					sep = c
					break
				}
			}
			sepShape[sep]++

			if numericTailRe.MatchString(id) {
				tailCovered++
			} else {
				tailUncovered++
				uncoveredShapes[regexp.MustCompile(`[0-9]+`).ReplaceAllString(id, "N")]++
			}

			trimmed := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(lines[i]), "- *"))
			base := baselineRe.MatchString(trimmed)
			live := parseSingleACLine(strings.TrimSpace(lines[i])) != nil

			if base {
				acceptedBase++
				baseFiles[p] = true
			}
			if live {
				acceptedLive++
				liveFiles[p] = true
			} else {
				rejectedLive++
				rejectedFiles[p] = true
				stillRejected = append(stillRejected, fmt.Sprintf("%s:%d:%s", p, i+1, strings.TrimSpace(lines[i])))
			}
			if live && !base {
				newlyAccepted = append(newlyAccepted, fmt.Sprintf("%s:%d:%s", p, i+1, strings.TrimSpace(lines[i])))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(readFiles)
	t528Write(t, out, "filelist.txt", readFiles)

	var pos []string
	for f := range liveFiles {
		pos = append(pos, f)
	}
	sort.Strings(pos)
	t528Write(t, out, "positive-needle.txt", pos)

	var basePos []string
	for f := range baseFiles {
		basePos = append(basePos, f)
	}
	sort.Strings(basePos)
	t528Write(t, out, "baseline-needle.txt", basePos)

	// FULLY-BLIND ROSTER. The count alone is not pickable: an AC that must name
	// a blind SPEC cannot do so from a number.
	var blindList []string
	for f := range rejectedFiles {
		if !liveFiles[f] {
			blindList = append(blindList, f)
		}
	}
	sort.Strings(blindList)
	t528Write(t, out, "blind-files.txt", blindList)
	blind := len(blindList)

	sort.Strings(rootAC)
	t528Write(t, out, "rootac.txt", rootAC)
	sort.Strings(reqMap)
	t528Write(t, out, "reqmap.txt", reqMap)
	t528Write(t, out, "newly-accepted.txt", newlyAccepted)
	t528Write(t, out, "still-rejected.txt", stillRejected)

	t.Logf("OUTDIR = %s", out)
	t.Logf("DENOMINATOR spec.md read = %d (list: %s/filelist.txt)", len(readFiles), out)
	t.Logf("IN-SECTION declarations = %d", declTotal)
	t.Logf("  accepted by FROZEN baseline anchor = %d", acceptedBase)
	t.Logf("  accepted by LIVE parseSingleACLine = %d", acceptedLive)
	t.Logf("  rejected by LIVE parseSingleACLine = %d", rejectedLive)
	t.Logf("  NEWLY accepted (live yes, baseline no) = %d (list: %s/newly-accepted.txt)", len(newlyAccepted), out)
	t.Logf("POSITIVE-NEEDLE files (live) = %d (list: %s/positive-needle.txt)", len(pos), out)
	t.Logf("BASELINE-NEEDLE files (frozen anchor) = %d (list: %s/baseline-needle.txt)", len(basePos), out)
	t.Logf("FULLY-BLIND files (>=1 decl, 0 accepted by live) = %d", blind)
	t.Logf("B.1 numeric-tail candidate: covered = %d  UNCOVERED = %d", tailCovered, tailUncovered)
	for k, v := range uncoveredShapes {
		t.Logf("  UNCOVERED-SHAPE %-24s %d", k, v)
	}
	type kv struct {
		k string
		v int
	}
	var s []kv
	for k, v := range sepShape {
		s = append(s, kv{k, v})
	}
	sort.Slice(s, func(i, j int) bool { return s[i].v > s[j].v })
	for _, e := range s {
		t.Logf("  SEP %-16s %d", e.k, e.v)
	}
}
