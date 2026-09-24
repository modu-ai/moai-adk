---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Acceptance — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.2.0"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Acceptance — SPEC-HOOK-STDIN-FAILCLOSED-001

모든 기준은 명령 출력이나 파일 판독으로 관측 가능하다. 각 PASS 는 실행한 명령과 원문 출력을, 주장하는 실행에서, 측정 대상 트리에 대해 남긴다(`verification-claim-integrity.md` §2).

0.2.0 개정: plan-audit 1회차(`.moai/reports/t1152/plan-audit.md`)의 D1(관측 이벤트 출력의 이벤트별 정정, AC-003 관측 술어), D4(depth 형태 추가), D5(AC-003 정적 검사를 AST 검사로 교체), D6(형태별 카나리), D7(사유 필수 요소 단언), D3(탈출 장치 출처 제약 음성 기준 AC-HSF-010 신설), D11(줄 번호) 을 반영했다.

## §0 공통 입력 — 「파손 4 형태」와 카나리

각 형태는 **서로 다른 카나리 문자열**을 품는다. 카나리는 테스트 실행 시 `CANARY-<FORM>-` 뒤에 `crypto/rand` 로 만든 16진 문자 64개를 붙여 만든다(형태마다 70바이트 이상, 서로 다름). 64바이트 이상으로 잡는 이유: `encoding/json` 오류 메시지는 입력의 한 글자나 숫자 리터럴 정도를 인용할 수 있어(깊이 초과 오류는 `invalid character '['` 로 관측됨), 짧은 표식은 합법적 오류 인용과 구별되지 않는다.

| 이름 | 내용 | 파싱 실패 원인 |
|---|---|---|
| malformed | `{"broken":"<CANARY-MALFORMED>` | 닫히지 않은 문자열 |
| truncated | `tool_input.note` 에 `<CANARY-TRUNCATED>` 를 담은 유효한 결정 이벤트 페이로드를, 카나리 **뒤**에서 잘라 앞부분만 넣는다(닫는 중괄호 없음) | 입력 조기 종료 |
| oversize | `{"hook_event_name":"<이벤트>","tool_input":{"note":"<CANARY-OVERSIZE>","content":"` + 6 MiB 의 `a` + `"}}` | 5 MiB 상한(`internal/hook/protocol.go:44`)에서 잘려 파싱 실패 |
| depth | `{"hook_event_name":"<이벤트>","tool_input":{"note":"<CANARY-DEPTH>","x":` + `[` 10001개 + `]` 10001개 + `}}` (약 20 KB) | `encoding/json` 중첩 깊이 상한 초과 — `normalizeHookInput` 첫 단계의 `json.Unmarshal`(`internal/hook/normalize.go:85-88`)에서 발생. 대조: 깊이 9000 은 같은 타입으로 파싱된다(spec.md §A.3 측정) |

depth 형태는 run 착수 시 대조군을 함께 잰다: 같은 구성의 깊이 9000 입력은 `ReadInput` 이 오류를 돌려주지 **않아야** 한다. 대조군이 실패하면 depth 형태의 결과를 판정에 쓰지 않고 원인을 먼저 보고한다.

「결정 이벤트」는 `codexadapter.DecisionBearingEvents()` 의 원소, 「관측 이벤트」는 `runHookEvent` 에 연결된 하위 명령 중 그 밖의 것(착지 시점 기준 22개)이다. 목록을 테스트 안에 손으로 적지 않고 두 원천(`DecisionBearingEvents()`, `hook.go` 의 하위 명령 표)에서 계산한다.

「거부 필드」는 spec.md §B.4 의 열이다: PreToolUse `hookSpecificOutput.permissionDecision == "deny"`, PermissionRequest `hookSpecificOutput.decision.behavior == "deny"`, Stop·UserPromptSubmit 최상위 `decision == "block"`.

## §1 요구사항 → 기준 대응과 수리 전 트리의 기대

두 칸 원칙(`verification-completeness.md` §2): 각 기준에 수리 전 트리에서 무엇이 보여야 하는지(RED 이유)와, 무엇이 그것을 뒤집는지를 함께 적는다. 수리 전 관측은 plan.md Pre-flight 4 의 기준선 커밋이 담는다.

| REQ | AC | 수리 전 트리(Pre-flight 기준선)에서의 기대 | 무엇이 뒤집는가 |
|---|---|---|---|
| REQ-HSF-001 | AC-HSF-001, AC-HSF-002 | RED — stdout 이 `{}` 라 (c) 가 실패 | M1 |
| REQ-HSF-002 | AC-HSF-003 | RED — 관측 집합이 공집합이라 (a) 가 실패. AST 검사는 수리 전에도 녹색일 수 있으므로 중복 `switch` 변이로 RED 를 따로 관측 | M1 |
| REQ-HSF-003 | AC-HSF-002 | RED — (c) 실패 | M1 |
| REQ-HSF-004 | AC-HSF-001 | RED — (c) 실패 | M1 |
| REQ-HSF-005 | AC-HSF-004 | RED — `--harness bogus` 에서 stdin 경고가 먼저 나온다 | M3 |
| REQ-HSF-006 | AC-HSF-005 | 녹색(특성 기준) — RED 는 관측 이벤트 출력을 바꾸는 변이로 관측 | 뒤집히지 않아야 함 |
| REQ-HSF-007 | AC-HSF-006, AC-HSF-010 | RED — 탈출 장치가 없다(AC-006). AC-010 은 수리 전에도 녹색일 수 있으므로 「설정면 값을 활성화로 인정하는」 변이로 RED 를 관측 | M2 |
| REQ-HSF-008 | AC-HSF-007 | RED — stderr 한 줄·기록 모두 없다 | M2 |
| REQ-HSF-009 | AC-HSF-008 | RED — `decision == "block"` 이 아니다 | M1 |
| REQ-HSF-010 | AC-HSF-001(e), AC-HSF-002(e) | RED — 사유 자체가 없다 | M1 |
| REQ-HSF-011 | AC-HSF-007(c) | 수리 전에는 기록이 없어 공허하게 녹색이다 — 원문을 기록에 싣는 변이로 RED 를 관측 | M2 |

## §2 기준

### AC-HSF-001 — Claude 하네스: 결정 이벤트 × 파손 4 형태 → 거부

**Given** `--harness` 미지정, 탈출 장치 꺼짐, 실제 `hook.NewProtocol()`, 호출 횟수를 세는 스파이 레지스트리, **When** 각 결정 이벤트의 하위 명령에 파손 4 형태를 각각 stdin 으로 넣고 `RunE` 를 실행하면, **Then** (a) `RunE` 가 nil 을 돌려준다(exit 0), (b) 디스패치 횟수가 0 이다, (c) stdout 이 JSON 값 정확히 하나이며 `{}` 도 빈 문자열도 아니다, (d) 그 바이트가 `Render(ev, Lookup(HarnessClaude, ev, DecisionFatalError).Outcome, <사유>)` 와 같고 그 이벤트의 거부 필드(§0)를 갖는다, (e) 사유 필드가 비어 있지 않으며 (e1) `fail-closed`, (e2) 원인이 stdin 파싱 실패라는 고정 문구(run-phase 상수), (e3) 운영자 문서 식별자(run-phase 상수)를 담고, (e4) 탈출 장치의 식별자(plan.md Q1 에서 정한 환경 변수 이름 또는 그에 해당하는 키)와 어떤 카나리도 담지 않는다. 4 × 4 = 16 경우 전부. (e2)·(e3)·(e4) 는 테스트가 리터럴이 아니라 구현의 상수를 참조해 단언한다.

### AC-HSF-002 — Codex 하네스: 결정 이벤트 × 파손 4 형태 → 거부

**Given** `--harness codex`, 나머지는 AC-HSF-001 과 같은 조건, **When** 같은 16 경우를 실행하면, **Then** (a)~(c) 가 성립하고, (d) stdout 바이트가 `codexadapter.TranslateCodex(ev, DecisionFatalError, <구현이 넘긴 원인 문자열>)` 의 출력과 같고 거부 필드를 가지며, (e) AC-HSF-001(e1)~(e4) 가 같은 방식으로 성립한다. PermissionRequest 는 Codex EventTable 에서 미적응(U)이지만 하위 명령은 존재하므로 같은 기준을 적용한다. Stop 은 plan.md Q1 판정이 Codex Stop 을 fail-closed 에서 제외하는 분기(A1)를 택한 경우 이 기준에서 빠지고 AC-HSF-005 의 관측 기준을 따른다 — 그 경우 AC 개정이 D-NEW-1 로 먼저 착지한다.

### AC-HSF-003 — 결정 이벤트 목록은 하나뿐이다

**(a) 동작 등가.** **Given** 착지된 구현, **When** 테스트가 `runHookEvent` 에 연결된 모든 하위 명령(26개)에 malformed 를 넣고, 두 하네스 각각에서 「stdout 이 JSON 으로 파싱되고 §0 의 거부 필드 가운데 하나를 갖는 이벤트 집합」을 관측하면, **Then** 그 집합이 `codexadapter.DecisionBearingEvents()` 와 원소 단위로 같다. 빈 stdout(worktree-create·worktree-remove)과 `{}` 는 거부 필드가 없으므로 집합에 들지 않는다.

**(b) AST 검사.** `internal/cli` 의 테스트가 아닌 Go 파일 전부를 `go/parser` 로 읽는 테스트(또는 같은 규칙의 `ast-grep` 규칙)가 다음을 단언한다:
- (b1) `runHookEvent` 안에서 `ReadInput` 호출 결과의 오류 분기(`if err != nil` 블록) 안에, 또는 그 블록이 직접 부르는 같은 패키지 함수 안에 `codexadapter.IsDecisionBearing` 호출이 1개 이상 있다.
- (b2) 원소 타입이 `hook.EventType` 인 슬라이스·배열 리터럴, 키가 `hook.EventType` 인 맵 리터럴, `switch` 문의 case 식 전체, `||` 로 이은 `== hook.EventX` 비교 사슬 가운데 **어느 것도** 네 결정 이벤트 식별자(`EventPreToolUse`, `EventPermissionRequest`, `EventStop`, `EventUserPromptSubmit`) 중 2개 이상을 포함하지 않는다. 하위 명령 표(`hook.go:53-78`, 원소가 구조체인 리터럴)는 원소 타입이 `hook.EventType` 이 아니므로 대상이 아니다.
- 기준선: 이 트리와 `fabc33812` 에서 네 식별자가 `internal/cli` 비테스트 파일에 나타나는 곳은 하위 명령 표와 `runAgentHook` 의 단일 대입(`event = hook.EventPreToolUse`)뿐이다(2026-09-24, `grep -rnE 'hook\.Event(PreToolUse|PermissionRequest|Stop|UserPromptSubmit)\b' internal/cli --include='*.go' | grep -v _test.go` → 6행, `git grep` 로 `fabc33812` 에서 같은 6행). run 착수 시 t1099 흡수 후 트리에서 다시 잰다.

**(c) 필수 RED 변이 둘.** run 증거에 다음을 각각 한 번 적용해 빨개지는 것을 관측하고 되돌린다: (c1) 구현의 분류 호출 결과를 `false` 로 바꾸는 변이 → (a) 가 실패한다. (c2) `internal/cli` 비테스트 파일에 `switch event { case hook.EventPreToolUse: … case hook.EventStop: … }` 처럼 **case 를 여러 줄로 나눈** 중복 목록을 넣는 변이 → (b2) 가 실패한다. plan-audit 에서 codex 백엔드는 줄 단위 grep 이 이 변이를 놓친다는 것을 재현했다(`duplicate-switch-was-not-detected`) — (b) 가 grep 이 아니라 AST 여야 하는 이유다.

### AC-HSF-004 — 하네스 모드는 stdin 보다 먼저 판정된다

**Given** 착지된 구현, **When** `--harness bogus` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료이며, stderr 에 stdin 파싱 경고가 나오지 않는다(판정이 읽기보다 앞섰다는 관측). **When** `--harness codex` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 출력이 AC-HSF-002 의 Codex 형태다 — 수리 전 트리에서는 같은 입력이 `{}` 였다(Pre-flight 기준선).

### AC-HSF-005 — 관측 이벤트는 `6a3603274` 동작을 유지한다 (특성 기준)

**Given** 착지된 구현, **When** 관측 이벤트 하위 명령 22개 각각에 파손 4 형태를 넣고 두 하네스 모드(미지정, `codex`)로 실행하면(22 × 4 × 2 = 176 경우), **Then** 모든 경우에서 `RunE` 가 nil, 디스패치 0 회, stderr 에 `invalid stdin JSON` 경고 한 줄이며, stdout 은 이벤트별 현재 출력과 바이트 단위로 같다 — worktree-create·worktree-remove 는 **빈 stdout(0 바이트)**, 나머지 20개는 정확히 `{}`. 빈 stdout 인 이유는 `writeHookOutput` 이 두 worktree 이벤트에서 `input == nil` 이면 아무것도 쓰지 않기 때문이다(`internal/cli/hook.go:389-392`). **And** 같은 176 경우의 출력이 Pre-flight 기준선 커밋에서 잰 수리 전 출력과 같다. TaskCreated·Notification 은 HOI 게이트보다 파싱이 앞서므로 게이트 설정과 무관하게 같은 결과여야 한다.

### AC-HSF-006 — 탈출 장치 (정당한 활성화 경로)

**Given** plan.md Q1 에서 정한 **정당한** 활성화 경로로 탈출 장치를 켠 환경(테스트는 `t.Setenv` 로 비병렬, `CLAUDE_PROJECT_DIR` 는 `t.TempDir()`), **When** AC-HSF-001·002 의 결정 이벤트 16 × 2 = 32 경우를 실행하면, **Then** stdout 이 `{}`, exit 0, stderr 에 탈출 장치가 적용됐다는 한 줄이 있고, 영속 기록에 탈출 적용을 나타내는 행이 한 건씩 생긴다. **When** 탈출 장치가 꺼져 있거나 설정되지 않았으면, **Then** AC-HSF-001·002 결과가 그대로다(기본 꺼짐).

### AC-HSF-007 — 관측 가능성과 페이로드 비노출

**Given** `CLAUDE_PROJECT_DIR` 를 `t.TempDir()` 로 둔 환경, 탈출 장치 꺼짐, **When** 결정 이벤트 4개 × 파손 4 형태 × 두 하네스(32 경우)마다 fail-closed 를 한 번씩 발화시키면, **Then** 각 경우에 (a) stderr 에 이벤트 이름·하네스 모드·파싱 오류 원인·`fail-closed` 를 담은 한 줄이 있고, (b) `.moai/logs/codex-adapter.jsonl`(또는 t1099 착지 시점의 `DiagnosticSinkRel`)에 `RecordDiscards` 형식의 행이 한 건 추가되며, 그 행의 키는 핸들러 fault 키 `hook-fault` 와 다른 파싱 실패 전용 키이고, 길이 필드는 디스패처가 관측한 stdin 바이트 수와 같다(oversize 는 5 MiB 상한 `5 << 20` 바이트, 나머지는 입력 길이), (c) 그 경우에 쓰인 카나리가 stdout·stderr·기록 파일 **어디에도** 나타나지 않는다 — 네 형태 모두에 대한 원문 비노출. **And** 필수 RED 변이: 기록의 Reason 이나 stderr 줄에 페이로드 앞부분을 싣는 변이를 한 번 적용해 (c) 가 실패하는 것을 관측하고 되돌린다.

### AC-HSF-008 — Stop 의 거부는 `stop_hook_active` 에 의존하지 않는다

**Given** `{"hook_event_name":"Stop","stop_hook_active":true,` (닫히지 않은 JSON) 과 `{"hook_event_name":"Stop","stop_hook_active":false,` 두 입력, **When** `stop` 하위 명령을 두 하네스에서 실행하면, **Then** Claude 쪽 두 경우는 `decision == "block"` 이고, Codex 쪽 두 경우는 plan.md Q1 판정 분기에 따른다(포함이면 `decision == "block"`, A1 제외 분기면 `{}`). **And** Stop fail-closed 코드 경로에 호스트 차단 상한 의존과 그 근거 등급을 적은 `@MX:WARN` 이 있다(`grep -n '@MX:WARN' <해당 파일>` 로 확인). **And** progress.md 에 plan.md Q2 의 판정(Kickoff 기록)과, 측정을 Pre-flight 로 넘겼다면 그 측정 명령·출력이 기록돼 있다.

### AC-HSF-009 — t1099 전제 재확인이 기록돼 있다

**Given** run 착수 시점, **Then** progress.md §E.2 에 spec.md §F.1 의 `git diff fabc33812 <착지 병합 커밋> -- …` 명령과 그 결과 요약이 있고, (a) `DecisionBearingEvents()` 원소, (b) fatal_error 열 Outcome, (c) `writeCodexFailClosed` 서명과 기록 키, (d) 네 이벤트의 `Render` 형태, (e) `Discard` 필드와 stderr 미러 형식 각각에 대해 「동일」 또는 「변경 → spec 개정 커밋 SHA」가 적혀 있다.

### AC-HSF-010 — 모델이 쓸 수 있는 설정면으로는 탈출 장치가 켜지지 않는다

**Given** `CLAUDE_PROJECT_DIR` 를 `t.TempDir()` 로 둔 환경, **When** 탈출 장치의 활성화 값을 다음 면에 각각 하나씩 두고 — (i) `.claude/settings.json` 의 `env` 블록, (ii) `.claude/settings.local.json` 의 `env` 블록, (iii) `.moai/config/sections/` 아래 YAML 키 — (i)·(ii) 는 호스트가 그 `env` 블록을 훅 프로세스 환경으로 전파한 상황을 같은 키·값의 프로세스 환경 변수(`t.Setenv`)로 함께 모사하여, 결정 이벤트 × 파손 4 형태를 두 하네스에서 실행하면, **Then** 모든 경우에서 결과가 AC-HSF-001·002 와 같다(거부, 탈출 적용 기록 없음). **And** 필수 RED 변이: 설정면의 값을 활성화로 인정하는 변이를 한 번 적용해 이 기준이 실패하는 것을 관측하고 되돌린다. 메커니즘이 환경 변수가 아닌 것으로 정해지면(plan.md Q1 (ii)), (i)·(ii) 의 모사는 그 메커니즘의 식별자를 같은 면에 두는 것으로 바꾼다.

### AC-HSF-GATE — 품질 게이트

**Given** 착지된 구현, **When** `go test ./internal/cli/... -run 'Hook|Stdin|FailClosed' -count=1`, `go test ./internal/codexadapter/... -count=1`, `go vet ./internal/cli/... ./internal/codexadapter/...`, `golangci-lint run ./internal/cli/... ./internal/codexadapter/...` 를 실행하면, **Then** 모두 0 으로 끝나며, `-run` 선택자가 실제로 이 SPEC 의 테스트를 1개 이상 실행했다는 것을 `-v` 출력의 `=== RUN` 행 수로 함께 보인다(빈 선택은 통과가 아니다). 전 패키지 판정은 develop push 의 CI 가 내린다.

## §D 갱신이 필요한 기존 테스트 (이 트리 `60017eb83` 에서 측정)

| 테스트 | 위치 | 현재 단언 | 이 SPEC 이후 |
|---|---|---|---|
| `TestRunHookEvent_MalformedStdinGraceful` | `internal/cli/hook_protocol_fix_test.go:60` (입력 `:71`, 단언 `:95-124`) | `pre-tool` 에 `{"broken` → nil, 디스패치 0, stdout 이 유효한 JSON 하나. 주석(`:3-5`, `:57-59`)은 「default output」을 약속 | 단언 자체는 거부 JSON 도 유효한 JSON 이므로 **그대로 통과할 수 있다** — 바로 그래서 갱신이 필요하다. 이 테스트는 수리 전후를 가르지 못한다. 거부 형태를 단언하도록 바꾸고, 주석의 「default output」을 결정 이벤트에 한해 고친다. 관측 이벤트 쪽 의도는 AC-HSF-005 가 이어받는다 |
| `TestRunHookEvent_ReadInputError` | `internal/cli/coverage_test.go:63` | `post-tool` + mock `ReadInput` 오류 → nil, `HookProtocol.WriteOutput` 으로 기본 출력 | `post-tool` 은 관측 이벤트이므로 **변경 불필요**. 다만 결정 이벤트 경로는 `HookProtocol.WriteOutput` 을 거치지 않을 수 있으므로, 이 테스트를 결정 이벤트로 복제하지 말 것 |
| t1099 의 `TestHookFaultInjection` | `fabc33812:internal/cli/hook_fault_injection_test.go:137` | 핸들러 fault 경로(파싱 성공 후) | 영향 없음을 확인해야 한다 — 특히 plan.md M1 이 `writeCodexFailClosed` 서명을 넓히면, 기존 fault 경로 호출부가 같은 기록 키(`hook-fault`)와 Reason 을 유지하는지 이 테스트로 확인한다 |

`internal/hook` 패키지의 `ReadInput` 테스트는 프로토콜 계층을 바꾸지 않으므로 영향이 없다. 위 표 밖에서 결정 이벤트에 파손 stdin 을 넣고 `{}` 를 단언하는 테스트는 이 트리에서 찾지 못했다(`grep -rn 'broken\|invalid JSON\|not valid json' internal/cli/*_test.go` 결과 판독). run 착수 시 t1099 흡수 후 트리에서 같은 검색을 다시 한다.

## §E 완료 정의

- AC-HSF-001~010 과 GATE 가 명령·원문 출력과 함께 progress.md §E.2 에 기록됐다. 필수 RED 변이(AC-HSF-003(c1)·(c2), AC-HSF-007, AC-HSF-010)의 실패 출력도 함께 기록됐다.
- Pre-flight 기준선 재현 커밋이 수리 커밋보다 앞선다.
- plan.md Q2·Q1·Q3 판정이 기록됐고, Q1 에 탈출 장치 메커니즘이 포함됐으며, Q3 을 포함으로 판정했다면 D-NEW-1 개정이 먼저 착지했다.
- depth 형태의 대조군(깊이 9000 파싱 성공)이 기록됐다.
- `internal/hook` 패키지 diff 가 0 이다.
