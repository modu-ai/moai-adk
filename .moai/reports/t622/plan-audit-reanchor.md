# SPEC 재고정 감사 보고서: SPEC-GIT-DELIVERY-PROCEDURE-001 (0.2.5, 변경분 한정)

Iteration: 재고정 변경분 감사 1회 (범위: `b24f2e184..bca80d7c5` 의 SPEC 네 파일 변경분과 그 파생 항목)
Verdict: **FAIL** — 변경분에 blocking 결함 2건(D1·D2). 수리는 acceptance.md·spec.md·plan.md 문구와 판정 명령에 한정되며 REQ 추가·범위 변경은 필요 없다.
Overall Score: 0.75 (Tier M 기준 0.80 미달)

작성자 추론 맥락은 M1 맥락 격리에 따라 무시했다. 판단 근거는 SPEC 파일, 트리, 이 감사에서 직접 실행한 명령의 출력이다.

- 측정 트리: 워크트리 `.claude/worktrees/t622`, HEAD `bca80d7c5cce45377f23fbfd0a3b7b6026497f0d`, BASE `255f88eb08df0d2cbb9f991f28aa8d9c2bd6f089`
- 같은 시점 ref: 로컬 `develop` = `0db675bedcae69c7ade2f02f106649a715fb14f6`, `origin/develop` = `b155c95f9942206d820ad3b7b4fc1cb132f41bb4`, `git merge-base develop HEAD` = `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf`
- 재현 파일: `.moai/reports/t622/reanchor-audit/`

## Must-Pass 결과 (변경분 기준)

- [PASS] MP-1 REQ 번호: 요구사항 층에서 판정. `spec.md` 의 `**REQ-GDP-0NN**` 표지 26개(001~026, 활성 12 + 자리표시 14), 새 번호 없음. `grep -c -E '^\*\*REQ-GDP-[0-9]+\*\*' spec.md` → 26.
- [PASS] MP-2 GEARS 형식: 요구사항 층에서 판정. 바뀐 REQ-GDP-002(`spec.md:152-153`)는 "…코드 블록은 … 지시해야 한다"(Ubiquitous), REQ-GDP-003(`spec.md:157-158`)은 "…수정되어서는 안 된다"(Unwanted). AC는 검증 층이라 이 항목에서 채점하지 않았다.
- [PASS] MP-3 프론트매터: `spec.md:4` `version: "0.2.5"`, `:5` `status: draft`, `:7` `updated: 2026-09-11`, `:15` `tier: M`. 나머지 필드는 변경분에서 바뀌지 않았다.
- [N/A] MP-4 언어 중립성: 변경분은 템플릿 내용을 새로 쓰지 않는다(이 카드는 `agent-common-protocol.md` 를 편집하지 않음).
- [PASS] MP-5 D7: 변경분이 새로 참조한 SPEC-ID 없음.
- [PASS] MP-6 D8: 변경분에 `syscall` 없음.
- [PASS] MP-7 확인 게이트: `grep -c '\[NEEDS CLARIFICATION' plan.md` → 0. research.md 없음.

## 영역별 점수

| 항목 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | REQ-GDP-002 의 "실질 동일" 주장이 정확하지 않다(D3, `spec.md:155`, `spec.md:34`) |
| Completeness | 0.75 | 0.75 | 통합 창의 develop 재흡수 뒤 판정 방법이 빠졌다(D1, `spec.md:335`, `acceptance.md:1055`, `plan.md:47`) |
| Testability | 0.50 | 0.50 | AC-GDP-002 가 요구사항을 어기는 뮤턴트를 통과시킨다(D2). AC-GDP-014·015·016·030 은 재흡수 뒤 잘못된 판정을 낸다(D1) |
| Traceability | 1.00 | 1.0 | REQ-GDP-002→AC-GDP-002, REQ-GDP-003→AC-GDP-003, AC-GDP-016 은 절차 점검으로 명시(`acceptance.md:44`). 판정 대상 AC 16개 유지(`acceptance.md:53`) |

## 점검 1 — REQ-GDP-002 새 문구와 실제 블록

**Claim**: 새 문구는 BASE 와 현재 트리에서 참이다. 다만 0.2.4 와 실질이 같지는 않다(좁아졌다). REQ-GDP-003 과는 충돌하지 않는다.

**Evidence**:
- 두 사본 블록 추출(`awk '/^### Pre-Spawn Sync Check/{s=1} s && /^```bash/{b=1; next} b && /^```/{exit} b' <파일>`) → `reanchor-audit/block-local.md`, `block-template.md`. `cmp` 두 블록 exit 0, 파일 전체 `cmp` exit 0.
- 블록 줄(로컬): 2 `git fetch origin main 2>&1` · 3 `fetch_status=$?` · 4-7 `if … -ne 0 … exit "$fetch_status"` · 8 `git rev-list --count --left-right origin/main...HEAD` · 12 `moai session list --json --filter-spec=<SPEC-ID>`.
- 파일 줄 인용: `agent-common-protocol.md:296` "Lane A (ordered): `git fetch origin main` MUST finish and its exit status be observed before …", `:297` "Lane B (independent): … may run concurrently", `:301`·`:302`·`:307`·`:311` 가 SPEC §A.1 표와 일치.
- `git diff --stat 255f88eb0… -- <acp 두 사본>` → 빈 출력(현재 트리 = BASE).

**판정**: BASE 에서 "fetch 완료 → exit status 관측 → rev-list" 가 참이다. 그러나 0.2.4 문구("순서가 보장되는 한 명령")는 exit status 관측을 요구하지 않았다. 새 문구에서는 SPEC 이 스스로 "정상 — 순서 보장"이라 부르는 한 줄 형태(`spec.md:73`, `git fetch origin main 2>&1; git rev-list …`)가 REQ-GDP-002 를 충족하지 못한다. `;` 는 exit status 를 관측하지 않기 때문이다. 요구가 좁아진 것이지 같은 것이 아니다(D3). REQ-GDP-003(Pre-Edit 절 불변)은 Pre-Spawn 만 묶는 REQ-GDP-002 와 부딪히지 않는다. Pre-Edit 절은 현재 트리 = BASE 이고, AC-GDP-003 대조를 b412↔BASE Pre-Spawn 절로 바꾼 것도 성립한다(`reanchor/ac003-control-section.diff` 비어 있지 않음).

## 점검 2 — AC-GDP-002 판정기

**Claim**: 판정기는 기존 뮤턴트 4개를 잡지만, REQ-GDP-002 를 어기는 뮤턴트 2개를 통과시킨다. 분류 서술은 대체로 맞지만 등급과의 관계가 흐리다.

**Evidence** (`acceptance.md:161` 판정기를 그대로 실행, `reanchor-audit/judge-new-mutants.txt`):

```
block-local.md     : fetch=1 revlist=1 joined=0 F=2 S=3 C=4 X=6 R=8 verdict=PASS
block-template.md  : fetch=1 revlist=1 joined=0 F=2 S=3 C=4 X=6 R=8 verdict=PASS
(i)~(iv)           : 모두 verdict=FAIL — 기존 기록(reanchor/ac002-judge-results.txt)과 같다
mut-v-background.md: fetch=1 revlist=1 joined=0 F=2 S=3 C=4 X=6 R=8 verdict=PASS
mut-vi-pipe.md     : fetch=1 revlist=1 joined=0 F=2 S=3 C=4 X=6 R=8 verdict=PASS
mut-vii-noabort.md : fetch=1 revlist=1 joined=0 F=2 S=3 C=4 X=6 R=8 verdict=PASS
mut-viii-laneb-serial.md: … verdict=PASS (session-count=1)
```

- v: `git fetch origin main 2>&1 &` — fetch 가 백그라운드로 가서 끝나기 전에 rev-list 가 돈다. `$?` 는 백그라운드 시작의 0. REQ-GDP-002 위반.
- vi: `git fetch origin main 2>&1 | tee …` — `$?` 는 tee 의 상태라 fetch 의 exit status 를 관측하지 않는다. REQ-GDP-002 위반.
- vii: `exit "$fetch_status"` 를 `exit_note="…"` 로 바꿈 — 판정기의 `^[[:space:]]*exit` 가 `exit_note=` 에 걸린다. 실패해도 멈추지 않는 블록이 PASS.
- viii: Lane B 주석을 "Lane A 가 끝나기 전 시작 금지"로 바꿈 — REQ-GDP-002 의 "세 번째 명령의 병렬 실행 허용은 바뀌어서는 안 된다" 위반. 판정기는 주석을 건너뛰고 session 줄 개수만 본다.

`verification-completeness.md §2` Mutant probe("요구를 어기면서 기준을 통과하는 뮤턴트를 쓸 수 있으면 채택하기엔 얕다")에 걸린다(D2).

**보강안 관측** (`reanchor-audit/judge-hardened-proposal.txt`): fetch 줄을 `^[[:space:]]*[g]it fetch origin main( 2>&1)?[[:space:]]*$` 로 끝까지 고정하고 `exit` 를 `^[[:space:]]*exit([[:space:]]|$)` 로 바꾸면 현재 두 블록 PASS, i~vii 전부 FAIL, viii 만 PASS. viii 는 절 전체 비교로 잡힌다: `sed -n '/^### Pre-Spawn Sync Check/,/^### Pre-Edit Sync Check/p'` 로 뽑은 BASE 절(52줄)과 현재 절 `diff` exit 0, viii 픽스처와의 `diff` exit 1(`prespawn-section-mut-viii.diff`). AC-GDP-002 는 이미 두 절 파일을 만들고(`acceptance.md:188-189`) 표 행만 비교하므로, 절 `diff` 한 줄을 더하는 것으로 충분하다.

**분류 서술**: `acceptance.md:150` 은 RED-now 셀이 없어 이 카드의 작업을 재지 않는다고 적고(§2 의 vacuous 방향과 일치), 채택 근거를 §1.1 관측된 실패로 둔다. 이 부분은 정확하다. 흐린 곳은 두 가지다. (1) §2 는 한 셀짜리 기준을 "unadopted" 로 보고, §2.1 의 regression-guard 분류는 release-blocking 자격을 잃는다고 적는데, AC 표(`acceptance.md:30`)는 MUST-PASS 를 유지한다. (2) §2 Mutant probe 를 형태 뮤턴트 4개로만 수행했다. 보존 판정을 MUST-PASS 로 두는 것은 AC-GDP-003 과 같은 선례가 있으므로, 근거를 §4(트리에 고정한 불변 단언)로 명시하면 된다(D4, optional).

## 점검 3 — AC-GDP-016 범위와 재흡수 (AC-GDP-030·014·015 포함)

**Claim**: 예정된 통합 창 흡수 뒤 `$BASE..HEAD`·`git diff $BASE` 범위는 develop 커밋과 변경을 이 카드의 것으로 읽는다. SPEC 이 적은 대응(BASE 를 흡수 병합으로 옮김)은 오히려 카드 커밋을 범위에서 빼 버린다. 전제("BASE 이후 재흡수 없음")는 계획 시점에 이미 거짓이다.

**Evidence**:
- 판정 재실행: `git log --no-merges --format=%H 255f88eb0…..HEAD -- <acp 두 사본>` → `ac016-judge-at-bca80d7c5.txt` `test -e` 0, `test -s` 1(빈 파일). 양성 대조 `b412f8a33…..255f88eb0…` → 2줄 `97ef8e3023e9…`, `6896eef3766a…`(`ac016-control.txt`). SPEC 기록과 같다.
- develop 이 이미 앞서 있다: `git rev-list --count --left-right develop...HEAD` → `13 26`, `origin/develop...HEAD` → `6 26`, `origin/develop...develop` → `0 7`.
- 재흡수가 AC-GDP-016 범위에 더할 비병합 커밋: `git log --no-merges 255f88eb0…..develop` → 7개(`ac016-emulated-reabsorb-range.txt`: ae0d07d4d, e526c8d5b, 92494400f, a9b15fb62, 29c16a236, 52a4de177, 5b7927b15). 지금 이 7개는 acp 를 건드리지 않는다(`git log --no-merges HEAD..develop -- <acp 두 사본>` 빈 출력). 앞으로 develop 커밋이 acp 를 건드리면 AC-GDP-016 은 거짓 FAIL 이다.
- 재흡수가 AC-GDP-015 범위에 더할 템플릿 변경: `git diff --stat HEAD...develop -- <범위 루트>` → 7개 파일, 그중 템플릿 5개(`trace-ledger.sh` 신규 87줄 등). 추가 줄 146줄이 AC-GDP-015 검사 대상이 된다(`develop-pending-template.diff`). 지금 그 줄에는 SPEC-ID·날짜·a-f 16진 낱말이 없어 판정이 바로 뒤집히지는 않지만, `acceptance.md:559` 의 "이 변경분은 이 카드의 템플릿 편집만 담는다(develop 의 템플릿 변경은 $BASE 에 이미 들어 있다)"는 재흡수 뒤 거짓이 된다.
- AC-GDP-030(`acceptance.md:1007`·`1011`·`1012`·`1018`)과 AC-GDP-014(`acceptance.md:506`)도 같은 리터럴 `$BASE` 범위를 쓴다. 지금 develop 쪽 변경은 이 경로들에 없다(`git diff --name-only 255f88eb0… develop -- <.codex/agents, .agents/skills, 명령 원본>` 빈 출력) — 잠재 결함이다.

**SPEC 의 전제·대응이 부족한 이유**:
1. 이 저장소의 통합 절차는 창에서 로컬 develop 을 카드 워크트리에 흡수한 뒤 **병합 트리에서 재측정**하게 한다(`CLAUDE.local.md` §4.1). 재흡수는 선택이 아니라 예정된 단계이고, 로컬 develop 은 이미 13커밋 앞서 있다. `acceptance.md:1055` 의 "BASE 이후 develop 재흡수가 없다는 전제"는 계획 시점에 거짓이다.
2. `spec.md:335`·`plan.md:47` 은 "run-phase **전에** 재흡수하면 BASE 를 옮긴다"만 다룬다. run·sync 뒤 창에서의 흡수는 다루지 않는다.
3. 그 대응을 창 흡수에 적용하면 결과가 틀린다. BASE 를 흡수 병합 커밋으로 옮기면 카드의 run·sync 커밋은 모두 새 BASE 의 조상이 되어 `$BASE..HEAD` 에서 빠진다. AC-GDP-016 은 빈 목록으로 공허하게 PASS(`test -e` 0·`test -s` 1 — 돌아서 아무것도 없는 경우와 구별 불가), AC-GDP-030 은 `test -s ac030-src-commits.txt` 가 exit 1 로 거짓 FAIL, AC-GDP-015 는 `test -s ac015-template.diff` 거짓 FAIL, AC-GDP-014 는 "한 줄" 기대 거짓 FAIL.
4. `acceptance.md:607` 의 대응(목록이 비어 있지 않으면 커밋 메시지의 `t622` 로 가린다)은 구조가 아니라 메시지 문자열에 기댄다. 리드 배치 커밋처럼 `t622` 를 언급하는 develop 커밋은 잘못 분류된다.
5. 로컬 규칙 `gitflow-lane-protocol.md` §8 [HARD] 는 "이 카드가 무엇을 바꿨는가"를 리터럴 base SHA 가 아니라 흡수한 ref 와의 merge-base 부터 재도록 한다. 이 SPEC 의 카드 변경 범위 네 곳은 그 규칙과 반대다.

구조적 대안은 이미 이 트리에서 동작한다: `git log --no-merges --format='%h %s' HEAD --not develop` → 카드 커밋 22개만(흡수된 develop 커밋 0), 같은 형태에 acp 경로를 붙이면 빈 출력. `git log --first-parent --no-merges 255f88eb0…..HEAD` → 카드 커밋 2개(`bca80d7c5`, `b24f2e184`)만. `git diff --stat f1f034bb4… 255f88eb0… -- <범위 루트>` → 빈 출력이라, merge-base 에서 반출한 기준 사본은 지금 BASE 반출본과 같다(기존 측정값 유지 가능).

## 점검 4 — 미러 테스트 비회귀 (M4·M6·AC-GDP-013 (e))

**Claim**: 집합 비교가 맞게 설계됐다. 새 FAIL 과 사라진 PASS 를 모두 잡고, 빈 선택으로 공허하게 초록이 되지 않는다.

**Evidence**:
- 재실행: `go test ./internal/template/ -count=1 -run 'TestSanitizedPairParity|TestRuleTemplateMirrorDrift' -v` → exit 1(`mirror-audit-rerun.txt`, `.exit`). `acceptance.md:454-469` 명령을 그대로 적용: PASS 17, new-fail `test -e` 0·`test -s` 1, lost-pass `test -e` 0·`test -s` 1.
- 정렬한 집합 `diff` exit 0(`post-sets.sorted` 대 `base-sets.sorted`).
- 빈 선택·컴파일 실패 대리 픽스처: 빈 PASS 파일로 lost-pass 를 계산하면 기준선 PASS 17줄이 전부 나온다(`lost-pass-compilefail.txt` 17줄) → FAIL. `mirror-post-pass-count ≥ 17` 검사도 같은 경우를 막는다.
- PASS→FAIL, PASS 누락 뮤턴트는 `reanchor/mutmirror/result.txt` 에 기록돼 있고 이 감사의 로직 재확인과 일치한다(`grep -v -x -F -f` 양방향).

**부수 발견**: 병렬 하위 테스트라 출력 순서가 매번 다르다. 정렬하지 않은 `diff` 는 같은 집합인데도 exit 1 이었다. `plan.md:58` §C 8단계는 "기준선 집합 파일과 같은지 본다, 다르면 멈춘다"고만 적어 정렬 여부가 빠졌다. 거짓 PASS 가 아니라 거짓 정지 쪽 위험이다(D5, optional).

## 점검 5 — 기준선과 로컬 줄 인용 표본 확인

**Evidence** (awk 로 줄 출력, 전부 SPEC §A.1 표와 일치):
- `manager-git.md` 로컬/템플릿: L34/T32 `merge_method = …`, L116/T114 `gh pr merge <PR> --squash --delete-branch`, L150/T148 `Auto-merge: only with the --auto-merge flag`, L158/T156 `Pre-flight status reads (git fetch, …)`, L168/T166 `Execute only with --auto-merge flag AND all approvals obtained:`.
- `delivery.md` 로컬/템플릿: L355/T330 `#### Step 3.4`, L362/T337 `is_worktree_context == true AND --no-merge …`, L368/T343·L380/T355 `gh pr merge --squash --delete-branch`, L373/T348 `--no-merge: Skip auto-merge …`, L381/T356 `Checkout target branch, fetch latest`, L421/T396 `#### Context-Aware Next Steps`, L429/T404 `Auto-Merge PR (/moai sync --merge)`.
- `workflows/sync.md` L95/T85 사용법 줄, L114/T104 `**Flags**:` 줄.
- `agent-common-protocol.md` 13·17·52·290·292·296·297·301·302·307·311·341·360·375 행 — 표와 일치.
- 사본 덩어리 머리 재측정(`pair-mg.diff`, `pair-delivery.diff`, `pair-qgc.diff`, `pair-sync.diff`): `5,7c5` / delivery 9개(`9,11c9,11` … `510,511c479`) / qgc 7개(`9,11c9,11` … `165c158`) / sync 3개(`29,31c29,31`, `65,74d64`, `81c71`) — `spec.md` §A.4 와 모두 같다.

판정: 표본 인용 29곳과 덩어리 머리 4쌍 전부 일치. 결함 없음.

## 점검 6 — 일관성

- 이 카드가 `agent-common-protocol.md` 를 편집한다고 적은 문장, 또는 그 편집이 마지막 지침 커밋이라는 문장: 남은 것은 `acceptance.md:591` 의 "0.2.4 의 판정은 … 전제였다"는 이력 설명뿐이다. `grep` 으로 acp 언급을 모두 확인했고, 나머지는 비편집·보존·이력 문맥이다.
- 판정 대상 AC: `acceptance.md:53` "판정 대상: 16개", AC 표 활성 행 16. 새 REQ 없음(001~026 그대로, 활성 12).
- 범위: 편집 파일이 19→17(경우 B 21→19)로 줄어든 것은 리드 판단 (a)의 직접 결과이며 `spec.md` §C.4·`plan.md:7` 에 명시돼 있다. 넓어진 곳은 없다.
- 작은 부정확: AC-GDP-016 GIVEN(`acceptance.md:585`)은 "run-phase·sync-phase 커밋들"이라 적지만, `$BASE..HEAD` 에는 plan 커밋 `b24f2e184`·`bca80d7c5` 도 들어 있다. 판정에는 영향이 없다(D6, optional).

## Defects Found

D1. RANGE-ABSORB — `acceptance.md:598`(AC-GDP-016), `:1007`·`:1011`·`:1012`·`:1018`(AC-GDP-030), `:506`(AC-GDP-014), `:557`·`:559`(AC-GDP-015), `:607`, `:1055`; `spec.md:335`; `plan.md:47` — "이 카드가 바꾼 것"을 리터럴 `$BASE` 범위로 재는데, 통합 창의 예정된 develop 흡수(로컬 develop 이 이미 13커밋 앞섬) 뒤에는 develop 커밋·변경이 섞이고, SPEC 의 대응(BASE 를 흡수 병합으로 옮김)은 카드 커밋을 범위에서 빼서 AC-GDP-016 을 공허한 PASS, AC-GDP-014·015·030 을 거짓 FAIL 로 만든다. 전제 "BASE 이후 재흡수 없음"은 계획 시점에 거짓이고 `gitflow-lane-protocol.md` §8 [HARD] 와도 어긋난다 — Severity: major — Class: blocking — Required fix: (a) 카드 변경 범위 네 기준의 `$BASE..HEAD` 를 `git log --no-merges --format=%H HEAD --not develop -- <경로>`(또는 `$(git merge-base develop HEAD)..HEAD`, 병합 전에만 유효하다고 명시)로, `git diff $BASE -- <경로>` 를 `git diff develop HEAD -- <경로>`(흡수 직후 기준) 또는 merge-base 형태로 바꾼다. AC-GDP-016 양성 대조는 같은 `<tip> --not <ref>` 형태로 적는다 — `git log --no-merges --format=%H 255f88eb0… --not b412f8a33… -- <acp 두 사본>` 는 지금의 `b412f8a33..BASE` 대조와 같은 집합(2커밋)이므로 측정값을 그대로 쓸 수 있다. `--first-parent` 형태를 고르면 이 대조는 흡수로 들어온 커밋을 보지 못하므로 대조 구간을 따로 정해야 한다. (b) `acceptance.md:1055` 의 전제를 지우고, `spec.md:335`·`plan.md:47` 에 "통합 창 흡수 뒤 재측정은 흡수한 develop 과의 merge-base 기준으로 한다 — BASE 를 흡수 병합으로 옮기지 않는다"를 적는다. (c) `acceptance.md:559` 의 괄호 주장을 새 범위 정의에 맞게 고친다. (d) 기준 사본 반출(`plan.md:48`)도 같은 merge-base 에서 하도록 적는다 — 지금은 `git diff --stat f1f034bb4… 255f88eb0… -- <범위 루트>` 가 비어 있어 값이 같다.

D2. AC002-MUTANT — `acceptance.md:161`(순서 판정기), `:136-196` — 판정기가 REQ-GDP-002 를 어기는 뮤턴트 v(`git fetch … &`)·vi(`git fetch … | tee`)를 PASS 로 통과시키고, vii(`exit_note=…` 가 `^exit` 에 걸림)·viii(Lane B 병렬 허용 삭제)도 통과시킨다. `verification-completeness.md §2` Mutant probe 불합격 — Severity: major — Class: blocking — Required fix: (1) fetch 줄 판별을 `^[[:space:]]*[g]it fetch origin main( 2>&1)?[[:space:]]*$` 로 끝까지 고정하고 fetch 개수는 `^[[:space:]]*[g]it fetch` 로 따로 센다. (2) `exit` 판별을 `^[[:space:]]*exit([[:space:]]|$)` 로 바꾼다. (3) 이미 만드는 `$E/ac002-base-section.md` 와 `$E/ac002-local-section.md`(템플릿도)를 `diff` 해 exit 0 을 기대하는 줄을 더한다(이 카드는 파일을 고치지 않으므로 절 전체 보존이 곧 REQ-GDP-002 보존이다). (4) 뮤턴트 v~viii 를 §D.1 뮤턴트 목록에 더하고 새 판정기로 FAIL 을 기록한다. 이 감사에서 (1)+(2) 적용 시 현재 블록 PASS·i~vii FAIL, (3) 적용 시 viii FAIL 을 관측했다(`reanchor-audit/judge-hardened-proposal.txt`, `prespawn-section-mut-viii.diff`). D1 을 merge-base 기준으로 고치면 (3)의 기준 사본도 같은 기준에서 반출한다.

D3. REQ002-SUBSTANCE — `spec.md:155`, `spec.md:34`(HISTORY "요구사항의 실질 … 그대로"), `spec.md:73` — 새 REQ-GDP-002 는 0.2.4 보다 좁다(fetch 의 exit status 관측을 새로 요구). 그래서 SPEC 이 "정상 — 순서 보장"이라 부르는 `fetch; rev-list` 한 줄 형태는 새 REQ-GDP-002 를 충족하지 못한다. 또 AC-GDP-002 는 실패 시 `exit` 까지 요구하지만 REQ-GDP-002 문구에는 없다(리드가 적은 착지 성질에는 "abort on failure" 가 있다) — Severity: minor — Class: blocking — Required fix: `spec.md:155` 의 "요구하는 성질 … 은 같다"를 "착지한 성질로 좁혔다 — 순서 보장에 더해 fetch 의 exit status 관측을 요구하며, 0.2.4 의 `;` 한 줄 형태는 이제 충족하지 않는다(Pre-Edit 절은 REQ-GDP-002 대상이 아니다)"로 고치고 HISTORY 0.2.5 행의 "(3)" 설명을 같이 맞춘다. REQ-GDP-002 에 "fetch 가 실패하면 rev-list 에 이르지 않아야 한다"를 더해 AC-GDP-002 의 검사·exit 요구와 맞추거나, 반대로 AC 에서 그 요구를 빼고 이유를 적는다(요구사항 수는 그대로다).

D4. AC002-CLASS — `acceptance.md:30`, `:150` — §2 를 근거로 "RED-now 없음, 회귀 방지 판정"이라 적으면서 MUST-PASS 등급을 유지한다. §2 는 한 셀짜리 기준을 unadopted 로, §2.1 의 regression-guard 는 release-blocking 자격 없음으로 적는다 — Severity: minor — Class: optional — Required fix: 근거를 `verification-completeness.md §4`(트리에 고정한 보존 단언, AC-GDP-003 과 같은 성격)로 명시하고 MUST-PASS 유지 이유를 한 문장 적거나, 등급을 낮춘다.

D5. MIRROR-ORDER — `plan.md:58` — 사전 점검 8단계가 "기준선 집합 파일과 같은지"만 적고 정렬을 지정하지 않는다. 병렬 하위 테스트 출력 순서가 달라 정렬 없는 `diff` 는 같은 집합에서도 exit 1 이었다(정렬 뒤 exit 0) — Severity: minor — Class: optional — Required fix: "양쪽을 `sort` 한 뒤 `diff`" 라고 적거나, AC-GDP-013 (e) 의 양방향 `grep -v -x -F -f` 판정을 그대로 쓰라고 가리킨다.

D6. AC016-GIVEN — `acceptance.md:585` — GIVEN 이 "run-phase·sync-phase 커밋들"이라 적지만 범위에는 plan 커밋 두 개도 들어 있다 — Severity: minor — Class: optional — Required fix: "이 카드의 커밋들(흡수 병합 제외)"로 고친다. D1 수리 때 함께 처리하면 된다.

## Recommendation

1. D1: 카드 변경 범위 판정 네 곳(AC-GDP-014·015·016·030)을 흡수한 develop 과의 구조적 범위(`HEAD --not develop` 또는 merge-base)로 바꾸고, "재흡수 없음" 전제와 "BASE 를 흡수 병합으로 옮김" 대응을 지운다. 통합 창에서의 재측정 방법을 spec §E.2·plan §C 에 한 문단으로 적는다.
2. D2: AC-GDP-002 판정기의 fetch 줄과 exit 판별을 고정하고, Pre-Spawn 절 전체 `diff` 를 더한다. 뮤턴트 v~viii 를 기록한다.
3. D3: REQ-GDP-002 의 "실질 동일" 서술을 "좁혔다"로 고치고, abort 요구를 REQ 와 AC 중 한쪽으로 맞춘다.
4. D4~D6 은 선택 사항이다. 비용이 작으니 같은 수정에서 처리할 것을 권한다.

재감사는 D1~D3 과 그로 인한 퇴행만 보면 된다. 점검 4·5·6 은 이번 측정으로 닫혔다.

## 증거 목록

`.moai/reports/t622/reanchor-audit/` — `block-local.md`, `block-template.md`, `mut-v-background.md`, `mut-vi-pipe.md`, `mut-vii-noabort.md`, `mut-viii-laneb-serial.md`, `judge-new-mutants.txt`, `judge-hardened-proposal.txt`, `prespawn-section-base.md`, `prespawn-section-local.md`, `prespawn-section-mut-viii.md`·`.diff`, `ac016-judge-at-bca80d7c5.txt`, `ac016-control.txt`, `ac016-emulated-reabsorb-range.txt`, `develop-pending-template.diff`, `develop-pending-hex.txt`, `mirror-audit-rerun.txt`·`.exit`, `post-*.txt`, `new-fail.txt`, `lost-pass.txt`, `lost-pass-compilefail.txt`, `post-sets.sorted`, `base-sets.sorted`, `pair-mg.diff`, `pair-delivery.diff`, `pair-qgc.diff`, `pair-sync.diff`.

## Gaps

- 재흡수를 실제로 수행해 AC-GDP-014·015·016·030 을 돌려 보지는 않았다(읽기 전용 제약). 범위 오염은 `255f88eb0..develop`·`HEAD...develop` 로 흉내 내 쟀다.
- BASE 를 흡수 병합으로 옮겼을 때의 결과(D1 의 3번)는 git 범위 의미로 추론한 것이며 실행하지 않았다.
- AC-GDP-002 의 D2 보강안은 이 감사의 픽스처 10개에서만 확인했다.

## Residual-risk

- develop 이 Pre-Spawn 절을 REQ 에 맞는 다른 모양으로 다시 쓰면, 절 전체 `diff`(D2 의 3)는 흡수 뒤 빨강이 된다. D1 의 merge-base 기준 반출을 쓰면 "카드가 흡수한 것 대비 바꾸지 않았다"로 뜻이 맞춰진다.
- 판정기는 여전히 줄 단위다. `if` 블록 경계와 비교 방향은 읽기로 확인해야 한다(`acceptance.md:214` 의 한계 서술 유지).
