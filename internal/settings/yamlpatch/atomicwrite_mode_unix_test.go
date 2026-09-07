//go:build !windows

package yamlpatch

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

// TestAtomicWriteAbsentModeUmaskIndependent는 absent 대상 생성이 umask와
// 무관하게 문서화된 0644를 남김을 검증한다 (SPEC-SEAM-GREENFIELD-001 REQ-2 +
// REQ-4). temp+rename 경로는 Chmod로 모드를 고정하므로 umask가 결과에
// 스며들지 않는다 — absent 브랜치가 직접 os.WriteFile(비원자 쓰기)로 대체되는
// 뮤턴트는 umask 0077에서 0600을 남겨 이 가드에 잡힌다 (M3 뮤턴트 B의 판별
// 증거). umask는 프로세스 전역이므로 t.Parallel()을 쓰지 않는다 — 비병렬
// 테스트는 서로 직렬로 돌고, 병렬 테스트들은 모든 비병렬 테스트가 끝난 뒤에
// 재개되므로 umask 창이 다른 테스트와 겹치지 않는다.
func TestAtomicWriteAbsentModeUmaskIndependent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unreachable on windows: POSIX umask has no analogue there")
	}
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	path := filepath.Join(t.TempDir(), "section.yaml")
	if err := atomicWrite(path, []byte("a: 1\n")); err != nil {
		t.Fatalf("atomicWrite on absent target: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Errorf("absent-target mode = %o under umask 0077, want 644 (temp+rename+chmod fixes the mode; a direct os.WriteFile would leak the umask into 0600)", got)
	}
}
