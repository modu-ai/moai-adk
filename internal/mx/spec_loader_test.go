package mx

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadSpecModules_StringFormat는 module 필드가 쉼표 구분 문자열일 때 올바르게 파싱하는지 검증합니다.
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

// TestLoadSpecModules_ArrayFormat는 module 필드가 YAML 시퀀스일 때 올바르게 파싱하는지 검증합니다.
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

// TestLoadSpecModules_EmptyModule는 module 필드가 빈 문자열일 때 빈 슬라이스를 반환하는지 검증합니다.
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

// TestLoadSpecModules_NoSpecsDir는 .moai/specs/ 디렉터리가 없을 때 빈 맵을 반환하는지 검증합니다.
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

// TestSpecAssociator_PathBased_FromLoader는 LoadSpecModules 결과를 SpecAssociator에 주입하면
// path-prefix 매칭이 올바르게 동작하는지 검증합니다.
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
