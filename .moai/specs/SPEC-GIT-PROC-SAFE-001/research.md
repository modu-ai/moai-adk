# SPEC-GIT-PROC-SAFE-001 — 조사 기록 (reproduction record)

> 본 문서의 모든 측정은 **본 워크트리(develop `c9ceff175`), 2026-09-14, 본 세션의 직접 파일 판독**으로 수행됐다. 명령 감사의 원 baseline은 main `2213871af`이지만, 아래 상태는 재검증치이며 가설이 아니다.

## 측정 환경

| 항목 | 값 |
|------|-----|
| 감사 baseline | main `2213871af` |
| 재검증 baseline | develop `c9ceff175` (본 워크트리 HEAD) |
| 측정 방법 | Read 도구 직접 판독 (행 범위 지정) + 사본당 `auto-merge` 대소문자 무시 적중 계수 |

## AC-01 (P1) — LIVE, 본 작업 핵심

`manager-git.md` § Late-Branch Invocation Pattern이 primary 공용 체크아웃에서 브랜치 상태를 바꾸는 수순을 지시한다(`AGENTS.md` §2 위반).

관측 (로컬 `.claude/agents/moai/manager-git.md` / 템플릿 `internal/template/templates/.claude/agents/moai/manager-git.md`):

| Phase | 명령 | 로컬 행 | 템플릿 행 |
|-------|------|---------|-----------|
| A | `git checkout main && git pull origin main` | 98 | 96 |
| C | `git switch -c feat/SPEC-XXX` | 112 | 110 |
| D | `git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main` | 121-124 | 119-122 |
| 복구 | `git fetch origin && git reset --hard origin/main` | 129 | 127 |
| Personal 옵션 | `main_late_branch` ... `reset --hard origin/main` cleanup | 139 | 137 |

SSOT 궤적 확장 (`.claude/rules/moai/workflow/spec-workflow.md`):
- 50행: Route B Late-branch 사전조건 — "Step 1 entry requires `git rev-parse --abbrev-ref HEAD == main`... plan-phase commits land directly on `main` and are pushed only after Phase C `git switch -c plan/SPEC-XXX` at PR creation time."
- 53-62행: Step 4 Late-branch closure — 정본 closure 명령 블록(`git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main`) + post-condition(`git status --porcelain` empty AND `git rev-parse main == origin/main`) + "For the complete 4-phase Late-branch invocation pattern (A→D), see `.claude/agents/moai/manager-git.md`" 교차참조.
- manager-git.md 131행은 spec-workflow.md에 "canonical step ordering"을 위임. → manager-git.md만 고치면 참조 정본에 모순이 남는다. **이 확장 범위는 의도적·문서화됨**(spec.md REQ-GP-003).

대체 모델의 기계적 근거: 하나의 브랜치는 하나의 워크트리에만 체크아웃 가능 → "main에 커밋 적립" 모델은 워크트리 격리와 원리적으로 양립 불가. 대체 모델 = "워크트리 자기 브랜치에 커밋 적립 → 브랜치를 구성된 통합 경로로 승격"(승격은 모드 조건부 — PR 통합이면 PR, git-flow이면 통합 창 규율에 따라 develop 통합 워크트리 병합). Phase D 소멸 근거: main이 애초에 커밋을 받지 않음.

## AC-11 (P2) — 이미 수리됨 (검증 전용)

관측: `.claude/agents/moai/manager-git.md` 158행 / 템플릿 156행에 감사가 처방한 수리 문구가 이미 존재 — "Pre-flight status reads are read-only, but `git rev-list --count --left-right` reads the remote-tracking refs that `git fetch` updates, so the two are not independent: run `git fetch` first and wait until it completes, then run `git rev-list --count --left-right` — never in the same batch as the fetch. Reads that do not consume the fetch result (`git status`, `gh pr checks --json`) may run in parallel with the fetch..."

→ 본 SPEC은 편집 없이 폐쇄를 기록한다. 카드 지시 "이미 사라졌다면 그 사실이 산출물"의 이행.

## SX-R04 (P1) — 부분 수리, 잔여 = 중복 레시피

관측 (`.claude/skills/moai/workflows/sync/delivery.md` 로컬 355-414행 / 템플릿 330-414행; 사본당 대소문자 무시 `auto-merge` 18적중 — 양 사본 동일):

이미 수리됨:
- 361행(로컬) / 336행(템플릿): "Merging is opt-in. The single criterion is the `--auto-merge` opt-in defined in `manager-git.md` § PR Auto-Merge; worktree context alone never triggers a merge."
- `merge_method`는 설정 해석(`git_strategy.<mode>.merge_method`, 기본 squash) — 하드코딩 squash 없음.

잔여 결함(중복):
- 367-369행(로컬) / 342-344행(템플릿): "Mode conditions (same as `manager-git.md` § PR Auto-Merge)" + 두 조건 원문 재인용
- 385-391행(로컬) / 360-366행(템플릿): "Auto-Merge Execution" 5단계 레시피 재인용 ("Checkout target branch, fetch latest" 포함)
- 393-397행(로컬) / 368-372행(템플릿): "Auto-Merge Failures" 실패 처리 중복

SSOT 귀속: `manager-git.md` § PR Auto-Merge (로컬 166-177행 / 템플릿 164-175행) — 깃발 전용 트리거, 폐기 별칭(`--merge`), 모드 조건, merge_method 해석 5단계.

delivery 고유 보존 대상: Step 3.4 "Only applies when a PR was created in Step 3.2" 프레이밍, Post-Merge Automatic Cleanup (`workflow.worktree.auto_cleanup` 키, 로컬 399-414행 / 템플릿 374-389행).

## 사본 분기 관측 (범위 밖)

- manager-git.md: 로컬 vs 템플릿 — frontmatter description 차이(~5행). 본문 위반 영역은 사실상 동일.
- delivery.md: 로컬 vs 템플릿 — ~94 diff 행(lint/크로스빌드 블록, Route A/B 산문, TRACE PROBE 주석).
- 두 분기 모두 선존재이며 본 SPEC 범위 밖 — 후속 후보로 기록만.

## Gaps 기록 — 배치 grep 오아티팩트

이전 배치 grep 출력이 로컬 delivery.md에 auto-merge 콘텐츠가 없다고 잘못 시사했다. 직접 재측정에서 사본당 18적중을 확인했다. **처리**: 해당 배치의 부재(absence) 주장은 전량 폐기했으며, 위의 모든 발견 건 상태는 본 트리에서의 직접 파일 판독으로 재확인했다. 부재 주장은 관측된 부재가 아니라 도구 출력의 왜곡이었음을 본 섹션에 남긴다.

## 미검증 항목 (Gaps)

- **plan-audit 1차 판정(F2)로 발각된 조사 공백**: 최초 조사는 `.moai/config/sections/git-strategy.yaml`의 `workflow` 키와 `auto_enabled` 키를 판독하지 않았다(당시 "설정 표면 불변"을 이유로 범위 밖 처리). 이 공백 때문에 "브랜치 → PR 승격" 단일 모델이 git-flow 모드 — 실제로 `auto_enabled == false`로 운영되는 유일 모드, 통합이 develop 통합 워크트리 병합 — 와 충돌한다는 사실이 plan-audit 측정 전까지 가려졌다. 수리 라운드에서 직접 재측정: 2행 `mode: manual`, 9행 `workflow: git-flow`, 29행 `auto_enabled: false`. 승격 단계는 모드 조건부로 재정의됐다(spec.md §1·REQ-GP-001, plan.md M1, acceptance.md AC-GP-01a).
- 6개 대상 파일 외 late-branch `reset --hard` 모델을 인용하는 표면의 전수 조사는 하지 않았다. run-phase M4의 2차 grep 스윕이 이를 보완한다.
