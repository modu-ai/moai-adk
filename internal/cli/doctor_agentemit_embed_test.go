package cli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// --- fixtures -------------------------------------------------------------

// newEmbedFixtureRoot builds a fake project root carrying a .moai/ marker and
// the committed emission set (one .toml per name).
func newEmbedFixtureRoot(t *testing.T, names ...string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	dir := filepath.Join(root, filepath.FromSlash(committedEmissionRelDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir committed set: %v", err)
	}
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("name = \""+n+"\"\n"), 0o644); err != nil {
			t.Fatalf("write committed %s: %v", n, err)
		}
	}
	return root
}

// newExtractedDir builds a fake extraction output carrying the deployed
// emission layout. Contents map name -> file body.
func newExtractedDir(t *testing.T, contents map[string]string) string {
	t.Helper()
	base := t.TempDir()
	dir := filepath.Join(base, filepath.FromSlash(deployedEmissionRelDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir extracted set: %v", err)
	}
	for n, body := range contents {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(body), 0o644); err != nil {
			t.Fatalf("write extracted %s: %v", n, err)
		}
	}
	return base
}

// writeFakeBinary creates an executable stand-in at root/bin/moai so the
// judgment-target existence probe succeeds without a real build.
func writeFakeBinary(t *testing.T, root string) string {
	t.Helper()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	p := filepath.Join(binDir, "moai")
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	return p
}

func staticExtractor(dir string) emissionExtractor {
	return func(string) (string, func(), error) { return dir, func() {}, nil }
}

// --- findEmbedCheckRoot ---------------------------------------------------

// TestFindEmbedCheckRoot_WalksUpFromSubdirectory pins the applicability
// anchor: the check resolves the tree under check ITSELF rather than trusting
// the raw os.Getwd() value doctor hands every check (doctor.go passes cwd
// verbatim), so a run from a subdirectory reaches the same root.
func TestFindEmbedCheckRoot_WalksUpFromSubdirectory(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	sub := filepath.Join(root, "internal", "cli")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	got, ok := findEmbedCheckRoot(sub)
	if !ok {
		t.Fatalf("findEmbedCheckRoot(%q) = not found, want the committed-set-bearing ancestor", sub)
	}
	if !embedSameDir(t, got, root) {
		t.Errorf("findEmbedCheckRoot(%q) = %q, want %q", sub, got, root)
	}
}

// TestFindEmbedCheckRoot_NoMarker reports not-found when no ancestor carries
// a committed emission set. t.TempDir() sits under the OS temp root, which
// carries neither the committed set nor a .moai marker.
func TestFindEmbedCheckRoot_NoMarker(t *testing.T) {
	dir := t.TempDir()
	if _, ok := findEmbedCheckRoot(dir); ok {
		t.Errorf("findEmbedCheckRoot(%q) = found, want not found", dir)
	}
}

func embedSameDir(t *testing.T, a, b string) bool {
	t.Helper()
	ra, err := filepath.EvalSymlinks(a)
	if err != nil {
		ra = a
	}
	rb, err := filepath.EvalSymlinks(b)
	if err != nil {
		rb = b
	}
	return ra == rb
}

// --- applicability --------------------------------------------------------

// TestAgentEmitEmbed_NotApplicable_NoRoot: outside any MoAI project the check
// reports ok and never touches a binary (REQ-AEL-004 applicability predicate).
func TestAgentEmitEmbed_NotApplicable_NoRoot(t *testing.T) {
	called := false
	extract := func(string) (string, func(), error) {
		called = true
		return "", func() {}, nil
	}

	c := checkAgentEmitEmbedAgainst(t.TempDir(), "", extract, false)

	if c.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok (not applicable)", c.Status)
	}
	if called {
		t.Error("extraction ran in a non-applicable tree")
	}
}

// TestAgentEmitEmbed_NotApplicable_NoCommittedSet: a deployed user project
// carries a .moai/ marker but NO committed emission set. The check reports ok,
// names the absent committed set as its reason, and leaves doctor's exit
// status unchanged (the check contributes no Fail).
func TestAgentEmitEmbed_NotApplicable_NoCommittedSet(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	// The deployment output the predicate must NOT be keyed on: a project-root
	// .codex/agents/moai/*.toml set. Its presence must not make the check
	// applicable (SPEC forbids substituting it for the committed set).
	depDir := filepath.Join(root, filepath.FromSlash(deployedEmissionRelDir))
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatalf("mkdir deployed set: %v", err)
	}
	if err := os.WriteFile(filepath.Join(depDir, "manager-git.toml"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write deployed toml: %v", err)
	}

	called := false
	extract := func(string) (string, func(), error) {
		called = true
		return "", func() {}, nil
	}

	c := checkAgentEmitEmbedAgainst(root, "", extract, false)

	if c.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok (not applicable); message: %s", c.Status, c.Message)
	}
	if called {
		t.Error("extraction ran against a deployment output — the predicate must key on the committed set only")
	}
	if !strings.Contains(c.Message, "no committed") {
		t.Errorf("message = %q, want it to name the absent committed emission set", c.Message)
	}
}

// --- applicable-but-unjudgeable ------------------------------------------

// TestAgentEmitEmbed_MissingBinarySkips is the informational-skip envelope
// (REQ-CDB-001, superseding the bin-absent-failure clause of REQ-AEL-004): in
// a tree that DOES carry the committed emission set, an absent judgment
// target is a judgment the tree cannot host — reported ok, never fail — and
// the skip must end before any attempt to execute the nonexistent binary.
// REQ-CDB-002: the message names the absent path and the remedy so
// "skipped: nothing to judge" stays distinguishable from "check disabled".
func TestAgentEmitEmbed_MissingBinarySkips(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml", "manager-docs.toml")
	absent := filepath.Join(root, "bin", "does-not-exist")

	called := false
	extract := func(string) (string, func(), error) {
		called = true
		return "", func() {}, nil
	}

	c := checkAgentEmitEmbedAgainst(root, absent, extract, false)

	if c.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok (applicable tree, no judgment target = skip); message: %s", c.Status, c.Message)
	}
	if called {
		t.Error("extraction ran against a nonexistent binary — the skip must end before execution")
	}
	if !strings.Contains(c.Message, absent) {
		t.Errorf("message = %q, want it to name the absent judgment-target path %q", c.Message, absent)
	}
	if !strings.Contains(c.Message, "make build") || !strings.Contains(c.Message, embedCheckBinEnvKey) {
		t.Errorf("message = %q, want it to name the remedy (`make build` or %s=<path>)", c.Message, embedCheckBinEnvKey)
	}
}

// TestAgentEmitEmbed_ExtractionErrorFails: an extraction that errors is a
// failure, never a silent pass.
func TestAgentEmitEmbed_ExtractionErrorFails(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	writeFakeBinary(t, root)

	extract := func(string) (string, func(), error) {
		return "", func() {}, errors.New("boom")
	}

	c := checkAgentEmitEmbedAgainst(root, "", extract, false)

	if c.Status != uikit.CheckFail {
		t.Errorf("status = %q, want fail (extraction error)", c.Status)
	}
}

// --- the judgment itself --------------------------------------------------

// TestAgentEmitEmbed_MatchReportsCardinality: a clean comparison passes AND
// reports how many paths it compared, matching the committed artifact count.
func TestAgentEmitEmbed_MatchReportsCardinality(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml", "manager-docs.toml")
	writeFakeBinary(t, root)
	extracted := newExtractedDir(t, map[string]string{
		"manager-git.toml":  "name = \"manager-git.toml\"\n",
		"manager-docs.toml": "name = \"manager-docs.toml\"\n",
	})

	c := checkAgentEmitEmbedAgainst(root, "", staticExtractor(extracted), false)

	if c.Status != uikit.CheckOK {
		t.Fatalf("status = %q, want ok; message: %s", c.Status, c.Message)
	}
	if !strings.Contains(c.Message, "2/2") {
		t.Errorf("message = %q, want it to report the compared cardinality as 2/2", c.Message)
	}
}

// TestAgentEmitEmbed_DriftFailsAndNamesPath is the mutation face: a byte
// difference between the binary's embedded artifact and the committed one
// fails and names every differing path.
func TestAgentEmitEmbed_DriftFailsAndNamesPath(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml", "manager-docs.toml")
	writeFakeBinary(t, root)
	extracted := newExtractedDir(t, map[string]string{
		"manager-git.toml":  "name = \"manager-git.toml\"\n# stale\n",
		"manager-docs.toml": "name = \"manager-docs.toml\"\n",
	})

	c := checkAgentEmitEmbedAgainst(root, "", staticExtractor(extracted), false)

	if c.Status != uikit.CheckFail {
		t.Fatalf("status = %q, want fail (embedded bytes differ)", c.Status)
	}
	if !strings.Contains(c.Message+c.Detail, "manager-git.toml") {
		t.Errorf("output = %q / %q, want it to name manager-git.toml", c.Message, c.Detail)
	}
	if strings.Contains(c.Message, "manager-docs.toml") {
		t.Errorf("message = %q, must not name the matching artifact as differing", c.Message)
	}
}

// TestAgentEmitEmbed_PartialExtractionFails: an extraction that yields only a
// subset of the committed set must NOT pass by comparing that subset. This is
// the cardinality gate — the analogue of golden_test.go's `count != 11`.
func TestAgentEmitEmbed_PartialExtractionFails(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml", "manager-docs.toml")
	writeFakeBinary(t, root)
	extracted := newExtractedDir(t, map[string]string{
		"manager-git.toml": "name = \"manager-git.toml\"\n",
	})

	c := checkAgentEmitEmbedAgainst(root, "", staticExtractor(extracted), false)

	if c.Status != uikit.CheckFail {
		t.Fatalf("status = %q, want fail (compared 1 of 2)", c.Status)
	}
	if !strings.Contains(c.Message+c.Detail, "manager-docs.toml") {
		t.Errorf("output = %q / %q, want it to name the uncompared artifact", c.Message, c.Detail)
	}
}

// --- doctor wiring --------------------------------------------------------

// TestAgentEmitEmbed_RegisteredAndFilterable: the check is registered in the
// doctor catalog and its name is the value `moai doctor --check` accepts,
// selecting exactly one check.
func TestAgentEmitEmbed_RegisteredAndFilterable(t *testing.T) {
	t.Chdir(t.TempDir())

	groups := runGroupedChecks(false, agentEmitEmbedCheckName)

	var got []DiagnosticCheck
	for _, g := range groups {
		got = append(got, g.checks...)
	}
	if len(got) != 1 {
		t.Fatalf("--check %q selected %d checks, want exactly 1", agentEmitEmbedCheckName, len(got))
	}
	if got[0].Name != agentEmitEmbedCheckName {
		t.Errorf("selected check name = %q, want %q", got[0].Name, agentEmitEmbedCheckName)
	}
}

// TestAgentEmitEmbed_BinEnvOverride: the judgment target is overridable so the
// same logic can be aimed at an installed binary (the `make embed-check BIN=`
// path). The env value wins over the default root/bin/moai.
func TestAgentEmitEmbed_BinEnvOverride(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	writeFakeBinary(t, root)
	other := filepath.Join(t.TempDir(), "elsewhere-moai")
	if err := os.WriteFile(other, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write override binary: %v", err)
	}
	t.Setenv(embedCheckBinEnvKey, other)

	var seen string
	extract := func(bin string) (string, func(), error) {
		seen = bin
		return newExtractedDir(t, map[string]string{
			"manager-git.toml": "name = \"manager-git.toml\"\n",
		}), func() {}, nil
	}

	c := checkAgentEmitEmbedAgainst(root, "", extract, false)

	if c.Status != uikit.CheckOK {
		t.Fatalf("status = %q, want ok; message: %s", c.Status, c.Message)
	}
	if seen != other {
		t.Errorf("judgment target = %q, want the %s override %q", seen, embedCheckBinEnvKey, other)
	}
}

// TestAgentEmitEmbed_WritesNothingInsideTree is REQ-AEL-002's face for M1: the
// extraction scratch lives outside the repository tree, so a run leaves the
// project root byte-identical.
func TestAgentEmitEmbed_WritesNothingInsideTree(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	writeFakeBinary(t, root)
	before := embedSnapshotTree(t, root)

	extracted := newExtractedDir(t, map[string]string{
		"manager-git.toml": "name = \"manager-git.toml\"\n",
	})
	c := checkAgentEmitEmbedAgainst(root, "", staticExtractor(extracted), false)
	if c.Status != uikit.CheckOK {
		t.Fatalf("status = %q, want ok; message: %s", c.Status, c.Message)
	}

	after := embedSnapshotTree(t, root)
	if len(before) != len(after) {
		t.Fatalf("tree entry count changed: %d -> %d", len(before), len(after))
	}
	for p, sz := range before {
		if after[p] != sz {
			t.Errorf("%s: changed by the check (%d -> %d)", p, sz, after[p])
		}
	}
}

func embedSnapshotTree(t *testing.T, root string) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = info.Size()
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

// TestExtractEmissionViaInit_ResolvesRelativeBin pins a defect the stubbed
// extractor could not see: the extraction runs the target binary with its
// working directory set to the scratch dir, so a RELATIVE judgment-target
// path (the `make embed-check` default `bin/moai`) must be resolved against
// the caller's working directory before exec — otherwise it is looked up
// inside the scratch dir and fails as "no such file or directory".
func TestExtractEmissionViaInit_ResolvesRelativeBin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stand-in binary is POSIX-only")
	}
	work := t.TempDir()
	binDir := filepath.Join(work, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	// A stand-in for `moai init <target> --non-interactive`: writes the
	// deployed emission layout under $2 and exits 0.
	script := "#!/bin/sh\nmkdir -p \"$2/" + deployedEmissionRelDir + "\"\n" +
		"printf 'name = \"manager-git.toml\"\\n' > \"$2/" + deployedEmissionRelDir + "/manager-git.toml\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "moai"), []byte(script), 0o755); err != nil {
		t.Fatalf("write stand-in binary: %v", err)
	}
	t.Chdir(work)

	dir, cleanup, err := extractEmissionViaInit(filepath.Join("bin", "moai"))
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		t.Fatalf("extractEmissionViaInit with a relative bin path: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(deployedEmissionRelDir), "manager-git.toml"))
	if err != nil {
		t.Fatalf("extracted artifact unreadable: %v", err)
	}
	if !strings.Contains(string(got), "manager-git.toml") {
		t.Errorf("extracted body = %q, want the stand-in payload", string(got))
	}
}

// TestBoundedTail keeps a failed extraction's output from flooding a doctor
// row: long output is trimmed to its tail, short output passes through.
func TestBoundedTail(t *testing.T) {
	if got := boundedTail([]byte("  short  ")); got != "short" {
		t.Errorf("boundedTail(short) = %q, want %q", got, "short")
	}
	long := strings.Repeat("x", 500) + "END"
	got := boundedTail([]byte(long))
	if len(got) > 310 {
		t.Errorf("boundedTail(long) length = %d, want a bounded tail", len(got))
	}
	if !strings.HasSuffix(got, "END") {
		t.Errorf("boundedTail(long) = %q, want it to keep the tail", got)
	}
	if !strings.HasPrefix(got, "...") {
		t.Errorf("boundedTail(long) = %q, want the elision marker", got)
	}
}

// TestFindEmbedCheckRoot_SkipsStrayMarker pins the subdirectory-anchor defect
// found by running the check from a package directory in the real repository:
// a test side effect leaves a bare .moai/state/ directory there, so a walk
// that stops at the NEAREST .moai-bearing ancestor anchors on that artifact
// instead of the repository root. The applicable tree then reports "not
// applicable" — an applicable tree misjudged as skippable, which is the
// narrow re-entry of the vacuity this check exists to close.
//
// The anchor is therefore the nearest ancestor carrying the COMMITTED
// EMISSION SET, which is what the applicability predicate actually names.
func TestFindEmbedCheckRoot_SkipsStrayMarker(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	sub := filepath.Join(root, "pkgdir")
	if err := os.MkdirAll(filepath.Join(sub, projectRootMarker, "state"), 0o755); err != nil {
		t.Fatalf("mkdir stray marker: %v", err)
	}

	got, ok := findEmbedCheckRoot(sub)
	if !ok {
		t.Fatalf("findEmbedCheckRoot(%q) = not found, want the root above the stray marker", sub)
	}
	if !embedSameDir(t, got, root) {
		t.Errorf("findEmbedCheckRoot(%q) = %q, want %q — a stray marker must not anchor the walk", sub, got, root)
	}
}

// TestAgentEmitEmbed_StrayMarkerStillJudges is the same defect at the check
// level: a run from a subdirectory carrying a stray marker must still judge,
// never report not-applicable.
func TestAgentEmitEmbed_StrayMarkerStillJudges(t *testing.T) {
	root := newEmbedFixtureRoot(t, "manager-git.toml")
	writeFakeBinary(t, root)
	sub := filepath.Join(root, "pkgdir")
	if err := os.MkdirAll(filepath.Join(sub, projectRootMarker, "state"), 0o755); err != nil {
		t.Fatalf("mkdir stray marker: %v", err)
	}
	extracted := newExtractedDir(t, map[string]string{
		"manager-git.toml": "name = \"manager-git.toml\"\n# stale\n",
	})

	c := checkAgentEmitEmbedAgainst(sub, "", staticExtractor(extracted), false)

	if c.Status != uikit.CheckFail {
		t.Fatalf("status = %q, want fail — an applicable tree must not flip to not-applicable when run from a subdirectory; message: %s",
			c.Status, c.Message)
	}
}
