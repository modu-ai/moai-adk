# Card t701 Verdict — 루트 .gitignore 171~174 스테일 규칙 정리 (B로 강등 실행)

- Date: 2026-09-14 · Lane: lane-4 · Branch: `WT-gitignore-stale` @ d3ba6f422 (base 9f5978d67, unpushed)
- Class: 배차 시 A 후보 → **실행 시 B로 강등** (A 요건 중 "CI 초록 head 인용" 불가 — 미병합 로컬 브랜치, CI 판정은 리드 develop push 후 원격 몫)

## Claim

루트 `.gitignore`의 `.agents/*`(133행)는 슬래시 포함 패턴이라 루트에 앵커드 — `internal/template/templates/.agents/`는 애초에 무시 대상이 아니었다. 따라서 171~174행(3줄 코멘트 + `!internal/template/templates/.agents/` 재포함)은 **no-op이며, "bare .agents/ 규칙이 템플릿 하위트리를 배제한다"는 코멘트는 패턴 의미를 오설명**. 4행 삭제로 정리했고 추적 파일 불변을 실측했다.

## Evidence

| 검증 | 명령 | 관측 |
|---|---|---|
| 무시 상태(삭제 전) | `git check-ignore -v internal/template/templates/.agents/skills/moai-plan/SKILL.md` | 출력 없음, exit 1 = 무시 아님 → 174행 재포함이 이미 no-op |
| diff 크기 | `git diff --stat` | `.gitignore \| 4 ----` (1파일 4행 삭제 — 배차 명세와 일치) |
| 추적 파일 불변 | `git ls-files .agents/ \| wc -l` / `git ls-files internal/template/templates/.agents/ \| wc -l` | **16 / 16** (불변) |
| 무시 상태(삭제 후) | `git check-ignore internal/template/templates/.agents/skills/moai-plan/SKILL.md` | exit 1 = 여전히 무시 아님 — 동작 불변 |
| 작업 트리 | `git status --short` | `.gitignore`만 수정 |

## Baseline-attribution

모든 관측은 lane-4 세션에서 이 워크트리(base 9f5978d67)에서 실행. CI 판정은 리드 push 후 원격 몫(B 강등 사유).

## Gaps

- CI 초록 인용 불가(위 강등 사유) — develop push 후 원격 CI가 전체 판정
- `templates/.gitignore`(템플릿 내부)의 자체 규칙은 이 카드 범위 밖(그 쪽 16명 재포함은 SPEC-CODEX-COMMAND-SKILLS-001 소관, 미접촉)

## Residual-risk

- 없음에 수렴: 삭제 대상이 no-op임을 양방향 check-ignore로 고정했고, 미래 파일도 루트 패턴이 앵커드라 무시될 수 없는 구조(패턴 읽기로 확인)
