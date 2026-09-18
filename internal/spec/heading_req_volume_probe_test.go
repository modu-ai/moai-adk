//go:build heading_req_probe

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

// Measurement instrument for card t894. It answers the question the card makes a
// precondition of design: if the REQ collector also picked up `### REQ-…`
// headings, how many findings would newly be REPORTED, and under which codes?
//
// It is a measurement, not an implementation: nothing here is wired into the
// live collector, and it runs only under `-tags heading_req_probe`. Run:
//
//	go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume -v
//
// Read the corpus SHA off the tree before citing any number it prints; the
// corpus moves.

// headingREQPattern mirrors reqLineWidePattern with the list bullet replaced by
// a level-3 heading. Everything else — the ID shape and the `—`/`:` separator —
// is deliberately identical, so the measurement reflects what an
// otherwise-unchanged collector would see.
var headingREQPattern = regexp.MustCompile(`^###\s+\**\s*(REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+)\s*\**\s*(?:\([^)]*\)\s*\**\s*)?(?:—|:)\s*(.*)$`)

// parseREQsHeading collects heading-form entries. useBody selects WHICH text
// becomes REQEntry.Text, which is the design question this probe exists to
// answer:
//
//	false — the heading's own trailing text. That text is a section TITLE
//	        ("Event-driven (When) — advisor rung trigger"), not a requirement
//	        sentence.
//	true  — the first non-empty paragraph BELOW the heading, which is where the
//	        SHALL-bearing requirement statement actually lives.
//
// Modality judgment is defined over requirement statements, so the two choices
// are not cosmetic: one of them feeds the judge a title.
func parseREQsHeading(body string, useBody bool) []REQEntry {
	var reqs []REQEntry
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		m := headingREQPattern.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		text := strings.TrimSpace(m[2])
		if useBody {
			if para := firstParagraphAfter(lines, i); para != "" {
				text = para
			}
		}
		reqs = append(reqs, REQEntry{ID: m[1], Text: text, Line: i + 1})
	}
	return reqs
}

// firstParagraphAfter returns the first non-empty, non-heading line following
// idx, stopping at the next heading so a REQ never borrows its successor's prose.
func firstParagraphAfter(lines []string, idx int) string {
	for j := idx + 1; j < len(lines); j++ {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "#") {
			return ""
		}
		return t
	}
	return ""
}

// locateSpecsRoot walks up from the test's working directory until it finds the
// tree's own .moai/specs. Ascending by directory name rather than by a fixed
// relative literal keeps the probe correct wherever the package sits, and keeps
// the measured corpus the one belonging to THIS worktree — which matters here,
// because the number this probe prints is only meaningful against a named tree.
func locateSpecsRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".moai", "specs")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .moai/specs found above %q", dir)
		}
		dir = parent
	}
}

type volume struct {
	files              int
	filesWithNewREQ    int
	newREQs            int
	coverageIncomplete int
	modalityMalformed  int
	modalityUnjudged   int
	legacyEARSKeyword  int
	invalidREQID       int
	duplicateREQID     int
}

func TestHeadingREQVolume(t *testing.T) {
	for _, tc := range []struct {
		name    string
		useBody bool
	}{
		{"A_heading_title_as_text", false},
		{"B_body_paragraph_as_text", true},
	} {
		t.Run(tc.name, func(t *testing.T) { measureVolume(t, tc.useBody) })
	}
}

func measureVolume(t *testing.T, useBodyText bool) {
	abs, err := locateSpecsRoot()
	if err != nil {
		t.Fatalf("locate specs root: %v", err)
	}

	var v volume
	perCode := map[string]int{}
	topFiles := map[string]int{}

	walkErr := filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		doc := parseSPECDoc(path)
		if doc.ParseError != nil || doc.Body == "" {
			return nil
		}
		v.files++

		// Entries the live collector already has, keyed by ID.
		existing := make(map[string]bool, len(doc.REQs))
		for _, r := range doc.REQs {
			existing[r.ID] = true
		}

		heading := parseREQsHeading(doc.Body, useBodyText)
		var fresh []REQEntry
		for _, r := range heading {
			if !existing[r.ID] {
				fresh = append(fresh, r)
			}
		}
		if len(fresh) == 0 {
			return nil
		}
		v.filesWithNewREQ++
		v.newREQs += len(fresh)

		covered := collectAllREQIDs(doc.Criteria)
		for id := range siblingAcceptanceCoveredREQIDs(doc.Path) {
			covered[id] = true
		}

		fileFindings := 0
		seen := make(map[string]bool, len(fresh))
		for _, r := range fresh {
			if !reqIDPattern.MatchString(r.ID) {
				v.invalidREQID++
				perCode["InvalidREQID"]++
				fileFindings++
				continue
			}
			if existing[r.ID] || seen[r.ID] {
				v.duplicateREQID++
				perCode["DuplicateREQID"]++
				fileFindings++
			}
			seen[r.ID] = true

			switch judgeModality(r.Text) {
			case modalityJudgedMalformed:
				v.modalityMalformed++
				perCode["ModalityMalformed"]++
				fileFindings++
			case modalityUnjudged:
				v.modalityUnjudged++
				perCode["ModalityUnjudged"]++
				fileFindings++
			}

			// Same loop as the modality switch above emits this one; counting it
			// here keeps the probe's denominator equal to the rules' own.
			if isLegacyEARSPattern(r.Text) {
				v.legacyEARSKeyword++
				perCode["LegacyEARSKeyword"]++
				fileFindings++
			}

			if !covered[r.ID] {
				v.coverageIncomplete++
				perCode["CoverageIncomplete"]++
				fileFindings++
			}
		}
		if fileFindings > 0 {
			topFiles[strings.TrimPrefix(path, abs+"/")] = fileFindings
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk: %v", walkErr)
	}

	// Emptiness guard: a corpus that yielded no documents would make every
	// number below a vacuous zero.
	if v.files == 0 {
		t.Fatalf("probe swept zero documents under %s — the numbers below would be vacuous", abs)
	}
	if v.newREQs == 0 {
		t.Fatalf("probe found zero heading-form REQs the live collector misses — either the corpus changed or the pattern is wrong")
	}

	total := 0
	for _, n := range perCode {
		total += n
	}

	t.Logf("corpus root: %s", abs)
	t.Logf("documents swept: %d", v.files)
	t.Logf("files gaining at least one REQ entry: %d", v.filesWithNewREQ)
	t.Logf("newly collected REQ entries: %d", v.newREQs)
	t.Logf("TOTAL new findings: %d", total)

	codes := make([]string, 0, len(perCode))
	for c := range perCode {
		codes = append(codes, c)
	}
	sort.Slice(codes, func(i, j int) bool { return perCode[codes[i]] > perCode[codes[j]] })
	for _, c := range codes {
		t.Logf("  %-20s %6d", c, perCode[c])
	}

	// The concentration question: is the volume spread thin or piled on a few
	// documents? A long tail and a short head imply different rollout designs.
	type fc struct {
		file string
		n    int
	}
	list := make([]fc, 0, len(topFiles))
	for f, n := range topFiles {
		list = append(list, fc{f, n})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	t.Logf("files carrying findings: %d", len(list))
	head := 0
	for i, e := range list {
		if i < 15 {
			t.Logf("  %4d  %s", e.n, e.file)
		}
		if i < 15 {
			head += e.n
		}
	}
	t.Logf("top-15 files carry %d of %d findings (%.0f%%)", head, total, float64(head)*100/float64(total))

	// Per-domain spread, for the staged-rollout option the card asks about.
	perDomain := map[string]int{}
	for f, n := range topFiles {
		parts := strings.SplitN(f, string(os.PathSeparator), 2)
		perDomain[parts[0]] += n
	}
	t.Logf("distinct SPEC directories carrying findings: %d", len(perDomain))

	fmt.Fprintln(os.Stderr, "probe complete")
}
