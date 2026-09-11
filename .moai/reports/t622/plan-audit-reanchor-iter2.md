# SPEC 재고정 감사 2회차: SPEC-GIT-DELIVERY-PROCEDURE-001 (0.2.6, 수리분 한정)

Iteration: 재고정 변경분 감사 2회차 (범위: 1회차 D1~D6 의 해소 여부와 수리 커밋 `33e0dac9c` 가 건드린 절의 퇴행)
Verdict: **PASS**
Overall Score: 0.875 (Tier M 기준 0.80 이상. 1회차 0.75 에서 올랐으므로 STOP 신호 없음)

작성자 추론 맥락은 M1 맥락 격리에 따라 무시했다. 수리 증거(`reanchor-fix/`)는 주장으로만 읽고, 판정은 이 감사에서 직접 실행한 명령의 출력으로 했다.

- 측정 트리: 워크트리 `.claude/worktrees/t622`, HEAD `33e0dac9ca898a3f62c849f722c8f490d0a803a2` (브랜치 `WT-git-procedure-fixes`)
- 같은 시점 ref: 로컬 `develop` = `origin/develop` = `85868148c30f39d654694f61d1424b2bb5a4f456`. `git rev-list --count --left-right develop...HEAD` → `25 28` (1회차 `13 26`, 0.2.6 작성 시점 `17 27` 에서 더 앞섰다). `git merge-base --all develop HEAD` → `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf` 1줄
- 재현 파일: `.moai/reports/t622/reanchor-audit2/` (`refs.txt` 에 위 세 값)

## Must-Pass 결과 (수리분 기준)

- [PASS] MP-1 REQ 번호: 요구사항 층. `grep -c -E '^\*\*REQ-GDP-[0-9]+\*\*' spec.md` → 26 (1회차와 같음). 새 번호 없음.
- [PASS] MP-2 GEARS: 요구사항 층. 바뀐 REQ-GDP-002(`spec.md:153-154`) "…코드 블록은 … 시작되도록 지시해야 하며, … 나열해서는 안 된다" — Ubiquitous(shall + shall not 복합, REQ-GDP-001 과 같은 형태). AC 는 검증 층이라 채점하지 않았다.
- [PASS] MP-3 프론트매터: `spec.md:4` `version: "0.2.6"`, `:5` `status: draft`, `:7` `updated: 2026-09-11`, `:15` `tier: M`. 나머지 필드는 수리분에서 바뀌지 않았다.
- [N/A] MP-4: 수리분은 템플릿 내용을 쓰지 않는다.
- [PASS] MP-5 D7: 수리분 diff 의 SPEC-ID 는 자기 ID 17회와 기존 `related_specs` 문맥 줄의 `SPEC-MERGE-METHOD-CONFIG-001` 3회뿐. 새 참조 없음.
- [PASS] MP-6 D8: 수리분 diff 의 `syscall` 0회.
- [PASS] MP-7: `grep -c '\[NEEDS CLARIFICATION' plan.md` → 0. research.md 없음.

## 영역별 점수

| 항목 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | `acceptance.md:173` "(a)는 REQ-GDP-002 의 순서 문장을 … 판정한다" 가 (a) 단독의 힘을 과장한다(N1, 뮤턴트 R3) |
| Completeness | 1.00 | 1.0 | 재흡수 절차가 `plan.md` §C 1단계·"통합 창 흡수 뒤 재측정", `spec.md` §E.2, `acceptance.md:17-36`, §D.3 에 모두 들어갔다 |
| Testability | 0.75 | 0.75 | 카드 경로에서는 AC-GDP-002 가 이진 판정이다. 스냅숏 재반출 분기에서는 (b)가 자기 자신과 비교하게 된다(N1) |
| Traceability | 1.00 | 1.0 | 판정 대상 AC 16(`acceptance.md:74`, 표 활성 행 16), REQ 26 표지 유지, REQ-GDP-002→AC-GDP-002 |

## 점검 1 — D1: 범위 판정의 왼쪽 끝

**Claim**: 해소됐다. "이 카드가 바꾼 것·커밋한 것" 을 재는 판정은 모두 읽는 시점 형태를 쓴다. 현재 트리와 현재 develop 을 흉내 흡수한 트리에서 카드 커밋만 골라낸다.

**Evidence — 네 파일 소탕**:
- `grep -n -E '\$BASE\.\.|255f88eb0[0-9a-f]*\.\.|diff (--name-only |--stat )?\$BASE|log [^`]*\$BASE|BASE\.\.HEAD' spec.md plan.md acceptance.md progress.md` → 남은 줄은 모두 이력·금지 서술(`spec.md:345`, `plan.md:170`, `acceptance.md:568·598·656·671·722·1087`, `progress.md:9`)이거나 신선도 점검(`acceptance.md:25·32`, `spec.md:346`, `plan.md:51`, 형태 `git diff --name-only $BASE $CARD_BASE`)이다. 리터럴 `$BASE` 범위 판정은 0개.
- `$BASE`/`255f88eb0` 를 담은 명령 줄 전수(`acceptance.md:8·25·32·161·198·240·242·284·431·434·659·714·788·929`, `plan.md:51·53`): 반출(`git show $BASE:`), 고정 과거 구간 양성 대조(`:659` `255f88eb0 --not b412f8a33`), 기준 절·사본 덩어리·미러 기준선·clause 개수 같은 고정 스냅숏뿐이다.
- 판정 형태: AC-GDP-014 `git diff --name-only develop...HEAD`(`acceptance.md` AC-014 판정), AC-GDP-015 `git diff develop...HEAD -- $T/`, AC-GDP-016 `git log --no-merges --format=%H HEAD --not develop`, AC-GDP-025 `git diff develop...HEAD -- <18경로>`, AC-GDP-030 `git log --format=%H HEAD --not develop`·`git diff --name-only develop...HEAD`, plan M5(`plan.md` M5 둘째 항목). 범위 대조 `git diff --name-only develop...HEAD` 1줄 이상, 0줄이면 "측정 불가"(`acceptance.md:17-21`, §D.3). `$(…)` 대신 `git merge-base --all develop HEAD > $E/card-base.txt` 두 단계 형태(`acceptance.md:15`, `plan.md` §C 1단계).
- 신선도 점검이 판정보다 먼저 온다(`acceptance.md:22-36`, `plan.md` §C 1단계 넷째 항목).

**Evidence — 현재 트리 재실행** (`reanchor-audit2/`):

| 명령 | 관측 |
|---|---|
| `git log --no-merges --format='%h %s' HEAD --not develop` | 24줄(`card-commits.txt`). 전부 이 카드 커밋 — 14줄은 `t622`, 나머지 10줄은 `SPEC-GIT-DELIVERY-PROCEDURE-001` 를 제목에 담는다. develop 이 앞선 25커밋은 한 줄도 없다 |
| `git diff --name-only develop...HEAD` | 309줄, 전부 `.moai/` 아래(`card-range-names.txt`, `grep -c -v '^\.moai/'` → 0). 0.2.6 작성 시점 192줄 + 수리 커밋의 보고서 117개 |
| AC-GDP-016 판정 `git log --no-merges --format=%H HEAD --not develop -- <acp 두 사본>` | exit 0, `test -e` 0, `test -s` 1 (`ac016-acp-commits.txt`) |
| AC-GDP-016 양성 대조 `255f88eb0… --not b412f8a33… -- <acp 두 사본>` | `97ef8e3023e9…`, `6896eef3766a…` 2줄 (`ac016-control.txt`) |

**Evidence — 현재 develop `85868148c` 흉내 흡수** (커밋·ref 없이 트리 객체만 만듦):

| 명령 | 관측 |
|---|---|
| `git merge-tree --write-tree HEAD develop` | exit 0, 트리 `7b91d60b1bdf4bb2b216e12c9f6b91a41dc95c96` (`emu-merge-tree.txt`) |
| 흡수 뒤 `develop...M` 에 해당하는 `git diff --name-only 85868148c 7b91d60b1` | `card-range-names.txt` 와 `cmp` exit 0 — 흡수 전후 카드 범위가 같다 |
| 흡수 뒤 `M --not develop` 에 해당하는 `git log --no-merges HEAD develop --not develop` | 현재 `HEAD --not develop` 과 `cmp` exit 0 |
| 흡수 뒤 AC-015 형태 `git diff 85868148c 7b91d60b1 -- internal/template/templates/` | `test -e` 0, `test -s` 1 |
| 대조: 리터럴 `git diff --name-only $BASE 7b91d60b1 -- internal/template/templates/ .claude/ .agents/` | 7개 파일(`trace-ledger.sh`, `spec-workflow.md` 두 사본 등) — 리터럴 형태였다면 develop 변경을 카드 것으로 읽었다 |
| 리터럴 `git log --no-merges 255f88eb0..develop` | 13줄 — 리터럴 `$BASE..HEAD` 가 흡수 뒤 추가로 담을 develop 커밋 수 |
| 흡수 뒤 신선도 점검(`$CARD_BASE`=`85868148c`, 21경로) | `test -e` 0, `test -s` 1 (`snapshot-stale-postabsorb.txt`) |
| 흡수 뒤 미러 기준선 점검 | 8줄, `spec-workflow.md` 두 사본 포함(`snapshot-stale-mirror-postabsorb.txt`) — 수리 증거의 흉내 흡수 값과 같다 |

**판정**: RESOLVED. 카드 범위 판정은 develop 이 더 앞서도 카드 커밋만 담고, 흡수 전후 결과가 바이트 동일하다. 고정 스냅숏은 신선도 점검 뒤에만 쓰도록 적혔다.

## 점검 2 — D2: AC-GDP-002 판정기

**Claim**: 카드 경로에서는 해소됐다. (a)+(b) 조합은 이 카드의 편집이 절 안에서 REQ-GDP-002 를 어기는 경우를 모두 막는다. 절 밖 문장(ix)은 받아들일 수 있는 한계다. 새로 찾은 생존 경로는 흡수로 절이 바뀌어 기준 절을 `$CARD_BASE` 에서 다시 반출하는 분기뿐이다(N1, optional).

**Evidence — 판정기 원문 재실행**: `acceptance.md:195` 의 awk 프로그램을 `sed` 로 떼어 `reanchor-fix/ac002/judge.awk` 와 `cmp` exit 0. 워크트리 가드가 `awk -f` 를 거부해서 acceptance.md 에 적힌 대로 인라인 한 줄로 실행했다(`reanchor-audit2/ac002/judge-rerun.txt`):

```
block-local.md             fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS
block-template.md          fetch=1 revlist=1 joined=0 J=0 F=2 S=3 C=4 X=6 R=8 order=PASS shape=PASS
mut-i-old.md               ... order=FAIL shape=FAIL
mut-ii-joined.md           ... J=2 ... order=PASS shape=FAIL
mut-iii-nostatus.md        ... order=FAIL shape=FAIL
mut-iv-reorder.md          ... order=FAIL shape=FAIL
mut-v-background.md        ... F=0 ... order=FAIL shape=FAIL
mut-vi-pipe.md             ... F=0 ... order=FAIL shape=FAIL
mut-vii-noabort.md         ... X=0 ... order=PASS shape=FAIL
mut-viii-laneb-serial.md   ... order=PASS shape=PASS
mut-ix-block.md            ... order=PASS shape=PASS
probe-joined-andand.md     joined=1 J=2 order=PASS      probe-joined-background.md joined=1 J=0 order=FAIL
```

`acceptance.md` §D.1 표와 줄마다 일치한다.

**Evidence — (b) 절 diff**: 기준 절을 직접 반출(`git show 255f88eb0…:<로컬·템플릿 acp>` → `sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p'`, 둘 다 52줄, 로컬·템플릿 기준 절끼리 diff exit 0). 현재 로컬 절 diff exit 0, 현재 템플릿 절 diff exit 0. 수리 증거의 절 픽스처를 이 감사의 기준 절과 비교: `ctl-section-current.md` exit 0, 뮤턴트 i~viii 여덟 개 모두 exit 1, `mut-ix-section.md` exit 0. 픽스처가 실제 뮤턴트인지 확인: `mut-v-background.section.diff` 는 12행 `git fetch origin main 2>&1 &`, `mut-viii-laneb-serial.section.diff` 는 20-21행 Lane B 주석 교체.

→ AC-GDP-002(= (a) AND (b)): 현재 두 사본 PASS, i~viii 전부 FAIL. 1회차 생존자 v·vi·vii·viii 가 모두 잡힌다.

**새 뮤턴트 시도** (`reanchor-audit2/ac002/`):

| 시도 | 모양 | (a) | (b) | AC-GDP-002 |
|---|---|---|---|---|
| direct-table | 절 안 `N 0` 행 동작을 "Proceed normally" 로 | — | exit 1 | FAIL |
| direct-fetchsuffix | 절 안 fetch 줄 끝에 ` && true` | — | exit 1 | FAIL |
| R3 bggroup | fetch·상태 검사를 `{ … } &` 로 묶어 백그라운드, rev-list 는 바로 다음 | `order=PASS shape=PASS` (`mutR3-judge.txt`) | 절 안 편집이라 exit 1 | FAIL (카드 경로) |
| R1 흡수 후 재반출 | develop 이 Lane B 주석을 직렬화로 바꿔 흡수됨(`mutR1-absorbed-acp.md`, 전체 diff 309-310행) | `order=PASS shape=PASS` | 재반출 기준 절 대비 exit 0 (`$BASE` 절 대비는 exit 1) | **PASS** — REQ-GDP-002 셋째 문장 위반 |
| R2 흡수 후 재반출 | develop 이 `N 0` 행을 바꿔 흡수됨(`mutR2-absorbed-acp.md`, 325행) | `order=PASS shape=PASS` | 재반출 절·표 행 대비 exit 0 (`$BASE` 대비 exit 1) | **PASS** — 두 해석 표 불변 문장 위반 |

(b)는 절 전체의 바이트 동일성이라 절 안의 편집으로는 통과할 방법을 찾지 못했다 — 구조상 그렇다. 통과하는 것은 두 부류다. (1) 절 밖 문장(ix). (2) 흡수가 절을 바꾸고, SPEC 절차(`acceptance.md:30`, `spec.md:347`)대로 기준 절을 `$CARD_BASE` 에서 다시 반출하면 (b)가 흡수된 절을 자기 자신과 비교하게 되는 분기. 이때 남는 판정은 (a)뿐인데, R3 이 보이듯 (a)는 순서 위반도 받는다.

**ix 판단 — 받아들일 수 있는 한계**: REQ-GDP-002 는 "Pre-Spawn Sync Check 코드 블록" 을 묶는다(`spec.md:154`). ix 는 절 밖 문장이라 글자로는 REQ 를 어기지 않는다. 이 카드가 넣으면 `plan.md` §D 가 acp 편집을 금지하고, AC-GDP-016 이 잡는다(`reanchor-fix/ac002/mut-ix-wholefile.diff` 에서 파일 전체 diff exit 1 확인). 흡수로 들어오면 신선도 점검에 acp 가 나와 리드로 올라간다. 한계 서술(`acceptance.md:272`, `spec.md:347`)이 정확하다. 덧붙일 점은 AC-GDP-016 이 SHOULD-PASS 라서 기계적으로는 출시 차단 자격이 없다는 것뿐인데, §D 금지 조항이 같은 행위를 막으므로 결함으로 세지 않는다.

**판정**: D2 RESOLVED. 재반출 분기의 생존(R1·R2)과 (a) 단독 판정력 과장(R3)은 N1 로 따로 적는다 — 이 분기는 카드가 일으킬 수 없고 SPEC 이 리드 보고로 돌리므로 blocking 이 아니다.

## 점검 3 — D3: REQ-GDP-002 실질과 Pre-Edit 한 줄

**Evidence**:
- 0.2.4 원문(`git show ccfe3005e:…/spec.md`, `version: "0.2.4"`, 131-132행): "…`git fetch origin main` 과 `git rev-list …` 를 순서가 보장되는 한 명령으로 지시해야 한다."
- 현재(`spec.md:153-154`): "…`git fetch origin main` 이 끝난 뒤에 `git rev-list …` 가 시작되도록 지시해야 하며, 두 명령을 서로 독립인 병렬 항목으로 나열해서는 안 된다." 셋째 명령 병렬 허용·두 표 불변 문장은 그대로다.
- exit status 관측과 실패 시 `exit` 는 비규범 주석으로 옮겨졌다(`spec.md` 0.2.6 인용 블록). AC-GDP-002 (a)는 그것을 요구하지 않고 `shape` 로 진단만 한다(`acceptance.md:176-177`). (b)가 착지 절 전체를 보존으로 지킨다.
- `spec.md:74`: Pre-Edit 의 `fetch 2>&1; rev-list` 한 줄은 "정상". 이 모양은 REQ-GDP-002 의 순서 성질을 충족하고 (a)도 `;`·`&&` 한 줄을 받는다(`probe-joined-andand.md` → `order=PASS`). Pre-Edit 절은 REQ-GDP-003 보존 대상이라고 적었다. acp 296행 Lane A 문장("…a shell line whose completion cannot be distinguished")은 착지 블록의 추가 성질로 설명되며, `grep -n` 으로 그 문장이 296행에 있음을 확인했다.
- HISTORY 0.2.5 행에 "[0.2.6 정정: …좁았다 …되돌렸다]" 괄호가 들어갔다.

**판정**: RESOLVED. 0.2.4 의 "한 명령" 은 한 가지 모양이고, 결함 AC-11 의 실질(fetch 완료 전 rev-list 금지, 독립 병렬 나열 금지)이 복원됐다. REQ-GDP-002, AC-GDP-002 (a), `spec.md:74` 가 서로 맞는다. (a)는 다른 줄 모양에 `fetch_status=$?` 를 요구해 REQ 보다 약간 엄격하지만, 그 이유(블록이 fetch 완료를 소비해야 독립 항목이 아니다)가 `acceptance.md:176` 에 적혀 있고, 이 카드가 파일을 고치지 않으므로 거짓 FAIL 은 흡수로만 생긴다(그때는 신선도 점검이 먼저 잡는다).

## 점검 4 — D4~D6

- D4 RESOLVED: `acceptance.md:181` "MUST-PASS 근거 (0.2.6, 재고정 감사 D4)" — §2 가 근거가 아님을 밝히고 §4 증거 고정(preserved-surface)으로 둔다. `verification-completeness.md:232` `## 4. Evidence pinning`, 234행 "Invariant assertions — byte-unchanged, preserved-surface" 로 인용이 성립한다.
- D5 RESOLVED: `plan.md:63` "비교 전에 양쪽을 `sort` 한다". 재현: `diff reanchor-audit/post-sets.txt reanchor/mirror-baseline-sets.txt` exit 1, 둘을 `sort` 한 `reanchor-audit2/d5-*.sorted` 끼리 exit 0.
- D6 RESOLVED: `acceptance.md:650` "GIVEN 이 카드의 커밋들(HEAD --not develop, 병합 커밋 제외 — plan·run·sync 커밋 모두)".

## 점검 5 — 범위·개수·lint

- 판정 대상 AC 16(`acceptance.md:74`, 표 활성 행 `grep -c` → 16). REQ 표지 26, 새 REQ 없음. 편집 대상 파일 수·범위 서술은 수리분에서 바뀌지 않았다(`plan.md:7`).
- `moai spec lint SPEC-GIT-DELIVERY-PROCEDURE-001`: 설치본 `/Users/goos/go/bin/moai` `v3.2.0-rc.7`, 빌드 `ed71054d3-dirty` — `git merge-base --is-ancestor ed71054d3 HEAD` exit 0 이라 HEAD 보다 오래된 빌드다. exit 0, 0 error 1 warning(`spec-lint.txt`). 그래서 이 트리에서 `go build -o <scratchpad>/moai-tree ./cmd/moai`(exit 0)로 만든 바이너리로 다시 돌렸다: exit 0, 0 error 1 warning(`spec-lint-tree-build.txt`). 경고는 `spec.md:349` 자리표시 REQ 표(§G.1)의 `REQTableRowsRejected` 로 수리분과 무관하다.

## Defects Found

D1~D6(1회차): 모두 RESOLVED — 아래 Regression Check.

N1. AC002-REEXPORT — `acceptance.md:30`(신선도 점검의 재반출 지시), `acceptance.md:173`("(a)는 … 순서 문장을 … 판정한다"), `acceptance.md:274`(재반출 시 읽기 기록 범위), `spec.md:347` — 흡수가 Pre-Spawn 절을 바꿔 기준 절을 `$CARD_BASE` 에서 다시 반출하면 (b)는 흡수된 절을 자기 자신과 비교해 늘 exit 0 이 되고, 남는 (a)는 REQ-GDP-002 위반을 통과시킨다. 증거: R1(Lane B 직렬화)·R2(`N 0` 행 변경)는 재반출 기준 대비 (b) exit 0·(a) `order=PASS` 로 AC-GDP-002 PASS, `$BASE` 절 대비로는 둘 다 exit 1. R3(`{ fetch; 상태 검사 } &` 뒤 rev-list)은 `order=PASS shape=PASS` — (a) 단독으로는 순서 위반을 받는다. 재반출 시 읽기 의무는 `if` 경계·비교 방향만 적혀 있다(`acceptance.md:274`). 이 분기는 카드가 일으킬 수 없고 리드 보고로 이어지므로 카드 경로의 판정은 옳다 — Severity: minor — Class: optional — Required fix: (1) `acceptance.md:173` 을 "(a)는 순서 모양 선별이며, 순서 판정은 (a)와 (b)가 함께 한다" 로 고친다. (2) 신선도 점검에 `agent-common-protocol.md` 가 나와 기준 절을 재반출한 경우, `diff <$BASE 기준 절> <$CARD_BASE 기준 절>` 을 기록하고 그 차이를 REQ-GDP-002 의 세 문장(fetch 완료 뒤 rev-list, 독립 병렬 나열 금지, Lane B 병렬 허용·두 표 불변)마다 읽어 답한 기록이 있어야 AC-GDP-002 를 PASS 로 적는다고 `acceptance.md:274` 에 한 문장 더한다. 두 해석 표 행 비교는 재반출과 무관하게 `$BASE` 표 행을 기준으로 두면 R2 는 기계적으로 잡힌다.

## Regression Check (1회차 결함)

- D1 RANGE-ABSORB — RESOLVED: 네 파일 소탕 결과 리터럴 `$BASE` 범위 판정 0. 현재 트리(develop 25커밋 앞섬)에서 `HEAD --not develop` 24줄 전부 카드 커밋, 현재 develop 흉내 흡수 전후 범위 이름·커밋 목록 `cmp` exit 0, 리터럴 형태였다면 develop 파일 7개·커밋 13개가 섞였다. 범위 대조·"측정 불가"·병합 전 전용·신선도 점검이 모두 적혔다.
- D2 AC002-MUTANT — RESOLVED: 판정기 원문 재실행이 §D.1 표와 일치, 뮤턴트 i~viii 전부 AC-GDP-002 FAIL(v·vi 는 (a)와 (b), vii·viii 는 (b)). 절 안 새 시도 둘도 FAIL. 남은 생존 경로는 절 밖(ix, 한계로 수용)과 재반출 분기(N1, optional).
- D3 REQ002-SUBSTANCE — RESOLVED: `spec.md:153-154` 가 순서 성질로 복원됐고 abort 요구는 비규범 주석과 (b)로 옮겨졌다. `spec.md:74` 와 (a)의 `;`·`&&` 수용이 맞는다.
- D4 AC002-CLASS — RESOLVED (`acceptance.md:181`).
- D5 MIRROR-ORDER — RESOLVED (`plan.md:63`, 재현 정렬 전 exit 1·정렬 뒤 exit 0).
- D6 AC016-GIVEN — RESOLVED (`acceptance.md:650`).

수리가 건드린 절(공통 관례, AC-GDP-002·014·015·016·025·030, plan §C 1·2·8단계·M5·§D, spec §A.1·REQ-GDP-002·§E.2)에서 새 blocking 퇴행은 찾지 못했다.

## Recommendation

PASS 근거: must-pass 일곱 항목이 모두 PASS 또는 N/A 이고(인용은 위 Must-Pass 절), 1회차 blocking 세 건(D1·D2·D3)이 이 트리에서 직접 재측정으로 해소됐다. 영역별 점수 합산 0.875 는 Tier M 기준 0.80 을 넘는다.

N1 은 선택 사항이다. 비용이 두 문장 수준이라 run-phase 착수 전에 함께 반영할 것을 권하지만, 반영하지 않아도 이 카드가 측정하는 경로의 판정은 옳다.

## 증거 목록

`.moai/reports/t622/reanchor-audit2/` — `refs.txt`, `card-base.txt`, `card-commits.txt`, `card-commits-h.txt`, `card-range-names.txt`, `ac016-acp-commits.txt`, `ac016-control.txt`, `emu-merge-tree.txt`, `emu-card-range-names.txt`, `emu-card-commits.txt`, `emu-literal-base-scope-names.txt`, `emu-ac015-template.diff`, `snapshot-stale-postabsorb.txt`, `snapshot-stale-mirror-postabsorb.txt`, `d5-post.sorted`, `d5-base.sorted`, `spec-lint.txt`, `spec-lint-tree-build.txt`, `ac002/`(`judge.awk`, `judge-rerun.txt`, `base-acp*.md`, `section-*.md`·`.diff`, `block-*.md`, `mutR1-*`, `mutR2-*`, `mutR3-*`, `mutR-judge.txt`, `try-direct-*`).

## Gaps

- 흡수는 `git merge-tree --write-tree` 로 만든 트리 객체로만 흉내 냈다(커밋·ref 이동 없음). 흡수 뒤 `develop...M`·`M --not develop` 은 merge-base 가 develop tip 이 된다는 git 의미에 기대어 `git diff develop-tip <tree>`·`git log HEAD develop --not develop` 로 대신 쟀다.
- AC-GDP-014·025·030 판정은 편집 전이라 빈 결과 쪽만 재실행했다. 양성 대조(dr0911·t645)는 수리 증거의 값을 다시 돌리지 않았다.
- R1·R2 는 흡수 분기를 흉내 낸 파일 픽스처다. 실제 develop 이 acp 를 바꾼 흡수는 관측하지 않았다.
- 교차 모델 감사(codex·glm)는 이번 제한 감사에서 호출하지 않았다.

## Residual-risk

- 로컬 develop 을 병합 뒤 재생성(`gitflow-lane-protocol.md` §11)해 흡수한 커밋이 develop 에서 사라지면 merge-base 가 흡수 전 분기점으로 돌아가 범위가 넓어질 수 있다. 범위 대조는 이 경우를 잡지 못한다(1줄 이상이면 통과).
- 판정기는 여전히 줄 단위이며, (b)가 없는 분기에서는 R3 같은 블록 구조 위반을 보지 못한다(N1).
