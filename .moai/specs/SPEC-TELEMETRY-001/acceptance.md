---
spec_id: SPEC-TELEMETRY-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. spec.md frontmatter는 "Implemented"로
  표시되어 있으며, 실제로 `internal/telemetry/` 패키지 전체(recorder.go,
  outcome.go, report.go, async_recorder.go, types.go)와 `internal/cli/telemetry.go`
  (CLI subcommand), `internal/hook/post_tool.go:169-173` (Skill tool PostToolUse
  훅 통합), `internal/hook/session_start.go:108-112,527-530` (90일 retention)이
  구현되어 있음. plan-auditor 2026-04-24의 "skill_usage missing" 관찰은
  함수 이름이 `logSkillUsage`(wrapper)이고 내부적으로 `telemetry.RecordSkillUsage`
  를 호출하기 때문에 문자열 "skill_usage" 직접 검색이 매칭되지 않았던 것으로
  판단됨. 본 backfill은 spec.md의 R1~R4를 실제 구현과 대조하여 AC 역도출.
---

# Acceptance Criteria — SPEC-TELEMETRY-001

Skill 사용 Telemetry(Skill invocation 기록, 결과 신호 추론, 보고서 생성, 프라이버시/저장 한계) 구현의 관찰 가능한 인수 기준.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| R1 (Skill Invocation Recording) | AC-001 ~ AC-004 | `internal/telemetry/recorder.go`, `types.go`, `hook/post_tool.go:169-173`, `hook/post_tool_metrics.go:44-84` |
| R2 (Outcome Signal Detection) | AC-005, AC-006 | `internal/telemetry/outcome.go`, `hook/session_start.go` stop integration |
| R3 (Telemetry Report) | AC-007, AC-008 | `internal/cli/telemetry.go`, `internal/telemetry/report.go` |
| R4 (Privacy and Storage) | AC-009, AC-010, AC-011 | `internal/telemetry/recorder.go:19-24, 107-150`, `hook/session_start.go:108-112,527-530` |

## AC-001: telemetry 패키지가 UsageRecord 스키마를 정의한다

**Given** `internal/telemetry/types.go` 타입 정의에서,
**When** `UsageRecord` 구조체를 조회하면,
**Then** 필드와 JSON 태그가 다음을 포함해야 한다: `Timestamp (ts)`, `SessionID (session_id)`, `SkillID (skill_id)`, `Trigger (trigger)`, `ContextHash (context_hash)`, `AgentType (agent_type)`, `Phase (phase)`, `DurationMs (duration_ms)`, `Outcome (outcome)`.

**Verification**: `internal/telemetry/types.go:7-17` — 정확한 9개 필드 및 JSON 태그 일치.

## AC-002: RecordSkillUsage가 일일 rotate JSONL 파일에 append한다

**Given** 프로젝트 루트와 `UsageRecord` 인스턴스(timestamp 포함)에서,
**When** `telemetry.RecordSkillUsage(projectRoot, r)`가 호출되면,
**Then** 기록은 `<projectRoot>/.moai/evolution/telemetry/usage-YYYY-MM-DD.jsonl`에 append되어야 하고(day key는 UTC), 파일 부재 시 0o644 권한으로 생성되며 디렉터리가 없으면 0o755로 생성되어야 한다.

**Verification**: `internal/telemetry/recorder.go:38-69` — `os.MkdirAll(dir, 0o755)`, `dayKey := r.Timestamp.UTC().Format("2006-01-02")`, `os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)`.

## AC-003: 쓰기는 package-level mutex로 보호된다

**Given** 복수 goroutine에서 동시에 `RecordSkillUsage`를 호출하는 상황에서,
**When** 동시성 접근이 발생하면,
**Then** `recordMu sync.Mutex`가 각 쓰기(marshal + append)를 직렬화하여 줄 단위 쓰기가 원자적이어야 하며, JSONL 라인 깨짐이 없어야 한다.

**Verification**: `internal/telemetry/recorder.go:15-17` (`var recordMu sync.Mutex`), `:56-57` (`recordMu.Lock(); defer recordMu.Unlock()`). 테스트: `recorder_test.go`.

## AC-004: PostToolUse 훅이 Skill 도구 호출을 기록한다

**Given** Claude Code가 Skill 도구를 호출한 PostToolUse 이벤트에서,
**When** `internal/hook/post_tool.go`가 이 이벤트를 처리하면,
**Then** `input.ToolName == "Skill"` 조건에서 `logSkillUsage(input)`이 호출되어야 하고, 내부에서 `telemetry.RecordSkillUsage(projectRoot, r)`로 위임되어야 한다. 오류는 로깅되되 전파되지 않아야 한다(best-effort).

**Verification**: `internal/hook/post_tool.go:169-173` (Skill tool interception), `internal/hook/post_tool_metrics.go:44-84` (`logSkillUsage` 함수 + `telemetry.RecordSkillUsage` 위임). 테스트: `internal/hook/post_tool_telemetry_test.go`.

## AC-005: DetermineOutcome은 OutcomeUnknown을 보수적 기본값으로 사용한다

**Given** 빈 이벤트 슬라이스 또는 어떤 signal flag도 설정되지 않은 이벤트에서,
**When** `telemetry.DetermineOutcome(events)`가 호출되면,
**Then** 반환값은 반드시 `OutcomeUnknown`("unknown")이어야 한다. 이는 실패 SPEC/성공 SPEC 신호가 부재할 때 evolution 제안을 오염시키지 않기 위한 보수적 기본값이다.

**Verification**: `internal/telemetry/outcome.go:17-20` (`if len(events) == 0 { return OutcomeUnknown }`), `:36-39` (모든 flag false일 때 unknown 반환). 테스트: `outcome_test.go`.

## AC-006: 혼합 signal(test pass + test fail 또는 + error)은 partial로 분류된다

**Given** 이벤트 목록에 `IsTestPass == true`이면서 `IsTestFail == true` (또는 `IsError == true`)인 경우에서,
**When** `DetermineOutcome(events)`가 호출되면,
**Then** 반환값은 `OutcomePartial`이어야 한다. 순수 error(test pass 없음)는 `OutcomeError`, 순수 test pass는 `OutcomeSuccess`여야 한다.

**Verification**: `internal/telemetry/outcome.go:41-49` (partial 분기), `:51-54` (error 분기), `:57-59` (success 분기). 테스트: `outcome_test.go`.

## AC-007: `moai telemetry report` 서브커맨드가 사용량 효과 보고서를 생성한다

**Given** `.moai/evolution/telemetry/usage-*.jsonl` 파일에 데이터가 존재하는 프로젝트에서,
**When** 사용자가 `moai telemetry report --days 30`을 실행하면,
**Then** stdout에 skill별 사용량(Uses/Success/Partial/Error), 상위 co-occurrence, 저사용 skill(< 3회/창) 섹션이 포함된 요약 보고서가 출력되어야 한다.

**Verification**: `internal/cli/telemetry.go:11-65` — `telemetryCmd`, `telemetryReportCmd`, `runTelemetryReport` 함수 및 `--days` flag(기본 30). `internal/telemetry/report.go` `GenerateReport` 구현.

## AC-008: --days는 양의 정수여야 하며 0/음수는 에러를 반환한다

**Given** `moai telemetry report --days 0` 또는 음수가 전달된 상황에서,
**When** 명령이 실행되면,
**Then** "telemetry report: --days must be a positive integer" 에러를 반환하고 exit code는 non-zero여야 한다.

**Verification**: `internal/cli/telemetry.go:49-51` — `if days <= 0 { return fmt.Errorf("telemetry report: --days must be a positive integer") }`.

## AC-009: HashContext는 SHA-256 첫 8자 hex digest를 반환하며 비가역적이다

**Given** 임의의 사용자 prompt 또는 task description 문자열에서,
**When** `telemetry.HashContext(input)`가 호출되면,
**Then** 반환값은 정확히 8자의 hex 문자열이어야 하고, 원본 입력이 PII(사용자 이름, 파일 내용, 경로 등)를 포함해도 8자 truncation으로 인해 복원이 비가역적이어야 한다.

**Verification**: `internal/telemetry/recorder.go:19-24` — `sum := sha256.Sum256([]byte(input)); return hex.EncodeToString(sum[:])[:8]`.

## AC-010: PruneOldFiles는 retentionDays보다 오래된 파일만 삭제한다

**Given** `.moai/evolution/telemetry/`에 `usage-2026-01-01.jsonl`(오래됨)과 `usage-<최근>.jsonl`(유효)이 공존하는 상황에서,
**When** `telemetry.PruneOldFiles(projectRoot, 90)`이 호출되면,
**Then** `fileDate.Before(now - 90d)`인 파일만 `os.Remove`로 삭제되어야 하고, `usage-` 접두사 / `.jsonl` 접미사 / 유효 날짜 parse 조건을 만족하지 않는 파일은 건드리지 않아야 한다. 디렉터리 부재 시 에러 대신 nil을 반환해야 한다.

**Verification**: `internal/telemetry/recorder.go:107-150` — `PruneOldFiles` 로직. `:114-117` 디렉터리 부재 처리, `:128-130` 접두사/접미사 필터, `:135-140` 날짜 parse 필터, `:142-147` retention 비교 및 삭제.

## AC-011: SessionStart 훅이 세션 시작 시 90일 retention을 강제한다

**Given** Claude Code가 SessionStart 이벤트를 발생시킨 상황에서,
**When** `internal/hook/session_start.go`의 핸들러가 실행되면,
**Then** `pruneTelemetry(projectDir)`가 호출되어 `telemetry.PruneOldFiles(projectDir, 90)`로 위임되어야 하고, 실패 시 "session start: telemetry pruning failed" 경고 로그만 남기고 세션 시작은 계속되어야 한다.

**Verification**: `internal/hook/session_start.go:15` (import telemetry), `:108-112` (pruning 호출 + 경고 로그), `:527-530` (`pruneTelemetry` wrapper).

## Edge Cases

- **EC-01**: projectRoot가 존재하지 않음 → `os.MkdirAll` 에러를 `telemetry: mkdir: %w`로 래핑하여 caller에 반환.
- **EC-02**: 세션이 자정을 넘김 → `LoadBySession`이 오늘과 어제 파일 모두 조회(`recorder.go:76-82`).
- **EC-03**: JSONL 파싱 중 손상된 라인 → 해당 라인만 skip, 다음 라인 계속 처리(`recorder.go:95-97`).
- **EC-04**: async recorder buffer 가득참 → record drop 후 warn 로그(`async_recorder.go:60`).
- **EC-05**: 모든 signal flag가 false → `OutcomeUnknown` 반환(보수적 기본값, R2 요구사항).

## Non-functional / Privacy Evidence

- **R4 context_hash 비가역성**: SHA-256 8자 truncation은 사전 공격에도 실용적으로 비가역(AC-009).
- **R4 raw prompt 미저장**: UsageRecord 스키마에 prompt 원문 필드 없음(types.go:7-17).
- **R4 90일 retention**: AC-010, AC-011.
- **R4 async write**: `internal/telemetry/async_recorder.go` — 1ms 목표 latency (`async_recorder.go` buffer pattern).

## Partial / Deferred

- **10MB 총 용량 캡 + LRU pruning (R4 마지막 항목)**: 현재 구현은 90일 날짜 기반 retention만 적용. 10MB 초과 시 LRU 프루닝 로직은 `PruneOldFiles` 범위 밖이며 아직 구현되지 않음. **PARTIAL: 90-day retention 구현됨, 10MB cap + LRU pruning은 후속 작업 필요**.
- **`.gitignore` 엔트리 (R4)**: SPEC-EVO-001 의존성이며 본 SPEC 범위 밖.
- **스킬 load duration 측정 (DurationMs 필드)**: 필드는 정의되어 있으나 실제 측정 구현은 async_recorder의 wrapper 수준에 머무름 — 정밀한 "skill load to next tool call" 측정은 후속 작업.

## Definition of Done

- [x] `internal/telemetry/` 패키지 5개 파일 구현 (types.go, recorder.go, outcome.go, report.go, async_recorder.go) (AC-001~007)
- [x] Skill tool PostToolUse 훅 통합 (AC-004)
- [x] `moai telemetry report` CLI 서브커맨드 (AC-007, AC-008)
- [x] HashContext SHA-256 8자 + PII 무저장 (AC-009)
- [x] PruneOldFiles 90일 retention + SessionStart 통합 (AC-010, AC-011)
- [x] 테스트 파일 존재: `recorder_test.go` (10KB), `outcome_test.go`, `report_test.go`, `async_recorder_test.go`, `post_tool_telemetry_test.go`
- [x] spec.md frontmatter `Status: Implemented` 표시
- [ ] **PARTIAL**: 10MB 용량 cap + LRU pruning (R4 마지막 항목) 미구현
- [ ] **PARTIAL**: 정밀 DurationMs 측정(load-to-next-tool-call) 구현 정확도 확인 필요
