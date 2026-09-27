# sync-audit — 카드 t1260 · 역할 로드 판정 테스트 판별력 부여 (독립 감사)

- 대상: 워크트리 `.claude/worktrees/t1260`, 브랜치 `WT-role-load-test-teeth`, HEAD `277954928`(기준 `develop b4f798dcc`, 커밋 3개: `515b05e06`·`4e4a87501`·`277954928`)
- 분류: Class B(SPEC 없음). 근거 문서: primary `.moai/reports/t1171/sync-audit.md`(F1~F10, 변이 m1·m2), 레인 판정서 `.moai/reports/t1260/verdict.md`
- 감사 중 작성자 없음. 작업 트리는 처음부터 끝까지 깨끗했다(`git status --short` 0행). 변이와 탐침은 전부 `go test -overlay` 로 주입했고, 추출본·오버레이·원시 출력은 머신 로컬 스크래치패드 `…/scratchpad/audit-t1260/` 에만 있다.
- 리드 지정 렌즈: ① 변이 판별력(m1·m2·m5·m6 직접 재실행) ② F7 의미 보존(판정 결과 전후 동일 + 핀 52행 재계산) ③ rosterguard. 추가로 F5 테스트 강도와 F2 범위를 판정했다.

---

## 1. 판정

**전체 판정: PASS-WITH-DEBT** — 차단 결함 0건. 부채는 레인이 이월한 F4·F6·F8~F10(원 감사 소관)과 이번에 새로 적은 선택 항목 N1~N6이다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%, must-pass) | 92/100 | PASS | 필수 변이 4개 전부 실패·실제 코드 통과(§2.1), 추가 변이 3개(m7·m8·m9)도 전부 실패. AC-RLP-003·004·006 판정식 원문 재실행 `true`. F7 치환 전후 판정 입력 필드 차이 0건(§2.2). 감점: 입력 타입을 거치지 않는 결합(전역 상태)은 여전히 못 잡음(레인이 잔여 위험으로 고지), `NOT_RUN` 분류는 생산 호출자 없음(N3) |
| Security (25%, must-pass) | 95/100 | PASS | 치환 역변환으로 52개 파일 전부 원본과 바이트 동일 — 두 치환 외 변경 없음. 잔여 `goos`·`/Users/`·`/var/folders/kt/` 0건, 키 형식 0건, 이메일 형식 0건. 감점: 계정에 묶인 암호화 블롭은 남음(치환 불가, 레인 고지) |
| Craft (20%) | 80/100 | 부분(커버리지 축 미측정) | vet·gofmt·golangci-lint 무결. 파일 커버리지 83.8%(원 감사 82.5% → 개선, 목표 85% 미달·F6 이월). 패키지 수준 커버리지는 이 감사에서도 미측정 — 임계값 판정 불가라 FAIL 이 아니라 미검증 |
| Consistency (15%) | 88/100 | PASS | 파일·명명·오류 처리 관례 준수, 커밋 3개 모두 카드 id 포함. 감점: 파일 머리 주석이 여전히 「TOML-specification parse」라 적어 함수 주석(좁혀짐)과 어긋남(N2), rosterguard 면제 사유의 「byte-frozen」 문안이 치환 뒤 사실과 어긋남(N6), `@MX:ANCHOR` 과장은 F4 로 이월 |

가중 합 89.8, 조화 평균 88.4. must-pass 두 차원 모두 통과.

## 2. 증거

환경 정리: 모든 `go test` 는 한 복합 호출 안에서 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test …` 형태로 실행했다(AC 판정식 재실행은 acceptance.md 원문의 5개 변수 형태 그대로).

### 2.1 렌즈 ① — 변이 판별력

오버레이 원본은 HEAD 의 `internal/cli/codex_role_fingerprint.go` 를 읽어 문자열 치환으로 만들었다(치환마다 적중 수 정확히 1을 단언). 레인 판정서의 오버레이는 쓰지 않았다.

| 변이 | 바꾼 것 | 출처 |
|---|---|---|
| m1 | `codexRoleLoadInput` 에 `Extra codexRoleBehaviour`, 판정을 `in.Matched[in.Role] \|\| in.Extra.NonceReturned` | 원 감사 §2.6 그대로 |
| m2 | 판정 앞에 `if in.Role == "manager-lead" \|\| in.Role == "mission-governor" { return true }` | 원 감사 §2.6 그대로 |
| m5 | `codexRoleLoadOutcome` 의 `DeveloperItems == 0 → NOT_RUN` 분기 삭제 | 레인 판정서 §2.3 정의 |
| m6 | 스캔의 `developerItems++` 삭제 | 레인 판정서 §2.3 정의 |
| m7(추가) | `developerItems++` 를 role 필터 **앞**으로 이동(모든 response_item 을 셈) | 이 감사 |
| m8(추가) | 입력 타입에 중립 이름 `Flag bool` 추가, 판정을 `in.Matched[in.Role] && !in.Flag`(AND 결합, 행동 타입 아님) | 이 감사 |
| m9(추가) | F2 수리 호출(`i = codexRoleSkipMultilineValue(…)`) 삭제 | 이 감사 |

명령(공통): `go test -json -count=1 -timeout=600s [-overlay <mN>.json] -run '^TestCodexRole' ./internal/cli/`

실제 코드(오버레이 없음):

```
real exit=0
  44 pass
   1 skip
skip TestCodexRoleLiveLoadAndReadOnly
```

(skip 은 `MOAI_CODEX_ROLE_LIVE` 게이트로 꺼진 LIVE 테스트 — 원래 상태 그대로.)

변이별 결과(원문, 게이트 LIVE 테스트의 고정 로그 줄은 생략):

```
m1 exit=1
fail TestCodexRoleLoadPredicateStructurallyIndependent/input_type_has_no_behaviour_field
fail TestCodexRoleLoadPredicateStructurallyIndependent/exhaustive_field_mutations
fail TestCodexRoleLoadPredicateStructurallyIndependent/half_true_arm
fail TestCodexRoleLoadPredicateStructurallyIndependent
codex_role_behaviour_test.go:201: codexRoleLoadInput must carry exactly {Role string, Matched map[string]bool}; got [Role string Matched map[string]bool Extra cli.codexRoleBehaviour]
codex_role_behaviour_test.go:224: field 0 mutation found 1 behaviour site(s) in codexRoleLoadInput
codex_role_behaviour_test.go:259: half-true arm changed the load verdict on 6 of 24 cases

m2 exit=1
fail TestCodexRoleLoadPredicateStructurallyIndependent
fail TestCodexRoleLoadPredicateContractNeutral/role_name_branch_mutant_rejected
fail TestCodexRoleLoadPredicateContractNeutral
fail TestCodexRoleBodyFingerprintDerive/N1_version_mismatch
fail TestCodexRoleBodyFingerprintDerive/N2_absent_roles
fail TestCodexRoleBodyFingerprintDerive/N5_real_parent_sessions
fail TestCodexRoleBodyFingerprintDerive
codex_role_behaviour_test.go:172: expected 12 load-true and 12 load-false cases, got 14 true of 24
codex_role_contract_test.go:211: codexRoleLoadPredicate is not uniform across roles: [manager-lead/empty=true manager-lead/others=true mission-governor/empty=true mission-governor/others=true]

m5 exit=1
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured/zero_developer_items
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured/empty_record
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured
codex_role_outcome_test.go:58: outcome = "FAIL", want "NOT_RUN"
codex_role_outcome_test.go:64: outcome = "FAIL", want "NOT_RUN"

m6 exit=1
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured/measured_body_absent
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured/measured_body_present
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured
codex_role_outcome_test.go:71: N3 fixture must still carry developer items for this arm to be a measured FAIL
codex_role_outcome_test.go:89: …/real/rollout-2026-09-24T18-39-44-01a0d2c9-0503-7700-becc-1bc8c98a8198.jsonl (role=builder-harness): outcome = "NOT_RUN", want "PASS"

m7 exit=1
fail TestCodexRoleLoadOutcomeSeparatesUnmeasured/zero_developer_items
codex_role_outcome_test.go:55: expected an unmeasured observation, got {Matched:map[] DeveloperItems:1}

m8 exit=1
fail TestCodexRoleLoadPredicateStructurallyIndependent/input_type_has_no_behaviour_field
codex_role_behaviour_test.go:201: codexRoleLoadInput must carry exactly {Role string, Matched map[string]bool}; got [Role string Matched map[string]bool Flag bool]

m9 exit=1
fail TestCodexRoleBodyExtractSkipsOtherMultilineValues/basic_desc_bracket_and_key
fail TestCodexRoleBodyExtractSkipsOtherMultilineValues/desc_line_starts_with_bracket
fail TestCodexRoleBodyExtractSkipsOtherMultilineValues/desc_line_mimics_key
codex_role_extract_test.go:28: extract: codex role body: developer_instructions key absent
codex_role_extract_test.go:28: extract: codex role body: developer_instructions value is not a TOML multi-line literal string
```

판독:

- **F1 해소 확인.** m1 은 원 감사에서 다섯 AC 테스트를 모두 통과했으나 이제 AC-006 세 부분 테스트에서 실패한다. `half_true_arm` 은 행동값을 실제로 판정까지 흘려 24건 중 6건에서 판정이 갈렸다 — 수리 전의 「만들고 넘기지 않는」 결함이 사라졌다.
- **F3 해소 확인.** m2 가 이제 AC-004 자신의 `role_name_branch_mutant_rejected` 에서 실패한다. 균일성 검사가 생산 판정을 직접 호출하고, 강제 참·강제 거짓 두 방향을 양성 대조로 둔다(`codex_role_contract_test.go:214-232`). m2 의 AC-006 실패는 모집단 전제(`:172`)에서 난 부수적 실패다.
- **F5 판별력.** m5·m6·m7 세 방향이 모두 잡힌다. 특히 m7 을 잡는 것은 `no-developer.jsonl` 픽스처가 assistant 항목 하나를 일부러 담고 있기 때문이다 — 「developer 가 아닌 항목까지 세는」 결함이 이 한 줄로 드러난다.
- **m8** 은 행동 타입이 아닌 필드로 AND 결합한 변이다. 허용목록 단언 하나가 잡고, 행동값 주입 부분 테스트는 반응하지 않는다(N1 참조 — 결함은 아니다).

레인 판정서 §2.2·§2.3 표의 실패 지점과 메시지는 이 재실행과 일치한다.

### 2.2 렌즈 ② — F7 의미 보존

추출: `git archive b4f798dcc internal/cli/testdata/codex-rollouts-t1171` 과 `git archive HEAD …` 를 스크래치패드 `old/`·`new/` 로 풀었다. `new/` 와 작업 트리 픽스처는 `diff -rq` 무출력(동일). 파일 수 양쪽 모두 `.jsonl` 28개, 전체 52개.

**(a) 치환 수와 범위**

```
old /Users/goos:       78  old tmpdir:     1133
new /home/fixture:       78  new tmpdir:     1133
old pre-existing /home/fixture:        0 /private/var/folders/fixture:        0
new residual goos|/Users/|kt/:        0
```

합계 1,211 건(78 + 1,133)으로 레인 주장과 일치한다. 치환 대상 문자열이 원본에 미리 있지 않았으므로 역변환이 모호하지 않다.

**(b) 역변환 바이트 동일성 + JSON 유효성** — `new` 에 두 치환을 거꾸로 적용하면 `old` 와 같아야 한다(두 치환 외의 변경이 있으면 여기서 드러난다):

```
files reverse-identical 52 mismatch 0
jsonl lines 738 unparseable 0
email-like 0
home-like paths ['/home/fixture']
```

**(c) 판정 결과 전후 비교.** HEAD 코드로 만든 테스트 바이너리에 감사 전용 탐침 테스트(`zz_audit_probe_test.go`, 오버레이로만 추가)를 넣어, 같은 바이너리로 `old` 와 `new` 픽스처 트리를 각각 읽었다. 기록 28개마다 AC 테스트가 읽는 입력을 모두 덤프했다 — 표찰·표찰 유무, 세 기대표(`roles`·`roles-other-version`·빈 표) 각각의 일치 역할 집합·developer 항목 수·판정·결과(PASS/FAIL/NOT_RUN), 마지막 assistant 응답 sha256(AC-004 nonce 판정의 입력), developer 항목별 본문 해시.

```
records old 28 new 28
SEMANTIC_FIELD_DIFFS 0
DEV_ITEMS_TOTAL 97 HASH_CHANGED 39 OF_WHICH_ROLE_TABLE_TAGGED 0
selected_by_label 14 of 28
positive_matches 12
n1_false 12
n2_false 12
synthetic [('synthetic/n3-label-without-body.jsonl', 'builder-harness', None, False, 'FAIL'), ('synthetic/n4-crossed-roles.jsonl', 'builder-harness', ['manager-git'], False, 'FAIL')]
outcomes_roles ['FAIL', 'PASS'] 12
probe_old.jsonl parents_unmatched=14 parents=14 nonce_last=12
probe_new.jsonl parents_unmatched=14 parents=14 nonce_last=12
```

판독: 판정 입력 필드 차이 **0건**. developer 항목 97개 중 39개의 본문 해시가 바뀌었지만(스킬 목록 등 경로를 담은 비역할 항목), 그중 역할 기대표 어느 쪽에든 걸리는 것은 0개다. 그래서 선별 14/28, 양성 12, N1·N2 각 12건 거짓, 부모 세션 14건 불일치, N3·N4 FAIL 이 치환 전후로 같다. 레인이 「테스트는 경로 문자열을 읽지 않는다」고만 적은 부분을 이 비교가 관측으로 확정한다.

AC 판정식 원문 재실행(HEAD, 출력 경로만 스크래치패드):

```
true
AC-RLP-006 exit=0
true
AC-RLP-004 exit=0
true
AC-RLP-003 exit=0
```

**(d) sha256 핀 52행 재계산** — `progress.md` 표의 `| \`<사본>\` | \`<sha256>\` |` 행을 파싱해 현재 파일과 대조하고, 재고정 문단이 인용한 이전 커밋 `1a8dfa26e` 의 표를 치환 전 바이트(`old/`)와 대조했다:

```
pin rows 52 match 52 mismatch 0 fixture files 52 unpinned []
1a8dfa26e pin rows 52 match pre-substitution bytes 52 mismatch 0
rows changed vs 1a8dfa26e 28
```

현재 표 52행 전부 일치, 핀이 없는 픽스처 파일 0개, 바뀐 행은 치환된 28개와 정확히 같다. 재고정 문단이 가리키는 이전 핀 위치(`1a8dfa26e`)도 사실이다.

### 2.3 렌즈 ③ — rosterguard

```
$ go test -count=1 -timeout=1200s ./internal/harness/rosterguard/
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	17.902s
rosterguard exit=0
```

`internal/harness/rosterguard/registry.go:768-774` 는 이 픽스처를 경로로만 등록한다(`git grep -nE "[0-9a-f]{64}" -- internal/harness/rosterguard | grep -c t1171` → `0`, 해시 핀 없음). 경로가 바뀌지 않았으므로 면제는 그대로 적중하고, 실측으로도 통과한다. 다만 그 면제 사유 문안은 이번 치환으로 사실과 조금 어긋났다(N6).

### 2.4 게이트

```
$ go vet ./internal/cli/
vet exit=0
$ gofmt -l internal/cli/codex_role_*.go
(출력 없음)
$ golangci-lint run --timeout=10m ./internal/cli/
0 issues.
$ go test -count=1 -run '^TestCodexRole' -coverprofile=… ./internal/cli/
cover exit=0
codexRoleBodyExtractTOML 87.1% · codexRoleSkipMultilineValue 90.0% · codexBuildRoleExpectationTable 86.2%
codexRoleSessionLabel 70.4% · codexRoleFingerprintDerive 75.0% · codexRoleLoadOutcome 100.0%
codexRoleFingerprintObserve 84.8% · codexRoleLoadPredicate 100.0%
stmts=142 covered=119 pct=83.8
```

(커버리지 함수별 수치는 `go tool cover -func` 출력의 해당 파일 행을 한 줄로 모은 것이다. 파일 합계는 `cover.out` 의 문장 수 가중 합이다.) 비밀 형식 탐침(`sk-…`·`ghp_…`·`AKIA…`·`xox[bp]-…`): `key-shape hits: 0`.

### 2.5 F5 강도와 F2 범위 판정

**F5 — 강함.** 네 부분 테스트가 세 값을 서로 다른 입력으로 고정한다: developer 항목 0개인 표찰 기록 → `NOT_RUN`, 빈 파일 → `NOT_RUN`, 측정됐으나 본문 없음(N3, developer 항목 수 > 0 을 전제로 단언) → `FAIL`, 실물 12개 → `PASS`. 대조 부분 테스트가 불리언 판정만으로는 앞의 두 경우가 둘 다 `false` 임을 남긴다. 세 방향 변이(분기 삭제 m5, 계수 삭제 m6, 계수 위치 오류 m7)가 모두 잡힌다. 판정 입력 타입(`Role`, `Matched`)은 바뀌지 않았고 `codexRoleLoadOutcome` 이 그 위에서 관측 수를 함께 보는 구조라, AC-006 의 허용목록과 충돌하지 않는다. 테스트 파일이 `NOT_RUN` 문자열을 출력하지 않아 AC 판정식의 `NOT_RUN|ABORTED` 거부 조건도 건드리지 않는다(AC-004·006 재실행 `true` 로 확인). 남은 한계는 N3 — 이 분류를 쓰는 생산 경로가 아직 없다.

**F2 — 범위 적정.** 원 감사가 F2 를 「같은 코드 경로의 비차단 부채」로 적었고, 수리는 생산 코드 17행(`codex_role_fingerprint.go:98-120`)과 호출 1행(`:66`)으로 작다. 원 감사의 두 탐침 입력이 그대로 테스트 부분 이름(`desc_line_starts_with_bracket`, `desc_line_mimics_key`)으로 들어갔고, 기본 문자열 변형과 한 줄 삼중따옴표 음성 대조, 닫히지 않은 값의 오류 보고까지 있다. m9(수리 호출 삭제)가 세 부분 테스트에서 원 감사 탐침과 같은 메시지로 실패하므로 테스트가 수리를 실제로 지킨다. 기존 12개 역할의 추출값 불변은 AC-003(`P_all_twelve` 12/12 — codex 가 실제로 주입한 본문과의 일치)이 독립 오라클로 확인한다. 한계(N4)는 레인이 주석에 일부 고지했다.

## 3. 결함 목록 (구조화)

차단(blocking) 결함 없음. 아래는 모두 선택(optional) 항목이다.

- **N1** [info][optional, 확신 높음] `internal/cli/codex_role_behaviour_test.go:218-261` — 생산 입력 타입에는 행동값을 넣을 자리가 0개라, `exhaustive_field_mutations`·`half_true_arm` 이 실제 코드에서 비교하는 두 입력은 같은 값이다. `BEHAVIOUR_FIELDS_MUTATED 2 OF 2` 는 변이를 **만든** 수이지 판정에 **전달한** 수가 아니다. 두 부분 테스트의 판별력은 행동 타입 자리가 생겼을 때만 살아나는데(m1: `sites != 0` 단언, 6/24 분기), 그 경우는 허용목록 단언이 이미 막는다. 행동 타입이 아닌 필드(m8)는 허용목록만 잡는다. 결함이 아니다 — 허용목록이 정확 집합이라 실질 방어로 충분하다. 필요 수정: 없음. 원하면 부분 테스트 주석에 「주 방어는 허용목록, 이 두 갈래는 행동 타입 자리가 생겼을 때의 2차 방어」라고 적는다.
- **N2** [minor][optional, 확신 높음] `internal/cli/codex_role_fingerprint.go:9-11` — 파일 머리 주석이 여전히 「extracted from the role TOML by a TOML-specification parse」라 적는다. 함수 주석(`:48-50`)은 이번 카드에서 「줄 스캐너, 완전한 TOML 파서 아님」으로 좁혔으므로 같은 파일 안에서 두 서술이 어긋난다. 필요 수정: 머리 주석도 같은 범위로 좁힌다(F4 의 `@MX:ANCHOR` 정정과 한 번에 처리하면 된다).
- **N3** [info][optional, 확신 높음] `internal/cli/codex_role_fingerprint.go:309` — `codexRoleLoadOutcome` 의 호출자는 테스트뿐이다(판정 함수와 같은 상태, F4). 원 감사 F5 가 요구한 구분은 API 수준에서 성립했지만, 아직 어느 경로도 `NOT_RUN` 을 실제로 찍지 않는다. LIVE 배선을 후속으로 넘긴 SPEC 과 일관되므로 수정 요구 없음. plan.md §G 1행이 「찍는다」고 적는 규약의 이행 주체는 후속 LIVE 카드임을 기록해 둔다.
- **N4** [info][optional, 확신 중간] `internal/cli/codex_role_fingerprint.go:51-66,104-120` — 스캐너는 여러 줄 **배열** 값 안의 `[` 로 시작하는 줄을 여전히 테이블 머리로 읽고, 기본 문자열 안의 이스케이프된 `\"""` 를 인식하지 않는다(후자는 주석에 고지). 방출 형태에는 둘 다 없고, 실패 방향은 조용한 오판이 아니라 키 부재 오류다. 필요 수정: 없음(방출기가 형태를 바꿀 때 재검토).
- **N5** [info][optional, 확신 높음] `internal/cli/testdata/codex-rollouts-t1171/real/*.jsonl` — 치환으로 developer 항목 39개(비역할)의 본문 바이트가 codex 가 실제로 기록한 것과 달라졌다. 판정에는 영향 없음(§2.2 c)이고 레인이 progress.md 재고정 문단과 판정서 §4 에 고지했다. 필요 수정: 없음.

- **N6** [minor][optional, 확신 높음] `internal/harness/rosterguard/registry.go:734-737`(주석 `:765-767`), 대상 행 `:774` — rosterguard 면제 사유가 이 픽스처를 「byte-frozen copy … Editing it would invalidate the capture」로, 주석이 「kept verbatim」으로 적는데, F7 이 면제 대상인 `real/rollout-…-8d51-….jsonl` 의 바이트를 바꿨다(핀 행 변경 28개에 포함). 면제의 실제 근거 — 캡처 시점의 명단 수 인용 — 는 치환이 경로 문자열만 바꿨으므로(§2.2 b 역변환 동일) 여전히 참이고, rosterguard 테스트도 통과한다. 그래서 차단하지 않는다. 필요 수정: 사유 문안에 「경로만 중립화, 명단 인용 문장은 캡처 그대로」를 덧붙이거나 그대로 수용한다(리드 처분).

레인이 이월한 원 감사 항목(F4 `@MX:ANCHOR` 과장, F6 파일 커버리지 83.8% < 85%, F8·F9·F10 SPEC 서술)은 이 감사가 다시 확인한 바 그대로 남아 있다. 이월 사유(manager-spec 소관·카드 범위 밖)는 타당하다.

## 4. 기준 귀속

- 트리: 이 워크트리, HEAD `277954928`, 작업 트리 깨끗. 감사 시작 시점과 보고서 작성 직전에 `git rev-parse --short HEAD` → `277954928`, `git branch --show-current` → `WT-role-load-test-teeth` 를 다시 읽었다.
- 치환 전 기준: `git archive b4f798dcc` 로 푼 픽스처. 이전 핀 기준: `git show 1a8dfa26e:.moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/progress.md`.
- 모든 수치는 이번 감사 실행에서 관측한 값이다. 레인 판정서의 수치는 대조용으로만 읽었고, 변이 오버레이·탐침·비교 스크립트는 이 감사가 따로 만들었다.

## 5. 미검증 (Gaps)

- **수리 전 재현(레인 주장 C1)** 은 다시 돌리지 않았다. 원 감사 §2.6 의 m1·m2 원문 출력이 같은 결과를 이미 기록하고 있어 교차 근거는 있으나, 이 감사의 관측은 아니다.
- **`internal/cli` 패키지 전체 스위트와 패키지 수준 커버리지**: 실행하지 않았다(지시에 따라). 판정은 develop push 뒤 CI 몫이다.
- 커밋 `515b05e06` 단독 상태(치환 전 픽스처 + 새 테스트)는 따로 측정하지 않았다. 다만 §2.2(c) 가 같은 코드에서 치환 전후 판정 입력이 동일함을 보였으므로 결과가 다를 경로는 관측상 없다.
- AC-RLP-001·002·008·009 판정식은 재실행하지 않았다(001·002 는 렌즈 ① 의 `TestCodexRole` 실행 안에 테스트로는 포함되어 통과).
- darwin 외 플랫폼, 교차 모델(codex/GLM) 감사는 실행하지 않았다.

## 6. 잔여 위험

- AC-006 의 방어는 입력 타입의 필드 집합을 고정한다. 판정 함수가 패키지 변수·전역 상태로 행동값을 읽는 결합은 입력 타입을 거치지 않아 이 테스트가 보지 못한다(레인 고지와 같음). 판정 함수 본문이 한 줄(`return in.Matched[in.Role]`)이라 현재 위험은 코드 리뷰로 충분히 보인다.
- 균일성 검사(AC-004)는 역할 12개와 세 가지 일치 집합에 대해서만 본다. 역할 이름이 아니라 일치 집합의 크기 같은 다른 속성으로 분기하는 변이는 AC-003 의 N 갈래가 잡는 범위에 기댄다.
- 픽스처는 이제 codex 원본과 바이트가 다르다. 향후 누군가 `.moai/reports/t1143/` 원본과의 동일성을 다시 재면 26건 불일치가 나오는데, 이는 의도된 상태다(progress.md 재고정 문단에 기록됨).
