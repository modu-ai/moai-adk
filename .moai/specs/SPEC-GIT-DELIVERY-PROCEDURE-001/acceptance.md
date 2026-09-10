# Acceptance — SPEC-GIT-DELIVERY-PROCEDURE-001

> 기준마다 Given-When-Then과 판정 명령을 둔다. 모든 부재 기준은 같은 검출기로 알려진 적중을 먼저 잡는 **양성 대조**를 가진다. 검증 출력은 파일로 보내고 exit code를 파이프 없이 따로 읽는다(`| head`·`| tail`·`| grep` 금지).

## 공통 변수

```bash
BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0   # R1 고정: 2026-09-10 이 워크트리에서 해석
E=.moai/reports/t622/run                         # 추적 증거 경로
T=internal/template/templates                    # 템플릿 루트
```

명령은 워크트리 루트에서 실행한다. "exit" 는 직전 명령의 exit code를 `echo "exit=$?"` 로 따로 기록한 값이다. 기준 트리 사본은 사전 점검에서 `git show $BASE:<경로> > $E/base-<이름>` 으로 반출해 둔다.

## §D AC 표

| AC ID | REQ | 등급 | 결정 의존 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 없음 | `manager-git.md` 동기화 절이 `git fetch` 를 독립 항목으로 분류하지 않고 순서를 지시 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 없음 | Pre-Spawn Sync Check가 fetch→rev-list 순서를 한 명령으로 보장, 세 번째 명령과 해석 표 유지 |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 없음 | Pre-Edit Sync Check 블록 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 없음 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 없음 | `gh pr merge <PR> --squash` 예시 부재, 32행 기본값 설명 유지 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | **DEFERRED: OD-2** | auto-merge 기본값 단일 기준 |
| AC-GDP-007 | REQ-GDP-007 | MUST-PASS | 분류 장부는 없음, "조건부" 분류는 OD-1 선택지 2 | 남은 명령 적중이 모두 분류되고 무조건 실행 안내가 0 |
| AC-GDP-008 | REQ-GDP-008 | MUST-PASS | 없음 | 절차 참조가 끊기지 않음 |
| AC-GDP-009 | REQ-GDP-009 | MUST-PASS | **DEFERRED: OD-1** | 네 지침이 같은 절차를 서술 |
| AC-GDP-010 | REQ-GDP-010 | MUST-PASS | **DEFERRED: OD-1 = 1** | 런처 진입 워크트리 흐름, 맨손 `git worktree add` 부재 |
| AC-GDP-011 | REQ-GDP-011 | MUST-PASS | **DEFERRED: OD-1 = 2** | 조건 문구가 금지문과 안내 위치에 같이 실림 |
| AC-GDP-012 | REQ-GDP-012 | MUST-PASS | **DEFERRED: OD-1 = 3** | 은퇴 선언, `main_late_branch` 부재, `--issue` 기본값 유지 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 없음 | 사본 일치와 `delivery.md` 의도된 차이 보존 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 없음 | `.toml` 재생성과 `agents-emit-check` exit 0 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 없음 | 템플릿 추가 문구에 SPEC ID·날짜 없음 |
| AC-GDP-016 | plan §D 제약 | SHOULD-PASS | 없음 | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |

## §D.1 Given-When-Then과 판정 명령

### AC-GDP-001 — `git fetch` 가 독립 항목에서 빠지고 순서가 지시됨

```
GIVEN manager-git.md 로컬·템플릿 사본의 "## Synchronization" 절
WHEN run-phase 편집 뒤 절을 추출해 검사하면
THEN 한 줄에 "git fetch" 와 "independent" 가 함께 나오는 줄이 없고
 AND git fetch 가 먼저 끝난다는 순서 문장이 있다
```

대조(기준 트리에서 적중해야 함):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $E/base-manager-git.md > $E/ac001-base-sync.md
grep -n -E 'git fetch.*independent|independent.*git fetch' $E/ac001-base-sync.md > $E/ac001-control.txt
# 기대: exit 0, 1줄 (기준 트리 156행 문장)
```

판정(로컬·템플릿 각각):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' .claude/agents/moai/manager-git.md > $E/ac001-local-sync.md
grep -n -E 'git fetch.*independent|independent.*git fetch' $E/ac001-local-sync.md > $E/ac001-local-absent.txt
# 기대: exit 1
grep -n -E 'git fetch[^.]*(first|before|then|completes)' $E/ac001-local-sync.md > $E/ac001-local-order.txt
# 기대: exit 0, 그리고 적중 줄을 읽어 fetch 가 rev-list 보다 먼저라는 뜻인지 확인 기록
```

템플릿 사본은 경로를 `$T/.claude/agents/moai/manager-git.md` 로 바꿔 같은 두 명령을 실행한다. 생성물 `$T/.codex/agents/moai/manager-git.toml` 에도 부재 명령을 실행한다(대조: `git show $BASE:$T/.codex/agents/moai/manager-git.toml > $E/base-manager-git.toml` 에서 `grep -n -E 'git fetch.*independent'` exit 0, 150행).

### AC-GDP-002 — Pre-Spawn Sync Check의 순서 보장

```
GIVEN agent-common-protocol.md 의 Pre-Spawn Sync Check 절
WHEN 편집 뒤 검사하면
THEN fetch 와 rev-list 를 ";" 또는 "&&" 로 이은 줄이 파일 전체에서 2개(Pre-Spawn, Pre-Edit)이고
 AND Pre-Spawn 절에 fetch 가 단독으로 선 줄이 없고
 AND 세 번째 명령 줄과 두 해석 표의 행이 기준 트리와 같다
```

대조:

```bash
grep -c -E 'git fetch origin main 2>&1(;| &&) git rev-list --count --left-right origin/main[.][.][.]HEAD' $E/base-agent-common-protocol.md > $E/ac002-control-count.txt
# 기대: 1 (Pre-Edit 347행 — 검출기가 알려진 형태를 잡음)
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' $E/base-agent-common-protocol.md > $E/ac002-base-prespawn.md
grep -n -E '^git fetch origin main 2>&1$' $E/ac002-base-prespawn.md > $E/ac002-control-standalone.txt
# 기대: exit 0 (296행)
```

판정(로컬·템플릿 각각):

```bash
grep -c -E 'git fetch origin main 2>&1(;| &&) git rev-list --count --left-right origin/main[.][.][.]HEAD' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-count.txt
# 기대: 2
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-prespawn.md
grep -n -E '^git fetch origin main 2>&1$' $E/ac002-local-prespawn.md > $E/ac002-local-standalone.txt
# 기대: exit 1
grep -c 'moai session list --json --filter-spec=' $E/ac002-local-prespawn.md > $E/ac002-local-session.txt
# 기대: 1 (기준 트리도 1)
grep -E '^[|] ' $E/ac002-base-prespawn.md > $E/ac002-base-matrix.txt
grep -E '^[|] ' $E/ac002-local-prespawn.md > $E/ac002-local-matrix.txt
diff $E/ac002-base-matrix.txt $E/ac002-local-matrix.txt > $E/ac002-matrix.diff
# 기대: exit 0
```

### AC-GDP-003 — Pre-Edit Sync Check 블록 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조(검출기가 변경을 잡는지): AC-GDP-002의 `ac002-base-prespawn.md` 와 `ac002-local-prespawn.md` 를 `diff` 하면 exit 1.

판정:

```bash
sed -n '/^### Pre-Edit Sync Check/,/^#### The sweep prohibition/p' $E/base-agent-common-protocol.md > $E/ac003-base.md
sed -n '/^### Pre-Edit Sync Check/,/^#### The sweep prohibition/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac003-local.md
diff $E/ac003-base.md $E/ac003-local.md > $E/ac003.diff
# 기대: exit 0 (템플릿 사본도 같은 방식)
```

### AC-GDP-004 — `delivery.md` 병합 명령 해석

```
GIVEN delivery.md 로컬·템플릿 사본의 Step 3.4 Auto-Merge Behavior
WHEN 편집 뒤 검사하면
THEN "gh pr merge --squash --delete-branch" 가 0회이고
 AND "gh pr merge --<merge_method> --delete-branch" 가 2회 이상이며
 AND merge_method 의 해석 출처(git_strategy 모드 설정, 기본 squash)가 적혀 있다
```

대조:

```bash
grep -c 'gh pr merge --squash --delete-branch' $E/base-delivery.md > $E/ac004-control.txt
# 기대: 2 (343·355행)
```

판정(로컬·템플릿 각각):

```bash
grep -c 'gh pr merge --squash --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-squash.txt
# 기대: 0
grep -c 'gh pr merge --<merge_method> --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-resolved.txt
# 기대: 2 이상
grep -n -E 'merge_method.*(squash|default)' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-source.txt
# 기대: exit 0, 적중 줄이 해석 출처를 설명하는지 읽어 기록
```

### AC-GDP-005 — 고정 `--squash` 예시 부재, 기본값 설명 유지

```
GIVEN manager-git.md 로컬·템플릿 사본과 생성물 .toml
WHEN 편집 뒤 검사하면
THEN "gh pr merge <PR> --squash" 가 없고
 AND 32행 기본값 설명 문장은 1회 남아 있다
```

이 기준은 OD-1이 어떤 선택지로 결정되어도 성립해야 한다(블록 유지+치환, 블록 제거 모두 부재).

대조:

```bash
grep -n 'gh pr merge <PR> --squash' $E/base-manager-git.md > $E/ac005-control.txt
# 기대: exit 0 (114행)
```

판정:

```bash
grep -n 'gh pr merge <PR> --squash' .claude/agents/moai/manager-git.md > $E/ac005-local.txt
# 기대: exit 1 (템플릿 사본, .toml 도 같은 명령으로 exit 1)
grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-local-default.txt
# 기대: 1 (기준 트리도 1)
```

### AC-GDP-006 — auto-merge 기본값 단일 기준 **[DEFERRED: OD-2]**

```
GIVEN OD-2 로 정한 기준(A: delivery.md 기본값 / B: manager-git.md 옵트인 / C: workflow.worktree.auto_merge)
WHEN delivery.md Step 3.4 와 manager-git.md Team Mode·PR Auto-Merge 절을 읽으면
THEN 기준이 아닌 쪽에 반대 기본값을 적은 문장이 0개이고
 AND 기준이 아닌 쪽은 기준을 이름으로 참조한다
```

대조: 기준 트리에서 `delivery.md:337`(기본 병합)과 `manager-git.md:148`·`166`(플래그 필수)이 동시에 적중한다(모순 존재 확인). 판정 명령은 결정된 선택지의 기준 문구를 고정한 뒤 확정한다.

### AC-GDP-007 — 명령 적중 분류 장부

```
GIVEN 범위 네 파일(로컬·템플릿)
WHEN 감사와 같은 검출식으로 명령 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND "primary checkout 무조건 실행 안내" 분류가 0행이다
```

대조(기준 트리에서 알려진 적중 수를 잡는지 — 편집 전 재측정):

```bash
git grep -n -E 'git checkout |git switch |git reset --hard' $BASE -- $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-base-hits.txt
# 기대: exit 0, 14줄 — manager-git 8, spec-workflow 3, spec-assembly 1, delivery 2
# (plan 작성 시점 측정: 이 워크트리, $BASE, `git grep -c -E` 같은 검출식. 재현 기록 c352330d3 의 값과 같음)
```

판정:

```bash
grep -n -E 'git checkout |git switch |git reset --hard' $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-post-hits.txt
# exit 0 이면 적중 줄마다 장부 행을 만든다. exit 1 이면 장부는 "적중 0" 한 줄
```

장부 `.moai/reports/t622/ac01-classification.md` 의 분류값은 `prohibition-context` · `worktree-internal` · `conditioned`(OD-1 선택지 2에서만, 조건 문구 인용) · `unconditioned-primary-instruction` 넷 중 하나. 합격 조건은 마지막 분류 0행.

### AC-GDP-008 — 절차 참조가 끊기지 않음

```
GIVEN spec-assembly.md 와 spec-workflow.md 가 참조하는 manager-git.md 절차 제목
WHEN 참조마다 대상 제목의 존재를 확인하면
THEN 끊긴 참조가 0개다
```

대조(검출기가 끊긴 참조를 잡는지):

```bash
grep -v '^### Late-Branch Invocation Pattern' $E/base-manager-git.md > $E/ac008-fixture-manager-git.md
grep -c '^### Late-Branch Invocation Pattern' $E/ac008-fixture-manager-git.md > $E/ac008-control.txt
# 기대: 0 — 기준 트리 참조 2곳(spec-assembly.md:340, spec-workflow.md:62)이 이 픽스처에서는 끊긴 것으로 판정되어야 함
```

판정:

```bash
grep -n -o -E 'manager-git.md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md > $E/ac008-refs.txt
# 제목은 대문자로 시작하는 단어만 잇는다. 소문자까지 받는 문자 클래스는 뒤따르는 "for the" 까지 삼켜
# 대상 제목 조회를 거짓 끊김으로 만든다(작성 시점 측정: spec-assembly.md:340 에서 발생).
# 적중한 제목마다: grep -c '^### <제목>' .claude/agents/moai/manager-git.md > $E/ac008-target-<n>.txt  기대: 1
```

적중이 0이면(참조가 모두 제거된 경우) OD-1 선택지 3의 AC-GDP-012 은퇴 선언 검사로 넘긴다.

### AC-GDP-009 — 네 지침의 절차 일치 **[DEFERRED: OD-1]**

```
GIVEN OD-1 로 정한 절차
WHEN spec-workflow.md Step 1·4, manager-git.md 절차, spec-assembly.md 사전 점검, delivery.md 복귀 단계의 코드 블록 명령을 뽑아 비교하면
THEN 진입 명령과 종결 명령의 집합이 같다
```

판정 명령은 결정 뒤 코드 블록 추출 범위를 고정해 확정한다. 대조는 기준 트리에서 `spec-workflow.md` 56-59행과 `manager-git.md` 119-122행이 같은 네 명령을 담는지 확인하는 것으로 한다.

### AC-GDP-010 — 런처 진입 워크트리 흐름 **[DEFERRED: OD-1 = 선택지 1]**

```
GIVEN 선택지 1로 재설계한 절차
WHEN 범위 파일을 검사하면
THEN "git worktree add" 적중이 0이고
 AND "moai cc -w" 또는 "EnterWorktree(" 가 절차 안에 있고
 AND progress.md 에 spec-workflow.md 49행 Frozen 구역 수정과 Implementation Kickoff Approval 기록이 있다
```

대조: `grep -n 'git worktree add' $T/.claude/rules/moai/workflow/main-checkout-branch-guard.md` exit 0(38행 — 검출기가 형태를 잡음).

### AC-GDP-011 — 조건 문구 공유 **[DEFERRED: OD-1 = 선택지 2]**

```
GIVEN 선택지 2로 정한 조건 문장 COND
WHEN 금지문과 남은 안내 위치를 검사하면
THEN 루트 AGENTS.md, 템플릿 AGENTS.md, main-checkout-branch-guard.md 로컬·템플릿, 남은 각 안내 위치에 COND 가 1회 이상 있다
 AND 두 AGENTS.md 의 편집 전후 바이트 수가 기록되어 있다
```

대조: `grep -c -F 'Never change branch state in the primary checkout' $E/base-AGENTS.md` 기대 1(고정 문자열 검출기가 이 파일에서 동작함).

### AC-GDP-012 — `main_late_branch` 은퇴 **[DEFERRED: OD-1 = 선택지 3]**

```
GIVEN 선택지 3으로 정한 은퇴
WHEN 검사하면
THEN manager-git.md 로컬·템플릿과 .toml 에 "main_late_branch" 가 없고
 AND manager-git.md 에 은퇴 선언 문장이 있고
 AND --issue opt-in 기본값 문구가 기준 트리와 같은 수로 남아 있다
```

대조:

```bash
grep -n 'main_late_branch' $E/base-manager-git.md > $E/ac012-control.txt
# 기대: exit 0 (137행)
grep -c 'default skips GitHub Issue creation' $E/base-SKILL.md > $E/ac012-issue-base.txt
# 기대: 기준 값 기록 (SKILL.md 124·227행이 해당)
```

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 다섯 파일의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md, spec-workflow.md, spec-assembly.md, agent-common-protocol.md 는 diff exit 0 이고
 AND delivery.md 는 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 미러 테스트가 PASS 줄을 1개 이상 내며 exit 0 이다
```

대조: 기준 트리에서 `delivery.md` 두 사본의 `diff` 는 exit 1(검출기가 차이를 보고함). 본문 비교 검출기는 빈 파일(`manager-git.md` 차이 본문)과 `delivery.md` 차이 본문을 `diff` 하면 exit 1.

판정:

```bash
diff .claude/agents/moai/manager-git.md $T/.claude/agents/moai/manager-git.md > $E/ac013-manager-git.diff
# 기대: exit 0 (spec-workflow.md, spec-assembly.md, agent-common-protocol.md 도 같은 형태)
diff $E/base-delivery.md $E/base-delivery-template.md > $E/ac013-delivery-base.diff
diff .claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac013-delivery-post.diff
grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-base.diff > $E/ac013-delivery-base.body
grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-post.diff > $E/ac013-delivery-post.body
diff $E/ac013-delivery-base.body $E/ac013-delivery-post.body > $E/ac013-delivery-body.diff
# 기대: exit 0
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0
grep -c -e '--- PASS' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 1 이상 (0이면 선택된 테스트가 없어 판정 불가 — FAIL 로 기록)
```

### AC-GDP-014 — 생성물 재생성

```
GIVEN 템플릿 manager-git.md 편집이 끝났고 .toml 은 아직 재생성하지 않은 상태
WHEN agents-emit-check → agents-emit → agents-emit-check 순서로 실행하면
THEN 첫 점검은 exit 1, 재생성은 exit 0, 두 번째 점검은 exit 0 이고
 AND 기준 트리 대비 .codex/agents/moai/ 아래 바뀐 파일은 manager-git.toml 뿐이다
```

```bash
make agents-emit-check > $E/ac014-red.txt 2>&1
# 기대: exit 1 (재생성 전 RED — 검출기 양성 대조)
make agents-emit > $E/ac014-emit.txt 2>&1
# 기대: exit 0
make agents-emit-check > $E/ac014-green.txt 2>&1
# 기대: exit 0
git diff --name-only $BASE -- $T/.codex/agents/moai/ > $E/ac014-changed.txt
# 기대: 한 줄, internal/template/templates/.codex/agents/moai/manager-git.toml
```

OD-2 선택지 B처럼 `manager-git.md` 템플릿이 전혀 바뀌지 않는 조합에서는 첫 점검 RED가 나오지 않으므로, 그때는 `ac014-changed.txt` 가 비어 있고 두 번째 점검 exit 0 인 것으로 합격한다.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분
WHEN 추가 줄을 검사하면
THEN SPEC ID 형태와 날짜 형태가 추가 줄에 없다
```

대조(검출기가 형태를 잡는지):

```bash
grep -c -E 'SPEC-[A-Z][A-Z0-9]*-[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-specid.txt
# 기대: 1 이상
grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-date.txt
# 기대: 1 이상
```

판정:

```bash
git diff $BASE -- $T/ > $E/ac015-template.diff
grep -n -E '^[+].*SPEC-[A-Z][A-Z0-9]*-[0-9]{3}' $E/ac015-template.diff > $E/ac015-specid.txt
# 기대: exit 1
grep -n -E '^[+].*20[0-9]{2}-[0-9]{2}-[0-9]{2}' $E/ac015-template.diff > $E/ac015-date.txt
# 기대: exit 1
grep -n -E '^[+].*CLAUDE[.]local' $E/ac015-template.diff > $E/ac015-local-ref.txt
# 기대: exit 1
```

`go test` 의 `TestTemplateNoInternalContentLeak` PASS는 AC-GDP-013에서 함께 확인한다.

### AC-GDP-016 — 항상 로드 규칙 편집이 마지막 (SHOULD)

```
GIVEN run-phase 커밋들
WHEN agent-common-protocol.md 를 고친 커밋 이후의 커밋을 범위 지침 파일로 거르면
THEN 결과가 비어 있다
```

대조(검출기가 범위 커밋을 보는지):

```bash
git log --format=%H $BASE..HEAD -- .claude/agents/moai/manager-git.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac016-control.txt
# 기대: 1줄 이상
```

판정:

```bash
git log --format=%H -1 $BASE..HEAD -- .claude/rules/moai/core/agent-common-protocol.md > $E/ac016-acp-commit.txt
git log --format=%H <ac016-acp-commit.txt 의 SHA>..HEAD -- .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac016-after.txt
# 기대: 빈 파일
```

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. 한 줄에 두 번 나오는 명령은 1로 센다 — AC-GDP-004의 "2 이상" 은 343·355행이 서로 다른 줄이라는 사실에 기댄다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. 추출 파일의 마지막 줄이 기대한 끝 제목인지 확인해 기록한다.
- 결정 대기 기준(AC-GDP-006, 009~012)은 결정 전에 FAIL로 기록하지 않고 DEFERRED로 기록한다.
- `delivery.md` 편집으로 줄 수가 바뀌면 479행 꼬리말의 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다).

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다.
- `go test` 선택 실행이 PASS 줄 0개이면 합격이 아니다.
- 로컬 전체 스위트는 돌리지 않는다. 전체 판정은 통합 뒤 CI가 한다.

## §D.4 완료 정의 (Definition of Done)

- 결정 불필요 기준 AC-GDP-001~005, 007, 008, 013~015가 PASS이고 AC-GDP-016이 PASS 또는 사유 기록.
- 결정된 선택지에 해당하는 기준만 PASS, 나머지 선택지 기준은 N/A로 기록.
- OD-1·OD-2 중 미결정이 남으면 해당 기준은 DEFERRED이고, SPEC은 `implemented` 로 넘어가지 않는다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
