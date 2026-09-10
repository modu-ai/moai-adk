# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 되돌리기 어려운 결정과 바뀔 가능성이 큰 항목(설계 항목 B1~B3, Frozen 구역 수정)을 먼저 적고, 기계적 편집은 뒤에 둔다. 마일스톤 번호는 리드가 정한 대로 AC-11을 M1으로 둔다.

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M, era V3R6.
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정): `BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` — 2026-09-10 이 워크트리에서 `git rev-parse b412f8a33` 로 해석.
- 결정(2026-09-10, 운영자·리드 경유): OD-1 = 선택지 1(런처 워크트리 흐름), OD-2 = 선택지 B(`manager-git.md` 옵트인). spec.md §C.1. 결정은 Implementation Kickoff Approval을 대신하지 않는다.
- 근거 기록: `.moai/reports/t622/repro.md`, plan-audit 1회차 `.moai/reports/t622/plan-audit.md`, 2회차 `.moai/reports/t622/plan-audit-iter2.md`.
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로. 인용 전에 여기로 반출한다).
- 편집 대상은 파일마다 로컬(`.claude/…`)과 템플릿(`internal/template/templates/.claude/…`) 두 사본이다.

## §B 알려진 문제

1. **`delivery.md` 두 사본은 의도적으로 다르다.** 차이는 275·278행(통합 워크트리 경로 표기)과 479-480행(꼬리말)뿐이다. 파일 통째 복사는 이 차이를 지운다.
2. **`doc-execution.md` 두 사본도 의도적으로 다르다.** 로컬 사본에만 138-143행(6줄)이 있다. 34-36행은 그 앞이라 두 사본에서 같은 줄번호다. 파일 통째 복사 금지.
3. **바이트 동일 테스트 대상이 아닌 파일**: `manager-git.md`, `agent-common-protocol.md`(`rule_template_mirror_test.go` 주석), `doc-execution.md`(파일에 이름 없음), `zone-registry.md`(파일에 이름 없음, 현재 두 사본 바이트 동일). 이들의 사본 일치는 AC-GDP-013·010의 `diff` 로만 보장된다.
4. **`spec-workflow.md` 와 `spec-assembly.md` 는 바이트 동일 테스트 대상이다** (`workflowOptMirroredPaths`, `lateBranchMirroredPaths`). 한쪽만 고치면 `RULE_TEMPLATE_MIRROR_DRIFT` 로 실패한다.
5. **`spec-workflow.md` Step 1·Step 4 문장은 헌법 레지스트리의 Frozen 항목이다** (`CONST-V3R5-027`, `CONST-V3R5-028`). 문장을 바꾸면 `moai constitution validate` 가 `DRIFT` 를 보고하므로 레지스트리 갱신이 따라온다(spec.md §C.4, §C.5 B3).
6. **`agent-common-protocol.md` 는 항상 로드되는 규칙이다.** 세션 도중 고치면 로드된 프롬프트 접두부가 무효화된다(cache-aware-execution 지침 3).
7. **`manager-git.md:32` 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 지우면 안 된다.
8. **생성물 `.toml` 은 손으로 고치지 않는다.** 템플릿 `manager-git.md` 를 고친 뒤 `make agents-emit` 으로만 재생성한다.
9. **워크트리 세션 가드는 복합 명령 안의 `git`·`parallel` 낱말, 셸 변수를 받는 `sed`·`perl`, 여러 명령을 이은 git 스크립트를 거부한다.** 검출식은 `[g]it`·`para[l]lel` 로, perl 코드 안에서는 `\x67it` 로 쓰고 경로는 글자 그대로 쓴다.
10. **셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛴다.** 판정은 `/usr/bin/grep` 으로 하고, 개수가 찍히지 않은 결과는 판정 불가로 읽는다. perl로 픽스처를 만들 때 비-ASCII 문자는 `-CSD` 를 함께 준다.

## §C 사전 점검 (run-phase 진입 시)

아래는 모두 읽기 전용이다. 결과는 `.moai/reports/t622/run/` 에 파일로 남기고 exit code를 따로 기록한다.

1. 트리 확인: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`. develop을 흡수했다면 `git diff --stat $BASE HEAD -- <범위 파일>` 이 비어 있는지 확인한다(비어 있지 않으면 줄번호 인용을 다시 잰다).
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`. 대상:
   - 로컬 범위 파일: `base-manager-git.md`, `base-spec-workflow.md`, `base-spec-assembly.md`, `base-agent-common-protocol.md`, `base-delivery.md`, `base-doc-execution.md`
   - 템플릿 사본: `base-delivery-template.md`, `base-doc-execution-template.md`, 생성물 `base-manager-git.toml`
   - 레지스트리: `base-zone-registry.md`
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행해 기대한 적중이 나오는지 먼저 본다. 적중이 안 나오면 검출기가 틀린 것이므로 편집 전에 멈추고 blocker로 보고한다.
4. `make agents-emit-check` 기준선: exit 0 확인.
5. 사본 diff 기준선: `manager-git.md`, `spec-workflow.md`, `spec-assembly.md`, `agent-common-protocol.md`, `zone-registry.md` exit 0, `delivery.md`·`doc-execution.md` exit 1.
6. 헌법 검증 기준선: `go run ./cmd/moai constitution validate --format json` → `status` 와 `drift_count` 기록.

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지.
- **생성물.** 템플릿 `manager-git.md` 편집 뒤 `make agents-emit` 실행, 재생성된 `.toml` 을 같은 카드에 커밋, `make agents-emit-check` exit 0 을 증거로 기록. `.toml` 손편집 금지. `make build` 와 embed 점검은 레인이 돌리지 않는다.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지.
- **항상 로드 규칙은 마지막에.** `agent-common-protocol.md` 편집은 번호가 M1이어도 run-phase의 **마지막** 지침 편집 커밋으로 실행한다.
- **설계 항목 선결.** spec.md §C.5 B1~B3이 정해지기 전에는 M4의 절차 블록·Step 1/4·레지스트리 편집을 시작하지 않는다.
- **Frozen 구역.** `spec-workflow.md` 의 Step ordering rules와 65-66행 안티패턴 편집은 Frozen 구역 수정이다. Implementation Kickoff Approval에서 명시하고, 레지스트리 경로(B3)를 따른 뒤 헌법 검증 `DRIFT` 0을 확인한다.
- **건드리지 않는 것.** `internal/hook/branch_guard.go`, 설정 키 `workflow.worktree.auto_merge`, `AGENTS.md` §2 금지문.
- **범위.** spec.md §D의 "카드 범위 밖에서 관측한 형제"는 리드 판단 없이 고치지 않는다.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/...`)만 돌린다.
- **git 인덱스.** 커밋·스테이징·푸시는 레인 규칙에 따르되 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

run-phase 완료 보고는 다음을 담는다.

- E1: AC-GDP-001~016 PASS/FAIL/N/A 표(011·012는 N/A). 각 행에 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과(기준 트리에서 기대 적중이 나왔다는 기록).
- E3: `make agents-emit-check` RED(재생성 전)와 GREEN(재생성 뒤) 출력.
- E4: `go test ./internal/template/ -run '…' -v -count=1` 출력과 최상위 PASS 줄 수(정확히 3).
- E5: 사본 diff 결과(다섯 파일 exit 0, `delivery.md`·`doc-execution.md` 차이 본문 동일).
- E6: 커밋 SHA 목록과 `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋이라는 확인.
- E7: AC-GDP-007 분류 장부 `.moai/reports/t622/ac01-classification.md`.
- E8: AC-GDP-001·006·009·010·015 읽기 단계 기록.
- E9: B1~B3 결정 기록, Frozen 구역 수정 기록, Implementation Kickoff Approval 기록(progress §E.2), 헌법 검증 JSON.

## §F 마일스톤

### F.0 결정과 선결 항목

| 항목 | 상태 | 걸린 마일스톤 |
|---|---|---|
| OD-1 = 선택지 1 | 결정됨 (2026-09-10) | M4 |
| OD-2 = 선택지 B | 결정됨 (2026-09-10) | M3 |
| B1 — PR 수와 PR 브랜치 접두 | **run-phase 착수 전 결정 필요** | M4 (절차 블록, Step 1/4, 66행 안티패턴, `CONST-V3R5-028`) |
| B2 — primary checkout의 PR 브랜치 경로 존속 | **run-phase 착수 전 확인 필요** | M4 (`delivery.md` Step 3.3.5 문구) |
| B3 — 레지스트리 경로와 파일 범위 | **run-phase 착수 전 결정 필요** | M4 (`zone-registry.md` L·T) |

M1·M2·M3는 B1~B3과 무관하게 진행할 수 있다.

### M1 — AC-11: fetch 순서 보장

- 대상 요구사항: REQ-GDP-001, 002, 003 (+ 013, 014, 015의 해당 파일분)
- 편집:
  - `manager-git.md` L·T 156 — `git fetch` 를 먼저 끝내고, 그 결과를 읽는 `git rev-list` 는 그 뒤에 실행한다는 순서로 고친다. fetch 와 rev-list 를 같은 배치·목록에 넣지 않는다.
  - `agent-common-protocol.md` L·T Pre-Spawn Sync Check 292·296·299 — 코드 블록 안에서 Pre-Edit 347행과 같은 형태(`git fetch origin main 2>&1; git rev-list …` 한 명령)로 순서를 보장한다. 세 번째 명령과 두 해석 표는 그대로 둔다.
- **실행 순서**: `manager-git.md` 편집분은 M2·M4의 같은 파일 편집과 함께 진행하고 `make agents-emit` 을 한 번에 돌린다. **`agent-common-protocol.md` 편집분은 run-phase의 마지막 지침 편집 커밋으로 실행한다.**

### M2 — SX-R04: 병합 방식 해석

- 대상 요구사항: REQ-GDP-004, 005
- 편집:
  - `delivery.md` L·T 343·355 — `gh pr merge --<merge_method> --delete-branch` 로 바꾸고 해석 출처를 한 번 명시한다. 275·278·479-480행의 사본 차이는 건드리지 않는다.
  - `manager-git.md` L·T 114 — M4의 절차 블록 재작성에 흡수된다(§C.2 C3의 `--<merge_method>`).

### M3 — SX-R04: auto-merge 기본값 단일 기준 (OD-2 = B)

- 대상 요구사항: REQ-GDP-006 (+ 013, 015)
- 편집:
  - `delivery.md` L·T 335-338(트리거) — 워크트리 문맥 기본 병합 조건을 없애고, 병합은 `manager-git.md` 옵트인(`--auto-merge` 와 전원 승인)을 따른다고 이름으로 밝힌다.
  - `delivery.md` L·T 348-349(플래그 설명) — `--no-merge`·`--merge` 설명과 폐기 경고 문구를 새 기본값에 맞춘다.
  - `doc-execution.md` L·T 34-36 — "worktree contexts default to auto-merge" 문장을 없애고, 병합 여부는 `manager-git.md` 옵트인이 정한다고 이름으로 밝힌다. `is_worktree_context` 감지 정의(32-35행)는 필요하면 유지한다. 138-143행의 사본 차이는 건드리지 않는다.
  - `manager-git.md` 148·166행 옵트인 문장은 그대로 둔다.
- `doc-execution.md` 는 미러 테스트 허용 목록에 없다 — 사본 일치는 AC-GDP-013의 본문 diff로 확인한다.

### M4 — AC-01: 런처 워크트리 흐름으로 재설계 (OD-1 = 1)

- 대상 요구사항: REQ-GDP-007, 008, 009, 010 (+ 013, 014, 015)
- 선결: B1~B3 (F.0).
- 편집 목록(spec.md §C.6 OD-1 선택지 1 행의 영향 파일 열과 §C.3):
  - `manager-git.md` L·T 88-129 절차 블록을 §C.2 진입(E1-E3)·PR(C1-C3)·종결(X1-X4)로 다시 쓴다. 137행 옵션 설명, 125·127행 복구 문장도 흐름에 맞춘다. 절 제목 `### Late-Branch Invocation Pattern` 은 유지한다(AC-GDP-008·009 추출 표지).
  - `manager-git.md` L·T 42(Checkpoint Rollback)·160(Synchronization pull) — §C.3 처리.
  - `spec-workflow.md` L·T 49-62 Step ordering rules — Step 1의 "main checkout, 워크트리 없음" 전제와 Step 4의 late-branch 종결을 §C.2에 맞춘다. 65-66행 안티패턴을 흐름과 맞춘다. **Frozen 구역 수정** — Kickoff에서 명시, B3 경로로 레지스트리 갱신.
  - `spec-assembly.md` L·T 332-340 Late-branch Pre-check — 진입 단계(E1)와 §C.2 참조로 바꾼다. 제목 유지.
  - `delivery.md` L·T 319-328 Step 3.3.5 — §C.3 처리(323-324·328).
  - `zone-registry.md` L·T — B3 결정에 따라 `CONST-V3R5-027`(해당하면 028) 갱신.
- `spec-assembly.md`·`spec-workflow.md` 는 `manager-git.md` 절차 절을 형식을 갖춘 참조로 계속 가리킨다(AC-GDP-008 하한).
- Frozen 수정 기록과 Kickoff 기록을 progress §E.2 에 남긴다(AC-GDP-010).

### M5 — 사본 일치·생성물·중립성·순서 검증 (마지막)

- 대상 요구사항: REQ-GDP-013, 014, 015 (+ AC-GDP-016 절차 점검)
- 절차:
  1. 템플릿 `manager-git.md` 의 모든 편집이 끝난 시점에 `make agents-emit-check` 를 먼저 돌려 exit 1 을 관측한다.
  2. `make agents-emit` → `make agents-emit-check` exit 0.
  3. 사본 diff(다섯 파일 exit 0, 두 파일 본문 diff), 미러 테스트 3개와 최상위 PASS 줄 3개.
  4. 템플릿 diff 추가 줄에서 SPEC ID·REQ 토큰·날짜·SHA 낱말·`CLAUDE.local` 검사.
  5. 헌법 검증(`go run ./cmd/moai constitution validate --format json`) `drift_count` 0.
  6. `agent-common-protocol.md` 커밋 뒤에 다른 범위 지침 파일을 고친 커밋이 없는지 확인(AC-GDP-016).

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것 (`delivery.md`·`doc-execution.md` 의도된 차이 소실).
- 브랜치 변경 명령 전체 부재를 합격 조건으로 쓰는 것 (금지 목록·경고표와 워크트리 내부 안내까지 지우게 된다).
- 금지 집합의 일부만 검출식에 넣는 것 (긴 옵션·`-C` 형태로 다시 쓴 안내가 장부를 빠져나간다).
- primary checkout의 로컬 `main` 을 갱신하려고 재설계 흐름에 `git pull`·`git reset` 을 다시 넣는 것.
- Frozen 절을 바꾸고 레지스트리를 두어 헌법 검증 `DRIFT` 를 남기는 것.
- 설정 키 `workflow.worktree.auto_merge` 를 기준으로 되살리는 것 (OD-2에서 교체됨).
- 예시 명령만 지우고 절차의 존재를 알리는 문장을 남기는 것 (REQ-GDP-008).
- `.toml` 을 손으로 맞추는 것.
- `agent-common-protocol.md` 를 run-phase 초반에 고치는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것.
- 개수가 찍히지 않은 grep 결과를 0으로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거), §C(결정 기록·흐름 정의·설계 항목), §D(제외 범위), §E(미검증·잔여 위험)
- `acceptance.md` (AC-GDP-001~016; 011·012 철회)
- `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`
- `internal/template/rule_template_mirror_test.go` — 바이트 동일 대상 목록
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R5-027`, `CONST-V3R5-028`, § Retiring an Entry
- `internal/constitution/validator.go`, `internal/cli/constitution.go` — DRIFT 판정과 `amend` 게이트
- `.claude/rules/moai/workflow/worktree-integration.md` — L1 진입·폐기, 워크트리 기준 브랜치
- `Makefile` 38·47행 — `agents-emit`, `agents-emit-check`
- SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001
