# t1226 — AC 일괄 검증 (SPEC-ALWAYS-LOADED-HEADROOM-001, run 단계 M7)

실행 트리: 워크트리 `.claude/worktrees/t1226`, 코드 기준 `build_head = 05d79c8b2`(이후 커밋은 `.moai/reports/t1226/` 만 변경 — `git diff --quiet 05d79c8b2 HEAD -- internal cmd pkg go.mod go.sum` 종료 코드 0). `$SCRATCH` 는 세션 임시 디렉터리다.

재현 절차(이 순서로 실행했다):

1. `rm -rf $SCRATCH/init-surface $SCRATCH/live`
2. `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_DISTRIBUTE_ALL && go test ./internal/cli/ -run '^TestHeadroomInitSurfaceExport$' -count=1 -v -args -headroom-export=$SCRATCH/init-surface` — 하네스 재실행(`BH` 와 코드 동일)
3. `mkdir -p $SCRATCH/live` → `git archive 05d79c8b27b1aacfc56946672e5187b0fcabf96e CLAUDE.md AGENTS.md .moai/config/sections .claude/rules/moai | tar -x -C $SCRATCH/live`
4. git 이 없는 검사: `SCRATCH=<dir> bash .moai/reports/t1226/ac-run.sh` — AC 명령 원문을 그대로 담았다(AC-ALH-001~007, 009 의 비-git 부분, 010). 출력 원문 `.moai/reports/t1226/ac-run.out`.
5. git 이 들어가는 검사: 워크트리 가드가 복합 명령 속 git 을 거부하므로 스크립트에 넣지 않고 한 줄 명령으로 따로 실행했다(아래 AC-ALH-008·009).

## 결과 요약

| AC | 판정 | 핵심 출력 |
|---|---|---|
| AC-ALH-001 | PASS | `1` `1` `2` `1` / `0` `0` |
| AC-ALH-002 | PASS | `UTF-8`, `1 1 1`, `PIN-live-OK`, `PIN-init-OK`, 재측정 `199111 total`·`203413 total` = 판정서, `_미측정` `0 0` |
| AC-ALH-003 | PASS | `KEYS-{init,live}-OK`, `SPLIT-{init,live}-OK`, ROWS/OVERLAP/DESTIN/M1P/NETNEG/EVID/DEST 전부 `BAD=0`·`MISSING=0` |
| AC-ALH-004 | PASS | 재실행 해시 `93de7321…4477`·`d97b33d9…c6c3` = 판정서, 해시 줄 `2`, `HASH-BAD=0` |
| AC-ALH-005 | PASS | sha256 `d0e61541…78547`, 선택 줄 `1 1 1`, 재실행 출력 존재, `/tmp` 인용 `0 0` |
| AC-ALH-006 | PASS | `CHECK-init=OK` `CHECK-live=OK` `CHECK-init_17=OK` `PAIRS-init=OK` `PAIRS-live=OK` |
| AC-ALH-007 | PASS | `TOKEN-live=OK` `TOKEN-init_17=OK` `TOKEN-init=OK`, 상신 항목 `4`, `RECOMMEND:` `1`, 결정 서술 `0` |
| AC-ALH-008 | PASS | `0` |
| AC-ALH-009 | PASS | PASS `1`, SKIP `0`, `harness_head` = `build_head`, sha256 = `harness_sha256`, `prepareSafeInitHome` `2`, `LOG-OK`, `moai … init` `0` |
| AC-ALH-010 | PASS (형태 2) | 형태 1 줄 전부 `0`, `runtime_observed = no` `1`, `count_set_init = 18` `1`, 17집합 줄 `6`, `verdict_init_17` `1` |

채무 검사(plan-audit 3회차 DEBT-1~3)는 AC-ALH-003·004 안에 들어 있다 — DEBT-1: ROWS 의 `chars ≤ gross`·M1 `chars == gross` 조건과 HASH 루프의 `pre_chars == gross`·`chars == pre − post` (`ROWS-* BAD=0`, `HASH-BAD=0`); DEBT-2: `OVERLAP-* BAD=0`; DEBT-3: `DESTIN-* BAD=0`. M1p 행은 없다(`M1P-* BAD=0`).

선택 채무 N5(포인터 원문 대조, 검토 항목): 커밋된 포인터 원문 6개의 `wc -m` 이 `pointers-<s>.tsv` 의 `pointer_chars` 와 같다 — 두 목록 모두 `23 64 68 87 89 146`(아래 원문). `net-negative` 행은 없다.

## ac-run.sh 출력 원문

```
### rerun harness tail
--- PASS: TestHeadroomInitSurfaceExport (0.65s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.472s
### AC-ALH-001
1
1
2
1
0
0
### AC-ALH-002
UTF-8
1
1
1
PIN-live-OK
PIN-init-OK
  199111 total
  203413 total
total_init = 203413
total_live = 199111
0
0
missing_init = -
### AC-ALH-003
KEYS-init-OK
KEYS-live-OK
SPLIT-init-OK
SPLIT-live-OK
ROWS-init BAD=0
OVERLAP-init BAD=0
DESTIN-init BAD=0
M1P-init BAD=0
NETNEG-init BAD=0
EVID-init MISSING=0
ROWS-live BAD=0
OVERLAP-live BAD=0
DESTIN-live BAD=0
M1P-live BAD=0
NETNEG-live BAD=0
EVID-live MISSING=0
DEST-init BAD=0
DEST-live BAD=0
### AC-ALH-004
93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477
d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3
2
HASH-BAD=0
### AC-ALH-005
d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547
1
1
1
.moai/reports/t1226/sec-172ef22eb.txt
0
0
### AC-ALH-006
CHECK-init=OK
CHECK-live=OK
CHECK-init_17=OK
PAIRS-init=OK
PAIRS-live=OK
# review: pointer_chars == wc -m of committed pointer text
64 .moai/reports/t1226/evidence/pointers/init/00b6deb8a2.txt
87 .moai/reports/t1226/evidence/pointers/init/02f8357cbb.txt
89 .moai/reports/t1226/evidence/pointers/init/3cf1ccd211.txt
146 .moai/reports/t1226/evidence/pointers/init/63d79ed95a.txt
23 .moai/reports/t1226/evidence/pointers/init/7609d7512a.txt
68 .moai/reports/t1226/evidence/pointers/init/c1549ca6b4.txt
23 64 68 87 89 146 
64 .moai/reports/t1226/evidence/pointers/live/00b6deb8a2.txt
87 .moai/reports/t1226/evidence/pointers/live/02f8357cbb.txt
89 .moai/reports/t1226/evidence/pointers/live/3cf1ccd211.txt
146 .moai/reports/t1226/evidence/pointers/live/63d79ed95a.txt
23 .moai/reports/t1226/evidence/pointers/live/7609d7512a.txt
68 .moai/reports/t1226/evidence/pointers/live/c1549ca6b4.txt
23 64 68 87 89 146 
### AC-ALH-007
TOKEN-live=OK
TOKEN-init_17=OK
TOKEN-init=OK
4
1
0
### AC-ALH-009 (non-git part)
1
0
05d79c8b27b1aacfc56946672e5187b0fcabf96e
build_head = 05d79c8b27b1aacfc56946672e5187b0fcabf96e
1
harness_sha256 = 8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5
LOG-OK
0
### AC-ALH-010
0
0
0
0
0
1
1
6
1
```

## git 이 들어가는 검사 (한 줄 명령, 출력 원문)

AC-ALH-008 — `git log --first-parent --no-merges --format=%H 7fe658815..HEAD -- <acceptance.md 의 36경로 그대로> | wc -l`:

```
       0
```

AC-ALH-009 — `git ls-files 'internal/cli/*headroom*_test.go'` / `shasum -a 256 internal/cli/init_headroom_export_test.go` / `grep -c 'prepareSafeInitHome' internal/cli/init_headroom_export_test.go`:

```
internal/cli/init_headroom_export_test.go
8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5  internal/cli/init_headroom_export_test.go
2
```

기준 트리 무변경 — `git diff --quiet 7fe658815eb0d4110b9acadad56e5a85bee3ed3f 05d79c8b27b1aacfc56946672e5187b0fcabf96e -- CLAUDE.md AGENTS.md .moai/config/sections/user.yaml .moai/config/sections/language.yaml .claude/rules/moai/core .claude/rules/moai/workflow internal/template/templates` → 출력 없음, 종료 코드 0. `git merge-base --is-ancestor 7fe658815… HEAD` → 종료 코드 0.

## 품질 게이트

`go build -o $SCRATCH/moai-run ./cmd/moai` → 종료 0. `$SCRATCH/moai-run spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001; echo "exit=$?"`:

```
INFO      OwnershipTransitionUnmeasured  …/SPEC-ALWAYS-LOADED-HEADROOM-001/spec.md  1     SPEC SPEC-ALWAYS-LOADED-HEADROOM-001 transition "(none)" → "draft" expected owner "manager-spec" but commit 10281a8570f4d9b870cfa1d69e1a9cfac125dd36 (docs(spec): t1226 plan SPEC-ALWAYS-LOADED-HEADROOM-001 — A_adm measurement and 150k verdict) has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 0 warning(s)
exit=0
```

(상태 전이 커밋 전 실행. 전이 뒤 재실행 결과는 `progress.md §E.2`.)

`golangci-lint version` → `golangci-lint has version v2.1.6 …`(CI 판). `golangci-lint run ./internal/cli/...; echo exit=$?` → `0 issues.` / `exit=0`. `go vet ./internal/cli/` → 종료 0.

## sync-audit 후속 수정 뒤 재검증 (F1·F2·F3·F5)

수정 커밋: F1 `7e509e4f1`(하네스 내보내기 경로 가드), F2·F3 `e1252de31`(판정서 상신 절차), F5 `6c84224b9`(progress.md SHA 보충, spec.md HISTORY 완료 행). 재검증은 `6c84224b9` 트리에서 했다. 워크트리 가드가 `bash <스크립트>` 호출을 거부해서, `ac-run.sh` 의 AC-ALH-007·009 줄을 같은 명령 그대로 한 줄씩 실행했다. AC-ALH-007 의 `rule()` 함수는 같은 판정식을 awk 한 줄로 펼쳐서 실행했다.

AC-ALH-007 — PASS:

```
TOKEN-live=OK
TOKEN-init_17=OK
TOKEN-init=OK
rule18=STRUCTURALLY-INFEASIBLE-UNDER-FREEZE verdict_init=STRUCTURALLY-INFEASIBLE-UNDER-FREEZE n=1 verdict_init_17=STRUCTURALLY-INFEASIBLE-UNDER-FREEZE
4
1
0
```

(상신 항목 `### (a)~(d)` 4, `^RECOMMEND: ` 1, 결정 서술 0. F2 는 하나뿐인 RECOMMEND 줄에 선택지 (4)로 넣고, 나머지는 불릿으로 적었다.)

AC-ALH-009 — **하네스 해시 줄 1건 불일치. F1 이 하네스 파일을 고쳤기 때문이다.** 나머지 줄은 전부 기대값과 같다.

```
1                                   # --- PASS
0                                   # --- SKIP
05d79c8b27b1aacfc56946672e5187b0fcabf96e      # harness_head
build_head = 05d79c8b27b1aacfc56946672e5187b0fcabf96e
1
harness_sha256 = 8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5
LOG-OK
0
internal/cli/init_headroom_export_test.go      # git ls-files
fdc74ea71b8963de49de943d2cad979b68a3a4356f1f3dedef5aab6c569dc573  internal/cli/init_headroom_export_test.go
2                                   # prepareSafeInitHome
```

- 현재 하네스의 sha256 `fdc74ea7…c573` 은 판정서 `harness_sha256`(`8776d7b5…e9c5`)과 다르다. AC 명령을 문자 그대로 읽으면 이 줄은 FAIL 이다.
- S_init 을 만든 하네스는 `build_head` 판이다. `git show 05d79c8b2:internal/cli/init_headroom_export_test.go | shasum -a 256` 은 `8776d7b5809b47a6215a79497f6036414ceb2967d32067db612d07289b49e9c5` 이고 판정서 값과 같다. 따라서 측정의 출처 결속은 유지된다.
- 판정서의 `build_head`·`harness_sha256`·측정값은 고치지 않았다. 새 하네스로 S_init 을 다시 재어 이 줄들을 갱신하는 일은 리드가 정한다(아래 F1 증거 참조. 새 하네스는 init 코드를 바꾸지 않았다).

F1 증거(`.moai/reports/t1226/evidence/`): `f1-red.txt`(헬퍼 부재 컴파일 RED), `f1-red2.txt`(무조건 `os.RemoveAll` 헬퍼에서 `--- FAIL: TestHeadroomExportDirGuard/refuses_non-empty_unmarked_dir`), `f1-green.txt`(4개 하위 테스트 PASS), `f1-harness.txt`(플래그 실행 `-count=2` 두 번 모두 PASS. 두 번째 실행은 표식이 붙은 첫 번째 내보내기를 지웠다), `f1-lint.txt`(golangci-lint v2.1.6 `0 issues.`).

부수 확인: 판정서 편집 뒤 AC-ALH-001 의 폐기 수치 줄 `0`, `^F = ` `0`. `moai spec lint --strict SPEC-ALWAYS-LOADED-HEADROOM-001` 은 `✓ No findings`, 종료 코드 0(F5 커밋 전 작업 트리에서 실행).

## F1 재측정과 AC 전체 재검증 (리드 결정: 재측정)

**새 `build_head` = `f8d6f51671535c613c5afb8b5ba0b984dcc3a585`** (F1 하네스 `7e509e4f1` 포함, 하네스 재실행 직전 `git rev-parse HEAD`). `$SCRATCH` 는 세션 임시 디렉터리 아래 `rm/` 이다. 재측정 절차와 전후 값 표는 `verdict.md` § 측정 환경 「F1 재측정」에 있다. 요지는 다음과 같다. 하네스 `--- PASS: TestHeadroomInitSurfaceExport (0.68s)`, `headroom-path` 18줄이 모두 `present` 이고, 이전 `harness-run.txt` 의 18줄과 `diff` 해도 출력이 없다. `git archive -o live.tar f8d6f5167 …` 뒤 `tar -x`. 모든 합계와 해시가 기록값과 같아서, 판정서에서는 `build_head`·`harness_sha256` 두 줄과 `harness-run.txt` 첫 줄 `harness_head` 만 바꿨다.

```
  203413 total                      # S_init (18)  == total_init
  198447 total                      # S_init (17)  == total_init_17
  199111 total                      # S_live       == total_live
93de7321ea747c584768af09d1908ad3e36fd076073a6f2a9d6fb5b7a22b4477  -   # == hash_init
d97b33d960c9801d4ec145ca263ed788425b337f43c585594c8d527c1318c6c3  -   # == hash_live
surface=init total=203413 A_adm=14201 R=477 U=79 T_min=189689
surface=init_17 total=198447 A_adm=13717 R=454 U=79 T_min=185184
surface=live total=199111 A_adm=13716 R=477 U=79 T_min=185872
```

(`build.py` 는 `--hash` 없이 돌렸다. 산출 트리는 `$SCRATCH` 에만 쓰였고, 실행 뒤 `git status --short` 는 무출력이었다.)

### 실행 방식

워크트리 가드가 `bash .moai/reports/t1226/ac-run.sh` 를 거부해서, AC 명령을 한 줄씩 따로 실행했다. 다음 두 곳은 가드가 셸 형태를 거부해 같은 판정식을 옮겨 적었다.
- AC-ALH-004 HASH 루프(셸 산술 `$((pre - post))` 거부) → `python3 -c`. 행 선택 awk 는 원문 그대로이고, 비교 조건 네 가지(ADMIT 은 `post_hash == H`, 비-ADMIT 은 `!= H`, `pre_chars == gross`, `chars == pre − post`)를 1:1 로 옮겼다.
- AC-ALH-006 `calc`(함수 안 awk 거부) → `python3 -c`. `C·A·U·R` 합산 조건과 `T_min = C − A + R` 을 1:1 로 옮기고, `PAIRS` 는 정렬 집합을 비교한다.
- AC-ALH-003 의 ROWS·OVERLAP·DEST awk 는 원문 그대로 한 파일씩 실행했다. M1P·NETNEG 루프는 대상 행 수를 직접 셌다. 두 표면 합쳐 `M1p` ADMIT 행 0, `net-negative` 행 0 이라서 루프가 검사할 대상이 없다(원 스크립트 출력 `BAD=0` 과 같은 의미).

### 결과 (재측정 트리, `build_head = f8d6f5167`)

| AC | 판정 | 핵심 출력 |
|---|---|---|
| AC-ALH-001 | PASS | `1` `1` `2` `1` / `0` `0` |
| AC-ALH-002 | PASS | `UTF-8`, `1 1 1`, `PIN-live-OK`, `PIN-init-OK`, 재측정 `199111 total`·`203413 total` = `total_live`·`total_init`, `_미측정` `0 0`, `missing_init = -` |
| AC-ALH-003 | PASS | `KEYS-init-OK` `KEYS-live-OK` `SPLIT-init-OK` `SPLIT-live-OK`, `ROWS-{init,live} BAD=0`, `OVERLAP-{init,live} BAD=0`, `DESTIN-{init,live} BAD=0`, M1p 0행·net-negative 0행, `EVID-{init,live} MISSING=0`, `DEST-{init,live} BAD=0` |
| AC-ALH-004 | PASS | 재실행 해시 `93de7321…4477`·`d97b33d9…c6c3` = 판정서, 해시 줄 `2`, `HASH-BAD-init=0` `HASH-BAD-live=0` (M2 대상 110·106행) |
| AC-ALH-005 | PASS | `d0e61541367abb06a170bd36b6376e51d51882380ca9f899f0d2934016a78547`, `1 1 1`, `sec-172ef22eb.txt` 존재, `/tmp` 인용 `0 0` |
| AC-ALH-006 | PASS | `CHECK-init=OK (C=203413 A=14201 R=477 U=79 T_min=189689)`, `CHECK-live=OK (C=199111 A=13716 R=477 U=79 T_min=185872)`, `CHECK-init_17=OK (C=198447 A=13717 R=454 U=79 T_min=185184)`, `PAIRS-init=OK` `PAIRS-live=OK`, 포인터 원문 `wc -m` = `23 64 68 87 89 146` 양쪽 |
| AC-ALH-007 | PASS | `TOKEN-live=OK` `TOKEN-init_17=OK` `TOKEN-init=OK`, 상신 항목 `4`, `RECOMMEND:` `1`, 결정 서술 `0` |
| AC-ALH-008 | PASS | 36경로 `git log --first-parent --no-merges --format=%H 7fe658815..HEAD -- …` → 무출력(0행). 양성 대조: 같은 형태를 `.moai/reports/t1226/verdict.md` 에 걸면 5행(`e1252de31 6a03d5272 7ae1a3b86 b231487ab 10281a857`)이 나온다 |
| AC-ALH-009 | PASS | PASS `1`, SKIP `0`, `harness_head` `f8d6f51671535c613c5afb8b5ba0b984dcc3a585` = `build_head`, `git ls-files` → `internal/cli/init_headroom_export_test.go`, sha256 `fdc74ea71b8963de49de943d2cad979b68a3a4356f1f3dedef5aab6c569dc573` = `harness_sha256`, `prepareSafeInitHome` `2`, `LOG-OK`, `moai … init` `0` |
| AC-ALH-010 | PASS (형태 2) | 형태 1 줄 `0 0 0 0 0`, `runtime_observed = no` `1`, `count_set_init = 18` `1`, 17집합 줄 `6`, `verdict_init_17` `1` |

바로 앞 절에 적은 AC-ALH-009 하네스 해시 불일치는 이 재측정으로 해소됐다.
