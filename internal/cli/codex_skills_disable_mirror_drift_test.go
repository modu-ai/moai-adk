package cli

// codex_skills_disable_mirror_drift_test.go — the drift-following guard for
// the disable verb (t699).
//
// resolveCodexSkillMirrorPath reads the producer's exported layout symbol
// (template.MirrorSkillsRelDir), the same binding the doctor-side reader has
// carried since t521. A reader that restated the path as its own literal
// passes every fixture pinned to the default location, so the discriminating
// run REPOINTS the producer and requires the verb to follow: under a
// re-injected literal the gate keeps resolving the old path and fails here
// with the exact false "mirror absent" resolution.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// writeMirrorCopySkill creates the copy-fallback mirror shape the disable
// gate resolves against: <root>/<mirrorRel>/<name>/SKILL.md as a regular
// file. (A symlink entry stat's to the same regular file; this helper pins
// the stricter shape.)
func writeMirrorCopySkill(t *testing.T, root, mirrorRel, name string) {
	t.Helper()
	dir := filepath.Join(root, mirrorRel, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestResolveCodexSkillMirrorPathFollowsProducerMirrorRelDir is the two-run
// guard. Run one pins normal operation at the producer's default location;
// run two repoints the producer and requires the gate to follow — the
// discriminator against a re-injected literal, which keeps resolving the old
// path and fails here. Non-parallel: the producer variable is package-global
// state, the same discipline the osStatFn and separator seams carry.
func TestResolveCodexSkillMirrorPathFollowsProducerMirrorRelDir(t *testing.T) {
	// Run 1 — the producer at its default location: a skill published there
	// must resolve. (A stale binding agrees with the default by accident, so
	// this run alone is not the discriminator; it keeps the guard from
	// passing vacuously once the binding exists.)
	root := t.TempDir()
	writeMirrorCopySkill(t, root, template.MirrorSkillsRelDir, "moai-demo")
	res := resolveCodexSkillMirrorPath(root, "", "moai-demo")
	if res.Outcome != codexSkillResolved {
		t.Fatalf("run 1 (default location %q): outcome = %v, want resolved (reason: %s)",
			template.MirrorSkillsRelDir, res.Outcome, res.Reason)
	}

	// Run 2 — repoint the producer; the gate must follow it, resolving the
	// file under the CURRENT mirror directory and nothing else.
	const moved = ".agents/moved"
	repointProducerMirrorRelDir(t, moved)

	root = t.TempDir()
	writeMirrorCopySkill(t, root, template.MirrorSkillsRelDir, "moai-demo")
	res = resolveCodexSkillMirrorPath(root, "", "moai-demo")
	if res.Outcome != codexSkillResolved {
		t.Fatalf("run 2 (producer repointed to %q): outcome = %v, want resolved — the gate did not follow the producer (reason: %s)",
			moved, res.Outcome, res.Reason)
	}
	if want := filepath.Join(root, moved, "moai-demo", "SKILL.md"); res.Path != want {
		t.Fatalf("run 2 resolved %q, want the repointed location %q", res.Path, want)
	}
}

// TestMirrorAbsentReasonNamesProducerRelDir requires the mirror-absent
// resolution to name the producer's CURRENT mirror directory. Under a
// restated literal the reason names the default path while the producer says
// otherwise, which sends the user looking at a directory the deployer no
// longer writes.
func TestMirrorAbsentReasonNamesProducerRelDir(t *testing.T) {
	repointProducerMirrorRelDir(t, filepath.Join(".agents", "moved"))

	root := t.TempDir() // no mirror anywhere on the tree
	res := resolveCodexSkillMirrorPath(root, "", "moai-demo")
	if res.Outcome != codexSkillMirrorAbsent {
		t.Fatalf("outcome = %v, want mirror-absent on a tree without a mirror (reason: %s)",
			res.Outcome, res.Reason)
	}
	if !strings.Contains(res.Reason, template.MirrorSkillsRelDir) {
		t.Fatalf("absent reason does not name the producer's mirror dir %q: %s",
			template.MirrorSkillsRelDir, res.Reason)
	}
}
