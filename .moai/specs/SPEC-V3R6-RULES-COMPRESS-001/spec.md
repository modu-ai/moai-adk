---
id: SPEC-V3R6-RULES-COMPRESS-001
title: "Always-Loaded Rule Body 압축 3건 — session-handoff / context-window / verification-batch (-4.5K tokens)"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "rules, compression, token-economy, wave-1, v3.0.0, layer-1"
tier: S
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-RULES-PATH-SCOPE-001, SPEC-V3R6-SKILL-CONSOLIDATE-001, SPEC-V3R6-SKILL-COMPRESS-001]
---

# SPEC-V3R6-RULES-COMPRESS-001: Always-Loaded Rule Body 압축 3건

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — v3.0 환골탈태 Wave 1 (Token Baseline 감축, Layer 1) 두번째 SPEC. design.md §Layer 1 line 165-184에서 도출. 3 always-loaded workflow rules (session-handoff.md 1,927w / context-window-management.md 712w / verification-batch-pattern.md 764w, 합계 3,403w) 본문을 압축 (1,000w / 500w / 400w, 합계 1,900w, **-1,503w ≈ -4.5K tokens, -44%**). HARD 조항 verbatim 보존 의무 + canonical artifact (model-specific threshold table, 6-block resume format, worktree Block 0, verification class taxonomy) 보존 의무. Tier S LEAN (2 artifacts: spec.md + acceptance.md, AC inline 통합). |

## 1. Goal

`.claude/rules/moai/workflow/` 디렉토리의 **always-loaded 3개 rule**을 본문 압축하여 매 세션 시작 시 항상 로딩되는 토큰 부담을 **약 4,500 tokens (~38%) 감축**한다.

- `workflow/session-handoff.md`: 1,927w → 1,000w (-927w, ~-2.8K tokens)
- `workflow/context-window-management.md`: 712w → 500w (-212w, ~-0.6K tokens)
- `workflow/verification-batch-pattern.md`: 764w → 400w (-364w, ~-1.1K tokens)
- **합계**: 3,403w → 1,900w (**-1,503w ≈ -4.5K tokens, -44%**)

**압축 원칙** (의미 보존, 본문 단축):

1. [ZONE:Frozen/Evolvable] [HARD] 조항은 **verbatim 보존** — 압축 시 의미 변형 절대 금지
2. Canonical artifact (model-specific threshold table, 6-block resume format, worktree Block 0 prepend pattern, verification class taxonomy table) **verbatim 보존**
3. 비핵심 prose (Why this matters / Detection heuristics 산문 / Anti-pattern 카탈로그 부연 설명) **단축**
4. Cross-reference로 **외부화 가능한 본문은 링크만 남기고 본문 단축** (예: `verification-batch-pattern.md` 7-item canonical example은 `agent-common-protocol.md` §Parallel Execution에 이미 verbatim 존재 → 인용 link만 보존)
5. HISTORY 섹션 (있는 경우) 또는 file-end footer 보존

## 2. Why

### 2.1 Token Baseline 부담의 누적 증거

매 세션 시작 시 Claude Code는 `.claude/rules/moai/**/*.md` 중 `paths:` frontmatter 제약이 없는 rule을 **모두 always-loaded**로 컨텍스트에 주입한다. v3.0 환골탈태 설계 문서 (`v3.0-design-2026-05-22.md` §Layer 1 line 165-184)가 식별한 9개 always-loaded rules 중:

| Rule | Baseline w | Layer 1 처리 | 출처 |
|------|------------|--------------|------|
| `workflow/session-handoff.md` | 1,927w | **본 SPEC 압축 대상** | design.md line 172 |
| `workflow/context-window-management.md` | 712w | **본 SPEC 압축 대상** | design.md line 174 |
| `workflow/verification-batch-pattern.md` | 764w | **본 SPEC 압축 대상** | design.md line 176 |
| `core/zone-registry.md` 외 4건 | ~5,000w | path-scope 전환 (Wave 1 SPEC `RULES-PATH-SCOPE-001` 담당) | design.md line 168-171 |

**근본 원인**: 본 3개 rule은 항상-load 의무가 있는 cross-cutting concern (Trigger #3 user explicit session-end은 SPEC 컨텍스트 무관 발생, model-specific threshold은 모든 모델 클래스에 적용, verification batch는 모든 run-phase 완료 시 적용)이므로 **path-scope 전환 불가**. 따라서 **본문 압축 only** 만이 유일한 감축 수단.

### 2.2 v3.0 §8 사용자 결정 반영

사용자 §8 4 결정 (2026-05-22, `design.md` line 449-456):

1. **목표 시점**: 6/15 이전 → 본 SPEC priority **P1**
2. **Wave 우선순위**: 토큰 감축 우선 → 본 SPEC = Wave 1 핵심
3. Profile preset / GLM 라우팅 — 본 SPEC 무관

### 2.3 Wave 0 정렬

선행 Wave 0 두 SPEC (`HOOK-CONTRACT-FIX-001` commit `48340e726` + `DOCS-USER-DRIFT-001` commit `d386cca0e`) 모두 main 직진 MERGED. Wave 0 가 hook 정합성 (active-creator contract regression guard) + 문서 drift (F8 sync CI doctrine + PR #1045 non-overlap)를 처리한 직후, Wave 1은 **토큰 baseline 감축**으로 진입.

## 3. Scope

### 3.1 In Scope

| 파일 | 현재 word count | 목표 word count | 절감 |
|------|---------------:|---------------:|-----:|
| `.claude/rules/moai/workflow/session-handoff.md` | 1,927w | ≤ 1,200w (목표 1,000w + 200w 여유) | -727w 이상 |
| `.claude/rules/moai/workflow/context-window-management.md` | 712w | ≤ 600w (목표 500w + 100w 여유) | -112w 이상 |
| `.claude/rules/moai/workflow/verification-batch-pattern.md` | 764w | ≤ 500w (목표 400w + 100w 여유) | -264w 이상 |
| **합계** | **3,403w** | **≤ 2,300w** | **-1,103w 이상 (~-3.3K tokens 최소)** |

압축 후 baseline은 목표값 (1,900w / -4.5K tokens) 달성을 권장하되 AC threshold 는 보수적 상한선 (2,300w)으로 설정 — 의미 손상 방지 우선.

### 3.2 보존 의무 Canonical Artifacts (verbatim, 압축 후에도 유지)

다음은 **압축 후에도 verbatim 유지** 의무 (변형 시 SSOT 위반):

#### 3.2.1 `session-handoff.md`

- **5 Trigger 테이블** (line 19-26): Trigger #1 model-specific threshold (1M 50% / 200K 90%) 포함 5개 행 — 컨텍스트 임계 진입 결정 SSOT
- **Canonical 6-block format Verbatim Spec** (line 32-58 + Field-by-Field Specification): Block 1-6 paste-ready 구조 — `paste-ready resume message` 의 contract
- **Auto-Memory Integration 4 step** (line 102-105): memory 파일 저장 의무 4 step
- **Worktree-Anchored Resume Pattern Block 0 format** (line 168-204): L3 `--worktree` opt-in 케이스의 cwd anchoring 의무
- **[ZONE:Evolvable] [HARD] 5 clauses** (line 20, 34, 100, 152, 203): verbatim

#### 3.2.2 `context-window-management.md`

- **Context Window Targets 3-row table** (line 15-19): Opus 4.7 1M / Sonnet-Opus 200K / Haiku 200K — model-specific threshold SSOT
- **[ZONE:Evolvable] [HARD] 5 clauses** (line 13, 27, 32, 41, 47): verbatim
- **Detection Heuristics 4 항목** (line 60-63): 4 signal (cumulative bytes / system reminder volume / tool result size / Agent() invocation count) — 단축 OK이나 4 항목 모두 보존

#### 3.2.3 `verification-batch-pattern.md`

- **Verification Class Taxonomy 8-row table** (line 25-34): batch-safe vs not 분류 — class-level decision SSOT
- **Default Grouping 5-row table** (Group A-E, line 71-77): functional / boundary / quality / smoke / benchmark — 운영 기본 그룹
- **7-item canonical example reference** (line 122-124): `agent-common-protocol.md` §Parallel Execution 인용 link verbatim (본문 inline 금지)

### 3.3 Out of Scope (Exclusions)

#### 3.3.1 Out of Scope

다음 항목은 **본 SPEC scope 외** — 별도 SPEC 또는 영원히 손대지 않음:

- **`core/zone-registry.md`, `design/constitution.md`, `development/manager-develop-prompt-template.md`, `workflow/agent-teams-pattern.md`** (4건): `SPEC-V3R6-RULES-PATH-SCOPE-001` 담당 — `paths:` frontmatter 전환으로 conditional loading 처리. 본문 압축 아닌 별도 메커니즘.
- **`core/agent-common-protocol.md`**: 본 SPEC 의 7-item canonical reference 의 anchor 역할 (보존 의무). 손대지 마라.
- **`core/askuser-protocol.md`, `core/moai-constitution.md`, `core/hooks-system.md`**: AskUserQuestion / agent core behavior / hook contract SSOT — 변경 risk 매우 높음, 본 SPEC scope 외.
- **`.claude/rules/moai/NOTICE.md`**: Apache 2.0 attribution — 법적 텍스트, 손대지 마라.
- **HARD 조항 의미 변형**: 압축 = body prose 단축이며, semantics 변경 절대 금지. 본 SPEC은 HARD 조항 verbatim 보존을 enforcing.
- **`paths:` frontmatter 추가**: 3개 rule 모두 `session-handoff.md` 자체 §Loading scope 주석이 `intentionally always-loaded`를 명시 (Trigger #3 user explicit session-end은 SPEC 컨텍스트 무관 발생). 압축 only, scope-narrowing 금지.
- **인접 워크플로우 rule 압축** (`ci-watch-protocol.md`, `ci-autofix-protocol.md`, `spec-workflow.md`, `worktree-integration.md`, etc.): 본 SPEC 3개 외 동작 변경 금지.

## 4. Requirements (EARS Format)

### REQ-RC-001 (Ubiquitous, Mandatory)

The system shall reduce the combined word count of the three always-loaded workflow rules (`session-handoff.md` + `context-window-management.md` + `verification-batch-pattern.md`) from **3,403w baseline to ≤ 2,300w** while preserving all HARD clauses and canonical artifacts verbatim.

### REQ-RC-002 (Event-Driven, Mandatory)

WHEN `session-handoff.md` is loaded into the always-loaded rule context THEN the compressed body MUST preserve **every `[ZONE:Frozen]` and `[ZONE:Evolvable] [HARD]` clause verbatim** (baseline count: 5 clauses) AND MUST preserve the canonical artifacts enumerated in §3.2.1 (5 Trigger table / 6-block format / Auto-Memory Integration / Worktree Block 0 pattern).

### REQ-RC-003 (Event-Driven, Mandatory)

WHEN `context-window-management.md` is loaded into the always-loaded rule context THEN the compressed body MUST preserve **every `[ZONE:Evolvable] [HARD]` clause verbatim** (baseline count: 5 clauses) AND MUST preserve the **model-specific Context Window Targets 3-row table** (Opus 4.7 1M 50% / Sonnet-Opus standard 200K 90% / Haiku 200K 90%) verbatim.

### REQ-RC-004 (Event-Driven, Mandatory)

WHEN `verification-batch-pattern.md` is loaded into the always-loaded rule context THEN the compressed body MUST link the 7-item canonical verification example **by cross-reference to `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution instead of inlining** AND MUST preserve the **Verification Class Taxonomy 8-row table** and **Default Grouping 5-row table** verbatim.

### REQ-RC-005 (Ubiquitous, Mandatory)

The system shall preserve cross-reference link integrity: every existing `.claude/rules/moai/...` and `.claude/output-styles/moai/moai.md` reference in the three compressed rule bodies MUST resolve to an existing path after compression (no dangling references).

### REQ-RC-006 (Unwanted, Mandatory)

The system shall NOT add or modify `paths:` frontmatter on the three compressed rule files (they remain always-loaded per intentional design — see `session-handoff.md` §Loading scope rationale).

### REQ-RC-007 (Ubiquitous, Mandatory)

The system shall preserve HISTORY / version footer / cross-reference list footers when present in the original files (e.g., `session-handoff.md` final `Status: HARD operational rule, applies to all multi-phase MoAI workflows` footer; `verification-batch-pattern.md` final `Version: 1.0.0 / Classification / Origin` block; `context-window-management.md` final `Status: HARD operational rule, applies to all sessions` footer).

## 5. Stakeholders

| 역할 | 책임 | 검증 책임 |
|------|------|----------|
| User (GOOS행님) | v3.0 환골탈태 목표 (-4.5K tokens) 승인 | Wave 1 통합 시 토큰 부담 체감 |
| manager-spec (본 SPEC 작성자) | EARS 요구사항 + AC 정의 | spec.md + acceptance.md self-audit |
| manager-develop (run-phase 위임 대상) | 3개 rule 본문 압축 실행 | AC PASS 검증 + HARD 조항 카운트 보존 |
| Future agents (always-loaded consumers) | 압축 후 본문에서 운영 지침 추출 | 압축 후 ambiguity 없는 retrieval 가능 |

## 6. Risks

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------:|-------:|-----------|
| R-RC-001 | 압축 과정에서 HARD 조항 의미가 paraphrase 되어 SSOT 위반 | Medium | High | AC-RC-002/003/004 grep 검증 + verbatim diff (압축 전후 HARD 조항 라인 정확히 동일 보장) |
| R-RC-002 | Cross-reference (`session-handoff.md` ↔ `context-window-management.md` ↔ `verification-batch-pattern.md` 상호 인용) 깨짐 | Low | Medium | AC-RC-005 `grep -F` link resolution 검증 + `.claude/output-styles/moai/moai.md` §6 / `feedback_large_spec_wave_split.md` / CLAUDE.md §11 등 외부 인용 보존 |
| R-RC-003 | `wc -w` 측정 시 YAML frontmatter / HISTORY / footer 포함 여부 모호로 baseline / target 불일치 | Low | Low | AC-RC-001/008 모두 **파일 전체 `wc -w`** 기준 단일 측정. Baseline 1,927w / 712w / 764w 도 동일 방법. |
| R-RC-004 | Compression 후 본문 retrieval 시 운영 지침 ambiguity 발생 (예: Trigger 5 항목 중 일부 단어 손실) | Low | High | REQ-RC-002/003 verbatim 보존 의무 → AC-RC-002 trigger 5-row table 모든 cell 보존 grep 검증 |
| R-RC-005 | `verification-batch-pattern.md` 7-item example 외부화 후 `agent-common-protocol.md` §Parallel Execution 갱신 시 동기화 누락 (SSOT 분리 risk) | Low | Medium | REQ-RC-004 cross-reference 의무 + `agent-common-protocol.md` 는 본 SPEC scope 외 (out-of-scope 명시) → 본 SPEC 후속 다른 SPEC이 7-item 변경할 때 sync 의무 별도 처리 |

## 7. Implementation Plan (inline, Tier S LEAN)

본 절은 Tier S LEAN workflow에 따라 `plan.md`를 생략하고 `spec.md` 내에 흡수.

### 7.1 Milestone Sequence

| Milestone | 산출물 | 검증 |
|-----------|--------|------|
| M1 — Baseline measure | 압축 전 `wc -w` + HARD/ZONE grep count 기록 | session-handoff: 1,927w/5 markers / context-window: 712w/5 markers / verification-batch: 764w/0 markers (HARD 0건이나 본 파일은 §Parallel Execution이 SSOT, 본 rule 은 reference doc) |
| M2 — `verification-batch-pattern.md` 압축 (가장 단순, dependency 없음) | 1 file edit | AC-RC-004/007 통과 |
| M3 — `context-window-management.md` 압축 | 1 file edit | AC-RC-003/007 통과 + threshold table verbatim |
| M4 — `session-handoff.md` 압축 (가장 큰, canonical artifact 다수) | 1 file edit | AC-RC-002/007 통과 + 5 Trigger table + 6-block format + Block 0 + Auto-Memory |
| M5 — Cross-reference 통합 검증 | grep audit report | AC-RC-005 0 dangling references |
| M6 — 종합 word count + token estimate | AC-RC-001/008 통과 | 합계 ≤ 2,300w (목표 1,900w) |

우선순위는 **단순도 오름차순** (verification-batch → context-window → session-handoff). HARD 조항 충돌 risk 낮은 파일부터 시작하여 압축 패턴 검증 후 가장 큰 파일에 적용.

### 7.2 Technical Approach (압축 기법)

1. **단축 우선순위**:
   - 1순위: Why this matters / Anti-pattern 카탈로그 / Detection heuristic 산문 prose 축약
   - 2순위: Example 블록 단축 (canonical format 자체는 보존하되 illustrative example 1개로 축소 가능)
   - 3순위: 중복 cross-reference 제거 (예: §Cross-references 와 본문 inline link 중복 시 §Cross-references 유지)

2. **보존 의무**: §3.2 enumerated artifacts (5 Trigger table / 6-block format / threshold table / class taxonomy / canonical example reference)

3. **단순 prose 압축 방식**:
   - 다단락 산문 → 2-3 sentence 핵심
   - Bulleted list 7+ items → 4-5 items + "etc." (단, HARD 조항 enumeration 은 100% 보존)
   - 중복 emphasis (예: 같은 HARD 조항 여러 곳 반복) → 1회만 유지 + 인용 link

### 7.3 Sequence 의무

- M2 → M3 → M4 순차 (단순 → 복잡)
- 각 milestone 종료 시 즉시 검증 (AC 부분 통과 확인)
- M6 종합 검증 전까지 commit 분리 또는 atomic MultiEdit 선택 가능 (manager-develop 판단)

## 8. Acceptance Criteria Reference

상세 AC는 `acceptance.md` 참조. 총 **8 binary ACs** (AC-RC-001 .. AC-RC-008) + 100% REQ↔AC traceability matrix + 3 Edge Cases 포함.

## 9. Cross-references

- 설계 문서: `.moai/research/v3.0-design-2026-05-22.md` §Layer 1 line 165-184 + line 358-365 (Wave 1 SPEC 카탈로그)
- Tier S workflow: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- Frontmatter SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Wave 0 패턴 참조: `.moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/spec.md` (Tier S structural pattern)
- 압축 대상 파일들:
  - `.claude/rules/moai/workflow/session-handoff.md`
  - `.claude/rules/moai/workflow/context-window-management.md`
  - `.claude/rules/moai/workflow/verification-batch-pattern.md`
- 보존 의무 SSOT 앵커: `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (verification 7-item canonical example)
- 관련 Wave 1 SPECs: `SPEC-V3R6-RULES-PATH-SCOPE-001` (path-scope 전환, 4건 별도) / `SPEC-V3R6-SKILL-CONSOLIDATE-001` / `SPEC-V3R6-SKILL-COMPRESS-001`

---

Version: 0.1.0
Tier: S (LEAN — 2 artifacts: spec.md + acceptance.md, plan.md inline 흡수 §7)
Status: draft
