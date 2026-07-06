---
id: SPEC-HANDOFF-THRESHOLD-001
title: "핸드오프 임계 완성 — 구현 계획 (Tier M)"
version: "0.1.0"
status: draft
created: 2026-07-06
updated: 2026-07-06
author: MoAI
priority: P2
phase: "v3.0.0"
module: "internal/statusline, internal/config"
lifecycle: spec-anchored
tags: "statusline, handoff-guide, context-window, threshold, two-stage, state-file, tier-m, epic-handoff-v2"
tier: M
era: V3R6
related_specs: [SPEC-HANDOFF-CTXGUIDE-001, SPEC-HANDOFF-MSGMODE-001, SPEC-HANDOFF-AUTORESUME-001]
---

# 구현 계획 — SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4/4, Tier M)

> 우선순위 라벨만 사용(시간 추정 금지). 2개 LOCKED 결정 FIXED. 각 milestone 독립 커밋 단위 + AC 바인딩. 실측 근거 research.md, 설계 design.md.

## §A — Context

M1이 M4로 이연한 4개 표면(2단계 suffix / HandoffConfig 소비 / context-usage.json 영속화 / Detection 독트린)을 완성한다. M3 landing(`HandoffConfig`)을 소비하되 statusline은 config 무관(M1 무회귀 불변식). 개발 방식: `quality.yaml` `development_mode` 따름 — Go 신규 코드(stage enum + writer + config 상수) → **TDD 권장**(RED→GREEN→REFACTOR).

## §B — Known Issues (본 SPEC 적용분)

- **B4 frontmatter 12-field**: spec.md canonical 이름 준수(snake_case alias 금지). 검증 `moai spec lint spec.md`.
- **B6 Out-of-Scope h3**: spec.md §B.2 + 본 plan.md §G에 `### Out of Scope — <topic>` h3 + `-` bullet. h2 단독 금지(MissingExclusions).
- **B8 working-tree hygiene**: runtime-managed(`.moai/state/*`, `.moai/cache/*`) 무접촉. context-usage.json은 **런타임 생성물**이므로 커밋 금지(코드만 커밋). 무관 untracked 커밋 금지(`git add` specific path).
- **B10 PRESERVE**: M1 `shouldShowHandoffGuide` 밴드 로직 verbatim 승계(재설계 아님). M3 `handoff_inject.go`/`handoff/pending.json` 흐름 무접촉.
- **§14 하드코딩**: 밴드 경계 → `config/defaults.go` 명명 상수. renderer.go inline 리터럴 금지.
- **§2 Template-First**: D4 독트린은 template mirror 먼저 편집 → `make build` → live sync. **task 전제("template 밖")는 오기 — mirror 존재(research §E).**
- **§25 template 중립성**: template 독트린 사본에 SPEC-ID/REQ 토큰/내부 날짜/commit SHA 금지.

## §C — Pre-flight (실측 완료, plan-phase)

- `grep 'handoffGuideStage\|writeContextUsage\|HandoffSoftLargePct' internal/` = 0 (greenfield in existing pkg).
- `getAutoCompactThreshold` = statusline 패키지 로컬(memory.go:39) → 직접 호출(research §A).
- `builder.Build`(138행) session_id + Memory 동시 스코프(research §C).
- `WriteModelCache` atomic 선례(model_cache.go) 확인(research §D).
- Detection 독트린 template mirror 존재 + 256K 행 기존재(research §E).
- renderer.go config 미import → D1 상수 참조 위해 import 추가 필요.

## §D — Constraints (위반 금지)

- Tier M artifact: spec/plan/acceptance (+ design/research/progress skeleton — task 명시).
- GEARS REQ (REQ-THRESHOLD-NNN) + AC (AC-THRESHOLD-NNN), 1:1 이상.
- 2 LOCKED FIXED: 기존 HandoffConfig 필드만 소비 / 밴드 경계 defaults.go 상수 하드코딩. 대안 제시 금지.
- **M1 무회귀 불변식**(REQ-006/AC-006): statusline suffix를 config에 게이팅 금지. default(guide=false)에서 soft suffix 유지.
- verification-claim-integrity: 독트린/CHANGELOG는 "stage-2 always fires" 미주장(REQ-005).
- authoring ONLY — git add/commit/push 금지(orchestrator가 plan-auditor + Kickoff Approval 후 커밋).
- AskUserQuestion 미호출. blocker는 구조화 "Missing Inputs" 보고.

## §E — Self-Verification (plan-phase audit-ready signal)

- [ ] 5-artifact + progress.md §E skeleton
- [ ] frontmatter 12-field + tier:M/era:V3R6/related_specs
- [ ] REQ 18개 ↔ AC 18개 1:1 (iter-1 D1/D2 SHOULD-FIX로 16→18)
- [ ] Out-of-Scope h3(spec §B.2 5개 + plan §G) + bullet
- [ ] 6 blocker 해소 명시(§F — 4 C-axis + D1 template drift + D2 concurrent-empty-id)
- [ ] M1 무회귀 불변식 = AC-006
- [ ] Template-First: template mirror 256K drift 정정(add) + section-level edit(full-sync 금지) 반영(AC-016/017)

## §F — Blocker 해소 (구현 계획 핵심 — 4 C-axis + 2 plan-auditor iter-1 SHOULD-FIX)

| # | Blocker | 채택 해소 |
|---|---------|-----------|
| B1 | stage-2 하드 상한 unreachable(auto-compact 85%) | `min(HandoffHardCeilingCapPct=95, getAutoCompactThreshold()+HandoffHardCeilingMarginPct=10)` + `hard<soft` clamp; reachability 한계는 독트린/CHANGELOG 명시(REQ-005). autoCompactThreshold는 memory.go 동일 패키지 → open question 아님. |
| B2 | write 호출부 misplacement(session_id 부재) | `builder.Build`에서 `collectAll` 직후 호출(`input.SessionID`+`data.Memory` 동시 스코프). StatusData 스키마 확장 불필요. |
| B3 | session_id guard false-match race | writer last-writer-wins 스탬프; reader session_id 불일치 → stale 폴백(cross-session false-resume 방지). |
| B4 | fallback-UUID dead-path | session_id 부재여도 write; reader는 양측 부재 시 `captured_at` freshness로 유효(single-session 공통 경로 생존, heuristics-always 방지). |
| D1 (iter-1) | template/live doctrine drift — 256K 행 LIVE만 존재, mirror 부재 | D4가 template mirror에 256K Targets 행 **추가**(parity, REQ-017/AC-017) + Detection 절만 **section-level 편집(BOTH files)**, full-file template→live overwrite/`moai update` full sync 금지(LIVE 256K 삭제 회귀 방지). AC-016는 `<doctrine>`를 LIVE로 명시. |
| D2 (iter-1) | concurrent empty-id cross-contamination hole — B4 empty-id fallback이 B3 guard를 UUID-less 2+ concurrent 세션에서 재개방 | `writer_pid` discriminator(REQ-018/AC-018): Go 헬퍼 `isFreshForSession(rec, curSession, curWriterID)`가 empty-id record에 대해 freshness ∧ writer_pid 일치 요구 → cross-read 기계적 차단. 잔여: 독트린-only reader는 curWriterID 미공급 → concurrent-empty-id는 doctrine-layer 보증 밖(보수적 폴백), 후속 Go reader가 완전 폐쇄. Tier M 비례 대응(trigger 드묾). design §D.4a/§D.4b + acceptance §D.4. |

## §G — Milestones (우선순위 순)

### M1 — config 상수 + 2단계 stage 게이트 (Priority: High)

**목표**: 밴드 경계 명명 상수화 + `handoffGuideStage` 열거형 + 2단계 렌더. M1(CTXGUIDE) 무회귀.

**파일 세트**:
- `internal/config/defaults.go` (+ `HandoffSoftLargePct` `HandoffSoftStandardPct` `HandoffLargeWindowCutoff` `HandoffHardCeilingCapPct` `HandoffHardCeilingMarginPct`)
- `internal/statusline/renderer.go` (`handoffGuideStage` 신규 + `shouldShowHandoffGuide` wrapper + `softThresholdPct`/`hardCeilingPct` + `renderBarsInline` switch; `internal/config` import 추가)
- `internal/statusline/stdinfields_test.go` 또는 `renderer_test.go` (stage enum 테이블 테스트 — soft/hard/none, 256K/1M/500K 경계, clamp, M1 기존 `TestShouldShowHandoffGuide_*` 무손상)

**AC 바인딩**: AC-THRESHOLD-001, 002, 003, 004, 005, 006.

**검증**: `go test ./internal/statusline/... ./internal/config/...`, `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`.

**주의**: soft 문자열 `(⚠️/clear)` verbatim(회귀 금지). hard 공식 clamp 경계 테스트(ac override 낮은 값 → hard=soft). REQ-005 문서 AC는 M3에서.

### M2 — context-usage.json 영속화 (Priority: High)

**목표**: `builder.Build`에서 atomic best-effort write + throttle + session_id guard + fallback-UUID.

**파일 세트**:
- `internal/statusline/context_usage.go` (신규 — `writeContextUsage`(+`writer_pid`=`os.Getpid()`), `readContextUsage`, `sameSemanticPayload`(writer_pid/captured_at 제외), `isFreshForSession`(session_id guard + empty-id writer_pid discriminator), `resolveProjectDir`, record struct; model_cache.go 패턴 재사용)
- `internal/statusline/context_usage_test.go` (신규 — atomic write, throttle skip, schema(+writer_pid), session_id guard, fallback-UUID freshness, **concurrent empty-id writer_pid discriminator(AC-018)**, best-effort silent-fail)
- `internal/statusline/builder.go` (`Build`에 `writeContextUsage(...)` 호출 1줄 추가, collectAll 직후, `os.Getpid()` 인자)
- `internal/statusline/builder_test.go` (Build 호출 시 파일 생성 e2e)

**AC 바인딩**: AC-THRESHOLD-007, 009, 010, 011, 012, 013, 014, 018.

**검증**: `go test ./internal/statusline/...`, `go test -race ./internal/statusline/...`(동시 write throttle + AC-018), coverage ≥85%. context-usage.json은 `t.TempDir()` 내에서만(테스트 격리, dev project 오염 금지).

**주의**: write는 무조건(config 무관, AC-007). Memory.Available==false → skip. throttle plateau 캐비어트는 reader freshness로 완화(design §D.3). `writer_pid`는 throttle 비교 제외(render-ephemeral, design §D.2 caveat). concurrent empty-id 잔여는 residual(acceptance §D.4).

### M3 — Detection 독트린 재작성 + Guide 화해 + reachability 문서 (Priority: High)

**목표**: `context-window-management.md` § Detection Heuristics state-file-first 재작성(Template-First). Guide/Mode 화해 문서. reachability 한계 명시.

**파일 세트** (BOTH files, section-level 편집만 — full-file overwrite/`moai update` full sync 금지):
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`: (a) § Detection Heuristics 절 재작성 — state-file-first + guard 매트릭스(§25 중립), (b) § Context Window Targets에 **누락된 256K 행 추가**(D1 drift parity, REQ-017)
- `.claude/rules/moai/workflow/context-window-management.md` (LIVE): § Detection Heuristics 절만 section-level 편집(`make build` 후 sync). **Targets 표 무접촉**(256K 이미 존재, 중복 금지)
- (design.md는 이미 Guide/Mode 화해 + reachability + writer_pid 서술 보유 — 산출물)
- CHANGELOG는 sync-phase(manager-docs) 소관 — reachability 한계 문구 포함

**AC 바인딩**: AC-THRESHOLD-005, 008, 015, 016, 017.

**검증**:
- `grep -c 'context-usage.json' <LIVE doctrine>` ≥1 ∧ `grep -c 'context-usage.json' <template doctrine>` ≥1 (state-file-first, BOTH)
- `grep -c '256,000' <template doctrine>` == 1 (D1 parity 회복, AC-017) ∧ `grep -c '256,000' <LIVE doctrine>` == 1 (중복 없음, AC-016/017)
- `make build` 성공, template 사본 `grep -E 'SPEC-|REQ-'` == 0(§25)

**주의**: Template-First — template 우선 편집 후 `make build`. **template mirror 256K 행 ADD(D1 drift 정정)**, LIVE Targets 표는 무접촉. section-level 편집만(full-sync 금지 — LIVE 256K 삭제 회귀). Guide advisory는 독트린 서술만(신규 Go 훅 금지).

## §H — Anti-Patterns (회피)

- statusline suffix를 `guide==true`에 게이팅 → default false에서 소멸 = M1 회귀. **무조건 렌더 필수.**
- 밴드 리터럴 renderer.go inline 유지 → §14 위반. defaults.go 상수.
- 하드 상한 naive 95% 고정 → auto-compact 무시 = "always fires 주장" 위험(verification-claim-integrity). auto-compact-aware 공식 + 한계 명시.
- write 호출부를 collectAll 내부/renderer에 배치 → session_id 부재. Build 배치 필수.
- Detection 독트린 live만 편집 → template drift. Template-First + make build(BOTH files).
- **LIVE Targets 표에 256K 행 재추가 → 중복(이미 존재). LIVE는 무접촉**; 반대로 **template mirror는 256K 행 누락 → 반드시 ADD(D1 drift parity)**. live/template를 혼동하지 말 것.
- **full-file template→live overwrite / `moai update` full sync → LIVE 256K 행 삭제 = M1 회귀. section-level 편집만.**
- concurrent empty-id에서 `captured_at` freshness만으로 own-session 판정 → session B가 A 스냅샷 오독. `writer_pid` discriminator 필수(Go 헬퍼, AC-018).
- `writer_pid`를 throttle semantic 비교에 포함 → render-ephemeral PID로 매 render write churn. throttle 제외 필수(design §D.3).
- context-usage.json을 커밋 → 런타임 생성물. 코드만 커밋.

### Out of Scope — plan-level (구현 계획 범위 밖)
- M1 밴드 로직 재설계 — verbatim 승계(soft 단계).
- M3 auto-resume 소비 경로 / handoff pending.json — 무접촉.
- 신규 HandoffConfig 필드 / 밴드 config override — LOCKED #1.
- state-file reader Go 파서 — 독트린 서술만(런타임 자동 파서는 후속 SPEC).
- cross-session 레지스트리(active-sessions.json) 통합 — 후속.

## §I — Cross-References

- research.md §A~G, design.md §A~G, acceptance.md §D
- `internal/statusline/{renderer,memory,builder,model_cache}.go`
- `internal/config/{types,defaults}.go`
- `.claude/rules/moai/workflow/context-window-management.md`(+ template mirror)
- M1 `SPEC-HANDOFF-CTXGUIDE-001` §1.3, M3 `SPEC-HANDOFF-AUTORESUME-001`
