# t622 재현 — manager-git·sync delivery git 절차 3건 (develop 기준)

- 카드: t622 (지침 감사 G1 · AC-01 P1 · AC-11 · SX-R04)
- 측정 트리: 로컬 develop `c352330d3` (카드 워크트리 `WT-git-procedure-fixes` = `281d1b664`, `HEAD^2` = `c352330d3`)
- 보고서 기준 트리: main `2213871af` — 인용 줄번호가 다르므로 문구로 찾았다. 아래 줄번호는 모두 `c352330d3` 기준이다.
- 측정 일자: 2026-09-10 · 코드와 지침 파일은 아직 한 줄도 바꾸지 않았다.

## 요약

| 결함 | 재현 | 바꿀 위치(카드 범위) | 카드 밖 형제 | 사본 | 설계 판단 |
|---|---|---|---|---|---|
| AC-01 | 재현됨 | `manager-git.md` 8줄 | 3파일 6줄 | 파일마다 로컬=템플릿 동일 2사본, `manager-git` 은 codex 생성물 추가 | **있음** |
| AC-11 | 재현됨 | `manager-git.md:156` 1곳 | 1파일 1곳 | 2사본 + codex 생성물 | 없음(기계적) |
| SX-R04 | 재현됨 | `delivery.md` 3곳 + `manager-git.md` 3곳 | — | `delivery.md` 두 사본은 갈라져 있음(해당 줄은 동일) | **있음** |

## AC-01 — primary checkout 에서 브랜치를 바꾸는 late-branch 절차

### 전제 확인 — 배포물 내부의 모순인가

| 근거 | 내용 |
|---|---|
| 템플릿 `AGENTS.md:64` | "Never change branch state in the primary checkout." — 조건 없는 금지 |
| 템플릿 `main-checkout-branch-guard.md:16` | `[HARD]` MUST NOT, 조건 없음 |
| 같은 파일 `:86` | "1인 저장소에는 해당 없음"은 **기계적 차단 훅의 기본값**(`false`)에만 붙은 조건 |

→ 배포되는 지침이 스스로 금지한 명령을 같은 배포물 안에서 안내한다. 이 저장소 규칙과만 충돌하는 것이 아니다.

### 위치 — 명령 줄 전수 분류

`git grep -n -E 'git checkout |git switch |git reset --hard'` (템플릿 `.claude` + `AGENTS.md` + `CLAUDE.md`) → 31줄. 실행 안내와 금지 목록을 갈랐다.

| 분류 | 파일 | 줄 |
|---|---|---|
| **실행 안내 — 카드 범위** | `agents/moai/manager-git.md` | 42 Rollback · 96 Phase A · 110 Phase C · 119·121 Phase D · 125 복구 · 127 실패 복구 · 137 Personal 모드 설명 |
| **실행 안내 — 카드 밖** | `rules/moai/workflow/spec-workflow.md` | 50 Step 1 main 진입 전제 · 56·58 Step 4 late-branch 종결 |
| | `skills/moai/workflows/plan/spec-assembly.md` | 336 Phase C 수동 `git switch -c` |
| | `skills/moai/workflows/sync/delivery.md` | 323·324 Step 3.3.5 기준 브랜치 복귀 |
| 위반 아님 | `AGENTS.md` 64–66 · 가드 규칙 20–22 · 가드 상세 120–124 | 금지 목록 자체 |
| | `coding-standards.md:141` · `moai-ref-git-workflow/SKILL.md:144–145` | 파괴 명령 경고표 |
| | `verification-claim-integrity-detail.md:161` | 인용 예시 |
| | `settings.json.tmpl:564` | 권한 항목 |
| | `moai-workflow-worktree/modules/troubleshooting.md:176·179` | 워크트리 안의 안내 — 허용 범위 |

`spec-workflow.md` 는 절차의 **계약 원문**이다(`manager-git.md` 가 이 절을 교차참조한다). `manager-git.md` 만 고치면 규칙과 에이전트가 반대를 말하게 된다.

### 사본

| 파일 | 로컬 ↔ 템플릿 |
|---|---|
| `manager-git.md` | 동일 (`git diff --no-index` exit 0) — 템플릿 수정 뒤 `make agents-emit` 으로 `.codex/agents/moai/manager-git.toml` 재생성 |
| `spec-workflow.md` · `spec-assembly.md` | 동일 (exit 0) |
| `delivery.md` | 갈라짐 (exit 1) — 아래 SX-R04 참조 |

### 설계 판단 — 있음

late-branch 는 문구가 아니라 워크플로다. `git-strategy.yaml` 의 `main_late_branch` 옵션, `team.branch_creation.auto_enabled == false` 기본값, `spec-workflow.md` 의 Step 1·Step 4 계약이 모두 이 절차에 묶여 있다. 선택지는 셋이다.

1. 절차를 launcher 워크트리 흐름으로 다시 설계한다(감사 보고서 권고).
2. 적용 범위를 조건절로 좁힌다 — 이 경우 템플릿 `AGENTS.md` 의 무조건 금지와 다시 부딪히므로 금지문 쪽 조건도 함께 정해야 한다.
3. `main_late_branch` 옵션을 은퇴시킨다.

어느 쪽이든 예시 명령을 지우기만 하면 기능을 말없이 철회하는 결과가 된다.

## AC-11 — `git fetch` 와 그 결과를 읽는 검사를 병렬로 지시

### 위치

| 파일 | 줄 | 문구 |
|---|---|---|
| `manager-git.md` (카드 범위) | 156 | "`git fetch`, `git status`, `git rev-list --count --left-right`, `gh pr checks --json` are independent and read-only: issue them as ONE single-turn multi-Bash batch" |
| `rules/moai/core/agent-common-protocol.md` (카드 밖) | 292 · 296 · 299 | Pre-Spawn Sync Check 가 "parallel batch" 를 선언하고 `git fetch origin main`(296)과 `git rev-list`(299)를 **별도 명령**으로 나열 |
| 같은 파일 — 정상 | 347 | Pre-Edit Sync Check 는 `git fetch …; git rev-list …` 를 한 명령에 묶어 순서가 보장된다 |

`git fetch` 는 원격 추적 ref 를 갱신하고 `rev-list` 는 그 ref 를 읽으므로 둘은 독립이 아니다.

줄 단위 검색(`fetch` 와 병렬·배치·독립이 한 줄에 함께 있는 경우)은 5줄을 잡았고 진짜는 `manager-git.md:156` 하나다(나머지 4줄은 낱말이 우연히 겹침). Pre-Spawn 형제는 선언과 명령이 다른 줄이라 줄 단위 검색으로는 잡히지 않아 직접 읽어 확인했다.

### 뿌리 확인

| 확인 | 결과 |
|---|---|
| `verification-batch-pattern.md`(156행이 인용하는 분류표)에 `fetch` 분류가 있는가 | 없음 — 같은 검색이 다른 낱말로 13줄을 잡았으므로 부재가 진짜다 |
| 표준 7개 명령 묶음(`agent-common-protocol-reference.md`)에 `git fetch`·`rev-list` 가 있는가 | 없음 — 같은 검색이 `gh pr checks` 로 7줄을 잡았다 |

→ 분류표를 잘못 옮긴 것이 아니라 두 문장이 스스로 틀린 주장을 한다. 다만 분류표의 병렬 금지 목록에 "한 항목이 바꾼 상태를 다른 항목이 읽는 경우"가 빠져 있다(공백).

### 사본 · 설계 판단

2사본(로컬=템플릿 동일) + `manager-git` codex 생성물. 설계 판단 없음 — `fetch` 를 먼저 끝낸 뒤 진정 독립인 읽기만 병렬로 두면 된다. Pre-Spawn 형제는 항상 로드되는 규칙이라 수정 시점을 작업 끝으로 미뤄야 한다(프롬프트 캐시).

## SX-R04 — auto-merge 조건 충돌과 squash 하드코딩

### 위치

| 파일 | 줄 | 문구 |
|---|---|---|
| `delivery.md` | 335–338 | Auto-Merge Trigger: `is_worktree_context == true` AND `--no-merge` 미지정 → **기본으로 병합** |
| `delivery.md` | 343 · 355 | `gh pr merge --squash --delete-branch` **하드코딩** |
| `manager-git.md` | 148 | Team Mode: "Auto-merge: only with the `--auto-merge` flag" |
| `manager-git.md` | 166 · 170 | "Execute only with `--auto-merge` flag AND all approvals" · `--<merge_method>` 해석 |
| `manager-git.md` | 32 | `merge_method` 해석 규칙(SSOT 후보) |
| `manager-git.md` | 114 | Phase C 예시도 `--squash` 로 적혀 있음 |

team 모드이면서 워크트리인 경우 두 절이 반대를 지시한다(무플래그 기본 병합 ↔ 플래그와 승인 필수). 하드코딩 검색(`gh pr merge … --squash`)에서 나머지 적중 `moai-ref-git-workflow/SKILL.md:133` 은 참조표이고 `manager-git.md:32` 는 기본값 설명이라 수정 대상이 아니다.

### 사본

`delivery.md` 두 사본은 갈라져 있다(3줄 추가·4줄 삭제). 차이는 WT 경로의 `develop` 표기(템플릿은 중립 표기)와 꼬리말뿐이고 **SX-R04 해당 줄(335–355)은 두 사본이 같다.** 의도된 차이이므로 통째 복사가 아니라 줄 단위로 고쳐야 한다.

### 설계 판단 — 있음

어느 기본값이 이기는지(워크트리 기본 병합 ↔ team 옵트인) 정해야 한다. `merge_method` 치환 자체는 기계적이다.

## 판정 권고

- AC-01·SX-R04 에 설계 결정이 들어 있고, 카드 밖 형제가 규칙 파일 2개(`spec-workflow.md`·`agent-common-protocol.md`)를 포함한다 → **SPEC 이 필요해 보인다(Class C 유지).**
- AC-11 은 기계적이므로 분리해 먼저 닫는 선택도 가능하다.
- 판정은 리드 몫이다.

## 미검증

- 이 문구들을 실제로 실행해 공유 checkout 이 바뀌는 것은 재현하지 않았다(파괴적 명령이라 실행하지 않음). 결함 판정은 템플릿 금지문과의 대조로 했다.
- `fetch` 와 `rev-list` 의 경합을 실제로 일으키지 않았다(감사 보고서도 신뢰도 medium).
- 로컬 `.claude` 사본 외의 설치본(사용자 프로젝트)은 보지 않았다.
- `git-strategy.yaml` 의 `main_late_branch` 를 실제로 쓰는 사용자가 있는지는 확인할 수 없다.

## 증거 파일

`copy-diff-*.txt` · `premise-template-branch-doctrine.txt` · `sweep-branch-change-commands.txt` · `sweep-branch-change-lines.txt` · `sweep-fetch-parallel.txt` · `sweep-squash-hardcode.txt` · `sweep-sync-check-batch.txt` · `spec-workflow-late-branch.txt` · `root-batch-pattern-taxonomy.txt` · `root-canonical-batch.txt` (모두 이 디렉터리, 커밋 전)
