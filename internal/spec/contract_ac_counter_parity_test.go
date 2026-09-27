// contract_ac_counter_parity_test.go — parity between the published awk AC
// counter (the MOAI-AC-COUNTER block of the manager-docs agent definition) and
// its Go port in internal/contract (SPEC-AUTONOMY-CONTRACT-001, REQ-CONTRACT-005).
//
// It lives here, beside extractCounterCommand, so the counter is EXTRACTED
// from the agent definition rather than restated. internal/contract never
// imports internal/spec, so the import below creates no cycle.
package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// awkCountResult is the awk counter's answer in the Go port's vocabulary.
type awkCountResult struct {
	ambiguous []string // non-empty exactly when awk exited 3
	live      int
	excluded  int
}

var awkTallyRe = regexp.MustCompile(`live=(\d+) excluded=(\d+) ambiguous=0`)

// writeTempAcceptance writes content to a fresh file under dir and returns its
// absolute path.
func writeTempAcceptance(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// runAwkCounter runs the extracted counter on content (written verbatim to a
// temp file) and decodes its three-shape output contract. ok is false, with a
// description, when the output fits none of the shapes.
func runAwkCounter(t *testing.T, counter, dir, name string, content []byte) (res awkCountResult, ok bool, why string) {
	t.Helper()
	path := writeTempAcceptance(t, dir, name, content)
	stdout, stderr, code := runCounter(t, counter, path)
	switch code {
	case 0:
		fields := strings.Fields(stdout)
		if len(fields) != 1 {
			return res, false, fmt.Sprintf("exit 0 but stdout %q is not one integer", stdout)
		}
		live, err := strconv.Atoi(fields[0])
		if err != nil {
			return res, false, fmt.Sprintf("exit 0 but stdout %q is not an integer", stdout)
		}
		m := awkTallyRe.FindStringSubmatch(stderr)
		if m == nil {
			return res, false, fmt.Sprintf("exit 0 but stderr tally absent: %q", stderr)
		}
		exc, _ := strconv.Atoi(m[2])
		return awkCountResult{live: live, excluded: exc}, true, ""
	case 3:
		first := strings.SplitN(stdout, "\n", 2)[0]
		if !strings.HasPrefix(first, "AMBIGUOUS") {
			return res, false, fmt.Sprintf("exit 3 but stdout %q lacks AMBIGUOUS", stdout)
		}
		ids := strings.Fields(strings.TrimPrefix(first, "AMBIGUOUS"))
		if len(ids) == 0 {
			return res, false, fmt.Sprintf("exit 3 but no identifiers named: %q", stdout)
		}
		return awkCountResult{ambiguous: ids}, true, ""
	default:
		return res, false, fmt.Sprintf("unexpected exit %d (stdout=%q stderr=%q)", code, stdout, stderr)
	}
}

// compareCounters runs both counters on the same normalized bytes and returns
// "" when they agree, or a description of the disagreement.
func compareCounters(t *testing.T, counter, dir, name string, normalized []byte) (awkCountResult, contract.ACCountResult, string) {
	t.Helper()
	awk, ok, why := runAwkCounter(t, counter, dir, name, normalized)
	if !ok {
		return awk, contract.ACCountResult{}, "awk counter: " + why
	}
	goRes, err := contract.CountAC(normalized)
	if err != nil {
		return awk, goRes, fmt.Sprintf("Go counter error: %v", err)
	}
	awkAmb := len(awk.ambiguous) > 0
	if awkAmb != goRes.IsAmbiguous() {
		return awk, goRes, fmt.Sprintf("ambiguity verdict differs: awk=%v (%v) go=%v (%v)",
			awkAmb, awk.ambiguous, goRes.IsAmbiguous(), goRes.Ambiguous)
	}
	if awkAmb {
		if !slices.Equal(awk.ambiguous, goRes.Ambiguous) {
			return awk, goRes, fmt.Sprintf("ambiguous identifiers differ: awk=%v go=%v", awk.ambiguous, goRes.Ambiguous)
		}
		return awk, goRes, ""
	}
	if awk.live != goRes.Live || awk.excluded != goRes.Excluded {
		return awk, goRes, fmt.Sprintf("counts differ: awk live=%d excluded=%d, go live=%d excluded=%d",
			awk.live, awk.excluded, goRes.Live, goRes.Excluded)
	}
	return awk, goRes, ""
}

// Synthetic per-branch fixtures (AC-CONTRACT-006 input (b)).
const (
	parityPrefixFixture = "<!-- moai-ac-prefix: CR -->\n" +
		"# Fixture — prefix declaration\n" +
		"### CR-01 — live under the declared prefix\n" +
		"### CR-02 — live under the declared prefix\n" +
		"### CR-03 [RETIRED] — retired under the declared prefix\n" +
		"- AC-EXT-01 [REF] cites another namespace\n"
	parityRetiredFixture = "# Fixture — retired marker\n" +
		"### AC-RET-001 — live\n" +
		"### AC-RET-002 [RETIRED] — retired\n" +
		"| AC-RET-003 | live row |\n"
	parityRefFixture = "# Fixture — ref marker\n" +
		"### AC-REF-001 — live\n" +
		"- see AC-OTHER-009 [REF] for the upstream criterion\n"
	parityAmbiguityFixture = "# Fixture — ambiguity\n" +
		"### AC-AMB-001 — live here\n" +
		"- AC-AMB-001 [RETIRED] but retired there\n" +
		"### AC-AMB-002 — unambiguous\n"
	parityCRLFBody = "# Fixture — CRLF line endings\n" +
		"### AC-CRLF-001 — live\n" +
		"### AC-CRLF-002 [RETIRED]\n" +
		"| AC-CRLF-003 | row |\n"
	parityBOMBody = "<!-- moai-ac-prefix: BOM -->\n" +
		"# Fixture — UTF-8 BOM\n" +
		"### BOM-01 — live\n" +
		"### BOM-02 [REF]\n"
)

// TestAC_CONTRACT_006 — AC counter parity, corpus and per-branch fixtures
// (REQ-CONTRACT-005).
func TestAC_CONTRACT_006(t *testing.T) {
	root := repoRoot(t)
	counter := extractCounterCommand(t, filepath.Join(root, acLocalClausePath))
	dir := t.TempDir()

	t.Run("corpus", func(t *testing.T) {
		matches, err := filepath.Glob(filepath.Join(root, ".moai/specs/*/acceptance.md"))
		if err != nil {
			t.Fatalf("glob corpus: %v", err)
		}
		sort.Strings(matches)
		n, nonZero := 0, 0
		for i, abs := range matches {
			rel, err := filepath.Rel(root, abs)
			if err != nil {
				t.Fatalf("relpath: %v", err)
			}
			rel = filepath.ToSlash(rel)
			if strings.HasPrefix(rel, ".moai/specs/_archive/") {
				continue
			}
			raw, err := os.ReadFile(abs)
			if err != nil {
				t.Fatalf("read %s: %v", rel, err)
			}
			n++
			normalized := contract.NormalizeAcceptance(raw)
			awk, _, diff := compareCounters(t, counter, dir, fmt.Sprintf("corpus-%04d.md", i), normalized)
			if diff != "" {
				t.Errorf("corpus input %s: %s", rel, diff)
			}
			if awk.live > 0 || len(awk.ambiguous) > 0 {
				nonZero++
			}
		}
		if n == 0 {
			t.Fatalf("corpus glob matched no acceptance.md — the parity check would be vacuous")
		}
		// Non-vacuity: two counters that both see empty input agree trivially.
		// Most real acceptance files declare criteria, so a corpus where
		// fewer than half count anything means the compared bytes are wrong.
		if nonZero*2 < n {
			t.Fatalf("only %d of %d corpus files produced a non-zero count — the compared bytes look empty", nonZero, n)
		}
		t.Logf("corpus-size=%d non-zero=%d", n, nonZero)
	})

	// wantLive / wantAmbiguous are hand-derived from the fixture text and
	// checked against the awk counter, so a normalizer that dropped content
	// could not make the two implementations agree vacuously.
	fixtures := []struct {
		name          string
		raw           []byte
		wantLive      int
		wantAmbiguous []string
	}{
		{"prefix-declaration", []byte(parityPrefixFixture), 2, nil},
		{"retired-marker", []byte(parityRetiredFixture), 2, nil},
		{"ref-marker", []byte(parityRefFixture), 1, nil},
		{"ambiguity", []byte(parityAmbiguityFixture), 0, []string{"AC-AMB-001"}},
		{"crlf", []byte(strings.ReplaceAll(parityCRLFBody, "\n", "\r\n")), 2, nil},
		{"utf8-bom", append([]byte("\xEF\xBB\xBF"), parityBOMBody...), 1, nil},
	}
	type outcome struct {
		awk awkCountResult
		go_ contract.ACCountResult
	}
	results := map[string]outcome{}
	t.Run("fixtures", func(t *testing.T) {
		for _, f := range fixtures {
			normalized := contract.NormalizeAcceptance(f.raw)
			awk, goRes, diff := compareCounters(t, counter, dir, "fixture-"+f.name+".md", normalized)
			if diff != "" {
				t.Errorf("fixture input %s: %s", f.name, diff)
				continue
			}
			if len(f.wantAmbiguous) > 0 {
				if !slices.Equal(awk.ambiguous, f.wantAmbiguous) {
					t.Errorf("fixture input %s: awk ambiguous = %v, hand-derived %v", f.name, awk.ambiguous, f.wantAmbiguous)
				}
			} else if len(awk.ambiguous) > 0 || awk.live != f.wantLive {
				t.Errorf("fixture input %s: awk live=%d ambiguous=%v, hand-derived live=%d", f.name, awk.live, awk.ambiguous, f.wantLive)
			}
			results[f.name] = outcome{awk, goRes}
		}
	})

	t.Run("positive control: ambiguity fixture ambiguous in both", func(t *testing.T) {
		o, ok := results["ambiguity"]
		if !ok {
			t.Fatal("ambiguity fixture produced no agreed result")
		}
		if len(o.awk.ambiguous) == 0 {
			t.Fatal("awk counter did not report the ambiguity fixture ambiguous")
		}
		if !o.go_.IsAmbiguous() {
			t.Fatal("Go counter did not report the ambiguity fixture ambiguous")
		}
	})

	t.Run("positive control: prefix declaration changes the count", func(t *testing.T) {
		o, ok := results["prefix-declaration"]
		if !ok {
			t.Fatal("prefix fixture produced no agreed result")
		}
		lines := strings.SplitN(parityPrefixFixture, "\n", 2)
		if !strings.HasPrefix(lines[0], "<!-- moai-ac-prefix:") {
			t.Fatalf("prefix fixture's first line is not the declaration: %q", lines[0])
		}
		undeclared := contract.NormalizeAcceptance([]byte(lines[1]))
		awkDefault, goDefault, diff := compareCounters(t, counter, dir, "fixture-prefix-default.md", undeclared)
		if diff != "" {
			t.Fatalf("prefix fixture under the default prefix: %s", diff)
		}
		if goDefault.Prefix != "AC" {
			t.Fatalf("default-prefix count used prefix %q", goDefault.Prefix)
		}
		if o.go_.Live == goDefault.Live || o.awk.live == awkDefault.live {
			t.Fatalf("prefix declaration did not change the count: declared awk=%d go=%d, default awk=%d go=%d",
				o.awk.live, o.go_.Live, awkDefault.live, goDefault.Live)
		}
	})

	t.Run("raw CRLF measurement", func(t *testing.T) {
		raw := []byte(strings.ReplaceAll(parityCRLFBody, "\n", "\r\n"))
		path := writeTempAcceptance(t, dir, "fixture-crlf-raw.md", raw)
		stdout, _, code := runCounter(t, counter, path)
		o, ok := results["crlf"]
		if !ok {
			t.Fatal("CRLF fixture produced no agreed result")
		}
		t.Logf("raw-crlf-count=%s (exit %d) normalized-count=%d", strings.TrimSpace(stdout), code, o.awk.live)
	})
}

// TestContractSpecIDPatternMatchesLint pins internal/contract's SPEC ID
// pattern to the spec-lint frontmatter pattern (design.md § Package Layout).
func TestContractSpecIDPatternMatchesLint(t *testing.T) {
	if got, want := contract.SpecIDPattern, specIDPattern.String(); got != want {
		t.Fatalf("contract.SpecIDPattern = %q, lint specIDPattern = %q", got, want)
	}
	contractRe := regexp.MustCompile(contract.SpecIDPattern)
	for _, id := range []string{
		"SPEC-AUTONOMY-CONTRACT-001", "SPEC-A-001", "SPEC-X1-Y2-123",
		"SPEC-001", "spec-A-001", "SPEC-A-01", "SPEC-A-0001", "SPEC-1A-001", "SPEC-A-001-x",
	} {
		if a, b := contractRe.MatchString(id), specIDPattern.MatchString(id); a != b {
			t.Errorf("%s: contract=%v lint=%v", id, a, b)
		}
		if a, b := contract.ValidSpecID(id), specIDPattern.MatchString(id); a != b {
			t.Errorf("%s: contract.ValidSpecID=%v lint=%v", id, a, b)
		}
	}
}
