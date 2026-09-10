package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// runStampCommit executes `graph stamp codemaps --root <root> [--commit <rev>]`
// and returns (stdout, stderr, execution error).
func runStampCommit(t *testing.T, root, commitRev string) (string, string, error) {
	t.Helper()
	cmd := newGraphCmd()
	args := []string{"stamp", "codemaps", "--root", root}
	if commitRev != "" {
		args = append(args, "--commit", commitRev)
	}
	cmd.SetArgs(args)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

// stampTargetPath returns the fixture's provenance.json path.
func stampTargetPath(root string) string {
	return filepath.Join(root, ".moai", "project", "codemaps", "provenance.json")
}

// TestGraphStampCmd_ExplicitCommitRecordsResolvedSHA pins AC-SP-005: a clean
// described-source tree plus --commit <parent rev> records the resolved FULL
// sha verbatim (not HEAD), dirty false, schema 1, default described roots.
func TestGraphStampCmd_ExplicitCommitRecordsResolvedSHA(t *testing.T) {
	root := stampFixture(t)
	parent := checkFixtureGit(t, root, "rev-parse", "HEAD")
	// Second commit so the fixture has a HEAD distinct from the named parent.
	if err := os.WriteFile(filepath.Join(root, "internal", "b.go"), []byte("package a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "-A")
	checkFixtureGit(t, root, "commit", "-q", "-m", "second")
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")

	out, _, err := runStampCommit(t, root, parent)
	if err != nil {
		t.Fatalf("stamp --commit: %v (out %s)", err, out)
	}
	data, err := os.ReadFile(stampTargetPath(root))
	if err != nil {
		t.Fatalf("provenance.json not written: %v", err)
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil {
		t.Fatalf("provenance.json not valid JSON: %v", err)
	}
	if pv.CommitSHA != parent {
		t.Errorf("commit_sha = %q, want named parent %q (not HEAD %q)", pv.CommitSHA, parent, head)
	}
	if pv.Dirty {
		t.Errorf("clean tree with --commit must set dirty=false")
	}
	if pv.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", pv.SchemaVersion)
	}
	if len(pv.DescribedRoots) != 3 || pv.DescribedRoots[0] != "internal" || pv.DescribedRoots[1] != "cmd" || pv.DescribedRoots[2] != "pkg" {
		t.Errorf("described_roots = %v, want [internal cmd pkg]", pv.DescribedRoots)
	}
	if !strings.Contains(out, "OK: stamped") {
		t.Errorf("output missing OK line: %s", out)
	}
}

// TestGraphStampCmd_ExplicitCommitShortRevResolvesFull pins the edge case: a
// 7-char short rev is accepted and the FULL sha lands in the file.
func TestGraphStampCmd_ExplicitCommitShortRevResolvesFull(t *testing.T) {
	root := stampFixture(t)
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")

	if _, _, err := runStampCommit(t, root, head[:7]); err != nil {
		t.Fatalf("stamp --commit <short>: %v", err)
	}
	data, err := os.ReadFile(stampTargetPath(root))
	if err != nil {
		t.Fatal(err)
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil {
		t.Fatal(err)
	}
	if pv.CommitSHA != head {
		t.Errorf("short rev recorded as %q, want full sha %q", pv.CommitSHA, head)
	}
}

// TestGraphStampCmd_UnresolvableCommitRejected pins AC-SP-006: an
// unresolvable revision exits non-zero with a path-free error and writes NO
// provenance file (neither target nor temp).
func TestGraphStampCmd_UnresolvableCommitRejected(t *testing.T) {
	root := stampFixture(t)

	_, errOut, err := runStampCommit(t, root, "deadbeef42")
	if err == nil {
		t.Fatal("unresolvable --commit must exit non-zero")
	}
	surface := err.Error() + "\n" + errOut
	if strings.Contains(surface, root) {
		t.Errorf("unresolvable error leaks the absolute local path: %q", surface)
	}
	if !strings.Contains(surface, "deadbeef42") {
		t.Errorf("unresolvable error must name the revision: %q", surface)
	}
	if _, statErr := os.Stat(stampTargetPath(root)); !os.IsNotExist(statErr) {
		t.Errorf("failed stamp must write no provenance file (target exists or stat error: %v)", statErr)
	}
	if _, statErr := os.Stat(stampTargetPath(root) + ".tmp"); !os.IsNotExist(statErr) {
		t.Errorf("failed stamp must leave no temp file (stat error: %v)", statErr)
	}
}

// TestGraphStampCmd_CommitWithDirtyTreeRejectedPreWrite pins AC-SP-007: a
// valid --commit over a tree with uncommitted described-source changes exits
// non-zero with a path-free error and writes NO file — the two anchors are
// mutually exclusive honesty claims (REQ-SR-006).
func TestGraphStampCmd_CommitWithDirtyTreeRejectedPreWrite(t *testing.T) {
	root := stampFixture(t)
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a // dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := runStampCommit(t, root, head)
	if err == nil {
		t.Fatal("--commit over a dirty tree must exit non-zero")
	}
	surface := err.Error() + "\n" + errOut
	if strings.Contains(surface, root) {
		t.Errorf("dirty-combination error leaks the absolute local path: %q", surface)
	}
	if !strings.Contains(surface, "uncommitted") && !strings.Contains(surface, "dirty") {
		t.Errorf("dirty-combination error must name the conflict: %q", surface)
	}
	if _, statErr := os.Stat(stampTargetPath(root)); !os.IsNotExist(statErr) {
		t.Errorf("rejected stamp must write no provenance file (stat error: %v)", statErr)
	}
	if _, statErr := os.Stat(stampTargetPath(root) + ".tmp"); !os.IsNotExist(statErr) {
		t.Errorf("rejected stamp must leave no temp file (stat error: %v)", statErr)
	}
	// The tree itself is untouched by the rejection.
	if _, err := os.Stat(filepath.Join(root, "internal", "a.go")); err != nil {
		t.Errorf("rejection must not modify the working tree: %v", err)
	}
}

// TestGraphStampCmd_FlaglessCharacterization pins AC-SP-008 (REQ-SR-007): the
// flagless default path is behavior-preserved — clean tree records HEAD,
// dirty tree records dirty + 64-hex fingerprint with no commit anchor, and
// the Describe line and JSON key set match the pre-SPEC shapes exactly.
func TestGraphStampCmd_FlaglessCharacterization(t *testing.T) {
	root := stampFixture(t)
	head := checkFixtureGit(t, root, "rev-parse", "HEAD")

	out, _, err := runStampCommit(t, root, "")
	if err != nil {
		t.Fatalf("flagless stamp: %v", err)
	}
	data, err := os.ReadFile(stampTargetPath(root))
	if err != nil {
		t.Fatal(err)
	}
	var pv mx.Provenance
	if err := json.Unmarshal(data, &pv); err != nil {
		t.Fatal(err)
	}
	if pv.CommitSHA != head || pv.Dirty {
		t.Errorf("flagless clean anchor = (%q, dirty=%v), want (%q, false)", pv.CommitSHA, pv.Dirty, head)
	}
	wantDescribe := "provenance: tree=" + root + " commit=" + head[:12]
	if !strings.Contains(out, wantDescribe) {
		t.Errorf("clean Describe line = missing %q in %q", wantDescribe, out)
	}
	// Clean anchor key set: no content_fingerprint key in the serialized JSON.
	if strings.Contains(string(data), "content_fingerprint") {
		t.Errorf("clean anchor must not serialize a content_fingerprint: %s", data)
	}

	// --commit equal to HEAD is a legal no-op relative to the default path:
	// the same anchor lands (edge case, acceptance.md §E).
	if _, _, err := runStampCommit(t, root, head); err != nil {
		t.Fatalf("--commit HEAD over clean tree must succeed: %v", err)
	}
	dataHead, err := os.ReadFile(stampTargetPath(root))
	if err != nil {
		t.Fatal(err)
	}
	var pvHead mx.Provenance
	if err := json.Unmarshal(dataHead, &pvHead); err != nil {
		t.Fatal(err)
	}
	if pvHead.CommitSHA != head || pvHead.Dirty {
		t.Errorf("--commit HEAD anchor = (%q, dirty=%v), want (%q, false)", pvHead.CommitSHA, pvHead.Dirty, head)
	}

	// Dirty variant: uncommitted described-source change flips the anchor to
	// dirty + fingerprint, commit_sha omitted.
	if err := os.WriteFile(filepath.Join(root, "internal", "a.go"), []byte("package a // dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out2, _, err := runStampCommit(t, root, "")
	if err != nil {
		t.Fatalf("flagless dirty stamp: %v", err)
	}
	data2, err := os.ReadFile(stampTargetPath(root))
	if err != nil {
		t.Fatal(err)
	}
	var pv2 mx.Provenance
	if err := json.Unmarshal(data2, &pv2); err != nil {
		t.Fatal(err)
	}
	if pv2.CommitSHA != "" || !pv2.Dirty || len(pv2.ContentFingerprint) != 64 {
		t.Errorf("dirty anchor = (commit %q, dirty=%v, fp len %d), want empty commit + dirty + 64-hex", pv2.CommitSHA, pv2.Dirty, len(pv2.ContentFingerprint))
	}
	if !strings.Contains(out2, "dirty fingerprint=") {
		t.Errorf("dirty Describe line missing fingerprint form: %q", out2)
	}
	if strings.Contains(string(data2), "commit_sha") {
		t.Errorf("dirty anchor must omit commit_sha entirely: %s", data2)
	}
}

// TestGraphStampCmd_LongTextCarriesRecipe pins the REQ-SR-010 CLI half: the
// command's Long text documents the explicit-commit mode, the merge-base
// recipe, and the branch-local prohibition.
func TestGraphStampCmd_LongTextCarriesRecipe(t *testing.T) {
	cmd := newGraphStampCmd()
	long := cmd.Long
	for _, token := range []string{"--commit", "merge-base HEAD origin/main", "branch-local"} {
		if !strings.Contains(long, token) {
			t.Errorf("Long text missing %q", token)
		}
	}
}
