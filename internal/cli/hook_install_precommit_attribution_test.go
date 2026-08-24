package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pinnedReleasedHookTag is the tag AC-PCP-005 sub-case (c) pins to: the most
// recent released tag whose hook body is byte-identical to the incoming body
// (acceptance.md §D.4). A legitimate hook-body change re-pins this constant in
// the same commit as the body change; it is never deleted.
const pinnedReleasedHookTag = "v3.1.2"

// previousPreCommitHookContent stands in for a hook body MoAI shipped in an
// earlier release: marker-bearing, and different from preCommitHookContent.
const previousPreCommitHookContent = `#!/bin/sh
# MoAI-ADK pre-commit hook — previous release body
set -eu
exit 0
`

// newPreCommitTestRepo creates a git repository in a temp dir and returns its
// root. git is required; its absence fails rather than skips, so an unmeasured
// criterion cannot read as a pass.
func newPreCommitTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init %s: %v\n%s", dir, err, out)
	}
	return dir
}

func precommitHookPath(root string) string {
	return filepath.Join(root, ".git", "hooks", "pre-commit")
}

func precommitRecordPath(root string) string {
	return filepath.Join(root, ".git", "hooks", moaiPreCommitProvenanceName)
}

// writeExistingHook plants an installed hook at the hook path.
func writeExistingHook(t *testing.T, root, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(precommitHookPath(root)), 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	if err := os.WriteFile(precommitHookPath(root), []byte(content), 0o755); err != nil {
		t.Fatalf("write hook: %v", err)
	}
}

// writeRecord plants a provenance record holding the given digest text.
func writeRecord(t *testing.T, root, digest string) {
	t.Helper()
	if err := os.WriteFile(precommitRecordPath(root), []byte(digest+"\n"), 0o644); err != nil {
		t.Fatalf("write record: %v", err)
	}
}

func digestOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func readRecord(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(precommitRecordPath(root))
	if err != nil {
		t.Fatalf("read provenance record: %v", err)
	}
	return strings.TrimSpace(string(raw))
}

func readHook(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(precommitHookPath(root))
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	return string(raw)
}

// assertNoBackup fails when any pre-commit.bak.* artifact exists.
func assertNoBackup(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".git", "hooks", "pre-commit.bak.*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected no backup file, found %v", matches)
	}
}

// installWithContent runs an install whose incoming content is the given string,
// returning the installer so the run's attribution can be inspected.
func installWithContent(t *testing.T, root, incoming string) *PreCommitInstaller {
	t.Helper()
	installer := NewPreCommitInstaller(root)
	installer.content = incoming
	if err := installer.InstallPreCommitHook(false); err != nil {
		t.Fatalf("InstallPreCommitHook: %v", err)
	}
	return installer
}

func assertAttribution(t *testing.T, got *preCommitAttribution, wantClass preCommitClass, wantBasis preCommitBasis) {
	t.Helper()
	if got == nil {
		t.Fatalf("no attribution recorded: expected class=%v basis=%v", wantClass, wantBasis)
	}
	if got.Class != wantClass || got.Basis != wantBasis {
		t.Errorf("attribution = {class:%v basis:%v}, want {class:%v basis:%v}",
			got.Class, got.Basis, wantClass, wantBasis)
	}
}

// TestPreCommitProvenanceRecorded — AC-PCP-001 (REQ-PCP-001).
//
// The second run is mandatory: it defeats the mutant that writes the sidecar
// once at first install and never refreshes it.
func TestPreCommitProvenanceRecorded(t *testing.T) {
	root := newPreCommitTestRepo(t)

	first := installWithContent(t, root, preCommitHookContent)
	if first.lastAttribution != nil {
		t.Errorf("fresh install classified something: %+v", first.lastAttribution)
	}
	if got, want := readRecord(t, root), digestOf(preCommitHookContent); got != want {
		t.Errorf("after run 1: record = %q, want %q", got, want)
	}
	if first.lastProvenanceErr != nil {
		t.Errorf("provenance write reported an error on a writable hooks dir: %v", first.lastProvenanceErr)
	}

	const bumped = preCommitHookContent + "\n# version bump\n"
	installWithContent(t, root, bumped)
	if got, want := readRecord(t, root), digestOf(bumped); got != want {
		t.Errorf("after run 2: record = %q, want %q (the record must be refreshed on every write)", got, want)
	}
	if got := readHook(t, root); got != bumped {
		t.Errorf("hook content after run 2 = %d bytes, want the bumped content (%d bytes)", len(got), len(bumped))
	}
}

// TestPreCommitVersionBumpIsSilent — AC-PCP-002 (REQ-PCP-002).
//
// The Given is reached through the real caller path: the first install writes a
// body X and stamps sha256(X); the second install comes through
// installPreCommitHookOptional, whose incoming content is the shipped constant
// and therefore differs from X. installed == recorded, so the run must be a
// silent version bump.
func TestPreCommitVersionBumpIsSilent(t *testing.T) {
	root := newPreCommitTestRepo(t)
	installWithContent(t, root, previousPreCommitHookContent)

	var buf bytes.Buffer
	installPreCommitHookOptional(root, false, &buf)

	if got := readHook(t, root); got != preCommitHookContent {
		t.Errorf("hook was not replaced: got %d bytes, want the incoming content (%d bytes)", len(got), len(preCommitHookContent))
	}
	assertNoBackup(t, root)

	const wantLine = "  Pre-commit hook installed (.git/hooks/pre-commit)\n"
	if buf.String() != wantLine {
		t.Errorf("output = %q, want exactly %q (a version bump produces no backup notice)", buf.String(), wantLine)
	}
	if got, want := readRecord(t, root), digestOf(preCommitHookContent); got != want {
		t.Errorf("record = %q, want %q", got, want)
	}

	// The same Given, run through the direct installer so the verdict itself is
	// observable. Without this arm the criterion is silent about the naive
	// two-way design: M1 takes no backup and prints no notice under any
	// classification, so the file-and-output assertions above pass for a
	// two-way implementation too. This arm is what makes AC-PCP-002's stated
	// failing input — "backs up whenever installed != incoming" — go red.
	root2 := newPreCommitTestRepo(t)
	installWithContent(t, root2, previousPreCommitHookContent)
	bump := installWithContent(t, root2, preCommitHookContent)
	assertAttribution(t, bump.lastAttribution, preCommitUnmodified, preCommitBasisRecord)
}

// TestPreCommitThreeWayAttribution — AC-PCP-014 (REQ-PCP-014).
//
// Case one is mandatory: the two-way implementation (installed vs incoming
// only) passes case two by luck and fails case one. A single-case version of
// this test is satisfiable by a two-way design and is not acceptable evidence.
func TestPreCommitThreeWayAttribution(t *testing.T) {
	const incoming = preCommitHookContent + "\n# upstream bump\n"

	t.Run("record_matches_installed_version_bump", func(t *testing.T) {
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, previousPreCommitHookContent)
		writeRecord(t, root, digestOf(previousPreCommitHookContent))

		installer := installWithContent(t, root, incoming)

		assertAttribution(t, installer.lastAttribution, preCommitUnmodified, preCommitBasisRecord)
		assertNoBackup(t, root)
		if got := readHook(t, root); got != incoming {
			t.Errorf("hook was not replaced with the incoming content")
		}
	})

	t.Run("record_differs_from_installed_user_edit", func(t *testing.T) {
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, previousPreCommitHookContent)
		writeRecord(t, root, digestOf("something else MoAI once wrote\n"))

		installer := installWithContent(t, root, incoming)

		assertAttribution(t, installer.lastAttribution, preCommitUserModified, preCommitBasisRecord)
	})
}

// TestPreCommitLegacyNoRecord — AC-PCP-005 (REQ-PCP-005).
//
// Sub-case (a) is mandatory: (b) and (c) alone are passed by the mutant that
// treats a missing record as "unmodified". Sub-case (c) measures the real
// installed population and reads its fixture from git; when git cannot be
// reached it FAILS rather than skipping — a skipped (c) is a gap, not a pass.
func TestPreCommitLegacyNoRecord(t *testing.T) {
	t.Run("a_no_record_content_differs", func(t *testing.T) {
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, previousPreCommitHookContent)

		installer := installWithContent(t, root, preCommitHookContent)

		assertAttribution(t, installer.lastAttribution, preCommitUserModified, preCommitBasisUndecidableLegacy)
	})

	t.Run("b_no_record_content_identical", func(t *testing.T) {
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, preCommitHookContent)

		installer := installWithContent(t, root, preCommitHookContent)

		assertAttribution(t, installer.lastAttribution, preCommitUnmodified, preCommitBasisUndecidableLegacy)
		assertNoBackup(t, root)
		if got, want := readRecord(t, root), digestOf(preCommitHookContent); got != want {
			t.Errorf("record = %q, want %q", got, want)
		}
	})

	t.Run("c_no_record_pinned_released_body", func(t *testing.T) {
		released := releasedHookBody(t)
		root := newPreCommitTestRepo(t)
		writeExistingHook(t, root, released)

		installer := installWithContent(t, root, preCommitHookContent)

		assertAttribution(t, installer.lastAttribution, preCommitUnmodified, preCommitBasisUndecidableLegacy)
		assertNoBackup(t, root)
		if got, want := readRecord(t, root), digestOf(preCommitHookContent); got != want {
			t.Errorf("record = %q, want %q", got, want)
		}
	})
}

// releasedHookBody reads the pinned released hook body from git. Any failure to
// reach git is fatal, never a skip: the criterion is red until it actually runs
// (acceptance.md AC-PCP-005 Decides). Forks, tarballs and shallow clones are the
// stated exposure.
func releasedHookBody(t *testing.T) string {
	t.Helper()
	root := precommitProjectRoot(t)
	ref := fmt.Sprintf("%s:internal/template/templates/.git_hooks/pre-commit", pinnedReleasedHookTag)
	cmd := exec.Command("git", "show", ref)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("cannot read the pinned released hook body %q from git in %s: %v\n"+
			"AC-PCP-005 sub-case (c) is red until it runs — a skip here would report an "+
			"unmeasured population as a pass. Fetch tags (fetch-depth: 0) and re-run.", ref, root, err)
	}
	if len(out) == 0 {
		t.Fatalf("pinned released hook body %q is empty", ref)
	}
	return string(out)
}
