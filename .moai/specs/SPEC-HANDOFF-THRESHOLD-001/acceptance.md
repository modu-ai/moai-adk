# Acceptance — SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4/4)

> 각 AC는 binary pass/fail, REQ와 1:1, Given-When-Then. Test 참조는 package-relative. 구현(run-phase) 시 이 AC가 완료 게이트. **AC-THRESHOLD-006 = "no M1 statusline regression" 불변식.**

## §D — Acceptance Criteria Matrix

| AC | REQ | Milestone | 검증 방식 |
|----|-----|-----------|-----------|
| AC-THRESHOLD-001 | REQ-001 | M1 | unit (stage enum + wrapper) |
| AC-THRESHOLD-002 | REQ-002 | M1 | unit (2-stage render string) |
| AC-THRESHOLD-003 | REQ-003 | M1 | unit (하드 상한 공식 + clamp) |
| AC-THRESHOLD-004 | REQ-004 | M1 | unit + grep (config 상수, inline literal 부재) |
| AC-THRESHOLD-005 | REQ-005 | M3 | grep (reachability 문서) |
| AC-THRESHOLD-006 | REQ-006 | M1 | unit (INVARIANT — M1 무회귀) |
| AC-THRESHOLD-007 | REQ-007 | M2 | unit (write config 무관) |
| AC-THRESHOLD-008 | REQ-008 | M3 | grep + review (Guide/Mode 화해 문서) |
| AC-THRESHOLD-009 | REQ-009 | M2 | unit (atomic best-effort write) |
| AC-THRESHOLD-010 | REQ-010 | M2 | unit (schema 필드) |
| AC-THRESHOLD-011 | REQ-011 | M2 | unit (Build 호출부 session_id+Memory) |
| AC-THRESHOLD-012 | REQ-012 | M2 | unit (write-if-changed throttle) |
| AC-THRESHOLD-013 | REQ-013 | M2 | unit (session_id guard stale) |
| AC-THRESHOLD-014 | REQ-014 | M2 | unit (fallback-UUID freshness) |
| AC-THRESHOLD-015 | REQ-015 | M3 | grep (state-file-first 재작성) |
| AC-THRESHOLD-016 | REQ-016 | M3 | grep (LIVE section-level + full-sync 금지) |
| AC-THRESHOLD-017 | REQ-017 | M3 | grep (template mirror 256K parity — D1 drift) |
| AC-THRESHOLD-018 | REQ-018 | M2 | unit (concurrent empty-id discriminator) |

---

### AC-THRESHOLD-001 → REQ-001 — stage enum + M1 wrapper
**GIVEN** `handoffGuideStage`가 도입되고 `shouldShowHandoffGuide`가 wrapper로 유지됨
**WHEN** 동일 `StatusData`(nil / cwSize≤0 / 각 밴드)에 대해 두 함수를 호출
**THEN** `shouldShowHandoffGuide(data) == (handoffGuideStage(data) != handoffStageNone)` 이 모든 케이스에서 성립 AND 기존 `TestShouldShowHandoffGuide_*`(M1) 전부 통과(무손상)
**Test**: `internal/statusline/renderer_test.go::TestHandoffGuideStage_WrapperEquivalence` + 기존 M1 함수

### AC-THRESHOLD-002 → REQ-002 — 2단계 렌더 문자열
**GIVEN** `renderBarsInline`이 stage로 분기
**WHEN** stage가 각각 none / soft / hard
**THEN** none → suffix 없음 AND soft → CW 바에 `" (⚠️/clear)"`(M1 verbatim) AND hard → `" (🛑/clear!)"`(구별 마커, soft와 다름)
**Test**: `internal/statusline/renderer_test.go::TestRenderBarsInline_TwoStageSuffix`

### AC-THRESHOLD-003 → REQ-003 — 하드 상한 공식 + clamp
**GIVEN** `getAutoCompactThreshold()` 기본 85, `HandoffHardCeilingCapPct=95`, `HandoffHardCeilingMarginPct=10`
**WHEN** (a) 기본 config로 `<500K` 밴드(예 256K) rawPct=95 / 90 / 89, (b) env `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=60`으로 하드 계산(60+10=70 < soft 90 → clamp)
**THEN** (a) rawPct≥95 → hard, 90≤rawPct<95 → soft, <90 → none (하드=min(95,95)=95); (b) clamp로 hard=soft=90 → rawPct≥90 → hard(soft 창 흡수), 저하 안전
**Test**: `internal/statusline/renderer_test.go::TestHardCeiling_Formula` + `TestHardCeiling_ClampBelowSoft`(t.Setenv override)

### AC-THRESHOLD-004 → REQ-004 — 밴드 경계 config 상수
**GIVEN** `internal/config/defaults.go`에 5개 명명 상수 추가
**WHEN** (a) 상수 값 검사, (b) renderer.go grep
**THEN** (a) `HandoffSoftLargePct==50` ∧ `HandoffSoftStandardPct==90` ∧ `HandoffLargeWindowCutoff==500000` ∧ `HandoffHardCeilingCapPct==95` ∧ `HandoffHardCeilingMarginPct==10`; (b) `grep -nE '\b(50\.0|90\.0|500_000)\b' internal/statusline/renderer.go` 가 밴드 판정 함수 내에서 0 매치(상수 참조만)
**Test**: `internal/config/defaults_test.go::TestHandoffThresholdConstants` + grep guard

### AC-THRESHOLD-005 → REQ-005 — reachability 한계 문서
**GIVEN** 독트린(context-window-management.md) 재작성 + CHANGELOG 엔트리
**WHEN** grep
**THEN** 독트린이 "auto-compact가 하드 상한을 자주 선점(pre-empt)한다 / stage-2는 드물게 발화" 취지 문구 포함 AND "stage-2 always fires" 류 주장 부재(verification-claim-integrity)
**Test**: `grep -iE 'auto-compact.*pre-empt|rarely fires|frequently pre-empted' <doctrine>` ≥1 AND `grep -i 'always fires' <doctrine>` == 0

### AC-THRESHOLD-006 → REQ-006 — INVARIANT: no M1 statusline regression
**GIVEN** `HandoffConfig` 기본값(Mode=="manual", Guide==false)
**WHEN** `<500K` 밴드에서 rawPct=90(soft 조건)로 `renderBarsInline` 렌더 (config 주입 없이)
**THEN** CW 바에 `" (⚠️/clear)"` soft suffix가 **여전히 렌더됨**(default guide=false에서 소멸 안 함) AND `handoffGuideStage`/`renderBarsInline` 시그니처가 `HandoffConfig`를 파라미터로 받지 않음(statusline이 config 미소비) AND `grep -n 'HandoffConfig\|Handoff\b' internal/statusline/renderer.go` 가 suffix 게이팅 목적 0 매치
**Test**: `internal/statusline/renderer_test.go::TestNoM1Regression_SuffixUnconditional` + static grep

### AC-THRESHOLD-007 → REQ-007 — state-file write config 무관
**GIVEN** `writeContextUsage`가 `HandoffConfig`를 파라미터로 받지 않음
**WHEN** (a) 시그니처 검사, (b) Build 렌더 시 (config 어떤 값이든) 파일 생성
**THEN** write는 Mode/Guide와 무관하게 발생(Memory.Available==true인 한) AND `writeContextUsage` 시그니처에 `HandoffConfig`/`cfg` 파라미터 없음
**Test**: `internal/statusline/context_usage_test.go::TestWriteContextUsage_ConfigIndependent`

### AC-THRESHOLD-008 → REQ-008 — Guide/Mode 소비 경계 화해
**GIVEN** design.md §C.2 화해 표 + D4 독트린
**WHEN** review + grep
**THEN** design.md가 "Mode → M3 auto-resume / Guide → advisory / 둘 다 statusline 미게이팅 / statusline은 HandoffConfig 미소비"를 명시 AND M3 `handoff_inject.go`가 무변경(`git diff internal/hook/handoff_inject.go` empty)
**Test**: `grep -c 'statusline.*미소비\|does not read HandoffConfig' design.md` ≥1 + `git diff --stat internal/hook/handoff_inject.go` empty

### AC-THRESHOLD-009 → REQ-009 — atomic best-effort write
**GIVEN** `writeContextUsage(projDir, sid, mem, stage)` in `t.TempDir()`
**WHEN** 정상 호출 / 쓰기 불가 디렉터리(권한) 주입
**THEN** 정상 → `<projDir>/.moai/state/context-usage.json` valid JSON 존재(temp+rename atomic) AND 실패 → error 미반환(silent), statusline 렌더 계속(패닉/에러 전파 없음)
**Test**: `internal/statusline/context_usage_test.go::TestWriteContextUsage_Atomic` + `TestWriteContextUsage_SilentFail`

### AC-THRESHOLD-010 → REQ-010 — state-file 스키마
**GIVEN** write 완료된 context-usage.json
**WHEN** 파싱
**THEN** `schema_version` ∧ `session_id` ∧ `writer_pid` ∧ `captured_at` ∧ `context_window_size` ∧ `tokens_used` ∧ `raw_pct` ∧ `stage`(none|soft|hard) ∧ `band`(large|standard) 필드 존재
**Test**: `internal/statusline/context_usage_test.go::TestContextUsage_Schema`

### AC-THRESHOLD-011 → REQ-011 — 호출부 배치 (session_id + Memory)
**GIVEN** `builder.Build`가 `collectAll` 이후 `writeContextUsage`를 호출
**WHEN** stdin에 `session_id` + `context_window` 제공하여 Build
**THEN** 생성된 파일의 `session_id`가 stdin `session_id`와 일치 AND `context_window_size`/`tokens_used`가 `data.Memory`와 일치 (즉 두 신호가 동일 스코프에서 캡처됨) AND 호출이 collectAll/renderer 내부가 아닌 Build에 위치(`grep -n 'writeContextUsage' internal/statusline/builder.go` ≥1, `internal/statusline/renderer.go` == 0)
**Test**: `internal/statusline/builder_test.go::TestBuild_WritesContextUsageWithSessionID`

### AC-THRESHOLD-012 → REQ-012 — write-if-changed throttle
**GIVEN** 동일 semantic payload(session_id/stage/int(raw_pct)/cwSize)로 2회 write
**WHEN** 2번째 write 시도
**THEN** 파일 mtime(또는 write 호출 카운터) 불변(2번째 skip) AND payload 변경(raw_pct 정수 달라짐) 시에는 write 발생(captured_at 갱신)
**Test**: `internal/statusline/context_usage_test.go::TestWriteContextUsage_ThrottleSkipUnchanged`

### AC-THRESHOLD-013 → REQ-013 — session_id guard stale
**GIVEN** context-usage.json이 session_id `A`로 스탬프됨
**WHEN** reader 유효성 판정 로직(`isFreshForSession(rec, current)`)을 current==`B`로 호출
**THEN** stale 판정(false — session_id 불일치) → 휴리스틱 폴백 신호; current==`A`이면 valid(true)
**Test**: `internal/statusline/context_usage_test.go::TestSessionGuard_MismatchStale`

### AC-THRESHOLD-014 → REQ-014 — fallback-UUID freshness (single-session 경로 생존)
**GIVEN** context-usage.json의 `session_id == ""`(environment-fallback)
**WHEN** (a) write는 여전히 발생, (b) 독트린-layer reader 유효성 판정을 current session_id==""(양측 부재) + captured_at 신선/오래됨 각각으로 호출
**THEN** (a) session_id 부재여도 파일 생성됨; (b) 양측 "" + captured_at 신선 → valid(single-session 공통 경로 생존, primary path 미사망), captured_at 만료 → stale; UUID X vs "" 혼합 → 보수적 stale
**Note**: concurrent same-checkout empty-id 세션의 cross-read 차단은 AC-018(`writer_pid` discriminator)이 담당 — 본 AC는 single-session 경로 생존만 검증.
**Test**: `internal/statusline/context_usage_test.go::TestFallbackUUID_FreshnessValidation`

### AC-THRESHOLD-015 → REQ-015 — Detection state-file-first 재작성
**GIVEN** `context-window-management.md` § Detection Heuristics 재작성
**WHEN** grep
**THEN** § Detection Heuristics가 `.moai/state/context-usage.json`를 우선 읽는다고 서술 AND "부재/stale/파싱실패 시 기존 휴리스틱 폴백" 명시 AND 기존 4-신호 휴리스틱이 폴백으로 보존됨
**Test**: `grep -c 'context-usage.json' <doctrine>` ≥1 AND `grep -iE 'fall back|폴백|absent|stale' <doctrine 내 Detection 절>` ≥1

### AC-THRESHOLD-016 → REQ-016 — Template-First + section-level edit (full-sync 금지)
**GIVEN** 독트린 편집
**WHEN** template mirror + LIVE grep
**THEN** LIVE copy(`.claude/rules/moai/workflow/context-window-management.md`)의 § Detection Heuristics가 `context-usage.json`를 언급(state-file-first 반영) AND template mirror(`internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`)의 § Detection Heuristics도 동일 반영(section-level) AND LIVE § Context Window Targets의 `grep -c '256,000'` == 1(중복 행 미추가 — Detection 절만 편집, Targets 표 무접촉) AND template 사본 `grep -E 'SPEC-|REQ-' <template doctrine>` == 0(§25 중립)
**Test**: LIVE·template 각각 Detection 절 grep(`context-usage.json`) + LIVE 256K count==1 + template 중립성 grep
**Note**: `<doctrine>`는 LIVE copy를 명시 지칭. full-file template→live overwrite / `moai update` full sync 금지(LIVE 256K 행 삭제 회귀).

### AC-THRESHOLD-017 → REQ-017 — template mirror 256K parity (D1 drift 정정)
**GIVEN** D4 완료(template mirror에 256K Targets 행 추가)
**WHEN** template mirror + LIVE 병렬 grep
**THEN** `grep -c '256,000' internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` == 1 (M1이 누락한 행 추가됨) AND `grep -c '256,000' .claude/rules/moai/workflow/context-window-management.md` == 1 (기존 유지) → 두 사본 parity(drift 제거) AND `make build` 성공
**Test**: 병렬 256K count(template==1 ∧ live==1) + `make build` exit 0
**Note**: plan-auditor iter-1 D1 SHOULD-FIX. 사전 상태(정정 전): template==0, live==1(drift). 정정 후: 둘 다 1.

### AC-THRESHOLD-018 → REQ-018 — concurrent empty-`session_id` discriminator
**GIVEN** `writer_pid`가 다른 2개의 empty-`session_id` record(recA.writer_pid=1001, recB.writer_pid=1002; 둘 다 captured_at 신선)
**WHEN** Go freshness 헬퍼 `isFreshForSession(rec, curSession="", curWriterID=1001)` 호출
**THEN** recA(writer_pid==1001 일치) → valid AND recB(writer_pid==1002 불일치) → **stale**(cross-read 차단) AND session_id==UUID 경로는 writer_pid 무관하게 기존 동작 유지(AC-013)
**Test**: `internal/statusline/context_usage_test.go::TestConcurrentEmptyID_WriterPIDGuard` (table: writer_pid match/mismatch × fresh/stale)
**Note**: plan-auditor iter-1 D2 SHOULD-FIX. B4 empty-id fallback이 재개방한 cross-session hole을 Go 헬퍼 레벨에서 기계적으로 닫음. 독트린-only reader의 잔여 한계는 §D.4 residual + design.md §D.4b 참조.

---

## §D.1 Edge Cases (명시적 처리)

- **Memory.Available==false**: `writeContextUsage` skip(원신호 부재) — write 안 함, 렌더 정상.
- **projectDir 유도 실패**(빈 CWD + Getwd 실패): write skip, 렌더 정상.
- **context-usage.json 손상 JSON**(reader): 파싱 실패 → stale 취급 → 휴리스틱 폴백.
- **getAutoCompactThreshold env 이상값**: memory.go:41이 1..100 검증, 벗어나면 85 폴백 → hard 공식 안전.
- **cwSize 정확히 500,000**: `>= 500_000` → large 밴드(soft 50) — M1 경계 verbatim.
- **동시 세션 동일 checkout write**: last-writer-wins(guard); reader session_id 불일치로 오재개 방지(AC-013).

## §D.2 Quality Gate (완료 게이트)

- `go test ./internal/statusline/... ./internal/config/...` 전부 통과
- `go test -race ./internal/statusline/...`(동시 write/throttle)
- `go vet ./...` + `golangci-lint run` 0 error(NEW)
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `make build` 성공(template embed — D4)
- `moai spec lint spec.md` / `plan.md` No findings
- coverage: 신규 파일(context_usage.go, renderer.go 신규 함수) ≥85%(critical statusline ≥90% 지향)
- M1 무회귀: 기존 `TestShouldShowHandoffGuide_*` + statusline 렌더 테스트 전부 green

## §D.3 Definition of Done

- [ ] AC-THRESHOLD-001..018 전부 pass (18 AC ↔ 18 REQ 1:1)
- [ ] M1/M2/M3 각 독립 커밋(RED 커밋 금지 — atomic GREEN, cycle_type=tdd)
- [ ] **M1 무회귀 불변식(AC-006) PASS** — default guide=false에서 soft suffix 유지
- [ ] 밴드 경계 config 상수화(AC-004), renderer.go inline literal 부재
- [ ] write 무조건(AC-007) + 호출부 Build(AC-011)
- [ ] session_id guard(AC-013) + fallback-UUID(AC-014) + concurrent empty-id discriminator(AC-018) 처리
- [ ] Detection state-file-first(AC-015) + section-level edit BOTH files, full-sync 금지(AC-016)
- [ ] **template mirror 256K parity(AC-017)** — D1 drift 정정, template+live 둘 다 256K==1
- [ ] reachability 한계 문서(AC-005) — "always fires" 미주장
- [ ] context-usage.json은 런타임 생성물 — 커밋 미포함(코드만)
- [ ] progress.md §E.2/§E.3 run-phase evidence 채움(manager-develop)

## §D.4 Residual Risk (잔여 위험)

- **throttle plateau**: usage 정체 시 write skip → captured_at 미갱신 → reader freshness 창이 짧으면 stale 오판. 완화: freshness 창 관대(session-scoped). 관측 필요.
- **stage-2 reachability**: auto-compact 85%가 raw에서 하드 상한(95%)을 자주 선점 → stage-2 드물게 발화. 의도된 tradeoff, 문서화(AC-005). 향후 관측으로 margin 재튜닝 가능.
- **reader는 Go 파서 아님**: state-file 읽기는 독트린 서술(orchestrator 행동). 런타임 자동 파서는 후속 SPEC. AC-013/014/018의 reader 로직은 유효성 헬퍼(`isFreshForSession`)로 테스트하되, 실제 소비는 독트린 경유.
- **concurrent empty-id hole (plan-auditor iter-1 D2)**: B4 empty-id fallback(AC-014)은 B3 session_id guard(AC-013)의 cross-session 보호를 **UUID 없는 2+ concurrent same-checkout 세션에 한해 약화**시킨다 — freshness 창 내에서 session B가 session A의 신선 스냅샷을 오독 가능. 정정: `writer_pid` discriminator(AC-018)가 Go 헬퍼 레벨에서 기계적으로 cross-read를 차단. **잔여**: 독트린-only reader는 `curWriterID`를 공급할 수 없어(Claude가 JSON을 Read tool로 읽음, PID 핸들 없음) concurrent empty-id 케이스가 doctrine-layer 기계적 보증 밖 → 보수적 휴리스틱 폴백. 게다가 statusline은 render마다 fresh 프로세스라 `writer_pid`가 완전 session-stable하지 않음(design §D.2 caveat) → single-session 가정 하에서만 primary path 생존 보장. 후속 Go reader(D3 파서 SPEC)가 session-stable 토큰으로 완전 폐쇄. trigger 조건(2+ concurrent same-checkout empty-id + freshness 창)이 드물어 Tier M 비례 대응으로 수용.

## §D.5 Cross-References

- spec.md §C(REQ), plan.md §F(blocker)/§G(milestone), design.md §B(하드 상한)/§C(화해)/§D(state-file)/§E(독트린), research.md §A~E(실측)
