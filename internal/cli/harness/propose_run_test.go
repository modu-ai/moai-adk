// Package harness — runPropose default-flag fallback + non-dry-run write 경로
// 커버리지 테스트.
//
// SPEC-HARNESS-CLI-COVERAGE-001 M4: 기존 propose_test.go는 cmd.Execute() 경로로
// --input/--output-dir를 항상 명시 지정하므로, runPropose의 default-flag fallback 블록
// (propose.go:80-90: InputPath/OutputDir/Limit 미지정 시 기본값 대입)과 non-dry-run
// WriteProposals 호출(propose.go:125-127)이 미커버 상태다. 본 파일은 runPropose를
// 직접 호출하여 두 경로를 도달시킨다.
//
// 격리(REQ-HCC-016): 모든 write는 t.TempDir() 내부. default fallback 테스트는
// DefaultInputPath가 cwd 상대로 존재하지 않아 no-op JSON을 반환하므로 디스크에 쓰지
// 않는다. HARD subagent boundary(C-HRA-008): AskUserQuestion 미호출.
package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// TestRunPropose_DefaultFlags_UsesDefaults — 모든 override flag를 생략하고 runPropose를
// 직접 호출하여 default-flag fallback 블록(propose.go:80-90)을 도달시킨다 (REQ-HCC-012).
// 기본 InputPath(proposalgen.DefaultInputPath)는 cwd 상대로 부재하므로 no-op JSON
// (reason="absent or empty")을 반환하고 WriteProposals를 실행하지 않는다.
func TestRunPropose_DefaultFlags_UsesDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewProposeCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&bytes.Buffer{})

	// OutputFlags{} — InputPath/OutputDir 모두 "" + Limit 0 → 세 default 분기 모두 도달.
	if err := runPropose(cmd, proposalgen.OutputFlags{}); err != nil {
		t.Fatalf("runPropose with default flags error = %v, want nil (graceful no-op)", err)
	}

	var got struct {
		Proposals []map[string]any `json:"proposals"`
		Reason    string           `json:"reason"`
	}
	if err := json.Unmarshal(outBuf.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout:\n%s", err, outBuf.String())
	}
	// DefaultInputPath는 패키지 테스트 cwd 기준으로 부재 → "absent or empty" no-op.
	if got.Reason != "tier-promotions.jsonl absent or empty" {
		t.Errorf("reason = %q, want absent-or-empty no-op (default InputPath resolves to a missing cwd-relative path)", got.Reason)
	}
	// Proposals nil → [] 정규화(propose.go:130-132) 도달.
	if got.Proposals == nil {
		t.Errorf("proposals must be normalized to an empty slice, not null")
	}
}

// TestRunPropose_WriteMode_PersistsProposals — actionable 패턴 1건을 담은
// tier-promotions.jsonl 픽스처에 대해 --dry-run=false로 runPropose를 직접 호출하여
// non-dry-run WriteProposals 경로(propose.go:125-127)를 도달시킨다 (REQ-HCC-011).
// 픽스처는 code_change: prefix + to_tier recommendation이어야 MapPromotions가
// candidates>0를 반환한다 (EC-3, TestPropose_AutoFlagWithActionableData와 동일 형식).
func TestRunPropose_WriteMode_PersistsProposals(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tier-promotions.jsonl")
	data := `{"ts":"2026-05-24T10:00:00Z","pattern_key":"code_change:func_extract:auth_module","from_tier":"observation","to_tier":"recommendation","observation_count":7,"confidence":0.85}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	outDir := filepath.Join(tmp, ".moai", "proposals")

	cmd := NewProposeCmd()
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&bytes.Buffer{})

	// DryRun: false + candidates>0 → WriteProposals 경로(125-127) 도달.
	err := runPropose(cmd, proposalgen.OutputFlags{
		InputPath: input,
		OutputDir: outDir,
		DryRun:    false,
	})
	if err != nil {
		t.Fatalf("runPropose write mode error = %v, want nil", err)
	}

	// WriteProposals가 <outDir>/<draft>/{spec.md,proposal.json}를 생성했는지 검증.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir outDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 draft dir written, got %d", len(entries))
	}
	draftDir := filepath.Join(outDir, entries[0].Name())
	for _, fname := range []string{"spec.md", "proposal.json"} {
		if _, err := os.Stat(filepath.Join(draftDir, fname)); err != nil {
			t.Errorf("expected %s in %s; stat err = %v", fname, draftDir, err)
		}
	}
}

// TestRunPropose_ReadError_Propagates — InputPath가 디렉터리일 때 ReadPromotions가
// non-ENOENT 에러를 반환하고, runPropose가 이를 wrapped error로 전파함을 검증한다
// (propose.go:93-95 error branch). os.Open은 디렉터리에 성공하지만 bufio.Scanner가
// "is a directory" 에러를 내므로 ReadPromotions가 error를 반환한다 — IsNotExist no-op이
// 아닌 진짜 IO 에러 경로다.
func TestRunPropose_ReadError_Propagates(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	// InputPath를 디렉터리로 지정 → ReadPromotions가 IsNotExist가 아닌 read 에러 반환.
	inputDir := filepath.Join(tmp, "promotions-as-dir")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("mkdir input-as-dir: %v", err)
	}

	cmd := NewProposeCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := runPropose(cmd, proposalgen.OutputFlags{
		InputPath: inputDir,
		OutputDir: filepath.Join(tmp, ".moai", "proposals"),
		DryRun:    true,
	})
	if err == nil {
		t.Fatal("runPropose must return an error when InputPath is a directory (read failure)")
	}
	if !strings.Contains(err.Error(), "read promotions") {
		t.Errorf("error should carry the read-promotions wrapping prefix, got: %v", err)
	}
}

// TestRunPropose_WriteError_Propagates — OutputDir 경로의 부모가 regular file이면
// WriteProposals의 MkdirAll이 실패하고, runPropose가 wrapped write error를 전파함을
// 검증한다 (propose.go:125-127 error branch). actionable 픽스처로 candidates>0 +
// DryRun=false를 만들어 WriteProposals 경로에 진입시킨다.
func TestRunPropose_WriteError_Propagates(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	input := filepath.Join(tmp, "tier-promotions.jsonl")
	data := `{"ts":"2026-05-24T10:00:00Z","pattern_key":"code_change:func_extract:auth_module","from_tier":"observation","to_tier":"recommendation","observation_count":7,"confidence":0.85}
`
	if err := os.WriteFile(input, []byte(data), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	// OutputDir의 부모 경로 컴포넌트를 regular file로 만든다 → 그 아래 MkdirAll 실패.
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	outDir := filepath.Join(blocker, "proposals") // blocker는 파일 → 하위 mkdir 불가.

	cmd := NewProposeCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := runPropose(cmd, proposalgen.OutputFlags{
		InputPath: input,
		OutputDir: outDir,
		DryRun:    false,
	})
	if err == nil {
		t.Fatal("runPropose must return an error when OutputDir parent is a regular file (write failure)")
	}
	if !strings.Contains(err.Error(), "write proposals") {
		t.Errorf("error should carry the write-proposals wrapping prefix, got: %v", err)
	}
}
