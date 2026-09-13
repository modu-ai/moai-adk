package cli

// doctor_codex_mirror_drift_test.go — the drift-following guard (t521).
//
// doctor's mirror check reads the producer's exported layout symbols
// (template.MirrorSkillsRelDir / template.CanonicalSkillsRelDir). A reader
// that restated the path as its own literal passes every fixture pinned to
// the default location, so these tests REPOINT the producer and require the
// reader to follow: under a hardcoded reader both tests fail, the first with
// the exact false "mirror absent" warning this card removes (t498 audit D8).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// repointProducerMirrorRelDir swaps template.MirrorSkillsRelDir for a
// non-default location and restores it on cleanup. The caller must not be
// parallel: the producer variable is package-global state.
func repointProducerMirrorRelDir(t *testing.T, rel string) {
	t.Helper()
	orig := template.MirrorSkillsRelDir
	template.MirrorSkillsRelDir = rel
	t.Cleanup(func() { template.MirrorSkillsRelDir = orig })
}

// writeCanonicalSkill creates .claude/skills/<name>/SKILL.md under root.
func writeCanonicalSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, template.CanonicalSkillsRelDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeMirrorLink creates the mirror entry <mirrorRel>/<name> as a symlink
// whose body is the producer's own MirrorLinkTarget, exactly as the deployer
// writes it.
func writeMirrorLink(t *testing.T, root, mirrorRel, name string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
	dir := filepath.Join(root, mirrorRel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(template.MirrorLinkTarget(name), filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
}

// TestInspectSkillMirrorFollowsProducerMirrorRelDir repoints the producer's
// mirror directory and requires the doctor-side reader to look there. This is
// the discriminating guard: a doctor that restated ".agents/skills" as a
// literal reads the repointed layout as absent and fails here.
func TestInspectSkillMirrorFollowsProducerMirrorRelDir(t *testing.T) {
	const moved = ".agents/moved"
	repointProducerMirrorRelDir(t, moved)

	root := t.TempDir()
	writeCanonicalSkill(t, root, "moai-demo")
	writeMirrorLink(t, root, template.MirrorSkillsRelDir, "moai-demo")

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("doctor did not follow the producer's mirror dir %q: mirror read as absent", moved)
	}
	if len(st.dangling) != 0 {
		t.Fatalf("dangling = %v, want none", st.dangling)
	}
	if st.unmirrored != 0 {
		t.Fatalf("unmirrored = %d, want 0", st.unmirrored)
	}
	problems, _ := codexMirrorObservations(st)
	for _, p := range problems {
		if strings.Contains(p.summary, "absent") {
			t.Fatalf("false absent finding raised against the producer's location %q: %s", moved, p.summary)
		}
	}
}

// TestMirrorAbsentFindingNamesProducerRelDir requires the absent finding to
// name the producer's CURRENT mirror directory, not a restated literal. Under
// a hardcoded ".agents/skills" reader the summary names the default path
// while the producer says ".agents/moved", and the test fails.
func TestMirrorAbsentFindingNamesProducerRelDir(t *testing.T) {
	repointProducerMirrorRelDir(t, filepath.Join(".agents", "moved"))

	root := t.TempDir() // no mirror anywhere on the tree
	problems, _ := codexMirrorObservations(inspectSkillMirror(root))
	if len(problems) == 0 {
		t.Fatal("no absent finding raised on a tree without a mirror")
	}
	for _, p := range problems {
		if !strings.Contains(p.summary, template.MirrorSkillsRelDir) {
			t.Fatalf("absent finding does not name the producer's mirror dir %q: %s",
				template.MirrorSkillsRelDir, p.summary)
		}
	}
}
