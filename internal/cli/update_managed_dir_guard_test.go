// SPEC-UPDATE-MIGRATION-FIX-001 M2-b (K2, REQ-UMF-005, AC-UMF-004): the
// project-side regression guard for the empty managed skill/agent directory
// class reported from the mo.ai.kr production run.
//
// Healthy state on this tree: the project payload carries NO common skill or
// agent file in any mode (isCommonAssetRoot, internal/template/deployer_mode.go),
// so a full sync creates no directory under the four managed roots. The swept
// count is therefore ZERO by design. A zero is only evidence when the sync is
// proven to have run, so the guard first asserts the sync's own output exists.
package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// managedSkillAgentRoots are the four project-side roots of the managed-surface
// contract. Each is a common-asset root excluded from the project payload.
var managedSkillAgentRoots = []string{
	filepath.Join(".claude", "skills"),
	filepath.Join(".claude", "agents", "moai"),
	filepath.Join(".agents", "skills"),
	filepath.Join(".codex", "agents", "moai"),
}

// errFoundRegularFile stops a walk early once one regular file is seen.
var errFoundRegularFile = errors.New("found regular file")

// dirHoldsRegularFile reports whether dir carries at least one regular file,
// searched recursively.
func dirHoldsRegularFile(dir string) bool {
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !d.IsDir() {
			return errFoundRegularFile
		}
		return nil
	})
	return errors.Is(err, errFoundRegularFile)
}

// sweepManagedSkillAgentDirs lists every directory strictly below each managed
// root under projectRoot. swept holds all of them; zeroFile holds the subset
// that carries no regular file, recursively. An absent root contributes nothing.
func sweepManagedSkillAgentDirs(projectRoot string) (swept, zeroFile []string) {
	for _, rel := range managedSkillAgentRoots {
		base := filepath.Join(projectRoot, rel)
		_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil || p == base || !d.IsDir() {
				return nil
			}
			swept = append(swept, p)
			if !dirHoldsRegularFile(p) {
				zeroFile = append(zeroFile, p)
			}
			return nil
		})
	}
	return swept, zeroFile
}

// TestTemplateSync_LeavesNoEmptyManagedSkillDirs runs a full v3 template sync and
// asserts the exclusion contract: no managed skill/agent directory is created,
// so no zero-file directory can be left behind.
func TestTemplateSync_LeavesNoEmptyManagedSkillDirs(t *testing.T) {
	root := t.TempDir()
	writeV3ProjectFixture(t, root)
	_ = runUpdateInFixture(t, root)

	// Anti-vacuity: a zero sweep proves nothing unless the sync actually ran.
	if _, err := os.Stat(filepath.Join(root, ".claude", "settings.json")); err != nil {
		t.Fatalf("sync did not run on the fixture (no .claude/settings.json), so a zero sweep would be vacuous: %v", err)
	}

	swept, zeroFile := sweepManagedSkillAgentDirs(root)
	t.Logf("managed skill/agent directories swept=%d zero_file=%d (healthy exclusion contract: swept=0)",
		len(swept), len(zeroFile))
	for _, d := range zeroFile {
		t.Errorf("zero-file managed directory left by the sync: %s", d)
	}
	if len(swept) != 0 {
		t.Errorf("exclusion contract violated: the project payload must carry no managed skill/agent directory; swept %d: %v",
			len(swept), swept)
	}
}

// TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir exercises the sweep's failure
// arm on a known failing input: a planted zero-file managed directory must be
// reported, and a directory holding a file must not be.
func TestTemplateSync_ManagedSweepFlagsPlantedEmptyDir(t *testing.T) {
	root := t.TempDir()
	planted := filepath.Join(root, ".claude", "skills", "moai-planted-empty")
	if err := os.MkdirAll(planted, 0o755); err != nil {
		t.Fatalf("plant empty directory: %v", err)
	}
	filled := filepath.Join(root, ".agents", "skills", "moai-planted-filled")
	if err := os.MkdirAll(filled, 0o755); err != nil {
		t.Fatalf("plant filled directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filled, "SKILL.md"), []byte("planted\n"), 0o644); err != nil {
		t.Fatalf("plant file: %v", err)
	}

	swept, zeroFile := sweepManagedSkillAgentDirs(root)
	if len(swept) != 2 {
		t.Fatalf("sweep must visit both planted directories, swept %d: %v", len(swept), swept)
	}
	if len(zeroFile) != 1 || zeroFile[0] != planted {
		t.Fatalf("sweep must flag exactly the planted empty directory %s, got zero_file=%v", planted, zeroFile)
	}
}
