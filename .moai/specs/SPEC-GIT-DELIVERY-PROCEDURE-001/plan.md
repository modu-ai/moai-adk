# Plan — SPEC-GIT-DELIVERY-PROCEDURE-001

> 구현 계획. 경로는 워크트리 루트 기준. 바뀔 가능성이 큰 결정(사용자에게 보이는 기본값·플래그·모드 조건 변경)을 먼저 적고, 기계적 치환과 생성물·검증은 뒤에 둔다. 항상 로드되는 규칙 `agent-common-protocol.md` 는 편집하지 않는다 — REQ-GDP-002 는 develop 카드 t635·dr0911 로 이미 충족돼 있다(0.2.5, 리드 판단 (a)).

## §A 맥락

- 카드: t622 (지침 감사 G1). Class C, Tier M, era V3R6. 판정 대상 REQ 12·AC 16으로 Tier M 상한(16/16) 안이다. 이 카드가 바꾸는 파일 17개(명령 게시본이 바뀌면 19개 — 0.2.5에서 `agent-common-protocol.md` 두 사본이 편집 대상에서 빠짐)는 Tier M 파일 수 안내를 넘지만, 리드가 "요구사항·수용 기준 수로 정하고 파일 수 안내는 참고" 로 판정했다(spec.md §C.1·§C.4).
- 워크트리: `.claude/worktrees/t622`, 브랜치 `WT-git-procedure-fixes`.
- 기준 트리(R1 고정, 0.2.5 재고정): `BASE=255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089` — 로컬 develop `f1f034bb4` 를 두 번째로 흡수한 병합 커밋. 이 카드는 아직 범위 파일을 고치지 않았으므로 이 커밋이 run-phase 이전 상태다. 0.2.5 작성 시점 HEAD `b24f2e184` 와의 차이는 보고서 두 파일뿐이고, 범위 루트 작업 트리는 BASE 와 차이 없다(spec.md §A.1, 증거 `.moai/reports/t622/reanchor/`). 0.2.4 까지의 기준 `b412f8a33` 은 HISTORY 기록이다. BASE 는 고정 스냅숏(기준 사본 반출·양성 대조·사본 덩어리·미러 기준선·AC-GDP-002·003 기준 절)에만 쓴다. "이 카드가 바꾼 것" 을 재는 범위의 왼쪽 끝은 읽는 시점의 `git merge-base develop HEAD`(`$CARD_BASE`)다 — spec.md §E.2, acceptance.md 관례(0.2.6).
- 줄번호: 이 문서의 줄번호는 템플릿 사본 기준이다. 로컬 사본이 다른 곳 — `manager-git.md` 6행부터 +2, `delivery.md` 템플릿 279~421행 구간 +25, `workflows/sync.md` 65행 이후 +10 — 은 괄호에 로컬 값을 적는다(spec.md §A.1 표). 편집·판정은 줄번호가 아니라 표지로 자리를 찾는다.
- 범위: AC-11 fetch 순서, SX-R04 병합 방식 해석, OD-2 = B auto-merge 옵트인 단일 기준, `/moai sync` 플래그 의미와 모드별 승인 조건, 소비자 X1~X3(사용법 줄·슬래시 명령 힌트·다음 단계 선택지), 명령 원본 편집에 따른 게시본 발행, 그 파일들의 부수 의무. late-branch 재설계는 카드 t658, amend 적용 도우미 스텁은 카드 t659(spec.md §G). docs-site 네 로케일(X4)은 후속 문서 카드(spec.md §D).
- 결정: OD-2 = 선택지 B(2026-09-10), T1 = 분할(2026-09-11), OD-2 하위 결정 플래그 의미(2026-09-11) — 운영자(리드 경유). 모드별 승인 조건(2026-09-11), 소비자 범위 판정 X1~X4(2026-09-11) — 리드 판정. 흡수 뒤 REQ-GDP-002 처리 선택지 (a)(2026-09-11) — 리드 판단(운영자 경유, 최종): t635 형태가 충족, `agent-common-protocol.md` 비편집·비미러, Kickoff 승인 유지. 결정은 Implementation Kickoff Approval을 대신하지 않는다.
- 0.2.5 재고정 뒤의 plan-auditor 는 REQ-GDP-002·AC-GDP-002 변경분과 그 파생 항목(BASE, 사본 기준선, M5, AC-GDP-016, 미러 테스트 비회귀 점검, 로컬 줄 인용)만 본다. 그 감사(`.moai/reports/t622/plan-audit-reanchor.md`, FAIL 0.75)의 D1~D6 을 0.2.6 에서 반영했다 — 다음 감사는 D1~D3 과 그로 인한 퇴행만 본다.
- plan-audit 회차: 분할 뒤 1회차 FAIL 0.71(`.moai/reports/t622/plan-audit-reduced-iter1.md`). 0.2.1 판에 대한 감사 파일은 `.moai/reports/t622/` 에 없다(0.2.2 작성 시점 목록). 이 판이 축소판 2회차 대상이다. 분할 전 두 회차(FAIL 0.67, FAIL 0.75)는 전체 범위 기록이다.
- run-phase 증거 디렉터리: `.moai/reports/t622/run/` (추적 경로).
- 남은 차단 항목: 없음.

## §B 알려진 문제

1. **의도된 사본 차이가 있는 파일이 여덟 개다(BASE 측정, spec.md §A.4).** `manager-git.md`(`5,7c5`), `delivery.md`(9개 덩어리), `doc-execution.md`(9개), `quality-gates-context.md`(7개), `moai/SKILL.md`(20개), `references/reference.md`(`229d228`), `workflows/sync.md`(`29,31c29,31`, 로컬 전용 65-74, 81), 명령 원본(`2c2` — 2행 `description`). develop 이 로컬 사본에만 넣은 편집이 그 가운데 넷(`manager-git.md`, `delivery.md`, `doc-execution.md`, `quality-gates-context.md`)과 `workflows/sync.md` 의 `29,31c29,31` 이다. 편집 자리는 모두 이 덩어리 밖이다(BASE 에서 확인). 바이트 동일로 남은 쌍은 `agent-common-protocol.md` 와 게시본 둘이다. 파일 통째 복사 금지 — 통째 복사는 develop 의 로컬 편집이나 템플릿 조건문을 지운다.
2. **사본마다 줄번호가 다른 파일이 셋이다.** `workflows/sync.md` — 플래그 줄 로컬 114 / 템플릿 104, 사용법 줄(X1) 로컬 95 / 템플릿 85. `manager-git.md` — 로컬이 6행부터 +2(예: 114 → 116, 156 → 158, 164-171 → 166-173). `delivery.md` — 템플릿 279~421행 구간이 로컬 +25(예: Step 3.4 제목 330 → 355, 404 → 429). 줄번호가 아니라 `**Flags**:`·`/moai sync [mode]`·`## Synchronization`·`## PR Auto-Merge`·`#### Step 3.4`·`#### Context-Aware Next Steps` 같은 표지로 찾는다.
3. **명령 원본은 로컬과 템플릿의 파일 이름과 형식이 다르다.** 로컬 `.claude/commands/moai/sync.md` 는 렌더링된 영어 사본, 템플릿 `.claude/commands/moai/sync.md.tmpl` 은 2행 `description` 이 로케일 조건문인 Go 템플릿이다. X2가 고치는 3행 `argument-hint` 에는 템플릿 액션이 없어 두 사본에 글자 그대로 같은 줄을 넣는다. 2행과 본문은 건드리지 않는다.
4. **발행기는 템플릿 트리의 게시본만 쓴다.** `internal/template/commandemit/golden_test.go:28` `templatesDir = "../templates"`, `:67-74` 갱신 분기가 템플릿 게시본 경로에 쓴다. 로컬 게시본 `.agents/skills/moai-sync/SKILL.md` 는 추적 파일이지만 발행기가 쓰지 않는다 — 게시본이 바뀌면(AC-GDP-030 경우 B) 로컬 사본은 템플릿 게시본을 바이트 그대로 복사해 맞춘다.
5. **`argument-hint` 는 게시본에 실리지 않는다.** `loader.go:4-6`("Claude-only keys and are NOT carried into the published skill"), `emit.go:122-130` `renderSkill` 은 생성 머리말·`name`·`description`·본문만 쓴다. 그래서 경우 (A)(변화 없음)가 예상이지만, 예상은 판정이 아니다 — 두 경우를 명령으로 가른다.
6. **`TestCommandSourcesUnmodified`(`golden_test.go:94-106`)는 한 실행 안에서 발행 전후 해시를 비교한다.** 원본 편집 자체로는 실패하지 않는다. 원본을 고치면 실패하는 검사는 `make commands-emit-check` 쪽이며, 원본 편집이 게시본을 바꾸는 경우에만 빨강이다.
7. **사본 일치 가드가 파일마다 다르다.** `agent-common-protocol.md` 만 `TestSanitizedPairParity` 와 diff, 나머지 여덟 파일은 diff만 가드다(acceptance.md AC-GDP-013 표). 미러 테스트(`TestSanitizedPairParity`·`TestRuleTemplateMirrorDrift`)는 BASE 에서 이미 exit 1 이다 — 범위 밖 `TestRuleTemplateMirrorDrift/spec-workflow.md` 하나가 빨강이다. 그래서 M4·M6 의 미러 테스트 실행은 exit 0 이 아니라 기준선 대비 집합 비교(새 FAIL 없음, 잃은 PASS 없음)로 판정한다(AC-GDP-013 (e)).
8. **`agent-common-protocol.md` 는 항상 로드되는 규칙이고, 이 카드는 편집하지 않는다(0.2.5).** Pre-Spawn 절은 develop 카드 t635 의 Lane A/B 형태로 REQ-GDP-002 를 이미 충족하고, 템플릿 미러는 카드 dr0911 이 했다(두 사본 `cmp` exit 0). 이 카드는 이 파일을 보존 판정(AC-GDP-002·003·013·016·025)으로만 다룬다.
9. **`manager-git.md:32`(로컬 34) 의 `gh pr merge --squash --delete-branch` 는 기본값 설명이다.** 지우면 안 된다. 148·166행(로컬 150·168) `--auto-merge` 조건도 남는다.
10. **`manager-git.md` 의 PR Auto-Merge 절은 지금 team 한정이다.** 절 제목은 "## PR Auto-Merge" 로 시작하게 유지하고(수용 기준 추출 표지), personal·manual 규칙을 이 절에 모드 이름을 담은 문장으로 더한다. `delivery.md` Step 3.4에도 같은 두 조건을 적는다.
11. **생성물 `.toml` 과 게시본은 손으로 고치지 않는다.** `.toml` 은 `make agents-emit`, 템플릿 게시본은 `make commands-emit` 으로만 다시 만든다(§B 4의 로컬 게시본 복사는 예외).
12. **`manager-git.md:114`(로컬 116) 는 t658이 다시 쓸 Late-Branch 절 안에 있다.** 이 SPEC은 그 줄의 병합 예시만 치환한다.
13. **`agent-common-protocol.md` 17행은 `[ZONE:Frozen]` 이고 13·17·52행에 등록 Frozen clause 가 있다(두 사본 같은 줄).** 이 카드는 이 파일을 편집하지 않으므로 건드릴 일이 없다(REQ-GDP-024). 나머지 여덟 파일(명령 원본 포함)에는 `[ZONE:]` 태그와 레지스트리 항목이 없다(BASE 재측정).
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

1. 트리 확인과 범위의 왼쪽 끝(0.2.6):
   - `git rev-parse --show-toplevel`, `git branch --show-current`, `git rev-parse HEAD`, `git status --porcelain`.
   - `git merge-base --all develop HEAD > .moai/reports/t622/run/card-base.txt` — 정확히 1줄이어야 한다(2줄 이상이면 범위 판정 불가로 보고). 이 값이 `$CARD_BASE` 다. SPEC 에 핀하지 않고, 판정할 때마다 다시 구해 명령에 글자 그대로 넣는다(워크트리 가드가 git 명령 안의 `$(…)` 를 거부한다).
   - 범위 대조: `git diff --name-only develop...HEAD > .moai/reports/t622/run/card-range-names.txt` → `test -e` exit 0, `test -s` exit 0(1줄 이상). 0줄이면 "측정 불가" 로 보고하고 멈춘다 — "변경 없음" 이 아니다.
   - 스냅숏 신선도 점검: `git diff --name-only $BASE $CARD_BASE -- <스냅숏 경로 21개: 범위 파일 로컬·템플릿 18, 게시본 로컬·템플릿 2, 생성물 manager-git.toml 1>` → `test -e` exit 0, `test -s` exit 1. 비어 있지 않으면 흡수가 그 파일을 바꾼 것이다 — 그 파일의 기준 사본은 2단계에서 `$CARD_BASE` 로 반출하고, 그 파일이 쓰는 기준선을 다시 재어 기록한 뒤 리드에 올린다(spec.md §E.2). 미러 기준선용 점검 `git diff --name-only $BASE $CARD_BASE -- internal/template/ .claude/rules/moai/` 도 함께 기록한다(비어 있지 않으면 8단계와 AC-GDP-013 (e)의 귀속 규칙을 쓴다).
   - 0.2.5 의 "BASE 이후 develop 을 다시 흡수했다면 멈추고 SPEC 재고정을 요청한다" 는 폐기했다. 재흡수는 통합 창의 예정된 단계이고, BASE 를 흡수 병합으로 옮기지 않는다(옮기면 카드 커밋이 범위에서 빠진다).
2. 기준 트리 사본 반출: `git show $BASE:<경로> > .moai/reports/t622/run/base-<이름>`(1단계 신선도 점검에서 낡은 것으로 나온 파일만 `$CARD_BASE` 에서 반출). 명령마다 경로를 글자 그대로 쓴다.
   - 로컬 경로 → `base-<이름>.md`: `manager-git`(`.claude/agents/moai/manager-git.md`), `agent-common-protocol`, `delivery`, `doc-execution`, `skill`(`.claude/skills/moai/SKILL.md`), `reference`, `qgc`(`quality-gates-context.md`), `sync`(`.claude/skills/moai/workflows/sync.md`), `command-sync`(`.claude/commands/moai/sync.md`), `published-sync`(`.agents/skills/moai-sync/SKILL.md`).
   - 템플릿 경로 → `base-<이름>-template.md`: `manager-git`(`internal/template/templates/.claude/agents/moai/manager-git.md`), `delivery`, `doc-execution`, `skill`, `reference`, `qgc`, `sync`, `command-sync`(`internal/template/templates/.claude/commands/moai/sync.md.tmpl`), `published-sync`(`internal/template/templates/.agents/skills/moai-sync/SKILL.md`). `manager-git`·`qgc` 는 0.2.5에서 더했다 — BASE 에서 두 사본이 달라져 AC-GDP-013 본문 비교 대상이 됐다. `agent-common-protocol`(`internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`)은 0.2.6에서 더했다 — AC-GDP-002 (b)의 템플릿 기준 절이다.
   - 생성물 `base-manager-git.toml`.
   - AC-GDP-002 뮤턴트 (i)·AC-GDP-003 대조용 옛 사본: `git show b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0:.claude/rules/moai/core/agent-common-protocol.md > .moai/reports/t622/run/b412-agent-common-protocol.md`.
3. 양성 대조(RED 셀) 측정 — `acceptance.md` 각 기준의 "대조" 명령을 기준 트리 사본에 실행한다. 기대 적중이 안 나오면 편집 전에 멈추고 blocker로 보고한다.
4. 생성물 점검 기준선: `make agents-emit-check` exit 0, `make commands-emit-check` exit 0.
5. AC-GDP-030 발행 점검 양성 대조(c): 템플릿 게시본 한 파일을 잠시 바꾼 상태에서 `make commands-emit-check` exit 1 → 백업으로 되돌림 → `cmp` exit 0 → `make commands-emit-check` exit 0 → 게시본 경로 `git status --porcelain` 빈 출력. 되돌림을 확인하지 못하면 멈추고 보고한다.
6. 사본 diff 기준선(BASE): `agent-common-protocol.md`·게시본 exit 0; 나머지 여덟 파일(`manager-git.md`, `delivery.md`, `doc-execution.md`, `quality-gates-context.md`, `moai/SKILL.md`, `references/reference.md`, `workflows/sync.md`, 명령 원본) exit 1. 덩어리 머리를 spec.md §A.4 목록과 대조해 기록한다(0.2.5 측정 `.moai/reports/t622/reanchor/pair-hunks.txt`).
7. Frozen 기준선: AC-GDP-025의 네 clause 개수, `[ZONE:Frozen]` 줄 위치, 레지스트리 양성 대조 13·기타 0, 기타 검출식 뮤턴트 픽스처 2.
8. 미러 테스트 기준선: `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v > .moai/reports/t622/run/mirror-pre.txt 2>&1` → exit 1 기대. PASS/FAIL 집합을 뽑아 `.moai/reports/t622/reanchor/mirror-baseline-sets.txt`(0.2.5 측정: PASS 17줄, FAIL {`TestRuleTemplateMirrorDrift`, `TestRuleTemplateMirrorDrift/spec-workflow.md`})와 같은지 본다. **비교 전에 양쪽을 `sort` 한다** — `sort <집합 파일> > <정렬본>` 두 번 뒤 두 정렬본을 `diff`. 하위 테스트가 병렬로 돌아 출력 순서가 매번 달라서, 정렬하지 않은 `diff` 는 같은 집합에서도 exit 1 이다(재고정 감사 D5, 재현 `.moai/reports/t622/reanchor-fix/d5-*.sorted`: 정렬 전 exit 1, 정렬 뒤 exit 0). AC-GDP-013 (e)의 양방향 `grep -v -x -F -f` 판정은 순서와 무관하므로 그대로 쓴다. 다르면 편집 전에 멈추고 blocker 로 보고한다 — 단 1단계 미러 기준선용 점검이 비어 있지 않았다면(흡수가 미러 대상 파일을 바꿈) 달라진 원소를 AC-GDP-013 (e)의 귀속 규칙으로 가려 기록한다.

**통합 창 흡수 뒤 재측정 (0.2.6).** 판정을 결정하는 검증은 통합 창에서 로컬 develop 을 흡수한 **뒤** 병합 트리에서 한다. 그때 1단계를 다시 한다 — `$CARD_BASE` 는 방금 흡수한 develop 커밋으로 바뀌고, 범위 판정(AC-GDP-014·015·016·025·030, M5)은 그 값에서 다시 잰다. BASE 는 옮기지 않는다. 스냅숏 신선도 점검이 비어 있지 않으면 해당 스냅숏만 다시 반출·재측정하고 리드에 올린다. 흡수를 여러 번 해도 같은 절차를 되풀이한다.

## §D 제약

- **줄 단위 편집.** 로컬과 템플릿 사본을 각각 같은 줄에서 고친다. 파일 통째 복사 금지(명령 원본은 형식이 달라 통째 복사하면 템플릿 조건문이 사라진다).
- **생성물 두 가지.**
  - 에이전트: 템플릿 `manager-git.md` 편집을 모두 마친 뒤 `make agents-emit`, 재생성된 `.toml` 을 같은 카드에 커밋, `make agents-emit-check` exit 0.
  - 명령 게시본: 템플릿 명령 원본 편집 직후 `make commands-emit` → `make commands-emit-check` exit 0. 바뀐 게시본이 있으면 원본 편집과 같은 커밋에 넣는다.
  - 두 발행기는 입력과 출력이 겹치지 않는다(명령 원본 → `.agents/skills/`, `manager-git.md` → `.codex/agents/`). 순서는 같은 커밋 의무에 따라 정한다: `commands-emit` 은 M1의 명령 원본 커밋에, `agents-emit` 은 M4에 둔다.
- **템플릿 중립성.** SPEC ID, REQ 토큰, 내부 날짜, 커밋 SHA, `CLAUDE.local` 참조, 특정 프로그래밍 언어 편향 금지.
- **`agent-common-protocol.md` 비편집.** 두 사본 어느 쪽도 이 카드의 커밋에서 바뀌지 않는다(AC-GDP-016). 미러도 하지 않는다 — 카드 dr0911 이 했다.
- **Frozen 비접촉.** `[ZONE:Frozen]` 줄과 등록 Frozen clause 는 건드리지 않는다(REQ-GDP-024).
- **건드리지 않는 것.** `agent-common-protocol.md`(로컬·템플릿), `spec-workflow.md`, `spec-assembly.md`, `zone-registry.md`, `worktree-integration.md`, `manager-git.md` 의 114행(로컬 116) 밖 Late-Branch 절 줄과 42·160·171행(로컬 44·162·173), `delivery.md` 356행(로컬 381)과 Step 3.2·3.3.5, develop 이 로컬 사본에만 넣은 차이 덩어리(spec.md §A.4), 명령 원본 2행 `description` 과 본문, 발행기 코드, docs-site(X4 — 후속 문서 카드), 설정 키 `workflow.worktree.auto_merge`, Go 코드.
- **검증 부하.** 로컬에서 `go test ./...` 금지. 영향 패키지(`./internal/template/`)의 선택 테스트(AC-GDP-013 의 두 테스트, 미러 테스트 비회귀 점검)와 두 `make` 점검 타깃만 돌린다.
- **git 인덱스.** 명시 경로로만 스테이징한다.

## §E 자기 검증 산출물

- E1: AC-GDP-001~006, 013~016, 025~030 PASS/FAIL 표(자리표시 기준은 N/A). 행마다 명령, 출력 파일 경로, exit code.
- E2: 양성 대조 결과(기준 트리에서 기대 적중이 나왔다는 기록).
- E3: 에이전트 생성물 — `make agents-emit-check` RED(재생성 전)와 GREEN(재생성 뒤) 출력, `.toml` 두 절 diff.
- E4: `go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1` 출력, 최상위 PASS 줄 수(정확히 2), 범위 파일 하위 테스트 흔적. 그리고 미러 테스트 비회귀 점검(M4·M6) 출력과 집합 비교 결과(`mirror-new-fail.txt`·`mirror-lost-pass.txt` 모두 빈 파일).
- E5: 사본 diff 결과(두 파일 exit 0, 여덟 파일 차이 본문이 BASE 와 동일, 명령 원본 `argument-hint` 줄 일치).
- E6: 커밋 SHA 목록과, 그중 `agent-common-protocol.md` 두 사본을 바꾼 커밋이 없다는 확인(AC-GDP-016 판정 출력과 양성 대조 출력).
- E7: 읽기 기록 `ac001-reading.md`·`ac006-reading.md`·`ac026-reading.md`·`ac028-reading.md`(PASS 전제), AC-GDP-015 읽기 목록.
- E8: AC-GDP-025 Frozen 확인 결과.
- E9: AC-GDP-026~029 아홉 조각별 판정 파일, `--merge` 존재 개수(`ac026-merge-presence.txt`), `--no-merge` 줄 수.
- E10: 명령 게시본 — AC-GDP-030 발행 대조(c)의 네 출력, `commands-emit`·`commands-emit-check` 출력, `ac030-outcome.md`(경우 이름과 근거 파일 내용).

## §F 마일스톤

### M1 — OD-2 = B와 하위 결정: 옵트인 단일 기준, 플래그 의미, 모드별 승인 조건, 소비자 X1~X3

- 대상: REQ-GDP-006, 025, 026, 그리고 X2의 발행 의무로 REQ-GDP-014
- `delivery.md` L·T Step 3.4 (템플릿 330~, 로컬 355~):
  - 335-338(트리거, 로컬 360-363) — 워크트리 문맥 기본 병합과 `--no-merge` 조건을 없애고, 트리거를 "`--auto-merge`(또는 폐기된 별칭 `--merge`)가 주어짐" 으로 바꾼다. 병합 기준이 `manager-git.md` 옵트인임을 이름으로 밝힌다.
  - 모드 조건 — "team mode" 는 `--auto-merge` 와 전원 승인, "personal and manual modes" 는 `--auto-merge` 로 승인 조건 없이 병합(승인할 팀원 없음)을 각각 한 줄로 적는다. CI 통과·충돌 없음 확인은 그대로 둔다.
  - 348-349(플래그 설명, 로컬 373-374) — `--auto-merge` 설명을 더하고, `--merge` 는 `--auto-merge` 의 폐기된 별칭(경고), `--no-merge` 는 폐기된 호환용 no-op(경고, 병합하지 않는 것이 이미 기본)으로 적는다.
- `delivery.md` L·T Context-Aware Next Steps 절 404행(로컬 429, X3) — 선택지를 `--auto-merge` 로 바꾼다. `--merge` 를 남기면 같은 줄에 폐기 표시와 `--auto-merge` 가 함께 있어야 한다.
- `doc-execution.md` L·T 34-36 — "worktree contexts default to auto-merge" 문장을 없애고, 병합 여부는 `manager-git.md` 옵트인(`--auto-merge`)이 정한다고 이름으로 밝힌다. 138-143행 사본 차이는 건드리지 않는다.
- `manager-git.md` L·T 164-171(로컬 166-173) PR Auto-Merge 절 — 제목은 "## PR Auto-Merge" 로 시작하게 두고, team 모드 조건 줄과 personal·manual 조건 줄을 적는다. 148행(로컬 150) Team Mode 항목의 `--auto-merge` 조건은 유지하고, 필요하면 133-140(로컬 135-142) Personal Mode 절에서 이 절을 가리킨다. 171행(로컬 173) 5단계는 건드리지 않는다(t658). 로컬 프론트매터 설명(`5,7c5` 덩어리)은 건드리지 않는다.
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
- `delivery.md` L·T 343·355(로컬 368·380) — `gh pr merge --<merge_method> --delete-branch` 로 바꾸고 해석 출처(`git_strategy.<mode>.merge_method`, 기본 `squash`)를 한 번 명시한다.
- `manager-git.md` L·T 114(로컬 116) — 예시를 `gh pr merge <PR> --<merge_method> --delete-branch` 로 바꾼다. 32행(로컬 34) 기본값 설명은 유지한다.

### M3 — AC-11: `manager-git.md` 동기화 절 순서

- 대상: REQ-GDP-001
- `manager-git.md` L·T 156(로컬 158) — `git fetch` 를 먼저 끝내고, 그 결과를 읽는 `git rev-list` 는 그 뒤에 실행한다고 고친다. fetch 결과를 읽지 않는 명령은 병렬로 둘 수 있다. fetch 와 rev-list 를 같은 목록·표에 두지 않는다.

### M4 — 생성물·사본·중립성·Frozen 확인

- 대상: REQ-GDP-013, 014, 015, 024
1. 템플릿 `manager-git.md` 편집(M1·M2·M3)이 끝난 시점에 `make agents-emit-check` 를 먼저 돌려 exit 1 을 관측한다.
2. `make agents-emit` → `make agents-emit-check` exit 0, `.toml` 두 절 diff exit 0, 재생성된 `.toml` 커밋.
3. `make commands-emit-check` 를 다시 돌려 exit 0 을 확인한다(읽기 전용 — M1 명령 원본 커밋 뒤 게시본 드리프트가 없다는 기록).
4. 사본 diff(두 파일 exit 0, 여덟 파일 차이 본문이 BASE 와 동일, 명령 원본 `argument-hint` 줄 일치).
5. 템플릿 diff 추가 줄 중립성 검사(AC-GDP-015).
6. Frozen 확인(AC-GDP-025).
7. 미러 테스트 비회귀 점검(AC-GDP-013 (e)): `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v > .moai/reports/t622/run/mirror-m4.txt 2>&1`. exit 는 기록만 한다(기준선이 exit 1). PASS/FAIL 집합을 뽑아 기준선과 비교한다 — 새 FAIL 이 없고(모든 FAIL 이 기준선 FAIL 집합 안) 기준선 PASS 가 하나도 빠지거나 FAIL 로 바뀌지 않아야 한다. 개수가 아니라 집합으로 비교한다.

### M5 — AC-11: Pre-Spawn Sync Check 보존 확인 (편집 없음)

- 대상: REQ-GDP-002, 003
- 0.2.5 리드 판단 (a): Pre-Spawn 절은 develop 카드 t635 의 Lane A/B 형태로 REQ-GDP-002 를 이미 충족하고, 템플릿 미러는 카드 dr0911 이 했다. **이 단계는 아무 파일도 고치지 않는다.**
- AC-GDP-002(회귀 방지 — (a) 두 사본 Pre-Spawn bash 블록에서 fetch 가 끝난 뒤 rev-list 가 시작되는 순서, (b) 두 사본 Pre-Spawn 절 전체가 고정 기준 절과 같음)와 AC-GDP-003(Pre-Edit 절이 기준 절과 같음)을 실행해 PASS 를 기록한다. 기준 절은 §C 1단계 신선도 점검이 빈 결과일 때 BASE 반출본이다.
- 두 판정이 빨강이면 이 카드의 편집이 원인인지(`git log --no-merges --format=%H HEAD --not develop -- .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` 목록 — 이 카드의 비병합 커밋만 담는다) 흡수가 원인인지(§C 1단계 신선도 점검에 `agent-common-protocol.md` 가 나옴) 가려 blocker 로 보고한다 — 이 카드가 그 파일을 고쳐 맞추지 않는다.

### M6 — 최종 검증 (편집 없음)

- 대상: 판정 대상 전체
1. 사본 diff와 AC-GDP-025를 다시 실행한다.
2. `make agents-emit-check`·`make commands-emit-check` 를 다시 실행해 둘 다 exit 0 을 확인한다(읽기 전용).
3. 선택 테스트 두 개와 최상위 PASS 줄 2개, 범위 파일 하위 테스트 흔적(AC-GDP-013 (d)).
4. 미러 테스트 비회귀 점검을 M4 7단계와 같은 방식으로 다시 실행한다(`mirror-m6.txt`, AC-GDP-013 (e)).
5. 읽기 기록 네 개 확인(`ac001`·`ac006`·`ac026`·`ac028`, acceptance.md §D.3).
6. 이 카드 커밋 가운데 `agent-common-protocol.md` 두 사본을 바꾼 커밋이 없는지 확인(AC-GDP-016, 양성 대조 포함).

## §G 안티패턴

- 파일 통째 복사로 사본을 맞추는 것(의도된 차이 소실 — 명령 원본은 템플릿 조건문까지 사라진다).
- 범위 파일을 덮지 않는 미러 테스트를 돌려 사본 일치 증거로 쓰는 것(M4·M6 의 미러 테스트 실행은 비회귀 점검이지 범위 파일 사본 일치 증거가 아니다). 미러 테스트 exit 1 을 이 카드의 실패로 읽거나, PASS·FAIL 개수만 세어 기준선과 같다고 적는 것.
- `--merge` 의 폐기를 취소하거나 `--auto-merge` 와 별개 플래그로 적는 것, `--no-merge` 에 건너뜀·조건 의미를 남기는 것.
- team 모드 승인 조건을 지우거나, personal·manual 모드에 승인 조건을 붙이는 것, 두 문서 중 한쪽에만 모드 조건을 적는 것.
- 설정 키 `workflow.worktree.auto_merge` 를 기준으로 되살리는 것.
- docs-site를 이 카드에서 고치는 것(X4는 후속 문서 카드 소관).
- 명령 원본만 커밋하고 게시본 발행을 다른 커밋으로 미루는 것, `make commands-emit` 을 실행하지 않고 경우 (A)라고 적는 것.
- 템플릿 게시본만 바꾸고 로컬 게시본 사본을 두고 오는 것, 게시본을 손으로 고치는 것.
- `manager-git.md:32` 기본값 설명이나 148·166행 `--auto-merge` 조건을 지우는 것.
- `manager-git.md:114` 를 고치면서 주변 late-branch 절차 줄까지 손대는 것(t658 소관).
- `.toml` 을 손으로 맞추는 것.
- `agent-common-protocol.md` 를 고치거나 미러하는 것(REQ-GDP-002 는 t635·dr0911 로 충족, 0.2.5), 17행 `[ZONE:Frozen]` 이나 등록 clause 를 건드리는 것.
- 파일 통째 복사나 되돌리기로 develop 이 로컬 사본에만 넣은 차이 덩어리를 지우는 것.
- 읽기 기록 없이 AC-GDP-001·006·026·028·029를 PASS로 적는 것.
- "이 카드가 바꾼 것" 을 리터럴 `$BASE..HEAD`·`git diff $BASE` 로 재거나, develop 흡수 뒤 BASE 를 흡수 병합으로 옮기는 것(앞은 흡수된 develop 커밋을 카드 것으로 읽고, 뒤는 카드 커밋을 범위에서 뺀다 — spec.md §E.2). 범위 대조가 0줄인데 "변경 없음" 으로 적는 것, 병합 뒤 빈 범위를 통과로 읽는 것.
- 검증 출력을 `| head`·`| tail`·`| grep` 로 잘라 exit code를 잃는 것, 개수가 찍히지 않은 grep 결과를 0으로 읽는 것, 빈 diff·빈 커밋 목록을 통과로 읽는 것.

## §H 교차 참조

- `spec.md` §A(측정 근거·Frozen 확인·소비자 조사), §C(결정 기록·번호 방식·티어), §D(제외 범위·X4 후속 카드 입력), §E(미검증·잔여 위험), §G(t658로 옮긴 항목)
- `acceptance.md` (판정 대상 AC-GDP-001~006, 013~016, 025~030)
- `.moai/reports/t622/repro.md`, `.moai/reports/t622/plan-audit.md`, `.moai/reports/t622/plan-audit-iter2.md`, `.moai/reports/t622/plan-audit-reduced-iter1.md`
- `.moai/reports/t622/absorb-ee99507fb.md`(두 번의 develop 흡수 기록), `.moai/reports/t622/absorb2-mirror-baseline.txt`(미러 테스트 기준선 원본), `.moai/reports/t622/reanchor/`(0.2.5 재고정 측정·뮤턴트, `index.md`)
- `internal/hook/wrapper_copies_contract_test.go:73` `TestHookWrapperCopiesStayIdentical` — 훅 래퍼 스크립트 여섯 개만 읽어 이 카드 파일과 닿지 않으므로 비회귀 점검에 넣지 않는다(spec.md §A.4)
- `internal/template/rule_template_mirror_test.go`, `sanitized_pair_parity_test.go`, `internal_content_leak_test.go`, `backlog_json_disclosure_mirror_test.go`, `template_neutrality_audit_test.go`, `agentless_audit_test.go`, `agent_frontmatter_audit_test.go`, `published_skills_deploy_test.go`
- `internal/template/commandemit/` — `commandemit.go:49-53`(`DefaultOptions` 루트), `loader.go:4-6`(발행에 실리지 않는 키), `emit.go:122-130`(`renderSkill`), `golden_test.go:28`(`templatesDir`)·`:57-74`(갱신 분기)·`:94-106`(`TestCommandSourcesUnmodified`)
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R2-006`·`036`·`037`·`038`
- `Makefile:38-49` — `agents-emit`, `agents-emit-check`; `Makefile:51-60` — `commands-emit`, `commands-emit-check`
- SPEC-MERGE-METHOD-CONFIG-001
