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

// inheritedFixture: codemaps stamped at the base commit; 41 files landed in an
// earlier (base-side) commit; this change adds ownFiles files on top; every
// other layer is fresh.
func inheritedFixture(t *testing.T, ownFiles int) string {
	t.Helper()
	root := newCheckCLIRepo(t)
	base := checkFixtureGit(t, root, "rev-parse", "HEAD")
	commitAgedFiles(t, root, 41)
	if ownFiles > 0 {
		dir := filepath.Join(root, "internal", "own")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := range ownFiles {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("o%03d.go", i)), []byte("package own\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		checkFixtureGit(t, root, "add", "-A")
		checkFixtureGit(t, root, "commit", "-q", "-m", "own change")
	}
	restampExceptCodemaps(t, root, base)
	return root
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
