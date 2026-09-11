# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 바뀔 가능성이 큰 결정(사용자에게 보이는 기본값·플래그·모드 조건 변경)을 먼저 적고, 기계적 치환과 생성물·검증은 뒤에 둔다. 항상 로드되는 규칙(`agent-common-protocol.md`) 편집은 마지막 지침 편집이다.

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M, era V3R6. 판정 대상 REQ 12·AC 16으로 Tier M 상한(16/16) 안이다. 영향 파일 19개(명령 게시본이 바뀌면 21개)는 Tier M 파일 수 안내를 넘지만, 리드가 "요구사항·수용 기준 수로 정하고 파일 수 안내는 참고" 로 판정했다(spec.md §C.1·§C.4).
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정): `BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0`. 0.2.2 작성 시점 HEAD `ea09ca650`(부모 `caa601d7c`, 둘 다 SPEC 파일만 바꿈)에서 범위 파일 아홉 개(L·T), `manager-git.toml`, 게시본 `moai-sync/SKILL.md`(L·T), `internal/template/commandemit`·`agentemit`, `zone-registry.md`(L·T), 사본·발행 관련 테스트 파일, `docs-site/content`, `Makefile` 은 `$BASE` 와 차이 없음.
- 범위: AC-11 fetch 순서, SX-R04 병합 방식 해석, OD-2 = B auto-merge 옵트인 단일 기준, `/moai sync` 플래그 의미와 모드별 승인 조건, 소비자 X1~X3(사용법 줄·슬래시 명령 힌트·다음 단계 선택지), 명령 원본 편집에 따른 게시본 발행, 그 파일들의 부수 의무. late-branch 재설계는 카드 t658, amend 적용 도우미 스텁은 카드 t659(spec.md §G). docs-site 네 로케일(X4)은 후속 문서 카드(spec.md §D).
- 결정: OD-2 = 선택지 B(2026-09-10), T1 = 분할(2026-09-11), OD-2 하위 결정 플래그 의미(2026-09-11) — 운영자(리드 경유). 모드별 승인 조건(2026-09-11), 소비자 범위 판정 X1~X4(2026-09-11) — 리드 판정. 결정은 Implementation Kickoff Approval을 대신하지 않는다.
- plan-audit 회차: 분할 뒤 1회차 FAIL 0.71(`.moai/reports/t622/plan-audit-reduced-iter1.md`). 0.2.1 판에 대한 감사 파일은 `.moai/reports/t622/` 에 없다(0.2.2 작성 시점 목록). 이 판이 축소판 2회차 대상이다. 분할 전 두 회차(FAIL 0.67, FAIL 0.75)는 전체 범위 기록이다.
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로).
- 남은 차단 항목: 없음.

## §B 알려진 문제

1. **의도된 사본 차이가 있는 파일이 여섯 개다.** `delivery.md`(275·278·479-480), `doc-execution.md`(로컬 전용 138-143), `moai/SKILL.md`(20개 덩어리), `references/reference.md`(229), `workflows/sync.md`(로컬 전용 65-74, 81), 명령 원본(`2c2` — 2행 `description`). 편집 자리는 모두 이 덩어리 밖이다. 파일 통째 복사 금지.
2. **`workflows/sync.md` 는 사본마다 줄번호가 다르다.** 플래그 줄은 로컬 114 / 템플릿 104, 사용법 줄(X1)은 로컬 95 / 템플릿 85. 줄번호가 아니라 `**Flags**:` 와 `/moai sync [mode]` 표지로 찾는다.
3. **명령 원본은 로컬과 템플릿의 파일 이름과 형식이 다르다.** 로컬 `.claude/commands/moai/sync.md` 는 렌더링된 영어 사본, 템플릿 `.claude/commands/moai/sync.md.tmpl` 은 2행 `description` 이 로케일 조건문인 Go 템플릿이다. X2가 고치는 3행 `argument-hint` 에는 템플릿 액션이 없어 두 사본에 글자 그대로 같은 줄을 넣는다. 2행과 본문은 건드리지 않는다.
4. **발행기는 템플릿 트리의 게시본만 쓴다.** `internal/template/commandemit/golden_test.go:28` `templatesDir = "../templates"`, `:67-74` 갱신 분기가 템플릿 게시본 경로에 쓴다. 로컬 게시본 `.agents/skills/moai-sync/SKILL.md` 는 추적 파일이지만 발행기가 쓰지 않는다 — 게시본이 바뀌면(AC-GDP-030 경우 B) 로컬 사본은 템플릿 게시본을 바이트 그대로 복사해 맞춘다.
5. **`argument-hint` 는 게시본에 실리지 않는다.** `loader.go:4-6`("Claude-only keys and are NOT carried into the published skill"), `emit.go:122-130` `renderSkill` 은 생성 머리말·`name`·`description`·본문만 쓴다. 그래서 경우 (A)(변화 없음)가 예상이지만, 예상은 판정이 아니다 — 두 경우를 명령으로 가른다.
6. **`TestCommandSourcesUnmodified`(`golden_test.go:94-106`)는 한 실행 안에서 발행 전후 해시를 비교한다.** 원본 편집 자체로는 실패하지 않는다. 원본을 고치면 실패하는 검사는 `make commands-emit-check` 쪽이며, 원본 편집이 게시본을 바꾸는 경우에만 빨강이다.
7. **사본 일치 가드가 파일마다 다르다.** `agent-common-protocol.md` 만 `TestSanitizedPairParity` 와 diff, 나머지 여덟 파일은 diff만 가드다(acceptance.md AC-GDP-013 표).
8. **`agent-common-protocol.md` 는 항상 로드되는 규칙이다.** 세션 도중 고치면 로드된 프롬프트 접두부가 무효화된다.
9. **`manager-git.md:32` 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 지우면 안 된다. 148·166행 `--auto-merge` 조건도 남는다.
10. **`manager-git.md` 의 PR Auto-Merge 절은 지금 team 한정이다.** 절 제목은 "## PR Auto-Merge" 로 시작하게 유지하고(수용 기준 추출 표지), personal·manual 규칙을 이 절에 모드 이름을 담은 문장으로 더한다. `delivery.md` Step 3.4에도 같은 두 조건을 적는다.
11. **생성물 `.toml` 과 게시본은 손으로 고치지 않는다.** `.toml` 은 `make agents-emit`, 템플릿 게시본은 `make commands-emit` 으로만 다시 만든다(§B 4의 로컬 게시본 복사는 예외).
12. **`manager-git.md:114` 는 t658이 다시 쓸 Late-Branch 절 안에 있다.** 이 SPEC은 그 줄의 병합 예시만 치환한다.
13. **`agent-common-protocol.md` 17행은 `[ZONE:Frozen]` 이고 13·17·52행에 등록 Frozen clause 가 있다.** 편집 자리(290-305)와 떨어져 있으며 건드리지 않는다(REQ-GDP-024). 나머지 여덟 파일(명령 원본 포함)에는 `[ZONE:]` 태그와 레지스트리 항목이 없다.
14. **줄 단위 검출식의 모양.** AC-GDP-026~029는 한 줄에 조건을 모아 적어야 잡힌다(검출식과 읽기 기록은 acceptance.md). 받아들이는 문구 모양:
    - `--merge` 를 말하는 줄은 `--merge` 뒤에 "deprecated alias of `--auto-merge`"(또는 "deprecated alias for") 구절을 담는다. 예: "- `--merge`: deprecated alias of `--auto-merge` (logs a warning)", 한 줄 목록이면 "--merge (deprecated alias of --auto-merge)". `--auto-merge` 를 폐기됐다고 적거나 `--merge` 를 별칭 구절 없이 적은 줄은 FAIL이다.
    - `--merge` 는 `moai/SKILL.md` Flags 줄, `workflows/sync.md` 의 `**Flags**:` 줄, Supported Flags 절, `delivery.md` Step 3.4 절에 각각 1줄 이상 남긴다(AC-GDP-026 (iii)). 나머지 조각에서는 빼도 된다.
    - `--no-merge` 를 말하는 줄은 "no-op" 과 "deprecated" 를 함께 담고, 효과를 말하는 낱말(skip, not set, prevent, unless, disable, override, turn off, suppress, bypass, cancel)을 담지 않는다. 예: "- `--no-merge`: Deprecated no-op kept for compatibility (logs a warning); not merging is already the default."
    - team 조건 줄은 "team mode"·`--auto-merge` 와 "all … approvals"(all 과 approv 사이 낱말 2개까지)를 함께 담는다. 예: "In team mode, `--auto-merge` merges only after all approvals are obtained."
    - personal·manual 조건 줄은 두 모드 이름과 `--auto-merge` 를 함께 담고, 승인·리뷰를 말하면 "without an approval condition", "without requiring approval", "not needed", "no teammates" 같은 없음 표현으로만 말한다. 예: "In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve)."
    - X1 사용법 줄과 X2 `argument-hint` 는 한 줄짜리 형식이라 별칭 구절을 넣기 어렵다 — `[--merge]` 를 남기려면 같은 줄에 별칭 구절이 있어야 하므로, 이 두 줄에서는 `--merge` 를 빼는 편이 쉽다. 폐기된 별칭 서술은 Flags 줄·Supported Flags 절·Step 3.4 절이 맡는다.
15. **워크트리 세션 가드**는 복합 명령 안의 `git`·`parallel` 낱말, 셸 변수를 받는 `sed`·`perl`, 경로를 만드는 반복문, 여러 명령을 이은 git 스크립트를 거부한다. 검출식은 `[g]it`·`para[l]lel`·`\x67it` 로 쓰고(`manager-[g]it` 포함) 경로는 글자 그대로 쓴다.
16. **셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛴다.** 판정은 `/usr/bin/grep` 으로 한다.

## §C 사전 점검 (run-phase 진입 시)

모두 읽기 전용이다(5번의 발행 대조만 추적 파일 하나를 잠시 바꿨다가 되돌린다). 결과는 `.moai/reports/t622/run/` 에 파일로 남기고 exit code를 따로 기록한다.

1. 트리 확인: `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`. develop을 흡수했다면 `git diff --stat $BASE HEAD -- <범위 파일·생성물·게시본·발행기>` 가 비어 있는지 확인한다.
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`. 명령마다 경로를 글자 그대로 쓴다.
   - 로컬 경로 → `base-<이름>.md`: `manager-git`(`.claude/agents/moai/manager-git.md`), `agent-common-protocol`, `delivery`, `doc-execution`, `skill`(`.claude/skills/moai/SKILL.md`), `reference`, `qgc`(`quality-gates-context.md`), `sync`(`.claude/skills/moai/workflows/sync.md`), `command-sync`(`.claude/commands/moai/sync.md`), `published-sync`(`.agents/skills/moai-sync/SKILL.md`).
   - 템플릿 경로 → `base-<이름>-template.md`: `delivery`, `doc-execution`, `skill`, `reference`, `sync`, `command-sync`(`internal/template/templates/.claude/commands/moai/sync.md.tmpl`), `published-sync`(`internal/template/templates/.agents/skills/moai-sync/SKILL.md`).
   - 생성물 `base-manager-git.toml`.
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행한다. 기대 적중이 안 나오면 편집 전에 멈추고 blocker로 보고한다.
4. 생성물 점검 기준선: `make agents-emit-check` exit 0, `make commands-emit-check` exit 0.
5. AC-GDP-030 발행 점검 양성 대조(c): 템플릿 게시본 한 파일을 잠시 바꾼 상태에서 `make commands-emit-check` exit 1 → 백업으로 되돌림 → `cmp` exit 0 → `make commands-emit-check` exit 0 → 게시본 경로 `git status --porcelain` 빈 출력. 되돌림을 확인하지 못하면 멈추고 보고한다.
6. 사본 diff 기준선: `manager-git.md`·`agent-common-protocol.md`·`quality-gates-context.md`·게시본 exit 0; 나머지 여섯 파일 exit 1(차이 줄번호 기록).
7. Frozen 기준선: AC-GDP-025의 네 clause 개수, `[ZONE:Frozen]` 줄 위치, 레지스트리 양성 대조 13·기타 0, 기타 검출식 뮤턴트 픽스처 2.

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지(명령 원본은 형식이 달라 통째 복사하면 템플릿 조건문이 사라진다).
- **생성물 두 가지.**
  - 에이전트: 템플릿 `manager-git.md` 편집을 모두 마친 뒤 `make agents-emit`, 재생성된 `.toml` 을 같은 카드에 커밋, `make agents-emit-check` exit 0.
  - 명령 게시본: 템플릿 명령 원본 편집 직후 `make commands-emit` → `make commands-emit-check` exit 0. 바뀐 게시본이 있으면 원본 편집과 같은 커밋에 넣는다.
  - 두 발행기는 입력과 출력이 겹치지 않는다(명령 원본 → `.agents/skills/`, `manager-git.md` → `.codex/agents/`). 순서는 같은 커밋 의무에 따라 정한다: `commands-emit` 은 M1의 명령 원본 커밋에, `agents-emit` 은 M4에 둔다.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지.
- **항상 로드 규칙은 마지막에.** `agent-common-protocol.md` 편집은 run-phase의 마지막 지침 편집 커밋이다(AC-GDP-016).
- **Frozen 비접촉.** `[ZONE:Frozen]` 줄과 등록 Frozen clause 는 건드리지 않는다(REQ-GDP-024).
- **건드리지 않는 것.** `spec-workflow.md`, `spec-assembly.md`, `zone-registry.md`, `worktree-integration.md`, `manager-git.md` 의 114행 밖 Late-Branch 절 줄과 42·160·171행, `delivery.md` 356행과 Step 3.2·3.3.5, 명령 원본 2행 `description` 과 본문, 발행기 코드, docs-site(X4 — 후속 문서 카드), 설정 키 `workflow.worktree.auto_merge`, Go 코드.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/`)의 선택 테스트와 두 `make` 점검 타깃만 돌린다.
- **git 인덱스.** 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

- E1: AC-GDP-001~006, 013~016, 025~030 PASS/FAIL 표(자리표시 기준은 N/A). 행마다 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과(기준 트리에서 기대 적중이 나왔다는 기록).
- E3: 에이전트 생성물 — `make agents-emit-check` RED(재생성 전)와 GREEN(재생성 뒤) 출력, `.toml` 두 절 diff.
- E4: `go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1` 출력, 최상위 PASS 줄 수(정확히 2), 범위 파일 하위 테스트 흔적.
- E5: 사본 diff 결과(네 파일 exit 0, 여섯 파일 차이 본문 동일, 명령 원본 `argument-hint` 줄 일치).
- E6: 커밋 SHA 목록과 `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋이라는 확인.
- E7: 읽기 기록 `ac001-reading.md`·`ac006-reading.md`·`ac026-reading.md`·`ac028-reading.md`(PASS 전제), AC-GDP-015 읽기 목록.
- E8: AC-GDP-025 Frozen 확인 결과.
- E9: AC-GDP-026~029 아홉 조각별 판정 파일, `--merge` 존재 개수(`ac026-merge-presence.txt`), `--no-merge` 줄 수.
- E10: 명령 게시본 — AC-GDP-030 발행 대조(c)의 네 출력, `commands-emit`·`commands-emit-check` 출력, `ac030-outcome.md`(경우 이름과 근거 파일 내용).

## §F 마일스톤

### M1 — OD-2 = B와 하위 결정: 옵트인 단일 기준, 플래그 의미, 모드별 승인 조건, 소비자 X1~X3

- 대상: REQ-GDP-006, 025, 026, 그리고 X2의 발행 의무로 REQ-GDP-014
- `delivery.md` L·T Step 3.4:
  - 335-338(트리거) — 워크트리 문맥 기본 병합과 `--no-merge` 조건을 없애고, 트리거를 "`--auto-merge`(또는 폐기된 별칭 `--merge`)가 주어짐" 으로 바꾼다. 병합 기준이 `manager-git.md` 옵트인임을 이름으로 밝힌다.
  - 모드 조건 — "team mode" 는 `--auto-merge` 와 전원 승인, "personal and manual modes" 는 `--auto-merge` 로 승인 조건 없이 병합(승인할 팀원 없음)을 각각 한 줄로 적는다. CI 통과·충돌 없음 확인은 그대로 둔다.
  - 348-349(플래그 설명) — `--auto-merge` 설명을 더하고, `--merge` 는 `--auto-merge` 의 폐기된 별칭(경고), `--no-merge` 는 폐기된 호환용 no-op(경고, 병합하지 않는 것이 이미 기본)으로 적는다.
- `delivery.md` L·T Context-Aware Next Steps 절 404행(X3) — 선택지를 `--auto-merge` 로 바꾼다. `--merge` 를 남기면 같은 줄에 폐기 표시와 `--auto-merge` 가 함께 있어야 한다.
- `doc-execution.md` L·T 34-36 — "worktree contexts default to auto-merge" 문장을 없애고, 병합 여부는 `manager-git.md` 옵트인(`--auto-merge`)이 정한다고 이름으로 밝힌다. 138-143행 사본 차이는 건드리지 않는다.
- `manager-git.md` L·T 164-171 PR Auto-Merge 절 — 제목은 "## PR Auto-Merge" 로 시작하게 두고, team 모드 조건 줄과 personal·manual 조건 줄을 적는다. 148행 Team Mode 항목의 `--auto-merge` 조건은 유지하고, 필요하면 133-140 Personal Mode 절에서 이 절을 가리킨다. 171행 5단계는 건드리지 않는다(t658).
- 플래그 표면 L·T — `moai/SKILL.md:140` Flags 줄, `references/reference.md:161`, `quality-gates-context.md:30`·`:101`, `workflows/sync.md` `**Flags**:` 줄(로컬 114 / 템플릿 104)에 `--auto-merge` 를 노출하고 `--merge` 를 "deprecated alias of `--auto-merge`" 로 적는다.
- `workflows/sync.md` L·T 사용법 줄(X1, 로컬 95 / 템플릿 85) — `[--auto-merge]` 를 노출한다(§B 14).
- **명령 원본(X2)은 별도 커밋으로 묶는다.** 순서:
  1. 로컬 `.claude/commands/moai/sync.md:3` 과 템플릿 `sync.md.tmpl:3` 의 `argument-hint` 에 `[--auto-merge]` 를 노출한다(두 줄 글자 그대로 같게, §B 14).
  2. 곧바로 `make commands-emit` → `make commands-emit-check`(exit 0) → `git status --porcelain -- internal/template/templates/.agents/skills/ .agents/skills/` 판독.
  3. 판독이 비어 있으면 경우 (A) 후보 — 원본 두 파일만 커밋한다. 비어 있지 않으면 경우 (B) 후보 — 로컬 게시본을 템플릿 게시본으로 복사하고, 원본 두 파일과 게시본 두 파일을 한 커밋에 넣는다.
  4. 커밋 뒤 AC-GDP-030 커밋 판정 명령을 실행해 `ac030-outcome.md` 를 남긴다.
- 설정 키 `workflow.worktree.auto_merge` 는 건드리지 않는다.

### M2 — SX-R04: 병합 방식 해석

- 대상: REQ-GDP-004, 005
- `delivery.md` L·T 343·355 — `gh pr merge --<merge_method> --delete-branch` 로 바꾸고 해석 출처(`git_strategy.<mode>.merge_method`, 기본 `squash`)를 한 번 명시한다.
- `manager-git.md` L·T 114 — 예시를 `gh pr merge <PR> --<merge_method> --delete-branch` 로 바꾼다. 32행 기본값 설명은 유지한다.

### M3 — AC-11: `manager-git.md` 동기화 절 순서

- 대상: REQ-GDP-001
- `manager-git.md` L·T 156 — `git fetch` 를 먼저 끝내고, 그 결과를 읽는 `git rev-list` 는 그 뒤에 실행한다고 고친다. fetch 결과를 읽지 않는 명령은 병렬로 둘 수 있다. fetch 와 rev-list 를 같은 목록·표에 두지 않는다.

### M4 — 생성물·사본·중립성·Frozen 확인

- 대상: REQ-GDP-013, 014, 015, 024
1. 템플릿 `manager-git.md` 편집(M1·M2·M3)이 끝난 시점에 `make agents-emit-check` 를 먼저 돌려 exit 1 을 관측한다.
2. `make agents-emit` → `make agents-emit-check` exit 0, `.toml` 두 절 diff exit 0, 재생성된 `.toml` 커밋.
3. `make commands-emit-check` 를 다시 돌려 exit 0 을 확인한다(읽기 전용 — M1 명령 원본 커밋 뒤 게시본 드리프트가 없다는 기록).
4. 사본 diff(네 파일 exit 0, 여섯 파일 차이 본문 동일, 명령 원본 `argument-hint` 줄 일치).
5. 템플릿 diff 추가 줄 중립성 검사(AC-GDP-015).
6. Frozen 확인(AC-GDP-025).

### M5 — AC-11: Pre-Spawn Sync Check (마지막 지침 편집)

- 대상: REQ-GDP-002, 003
- `agent-common-protocol.md` L·T 292·296·299 — 코드 블록 안에서 Pre-Edit 347행과 같은 형태(`git fetch origin main 2>&1; git rev-list …` 한 명령)로 순서를 보장한다. 세 번째 명령과 두 해석 표는 그대로 둔다. Pre-Edit 절은 건드리지 않는다.
- **이 편집은 run-phase의 마지막 지침 편집 커밋이다.** 이 커밋 뒤에는 범위 지침 파일(명령 원본 포함)을 고친 커밋이 없어야 한다.

### M6 — 최종 검증 (편집 없음)

- 대상: 판정 대상 전체
1. M5 뒤에 사본 diff와 AC-GDP-025를 다시 실행한다.
2. `make agents-emit-check`·`make commands-emit-check` 를 다시 실행해 둘 다 exit 0 을 확인한다(읽기 전용).
3. 선택 테스트 두 개와 최상위 PASS 줄 2개, 범위 파일 하위 테스트 흔적(AC-GDP-013).
4. 읽기 기록 네 개 확인(`ac001`·`ac006`·`ac026`·`ac028`, acceptance.md §D.3).
5. `agent-common-protocol.md` 커밋 뒤에 다른 범위 지침 파일을 고친 커밋이 없는지 확인(AC-GDP-016).

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것(의도된 차이 소실 — 명령 원본은 템플릿 조건문까지 사라진다).
- 범위 파일을 덮지 않는 미러 테스트를 돌려 사본 일치 증거로 쓰는 것.
- `--merge` 의 폐기를 취소하거나 `--auto-merge` 와 별개 플래그로 적는 것, `--no-merge` 에 건너뜀·조건 의미를 남기는 것.
- team 모드 승인 조건을 지우거나, personal·manual 모드에 승인 조건을 붙이는 것, 두 문서 중 한쪽에만 모드 조건을 적는 것.
- 설정 키 `workflow.worktree.auto_merge` 를 기준으로 되살리는 것.
- docs-site를 이 카드에서 고치는 것(X4는 후속 문서 카드 소관).
- 명령 원본만 커밋하고 게시본 발행을 다른 커밋으로 미루는 것, `make commands-emit` 을 실행하지 않고 경우 (A)라고 적는 것.
- 템플릿 게시본만 바꾸고 로컬 게시본 사본을 두고 오는 것, 게시본을 손으로 고치는 것.
- `manager-git.md:32` 기본값 설명이나 148·166행 `--auto-merge` 조건을 지우는 것.
- `manager-git.md:114` 를 고치면서 주변 late-branch 절차 줄까지 손대는 것(t658 소관).
- `.toml` 을 손으로 맞추는 것.
- `agent-common-protocol.md` 를 run-phase 초반에 고치는 것, 17행 `[ZONE:Frozen]` 이나 등록 clause 를 건드리는 것.
- 읽기 기록 없이 AC-GDP-001·006·026·028·029를 PASS로 적는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것, 개수가 찍히지 않은 grep 결과를 0으로 읽는 것, 빈 diff·빈 커밋 목록을 통과로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거·Frozen 확인·소비자 조사), §C(결정 기록·번호 방식·티어), §D(제외 범위·X4 후속 카드 입력), §E(미검증·잔여 위험), §G(t658로 옮긴 항목)
- `acceptance.md` (판정 대상 AC-GDP-001~006, 013~016, 025~030)
- `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`, `.moai/reports/t622/plan-audit-reduced-iter1.md`
- `internal/template/rule_template_mirror_test.go`, `sanitized_pair_parity_test.go`, `internal_content_leak_test.go`, `backlog_json_disclosure_mirror_test.go`, `template_neutrality_audit_test.go`, `agentless_audit_test.go`, `agent_frontmatter_audit_test.go`, `published_skills_deploy_test.go`
- `internal/template/commandemit/` — `commandemit.go:49-53`(`DefaultOptions` 루트), `loader.go:4-6`(발행에 실리지 않는 키), `emit.go:122-130`(`renderSkill`), `golden_test.go:28`(`templatesDir`)·`:57-74`(갱신 분기)·`:94-106`(`TestCommandSourcesUnmodified`)
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R2-006`·`036`·`037`·`038`
- `Makefile:38-49` — `agents-emit`, `agents-emit-check`; `Makefile:51-60` — `commands-emit`, `commands-emit-check`
- SPEC-MERGE-METHOD-CONFIG-001
