package spec

// zz_t565_anchor_after_census_test.go — card t565 corpus census of the FIXED
// heading anchor and section end, side by side with a re-implementation of the
// unfixed rule and with a counterfactual vocabulary. The committed copy lives
// under .moai/reports/t565/probe/; it is copied into internal/spec only for the
// run and removed afterwards. It writes nothing; every figure is a t.Logf line.
//
// RE-DERIVATION
//
//	cp .moai/reports/t565/probe/zz_t565_anchor_after_census_test.go internal/spec/
//	cp .moai/reports/t565/probe/zz_t565_heading_census_test.go internal/spec/   # t565SortedKeys, t565Itoa
//	go test ./internal/spec -count=1 -run '^TestT565AnchorAfterCensus$' -v
//	rm internal/spec/zz_t565_anchor_after_census_test.go internal/spec/zz_t565_heading_census_test.go
//
// CONTROL. The re-implemented unfixed rule must reproduce the before census
// (probe/census-develop-before.log): anchor kinds h2/word 319, h2/file+word 29,
// h2/file-only 10, h3+/file-only 4, h3+/word 5, none 467, and in-section
// parser-accepted lines 994 + 101 + 5 = 1100. A mismatch means the old-rule
// columns below measure something other than the parser that shipped.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// t565aExtendedVocabulary is the counterfactual: the ruled list plus the three
// phrases the before census showed on file-only headings the ruled list misses.
var t565aExtendedVocabulary = append(append([]string{}, acSectionVocabulary...), "수용 기준", "성공 기준", "ac summary")

func t565aOldSection(lines []string) (int, int) {
	start := -1
	for i, l := range lines {
		tr := strings.TrimSpace(l)
		if strings.HasPrefix(tr, "##") && strings.Contains(strings.ToLower(tr), "acceptance") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return -1, -1
	}
	end := start
	for ; end < len(lines); end++ {
		if strings.HasPrefix(strings.TrimSpace(lines[end]), "##") {
			break
		}
	}
	return start, end
}

func t565aHeadingWith(trimmed string, vocab []string) bool {
	if markdownHeadingLevel(trimmed) < 2 {
		return false
	}
	text := strings.ReplaceAll(strings.ToLower(trimmed), "acceptance.md", "")
	for _, m := range acNegativeSectionMarkers {
		if strings.Contains(text, m) {
			return false
		}
	}
	for _, p := range vocab {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

func t565aStartWith(lines []string, vocab []string) int {
	for i, l := range lines {
		if t565aHeadingWith(strings.TrimSpace(l), vocab) {
			return i + 1
		}
	}
	return -1
}

func t565aEnd(lines []string, start int) int {
	level := markdownHeadingLevel(strings.TrimSpace(lines[start-1]))
	end := start
	for ; end < len(lines); end++ {
		if l := markdownHeadingLevel(strings.TrimSpace(lines[end])); l > 0 && l <= level {
			break
		}
	}
	return end
}

func t565aLive(lines []string, start, end int) int {
	n := 0
	for i := start; i < end; i++ {
		if tr := strings.TrimSpace(lines[i]); tr != "" && parseSingleACLine(tr) != nil {
			n++
		}
	}
	return n
}

func t565aQualifyingStarts(lines []string, vocab []string) []int {
	var starts []int
	for i, l := range lines {
		if t565aHeadingWith(strings.TrimSpace(l), vocab) {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func t565aOldKind(h string) string {
	level := "h2"
	if !strings.HasPrefix(h, "## ") {
		level = "h3+"
	}
	lower := strings.ToLower(h)
	file := strings.Contains(lower, "acceptance.md")
	word := strings.Contains(strings.ReplaceAll(lower, "acceptance.md", ""), "acceptance")
	switch {
	case file && !word:
		return level + "/file-only"
	case file && word:
		return level + "/file+word"
	default:
		return level + "/word"
	}
}

func TestT565AnchorAfterCensus(t *testing.T) {
	root := "../../.moai/specs"
	var files, oldAnch, newAnch, extAnch, oldLive, newLive, extLive int
	oldKinds := map[string]int{}
	oldKindLive := map[string]int{}
	var changed, fileOnly, deepNew, laterLost, negExcluded, extChanged []string
	var laterLostFiles, laterLostLive, changedCount, extChangedCount, extChangedLiveDelta, oldNotBullet int
	var skipLive, unionLive, comboLive int
	var skipChanged, unionChanged, comboChanged, ruledLoss, comboLoss []string
	var reqGained, reqLost []string
	var oldReqTotal, newReqTotal int

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		files++
		rel := strings.TrimPrefix(p, "../../")
		lines := strings.Split(string(b), "\n")

		oStart, oEnd := t565aOldSection(lines)
		oHeading, oKind, oN := "", "none", 0
		if oStart >= 0 {
			oldAnch++
			oHeading = strings.TrimSpace(lines[oStart-1])
			oKind = t565aOldKind(oHeading)
			oN = t565aLive(lines, oStart, oEnd)
			oldLive += oN
			oldKindLive[oKind] += oN
			// The before census counted only bullet declaration-shaped lines
			// (t565DeclRe); this census counts every line the parser accepts.
			for i := oStart; i < oEnd; i++ {
				if tr := strings.TrimSpace(lines[i]); tr != "" && parseSingleACLine(tr) != nil && !t565DeclRe.MatchString(lines[i]) {
					oldNotBullet++
				}
			}
		}
		oldKinds[oKind]++

		nStart := findACSectionStart(lines)
		nHeading, nN := "", 0
		if nStart >= 0 {
			newAnch++
			nHeading = strings.TrimSpace(lines[nStart-1])
			nN = len(extractACLines(lines, nStart, false))
			newLive += nN
			if markdownHeadingLevel(nHeading) >= 3 {
				deepNew = append(deepNew, rel+"  "+nHeading+"  live="+t565Itoa(nN))
			}
			// A later heading that also names the section is never read: only
			// the first anchor is.
			for j := t565aEnd(lines, nStart); j < len(lines); j++ {
				tr := strings.TrimSpace(lines[j])
				if !isACSectionHeading(tr) || markdownHeadingLevel(tr) < 2 {
					continue
				}
				if k := t565aLive(lines, j+1, t565aEnd(lines, j+1)); k > 0 {
					laterLostFiles++
					laterLostLive += k
					laterLost = append(laterLost, rel+"  anchor: "+nHeading+"  |  later: "+tr+"  live="+t565Itoa(k))
				}
			}
		}

		// COUNTERFACTUAL selection policies over the ruled vocabulary:
		// skip-empty takes the first section-naming heading whose section the
		// parser reads lines from; union reads every such section (lines deduped).
		starts := t565aQualifyingStarts(lines, acSectionVocabulary)
		sN := 0
		for _, s := range starts {
			if n := len(extractACLines(lines, s, false)); n > 0 {
				sN = n
				break
			}
		}
		skipLive += sN
		if sN != nN {
			skipChanged = append(skipChanged, rel+"  ruled[live="+t565Itoa(nN)+"] -> skip-empty[live="+t565Itoa(sN)+"]")
		}
		seen := map[int]bool{}
		for _, s := range starts {
			for i := s; i < t565aEnd(lines, s); i++ {
				if tr := strings.TrimSpace(lines[i]); tr != "" && parseSingleACLine(tr) != nil {
					seen[i] = true
				}
			}
		}
		unionLive += len(seen)
		if len(seen) != nN {
			unionChanged = append(unionChanged, rel+"  ruled[live="+t565Itoa(nN)+"] -> union[live="+t565Itoa(len(seen))+"]")
		}

		cN := 0
		for _, s := range t565aQualifyingStarts(lines, t565aExtendedVocabulary) {
			if n := len(extractACLines(lines, s, false)); n > 0 {
				cN = n
				break
			}
		}
		comboLive += cN
		if cN != nN {
			comboChanged = append(comboChanged, rel+"  ruled[live="+t565Itoa(nN)+"] -> skip-empty+extended[live="+t565Itoa(cN)+"]")
		}
		if nN < oN {
			ruledLoss = append(ruledLoss, rel+"  old="+t565Itoa(oN)+" ruled="+t565Itoa(nN))
		}
		if cN < oN {
			comboLoss = append(comboLoss, rel+"  old="+t565Itoa(oN)+" skip-empty+extended="+t565Itoa(cN))
		}

		// REQ mappings the read lines carry, old rule vs ruled: the coverage rule
		// only moves when this set moves.
		oldReqs := map[string]bool{}
		if oStart >= 0 {
			for i := oStart; i < oEnd; i++ {
				if tr := strings.TrimSpace(lines[i]); tr != "" {
					if p := parseSingleACLine(tr); p != nil {
						for _, r := range p.reqIDs {
							oldReqs[r] = true
						}
					}
				}
			}
		}
		newReqs := map[string]bool{}
		if nStart >= 0 {
			for _, l := range extractACLines(lines, nStart, false) {
				for _, r := range l.reqIDs {
					newReqs[r] = true
				}
			}
		}
		oldReqTotal += len(oldReqs)
		newReqTotal += len(newReqs)
		for r := range newReqs {
			if !oldReqs[r] {
				reqGained = append(reqGained, rel+"\t"+r)
			}
		}
		for r := range oldReqs {
			if !newReqs[r] {
				reqLost = append(reqLost, rel+"\t"+r)
			}
		}

		if oStart != nStart || oN != nN {
			changedCount++
			changed = append(changed, rel+"  old["+oKind+" live="+t565Itoa(oN)+"]: "+oHeading+"  ->  new[live="+t565Itoa(nN)+"]: "+nHeading)
		}

		eStart := t565aStartWith(lines, t565aExtendedVocabulary)
		eN := 0
		if eStart >= 0 {
			extAnch++
			eN = len(extractACLines(lines, eStart, false))
			extLive += eN
		}
		if eStart != nStart || eN != nN {
			extChangedCount++
			extChangedLiveDelta += eN - nN
			eHeading := ""
			if eStart >= 0 {
				eHeading = strings.TrimSpace(lines[eStart-1])
			}
			extChanged = append(extChanged, rel+"  ruled[live="+t565Itoa(nN)+"]: "+nHeading+"  ->  extended[live="+t565Itoa(eN)+"]: "+eHeading)
		}

		if oKind == "h2/file-only" {
			fileOnly = append(fileOnly, rel+"  old: "+oHeading+"  |  new: "+nHeading+" (live="+t565Itoa(nN)+")  |  caught-by-ruled-vocabulary="+
				map[bool]string{true: "yes", false: "no"}[nStart == oStart])
		}

		for _, l := range lines {
			tr := strings.TrimSpace(l)
			if markdownHeadingLevel(tr) < 2 {
				continue
			}
			text := strings.ReplaceAll(strings.ToLower(tr), "acceptance.md", "")
			neg := false
			for _, m := range acNegativeSectionMarkers {
				neg = neg || strings.Contains(text, m)
			}
			if !neg {
				continue
			}
			for _, ph := range t565aExtendedVocabulary {
				if strings.Contains(text, ph) {
					negExcluded = append(negExcluded, rel+"  "+tr)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 800 {
		t.Fatalf("walked only %d spec.md files; a short walk makes every count meaningless", files)
	}

	t.Logf("spec.md walked = %d", files)
	t.Logf("CONTROL old-rule anchor kinds:")
	for _, k := range t565SortedKeys(oldKinds) {
		t.Logf("  %-16s files=%-4d in-section parser-accepted=%d", k, oldKinds[k], oldKindLive[k])
	}
	t.Logf("CONTROL old-rule in-section parser-accepted lines that are not bullet declaration-shaped = %d (before census counted bullet-shaped only: 1100)", oldNotBullet)
	t.Logf("anchored files: old=%d ruled=%d extended=%d", oldAnch, newAnch, extAnch)
	t.Logf("in-section parser-accepted lines: old=%d ruled=%d extended=%d", oldLive, newLive, extLive)
	t.Logf("files whose anchor or read count changed old->ruled = %d", changedCount)
	for _, s := range changed {
		t.Logf("  changed: %s", s)
	}
	t.Logf("old h2/file-only anchors = %d", len(fileOnly))
	for _, s := range fileOnly {
		t.Logf("  file-only: %s", s)
	}
	t.Logf("ruled anchors at ### or deeper = %d", len(deepNew))
	for _, s := range deepNew {
		t.Logf("  deep: %s", s)
	}
	t.Logf("later section-naming headings with parser-accepted lines never read (ruled): files=%d lines=%d", laterLostFiles, laterLostLive)
	for _, s := range laterLost {
		t.Logf("  later-lost: %s", s)
	}
	t.Logf("headings excluded by the negative markers that carry vocabulary = %d", len(negExcluded))
	for _, s := range negExcluded {
		t.Logf("  negative: %s", s)
	}
	t.Logf("COUNTERFACTUAL skip-empty selection (ruled vocabulary): lines=%d, files changed vs ruled = %d", skipLive, len(skipChanged))
	for _, s := range skipChanged {
		t.Logf("  skip-empty: %s", s)
	}
	t.Logf("COUNTERFACTUAL union of every section (ruled vocabulary): lines=%d, files changed vs ruled = %d", unionLive, len(unionChanged))
	for _, s := range unionChanged {
		t.Logf("  union: %s", s)
	}
	sort.Strings(reqGained)
	sort.Strings(reqLost)
	t.Logf("CONTROL distinct REQ ids mapped by read lines, summed per file: old=%d ruled=%d", oldReqTotal, newReqTotal)
	t.Logf("REQ ids mapped by read lines: newly mapped (ruled vs old) = %d, no longer mapped = %d", len(reqGained), len(reqLost))
	for _, s := range reqGained {
		t.Logf("  req-gained: %s", s)
	}
	for _, s := range reqLost {
		t.Logf("  req-lost: %s", s)
	}
	t.Logf("COUNTERFACTUAL skip-empty + extended vocabulary: lines=%d, files changed vs ruled = %d", comboLive, len(comboChanged))
	for _, s := range comboChanged {
		t.Logf("  combo: %s", s)
	}
	t.Logf("files reading FEWER lines than the unfixed parser: ruled=%d skip-empty+extended=%d", len(ruledLoss), len(comboLoss))
	for _, s := range ruledLoss {
		t.Logf("  loss-ruled: %s", s)
	}
	for _, s := range comboLoss {
		t.Logf("  loss-combo: %s", s)
	}
	t.Logf("COUNTERFACTUAL extended vocabulary: files changed vs ruled = %d, parser-accepted line delta = %+d", extChangedCount, extChangedLiveDelta)
	for _, s := range extChanged {
		t.Logf("  extended: %s", s)
	}
}
