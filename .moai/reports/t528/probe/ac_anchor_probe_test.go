// ac_anchor_probe_test.go — t528 plan-phase measurement probe.
//
// RE-DERIVATION. This file IS the discriminator. Copy it to
// internal/spec/zz_probe_test.go and run:
//
//	go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s
//
// It is kept here (not in internal/) so the plan-phase tree carries no
// uncommitted test, while the figures in the report stay re-derivable by
// anyone. The run phase promotes it to a committed test.
//
// DENOMINATOR. The corpus is self-modifying: authoring this card's own SPEC
// added one spec.md. The probe therefore writes the exact file list it read
// to filelist.txt, so the denominator is fixed by an artifact rather than by
// a `find` re-run at an unknown later time.
package spec

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// declRe — an AC DECLARATION line. A bullet is REQUIRED, so a prose sentence
// mentioning an AC id at column 0 is excluded. (The first measurement pass
// omitted this and counted prose as declarations; see the report's correction.)
var declRe = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)

// currentRe — what parseSingleACLine accepts today, applied after its own
// strings.TrimLeft(trimmed, "- *").
var currentRe = regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)

// numericTailRe — the §B.1 candidate decision under test: last segment numeric,
// optional .a / .a.i sub-id suffix preserved.
var numericTailRe = regexp.MustCompile(`^AC-(?:[A-Za-z0-9]+-)*[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?$`)

func TestT528Anchor(t *testing.T) {
	root := "../../.moai/specs"

	var readFiles []string
	var declTotal, accepted, rejected int
	var tailCovered, tailUncovered int
	uncoveredShapes := map[string]int{}
	sepShape := map[string]int{}
	acceptedFiles := map[string]bool{}
	rejectedFiles := map[string]bool{}

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

			// §B.1 candidate: does a numeric last segment cover this id?
			if numericTailRe.MatchString(id) {
				tailCovered++
			} else {
				tailUncovered++
				uncoveredShapes[regexp.MustCompile(`[0-9]+`).ReplaceAllString(id, "N")]++
			}

			trimmed := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(lines[i]), "- *"))
			if currentRe.MatchString(trimmed) {
				accepted++
				acceptedFiles[p] = true
			} else {
				rejected++
				rejectedFiles[p] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(readFiles)
	if e := os.WriteFile("../../.moai/reports/t528/probe/filelist.txt",
		[]byte(strings.Join(readFiles, "\n")+"\n"), 0o644); e != nil {
		t.Fatal(e)
	}

	var pos []string
	for f := range acceptedFiles {
		pos = append(pos, f)
	}
	sort.Strings(pos)
	if e := os.WriteFile("../../.moai/reports/t528/probe/positive-needle.txt",
		[]byte(strings.Join(pos, "\n")+"\n"), 0o644); e != nil {
		t.Fatal(e)
	}

	// FULLY-BLIND ROSTER. The count alone is not pickable: an AC that must name
	// a blind SPEC cannot do so from a number. Emitted as a list for the same
	// reason positive-needle.txt is — a roster is re-derivable, a count is not.
	var blindList []string
	for f := range rejectedFiles {
		if !acceptedFiles[f] {
			blindList = append(blindList, f)
		}
	}
	sort.Strings(blindList)
	if e := os.WriteFile("../../.moai/reports/t528/probe/blind-files.txt",
		[]byte(strings.Join(blindList, "\n")+"\n"), 0o644); e != nil {
		t.Fatal(e)
	}
	blind := len(blindList)

	t.Logf("DENOMINATOR spec.md read = %d (list: .moai/reports/t528/probe/filelist.txt)", len(readFiles))
	t.Logf("IN-SECTION declarations = %d", declTotal)
	t.Logf("  accepted by current parser = %d", accepted)
	t.Logf("  rejected                   = %d", rejected)
	t.Logf("POSITIVE-NEEDLE files = %d (list: .moai/reports/t528/probe/positive-needle.txt)", len(pos))
	t.Logf("FULLY-BLIND files (>=1 decl, 0 accepted) = %d", blind)
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
