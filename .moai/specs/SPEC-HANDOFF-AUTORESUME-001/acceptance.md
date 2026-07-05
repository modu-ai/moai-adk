# Acceptance — SPEC-HANDOFF-AUTORESUME-001 (Handoff-v2 M3/4, auto-resume)

> 각 AC는 binary pass/fail, REQ와 1:1, Given-When-Then. Test 참조는 package-relative. 구현(run-phase) 시 이 AC가 완료 게이트.

## §D — Acceptance Criteria Matrix

| AC | REQ | Milestone | 검증 방식 |
|----|-----|-----------|-----------|
| AC-AUTORESUME-001 | REQ-001 | M1 | unit (default) |
| AC-AUTORESUME-002 | REQ-002 | M1 | audit parity |
| AC-AUTORESUME-003 | REQ-003 | M1 | unit (partial-override) |
| AC-AUTORESUME-004 | REQ-004 | M1 | template neutrality grep |
| AC-AUTORESUME-005 | REQ-005 | M2 | unit + path isolation grep |
| AC-AUTORESUME-006 | REQ-006 | M2 | unit (schema) |
| AC-AUTORESUME-007 | REQ-007 | M2 | unit (clear CLI only) |
| AC-AUTORESUME-008 | REQ-008 | M3 | branch table unit (auto 4-cell) |
| AC-AUTORESUME-009 | REQ-009 | M3 | unit (manual 4-cell no-op + stale) |
| AC-AUTORESUME-010 | REQ-010 | M3 | unit (notice-only, guide∈{true,false}) |
| AC-AUTORESUME-011 | REQ-011 | M3 | unit (degrade string) |
| AC-AUTORESUME-012 | REQ-012 | M3 | unit (rename + audit preserve) |
| AC-AUTORESUME-013 | REQ-013 | M3 | unit (rename-fail → fail-open) |
| AC-AUTORESUME-014 | REQ-014 | M3 | unit (nonce filename shape) |
| AC-AUTORESUME-015 | REQ-015 | M3 | unit (i18n) |
| AC-AUTORESUME-016 | REQ-016 | M3 | static grep |
| AC-AUTORESUME-017 | REQ-017 | M3 | unit (fail-open) |
| AC-AUTORESUME-018 | REQ-018 | M3 | e2e (3-handler coexist) |
| AC-AUTORESUME-019 | REQ-019 | M3 | unit (auto-mode TTL cleanup) |

---

### AC-AUTORESUME-001 → REQ-001 — HandoffConfig manual default
**GIVEN** 설정 없이 `NewDefaultHandoffConfig()` 호출
**WHEN** 반환 struct 검사
**THEN** `Mode == "manual"` AND `Guide == false` AND struct에 `Consume` 필드가 존재하지 않음 (YAGNI 제거)
**Test**: `internal/config/handoff_test.go::TestNewDefaultHandoffConfig`

### AC-AUTORESUME-002 → REQ-002 — audit parity 바인딩
**GIVEN** `yamlToStructRegistry`에 `handoff` 등록 + live/template에 `handoff.yaml` 존재
**WHEN** `TestAuditParity` 실행
**THEN** `handoff.yaml`이 orphan으로 보고되지 않음, `IsRegisteredOrException("handoff") == true`
**Test**: `internal/config/audit_test.go::TestAuditParity` (기존) + `TestHandoffRegistered`

### AC-AUTORESUME-003 → REQ-003 — partial-override 로드
**GIVEN** `handoff.yaml`이 `mode: auto`만 명시(guide 생략)
**WHEN** `Loader.Load`로 로드
**THEN** `cfg.Handoff.Mode == "auto"` AND `cfg.Handoff.Guide == false` (생략 키 default 유지, zero-value 붕괴 없음)
**Test**: `internal/config/handoff_test.go::TestLoadHandoffSection_PartialOverride`

### AC-AUTORESUME-004 → REQ-004 — 중립 template
**GIVEN** `internal/template/templates/.moai/config/sections/handoff.yaml`
**WHEN** 내용 grep
**THEN** SPEC ID 토큰(`SPEC-`), REQ 토큰(`REQ-`), 내부 날짜/commit SHA 없음; `mode`/`guide` 키만 존재 (`consume` 키 없음 — D3 필드 제거로 unbound YAML 키 금지)
**Test**: `internal/template/internal_content_leak_test.go` (기존 CI guard 통과) + `grep -E 'SPEC-|REQ-' handoff.yaml` = 0 + `grep -c 'consume' handoff.yaml` = 0

### AC-AUTORESUME-005 → REQ-005 — save 별도 경로 atomic
**GIVEN** `moai handoff save --body "<resume>"` in `<projectDir>`
**WHEN** 명령 완료
**THEN** `<projectDir>/.moai/state/handoff/pending.json` 존재 (valid JSON) AND `<projectDir>/.moai/state/session-handoff/pending.md` **미생성/미변경**
**Test**: `internal/cli/handoff_test.go::TestHandoffSave_WritesJSONNotMarkdown` (decoy: 기존 session-handoff/pending.md mtime 불변 assert)

### AC-AUTORESUME-006 → REQ-006 — pending.json 스키마
**GIVEN** `moai handoff save --body "<resume>" --ultrathink`
**WHEN** pending.json 파싱
**THEN** `schema_version`, `body`(verbatim resume), `directives.ultrathink == true`, `conversation_language`, `saved_at` 필드 존재
**Test**: `internal/cli/handoff_test.go::TestHandoffSave_Schema`

### AC-AUTORESUME-007 → REQ-007 — clear CLI (M2, TTL 분리)
**GIVEN** pending.json 존재 + `moai handoff clear` 실행
**WHEN** clear 완료
**THEN** `handoff/pending.json` 제거됨 AND `session-handoff/pending.md` 미접촉(decoy mtime 불변)
**Test**: `internal/cli/handoff_test.go::TestHandoffClear` (TTL cleanup은 AC-019로 분리 — milestone 경계 정합)

### AC-AUTORESUME-008 → REQ-008 — 유일 소비 셀 (branch table)
**GIVEN** pending.json 존재 + `mode == "auto"`
**WHEN** SessionStart가 각 source(startup/resume/clear/compact)로 발화
**THEN** `source == "clear"`에서만 additionalContext에 handoff 주입 AND pending.json이 consumed/로 rename됨; 나머지 3 source에서는 pending.json 보존(주입 없음)
**Test**: `internal/hook/handoff_inject_test.go::TestBranchTable_AutoMode` (table-driven 4 source)

### AC-AUTORESUME-009 → REQ-009 — manual mode pure no-op (4-cell + stale)
**GIVEN** pending.json 존재 + `mode == "manual"`
**WHEN** SessionStart가 각 source(startup/resume/clear/compact)로 발화 — 그리고 별도 sub-case로 pending.json이 TTL 초과(stale)인 경우
**THEN** 4개 source 모두에서 주입 없음 AND pending.json 바이트 불변(pure no-op); **stale sub-case**: manual mode는 stale pending도 제거/rename하지 않음(바이트 불변) — REQ-019 auto-only TTL과의 모순 해소 잠금
**Test**: `internal/hook/handoff_inject_test.go::TestManualMode_NoOp` (table-driven 4 source) + `TestManualMode_StalePendingPreserved`

### AC-AUTORESUME-010 → REQ-010 — non-clear notice-only (guide 양분기)
**GIVEN** **live(non-stale)** pending.json 존재 + `mode == "auto"`, source∈{startup,resume,compact} (stale pending의 우선순위는 AC-019 소관 — N1 precedence)
**WHEN** (a) `guide == true`로 발화 / (b) `guide == false`로 발화
**THEN** 두 경우 모두 pending.json 미소비(보존) AND additionalContext에 handoff body 미주입; (a) `guide==true` → stderr에 힌트 1건 방출 / (b) `guide==false` → stderr 힌트 방출 안 함(hint suppression)
**Test**: `internal/hook/handoff_inject_test.go::TestNonClearSource_NoticeOnly` (table-driven source × guide∈{true,false}, live pending)

### AC-AUTORESUME-011 → REQ-011 — degrade-to-guidance
**GIVEN** pending.json `directives.ultrathink == true`, mode==auto, source==clear
**WHEN** 주입 콘텐츠 렌더
**THEN** additionalContext는 "ultrathink가 활성/xhigh 적용됨" 문구를 **포함하지 않음** AND "복원하려면 `ultrathink` 입력" 형태의 안내 문구를 포함
**Test**: `internal/hook/handoff_inject_test.go::TestDegradeToGuidance` (부정 assert: `NotContains("xhigh")`, 긍정 assert: 안내 문구)

### AC-AUTORESUME-012 → REQ-012 — claim-then-inject + audit preserve
**GIVEN** pending.json 존재, mode==auto, source==clear, memory dir에 `project_*.md` audit 존재
**WHEN** 소비 완료
**THEN** `handoff/pending.json` 부재 AND `handoff/consumed/<ts>-<nonce>.json` 존재(내용 == 원 pending) AND memory `project_*.md` 및 consumed/ 삭제 안 됨 AND 주입은 rename 성공 후 발생
**Test**: `internal/hook/handoff_inject_test.go::TestClaimThenInject_AuditPreserved`

### AC-AUTORESUME-013 → REQ-013 — rename 실패 → fail-open (errno 무관, cross-platform)
**GIVEN** (a) 동일 pending.json에 2 goroutine이 동시 consume 시도, (b) rename이 임의 원인으로 실패(injected error — ENOENT 아닌 경우 포함)
**WHEN** rename 시도
**THEN** (a) 정확히 1개 goroutine만 주입+consumed 파일 1개 생성, 나머지는 주입 생략 + 에러 없이 정상 반환; (b) rename이 어떤 errno로 실패하든 핸들러는 주입을 생략하고 정상 반환(특정 `os.IsNotExist`에 의존하지 않음 — Windows `MoveFileEx` 호환)
**Test**: `internal/hook/handoff_inject_test.go::TestConcurrentConsume_SingleWinner` (`go test -race`) + `TestRenameFailure_FailOpen` (rename 함수 주입으로 임의 errno 강제)

### AC-AUTORESUME-014 → REQ-014 — NULL session_id nonce (filename shape)
**GIVEN** pending.json `saved_by_session == ""`, mode==auto, source==clear
**WHEN** 소비 완료
**THEN** consumed 파일명이 정규식 `^\d+-[0-9a-f]{8}\.json$` (즉 `<ts>-<8hex>.json`, session8 미의존)에 매칭됨 — binary shape assertion
**Test**: `internal/hook/handoff_inject_test.go::TestNonceFallback_FilenameShape` (regexp.MatchString)
**Note**: 충돌 안전성(cross-session collision 도달 불가)은 atomic-rename-as-claim 논증으로 design.md §C.4 prose에 문서화 — runtime AC가 아닌 설계 불변식.

### AC-AUTORESUME-015 → REQ-015 — i18n header
**GIVEN** `conversation_language == "ko"` (그리고 별도 케이스 en/ja/zh/fr)
**WHEN** 주입 header 렌더
**THEN** header가 해당 언어로 렌더; `fr`(테이블 외) → en fallback
**Test**: `internal/hook/handoff_inject_test.go::TestInjectionHeader_I18n` (table-driven locale)

### AC-AUTORESUME-016 → REQ-016 — 서브에이전트 경계
**GIVEN** 신규 hook 코드 배치 후
**WHEN** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "^[^:]*:[0-9]*:[ \t]*//"`
**THEN** exit code 1 (0 matches)
**Test**: static guard (CI) + `internal/hook/handoff_inject_test.go::TestNoUserInteraction`

### AC-AUTORESUME-017 → REQ-017 — fail-open
**GIVEN** pending.json이 손상된 JSON, mode==auto, source==clear
**WHEN** 핸들러 실행
**THEN** SessionStart 훅이 allow 반환(block 안 함) AND `slog.Warn("session_start: handoff: ...")` 1건 AND pending.json 보존(rename 안 함)
**Test**: `internal/hook/handoff_inject_test.go::TestFailOpen_CorruptPending`

### AC-AUTORESUME-018 → REQ-018 — additionalContext 공존
**GIVEN** sessionStartHandler + autoUpdateHandler + handoffInjectHandler 3개 등록, mode==auto, source==clear, pending 존재
**WHEN** registry.Dispatch(EventSessionStart)
**THEN** merged additionalContext가 sessionStartHandler의 attribution/guardrail 콘텐츠 AND handoffInjectHandler의 handoff 안내를 **모두** 포함(`\n`-join, 드롭 없음)
**Test**: `internal/hook/handoff_inject_test.go::TestThreeHandlerCoexist` (registry 통합)

### AC-AUTORESUME-019 → REQ-019 — auto-mode stale TTL cleanup (M3, N1 precedence)
**GIVEN** pending.json의 `saved_at`이 TTL 초과(stale) + `mode == "auto"` (source∈{startup,resume,compact} + `guide==true` 하위 케이스 포함)
**WHEN** SessionStart consume-eligibility 체크
**THEN** stale pending.json이 조용히 제거됨(주입 없음, best-effort) AND consumed/로 rename하지 않음(정상 소비 아님, 단순 청소) AND **notice-only 힌트 생략**(N1: REQ-019 cleanup이 REQ-010 hint보다 우선 — `guide==true`여도 힌트 방출 안 함, 재개할 live 컨텍스트 없음); 대조군으로 `mode == "manual"` + 동일 stale → pending.json 바이트 불변(AC-009 stale sub-case와 정합)
**Test**: `internal/hook/handoff_inject_test.go::TestStaleTTLCleanup_AutoOnly` (+ stale ∧ guide==true → no-hint sub-case)

---

## §D.1 Edge Cases (명시적 처리)

- **pending.json 부재 + source==clear + mode==auto**: no-op, slog 없음(정상).
- **consumed/ 디렉터리 부재**: rename 전 `os.MkdirAll(handoff/consumed, 0o700)` best-effort. 실패 시 주입 생략 + slog.Warn.
- **body가 64 KiB 근처**: 3-handler 합산이 64 KiB 초과 시 `ValidateHookResponse` 절단 + SystemMessage notice. 훅 계속 진행(best-effort). (선택 AC — plan-auditor 판단.)
- **mode 값이 "auto"/"manual" 외**(오타): unknown → manual로 안전 처리(no-op) + slog.Warn 1건.
- **worktree cwd**: `resolveProjectDir(input)`가 CWD 우선 → worktree 내 `handoff/` 정확 해석.
- **동일 세션 내 재-clear**: 첫 clear가 소비(pending 제거) → 둘째 clear는 pending 부재 → no-op.

## §D.2 Quality Gate (완료 게이트)

- `go test ./internal/config/... ./internal/cli/... ./internal/hook/...` 전부 통과
- `go test -race ./internal/hook/...` (AC-013)
- `go vet ./...` + `golangci-lint run` 0 error
- `moai spec lint spec.md` / `plan.md` No findings
- 경계 grep(AC-016) 0 matches
- `make build` 성공 (template embed)
- coverage: 신규 패키지 ≥ 85% (critical hook ≥ 90% 지향)

## §D.3 Definition of Done

- [ ] AC-AUTORESUME-001..019 전부 pass (19 AC ↔ 19 REQ 1:1)
- [ ] M1/M2/M3 각 독립 커밋 (RED 커밋 금지 — atomic GREEN)
- [ ] `session-handoff/pending.md` 무접촉 확인 (경로 격리 grep)
- [ ] settings.json 무변경 확인 (matcher 이미 clear 포함, assertion only)
- [ ] verification-claim-integrity 준수 (주입 콘텐츠 xhigh 미주장)
- [ ] 3개 SessionStart 핸들러 공존 e2e 통과
- [ ] progress.md §E.2/§E.3 run-phase evidence 채움(manager-develop)

## §D.4 Cross-References

- spec.md §C (REQ), design.md §C.2 branch table / §C.4 nonce / §E config, research.md §C/§D 실측
