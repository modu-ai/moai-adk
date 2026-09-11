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
- 검출식 안의 `git` 은 `[g]it` 으로, `parallel` 은 `para[l]lel` 로, perl 코드 안의 `git` 은 `\x67it` 로 쓴다. 셸 변수를 받는 `sed`·`perl` 과 경로를 만드는 반복문은 가드가 거부할 수 있으므로 경로를 글자 그대로 쓴다.
- 판정용 grep은 `/usr/bin/grep` 으로 실행한다. 개수가 찍히지 않은 결과는 판정 불가로 기록한다.
- **범위 파일 집합(열 개, 로컬·템플릿)**: `.claude/agents/moai/manager-git.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/skills/moai/workflows/sync/delivery.md`, `.claude/skills/moai/workflows/sync/doc-execution.md`, `.claude/skills/moai/SKILL.md`, `.claude/skills/moai/references/reference.md`, `.claude/skills/moai/workflows/sync/quality-gates-context.md`, `.claude/skills/moai/workflows/sync.md`, 명령 원본 `.claude/commands/moai/sync.md`(로컬) / `.claude/commands/moai/sync.md.tmpl`(템플릿). 생성물 `$T/.codex/agents/moai/manager-git.toml`, 게시본 `$T/.agents/skills/moai-sync/SKILL.md`(로컬 사본 `.agents/skills/moai-sync/SKILL.md`).
- **기준 트리 측정**: 0.2.2 작성 시점 HEAD `caa601d7c` 에서 범위 파일·생성물·게시본·발행기(`internal/template/commandemit`)·관련 테스트 파일·`docs-site/content`·`Makefile` 은 `$BASE` 와 차이가 없다(`git diff --stat` 출력 없음, exit 0). "plan 작성 시점 측정" 값은 이 템플릿 사본으로 잰 기준 트리 값이다.

## §D AC 표

**번호 방식**: 0.1.3 번호를 유지한다(spec.md §C.2). 카드 t658로 옮긴 번호와 철회된 번호는 판정하지 않는 자리표시로 남기고 spec.md §G에서 추적한다. 새 기준은 AC-GDP-025~030.

| AC ID | REQ | 등급 | 상태 | 요약 |
|---|---|---|---|---|
| AC-GDP-001 | REQ-GDP-001 | MUST-PASS | 활성 | `manager-git.md`·`.toml` 동기화 절에서 fetch 가 rev-list 와 같은 배치로 묶이지 않고 순서가 지시됨 |
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
| AC-GDP-013 | REQ-GDP-013 | MUST-PASS | 활성 (파일 확장) | 범위 파일 열 개 사본 일치, 의도된 차이 보존, 범위 파일을 덮는 테스트 2개 실행·통과 |
| AC-GDP-014 | REQ-GDP-014 | MUST-PASS | 활성 | `.toml` 재생성, `agents-emit-check` exit 0, `.toml` 두 절이 템플릿 `manager-git.md` 와 같음 |
| AC-GDP-015 | REQ-GDP-015 | MUST-PASS | 활성 | 템플릿 추가 줄에 SPEC ID·REQ 토큰·날짜·SHA·`CLAUDE.local` 없음 |
| AC-GDP-016 | 없음 — 절차 점검(plan.md §D 제약) | SHOULD-PASS | 활성 (파일 확장) | `agent-common-protocol.md` 커밋이 마지막 지침 편집 커밋 |
| AC-GDP-017 ~ AC-GDP-024 | — | — | 카드 t658로 이동 | spec.md §G.1 표 |
| AC-GDP-025 | REQ-GDP-024 | MUST-PASS | 활성 (파일 확장) | 범위 파일의 `[ZONE:Frozen]` 줄과 등록 Frozen clause 불변 |
| AC-GDP-026 | REQ-GDP-025 | MUST-PASS | 활성 (조각 확장) | 플래그 표면 아홉 조각이 `--auto-merge` 를 노출하고 `--merge` 를 폐기된 별칭으로만 서술 |
| AC-GDP-027 | REQ-GDP-025 | MUST-PASS | 활성 (조각 확장) | `--no-merge` 는 폐기된 no-op으로만 서술되고 병합 조건·동작에 쓰이지 않음 |
| AC-GDP-028 | REQ-GDP-026 | MUST-PASS | 활성 | team 모드: `--auto-merge` 가 전원 승인 조건과 함께 적힘, 승인 없는 병합 문장 없음 |
| AC-GDP-029 | REQ-GDP-026 | MUST-PASS | 활성 | personal·manual 모드: `--auto-merge` 가 승인 조건 없이 병합한다고 적힘, 승인을 요구하는 문장 없음 |
| AC-GDP-030 | REQ-GDP-014 | MUST-PASS | 활성 (신규) | 명령 원본 편집 뒤 `commands-emit`·`commands-emit-check` exit 0, 게시본 변화 여부를 두 경우 모두 판정, 바뀐 게시본은 원본과 같은 커밋 |

판정 대상: 16개(AC-GDP-001~006, 013~016, 025~030). AC-GDP-016은 요구사항 추적 밖의 절차 점검이다.

**감사가 지목한 잘못된 구현과 이를 잡는 기준**

| 잘못된 구현 | 잡는 기준 |
|---|---|
| `--merge` 를 별개이거나 폐기되지 않은 auto-merge 플래그로 서술 (결정 목록 줄, X1 사용법 줄, X2 `argument-hint`, X3 다음 단계 선택지 포함) | AC-GDP-026 (ii) — 그리고 `--auto-merge` 누락은 (i) |
| `--no-merge` 가 여전히 동작을 바꿈(건너뜀·트리거 조건) | AC-GDP-027 (i)·(ii) |
| 워크트리 문맥에서 여전히 기본 병합 | AC-GDP-006 (확장 검출식), AC-GDP-027 (ii) 트리거 조건 |
| team 모드가 승인 없이 병합 | AC-GDP-028 (a)·(b) |
| personal·manual 모드가 승인을 요구 | AC-GDP-029 (a)·(b) |
| 명령 원본을 고치고 게시본을 다시 만들지 않거나 다른 커밋에 넣음 | AC-GDP-030 |

## §D.1 Given-When-Then과 판정 명령

### AC-GDP-001 — fetch 와 rev-list 가 같은 배치로 묶이지 않고 순서가 지시됨

```
GIVEN manager-git.md 로컬·템플릿 사본과 생성물 manager-git.toml 의 "## Synchronization" 절
WHEN 절을 빈 줄로 나눈 문단(목록 묶음은 한 문단)마다 검사하면
THEN fetch 와 rev-list 를 함께 담은 문단이 1개 이상 있고
 AND 그중 배치·병렬 낱말을 담으면서 순서 낱말이 없는 문단이 0개이고
 AND 읽기 기록 $E/ac001-reading.md 가 존재하며, fetch 와 rev-list 를 함께 담은 모든 문단에 대해
     (1) fetch 가 끝난 뒤 rev-list 를 실행한다고 말하는지
     (2) fetch 와 rev-list 를 같은 배치·같은 목록·표·병렬 묶음에 넣지 않는지
     를 문단마다 예/아니오로 답하고 모두 예다
```

판정은 줄이 아니라 문단 단위다. 자동 검출은 보조이고, 읽기 기록이 PASS의 전제다(§D.3).

대조(기준 트리):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $E/base-manager-git.md > $E/ac001-base-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-base-sync.md > $E/ac001-base-paras.txt
# 기대: 1
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-base-sync.md > $E/ac001-base-autofail.md
test -s $E/ac001-base-autofail.md
# 기대: exit 0 — 자동 실패 검출기는 기준 트리에서 빨강(156행 문단)
```

판정(로컬·템플릿·생성물 각각 — 아래는 로컬·생성물 명령, 템플릿은 경로만 `$T/.claude/agents/moai/manager-git.md` 로 바꾼다):

```bash
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' .claude/agents/moai/manager-git.md > $E/ac001-local-sync.md
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-local-sync.md > $E/ac001-local-paras.txt
# 기대: 1 이상
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/' $E/ac001-local-sync.md > $E/ac001-local-paras.md
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-local-sync.md > $E/ac001-local-autofail.md
test -s $E/ac001-local-autofail.md
# 기대: exit 1
awk 'BEGIN{RS=""; ORS="\n\n"} /(^|\n)[-*|] [^\n]*fetch/ && /(^|\n)[-*|] [^\n]*rev-list/' $E/ac001-local-sync.md > $E/ac001-local-listgroup.md
# 비어 있지 않으면 목록·표 안에 fetch 와 rev-list 가 함께 있다는 뜻 — 읽기 기록 (2)에서 반드시 판정
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.codex/agents/moai/manager-git.toml > $E/ac001-toml-sync.md
test -s $E/ac001-toml-sync.md
# 기대: exit 0 — 절이 비면 판정 불가
awk 'BEGIN{RS=""} /fetch/ && /rev-list/ {c++} END{print c+0}' $E/ac001-toml-sync.md > $E/ac001-toml-paras.txt
# 기대: 1 이상
awk 'BEGIN{RS=""; ORS="\n\n"} /fetch/ && /rev-list/ && /(batch|para[l]lel|single-turn|multi-Bash|independent)/ && !/(first|before|once|after|completes|wait)/' $E/ac001-toml-sync.md > $E/ac001-toml-autofail.md
test -s $E/ac001-toml-autofail.md
# 기대: exit 1
test -s $E/ac001-reading.md
# 기대: exit 0 — 읽기 기록이 없으면 PASS 불가
```

plan 작성 시점 측정: 기준 트리 절 문단 1·autofail 1·목록 묶음 0. 기준 트리 `.toml` 의 `## Synchronization` 절은 템플릿 `manager-git.md` 의 같은 절과 diff exit 0(11줄)이므로 같은 값을 낸다.

뮤턴트 재실행:

```
"`git fetch` then … all in parallel"                               → 문단 1, autofail 1 → FAIL
2회차 목록 뮤턴트                                                   → autofail 0, listgroup 1 → 읽기 기록 (2)에서 FAIL
감사 뮤턴트 "… ONE single-turn multi-Bash batch after the checkpoint" → autofail 0 → 읽기 기록 (1)·(2)에서 FAIL
감사 표 뮤턴트(| remote | `fetch` | / | divergence | `rev-list …` |)   → autofail 0, listgroup(표 행 포함) 비어 있지 않음 → 읽기 기록에서 FAIL
2회차 올바른 문장                                                   → autofail 0, listgroup 0 → 빨강 아님
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
# 기대: exit 0
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

뮤턴트 재실행: fetch 줄 끝 주석 → a 1, b exit 0, c 0 → FAIL; 이어 붙인 줄 + 두 번째 `git -C . rev-list` 줄 → rev-list 2 → FAIL; 올바른 픽스처 → a 0, b exit 1, c 1, rev-list 1 → PASS.

### AC-GDP-003 — Pre-Edit Sync Check 절 불변

```
GIVEN 기준 트리와 편집 뒤의 Pre-Edit Sync Check 절(제목부터 #### The sweep prohibition 앞까지)
WHEN 두 절을 추출해 비교하면
THEN 차이가 없다
```

대조: AC-GDP-002의 `ac002-base-section.md` 와 편집 뒤 `ac002-local-section.md` 를 `diff` 하면 exit 1.

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

템플릿 사본도 같은 방식으로 판정한다.

### AC-GDP-005 — 고정 `--squash` 병합 예시 부재, 기본값 설명 유지

```
GIVEN 범위 파일 열 개(로컬·템플릿)와 생성물 manager-git.toml
WHEN 편집 뒤 --squash 가 붙은 gh pr merge 를 모두 찾으면
THEN manager-git.md 와 .toml 에서만 1줄씩 나오고, 그 줄은 기본값 설명 문장이며
 AND 나머지 여덟 파일에서는 0줄이다
```

대조(기준 트리):

```bash
/usr/bin/grep -n -E 'gh pr merge[^|]*--squash' $E/base-manager-git.md $E/base-delivery.md $E/base-manager-git.toml > $E/ac005-control.txt
# 기대: 6줄 — manager-git 32·114, delivery 343·355, .toml 26·108
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $E/base-agent-common-protocol.md $E/base-doc-execution.md $E/base-skill.md $E/base-reference.md $E/base-qgc.md $E/base-sync.md $E/base-command-sync.md > $E/ac005-control-zero.txt
# 기대: 파일마다 0
```

판정(로컬·템플릿 각각, 파일마다 따로 셈):

```bash
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/agents/moai/manager-git.md > $E/ac005-mg.txt
# 기대: 1
/usr/bin/grep -c -F 'which under the squash default renders `gh pr merge --squash --delete-branch`' .claude/agents/moai/manager-git.md > $E/ac005-mg-default.txt
# 기대: 1
/usr/bin/grep -c -F 'gh pr merge <PR> --<merge_method> --delete-branch' .claude/agents/moai/manager-git.md > $E/ac005-mg-example.txt
# 기대: 1 — 114행 예시가 치환됨
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' .claude/skills/moai/workflows/sync/delivery.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md > $E/ac005-others.txt
# 기대: 파일마다 0
/usr/bin/grep -c -E 'gh pr merge[^|]*--squash' $T/.codex/agents/moai/manager-git.toml > $E/ac005-toml.txt
# 기대: 1 (기본값 설명 문장)
```

1회차 감사 뮤턴트(`gh pr merge 42 --squash --delete-branch`, `gh pr merge --squash <PR> --delete-branch`)는 이 검출식에 걸려 FAIL로 판정된다.

### AC-GDP-006 — auto-merge 기본값 단일 기준 (OD-2 = B)

```
GIVEN delivery.md 로컬·템플릿 사본의 Step 3.4 절과 doc-execution.md 로컬·템플릿 사본의 "Worktree Context Detection" 소절
WHEN 편집 뒤 두 절을 추출해 검사하면
THEN 워크트리 문맥을 기본 병합과 묶는 문구(확장 검출식)가 두 절 모두 0개이고
 AND 두 절이 각각 manager-git.md 를 기준으로 1회 이상 이름으로 밝히고
 AND manager-git.md 의 옵트인 문장 두 개(148행, 166행)의 --auto-merge 조건이 남아 있고
 AND 읽기 기록 $E/ac006-reading.md 가 존재하며, 두 절의 모든 문장에 대해 "워크트리 문맥만으로 병합이 일어난다고 말하는가" 에 아니오로 답하고, 병합 조건이 --auto-merge 로 적혀 있음을 확인한다
```

확장 검출식(대조·판정 공통):

```
default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)
```

doc-execution 소절에는 기존 검출식 `default (to )?auto-merge|worktree contexts default` 도 함께 쓴다.

절 추출(대조·판정 공통):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery.md 경로> > <delivery 절 파일>
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' <doc-execution.md 경로> > <doc-execution 절 파일>
```

대조(기준 트리):

```bash
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' $E/base-delivery.md > $E/ac006-base-dl.md
awk '/^##### Worktree Context Detection/{s=1; print; next} s && /^#/{exit} s' $E/base-doc-execution.md > $E/ac006-base-de.md
awk '/default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)/ {print FNR}' $E/ac006-base-dl.md > $E/ac006-base-dl-default.txt
# 기대: 8·20 (파일 기준 337·349행)
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default|merges? (automatically|by default)' $E/ac006-base-de.md > $E/ac006-base-de-default.txt
# 기대: exit 0 — 절 안 8행, 파일 기준 36행
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
awk '/default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)/ {print FNR}' $E/ac006-local-dl.md > $E/ac006-local-dl-default.txt
test -s $E/ac006-local-dl-default.txt
# 기대: exit 1
/usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default|merges? (automatically|by default)' $E/ac006-local-de.md > $E/ac006-local-de-default.txt
# 기대: exit 1
/usr/bin/grep -c 'manager-[g]it[.]md' $E/ac006-local-dl.md $E/ac006-local-de.md > $E/ac006-local-source.txt
# 기대: 파일마다 1 이상
/usr/bin/grep -c -e '--auto-merge' .claude/agents/moai/manager-git.md > $E/ac006-local-optin.txt
# 기대: 2 이상 (148·166행 옵트인 조건 유지)
test -s $E/ac006-reading.md
# 기대: exit 0 — 읽기 기록이 없으면 PASS 불가
```

뮤턴트 재실행(스크래치 픽스처 4줄): "Merge policy follows manager-git.md." · "When the session runs inside a worktree and the no-merge flag is absent, merge once checks pass" · "Worktree sessions merge automatically." · "Auto-merge is not the default; pass `--auto-merge` per manager-git.md." → 확장 검출식이 2·3줄을 잡고 올바른 4줄은 잡지 않음.

### AC-GDP-007 ~ AC-GDP-012 — 자리표시

AC-GDP-007·008·009·010은 카드 t658로 옮겼고, AC-GDP-011·012는 0.1.2에서 철회됐다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-013 — 사본 일치와 의도된 차이 보존

```
GIVEN 범위 파일 열 개의 로컬·템플릿 사본과 게시본 moai-sync/SKILL.md 의 로컬·템플릿 사본
WHEN 편집 뒤 비교하면
THEN manager-git.md, agent-common-protocol.md, quality-gates-context.md, 게시본 SKILL.md 는 diff exit 0 이고
 AND delivery.md, doc-execution.md, moai/SKILL.md, references/reference.md, workflows/sync.md, 명령 원본(sync.md 대 sync.md.tmpl)은 줄번호 머리를 뺀 차이 본문이 기준 트리와 같고
 AND 명령 원본 두 사본의 argument-hint 줄이 서로 같고
 AND 범위 파일을 덮는 테스트 두 개(TestSanitizedPairParity, TestTemplateNoInternalContentLeak)가 각자 최상위 PASS 줄을 내며 exit 0 이다
```

**파일별 사본 가드** (plan 작성 시점 테스트 파일 읽기, `$BASE` 와 차이 없음):

| 범위 파일 | 기준 트리 L·T diff | 사본 가드 | 근거 |
|---|---|---|---|
| `agent-common-protocol.md` | exit 0 | `diff` + `TestSanitizedPairParity` | `sanitized_pair_parity_test.go:71` |
| `manager-git.md` | exit 0 | `diff` 만 | `rule_template_mirror_test.go` 주석이 바이트 동일 목록에서 제거를 명시 |
| `quality-gates-context.md` | exit 0 | `diff` 만 | 어떤 사본 테스트에도 없음 |
| `delivery.md` | exit 1 (`275c275`, `278c278`, `479,480c479`) | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음. X3(404행)도 이 판정에 들어간다 |
| `doc-execution.md` | exit 1 (`138,143d137`) | 차이 본문 `diff` 만 | 어떤 사본 테스트에도 없음 |
| `moai/SKILL.md` | exit 1 (20개 덩어리, `125c125` … `392d391`) | 차이 본문 `diff` 만 | `backlog_json_disclosure_mirror_test.go:24` 는 임베드 사본 = 템플릿 원본을 볼 뿐 로컬 사본을 보지 않는다 |
| `references/reference.md` | exit 1 (`229d228`) | 차이 본문 `diff` 만 | `agent_frontmatter_audit_test.go:407` 은 프론트매터만 본다 |
| `workflows/sync.md` | exit 1 (`65,74d64`, `81c71`) | 차이 본문 `diff` 만 | `agentless_audit_test.go:44` 는 사본 일치가 아닌 지침 내용을 본다. X1 사용법 줄(로컬 95 / 템플릿 85)도 이 판정에 들어간다 |
| 명령 원본 `.claude/commands/moai/sync.md` (로컬) 대 `sync.md.tmpl` (템플릿) | exit 1 (`2c2` — `description` 줄만 다름) | 차이 본문 `diff` + `argument-hint` 줄 비교 | 파일 형식이 다르다: 템플릿은 `moai init` 때 `ConversationLanguage` 로 렌더링되는 Go 템플릿이고 2행 `description` 이 로케일 조건문이다. 로컬은 렌더링된 영어 사본이다. X2가 고치는 3행에는 템플릿 액션이 없어 두 사본에서 글자 그대로 같아야 한다. `commandemit/golden_test.go` 는 템플릿 원본과 게시본의 관계를 보며 로컬 원본을 보지 않는다 |
| 게시본 `.agents/skills/moai-sync/SKILL.md` (로컬) 대 템플릿 게시본 | exit 0 | `diff` | 발행기는 템플릿 게시본만 쓴다(`golden_test.go` `templatesDir`). 로컬 사본은 추적 파일이다 |
| 템플릿 사본 전체 | — | `TestTemplateNoInternalContentLeak` (사본 일치가 아니라 템플릿 청결) | `internal_content_leak_test.go:1535`, 템플릿 루트 전체를 걷는다 |

`TestRuleTemplateMirrorDrift`·`TestLateBranchTemplateMirror` 는 허용 목록에 범위 파일이 하나도 없어 선택하지 않는다.

대조: 위 표의 기준 트리 diff 결과, 명령 원본 `argument-hint` 줄 L·T diff exit 0(plan 작성 시점 측정). 본문 비교 검출기는 빈 파일과 차이 본문을 `diff` 하면 exit 1.

판정(바이트 동일 네 파일):

```bash
diff .claude/agents/moai/manager-git.md $T/.claude/agents/moai/manager-git.md > $E/ac013-manager-git.diff
# 기대: exit 0
diff .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac013-acp.diff
# 기대: exit 0
diff .claude/skills/moai/workflows/sync/quality-gates-context.md $T/.claude/skills/moai/workflows/sync/quality-gates-context.md > $E/ac013-qgc.diff
# 기대: exit 0
diff .agents/skills/moai-sync/SKILL.md $T/.agents/skills/moai-sync/SKILL.md > $E/ac013-published.diff
# 기대: exit 0
```

판정(의도된 차이 여섯 파일 — 파일마다 아래 다섯 줄을 경로만 바꿔 실행한다. 기준 트리 반출 이름은 `base-<이름>.md`·`base-<이름>-template.md`):

```bash
diff $E/base-delivery.md $E/base-delivery-template.md > $E/ac013-delivery-base.diff
diff .claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/delivery.md > $E/ac013-delivery-post.diff
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-base.diff > $E/ac013-delivery-base.body
/usr/bin/grep -v -E '^[0-9]+(,[0-9]+)?[acd][0-9]+(,[0-9]+)?$' $E/ac013-delivery-post.diff > $E/ac013-delivery-post.body
diff $E/ac013-delivery-base.body $E/ac013-delivery-post.body > $E/ac013-delivery-body.diff
# 기대: exit 0
# 같은 형태: doc-execution.md(base-doc-execution), moai/SKILL.md(base-skill), references/reference.md(base-reference),
#            workflows/sync.md(base-sync), 명령 원본(base-command-sync 대 base-command-sync-template;
#            로컬 .claude/commands/moai/sync.md 대 $T/.claude/commands/moai/sync.md.tmpl)
```

판정(명령 원본 `argument-hint` 줄):

```bash
/usr/bin/grep -E '^argument-hint:' .claude/commands/moai/sync.md > $E/ac013-hint-local.txt
/usr/bin/grep -E '^argument-hint:' $T/.claude/commands/moai/sync.md.tmpl > $E/ac013-hint-template.txt
test -s $E/ac013-hint-local.txt
test -s $E/ac013-hint-template.txt
# 기대: 둘 다 exit 0
diff $E/ac013-hint-local.txt $E/ac013-hint-template.txt > $E/ac013-hint.diff
# 기대: exit 0
```

판정(테스트):

```bash
go test ./internal/template/ -run '^(TestSanitizedPairParity|TestTemplateNoInternalContentLeak)$' -v -count=1 > $E/ac013-gotest.txt 2>&1
# 기대: exit 0
/usr/bin/grep -c -E '^--- PASS: (TestSanitizedPairParity|TestTemplateNoInternalContentLeak) ' $E/ac013-gotest.txt > $E/ac013-pass-count.txt
# 기대: 정확히 2
/usr/bin/grep -c -F 'agent-common-protocol.md' $E/ac013-gotest.txt > $E/ac013-acp-subtest.txt
# 기대: 1 이상 — TestSanitizedPairParity 가 범위 파일 하위 테스트를 실제로 돌렸다는 기록(빈 선택 방지)
```

뮤턴트: 로컬 명령 원본의 `argument-hint` 만 `--auto-merge` 로 바꾸고 템플릿은 그대로 둔 상태 → `ac013-hint.diff` exit 1, 차이 본문에 `argument-hint` 줄 쌍이 더해져 본문 비교 exit 1 → FAIL.

### AC-GDP-014 — 에이전트 생성물 재생성

```
GIVEN 템플릿 manager-git.md 편집(114·156행과 PR Auto-Merge 절)이 끝났고 .toml 은 아직 재생성하지 않은 상태
WHEN agents-emit-check → agents-emit → agents-emit-check 순서로 실행하면
THEN 첫 점검은 exit 1, 재생성은 exit 0, 두 번째 점검은 exit 0 이고
 AND 기준 트리 대비 .codex/agents/moai/ 아래 바뀐 파일은 manager-git.toml 뿐이고
 AND .toml 의 병합 예시가 --<merge_method> 이고
 AND .toml 의 "## Synchronization" 절과 "## PR Auto-Merge" 절이 템플릿 manager-git.md 의 같은 절과 diff exit 0 이다
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
# 기대: 1 (기준 트리 .toml 108행은 --squash, 이 개수 0)
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.claude/agents/moai/manager-git.md > $E/ac014-md-sync.md
sed -n '/^## Synchronization/,/^## PR Auto-Merge/p' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-sync.md
diff $E/ac014-md-sync.md $E/ac014-toml-sync.md > $E/ac014-sync.diff
# 기대: exit 0 — 156행 새 순서 문장이 .toml 에 같은 문장으로 들어감
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' $T/.claude/agents/moai/manager-git.md > $E/ac014-md-pram.md
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' $T/.codex/agents/moai/manager-git.toml > $E/ac014-toml-pram.md
test -s $E/ac014-toml-pram.md
# 기대: exit 0
diff $E/ac014-md-pram.md $E/ac014-toml-pram.md > $E/ac014-pram.diff
# 기대: exit 0 — 모드별 승인 규칙이 .toml 에 같은 문장으로 들어감
```

plan 작성 시점 측정: 기준 트리에서 두 절의 diff 는 exit 0(11줄, 9줄). 뮤턴트: 기준 트리 `.toml` 절에서 "AND all approvals obtained" 를 "once checks pass" 로 바꾼 픽스처 → diff exit 1.

### AC-GDP-015 — 템플릿 중립성

```
GIVEN 기준 트리 대비 템플릿 변경분(범위 파일 열 개의 템플릿 사본, 명령 원본 sync.md.tmpl, 게시본 포함)
WHEN 추가 줄(+++ 머리 줄 제외)을 검사하면
THEN SPEC ID, REQ 토큰, 날짜, CLAUDE.local 참조가 추가 줄에 없고
 AND 7~40자 16진 낱말 가운데 a-f 문자를 담은 것이 없으며
 AND 숫자로만 된 7~40자 낱말은 목록으로 뽑혀 읽기 단계에서 커밋 SHA가 아님이 기록된다
```

`git diff $BASE -- $T/` 는 템플릿 루트 전체를 담으므로 X1~X3의 템플릿 사본과 게시본 변화도 이 판정에 들어간다. 프로그래밍 언어 편향은 추가 줄을 읽어 기록하고 CI의 `template-neutrality-check` 결과를 함께 적는다.

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

검출 한계: 대문자 16진, 7자 미만 약식 SHA, 영숫자에 바로 붙은 16진 낱말은 잡지 않는다.

### AC-GDP-016 — 항상 로드 규칙 편집이 마지막 (SHOULD, 절차 점검)

```
GIVEN run-phase 커밋들
WHEN agent-common-protocol.md 를 고친 커밋 이후의 커밋을 나머지 범위 지침 파일로 거르면
THEN 결과가 비어 있다
```

```bash
git log --format=%H $BASE..HEAD -- .claude/agents/moai/manager-git.md .claude/skills/moai/workflows/sync/delivery.md > $E/ac016-control.txt
# 대조 기대: 1줄 이상
git log --format=%H -1 $BASE..HEAD -- .claude/rules/moai/core/agent-common-protocol.md > $E/ac016-acp-commit.txt
git log --format=%H <ac016-acp-commit.txt 의 SHA>..HEAD -- .claude/agents/moai/manager-git.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md internal/template/templates/.claude/agents/moai/manager-git.md internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md internal/template/templates/.claude/skills/moai/workflows/sync/doc-execution.md internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/references/reference.md internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-context.md internal/template/templates/.claude/skills/moai/workflows/sync.md internal/template/templates/.claude/commands/moai/sync.md.tmpl > $E/ac016-after.txt
# 기대: 빈 파일
```

### AC-GDP-017 ~ AC-GDP-024 — 자리표시

카드 t658로 옮겼다. 판정하지 않으며 원문은 커밋 `87988e946` 의 acceptance.md, 추적 표는 spec.md §G.1이다.

### AC-GDP-025 — Frozen 줄과 등록 Frozen clause 불변

```
GIVEN 범위 파일 열 개의 로컬·템플릿 사본
WHEN 기준 트리 대비 변경분과 등록 Frozen clause 를 검사하면
THEN 변경분에 [ZONE:Frozen] 을 담은 추가·삭제 줄이 없고
 AND agent-common-protocol.md 두 사본 모두 등록 Frozen clause 네 문장이 기준 트리와 같은 개수로 남아 있고
 AND zone-registry.md 에서 나머지 아홉 파일을 가리키는 항목이 여전히 0개이며, 같은 검출식 형태로 센 agent-common-protocol.md 항목이 13개다(양성 대조)
```

등록 Frozen clause (`zone-registry.md`, 모두 `file: .claude/rules/moai/core/agent-common-protocol.md`, `#user-interaction-boundary`):

| ID | clause | 기준 트리 위치 |
|---|---|---|
| `CONST-V3R2-006` | `` `AskUserQuestion` is the **only** user-facing question channel `` | 13행 |
| `CONST-V3R2-036` | `Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.` | 17행 |
| `CONST-V3R2-037` | `` Preload `AskUserQuestion` via `ToolSearch(query: `` | 52행 |
| `CONST-V3R2-038` | `AskUserQuestion is reserved exclusively for the MoAI orchestrator` | 17행 |

대조(plan 작성 시점 측정, 템플릿·로컬 사본): `[ZONE:Frozen]` 줄 — `agent-common-protocol.md` 17행 1줄, 나머지 아홉 파일 0줄(명령 원본 `sync.md`·`sync.md.tmpl` 포함). 네 clause 개수 — 사본마다 1·1·1·1. 레지스트리에서 나머지 아홉 파일을 가리키는 `file:` 항목 0개(명령 원본 `file: .claude/commands/moai/sync.md` 0개 포함), `agent-common-protocol.md` 를 가리키는 항목 13개(로컬·템플릿). 뮤턴트: `agent-common-protocol.md` 템플릿 사본의 "MUST NOT prompt" 를 소문자로 바꾼 픽스처 → `CONST-V3R2-036` 개수 0, 기준 사본과의 `diff` 에 `[ZONE:Frozen]` 을 담은 줄 2개 → FAIL. Frozen 줄 판정 뮤턴트: `-[ZONE:Frozen] a` · `+[ZONE:Frozen] b` · ` [ZONE:Frozen] c`(문맥) 세 줄 diff 픽스처 → 2줄 적중, 문맥 줄 미적중.

판정:

```bash
git diff $BASE -- .claude/agents/moai/manager-git.md .claude/rules/moai/core/agent-common-protocol.md .claude/skills/moai/workflows/sync/delivery.md .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/SKILL.md .claude/skills/moai/references/reference.md .claude/skills/moai/workflows/sync/quality-gates-context.md .claude/skills/moai/workflows/sync.md .claude/commands/moai/sync.md $T/.claude/agents/moai/manager-git.md $T/.claude/rules/moai/core/agent-common-protocol.md $T/.claude/skills/moai/workflows/sync/delivery.md $T/.claude/skills/moai/workflows/sync/doc-execution.md $T/.claude/skills/moai/SKILL.md $T/.claude/skills/moai/references/reference.md $T/.claude/skills/moai/workflows/sync/quality-gates-context.md $T/.claude/skills/moai/workflows/sync.md $T/.claude/commands/moai/sync.md.tmpl > $E/ac025-scope.diff
/usr/bin/grep -F '[ZONE:Frozen]' $E/ac025-scope.diff > $E/ac025-frozen-any.txt
/usr/bin/grep -n -E '^[-+].*\[ZONE:Frozen\]' $E/ac025-frozen-any.txt > $E/ac025-frozen-lines.txt
# 기대: exit 1 (문맥 줄 ' …[ZONE:Frozen]' 은 허용, 추가·삭제 줄은 불허).
# 첫 grep 에 -n 을 붙이면 줄번호 머리 때문에 둘째 grep 의 ^[-+] 가 아무 줄도 잡지 못해 공허하게 통과한다 — 붙이지 않는다.
/usr/bin/grep -c -F '`AskUserQuestion` is the **only** user-facing question channel' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const006.txt
/usr/bin/grep -c -F 'Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const036.txt
/usr/bin/grep -c -F 'Preload `AskUserQuestion` via `ToolSearch(query:' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const037.txt
/usr/bin/grep -c -F 'AskUserQuestion is reserved exclusively for the MoAI orchestrator' .claude/rules/moai/core/agent-common-protocol.md $T/.claude/rules/moai/core/agent-common-protocol.md > $E/ac025-const038.txt
# 기대: 네 파일 모두 사본마다 1
/usr/bin/grep -c -E 'file: \.claude/rules/moai/core/agent-common-protocol\.md' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac025-registry-control.txt
# 기대: 사본마다 13 — 검출식 형태가 레지스트리 항목을 실제로 잡는다는 양성 대조
/usr/bin/grep -c -E 'file: \.claude/(agents/moai/manager-git\.md|skills/moai/workflows/sync/(delivery|doc-execution|quality-gates-context)\.md|skills/moai/SKILL\.md|skills/moai/references/reference\.md|skills/moai/workflows/sync\.md|commands/moai/sync\.md)' .claude/rules/moai/core/zone-registry.md $T/.claude/rules/moai/core/zone-registry.md > $E/ac025-registry-others.txt
# 기대: 사본마다 0
```

### AC-GDP-026 — 플래그 표면: `--auto-merge` 노출, `--merge` 는 폐기된 별칭으로만

```
GIVEN 플래그 표면 아홉 조각의 로컬·템플릿 사본
      (1) moai/SKILL.md 의 "Modes: auto, force, status, project. Flags:" 줄
      (2) references/reference.md 의 "- Modes (positional): auto (default), force, status, project" 로 시작하는 목록(빈 줄까지)
      (3) quality-gates-context.md 의 "- $ARGUMENTS: Mode and optional path" 로 시작하는 목록(빈 줄까지)
      (4) quality-gates-context.md 의 "## Supported Flags" 절
      (5) workflows/sync.md 의 "**Flags**:" 줄
      (6) delivery.md 의 "#### Step 3.4" 절
      (7) workflows/sync.md 의 "/moai sync [mode]" 로 시작하는 사용법 줄 (X1)
      (8) 명령 원본의 "argument-hint:" 줄 — 로컬 .claude/commands/moai/sync.md, 템플릿 .claude/commands/moai/sync.md.tmpl (X2)
      (9) delivery.md 의 "#### Context-Aware Next Steps" 절 (X3, 404행 선택지를 담음)
WHEN 편집 뒤 조각을 추출해 검사하면
THEN (i) 조각 (1)·(2)·(5)·(6)·(7)·(8)·(9)와 조각 (3)·(4)를 합친 quality-gates-context 조각이 각각 --auto-merge 를 1회 이상 담고
 AND (ii) --merge 낱말(앞뒤가 영문자·하이픈이 아닌 --merge)을 담은 줄은 모두 "deprecat" 와 "--auto-merge" 를 함께 담는다
     (위반 줄 0개)
```

조각 추출(대조·판정 공통, `<SKILL>` 등은 로컬·템플릿 경로):

```bash
/usr/bin/grep -E '^Modes: auto, force, status, project\. Flags:' <SKILL> > $E/ac026-skill.md
awk '/^- Modes \(positional\): auto \(default\), force, status, project/{s=1; print; next} s && /^$/{exit} s' <reference> > $E/ac026-ref.md
awk '/^- \$ARGUMENTS: Mode and optional path/{s=1; print; next} s && /^$/{exit} s' <quality-gates-context> > $E/ac026-qgc-args.md
awk '/^## Supported Flags/{s=1; print; next} s && /^## /{exit} s' <quality-gates-context> > $E/ac026-qgc-flags.md
/usr/bin/grep -E '^\*\*Flags\*\*: ' <sync.md> > $E/ac026-sync.md
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery> > $E/ac026-dl.md
/usr/bin/grep -E '^/moai sync \[mode\]' <sync.md> > $E/ac026-sync-usage.md
/usr/bin/grep -E '^argument-hint:' <command source> > $E/ac026-hint.md
awk '/^#### Context-Aware Next Steps/{s=1; print; next} s && /^(####|###|##) /{exit} s' <delivery> > $E/ac026-dl-next.md
```

조각의 추출 표지(`Modes: auto, force, status, project. Flags:`, `- Modes (positional): …`, `- $ARGUMENTS: Mode and optional path`, `## Supported Flags`, `**Flags**:`, `#### Step 3.4`, `/moai sync [mode]`, `argument-hint:`, `#### Context-Aware Next Steps`)는 편집 뒤에도 남아야 한다.

대조(기준 트리, 템플릿 사본; 조각 (7)·(8)·(9)는 로컬 사본도 같은 값):

```
조각 줄 수: skill 1, ref 3, qgc-args 4, qgc-flags 6, sync 1, dl 52, sync-usage 1, hint 1, dl-next 24
(i) --auto-merge 개수: 아홉 조각 모두 0 → 빨강
(ii) 위반 줄: skill 1(140행), ref 2(161행), qgc-args 4(30행), qgc-flags 4(101행), sync 1(104행), dl 9(338행)·20(349행),
     sync-usage 1(템플릿 85 / 로컬 95), hint 1(3행), dl-next 9(404행) → 10줄 → 빨강
```

판정(로컬·템플릿 각각):

```bash
test -s $E/ac026-skill.md
test -s $E/ac026-ref.md
test -s $E/ac026-qgc-args.md
test -s $E/ac026-qgc-flags.md
test -s $E/ac026-sync.md
test -s $E/ac026-dl.md
test -s $E/ac026-sync-usage.md
test -s $E/ac026-hint.md
test -s $E/ac026-dl-next.md
# 기대: 모두 exit 0 — 조각이 비면 판정 불가
/usr/bin/grep -c -e '--auto-merge' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac026-i.txt
/usr/bin/grep -c -e '--auto-merge' $E/ac026-qgc-args.md $E/ac026-qgc-flags.md > $E/ac026-i-qgc.txt
# 기대: ac026-i.txt 파일마다 1 이상, ac026-i-qgc.txt 합계 1 이상
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && !(/[Dd]eprecat/ && /--auto-merge/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac026-ii.txt
test -s $E/ac026-ii.txt
# 기대: exit 1
```

뮤턴트(plan 작성 시점, 스크래치 픽스처):

```
결정 목록 조각 픽스처 5줄: "Modes: … Flags: --merge, --skip-mx"(그대로) · "Modes: … Flags: --auto-merge, --merge, --skip-mx"(별개·폐기 표시 없음)
  · "- `--merge` (deprecated)"(별칭 관계 없음) · "- `--merge`: deprecated alias of `--auto-merge` (logs a warning)"(올바름)
  · "moai worktree clean --merged-only"(다른 플래그)
  → (ii)가 1·2·3줄을 잡고 4·5줄은 잡지 않음
X1~X3 픽스처 6줄: "/moai sync [mode] [--pr] [--auto-merge] [--merge] [--skip-mx]" · "/moai sync [mode] [--pr] [--auto-merge] [--skip-mx]"
  · "argument-hint: \"[SPEC-XXX] [--merge] [--skip-mx]\"" · "argument-hint: \"[SPEC-XXX] [--auto-merge] [--skip-mx]\""
  · "- Auto-Merge PR (/moai sync --merge)" · "- Auto-Merge PR (/moai sync --auto-merge)"
  → (ii)가 1·3·5줄(옛 별칭을 폐기 표시 없이 남긴 줄)을 잡고, --auto-merge 를 담은 줄은 1·2·4·6줄
    → 3줄이나 5줄만 남은 조각은 (i) 0 과 (ii) 적중으로 FAIL, 2·4·6줄만 남은 조각은 PASS
```

### AC-GDP-027 — `--no-merge` 는 폐기된 no-op으로만

```
GIVEN AC-GDP-026 의 아홉 조각(로컬·템플릿)
WHEN 편집 뒤 --no-merge 를 담은 줄을 검사하면
THEN (i) --no-merge 를 담은 줄은 모두 "no-op" 과 "deprecat" 를 함께 담고
 AND (ii) --no-merge 를 담은 줄 가운데 건너뜀·조건 낱말(skip, not set, prevent, unless)을 담은 줄이 0개다
```

(ii)는 `--no-merge` 가 병합을 건너뛰게 하거나 트리거 조건(`--no-merge flag NOT set`)으로 쓰이는 서술을 잡는다. `--no-merge` 가 어느 조각에도 없으면 두 조건은 공허하게 참이 되므로, 판정 기록에 조각별 `--no-merge` 줄 수를 함께 적는다(REQ-GDP-025는 `--no-merge` 를 호환용 no-op으로 서술할 것을 요구하므로 `delivery.md` Step 3.4 조각에는 1줄 이상 있어야 한다).

대조(기준 트리, 템플릿 사본): `--no-merge` 줄 — dl 조각 8(337행)·19(348행), 나머지 여덟 조각 0(X1~X3 조각 포함). (i) 위반 8·19 → 빨강. (ii) 위반 8("NOT set")·19("Skip") → 빨강.

판정(로컬·템플릿 각각):

```bash
/usr/bin/grep -c -e '--no-merge' $E/ac026-dl.md > $E/ac027-dl-count.txt
# 기대: 1 이상
awk '/--no-merge/ && !(/no-op/ && /[Dd]eprecat/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac027-i.txt
test -s $E/ac027-i.txt
# 기대: exit 1
awk '/--no-merge/ && (/[Ss]kip/ || /[Nn][Oo][Tt] set/ || /[Pp]revent/ || /unless/) {print FILENAME ":" FNR ": " $0}' $E/ac026-skill.md $E/ac026-ref.md $E/ac026-qgc-args.md $E/ac026-qgc-flags.md $E/ac026-sync.md $E/ac026-dl.md $E/ac026-sync-usage.md $E/ac026-hint.md $E/ac026-dl-next.md > $E/ac027-ii.txt
test -s $E/ac027-ii.txt
# 기대: exit 1
```

뮤턴트(plan 작성 시점, 스크래치 픽스처 4줄): "- `--no-merge`: Skip auto-merge even in worktree context." · "- `--no-merge`: Deprecated no-op; skips auto-merge." · "- `is_worktree_context == true` AND `--no-merge` flag not set" · "- `--no-merge`: Deprecated no-op kept for compatibility (logs a warning); not merging is already the default."(올바름) → (i)이 1·3줄, (ii)가 1·2·3줄을 잡고 4줄은 둘 다 잡지 않음.

### AC-GDP-028 — team 모드: 전원 승인 조건

```
GIVEN manager-git.md 로컬·템플릿 사본의 "## PR Auto-Merge" 로 시작하는 절(다음 "## " 제목 앞까지)과 delivery.md 로컬·템플릿 사본의 Step 3.4 절
WHEN 편집 뒤 두 절을 검사하면
THEN (a) 두 절 각각에 "team mode"·"--auto-merge"·"approv" 를 함께 담은 줄이 1개 이상 있고
 AND (b) 두 절 어디에도 "team mode" 를 담으면서 승인 없이 병합한다는 문구(without approval, no approval, approvals not required, regardless of approval)를 담은 줄이 없다
```

검출식은 "team mode" 두 낱말로 판정한다. "team" 한 낱말은 "teammates" 에 걸려 personal·manual 문장(승인할 팀원 없음)을 team 규칙으로 잘못 읽는다.

절 추출(대조·판정 공통):

```bash
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' <manager-git.md> > <mg 절 파일>
awk '/^#### Step 3\.4/{s=1; print; next} s && /^(####|###) /{exit} s' <delivery.md> > <dl 절 파일>
```

대조(기준 트리, 템플릿 사본): mg 절 9줄. (a) mg 0(승인 조건 줄에 모드 이름이 없고 절 제목만 team을 말함), dl 0. (b) 0·0.

판정(로컬·템플릿 각각):

```bash
awk '/^## PR Auto-Merge/{s=1; print; next} s && /^## /{exit} s' .claude/agents/moai/manager-git.md > $E/ac028-mg.md
test -s $E/ac028-mg.md
# 기대: exit 0
awk '/[Tt]eam mode/ && /--auto-merge/ && /[Aa]pprov/ {c++} END{print c+0}' $E/ac028-mg.md > $E/ac028-a-mg.txt
awk '/[Tt]eam mode/ && /--auto-merge/ && /[Aa]pprov/ {c++} END{print c+0}' $E/ac026-dl.md > $E/ac028-a-dl.txt
# 기대: 각각 1 이상
awk '/[Tt]eam mode/ && (/without (any |an )?approv/ || /no approv/ || /approvals? (are |is )?not required/ || /regardless of approv/) {print FILENAME ":" FNR ": " $0}' $E/ac028-mg.md $E/ac026-dl.md > $E/ac028-b.txt
test -s $E/ac028-b.txt
# 기대: exit 1
```

뮤턴트(스크래치 픽스처 6줄): "In team mode, `--auto-merge` merges once checks pass."(승인 없음) · "In team mode, `--auto-merge` merges regardless of approvals." · "In team mode, `--auto-merge` merges only after all approvals are obtained."(올바름) · "In personal and manual modes, `--auto-merge` merges after all approvals." · "In personal mode, `--auto-merge` merges without an approval condition." · "In personal and manual modes, `--auto-merge` merges without an approval condition (no teammates to approve)." → (a)가 2·3줄, (b)가 2줄을 잡음. 1줄만 있는 절은 (a) 0 → FAIL. 6줄은 (a)·(b) 모두 잡지 않음.

### AC-GDP-029 — personal·manual 모드: 승인 조건 없음

```
GIVEN AC-GDP-028 과 같은 두 절(로컬·템플릿)
WHEN 편집 뒤 검사하면
THEN (a) 두 절 각각에 "personal"·"manual"·"--auto-merge" 를 함께 담은 줄이 1개 이상 있고
 AND (b) 두 절 어디에도 personal 또는 manual 을 담고 "approv" 를 담으면서 승인이 없음을 말하지 않는(without approval, no approval, not required, no teammates 가 없는) 줄이 없다
```

대조(기준 트리, 템플릿 사본): (a) mg 절 0, dl 절 0. (b) 0·0.

판정(로컬·템플릿 각각):

```bash
awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $E/ac028-mg.md > $E/ac029-a-mg.txt
awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $E/ac026-dl.md > $E/ac029-a-dl.txt
# 기대: 각각 1 이상
awk '(/[Pp]ersonal/ || /[Mm]anual/) && /[Aa]pprov/ && !(/without (any |an )?approv/ || /no approv/ || /not required/ || /no teammates/) {print FILENAME ":" FNR ": " $0}' $E/ac028-mg.md $E/ac026-dl.md > $E/ac029-b.txt
test -s $E/ac029-b.txt
# 기대: exit 1
```

뮤턴트: AC-GDP-028과 같은 픽스처 → (a)가 4·6줄, (b)가 4줄을 잡음. 5줄(manual 누락)만 있는 절은 (a) 0 → FAIL. 올바른 6줄은 (b)에 걸리지 않음.

### AC-GDP-030 — 명령 원본 편집 뒤 게시본 발행

```
GIVEN 명령 원본 argument-hint 편집(로컬 .claude/commands/moai/sync.md, 템플릿 .claude/commands/moai/sync.md.tmpl)
WHEN run-phase 가 사전 점검에서 발행 점검의 양성 대조를 실행하고,
     원본 편집 직후 make commands-emit 과 make commands-emit-check 를 실행하고,
     원본 편집 커밋이 만들어진 뒤 게시본 변화를 판정하면
THEN (c) 양성 대조: 템플릿 게시본 한 파일을 잠시 바꾼 상태에서 make commands-emit-check 가 exit 1 이고,
     백업으로 되돌린 뒤 cmp 가 exit 0, 다시 실행한 make commands-emit-check 가 exit 0 이며, 게시본 경로의 git status 가 비어 있고
 AND make commands-emit 이 exit 0, 이어서 make commands-emit-check 가 exit 0 이고
 AND 게시본 변화가 아래 두 경우 중 하나로 판정되어 progress §E.2 에 경우 이름과 근거 출력과 함께 기록된다:
     (A) 변화 없음 — 기준 트리 대비 템플릿·로컬 게시본(.agents/skills/) 변경 경로 목록이 비어 있다
     (B) 변화 있음 — 변경 경로가 템플릿 게시본 moai-sync/SKILL.md 와 로컬 사본 .agents/skills/moai-sync/SKILL.md 뿐이고,
         두 게시본 파일을 바꾼 커밋이 모두 템플릿 명령 원본 sync.md.tmpl 을 바꾼 커밋과 같으며,
         로컬 게시본 사본이 템플릿 게시본과 diff exit 0 이고,
         게시본이 AC-GDP-026 (ii)·AC-GDP-027 규칙을 어기는 줄을 담지 않는다
```

**plan 작성 시점 예상: 경우 (A).** 근거: 발행기는 명령 원본의 `argument-hint`·`allowed-tools` 를 게시본에 옮기지 않는다("Claude-only keys and are NOT carried into the published skill", `internal/template/commandemit/loader.go:4-6`). 게시본은 생성 머리말·`name`·영어 `description`·원본 본문만 담는다(`emit.go` `renderSkill`). X2 편집은 3행만 바꾸고 2행 `description` 과 본문을 바꾸지 않는다. 기준 트리의 템플릿·로컬 게시본은 서로 바이트 동일하고 `merge` 를 담지 않는다(`grep -i merge` exit 1). 예상은 판정을 대신하지 않는다 — 두 경우를 모두 명령으로 판정한다. 발행기는 템플릿 게시본만 쓰므로(`golden_test.go` `templatesDir = "../templates"`) 경우 (B)가 되면 로컬 게시본 사본은 템플릿 게시본을 그대로 복사해 같은 커밋에 넣는다.

양성 대조(run-phase 사전 점검, 원본 편집 전):

```bash
cp $T/.agents/skills/moai-sync/SKILL.md /tmp/t622-moai-sync-SKILL.backup.md
printf '\n' >> $T/.agents/skills/moai-sync/SKILL.md
make commands-emit-check > $E/ac030-red.txt 2>&1
# 기대: exit 1 — 발행 점검이 게시본 드리프트를 실제로 잡는다
cp /tmp/t622-moai-sync-SKILL.backup.md $T/.agents/skills/moai-sync/SKILL.md
cmp /tmp/t622-moai-sync-SKILL.backup.md $T/.agents/skills/moai-sync/SKILL.md
# 기대: exit 0
make commands-emit-check > $E/ac030-control-green.txt 2>&1
# 기대: exit 0
git status --porcelain -- $T/.agents/skills/ .agents/skills/ > $E/ac030-control-clean.txt
test -s $E/ac030-control-clean.txt
# 기대: exit 1 — 되돌림 뒤 게시본 경로에 변경 없음
```

판정(원본 편집 직후, 커밋 전):

```bash
make commands-emit > $E/ac030-emit.txt 2>&1
# 기대: exit 0
make commands-emit-check > $E/ac030-check.txt 2>&1
# 기대: exit 0
git status --porcelain -- $T/.agents/skills/ .agents/skills/ > $E/ac030-emit-status.txt
# 비어 있으면 경우 (A) 후보, 비어 있지 않으면 경우 (B) 후보 — 비어 있지 않으면 바뀐 게시본을 원본과 함께 스테이징한다
```

판정(원본 편집 커밋 뒤):

```bash
git diff --name-only $BASE -- $T/.agents/skills/ .agents/skills/ > $E/ac030-changed.txt
# (A): 빈 파일. (B): 두 줄 이하이며 internal/template/templates/.agents/skills/moai-sync/SKILL.md, .agents/skills/moai-sync/SKILL.md 만
git log --format=%H $BASE..HEAD -- $T/.claude/commands/moai/sync.md.tmpl > $E/ac030-src-commits.txt
git log --format=%H $BASE..HEAD -- $T/.agents/skills/moai-sync/SKILL.md .agents/skills/moai-sync/SKILL.md > $E/ac030-artifact-commits.txt
# (A): ac030-artifact-commits.txt 빈 파일. (B): ac030-artifact-commits.txt 의 모든 줄이 ac030-src-commits.txt 에 있음(같은 커밋)
diff .agents/skills/moai-sync/SKILL.md $T/.agents/skills/moai-sync/SKILL.md > $E/ac030-published-lt.diff
# 기대: 두 경우 모두 exit 0
awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && !(/[Dd]eprecat/ && /--auto-merge/) {print FNR ": " $0}' $T/.agents/skills/moai-sync/SKILL.md > $E/ac030-published-flags.txt
test -s $E/ac030-published-flags.txt
# 기대: 두 경우 모두 exit 1
```

판정 기록: `$E/ac030-outcome.md` 에 경우 이름(A 또는 B), `ac030-changed.txt`·`ac030-src-commits.txt`·`ac030-artifact-commits.txt` 내용을 적는다. 경우를 가르는 명령 출력 없이 경우 이름만 적은 기록은 PASS가 아니다.

뮤턴트(판정 논리): 원본을 커밋한 뒤 별도 커밋으로 게시본만 바꾼 이력 → `ac030-artifact-commits.txt` 의 SHA가 `ac030-src-commits.txt` 에 없음 → FAIL. 게시본을 다시 만들지 않고 원본의 `description` 을 바꾼 상태 → `make commands-emit-check` exit 1 → FAIL. 템플릿 게시본만 바뀌고 로컬 사본을 두고 온 상태 → `ac030-published-lt.diff` exit 1 → FAIL.

## §D.2 경계 사례

- `grep -c` 는 줄 수를 센다. SHA는 AC-GDP-015에서 낱말 단위로 센다.
- `sed -n '/A/,/B/p'` 의 끝 제목이 편집으로 바뀌면 절이 파일 끝까지 늘어난다. `awk` 절 추출은 시작 표지가 사라지면 빈 파일을 낸다 — 빈 파일은 판정 불가로 기록한다. 추출 표지(`## Synchronization`, `## PR Auto-Merge`, `### Pre-Spawn Sync Check`, `### Pre-Edit Sync Check`, `#### The sweep prohibition`, `#### Step 3.4`, `##### Worktree Context Detection`, AC-GDP-026의 아홉 표지)는 편집 뒤에도 남아야 한다. `manager-git.md` 절 제목은 "## PR Auto-Merge" 로 시작하기만 하면 뒤의 괄호를 바꿔도 된다.
- 자리표시 기준(AC-GDP-007~012, 017~024)은 판정하지 않고 N/A로 기록한다.
- 의도된 사본 차이가 있는 여섯 파일은 편집으로 줄 수가 바뀌면 차이 줄번호가 밀린다. AC-GDP-013은 줄번호 머리를 빼고 본문만 비교한다. 명령 원본은 로컬 `.md` 와 템플릿 `.md.tmpl` 로 파일 이름 자체가 다르므로 두 경로를 글자 그대로 짝지어 비교한다.
- 검출식에 `\b` 를 쓰지 않는다(POSIX ERE에서 단어 경계가 아니다). AC-GDP-026의 `--merge` 낱말 경계는 앞뒤 문자 클래스로 표현하며 `--merged-only`·`--auto-merge` 는 걸리지 않고, `[--merge]` 처럼 대괄호에 둘러싸인 형태는 걸린다.
- AC-GDP-025의 Frozen 줄 판정은 `[ZONE:Frozen]` 을 담은 diff 줄을 먼저 모은 뒤 `-`·`+` 로 시작하는 줄만 고른다. 첫 grep 에 `-n` 을 붙이면 공허하게 통과한다.
- AC-GDP-026~029는 줄 단위다. 한 조건을 여러 줄에 나눠 적으면 (a)·(i)가 0이 되어 FAIL로 기울고, 검출식이 예상하지 않은 표현은 통과할 수 있다(spec.md §E.2).
- AC-GDP-030의 양성 대조는 추적 파일을 잠시 바꾼다. 되돌림을 `cmp` 와 `git status` 로 확인하지 못하면 사전 점검을 멈추고 보고한다.
- `go test -run` 선택자는 `^…$` 로 고정하고, 최상위 PASS 줄 수와 범위 파일 하위 테스트 흔적을 함께 본다.

## §D.3 품질 게이트

- 사전 점검의 양성 대조가 모두 기대값을 냈다는 기록이 있어야 판정이 유효하다(AC-GDP-030의 발행 점검 대조 포함).
- **읽기 단계가 있는 기준(AC-GDP-001, AC-GDP-006)은 읽기 기록 파일(`$E/ac001-reading.md`, `$E/ac006-reading.md`)이 존재하고 대상 문단·절의 모든 질문에 답했을 때만 PASS다.** 자동 검출이 통과해도 읽기 기록이 없거나 한 문단이라도 답이 비면 PASS가 아니다.
- `go test` 선택 실행이 최상위 PASS 줄 2개를 내지 않으면 합격이 아니다. 로컬 전체 스위트는 돌리지 않는다.

## §D.4 완료 정의 (Definition of Done)

- 판정 대상 기준 AC-GDP-001~006, 013~015, 025~030이 PASS이고 AC-GDP-016이 PASS 또는 사유 기록. 자리표시 기준은 N/A.
- AC-GDP-030의 경우 이름(A 또는 B)과 근거 출력이 progress 기록에 남는다.
- 모든 증거 파일이 `.moai/reports/t622/run/` 에 커밋되어 인용 경로가 해석된다.
