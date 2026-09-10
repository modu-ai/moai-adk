# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 되돌리기 어려운 결정과 바뀔 가능성이 큰 결정(OD-1·OD-2)을 먼저 적고, 기계적 편집은 뒤에 둔다. 마일스톤 번호는 리드가 정한 대로 AC-11을 M1으로 둔다.

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M, era V3R6.
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정): `BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` — 2026-09-10 이 워크트리에서 `git rev-parse b412f8a33` 로 해석.
- 재현 근거: `.moai/reports/t622/repro.md` 와 같은 디렉터리의 sweep 파일들(추적 파일). plan-audit 1회차: `.moai/reports/t622/plan-audit.md`.
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로. 인용 전에 여기로 반출한다).
- 편집 대상은 파일마다 로컬(`.claude/…`)과 템플릿(`internal/template/templates/.claude/…`) 두 사본이다.

## §B 알려진 문제

1. **`delivery.md` 두 사본은 의도적으로 다르다.** 차이는 275·278행(통합 워크트리 경로 표기, 템플릿은 중립 표기)과 479-480행(꼬리말)뿐이다. 파일 통째 복사는 이 차이를 지운다.
2. **`manager-git.md` 와 `agent-common-protocol.md` 는 바이트 동일 테스트 대상이 아니다.** `internal/template/rule_template_mirror_test.go` 주석이 두 파일을 byte-parity 허용 목록에서 뺀 사유를 적어 두었다(`agent-common-protocol.md` 는 "17 tokens"). 현재 두 파일 모두 사본 diff exit 0 이지만, 일치는 AC-GDP-013의 `diff` 로만 보장된다.
3. **`spec-workflow.md` 와 `spec-assembly.md` 는 바이트 동일 테스트 대상이다** (`workflowOptMirroredPaths`, `lateBranchMirroredPaths`). 한쪽만 고치면 `RULE_TEMPLATE_MIRROR_DRIFT` 로 실패한다.
4. **`agent-common-protocol.md` 는 항상 로드되는 규칙이다.** 세션 도중 고치면 로드된 프롬프트 접두부가 무효화된다(cache-aware-execution 지침 3).
5. **`manager-git.md:32` 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 병합 방식 고정을 찾는 검색에 걸리지만 지우면 안 된다.
6. **생성물 `.toml` 은 손으로 고치지 않는다.** 템플릿 `manager-git.md` 를 고친 뒤 `make agents-emit` 으로만 재생성한다.
7. **워크트리 세션 가드는 복합 명령 안의 `git`·`parallel` 낱말, 셸 변수를 받는 `sed`, 여러 명령을 이은 git 스크립트를 거부한다.** 검출식은 `[g]it`·`para[l]lel` 로 쓰고, 경로는 변수 대신 글자 그대로 쓴다(acceptance.md 공통 관례).
8. **셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛴다.** plan 작성 중 픽스처 하나가 Latin-1 바이트로 만들어져 개수가 아예 찍히지 않았다. 판정은 `/usr/bin/grep` 으로 하고, 개수가 찍히지 않은 결과는 0이 아니라 판정 불가로 읽는다.

## §C 사전 점검 (run-phase 진입 시)

아래는 모두 읽기 전용이다. 결과는 `.moai/reports/t622/run/` 에 파일로 남기고 exit code를 따로 기록한다.

1. 트리 확인: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`.
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`. 대상:
   - 범위 파일 로컬 다섯(`manager-git.md`, `spec-workflow.md`, `spec-assembly.md`, `agent-common-protocol.md`, `delivery.md`)
   - 템플릿 `delivery.md`(`base-delivery-template.md`)와 생성물 `manager-git.toml`
   - 결정 대기 기준의 대조용: 루트 `AGENTS.md`(`base-AGENTS.md`), 템플릿 `AGENTS.md`(`base-AGENTS-template.md`), `.claude/skills/moai/SKILL.md`(`base-SKILL.md`)
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행해 기대한 적중이 나오는지 먼저 본다. 적중이 안 나오면 검출기가 틀린 것이므로 편집 전에 멈추고 blocker로 보고한다.
4. `make agents-emit-check` 기준선: exit 0 확인.
5. 사본 diff 기준선: 네 파일(`manager-git.md`, `spec-workflow.md`, `spec-assembly.md`, `agent-common-protocol.md`) exit 0, `delivery.md` exit 1.

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지.
- **생성물.** 템플릿 `manager-git.md` 편집 뒤 `make agents-emit` 실행, 재생성된 `internal/template/templates/.codex/agents/moai/manager-git.toml` 을 같은 카드에 커밋, `make agents-emit-check` exit 0 을 증거로 기록. `.toml` 손편집 금지. `make build` 와 embed 점검은 레인이 돌리지 않는다.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지.
- **항상 로드 규칙은 마지막에.** `agent-common-protocol.md` 편집은 번호가 M1이어도 run-phase의 **마지막** 지침 편집 커밋으로 실행한다. 세션 도중에 로드된 접두부를 무효화하지 않기 위해서다.
- **결정 대기.** OD-1·OD-2가 결정되기 전에는 M3·M4를 시작하지 않는다.
- **범위.** spec.md §D의 "카드 범위 밖에서 관측한 형제"는 리드 판단 없이 고치지 않는다.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/...`)만 돌린다.
- **git 인덱스.** 커밋·스테이징·푸시는 레인 규칙에 따르되 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

run-phase 완료 보고는 다음을 담는다.

- E1: AC-GDP-001~016 PASS/FAIL/DEFERRED 표. 각 행에 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과(기준 트리에서 기대 적중이 나왔다는 기록).
- E3: `make agents-emit-check` RED(재생성 전)와 GREEN(재생성 뒤) 출력.
- E4: `go test ./internal/template/ -run '…' -v -count=1` 출력과 최상위 PASS 줄 수(정확히 3이 아니면 판정 불가).
- E5: 사본 diff 결과(네 파일 exit 0, `delivery.md` 차이 본문 동일).
- E6: 커밋 SHA 목록과 `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋이라는 확인.
- E7: AC-GDP-007 분류 장부 `.moai/reports/t622/ac01-classification.md`.

## §F 마일스톤

### F.0 결정 대기 (가장 먼저 확정되어야 하는 것)

| 결정 | 걸린 마일스톤 | 걸린 요구사항 |
|---|---|---|
| OD-1 (Late-branch 처리) | M4 | REQ-GDP-009, 010, 011, 012 (선택지별로 하나만 활성) |
| OD-2 (auto-merge 기본값 기준) | M3 | REQ-GDP-006 |

M1·M2·M5는 두 결정과 무관하게 진행할 수 있다. 다만 M2의 `manager-git.md:114` 편집은 OD-1 선택지 3에서 블록 자체가 사라질 수 있으므로, OD-1이 먼저 결정되면 그 결정에 맞춰 M4와 합쳐 진행한다(REQ-GDP-005의 부재 조건은 어느 쪽이든 성립한다).

### M1 — AC-11: fetch 순서 보장 (결정 불필요)

- 대상 요구사항: REQ-GDP-001, 002, 003 (+ 013, 014, 015의 해당 파일분)
- 편집:
  - `manager-git.md` L·T 156 — `git fetch` 를 먼저 끝내고, 그 결과를 읽는 `git rev-list` 는 그 뒤에 실행한다는 순서로 문장을 고친다. `git fetch` 를 "independent"·병렬·배치 목록에서 뺀다.
  - `agent-common-protocol.md` L·T Pre-Spawn Sync Check 292·296·299 — 코드 블록 안에서 Pre-Edit Sync Check 347행과 같은 형태(`git fetch origin main 2>&1; git rev-list …` 한 명령)로 순서를 보장한다. 세 번째 명령과 두 해석 표는 그대로 둔다. Pre-Edit 절은 건드리지 않는다.
- **실행 순서**: `manager-git.md` 편집분은 M2·M4의 같은 파일 편집과 함께 진행하고 `make agents-emit` 을 한 번에 돌린다. **`agent-common-protocol.md` 편집분은 run-phase의 마지막 지침 편집 커밋으로 실행한다** (항상 로드되는 규칙 — 세션 도중 접두부 무효화 방지).

### M2 — SX-R04: 병합 방식 해석 (결정 불필요)

- 대상 요구사항: REQ-GDP-004, 005 (+ 013, 014, 015의 해당 파일분)
- 편집:
  - `delivery.md` L·T 343·355 — `gh pr merge --squash --delete-branch` 를 `git_strategy.<mode>.merge_method`(기본 `squash`)로 해석한 `gh pr merge --<merge_method> --delete-branch` 로 바꾼다. 해석 출처를 한 번 명시한다. 275·278·479-480행의 사본 차이는 건드리지 않는다.
  - `manager-git.md` L·T 114 — Phase C 예시를 `--<merge_method>` 로 바꾼다. 32행 기본값 설명은 유지. OD-1이 먼저 결정되어 블록이 사라지면 이 편집은 M4에 흡수된다.

### M3 — SX-R04: auto-merge 기본값 단일 기준 **[DEFERRED: OD-2]**

- 대상 요구사항: REQ-GDP-006
- OD-2 결정 전에는 시작하지 않는다. 결정되면 spec.md §C.2 표의 해당 행 "영향 파일" 열이 편집 목록이다.
- 선택지 B·C에서는 `sync/doc-execution.md:36` 이 반대 기본값을 적은 문장이 되므로, 리드가 범위에 넣을지 함께 결정해야 한다(spec.md §D 형제 목록).

### M4 — AC-01: Late-branch 절차 처리 **[DEFERRED: OD-1]**

- 대상 요구사항: REQ-GDP-007, 008 (결정과 무관한 불변식 — 결정 뒤 검증), REQ-GDP-009 + 선택지별 REQ-GDP-010 / 011 / 012 중 하나
- OD-1 결정 전에는 시작하지 않는다. 결정되면 spec.md §C.1 표의 해당 행 "영향 파일" 열이 편집 목록이다.
- 어느 선택지든 late-branch 블록 밖의 네 자리 — `manager-git.md:42`(Checkpoint Rollback), `manager-git.md:160`(Synchronization의 `git pull`), `delivery.md:323-324`(기준 브랜치 복귀), `delivery.md:328`(`git branch -d` 안내) — 는 별도 처리 방식이 필요하다. 선택지 2는 조건으로 덮을 수 있고, 선택지 1·3은 네 자리를 따로 정해야 한다.
- 선택지 1은 `spec-workflow.md` 49행 `[ZONE:Frozen]` 구역을 고치므로 plan-auditor 재검토와 Implementation Kickoff Approval에서 Frozen 구역 수정으로 명시한다.
- 선택지 1·3은 SPEC-WORKTREE-BRANCH-GUARD-001 REQ-WBG-011이 적은 manager-git 면제 근거를 없앤다. 훅 변경은 범위 밖이지만 결정 기록에 그 사실을 남긴다.
- 선택지 2는 루트 `AGENTS.md` 와 템플릿 `AGENTS.md` 의 바이트 수를 편집 전후로 기록한다.

### M5 — 사본 일치·생성물·중립성·순서 검증 (결정 불필요, 마지막)

- 대상 요구사항: REQ-GDP-013, 014, 015 (+ AC-GDP-016 절차 점검)
- 절차:
  1. 템플릿 `manager-git.md` 의 모든 편집이 끝난 시점에 `make agents-emit-check` 를 먼저 돌려 exit 1 을 관측한다(재생성 전 RED — 검출기 양성 대조).
  2. `make agents-emit` → `make agents-emit-check` exit 0.
  3. 사본 diff, `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak' -v -count=1` 와 최상위 PASS 줄 3개 확인.
  4. 템플릿 diff 추가 줄에서 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 검사.
  5. `agent-common-protocol.md` 커밋 뒤에 다른 범위 지침 파일을 고친 커밋이 없는지 확인(AC-GDP-016).

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것 (`delivery.md` 의도된 차이 소실).
- `git checkout`·`git switch`·`git reset --hard` 전체 부재를 합격 조건으로 쓰는 것 (금지 목록·경고표까지 지우게 된다).
- 금지 집합의 일부만 검출식에 넣는 것 (`git pull`·`git merge`·변경형 `git branch` 로 다시 쓴 안내가 장부를 빠져나간다).
- 예시 명령만 지우고 절차의 존재를 알리는 문장을 남기는 것 (REQ-GDP-008).
- `.toml` 을 손으로 맞추는 것.
- 결정 대기 요구사항을 추정한 선택지로 미리 구현하는 것.
- `agent-common-protocol.md` 를 run-phase 초반에 고치는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것.
- 개수가 찍히지 않은 grep 결과를 0으로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거), §C(OD-1·OD-2 표), §D(제외 범위), §E(미검증)
- `acceptance.md` (AC-GDP-001~016)
- `.moai/reports/t622/repro.md` — 재현 기록
- `.moai/reports/t622/plan-audit.md` — plan-audit 1회차
- `internal/template/rule_template_mirror_test.go` — 바이트 동일 대상 목록
- `Makefile` 38·47행 — `agents-emit`, `agents-emit-check`
- SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001
