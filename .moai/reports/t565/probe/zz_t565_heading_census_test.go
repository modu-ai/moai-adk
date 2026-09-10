package spec

// zz_t565_heading_census_test.go — card t565 corpus census of the AC section
// heading anchor. The committed copy lives under .moai/reports/t565/probe/; it
// is copied into internal/spec only for the run and removed afterwards. It
// writes nothing; every figure is a t.Logf line.
//
// RE-DERIVATION
//
//	cp .moai/reports/t565/probe/zz_t565_heading_census_test.go internal/spec/
//	go test ./internal/spec -count=1 -run '^TestT565HeadingAnchorCensus$' -v
//	rm internal/spec/zz_t565_heading_census_test.go
//
// WHAT IT MEASURES. findACSectionStart anchors on the FIRST heading that starts
// with "##" (so ### and deeper too) and contains "acceptance" after lower-casing
// (so a heading that only names the file acceptance.md too). extractACLines then
// reads until the next heading starting with "##". The census splits the anchor
// by kind, counts declaration-shaped lines inside and outside the anchored
// section, and for misanchored files looks for a later heading that names the
// acceptance criteria and was therefore never read.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// t565DeclRe is discriminator B from t528 (bullet required, dots allowed, id may
// end in a letter): the declaration-SHAPED line, independent of the parser.
var t565DeclRe = regexp.MustCompile(`^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`)

var t565ACToken = regexp.MustCompile(`(^|[^A-Za-z])AC([^A-Za-z-]|$)`)

// t565AnchorKind classifies the heading findACSectionStart anchored on.
func t565AnchorKind(h string) string {
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

// t565HeadingBucket classifies the heading a declaration line sits under.
func t565HeadingBucket(h string) string {
	lower := strings.ToLower(h)
	noFile := strings.ReplaceAll(lower, "acceptance.md", "")
	switch {
	case h == "":
		return "no-heading"
	case strings.Contains(noFile, "acceptance"):
		return "english-acceptance-word"
	case strings.Contains(h, "수락") || strings.Contains(h, "인수") || strings.Contains(h, "검수"):
		return "korean-acceptance-word"
	case strings.Contains(lower, "acceptance.md"):
		return "file-mention-only"
	case t565ACToken.MatchString(h):
		return "ac-token"
	default:
		return "other"
	}
}

func TestT565HeadingAnchorCensus(t *testing.T) {
	root := "../../.moai/specs"
	var files, anchored, unanchored int
	kinds := map[string]int{}
	inDecl := map[string]int{}
	inLive := map[string]int{}
	var outDeclAnchored, outLiveAnchored, outDeclUnanchored, outLiveUnanchored int
	outBucketAnchored := map[string]int{}
	outBucketUnanchored := map[string]int{}
	var laterRealFiles, laterRealLive int
	var samplesMis, samplesLater []string
	samplesOutside := map[string][]string{}

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		files++
		lines := strings.Split(string(b), "\n")
		start := findACSectionStart(lines)
		end := start
		kind := "none"
		if start >= 0 {
			anchored++
			kind = t565AnchorKind(strings.TrimSpace(lines[start-1]))
			for end = start; end < len(lines); end++ {
				if strings.HasPrefix(strings.TrimSpace(lines[end]), "##") {
					break
				}
			}
		} else {
			unanchored++
		}
		kinds[kind]++

		misanchored := strings.HasPrefix(kind, "h3+") || strings.HasSuffix(kind, "file-only")
		if misanchored && len(samplesMis) < 15 {
			samplesMis = append(samplesMis, kind+"  "+p+"  ->  "+strings.TrimSpace(lines[start-1]))
		}

		lastHeading := ""
		for i, line := range lines {
			tr := strings.TrimSpace(line)
			if strings.HasPrefix(tr, "#") {
				lastHeading = tr
			}
			if !t565DeclRe.MatchString(line) {
				continue
			}
			live := parseSingleACLine(tr) != nil
			if start >= 0 && i >= start && i < end {
				inDecl[kind]++
				if live {
					inLive[kind]++
				}
				continue
			}
			if start >= 0 {
				outDeclAnchored++
				if live {
					outLiveAnchored++
					bucket := t565HeadingBucket(lastHeading)
					outBucketAnchored[bucket]++
					if len(samplesOutside[bucket]) < 8 {
						samplesOutside[bucket] = append(samplesOutside[bucket],
							p+":"+t565Itoa(i+1)+"  under: "+lastHeading+"  |  anchor: "+strings.TrimSpace(lines[start-1]))
					}
				}
			} else {
				outDeclUnanchored++
				if live {
					outLiveUnanchored++
					outBucketUnanchored[t565HeadingBucket(lastHeading)]++
				}
			}
		}

		// A misanchored file whose REAL acceptance heading comes later is never read.
		if misanchored {
			for j := start; j < len(lines); j++ {
				tr := strings.TrimSpace(lines[j])
				if !strings.HasPrefix(tr, "##") {
					continue
				}
				if strings.Contains(strings.ReplaceAll(strings.ToLower(tr), "acceptance.md", ""), "acceptance") {
					laterRealFiles++
					n := 0
					for k := j + 1; k < len(lines); k++ {
						t2 := strings.TrimSpace(lines[k])
						if strings.HasPrefix(t2, "##") {
							break
						}
						if t565DeclRe.MatchString(lines[k]) && parseSingleACLine(t2) != nil {
							n++
						}
					}
					laterRealLive += n
					if len(samplesLater) < 15 {
						samplesLater = append(samplesLater, p+"  anchor: "+strings.TrimSpace(lines[start-1])+"  |  later: "+tr+"  |  live decls under later = "+t565Itoa(n))
					}
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

	t.Logf("spec.md walked = %d  anchored = %d  unanchored = %d", files, anchored, unanchored)
	for _, k := range t565SortedKeys(kinds) {
		t.Logf("  anchor kind %-16s files=%-4d in-section decl-shaped=%-5d in-section parser-accepted=%d", k, kinds[k], inDecl[k], inLive[k])
	}
	t.Logf("OUTSIDE the anchored section (anchored files): decl-shaped=%d parser-accepted=%d", outDeclAnchored, outLiveAnchored)
	for _, k := range t565SortedKeys(outBucketAnchored) {
		t.Logf("    under heading bucket %-26s parser-accepted=%d", k, outBucketAnchored[k])
	}
	t.Logf("IN UNANCHORED files: decl-shaped=%d parser-accepted=%d", outDeclUnanchored, outLiveUnanchored)
	for _, k := range t565SortedKeys(outBucketUnanchored) {
		t.Logf("    under heading bucket %-26s parser-accepted=%d", k, outBucketUnanchored[k])
	}
	t.Logf("MISANCHORED files with a later real acceptance heading = %d, parser-accepted decls under it (never read) = %d", laterRealFiles, laterRealLive)
	for _, s := range samplesMis {
		t.Logf("  misanchored sample: %s", s)
	}
	for _, s := range samplesLater {
		t.Logf("  later-real sample: %s", s)
	}
	for _, k := range t565SortedKeys(outBucketAnchored) {
		for _, s := range samplesOutside[k] {
			t.Logf("  outside sample [%s]: %s", k, s)
		}
	}
}

func t565SortedKeys(m map[string]int) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func t565Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
