# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 바뀔 가능성이 큰 결정(사용자에게 보이는 기본값 변경 OD-2 = B)을 먼저 적고, 기계적 치환과 생성물·검증은 뒤에 둔다. 항상 로드되는 규칙(`agent-common-protocol.md`) 편집은 마지막 지침 편집이다.

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M(spec.md §C.4), era V3R6.
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정): `BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0`. 0.2.0 작성 시점 HEAD `87988e946` 에서 범위 파일 네 개(L·T), `manager-git.toml`, 관련 테스트 파일, `Makefile` 은 `$BASE` 와 차이 없음.
- 범위: AC-11 fetch 순서, SX-R04 병합 방식 해석, OD-2 = B auto-merge 기본값 단일 기준과 그 파일들의 부수 의무. late-branch 재설계는 카드 t658, amend 적용 도우미 스텁은 카드 t659(spec.md §G).
- 결정: OD-2 = 선택지 B(2026-09-10), T1 = 분할(2026-09-11), 모두 운영자(리드 경유). 결정은 Implementation Kickoff Approval을 대신하지 않는다.
- plan-audit 회차: 이 판부터 1회차로 다시 센다. 이전 두 회차(FAIL 0.67, FAIL 0.75)는 전체 범위에 대한 기록이다(spec.md HISTORY).
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로).

## §B 알려진 문제

1. **`delivery.md` 두 사본은 의도적으로 다르다.** 차이는 275·278행과 479-480행뿐이다. 편집 자리(335-338, 343, 348-349, 355)는 두 사본에서 같은 줄이다. 파일 통째 복사 금지.
2. **`doc-execution.md` 두 사본도 의도적으로 다르다.** 로컬 사본에만 138-143행이 있다. 34-36행은 두 사본에서 같은 줄이다.
3. **사본 일치 가드가 파일마다 다르다.** `agent-common-protocol.md` 는 `TestSanitizedPairParity` 와 diff, `manager-git.md`·`delivery.md`·`doc-execution.md` 는 diff만 가드다. 바이트 동일 미러 테스트(`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`)는 범위 파일을 덮지 않는다(acceptance.md AC-GDP-013).
4. **`agent-common-protocol.md` 는 항상 로드되는 규칙이다.** 세션 도중 고치면 로드된 프롬프트 접두부가 무효화된다.
5. **`manager-git.md:32` 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 지우면 안 된다. 148·166행 옵트인 문장도 남는다.
6. **생성물 `.toml` 은 손으로 고치지 않는다.** 템플릿 `manager-git.md` 편집 뒤 `make agents-emit` 으로만 재생성한다.
7. **`manager-git.md:114` 는 t658이 다시 쓸 Late-Branch 절 안에 있다.** 이 SPEC은 그 줄의 병합 예시만 치환하고 주변 절차 줄은 건드리지 않는다.
8. **`agent-common-protocol.md` 17행은 `[ZONE:Frozen]` 이고 13·17·52행에 등록 Frozen clause 가 있다.** 편집 자리(290-305)와 떨어져 있으며 건드리지 않는다(REQ-GDP-024).
9. **워크트리 세션 가드**는 복합 명령 안의 `git`·`parallel` 낱말, 셸 변수를 받는 `sed`·`perl`, 경로를 만드는 반복문, 여러 명령을 이은 git 스크립트를 거부한다. 검출식은 `[g]it`·`para[l]lel`·`\x67it` 로 쓰고 경로는 글자 그대로 쓴다.
10. **셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛴다.** 판정은 `/usr/bin/grep` 으로 한다.

## §C 사전 점검 (run-phase 진입 시)

모두 읽기 전용이다. 결과는 `.moai/reports/t622/run/` 에 파일로 남기고 exit code를 따로 기록한다.

1. 트리 확인: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`. develop을 흡수했다면 `git diff --stat $BASE HEAD -- <범위 파일>` 이 비어 있는지 확인한다.
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`. 대상: `base-manager-git.md`, `base-agent-common-protocol.md`, `base-delivery.md`, `base-doc-execution.md`(로컬 경로), `base-delivery-template.md`, `base-doc-execution-template.md`(템플릿 경로), `base-manager-git.toml`.
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행한다. 기대 적중이 안 나오면 편집 전에 멈추고 blocker로 보고한다.
4. `make agents-emit-check` 기준선 exit 0.
5. 사본 diff 기준선: `manager-git.md`, `agent-common-protocol.md` exit 0; `delivery.md`·`doc-execution.md` exit 1(차이 줄번호 기록).
6. Frozen 기준선: AC-GDP-025의 네 clause 개수와 `[ZONE:Frozen]` 줄 위치 기록.

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지.
- **생성물.** 템플릿 `manager-git.md` 편집 뒤 `make agents-emit` 실행, 재생성된 `.toml` 을 같은 카드에 커밋, `make agents-emit-check` exit 0 을 증거로 기록. `.toml` 손편집 금지.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지.
- **항상 로드 규칙은 마지막에.** `agent-common-protocol.md` 편집은 run-phase의 마지막 지침 편집 커밋이다(AC-GDP-016).
- **Frozen 비접촉.** `[ZONE:Frozen]` 줄과 등록 Frozen clause 는 건드리지 않는다(REQ-GDP-024).
- **건드리지 않는 것.** `spec-workflow.md`, `spec-assembly.md`, `zone-registry.md`, `worktree-integration.md`, `manager-git.md` 의 114행 밖 Late-Branch 절 줄과 42·160·171행, `delivery.md` 356행과 Step 3.2·3.3.5, 설정 키 `workflow.worktree.auto_merge`, Go 코드.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/`)의 선택 테스트만 돌린다.
- **git 인덱스.** 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

- E1: AC-GDP-001~006, 013~016, 025 PASS/FAIL 표(자리표시 기준은 N/A). 행마다 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과(기준 트리에서 기대 적중이 나왔다는 기록).
- E3: `make agents-emit-check` RED(재생성 전)와 GREEN(재생성 뒤) 출력.
- E4: `go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1` 출력, 최상위 PASS 줄 수(정확히 2), 범위 파일 하위 테스트 흔적.
- E5: 사본 diff 결과(두 파일 exit 0, 두 파일 차이 본문 동일).
- E6: 커밋 SHA 목록과 `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋이라는 확인.
- E7: AC-GDP-001·006 읽기 단계 기록, AC-GDP-015 읽기 목록.
- E8: AC-GDP-025 Frozen 확인 결과.

## §F 마일스톤

### M1 — OD-2 = B: auto-merge 기본값 단일 기준 (사용자에게 보이는 기본값 변경)

- 대상: REQ-GDP-006
- `delivery.md` L·T 335-338(트리거) — 워크트리 문맥 기본 병합 조건을 없애고, 병합은 `manager-git.md` 옵트인(`--auto-merge` 와 전원 승인)을 따른다고 이름으로 밝힌다.
- `delivery.md` L·T 348-349(플래그 설명) — `--no-merge`·`--merge` 설명과 폐기 경고 문구를 새 기본값에 맞춘다.
- `doc-execution.md` L·T 34-36 — "worktree contexts default to auto-merge" 문장을 없애고, 병합 여부는 `manager-git.md` 옵트인이 정한다고 이름으로 밝힌다. 138-143행 사본 차이는 건드리지 않는다.
- `manager-git.md` 148·166행은 그대로 둔다. 설정 키 `workflow.worktree.auto_merge` 는 건드리지 않는다.

### M2 — SX-R04: 병합 방식 해석

- 대상: REQ-GDP-004, 005
- `delivery.md` L·T 343·355 — `gh pr merge --<merge_method> --delete-branch` 로 바꾸고 해석 출처(`git_strategy.<mode>.merge_method`, 기본 `squash`)를 한 번 명시한다. 275·278·479-480행 사본 차이는 건드리지 않는다.
- `manager-git.md` L·T 114 — 예시를 `gh pr merge <PR> --<merge_method> --delete-branch` 로 바꾼다. 32행 기본값 설명은 유지한다.

### M3 — AC-11: `manager-git.md` 동기화 절 순서

- 대상: REQ-GDP-001
- `manager-git.md` L·T 156 — `git fetch` 를 먼저 끝내고, 그 결과를 읽는 `git rev-list` 는 그 뒤에 실행한다고 고친다. fetch 결과를 읽지 않는 명령은 병렬로 둘 수 있다.

### M4 — 생성물·사본·중립성·Frozen 확인

- 대상: REQ-GDP-013, 014, 015, 024
1. 템플릿 `manager-git.md` 편집(M2·M3)이 끝난 시점에 `make agents-emit-check` 를 먼저 돌려 exit 1 을 관측한다.
2. `make agents-emit` → `make agents-emit-check` exit 0, 재생성된 `.toml` 커밋.
3. 사본 diff(두 파일 exit 0, 두 파일 차이 본문 동일).
4. 템플릿 diff 추가 줄 중립성 검사(AC-GDP-015).
5. Frozen 확인(AC-GDP-025).

### M5 — AC-11: Pre-Spawn Sync Check (마지막 지침 편집)

- 대상: REQ-GDP-002, 003
- `agent-common-protocol.md` L·T 292·296·299 — 코드 블록 안에서 Pre-Edit 347행과 같은 형태(`git fetch origin main 2>&1; git rev-list …` 한 명령)로 순서를 보장한다. 세 번째 명령과 두 해석 표는 그대로 둔다. Pre-Edit 절(328행~`#### The sweep prohibition` 앞)은 건드리지 않는다.
- **이 편집은 run-phase의 마지막 지침 편집 커밋이다.**

### M6 — 최종 검증 (편집 없음)

- 대상: 판정 대상 전체
1. M5 뒤에 사본 diff와 AC-GDP-025를 다시 실행한다(`agent-common-protocol.md` 가 M5에서 바뀌므로).
2. 선택 테스트 두 개와 최상위 PASS 줄 2개, 범위 파일 하위 테스트 흔적(AC-GDP-013).
3. `agent-common-protocol.md` 커밋 뒤에 다른 범위 지침 파일을 고친 커밋이 없는지 확인(AC-GDP-016).

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것 (`delivery.md`·`doc-execution.md` 의도된 차이 소실).
- 범위 파일을 덮지 않는 미러 테스트를 돌려 사본 일치 증거로 쓰는 것.
- 설정 키 `workflow.worktree.auto_merge` 를 기준으로 되살리는 것.
- `manager-git.md:32` 기본값 설명이나 148·166행 옵트인 문장을 지우는 것.
- `manager-git.md:114` 를 고치면서 주변 late-branch 절차 줄까지 손대는 것(t658 소관).
- `.toml` 을 손으로 맞추는 것.
- `agent-common-protocol.md` 를 run-phase 초반에 고치는 것, 17행 `[ZONE:Frozen]` 이나 등록 clause 를 건드리는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것, 개수가 찍히지 않은 grep 결과를 0으로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거·Frozen 확인), §C(결정 기록·번호 방식·티어), §D(제외 범위), §E(미검증·잔여 위험), §G(t658로 옮긴 항목)
- `acceptance.md` (판정 대상 AC-GDP-001~006, 013~016, 025)
- `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`
- `internal/template/rule_template_mirror_test.go`, `internal/template/sanitized_pair_parity_test.go`, `internal/template/internal_content_leak_test.go`
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R2-006`·`036`·`037`·`038`
- `Makefile` — `agents-emit`, `agents-emit-check`
- SPEC-MERGE-METHOD-CONFIG-001
