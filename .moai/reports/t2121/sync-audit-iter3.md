auditor-model: claude-opus-5-5

# t2121 sync 단계 재감사 판정 (iter3, 수리 델타 2차, Claude 단독)

verdict: PASS-WITH-DEBT
audited_sha: 54f6e04e135d0a632a56a156b155b1d9226e3c45

- 카드: t2121 (P0, Class B, SPEC 없음), 브랜치 `WT-complete-merge-readiness`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2121`
- 판정: **PASS-WITH-DEBT**, 종합 점수 **0.89**
- 감사 범위: 수리 델타 `git diff 0a1192eeacf2a412f371a62f7625361a37c7e762..HEAD -- internal/ .moai/reports/t2121/progress.md` (커밋 4ed05288c1 RED, de262310a7 GREEN, 54f6e04e13 기록). 감사 시작과 판정 작성 직전 모두 `git rev-parse HEAD` = `54f6e04e135d0a632a56a156b155b1d9226e3c45`, `git status --short` = 빈 출력. 프로브와 변이는 모두 `go test -overlay`로 돌렸고 추적 파일은 건드리지 않았다.
- 한 줄 근거: 지난 감사의 차단 결함 G1(감사 시점과 트리가 같아지는 흡수 merge가 흡수 검사를 건너뜀)은 닫혔다. 재현 프로브 N8·N4c와 새 G1 변형 6건(Q1a·Q2a·Q3c·Q4b·Q7·Q8a)이 모두 거부되고, 정상 흡수(N2a·N3a·P10·Q1b·Q2b·Q3b)는 그대로 통과한다. 남은 것은 선택 항목뿐이다: 목록 실패 분기를 잠그는 테스트 부재(H1), 증거 디렉터리만 바꾸는 merge의 새 거짓 거부(H2), 범위 밖으로 넘긴 G2.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | 프로브 27건 전부 기대대로: `RESULT N8_X_ours_absorb_drops_main_change passed=false want=false OK`, `RESULT N4c_delete_main_file_in_absorb passed=false want=false OK`, `RESULT N2a_clean_absorb_of_older_main_commit passed=true want=true OK`, `RESULT N3a_two_clean_absorbs passed=true want=true OK`, `RESULT P10_clean_absorb_shifted_lines passed=true want=true OK`. 감점: H2(증거 전용 merge 거짓 거부, 실패 시 닫힘 방향) |
| Security (25%) | 90/100 | PASS | 머지 게이트 우회 시도 18건(지난 감사 재현 5 + 새 13) 중 통과해서는 안 되는 것은 전부 거부. 새 코드는 셸 없이 내부 SHA만 인자로 쓰는 `git rev-list` 1회 추가(`grep` 결과 `exec.Command`·비밀값 패턴 0건, `grep_exit=1`). 감점: G2(태그 그림자) 미처리 — 범위 밖 후속 카드 |
| Craft (20%) | 86/100 | PASS | `go vet` exit 0, `gofmt -l` 빈 출력, 패키지 테스트 `ok … factorylane 66.584s`. 새 분기 변이 6개 중 5개 사망, M2(목록 실패 시 검사 생략) 생존(H1). G3 플래그 변이 2개(M7 `--binary` 제거, M8 `--no-textconv` 제거) 사망 → G3 닫힘. 커버리지 수치 미측정 |
| Consistency (15%) | 90/100 | PASS | 기존 `refuse`·`GitRunner`·실패 시 닫힘 관례 유지, RED(4ed05288c1: 테스트+RED 출력만) → GREEN(de262310a7: 소스 15줄만) → 기록(54f6e04e13) 커밋 분리, 진행 기록에 수리 2차 기록. 사소: `injectingMergeDriver`의 `file` 인자 미사용, 커밋 테스트 `N2a` 이름이 감사 프로브 N2a(오래된 main 커밋)와 다른 형태 |

must-pass 방화벽: Functionality·Security 모두 기준 통과. 가중 평균 90.0, 차원 조화평균 0.894 → 보고 점수 0.89. 차단 결함 없음, 선택 부채가 남아 PASS-WITH-DEBT.

## 지난 결함 처리 상태

| 결함 | 상태 | 근거 |
|---|---|---|
| G1 트리 동일 흡수 merge가 검사 우회 [blocking] | **닫힘** | HEAD에서 N8·N4c 거부(사유: `merge <sha> was resolved by hand: re-merging its parents conflicts`, `merge <sha> differs from a plain re-merge of its parents`). 커밋 테스트 `TestT2121AbsorbRepair2/N8…`, `/N4c…`가 RED 출력(`sync-audit passed=true, want false`)에서 GREEN으로 바뀜. 옛 동작으로 되돌린 변이 M4가 같은 두 테스트에서 사망 |
| G2 `main` 태그가 브랜치를 가림 [optional] | 미처리(지시상 범위 밖, 후속 카드) | 진행 기록 Gaps에 명시 |
| G3 `--binary`·`--no-textconv` 미잠금 [optional] | **닫힘** | M7 → `--- FAIL: TestT2121AbsorbRepair2BinaryPinned`, M8 → `--- FAIL: TestT2121AbsorbRepair2TextconvPinned` |
| G4 옥토퍼스 메시지 고정 [info] | 변화 없음 | 기록만 |

## 결함 목록

수리 유래 차단 결함 없음 → 반복 정지 규칙 해당 없음.

### H1 [Low] [optional] [in-diff] 목록 실패 분기(`mErr != nil`)를 잠그는 테스트가 없음

- 신뢰도: 높음(변이로 확인).
- 변이 M2(`mErr != nil`이면 흡수 검사를 건너뜀)가 커밋된 T2121 테스트 전체를 통과했다: `rc=0`, `ok … factorylane 38.279s`.
- 감사관 프로브 Q8a(`--min-parents=2` 호출만 실패시키는 러너 + `-s ours` 흡수)로 차이를 보였다: HEAD에서 `RESULT Q8a_minparents_fails_with_dropping_merge passed=false want=false OK`, M2에서 `RESULT Q8a_minparents_fails_with_dropping_merge passed=true want=false UNEXPECTED`. 즉 코드는 맞지만 잠겨 있지 않다.
- 선택 처리 이유: 다른 git 호출은 성공하고 이 호출만 실패하는 상황이 필요하다.
- 권장 수정: 기존 `noMergeTreeGit`처럼 `--min-parents=2` 인자에만 오류를 내는 러너를 두고, `-s ours` 흡수(트리 = 감사 트리) 저장소에서 거부를 기대하는 테스트 1건 추가.
- 덧붙임(Info): 주석 "A failure to list merges fails closed"는 실제로는 "전체 흡수 검사로 넘긴다"는 뜻이다. Q8b(실패 + 증거 커밋만)는 `passed=true`로 통과하는데, `absorbOnlyChange`가 첫 부모 경로를 직접 다시 훑으므로 안전한 대체다. 주석 표현만 약간 과장.

### H2 [Low] [optional] [in-diff, 수리로 바뀐 동작] 증거 디렉터리만 바꾸는 merge가 이제 거부됨 — 거부 문구와 어긋남

- 신뢰도: 높음(재현함).
- Q6a(감사 커밋에서 갈라진 브랜치에 증거 파일만 커밋 → 카드가 `--no-ff` merge): `passed=false`, 사유 `merge <sha> brings in <sha>, which is not part of main`. Q6b(`commit-tree`로 만든, 이미 조상인 main과의 merge에 증거 파일만 추가): `passed=false`, 사유 `differs from a plain re-merge of its parents`. iter2 코드(변이 M4)에서는 둘 다 `passed=true`였다.
- 실패 시 닫힘 방향이라 착지 안전에는 영향이 없다. 다만 거부 문구는 여전히 "the audited commit must … differ from HEAD only under .moai/reports/<card>/"를 허용 형태로 안내하므로, 트리가 증거 디렉터리에서만 다른 Q6a를 거부당한 레인은 문구와 동작이 어긋난다고 읽게 된다.
- 권장 수정(택1): 문구에 "merge 커밋 없이"를 넣거나, 증거 전용 merge를 받아들이도록 할 때는 merge 트리와 재merge 트리의 차이가 증거 디렉터리 안에만 있는지 비교. 실제 레인 흐름(증거를 카드 브랜치에 직접 커밋)에서는 드문 형태라 선택 항목으로 둔다.

### H3 [Info] [optional] [in-diff] 트리가 같을 때 거부 문구의 `first path outside …: ""`

- G1 형태 거부에서 `(first path outside .moai/reports/t9/: "")`처럼 빈 따옴표가 찍힌다. 뒤에 붙는 사유가 원인을 밝히므로 기능 영향 없음. 기록만.

### 범위 밖 (판정 영향 없음)

- G2 태그 그림자: 지시대로 후속 카드. 이번 프로브에서 재측정하지 않았다.

## 프로브 결과 (HEAD 54f6e04e13, overlay, 트리 무변경)

`go test -overlay <scratch>/probe.json ./internal/factorylane/ -run TestAuditProbeIter3 -count=1 -v` → exit 0, `ok … factorylane 31.347s`. "iter2 동작" 열은 같은 프로브를 G1 수리 이전 분기(변이 M4 = 0a1192eeac의 블록)로 돌린 결과다.

### (a) 지난 감사 프로브 재현

| 프로브 | 내용 | HEAD | 기대 | iter2 동작 | 맞음? |
|---|---|---|---|---|---|
| N8 | `-X ours` 흡수, main 변경 버림 | 거부 | 거부 | 통과 | 예 (G1 닫힘) |
| N4c | 흡수 안에서 main 새 파일 삭제 | 거부 | 거부 | 통과 | 예 (G1 닫힘) |
| N2a | 오래된 main 커밋을 깨끗이 흡수 | 통과 | 통과 | 통과 | 예 |
| N3a | 깨끗한 흡수 두 번 | 통과 | 통과 | 통과 | 예 |
| P10 | main이 줄 번호를 밀어낸 깨끗한 흡수 | 통과 | 통과 | 통과 | 예 |
| P1 | merge 안에서 바이너리 내용 교체 | 거부 | 거부 | 거부 | 예 |
| P7 | `-diff` 파일 내용 교체 | 거부 | 거부 | 거부 | 예 |
| P2b | 카드 hunk 이동 | 거부 | 거부 | 거부 | 예 |

### (b) 새 공격 프로브

| 프로브 | 내용 | HEAD | 기대 | iter2 동작 | 맞음? / 이유 |
|---|---|---|---|---|---|
| Q1a | 증거 커밋 → `-X ours` 흡수 → 증거 커밋 (merge가 증거 커밋보다 뒤) | 거부 | 거부 | 통과 | 예. `rev-list` 범위가 위치와 무관하게 merge를 찾음 |
| Q1b | 같은 배치에 깨끗한 흡수 | 통과 | 통과 | 통과 | 예 |
| Q2a | 곁가지에서 `-s ours`로 main 흡수 → 카드가 곁가지를 merge (첫 부모 경로 밖 merge) | 거부 | 거부 | 통과 | 예. 첫 부모 merge의 두 번째 부모(곁가지)가 main에 없음 |
| Q2b | main 내부에 merge가 있는 상태에서 깨끗한 흡수 (목록에 main 내부 merge 포함) | 통과 | 통과 | 통과 | 예. 거짓 거부 없음 — 첫 부모 경로만 판정 |
| Q2c | 부모 순서 반전(main 위에 카드 merge 후 fast-forward) | 거부 | 거부(거짓 거부 허용) | 거부 | 예. 실패 시 닫힘 |
| Q3a | 깨끗한 흡수 → 두 번째에서 `-X ours` | 거부 | 거부 | 거부 | 예 (트리가 달라 iter2도 잡음) |
| Q3b | 두 번 다 깨끗, main이 추가 후 삭제해 트리 = 감사 트리 | 통과 | 통과 | 통과 | 예. 트리 동일 + merge 존재여도 정상 흡수는 통과 |
| Q3c | 깨끗한 흡수 → 두 번째는 `-s ours` + 파일 삭제로 트리 = 감사 트리 | 거부 | 거부 | 통과 | 예. 두 번째 merge만 재merge 불일치 |
| Q4a | main 아닌 기능 브랜치 흡수(새 파일) | 거부 | 거부 | 거부 | 예 |
| Q4b | 기능 브랜치를 `-s ours`로 흡수, 카드 diff·트리 무변화 | 거부 | 거부 | 통과 | 예. 두 번째 부모가 main에 없음 |
| Q5a | merge 없음, 증거 하위 디렉터리 커밋 | 통과 | 통과 | 통과 | 예 |
| Q5b | 판정 파일이 하위 디렉터리에만 있음 | 거부(`verdict.md does not exist`) | 거부 | 거부 | 예. 읽기 위치는 증거 루트뿐 |
| Q5c | 접두가 같은 형제 디렉터리 `.moai/reports/t9x/` | 거부 | 거부 | 거부 | 예. 접두에 `/` 포함 |
| Q6a | 증거 전용 브랜치 merge | 거부 | 거부(거짓 거부) | 통과 | 예, 단 H2 |
| Q6b | 증거 파일만 더한 조작 merge | 거부 | 거부(거짓 거부) | 통과 | 예, 단 H2 |
| Q7 | `-s ours` main 흡수 (충돌 없이 트리 = 감사 트리) | 거부 | 거부 | 통과 | 예. G1의 가장 순수한 형태 |
| Q8a | `--min-parents=2` 실패 주입 + `-s ours` 흡수 | 거부 | 거부 | 통과 | 예. 실패가 전체 검사로 넘어감 |
| Q8b | 같은 실패 주입 + 증거 커밋만 | 통과 | 통과 | 통과 | 예. 대체 검사가 정상 이력을 통과시킴 |
| Q9 | 감사 커밋 = HEAD | 통과 | 통과 | 통과 | 예 |

## 변이 시험 (커밋된 T2121 테스트 기준)

`go test -overlay <scratch>/<mutant>.json ./internal/factorylane/ -run T2121 -count=1`

```
M1_drop_minparents_clause   rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2/N8_X_ours_absorb_dropping_main_change_refuses, /N4c_absorb_deleting_a_file_main_added_refuses
M2_failure_skips_check      rc=0 SURVIVED ok  github.com/modu-ai/moai-adk/internal/factorylane 38.279s   (H1)
M3_drop_path_check          rc=1 KILLED   --- FAIL: TestT2121AbsorbTolerance/late_code_change_without_absorb_refuses
M4_old_behavior_paths_only  rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2/N8…, /N4c…
M5_minparents3              rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2/N8…, /N4c…
M6_reversed_range           rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2/N8…, /N4c…
M7_drop_binary_flag         rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2BinaryPinned
M8_drop_no_textconv_flag    rc=1 KILLED   --- FAIL: TestT2121AbsorbRepair2TextconvPinned
```

## 실행한 명령과 관측 결과

### Claim
수리 커밋 de262310a7은 G1을 닫았고, 정상 흡수 경로를 깨지 않았으며, 정적 검사와 지정 테스트를 모두 통과한다. 남은 결함은 모두 선택 항목이다.

### Evidence
- `timeout 580 go test ./internal/factorylane/ -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/factorylane	66.584s`
- `timeout 580 go test ./internal/cli/ -run MergeReady -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	23.108s`
- `timeout 580 go test ./internal/cli/ -run SpecLess -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	15.761s`
- `timeout 580 go test ./internal/cli/ -run FactoryMerge -count=1` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	14.981s`
- `go vet ./internal/factorylane/` → `vet_exit=0`; `gofmt -l internal/factorylane` → 빈 출력, `gofmt_exit=0`
- 프로브(HEAD): 위 표, `ok … factorylane 31.347s` (첫 실행 `39.117s`도 같은 판정)
- 프로브(iter2 동작 = 변이 M4): `RESULT N8_… passed=true want=false UNEXPECTED`, `RESULT N4c_… passed=true want=false UNEXPECTED`, `Q1a`·`Q2a`·`Q3c`·`Q4b`·`Q6a`·`Q6b`·`Q7`·`Q8a` 모두 `passed=true`; 나머지는 HEAD와 같음
- 프로브(M2): `RESULT Q8a_minparents_fails_with_dropping_merge passed=true want=false UNEXPECTED`
- 변이: 위 블록
- 커밋 구성: `git show --stat` → 4ed05288c1 = `red-repair2-output.txt` + `merge_specless_repair2_t2121_test.go`, de262310a7 = `merge_specless.go`만(12+/3-), 54f6e04e13 = `progress.md`만. RED 출력 파일에 `--- FAIL: …/N8…`, `--- FAIL: …/N4c…`, `FAIL … factorylane 19.712s`
- 보안 grep: `git diff 0a1192eeac..HEAD -- internal/ | grep -nE '^\+.*(exec\.Command|"sh"|os\.Getenv|Sprintf\("%s"|password|token|secret)'` → 매치 없음(`grep_exit=1`)
- Binding run conditions: `grep -c "Binding run conditions" .moai/reports/t2121/progress.md` → `0` (해당 항목 없음)

### Baseline-attribution
모든 명령은 이번 실행에서 트리 HEAD `54f6e04e135d0a632a56a156b155b1d9226e3c45`(작업 트리 깨끗)에 대해 돌렸다. Go 테스트는 이 트리 소스를 `go test`(go1.27.3, git 2.54.0)가 직접 빌드했고 설치된 `moai` 바이너리는 쓰지 않았다. "iter2 동작" 기준선은 `git show 0a1192eeac:internal/factorylane/merge_specless.go`로 읽은 옛 블록을 HEAD 소스에 되돌려 넣은 overlay(M4)로 측정했다.

### Gaps
- **Codex 미호출**: 이번 반복에서는 `codex_audit`·`audit_multi`를 부르지 않았다. 카드의 Codex 예산(2회)이 이미 소진됐다는 리더 지시 때문이다. 따라서 이 판정은 Claude 단독이며, 감사 계획(`moai verify audit-plan`)도 이번에 다시 읽지 않았다. 지난 감사 기준으로 codex는 `required` 게이트였으므로, 이번 판정에 Codex 영수증은 없다(`receipts=none`).
- **영수증 가드 거부**: 이 판정을 보고하자 SubagentStop 훅이 `AUDIT_RECEIPT_VIOLATION`(영수증 미인용)으로 종료를 거부하고 codex 감사 호출을 요구했다. 리더가 Codex 예산 소진을 이유로 호출을 금지했으므로 감사관은 호출하지 않았다. 세 번째 Codex 호출을 승인할지, 리더가 직접 영수증을 받을지는 리더가 정한다. 그 전까지 이 PASS-WITH-DEBT는 Codex 영수증으로 뒷받침되지 않은 Claude 단독 판정이다.
- **작업 트리 격리 가드 거부 1건**: `git show 0a1192eeac:… > <scratch>/iter2_merge_specless.go` 리다이렉트 명령이 "too complex" 사유로 거부됐다. 대신 리다이렉트 없는 `git show`로 옛 소스를 읽고, 옛 블록을 `python3 -I` 스크립트(`mkmut.py`) 안에 문자열로 넣어 M4 변이를 만들었다. 측정 대상은 바뀌지 않았다.
- CI 미확인: 저장소 전체·`./internal/cli/` 전체 테스트는 돌리지 않았다(CI 몫, PENDING). CI 러너의 git 버전도 확인하지 않았다.
- 커버리지 수치 미측정.
- G2는 지시대로 다시 측정하지 않았다.
- Windows에서 `BinaryPinned`·`TextconvPinned`는 건너뛰므로(POSIX 셸 필요) 그 환경에서는 G3 잠금이 없다 — 이 감사는 darwin에서만 돌렸다.

### Residual-risk
- main 변경과 카드 변경이 같은 파일에서 충돌 없이 자동 merge되며 의미가 상호작용하는 경우는 텍스트 수준 검사로 잡을 수 없다(지난 감사와 같은 잔여 위험).
- G2 태그 그림자는 후속 카드로 넘어갔으므로 이 카드 착지 시점에는 열려 있다.
- 증거 파일을 커밋이 아니라 작업 트리에서 읽는 기존 동작(이전 F6)은 그대로다.
- 프로브·변이 원본 출력은 머신 로컬 스크래치 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/1df9d274-f8c9-4593-8790-efd5997fbc70/scratchpad/t2121-iter3/`에만 있다. 판정을 가른 줄은 위에 옮겨 적었고, 나머지는 사라질 수 있다.

## 반복 이력
- iter1: FAIL 0.70 @7723481e7e — F1(바이너리 내용), F2(hunk 이동) 차단.
- iter2: FAIL 0.77 @0a1192eeac — F1·F2 닫힘, 새 차단 G1(기존 결함, 수리 유래 아님).
- iter3(이번): PASS-WITH-DEBT 0.89 @54f6e04e13 — G1·G3 닫힘, 차단 결함 없음. 선택 부채 H1(테스트 공백), H2(증거 전용 merge 거짓 거부·문구 불일치), G2(후속 카드). 수리 유래 차단 결함 없음 → 반복 정지 규칙 해당 없음.

## 부채 목록 (PASS-WITH-DEBT)
1. G2 — `develop+"^{commit}"`를 `refs/heads/<develop>`로 해석 (후속 카드, 리더 지시상 범위 밖).
2. H1 — `--min-parents=2` 실패 주입 테스트 1건.
3. H2 — 거부 문구에 merge 조건 명시, 또는 증거 전용 merge 허용.
