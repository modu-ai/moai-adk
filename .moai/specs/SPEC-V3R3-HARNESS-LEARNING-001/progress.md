# SPEC-V3R3-HARNESS-LEARNING-001 — Progress

> Last Updated: 2026-04-27 (Wave B.2 완료, Wave B.3 대기)
> Branch: `feat/SPEC-V3R3-HARNESS-LEARNING-001-impl`
> Methodology: TDD (RED → GREEN → REFACTOR)
> Coverage Target: ≥85% per new file

---

## Wave Plan (5-Wave 분할)

행님 결정 (2026-04-27): 39 tasks를 Phase별 5 wave로 순차 진행. 각 wave마다 commit + push, PR은 5 wave 모두 완료 후 단일 생성.

| Wave | Phase | Status | Commit | Coverage | Tasks |
|------|-------|--------|--------|----------|-------|
| B.1 | Phase 1 — Observer + JSONL Schema | ✅ Done | `9c965d87d` | 89.3% | T-P1-01~05 (5) |
| B.1.1 | .gitignore 학습 데이터 패턴 보완 | ✅ Done | `4660812fb` | — | follow-up |
| B.2 | Phase 2 — Tier Classifier + Description/Trigger Writer | ✅ Done | `92a80a14f` | 90.0% | T-P2-01~07 (7) |
| B.3 | Phase 3 — 5-Layer Safety Architecture | ⏸️ Pending | — | — | T-P3-NN |
| B.4 | Phase 4 — Applier + CLI + Coordinator Skill | ⏸️ Pending | — | — | T-P4-NN |
| B.5 | Phase 5 — Integration Tests + Documentation | ⏸️ Pending | — | — | T-P5-NN |

---

## Wave B.1 산출물 (Done)

**Files**:
- 신규: `internal/harness/types.go` (52L), `observer.go` (147L), `retention.go` (221L), `observer_test.go` (338L), `integration_test.go` (191L)
- 신규: `.claude/hooks/moai/handle-harness-observe.sh` (29L)
- 신규: `internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl` (34L)
- 수정: `internal/cli/hook.go` (+57L, harness-observe subcommand 라우팅)
- 수정: `.gitignore` (+4L, 학습 데이터 패턴)

**Verification**:
- go test -race PASS (1.326s)
- golangci-lint 0 issues
- coverage 89.3% (`internal/harness`)

**MX Tags**:
- `RecordEvent`: `@MX:ANCHOR` + `@MX:TODO` (Phase 4 learning.enabled gate)
- `PruneStaleEntries`: `@MX:ANCHOR`
- `appendToGzip`: `@MX:WARN` (REASON: gzip concat 동시 쓰기 stream 분리 필요)

---

## Wave B.2 산출물 (Done)

**Files**:
- 신규: `internal/harness/learner.go` (AggregatePatterns + ClassifyTier + WritePromotion)
- 신규: `internal/harness/applier.go` (EnrichDescription + InjectTrigger, feature-gated)
- 신규: `internal/harness/learner_test.go`, `applier_test.go`
- 수정: `internal/harness/types.go` (+Pattern, +Tier enum, +Promotion)

**Verification**:
- go test -race PASS (전체 + Phase 2 신규 29 tests)
- golangci-lint 0 issues
- coverage 90.0% 전체, learner.go ~91%, applier.go ~95%
- Feature flag (`enableTriggerInjectionWrites = false`) 동작 검증
- Frontmatter preservation golden fixture pass

---

## Plan-Auditor Findings (Sync 단계 deferred)

Plan Audit Gate: PASS, 보고서 `.moai/reports/plan-audit/SPEC-V3R3-HARNESS-LEARNING-001-2026-04-27.md`.

3 non-blocking findings (모두 `/moai sync` 단계에서 처리):

1. **AC-02 self-contradicting** (`acceptance.md:L31`): `from_tier: "rule", to_tier: "rule"` 자기모순 → `from_tier: "heuristic", to_tier: "rule"`로 수정 필요
2. **REQ-HL-008 compound** (`spec.md:L133-135`): Contradiction Detector + Rate Limiter 한 REQ에 묶임 → REQ-HL-008a/008b로 분리 권장
3. **target_release stale** (`spec.md:L23`): `v2.17.0` → `v2.19.0` 업데이트 (행님 release 계획 일치)

---

## Resume Message (다음 세션용 — paste verbatim)

행님이 `/clear` 실행 후 새 세션에 붙여넣을 메시지:

```
ultrathink. Wave B.3 이어서 진행. SPEC-V3R3-HARNESS-LEARNING-001 Phase 3 (5-Layer Safety Architecture)부터.

applied lessons: feedback_large_spec_wave_split.md, lessons.md #9 (Agent 1M context limit + wave 분할).

branch: feat/SPEC-V3R3-HARNESS-LEARNING-001-impl (B.1+B.2 push 완료, 3 commits 누적).
완료 wave: B.1 (Phase 1 Observer, 9c965d87d), B.1.1 (gitignore, 4660812fb), B.2 (Phase 2 Tier Classifier, 92a80a14f).
누적 coverage: 90.0%, all gates GREEN.

progress.md 경로: .moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/progress.md

다음 단계: Phase 3 (5-Layer Safety) tasks 확인 후 manager-tdd에 Wave B.3 위임.
- L1 Frozen Guard: path-prefix matcher, FROZEN zone 차단 (REQ-HL-006)
- L2 Canary Check: shadow-eval, 0.10 score drop reject (REQ-HL-007)
- L3 Contradiction Detector: user customization 충돌 → AskUserQuestion (REQ-HL-008a)
- L4 Rate Limiter: 3 updates/week, 24h cooldown (REQ-HL-008b)
- L5 Human Oversight: orchestrator AskUserQuestion 게이트 (REQ-HL-005)

완료 후: Wave B.4 (Phase 4 Applier+CLI+Coordinator Skill) → Wave B.5 (Phase 5 Tests+Docs) → PR 생성 → 머지 → /moai run SPEC-V3R3-DESIGN-PIPELINE-001 (Wave C) → ./scripts/release.sh v2.19.0 (행님 manual, Wave D).

target_release: v2.19.0 (sync 단계에서 spec.md v2.17.0 → v2.19.0 업데이트, plan-auditor finding #3).
```

---

## Next Wave (B.3) Pre-Brief

**Phase 3 — 5-Layer Safety Architecture**

핵심 컴포넌트 (예상, plan.md §3 + tasks.md §"## Phase 3" 정확 확인 필요):
- `internal/harness/safety.go` (L1~L5 통합 또는 분리 파일)
- L1 Frozen Guard: `.claude/agents/moai/`, `.claude/skills/moai-*/`, `.claude/rules/moai/`, `.moai/project/brand/` 차단
- L2 Canary Check: 최근 3 세션 shadow-evaluation, projected effectiveness drop > 0.10 → reject
- L3 Contradiction Detector: trigger keyword 중첩 / chaining rule 모순 탐지 → flag + AskUserQuestion (orchestrator level)
- L4 Rate Limiter: max 3 auto-updates/7-day window, 24h cooldown
- L5 Human Oversight: AskUserQuestion 호출은 orchestrator만 (subagent boundary)

**HARD constraints**:
- No FROZEN zone writes (L1이 자기 자신 검증)
- Local-only logging
- Korean comments
- Subagent must NOT call AskUserQuestion — L5는 orchestrator API 호출 인터페이스만 제공
- Feature flag (Phase 4까지 actual write 미활성)

**예상 분량**: 6~8 tasks, ~600 LOC implementation + ~400 LOC tests (B.2와 유사하거나 약간 큼)

---

## Decisions Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-04-27 | 5-wave 분할 | 39 tasks → wave당 ~5-9 tasks, SSE stall 회피 (memory `large_spec_wave_split.md`) |
| 2026-04-27 | target_release update는 sync 단계 | 별도 commit 없이 자연스럽게 머지 |
| 2026-04-27 | 5 waves 완료 후 단일 PR | reviewable 단위 = SPEC, 중간 PR 노이즈 회피 |
| 2026-04-27 | Plan Audit PASS 후 진행 | 3 non-blocking findings는 sync에서 처리 |
| 2026-04-27 | Wave B.3 시작 전 /clear | 컨텍스트 ~70% 도달, 75% 임계 회피 (HARD per context-window-management.md) |
