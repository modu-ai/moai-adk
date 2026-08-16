package mx

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadSpecModules_StringFormat verifies correct parsing when the module field is a comma-separated string.
func TestLoadSpecModules_StringFormat(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-X")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-X\nmodule: \"internal/mx/, cmd/moai/\"\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSpecModules(dir)
	if err != nil {
		t.Fatalf("LoadSpecModules 에러: %v", err)
	}

	modules, ok := got["SPEC-X"]
	if !ok {
		t.Fatalf("SPEC-X 키 없음 (got=%v)", got)
	}

	want := []string{"internal/mx/", "cmd/moai/"}
	if len(modules) != len(want) {
		t.Fatalf("모듈 수: 기대 %d, 실제 %d (got=%v)", len(want), len(modules), modules)
	}
	for i, w := range want {
		if modules[i] != w {
			t.Errorf("modules[%d]: 기대 %q, 실제 %q", i, w, modules[i])
		}
	}
}

// TestLoadSpecModules_ArrayFormat verifies correct parsing when the module field is a YAML sequence.
func TestLoadSpecModules_ArrayFormat(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-X")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-X\nmodule:\n  - internal/foo/\n  - internal/bar/\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSpecModules(dir)
	if err != nil {
		t.Fatalf("LoadSpecModules 에러: %v", err)
	}

	modules, ok := got["SPEC-X"]
	if !ok {
		t.Fatalf("SPEC-X 키 없음 (got=%v)", got)
	}

	want := []string{"internal/foo/", "internal/bar/"}
	if len(modules) != len(want) {
		t.Fatalf("모듈 수: 기대 %d, 실제 %d (got=%v)", len(want), len(modules), modules)
	}
	for i, w := range want {
		if modules[i] != w {
			t.Errorf("modules[%d]: 기대 %q, 실제 %q", i, w, modules[i])
		}
	}
}

// TestLoadSpecModules_EmptyModule verifies that an empty module field returns an empty slice.
func TestLoadSpecModules_EmptyModule(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-X")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-X\nmodule: \"\"\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSpecModules(dir)
	if err != nil {
		t.Fatalf("LoadSpecModules 에러: %v", err)
	}

	modules, ok := got["SPEC-X"]
	if !ok {
		t.Fatalf("SPEC-X 키 없음 (got=%v)", got)
	}

	if len(modules) != 0 {
		t.Errorf("빈 module 필드: 슬라이스 길이 기대 0, 실제 %d (got=%v)", len(modules), modules)
	}
}

// TestLoadSpecModules_NoSpecsDir verifies that an empty map is returned when .moai/specs/ is absent.
func TestLoadSpecModules_NoSpecsDir(t *testing.T) {
	dir := t.TempDir()

	got, err := LoadSpecModules(dir)
	if err != nil {
		t.Fatalf("LoadSpecModules 에러 (specs 없음): %v", err)
	}

	if len(got) != 0 {
		t.Errorf("specs 디렉터리 없음: 빈 맵 기대, 실제 %v", got)
	}
}

// TestSpecAssociator_PathBased_FromLoader verifies that injecting LoadSpecModules
// results into SpecAssociator yields correct path-prefix matching.
func TestSpecAssociator_PathBased_FromLoader(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-X")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-X\nmodule: \"internal/mx/, cmd/moai/\"\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	specModules, err := LoadSpecModules(dir)
	if err != nil {
		t.Fatalf("LoadSpecModules 에러: %v", err)
	}

	associator := NewSpecAssociator(specModules)

	tag := Tag{
		Kind:     MXAnchor,
		File:     "internal/mx/scanner.go",
		Line:     10,
		Body:     "scanner anchor",
		AnchorID: "anchor-scanner",
	}

	specs := associator.Associate(tag)

	found := false
	for _, s := range specs {
		if s == "SPEC-X" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("path-based 연결 실패: SPEC-X 없음 (got=%v, file=%s)", specs, tag.File)
	}
}

// TestLoadSpecDependencies_FlowSequence verifies depends_on parsing when the
// frontmatter carries a YAML flow sequence.
func TestLoadSpecDependencies_FlowSequence(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-DEP-X")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-DEP-X\ntitle: \"t\"\ndepends_on: [SPEC-DEP-A, SPEC-DEP-B]\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSpecDependencies(dir)
	if err != nil {
		t.Fatalf("LoadSpecDependencies error: %v", err)
	}
	deps, ok := got["SPEC-DEP-X"]
	if !ok {
		t.Fatalf("SPEC-DEP-X key missing (got=%v)", got)
	}
	if len(deps) != 2 || deps[0] != "SPEC-DEP-A" || deps[1] != "SPEC-DEP-B" {
		t.Errorf("depends_on=%v; want [SPEC-DEP-A SPEC-DEP-B]", deps)
	}
}

// TestLoadSpecDependencies_AbsentField verifies a spec without depends_on
// yields an empty (non-nil) slice, not a missing key.
func TestLoadSpecDependencies_AbsentField(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, ".moai", "specs", "SPEC-DEP-PLAIN")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "---\nid: SPEC-DEP-PLAIN\ntitle: \"t\"\n---\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSpecDependencies(dir)
	if err != nil {
		t.Fatalf("LoadSpecDependencies error: %v", err)
	}
	deps, ok := got["SPEC-DEP-PLAIN"]
	if !ok {
		t.Fatalf("SPEC-DEP-PLAIN key missing (got=%v)", got)
	}
	if len(deps) != 0 {
		t.Errorf("depends_on=%v; want empty", deps)
	}
}

// TestLoadSpecDependencies_NoSpecsDir verifies the fail-open contract: a
// project without .moai/specs/ returns an empty map without error.
func TestLoadSpecDependencies_NoSpecsDir(t *testing.T) {
	got, err := LoadSpecDependencies(t.TempDir())
	if err != nil {
		t.Fatalf("LoadSpecDependencies error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want empty map, got %v", got)
	}
}
