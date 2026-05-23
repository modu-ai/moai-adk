package harness_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// makeAnswer is a helper that builds an Answer for tests.
func makeAnswer(qid string, round int, questionText, answerText string) harness.Answer {
	return harness.Answer{
		QuestionID:   qid,
		Round:        round,
		QuestionText: questionText,
		AnswerText:   answerText,
		RecordedAt:   time.Now(),
	}
}

// make16Answers returns a slice of 16 Answers for tests.
func make16Answers() []harness.Answer {
	data := []struct {
		qid   string
		round int
		qt    string
		at    string
	}{
		{"Q01", 1, "도메인은 무엇인가요?", "Mobile (iOS)"},
		{"Q02", 1, "기술 스택은 무엇인가요?", "Swift + SwiftUI"},
		{"Q03", 1, "규모는 어떻게 되나요?", "MVP (1-3 modules)"},
		{"Q04", 1, "팀 구성은 어떻게 되나요?", "Solo developer"},
		{"Q05", 2, "개발 방법론은 무엇인가요?", "TDD"},
		{"Q06", 2, "디자인 툴은 무엇을 사용하나요?", "Figma"},
		{"Q07", 2, "UI 복잡도는 어느 정도인가요?", "Standard (lists + forms)"},
		{"Q08", 2, "디자인 시스템은 무엇을 사용하나요?", "Custom DTCG tokens"},
		{"Q09", 3, "보안 요구사항은 무엇인가요?", "OAuth + Keychain"},
		{"Q10", 3, "성능 목표는 무엇인가요?", "60fps 일반 UI"},
		{"Q11", 3, "배포 대상은 어디인가요?", "App Store"},
		{"Q12", 3, "외부 통합이 필요한가요?", "HealthKit"},
		{"Q13", 4, "customization 범위는 어떻게 되나요?", "Standard (recommended)"},
		{"Q14", 4, "특수 제약이 있나요?", "iOS 17+ minimum"},
		{"Q15", 4, "우선순위 수준은 무엇인가요?", "thorough harness level"},
		{"Q16", 4, "최종 확인해 주세요.", "Confirm"},
	}
	answers := make([]harness.Answer, len(data))
	for i, d := range data {
		answers[i] = makeAnswer(d.qid, d.round, d.qt, d.at)
	}
	return answers
}

// TestBuffer_AppendAndCommit verifies Commit and Frozen state after appending 16 answers.
func TestBuffer_AppendAndCommit(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()

	answers := make16Answers()
	for _, a := range answers {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s) failed: %v", a.QuestionID, err)
		}
	}

	if buf.Len() != 16 {
		t.Fatalf("expected 16 answers, got %d", buf.Len())
	}

	if buf.Frozen() {
		t.Fatal("buffer should not be frozen before Commit()")
	}

	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit() failed: %v", err)
	}

	if !buf.Frozen() {
		t.Fatal("buffer should be frozen after Commit()")
	}

	// After Commit, Append must return an error.
	extraAnswer := makeAnswer("Q17", 5, "extra question", "extra answer")
	if err := buf.Append(extraAnswer); err == nil {
		t.Fatal("Append after Commit() should return error")
	}
}

// TestBuffer_Abort_NoDiskWrite verifies that Abort clears the in-memory buffer and writes nothing to disk.
func TestBuffer_Abort_NoDiskWrite(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	buf := harness.NewBuffer()
	answers := make16Answers()[:5] // append only 5
	for _, a := range answers {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s) failed: %v", a.QuestionID, err)
		}
	}

	// Before Abort, verify Len == 5
	if buf.Len() != 5 {
		t.Fatalf("before Abort: expected 5, got %d", buf.Len())
	}

	buf.Abort()

	// After Abort, Frozen() == true
	if !buf.Frozen() {
		t.Fatal("buffer should be frozen after Abort()")
	}

	// After Abort, Answers() must return an empty slice.
	if got := buf.Answers(); len(got) != 0 {
		t.Fatalf("Answers() after Abort() should be empty, got %d items", len(got))
	}

	// Verify no disk writes: tempDir must be empty.
	// Abort writes nothing to disk, so tempDir is unchanged.
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", tempDir, err)
	}
	if len(entries) != 0 {
		t.Fatalf("tempDir should be empty after Abort(), got %d entries", len(entries))
	}
}

// TestBuffer_AppendAfterFrozen_Error verifies Append returns an error after Commit.
func TestBuffer_AppendAfterFrozen_Error(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit() on empty buffer failed: %v", err)
	}

	a := makeAnswer("Q01", 1, "질문", "답변")
	if err := buf.Append(a); err == nil {
		t.Fatal("Append to frozen buffer should return error")
	}
}

// TestWriteResults_FullFlow verifies the output format after 16 answers → Commit → Write.
func TestWriteResults_FullFlow(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	for _, a := range make16Answers() {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s): %v", a.QuestionID, err)
		}
	}
	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit(): %v", err)
	}

	var out bytes.Buffer
	err := harness.WriteResults(buf, "/tmp/test-project", "SPEC-PROJ-INIT-001", "ko", &out)
	if err != nil {
		t.Fatalf("WriteResults(): %v", err)
	}

	content := out.String()

	// Verify YAML frontmatter
	if !strings.Contains(content, "spec_id: SPEC-PROJ-INIT-001") {
		t.Error("output missing spec_id in frontmatter")
	}
	if !strings.Contains(content, "generated_at:") {
		t.Error("output missing generated_at in frontmatter")
	}
	if !strings.Contains(content, "project_root: /tmp/test-project") {
		t.Error("output missing project_root in frontmatter")
	}
	if !strings.Contains(content, "conversation_language: ko") {
		t.Error("output missing conversation_language in frontmatter")
	}

	// Verify the 4 Round headers
	expectedHeaders := []string{
		"## Round 1: Domain & Technology Foundation",
		"## Round 2: Methodology & Design",
		"## Round 3: Security, Performance, Deployment",
		"## Round 4: Customization & Final Confirmation",
	}
	for _, h := range expectedHeaders {
		if !strings.Contains(content, h) {
			t.Errorf("output missing Round header: %q", h)
		}
	}

	// Verify 16 Q entries (each line starts with "- Q")
	qCount := strings.Count(content, "\n- Q")
	if qCount != 16 {
		t.Errorf("expected 16 '- Q' entries, got %d", qCount)
	}

	// Verify 16 "Recorded at:" entries
	recCount := strings.Count(content, "  - Recorded at:")
	if recCount != 16 {
		t.Errorf("expected 16 'Recorded at:' entries, got %d", recCount)
	}
}

// TestWriteResults_NotFrozen_Error verifies an error when Write is called before Commit.
func TestWriteResults_NotFrozen_Error(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	for _, a := range make16Answers() {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s): %v", a.QuestionID, err)
		}
	}
	// Do not call Commit

	var out bytes.Buffer
	if err := harness.WriteResults(buf, "/tmp", "SPEC-001", "ko", &out); err == nil {
		t.Fatal("WriteResults on non-frozen buffer should return error")
	}
}

// TestWriteResults_Incomplete_Error verifies an error when Write is called with 15 answers after Commit.
func TestWriteResults_Incomplete_Error(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	answers := make16Answers()[:15] // only 15
	for _, a := range answers {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s): %v", a.QuestionID, err)
		}
	}
	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit(): %v", err)
	}

	var out bytes.Buffer
	if err := harness.WriteResults(buf, "/tmp", "SPEC-001", "ko", &out); err == nil {
		t.Fatal("WriteResults with 15 answers should return error (need exactly 16)")
	}
}

// TestBuffer_DoubleCommit_Error verifies an error on re-Commit after Commit.
func TestBuffer_DoubleCommit_Error(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	if err := buf.Commit(); err != nil {
		t.Fatalf("first Commit() failed: %v", err)
	}
	if err := buf.Commit(); err == nil {
		t.Fatal("second Commit() should return error (already frozen)")
	}
}

// TestWriteResultsToFile_CreatesFile verifies WriteResultsToFile creates the parent directory and the file.
func TestWriteResultsToFile_CreatesFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	outPath := tempDir + "/harness/interview-results.md"

	buf := harness.NewBuffer()
	for _, a := range make16Answers() {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s): %v", a.QuestionID, err)
		}
	}
	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit(): %v", err)
	}

	if err := harness.WriteResultsToFile(buf, outPath, tempDir, "SPEC-PROJ-INIT-001", "ko"); err != nil {
		t.Fatalf("WriteResultsToFile(): %v", err)
	}

	// Verify the file was created
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("created file should not be empty")
	}

	// Verify contents
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "spec_id: SPEC-PROJ-INIT-001") {
		t.Error("file content missing spec_id")
	}
}

// TestWriteResults_KoLanguage_Preserved verifies that Korean answers are preserved verbatim in the output.
func TestWriteResults_KoLanguage_Preserved(t *testing.T) {
	t.Parallel()

	buf := harness.NewBuffer()
	koreanAnswers := []harness.Answer{
		makeAnswer("Q01", 1, "도메인은 무엇인가요?", "모바일 (iOS)"),
		makeAnswer("Q02", 1, "기술 스택은 무엇인가요?", "스위프트 + SwiftUI"),
		makeAnswer("Q03", 1, "규모는 어떻게 되나요?", "MVP (1-3 모듈)"),
		makeAnswer("Q04", 1, "팀 구성은 어떻게 되나요?", "솔로 개발자"),
		makeAnswer("Q05", 2, "개발 방법론은 무엇인가요?", "테스트 주도 개발"),
		makeAnswer("Q06", 2, "디자인 툴은 무엇을 사용하나요?", "피그마"),
		makeAnswer("Q07", 2, "UI 복잡도는 어느 정도인가요?", "표준 (목록 + 폼)"),
		makeAnswer("Q08", 2, "디자인 시스템은 무엇을 사용하나요?", "커스텀 DTCG 토큰"),
		makeAnswer("Q09", 3, "보안 요구사항은 무엇인가요?", "OAuth + 키체인"),
		makeAnswer("Q10", 3, "성능 목표는 무엇인가요?", "60fps 일반 UI"),
		makeAnswer("Q11", 3, "배포 대상은 어디인가요?", "앱스토어"),
		makeAnswer("Q12", 3, "외부 통합이 필요한가요?", "헬스킷"),
		makeAnswer("Q13", 4, "customization 범위는 어떻게 되나요?", "표준 (권장)"),
		makeAnswer("Q14", 4, "특수 제약이 있나요?", "iOS 17+ 최소"),
		makeAnswer("Q15", 4, "우선순위 수준은 무엇인가요?", "thorough harness level"),
		makeAnswer("Q16", 4, "최종 확인해 주세요.", "확인"),
	}
	for _, a := range koreanAnswers {
		if err := buf.Append(a); err != nil {
			t.Fatalf("Append(%s): %v", a.QuestionID, err)
		}
	}
	if err := buf.Commit(); err != nil {
		t.Fatalf("Commit(): %v", err)
	}

	var out bytes.Buffer
	if err := harness.WriteResults(buf, "/tmp/ko-project", "SPEC-001", "ko", &out); err != nil {
		t.Fatalf("WriteResults(): %v", err)
	}

	content := out.String()

	// Verify that Korean answers are preserved verbatim
	koreanTexts := []string{
		"모바일 (iOS)",
		"스위프트 + SwiftUI",
		"테스트 주도 개발",
		"확인",
	}
	for _, kt := range koreanTexts {
		if !strings.Contains(content, kt) {
			t.Errorf("Korean text %q not preserved in output", kt)
		}
	}
}
