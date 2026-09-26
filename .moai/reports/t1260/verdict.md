# 카드 t1260 — 역할 로드 판정 테스트에 판별력 부여 (t1171 sync-audit 후속)

- 분류: Tier S, Class B(SPEC 없음, plan 생략), cycle_type=tdd
- 워크트리: `.claude/worktrees/t1260`, 브랜치 `WT-role-load-test-teeth`, 기준 `develop b4f798dcc`
- 근거 문서: primary 체크아웃 `.moai/reports/t1171/sync-audit.md`(읽기 전용) — F1~F10, 변이 m1·m2
- 커밋: `515b05e06`(테스트·코드: F1·F3·F5·F2), `4e4a87501`(픽스처 중립화·핀 갱신: F7), 이 판정서 커밋
- push 하지 않았다.

---

## 1. 주장

| # | 주장 | 근거 절 |
|---|---|---|
| C1 | 수리 전 트리에서 변이 m1 은 AC 테스트 다섯 개를 모두 통과했다(결함 재현). m2 는 AC-004·AC-006 을 통과하고 AC-003 에서만 걸렸다 | §2.1 |
| C2 | 수리 뒤 m1 은 AC-006 의 세 부분 테스트에서, m2 는 AC-003·AC-004·AC-006 에서 실패한다. 실제 코드에서는 전부 통과한다 | §2.2 |
| C3 | F5: developer 항목이 0개인 기록은 `NOT_RUN`, 측정됐으나 본문이 없으면 `FAIL`, 본문이 있으면 `PASS` 로 갈린다. 두 변이(m5·m6)가 이 테스트에서 실패한다 | §2.3 |
| C4 | F2: 다른 키의 여러 줄 문자열 안 줄(`[`로 시작, 키 흉내)이 더 이상 추출을 깨지 않는다 | §2.4 |
| C5 | F7: 픽스처 28개에서 사용자·머신 고유 절대경로가 0건이 됐고, 진행 기록의 sha256 핀 52행이 현재 트리와 전부 일치한다 | §2.5 |
| C6 | AC-RLP-001·002·003·004·006 판정식(acceptance.md 원문, 출력 경로만 스크래치패드)이 모두 `true` 다 | §2.6 |
| C7 | vet·gofmt·golangci-lint 무결, rosterguard 패키지 통과 | §2.7 |

## 2. 증거

환경 정리: 모든 `go test` 는 같은 프로세스 안에서 13개 변수(`MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID`)를 `unset` 한 뒤 실행했다(래퍼 스크립트 한 호출 안에서 `unset` → `go test`).

변이는 감사와 같은 방식으로 `go test -overlay` 로 주입했다 — 작업 트리의 생산 파일은 한 번도 바꾸지 않았으므로 되돌릴 것도 없다. 오버레이 원본은 그 시점의 `internal/cli/codex_role_fingerprint.go` 를 읽어 문자열 치환으로 만들었다.

- **m1**(감사 §2.6 그대로): `codexRoleLoadInput` 에 `Extra codexRoleBehaviour` 추가, 판정을 `return in.Matched[in.Role] || in.Extra.NonceReturned` 로
- **m2**(감사 §2.6 그대로): 판정 앞에 `if in.Role == "manager-lead" || in.Role == "mission-governor" { return true }` 삽입

실행 명령(요약 표기 `R`):

```
go test -json -count=1 -timeout=300s -run '^(TestCodexRoleBodyFingerprintParse|TestCodexRoleBodyExpectationTable|TestCodexRoleBodyFingerprintDerive|TestCodexRoleLoadPredicateContractNeutral|TestCodexRoleLoadPredicateStructurallyIndependent)$' [-overlay <mN>.json] ./internal/cli
```

### 2.1 원인 확립 — 수리 전 재현 (트리 `b4f798dcc`, 테스트 무변경)

m1 (`R -overlay m1.json`) — 결함: 전부 통과

```
== before-m1 exit=0
pass TestCodexRoleLoadPredicateStructurallyIndependent/input_type_has_no_behaviour_field
pass TestCodexRoleLoadPredicateStructurallyIndependent/exhaustive_field_mutations
pass TestCodexRoleLoadPredicateStructurallyIndependent/half_true_arm
pass TestCodexRoleLoadPredicateStructurallyIndependent/coupled_mutant_rejected
pass TestCodexRoleLoadPredicateStructurallyIndependent
pass TestCodexRoleLoadPredicateContractNeutral/manager-lead
pass TestCodexRoleLoadPredicateContractNeutral/mission-governor
pass TestCodexRoleLoadPredicateContractNeutral/all_twelve_same_function
pass TestCodexRoleLoadPredicateContractNeutral/role_name_branch_mutant_rejected
pass TestCodexRoleLoadPredicateContractNeutral
pass TestCodexRoleBodyFingerprintDerive/S_selection_by_label
pass TestCodexRoleBodyFingerprintDerive/P_all_twelve
pass TestCodexRoleBodyFingerprintDerive/N1_version_mismatch
pass TestCodexRoleBodyFingerprintDerive/N2_absent_roles
pass TestCodexRoleBodyFingerprintDerive/N3_label_without_body
pass TestCodexRoleBodyFingerprintDerive/N4_crossed_roles
pass TestCodexRoleBodyFingerprintDerive/N5_real_parent_sessions
pass TestCodexRoleBodyFingerprintDerive/hybrid_selection_mutant_rejected
pass TestCodexRoleBodyFingerprintDerive
pass TestCodexRoleBodyFingerprintParse/... (5 부분 테스트 전부 pass)
pass TestCodexRoleBodyFingerprintParse
pass TestCodexRoleBodyExpectationTable/... (4 부분 테스트 전부 pass)
pass TestCodexRoleBodyExpectationTable
BEHAVIOUR_FIELDS_MUTATED 2 OF 2
COUPLED_MUTANT_DIVERGED 24
```

m2 (`R -overlay m2.json`) — AC-004·AC-006 은 통과, AC-003 만 실패

```
== before-m2 exit=1
pass TestCodexRoleLoadPredicateStructurallyIndependent   (4 부분 테스트 전부 pass)
pass TestCodexRoleLoadPredicateContractNeutral/role_name_branch_mutant_rejected
pass TestCodexRoleLoadPredicateContractNeutral
fail TestCodexRoleBodyFingerprintDerive/N1_version_mismatch
fail TestCodexRoleBodyFingerprintDerive/N2_absent_roles
fail TestCodexRoleBodyFingerprintDerive/N5_real_parent_sessions
fail TestCodexRoleBodyFingerprintDerive
```

같은 트리의 실제 코드(`R`, 오버레이 없음)는 `== before-real exit=0`, 30개 pass.

**원인 판단 — 증상이 아니라 원인인가.** m1 이 통과한 이유를 테스트 본문에서 확인했다: (1) 반사 검사가 필드 **이름**만 보고(`behaviour`/`nonce` 부분 문자열) `Extra` 는 통과시킨다, (2) 「전수 변이」는 행동 값 `b` 를 만들고 어디에도 넘기지 않는다, (3) 「결합 변이 기각」은 테스트 파일 안의 함수를 대상으로 해 생산 판정을 보지 않는다, (4) 비교 모집단이 로드 참 12건뿐이라 `||` 형 결합은 참을 참으로 둔다. (4)는 감사가 짚지 않은 추가 원인이다 — 행동 값을 실제로 주입해도 참만 있는 모집단에서는 OR 결합이 보이지 않는다. m2 가 AC-004 를 통과한 이유는 `role_name_branch_mutant_rejected` 가 테스트 안 람다만 검사했기 때문이다(감사 F3 과 일치).

### 2.2 수리 뒤 — 변이 전후 표 (최종 트리, 오버레이는 최종 소스에서 재생성)

| 변이 | 수리 전 | 수리 뒤 | 수리 뒤 실패 지점 |
|---|---|---|---|
| m1 | exit=0, 5개 AC 테스트 전부 PASS | **exit=1** | AC-006 `input_type_has_no_behaviour_field`, `exhaustive_field_mutations`, `half_true_arm` |
| m2 | exit=1, AC-003 만 FAIL(N1·N2·N5) | **exit=1** | AC-003 N1·N2·N5, **AC-004 `role_name_branch_mutant_rejected`**, **AC-006**(모집단 전제) |
| 실제 코드 | exit=0 | exit=0 | — |

수리 뒤 m1 원문(발췌):

```
== final-m1 exit=1
fail TestCodexRoleLoadPredicateStructurallyIndependent/input_type_has_no_behaviour_field
fail TestCodexRoleLoadPredicateStructurallyIndependent/exhaustive_field_mutations
fail TestCodexRoleLoadPredicateStructurallyIndependent/half_true_arm
pass TestCodexRoleLoadPredicateStructurallyIndependent/coupled_mutant_rejected
fail TestCodexRoleLoadPredicateStructurallyIndependent
    codex_role_behaviour_test.go:201: codexRoleLoadInput must carry exactly {Role string, Matched map[string]bool}; got [Role string Matched map[string]bool Extra cli.codexRoleBehaviour]
    codex_role_behaviour_test.go:224: field 0 mutation found 1 behaviour site(s) in codexRoleLoadInput
    codex_role_behaviour_test.go:259: half-true arm changed the load verdict on 6 of 24 cases
```

수리 뒤 m2 원문(발췌):

```
== final-m2 exit=1
fail TestCodexRoleLoadPredicateStructurallyIndependent
fail TestCodexRoleLoadPredicateContractNeutral/role_name_branch_mutant_rejected
fail TestCodexRoleLoadPredicateContractNeutral
fail TestCodexRoleBodyFingerprintDerive/N1_version_mismatch
fail TestCodexRoleBodyFingerprintDerive/N2_absent_roles
fail TestCodexRoleBodyFingerprintDerive/N5_real_parent_sessions
fail TestCodexRoleBodyFingerprintDerive
    codex_role_behaviour_test.go:172: expected 12 load-true and 12 load-false cases, got 14 true of 24
    codex_role_contract_test.go:211: codexRoleLoadPredicate is not uniform across roles: [manager-lead/empty=true manager-lead/others=true mission-governor/empty=true mission-governor/others=true]
```

수리 뒤 실제 코드:

```
== final-real exit=0
(30개 pass, 판정식이 보는 토큰 원문)
BEHAVIOUR_FIELDS_MUTATED 2 OF 2
COUPLED_MUTANT_DIVERGED 24
CONTRACT_NEUTRAL_LOAD_TRUE 12 NONCE_TRUE 10
FINGERPRINT_POSITIVE_MATCHES 12
PARENT_SESSIONS_FALSE 14
SELECTED_BY_LABEL 14 OF 28
PARSE_LEADING_DELTA 1 0x0a
TABLE_DISTINCT_REASONS 3
```

`half_true_arm` 은 첫 수리안에서 m1 을 통과시켰다(`i%2` 분할이 세션당 두 행 중 로드 참 행에만 행동 참을 줬다). 세션 단위 분할 `(i/2)%2` 로 고친 뒤 위 결과가 나왔다 — 첫 수리안의 m1 실행 기록에 `pass …/half_true_arm` 이 남아 있다.

### 2.3 F5 — 미측정과 거짓의 분리

테스트 먼저 작성한 뒤 생산 코드를 수리 전 판(`git show HEAD:internal/cli/codex_role_fingerprint.go`)으로 되돌린 오버레이에서 실행:

```
$ go test -json -count=1 -run '^TestCodexRoleLoadOutcomeSeparatesUnmeasured$' -overlay head.json ./internal/cli
== f5-red-head exit=1
internal/cli/codex_role_outcome_test.go:44:28: undefined: codexRoleFingerprintObservation
internal/cli/codex_role_outcome_test.go:45:15: undefined: codexRoleFingerprintObserve
internal/cli/codex_role_outcome_test.go:57:13: undefined: codexRoleLoadOutcome
```

작성 순서에 관한 고지: F5 의 생산 코드는 이 테스트보다 **먼저** 편집기에 썼다. 위 RED 는 테스트를 쓴 뒤 수리 전 판으로 되돌려 얻은 것이고, 판별력은 아래 두 행동 변이로 따로 확인했다.

| 변이 | 바꾼 것 | 결과 |
|---|---|---|
| m5 | `codexRoleLoadOutcome` 에서 `DeveloperItems == 0 → NOT_RUN` 분기 삭제 | exit=1 — `zero_developer_items`, `empty_record` FAIL (`outcome = "FAIL", want "NOT_RUN"`) |
| m6 | 스캔에서 `developerItems++` 삭제 | exit=1 — `measured_body_absent`, `measured_body_present` FAIL (`outcome = "NOT_RUN", want "PASS"`) |
| 실제 코드 | — | `== f5-real exit=0`, 부분 테스트 5개 pass |

`predicate_alone_conflates` 부분 테스트는 불리언 판정만으로는 두 기록이 모두 `false` 라 구분되지 않음을 대조로 남긴다. 판정 입력 타입(`Role`, `Matched`)은 바꾸지 않았다 — `codexRoleLoadOutcome` 은 그 위에서 관측 수를 함께 본다. 테스트 파일은 `NOT_RUN` 문자열을 출력하지 않는다(AC 판정식이 그 토큰을 거부하므로).

### 2.4 F2 — 다른 키의 여러 줄 문자열

같은 코드 경로(`codexRoleBodyExtractTOML`)의 작은 수리라 이 카드에서 처리했다. 테스트 먼저:

```
== f2-red exit=1
fail …/desc_line_starts_with_bracket      extract: codex role body: developer_instructions key absent
fail …/desc_line_mimics_key               extract: codex role body: developer_instructions value is not a TOML multi-line literal string
fail …/basic_desc_bracket_and_key         extract: codex role body: developer_instructions key absent
pass …/single_line_triple_quoted
pass …/unclosed_other_value_is_an_error
```

처음 두 실패 메시지는 감사 §2.6 탐침 출력과 같다. 수리(`codexRoleSkipMultilineValue` — 다른 키의 `'''`/`"""` 값을 닫는 줄까지 건너뜀) 뒤:

```
== f2-green exit=0
pass TestCodexRoleBodyExtractSkipsOtherMultilineValues (5 부분 테스트)
pass TestCodexRoleBodyFingerprintParse (5 부분 테스트)
pass TestCodexRoleBodyExpectationTable (4 부분 테스트)
PARSE_LEADING_DELTA 1 0x0a
TABLE_DISTINCT_REASONS 3
```

역할 TOML 12개의 추출값은 바뀌지 않았다 — `P_all_twelve`(12/12 일치)와 AC-RLP-002 가 그대로 통과한다. 함수 주석도 「TOML 명세 파서」가 아니라 방출 형태를 다루는 줄 스캐너임을 밝히도록 좁혔다.

### 2.5 F7 — 픽스처 경로 중립화

`internal/cli/testdata/codex-rollouts-t1171/` 의 `.jsonl` 에 바이트 치환(스크립트가 치환 뒤 모든 줄을 JSON 으로 다시 파싱):

```
/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn 1133   → /private/var/folders/fixture
/Users/goos 78                                                → /home/fixture
files changed: 28 mode: apply
residual /Users/|/var/folders/kt/|goos hits: 0
```

감사가 센 78건 외에 macOS 사용자별 임시 디렉터리 경로 1133건도 같은 성격(사용자·머신 고유)이라 함께 치환했다. 치환 뒤 이메일·`/home|/Users/<name>` 재탐침 결과는 `[('/home/fixture', 78)]` 뿐이다. `roles/`·`roles-other-version/` 24개는 경로가 없어 무변경이다.

핀 갱신 — 이 픽스처의 sha256 핀은 저장소 안에서 `.moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/progress.md` 표 한 곳뿐이다(`git grep` 로 확인, 테스트가 핀을 강제하는 곳은 없다):

```
rows: 52 changed: 28
post-apply mismatches: 0
```

표 바로 위에 재고정 사유와 이전 핀의 위치(커밋 `1a8dfa26e`)를 한 문단으로 적었다.

### 2.6 AC 판정식 재실행 (acceptance.md §B 원문, 출력 경로만 스크래치패드)

```
AC-RLP-001: true exit=0
AC-RLP-002: true exit=0
AC-RLP-003: true exit=0
AC-RLP-004: true exit=0
AC-RLP-006: true exit=0
```

### 2.7 영향 패키지 검증

```
$ go test -count=1 -timeout=1200s -run '^TestCodexRole' -v ./internal/cli/
--- PASS: TestCodexRoleLoadPredicateStructurallyIndependent (0.07s)
--- PASS: TestCodexRoleLoadPredicateContractNeutral (0.06s)
--- PASS: TestCodexRoleBodyFingerprintDerive (0.13s)
--- PASS: TestCodexRoleBodyExtractSkipsOtherMultilineValues (0.00s)
--- PASS: TestCodexRoleBodyFingerprintParse (0.00s)
--- PASS: TestCodexRoleBodyExpectationTable (0.06s)
--- SKIP: TestCodexRoleLiveLoadAndReadOnly (0.00s)
--- PASS: TestCodexRoleLoadNegativeControl (0.03s)
--- PASS: TestCodexRoleLoadProbeWordingIsUniform (0.00s)
--- PASS: TestCodexRoleLoadOutcomeSeparatesUnmeasured (0.05s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.164s

$ go test -count=1 -timeout=1200s ./internal/harness/rosterguard/
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	26.674s

$ go vet ./internal/cli/ ./internal/harness/rosterguard/
vet_exit=0
$ gofmt -l internal/cli/codex_role_*.go
(출력 없음)
$ golangci-lint run --timeout=10m ./internal/cli/... ./internal/harness/rosterguard/...
0 issues.
lint_exit=0
```

`./internal/cli/...` 전체 실행 결과: 하위 패키지 17개는 모두 `ok`, `internal/cli` 본 패키지는 `panic: test timed out after 20m0s`(당시 실행 중 테스트 `TestTodoQueueRootGuard_SilentOnHomeFallbackFixture_NonTemp (0s)`, `--- FAIL` 행 0건). 실패가 아니라 시간 상한 초과이며, 본 패키지 전체 판정은 CI 몫으로 남긴다(§4).

`codex_role_fingerprint.go` 파일 커버리지(`-run '^TestCodexRole'`): `stmts=142 covered=119 pct=83.8` (감사 시점 82.5%). `codexRoleLoadOutcome`·`codexRoleLoadPredicate` 100%, `codexRoleSkipMultilineValue` 90.0%.

## 3. 기준 귀속

- 트리: 이 워크트리. 수리 전 측정은 HEAD `b4f798dcc`(작업 트리 깨끗)에서, 수리 뒤 측정은 커밋 `515b05e06`·`4e4a87501` 의 내용이 작업 트리에 있던 상태에서 실행했다.
- 모든 수치는 이번 실행에서 관측한 값이다. 감사 문서의 수치는 대조용으로만 인용했고(m1·m2 재현, 탐침 메시지), 재측정으로 같은 값을 얻었다.
- 변이 주입은 `-overlay` — 오버레이 파일과 원시 출력은 머신 로컬 스크래치패드(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/f3926fad-6b6a-4bdd-8856-1152613006aa/scratchpad/`)에 있고, 판정을 가른 줄은 위에 원문으로 옮겼다.

## 4. 미검증 (Gaps)

- **`internal/cli` 본 패키지 전체 스위트**: 20분 상한에서 시간 초과로 끝났다(실패 행 0). 전체 판정은 develop push 뒤 CI 가 낸다.
- 커밋 `515b05e06` **단독** 상태(픽스처 치환 전)에서는 테스트를 따로 돌리지 않았다. 수리 뒤 측정은 두 커밋의 내용이 함께 있는 트리에서 했다. 테스트는 경로 문자열을 읽지 않으므로 결과가 다를 이유는 없으나, 관측하지는 않았다.
- AC-RLP-008·009 판정식은 재실행하지 않았다(이 카드는 형제 SPEC 디렉터리와 acceptance.md 를 건드리지 않았다).
- darwin 외 플랫폼, 교차 모델 감사 미실행.
- `.moai/reports/t1143/` 원본과의 바이트 동일성(감사 §2.4 `26 same`)은 F7 치환으로 의도적으로 깨졌다 — 재확인 대상이 아니다.

## 5. 잔여 위험

- F1 방어는 반사로 「입력 타입이 정확히 두 필드」임을 고정한다. 판정 함수가 전역 상태나 패키지 변수를 통해 행동값을 읽는 결합은 입력 타입을 거치지 않으므로 이 테스트가 보지 못한다.
- F2 수리는 기본 문자열 안의 이스케이프된 `\"""` 를 인식하지 않는다(주석에 적음). 역할 파일은 리터럴 문자열만 쓴다.
- 암호화 추론 블롭(감사 F7 의 둘째 항목)은 계정에 묶인 불투명 데이터라 치환할 수 없어 그대로 남았다.

## 6. 이월 항목과 사유

| 항목 | 처분 | 사유 |
|---|---|---|
| F1 의 `spec.md` §C.5 3행 문장 정정 | 이월 | SPEC 본문 수정은 manager-spec 소관(이 에이전트 금지 범위). 이제 그 행이 적은 담보가 테스트로 성립하므로 정정 필요성은 줄었다 |
| F4 `@MX:ANCHOR` 과장 | 이월 | 카드 범위(F1·F5·F7, F2·F3 선택) 밖 |
| F6 파일 커버리지 85% 미달(83.8%) | 이월 | 카드 범위 밖. 남은 미커버는 `codexRoleSessionLabel`(70.4%)·스캔 오류 분기 |
| F8·F9·F10 | 이월 | SPEC 본문 서술 문제 — manager-spec 소관 |
| plan.md §G 1행과의 정합 | 해소 | 이제 `codexRoleLoadOutcome` 이 그 행의 `NOT_RUN` 규약을 구현한다. 문서 수정은 불필요 |

## 7. 바꾼 파일

- `internal/cli/codex_role_behaviour_test.go` — AC-006 재작성(허용목록, 반사 주입 하네스, 참·거짓 두 모집단, 필드별 양성 대조)
- `internal/cli/codex_role_contract_test.go` — AC-004 균일성 검사를 생산 판정에 적용, 양방향 분기 람다를 양성 대조로
- `internal/cli/codex_role_fingerprint.go` — `codexRoleFingerprintObserve`·`codexRoleFingerprintObservation`·`codexRoleLoadOutcome`(F5), `codexRoleSkipMultilineValue`(F2), 추출 함수 주석 범위 정정
- `internal/cli/codex_role_outcome_test.go` — 신규(F5)
- `internal/cli/codex_role_extract_test.go` — 신규(F2)
- `internal/cli/testdata/codex-rollouts-t1171/real/*.jsonl`(26), `synthetic/*.jsonl`(2) — 경로 중립화(F7)
- `.moai/specs/SPEC-ROLE-LOAD-PREDICATE-001/progress.md` — 픽스처 sha256 핀 28행 갱신 + 재고정 문단
- `.moai/reports/t1260/verdict.md` — 이 문서
