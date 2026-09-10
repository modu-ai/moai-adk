# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 되돌리기 어려운 결정과 바뀔 가능성이 큰 항목(남은 차단 항목 B4·B6·T1, Frozen 구역 수정, 헌법 개정)을 먼저 적고, 기계적 편집은 뒤에 둔다. 마일스톤 번호는 리드가 정한 대로 AC-11을 M1으로 둔다.

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M(요구사항 수가 상한을 넘음 — spec.md §C.5 T1), era V3R6.
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정): `BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0`. plan 0.1.3 작성 시점 HEAD `880c0c702` 에서 범위 파일·`zone-registry.md` L·T·`internal/constitution`·`internal/cli/constitution.go` 는 `$BASE` 와 차이 없음(`git diff --stat` 출력 없음).
- 결정(2026-09-10, 운영자·리드 경유): OD-1 = 선택지 1, OD-2 = 선택지 B, B1 = SPEC당 PR 하나(`feat/SPEC-*`), B2 = primary checkout feature 브랜치 경로도 워크트리 흐름으로, B3 = `moai constitution amend` 로 `CONST-V3R5-027`·`028` 개정. spec.md §C.1. 결정은 Implementation Kickoff Approval을 대신하지 않는다.
- 근거 기록: `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`.
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로).
- 편집 대상은 파일마다 로컬(`.claude/…`)과 템플릿(`internal/template/templates/.claude/…`) 두 사본이다.

## §B 알려진 문제

1. **`delivery.md` 두 사본은 의도적으로 다르다.** 차이는 275·278행과 479-480행뿐이다. Step 3.1(50행)·Step 3.2 github-flow(238-257행)·Step 3.3.5(319-328행)는 두 사본에서 같은 줄이다. 파일 통째 복사 금지.
2. **`doc-execution.md` 두 사본도 의도적으로 다르다.** 로컬 사본에만 138-143행이 있다.
3. **바이트 동일 테스트 대상이 아닌 파일**: `manager-git.md`, `agent-common-protocol.md`, `doc-execution.md`, `zone-registry.md`. 사본 일치는 AC-GDP-013·023의 `diff` 로만 보장된다.
4. **`spec-workflow.md` 와 `spec-assembly.md` 는 바이트 동일 테스트 대상이다.** 한쪽만 고치면 `RULE_TEMPLATE_MIRROR_DRIFT` 로 실패한다.
5. **`spec-workflow.md` 에는 `[ZONE:Frozen]` 태그가 세 곳(23·49·166행) 있고, 그중 등록된 clause 는 49행 블록 안의 `CONST-V3R5-027`·`028` 뿐이다.** 등록된 두 문장을 바꾸면 검증기가 `DRIFT` 를 보고한다. 등록되지 않은 23행 블록·166행은 검증기가 보지 않는다(spec.md §A.5).
6. **`moai constitution amend` 의 적용 단계는 스텁이다.** 승인 뒤 `updateSourceFile` 에서 실패하고 아무 파일도 바뀌지 않는다(spec.md §A.7). 레지스트리 해석도 CLI 대조와 파이프라인이 다르고, 템플릿 레지스트리는 어느 쪽도 다루지 않는다.
7. **HumanOversight 층은 표준입력 Y/N 질문이다.** 비대화형 셸에서는 입력이 없어 오류로 끝난다. 이 질문에 답하는 것은 운영자다.
8. **`agent-common-protocol.md` 는 항상 로드되는 규칙이다.** 세션 도중 고치면 로드된 접두부가 무효화된다.
9. **`manager-git.md:32` 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 지우면 안 된다.
10. **생성물 `.toml` 은 손으로 고치지 않는다.** `make agents-emit` 으로만 재생성한다.
11. **워크트리 세션 가드**는 복합 명령 안의 `git`·`parallel` 낱말, 셸 변수를 받는 `sed`·`perl`, 여러 명령을 이은 git 스크립트를 거부한다. 검출식은 `[g]it`·`para[l]lel`·`\x67it` 로 쓰고 경로는 글자 그대로 쓴다.
12. **셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛴다.** 판정은 `/usr/bin/grep` 으로 한다.
13. **헌법 검증기는 프로젝트 디렉터리 밖의 레지스트리를 거부한다**("escapes project dir", exit 1, `"status": ""`). 대조용 픽스처는 트리 밖 스크래치 프로젝트 사본 안에서 돌린다.

## §C 사전 점검 (run-phase 진입 시)

모두 읽기 전용이다(1번의 빌드 산출물은 트리 밖 `/tmp`). 결과는 `.moai/reports/t622/run/` 에 파일로 남기고 exit code를 따로 기록한다.

1. 트리 확인: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`. develop을 흡수했다면 `git diff --stat $BASE HEAD -- <범위 파일>` 이 비어 있는지 확인한다.
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`. 대상: `base-manager-git.md`, `base-spec-workflow.md`, `base-spec-assembly.md`, `base-agent-common-protocol.md`, `base-delivery.md`, `base-doc-execution.md`, `base-delivery-template.md`, `base-doc-execution-template.md`, `base-manager-git.toml`, `base-zone-registry.md`.
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행한다. 기대 적중이 안 나오면 편집 전에 멈추고 blocker로 보고한다.
4. `make agents-emit-check` 기준선 exit 0.
5. 사본 diff 기준선: `manager-git.md`, `spec-workflow.md`, `spec-assembly.md`, `agent-common-protocol.md`, `zone-registry.md` exit 0; `delivery.md`·`doc-execution.md` exit 1.
6. 트리 빌드: `go build -o /tmp/t622-moai-tree ./cmd/moai`, `go version -m /tmp/t622-moai-tree` 의 `vcs.revision` 기록.
7. 헌법 검증 기준선(트리 빌드): `unset CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_REGISTRY && /tmp/t622-moai-tree constitution validate --format json` → `status`·`drift_count` 기록. AC-GDP-022의 짝 대조(픽스처 DRIFT 1 / 사본 ok 0)를 트리 빌드로 다시 잰다.
8. `.moai/research/evolution-log.md` 상태와 `.moai/research/.amendment.lock` 부재 확인(4층·잠금).

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지.
- **생성물.** 템플릿 `manager-git.md` 편집 뒤 `make agents-emit`, 재생성된 `.toml` 커밋, `make agents-emit-check` exit 0. `.toml` 손편집 금지.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지. spec.md §C.4의 `--after` 문장도 이 조건을 지킨다.
- **항상 로드 규칙은 마지막에.** `agent-common-protocol.md` 편집은 run-phase의 마지막 지침 편집 커밋이다.
- **선결 항목.** spec.md §C.5 B4·B6·T1이 정해지기 전에는 M5(헌법 개정)와 등록된 두 문장(Step 1·Step 4)의 편집을 시작하지 않는다. B6 결정 전에는 `worktree-integration.md` 를 고치지 않는다.
- **[HARD] 사람 승인 경계.** 레인은 `constitution amend --dry-run` 까지만 실행한다. `--dry-run` 없는 호출은 운영자가 대화형 터미널에서 실행하고 Y/N에 직접 답한다. 레인은 승인 입력을 파이프·here-string·파일 리다이렉트로 넘기지 않고(`MOAI_CONSTITUTION_DRY_RUN` 을 꺼서 실제 실행을 시도하는 것도 포함), 게이트가 사람의 승인을 요청하는 단계에 이르면 멈추고 리드에게 올린다. 운영자 대신 승인하지 않는다.
- **도구 출처.** 헌법 개정·검증은 이 트리에서 빌드해 경로로 호출한 `/tmp/t622-moai-tree` 로 하고, `CLAUDE_PROJECT_DIR`·`MOAI_CONSTITUTION_REGISTRY` 를 같은 호출 안에서 해제한다. 작업 디렉터리는 워크트리 루트다.
- **Frozen 구역.** spec-workflow.md 23행 블록·49행 블록·166행과(B6 (a)이면) worktree-integration.md 556행의 편집은 Frozen 구역 수정이다. Implementation Kickoff Approval에서 명시하고 progress §E.2에 줄마다 이전·이후 문장을 기록한다.
- **건드리지 않는 것.** `internal/hook/branch_guard.go`, `internal/constitution/**`(B4 (a)는 별도 카드), 설정 키 `workflow.worktree.auto_merge`, `AGENTS.md` §2 금지문, Route A 절차.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/...`)만 돌린다.
- **git 인덱스.** 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

- E1: AC-GDP-001~024 PASS/FAIL/N/A 표(011·012는 N/A). 행마다 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과.
- E3: `make agents-emit-check` RED·GREEN 출력.
- E4: `go test ./internal/template/ -run '…' -v -count=1` 출력과 최상위 PASS 줄 수(정확히 3).
- E5: 사본 diff 결과(다섯 파일 exit 0, 두 파일 차이 본문 동일, 레지스트리 L·T exit 0).
- E6: 커밋 SHA 목록과 `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋이라는 확인.
- E7: 분류 장부 세 개 — `ac01-classification.md`(명령), `b2-prose-classification.md`(서술형), `b1-pr-classification.md`(단계별 PR).
- E8: AC-GDP-001·006·009·010·018·019·021 읽기 단계 기록.
- E9: B4·B6·T1 결정 기록, Frozen 구역 수정 기록, Implementation Kickoff Approval 기록(progress §E.2).
- E10: 헌법 개정 증거 — `const-amend-evidence.md`, `const-amend-commands.txt`, dry-run 출력 두 개, 운영자 실행 출력 두 개, 리드 에스컬레이션 기록, 트리 빌드 정보, 검증 JSON.

## §F 마일스톤

### F.0 결정과 선결 항목

| 항목 | 상태 | 걸린 마일스톤 |
|---|---|---|
| OD-1 = 선택지 1 | 결정됨 (2026-09-10) | M4 |
| OD-2 = 선택지 B | 결정됨 (2026-09-10) | M3 |
| B1 = SPEC당 PR 하나 | 결정됨 (2026-09-10) | M4 |
| B2 = primary feature 브랜치 경로도 이동 | 결정됨 (2026-09-10) | M4 |
| B3 = `amend` 로 027·028 개정 | 결정됨 (2026-09-10) | M5 |
| B4 — amend 적용 단계 스텁 | **run-phase 착수 전 결정 필요** | M5, M4의 Step 1·Step 4 두 문장 |
| B6 — worktree-integration.md 범위 | **run-phase 착수 전 결정 필요** | M4 |
| T1 — Tier 재분류 | **리드 결정 필요** | 전체 |

M1·M2·M3와 M4의 대부분(등록된 두 문장 제외)은 B4와 무관하게 진행할 수 있다.

### M1 — AC-11: fetch 순서 보장

- 대상: REQ-GDP-001, 002, 003
- `manager-git.md` L·T 156 — fetch 를 먼저 끝내고 rev-list 는 그 뒤에 실행하도록 고친다.
- `agent-common-protocol.md` L·T Pre-Spawn 292·296·299 — Pre-Edit 347행 형태의 한 명령으로 순서를 보장한다. **이 편집은 run-phase의 마지막 지침 편집 커밋이다.**

### M2 — SX-R04: 병합 방식 해석

- 대상: REQ-GDP-004, 005
- `delivery.md` L·T 343·355 — `gh pr merge --<merge_method> --delete-branch` 와 해석 출처.
- `manager-git.md` L·T 114 — M4 절차 블록 재작성에 흡수.

### M3 — SX-R04: auto-merge 기본값 단일 기준 (OD-2 = B)

- 대상: REQ-GDP-006
- `delivery.md` L·T 335-338·348-349, `doc-execution.md` L·T 34-36 — 워크트리 기본 병합 문구 제거, `manager-git.md` 옵트인을 이름으로 밝힘. `manager-git.md` 148·166행은 유지.

### M4 — AC-01: Route B 흐름 재설계 (OD-1, B1, B2)

- 대상: REQ-GDP-007, 008, 009, 010, 016, 017, 018, 019
- `manager-git.md` L·T 88-137 — 절차 블록을 spec.md §C.2(E1-E3, C1-C3, X1-X4)로 다시 쓰고 C1 접두는 `feat/SPEC-<ID>`. 125·127·137행 문장도 맞춘다. 절 제목 유지. 42·160행은 §C.3 처리.
- `spec-workflow.md` L·T:
  - 21행 인용문 — 기본 흐름 서술을 Route A와 Route B 워크트리 흐름으로.
  - 23-26행 `[ZONE:Frozen]` Route 정의 — Route B "PR per phase" 를 SPEC당 PR 하나로(Frozen·미등록 → 기록).
  - 40-47행 Route B 표와 각주 — Step 1~3 위치를 같은 워크트리로, 브랜치를 `feat/SPEC-XXX` 하나로, 트리거를 SPEC PR 병합으로, Step 4를 L1 워크트리 폐기로; "one squash commit per phase" 를 SPEC당 한 번으로.
  - 49-62행 Step 규칙 — Step 1·Step 4의 등록 문장은 §C.4의 `--after` 문장으로(M5와 같은 커밋, B4 결정 뒤), Step 2·3은 같은 워크트리, late-branch 종결 명령 제거. `manager-git.md` § Late-Branch Invocation Pattern 참조는 유지.
  - 64-66행 안티패턴 — plan 워크트리 금지와 plan PR 병합 전제를 흐름에 맞춘다.
  - 166행 `[ZONE:Frozen]` Plan Phase — Route B는 워크트리 안에서(Frozen·미등록 → 기록).
  - 192·286행 Run·Sync Phase `[SHOULD]`, 316-362행 Plan to Run(Route B 트리거·전제·실행 위치, skip 정책 336행, 동시 plan-run 파이프라인 356-362행), 424-440행 Run to Sync·Sync close·Cleanup.
- `spec-assembly.md` L·T 332-340 — 진입 단계(E1)와 §C.2 참조로. 제목 유지.
- `delivery.md` L·T 50(Step 3.1 Route B 동기화 커밋 브랜치), 238-257(Step 3.2 github-flow 경로 통합과 primary 비-main 브랜치 멈춤), 319-328(Step 3.3.5 — spec.md §C.3 결과; 제목의 `Step 3.3.5` 유지).
- `worktree-integration.md` L·T 45·543-556 — B6 (a)일 때만.
- Frozen 수정 기록과 Kickoff 기록을 progress §E.2에 남긴다.

### M5 — 헌법 레지스트리 개정 (B3)

- 대상: REQ-GDP-020, 021, 022, 023
- 선결: B4 결정. 아래 순서는 B4 선택지 (a)·(b)에 공통인 게이트 기록 단계이며, 적용 단계는 B4 결정에 따라 확정한다.
1. 트리 빌드 `/tmp/t622-moai-tree` 와 빌드 정보 기록(§C 6).
2. 근거 문서 `.moai/reports/t622/run/const-amend-evidence.md` 작성 — 운영자 결정(B1·B3, 2026-09-10), 두 규칙의 `--before`·`--after`, 게이트 예측, 적용 단계 상태.
3. 레인 dry-run 두 번(spec.md §C.4 명령 원문), 명령 원문을 `const-amend-commands.txt` 에, 출력은 `const-amend-027-dryrun.txt`·`const-amend-028-dryrun.txt` 에 남긴다. 레지스트리가 아직 이전 clause 를 담고 있을 때 실행한다(`--before` 대조).
4. **멈추고 리드에게 올린다.** 명령 원문 두 줄(`--dry-run` 없는 형태), dry-run 출력, 근거 문서 경로를 담는다. progress §E.2에 "HumanOversight 단계에서 멈춤, 리드에게 올림" 을 기록한다.
5. 운영자가 대화형 터미널에서 두 명령을 실행한다. 첫 개정이 기록되면 두 번째는 24시간 뒤에만 4층을 통과한다(`rate_limiter.go:92`). 운영자 실행 출력은 리드를 거쳐 `const-amend-027-operator.txt`·`const-amend-028-operator.txt` 로 받는다.
6. 적용: B4 결정대로. 로컬 레지스트리가 바뀌면 템플릿 레지스트리를 줄 단위로 맞추고, spec-workflow.md L·T의 Step 1·Step 4 문장을 같은 커밋에서 `--after` 문장과 같게 한다(`DRIFT` 창을 남기지 않음).
7. 커밋 뒤 트리 빌드를 다시 만들어 `constitution validate --format json` 을 실행한다(AC-GDP-022, `vcs.revision` = HEAD).

### M6 — 사본 일치·생성물·중립성·순서 검증 (마지막)

- 대상: REQ-GDP-013, 014, 015 (+ AC-GDP-016)
1. 템플릿 `manager-git.md` 편집이 끝난 시점에 `make agents-emit-check` exit 1 관측 → `make agents-emit` → `make agents-emit-check` exit 0.
2. 사본 diff, 미러 테스트 3개와 최상위 PASS 줄 3개.
3. 템플릿 diff 추가 줄 중립성 검사.
4. 세 분류 장부 완성.
5. `agent-common-protocol.md` 커밋 뒤에 다른 범위 지침 파일(`zone-registry.md` 포함)을 고친 커밋이 없는지 확인.

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것.
- 브랜치 변경 명령 전체 부재를 합격 조건으로 쓰는 것.
- 금지 집합의 일부만 검출식에 넣는 것, 명령 검출식만으로 서술형 안내가 없다고 판정하는 것.
- primary checkout의 로컬 `main` 을 갱신하려고 흐름에 `git pull`·`git reset` 을 다시 넣는 것.
- Route B 표만 PR 하나로 고치고 Phase Transitions·skip 정책·Cleanup·`delivery.md` Step 3.1에 단계별 PR을 남기는 것.
- **`Y` 를 파이프·here-string·파일 리다이렉트로 amend 에 넘기는 것, 운영자 대신 승인하는 것.**
- `CLAUDE_PROJECT_DIR` 가 설정된 채로, 또는 설치 빌드로 amend·validate 를 실행하는 것.
- 로컬 레지스트리만 바꾸고 템플릿 레지스트리를 두는 것, 레지스트리를 바꾸고 원본 문장을 다른 커밋에 두어 `DRIFT` 창을 남기는 것.
- 검증기의 "escapes project dir" 거부 JSON(`"status": ""`, `"drift_count": 0`)을 통과로 읽는 것.
- 설정 키 `workflow.worktree.auto_merge` 를 기준으로 되살리는 것.
- 예시 명령만 지우고 절차의 존재를 알리는 문장을 남기는 것.
- `.toml` 을 손으로 맞추는 것, `agent-common-protocol.md` 를 run-phase 초반에 고치는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것, 개수가 찍히지 않은 grep 결과를 0으로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거·amend 동작), §C(결정 기록·흐름·개정 절차·남은 차단 항목), §D(제외 범위), §E(미검증·잔여 위험)
- `acceptance.md` (AC-GDP-001~024; 011·012 철회)
- `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`
- `internal/template/rule_template_mirror_test.go` — 바이트 동일 대상 목록
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R5-027`, `CONST-V3R5-028`, § Retiring an Entry
- `internal/cli/constitution.go`(143-153, 463-505, 510, 528-529), `internal/constitution/pipeline.go`(66, 115, 120, 140, 190, 227, 256-266), `frozen_guard.go:22`, `canary.go`(16, 18, 38), `contradiction.go`(87, 110, 124), `rate_limiter.go`(10, 12, 92), `human_oversight.go`(32, 44-45), `pipeline_test.go`(334-356, 406-423), `validator.go`
- `.claude/rules/moai/workflow/worktree-integration.md` — L1 진입·폐기, 워크트리 기준 브랜치, SPEC-to-Worktree Mapping(B6)
- `Makefile` — `agents-emit`, `agents-emit-check`
- SPEC-V3R5-LATE-BRANCH-001, SPEC-WORKTREE-BRANCH-GUARD-001, SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001, SPEC-MERGE-METHOD-CONFIG-001, SPEC-V3R2-CON-002
