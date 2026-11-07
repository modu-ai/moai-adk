package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// commitAgedFiles commits n new described-source files as one commit.
func commitAgedFiles(t *testing.T, root string, n int) {
	t.Helper()
	dir := filepath.Join(root, "internal", "aged")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := range n {
		p := filepath.Join(dir, fmt.Sprintf("f%03d.go", i))
		if err := os.WriteFile(p, []byte("package aged\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", fmt.Sprintf("add %d files", n))
}

// restampExceptCodemaps re-stamps every layer, then puts the codemaps stamp
// back on the given (older) commit, leaving only codemaps non-fresh.
func restampExceptCodemaps(t *testing.T, root, codemapsCommit string) {
	t.Helper()
	stampAllLayers(t, root)
	pv := filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
	if err := os.WriteFile(pv, marshalCodemapsProvenance(t, root, codemapsCommit), 0o644); err != nil {
		t.Fatal(err)
	}
	// The edges fingerprint covers the codemaps source set, so it must be
	// re-recorded after the codemaps stamp moves.
	meta := filepath.Join(root, ".moai", "project", "graph", graph.MetaFileName)
	if err := graph.WriteEdgesMeta(meta, root, graph.SourceFingerprintsForEdges(root), 0); err != nil {
		t.Fatal(err)
	}
}

// commitCodemaps commits the codemaps body and a provenance stamp pointing at
// stampSHA, exactly the bytes stampAllLayers/restampExceptCodemaps write, so a
// later re-stamp leaves the working tree identical to the commit.
func commitCodemaps(t *testing.T, root, stampSHA string) {
	t.Helper()
	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmDir, "provenance.json"), marshalCodemapsProvenance(t, root, stampSHA), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmDir, "modules.md"), []byte("# modules\n\ninternal/alpha/alpha.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "--", ".moai/project/codemaps")
	checkFixtureGit(t, root, "commit", "-q", "-m", "codemaps")
}

// rewoundStampFixture is the F2 exploit shape. The BASE is fresh: 41 described
// files landed in c1 and the committed stamp points at c1. The change under
// judgment is docs-only (contribution 0) but rewinds the stamp to c0, which
// drags the measurement window back over those 41 files.
func rewoundStampFixture(t *testing.T) string {
	t.Helper()
	root := newCheckCLIRepo(t)
	c0 := checkFixtureGit(t, root, "rev-parse", "HEAD")
	commitAgedFiles(t, root, 41)
	c1 := checkFixtureGit(t, root, "rev-parse", "HEAD")
	commitCodemaps(t, root, c1)
	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte("docs only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pv := filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
	if err := os.WriteFile(pv, marshalCodemapsProvenance(t, root, c0), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "--", "NOTES.md", ".moai/project/codemaps/provenance.json")
	checkFixtureGit(t, root, "commit", "-q", "-m", "docs + rewound stamp")
	restampExceptCodemaps(t, root, c0)
	return root
}

// F2 RED: a change that rewinds the codemaps stamp moved the measurement window,
// so Value - Contribution is not the base's count. It must block, and say why.
func TestGraphCheckPolicy_RewoundStampBlocks(t *testing.T) {
	root := rewoundStampFixture(t)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 1 {
		t.Fatalf("a rewound stamp over a fresh base must exit 1, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "anchor-moved"`) {
		t.Errorf("origin must be anchor-moved:\n%s", out)
	}
}

// Control: with the stamp and body untouched by the change, a base that already
// carries the count still warns (the narrow R208 downgrade is intact).
func TestGraphCheckPolicy_UntouchedStampStillWarns(t *testing.T) {
	root := driftFixture(t, 41, 0)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 0 {
		t.Fatalf("untouched stamp, inherited count must exit 0, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"codemaps_inputs_stable": true`) {
		t.Errorf("the stability fact must be reported as true:\n%s", out)
	}
}

// inheritedFixture: codemaps stamped at the base commit; 41 files landed in an
// earlier (base-side) commit; this change adds ownFiles files on top; every
// other layer is fresh.
func inheritedFixture(t *testing.T, ownFiles int) string {
	t.Helper()
	return driftFixture(t, 41, ownFiles)
}

// driftFixture: codemaps stamped at the base commit; baseFiles described files
// landed base-side; this change then adds ownFiles described files (0 = a
// docs-only commit that contributes nothing). Every other layer is fresh.
func driftFixture(t *testing.T, baseFiles, ownFiles int) string {
	t.Helper()
	root := newCheckCLIRepo(t)
	base := checkFixtureGit(t, root, "rev-parse", "HEAD")
	// Real CI checks out codemaps as a TRACKED artifact, so the base commit
	// carries the body and its stamp.
	commitCodemaps(t, root, base)
	if baseFiles > 0 {
		commitAgedFiles(t, root, baseFiles)
	}
	dir := filepath.Join(root, "internal", "own")
	if ownFiles > 0 {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := range ownFiles {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("o%03d.go", i)), []byte("package own\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	} else if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte("docs only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", "own change")
	restampExceptCodemaps(t, root, base)
	return root
}

// R216 RED: base 25 + this change 16 = 41. The change took the count over the
// line, so it is not inherited and must block.
func TestGraphCheckPolicy_CrossedByChangeBlocks(t *testing.T) {
	root := driftFixture(t, 25, 16)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 1 {
		t.Fatalf("base 25 + pr 16 must exit 1, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "crossed-by-change"`) {
		t.Errorf("origin must be stated as crossed-by-change:\n%s", out)
	}
}

// Base alone is at the line and this change contributed nothing: inherited.
func TestGraphCheckPolicy_BaseOverLinePrZeroWarns(t *testing.T) {
	root := driftFixture(t, 41, 0)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 0 {
		t.Fatalf("base 41 + pr 0 must exit 0, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "inherited"`) {
		t.Errorf("origin must be inherited:\n%s", out)
	}
}

// This change alone is over the line: originated, blocks.
func TestGraphCheckPolicy_OriginatedByChangeBlocks(t *testing.T) {
	root := driftFixture(t, 0, 41)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 1 {
		t.Fatalf("base 0 + pr 41 must exit 1, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "originated"`) {
		t.Errorf("origin must be originated:\n%s", out)
	}
}

func writeGatePolicy(t *testing.T, root, policy string) {
	t.Helper()
	cfg := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	y := "gate:\n  graph_freshness:\n    enabled: true\n    blocking: false\n    codemaps_changed_files: 40\n    mx_index_changed_files: 1\n    update_budget_ms: 2000\n    inherited_codemaps_drift: " + policy + "\n"
	if err := os.WriteFile(filepath.Join(cfg, "gate.yaml"), []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGraphCheckPolicy(t *testing.T, root string, extra ...string) (string, string, int) {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SetArgs(append([]string{"check", "--root", root}, extra...))
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	if err == nil {
		return out.String(), errOut.String(), 0
	}
	var ce interface{ ExitCode() int }
	if !errors.As(err, &ce) {
		t.Fatalf("error without exit code: %T %v", err, err)
	}
	return out.String(), errOut.String(), ce.ExitCode()
}

// R208 RED→GREEN: an inherited codemaps drift on a pull_request exits 0 with a
// warning, and the verdict stays stale.
func TestGraphCheckPolicy_InheritedPullRequestWarns(t *testing.T) {
	root := inheritedFixture(t, 1)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 0 {
		t.Fatalf("inherited pull_request drift must exit 0, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"verdict": "stale"`) || !strings.Contains(out, `"verdict_class": "warn"`) {
		t.Errorf("verdict must stay stale with class warn:\n%s", out)
	}
	if !strings.Contains(errOut, "::warning") {
		t.Errorf("a ::warning annotation is required:\n%s", errOut)
	}
}

// Negative: self-contribution at/over the threshold still blocks.
func TestGraphCheckPolicy_SelfContributionBlocks(t *testing.T) {
	root := inheritedFixture(t, 41)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("self-contribution 41 must exit 1, got %d\n%s", code, errOut)
	}
}

// Negative: the same inherited drift on a push event blocks.
func TestGraphCheckPolicy_PushInheritedBlocks(t *testing.T) {
	root := inheritedFixture(t, 1)
	t.Setenv("GITHUB_EVENT_NAME", "push")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("push inherited drift must exit 1, got %d\n%s", code, errOut)
	}
}

// Fail-closed: no event at all blocks.
func TestGraphCheckPolicy_NoEventBlocks(t *testing.T) {
	root := inheritedFixture(t, 1)
	t.Setenv("GITHUB_EVENT_NAME", "")
	if _, _, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("empty event must exit 1, got %d", code)
	}
}

// --event overrides the environment.
func TestGraphCheckPolicy_EventFlagOverridesEnv(t *testing.T) {
	root := inheritedFixture(t, 1)
	t.Setenv("GITHUB_EVENT_NAME", "push")
	if _, errOut, code := runGraphCheckPolicy(t, root, "--event", "pull_request"); code != 0 {
		t.Fatalf("--event pull_request must exit 0, got %d\n%s", code, errOut)
	}
}

// Negative: an absent layer blocks even on a pull_request.
func TestGraphCheckPolicy_AbsentLayerBlocks(t *testing.T) {
	root := inheritedFixture(t, 1)
	if err := os.Remove(filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("absent edges must exit 1, got %d\n%s", code, errOut)
	}
}

// Negative: a second non-fresh layer (source drift under mx/edges) blocks.
func TestGraphCheckPolicy_OtherLayerStaleBlocks(t *testing.T) {
	root := inheritedFixture(t, 1)
	if err := os.WriteFile(filepath.Join(root, "internal", "alpha", "alpha.go"), []byte("package alpha\n// changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("other-layer drift must exit 1, got %d\n%s", code, errOut)
	}
}

func TestGraphCheckPolicy_PolicyBlockRestoresExit1(t *testing.T) {
	root := inheritedFixture(t, 1)
	writeGatePolicy(t, root, "block")
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 1 {
		t.Fatalf("policy=block must exit 1, got %d\n%s", code, errOut)
	}
}

func TestGraphCheckPolicy_InvalidPolicyExitsTwo(t *testing.T) {
	root := inheritedFixture(t, 1)
	writeGatePolicy(t, root, "nope")
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	if _, errOut, code := runGraphCheckPolicy(t, root); code != 2 {
		t.Fatalf("invalid policy must exit 2, got %d\n%s", code, errOut)
	}
}

// symlinkedStampFixture is the N1 shape: the codemaps stamp is a tracked
// symlink to a file OUTSIDE the codemaps directory. The BASE is fresh (the
// target stamps c1, after 41 described files). The change under judgment is
// docs-only but rewinds the symlink TARGET to c0, which git sees as a change
// to a different path: the codemaps directory diff stays empty while the bytes
// read through the link move the measurement window. It skips where symlinks
// cannot be created.
func symlinkedStampFixture(t *testing.T) string {
	t.Helper()
	root := newCheckCLIRepo(t)
	c0 := checkFixtureGit(t, root, "rev-parse", "HEAD")
	commitAgedFiles(t, root, 41)
	c1 := checkFixtureGit(t, root, "rev-parse", "HEAD")

	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "stamp-target.json")
	if err := os.WriteFile(target, marshalCodemapsProvenance(t, root, c1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmDir, "modules.md"), []byte("# modules\n\ninternal/alpha/alpha.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", "..", "..", "stamp-target.json"), filepath.Join(cmDir, "provenance.json")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	checkFixtureGit(t, root, "add", "--", "stamp-target.json", ".moai/project/codemaps")
	checkFixtureGit(t, root, "commit", "-q", "-m", "codemaps with linked stamp")

	if err := os.WriteFile(target, marshalCodemapsProvenance(t, root, c0), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte("docs only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "--", "stamp-target.json", "NOTES.md")
	checkFixtureGit(t, root, "commit", "-q", "-m", "docs + rewound linked stamp target")
	restampExceptCodemaps(t, root, c0)
	return root
}

// N1 RED: stability read from git path lists cannot see a rewound symlink
// target; the stamp actually read is not what the first parent carried, so the
// layer must block as anchor-moved.
func TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks(t *testing.T) {
	root := symlinkedStampFixture(t)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 1 {
		t.Fatalf("a rewound symlinked stamp must exit 1, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "anchor-moved"`) {
		t.Errorf("origin must be anchor-moved:\n%s", out)
	}
}

// ignoredStampFixture is the N1 variant where provenance.json is ignored by git:
// it never shows in any path list, tracked or untracked, yet it is the stamp
// that gets read. Its on-disk content points at c0.
func ignoredStampFixture(t *testing.T) string {
	t.Helper()
	root := newCheckCLIRepo(t)
	c0 := checkFixtureGit(t, root, "rev-parse", "HEAD")
	commitAgedFiles(t, root, 41)

	cmDir := filepath.Join(root, ".moai", "project", "codemaps")
	if err := os.MkdirAll(cmDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".moai/project/codemaps/provenance.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmDir, "modules.md"), []byte("# modules\n\ninternal/alpha/alpha.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "--", ".gitignore", ".moai/project/codemaps/modules.md")
	checkFixtureGit(t, root, "commit", "-q", "-m", "codemaps body, ignored stamp")
	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte("docs only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "--", "NOTES.md")
	checkFixtureGit(t, root, "commit", "-q", "-m", "docs only")
	restampExceptCodemaps(t, root, c0)
	return root
}

// N1 RED: an ignored stamp cannot be compared with the first parent at all, so
// sameness is unprovable and the layer must block as anchor-moved.
func TestGraphCheckPolicy_IgnoredStampBlocks(t *testing.T) {
	root := ignoredStampFixture(t)
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	out, errOut, code := runGraphCheckPolicy(t, root, "--json")
	if code != 1 {
		t.Fatalf("an ignored stamp must exit 1, got %d\nstdout=%s\nstderr=%s", code, out, errOut)
	}
	if !strings.Contains(out, `"drift_origin": "anchor-moved"`) {
		t.Errorf("origin must be anchor-moved:\n%s", out)
	}
}
