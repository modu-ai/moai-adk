# t401 preflight — 카드 진입 전 상태

card: t401
issue: modu-ai/moai-adk#1683 (Seung-zedd, 2026-08-30, type:feature)
worktree: .claude/worktrees/t401
branch: WT-analysis-pull
base: origin/develop @ ad272be20 (origin/main 7ad9f8534 에서 ff-only merge, 979커밋 흡수, 0 ahead)

## 선행: t336 워크트리 폐기 (리드 승인 2026-09-02)

| 검증 | 명령 | 관측 |
|---|---|---|
| 원격 착지 | `git merge-base --is-ancestor 538f9dc77 origin/develop` | rc=0 |
| WT HEAD 착지 | `git merge-base --is-ancestor 81ece2481 origin/develop` | rc=0 |
| 미푸시 | `git -C .claude/worktrees/t336 rev-list --count origin/develop..HEAD` | 0 |
| 미커밋 | `git status --porcelain` | `?? .moai/reports/t336/verdict.md` 1건 (primary 사본과 바이트 동일) |

증거 반출: 워크트리에만 있던 6건(plan-audit.md · plan-audit-iter2.md · preflight.md ·
spec-lint.txt · spec-lint-iter2.txt · sync-audit.md)을 primary `.moai/reports/t336/` 로 복사.
verdict.md 는 이미 primary 에 동일 바이트로 존재(`diff -q` 일치).

폐기 실행: `git worktree remove --force` → 레지스트리 항목은 제거됐으나 디렉터리 잔존
(`Directory not empty`). 원인은 자기 세션 statusline 이
`.moai/state/context-usage/<own-session>.json` 을 재생성하는 레이스 + gopls(pid 2780) cwd 점유.
잔여 3파일 전부 `.moai/state/` 하위 휘발성 — 비휘발 산출물 0건 확인
(`find .moai -type f | grep -vE 'cache|logs|state'` → 0행). 재시도 `rm -rf` rc=0,
`ls -d .claude/worktrees/t336` → No such file or directory.

**잔재 보고(삭제하지 않음)**: 로컬 브랜치 `WT-integration-lock-atomic` 존속.
primary 에서의 `git branch -d` 는 BranchGuard 가 거부 — t264 소관.

## 이슈 #1683 원문 판독

첨부 문서 `moai-adk-human-decision-authority-feature-issue.md` (10,209바이트) 전문 확인.

제안의 뼈대:

```
Stage 1 Mechanical Verification
  ↓
Stage 2 Adversarial Interrogation
  ↓
Decision Index / Authority Routing (DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER)
  ↓
Human Decision Gate
  ↓
Stage 3 Apply Confirmed Verdict
```

제안이 명시한 조건 3개:

- Detect → Explain → Ask, but never decide.
- An LLM "best practice" is not a policy.
- When uncertain, escalate. Never downgrade.

제보자 본인이 붙인 호환성 단서: 설계·구현은 **v3.1.2 이전 legacy EARS 기반**이며 그대로 적용을
제안하는 것이 아니다. 이번 이슈에서 제안하는 것은 구현이 아니라 **Human Decision Authority
모델**이고, 방향성에 대한 maintainer 의견을 먼저 묻고 있다.

## 리드가 못박은 채택 순서 [HARD]

2번(Analysis is Pull, Not Push — anchoring 제거)을 **먼저 좁게** 시험한다. 구조 변경 없이
출력 규약만으로 상당 부분 얻어지고 효과가 즉시 관측된다. 1번(SPEC Review 내부 게이트)은
파이프라인 구조 변경이라 별건.

→ 이 카드의 범위는 **pull 전환** 하나다.
