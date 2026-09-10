# Acceptance — SPEC-GIT-DELIVERY-PROCEDURE-001

> 기준마다 Given-When-Then과 판정 명령을 둔다. 모든 부재 기준은 같은 검출기로 알려진 적중을 먼저 잡는 **양성 대조**를 가진다. 검증 출력은 파일로 보내고 exit code를 파이프 없이 따로 읽는다(`| head`·`| tail`·`| grep` 금지).

## 공통 변수와 명령 관례

```bash
BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0   # R1 고정: 2026-09-10 이 워크트리에서 해석
E=.moai/reports/t622/run                         # 추적 증거 경로
T=internal/template/templates                    # 템플릿 루트
```

- 명령은 워크트리 루트에서 실행한다. "exit" 는 직전 명령의 exit code를 `echo "exit=$?"` 로 따로 기록한 값이다.
- 기준 트리 사본은 사전 점검에서 `git show $BASE:<경로> > $E/base-<이름>` 으로 반출해 둔다(대상 목록은 plan.md §C 2단계).
- 검출식 안의 `git` 은 `[g]it` 으로 쓴다. 뜻은 같고, 워크트리 세션 가드가 복합 명령 안의 `git` 낱말을 git 실행으로 보고 거부하는 일을 피한다. 같은 이유로 `parallel` 은 `para[l]lel` 로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 이 환경의 셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛰어, 개수가 찍히지 않은 결과가 0처럼 읽힐 수 있다.

## §D AC 표

| AC ID | REQ | 등급 | 결정 의존 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 없음 | `manager-git.md` 동기화 절이 `git fetch` 를 독립·병렬·배치 항목으로 묶지 않고 순서를 지시 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 없음 | Pre-Spawn 코드 블록 안에 단독 fetch·단독 rev-list 줄이 없고 이어 붙인 줄이 정확히 하나, 세 번째 명령과 해석 표 유지 |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 없음 | Pre-Edit Sync Check 절 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 없음 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 없음 | 범위 파일 전체에서 `--squash` 고정 `gh pr merge` 가 기본값 설명 문장뿐 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | **DEFERRED: OD-2** | auto-merge 기본값 단일 기준 |
| AC-GDP-007 | REQ-GDP-007 | MUST-PASS | 분류 장부는 없음, "조건부" 분류는 OD-1 선택지 2 | 금지 집합 전체 검출식의 적중이 모두 분류되고 무조건 실행 안내가 0 |
| AC-GDP-008 | REQ-GDP-008 | MUST-PASS | 참조 0건 경로만 OD-1 = 3 | 절차 참조가 끊기지 않고, 문구가 바뀐 참조도 놓치지 않음 |
| AC-GDP-009 | REQ-GDP-009 | MUST-PASS | **DEFERRED: OD-1** | 네 지침이 같은 절차를 서술 |
| AC-GDP-010 | REQ-GDP-010 | MUST-PASS | **DEFERRED: OD-1 = 1** | 런처 진입 워크트리 흐름, 맨손 `git worktree add` 부재 |
| AC-GDP-011 | REQ-GDP-011 | MUST-PASS | **DEFERRED: OD-1 = 2** | 조건 문구가 금지문과 안내 위치에 같이 실림 |
| AC-GDP-012 | REQ-GDP-012 | MUST-PASS | **DEFERRED: OD-1 = 3** | 은퇴 선언, `main_late_branch` 부재, `--issue` 기본값 유지 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 없음 | 사본 일치, `delivery.md` 의도된 차이 보존, 선택한 미러 테스트 3개가 모두 실행·통과 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 없음 | `.toml` 재생성과 `agents-emit-check` exit 0 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 없음 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 없음 | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |

AC-GDP-016은 요구사항 추적 밖의 절차 점검이다. 항상 로드되는 규칙을 세션 도중에 고치지 않기 위한 실행 순서 제약이며, 결과물의 성질이 아니라 run-phase의 커밋 순서를 본다.

## §D.1 Given-When-Then과 판정 명령

### AC-GDP-001 — `git fetch` 가 독립·병렬 항목에서 빠지고 순서가 지시됨

```
GIVEN manager-git.md 로컬·템플릿 사본의 "## Synchronization" 절
WHEN run-phase 편집 뒤 절을 추출해 검사하면
THEN 한 줄에 "fetch" 와 independent·parallel·batch 중 하나가 함께 나오는 줄이 없고
 AND fetch 가 먼저 끝난다는 순서 문장이 있다
```

대조(기준 트리):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $E/base-manager-git.md > $E/ac001-base-sync.md
/usr/bin/grep -n -E 'fetch.*(independent|para[l]lel|batch)|(independent|para[l]lel|batch).*fetch' $E/ac001-base-sync.md > $E/ac001-control-absent.txt
# 기대: exit 0, 1줄 (기준 트리 156행 문장)
/usr/bin/grep -n -E 'fetch[^.]*(first|before|then|completes)' $E/ac001-base-sync.md > $E/ac001-control-order.txt
# 기대: exit 1 — 순서 검출기는 기준 트리에서 빨강
```

판정(로컬·템플릿 각각):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' .claude/agents/moai/manager-git.md > $E/ac001-local-sync.md
/usr/bin/grep -n -E 'fetch.*(independent|para[l]lel|batch)|(independent|para[l]lel|batch).*fetch' $E/ac001-local-sync.md > $E/ac001-local-absent.txt
# 기대: exit 1
/usr/bin/grep -n -E 'fetch[^.]*(first|before|then|completes)' $E/ac001-local-sync.md > $E/ac001-local-order.txt
# 기대: exit 0, 적중 줄이 fetch 가 rev-list 보다 먼저라는 뜻인지 읽어 기록
```

템플릿 사본은 경로를 `$T/.claude/agents/moai/manager-git.md` 로 바꿔 같은 명령을 실행한다. 생성물 `$T/.codex/agents/moai/manager-git.toml` 에도 부재 명령을 실행한다(대조: `$E/base-manager-git.toml` 에서 같은 부재 검출식 exit 0, 150행).

1회차 감사 뮤턴트("`git fetch` then … all in parallel" 를 한 배치로 지시)에 대한 재실행(plan 작성 시점, 세션 스크래치 파일):

```
absent 검출식 → 3:Issue these as ONE single-turn multi-Bash batch: `git fetch` then …, all in parallel.
mutant-absent-exit=0          # 기대 exit 1 과 달라 판정 FAIL — 뮤턴트를 막음
올바르게 고친 문장 픽스처 → fixed-absent-exit=1, 순서 검출 1줄(exit 0)
```

### AC-GDP-002 — Pre-Spawn 코드 블록의 순서 보장

```
GIVEN agent-common-protocol.md 의 Pre-Spawn Sync Check 절에 있는 bash 코드 블록
WHEN 편집 뒤 그 블록만 추출해 검사하면
THEN 앞 공백을 허용해 fetch 로 시작하면서 같은 줄에 rev-list 가 없는 줄이 0개이고
 AND 앞 공백을 허용해 rev-list 로 시작하는 줄이 0개이고
 AND fetch 와 rev-list 를 ";" 또는 "&&" 로 이은 줄이 정확히 1개이고
 AND 세 번째 명령 줄과 두 해석 표의 행이 기준 트리와 같다
```

블록 추출(대조·판정 공통):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' <파일> > <블록 파일>
```

대조(기준 트리 — 세 검사 모두 빨강):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' $E/base-agent-common-protocol.md > $E/ac002-base-block.md
awk '/^[[:space:]]*[g]it fetch/ && !/[g]it rev-list/' $E/ac002-base-block.md > $E/ac002-base-a.txt
# 기대: 1줄 (블록 2행 — 기준 트리 296행)
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-base-block.md > $E/ac002-base-b.txt
# 기대: exit 0 (블록 5행 — 기준 트리 299행)
/usr/bin/grep -c -E '^[[:space:]]*[g]it fetch origin main 2>&1[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin/main[.][.][.]HEAD' $E/ac002-base-block.md > $E/ac002-base-c.txt
# 기대: 0
```

판정(로컬·템플릿 각각):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-block.md
awk '/^[[:space:]]*[g]it fetch/ && !/[g]it rev-list/' $E/ac002-local-block.md > $E/ac002-local-a.txt
test -s $E/ac002-local-a.txt
# 기대: exit 1 (파일이 비어 있음)
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-local-block.md > $E/ac002-local-b.txt
# 기대: exit 1
/usr/bin/grep -c -E '^[[:space:]]*[g]it fetch origin main 2>&1[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin/main[.][.][.]HEAD' $E/ac002-local-block.md > $E/ac002-local-c.txt
# 기대: 1
/usr/bin/grep -c 'moai session list --json --filter-spec=' $E/ac002-local-block.md > $E/ac002-local-session.txt
# 기대: 1 (기준 트리도 1)
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' $E/base-agent-common-protocol.md > $E/ac002-base-section.md
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-section.md
/usr/bin/grep -E '^[|] ' $E/ac002-base-section.md > $E/ac002-base-matrix.txt
/usr/bin/grep -E '^[|] ' $E/ac002-local-section.md > $E/ac002-local-matrix.txt
diff $E/ac002-base-matrix.txt $E/ac002-local-matrix.txt > $E/ac002-matrix.diff
# 기대: exit 0 (기준 트리의 표 행은 8개 — 구분 행 `|-` 은 세지 않음)
```

보조 검사(판정 근거가 아님): 파일 전체에서 이어 붙인 줄 개수는 2(Pre-Spawn, Pre-Edit). 1회차에서 이 파일 전체 개수만으로는 산문에 넣은 이어 붙인 형태에 속았으므로 판정에 쓰지 않는다.

1회차 감사 뮤턴트(fetch 줄 끝에 `# step 1` 주석, 산문에 이어 붙인 형태 추가)에 대한 재실행(plan 작성 시점, 세션 스크래치 파일):

```
블록 추출 결과 2행: git fetch origin main 2>&1  # step 1 / 5행: git rev-list --count --left-right origin/main...HEAD
a-lines=1        # 기대 0 → FAIL
b-exit=0         # 기대 1 → FAIL
c=0 (exit 1)     # 기대 1 → FAIL
old-filewide=2   # 옛 판정은 이 값으로 통과했음
올바르게 고친 픽스처(블록 안에서 fetch; rev-list 한 줄) → a-lines=0, b-exit=1, c=1 — 세 검사 모두 통과
```

### AC-GDP-003 — Pre-Edit Sync Check 절 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절(제목부터 #### The sweep prohibition 앞까지)
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조(검출기가 변경을 잡는지): AC-GDP-002의 `ac002-base-section.md` 와 `ac002-local-section.md` 를 `diff` 하면 exit 1.

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
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' $E/base-delivery.md > $E/ac004-control-squash.txt
# 기대: 2 (343·355행)
/usr/bin/grep -n 'merge_method' $E/base-delivery.md > $E/ac004-control-source.txt
# 기대: exit 1 — 출처 검출기는 기준 트리에서 빨강
```

판정(로컬·템플릿 각각):

```bash
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-squash.txt
# 기대: 0
/usr/bin/grep -c 'gh pr merge --<merge_method> --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-resolved.txt
# 기대: 2 이상
/usr/bin/grep -n -E 'merge_method.*(squash|default)' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-source.txt
# 기대: exit 0, 적중 줄이 해석 출처를 설명하는지 읽어 기록
```

### AC-GDP-005 — 고정 `--squash` 병합 예시 부재, 기본값 설명 유지

```
GIVEN 범위 다섯 파일(로컬·템플릿)과 생성물 .toml
WHEN 편집 뒤 --squash 가 붙은 gh pr merge 를 모두 찾으면
THEN manager-git.md 와 .toml 에서만 1줄씩 나오고, 그 줄은 기본값 설명 문장이며
 AND 나머지 파일에서는 0줄이다
```

이 기준은 OD-1이 어떤 선택지로 결정되어도 성립해야 한다(블록 유지+치환, 블록 제거 모두 해당).

대조(기준 트리):

```bash
/usr/bin/grep -n -E 'gh pr merge[^|]*--squash' $E/base-manager-git.md $E/base-delivery.md $E/base-manager-git.toml > $E/ac005-control.txt
# 기대: 6줄 — manager-git 32·114, delivery 343·355, .toml 26·108
```

판정(로컬·템플릿 각각, 파일마다 따로 셈):

```bash
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/agents/moai/manager-git.md > $E/ac005-mg.txt
# 기대: 1
/usr/bin/grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-mg-default.txt
# 기대: 1 — 위의 1줄이 기본값 설명 문장임을 확인
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/skills/moai/workflows/sync/delivery.md > $E/ac005-delivery.txt
# 기대: 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/core/agent-common-protocol.md > $E/ac005-others.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $T/.codex/agents/moai/manager-git.toml > $E/ac005-toml.txt
# 기대: 1 (기본값 설명 문장)
```

1회차 감사 뮤턴트(`gh pr merge 42 --squash --delete-branch`, `gh pr merge --squash <PR> --delete-branch`)는 이 검출식에 걸리므로 해당 파일의 개수가 기대값을 넘어 FAIL로 판정된다.

### AC-GDP-006 — auto-merge 기본값 단일 기준 **[DEFERRED: OD-2]**

```
GIVEN OD-2 로 정한 기준(A: delivery.md 기본값 / B: manager-git.md 옵트인 / C: workflow.worktree.auto_merge)
WHEN delivery.md Step 3.4 와 manager-git.md Team Mode·PR Auto-Merge 절을 읽으면
THEN 기준이 아닌 쪽에 반대 기본값을 적은 문장이 0개이고
 AND 기준이 아닌 쪽은 기준을 이름으로 참조한다
```

대조: 기준 트리에서 `delivery.md:337`(기본 병합)과 `manager-git.md:148`·`166`(플래그 필수)이 동시에 적중한다(모순 존재 확인). 판정 명령은 결정된 선택지의 기준 문구를 고정한 뒤 확정한다.

### AC-GDP-007 — 금지 집합 전체 검출식의 분류 장부

```
GIVEN 범위 네 파일(로컬·템플릿)
WHEN AGENTS.md §2 금지 집합 전체를 잡는 검출식으로 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND unconditioned-primary-instruction 분류가 0행이다
```

검출식(대조·판정 공통) — `checkout`·`switch`·`reset --hard`·`pull`·`merge`·`rebase`·`stash` 와 변경형 `branch`:

```
[g]it (checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it branch (-[a-zA-Z]*[fDmMcCdtu]|[^-])
```

대조(기준 트리 — 편집 전 재측정):

```bash
git grep -n -E '[g]it (checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it branch (-[a-zA-Z]*[fDmMcCdtu]|[^-])' $BASE -- $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-base-hits.txt
# 기대: exit 0, 20줄 — manager-git 10(42·96·110·119·121·122·125·127·137·160),
#   spec-workflow 5(50·56·58·59·62), spec-assembly 1(336), delivery 4(276·323·324·328)
# (plan 0.1.1 작성 시점 측정: 이 워크트리, $BASE. 0.1.0의 좁은 검출식은 14줄이었다)
```

1회차 감사 뮤턴트(`git pull --rebase origin main`, `git branch -f main origin/main`, `git merge --ff-only origin/main`)에 대한 재실행(plan 작성 시점): 넓힌 검출식이 3줄 모두 잡음(exit 0). 옛 검출식은 0줄(exit 1)이었다.

판정:

```bash
/usr/bin/grep -n -E '[g]it (checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it branch (-[a-zA-Z]*[fDmMcCdtu]|[^-])' $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-post-hits.txt
# exit 0 이면 적중 줄마다 장부 행을 만든다. exit 1 이면 장부는 "적중 0" 한 줄
```

장부 `.moai/reports/t622/ac01-classification.md` 의 분류값은 다섯 중 하나다.

| 분류 | 뜻 | 기준 트리에서의 예 |
|---|---|---|
| `prohibition-context` | 금지 목록·경고 속의 명령 | — |
| `narrative` | 명령을 실행하라고 하지 않고 서술만 함 | `spec-workflow.md:62` |
| `worktree-internal` | 진입한 워크트리 안에서 실행 | `delivery.md:276` |
| `conditioned` | OD-1 선택지 2에서만. 조건 문구를 인용 | — |
| `unconditioned-primary-instruction` | primary checkout에서 조건 없이 실행하라는 안내 | 0.1.0 §A.2 표의 안내 줄 |

합격 조건은 마지막 분류 0행.

### AC-GDP-008 — 절차 참조가 끊기지 않음

```
GIVEN spec-assembly.md, spec-workflow.md, delivery.md
WHEN (1) 형식을 갖춘 manager-git.md 절 참조를 모으고
 AND (2) manager-git 절 참조나 절차 제목을 담은 줄을 형식과 무관하게 모아 파일마다 개수를 비교하고
 AND (3) 형식을 갖춘 참조의 제목을 manager-git.md 제목 줄에서 조회하면
THEN 파일마다 (1) 개수와 (2) 개수가 같고
 AND (3) 조회에서 제목마다 대상이 정확히 한 줄 있고
 AND (1)·(2) 합계가 모두 0인 경우는 OD-1 = 선택지 3이면서 AC-GDP-012가 PASS일 때만 PASS이며, 그 밖에는 FAIL이다
```

판정:

```bash
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-strict-count.txt
/usr/bin/grep -c -E 'manager-[g]it[^§]*§ [A-Z]|Invocation Pattern' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-loose-count.txt
diff $E/ac008-strict-count.txt $E/ac008-loose-count.txt > $E/ac008-count.diff
# 기대: exit 0 — 차이가 있으면 문구가 바뀐 참조가 조회를 빠져나간 것이므로 FAIL
/usr/bin/grep -n -o -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-refs.txt
sed -n -e 's/.*§ /### /p' -- $E/ac008-refs.txt > $E/ac008-headings-raw.txt
sort -u -o $E/ac008-headings.txt $E/ac008-headings-raw.txt
wc -l < $E/ac008-headings.txt > $E/ac008-heading-count.txt
/usr/bin/grep -c -x -F -f $E/ac008-headings.txt .claude/agents/moai/manager-git.md > $E/ac008-lookup.txt
# 기대: ac008-lookup.txt 의 값 = ac008-heading-count.txt 의 값
```

제목은 대문자로 시작하는 낱말만 잇는다. 소문자까지 받는 문자 클래스는 뒤따르는 "for the" 까지 삼켜 조회를 거짓 끊김으로 만든다(0.1.0 작성 시점 측정: spec-assembly.md:340).

대조(plan 작성 시점 측정, 세션 스크래치 파일):

```
기준 트리: 형식 참조 개수 spec-assembly 1 · spec-workflow 1 · delivery 0 / 느슨한 개수 1 · 1 · 0 → 같음
기준 트리 조회: 제목 1개(### Late-Branch Invocation Pattern) → manager-git.md 복사본에서 1
제목 줄을 지운 manager-git.md 픽스처 조회 → 0 (exit 1) — 전체 조회가 끊긴 참조를 보고함
```

1회차 감사 뮤턴트("see the manager-git agent § Late-Branch Invocation Pattern")에 대한 재실행: 형식 참조 0, 느슨한 개수 1 → 개수 불일치로 FAIL. 옛 판정은 형식 참조 0을 은퇴 경로로 넘겨 검사 없이 통과시켰다.

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

대조: `/usr/bin/grep -n '[g]it worktree add' $T/.claude/rules/moai/workflow/main-checkout-branch-guard.md` exit 0(38행 — 검출기가 형태를 잡음).

### AC-GDP-011 — 조건 문구 공유 **[DEFERRED: OD-1 = 선택지 2]**

```
GIVEN 선택지 2로 정한 조건 문장 COND
WHEN 금지문과 남은 안내 위치를 검사하면
THEN 루트 AGENTS.md, 템플릿 AGENTS.md, main-checkout-branch-guard.md 로컬·템플릿, 남은 각 안내 위치에 COND 가 1회 이상 있다
 AND 두 AGENTS.md 의 편집 전후 바이트 수가 기록되어 있다
```

대조: `/usr/bin/grep -c -F 'Never change branch state in the primary checkout' $E/base-AGENTS.md` 기대 1(고정 문자열 검출기가 이 파일에서 동작함). `$E/base-AGENTS.md` 는 plan.md §C 2단계에서 반출한다.

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
/usr/bin/grep -n 'main_late_branch' $E/base-manager-git.md > $E/ac012-control.txt
# 기대: exit 0 (137행)
/usr/bin/grep -c 'default skips GitHub Issue creation' $E/base-SKILL.md > $E/ac012-issue-base.txt
# 기대: 2 (SKILL.md 124·227행). $E/base-SKILL.md 는 plan.md §C 2단계에서 반출한다
```

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 다섯 파일의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md, spec-workflow.md, spec-assembly.md, agent-common-protocol.md 는 diff exit 0 이고
 AND delivery.md 는 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 선택한 미러 테스트 3개가 각자 최상위 PASS 줄을 내며 exit 0 이다
```

`manager-git.md` 와 `agent-common-protocol.md` 는 미러 테스트의 바이트 동일 대상에서 빠져 있다(`rule_template_mirror_test.go` 주석). 이 두 파일의 사본 일치를 지키는 것은 아래 `diff` 뿐이다.

대조: 기준 트리에서 `delivery.md` 두 사본의 `diff` 는 exit 1(검출기가 차이를 보고함). 본문 비교 검출기는 빈 파일(`manager-git.md` 차이 본문)과 `delivery.md` 차이 본문을 `diff` 하면 exit 1.

판정:

```bash
diff .claude/agents/moai/manager-git.md $T/.claude/agents/moai/manager-git.md > $E/ac013-manager-git.diff
# 기대: exit 0 (spec-workflow.md, spec-assembly.md, agent-common-protocol.md 도 같은 형태)
diff $E/base-delivery.md $E/base-delivery-template.md > $E/ac013-delivery-base.diff
diff .claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac013-delivery-post.diff
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-base.diff > $E/ac013-delivery-base.body
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-post.diff > $E/ac013-delivery-post.body
diff $E/ac013-delivery-base.body $E/ac013-delivery-post.body > $E/ac013-delivery-body.diff
# 기대: exit 0
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0
/usr/bin/grep -c -E '^--- PASS: (TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 정확히 3 — 셋 중 하나라도 이름이 바뀌어 선택에서 빠지면 3이 되지 않는다
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

M1이 템플릿 `manager-git.md:156` 을, M2가 `:114` 를 항상 고치므로 첫 점검 RED는 어느 결정 조합에서도 나온다.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, 커밋 SHA 형태, CLAUDE.local 참조가 추가 줄에 없다
```

REQ-GDP-015의 프로그래밍 언어 편향 절은 기계 판정이 없다. 추가 줄을 읽어 16개 지원 프로그래밍 언어 중 하나를 앞세운 표현이 없는지 기록하고, CI의 `template-neutrality-check` 결과를 함께 적는다. `TestTemplateNoInternalContentLeak` 는 CI 보조 방어선으로 남기되, 날짜·SHA 분류의 적용 경로를 확인하지 않았으므로(spec.md §E) 이 기준은 그 테스트에 기대지 않는다.

대조(검출기가 형태를 잡는지 — 이 SPEC의 spec.md에서 각각 1 이상):

```bash
/usr/bin/grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-specid.txt
/usr/bin/grep -c -E 'REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-req.txt
/usr/bin/grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-date.txt
/usr/bin/grep -c -E '(^|[^0-9A-Za-z])[0-9]*[a-f][0-9a-f]{6,39}([^0-9A-Za-z]|$)' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-sha.txt
# 기대: 네 값 모두 1 이상. plan 작성 시점 값은 progress.md §E.1 에 따로 기록
```

1회차 대조는 한 세그먼트만 받는 SPEC-ID 검출식이라 이 저장소의 다중 세그먼트 ID를 하나도 잡지 못했다(0). 위 검출식은 다중 세그먼트를 받는다.

판정:

```bash
git diff $BASE -- $T/ > $E/ac015-template.diff
/usr/bin/grep -n -E '^[+][^+].*SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-specid.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-req.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*20[0-9]{2}-[0-9]{2}-[0-9]{2}' $E/ac015-template.diff > $E/ac015-date.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*(^|[^0-9A-Za-z])[0-9]*[a-f][0-9a-f]{6,39}([^0-9A-Za-z]|$)' $E/ac015-template.diff > $E/ac015-sha.txt
# 기대: exit 1. 적중이 있으면 줄을 읽어, 커밋 SHA가 아닌 16진 낱말이면 그 사실을 기록하고 SHA면 FAIL
/usr/bin/grep -n -E '^[+][^+].*CLAUDE[.]local' $E/ac015-template.diff > $E/ac015-local-ref.txt
# 기대: exit 1
```

SHA 검출식은 16진 문자 a-f를 하나 이상 요구해 숫자만 있는 값(예: 토큰 수)을 거르지만, 숫자로만 된 SHA는 놓친다. 이 한계는 적중 없음이 SHA 부재를 완전히 보장하지 않는다는 뜻이며 잔여 위험으로 기록한다.

### AC-GDP-016 — 항상 로드 규칙 편집이 마지막 (SHOULD, 절차 점검)

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
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. 추출 파일의 마지막 줄이 기대한 끝 제목인지 확인해 기록한다. AC-GDP-002의 `awk` 블록 추출은 절 제목 뒤 첫 `bash` 코드 블록만 읽으므로, 편집으로 코드 블록 언어 표기가 바뀌면 빈 블록이 나온다 — 블록 파일이 비어 있으면 판정 불가로 기록한다.
- 결정 대기 기준(AC-GDP-006, 009~012)은 결정 전에 FAIL로 기록하지 않고 DEFERRED로 기록한다.
- `delivery.md` 편집으로 줄 수가 바뀌면 479행 꼬리말의 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다).
- `[^-]` 로 변경형 `git branch` 를 잡는 검출식은 산문 속 "`git branch`" 에도 걸린다. 그런 적중은 AC-GDP-007 장부에서 `narrative` 로 분류한다.

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다.
- `go test` 선택 실행이 최상위 PASS 줄 3개를 내지 않으면 합격이 아니다.
- 로컬 전체 스위트는 돌리지 않는다. 전체 판정은 통합 뒤 CI가 한다.

## §D.4 완료 정의 (Definition of Done)

- 결정 불필요 기준 AC-GDP-001~005, 007, 008, 013~015가 PASS이고 AC-GDP-016이 PASS 또는 사유 기록.
- 결정된 선택지에 해당하는 기준만 PASS, 나머지 선택지 기준은 N/A로 기록.
- OD-1·OD-2 중 미결정이 남으면 해당 기준은 DEFERRED이고, SPEC은 `implemented` 로 넘어가지 않는다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
