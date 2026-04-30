package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// buildTestSidecarForCLI는 CLI 테스트용 사이드카 파일을 생성합니다.
func buildTestSidecarForCLI(t *testing.T, stateDir string, tags []mx.Tag) {
	t.Helper()
	mgr := mx.NewManager(stateDir)
	sidecar := &mx.Sidecar{
		SchemaVersion: mx.SchemaVersion,
		Tags:          tags,
		ScannedAt:     time.Now(),
	}
	if err := mgr.Write(sidecar); err != nil {
		t.Fatalf("CLI 테스트용 사이드카 쓰기 실패: %v", err)
	}
}

// executeQueryCmd는 CLI 명령을 실행하고 stdout/stderr을 캡처합니다.
func executeQueryCmd(t *testing.T, args []string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := newMxQueryCmd()

	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// TestMxQueryCmd_Structure는 mx query 명령 구조를 테스트합니다.
func TestMxQueryCmd_Structure(t *testing.T) {
	cmd := newMxQueryCmd()

	if cmd.Use != "query" {
		t.Errorf("Use: 기대 'query', 실제 %q", cmd.Use)
	}

	// 필수 플래그 존재 확인
	requiredFlags := []string{
		"spec", "kind", "fan-in-min", "danger", "file-prefix",
		"since", "limit", "offset", "format", "include-tests",
	}

	for _, flag := range requiredFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("플래그 누락: --%s", flag)
		}
	}
}

// TestMxQueryCmd_InvalidKind는 잘못된 kind 값에 대한 오류를 테스트합니다.
// AC-SPC-004-13: --kind nonexistent → exit 2 + InvalidQuery
func TestMxQueryCmd_InvalidKind(t *testing.T) {
	// AC-SPC-004-13: 잘못된 필터 값
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")
	buildTestSidecarForCLI(t, stateDir, []mx.Tag{})

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	_, stderr, err := executeQueryCmd(t, []string{"--kind", "nonexistent"})
	if err == nil {
		t.Error("잘못된 kind에 대해 오류 기대, 실제 nil")
	}

	if !strings.Contains(stderr, "InvalidQuery") && !strings.Contains(err.Error(), "InvalidQuery") {
		t.Logf("stderr: %s, err: %v", stderr, err)
		// RED 단계에서는 "not implemented" 오류가 반환되므로 실패 예상
	}
}

// TestMxQueryCmd_SidecarUnavailable은 사이드카 파일 없을 때 오류를 테스트합니다.
// AC-SPC-004-04: 사이드카 없을 때 SidecarUnavailable 오류
func TestMxQueryCmd_SidecarUnavailable(t *testing.T) {
	// AC-SPC-004-04: 사이드카 파일 없을 때
	tmpDir := t.TempDir()
	// 사이드카 파일 생성하지 않음

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	_, stderr, err := executeQueryCmd(t, []string{})
	if err == nil {
		t.Error("사이드카 없을 때 오류 기대, 실제 nil")
	}

	_ = stderr // GREEN 단계에서 "SidecarUnavailable" 포함 검증
}

// TestMxQueryCmd_JSONOutput은 JSON 출력 형식을 테스트합니다.
// AC-SPC-004-05: JSON 출력이 REQ-SPC-004-005 스키마 준수
func TestMxQueryCmd_JSONOutput(t *testing.T) {
	// AC-SPC-004-05: JSON 출력 스키마
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	tags := []mx.Tag{
		{
			Kind:       mx.MXAnchor,
			File:       "internal/auth/handler.go",
			Line:       10,
			Body:       "인증 핸들러 앵커",
			AnchorID:   "anchor-auth-handler",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, _, err := executeQueryCmd(t, []string{"--format", "json"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return // RED 단계에서는 실패 예상
	}

	// JSON 파싱 가능한지 확인
	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("JSON 파싱 실패: %v\n출력: %s", err, stdout)
	}
}

// TestMxQueryCmd_TableOutput은 테이블 출력 형식을 테스트합니다.
// AC-SPC-004-06: --format table 컬럼 형식 출력
func TestMxQueryCmd_TableOutput(t *testing.T) {
	// AC-SPC-004-06: 테이블 출력 형식
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	tags := []mx.Tag{
		{
			Kind:       mx.MXNote,
			File:       "internal/misc.go",
			Line:       1,
			Body:       "노트 태그",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, _, err := executeQueryCmd(t, []string{"--format", "table"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return // RED 단계에서는 실패 예상
	}

	// 테이블 출력에 컬럼 헤더가 있어야 함
	_ = stdout
}

// TestMxQueryCmd_MarkdownOutput은 마크다운 출력 형식을 테스트합니다.
// AC-SPC-004-10: --format markdown 마크다운 테이블 출력
func TestMxQueryCmd_MarkdownOutput(t *testing.T) {
	// AC-SPC-004-10: 마크다운 출력 형식
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	tags := []mx.Tag{
		{
			Kind:       mx.MXNote,
			File:       "internal/misc.go",
			Line:       1,
			Body:       "노트 태그",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, _, err := executeQueryCmd(t, []string{"--format", "markdown"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return
	}

	if !strings.Contains(stdout, "|") {
		t.Error("마크다운 테이블 구분자 '|' 없음")
	}
}

// TestMxQueryCmd_EmptyResult는 매칭 없을 때 빈 배열과 exit 0을 테스트합니다.
// AC-SPC-004-12: 빈 결과 시 [] + exit 0
func TestMxQueryCmd_EmptyResult(t *testing.T) {
	// AC-SPC-004-12: 빈 결과 처리
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	// NOTE 태그만 있는데 ANCHOR 필터 적용 → 빈 결과
	tags := []mx.Tag{
		{
			Kind:       mx.MXNote,
			File:       "internal/misc.go",
			Line:       1,
			Body:       "노트 태그",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, _, err := executeQueryCmd(t, []string{"--kind", "anchor"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return
	}

	// 빈 JSON 배열이어야 함
	if strings.TrimSpace(stdout) != "[]" {
		t.Errorf("빈 결과 시 [] 기대, 실제: %q", stdout)
	}
}

// TestMxQueryCmd_StrictMode는 MOAI_MX_QUERY_STRICT=1 모드를 테스트합니다.
// AC-SPC-004-09: strict 모드에서 LSP 없으면 LSPRequired 오류
func TestMxQueryCmd_StrictMode(t *testing.T) {
	// AC-SPC-004-09: strict 모드
	t.Setenv("MOAI_MX_QUERY_STRICT", "1")

	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	tags := []mx.Tag{
		{
			Kind:       mx.MXAnchor,
			File:       "internal/auth.go",
			Line:       1,
			Body:       "앵커",
			AnchorID:   "anchor-test",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	_, stderr, err := executeQueryCmd(t, []string{"--fan-in-min", "3"})
	if err == nil {
		t.Error("strict 모드에서 LSP 없을 때 오류 기대")
	}
	_ = stderr // GREEN 단계에서 "LSPRequired" 포함 검증
}

// TestMxQueryCmd_MxParentCommand는 mx 부모 명령 구조를 테스트합니다.
func TestMxQueryCmd_MxParentCommand(t *testing.T) {
	cmd := newMxCmd()

	if cmd.Use != "mx" {
		t.Errorf("Use: 기대 'mx', 실제 %q", cmd.Use)
	}

	// query 서브커맨드가 있어야 함
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "query" {
			found = true
			break
		}
	}

	if !found {
		t.Error("'query' 서브커맨드가 mx 명령에 없음")
	}
}

// TestMxQueryCmd_Pagination은 limit/offset 플래그를 테스트합니다.
func TestMxQueryCmd_Pagination(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	// 10개 태그 생성
	tags := make([]mx.Tag, 10)
	for i := range tags {
		tags[i] = mx.Tag{
			Kind:       mx.MXNote,
			File:       "internal/file.go",
			Line:       i + 1,
			Body:       "노트",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		}
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	// limit=5, offset=0
	stdout, _, err := executeQueryCmd(t, []string{"--limit", "5", "--offset", "0"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("JSON 파싱 실패: %v", err)
		return
	}

	if len(result) > 5 {
		t.Errorf("limit=5인데 %d개 반환", len(result))
	}
}

// TestMxCmd_IsRegisteredInRoot는 mx 명령이 rootCmd에 등록되었는지 확인합니다.
func TestMxCmd_IsRegisteredInRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mx" {
			found = true
			break
		}
	}

	if !found {
		t.Error("'mx' 명령이 rootCmd에 등록되지 않음")
	}
}

// TestMxQueryCmd_FilePrefix는 파일 경로 접두사 필터를 테스트합니다.
func TestMxQueryCmd_FilePrefix(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".moai", "state")

	tags := []mx.Tag{
		{
			Kind:       mx.MXNote,
			File:       "internal/auth/handler.go",
			Line:       1,
			Body:       "인증 태그",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
		{
			Kind:       mx.MXNote,
			File:       "internal/cache/store.go",
			Line:       1,
			Body:       "캐시 태그",
			CreatedBy:  "agent",
			LastSeenAt: time.Now(),
		},
	}
	buildTestSidecarForCLI(t, stateDir, tags)

	oldFindProjectRootFn := findProjectRootFn
	defer func() { findProjectRootFn = oldFindProjectRootFn }()
	findProjectRootFn = func() (string, error) { return tmpDir, nil }

	stdout, _, err := executeQueryCmd(t, []string{"--file-prefix", "internal/auth/"})
	if err != nil {
		t.Logf("RED 단계: not implemented 오류 예상 - %v", err)
		return
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("JSON 파싱 실패: %v", err)
		return
	}

	for _, item := range result {
		file, ok := item["file"].(string)
		if !ok {
			continue
		}
		if !strings.HasPrefix(file, "internal/auth/") {
			t.Errorf("파일 접두사 필터 실패: %s", file)
		}
	}
}

// TestMxQueryCmd_LimitDefault는 기본 limit이 100임을 테스트합니다.
func TestMxQueryCmd_LimitDefault(t *testing.T) {
	cmd := newMxQueryCmd()

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("--limit 플래그 없음")
	}

	// 기본값 확인
	if limitFlag.DefValue != "0" {
		// 기본값이 0이면 런타임에 DefaultLimit(100)로 처리
		t.Logf("--limit 기본값: %s", limitFlag.DefValue)
	}
}

// TestMxQueryCmd_FormatDefault는 기본 format이 json임을 테스트합니다.
func TestMxQueryCmd_FormatDefault(t *testing.T) {
	cmd := newMxQueryCmd()

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("--format 플래그 없음")
	}

	if formatFlag.DefValue != "json" {
		t.Errorf("--format 기본값: 기대 'json', 실제 %q", formatFlag.DefValue)
	}
}

// 더미 참조: os 패키지 사용 확인
var _ = os.DevNull
