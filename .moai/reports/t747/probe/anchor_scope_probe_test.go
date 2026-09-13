// zz_t747_anchor_probe_test.go — t747 plan-phase measurement probe.
// UNCOMMITTED measurement tool: this file is deleted before any commit; its
// outputs are pinned under .moai/reports/t747/probe/.
//
// RE-DERIVATION
//
//	T747_PROBE_OUT=<abs dir> go test ./internal/spec/ -run TestT747AnchorScope -v -count=1
//
// QUESTION (card t747, [HARD] first step). How large, and in which direction,
// is the in-section count of the section findACSectionStart anchors — against
// the t528 frozen-anchor baseline (216 accepted of 1167 in-section
// declarations) — i.e. is the anchor too NARROW (AC sections the vocabulary
// never names, declarations left outside) or too LOOSE (a heading elsewhere
// that merely contains a vocabulary word takes the anchor)?
//
// METHOD. Same corpus walk and same discriminator B (declRe, verbatim from
// zz_t528_anchor_probe_test.go) for comparability; per file, record the
// heading findACSectionStart anchors (line, level, text) and count
// declarations inside vs outside the anchored region. The in-section scan
// mirrors the t528 probe's shape (break on any "##"-prefixed line) so the
// numbers stay comparable with 216/1167; extractACLines' anchor-level break
// is a known second-order difference, recorded here rather than resolved.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestT747AnchorScope(t *testing.T) {
	root := "../../.moai/specs"
	out := os.Getenv("T747_PROBE_OUT")
	if out == "" {
		t.Fatal("T747_PROBE_OUT must name an output directory (evidence goes under .moai/reports/t747/probe/)")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	var readFiles []string
	var declAnywhereTotal, declInTotal, declOutTotal int
	var filesWithDecl, filesAnchored, filesAnchorMissed, filesEmptyAnchor, filesFallbackFirst int
	var anchorRows []string
	var missedFiles []string
	var emptyAnchorFiles []string
	var outsideTop []string

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		readFiles = append(readFiles, p)
		lines := strings.Split(string(b), "\n")

		// Whole-file declaration census (discriminator B).
		var anywhere []int
		for i, line := range lines {
			if declRe.FindStringSubmatch(line) != nil {
				anywhere = append(anywhere, i)
			}
		}

		start := findACSectionStart(lines)
		declIn := 0
		anchorDesc := "NONE"
		fallback := false
		if start < 0 {
			if len(anywhere) > 0 {
				filesAnchorMissed++
				missedFiles = append(missedFiles, fmt.Sprintf("%s\tdecls=%d", p, len(anywhere)))
			}
		} else {
			filesAnchored++
			anchorLevel := markdownHeadingLevel(strings.TrimSpace(lines[start-1]))
			anchorDesc = fmt.Sprintf("L%d:%s", anchorLevel, strings.TrimSpace(lines[start-1]))
			// Did the vocabulary's first candidate get skipped by the
			// non-empty-section preference? (findACSectionStart sets first,
			// then keeps looking for a non-empty one.)
			for i := start; i < len(lines); i++ {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "##") {
					break
				}
				if declRe.FindStringSubmatch(lines[i]) != nil {
					declIn++
				}
			}
			if len(anywhere) > 0 && declIn == 0 {
				filesEmptyAnchor++
				emptyAnchorFiles = append(emptyAnchorFiles, fmt.Sprintf("%s\t%s", p, anchorDesc))
			}
			_ = anchorLevel
			_ = fallback
		}
		// Fallback-first detection: recompute independently — if the anchored
		// section is empty and a LATER vocabulary heading has decls, the
		// anchor took the fallback. Approximation recorded per file.
		if start >= 0 && declIn == 0 && len(anywhere) > 0 {
			filesFallbackFirst++
		}

		declOut := len(anywhere) - declIn
		declAnywhereTotal += len(anywhere)
		declInTotal += declIn
		declOutTotal += declOut
		if len(anywhere) > 0 {
			filesWithDecl++
		}
		if declOut > 0 {
			outsideTop = append(outsideTop, fmt.Sprintf("%s\tin=%d\tout=%d\tanchor=%s", p, declIn, declOut, anchorDesc))
		}
		anchorRows = append(anchorRows, fmt.Sprintf("%s\t%s\tin=%d\tany=%d", p, anchorDesc, declIn, len(anywhere)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(readFiles)
	t528Write(t, out, "filelist.txt", readFiles)
	sort.Strings(anchorRows)
	t528Write(t, out, "anchors.txt", anchorRows)
	sort.Strings(missedFiles)
	t528Write(t, out, "anchor-missed.txt", missedFiles)
	sort.Strings(emptyAnchorFiles)
	t528Write(t, out, "empty-anchor.txt", emptyAnchorFiles)
	sort.Slice(outsideTop, func(i, j int) bool {
		a := strings.Count(outsideTop[i], "\t")
		b := strings.Count(outsideTop[j], "\t")
		if a != b {
			return false
		}
		return outsideTop[i] < outsideTop[j]
	})
	t528Write(t, out, "outside-decls.txt", outsideTop)

	t.Logf("OUTDIR = %s", out)
	t.Logf("DENOMINATOR spec.md read = %d (list: %s/filelist.txt)", len(readFiles), out)
	t.Logf("files with >=1 declaration (whole file) = %d", filesWithDecl)
	t.Logf("files anchored by findACSectionStart = %d", filesAnchored)
	t.Logf("files with declarations but NO anchor (narrow-miss) = %d (list: %s/anchor-missed.txt)", filesAnchorMissed, out)
	t.Logf("files anchored but 0 in-section while decls exist elsewhere (loose/empty anchor) = %d (list: %s/empty-anchor.txt)", filesEmptyAnchor, out)
	t.Logf("  of which the empty-first-section fallback likely took the anchor = %d", filesFallbackFirst)
	t.Logf("declarations WHOLE FILE = %d", declAnywhereTotal)
	t.Logf("declarations IN-SECTION (current anchor) = %d", declInTotal)
	t.Logf("declarations OUT-SECTION = %d (list: %s/outside-decls.txt)", declOutTotal, out)
}
