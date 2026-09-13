# SPEC-GIT-PROC-SAFE-001 — 수락 기준

문서 전용 변경 SPEC이다. TDD RED/GREEN은 문서 도메인으로 변환한다: **RED-now 셀** = baseline 트리에서 관측된 위반 텍스트(file:line + 인용), **GREEN 셀** = 수리 후 대체 텍스트 검사(관측 가능한 텍스트 조건). 모든 baseline 귀속은 develop `c9ceff175`(본 워크트리, 2026-09-14 측정)다.

## §D AC Matrix

| AC ID | 발견 건 | 종류 | 대상 | 판정 방법 |
|-------|---------|------|------|-----------|
| AC-GP-01a | AC-01 | 편집 | manager-git.md 양 사본 | RED/GREEN 텍스트 검사 + REQ-GP-002 grep |
| AC-GP-01b | AC-01(확장) | 편집 | spec-workflow.md 양 사본 | RED/GREEN 텍스트 검사 |
| AC-GP-01c | AC-01 | 편집 | manager-git.md 139행 옵션 행(양 사본) | RED/GREEN 텍스트 검사 |
| AC-SX-01 | SX-R04 | 편집 | delivery.md Step 3.4 (양 사본) | 중복 제거 + 보존 절 존재 |
| AC-AC11-01 | AC-11 | 검증 전용 | manager-git.md 158/156행 | 관측 기록, 편집 0건 |
| AC-XREF-01 | AC-01 | 편집 수반 | 양 파일 교차참조 | 참조 해석 가능성 검사 |
| AC-MIRROR-01 | 범위 방어 | 회귀 가드 | 6개 파일 전체 diff | 범위 밖 행 불변 |
| AC-TN-01 | REQ-TN-001 | 편집 수반 검사 | 템플릿 사본 3개 | grep 적중 판정 |

## AC-GP-01a — manager-git.md Late-Branch 절차 (양 사본)

**RED-now** (baseline `c9ceff175`, 직접 판독):
- 로컬 `.claude/agents/moai/manager-git.md` 98행: `git checkout main && git pull origin main` (Phase A)
- 로컬 112행: `git switch -c feat/SPEC-XXX` (Phase C)
- 로컬 121-124행: `git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main` (Phase D)
- 로컬 129행: `git fetch origin && git reset --hard origin/main` (실패 복구)
- 템플릿 `internal/template/templates/.claude/agents/moai/manager-git.md` 96, 110, 119-122, 127행 동일 내용
- RED 사유: 이 수순들은 AGENTS.md §2가 primary 체크아웃에서 금지하는 브랜치 상태 변이이며, 절차 텍스트가 이를 지시하므로 교리 자체가 규율을 위반한다.

**GREEN** (M1 완료 후):
- Given: 편집된 manager-git.md 양 사본을 판독할 때
- When: § Late-Branch Invocation Pattern 섹션을 읽으면
- Then: (1) 모든 브랜치 변이가 런처 경유 워크트리 내부로 한정됨을 서술하고, (2) primary 체크아웃 문맥의 `git checkout main`·`git switch -c`·`git reset --hard` 인용이 0건이며, (3) 커밋 적립-브랜치 → 승격 모델이 명시되고 승격 단계는 구성된 workflow에 조건부(PR 통합이면 PR 승격, git-flow이면 통합 창 규율에 따라 develop 통합 워크트리 병합)로 서술되고, (4) Phase D main 정렬 수순이 존재하지 않는다.
- 검사: `grep -n 'reset --hard\|checkout main\|switch -c'` 적중 0건(워크트리 내부 명령 문맥 근거 포함 판독) — run-phase §E 기록.

## AC-GP-01b — spec-workflow.md SSOT 궤적 (양 사본)

**RED-now** (baseline `c9ceff175`, 직접 판독):
- `.claude/rules/moai/workflow/spec-workflow.md` 50행: Route B 사전조건 "Step 1 entry requires `git rev-parse --abbrev-ref HEAD == main`... plan-phase commits land directly on `main` and are pushed only after Phase C `git switch -c plan/SPEC-XXX`"
- 55-60행 closure 명령 블록: `git checkout main` / `git fetch origin` / `git reset --hard origin/main` / `git pull origin main`
- 62행: post-condition `git status --porcelain` empty + `git rev-parse main == origin/main`, 교차참조 "For the complete 4-phase Late-branch invocation pattern (A→D), see `.claude/agents/moai/manager-git.md`"
- 템플릿 `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` 50행, 55-60행, 62행 — 로컬 사본과 동일 내용·동일 행(2026-09-14 수리 라운드에서 직접 판독 재확인)

**GREEN** (M2 완료 후):
- Given: 편집된 spec-workflow.md 양 사본을 판독할 때
- When: Route B 사전조건과 Step 4 closure 단락을 읽으면
- Then: (1) late-branch 흐름이 워크트리 브랜치 모델로 서술되고, (2) `reset --hard` 포함 closure 블록이 존재하지 않으며, (3) 새 post-condition(워크트리 브랜치가 PR로 승격되며 main은 커밋을 받지 않음)이 검증 가능한 형태로 서술되고, (4) manager-git.md 교차참조가 갱신된 섹션명으로 해석 가능하다.

## AC-GP-01c — Personal 모드 옵션 행 (양 사본)

**RED-now**: 로컬 139행 / 템플릿 137행 — `main_late_branch` 설명이 "local main `reset --hard origin/main` cleanup (4-phase procedure)"을 포함.
**GREEN** (M1 완료 후): 옵션 설명이 워크트리 브랜치 모델을 가리키며 `reset --hard` 문구가 없다.

## AC-SX-01 — delivery.md Step 3.4 축소 (양 사본)

**RED-now** (baseline `c9ceff175`, 직접 판독; 로컬 355-414행 / 템플릿 330-414행, 사본당 대소문자 무시 `auto-merge` 18적중 재측정):
- 367-369행(로컬) / 342-344행(템플릿): "Mode conditions (same as `manager-git.md` § PR Auto-Merge)" 뒤 두 조건 원문 재인용
- 385-391행(로컬) / 360-366행(템플릿): "Auto-Merge Execution" 5단계 레시피 재인용 (4단계 "Checkout target branch, fetch latest" 포함)
- 393-397행(로컬) / 368-372행(템플릿): "Auto-Merge Failures" 실패 처리 중복

**GREEN** (M3 완료 후):
- Given: 편집된 delivery.md Step 3.4를 판독할 때
- When: Auto-Merge 섹션을 읽으면
- Then: (1) 모드 조건·실행 레시피·실패 처리 중복이 제거되고 `manager-git.md` § PR Auto-Merge 포인터로 위임되며, (2) "PR이 Step 3.2에서 생성된 경우에만 적용" 프레이밍이 유지되고, (3) Post-Merge Automatic Cleanup 절(`workflow.worktree.auto_cleanup` 키)이 delivery에 유지되고, (4) 트리거 단일 기준 문장과 `merge_method` 설정 해석 1행이 유지된다.

## AC-AC11-01 — AC-11 검증 전용 폐쇄 (편집 0건)

**Given**: run-phase가 6개 대상 파일을 편집한 뒤
**When**: `git diff`와 직접 판독으로 AC-11 대상 행을 검사하면
**Then**: (1) 로컬 manager-git.md 158행과 템플릿 156행의 fetch→rev-list 순서 문구가 불변이고, (2) 해당 텍스트가 감사 처방 수리 문구(기준이 되는 `run git fetch first and wait until it completes, then run git rev-list --count --left-right`)와 일치함이 acceptance 증거로 기록되며, (3) AC-11 관련 편집 커밋이 존재하지 않는다. — "이미 사라졌다면 그 사실이 산출물" 이행.

## AC-XREF-01 — 교차참조 무결성

**Given**: M1·M2 완료 후
**When**: manager-git.md ↔ spec-workflow.md의 교차참조를 따라가면
**Then**: 양방향 참조가 실존 섹션명을 가리키며 파싱 가능하다(끊긴 앵커 0건).

## AC-TN-01 — 템플릿 중립성 (REQ-TN-001 수락 기준)

**Given**: M1-M3 완료 후
**When**: 범위 내 템플릿 사본 3개(`internal/template/templates/.claude/agents/moai/manager-git.md`, `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`, `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`)에 grep을 실행하면
**Then**: (1) 내부 SPEC ID 패턴(`SPEC-GIT-PROC-SAFE-001` 및 동류) 적중 0건이고, (2) dev-only 로컬 경로(`.claude/rules/local/`) 인용이 0건이며, (3) 16개 프로그래밍 언어 중립성을 훼손하는 표현(특정 언어의 PRIMARY 격상, enabled/planned 격하)이 이 이슈의 편집으로 새로 도입되지 않는다.

## AC-MIRROR-01 — 범위 밖 무손상

**Given**: M1-M4 완료 후
**When**: 6개 대상 파일의 전체 `git diff`를 판독하면
**Then**: 변경 행이 발견 건 범위(REQ-GP-001/003/004, REQ-SX-001)에 귀속되고, 범위 밖 분기(frontmatter description 차이, lint/크로스빌드 블록, TRACE PROBE 주석, Route A/B 산문) 행은 불변이다.

## 품질 게이트 (Definition of Done)

- 모든 AC GREEN 셀 충족 + AC-AC11-01 편집 0건 증명
- §E 자체검증 전 항목 실행 기록
- plan-audit PASS (harness standard)
- 병합 후 CI 녹색 (문서 전용이므로 lint/빌드 게이트 중심)
