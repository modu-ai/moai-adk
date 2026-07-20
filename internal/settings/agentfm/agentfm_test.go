package agentfm

// SPEC-WEB-CONSOLE-011 M3 frontmatter patch layer 테스트
// (AC-WC11-027 idempotency + body byte 보존, EC-7 effort 부재/삭제).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleAgent는 실제 agent 파일 형상(중첩 hooks + 주석 + 본문)을 모사한다.
const sampleAgent = `---
name: manager-develop
description: run-phase implementation agent
model: inherit
effort: xhigh
# isolation은 worktree 기본이다.
isolation: worktree
memory: project
hooks:
  Stop:
    - command: moai hook subagent-stop
tools: Read, Write, Edit
---
# Body Heading

Body content with --- inline dashes and code:

` + "```yaml\nmodel: fake\n---\n```" + `

END OF BODY (byte-preservation sentinel: 한글 · émoji ✂)
`

// noEffortAgent는 effort 키 부재 형상(manager-docs 선례)이다.
const noEffortAgent = `---
name: manager-docs
model: haiku
memory: project
---
Body only.
`

func writeAgent(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func bodyOf(t *testing.T, path string) []byte {
	t.Helper()
	_, body, err := splitFrontmatter(path)
	if err != nil {
		t.Fatalf("splitFrontmatter: %v", err)
	}
	return body
}

// TestPatchIdempotencyAndBodyPreserved는 AC-WC11-027의 기계 증거다: 동일 패치
// 2회 → 파일 byte-identical, body 구간은 원본과 bytes.Equal.
func TestPatchIdempotencyAndBodyPreserved(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeAgent(t, dir, "manager-develop.md", sampleAgent)
	originalBody := bodyOf(t, path)

	if err := Patch(path, "opus", "high", false); err != nil {
		t.Fatalf("Patch #1: %v", err)
	}
	after1, _ := os.ReadFile(path)

	if err := Patch(path, "opus", "high", false); err != nil {
		t.Fatalf("Patch #2: %v", err)
	}
	after2, _ := os.ReadFile(path)

	if !bytes.Equal(after1, after2) {
		t.Errorf("patch is not idempotent:\n#1:\n%s\n#2:\n%s", after1, after2)
	}
	if !bytes.Equal(bodyOf(t, path), originalBody) {
		t.Error("body bytes changed (must be byte-untouched)")
	}
	content := string(after1)
	if !strings.Contains(content, "model: opus") || !strings.Contains(content, "effort: high") {
		t.Errorf("patched values missing:\n%s", content)
	}
	// frontmatter 주석 + 중첩 구조 보존.
	if !strings.Contains(content, "# isolation은 worktree 기본이다.") {
		t.Error("frontmatter comment lost")
	}
	if !strings.Contains(content, "command: moai hook subagent-stop") {
		t.Error("nested hooks block lost")
	}
}

// TestPatchEffortAbsentStaysAbsent는 EC-7의 preserve 반경이다: effort 부재
// agent에 model만 패치하면 effort 키가 주입되지 않는다 (REQ-WC11-029).
func TestPatchEffortAbsentStaysAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeAgent(t, dir, "manager-docs.md", noEffortAgent)

	if err := Patch(path, "sonnet", "", false); err != nil {
		t.Fatalf("Patch: %v", err)
	}
	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "effort:") {
		t.Errorf("effort key injected into effort-absent agent:\n%s", content)
	}
	if !strings.Contains(string(content), "model: sonnet") {
		t.Errorf("model not patched:\n%s", content)
	}
}

// TestPatchEffortUpsertAndDelete는 effort upsert(부재 → 설정)와 삭제("(absent)"
// 복귀)를 검증한다 (EC-7 — 섹션 seam에 없는 삭제 연산은 이 레이어 소관).
func TestPatchEffortUpsertAndDelete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeAgent(t, dir, "manager-docs.md", noEffortAgent)

	if err := Patch(path, "", "medium", false); err != nil {
		t.Fatalf("Patch upsert: %v", err)
	}
	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "effort: medium") {
		t.Errorf("effort not upserted:\n%s", content)
	}

	if err := Patch(path, "", "", true); err != nil {
		t.Fatalf("Patch delete: %v", err)
	}
	content, _ = os.ReadFile(path)
	if strings.Contains(string(content), "effort:") {
		t.Errorf("effort not deleted:\n%s", content)
	}
	if !strings.Contains(string(content), "Body only.") {
		t.Error("body lost across delete")
	}
}

// TestPatchGuards는 오류/no-op 경로를 검증한다.
func TestPatchGuards(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeAgent(t, dir, "a.md", sampleAgent)
	before, _ := os.ReadFile(path)

	if err := Patch(path, "", "", false); err != nil {
		t.Fatalf("no-op Patch errored: %v", err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Error("no-op patch mutated the file")
	}

	if err := Patch(path, "", "high", true); err == nil {
		t.Error("set+delete effort simultaneously must error")
	}
	if err := Patch(filepath.Join(dir, "missing.md"), "opus", "", false); err == nil {
		t.Error("missing file must error")
	}

	noFM := writeAgent(t, dir, "nofm.md", "just a body\n")
	if err := Patch(noFM, "opus", "", false); err == nil {
		t.Error("file without frontmatter must error")
	}
	unterminated := writeAgent(t, dir, "open.md", "---\nname: x\nno end\n")
	if err := Patch(unterminated, "opus", "", false); err == nil {
		t.Error("unterminated frontmatter must error")
	}
}

// TestPatchReadOnlyDirFails는 temp 파일 생성 실패(읽기 전용 디렉터리)의 오류
// 경로와 원본 무변경을 검증한다.
func TestPatchReadOnlyDirFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeAgent(t, dir, "a.md", sampleAgent)
	before, _ := os.ReadFile(path)

	if err := os.Chmod(dir, 0o555); err != nil {
		t.Skipf("cannot chmod dir read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := Patch(path, "opus", "", false); err == nil {
		t.Fatal("want error when temp file cannot be created")
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil || !bytes.Equal(before, after) {
		t.Error("failed patch mutated the target file")
	}
}

// TestPatchInvalidFrontmatterYAML은 frontmatter 파싱 불가/비매핑 형상의 오류
// 경로를 검증한다.
func TestPatchInvalidFrontmatterYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	badYAML := writeAgent(t, dir, "bad.md", "---\n: [unclosed\n---\nbody\n")
	if err := Patch(badYAML, "opus", "", false); err == nil {
		t.Error("invalid frontmatter yaml must error")
	}

	seqFM := writeAgent(t, dir, "seq.md", "---\n- a\n- b\n---\nbody\n")
	if err := Patch(seqFM, "opus", "", false); err == nil {
		t.Error("non-mapping frontmatter must error")
	}
}

// TestListDirReadError는 디렉터리가 아닌 경로 스캔의 오류 경로를 검증한다.
func TestListDirReadError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	notDir := writeAgent(t, dir, "file.md", sampleAgent)
	if _, err := List(notDir); err == nil {
		t.Error("List on a non-directory must error")
	}
}

// TestList는 디렉터리 스캔의 상태 추출을 검증한다: effort 부재 표시(EC-7),
// 파싱 실패 행 강등(ParseOK=false), 이름순 정렬, 디렉터리 부재 무오류.
func TestList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeAgent(t, dir, "b-develop.md", sampleAgent)
	writeAgent(t, dir, "a-docs.md", noEffortAgent)
	writeAgent(t, dir, "broken.md", "no frontmatter here\n")

	agents, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(agents) != 3 {
		t.Fatalf("List length = %d, want 3", len(agents))
	}
	if agents[0].Name != "a-docs" || agents[1].Name != "b-develop" || agents[2].Name != "broken" {
		t.Errorf("List order = %v", []string{agents[0].Name, agents[1].Name, agents[2].Name})
	}
	if agents[0].EffortPresent || agents[0].Model != "haiku" || !agents[0].ParseOK {
		t.Errorf("a-docs state = %+v (want model haiku, effort absent, ParseOK)", agents[0])
	}
	if !agents[1].EffortPresent || agents[1].Effort != "xhigh" {
		t.Errorf("b-develop state = %+v (want effort xhigh present)", agents[1])
	}
	if agents[2].ParseOK {
		t.Error("broken.md must degrade to ParseOK=false, not fail the listing")
	}

	if missing, err := List(filepath.Join(dir, "nope")); err != nil || missing != nil {
		t.Errorf("missing dir: agents=%v err=%v (want nil, nil)", missing, err)
	}
}
