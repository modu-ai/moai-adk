package hook

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// M1 of SPEC-ASTGREP-LANG16-001: this is the single new file added under the
// otherwise read-only internal/hook tree. It mechanically pins the differential
// corpus gate's real semantics instead of leaving them as prose:
//
//   TestAstgrepCorpusFixturesPinned — the twelve fixtures under
//     internal/hook/security/testdata/scan-corpus/ are byte-identical to
//     their blobs at base 294b4b6ab (the plan-measured tree; digests verified
//     equal on this branch, whose fork point is develop d29b8942e), so an
//     edited fixture cannot turn a red
//     differential row green invisibly.
//
//   TestAstgrepCorpusTablePinned — the recorded wantDeny values in
//     pre_tool_scan_differential_test.go are identical to the base table, and
//     the corpus-validity gate really calls t.Skip (positionally, between the
//     observation loop and the assertion loop).
//
//   TestAstgrepCorpusRunDoesNotSkip — a live, filtered re-run of the
//     differential test reports --- PASS and neither --- SKIP nor
//     "corpus rejected:", proving the twelve assertions actually executed.

// corpusBaseSHA notes the tree every constant below was derived from.
const corpusBaseSHA = "294b4b6ab"

// corpusFixtureDigests maps each scan-corpus fixture to the SHA-256 of its
// blob at base 294b4b6ab (git cat-file blob 294b4b6ab:<path> | sha256sum),
// cross-checked against the working tree at authoring time.
var corpusFixtureDigests = map[string]string{
	"go_clean.go":           "58a994a9e6e3197867d634abe35aeca17ca4b43a30353786f713a40419231a39",
	"go_deny_credential.go": "9fd3ab48c397457309f7728ebeb815f5947f0f634bae0be61168ddf41ffd0582",
	"go_deny_md5.go":        "c5472a59befa311d83c90fd3d4a612867e4ae2e9eadbeb60cfddbe8414090d99",
	"go_warning_only.go":    "8535850413f3f7b56934a5234c52f891e26621a4d3dab77ad8d8701ddc025a1d",
	"java_uncovered.java":   "76f371af6ad07488c496113f1b67d4ec0007b3b0b069fc99afa511b4ead8e6ac",
	"js_clean.js":           "aaf7cf274e800934790bcb31029964753f0387d37054151574526ef230be0d74",
	"js_deny_exec.js":       "dd780240d6ee2d6448fedc812704f819b1d7c536429ca07d39d5c154b9a970b8",
	"py_clean.py":           "ba1a531f581d2e6094e978ed6f7aca7a8d92eeb62c6e7ad73ee692f7f18bc772",
	"py_deny_os_system.py":  "3d2b44884b0cc4e38719cbb6504bedfabce9a4f6e22479df76e727a56a45fcb2",
	"rs_uncovered.rs":       "8d6b2e8d7cfd3992e4fffce1dc5388975b4a5070d329ffa35dda2c75d8d58a13",
	"ts_clean.ts":           "9a30ac96b5d5c1b67eca69e1e2cf0798817d9578c8d7d904a81a67b983b35cba",
	"ts_deny_exec.ts":       "9a73d0dd1e9c2ae6fd11834ed510db90612fe488fb1d5fbb5540e5c50007afcb",
}

// baseWantDenyRows is the twelve-row (fixture, wantDeny) table extracted at
// base 294b4b6ab. A future edit that flips any recorded verdict makes the live
// parse below diverge from this pin.
var baseWantDenyRows = map[string]bool{
	"go_deny_credential.go": true,
	"go_deny_md5.go":        true,
	"go_clean.go":           false,
	"go_warning_only.go":    false,
	"js_deny_exec.js":       true,
	"js_clean.js":           false,
	"ts_deny_exec.ts":       true,
	"ts_clean.ts":           false,
	"py_deny_os_system.py":  true,
	"py_clean.py":           false,
	"rs_uncovered.rs":       false,
	"java_uncovered.java":   false,
}

// baseCoveredLanguages is the covered-language list at base; adding a language
// here without its denying fixture turns the differential green-by-skip, which
// this divergence check exposes.
var baseCoveredLanguages = []string{"go", "javascript", "typescript", "python"}

// differentialSource reads pre_tool_scan_differential_test.go relative to this
// package directory.
func differentialSource(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("pre_tool_scan_differential_test.go"))
	if err != nil {
		t.Fatalf("read differential source: %v", err)
	}
	return string(data)
}

func sha256File(t *testing.T, path string) string {
	t.Helper()
	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = fh.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, fh); err != nil {
		t.Fatalf("hash %s: %v", path, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// indexOf returns the byte offset of needle or fails the test; ordering checks
// use these offsets rather than fragile single greps.
func indexOf(t *testing.T, src, needle, what string) int {
	t.Helper()
	i := strings.Index(src, needle)
	if i < 0 {
		t.Fatalf("%s: sentinel %q not found in differential source", what, needle)
	}
	return i
}

func TestAstgrepCorpusFixturesPinned(t *testing.T) {
	corpusDir := filepath.Join("security", "testdata", "scan-corpus")
	for name, want := range corpusFixtureDigests {
		got := sha256File(t, filepath.Join(corpusDir, name))
		if got != want {
			t.Errorf("%s changed since %s: sha256 = %s, want %s — revert the fixture or consciously re-pin both digests with evidence",
				name, corpusBaseSHA, got, want)
		}
	}
	if len(corpusFixtureDigests) != 12 {
		t.Fatalf("pin table carries %d fixtures, want 12", len(corpusFixtureDigests))
	}
}

// extractWantDeny parses the live `(file, wantDeny)` pairs out of the corpus
// table literal in the differential test source.
func extractWantDeny(t *testing.T, src string) map[string]bool {
	t.Helper()
	re := regexp.MustCompile(`file:\s*"([^"]+)"[^}]*wantDeny:\s*(true|false)`)
	rows := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(src, -1) {
		rows[m[1]] = m[2] == "true"
	}
	if len(rows) == 0 {
		t.Fatal("no wantDeny rows parsed; the extraction regex no longer matches the corpus literal")
	}
	return rows
}

func TestAstgrepCorpusTablePinned(t *testing.T) {
	src := differentialSource(t)

	live := extractWantDeny(t, src)
	for name, want := range baseWantDenyRows {
		got, ok := live[name]
		if !ok {
			t.Errorf("row for %s disappeared from the corpus table since %s", name, corpusBaseSHA)
			continue
		}
		if got != want {
			t.Errorf("%s: wantDeny flipped from %v to %v since %s — a flip is how a red differential row turns green without any rule change; re-record it through the normal audit path, not silently",
				name, want, got, corpusBaseSHA)
		}
	}
	if len(live) != len(baseWantDenyRows) {
		t.Errorf("corpus table now holds %d rows, want %d — update the pin together with the table", len(live), len(baseWantDenyRows))
	}

	langs := regexp.MustCompile(`coveredCorpusLanguages\s*=\s*\[\]string\{([^}]*)\}`).FindStringSubmatch(src)
	if langs == nil {
		t.Fatal("coveredCorpusLanguages literal not found")
	}
	var found []string
	for _, part := range strings.Split(langs[1], ",") {
		s := strings.Trim(strings.TrimSpace(part), `"`)
		if s != "" {
			found = append(found, s)
		}
	}
	if strings.Join(found, "|") != strings.Join(baseCoveredLanguages, "|") {
		t.Errorf("covered languages changed since %s: got %v, pinned %v — new entries must arrive with their own denying fixture in the same change",
			corpusBaseSHA, found, baseCoveredLanguages)
	}

	// Positional proof of the validity gate: observation loop -> missing-guard
	// -> t.Skip -> assertion loop. If a refactor moved the skip after the
	// assertions (or deleted it), the guard's silence could no longer be told
	// apart from a meaningful pass.
	firstObs := indexOf(t, src, "for _, fx := range scanCorpus {", "observation loop")
	guard := indexOf(t, src, "if len(missing) > 0 {", "validity guard")
	skipCall := indexOf(t, src, "t.Skip(b.String())", "validity skip call")
	assertLoop := indexOf(t, src, "gotDeny := o.decision == DecisionDeny", "assertion loop")
	for i := range scanCorpus {
		_ = i // touch package symbol to keep the reference compile-time honest
	}
	if firstObs > guard || guard > skipCall || skipCall > assertLoop {
		t.Errorf("validity gate is mispositioned: obs(%d) < guard(%d) < skip(%d) < assert(%d) must hold — spec.md A.7 relies on this exact shape",
			firstObs, guard, skipCall, assertLoop)
	}
	if !strings.Contains(src, "no denying fixture for covered language(s)") {
		t.Error("the validity skip no longer explains itself; keep the self-describing message")
	}
}

func TestAstgrepCorpusRunDoesNotSkip(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("go toolchain not on PATH; cannot re-run the differential test (%v)", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, goBin, "test", "-count=1", "-run", "^TestScanWriteContentDifferential$", "-v", "./internal/hook/")
	cmd.Dir = repoRootForTest(t)
	out, err := cmd.CombinedOutput()
	output := string(out)
	t.Logf("nested differential run rc=%v, %d bytes of output", err, len(output))

	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("nested differential re-run timed out; output tail:\n%s", tailBytes(output, 2000))
	}
	if !strings.Contains(output, "--- PASS: TestScanWriteContentDifferential") {
		t.Errorf("differential re-run did not report --- PASS; the validity gate may have silenced all assertions.\nrc=%v\n%s", err, tailBytes(output, 4000))
	}
	if strings.Contains(output, "--- SKIP") {
		t.Errorf("differential re-run SKIPPED; green-by-skip proves nothing.\n%s", tailBytes(output, 4000))
	}
	if strings.Contains(output, "corpus rejected:") {
		t.Errorf("differential re-run reported a rejected corpus: %s", tailBytes(output, 4000))
	}
	passes := strings.Count(output, "-- PASS: TestScanWriteContentDifferential")
	if passes == 0 {
		t.Errorf("counted zero explicit passing runs; output was:\n%s", tailBytes(output, 4000))
	}
}

func tailBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return fmt.Sprintf("...(truncated)...%s", s[len(s)-n:])
}
