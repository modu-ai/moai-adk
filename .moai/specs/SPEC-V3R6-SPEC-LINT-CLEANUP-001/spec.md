---
id: SPEC-V3R6-SPEC-LINT-CLEANUP-001
title: "spec-lint MissingExclusions baseline cleanup — H3 sub-heading retroactive application to 8 sibling SPECs"
version: "0.1.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, missing-exclusions, baseline-cleanup, h3-pattern, retroactive, tier-s"
sync_commit_sha: "<pending>"
---

# SPEC-V3R6-SPEC-LINT-CLEANUP-001 — spec-lint MissingExclusions baseline cleanup

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | GOOS행님 | 최초 작성 (plan-phase). Tier S minimal Section A 단일 섹션 변형. ARR-001 + PROPOSAL-GEN-001가 확립한 `### N.M Out of Scope — <topic>` H3 sub-heading 패턴을 8개 baseline-failing sibling SPEC에 소급 적용하기 위한 정리 SPEC. Sprint 8 P4 baseline cleanup 진입점. |

## §1. 목적 (Goal)

`spec-lint` 도구의 `MissingExclusions` (`internal/spec/lint.go` `OutOfScopeRule`) 규칙이 **8개 sibling SPEC**에서 ERROR로 baseline failure를 발생시킨다. 본 SPEC은 8개 SPEC의 `spec.md`에 canonical H3 sub-heading 패턴을 소급 적용하여 baseline failure를 0으로 떨어뜨린다. 본 SPEC의 plan-phase는 (a) 8개 sibling SPEC 식별과 (b) canonical H3 패턴 정의에 집중하며, (c) 실제 edit 작업은 후속 run-phase로 명시 연기한다.

## §2. 배경 (Background)

### §2.1 lint rule mechanics

`internal/spec/lint.go:678-728` `OutOfScopeRule.Check()` 검사 알고리즘:

1. `spec.md` body를 lowercase로 변환 후 `"out of scope"` 문자열 포함 여부 확인. 없으면 `MissingExclusions` ERROR ("section missing").
2. `###`로 시작하면서 `"out of scope"` 포함하는 H3 sub-heading 검색. 발견 시 `inOutOfScope = true`.
3. 다음 `##` H2 등장 전까지 `-`로 시작하는 list item을 1개 이상 발견하면 PASS, 0개면 `MissingExclusions` ERROR ("no items").

### §2.2 8 sibling SPEC baseline failure (2026-05-25 `moai spec lint` 실행 결과)

| # | SPEC ID | 실패 메시지 | 분류 |
|---|---------|-----------|------|
| 1 | SPEC-V3R6-CI-BASELINE-DRIFT-001 | 'Out of Scope' section missing | A — `## Exclusions` 사용 중, "out of scope" 문자열 부재 |
| 2 | SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 | section has no items | B — H3 sub-heading 부재 또는 list item 부재 |
| 3 | SPEC-V3R6-LEGACY-CLEANUP-001 | section has no items | B |
| 4 | SPEC-V3R6-LEGACY-CLEANUP-002 | section has no items | B |
| 5 | SPEC-V3R6-LEGACY-CLEANUP-003 | section has no items | B |
| 6 | SPEC-V3R6-PROMPT-CACHE-001 | section has no items | B |
| 7 | SPEC-V3R6-SESSION-HANDOFF-AUTO-001 | section has no items | B |
| 8 | SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001 | 'Out of Scope' section missing | A — `## §2 Non-Goals` 사용 중, "out of scope" 문자열 부재 |

분류 A (2개): "out of scope" 문자열 자체가 spec.md body에 등장하지 않음. H3 sub-heading 신규 추가 필요.
분류 B (6개): H3 sub-heading은 있으나 그 아래 `-` list item이 비어 있거나, "out of scope" 단어가 등장하지만 H3 형식이 아니거나 `##` H2 뒤 list item이 다음 `##`까지 가지 못함. 본 SPEC plan-phase의 정밀 분류는 run-phase에서 파일별 진단으로 확정한다.

**Parallel session drift note (2026-05-25 plan-phase 중 발견)**: plan-phase 작성 중 parallel session이 2개 신규 SPEC을 작성하면서 `MissingExclusions` failure를 추가 발생시켰다 (SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 + SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001). 본 SPEC scope는 **2026-05-25 plan-phase 시작 시점 baseline 8개**에 명시적으로 한정한다. 신규 2개 SPEC은 (a) 작성자(parallel session)가 작성 시점에 §3 canonical pattern을 직접 적용하거나, (b) 본 SPEC run-phase 시점 재진단으로 scope 확장 결정 — 두 path 모두 별도 사이클에서 처리한다 (§5.3 참조).

### §2.3 Canonical H3 패턴 선례

- **SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001** (`spec.md` 라인 102): `### §B.2 Out of Scope (deferred to follow-up SPECs or explicitly NOT done)` H3 + 9개 `-` list item. lint rule PASS.
- **SPEC-V3R6-HARNESS-PROPOSAL-GEN-001** (`spec.md` 라인 158-183): `## §7. Out of Scope (Sub-section for spec-lint MissingExclusions compliance)` H2 + 5개 `### §7.1..§7.5 Out of Scope — <topic>` H3 sub-section 각각 하나 이상의 `-` list item. lint rule PASS.

두 선례 모두 다음 조건을 만족한다: (1) H3 heading에 "Out of Scope" 문자열 포함, (2) H3 아래 다음 H2 이전까지 1개 이상의 `-` list item 존재. 본 SPEC은 두 선례의 공통 패턴을 canonical pattern으로 codify한다.

## §3. Canonical H3 패턴 정의

본 SPEC이 8개 sibling SPEC에 소급 적용하기 위한 표준 패턴은 다음과 같다.

### §3.1 Minimum form (모든 cleanup 대상 SPEC에 적용)

```markdown
## <existing H2 — e.g., "Exclusions" or "Non-Goals" or "§N. Out of Scope">

### <N.1> Out of Scope — <SPEC-specific topic>

- <Item 1 — non-empty exclusion description>
- <Item 2 ...> (optional, ≥1 required)
```

조건:
- H3 heading 텍스트에 **literal "Out of Scope" 문자열** (case-insensitive) 포함.
- H3와 다음 `##` H2 사이에 **`-`로 시작하는 list item 최소 1개** 존재.
- list item 본문은 비어 있지 않아야 함 (`-` 단독 라인 금지).

### §3.2 분류 A SPEC (2개) — H2 + H3 신규 추가

기존 `## Exclusions` 또는 `## §N Non-Goals` 섹션의 H2 텍스트를 유지하되, 그 아래 `### N.1 Out of Scope — <topic>` H3 sub-heading을 추가하고 기존 list item을 그 아래로 이동(또는 복제) 한다.

### §3.3 분류 B SPEC (6개) — H3 sub-heading 보강 또는 list item 추가

진단 결과에 따라:
- (B-i) `## §C Exclusions` 같은 H2는 있지만 그 아래 `### Out of Scope ...` H3가 없는 경우: H3 sub-heading 추가 + 기존 H2 list item 복제 또는 이동.
- (B-ii) `### §X Out-of-scope ...` H3는 있지만 그 아래 list item이 모두 다음 `##` 뒤로 밀린 경우: H3 직속에 list item 1개 이상 추가.
- (B-iii) H3 텍스트가 "out of scope" 문자열을 포함하지 않는 경우(예: `### N.M Non-Goals`): H3 텍스트에 "Out of Scope" 키워드 병기.

### §3.4 self-compliance — 본 SPEC도 패턴 준수

본 spec.md는 자기 정의 패턴을 준수한다. 본 §7 OutOfScopeRule 섹션이 `### §7.1 Out of Scope — Run-phase edits`를 포함하며 list item 1개 이상 존재한다 (§7 참조).

## §4. Requirements (EARS Format)

### REQ-SLC-001 [Ubiquitous] Sibling SPEC enumeration
본 SPEC plan-phase의 `plan.md §A`는 2026-05-25 `moai spec lint` 출력 기반의 8개 sibling SPEC 명단을 명시 SHALL 한다.

### REQ-SLC-002 [Ubiquitous] Canonical H3 pattern codification
본 SPEC의 `spec.md §3`은 lint rule을 통과하는 minimum form (§3.1)을 코드 블록과 함께 명문화 SHALL 한다.

### REQ-SLC-003 [Ubiquitous] Classification taxonomy
본 SPEC의 `spec.md §2.2`는 8개 SPEC을 분류 A(missing) / 분류 B(no items)로 일관되게 분류 SHALL 한다.

### REQ-SLC-004 [State-Driven] Run-phase scope contract
WHILE run-phase가 본 SPEC을 구현하는 동안, run-phase는 §2.2에 명시된 8개 sibling SPEC의 `spec.md`만 수정 SHALL 한다 (그 외 SPEC 또는 plan/acceptance/progress 파일 수정 금지).

### REQ-SLC-005 [Unwanted] No body content semantic change
본 SPEC의 run-phase는 8개 sibling SPEC의 기존 exclusion 의미(semantic content)를 변경 SHALL NOT 한다. H3 sub-heading 추가와 list item 위치 조정은 허용되지만 list item 본문 텍스트의 의미 변경 금지.

### REQ-SLC-006 [Event-Driven] lint baseline verification
WHEN run-phase가 모든 8개 SPEC 수정을 완료한 직후, `moai spec lint 2>&1 | grep -c MissingExclusions`가 **0**을 출력 SHALL 한다.

### REQ-SLC-007 [Optional] Self-compliance reinforcement
WHERE 가능한 경우, 본 SPEC의 spec.md는 자기 정의 패턴(§3)을 모범적으로 준수 SHOULD 하여 향후 SPEC 작성자에게 reference 역할 수행.

## §5. Out of Scope (자기 준수 - §3 self-compliance per REQ-SLC-007)

### §5.1 Out of Scope — Run-phase edits (deferred to follow-up implementation)

본 SPEC의 plan-phase는 다음을 명시적으로 out of scope로 둔다:

- **8개 sibling spec.md 파일의 실제 H3 sub-heading 추가 edit** — run-phase에서 manager-develop이 수행. 본 plan-phase는 어떤 edit도 수행하지 않는다.
- **8개 SPEC의 plan.md, acceptance.md, progress.md 수정** — 영향 범위 밖. lint rule은 `spec.md`만 검사하므로 sibling 산출물은 untouched.
- **8개 SPEC의 frontmatter 변경** — version bump, updated 갱신 등은 본 정리 SPEC scope 외.

### §5.2 Out of Scope — Lint rule code modification

`internal/spec/lint.go` `OutOfScopeRule` Go 코드 자체는 본 SPEC scope 외. 본 SPEC은 8개 SPEC을 rule에 적합화하는 방향이지, rule 자체를 완화하지 않는다. 만약 rule이 과도하게 엄격하다는 판단이 든다면 별도 SPEC에서 다룬다.

### §5.3 Out of Scope — 시간 외 발견되는 baseline failure

만약 본 SPEC의 plan-phase 또는 run-phase 진행 중 8개 외의 새로운 `MissingExclusions` failure가 등장(parallel session에서 신규 SPEC 작성 시)하면, 해당 신규 SPEC은 작성자가 spec.md 작성 시점에 직접 §3 canonical pattern을 따라야 한다. 본 SPEC scope는 2026-05-25 baseline의 8개 SPEC에 한정.

### §5.4 Out of Scope — Other lint rule failures

본 SPEC은 `MissingExclusions` 단일 rule에만 집중한다. 다른 spec-lint rule(`FrontmatterInvalid`, `BreakingChangeMissingID`, etc.)의 baseline failure 정리는 별도 SPEC scope.

## §6. Dependencies

- **Required (reference precedent)**: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 (`a25476e7e` merge) + SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (`b47eb4428` merge). 두 SPEC이 canonical H3 패턴의 운영 검증을 완료했다.
- **Lint tool**: `~/go/bin/moai spec lint` (moai-adk v3.0.0-rc1). 본 SPEC plan-phase 진단 시점 baseline = 8개 ERROR.
- **No code dependencies**: 본 SPEC은 `.go` 코드 수정 없이 markdown 파일만 다룬다. run-phase는 Go test 영향 zero.

## §7. Cross-references

- `internal/spec/lint.go:678-728` — `OutOfScopeRule.Check()` 알고리즘.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter SSOT (본 SPEC의 12-field frontmatter 준수).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S minimal 정의.
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate 책임자.
- MEMORY.md `Sprint 8 ARR-001 4-phase CLOSE` 엔트리 — 8개 sibling MissingExclusions baseline cleanup이 본 SPEC scope임을 명시한 다음-SPEC 후보.
