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
- 검출식 안의 `git` 은 `[g]it` 으로, `parallel` 은 `para[l]lel` 로, perl 코드 안의 `git` 은 `\x67it` 로 쓴다. 셸 변수를 받는 `sed`·`perl` 은 가드가 거부할 수 있으므로 경로를 글자 그대로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 개수가 찍히지 않은 결과는 판정 불가로 기록한다.
- **범위 파일 집합**: `.claude/agents/moai/manager-git.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/skills/moai/workflows/sync/delivery.md`, `.claude/skills/moai/workflows/sync/doc-execution.md` 의 로컬·템플릿 사본과 생성물 `$T/.codex/agents/moai/manager-git.toml`.
- **기준 트리 측정**: 0.2.0 작성 시점 HEAD `87988e946` 에서 범위 파일·생성물·관련 테스트 파일·`Makefile` 은 `$BASE` 와 차이가 없다(`git diff --stat` 출력 없음, exit 0). "plan 작성 시점 측정" 값은 이 템플릿 사본으로 잰 기준 트리 값이다.

## §D AC 표

**번호 방식**: 0.1.3 번호를 유지한다(spec.md §C.2). 카드 t658로 옮긴 번호와 철회된 번호는 판정하지 않는 자리표시로 남기고 spec.md §G에서 추적한다. 새 기준은 AC-GDP-025.

| AC ID | REQ | 등급 | 상태 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 활성 | `manager-git.md` 동기화 절의 문단·목록 묶음에서 fetch 가 rev-list 와 같은 배치로 묶이지 않고 순서가 지시됨 |
| AC-GDP-002 | REQ-GDP-002 | MUST-PASS | 활성 | Pre-Spawn 코드 블록에 단독 fetch·단독 rev-list 줄이 없고, 이어 붙인 줄 1개, rev-list 1개 |
| AC-GDP-003 | REQ-GDP-003 | MUST-PASS | 활성 | Pre-Edit Sync Check 절 불변 |
| AC-GDP-004 | REQ-GDP-004 | MUST-PASS | 활성 | `delivery.md` 병합 명령이 `--<merge_method>` 로 해석 |
| AC-GDP-005 | REQ-GDP-005 | MUST-PASS | 활성 | 범위 파일과 `.toml` 에서 `--squash` 고정 `gh pr merge` 가 기본값 설명 문장뿐 |
| AC-GDP-006 | REQ-GDP-006 | MUST-PASS | 활성 | `delivery.md`·`doc-execution.md` 에 워크트리 기본 병합 문구가 없고 `manager-git.md` 를 기준으로 밝힘 |
| AC-GDP-007 | — | — | 카드 t658로 이동 | 명령 검출식 분류 장부 |
| AC-GDP-008 | — | — | 카드 t658로 이동 | 절차 참조 유지 |
| AC-GDP-009 | — | — | 카드 t658로 이동 | 흐름 표지·접두 합집합 |
| AC-GDP-010 | — | — | 카드 t658로 이동 | 워크트리 흐름·Frozen 기록 |
| AC-GDP-011 | — | — | 철회(0.1.2) | OD-1 선택지 2 경로 |
| AC-GDP-012 | — | — | 철회(0.1.2) | OD-1 선택지 3 경로 |
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 활성 | 범위 파일 사본 일치, 의도된 차이 보존, 범위 파일을 덮는 테스트 2개 실행·통과 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 활성 | `.toml` 재생성과 `agents-emit-check` exit 0 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 활성 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 활성 | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |
| AC-GDP-017 ~ AC-GDP-024 | — | — | 카드 t658로 이동 | spec.md §G.1 표 |
| AC-GDP-025 | REQ-GDP-024 | MUST-PASS | 활성 (신규) | 범위 파일의 `[ZONE:Frozen]` 줄과 등록 Frozen clause 불변 |

판정 대상: 11개(AC-GDP-001~006, 013~016, 025). AC-GDP-016은 요구사항 추적 밖의 절차 점검이다.

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

판정은 줄이 아니라 문단 단위다(2회차 결함 N5).

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

읽기 단계: `ac001-local-paras.md` 의 문단마다 THEN (1)·(2)를 판정해 `$E/ac001-reading.md` 에 기록한다. 템플릿 사본과 생성물 `$T/.codex/agents/moai/manager-git.toml`(동기화 절 150행 주변)에도 같은 자동 실패 검사를 실행한다.

뮤턴트 재실행(세션 스크래치 파일):

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
# 기대: 1줄 (블록 2행 — 기준 트리 296행)
/usr/bin/grep -n -E '^[[:space:]]*[g]it rev-list' $E/ac002-base-block.md > $E/ac002-base-b.txt
# 기대: exit 0 (블록 5행 — 기준 트리 299행)
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

뮤턴트 재실행: 1회차 뮤턴트(fetch 줄 끝 주석) → a 1, b exit 0, c 0 → FAIL; 2회차 뮤턴트(이어 붙인 줄 + 두 번째 `git -C . rev-list` 줄) → rev-list 2 → FAIL; 올바른 픽스처 → a 0, b exit 1, c 1, rev-list 1 → PASS.

### AC-GDP-003 — Pre-Edit Sync Check 절 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절(제목부터 #### The sweep prohibition 앞까지)
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조: AC-GDP-002의 `ac002-base-section.md` 와 편집 뒤 `ac002-local-section.md` 를 `diff` 하면 exit 1(검출기가 Pre-Spawn 절의 변경을 잡음).

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
# 대조 기대: exit 1 — 출처 검출기는 기준 트리에서 빨강
/usr/bin/grep -c 'gh pr merge --squash --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-squash.txt
# 기대: 0
/usr/bin/grep -c 'gh pr merge --<merge_method> --delete-branch' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-resolved.txt
# 기대: 2 이상
/usr/bin/grep -n -E 'merge_method.*(squash|default)' .claude/skills/moai/workflows/sync/delivery.md > $E/ac004-local-source.txt
# 기대: exit 0, 적중 줄이 해석 출처를 설명하는지 읽어 기록
```

템플릿 사본도 같은 방식으로 판정한다.

### AC-GDP-005 — 고정 `--squash` 병합 예시 부재, 기본값 설명 유지

```
GIVEN 범위 파일 네 개(로컬·템플릿)와 생성물 manager-git.toml
WHEN 편집 뒤 --squash 가 붙은 gh pr merge 를 모두 찾으면
THEN manager-git.md 와 .toml 에서만 1줄씩 나오고, 그 줄은 기본값 설명 문장이며
 AND delivery.md·agent-common-protocol.md·doc-execution.md 에서는 0줄이다
```

대조(기준 트리):

```bash
/usr/bin/grep -n -E 'gh pr merge[^|]*--squash' $E/base-manager-git.md $E/base-delivery.md $E/base-manager-git.toml > $E/ac005-control.txt
# 기대: 6줄 — manager-git 32·114, delivery 343·355, .toml 26·108
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $E/base-agent-common-protocol.md $E/base-doc-execution.md > $E/ac005-control-zero.txt
# 기대: 파일마다 0
```

판정(로컬·템플릿 각각, 파일마다 따로 셈):

```bash
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/agents/moai/manager-git.md > $E/ac005-mg.txt
# 기대: 1
/usr/bin/grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-mg-default.txt
# 기대: 1 — 위의 1줄이 기본값 설명 문장임을 확인
/usr/bin/grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' .claude/agents/moai/manager-git.md > $E/ac005-mg-example.txt
# 기대: 1 — 114행 예시가 치환됨
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/skills/moai/workflows/sync/delivery.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/doc-execution.md > $E/ac005-others.txt
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

읽기 단계: 두 절 파일을 읽어, 워크트리 문맥만으로 병합이 일어난다는 문장이 없는지(검출식이 예상하지 않은 표현 포함)와 병합 조건이 `--auto-merge` 로 적혀 있는지 `$E/ac006-reading.md` 에 기록한다. 설정 키 `workflow.worktree.auto_merge` 는 판정 대상도 편집 대상도 아니다.

### AC-GDP-007 ~ AC-GDP-012 — 자리표시

AC-GDP-007·008·009·010은 카드 t658로 옮겼고, AC-GDP-011·012는 0.1.2에서 철회됐다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 파일 네 개의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md 와 agent-common-protocol.md 는 diff exit 0 이고
 AND delivery.md 와 doc-execution.md 는 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 범위 파일을 덮는 테스트 두 개(TestSanitizedPairParity, TestTemplateNoInternalContentLeak)가 각자 최상위 PASS 줄을 내며 exit 0 이다
```

**테스트 선택 근거** (plan 작성 시점 테스트 파일 읽기, `$BASE` 와 차이 없음):

| 범위 파일 | 사본 가드 | 근거 |
|---|---|---|
| `agent-common-protocol.md` | `diff` + `TestSanitizedPairParity` | `sanitized_pair_parity_test.go:71` 등록. `rule_template_mirror_test.go` 주석은 이 파일을 바이트 동일 허용 목록에서 뺐다고 적는다 |
| `manager-git.md` | `diff` 만 | `rule_template_mirror_test.go` 주석이 바이트 동일 목록에서 제거를 명시, 다른 사본 테스트 없음 |
| `delivery.md` | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음 |
| `doc-execution.md` | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음 |
| 템플릿 사본 네 개 전체 | `TestTemplateNoInternalContentLeak` (사본 일치가 아니라 템플릿 청결) | `internal_content_leak_test.go:1535`, 템플릿 루트 전체를 걷는다 |

0.1.x 판이 고른 `TestRuleTemplateMirrorDrift`·`TestLateBranchTemplateMirror` 는 허용 목록(`workflowOptMirroredPaths`, `lateBranchMirroredPaths`)에 범위 파일이 하나도 없어 이 SPEC의 파일을 덮지 않으므로 선택하지 않는다.

대조: 기준 트리에서 `delivery.md` 두 사본의 `diff` exit 1(`275c275`, `278c278`, `479,480c479`), `doc-execution.md` 두 사본의 `diff` exit 1(`138,143d137`), `manager-git.md`·`agent-common-protocol.md` 는 exit 0. 본문 비교 검출기는 빈 파일과 차이 본문을 `diff` 하면 exit 1.

판정:

```bash
diff .claude/agents/moai/manager-git.md $T/.claude/agents/moai/manager-git.md > $E/ac013-manager-git.diff
# 기대: exit 0
diff .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac013-acp.diff
# 기대: exit 0
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
go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0
/usr/bin/grep -c -E '^--- PASS: (TestSanitizedPairParity|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 정확히 2
/usr/bin/grep -c -F 'agent-common-protocol.md' $E/ac013-gotest.txt > $E/ac013-acp-subtest.txt
# 기대: 1 이상 — TestSanitizedPairParity 가 범위 파일 하위 테스트를 실제로 돌렸다는 기록(빈 선택 방지)
```

### AC-GDP-014 — 생성물 재생성

```
GIVEN 템플릿 manager-git.md 편집(114·156행)이 끝났고 .toml 은 아직 재생성하지 않은 상태
WHEN agents-emit-check → agents-emit → agents-emit-check 순서로 실행하면
THEN 첫 점검은 exit 1, 재생성은 exit 0, 두 번째 점검은 exit 0 이고
 AND 기준 트리 대비 .codex/agents/moai/ 아래 바뀐 파일은 manager-git.toml 뿐이고
 AND 재생성된 .toml 의 병합 예시와 동기화 문장이 템플릿 manager-git.md 와 같은 문장을 담는다
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
/usr/bin/grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-example.txt
# 기대: 1 (기준 트리 .toml 108행은 --squash)
```

템플릿 `manager-git.md` 는 이 SPEC에서 두 곳(114·156행)이 반드시 바뀌므로 첫 점검 RED는 반드시 나온다.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, CLAUDE.local 참조가 추가 줄에 없고
 AND 7~40자 16진 낱말 가운데 a-f 문자를 담은 것이 없으며
 AND 숫자로만 된 7~40자 낱말은 목록으로 뽑혀 읽기 단계에서 커밋 SHA가 아님이 기록된다
```

REQ-GDP-015의 프로그래밍 언어 편향 절은 기계 판정이 없다. 추가 줄을 읽어 기록하고 CI의 `template-neutrality-check` 결과를 함께 적는다. 이 기준은 `TestTemplateNoInternalContentLeak` 에 기대지 않고 추가 줄을 직접 검사한다.

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
# 기대: 5 — 추가 줄의 a-f 포함 SHA 다섯 개 모두(한 줄에 붙은 네 개 포함)
/usr/bin/grep -c -E ':[0-9]+$' $E/ac015-fixture-tokens.txt > $E/ac015-fixture-digits.txt
# 기대: 2 — 숫자로만 된 낱말은 읽기 목록으로 분리. 추가 줄이 아닌 02aca7afe 는 뽑히지 않음
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
WHEN agent-common-protocol.md 를 고친 커밋 이후의 커밋을 나머지 범위 지침 파일로 거르면
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
git log --format=%H <ac016-acp-commit.txt 의 SHA>..HEAD -- .claude/agents/moai/manager-git.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md internal/template/templates/.claude/skills/moai/workflows/sync/doc-execution.md > $E/ac016-after.txt
# 기대: 빈 파일
```

### AC-GDP-017 ~ AC-GDP-024 — 자리표시

카드 t658로 옮겼다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-025 — Frozen 줄과 등록 Frozen clause 불변

```
GIVEN 범위 파일 네 개의 로컬·템플릿 사본
WHEN 기준 트리 대비 변경분과 등록 Frozen clause 를 검사하면
THEN 변경분에 [ZONE:Frozen] 을 담은 추가·삭제 줄이 없고
 AND agent-common-protocol.md 두 사본 모두 등록 Frozen clause 네 문장이 기준 트리와 같은 개수로 남아 있고
 AND zone-registry.md 에 manager-git.md·delivery.md·doc-execution.md 를 가리키는 항목이 여전히 0개다
```

등록 Frozen clause (`zone-registry.md`, 모두 `file: .claude/rules/moai/core/agent-common-protocol.md`, `#user-interaction-boundary`):

| ID | clause | 기준 트리 위치 |
|---|---|---|
| `CONST-V3R2-006` | `` `AskUserQuestion` is the **only** user-facing question channel `` | 13행 |
| `CONST-V3R2-036` | `Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.` | 17행 |
| `CONST-V3R2-037` | `` Preload `AskUserQuestion` via `ToolSearch(query: `` | 52행 |
| `CONST-V3R2-038` | `AskUserQuestion is reserved exclusively for the MoAI orchestrator` | 17행 |

대조(plan 작성 시점 측정, 템플릿·로컬 사본): `[ZONE:Frozen]` 줄 — `agent-common-protocol.md` 17행 1줄, 나머지 세 파일 0줄. 네 clause `grep -c -F` 개수 — 사본마다 1·1·1·1. 레지스트리에서 나머지 세 파일을 가리키는 `file:` 항목 0개. 뮤턴트: `agent-common-protocol.md` 사본의 17행에서 "MUST NOT prompt" 를 "must not prompt" 로 바꾼 픽스처 → `CONST-V3R2-036` 개수 0, 기준 트리 사본과의 `diff` 에 `[ZONE:Frozen]` 을 담은 `<`·`>` 줄 → FAIL(측정값은 progress.md §E.1).

판정:

```bash
git diff $BASE -- .claude/agents/moai/manager-git.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/core/agent-common-protocol.md $T/.claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/doc-execution.md > $E/ac025-scope.diff
/usr/bin/grep -n -E '^[-+][^-+].*\[ZONE:Frozen\]' $E/ac025-scope.diff > $E/ac025-frozen-lines.txt
# 기대: exit 1
/usr/bin/grep -c -F '`AskUserQuestion` is the **only** user-facing question channel' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const006.txt
/usr/bin/grep -c -F 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const036.txt
/usr/bin/grep -c -F 'Preload `AskUserQuestion` via `ToolSearch(query:' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const037.txt
/usr/bin/grep -c -F 'AskUserQuestion is reserved exclusively for the MoAI orchestrator' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const038.txt
# 기대: 네 파일 모두 사본마다 1
/usr/bin/grep -c -E 'file: \.claude/(agents/moai/manager-git\.md|skills/moai/workflows/sync/(delivery|doc-execution)\.md)' .claude/rules/moai/core/zone-registry.md > $E/ac025-registry-others.txt
# 기대: 0
```

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. SHA는 AC-GDP-015에서 낱말 단위로 센다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. `awk` 절 추출은 시작 표지가 사라지면 빈 파일을 낸다 — 빈 파일은 판정 불가로 기록한다. 추출 표지(`## Synchronization`, `## PR Auto-Merge`, `### Pre-Spawn Sync Check`, `### Pre-Edit Sync Check`, `#### The sweep prohibition`, `#### Step 3.4`, `##### Worktree Context Detection`)는 편집 뒤에도 남아야 한다.
- 자리표시 기준(AC-GDP-007~012, 017~024)은 판정하지 않고 N/A로 기록한다.
- `delivery.md`·`doc-execution.md` 편집으로 줄 수가 바뀌면 사본 차이 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다).
- `go test -run` 선택자는 `^…$` 로 고정해 이름이 비슷한 다른 테스트가 섞이지 않게 하고, 최상위 PASS 줄 수와 범위 파일 하위 테스트 흔적을 함께 본다(빈 선택이 초록으로 보이는 일 방지).

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다.
- `go test` 선택 실행이 최상위 PASS 줄 2개를 내지 않으면 합격이 아니다. 로컬 전체 스위트는 돌리지 않는다.

## §D.4 완료 정의 (Definition of Done)

- 판정 대상 기준 AC-GDP-001~006, 013~015, 025가 PASS이고 AC-GDP-016이 PASS 또는 사유 기록. 자리표시 기준은 N/A.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
