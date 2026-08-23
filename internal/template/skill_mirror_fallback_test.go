package template

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

// threeSkillFS is a small synthetic catalog used by the mirror behaviour tests.
func threeSkillFS() fstest.MapFS {
	return fstest.MapFS{
		".claude/skills/moai-alpha/SKILL.md":        &fstest.MapFile{Data: []byte("alpha")},
		".claude/skills/moai-beta/SKILL.md":         &fstest.MapFile{Data: []byte("beta")},
		".claude/skills/moai-beta/modules/extra.md": &fstest.MapFile{Data: []byte("beta extra")},
		".claude/skills/moai-gamma/SKILL.md":        &fstest.MapFile{Data: []byte("gamma")},
		".claude/rules/moai/unrelated.md":           &fstest.MapFile{Data: []byte("rule")},
		".moai/config/sections/quality.yaml":        &fstest.MapFile{Data: []byte("quality: {}")},
	}
}

func deployWithResult(t *testing.T, fsys fs.FS, opts ...DeployerOption) (string, *DeployResult) {
	t.Helper()
	root, mgr := setupDeployProject(t)
	dep, ok := NewDeployer(fsys, opts...).(ResultDeployer)
	if !ok {
		t.Fatal("deployer does not implement ResultDeployer")
	}
	res, err := dep.DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("DeployWithResult: %v", err)
	}
	return root, res
}

var errInjectedSymlink = errors.New("injected: symlink unavailable")
var errInjectedCopy = errors.New("injected: copy unavailable")

func failingSymlink() DeployerOption {
	return withSymlinkFunc(func(_, _ string) error { return errInjectedSymlink })
}

func failingCopy() DeployerOption {
	return withMirrorCopyFunc(func(_, _ string) error { return errInjectedCopy })
}

// claudeTreeFingerprint lists (relative path, sha256, permission) for every
// file under .claude/ — the comparison subject for the no-regression
// invariant. Both sides are produced in this same process, by this same
// binary; a "baseline from the previous commit" is not something a Go test can
// produce, and would stop being re-runnable the moment the change landed.
func claudeTreeFingerprint(t *testing.T, root string) []string {
	t.Helper()
	base := filepath.Join(root, ".claude")
	var lines []string
	err := filepath.WalkDir(base, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(base, p)
		if relErr != nil {
			return relErr
		}
		content, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		info, statErr := entry.Info()
		if statErr != nil {
			return statErr
		}
		sum := sha256.Sum256(content)
		lines = append(lines, fmt.Sprintf("%s %s %04o", filepath.ToSlash(rel), hex.EncodeToString(sum[:]), info.Mode().Perm()))
		return nil
	})
	if err != nil {
		t.Fatalf("fingerprint %q: %v", base, err)
	}
	sort.Strings(lines)
	return lines
}

func diffFingerprints(t *testing.T, want, got []string) {
	t.Helper()
	if sameStringSlice(want, got) {
		return
	}
	t.Errorf(".claude tree differs: %d entries without mirror, %d with", len(want), len(got))
	wantSet := map[string]struct{}{}
	for _, l := range want {
		wantSet[l] = struct{}{}
	}
	for _, l := range got {
		if _, ok := wantSet[l]; !ok {
			t.Errorf("  unexpected/changed entry: %s", l)
		}
	}
	gotSet := map[string]struct{}{}
	for _, l := range got {
		gotSet[l] = struct{}{}
	}
	for _, l := range want {
		if _, ok := gotSet[l]; !ok {
			t.Errorf("  missing entry: %s", l)
		}
	}
}

// TestSkillMirror_LinkIsRelativeAndResolves covers AC-CSC-004 (SHOULD): the
// link body is relative and resolves to the canonical directory. On a platform
// where symlink creation is unavailable the criterion does not apply.
func TestSkillMirror_LinkIsRelativeAndResolves(t *testing.T) {
	root, res := deployWithResult(t, threeSkillFS())

	for _, skill := range []string{"moai-alpha", "moai-beta", "moai-gamma"} {
		mode, ok := res.MirrorMode(skill)
		if !ok {
			t.Fatalf("no mirror entry recorded for %q", skill)
		}
		if mode != MirrorModeSymlink {
			t.Skipf("symlink mode unavailable on this platform (mode=%s) — AC-CSC-004 not applicable", mode)
		}
		mirrorPath := filepath.Join(root, ".agents", "skills", skill)
		target, err := os.Readlink(mirrorPath)
		if err != nil {
			t.Fatalf("readlink %q: %v", mirrorPath, err)
		}
		want := "../../.claude/skills/" + skill
		if target != want {
			t.Errorf("link body for %q = %q, want %q (an absolute path breaks when the project is moved)", skill, target, want)
		}
		resolved, err := filepath.EvalSymlinks(mirrorPath)
		if err != nil {
			t.Fatalf("eval symlinks %q: %v", mirrorPath, err)
		}
		wantResolved, err := filepath.EvalSymlinks(filepath.Join(root, ".claude", "skills", skill))
		if err != nil {
			t.Fatalf("eval canonical: %v", err)
		}
		if resolved != wantResolved {
			t.Errorf("link for %q resolves to %q, want %q", skill, resolved, wantResolved)
		}
	}
}

// TestSkillMirror_CopyFallbackIsReadable covers AC-CSC-005.
func TestSkillMirror_CopyFallbackIsReadable(t *testing.T) {
	root, res := deployWithResult(t, threeSkillFS(), failingSymlink())

	for _, skill := range []string{"moai-alpha", "moai-beta", "moai-gamma"} {
		if mode, _ := res.MirrorMode(skill); mode != MirrorModeCopy {
			t.Errorf("mirror mode for %q = %q, want %q", skill, mode, MirrorModeCopy)
		}
		mirrorFile := filepath.Join(root, ".agents", "skills", skill, "SKILL.md")
		info, err := os.Lstat(mirrorFile)
		if err != nil {
			t.Errorf("lstat %q: %v", mirrorFile, err)
			continue
		}
		if !info.Mode().IsRegular() {
			t.Errorf("%q is not a regular file (mode %v) — the fallback must materialize real files", mirrorFile, info.Mode())
		}
		got, err := os.ReadFile(mirrorFile)
		if err != nil {
			t.Errorf("read %q: %v", mirrorFile, err)
			continue
		}
		want, err := os.ReadFile(filepath.Join(root, ".claude", "skills", skill, "SKILL.md"))
		if err != nil {
			t.Fatalf("read canonical: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("copied mirror content for %q differs from canonical", skill)
		}
	}
	// Nested files come along too.
	nested := filepath.Join(root, ".agents", "skills", "moai-beta", "modules", "extra.md")
	if _, err := os.ReadFile(nested); err != nil {
		t.Errorf("nested copied file %q unreadable: %v", nested, err)
	}
}

// TestSkillMirror_FallbackIsObservable covers AC-CSC-006 in both directions:
// the copy fallback shows up in the returned result, and a run where linking
// succeeded does not claim a fallback.
func TestSkillMirror_FallbackIsObservable(t *testing.T) {
	t.Run("fallback_reported", func(t *testing.T) {
		_, res := deployWithResult(t, threeSkillFS(), failingSymlink())
		if !res.CopyFallbackUsed() {
			t.Error("CopyFallbackUsed() = false after an injected symlink failure — the fallback is silent")
		}
		warnings := res.Warnings()
		if len(warnings) == 0 {
			t.Fatal("no warning recorded for the copy fallback")
		}
		if !strings.Contains(strings.Join(warnings, "\n"), "copied instead of linked") {
			t.Errorf("warnings do not explain the fallback: %v", warnings)
		}
	})

	t.Run("no_false_report_when_linked", func(t *testing.T) {
		_, res := deployWithResult(t, threeSkillFS())
		if mode, _ := res.MirrorMode("moai-alpha"); mode != MirrorModeSymlink {
			t.Skipf("symlink mode unavailable on this platform (mode=%s)", mode)
		}
		if res.CopyFallbackUsed() {
			t.Error("CopyFallbackUsed() = true on a run where every link succeeded")
		}
		if w := res.Warnings(); len(w) != 0 {
			t.Errorf("unexpected warnings on a clean run: %v", w)
		}
	})
}

// TestSkillMirror_ClaudePathUnchangedByMirror covers AC-CSC-010: the .claude/
// output is identical whether the mirror feature runs or not.
func TestSkillMirror_ClaudePathUnchangedByMirror(t *testing.T) {
	offRoot, _ := deployWithResult(t, threeSkillFS(), WithSkillMirror(false))
	onRoot, _ := deployWithResult(t, threeSkillFS())

	diffFingerprints(t, claudeTreeFingerprint(t, offRoot), claudeTreeFingerprint(t, onRoot))

	// The disabled seam must genuinely produce no mirror, or the comparison
	// above would be comparing the feature against itself.
	if names := mirrorEntryNames(t, offRoot); len(names) != 0 {
		t.Errorf("mirror-disabled deploy produced %v under .agents/skills", names)
	}
}

// TestSkillMirror_PreoccupiedTargets covers AC-CSC-011: the three states a
// mirror path can already be in on a re-deploy.
func TestSkillMirror_PreoccupiedTargets(t *testing.T) {
	root, mgr := setupDeployProject(t)
	dep := NewDeployer(threeSkillFS()).(ResultDeployer)
	first, err := dep.DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("first Deploy: %v", err)
	}
	if mode, _ := first.MirrorMode("moai-alpha"); mode != MirrorModeSymlink {
		t.Skipf("symlink mode unavailable on this platform (mode=%s) — the three states are symlink-shaped", mode)
	}

	agentsSkills := filepath.Join(root, ".agents", "skills")
	// (i) moai-alpha: left exactly as deployed.
	// (ii) moai-beta: repointed somewhere else.
	betaMirror := filepath.Join(agentsSkills, "moai-beta")
	if err := os.Remove(betaMirror); err != nil {
		t.Fatalf("remove beta link: %v", err)
	}
	if err := os.Symlink("../../elsewhere", betaMirror); err != nil {
		t.Fatalf("create stale beta link: %v", err)
	}
	// (iii) moai-gamma: replaced by a real user directory holding USER.md.
	gammaMirror := filepath.Join(agentsSkills, "moai-gamma")
	if err := os.Remove(gammaMirror); err != nil {
		t.Fatalf("remove gamma link: %v", err)
	}
	if err := os.MkdirAll(gammaMirror, 0o755); err != nil {
		t.Fatalf("mkdir gamma: %v", err)
	}
	userFile := filepath.Join(gammaMirror, "USER.md")
	if err := os.WriteFile(userFile, []byte("user content"), 0o644); err != nil {
		t.Fatalf("write USER.md: %v", err)
	}

	second, err := dep.DeployWithResult(context.Background(), root, mgr, nil)
	if err != nil {
		t.Fatalf("second Deploy returned an error: %v", err)
	}

	// 1. (i) untouched.
	alphaTarget, err := os.Readlink(filepath.Join(agentsSkills, "moai-alpha"))
	if err != nil {
		t.Fatalf("readlink alpha: %v", err)
	}
	if alphaTarget != "../../.claude/skills/moai-alpha" {
		t.Errorf("alpha link body changed: %q", alphaTarget)
	}

	// 2. (ii) repointed back at the canonical directory.
	betaTarget, err := os.Readlink(betaMirror)
	if err != nil {
		t.Fatalf("readlink beta: %v", err)
	}
	if betaTarget != "../../.claude/skills/moai-beta" {
		t.Errorf("stale beta link was not replaced: %q", betaTarget)
	}

	// 3. (iii) the user's real directory survives, and the skip is reported.
	got, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("USER.md was destroyed by the re-deploy: %v", err)
	}
	if string(got) != "user content" {
		t.Errorf("USER.md content changed: %q", got)
	}
	if mode, _ := second.MirrorMode("moai-gamma"); mode != MirrorModeSkipped {
		t.Errorf("mirror mode for the pre-occupied path = %q, want %q", mode, MirrorModeSkipped)
	}
	joined := strings.Join(second.Warnings(), "\n")
	if !strings.Contains(joined, "moai-gamma") {
		t.Errorf("no warning naming the skipped skill in the returned result: %v", second.Warnings())
	}
}

// TestSkillMirror_FailOpen covers AC-CSC-013: with both symlink and copy
// failing the deployment still succeeds, the .claude/ output is unaffected, and
// the failure is reported in the returned result.
func TestSkillMirror_FailOpen(t *testing.T) {
	root, mgr := setupDeployProject(t)
	dep := NewDeployer(threeSkillFS(), failingSymlink(), failingCopy()).(ResultDeployer)
	res, err := dep.DeployWithResult(context.Background(), root, mgr, nil)

	// 1. no error.
	if err != nil {
		t.Fatalf("Deploy failed because mirroring failed: %v", err)
	}

	// 2. .claude/ output matches the mirror-disabled seam's output.
	offRoot, _ := deployWithResult(t, threeSkillFS(), WithSkillMirror(false))
	diffFingerprints(t, claudeTreeFingerprint(t, offRoot), claudeTreeFingerprint(t, root))

	// 3. the failure is in the returned result.
	if mode, _ := res.MirrorMode("moai-alpha"); mode != MirrorModeFailed {
		t.Errorf("mirror mode = %q, want %q", mode, MirrorModeFailed)
	}
	joined := strings.Join(res.Warnings(), "\n")
	if !strings.Contains(joined, "symlink and copy both failed") {
		t.Errorf("failure not reported in the result: %v", res.Warnings())
	}
}
