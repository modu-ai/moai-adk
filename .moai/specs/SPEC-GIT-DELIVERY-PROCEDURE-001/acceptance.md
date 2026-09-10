# Acceptance — SPEC-GIT-DELIVERY-PROCEDURE-001

> 기준마다 Given-When-Then과 판정 명령을 둔다. 모든 부재 기준은 같은 검출기로 알려진 적중을 먼저 잡는 **양성 대조**를 가진다. 검증 출력은 파일로 보내고 exit code를 파이프 없이 따로 읽는다(`| head`·`| tail`·`| grep` 금지).

## 공통 변수와 명령 관례

```bash
BASE=b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0   # R1 고정: 2026-09-10 이 워크트리에서 해석
E=.moai/reports/t622/run                         # 추적 증거 경로
T=internal/template/templates                    # 템플릿 루트
TB=/tmp/t622-moai-tree                           # 이 트리에서 빌드한 moai (AC-GDP-020·022)
S=<세션 스크래치 디렉터리>                         # 트리 밖 임시 경로 (대조용 프로젝트 사본)
```

- 명령은 워크트리 루트에서 실행한다. "exit" 는 직전 명령의 exit code를 `echo "exit=$?"` 로 따로 기록한 값이다.
- 기준 트리 사본은 사전 점검에서 `git show $BASE:<경로> > $E/base-<이름>` 으로 반출해 둔다(대상 목록은 plan.md §C 2단계).
- 검출식 안의 `git` 은 `[g]it` 으로 쓴다. 같은 이유로 `parallel` 은 `para[l]lel` 로, perl 코드 안의 `git` 은 `\x67it` 로 쓴다. 셸 변수를 받는 `sed`·`perl` 은 가드가 거부할 수 있으므로 경로를 글자 그대로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 개수가 찍히지 않은 결과는 판정 불가로 기록한다.
- 기준 트리 측정: plan 작성 시점 이 워크트리 HEAD(`880c0c702`)의 범위 파일은 `$BASE` 와 바이트 동일하다(`git diff --stat $BASE HEAD -- <범위 파일, zone-registry.md L·T, internal/constitution, internal/cli/constitution.go>` 출력 없음, exit 0). 아래 "plan 작성 시점 측정" 값은 이 템플릿 사본으로 쟀다.

## §D AC 표

| AC ID | REQ | 등급 | 상태 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 활성 | `manager-git.md` 동기화 절의 문단·목록 묶음에서 fetch 가 rev-list 와 같은 배치로 묶이지 않고 순서가 지시됨 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 활성 | Pre-Spawn 코드 블록에 단독 fetch·단독 rev-list 줄이 없고, 이어 붙인 줄 1개, rev-list 1개 |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 활성 | Pre-Edit Sync Check 절 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 활성 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 활성 | 범위 파일 전체에서 `--squash` 고정 `gh pr merge` 가 기본값 설명 문장뿐 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | 활성 | `delivery.md`·`doc-execution.md` 에 워크트리 기본 병합 문구가 없고 `manager-git.md` 를 기준으로 밝힘 |
| AC-GDP-007 | REQ-GDP-007 | MUST-PASS | 활성 | 금지 집합 명령 검출식의 적중이 모두 분류되고 primary checkout 실행 안내가 0 |
| AC-GDP-008 | REQ-GDP-008 | MUST-PASS | 활성 | 파일별 참조 하한, 형식·느슨한 개수 일치, 제목 조회 |
| AC-GDP-009 | REQ-GDP-009 | MUST-PASS | 활성 (개편) | 네 지침 절이 §C.2 흐름의 명령을 담고, PR 브랜치 접두 합집합이 정확히 `feat/SPEC-` 하나 |
| AC-GDP-010 | REQ-GDP-010 | MUST-PASS | 활성 (개편) | 맨손 `git worktree add` 부재, `manager-git.md` 42·160 처리, Frozen 줄 수정 기록과 Kickoff 기록 |
| AC-GDP-011 | REQ-GDP-011 | — | **철회** | OD-1 선택지 2 경로 |
| AC-GDP-012 | REQ-GDP-012 | — | **철회** | OD-1 선택지 3 경로 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 활성 | 사본 일치, `delivery.md`·`doc-execution.md` 의도된 차이 보존, 미러 테스트 3개 실행·통과 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 활성 | `.toml` 재생성과 `agents-emit-check` exit 0 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 활성 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 활성 | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |
| AC-GDP-017 | REQ-GDP-016 | MUST-PASS | 활성 (신규) | 단계별 PR 검출식의 적중이 모두 분류되고 단계별 PR 안내가 0 |
| AC-GDP-018 | REQ-GDP-018 | MUST-PASS | 활성 (신규) | `delivery.md` Step 3.2 github-flow가 워크트리 안의 PR 브랜치만 전달하고 primary checkout 비-main 브랜치에서 멈춤 |
| AC-GDP-019 | REQ-GDP-019 | MUST-PASS | 활성 (신규) | Step 3.3.5가 `ExitWorktree`·병합 착지 뒤 폐기를 담고 브랜치 명령·"return to base branch" 가 없음 |
| AC-GDP-020 | REQ-GDP-020 | MUST-PASS | 활성 (신규) | amend 명령 원문·근거 문서·dry-run 출력·운영자 실행 출력이 추적 경로에 있음 |
| AC-GDP-021 | REQ-GDP-017 | MUST-PASS | 활성 (신규) | 서술형 검출식의 적중이 모두 분류되고 primary checkout feature 브랜치 안내가 0 |
| AC-GDP-022 | REQ-GDP-022 | MUST-PASS | 활성 (신규) | 트리 빌드를 경로로 호출한 `constitution validate` 가 ok·drift 0, 빌드 커밋 = HEAD |
| AC-GDP-023 | REQ-GDP-023 | MUST-PASS | 활성 (신규) | `zone-registry.md` L·T 바이트 동일, 두 사본 모두 새 clause 1회·이전 clause 0회 |
| AC-GDP-024 | REQ-GDP-021 | MUST-PASS | 활성 (신규) | 기록된 amend 명령에 표준입력 주입이 없고, 운영자 대화형 실행과 레인의 에스컬레이션 기록이 있음 |

AC-GDP-016은 요구사항 추적 밖의 절차 점검이다.

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

읽기 단계: `ac001-local-paras.md` 의 문단마다 THEN (1)·(2)를 판정해 `$E/ac001-reading.md` 에 기록한다. 템플릿 사본과 생성물 `$T/.codex/agents/moai/manager-git.toml` 에도 같은 자동 실패 검사를 실행한다.

뮤턴트 재실행(plan 작성 시점, 세션 스크래치 파일; 기준 트리 값은 0.1.3 작성 때 다시 잼):

```
기준 트리 절                          → 문단 1, autofail 1, listgroup 0
"`git fetch` then … all in parallel"  → 문단 1, autofail 1 → FAIL
2회차 목록 뮤턴트                      → 문단 1, autofail 0, listgroup 1 → 읽기 단계 (2)에서 FAIL
2회차 올바른 문장                      → 문단 1, autofail 0, listgroup 0 → 빨강 아님
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
# 기대: 1줄
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-base-block.md > $E/ac002-base-b.txt
# 기대: exit 0
/usr/bin/grep -c -E '^[[:space:]]*[g]it fetch origin main 2>&1[[:space:]]*(;|&&)[[:space:]]*[g]it rev-list --count --left-right origin/main[.][.][.]HEAD' $E/ac002-base-block.md > $E/ac002-base-c.txt
# 기대: 0
/usr/bin/grep -c 'rev-list' $E/ac002-base-block.md > $E/ac002-base-d.txt
# 기대: 1
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
# 기대: exit 0 (기준 트리의 표 행은 8개)
```

plan 작성 시점 측정(0.1.3): 기준 트리 블록 9줄, rev-list 1, 표 행 8. 뮤턴트: 이어 붙인 줄 + 두 번째 `git -C . rev-list` 줄 → rev-list 2 → FAIL; 올바른 픽스처 → a 0, b exit 1, c 1, rev-list 1 → PASS.

### AC-GDP-003 — Pre-Edit Sync Check 절 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절(제목부터 #### The sweep prohibition 앞까지)
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조: AC-GDP-002의 `ac002-base-section.md` 와 `ac002-local-section.md` 를 `diff` 하면 exit 1.

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

```bash
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' $E/base-delivery.md > $E/ac004-control-squash.txt
# 대조 기대: 2 (343·355행)
/usr/bin/grep -n 'merge_method' $E/base-delivery.md > $E/ac004-control-source.txt
# 대조 기대: exit 1
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

```bash
/usr/bin/grep -n -E 'gh pr merge[^|]*--squash' $E/base-manager-git.md $E/base-delivery.md $E/base-manager-git.toml > $E/ac005-control.txt
# 대조 기대: 6줄 — manager-git 32·114, delivery 343·355, .toml 26·108
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/agents/moai/manager-git.md > $E/ac005-mg.txt
# 기대: 1
/usr/bin/grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-mg-default.txt
# 기대: 1
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/skills/moai/workflows/sync/delivery.md > $E/ac005-delivery.txt
# 기대: 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac005-others.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $T/.codex/agents/moai/manager-git.toml > $E/ac005-toml.txt
# 기대: 1
```

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
# 기대: exit 0 — 절 안 8행·20행
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default' $E/ac006-base-de.md > $E/ac006-base-de-default.txt
# 기대: exit 0 — 절 안 8행
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-base-dl.md $E/ac006-base-de.md > $E/ac006-base-source.txt
# 기대: 파일마다 0
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
# 기대: 둘 다 exit 0
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

읽기 단계: 두 절 파일을 읽어 `$E/ac006-reading.md` 에 기록한다. plan 작성 시점 측정(0.1.3): delivery 절 52줄(적중 8·20행), doc-execution 소절 9줄(적중 8행), `manager-[g]it[.]md` 0·0.

### AC-GDP-007 — 금지 집합 명령 검출식의 분류 장부

```
GIVEN 범위 네 파일(로컬·템플릿)
WHEN AGENTS.md §2 금지 집합 전체를 잡는 검출식으로 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND unconditioned-primary-instruction 분류가 0행이다
```

검출식(대조·판정 공통):

```
[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])
```

대조(기준 트리):

```bash
git grep -n -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $BASE -- $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-base-hits.txt
# 기대: exit 0, 20줄 — manager-git 10(42·96·110·119·121·122·125·127·137·160),
#   spec-workflow 5(50·56·58·59·62), spec-assembly 1(336), delivery 4(276·323·324·328)
```

뮤턴트 재실행(plan 작성 시점): 2회차 뮤턴트 8줄 중 1-7줄 적중, 8줄 `reset --keep` 은 AGENTS.md §2 목록 밖이라 미적중.

판정:

```bash
/usr/bin/grep -n -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac007-post-hits.txt
# exit 0 이면 적중 줄마다 장부 행을 만든다. exit 1 이면 장부는 "적중 0" 한 줄
```

장부 `.moai/reports/t622/ac01-classification.md` 분류값: `prohibition-context`, `narrative`, `worktree-internal`, `unconditioned-primary-instruction`. 합격 조건은 마지막 분류 0행. 서술형 안내는 AC-GDP-021이, 단계별 PR 안내는 AC-GDP-017이 따로 판정한다.

### AC-GDP-008 — 절차 참조가 끊기지 않음

```
GIVEN spec-assembly.md, spec-workflow.md, delivery.md
WHEN (1) 형식을 갖춘 manager-git.md 절 참조를 파일마다 세고
 AND (2) 대소문자를 가리는 느슨한 검출식으로 파일마다 세고
 AND (3) 형식을 갖춘 참조의 제목을 manager-git.md 제목 줄에서 조회하면
THEN spec-assembly.md 와 spec-workflow.md 의 (1) 개수가 각각 1 이상이고
 AND 파일마다 (1) 개수와 (2) 개수가 같고
 AND (3) 조회에서 제목마다 대상이 정확히 한 줄 있다
```

참조가 0건인 경우는 PASS가 될 수 없다. 느슨한 검출식에 `-i` 를 쓰지 않는 이유: 기준 트리 `spec-workflow.md` 17행의 에이전트 목록 줄이 `-i` 형태에 걸려 느슨한 개수가 2가 된다.

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
# 기대: 0, exit 1
perl -CSD -pe 's/ \x{a7} Late-Branch Invocation Pattern\./ \x{a7} late-branch invocation pattern./' -- $E/base-spec-workflow.md > $E/ac008-mutant-lower.md
perl -CSD -pe 's/see `\.claude\/agents\/moai\/manager-\x67it\.md` \x{a7} Late-Branch Invocation Pattern\./see the \x67it agent definition, Late-branch section./' -- $E/base-spec-workflow.md > $E/ac008-mutant-prose.md
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' $E/ac008-mutant-lower.md $E/ac008-mutant-prose.md > $E/ac008-control-mutants.txt
# 기대: 파일마다 0 — 하한 1 미만이므로 두 뮤턴트 모두 FAIL
```

판정:

```bash
/usr/bin/grep -c -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-strict-count.txt
# 기대: spec-assembly.md ≥ 1, spec-workflow.md ≥ 1
/usr/bin/grep -c -E 'manager-[g]it[^§]*§ [A-Z]|Invocation Pattern' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-loose-count.txt
diff $E/ac008-strict-count.txt $E/ac008-loose-count.txt > $E/ac008-count.diff
# 기대: exit 0
/usr/bin/grep -n -o -E 'manager-[g]it[.]md` § [A-Z][A-Za-z-]*( [A-Z][A-Za-z-]*)*' .claude/skills/moai/workflows/plan/spec-assembly.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac008-refs.txt
sed -n -e 's/.*§ /### /p' -- $E/ac008-refs.txt > $E/ac008-headings-raw.txt
sort -u -o $E/ac008-headings.txt $E/ac008-headings-raw.txt
wc -l < $E/ac008-headings.txt > $E/ac008-heading-count.txt
/usr/bin/grep -c -x -F -f $E/ac008-headings.txt .claude/agents/moai/manager-git.md > $E/ac008-lookup.txt
# 기대: ac008-lookup.txt 의 값 = ac008-heading-count.txt 의 값, 두 값 모두 1 이상
```

### AC-GDP-009 — 네 지침 절의 흐름 일치와 단일 PR 접두 (OD-1 = 1, B1)

```
GIVEN 네 지침의 절: manager-git.md "### Late-Branch Invocation Pattern" 절,
      spec-workflow.md "[ZONE:Frozen] [HARD] Step ordering rules" 블록,
      spec-assembly.md "#### Late-branch Pre-check" 절, delivery.md "#### Step 3.3.5" 절
WHEN 편집 뒤 절을 추출해 §C.2 흐름의 명령 표지를 세고 PR 브랜치 접두의 합집합을 구하면
THEN 아래 표의 필수 표지가 절마다 1회 이상 있고
 AND spec-workflow.md 전체, manager-git.md 절, spec-assembly.md 절, delivery.md 전체에서 모은
     (plan|feat|feature|sync|chore)/SPEC- 접두의 합집합이 정확히 {feat/SPEC-} 하나이고
 AND 읽기 단계에서, 네 절이 진입 → PR → 종결 순서를 §C.2와 같게 서술하고 primary checkout에서 브랜치 명령을 실행하라고 하지 않는다
```

`spec-assembly.md` 388행(`--branch` 경로의 `feature/SPEC-{ID}`)과 `manager-git.md` 82·144행은 절 밖 형제라 합집합에서 뺀다(spec.md §D).

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
/usr/bin/grep -h -o -E '(plan|feat|feature|sync|chore)/SPEC-' $E/base-spec-workflow.md $E/ac009-base-mg.md $E/ac009-base-sa.md $E/base-delivery.md > $E/ac009-base-prefix-raw.txt
sort -u -o $E/ac009-base-prefix.txt $E/ac009-base-prefix-raw.txt
printf 'feat/SPEC-\n' > $E/ac009-expected.txt
diff $E/ac009-expected.txt $E/ac009-base-prefix.txt > $E/ac009-base-prefix.diff
# 기대: exit 1 — 기준 트리 합집합 {chore/SPEC-, feat/SPEC-, plan/SPEC-, sync/SPEC-}
```

plan 작성 시점 측정(0.1.3): manager-git 절 43줄(`moai cc -w` 1·`push -u origin` 1·`gh pr create` 1, 나머지 표지 0), spec-workflow 블록 15줄·spec-assembly 절 10줄·delivery 절 11줄 모든 표지 0. 접두 토큰: spec-workflow 전체 chore 1·feat 5·plan 3·sync 1, manager-git 절 feat만, spec-assembly 절 feat 1, delivery 전체 chore 1·sync 1(50행). 픽스처: `feat/SPEC-XXX` 만 담은 파일 → 합집합 diff exit 0; 여기에 `sync/SPEC-XXX` 한 줄을 더한 뮤턴트 → diff exit 1(`> sync/SPEC-`).

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
# 기대: 모두 exit 0
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
/usr/bin/grep -h -o -E '(plan|feat|feature|sync|chore)/SPEC-' .claude/rules/moai/workflow/spec-workflow.md $E/ac009-mg.md $E/ac009-sa.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac009-prefix-raw.txt
sort -u -o $E/ac009-prefix.txt $E/ac009-prefix-raw.txt
diff $E/ac009-expected.txt $E/ac009-prefix.txt > $E/ac009-prefix.diff
# 기대: exit 0
```

읽기 단계: 네 절 파일을 읽어 THEN 셋째 항목을 `$E/ac009-reading.md` 에 기록한다.

### AC-GDP-010 — 워크트리 흐름·manager-git 42/160·Frozen 줄 수정 기록

```
GIVEN 편집 뒤의 범위 파일과 progress 기록
WHEN 검사하면
THEN 범위 지침 파일(로컬·템플릿)에 맨손 git worktree add 가 없고
 AND manager-git.md 의 Checkpoint 절과 Synchronization 절이 되돌리기와 pull 을 워크트리 안에서만 실행한다고 말하고(읽기 단계)
 AND progress 기록 §E.2 에 spec-workflow.md 의 바뀐 [ZONE:Frozen] 줄 전부(등록된 CONST-V3R5-027·028, 등록되지 않은 23행 블록·166행)와 Implementation Kickoff Approval 기록이 있다
```

대조(기준 트리):

```bash
/usr/bin/grep -n '[g]it worktree add' $T/.claude/rules/moai/workflow/main-checkout-branch-guard.md > $E/ac010-control-add.txt
# 기대: exit 0 (38행)
sed -n '/^## §E.2/,/^## §E.3/p' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/progress.md > $E/ac010-base-e2.md
/usr/bin/grep -c 'CONST-V3R5-027' $E/ac010-base-e2.md > $E/ac010-base-e2-const.txt
# 기대: 0 — run-phase 시작 전 §E.2 는 비어 있음
```

판정:

```bash
/usr/bin/grep -n '[g]it worktree add' .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/workflow/spec-workflow.md $T/.claude/skills/moai/workflows/plan/spec-assembly.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac010-add.txt
# 기대: exit 1
sed -n '/^## §E.2/,/^## §E.3/p' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/progress.md > $E/ac010-e2.md
/usr/bin/grep -c 'CONST-V3R5-027' $E/ac010-e2.md > $E/ac010-e2-027.txt
/usr/bin/grep -c 'CONST-V3R5-028' $E/ac010-e2.md > $E/ac010-e2-028.txt
/usr/bin/grep -c -i 'unregistered' $E/ac010-e2.md > $E/ac010-e2-unregistered.txt
/usr/bin/grep -c -i 'kickoff' $E/ac010-e2.md > $E/ac010-e2-kickoff.txt
# 기대: 네 값 모두 1 이상
```

읽기 단계: `manager-git.md` 의 `## Checkpoint System`·`## Synchronization` 절과 progress §E.2의 Frozen 기록(줄마다 이전·이후 문장)을 `$E/ac010-reading.md` 에 기록한다.

### AC-GDP-011 — **[철회]**

OD-1 = 선택지 1 채택(2026-09-10, 운영자·리드 경유)으로 선택지 2 경로의 기준을 철회한다. 번호는 추적을 위해 유지하며 판정 대상이 아니다.

### AC-GDP-012 — **[철회]**

OD-1 = 선택지 1 채택(2026-09-10, 운영자·리드 경유)으로 선택지 3 경로의 기준을 철회한다. 번호는 추적을 위해 유지하며 판정 대상이 아니다.

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 여섯 파일의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md, spec-workflow.md, spec-assembly.md, agent-common-protocol.md 는 diff exit 0 이고
 AND delivery.md 와 doc-execution.md 는 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 선택한 미러 테스트 3개가 각자 최상위 PASS 줄을 내며 exit 0 이다
```

레지스트리 사본 일치는 AC-GDP-023이 판정한다. 대조: 기준 트리에서 `delivery.md` 두 사본의 `diff` exit 1(275·278·479-480), `doc-execution.md` 두 사본의 `diff` exit 1(`138,143d137`).

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
# 기대: exit 1
make agents-emit > $E/ac014-emit.txt 2>&1
# 기대: exit 0
make agents-emit-check > $E/ac014-green.txt 2>&1
# 기대: exit 0
git diff --name-only $BASE -- $T/.codex/agents/moai/ > $E/ac014-changed.txt
# 기대: 한 줄, internal/template/templates/.codex/agents/moai/manager-git.toml
```

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, CLAUDE.local 참조가 추가 줄에 없고
 AND 7~40자 16진 낱말 가운데 a-f 문자를 담은 것이 없으며
 AND 숫자로만 된 7~40자 낱말은 목록으로 뽑혀 읽기 단계에서 커밋 SHA가 아님이 기록된다
```

REQ-GDP-015의 프로그래밍 언어 편향 절은 기계 판정이 없다. 추가 줄을 읽어 기록하고 CI의 `template-neutrality-check` 결과를 함께 적는다. §C.4의 `--after` 문장 두 개는 SPEC ID·날짜·SHA를 담지 않도록 쓰였다.

대조:

```bash
/usr/bin/grep -c -E 'SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-specid.txt
/usr/bin/grep -c -E 'REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-req.txt
/usr/bin/grep -c -E '20[0-9]{2}-[0-9]{2}-[0-9]{2}' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-date.txt
# 기대: 세 값 모두 1 이상
perl -ne 'while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/spec.md > $E/ac015-control-sha-tokens.txt
/usr/bin/grep -c -E ':980ccdc56$' $E/ac015-control-sha-tokens.txt > $E/ac015-control-sha-980.txt
# 기대: 1 이상. 전체 낱말 수는 progress.md §E.1 에 기록
perl -e 'print "+++ b/file.md\n+sha 538684c47 here\n+one line 7374b183e 2213871af 980ccdc56 b412f8a33\n+all digits 647460835 and 1000000 here\n context 02aca7afe not added\n"' > $E/ac015-fixture.diff
perl -ne 'next unless /^\+(?!\+\+)/; while (/(?<![0-9A-Za-z])([0-9a-f]{7,40})(?![0-9A-Za-z])/g) { print "$.:$1\n" }' -- $E/ac015-fixture.diff > $E/ac015-fixture-tokens.txt
/usr/bin/grep -c -v -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-letter.txt
# 기대: 5
/usr/bin/grep -c -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-digits.txt
# 기대: 2
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
# 기대: exit 1. 비어 있지 않으면 낱말마다 읽어 커밋 SHA인지 판정
/usr/bin/grep -E ':[0-9]+$' $E/ac015-hex-tokens.txt > $E/ac015-sha-digits.txt
# 읽기 단계: 숫자로만 된 낱말마다 커밋 SHA가 아닌지 $E/ac015-reading.md 에 기록
```

검출 한계: 대문자 16진, 7자 미만 약식 SHA, 영숫자에 바로 붙은 16진 낱말은 잡지 않는다. 숫자로만 된 SHA는 읽기 목록으로만 넘긴다.

### AC-GDP-016 — 항상 로드 규칙 편집이 마지막 (SHOULD, 절차 점검)

```
GIVEN run-phase 커밋들
WHEN agent-common-protocol.md 를 고친 커밋 이후의 커밋을 범위 지침 파일로 거르면
THEN 결과가 비어 있다
```

```bash
git log --format=%H $BASE..HEAD -- .claude/agents/moai/manager-git.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac016-control.txt
# 대조 기대: 1줄 이상
git log --format=%H -1 $BASE..HEAD -- .claude/rules/moai/core/agent-common-protocol.md > $E/ac016-acp-commit.txt
git log --format=%H <ac016-acp-commit.txt 의 SHA>..HEAD -- .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/rules/moai/core/zone-registry.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac016-after.txt
# 기대: 빈 파일
```

### AC-GDP-017 — 단계별 PR 안내의 분류 장부 (B1)

```
GIVEN 범위 네 지침 파일(manager-git.md, spec-workflow.md, spec-assembly.md, delivery.md; 로컬·템플릿)
WHEN 단계별 PR 검출식으로 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND multi-pr-instruction 분류가 0행이다
```

검출식(awk, 대조·판정 공통):

```
/PR per phase|pull request per phase|own pull request|squash commit per phase|[Pp]lan[ -]PR|[Rr]un[ -]PR|[Ss]ync[ -]PR|BOTH run AND sync|run-merge|sync-merge|(plan|sync|chore)\/SPEC-/
```

대조(기준 트리):

```bash
awk '/PR per phase|pull request per phase|own pull request|squash commit per phase|[Pp]lan[ -]PR|[Rr]un[ -]PR|[Ss]ync[ -]PR|BOTH run AND sync|run-merge|sync-merge|(plan|sync|chore)\/SPEC-/ {print FILENAME ":" FNR}' $E/base-manager-git.md $E/base-spec-workflow.md $E/base-spec-assembly.md $E/base-delivery.md > $E/ac017-base-hits.txt
# 기대: 19줄 — spec-workflow 18(26·42·43·44·47·50·53·65·66·319·320·336·360·361·428·433·438·439), delivery 1(50), manager-git 0, spec-assembly 0
```

뮤턴트(plan 작성 시점, 스크래치 픽스처 6줄): "Open a run PR and then a sync PR." · "Merge the plan-PR before running." · "Each phase gets its own pull request." · "Create plan/SPEC-XXX at PR time." · "One PR per SPEC on feat/SPEC-XXX." · "Cleanup waits until BOTH run AND sync PRs land." → 1·2·3·4·6줄 적중, 단일 PR 문장인 5줄은 미적중.

판정(로컬·템플릿 각각):

```bash
awk '/PR per phase|pull request per phase|own pull request|squash commit per phase|[Pp]lan[ -]PR|[Rr]un[ -]PR|[Ss]ync[ -]PR|BOTH run AND sync|run-merge|sync-merge|(plan|sync|chore)\/SPEC-/ {print FILENAME ":" FNR ": " $0}' .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac017-post-hits.txt
wc -l < $E/ac017-post-hits.txt > $E/ac017-post-count.txt
```

장부 `.moai/reports/t622/b1-pr-classification.md` 분류값: `single-pr`(SPEC PR 하나를 말함), `prohibition-context`(단계별 PR을 쓰지 않는다는 문장), `multi-pr-instruction`. 합격 조건: 장부 행 수 = `ac017-post-count.txt`, `multi-pr-instruction` 0행.

### AC-GDP-018 — Step 3.2 github-flow 전달 경로 (B2)

```
GIVEN delivery.md 로컬·템플릿 사본의 "##### Strategy: github-flow" 절
WHEN 편집 뒤 절을 추출해 검사하면
THEN "**Feature branch** (any branch other than main)" 경로 문구가 0회이고
 AND 런처 진입 표지(moai cc -w 또는 EnterWorktree()가 1회 이상이고
 AND "primary checkout" 이 1회 이상이고
 AND AC-GDP-021 장부에서 이 절의 줄이 primary-branch-instruction 으로 분류된 행이 0개이고
 AND 읽기 단계에서, primary checkout의 main 이 아닌 브랜치를 만나면 push와 PR 생성 없이 멈춰 보고한다는 경로가 있고 Route A의 main 직접 push 경로가 남아 있다
```

대조(기준 트리):

```bash
awk '/^##### Strategy: github-flow/{s=1; print; next} s && /^##### /{exit} s' $E/base-delivery.md > $E/ac018-base-gf.md
/usr/bin/grep -c -F '**Feature branch** (any branch other than main)' $E/ac018-base-gf.md > $E/ac018-base-feature.txt
# 기대: 1
/usr/bin/grep -c -E 'moai cc -w|EnterWorktree\(' $E/ac018-base-gf.md > $E/ac018-base-launcher.txt
# 기대: 0
/usr/bin/grep -c -i 'primary checkout' $E/ac018-base-gf.md > $E/ac018-base-primary.txt
# 기대: 0
```

plan 작성 시점 측정(0.1.3): 절 30줄, Feature branch 경로 1, 런처 표지 0, primary checkout 0, `**Worktree context**` 1, `stop and report` 1. 뮤턴트("**PR branch in the primary checkout**: push and create the PR." + "Enter the worktree with moai cc -w <name> first.") → Feature 경로 0·런처 1·primary 1로 세 개수를 모두 통과하지만 AC-GDP-021 서술형 검출식이 3줄을 잡는다 — 개수만으로는 부족해 AC-GDP-021 장부 조건을 함께 둔다.

판정(로컬·템플릿 각각):

```bash
awk '/^##### Strategy: github-flow/{s=1; print; next} s && /^##### /{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac018-gf.md
test -s $E/ac018-gf.md
# 기대: exit 0
/usr/bin/grep -c -F '**Feature branch** (any branch other than main)' $E/ac018-gf.md > $E/ac018-feature.txt
# 기대: 0
/usr/bin/grep -c -E 'moai cc -w|EnterWorktree\(' $E/ac018-gf.md > $E/ac018-launcher.txt
# 기대: 1 이상
/usr/bin/grep -c -i 'primary checkout' $E/ac018-gf.md > $E/ac018-primary.txt
# 기대: 1 이상
```

읽기 단계는 `$E/ac018-reading.md` 에 기록한다. 사본의 의도된 차이 보존은 AC-GDP-013이 판정한다.

### AC-GDP-019 — Step 3.3.5의 결과 (B2)

```
GIVEN delivery.md 로컬·템플릿 사본의 "#### Step 3.3.5" 절
WHEN 편집 뒤 절을 추출해 검사하면
THEN ExitWorktree 가 1회 이상이고
 AND AC-GDP-007 금지 집합 명령 검출식의 적중이 0줄이고
 AND "return to (the) base branch" (대소문자 무시)가 0회이고
 AND 병합 착지 표지(MERGED 또는 merge (has) land)가 1회 이상이고
 AND 읽기 단계에서, 워크트리를 남겨 두고 떠나며 원격 병합 착지 뒤에만 폐기한다고 말하고 WT-* 문장이 남아 있다
```

대조(기준 트리):

```bash
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' $E/base-delivery.md > $E/ac019-base-dl.md
/usr/bin/grep -c 'ExitWorktree' $E/ac019-base-dl.md > $E/ac019-base-exit.txt
# 기대: 0
/usr/bin/grep -c -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $E/ac019-base-dl.md > $E/ac019-base-cmd.txt
# 기대: 3 (323·324·328행)
/usr/bin/grep -c -i -E 'return to (the )?base branch' $E/ac019-base-dl.md > $E/ac019-base-return.txt
# 기대: 2 (제목·321행)
/usr/bin/grep -c -E 'MERGED|merge (has )?land' $E/ac019-base-dl.md > $E/ac019-base-merged.txt
# 기대: 0
```

뮤턴트(plan 작성 시점): 제목을 "Leave the Worktree" 로 바꾸고 "Then run: git switch {main_branch}" 와 "ExitWorktree once the PR is MERGED." 를 담은 픽스처 → ExitWorktree 1·병합 표지 1이지만 명령 검출식 1줄 → FAIL.

판정(로컬·템플릿 각각):

```bash
awk '/^#### Step 3\.3\.5/{s=1; print; next} s && /^####/{exit} s' .claude/skills/moai/workflows/sync/delivery.md > $E/ac019-dl.md
test -s $E/ac019-dl.md
# 기대: exit 0
/usr/bin/grep -c 'ExitWorktree' $E/ac019-dl.md > $E/ac019-exit.txt
# 기대: 1 이상
/usr/bin/grep -c -E '[g]it (-C [^ ]+ )?(checkout|switch|reset --hard|pull|merge( |$)|rebase|stash)|[g]it (-C [^ ]+ )?branch (-[a-zA-Z]*[fDmMcCdtu]|--(force|move|copy|delete|set-upstream-to|unset-upstream)|[^-])' $E/ac019-dl.md > $E/ac019-cmd.txt
# 기대: 0
/usr/bin/grep -c -i -E 'return to (the )?base branch' $E/ac019-dl.md > $E/ac019-return.txt
# 기대: 0
/usr/bin/grep -c -E 'MERGED|merge (has )?land' $E/ac019-dl.md > $E/ac019-merged.txt
# 기대: 1 이상
```

### AC-GDP-020 — 헌법 개정 명령과 근거의 기록 (B3)

```
GIVEN run-phase 의 헌법 개정 단계(plan.md M5)
WHEN 추적 증거 경로를 검사하면
THEN 근거 문서 .moai/reports/t622/run/const-amend-evidence.md 가 추적되고 두 규칙 ID와 --before·--after 문장을 담고
 AND 명령 기록 .moai/reports/t622/run/const-amend-commands.txt 가 레인 dry-run 두 줄과 운영자 실행 두 줄을 원문 그대로 담으며,
     규칙 ID마다 --rule 2회, --dry-run 합계 2회, --evidence 합계 4회, §C.4의 --before·--after 원문이 규칙마다 각각 2회이고
 AND dry-run 출력 두 파일이 각각 "Dry-run success: files were not modified." 와 해당 "Rule ID:" 줄을 담고
 AND 운영자 실행 출력 두 파일(const-amend-027-operator.txt, const-amend-028-operator.txt)이 각각 "Amendment success:" 줄을 담는다
```

**현재 트리에서 마지막 조건의 초록 경로는 막혀 있다.** 적용 단계의 두 도우미가 스텁이라(spec.md §A.7) 운영자가 승인해도 "Amendment success:" 가 나오지 않는다. 이 조건은 spec.md §C.5 B4가 정해지기 전에는 이 작업으로 뒤집을 수 없으며, B4 결정에 따라 이 기준의 마지막 조건을 다시 쓴다.

대조(기준 트리·plan 작성 시점):

```bash
test -e .moai/reports/t622/run/const-amend-commands.txt
# 기대: exit 1 — 증거 없음 (RED-now)
/usr/bin/grep -c -F 'clause: "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac020-control-before027.txt
/usr/bin/grep -c -F 'clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac020-control-before028.txt
# 기대: 파일마다 1 — --before 원문이 현재 등록 clause 와 같음
```

뮤턴트: `--before` 끝의 " at this step" 을 뺀 명령 기록 → 원문 개수 0 → FAIL (CLI도 `constitution.go:528-529` 에서 "clause mismatch" 로 거부).

판정:

```bash
git ls-files --error-unmatch .moai/reports/t622/run/const-amend-evidence.md > $E/ac020-tracked.txt 2>&1
# 기대: exit 0
/usr/bin/grep -c -F -e '--rule CONST-V3R5-027' $E/const-amend-commands.txt > $E/ac020-rule027.txt
/usr/bin/grep -c -F -e '--rule CONST-V3R5-028' $E/const-amend-commands.txt > $E/ac020-rule028.txt
# 기대: 각각 2
/usr/bin/grep -c -F -e '--dry-run' $E/const-amend-commands.txt > $E/ac020-dryrun.txt
# 기대: 2
/usr/bin/grep -c -F -e '--evidence' $E/const-amend-commands.txt > $E/ac020-evidence.txt
# 기대: 4
/usr/bin/grep -c -F -e "--before 'Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step'" $E/const-amend-commands.txt > $E/ac020-before027.txt
/usr/bin/grep -c -F -e "--after 'Step 1 (plan) MUST execute in main checkout on Route A and inside a launcher-entered worktree on Route B'" $E/const-amend-commands.txt > $E/ac020-after027.txt
/usr/bin/grep -c -F -e "--before 'Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged'" $E/const-amend-commands.txt > $E/ac020-before028.txt
/usr/bin/grep -c -F -e "--after 'Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after the single SPEC PR is merged'" $E/const-amend-commands.txt > $E/ac020-after028.txt
# 기대: 네 값 모두 2
/usr/bin/grep -c -F 'Dry-run success: files were not modified.' $E/const-amend-027-dryrun.txt $E/const-amend-028-dryrun.txt > $E/ac020-dryrun-ok.txt
/usr/bin/grep -c -F 'Rule ID: CONST-V3R5-027' $E/const-amend-027-dryrun.txt > $E/ac020-dryrun-027.txt
/usr/bin/grep -c -F 'Rule ID: CONST-V3R5-028' $E/const-amend-028-dryrun.txt > $E/ac020-dryrun-028.txt
# 기대: 파일마다 1
/usr/bin/grep -c -F 'Amendment success:' $E/const-amend-027-operator.txt $E/const-amend-028-operator.txt > $E/ac020-operator-ok.txt
# 기대: 파일마다 1 (B4 결정 전에는 도달 불가)
```

### AC-GDP-021 — 서술형 primary checkout 브랜치 안내의 분류 장부 (B2)

```
GIVEN 범위 다섯 지침 파일(manager-git.md, spec-workflow.md, spec-assembly.md, delivery.md, doc-execution.md; 로컬·템플릿)
WHEN 서술형 검출식으로 적중을 모으고 장부로 분류하면
THEN 장부 행 수가 적중 수와 같고
 AND primary-branch-instruction 분류가 0행이다
```

AC-GDP-007의 명령 검출식은 명령 문구가 없는 문장("continue on the feature branch in main checkout")을 보지 못한다. REQ-GDP-017이 읽기에만 기대지 않도록 이 기준을 둔다.

검출식(awk, 대조·판정 공통):

```
((/[Mm]ain (checkout|working tree)|[Pp]rimary (checkout|working tree)|[Hh]ost checkout|[Ss]hared checkout/) && (/feature branch|feat\/SPEC|plan\/SPEC|sync\/SPEC|chore\/SPEC|branch in|SPEC branch|PR branch/)) || /[Rr]eturn to (the )?base branch/
```

대조(기준 트리):

```bash
awk '((/[Mm]ain (checkout|working tree)|[Pp]rimary (checkout|working tree)|[Hh]ost checkout|[Ss]hared checkout/) && (/feature branch|feat\/SPEC|plan\/SPEC|sync\/SPEC|chore\/SPEC|branch in|SPEC branch|PR branch/)) || /[Rr]eturn to (the )?base branch/ {print FILENAME ":" FNR}' $E/base-manager-git.md $E/base-spec-workflow.md $E/base-spec-assembly.md $E/base-delivery.md $E/base-doc-execution.md > $E/ac021-base-hits.txt
# 기대: 11줄 — spec-workflow 10(21·42·43·50·51·52·192·286·321·429), delivery 1(321)
```

뮤턴트(plan 작성 시점, 스크래치 픽스처 6줄): "Step 2 (run) continues on the feature branch in the primary checkout." · "Run /moai run SPEC-XXX on feat/SPEC-XXX in the host checkout." · "After the PR, return to the base branch." · "Keep working on the SPEC branch in the main working tree." · "Stay on your SPEC branch at the repository root." · "Route A commits directly to main in main checkout." → 1~4줄 적중, 5줄 미적중(검출 한계 — "repository root" 표현), 6줄(Route A) 미적중.

판정(로컬·템플릿 각각):

```bash
awk '((/[Mm]ain (checkout|working tree)|[Pp]rimary (checkout|working tree)|[Hh]ost checkout|[Ss]hared checkout/) && (/feature branch|feat\/SPEC|plan\/SPEC|sync\/SPEC|chore\/SPEC|branch in|SPEC branch|PR branch/)) || /[Rr]eturn to (the )?base branch/ {print FILENAME ":" FNR ": " $0}' .claude/agents/moai/manager-git.md .claude/rules/moai/workflow/spec-workflow.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac021-post-hits.txt
wc -l < $E/ac021-post-hits.txt > $E/ac021-post-count.txt
```

장부 `.moai/reports/t622/b2-prose-classification.md` 분류값: `route-a-main-direct`, `prohibition-context`, `worktree-flow`(워크트리 안에서 PR 브랜치를 다룬다고 말함), `primary-branch-instruction`. 합격 조건: 장부 행 수 = `ac021-post-count.txt`, `primary-branch-instruction` 0행. 읽기 단계에서 spec-workflow.md 166행 Plan Phase 문장(브랜치 낱말이 없어 검출식 밖)이 워크트리 흐름과 맞는지도 기록한다.

### AC-GDP-022 — 트리 빌드로 잰 헌법 검증 (B3)

```
GIVEN 헌법 개정과 spec-workflow.md 문장 변경이 커밋된 트리
WHEN 이 트리에서 빌드한 moai 를 경로로 호출해 constitution validate --format json 을 실행하면
THEN 빌드와 호출이 exit 0 이고
 AND 빌드 정보의 vcs.revision 이 git rev-parse HEAD 와 같고 vcs.modified 가 false 이고
 AND JSON 이 "status": "ok" 와 "drift_count": 0 을 각각 1회 담는다
```

판정:

```bash
go build -o /tmp/t622-moai-tree ./cmd/moai > $E/ac022-build.txt 2>&1
# 기대: exit 0
git rev-parse HEAD > $E/ac022-head.txt
go version -m /tmp/t622-moai-tree > $E/ac022-buildinfo.txt 2>&1
# 기대: exit 0
/usr/bin/grep -E 'vcs\.(revision|modified)=' $E/ac022-buildinfo.txt > $E/ac022-vcs.txt
# 읽기: vcs.revision 값 = ac022-head.txt, vcs.modified=false (VCI §2.2 도구 출처)
unset CLAUDE_PROJECT_DIR MOAI_CONSTITUTION_REGISTRY && /tmp/t622-moai-tree constitution validate --format json > $E/ac022-validate.json 2> $E/ac022-validate.err
# 기대: exit 0
/usr/bin/grep -c -F '"status": "ok"' $E/ac022-validate.json > $E/ac022-status.txt
/usr/bin/grep -c -F '"drift_count": 0' $E/ac022-validate.json > $E/ac022-drift.txt
# 기대: 각각 1
```

대조(plan 작성 시점, 설치 빌드 `84fa4ece4` — 트리 빌드 아님, 검출기 동작 확인용). 검증기는 프로젝트 디렉터리 밖의 레지스트리를 거부하므로 대조는 트리 밖 스크래치 프로젝트 사본(`.claude` 트리와 `CLAUDE.md` 복사) 안에서 한다:

```bash
# 픽스처: 레지스트리의 027 clause 만 --after 문장으로 바꾸고 원본 spec-workflow.md 는 그대로
cd $S/v013-proj-fixture && moai constitution validate --format json > ../v013-ac022-fixture2.json
# 측정: exit 1, "status": "drift", "drift_count": 1, CONST-V3R5-027 DRIFT
cd $S/v013-proj-copy && moai constitution validate --format json > ../v013-ac022-copy2.json
# 측정: exit 0, "status": "ok", "drift_count": 0, "retired_count": 4 (바꾸지 않은 사본)
MOAI_CONSTITUTION_REGISTRY=$S/v013-ac022-registry-fixture.md moai constitution validate --format json > $S/v013-ac022-fixture-validate.json
# 측정: exit 1, "escapes project dir" 오류, "status": "", "drift_count": 0 — "status": "ok" 개수 0 이므로 판정식은 이 거부를 통과로 읽지 않는다
```

run-phase에서는 같은 짝 대조를 트리 빌드로 다시 잰다.

### AC-GDP-023 — 레지스트리 사본 일치와 새 clause (B3)

```
GIVEN 개정 뒤 zone-registry.md 로컬·템플릿 사본과 spec-workflow.md 로컬·템플릿 사본
WHEN 비교하고 clause 문장을 세면
THEN 두 레지스트리 사본의 diff 가 exit 0 이고
 AND 두 사본 모두 027·028 의 새 clause 줄이 각각 정확히 1회, 이전 clause 줄이 각각 0회이고
 AND spec-workflow.md 두 사본 모두 새 문장 두 개를 각각 1회 이상 담는다
```

대조(기준 트리·plan 작성 시점 측정): 레지스트리 L·T `diff` exit 0; 이전 clause 줄 L·T 각각 1; 새 문장은 레지스트리 L·T와 spec-workflow.md L·T 모두 0. 뮤턴트(로컬 레지스트리만 바뀐 상태 — amend 파이프라인이 로컬만 다루는 경우, spec.md §A.7): 템플릿 쪽 새 clause 0, `diff` exit 1 → FAIL.

판정:

```bash
diff .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac023-registry.diff
# 기대: exit 0
/usr/bin/grep -c -F 'clause: "Step 1 (plan) MUST execute in main checkout on Route A and inside a launcher-entered worktree on Route B"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac023-after027.txt
/usr/bin/grep -c -F 'clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after the single SPEC PR is merged"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac023-after028.txt
# 기대: 파일마다 1
/usr/bin/grep -c -F 'clause: "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac023-before027.txt
/usr/bin/grep -c -F 'clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac023-before028.txt
# 기대: 파일마다 0
/usr/bin/grep -c -F 'Step 1 (plan) MUST execute in main checkout on Route A and inside a launcher-entered worktree on Route B' .claude/rules/moai/workflow/spec-workflow.md $T/.claude/rules/moai/workflow/spec-workflow.md > $E/ac023-src027.txt
/usr/bin/grep -c -F 'It MUST happen ONLY after the single SPEC PR is merged' .claude/rules/moai/workflow/spec-workflow.md $T/.claude/rules/moai/workflow/spec-workflow.md > $E/ac023-src028.txt
# 기대: 파일마다 1 이상
```

### AC-GDP-024 — HumanOversight 승인을 레인이 대신하지 않음 (B3)

```
GIVEN 명령 기록 const-amend-commands.txt, 운영자 실행 출력, progress §E.2
WHEN 검사하면
THEN 명령 기록의 어떤 amend 줄도 표준입력을 파이프·here-string·파일 리다이렉트로 받지 않고
 AND 운영자 실행 출력 두 파일이 각각 대화형 질문 "(Y/N)" 을 담고
 AND progress §E.2 가 HumanOversight 단계에서 멈추고 리드에게 올린 기록을 담는다
```

검출식 한계: 레인이 스스로 기록한 명령 파일을 검사하므로, 기록되지 않은 실행은 보지 못한다(spec.md §E.2 성격의 잔여 위험). `--evidence` 값은 파이프 문자와 `<` 를 담지 않는다(spec.md §C.4).

대조(plan 작성 시점, 스크래치 픽스처 5줄): dry-run 한 줄 · `printf "Y\n" | moai constitution amend …` · `yes | moai constitution amend …` · `moai constitution amend … <<< Y` · `moai constitution amend … < answers.txt` → 검출식이 2~5줄을 잡고 dry-run 줄은 잡지 않음(exit 0). 기준 트리: 명령 기록 없음, progress §E.2 는 자리표시자.

판정:

```bash
/usr/bin/grep -n -E '\|[[:space:]]*[^|]*constitution amend|constitution amend.*[[:space:]]<' $E/const-amend-commands.txt > $E/ac024-stdin.txt
# 기대: exit 1
/usr/bin/grep -c -F '(Y/N)' $E/const-amend-027-operator.txt $E/const-amend-028-operator.txt > $E/ac024-prompt.txt
# 기대: 파일마다 1 이상
sed -n '/^## §E.2/,/^## §E.3/p' .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001/progress.md > $E/ac024-e2.md
/usr/bin/grep -c 'HumanOversight' $E/ac024-e2.md > $E/ac024-e2-oversight.txt
/usr/bin/grep -c -i 'lead' $E/ac024-e2.md > $E/ac024-e2-lead.txt
# 기대: 각각 1 이상
```

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. SHA는 AC-GDP-015에서 낱말 단위로 센다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. `awk` 절 추출은 시작 표지가 사라지면 빈 파일을 낸다 — 빈 파일은 판정 불가로 기록한다. 추출 표지(`### Late-Branch Invocation Pattern`, `[ZONE:Frozen] [HARD] Step ordering rules`, `[SHOULD] Anti-patterns`, `#### Late-branch Pre-check`, `#### Step 3.3.5`, `#### Step 3.4`, `##### Strategy: github-flow`, `##### Worktree Context Detection`)는 편집 뒤에도 남아야 한다.
- 철회된 기준(AC-GDP-011, 012)은 판정하지 않고 N/A로 기록한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다).
- 세 검출식(명령·서술형·단계별 PR)은 서로 다른 줄을 잡을 수 있고 같은 줄을 함께 잡을 수도 있다. 장부는 검출식마다 따로 만든다.
- 헌법 검증은 레지스트리가 프로젝트 디렉터리 안에 있어야 한다. 밖이면 exit 1·`"status": ""` 로 끝나며 이는 통과도 DRIFT도 아니다.

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다.
- `go test` 선택 실행이 최상위 PASS 줄 3개를 내지 않으면 합격이 아니다. 로컬 전체 스위트는 돌리지 않는다.
- 도구 측정(헌법 개정·검증)은 이 트리에서 빌드해 경로로 호출한 바이너리로 한다(`/tmp/t622-moai-tree`).

## §D.4 완료 정의 (Definition of Done)

- 활성 기준 AC-GDP-001~010, 013~015, 017~024가 PASS이고 AC-GDP-016이 PASS 또는 사유 기록. AC-GDP-011·012는 N/A.
- spec.md §C.5 B4·B6·T1이 run-phase 착수 전에 정해지고 그 결정이 progress 기록에 남는다. B4 결정에 따라 AC-GDP-020의 마지막 조건을 다시 쓴다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
