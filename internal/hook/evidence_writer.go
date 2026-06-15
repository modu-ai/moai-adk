// SPEC-STOP-EVIDENCE-WRITER-001: 기록 시점 증거 writer (IMP-WRITER).
// GATE-001 이 만든 dormant Stop 증거 게이트를 production 에서 genuinely 활성화한다.
// Bash 테스트 결과(IsTestPass/IsTestFail) 와 Edit/Write path-kind(PathKind) 증거를
// PostToolUse 시점에 기록해, 기존 buildSessionLedger → evaluateEvidence 가
// 발화 조건(code-change 성공 주장 + 관측된 pass 없음)에 도달하게 한다.
//
// 게이트 read-side(session_ledger.go / stop.go / telemetry.types.go) 는 PRESERVE —
// 본 파일은 write-side 만 추가한다. 새 저장소를 만들지 않고 기존
// telemetry.RecordSkillUsage 경로를 재사용한다 (REQ-SEW-005).
package hook

import (
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// testCommandSignatures 는 다중 언어 테스트 러너의 고정 taxonomy 다 (REQ-SEW-006).
// 명령 텍스트가 이 중 하나를 토큰으로 포함하면 테스트 명령으로 인식한다.
// config 파일이 아닌 코드 편집으로만 확장한다 (spec.md §A.5 scope discipline).
var testCommandSignatures = []string{
	"go test",
	"pytest",
	"cargo test",
	"npm test", "npm run test",
	"pnpm test", "yarn test",
	"jest", "vitest",
}

// codeExtensions 는 source-code 확장자 고정 taxonomy 다 (REQ-SEW-009).
var codeExtensions = []string{
	".go", ".py", ".ts", ".js", ".rs", ".java", ".kt", ".cs",
	".rb", ".php", ".ex", ".cpp", ".scala", ".r", ".dart", ".swift",
}

// docsExtensions 는 documentation/prose 확장자 고정 taxonomy 다 (REQ-SEW-010).
var docsExtensions = []string{".md", ".mdx", ".txt", ".rst"}

// classifyTestCommand 는 Bash 명령 텍스트 + 관측된 결과로부터
// (isTest, isPass, isFail) 을 도출하는 부수효과 없는 pure 함수다 (REQ-SEW-003).
// I/O 를 수행하지 않는다 — 호출자가 이미 확보한 바이트를 넘긴다.
//
//  1. 인식(REQ-SEW-006): 명령이 testCommandSignatures 중 하나를 토큰으로
//     포함하면 isTest=true. 비테스트 명령(go build, ls, ...)은 isTest=false 이며
//     두 이진 플래그 모두 false (REQ-SEW-008).
//  2. pass/fail 도출(REQ-SEW-007): isTest 일 때만, 먼저 구조화된 exit-code 신호
//     ({"exit_code":0}=pass, 비0=fail, {"interrupted":true}=fail) 를 보고,
//     없으면 출력 텍스트 휴리스틱(go --- FAIL / ok, pytest passed/failed,
//     cargo test result) 으로 보강한다.
//  3. graceful degradation(REQ-SEW-013, R1): 결과가 부재/해석불가/신호없음이면
//     어느 플래그도 set 하지 않는다. 부재 ≠ 실패 — 게이트를 보수적으로 유지한다.
func classifyTestCommand(command string, result []byte) (isTest, isPass, isFail bool) {
	trimmed := strings.TrimSpace(command)
	for _, sig := range testCommandSignatures {
		if isTestCommandToken(trimmed, sig) {
			isTest = true
			break
		}
	}
	if !isTest {
		return false, false, false
	}

	// (2) 구조화된 exit-code 신호 우선.
	if pass, fail, ok := deriveFromExitCode(result); ok {
		return true, pass, fail
	}

	// (2b) 출력 텍스트 휴리스틱 (exit-code 신호 부재 시).
	if pass, fail, ok := deriveFromOutputText(result); ok {
		return true, pass, fail
	}

	// (3) 신호 없음 → 어느 플래그도 set 하지 않음 (부재 ≠ 실패).
	return true, false, false
}

// isTestCommandToken 은 명령이 signature 로 시작하는 테스트 호출인지 판정한다.
// signature 가 명령 문두에 위치해야 한다 (예: "cat foo_test.go" 는 매칭하지 않음 —
// "go test" 가 인자가 아닌 명령 토큰이어야 한다).
func isTestCommandToken(command, sig string) bool {
	if command == sig {
		return true
	}
	// signature 다음에 공백이 오면 명령 토큰 (예: "go test ./...").
	return strings.HasPrefix(command, sig+" ")
}

// exitCodeSignal 은 Bash ToolResponse/ToolOutput 의 구조화된 종료 신호를 담는다.
type exitCodeSignal struct {
	ExitCode    *int  `json:"exit_code"`
	Interrupted *bool `json:"interrupted"`
}

// deriveFromExitCode 는 구조화된 종료 신호로부터 pass/fail 을 도출한다.
// ok=false 면 종료 신호가 없어 출력 텍스트 휴리스틱으로 넘어가야 한다.
func deriveFromExitCode(result []byte) (pass, fail, ok bool) {
	if len(result) == 0 {
		return false, false, false
	}
	var sig exitCodeSignal
	if err := json.Unmarshal(result, &sig); err != nil {
		return false, false, false
	}
	if sig.Interrupted != nil && *sig.Interrupted {
		return false, true, true // 중단 = clean exit 아님 = fail.
	}
	if sig.ExitCode != nil {
		if *sig.ExitCode == 0 {
			return true, false, true
		}
		return false, true, true
	}
	return false, false, false
}

// deriveFromOutputText 는 테스트 러너 출력 텍스트로부터 pass/fail 을 도출한다
// (exit-code 신호 부재 시 fallback). fail 마커를 pass 마커보다 우선한다 (보수적).
// ok=false 면 인식 가능한 마커가 없어 어느 플래그도 set 하지 않아야 한다.
func deriveFromOutputText(result []byte) (pass, fail, ok bool) {
	if len(result) == 0 {
		return false, false, false
	}
	text := string(result)

	// (a) 정밀 마커 우선 — go/cargo 의 명시적 결과 라인은 상호배타적이라 신뢰도가 높다.
	//     pass 출력이 "0 failed" 같은 카운트 문구를 포함할 수 있으므로 정밀 마커를
	//     일반 카운트 휴리스틱보다 먼저 본다.
	hasPreciseFail := strings.Contains(text, "--- FAIL") ||
		strings.Contains(text, "FAIL\t") ||
		strings.Contains(text, "test result: FAILED")
	hasPrecisePass := strings.Contains(text, "ok  \t") ||
		strings.Contains(text, "ok \t") ||
		strings.Contains(text, "test result: ok")
	if hasPreciseFail {
		return false, true, true
	}
	if hasPrecisePass {
		return true, false, true
	}

	// (b) 카운트-단어 휴리스틱 (pytest "1 failed, 2 passed" / "3 passed").
	//     "failed" 가 있으면 fail 우선 (보수적).
	if strings.Contains(text, " failed") {
		return false, true, true
	}
	if strings.Contains(text, " passed") {
		return true, false, true
	}
	return false, false, false
}

// classifyPathKind 는 파일 경로를 telemetry.PathKind* 상수로 분류하는
// 부수효과 없는 pure 함수다 (REQ-SEW-004). first match wins:
//  1. 확장자 ∈ codeExtensions → PathKindCodeChange (REQ-SEW-009).
//  2. 확장자 ∈ docsExtensions, 또는 base 가 CHANGELOG*/README* → PathKindDocsOnly
//     (REQ-SEW-010).
//  3. 그 외 → PathKindUnknown (REQ-SEW-011) — 게이트가 보수적으로 처리.
func classifyPathKind(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	base := filepath.Base(filePath)

	for _, e := range codeExtensions {
		if ext == e {
			return telemetry.PathKindCodeChange
		}
	}
	for _, e := range docsExtensions {
		if ext == e {
			return telemetry.PathKindDocsOnly
		}
	}
	// 확장자 없는 docs base name (예: README, CHANGELOG).
	if hasDocsBaseName(base) {
		return telemetry.PathKindDocsOnly
	}
	return telemetry.PathKindUnknown
}

// hasDocsBaseName 은 base name 이 README/CHANGELOG 계열인지 판정한다.
func hasDocsBaseName(base string) bool {
	upper := strings.ToUpper(base)
	return strings.HasPrefix(upper, "README") || strings.HasPrefix(upper, "CHANGELOG")
}

// bashToolInput 은 Bash 이벤트의 ToolInput 모양이다 ({"command": "..."}).
type bashToolInput struct {
	Command string `json:"command"`
}

// fileToolInput 은 Edit/Write 이벤트의 ToolInput 모양이다 ({"file_path": "..."}).
type fileToolInput struct {
	FilePath string `json:"file_path"`
}

// buildEvidenceRecord 는 PostToolUse 입력으로부터 증거 레코드를 조립하는
// pure assembler 다. ToolInput 을 파싱해 §2 분류기를 호출하고 §3 Outcome 표를
// 적용한다. 증거가 없는 이벤트(비테스트 Bash, unknown-ext 경로)면 ok=false 를
// 반환해 기록을 건너뛴다 (write-volume discipline, design.md §2.3).
//
// @MX:NOTE: [AUTO] code-change Edit/Write 가 Outcome=success 를 갖는 이유 —
// "코드를 바꾸고 완료로 간주한다"는 암묵적 성공 주장이다. 이 success 주장이
// code-change 세션에서 관측된 test-pass 없이 존재하면(혹은 test-fail 과 짝지으면)
// evaluateEvidence 가 advisory finding 을 낸다 (design.md §3). 비자명한 business rule.
func buildEvidenceRecord(input *HookInput) (telemetry.UsageRecord, bool) {
	rec := telemetry.UsageRecord{
		Timestamp: time.Now().UTC(),
		SessionID: input.SessionID, // REQ-SEW-016: 세션 상관관계.
		AgentType: input.AgentType,
	}

	switch input.ToolName {
	case "Bash":
		return buildBashRecord(input, rec)
	case "Edit", "Write":
		return buildFileRecord(input, rec)
	default:
		return telemetry.UsageRecord{}, false
	}
}

// buildBashRecord 는 Bash 이벤트로부터 테스트 증거 레코드를 조립한다 (§3 표).
// 비테스트 명령이면 ok=false (기록 없음).
func buildBashRecord(input *HookInput, rec telemetry.UsageRecord) (telemetry.UsageRecord, bool) {
	var bi bashToolInput
	if len(input.ToolInput) > 0 {
		_ = json.Unmarshal(input.ToolInput, &bi)
	}
	if bi.Command == "" {
		return telemetry.UsageRecord{}, false
	}

	result := input.ToolResponse
	if len(result) == 0 {
		result = input.ToolOutput // 레거시 fallback.
	}

	isTest, isPass, isFail := classifyTestCommand(bi.Command, result)
	if !isTest {
		return telemetry.UsageRecord{}, false // 비테스트 Bash → 기록 없음 (§2.3).
	}

	rec.IsTestPass = isPass
	rec.IsTestFail = isFail
	switch {
	case isPass:
		// 관측된 pass → BinaryPass=true → 게이트가 nil 반환 (올바름: 실제 pass).
		rec.Outcome = telemetry.OutcomeSuccess
	case isFail:
		// fail 은 성공 주장이 아니다; BinaryFail 에 기여(이진 신호 관측됨).
		rec.Outcome = telemetry.OutcomeError
	default:
		// ambiguous → 신호 없음 (graceful degradation).
		rec.Outcome = telemetry.OutcomeUnknown
	}
	return rec, true
}

// buildFileRecord 는 Edit/Write 이벤트로부터 path-kind 증거 레코드를 조립한다 (§3 표).
// unknown-ext 경로면 ok=false (기록 없음 — false code-change 주장 방지).
func buildFileRecord(input *HookInput, rec telemetry.UsageRecord) (telemetry.UsageRecord, bool) {
	var fi fileToolInput
	if len(input.ToolInput) > 0 {
		_ = json.Unmarshal(input.ToolInput, &fi)
	}
	if fi.FilePath == "" {
		return telemetry.UsageRecord{}, false
	}

	kind := classifyPathKind(fi.FilePath)
	if kind == telemetry.PathKindUnknown {
		return telemetry.UsageRecord{}, false // unknown → 기록 없음 (§2.3, REQ-SEW-011).
	}

	rec.PathKind = kind
	if kind == telemetry.PathKindCodeChange {
		// code-change Edit/Write = 암묵적 성공 주장 (@MX:NOTE 참조).
		rec.Outcome = telemetry.OutcomeSuccess
	} else {
		// docs-only 는 면제 — 성공 주장 불필요.
		rec.Outcome = telemetry.OutcomeUnknown
	}
	return rec, true
}

// @MX:ANCHOR: [AUTO] logEvidence 는 증거 게이트 활성화의 production write 경로다 —
// postToolHandler.Handle 이 Bash/Edit/Write 마다 호출
// @MX:REASON: [AUTO] fan_in>=3 (Handle + 테스트들). 본 함수가 레코드를 기록하지
// 않으면 GATE-001 게이트가 production 에서 발화하지 못한다 (활성화 hinge).
//
// logEvidence 는 증거를 갖는 PostToolUse 이벤트(Bash 테스트 결과, Edit/Write
// path-kind)를 세션 텔레메트리 저장소에 기록해 Stop 증거 게이트가 발화하게 한다
// (REQ-SEW-005, REQ-SEW-012, REQ-SEW-013). best-effort — 에러는 slog.Warn 으로
// 남기고 반환하지 않으며, PostToolUse 훅을 절대 차단하지 않는다 (fail-open).
func logEvidence(input *HookInput) {
	projectRoot := resolveProjectRoot(input) // REUSE — no-project-root skip (REQ-SEW-013).
	if projectRoot == "" {
		return
	}
	rec, ok := buildEvidenceRecord(input) // pure: classify + assemble (§2, §3).
	if !ok {
		return // 비증거 이벤트 → 기록 없음 (§2.3).
	}
	if err := telemetry.RecordSkillUsage(projectRoot, rec); err != nil { // REUSE store (REQ-SEW-005).
		slog.Warn("evidence writer: failed to record",
			"tool", input.ToolName,
			"session_id", input.SessionID,
			"error", err,
		)
	}
}
