//go:build !windows

package yamlpatch

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPatchFileWithModeFixesTheModeOnCreateAndRewrite는 perm 인자가 새로 만든 파일과
// 재기록한 파일 모두에 적용됨을 검증한다. 0644 파일을 재기록해도 0600 이 된다. 같은
// 재기록을 PatchFile 로 하면 기존 모드 0644 가 보존된다.
func TestPatchFileWithModeFixesTheModeOnCreateAndRewrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	edits := []KeyEdit{{Path: []string{"a"}, Value: "1"}}

	created := filepath.Join(dir, "created.yaml")
	if err := PatchFileWithMode(created, edits, 0o600); err != nil {
		t.Fatalf("PatchFileWithMode create: %v", err)
	}
	assertFileMode(t, created, 0o600)

	rewritten := filepath.Join(dir, "rewritten.yaml")
	if err := os.WriteFile(rewritten, []byte("a: 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := PatchFileWithMode(rewritten, []KeyEdit{{Path: []string{"a"}, Value: "2"}}, 0o600); err != nil {
		t.Fatalf("PatchFileWithMode rewrite: %v", err)
	}
	assertFileMode(t, rewritten, 0o600)
}

// assertFileMode fails when the file's permission bits differ from want.
func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode = %v, want %v", filepath.Base(path), got, want)
	}
}
