---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Acceptance — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.1.0"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Acceptance — SPEC-HOOK-STDIN-FAILCLOSED-001

모든 기준은 명령 출력이나 파일 판독으로 관측 가능하다. 각 PASS 는 실행한 명령과 원문 출력을, 주장하는 실행에서, 측정 대상 트리에 대해 남긴다(`verification-claim-integrity.md` §2).

공통 입력 — 「파손 3 형태」:

| 이름 | 내용 |
|---|---|
| malformed | `{"broken` |
| truncated | 유효한 결정 이벤트 페이로드의 앞 절반(닫는 중괄호 없음) |
| oversize | `{"hook_event_name":"<이벤트>","tool_input":{"content":"` + 6 MiB 의 `a` + `"}}` — 5 MiB 에서 잘려 파싱 실패가 되는 경로 |

「결정 이벤트」는 `codexadapter.DecisionBearingEvents()` 의 원소, 「관측 이벤트」는 `runHookEvent` 에 연결된 하위 명령 중 그 밖의 것(착지 시점 기준 22개)이다. 목록을 테스트 안에 손으로 적지 않고 두 원천(`DecisionBearingEvents()`, `hook.go` 의 하위 명령 표)에서 계산한다.

## §1 요구사항 → 기준 대응

| REQ | AC |
|---|---|
| REQ-HSF-001 | AC-HSF-001, AC-HSF-002 |
| REQ-HSF-002 | AC-HSF-003 |
| REQ-HSF-003 | AC-HSF-002 |
| REQ-HSF-004 | AC-HSF-001 |
| REQ-HSF-005 | AC-HSF-004 |
| REQ-HSF-006 | AC-HSF-005 |
| REQ-HSF-007 | AC-HSF-006 |
| REQ-HSF-008 | AC-HSF-007 |
| REQ-HSF-009 | AC-HSF-008 |
| REQ-HSF-010 | AC-HSF-001, AC-HSF-002 |
| REQ-HSF-011 | AC-HSF-007 |

## §2 기준

### AC-HSF-001 — Claude 하네스: 결정 이벤트 × 파손 3 형태 → 거부

**Given** `--harness` 미지정, 실제 `hook.NewProtocol()`, 호출 횟수를 세는 스파이 레지스트리, **When** 각 결정 이벤트의 하위 명령에 파손 3 형태를 각각 stdin 으로 넣고 `RunE` 를 실행하면, **Then** (a) `RunE` 가 nil 을 돌려준다(exit 0), (b) 디스패치 횟수가 0 이다, (c) stdout 이 JSON 값 정확히 하나이며 `{}` 도 빈 문자열도 아니다, (d) 그 바이트가 `Render(ev, Lookup(HarnessClaude, ev, DecisionFatalError).Outcome, <사유>)` 와 같은 형태다 — PreToolUse 는 `hookSpecificOutput.permissionDecision == "deny"`, PermissionRequest 는 `hookSpecificOutput.decision.behavior == "deny"`, Stop·UserPromptSubmit 은 `decision == "block"` — (e) 사유 필드가 비어 있지 않다. 4 × 3 = 12 경우 전부.

### AC-HSF-002 — Codex 하네스: 결정 이벤트 × 파손 3 형태 → 거부

**Given** `--harness codex`, 나머지는 AC-HSF-001 과 같은 조건, **When** 같은 12 경우를 실행하면, **Then** (a)~(c) 가 성립하고, (d) stdout 바이트가 `codexadapter.TranslateCodex(ev, DecisionFatalError, <파싱 오류>)` 의 출력과 같으며, (e) 사유에 `fail-closed` 가 들어 있다. PermissionRequest 는 Codex EventTable 에서 미적응(U)이지만 하위 명령은 존재하므로 같은 기준을 적용한다.

### AC-HSF-003 — 결정 이벤트 목록은 하나뿐이다

**Given** 착지된 구현, **When** 테스트가 `runHookEvent` 에 연결된 모든 하위 명령(26개)에 malformed 를 넣고 「출력이 `{}` 가 아닌 이벤트 집합」을 두 하네스 각각에서 관측하면, **Then** 그 집합이 `codexadapter.DecisionBearingEvents()` 와 원소 단위로 같다. **And** 정적 검사 `grep -rnE 'EventPermissionRequest.*EventStop|EventStop.*EventUserPromptSubmit' internal/cli --include='*.go' | grep -v _test.go` 가 0 행이고, `grep -n 'IsDecisionBearing' internal/cli/hook*.go` 가 파싱 실패 분기 안에서 1 행 이상을 낸다. **And** 변이 확인: 테스트 안에서 `DecisionBearingEvents()` 를 바꿀 수 없으므로, 대신 구현의 판정 호출을 `false` 로 바꾸는 변이를 run 증거에서 한 번 적용해 이 테스트가 빨개지는 것을 관측하고 되돌린다.

### AC-HSF-004 — 하네스 모드는 stdin 보다 먼저 판정된다

**Given** 착지된 구현, **When** `--harness bogus` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료이며, stderr 에 stdin 파싱 경고가 나오지 않는다(판정이 읽기보다 앞섰다는 관측). **When** `--harness codex` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 출력이 AC-HSF-002 의 Codex 형태다 — 수리 전 트리에서는 같은 입력이 `{}` 였다(Pre-flight 기준선).

### AC-HSF-005 — 관측 이벤트는 `6a3603274` 동작을 유지한다 (특성 테스트)

**Given** 착지된 구현, **When** 관측 이벤트 하위 명령 22개 각각에 파손 3 형태를 넣고 두 하네스 모드(미지정, `codex`)로 실행하면, **Then** 모든 경우에서 `RunE` 가 nil, 디스패치 0 회, stdout 이 정확히 `{}`, stderr 에 `invalid stdin JSON` 경고 한 줄이다. TaskCreated·Notification 은 HOI 게이트보다 파싱이 앞서므로 게이트 설정과 무관하게 같은 결과여야 한다.

### AC-HSF-006 — 탈출 장치

**Given** 탈출 장치가 켜진 환경(plan Q1 판정 형태, 테스트는 `t.Setenv` 로 비병렬), **When** AC-HSF-001·002 의 결정 이벤트 12 × 2 경우를 실행하면, **Then** stdout 이 `{}`, exit 0, stderr 에 탈출 장치가 적용됐다는 한 줄이 있고, 영속 기록에 탈출 적용을 나타내는 행이 한 건씩 생긴다. **When** 탈출 장치가 꺼져 있거나 설정되지 않았으면, **Then** AC-HSF-001·002 결과가 그대로다(기본 꺼짐).

### AC-HSF-007 — 관측 가능성과 페이로드 비노출

**Given** `CLAUDE_PROJECT_DIR` 를 `t.TempDir()` 로 둔 환경, **When** oversize 입력으로 결정 이벤트의 fail-closed 가 한 번 발화하면, **Then** (a) stderr 에 이벤트 이름·하네스 모드·파싱 오류 원인·`fail-closed` 를 담은 한 줄이 있고, (b) `.moai/logs/codex-adapter.jsonl`(또는 t1099 착지 시점의 `DiagnosticSinkRel`)에 한 행이 추가되며 그 행은 `RecordDiscards` 가 쓴 형식이다, (c) stdout·stderr·기록 어디에도 입력의 `aaaa` 연속(64 바이트 이상)이 나타나지 않는다 — 원문 비노출.

### AC-HSF-008 — Stop 의 거부는 `stop_hook_active` 에 의존하지 않는다

**Given** `{"hook_event_name":"Stop","stop_hook_active":true,` (닫히지 않은 JSON) 과 `{"hook_event_name":"Stop","stop_hook_active":false,` 두 입력, **When** `stop` 하위 명령을 두 하네스에서 실행하면, **Then** 네 경우 모두 `decision == "block"` 이다. **And** Stop fail-closed 코드 경로에 호스트 차단 상한 의존을 적은 `@MX:WARN` 이 있다(`grep -n '@MX:WARN' <해당 파일>` 로 확인). **And** progress.md §E.2 에 plan Q2(Codex 상한) 측정 결과가 기록돼 있다.

### AC-HSF-009 — t1099 전제 재확인이 기록돼 있다

**Given** run 착수 시점, **Then** progress.md §E.2 에 spec.md §F.1 의 `git diff fabc33812 <착지 병합 커밋> -- …` 명령과 그 결과 요약이 있고, (a) `DecisionBearingEvents()` 원소, (b) fatal_error 열 Outcome, (c) `writeCodexFailClosed` 서명, (d) 네 이벤트의 `Render` 형태 각각에 대해 「동일」 또는 「변경 → spec 개정 커밋 SHA」가 적혀 있다.

### AC-HSF-GATE — 품질 게이트

**Given** 착지된 구현, **When** `go test ./internal/cli/... -run 'Hook|Stdin|FailClosed' -count=1`, `go test ./internal/codexadapter/... -count=1`, `go vet ./internal/cli/... ./internal/codexadapter/...`, `golangci-lint run ./internal/cli/... ./internal/codexadapter/...` 를 실행하면, **Then** 모두 0 으로 끝난다. 전 패키지 판정은 develop push 의 CI 가 내린다.

## §D 갱신이 필요한 기존 테스트 (이 트리 `60017eb83` 에서 측정)

| 테스트 | 위치 | 현재 단언 | 이 SPEC 이후 |
|---|---|---|---|
| `TestRunHookEvent_MalformedStdinGraceful` | `internal/cli/hook_protocol_fix_test.go:60` (입력 `:71`, 단언 `:95-124`) | `pre-tool` 에 `{"broken` → nil, 디스패치 0, stdout 이 유효한 JSON 하나. 주석(`:3-5`, `:58-60`)은 「default output」을 약속 | 단언 자체는 거부 JSON 도 유효한 JSON 이므로 **그대로 통과할 수 있다** — 바로 그래서 갱신이 필요하다. 이 테스트는 수리 전후를 가르지 못한다. 거부 형태를 단언하도록 바꾸고, 주석의 「default output」을 결정 이벤트에 한해 고친다. 관측 이벤트 쪽 의도는 AC-HSF-005 가 이어받는다 |
| `TestRunHookEvent_ReadInputError` | `internal/cli/coverage_test.go:63` | `post-tool` + mock `ReadInput` 오류 → nil, `HookProtocol.WriteOutput` 으로 기본 출력 | `post-tool` 은 관측 이벤트이므로 **변경 불필요**. 다만 결정 이벤트 경로는 `HookProtocol.WriteOutput` 을 거치지 않을 수 있으므로, 이 테스트를 결정 이벤트로 복제하지 말 것 |
| t1099 의 `TestHookFaultInjection` | `fabc33812:internal/cli/hook_fault_injection_test.go:137` | 핸들러 fault 경로(파싱 성공 후) | 영향 없음 — 파싱 실패 경로는 이 SPEC 의 새 테스트가 덮는다. 병합 후 함께 통과하는지만 확인 |

`internal/hook` 패키지의 `ReadInput` 테스트는 프로토콜 계층을 바꾸지 않으므로 영향이 없다. 위 표 밖에서 결정 이벤트에 파손 stdin 을 넣고 `{}` 를 단언하는 테스트는 이 트리에서 찾지 못했다(`grep -rn 'broken\|invalid JSON\|not valid json' internal/cli/*_test.go` 결과 판독). run 착수 시 t1099 흡수 후 트리에서 같은 검색을 다시 한다.

## §E 완료 정의

- AC-HSF-001~009 와 GATE 가 명령·원문 출력과 함께 progress.md §E.2 에 기록됐다.
- Pre-flight 기준선 재현 커밋이 수리 커밋보다 앞선다.
- plan Q1·Q3 판정이 기록됐고, Q3 을 포함으로 판정했다면 D-NEW-1 개정이 먼저 착지했다.
- `internal/hook` 패키지 diff 가 0 이다.
