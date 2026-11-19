---
id: SPEC-20261119-c71-evidence
title: "C-71 evidence: the refusal-list follow-ups are already met on the integration tip (PR #3169)"
status: draft
revision: 1
created: 2026-11-19
---

# C-71 evidence (PR #3169)

## Intent

카드 C-71 (p3, C-69 감사 비차단 R2-N1·R2-N2 후속) 을 통합 가지 tip(`9bd09c2effb91e9a6901f6d81d8e876d1e484d15`)에서 확인한 결과, 두 기준 모두 **이미 PR #3169**(`spec/20261117-versioning-refusals`, 병합 2026-11-17)과 그 감사 수리 커밋들(`5c87e5b8e9`, `a091e3e97a`)로 들어와 있어요. 이 SPEC 은 새 코드 없이 그 증거를 기록해요(리더 배차의 "이미 됐으면" 경로).

## Non-goals

- 새 구현·시험 — tip 에 완성돼 있어요.
- 이 증거 SPEC 외 경로 쓰기(축소 take: `.moai/specs/SPEC-20261119-c71-evidence/**` 만).

## Requirements

- REQ-1: When the custom-text refusal rules are read from the internal documents on the integration tip, the list shall name `/` and `\` beside the default-ignorable set. · failure: none — 위반 시 PR #3169 가 추가한 문서·시험이 어긋나요; 이 SPEC 은 tip 증거 기록이에요.
- REQ-2: When a custom versioning text holds any default-ignorable code point, the `custom` check shall refuse it. · failure: none — 위반 시 PR #3169 의 `DEFAULT_IGNORABLE` 시험과 감사 수리(a091e3e97a)가 어긋나요.

## Acceptance

- AC-1 (REQ-1): tip 문서 증거 — `internal/versioning.md:5` 와 `internal/interview.md:39` 가 둘 다 `/`, `\` 와 기본 무시 목록을 적어요(`internal/project_plan.md` 는 tip 에 없는 파일 — 통합됐어요).
  verify: `git show origin/v2026.11.0:skills/moai/references/internal/versioning.md | sed -n 5p`
  verify: `git show origin/v2026.11.0:skills/moai/references/internal/interview.md | rg -n "backtick"`
- AC-2 (REQ-2): tip 코드 증거 — `src/versioning.rs` 의 `DEFAULT_IGNORABLE`(Unicode 18.0.0, 17 범위, U+00AD–U+E0FFF)이 감사가 지적한 범위 전부(U+034F, U+115F–1160, U+17B4–17B5, U+180B–180F, U+3164, U+FFA0, U+FFF0–FFF8, U+1BCA0–1BCA3, U+1D173–1D17A, U+E0000–E0FFF)를 품고, 시험이 바깥 이웃·지적 코드점 거부를 못박아요.
  verify: `git show origin/v2026.11.0:src/versioning.rs | rg -n "DEFAULT_IGNORABLE|034F|E0000"`

## Scope

- write: `.moai/specs/SPEC-20261119-c71-evidence/**`
- read-only: `src/versioning.rs`, `src/profile.rs`, `skills/moai/references/internal/versioning.md`, `skills/moai/references/internal/interview.md`, PR #3169
- forbidden: 이 SPEC 외 쓰기, 새 구현

## Plan

1. tip 확인(완료): 두 AC 모두 PR #3169 로 충족 — 위 verify 줄.
2. 축소 take(SPEC 폴더) → 이 SPEC → push(PR 없음) → 판정 기준 met → 근거 PR #3169 연결 → card done.

## Risks

- 증거 SPEC 이므로 새 위험 없음; PR #3169 의 감사(수리 2건 포함)가 이미 지나갔어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "이미 됨" 경로(리더 배차 1단계) — 구현 없이 증거 기록만 | tip `9bd09c2eff` 에서 두 AC 확인(PR #3169 + 감사 수리) | 구현 카드로 재발행 |

## Result

- AC-1: met — tip 에서 `internal/versioning.md:5`(“no `<`, `>`, backtick, `[`, `]`, `*`, `_`, `/` or `\` … Default_Ignorable_Code_Point … U+E0000–U+E0FFF”)과 `internal/interview.md:39`(`_`, `/`, `\` 명시) 확인(2026-11-19T02:06Z, `git show origin/v2026.11.0:…`). 근거 PR #3169.
- AC-2: met — tip `src/versioning.rs:716` `DEFAULT_IGNORABLE`[(u32,u32);17](U+00AD–U+E0FFF, Unicode 18.0.0 발췄 파일 시험 포함)이 지적 범위 전부를 담고, `:1519` 시험이 지적 코드점 거부를, `:1448` 주석이 이웃 수(32)를 못박아요. 감사 수리 커밋 5c87e5b8e9·a091e3e97a(Cf 전체·이웃 보정). 근거 PR #3169.
