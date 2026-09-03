# SPEC-GRAPH-GATE-RESTAMP-001 — 진행 기록 (카드 t478)

## §E.1 Plan-phase Audit-Ready Signal

- 저작 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t478`, 브랜치 `WT-graph-gate-restamp`,
  HEAD `456665e8d` (plan-phase 저작 시점).
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID 정규식 자체검사: `[[ "SPEC-GRAPH-GATE-RESTAMP-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`
  → `PASS`.
- plan-phase에서 이 트리에서 직접 측정한 값(`spec.md` §B.2 / §B.3):
  - `git log -1 --format=%H ad272be20 -- .moai/project/codemaps/ ':(exclude).moai/project/codemaps/provenance.json'`
    → `2f28bc394718c2cb49a6dc96577d12c6da35b05d`
  - `git merge-base --is-ancestor 2f28bc394 ad272be20` → rc 0
  - `git diff --name-only ad272be20 -- internal cmd pkg | wc -l` → `405`
  - `git diff --name-only 2f28bc394 -- internal cmd pkg | wc -l` → `570`
  - `git diff --name-only ad272be20 -- .moai/project/codemaps/ ':(exclude)…/provenance.json'`
    → 본문 6개 파일 출력 (규칙 A 프로브 발화 — `spec.md` §D.2)
- 레인 결정 2건 반영 완료 (2026-09-04). **미해소 항목 없음** — `[NEEDS CLARIFICATION]` 표식 0개.
  1. 규칙 A `content_anchor_source` 토큰 → `working-tree-differs-from-stamp` (로직 불변; `spec.md` §D.2).
  2. 규칙 A 프로브 → `git diff` ∪ `git ls-files --others --exclude-standard` (`spec.md` §B.4, §D.1).
     근거 실측: `internal/graph/check_test.go` `writeCodemapsProvenance`가 base 커밋 이후
     `.moai/project/codemaps/modules.md`를 `os.WriteFile`로 쓰고 커밋하지 않음 → 기존 픽스처 본문은 미추적.
     정정 전이라면 규칙 A·B 모두 발화하지 않아 규칙 C(absent)로 떨어져 AC-5가 구조적으로 실패.
- 인수 기준 신설: **AC-7** (미추적 본문 픽스처 → 규칙 A로 해석, verdict는 `absent`가 아님). 기존 AC 번호
  재부여 없음 (AC-1~AC-6 그대로).
- 코드는 작성하지 않았다. `provenance.json`은 만지지 않았다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
