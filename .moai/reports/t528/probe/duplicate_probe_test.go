package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// zz_probe3_test.go — t528: duplicate-ID exposure under the ACTUAL §B.1
// candidate grammar, not under a loose scanning regex.
//
// CORRECTION RECORD. A first pass used a scanning regex whose id class omitted
// `.` and required a digit-final char. It captured `AC-LCLN-001` out of
// `AC-LCLN-001.1/.2/.3` and `AC-HFC-001` out of `AC-HFC-001a/001b`, collapsing
// distinct sub-ids into false duplicates and reporting "4 files / 24 dropped".
// Those ids are DISTINCT in the corpus and are, moreover, among the 39 the
// §B.1 candidate does not admit at all. The figure was an artifact of the
// discriminator. This probe admits a line only if the candidate grammar would.

// declScan — permissive line finder. Its ONLY job is to locate a candidate
// bullet and hand the raw id to the admission test below; it never decides.
var declScan = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)

// admitRe — the §B.1 candidate: variable segments, alphabetic middles allowed,
// LAST segment numeric, existing `.a` / `.a.i` sub-id suffix preserved.
var admitRe = regexp.MustCompile(`^AC-(?:[A-Za-z0-9]+-)*[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?$`)

// currentAdmit — what the parser accepts today, post TrimLeft("- *").
var currentAdmit = regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)

func TestT528DuplicateExposure(t *testing.T) {
	root := "../../.moai/specs"
	var filesToday, dropToday, filesWide, dropWide int
	var samples []string

	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		lines := strings.Split(string(b), "\n")
		start := findACSectionStart(lines)
		if start < 0 {
			return nil
		}
		seenToday := map[string]int{}
		seenWide := map[string]int{}
		for i := start; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "##") {
				break
			}
			m := declScan.FindStringSubmatch(lines[i])
			if m == nil {
				continue
			}
			id := m[1]
			if admitRe.MatchString(id) {
				seenWide[id]++
			}
			trimmed := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(lines[i]), "- *"))
			if currentAdmit.MatchString(trimmed) {
				seenToday[id]++
			}
		}
		dt, dw := 0, 0
		var names []string
		for _, n := range seenToday {
			if n > 1 {
				dt += n - 1
			}
		}
		for id, n := range seenWide {
			if n > 1 {
				dw += n - 1
				names = append(names, id)
			}
		}
		if dt > 0 {
			filesToday++
			dropToday += dt
		}
		if dw > 0 {
			filesWide++
			dropWide += dw
			if len(samples) < 8 {
				sort.Strings(names)
				samples = append(samples, p+"  dup="+strings.Join(names, ","))
			}
		}
		return nil
	})

	t.Logf("TODAY   (current parser grammar) : files with duplicate ids = %d, dropped lines = %d", filesToday, dropToday)
	t.Logf("WIDENED (B.1 candidate grammar)  : files with duplicate ids = %d, dropped lines = %d", filesWide, dropWide)
	for _, s := range samples {
		t.Logf("  %s", s)
	}
}
