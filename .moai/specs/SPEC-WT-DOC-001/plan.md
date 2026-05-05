---
id: SPEC-WT-DOC-001
plan_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Implementation Plan — SPEC-WT-DOC-001

## 1. Overview

`.claude/rules/moai/workflow/worktree-integration.md`에 Shared State Policy / Concurrency Model / Anti-Patterns / Termination Conditions 절 4종 추가. 코드 변경 0, 정책 강화만으로 worktree shared state 위험 감소.

## 2. Approach Summary

**전략**: Documentation-Only, Incident-Based-Patterns, Cross-Reference-Strong.

1. v2.14.0 case study (CLAUDE.local.md §18.11)에서 5 anti-pattern 도출
2. SPEC-LOOP-TERM-001 (Wave 2) termination schema 의미 확장 인용
3. CLAUDE.md §14 + 본 문서 양방향 cross-ref
4. Template-First 동기화

## 3. Milestones (Priority-based, no time estimates)

### M0 — Pre-flight (Priority: Critical)

- [ ] `.claude/rules/moai/workflow/worktree-integration.md` 현재 내용 verbatim 캡처
- [ ] CLAUDE.md §14 Worktree Isolation Rules verbatim 캡처
- [ ] CLAUDE.local.md §18.11 v2.14.0 case study 사례 추출
- [ ] SPEC-LOOP-TERM-001 (Wave 2) termination schema 확인
- [ ] 본 프로젝트의 worktree 사용 사례 5건 발굴 (PR 검색)

**Exit Criteria**: 정책 문서 baseline + incident 데이터 확보

### M1 — Shared State Policy 절 작성 (Priority: High)

- [ ] 새 절 "Shared State Policy" 추가:
  - per-file ownership 정의 (예: `.moai/specs/<ID>/spec.md`은 SPEC author 소유)
  - writer-of-last-resort: PR merge → main이 최종 권위
  - eventual consistency model 명시
  - 동시 쓰기 정책: worktree A가 worktree B의 파일에 직접 쓰기 금지
- [ ] 사례 표 작성 (file → owner → write timing)

**Exit Criteria**: Shared State Policy 절 완성

### M2 — Concurrency Model 절 작성 (Priority: High)

- [ ] 새 절 "Concurrency Model" 추가:
  - "Consistency: eventual via PR merge to main"
  - "No direct cross-worktree file writes"
  - "Worktree-local writes are isolated and safe"
  - "PR merge is the synchronization barrier"
- [ ] 결정 트리 (간단 flowchart): "shared file 변경 필요? → 자기 worktree 내에서 → PR로 main에"

**Exit Criteria**: Concurrency Model 절 완성

### M3 — Anti-Patterns 카탈로그 (Priority: High)

- [ ] 새 절 "Anti-Patterns" 추가, 5+ 항목:

  **AP-1: Direct cross-worktree write**
  - Symptom: worktree A에서 `../wave-B/.moai/specs/.../progress.md` 직접 쓰기
  - Why bad: race condition, git semantics 위반
  - Mitigation: 자기 worktree 내 작업 → PR merge

  **AP-2: Concurrent SPEC file modification**
  - Symptom: 두 worktree에서 동일 SPEC의 spec.md 동시 변경
  - Why bad: PR conflict, manual merge 비용
  - Mitigation: SPEC ownership 명확화 (`.moai-worktree-registry.json`)

  **AP-3: Unterminated reactive loop**
  - Symptom: shared state 무한 재읽기
  - Why bad: SPEC-LOOP-TERM-001 schema 위반
  - Mitigation: max_iterations + improvement_threshold 적용

  **AP-4: Schema drift across worktrees**
  - Symptom: progress.md 형식이 worktree마다 다름
  - Why bad: fan-in 실패, automation 오작동
  - Mitigation: SPEC-CONTEXT-INJ-001 권장 schema 준수

  **AP-5: Implicit shared mutable state**
  - Symptom: `.moai/state/` 파일을 cross-worktree에서 읽기
  - Why bad: undefined behavior, 동기화 미보장
  - Mitigation: state 파일은 worktree-local, 공유는 PR 경유

- [ ] 각 anti-pattern에 concrete code/operation example

**Exit Criteria**: 5+ anti-pattern entries 완성

### M4 — Termination Conditions 절 (Priority: Medium)

- [ ] 새 절 "Termination Conditions" 추가:
  - SPEC-LOOP-TERM-001 인용 (verbatim 발췌)
  - max_iterations / improvement_threshold / escalation_after 의미 확장
  - shared state context: "loop가 shared state에 의존할 때 termination 명시 의무"
- [ ] cross-ref to SPEC-LOOP-TERM-001 (Wave 2)

**Exit Criteria**: Termination 절 완성 + LOOP-TERM-001 cross-ref

### M5 — Cross-reference 갱신 (Priority: High)

- [ ] `CLAUDE.md` §14 "Parallel Execution Safeguards"에 추가:
  > "Worktree shared state 정책: .claude/rules/moai/workflow/worktree-integration.md §Shared State Policy 참조"
- [ ] worktree-integration.md 도입부에 cross-ref:
  - "termination schema는 SPEC-LOOP-TERM-001 참조"
  - "context injection은 SPEC-CONTEXT-INJ-001 참조"
- [ ] Template-First 동기화

**Exit Criteria**: 양방향 cross-ref 검증

### M6 — Living Document 마킹 (Priority: Low)

- [ ] 문서 도입부에 living document 마킹
- [ ] "분기별 review 의무, 신규 anti-pattern 등재 절차 명시"
- [ ] HISTORY 섹션 추가 (현재 SPEC 기록)

**Exit Criteria**: living document 마커 + HISTORY 작성

### M7 — Validation + Acceptance Sign-off (Priority: High)

- [ ] acceptance.md 시나리오 모두 PASS
- [ ] anti-pattern count >= 5 검증
- [ ] cross-ref (CLAUDE.md, LOOP-TERM, CONTEXT-INJ) 모두 검증
- [ ] code change = 0 (Go files diff = 0) 검증
- [ ] Template-First sync clean
- [ ] plan-auditor PASS

**Exit Criteria**: 모든 acceptance + plan-auditor PASS

## 4. Technical Approach

### 4.1 새 절 4개 구조

```markdown
## Shared State Policy

### Per-file Ownership Matrix
| File | Owner | Write Timing | Sharing |
|------|-------|--------------|---------|
| spec.md | SPEC author | Plan phase only | read-only after plan |
| plan.md | SPEC author | Plan phase | read-only after plan |
| acceptance.md | SPEC author | Plan phase | read-only after plan |
| progress.md | active runner | Run phase | worktree-local until PR |
| _status.json | active runner | hook-driven | worktree-local |

### Writer-of-Last-Resort
PR merge to `main` is the synchronization barrier and final authoritative writer.

## Concurrency Model

[Mermaid flowchart TD]

## Anti-Patterns

### AP-1: Direct cross-worktree write
... (M3)

## Termination Conditions

For shared state with reactive loops, termination conditions MUST be defined per
SPEC-LOOP-TERM-001 schema:
- max_iterations
- improvement_threshold
- escalation_after
```

### 4.2 cross-ref 텍스트 표준 (CLAUDE.md §14)

```markdown
For shared state policy across worktrees, see
`.claude/rules/moai/workflow/worktree-integration.md` §Shared State Policy.
```

## 5. Risks and Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| 문서만으로 enforcement 불가 | High | Medium | living document + 분기별 audit, 별도 SPEC으로 hook 차단 후속 |
| 5 anti-pattern 가공 우려 | Medium | Medium | M0에서 incident-based 도출 강제 (CLAUDE.local.md §18.11 인용) |
| LOOP-TERM-001 의미 충돌 | Low | Medium | LOOP-TERM-001은 인용만, 요건 추가 금지 |
| cross-ref 누락 | Low | Medium | M5 checklist 명시 (CLAUDE.md, LOOP-TERM, CONTEXT-INJ 3건) |
| 사용자 학습 곡선 | High | Low | 결정 트리 flowchart, 5 사례 명시 |

## 6. Dependencies

- 선행 SPEC: 없음 (독립)
- 의존 (인용): **SPEC-LOOP-TERM-001** (Wave 2) — termination schema source
- sibling SPEC: SPEC-CONTEXT-INJ-001 (이번 wave) — schema cross-ref
- 도구: `make build`, plan-auditor

## 7. Open Questions Resolution

- **OQ1** (anti-pattern 자동 감지): 본 SPEC scope 외, 별도 SPEC 후보 (plan-auditor 확장)
- **OQ2** (frontmatter `shared_state_termination`): 본 SPEC 외, 추가 의미 부담 회피
- **OQ3** (cross-worktree merge 충돌 해결): "PR merge with manual review" — Git 기본 정책 명시
- **OQ4** (위반 강도): warning (text guide), 강제 enforcement 별도 SPEC

## 8. Rollout Plan

1. M1-M5 작업 → 본 프로젝트의 다음 worktree 사용 PR에 적용 (dogfooding)
2. CHANGELOG entry (Unreleased)
3. v2.x.0 minor release (코드 변경 없으니 patch도 가능)
4. 분기별 review에서 anti-pattern 카탈로그 확장

End of plan.md (SPEC-WT-DOC-001).
