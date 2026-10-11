# SPEC-CODEX-CONFORMANCE-001 — acceptance.md

## §A. 판정 원리

- 기계 판정 우선: 명명된 시험 계열 + grep 가능 행동 + 단일 명령 증거. 카드 워크트리에서 `internal/cli` 전체 스위트는 구조적 적색이므로 레인-로컬 판정은 아래 명명 계열로 한정하고, 전체 스위트는 이 카드의 main 통합 PR 병합 헤드 CI run이 등판한다(§E DoD-5).
- 두-셀 채용(verification-completeness §2): 차단 기준에는 RED-now 셀(§B 증거 장부, 본 계획 기준 트리에서 실측)과 green 경로 셀(뒤집는 마일스톤 명시)을 함께 둔다. 보존 기준은 회귀 가드로 분류한다(사전 관측 가능한 RED가 없는 "지금 green이 유지된다" 계열 — 계획 시점 판정으로 기록하지 않는다). AC-CONF-005/006은 RED가 M2 관측·M3 생성 시점에만 존재하므로 차단 채용을 각 E8 원문 시점으로 미룬다(§B).
- 운영자 환경 의존 검사(0.161.0 라이브 관측)는 자동 AC로 쓰지 않는다 — sanitized vendored 샘플 또는 증거를 실은 수동 관측 기록(명령+verbatim 출력+바이너리 버전)으로만 판정한다(REQ-CONF-007).

## §B. 증거 장부 (RED-now 셀 — 트리 81786284e, 2026-10-08/09 실측)

| id | 명령 (단일 호출) | verbatim stdout | exit | 왜 RED인가 | 뒤집는 마일스톤 |
|----|---|---|---|---|---|
| LEDGER-1 | `grep -rn "codex-0.161.0" internal/cli` | (출력 없음) | 1 | 0.161.0 fixture와 소비자 참조가 아직 존재하지 않는다 — REQ-CONF-001/002 미충족 상태의 관측 | M1 |
| LEDGER-2 | `ls internal/cli/testdata` | `drwxr-xr-x@   12 goos  staff    384 10월  8 19:02 codex-0.160.0` 행 존재, `codex-0.161.0`로 시작하는 행 부재 — 41행 목록 중 위 발췌 1행이 원문 바이트고 나머지 40행은 본 AC와 무관한 타 fixture다(생략 명시) | 0 | 대상 fixture 디렉터가 없다 — 목록면 관측(부재의 결정 관측은 LEDGER-3) | M1 |
| LEDGER-3 | `ls internal/cli/testdata/codex-0.161.0` | stdout 없음 / stderr: `ls: internal/cli/testdata/codex-0.161.0: No such file or directory` | 1 | REQ-CONF-001 미충족의 직접 관측 — AC-CONF-001/002/003의 **공유 RED 앵커** | M1 |

**공유 RED 앵커의 AC별 실패 입력** (LEDGER-1/3이 같은 원인 — 대상 fixture 부재 — 의 세 AC 확산이다):

- AC-CONF-001: LEDGER-1/3 그 자체가 RED다(파일 셋 부재 + 리터럴 잔존).
- AC-CONF-002: 실패 입력은 소비자가 0.161.0 디렉터를 가리킬 때 누락·훼손된 스키마 파일 — `hardenCompileSchema`의 `t.Fatalf("read vendored schema %s: %v", file, err)`(`internal/cli/managed_hardening_test.go`)이 관측 가능한 실패다. 현 시점 시험이 0.160.0 세트로 green인 것은 본 AC의 RED가 아니라 그 가드가 아직 0.161.0을 소비하지 않는다는 뜻이며, RED는 LEDGER-3이 대리 관측한다.
- AC-CONF-003: 실패 입력은 (a) resume-help.txt 부재 시 `os.ReadFile` 실패(`t.Fatal`, `internal/cli/managed_codex_tui_test.go` `TestManagedCodexRemoteSupportProbe`의 `real_help_supported` 다리) 또는 (b) 옵션 제거 세계에서의 양제 단언 실패 — (b)는 REQ-CONF-004의 When 분기가 발동한 upstream 변화의 관측이며 AC-CONF-003의 명시적 완결 경로(아래)로 처리된다.

**신설 AC의 채용 지연(E8 규약)**: AC-CONF-006 — 시험과 fixture는 M3에서 새로 만들어지므로 plan 시점에 실행 가능한 RED 셀은 존재할 수 없고, 차단 채용은 M3의 E8 원문(구현 전 적색 출력, 그 시점 트리 SHA와 함께 progress.md §E.2 첨부)이 붙는 시점으로 미뤄진다. AC-CONF-005 — LEDGER-1~3은 디렉터·참조 부재만 관측하므로(인증 분류 실패의 관측이 아님) 차단 채용은 M2 최소 관측(미인증 0.161.0 `codex login status` 출력 포획, E8 원문 첨부) 시점으로 미뤄진다. 두 AC 모두 부재 관측으로 차단 지위를 주장하지 않는다.

## §C. AC-CONF 행렬

### AC-CONF-001 — 0.161.0 fixture 세트 착지 + 구(舊) 리터럴 소멸 (차단; LEDGER-1/3) (maps REQ-CONF-001, REQ-CONF-002)

- **Given** M1 시작 시점의 트리(0.160.0 핀만 존재) **When** fixture 재생성과 소비자 이전이 완료되면 **Then** `internal/cli/testdata/codex-0.161.0/`이 스키마 8종(ApplyPatchApprovalResponse, ExecCommandApprovalResponse, CommandExecutionRequestApprovalResponse, FileChangeRequestApprovalResponse, PermissionsRequestApprovalResponse, McpServerElicitationRequestResponse, DynamicToolCallResponse, JSONRPCError) + `resume-help.txt` + README(0.161.0 명기)로 존재하고, `grep -rn "codex-0.160.0" internal/cli`의 발생 수는 0이다.
- 검증: `ls internal/cli/testdata/codex-0.161.0`(8+2 파일), `grep -rn "codex-0.160.0" internal/cli | wc -l` → `0`. green 경로: M1.
- RED 소멸 조건 주의: `grep -rn "codex-0.160.0" internal/cli | wc -l`가 0이 되려면 `hardenCompileSchema`의 경로·URL 리터럴과 `tuiResumeHelpFixture`가 모두 이전돼야 한다 — 리터럴 3행 중 하나라도 남으면 FAIL.

### AC-CONF-002 — 스키마 가드가 0.161.0 스키마 위에서 통과 (차단) (maps REQ-CONF-003)

- **Given** 0.161.0에서 생성된 vendored 스키마 8종 **When** `go test ./internal/cli -run '^TestManagedServerRequestPolicyMatchesCodexSchema$' -count=1` **Then** 통과 — 기각 답 객체 포함 소유자의 일곱 종 답 본문이 0.161.0 응답 스키마를 만족한다. green 경로: M1. 상위 스키마에서 필드가 줄거나 필수가 바뀌면 이 시험이 적색으로 검출한다(기기의 실패 관측성 보장).

### AC-CONF-003 — resume-help 프로브가 0.161.0 텍스트와 일치 (차단) (maps REQ-CONF-004)

- **Given** 0.161.0에서 재생성된 `resume-help.txt` **When** `go test ./internal/cli -run '^TestManagedCodexRemoteSupportProbe$' -count=1` **Then** 통과 — `real_help_supported` 다리가 새 fixture를 양제(`--remote`+`--remote-auth-token-env`)로 읽는다.
- **옵션 제거 세계의 명시적 완결 경로**: 0.161.0 텍스트가 두 옵션을 모두 제공하지 않으면, ① `real_help_supported` 다리의 기대값을 **관측된 0.161.0 현실(미지원)로 갱신**한다 — 이 기대 갱신은 승인된 메커니즘이며 갱신 근거(실제 새 help 텍스트의 원문 관측)를 기록에 남긴다. ② 갱신 뒤 프로브 시험과 폴백 검증(`TestManagedCodexTUIPreconditionsAndFallback`, `internal/cli/managed_codex_tui_test.go:1084`)을 **실제 새 help 텍스트를 입력으로** 실행해 명명 계열 전부 green을 만든다 — §D.1·DoD-2의 family-green 요구와 정합이며, 합성 텍스트만의 녹색으로 대체하지 않는다. ③ 어댑터 후속(REQ-CONF-004의 When 분기 판정)은 blocker 보고로 남긴다. 옵션 존재 세계에서는 기대 갱신 없이 첫 Then(통과)이 완결 경로다. **두 세계 모두에서 family green이 M1 폐쇄 조건이다.**
- green 경로: M1. RED 앵커: LEDGER-1/3 + 실패 입력(§B).

### AC-CONF-004 — auth 부재 = stage 2 하강, 미인증 아님 (회귀 가드) (maps REQ-CONF-006)

- **Given** `auth.json`이 없는 CODEX_HOME fixture와 문법 미일치를 돌려주는 stage 2 러너 시임 **When** `classifyCodexAuth`가 실행되면 **Then** 판정은 `codexAuthUnknown`이다(프로바이더 문자열 아님). 검증: `internal/cli/codex_auth_ladder_test.go`·`mcp_codex_test.go`의 사다리 계열 — run-phase에서 부재-파일 케이스가 명시적으로 존재하는지 확인하고 없으면 M2가 RED/GREEN 한 쌍으로 추가한다. green-now 기준선: Pre-flight의 사다리 계열 실행.

### AC-CONF-005 — stage 2 문법의 0.161.0 출력 정합 (차단 — 채용은 M2 최소 관측 시점, RED는 E8 규약) (maps REQ-CONF-007)

- **Given** 0.161.0에서 포획해 정화한 `codex login status` 출력 샘플 **When** `parseCodexAuthLine`이 그 샘플을 분류하면 **Then** 0.160 세대 동일 상태 샘플과 같은 판정을 내린다(자동 시험, vendored 샘플 fixture).
- **채용 지연(E8 규약)**: LEDGER-1~3은 디렉터·참조 부재를 관측하지 auth 분류 실패를 관측하지 않으므로 본 AC의 RED는 plan 시점에 존재할 수 없다 — 차단 채용은 M2의 최소 필수 관측(아래 관측 의무)이 수행되는 시점이며, 그 관측의 명령·verbatim 출력·바이너리 버전이 E8 원문으로 progress.md §E.2에 첨부될 때 성립한다.
- **관측 의무(DoD-1과 짝)**: M1/M2 환경은 0.161.0 바이너리를 구조적으로 보유한다(plan §A — `npx -y @openai/codex@0.161.0`). 따라서 **최소 관측은 필수**다: 미인증 상태의 0.161.0 `codex login status` 출력 형태를 포획한다. 수동 검증 기록은 증거를 실어야 한다 — 포획 명령 + verbatim 출력 + 바이너리 버전(verification-claim-integrity §3 형식). **keyring 로그인 변형만** 포획 불가 시 기록 갭으로 남을 수 있다; 인증 표면 전체가 미관층으로 남는 것은 열린 항목(pending)이지 DoD 성립이 아니다.
- green 경로: M2.

### AC-CONF-006 — 재생 페이로드 내성 (차단 — 채용은 M3, RED는 E8 규약) (maps REQ-CONF-005)

- **Given** 어댑터가 모델링하지 않는 추가 필드를 실은 재개 재생 프레임 fake-server fixture — 그 프레임 형상의 **근거를 기록으로 남긴다**: 1차 근거는 **0.161.0 바이너리 포획 또는 해당 버전의 상위 정의(rust-v0.161.0)에서 유도된 프레임**이며, **fake 서버가 스스로 생성한 응답은 출처로 인정하지 않는다**(자기-생성 순환 — 그 경로로는 발명된 필드가 근거 절을 통과한다). 0.161.0 근거를 확보하지 못하면 #49599 변화 요약만을 근거로 한 추정임을 명시한다(0.160.0 vendored 스키마에는 재생 프레임 정의가 없고 그 디렉터는 M1이 삭제하므로 그 참조는 근거가 아니다). 근거를 전혀 확보하지 못하면 본 AC는 채용 지연과 같이 pending으로 보류한다(상상 필드 금지) — **When** 관리 소유자가 재개 스레드를 소비하면 **Then** 디코드 실패도 기표된 턴 실패(`Factory turn failed`)도 기록되지 않고, **양성 대조**로 재생 프레임 뒤의 정상 턴이 완료된다 — 소비가 실제로 일어났음을 보이는 통제다(기존 fake TUI가 재생 프레임을 읽고 버리는 경로만으로는 본 AC를 만족시킬 수 없다).
- RED 관측 계획(E8 규약): 시험명은 M3에서 확정되고, 구현 전 적색 출력을 그 시점 트리에서 원문 캡처해 progress.md §E.2에 첨부한다. 본 AC의 차단 채용은 E8 원문이 붙는 시점(M3)에 성립한다. green 경로: M3.

### AC-CONF-007 — escalation NO-OP: 기각 정책 불변 (회귀 가드) (maps REQ-CONF-008)

- **Given** 갱신된 fixture **When** `go test ./internal/cli -run '^TestManagedCodexServerRequestPolicy$' -count=1` **Then** 통과 — 승인된 파일시스템 escalation(#49353) 의미 확장과 무관하게 소유자의 답 정책표(decline/거부 객체/오류 코드)는 불변이다. green-now 기준선: Pre-flight.

### AC-CONF-008 — 모델 핀 정합 (회귀 가드) (maps REQ-CONF-008)

- **Given** 0.161.0의 GPT-6.1 Sol 기본 카탈로그 승격(#49318/#49339) **When** `go test ./internal/config -count=1` **Then** 통과 — `DefaultCodexAuditModel = "gpt-6.1-sol"`(`internal/config/closed_sets.go:96`)과 defaults의 `{gpt-6.1-sol, high}` Codex 핀은 변경 없음이 단언된다. 어댑터 코드 변경 0건. green-now 기준선: Pre-flight.

## §D.1 품질 게이트 기준 (TRUST 5)

- Tested: 영향 계열 전부 green + 신규 시험의 RED 원문 보관(E8). Unified: gofmt 대상 파일 무변경 또는 포맷 완료. Secured: 정화 규약(토큰·계정 식별자 무포함) — vendored 샘플에 자격증명 문자열이 있으면 즉시 FAIL. Trackable: 커밋에 카드 id + SPEC id. Readable: fixture README가 생성 명령·버전을 새로 기술.

## §E. Definition of Done

1. AC-CONF-001~008 전부 PASS(AC-CONF-005는 자동 PASS, 또는 **증거를 실은** 수동 관측 기록 — 포획 명령 + verbatim 출력 + 바이너리 버전, 미인증 0.161.0 출력 형태 포획을 최소 관측으로 하고 keyring 로그인 변형만 기록 갭 허용 — 둘 중 하나로 성립한다. 증거 없는 갭 기록은 성립이 아니다).
2. 레인-로컬 명명 계열 green + `GOOS=windows go build ./...` 통과.
3. `.moai/docs/factory-managed-session.md` 버전 표기 갱신(M4).
4. progress.md §E.2에 마일스톤별 관측 기록 + §E.3 run audit-ready.
5. **전체 스위트 등판면은 이 카드의 main 통합 PR과 그 병합 헤드에서 실행된 CI run이다.** develop push에는 CI 트리거가 없다 — `.github/workflows/ci.yml`의 트리거는 `on.push.branches: [main]`과 `on.pull_request.branches: [main]`뿐이다(실측 2026-10-09). 레인-로컬 명명 계열 green은 조기 신호일 뿐이고 병합 헤드의 CI가 등판 판정면이다.

## §F. 전방향 확인 (forward-looking)

- 다음 스윕 기준(상위 리포 인용): 업스트림 안정이 0.162로 승격되거나 0.161.0 이후 델타가 쌓이면 이 fixture 세트가 다시 낡는다 — 본 SPEC의 재생성 절차(README 기재 명령)가 그 다음 카드의 입력이다.
