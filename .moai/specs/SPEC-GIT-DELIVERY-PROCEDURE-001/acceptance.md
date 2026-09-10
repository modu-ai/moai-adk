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
- 검출식 안의 `git` 은 `[g]it` 으로 쓴다. 워크트리 세션 가드가 복합 명령 안의 `git` 낱말을 git 실행으로 보고 거부하는 일을 피한다. 같은 이유로 `parallel` 은 `para[l]lel` 로, perl 코드 안의 `git` 은 `\x67it` 로 쓴다. 셸 변수를 받는 `sed`·`perl` 은 가드가 거부할 수 있으므로 경로를 글자 그대로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 셸 `grep` 래퍼는 UTF-8이 아닌 파일을 출력 없이 건너뛰어, 개수가 찍히지 않은 결과가 0처럼 읽힐 수 있다. 개수가 찍히지 않은 결과는 판정 불가로 기록한다.

## §D AC 표

| AC ID | REQ | 등급 | 상태 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 활성 | `manager-git.md` 동기화 절의 문단·목록 묶음에서 fetch 가 rev-list 와 같은 배치로 묶이지 않고 순서가 지시됨 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 활성 | Pre-Spawn 코드 블록에 단독 fetch·단독 rev-list 줄이 없고, 이어 붙인 줄 1개, rev-list 1개 |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 활성 | Pre-Edit Sync Check 절 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 활성 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 활성 | 범위 파일 전체에서 `--squash` 고정 `gh pr merge` 가 기본값 설명 문장뿐 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | 활성 (OD-2 = B) | `delivery.md`·`doc-execution.md` 에 워크트리 기본 병합 문구가 없고 `manager-git.md` 를 기준으로 밝힘 |
| AC-GDP-007 | REQ-GDP-007 | MUST-PASS | 활성 | 금지 집합 전체 검출식의 적중이 모두 분류되고 primary checkout 실행 안내가 0 |
| AC-GDP-008 | REQ-GDP-008 | MUST-PASS | 활성 | 파일별 참조 하한, 형식·느슨한 개수 일치, 제목 조회 |
| AC-GDP-009 | REQ-GDP-009 | MUST-PASS | 활성 (OD-1 = 1) | 네 지침 절이 §C.2 흐름의 명령을 담고 PR 브랜치 접두 집합이 같음 |
| AC-GDP-010 | REQ-GDP-010 | MUST-PASS | 활성 (OD-1 = 1) | 맨손 `git worktree add` 부재, 블록 밖 자리 처리, Frozen 수정 기록과 헌법 검증 DRIFT 없음 |
| AC-GDP-011 | REQ-GDP-011 | — | **철회** | OD-1 선택지 2 경로 |
| AC-GDP-012 | REQ-GDP-012 | — | **철회** | OD-1 선택지 3 경로 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 활성 | 사본 일치, `delivery.md`·`doc-execution.md` 의도된 차이 보존, 미러 테스트 3개 실행·통과 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 활성 | `.toml` 재생성과 `agents-emit-check` exit 0 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 활성 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 활성 | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |

AC-GDP-016은 요구사항 추적 밖의 절차 점검이다. 항상 로드되는 규칙을 세션 도중에 고치지 않기 위한 실행 순서 제약이며, 결과물의 성질이 아니라 run-phase의 커밋 순서를 본다.

## §D.1 Given-When-Then과 판정 명령

### AC-GDP-001 — fetch 와 rev-list 가 같은 배치로 묶이지 않고 순서가 지시됨

```
GIVEN manager-git.md 로컬·템플릿 사본의 "## Synchronization" 절
WHEN 절을 빈 줄로 나눈 문단(목록 묶음은 한 문단)마다 검사하면
THEN fetch 와 rev-list 를 함께 담은 문단이 1개 이상 있고
 AND 그중 배치·병렬 낱말을 담으면서 순서 낱말이 없는 문단이 0개이고
 AND 읽기 단계에서, fetch 와 rev-list 를 함께 담은 모든 문단이
     (1) fetch 가 끝난 뒤 rev-list 를 실행한다고 말하고
     (2) fetch 와 rev-list 를 같은 배치·같은 목록·병렬 묶음에 넣지 않는다
```

판정은 줄이 아니라 문단 단위다. 1회차 줄 단위 검출식은 올바른 한 문장("Run `git fetch` first; once it completes, issue the independent reads … as one batch")을 빨강으로 만들고, fetch 를 목록 항목 하나로 떼어 놓은 형태는 놓쳤다.

대조(기준 트리):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $E/base-manager-git.md > $E/ac001-base-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-base-sync.md > $E/ac001-base-paras.txt
# 기대: 1
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-base-sync.md > $E/ac001-base-autofail.md
test -s $E/ac001-base-autofail.md
# 기대: exit 0 — 자동 실패 검출기는 기준 트리에서 빨강(156행 문단)
```

판정(로컬·템플릿 각각):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' .claude/agents/moai/manager-git.md > $E/ac001-local-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-local-sync.md > $E/ac001-local-paras.txt
# 기대: 1 이상 (0이면 순서 문장이 사라진 것 — FAIL)
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/' $E/ac001-local-sync.md > $E/ac001-local-paras.md
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-local-sync.md > $E/ac001-local-autofail.md
test -s $E/ac001-local-autofail.md
# 기대: exit 1
awk 'BEGIN{RS=""; ORS="\n\n"} /(^|\n)[-*] [^\n]*fetch/ && /(^|\n)[-*] [^\n]*rev-list/' $E/ac001-local-sync.md > $E/ac001-local-listgroup.md
# 비어 있지 않으면 목록 안에 fetch 와 rev-list 가 함께 있다는 뜻 — 읽기 단계 (2)에서 반드시 판정
```

읽기 단계: `ac001-local-paras.md` 의 문단마다 위 THEN (1)·(2)를 판정해 `$E/ac001-reading.md` 에 기록한다. `ac001-local-listgroup.md` 가 비어 있지 않으면 그 목록이 한 배치를 뜻하는지 반드시 적는다. 템플릿 사본은 경로를 `$T/.claude/agents/moai/manager-git.md` 로 바꾼다. 생성물 `$T/.codex/agents/moai/manager-git.toml` 에도 같은 자동 실패 검사를 실행한다.

뮤턴트 재실행(plan 작성 시점, 세션 스크래치 파일):

```
기준 트리 절                          → 문단 1, autofail 1, listgroup 0
"`git fetch` then … all in parallel"  → 문단 1, autofail 1 → FAIL
2회차 목록 뮤턴트 (ONE single-turn multi-Bash call: - `git fetch` (first) - `git status` - `git rev-list`)
                                      → 문단 1, autofail 0, listgroup 1 → 읽기 단계 (2)에서 FAIL
2회차 올바른 문장 (… first; once it completes, issue the independent reads … as one batch)
                                      → 문단 1, autofail 0, listgroup 0 → 빨강 아님
```

### AC-GDP-002 — Pre-Spawn 코드 블록의 순서 보장

```
GIVEN agent-common-protocol.md 의 Pre-Spawn Sync Check 절에 있는 bash 코드 블록
WHEN 편집 뒤 그 블록만 추출해 검사하면
THEN 앞 공백을 허용해 fetch 로 시작하면서 같은 줄에 rev-list 가 없는 줄이 0개이고
 AND 앞 공백을 허용해 rev-list 로 시작하는 줄이 0개이고
 AND fetch 와 rev-list 를 ";" 또는 "&&" 로 이은 줄이 정확히 1개이고
 AND rev-list 를 담은 줄이 정확히 1개이고
 AND 세 번째 명령 줄과 두 해석 표의 행이 기준 트리와 같다
```

블록 추출(대조·판정 공통):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' <파일> > <블록 파일>
```

대조(기준 트리):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' $E/base-agent-common-protocol.md > $E/ac002-base-block.md
awk '/^[[:space:]]*[g]it fetch/ && !/[g]it rev-list/' $E/ac002-base-block.md > $E/ac002-base-a.txt
# 기대: 1줄 (블록 2행 — 기준 트리 296행)
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-base-block.md > $E/ac002-base-b.txt
# 기대: exit 0 (블록 5행 — 기준 트리 299행)
/usr/bin/grep -c -E '^[[:space:]]*[g]it fetch origin main 2>&1[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin/main[.][.][.]HEAD' $E/ac002-base-block.md > $E/ac002-base-c.txt
# 기대: 0
/usr/bin/grep -c 'rev-list' $E/ac002-base-block.md > $E/ac002-base-d.txt
# 기대: 1 — 개수 검출기가 동작함을 보이는 기록. 빨강은 a·b·c가 담당한다
```

판정(로컬·템플릿 각각):

```bash
awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-block.md
test -s $E/ac002-local-block.md
# 기대: exit 0 — 블록이 비면 판정 불가
awk '/^[[:space:]]*[g]it fetch/ && !/[g]it rev-list/' $E/ac002-local-block.md > $E/ac002-local-a.txt
test -s $E/ac002-local-a.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-local-block.md > $E/ac002-local-b.txt
# 기대: exit 1
/usr/bin/grep -c -E '^[[:space:]]*[g]it fetch origin main 2>&1[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin/main[.][.][.]HEAD' $E/ac002-local-block.md > $E/ac002-local-c.txt
# 기대: 1
/usr/bin/grep -c 'rev-list' $E/ac002-local-block.md > $E/ac002-local-d.txt
# 기대: 1
/usr/bin/grep -c 'moai session list --json --filter-spec=' $E/ac002-local-block.md > $E/ac002-local-session.txt
# 기대: 1
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' $E/base-agent-common-protocol.md > $E/ac002-base-section.md
sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p' .claude/rules/moai/core/agent-common-protocol.md > $E/ac002-local-section.md
/usr/bin/grep -E '^[|] ' $E/ac002-base-section.md > $E/ac002-base-matrix.txt
/usr/bin/grep -E '^[|] ' $E/ac002-local-section.md > $E/ac002-local-matrix.txt
diff $E/ac002-base-matrix.txt $E/ac002-local-matrix.txt > $E/ac002-matrix.diff
# 기대: exit 0 (기준 트리의 표 행은 8개 — 구분 행 `|-` 은 세지 않음)
```

뮤턴트 재실행(plan 작성 시점):

```
1회차 뮤턴트(fetch 줄 끝 `# step 1` 주석, 산문에 이어 붙인 형태) → a 1, b exit 0, c 0 → FAIL
2회차 뮤턴트(이어 붙인 줄 1개 + 블록 안 두 번째 `git -C . rev-list …` 줄) → rev-list 개수 2 → FAIL
올바르게 고친 픽스처(블록 안 fetch; rev-list 한 줄) → a 0, b exit 1, c 1, rev-list 1 → PASS
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
GIVEN 범위 여섯 파일(로컬·템플릿)과 생성물 .toml
WHEN 편집 뒤 --squash 가 붙은 gh pr merge 를 모두 찾으면
THEN manager-git.md 와 .toml 에서만 1줄씩 나오고, 그 줄은 기본값 설명 문장이며
 AND 나머지 파일에서는 0줄이다
```

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
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac005-others.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $T/.codex/agents/moai/manager-git.toml > $E/ac005-toml.txt
# 기대: 1 (기본값 설명 문장)
```

1회차 감사 뮤턴트(`gh pr merge 42 --squash --delete-branch`, `gh pr merge --squash <PR> --delete-branch`)는 이 검출식에 걸리므로 해당 파일의 개수가 기대값을 넘어 FAIL로 판정된다.

### AC-GDP-006 — auto-merge 기본값 단일 기준 (OD-2 = B)

```
GIVEN delivery.md 로컬·템플릿 사본의 Step 3.4 절과 doc-execution.md 로컬·템플릿 사본의 "Worktree Context Detection" 소절
WHEN 편집 뒤 두 절을 추출해 검사하면
THEN 워크트리 문맥을 기본 병합과 묶는 문구가 두 절 모두 0개이고
 AND 두 절이 각각 manager-git.md 를 기준으로 1회 이상 이름으로 밝히고
 AND manager-git.md 의 옵트인 문장 두 개(148행, 166행)가 각각 1회 남아 있고
 AND 읽기 단계에서, 두 절 어디에도 워크트리 문맥만으로 병합이 일어난다는 문장이 없다
```

절 추출(대조·판정 공통):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery.md 경로> > <delivery 절 파일>
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' <doc-execution.md 경로> > <doc-execution 절 파일>
```

대조(기준 트리):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' $E/base-delivery.md > $E/ac006-base-dl.md
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' $E/base-doc-execution.md > $E/ac006-base-de.md
/usr/bin/grep -n -i -E 'default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set' $E/ac006-base-dl.md > $E/ac006-base-dl-default.txt
# 기대: exit 0 — 절 안 8행(트리거)·20행(`--merge` 폐기 경고), 파일 기준 337·349행
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default' $E/ac006-base-de.md > $E/ac006-base-de-default.txt
# 기대: exit 0 — 절 안 8행, 파일 기준 36행
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-base-dl.md $E/ac006-base-de.md > $E/ac006-base-source.txt
# 기대: 파일마다 0 — 기준 명시 검출기는 기준 트리에서 빨강
/usr/bin/grep -c -F 'Execute only with `--auto-merge` flag AND all approvals obtained' $E/base-manager-git.md > $E/ac006-base-optin1.txt
/usr/bin/grep -c -F 'Auto-merge: only with the `--auto-merge` flag' $E/base-manager-git.md > $E/ac006-base-optin2.txt
# 기대: 각각 1
```

판정(로컬·템플릿 각각):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac006-local-dl.md
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac006-local-de.md
test -s $E/ac006-local-dl.md
test -s $E/ac006-local-de.md
# 기대: 둘 다 exit 0 — 절이 비면 판정 불가
/usr/bin/grep -n -i -E 'default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set' $E/ac006-local-dl.md > $E/ac006-local-dl-default.txt
# 기대: exit 1
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default' $E/ac006-local-de.md > $E/ac006-local-de-default.txt
# 기대: exit 1
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-local-dl.md $E/ac006-local-de.md > $E/ac006-local-source.txt
# 기대: 파일마다 1 이상
/usr/bin/grep -c -F 'Execute only with `--auto-merge` flag AND all approvals obtained' .claude/agents/moai/manager-git.md > $E/ac006-local-optin1.txt
/usr/bin/grep -c -F 'Auto-merge: only with the `--auto-merge` flag' .claude/agents/moai/manager-git.md > $E/ac006-local-optin2.txt
# 기대: 각각 1
```

읽기 단계: 두 절 파일을 읽어, 워크트리 문맥만으로 병합이 일어난다는 문장이 없는지(검출식이 예상하지 않은 표현 포함)와 병합 조건이 `--auto-merge` 로 적혀 있는지 `$E/ac006-reading.md` 에 기록한다.

### AC-GDP-007 — 금지 집합 전체 검출식의 분류 장부

```
GIVEN 범위 네 파일(로컬·템플릿)
WHEN AGENTS.md §2 금지 집합 전체를 잡는 검출식으로 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND unconditioned-primary-instruction 분류가 0행이다
```

검출식(대조·판정 공통) — `checkout`·`switch`·`reset --hard`·`pull`·`merge`·`rebase`·`stash`, 짧은·긴 옵션의 변경형 `branch`, 그리고 `git` 과 하위 명령 사이의 선택적 `-C <경로>`:

```
[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])
```

대조(기준 트리 — 편집 전 재측정):

```bash
git grep -n -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $BASE -- $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-base-hits.txt
# 기대: exit 0, 20줄 — manager-git 10(42·96·110·119·121·122·125·127·137·160),
#   spec-workflow 5(50·56·58·59·62), spec-assembly 1(336), delivery 4(276·323·324·328)
# (plan 0.1.2 작성 시점 측정: 이 워크트리, $BASE. 긴 옵션과 -C 형태를 더해도 기준 트리 값은 그대로)
```

뮤턴트 재실행(plan 작성 시점): 2회차 뮤턴트 8줄(`pull --rebase`, `branch -f`, `merge --ff-only`, `branch --force`, `branch --move`, `branch --delete`, `-C . switch`, `reset --keep`) 가운데 1-7줄이 모두 잡힘. 8줄 `reset --keep` 은 `AGENTS.md` §2 목록 밖이라 잡지 않는다.

판정:

```bash
/usr/bin/grep -n -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-post-hits.txt
# exit 0 이면 적중 줄마다 장부 행을 만든다. exit 1 이면 장부는 "적중 0" 한 줄
```

장부 `.moai/reports/t622/ac01-classification.md` 의 분류값은 넷 중 하나다. OD-1 선택지 2 철회로 0.1.1의 `conditioned` 분류는 없앴다.

| 분류 | 뜻 | 예 |
|---|---|---|
| `prohibition-context` | 금지 목록·경고 속의 명령 | — |
| `narrative` | 명령을 실행하라고 하지 않고 서술만 함 | 기준 트리 `spec-workflow.md:62` |
| `worktree-internal` | 런처로 진입한 워크트리 안에서 실행한다고 명시된 안내 | 기준 트리 `delivery.md:276`; 재설계 뒤 §C.2 C1의 `git branch -m`, §C.3의 `manager-git.md` 42·160행 |
| `unconditioned-primary-instruction` | primary checkout에서 실행하라는 안내 | 기준 트리 §A.2 표의 안내 줄 |

합격 조건은 마지막 분류 0행.

### AC-GDP-008 — 절차 참조가 끊기지 않음

```
GIVEN spec-assembly.md, spec-workflow.md, delivery.md
WHEN (1) 형식을 갖춘 manager-git.md 절 참조를 파일마다 세고
 AND (2) 대소문자를 가리는 느슨한 검출식으로 파일마다 세고
 AND (3) 형식을 갖춘 참조의 제목을 manager-git.md 제목 줄에서 조회하면
THEN spec-assembly.md 와 spec-workflow.md 의 (1) 개수가 각각 1 이상이고 (기준 트리 값이 각각 1인 파일별 하한)
 AND 파일마다 (1) 개수와 (2) 개수가 같고
 AND (3) 조회에서 제목마다 대상이 정확히 한 줄 있다
```

참조가 0건인 경우는 PASS가 될 수 없다. 0.1.1은 합계 0을 OD-1 선택지 3의 은퇴 경로로 넘겼지만, 선택지 3이 철회됐으므로 그 경로도 없다.

재설계(§C.2)는 참조 대상 절의 제목을 바꿀 수 있다. 이 기준은 제목에 무관하게 "대문자로 시작하는 낱말로 된 제목" 을 잡으므로, 제목이 바뀌어도 두 파일이 형식을 갖춘 참조를 유지하고 그 제목이 `manager-git.md` 에 있으면 통과한다. 느슨한 검출식에 대소문자 무시(`-i`)를 쓰지 않는 이유: 기준 트리 `spec-workflow.md` 17행의 에이전트 목록 줄이 `-i` 형태에 걸려 느슨한 개수가 2가 되어, 올바른 파일에서도 (1)=(2) 가 깨진다(plan 작성 시점 측정).

대조:

```bash
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' $E/base-spec-assembly.md $E/base-spec-workflow.md $E/base-delivery.md > $E/ac008-base-strict.txt
# 기대: 1 · 1 · 0
/usr/bin/grep -n -o -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' $E/base-spec-assembly.md $E/base-spec-workflow.md > $E/ac008-base-refs.txt
sed -n -e 's/.*§ /### /p' -- $E/ac008-base-refs.txt > $E/ac008-base-headings-raw.txt
sort -u -o $E/ac008-base-headings.txt $E/ac008-base-headings-raw.txt
/usr/bin/grep -v -x -F -f $E/ac008-base-headings.txt $E/base-manager-git.md > $E/ac008-fixture-manager-git.md
/usr/bin/grep -c -x -F -f $E/ac008-base-headings.txt $E/base-manager-git.md > $E/ac008-control-lookup-base.txt
# 기대: 1
/usr/bin/grep -c -x -F -f $E/ac008-base-headings.txt $E/ac008-fixture-manager-git.md > $E/ac008-control-lookup-fixture.txt
# 기대: 0, exit 1 — 제목을 지운 픽스처에서 전체 조회가 끊긴 참조를 보고함
perl -CSD -pe 's/ \x{a7} Late-Branch Invocation Pattern\./ \x{a7} late-branch invocation pattern./' -- $E/base-spec-workflow.md > $E/ac008-mutant-lower.md
perl -CSD -pe 's/see `\.claude\/agents\/moai\/manager-\x67it\.md` \x{a7} Late-Branch Invocation Pattern\./see the \x67it agent definition, Late-branch section./' -- $E/base-spec-workflow.md > $E/ac008-mutant-prose.md
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' $E/ac008-mutant-lower.md $E/ac008-mutant-prose.md > $E/ac008-control-mutants.txt
# 기대: 파일마다 0 — 하한 1 미만이므로 두 뮤턴트 모두 FAIL
```

판정:

```bash
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-strict-count.txt
# 기대: spec-assembly.md ≥ 1, spec-workflow.md ≥ 1 (파일별 하한)
/usr/bin/grep -c -E 'manager-[g]it[^§]*§ [A-Z]|Invocation Pattern' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-loose-count.txt
diff $E/ac008-strict-count.txt $E/ac008-loose-count.txt > $E/ac008-count.diff
# 기대: exit 0
/usr/bin/grep -n -o -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-refs.txt
sed -n -e 's/.*§ /### /p' -- $E/ac008-refs.txt > $E/ac008-headings-raw.txt
sort -u -o $E/ac008-headings.txt $E/ac008-headings-raw.txt
wc -l < $E/ac008-headings.txt > $E/ac008-heading-count.txt
/usr/bin/grep -c -x -F -f $E/ac008-headings.txt .claude/agents/moai/manager-git.md > $E/ac008-lookup.txt
# 기대: ac008-lookup.txt 의 값 = ac008-heading-count.txt 의 값, 그리고 두 값 모두 1 이상
```

뮤턴트 재실행(plan 작성 시점, 세션 스크래치 파일):

```
기준 트리 형식 참조 1 · 1 · 0, 느슨한 개수 1 · 1 · 0
2회차 소문자 뮤턴트(spec-workflow.md 62행 제목 소문자화)   → spec-workflow 형식 0 < 하한 1 → FAIL
2회차 산문 뮤턴트("see the git agent definition, Late-branch section.") → spec-workflow 형식 0 < 하한 1 → FAIL
1회차 뮤턴트("see the manager-git agent § Late-Branch Invocation Pattern") → 형식 0, 느슨한 1 → 불일치 + 하한 미달 → FAIL
제목을 지운 manager-git.md 픽스처 조회 → 0 (exit 1)
```

### AC-GDP-009 — 네 지침 절의 흐름 일치 (OD-1 = 1)

```
GIVEN 네 지침의 절: manager-git.md "### Late-Branch Invocation Pattern" 절,
      spec-workflow.md "[ZONE:Frozen] [HARD] Step ordering rules" 블록,
      spec-assembly.md "#### Late-branch Pre-check" 절, delivery.md "#### Step 3.3.5" 절
WHEN 편집 뒤 절을 추출해 §C.2 흐름의 명령 표지를 세고 PR 브랜치 접두 집합을 비교하면
THEN 아래 표의 필수 표지가 절마다 1회 이상 있고
 AND spec-workflow.md 블록과 manager-git.md 절의 PR 브랜치 접두 집합이 같고
 AND 읽기 단계에서, 네 절이 진입 → PR → 종결 순서를 §C.2와 같게 서술하고 primary checkout에서 브랜치 명령을 실행하라고 하지 않는다
```

절 추출 표지는 편집 뒤에도 남아야 한다. 제목을 바꿔야 하면 run-phase가 새 표지를 progress 기록에 적고 추출 명령을 같은 방식으로 바꾼다.

| 절 | 필수 표지 (각 1회 이상) |
|---|---|
| manager-git.md | `moai cc -w` · `EnterWorktree(` · `[g]it branch -m` · `[g]it push -u origin` · `gh pr create` · `--<merge_method>` · `ExitWorktree` · `[g]it worktree remove` |
| spec-workflow.md 블록 | (`moai cc -w` 또는 `EnterWorktree(`) · `ExitWorktree` · `[g]it worktree remove` |
| spec-assembly.md 절 | `moai cc -w` 또는 `EnterWorktree(` |
| delivery.md Step 3.3.5 절 | `ExitWorktree` |

절 추출(대조·판정 공통):

```bash
awk '/^### Late-Branch Invocation Pattern/{s=1; print; next} s && /^##/{exit} s' <manager-git.md> > <mg 절>
awk '/^\[ZONE:Frozen\] \[HARD\] Step ordering rules/{s=1} s && /^\[SHOULD\] Anti-patterns/{exit} s' <spec-workflow.md> > <sw 블록>
awk '/^#### Late-branch Pre-check/{s=1; print; next} s && /^####/{exit} s' <spec-assembly.md> > <sa 절>
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' <delivery.md> > <dl 절>
```

대조(기준 트리):

```bash
awk '/^### Late-Branch Invocation Pattern/{s=1; print; next} s && /^##/{exit} s' $E/base-manager-git.md > $E/ac009-base-mg.md
awk '/^\[ZONE:Frozen\] \[HARD\] Step ordering rules/{s=1} s && /^\[SHOULD\] Anti-patterns/{exit} s' $E/base-spec-workflow.md > $E/ac009-base-sw.md
awk '/^#### Late-branch Pre-check/{s=1; print; next} s && /^####/{exit} s' $E/base-spec-assembly.md > $E/ac009-base-sa.md
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' $E/base-delivery.md > $E/ac009-base-dl.md
/usr/bin/grep -c 'ExitWorktree' $E/ac009-base-mg.md $E/ac009-base-sw.md $E/ac009-base-dl.md > $E/ac009-base-exit.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'moai cc -w|EnterWorktree\(' $E/ac009-base-sw.md $E/ac009-base-sa.md > $E/ac009-base-entry.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'EnterWorktree\(|[g]it branch -m|--<merge_method>|[g]it worktree remove' $E/ac009-base-mg.md > $E/ac009-base-mg-markers.txt
# 기대: 0
/usr/bin/grep -o -E '(plan|feat|sync|chore)/SPEC-' $E/ac009-base-sw.md > $E/ac009-base-sw-prefix-raw.txt
sort -u -o $E/ac009-base-sw-prefix.txt $E/ac009-base-sw-prefix-raw.txt
/usr/bin/grep -o -E '(plan|feat|sync|chore)/SPEC-' $E/ac009-base-mg.md > $E/ac009-base-mg-prefix-raw.txt
sort -u -o $E/ac009-base-mg-prefix.txt $E/ac009-base-mg-prefix-raw.txt
diff $E/ac009-base-sw-prefix.txt $E/ac009-base-mg-prefix.txt > $E/ac009-base-prefix.diff
# 기대: exit 1 — 기준 트리 spec-workflow 블록 {feat/SPEC-, plan/SPEC-} vs manager-git 절 {feat/SPEC-}
```

plan 작성 시점 기준 트리 측정: manager-git 절은 `moai cc -w` 1(92행 탐지 단서 문장)·`git push -u origin` 1·`gh pr create` 1, 나머지 표지 0. spec-workflow 블록·spec-assembly 절·delivery 절은 모든 표지 0.

판정(로컬·템플릿 각각):

```bash
awk '/^### Late-Branch Invocation Pattern/{s=1; print; next} s && /^##/{exit} s' .claude/agents/moai/manager-git.md > $E/ac009-mg.md
awk '/^\[ZONE:Frozen\] \[HARD\] Step ordering rules/{s=1} s && /^\[SHOULD\] Anti-patterns/{exit} s' .claude/rules/moai/workflow/spec-workflow.md > $E/ac009-sw.md
awk '/^#### Late-branch Pre-check/{s=1; print; next} s && /^####/{exit} s' .claude/skills/moai/workflows/plan/spec-assembly.md > $E/ac009-sa.md
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac009-dl.md
test -s $E/ac009-mg.md
test -s $E/ac009-sw.md
test -s $E/ac009-sa.md
test -s $E/ac009-dl.md
# 기대: 모두 exit 0 — 절이 비면 판정 불가
/usr/bin/grep -c 'moai cc -w' $E/ac009-mg.md > $E/ac009-mg-ccw.txt
/usr/bin/grep -c 'EnterWorktree(' $E/ac009-mg.md > $E/ac009-mg-enter.txt
/usr/bin/grep -c '[g]it branch -m' $E/ac009-mg.md > $E/ac009-mg-rename.txt
/usr/bin/grep -c '[g]it push -u origin' $E/ac009-mg.md > $E/ac009-mg-push.txt
/usr/bin/grep -c 'gh pr create' $E/ac009-mg.md > $E/ac009-mg-pr.txt
/usr/bin/grep -c -- '--<merge_method>' $E/ac009-mg.md > $E/ac009-mg-merge.txt
/usr/bin/grep -c 'ExitWorktree' $E/ac009-mg.md > $E/ac009-mg-exit.txt
/usr/bin/grep -c '[g]it worktree remove' $E/ac009-mg.md > $E/ac009-mg-remove.txt
# 기대: 여덟 값 모두 1 이상
/usr/bin/grep -c -E 'moai cc -w|EnterWorktree\(' $E/ac009-sw.md $E/ac009-sa.md > $E/ac009-entry.txt
# 기대: 파일마다 1 이상
/usr/bin/grep -c 'ExitWorktree' $E/ac009-sw.md $E/ac009-dl.md > $E/ac009-exit.txt
# 기대: 파일마다 1 이상
/usr/bin/grep -c '[g]it worktree remove' $E/ac009-sw.md > $E/ac009-sw-remove.txt
# 기대: 1 이상
/usr/bin/grep -o -E '(plan|feat|sync|chore)/SPEC-' $E/ac009-sw.md > $E/ac009-sw-prefix-raw.txt
sort -u -o $E/ac009-sw-prefix.txt $E/ac009-sw-prefix-raw.txt
/usr/bin/grep -o -E '(plan|feat|sync|chore)/SPEC-' $E/ac009-mg.md > $E/ac009-mg-prefix-raw.txt
sort -u -o $E/ac009-mg-prefix.txt $E/ac009-mg-prefix-raw.txt
diff $E/ac009-sw-prefix.txt $E/ac009-mg-prefix.txt > $E/ac009-prefix.diff
# 기대: exit 0 (어느 쪽 PR 수로 정해지든 두 집합이 같아야 한다 — spec.md §C.5 B1)
```

읽기 단계: 네 절 파일을 읽어 THEN 셋째 항목을 `$E/ac009-reading.md` 에 기록한다. 이 기준은 PR 수(§C.5 B1)를 정하지 않는다. 정해진 값과 무관하게 두 문서가 같은 접두 집합을 쓰는지만 본다.

### AC-GDP-010 — 워크트리 흐름·블록 밖 자리·Frozen 수정 (OD-1 = 1)

```
GIVEN 편집 뒤의 범위 파일, 헌법 레지스트리, progress 기록
WHEN 검사하면
THEN 범위 지침 파일(로컬·템플릿)에 맨손 git worktree add 가 없고
 AND delivery.md Step 3.3.5 절에 git branch -d 명령이 없고
 AND manager-git.md 의 Checkpoint 절과 Synchronization 절이 되돌리기와 pull 을 워크트리 안에서만 실행한다고 말하고(읽기 단계)
 AND 이 트리에서 빌드한 moai 로 헌법 검증을 돌리면 status ok, drift_count 0 이고
 AND 레지스트리 두 사본이 바이트 동일하고
 AND progress 기록 §E.2 에 Frozen 구역 수정(CONST-V3R5-027, 해당하면 028)과 Implementation Kickoff Approval 기록이 있다
```

대조(기준 트리):

```bash
/usr/bin/grep -n '[g]it worktree add' $T/.claude/rules/moai/workflow/main-checkout-branch-guard.md > $E/ac010-control-add.txt
# 기대: exit 0 (38행 — 검출기가 형태를 잡음)
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' $E/base-delivery.md > $E/ac010-base-dl.md
/usr/bin/grep -c '[g]it branch -d' $E/ac010-base-dl.md > $E/ac010-base-branchd.txt
# 기대: 1 (기준 트리 328행)
/usr/bin/grep -c -F 'Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step' $E/base-spec-workflow.md > $E/ac010-base-clause027.txt
/usr/bin/grep -c -F 'Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged' $E/base-spec-workflow.md > $E/ac010-base-clause028.txt
# 기대: 각각 1 — 검증기가 부분 문자열로 찾는 clause 가 기준 트리에 있음
sed -n '/^## §E.2/,/^## §E.3/p' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/progress.md > $E/ac010-base-e2.md
/usr/bin/grep -c 'CONST-V3R5-027' $E/ac010-base-e2.md > $E/ac010-base-e2-const.txt
# 기대: 0 — run-phase 시작 전 §E.2 는 비어 있음
```

plan 작성 시점 기준선: 설치 빌드(`v3.2.0-rc.5`, `84fa4ece4`)의 `moai constitution validate --format json` → `status: ok`, `drift_count: 0`. 이 트리에서 빌드한 바이너리가 아니므로 run-phase가 다시 잰다.

판정:

```bash
/usr/bin/grep -n '[g]it worktree add' .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac010-add.txt
# 기대: exit 1
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac010-dl.md
/usr/bin/grep -c '[g]it branch -d' $E/ac010-dl.md > $E/ac010-branchd.txt
# 기대: 0
go run ./cmd/moai constitution validate --format json > $E/ac010-validate.json 2> $E/ac010-validate.err
# 기대: exit 0, JSON 의 status "ok" 와 drift_count 0 을 읽어 기록 (VCI §2.2: 이 트리에서 빌드한 도구로 측정)
diff .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac010-registry.diff
# 기대: exit 0
sed -n '/^## §E.2/,/^## §E.3/p' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/progress.md > $E/ac010-e2.md
/usr/bin/grep -c 'CONST-V3R5-027' $E/ac010-e2.md > $E/ac010-e2-const.txt
/usr/bin/grep -c -i 'kickoff' $E/ac010-e2.md > $E/ac010-e2-kickoff.txt
# 기대: 각각 1 이상
```

읽기 단계: `manager-git.md` 의 `## Checkpoint System` 절과 `## Synchronization` 절을 읽어, 되돌리기(`reset --hard`)와 `pull` 이 워크트리 안에서만(되돌리기는 사용자 확인 뒤) 실행된다고 적혀 있는지 `$E/ac010-reading.md` 에 기록한다(spec.md §C.3).

### AC-GDP-011 — **[철회]**

OD-1 = 선택지 1 채택(2026-09-10, 운영자·리드 경유)으로 선택지 2(조건부 유지) 경로의 기준을 철회한다. 번호는 추적을 위해 유지하며 판정 대상이 아니다.

### AC-GDP-012 — **[철회]**

OD-1 = 선택지 1 채택(2026-09-10, 운영자·리드 경유)으로 선택지 3(`main_late_branch` 은퇴) 경로의 기준을 철회한다. 번호는 추적을 위해 유지하며 판정 대상이 아니다.

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 여섯 파일의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md, spec-workflow.md, spec-assembly.md, agent-common-protocol.md 는 diff exit 0 이고
 AND delivery.md 와 doc-execution.md 는 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 선택한 미러 테스트 3개가 각자 최상위 PASS 줄을 내며 exit 0 이다
```

`manager-git.md`·`agent-common-protocol.md`·`doc-execution.md` 는 미러 테스트의 바이트 동일 대상이 아니다(`rule_template_mirror_test.go`; `doc-execution` 은 파일에 이름이 없음, plan 작성 시점 `grep` exit 1). 이 세 파일의 사본 일치를 지키는 것은 아래 `diff` 뿐이다.

대조: 기준 트리에서 `delivery.md` 두 사본의 `diff` 는 exit 1, `doc-execution.md` 두 사본의 `diff` 는 exit 1(`138,143d137`, 본문 6줄). 본문 비교 검출기는 빈 파일과 차이 본문을 `diff` 하면 exit 1.

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
diff $E/base-doc-execution.md $E/base-doc-execution-template.md > $E/ac013-docexec-base.diff
diff .claude/skills/moai/workflows/sync/doc-execution.md $T/.claude/skills/moai/workflows/sync/doc-execution.md > $E/ac013-docexec-post.diff
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-docexec-base.diff > $E/ac013-docexec-base.body
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-docexec-post.diff > $E/ac013-docexec-post.body
diff $E/ac013-docexec-base.body $E/ac013-docexec-post.body > $E/ac013-docexec-body.diff
# 기대: exit 0
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0
/usr/bin/grep -c -E '^--- PASS: (TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 정확히 3
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

M1이 템플릿 `manager-git.md:156` 을, M2가 `:114` 를, M4가 절차 블록을 항상 고치므로 첫 점검 RED는 반드시 나온다.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, CLAUDE.local 참조가 추가 줄에 없고
 AND 7~40자 16진 낱말 가운데 a-f 문자를 담은 것이 없으며
 AND 숫자로만 된 7~40자 낱말은 목록으로 뽑혀 읽기 단계에서 커밋 SHA가 아님이 기록된다
```

REQ-GDP-015의 프로그래밍 언어 편향 절은 기계 판정이 없다. 추가 줄을 읽어 16개 지원 프로그래밍 언어 중 하나를 앞세운 표현이 없는지 기록하고, CI의 `template-neutrality-check` 결과를 함께 적는다. `TestTemplateNoInternalContentLeak` 는 CI 보조 방어선으로 남기되 이 기준은 그 테스트에 기대지 않는다(spec.md §E.1).

SHA 검출은 줄이 아니라 낱말 단위다. perl 뒤돌아보기·앞보기로 경계 문자를 소비하지 않으므로 한 줄에 SHA가 여러 개 붙어 있어도 모두 뽑힌다. 2회차 검출식(`[0-9]*[a-f][0-9a-f]{6,39}`)은 첫 문자 a-f 뒤에 6자 이상을 요구해 `538684c47`·`7374b183e`·`2213871af`·`980ccdc56` 을 놓쳤다.

대조(검출기가 형태를 잡는지):

```bash
/usr/bin/grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-specid.txt
/usr/bin/grep -c -E 'REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-req.txt
/usr/bin/grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-date.txt
# 기대: 세 값 모두 1 이상
perl -ne 'while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-sha-tokens.txt
/usr/bin/grep -c -E ':980ccdc56$' $E/ac015-control-sha-tokens.txt > $E/ac015-control-sha-980.txt
# 기대: 1 이상 — 2회차 검출식이 놓친 SHA를 낱말 단위로 잡음. 전체 낱말 수는 progress.md §E.1 에 기록
perl -e 'print "+++ b/file.md\n+sha 538684c47 here\n+one line 7374b183e 2213871af 980ccdc56 b412f8a33\n+all digits 647460835 and 1000000 here\n context 02aca7afe not added\n"' > $E/ac015-fixture.diff
perl -ne 'next unless /^\+(?!\+\+)/; while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- $E/ac015-fixture.diff > $E/ac015-fixture-tokens.txt
/usr/bin/grep -c -v -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-letter.txt
# 기대: 5 — 추가 줄의 a-f 포함 SHA 다섯 개 모두(한 줄에 붙은 네 개 포함)
/usr/bin/grep -c -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-digits.txt
# 기대: 2 — 숫자로만 된 낱말(647460835, 1000000)은 읽기 목록으로 분리. 추가 줄이 아닌 02aca7afe 는 뽑히지 않음
```

판정:

```bash
git diff $BASE -- $T/ > $E/ac015-template.diff
/usr/bin/grep -n -E '^[+][^+].*SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-specid.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' $E/ac015-template.diff > $E/ac015-req.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*20[0-9]{2}-[0-9]{2}-[0-9]{2}' $E/ac015-template.diff > $E/ac015-date.txt
# 기대: exit 1
/usr/bin/grep -n -E '^[+][^+].*CLAUDE[.]local' $E/ac015-template.diff > $E/ac015-local-ref.txt
# 기대: exit 1
perl -ne 'next unless /^\+(?!\+\+)/; while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- $E/ac015-template.diff > $E/ac015-hex-tokens.txt
/usr/bin/grep -v -E ':[0-9]+$' $E/ac015-hex-tokens.txt > $E/ac015-sha-letter.txt
test -s $E/ac015-sha-letter.txt
# 기대: exit 1. 비어 있지 않으면 낱말마다 읽어 커밋 SHA인지 판정 — SHA면 FAIL, 16진 영단어 같은 비-SHA면 사유를 기록
/usr/bin/grep -E ':[0-9]+$' $E/ac015-hex-tokens.txt > $E/ac015-sha-digits.txt
# 읽기 단계: 숫자로만 된 낱말마다 커밋 SHA가 아닌지(개수·크기 값 등) $E/ac015-reading.md 에 기록
```

검출 한계: 대문자 16진(`ABCDEF`)은 잡지 않는다. 7자 미만 약식 SHA와, 영숫자에 바로 붙은 16진 낱말(`abc980ccdc56`, `v1234567a`)도 잡지 않는다. 숫자로만 된 SHA는 자동으로 FAIL시키지 않고 읽기 목록으로만 넘긴다.

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
git log --format=%H <ac016-acp-commit.txt 의 SHA>..HEAD -- .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac016-after.txt
# 기대: 빈 파일
```

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. 한 줄에 두 번 나오는 명령은 1로 센다 — AC-GDP-004의 "2 이상" 은 343·355행이 서로 다른 줄이라는 사실에 기댄다. SHA는 AC-GDP-015에서 낱말 단위로 센다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. 추출 파일의 마지막 줄이 기대한 끝 제목인지 확인해 기록한다. `awk` 절 추출은 시작 표지가 사라지면 빈 파일을 낸다 — 빈 파일은 판정 불가로 기록한다.
- 철회된 기준(AC-GDP-011, 012)은 판정하지 않고 N/A로 기록한다.
- `delivery.md`·`doc-execution.md` 편집으로 줄 수가 바뀌면 사본 차이 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다). 낱말 경계가 필요한 SHA 검출은 perl 뒤돌아보기·앞보기로 한다.
- `[^-]` 로 변경형 `git branch` 를 잡는 검출식은 산문 속 "`git branch`" 에도 걸린다. 그런 적중은 AC-GDP-007 장부에서 `narrative` 로 분류한다.

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다.
- `go test` 선택 실행이 최상위 PASS 줄 3개를 내지 않으면 합격이 아니다.
- 로컬 전체 스위트는 돌리지 않는다. 전체 판정은 통합 뒤 CI가 한다.
- 도구 측정(헌법 검증)은 이 트리에서 빌드한 바이너리로 한다(`go run ./cmd/moai …`).

## §D.4 완료 정의 (Definition of Done)

- 활성 기준 AC-GDP-001~010, 013~015가 PASS이고 AC-GDP-016이 PASS 또는 사유 기록. AC-GDP-011·012는 N/A.
- spec.md §C.5 B1~B3이 run-phase 착수 전에 정해지고 그 결정이 progress 기록에 남는다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
