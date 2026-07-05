---
id: SPEC-HANDOFF-MSGMODE-001
title: "핸드오프 메시지 오케스트레이션-모드 내장 (message-v2) — 구현 계획"
version: "0.1.0"
status: in-progress
created: 2026-07-04
updated: 2026-07-05
author: MoAI
priority: P2
phase: "v3.0.0"
module: "docs-handoff"
lifecycle: spec-anchored
tags: "session-handoff, orchestration-mode, message-v2, mode-seed, handoff-protocol, tier-s, epic-handoff-v2"
era: V3R6
---

# 구현 계획 — SPEC-HANDOFF-MSGMODE-001 (Tier S)

## §A.1 Tier 판정

Tier **S** (S/M 경계 — 아래 근거로 S 확정).

- **S 방향(우세)**: doctrine-only(Go/config/상태 파일 변경 0), 영향 파일 4개(< 5), 신규 의존성 0, 신규 임계값 0. 런타임 동작 변경 없음(paste-ready 텍스트 포맷). 회귀 가능한 테스트 없음.
- **M 방향(비우세)**: 6개 doctrine concern(mode-seed core + B1/B2/B3 + ultracode 재정합 + JSON-twin note) + live↔mirror parity + §25 neutrality 조율.
- **경계 해소**: `orchestration-mode-selection.md §B.2` "경계에서는 더 단순한 쪽으로" 및 Tier anti-pattern(대규모를 S로 낮추는 게이밍 방지 — 본건은 ~150-250 LOC markdown으로 게이밍 아님)에 따라 **S** 확정. 형제 M1(SPEC-HANDOFF-CTXGUIDE-001)도 Tier S.
- plan-auditor PASS 임계 **0.75**. 산출물: spec.md + plan.md(2 files, AC는 spec.md §3 인라인) + progress.md(lifecycle 추적). acceptance.md 미분리(consolidated lifecycle — small doctrine SPEC).

## §A.2 변경 파일 (정확한 타깃 — 모두 git-clean 확인됨)

1. **`.claude/rules/moai/workflow/session-handoff.md`** (live SSOT)
   - § Canonical Format 6-block 스켈레톤(L84 인근): 조건부 `mode:` 라인 추가 + `# /effort ultracode` → bare `ultracode` opener 재정합.
   - § Field-by-Field Specification Block 1(L78-L80): 4-enum↔Mode 매핑 표 + 지시어 바인딩 규칙 + SEED-not-permission 불변식 + ultracode bare/slash 구분 추가. **B1**: L79 "sits immediately after `ultrathink.`" → "(or after the `mode:` line when present)" 일반화; L80 goal 배치 절도 mode 라인 순서 반영.
   - § Diet Constraints pre-emit self-check(L288 "paste-ready budget — 9 items"): **B3 — 불변 유지(무접촉)**. 이 self-check는 `paste-ready budget` concern이며 M2에서 항목 수·내용을 변경하지 않는다(신규 template-completeness self-check 섹션도 SH에 생성하지 않음). `mode:` 검증 항목 추가(10→11)는 별개 concern인 `moai.md §8` "session-handoff template completeness" self-check(item #2) 소관 — §A.5 R1 참조. SH 측 유일 조치는 concern-name qualifier(`paste-ready budget`) 표기 유지.
   - § Localization Table: `mode:`가 protocol token(verbatim 보존)임을 명시 — **신규 locale 행 0**.
   - JSON-twin `schema_version: 2` forward-compat note(신규 소절): "현재 JSON twin 부재(doctrine-only)".
   - **B2**: solo-sequential "emit-discouraged" 단일 정책 소절 + 기존 "never emitted" 상충 서술 제거.
2. **`.claude/output-styles/moai/moai.md`** (live 렌더 표면, §8)
   - §8 Session Handoff 6-block 스켈레톤(L679-L697): 조건부 `mode:` 라인 + bare `ultracode` opener 재정합(L682 `# /effort ultracode` 정정).
   - §8 4-enum↔Mode compact 참조(4 토큰 + Mode 3/4/5/6 grep 가능하게).
   - §8 "session-handoff template completeness" pre-emit self-check(L720 "10 items"): **B3** → `mode:` 검증 항목 1개 추가 → **10→11 items**(M8 표면 내부 갱신; SH의 `paste-ready budget` 9-item과는 별개 concern이므로 "통일"이 아니라 M8-local 갱신). concern-name qualifier(`session-handoff template completeness`) 표기 유지.
   - §8 translation 표(Cut-line/Header): locale 컬럼 4 유지(신규 행 0).
3. **`internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`** (미러) — #1과 동일 편집, **§25 neutrality**(SPEC-ID·날짜·SHA·REQ 토큰 무유출).
4. **`internal/template/templates/.claude/output-styles/moai/moai.md`** (미러) — #2와 동일 편집, §25 neutrality.

## §A.3 구현 순서 (doctrine-only; cycle_type=ddd 성격 — behavior-preserving 문서 편집)

1. **SSOT 먼저 (`session-handoff.md`)**: mode-seed core(REQ-001..009) → B1 placement(REQ-010) → B2 emit-discouraged(REQ-011) → B3(REQ-012: SH의 `paste-ready budget` 9-item self-check는 불변 유지 — 10→11 편집은 step 2 M8 소관) → JSON-twin note(REQ-008). SSOT가 canonical이므로 최초 확정.
2. **렌더 표면 (`moai.md §8`)**: SSOT 확정본을 render skeleton + compact 참조 + "session-handoff template completeness" self-check 10→11(`mode:` 항목 추가)로 반영(REQ-002/006/012). parity sentinel(SH § Cross-references ↔ M8 §8 sentinel)의 3개 concern-name qualifier + locale 컬럼 카운트(4) 상호 일치 유지.
3. **미러 2개**: #1/#2 편집을 `internal/template/templates/...`에 반영 + §25 neutrality 확인(REQ-013).
4. **검증**: AC-MSGMODE-001..016 grep 행렬 + `moai spec lint`.

> 배치 순서 근거: SSOT→render→mirror. SSOT를 먼저 확정해 render/mirror가 canonical을 복제하도록 함(drift 최소화). B3의 9-vs-10은 **동일 self-check의 count drift가 아니라 별개 concern**임을 §A.5 R1에서 확정 — M2는 M8 template-completeness self-check만 10→11 갱신하고 SH paste-ready-budget 9는 불변 유지한다(SSOT-attribution 불일치는 §A.7 이연).

## §A.4 PRESERVE (변경 금지)

- **병렬 세션 dirty working tree**(절대 무접촉): `.claude/rules/moai/workflow/dynamic-workflows.md`(+미러), `.moai/config/sections/workflow.yaml`, `context-window-management.md`, `internal/statusline/*`, `internal/web/*`, `internal/hook/*`, `internal/settings/*`, `internal/cli/*`, `README.ko.md`, `docs-site/*`.
- **cross-ref-only(READ만, 편집 금지)**: `orchestration-mode-selection.md`(Mode 카탈로그 — mirror-drift resync는 Out of Scope), `goal-directive.md`(/goal 원천).
- **타 SPEC 디렉토리**: `.moai/specs/SPEC-WEB-CONSOLE-011/`, `.moai/specs/SPEC-HANDOFF-CTXGUIDE-001/`.
- **v1 byte-identity 불변식**: solo-sequential 케이스는 라인 생략으로 v1과 byte-identical 유지(공통 케이스 zero-diff) — 스켈레톤 편집이 기본 케이스 출력을 바꾸지 않도록.
- runtime-managed: `.moai/state/*`, `.moai/cache/*`, `.moai/logs/*`.
- **git 작업 금지**: 본 plan-phase는 산출물 authoring만. `git add`/`commit`/`push` 미수행(uncommitted 유지).

## §A.5 위험 · 완화

- **위험 R1 (B3 대상 표면 확정 — Option A)**: SSOT § Diet Constraints(SH)는 "paste-ready budget" **9-item** self-check, M8 §8은 이와 **별개 concern**인 "session-handoff template completeness" **10-item**을 참조한다 — 이 둘은 동일 self-check의 count drift가 아니라 **서로 다른 두 self-check**다. → **완화/결정(Option A, M8-only 최소 범위)**:
  1. B3의 canonical 편집 대상은 **`moai.md §8` "session-handoff template completeness" self-check 단 하나**다. 여기에 `mode:` 검증 1개를 추가해 **10→11**로 갱신한다.
  2. SH의 **"paste-ready budget" 9-item self-check는 불변 유지**(무접촉)한다. M2는 SH에 **신규 template-completeness self-check 섹션을 생성하지 않는다**.
  3. 두 self-check가 표면 간 혼동되지 않도록 concern-name qualifier(`paste-ready budget` vs `session-handoff template completeness`)를 명확히 유지한다.
  4. **사전 존재 drift는 이연**: `moai.md §8`이 template-completeness self-check를 `session-handoff.md`가 SSOT인 것처럼 귀속하나 SH에는 해당 섹션이 부재하고(§ Cross-references sentinel은 3개 qualifier `paste-ready budget`/`localization render`/`session-handoff template completeness`를 명명) 실제로는 `paste-ready budget`만 보유하는 **SSOT-attribution 불일치**는 M2에서 해소하지 않고 §A.7 별도 doctrine-reconciliation chore로 이연한다.
  plan-auditor가 이 결정(M8-only 갱신 + SH 9 보존 + attribution drift 이연)을 독립 검증.
- **위험 R2 (mode:과 goal/ultracode 배치 상호작용)**: 새 `mode:` 라인이 ultracode(opener 부가) 및 `/goal`(옵션 라인)과 순서 충돌 가능. → 완화: §4 배치 불변식(opener → mode → goal → lessons → session_id) 단일 정의. bare `ultracode`는 opener 라인 내부(별도 라인 아님)이므로 `mode:` 라인 슬롯과 겹치지 않음.
- **위험 R3 (mirror parity + neutrality 충돌)**: §24 mirror-parity(live↔mirror 동일)와 §25 neutrality(미러는 internal-content 무유출)가 상충 가능(내부 참조 문구 시). → 완화: 편집 콘텐츠는 generic doctrine prose만(SPEC-ID/날짜/SHA 미포함)이므로 parity와 neutrality 동시 만족. AC-013 grep으로 확인.
- **위험 R4 (v1 byte-identity 회귀)**: 스켈레톤 편집이 solo-sequential 기본 출력을 바꾸면 M1 이래 공통 케이스 zero-diff가 깨짐. → 완화: `mode:` 라인은 mode ≠ solo-sequential일 때만 발화(조건부) — 기본 케이스 출력 무변. AC-014로 서술 확인.
- **위험 R5 (병렬 세션 shared-checkout race)**: 4개 타깃이 병렬 세션과 겹치지 않음(git-clean 확인). → 완화: run-phase 진입 시 pre-spawn `git fetch` + 4개 타깃 재-clean 확인(agent-common-protocol § Pre-Spawn Sync Check). specific-path commit.

## §A.6 자가 검증 (완료 시)

- AC-MSGMODE-001..016 PASS/FAIL 행렬 + 실제 grep 명령 출력(verbatim).
- `grep -c '11 items' .claude/output-styles/moai/moai.md` ≥ 1 (M8 template-completeness self-check 10→11) **AND** `grep -c '9 items' .claude/rules/moai/workflow/session-handoff.md` ≥ 1 (SH paste-ready-budget 9-item 불변 보존). SH에 `11 items` 요구 없음.
- `grep -n 'immediately after' .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` (모든 placement clause에 "(or after the `mode:` line when present)").
- `grep -rn 'never emitted' .claude/rules/moai/workflow/session-handoff.md` = 0 (B2 상충 제거).
- `grep -E 'SPEC-HANDOFF-MSGMODE|SPEC-MSGMODE|2026-07-04' internal/template/templates/.claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/output-styles/moai/moai.md` = 0 (neutrality).
- `moai spec lint .moai/specs/SPEC-HANDOFF-MSGMODE-001/spec.md` **AND** `moai spec lint .moai/specs/SPEC-HANDOFF-MSGMODE-001/plan.md` 각각 "No findings"(exit 0), 신규 ERROR 0. **파일 경로 형식 사용** — 디렉터리 형식은 `ParseFailure ... is a directory`(exit 1)를 콘텐츠와 무관하게 반환하므로 clean lint 입증 불가.
- 4 enum 토큰 + Mode 3/4/5/6 존재: `grep -c 'solo-sequential\|parallel-subagents\|agent-team\|dynamic-workflow' <SH> <M8>`.

### §A.7 Out of Scope

이 절은 spec-lint Exclusions 요구를 만족한다. 상세는 spec.md §1.3 참조.

- `orchestration-mode-selection.md` mirror-drift resync(3074B) — 별도 사전 chore.
- null session_id filename nonce — M3 소관.
- `handoff.yaml` / `HandoffConfig` struct+loader — M3 landing · M4 소비.
- Go 코드 / config / 상태 파일 변경 — M2는 doctrine-only(JSON-twin은 개념만 문서화).
- `dynamic-workflows.md` / `goal-directive.md` / `orchestration-mode-selection.md` 본문 편집 — cross-ref-only(READ).
- `session-handoff.md` template-completeness self-check **SSOT-attribution drift** — `moai.md §8`이 "session-handoff template completeness" self-check를 `session-handoff.md`에 존재하는 것으로 귀속하나 SH에는 해당 섹션이 부재(SH는 `paste-ready budget` self-check만 보유; § Cross-references sentinel은 3개 qualifier `paste-ready budget`/`localization render`/`session-handoff template completeness`를 명명). 이 귀속 불일치 해소(SH에 template-completeness self-check 신설 vs M8 귀속 문구 정정)는 별도 doctrine-reconciliation chore 소관 — M2 아님(§A.5 R1 결정).
