package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// M4 — Provenance correctness tests. These are the load-bearing correctness
// assertions for REQ-TBS-010 (template-blessed key adopted, not misread as a
// user edit) and REQ-TBS-012 (quality.yaml real-3-way correctness).
//
// IMPORTANT — why these tests use NON-system fields (test_coverage_target,
// enforce_quality), NOT `version`: MergeYAML3Way treats `version` /
// `template_version` as SYSTEM FIELDS (node_merge.go systemFields) that always
// take the NEW template value regardless of base. So a provenance test on
// `version` is VACUOUS — it passes against both the correct (snapshot) base
// and the wrong (embedded-raw) base. The provenance fix only matters for
// NON-system fields where the old==base vs old!=base decision distinguishes
// "adopt new" from "preserve old". quality.yaml's test_coverage_target /
// enforce_quality are the canonical non-system placeholder keys (research.md
// §D), so they are the correct falsifiability targets.

// plantSnapshot plants a rendered snapshot at
// <projectRoot>/.moai/cache/template-snapshot/sections/<name> with the given
// bytes. This mirrors what WriteSnapshot produces, but is self-contained so
// the falsifiability test (M5) can run against a tree where snapshot.go /
// base_loader.go are stashed.
func plantSnapshot(t *testing.T, projectRoot, sectionName string, sectionBytes []byte) {
	t.Helper()
	snapDir := filepath.Join(projectRoot, defs.MoAIDir, "cache", "template-snapshot", "sections")
	if err := os.MkdirAll(snapDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir snap sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, sectionName), sectionBytes, defs.FilePerm); err != nil {
		t.Fatalf("write snapshot %s: %v", sectionName, err)
	}
}

// assertTopLevelScalar extracts a top-level scalar value for a key from a YAML
// document and asserts it equals want.
func assertTopLevelScalar(t *testing.T, doc []byte, key, want string, msg string) {
	t.Helper()
	root := &yaml.Node{}
	if err := yaml.Unmarshal(doc, root); err != nil {
		t.Fatalf("%s: unmarshal merged doc: %v", msg, err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		t.Fatalf("%s: expected document node", msg)
	}
	m := root.Content[0]
	if m.Kind != yaml.MappingNode {
		t.Fatalf("%s: expected mapping at root", msg)
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			got := m.Content[i+1].Value
			if got != want {
				t.Errorf("%s: key %q = %q, want %q", msg, key, got, want)
			}
			return
		}
	}
	t.Errorf("%s: key %q not found in merged output", msg, key)
}

// TestMerge_BaseFromSnapshot_NotFromEmbedded is the LOAD-BEARING provenance
// test (AC-TBS-011 / REQ-TBS-010) AND the AC-TBS-010 falsifiability target.
//
// It uses BackupMoaiConfig as the ONLY entry point (no references to
// SaveTemplateBase / WriteSnapshot / SnapshotDir / HasSnapshot), so the M5
// stash procedure can revert ONLY the backup.go edit
// (SaveTemplateBase → SaveTemplateDefaults) and run this test against the
// wrong-base implementation.
//
// Setup:
//   - on-disk rendered section: test_coverage_target: 80
//   - snapshot (rendered): test_coverage_target: 80 (matches on-disk)
//   - OLD (user): test_coverage_target: 80 (matches snapshot)
//   - NEW template: test_coverage_target: 85
//
// With the CORRECT (snapshot) base: old == base → adopt NEW (85).
// With the WRONG (embedded-raw) base: base = "{{.TestCoverageTarget}}" (a
// nested flow-mapping node) → old != base → preserve OLD (80, stale).
//
// The test asserts 85. It FAILS against the wrong-base implementation.
func TestMerge_BaseFromSnapshot_NotFromEmbedded(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()

	// 1. Plant rendered on-disk section (what BackupMoaiConfig backs up).
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"),
		[]byte("test_coverage_target: 80\n"), defs.FilePerm); err != nil {
		t.Fatalf("write on-disk quality.yaml: %v", err)
	}

	// 2. Plant a rendered snapshot matching the on-disk state (prior install).
	plantSnapshot(t, projectRoot, "quality.yaml", []byte("test_coverage_target: 80\n"))

	// 3. BackupMoaiConfig populates .template-defaults/sections/ via either
	//    SaveTemplateBase (correct: snapshot bytes) or SaveTemplateDefaults
	//    (wrong: embedded-raw bytes). This is the ONLY call site the M5
	//    procedure reverts.
	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}

	// 4. Read the BASE bytes (the production read path at restore.go:118).
	baseBytes, err := os.ReadFile(filepath.Join(backupDir, ".template-defaults", "sections", "quality.yaml"))
	if err != nil {
		t.Fatalf("read base quality.yaml: %v", err)
	}

	// 5. Run the 3-way merge with OLD (user) == snapshot, NEW template = 85.
	oldBytes := []byte("test_coverage_target: 80\n")
	newBytes := []byte("test_coverage_target: 85\n")
	merged, err := MergeYAML3Way(newBytes, oldBytes, baseBytes)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}

	// 6. Assert the NEW value is adopted (old==base). Against the wrong base
	//    (embedded-raw), old!=base and the stale 80 is preserved → FAIL.
	assertTopLevelScalar(t, merged, "test_coverage_target", "85",
		"TestMerge_BaseFromSnapshot_NotFromEmbedded (old==base → adopt NEW)")
}

// TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal is AC-TBS-011:
// when old == base (snapshot matches local), the merge adopts the NEW value.
func TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"),
		[]byte("test_coverage_target: 80\n"), defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}
	plantSnapshot(t, projectRoot, "quality.yaml", []byte("test_coverage_target: 80\n"))

	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	baseBytes, _ := os.ReadFile(filepath.Join(backupDir, ".template-defaults", "sections", "quality.yaml"))

	oldBytes := []byte("test_coverage_target: 80\n")
	newBytes := []byte("test_coverage_target: 85\n")
	merged, err := MergeYAML3Way(newBytes, oldBytes, baseBytes)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	assertTopLevelScalar(t, merged, "test_coverage_target", "85", "adopt NEW when old==base")
}

// TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal is
// AC-TBS-012 (REQ-TBS-011): when old != base (user customized), the merge
// preserves the USER value.
func TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	// On-disk user value = 95 (customized away from template default 80).
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"),
		[]byte("test_coverage_target: 95\n"), defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}
	// Snapshot = template-blessed default 80 (the prior install's value).
	plantSnapshot(t, projectRoot, "quality.yaml", []byte("test_coverage_target: 80\n"))

	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	baseBytes, _ := os.ReadFile(filepath.Join(backupDir, ".template-defaults", "sections", "quality.yaml"))

	oldBytes := []byte("test_coverage_target: 95\n") // user customized
	newBytes := []byte("test_coverage_target: 85\n") // new template default
	merged, err := MergeYAML3Way(newBytes, oldBytes, baseBytes)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	assertTopLevelScalar(t, merged, "test_coverage_target", "95",
		"preserve USER when old!=base")
}

// TestMerge_QualityYaml_Real3Way_WithSnapshot is AC-TBS-013 (REQ-TBS-012):
// quality.yaml merges correctly with a snapshot-sourced rendered BASE. The
// template-blessed key (enforce_quality) is NOT misread as a user edit
// (old == base); the updated default (test_coverage_target) is adopted.
func TestMerge_QualityYaml_Real3Way_WithSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	renderedQuality := []byte("enforce_quality: true\ntest_coverage_target: 80\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "quality.yaml"), renderedQuality, defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}
	// Snapshot = rendered (placeholders resolved), matches on-disk.
	plantSnapshot(t, projectRoot, "quality.yaml", renderedQuality)

	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	baseBytes, _ := os.ReadFile(filepath.Join(backupDir, ".template-defaults", "sections", "quality.yaml"))

	oldBytes := renderedQuality // user did not customize
	newBytes := []byte("enforce_quality: true\ntest_coverage_target: 85\n")
	merged, err := MergeYAML3Way(newBytes, oldBytes, baseBytes)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	assertTopLevelScalar(t, merged, "test_coverage_target", "85",
		"adopt NEW default when old==base (quality.yaml)")
	assertTopLevelScalar(t, merged, "enforce_quality", "true",
		"enforce_quality not misread as user edit (old==base)")
}

// TestRestore_2WayFallback_UnparseableBase is AC-TBS-009: when the BASE is
// unparseable, MergeYAML3Way errors and the caller falls through to the 2-way
// MergeYAMLDeep. Uses a NON-system field (foo) so the 2-way value semantics
// are observable.
func TestRestore_2WayFallback_UnparseableBase(t *testing.T) {
	t.Parallel()
	newBytes := []byte("foo: new\n")
	oldBytes := []byte("foo: old\n")
	corruptBase := []byte("foo: [unclosed\n  : : :")

	_, err := MergeYAML3Way(newBytes, oldBytes, corruptBase)
	if err == nil {
		t.Fatalf("MergeYAML3Way must error on an unparseable base")
	}

	// The 2-way fallback succeeds and preserves the old (user) value for a
	// non-system key.
	merged, err := MergeYAMLDeep(newBytes, oldBytes)
	if err != nil {
		t.Fatalf("MergeYAMLDeep (2-way fallback): %v", err)
	}
	assertTopLevelScalar(t, merged, "foo", "old", "2-way fallback preserves user value")
}

// TestMergeYAML3Way_SignatureUnchanged is AC-TBS-008: the signature is
// literally func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error).
func TestMergeYAML3Way_SignatureUnchanged(t *testing.T) {
	t.Parallel()
	// Compile-time assertion: passing MergeYAML3Way to a parameter of the
	// exact signature type fails to compile if the signature drifted.
	requireSig3Way(MergeYAML3Way)
}

// requireSig3Way enforces the MergeYAML3Way signature at compile time.
func requireSig3Way(f func(newData, oldData, baseData []byte) ([]byte, error)) {
	_ = f
}

// TestBackupMoaiConfig_PrefersSnapshot is AC-TBS-006a: BackupMoaiConfig, when
// a snapshot exists, writes the snapshot bytes (rendered) into the per-backup
// .template-defaults/sections/, NOT the embedded-raw bytes.
func TestBackupMoaiConfig_PrefersSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}
	plantSnapshot(t, projectRoot, "system.yaml", []byte("version: \"3.0.1\"\n"))

	backupDir, err := BackupMoaiConfig(projectRoot)
	if err != nil {
		t.Fatalf("BackupMoaiConfig: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(backupDir, ".template-defaults", "sections", "system.yaml"))
	if err != nil {
		t.Fatalf("read backup base: %v", err)
	}
	want := "version: \"3.0.1\"\n"
	if string(got) != want {
		t.Errorf("BackupMoaiConfig base = %q, want snapshot bytes %q", got, want)
	}
}
