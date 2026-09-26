# Acceptance — SPEC-HOOK-STOP-PARSE-CAP-001

이 문서는 검증 층이다. 요구사항(GEARS)은 spec.md §C 에 있고, 여기서는 각 요구를 Given-When-Then 으로 이진 판정 가능한 형태로 옮긴다.

## §0 공통 조건

- **파손 stdin**: `not json{` (9바이트). 선행 SPEC 의 파손 4 형태 중 형식 불량 하나를 대표로 쓴다. 셈은 파싱 실패의 종류와 무관하므로 4 형태 전부를 돌릴 필요는 없다 — AC-SPC-001 만 4 형태를 모두 돌려 종류 무관성을 확인한다.
- **유효 stdin**: `{"hook_event_name":"Stop","session_id":"s","cwd":"<임시 디렉터리>"}` 처럼 `ReadInput` 이 오류 없이 받는 최소 페이로드.
- **열쇠 주입**: 테스트는 부모 pid 공급원을 주입해 열쇠를 정한다. 같은 열쇠 = 같은 호스트 프로세스. 셈 기록 위치는 `t.TempDir()` 아래 프로젝트 루트로 돌린다.
- **Claude Stop 거부 바이트**: `Render(Stop, Lookup(HarnessClaude, Stop, DecisionFatalError).Outcome, <REQ-SPC-010 문자열>)` — 기대 사유는 테스트가 REQ-SPC-010 의 리터럴로 조립한다(구현이 넘긴 값을 되받지 않는다).
- **상한 해제 응답**: stdout 이 정확히 `{}` + 개행, `RunE` 가 nil.
- **분류**: 「차단」은 run-phase 완료를 막는 AC, 「회귀 가드」는 구현 뒤에만 의미 있는 판정이 나오는 AC 다(verification-completeness.md §2.1 의 판정 불가 처분).

---

## §0.1 요구사항 → 기준 대응 (두 칸: 수리 전 기대와 그것을 뒤집는 마일스톤)

| REQ | AC | 수리 전 트리(`e464fd5d0`)에서의 기대와 그 이유 | 무엇이 뒤집는가 |
|---|---|---|---|
| REQ-SPC-001 | AC-SPC-001 | RED — (d) 셈 기록이 존재하지 않는다(셈이 없는 코드). (a)~(c) 는 거부 사유를 빼면 초록이지만 (c) 의 사유 바이트가 새 문구라 RED | M1 + M3 |
| REQ-SPC-002 | AC-SPC-001 | 위와 같다 | M2 + M3 |
| REQ-SPC-003 | AC-SPC-002 | RED — 9번째 호출도 거부를 낸다(상태 없는 분기, 원장 L1 + 코드 판독) | M3 |
| REQ-SPC-004 | AC-SPC-005 | 공허한 초록 — 셈이 없어 세 환경이 같다(원장 L4). 변이로만 RED 관측 | M3 뒤 변이 |
| REQ-SPC-005 | AC-SPC-003 | RED — (b) 지울 셈 기록이 애초에 없어 판정 불가, (c) 의 사유 바이트가 새 문구가 아니다 | M3 |
| REQ-SPC-006 | AC-SPC-004 | RED — (b) 누적 9번째가 거부 | M3 |
| REQ-SPC-007 | AC-SPC-006 | RED — 셈 기록을 읽지 않으므로 대조(만료 전 → 해제)가 거부를 낸다 | M1 |
| REQ-SPC-008 | AC-SPC-007, AC-SPC-008 | AC-SPC-007 은 (c) 의 stderr 줄이 없어 RED, AC-SPC-008 (ii) 는 9번째 해제가 없어 RED | M1 + M3 |
| REQ-SPC-009 | AC-SPC-009 | 초록 — 보존 기준이다. 사유를 하네스 공통으로 바꾸는 변이가 (ii) 를 빨갛게 해야 한다 | 보존(M2 에서 깨지지 않을 것) |
| REQ-SPC-010 | AC-SPC-010 | RED — 개정 전 사유(원장 L1, L3) | M2 |
| REQ-SPC-011 | AC-SPC-011 | RED — 9·10번째 상한 해제 호출 자체가 없다 | M3 |
| REQ-SPC-012 | AC-SPC-012 | RED — 문서에 moai 자체 상한 서술이 없다 | M4 또는 sync |
| REQ-SPC-013 | AC-SPC-013 | RED — 원장 L2 | M4 |

AC-SPC-014 는 선행 SPEC 정정(F6)을, AC-SPC-015 는 M5 LIVE 재측정을 다룬다 — 둘 다 이 SPEC 의 REQ 에 직접 대응하지 않는다.

---

## §1 AC 목록

### AC-SPC-001 — N 이하에서는 차단이 유지된다 (REQ-SPC-001·002) — 차단

**Given** 셈 기록이 없는 열쇠 K, Claude 하네스, 호출 횟수를 세는 스파이 레지스트리, **When** 같은 K 로 `moai hook stop` 에 파손 stdin 을 8번 연달아 넣으면(파손 4 형태를 두 바퀴 돌려 8회), **Then** 8번 모두 (a) `RunE` 가 nil, (b) 디스패치 0회, (c) stdout 이 §0 의 Claude Stop 거부 바이트와 같고, (d) k번째 호출 뒤 셈 기록의 연속 횟수가 k 다.

### AC-SPC-002 — N 을 넘으면 상한 해제 응답이 나온다 (REQ-SPC-003) — 차단

**Given** AC-SPC-001 의 8회를 마친 K, **When** 파손 stdin 으로 9번째·10번째 호출을 하면, **Then** 두 호출 모두 (a) stdout 이 상한 해제 응답, (b) 디스패치 0회, (c) stderr 에 이벤트 `Stop`·하네스 `claude`·연속 횟수(9, 10)·N(8)을 담은 줄이 정확히 하나, (d) 영속 기록 한 건의 키가 `stdin-parse-fail-closed`·`stdin-parse-exempt`·`hook-fault` 셋 모두와 다르고 두 호출에서 같은 값이며, (e) 셈 기록의 연속 횟수가 각각 9, 10 이다.
**필수 RED 변이**: N 비교를 `>` 에서 `>=` 로 바꾼 구현은 8번째 호출에서 `{}` 를 내 AC-SPC-001(c) 가 빨개진다. N 판정을 없앤 구현(현재 코드)은 9번째에서 거부를 내 (a) 가 빨개진다.

### AC-SPC-003 — 파싱에 성공한 Stop 이 셈을 되돌린다 (REQ-SPC-005) — 차단

**Given** 파손 stdin 으로 8회를 마친 K, **When** 같은 K 로 유효 stdin 의 Stop 을 한 번 부른 뒤 파손 stdin 으로 한 번 더 부르면, **Then** (a) 유효 호출에서 디스패치가 1회 일어나고, (b) 유효 호출 직후 K 의 셈 기록 파일이 존재하지 않으며, (c) 이어진 파손 호출의 stdout 이 Claude Stop 거부 바이트이고 셈 기록의 연속 횟수가 1 이다.
**변형**: 10회(상한 해제 상태) 뒤 유효 Stop 한 번, 파손 한 번 → 마지막 호출이 거부를 낸다.

### AC-SPC-004 — 도구 사용은 셈을 되돌리지 않는다 (REQ-SPC-006) — 차단

**Given** 셈 기록이 없는 K, **When** 같은 K 로 다음 순서를 실행하면 — 파손 Stop 5회, 유효 PreToolUse 1회, 파손 PreToolUse 1회, 유효 PostToolUse 1회, 파손 PostToolUse 1회, 파손 Stop 4회 — **Then** (a) 끼어든 네 호출 전후로 셈 기록의 연속 횟수가 5 로 같고, (b) 마지막 파손 Stop 4회 가운데 앞의 3회(누적 6·7·8)는 거부, 4번째(누적 9)는 상한 해제 응답이며, (c) 파손 PreToolUse 는 선행 SPEC 대로 fail-closed 거부를 낸다(상한 없음).
**필수 RED 변이**: 모든 이벤트의 파싱 성공에서 셈을 지우는 구현은 (b) 의 9번째가 거부가 되어 빨개진다.

### AC-SPC-005 — 호스트 상한 값과 무관하다 (REQ-SPC-004) — 회귀 가드

**Given** 같은 열쇠 규칙, **When** AC-SPC-001·002 의 10회 순서를 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 가 (i) unset, (ii) `8`, (iii) `200` 인 세 환경에서 각각 새 열쇠로 실행하면(테스트는 병렬로 돌리지 않는다 — 환경 변수를 바꾸므로), **Then** 세 환경의 stdout 바이트 열 10개가 서로 같다. 그리고 정적 검사로, `internal/cli` 의 비테스트 Go 파일 중 셈·상한 판정을 담은 파일에 `EnvClaudeCodeStopHookBlockCap` 식별자와 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 리터럴이 없다.
**RED 경로**: 현재 코드는 상태가 없어 이 AC 가 공허하게 초록이다(원장 L4). 의미 있는 RED 는 run-phase 에서 변이로 관측한다 — 환경 변수가 200 일 때 N 을 200 으로 바꾸는 변이 구현이 (iii) 의 9번째에서 거부를 내 빨개져야 한다.

### AC-SPC-006 — 만료된 셈 기록은 없는 것으로 본다 (REQ-SPC-007) — 차단

**Given** K 의 셈 기록이 연속 횟수 8, 마지막 갱신 시각이 (만료 시간 + 1분) 전인 상태, 그리고 다른 열쇠 K2 의 셈 기록이 역시 만료된 상태, **When** K 로 파손 Stop 을 한 번 부르면, **Then** (a) stdout 이 거부(상한 해제가 아님), (b) K 의 셈 기록 연속 횟수가 1, (c) K2 의 셈 기록 파일이 지워져 있다. **대조**: 마지막 갱신이 (만료 시간 − 1분) 전이면 같은 호출이 상한 해제 응답을 낸다(횟수 9).

### AC-SPC-007 — 호스트 프로세스를 식별할 수 없으면 차단을 유지한다 (REQ-SPC-008) — 차단

**Given** 부모 pid 공급원이 1 을 돌려주는 설정, **When** 파손 Stop 을 12번 연달아 부르면, **Then** (a) 12번 모두 Claude Stop 거부 바이트, (b) 상태 영역에 셈 기록 파일이 하나도 생기지 않고, (c) 매 호출의 stderr 에 셈을 적용하지 못했다는 줄이 있다. 같은 확인을 부모 pid 0 으로 한 번 더 한다.

### AC-SPC-008 — 셈 기록을 쓰거나 해석할 수 없으면 차단을 유지한다 (REQ-SPC-008) — 차단

**Given** (i) 상태 영역 디렉터리 자리에 일반 파일이 있어 기록을 만들 수 없는 프로젝트 루트, (ii) K 의 셈 기록 내용이 `{` 인 프로젝트 루트, **When** 각각 파손 Stop 을 10번 부르면, **Then** (i) 에서는 10번 모두 거부이고 매 호출 stderr 에 적용 실패 줄이 있으며, (ii) 에서는 첫 호출이 거부이고 그 뒤 셈 기록이 해석 가능한 내용(연속 횟수 1)으로 바뀌어 있으며 9번째 호출에서 처음 상한 해제 응답이 나온다.

### AC-SPC-009 — Codex 하네스는 바뀌지 않는다 (REQ-SPC-009) — 차단

**Given** `--harness codex`, **When** (i) 파손 Stop 을 12번, (ii) 파손 PreToolUse 를 1번 부르면, **Then** (i) 12번 모두 stdout `{}` 에 면제 키 `stdin-parse-exempt` 기록(선행 SPEC AC-HSF-011 과 같은 결과)이고 상태 영역에 셈 기록 파일이 생기지 않으며, (ii) stdout 바이트가 `TranslateCodex(PreToolUse, DecisionFatalError, "fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)")` 와 같다 — Codex 사유는 개정 전 문자열 그대로다.

### AC-SPC-010 — Claude 하네스 사유 문구 (REQ-SPC-010) — 차단

**Given** `--harness` 미지정, **When** 네 결정 이벤트(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)와 `moai hook agent x-validation` 을 파손 stdin 으로 각각 부르면(Stop 은 새 열쇠로 1회), **Then** 다섯 출력 모두 (a) 그 이벤트의 거부 필드에 담긴 사유가 REQ-SPC-010 의 리터럴과 바이트 단위로 같고, (b) 사유에 `disableAllHooks` 와 `moai update` 가 없으며, (c) 사유가 `fail-closed: ` 로 시작하고 `(.moai/docs/hook-stdin-fail-closed.md)` 로 끝난다.
**필수 RED 변이**: 개정 전 사유를 내는 구현(현재 코드)은 (a) 에서 빨개진다 — 원장 L1 의 출력이 그 실패의 모양이다.

### AC-SPC-011 — 페이로드 원문이 새지 않는다 (REQ-SPC-011) — 차단

**Given** 파손 stdin 에 카나리 문자열 `CANARY_7f3a` 를 넣은 페이로드(`{"tool_input":"CANARY_7f3a"` 처럼 잘린 JSON), **When** 같은 열쇠로 10번 불러 9·10번째에서 상한 해제 응답이 나오게 하면, **Then** 10번의 stdout·stderr·영속 기록 파일·셈 기록 파일 어디에도 `CANARY_7f3a` 가 없다.

### AC-SPC-012 — 운영자 문서 (REQ-SPC-012) — 차단

**Given** 템플릿 원본 `internal/template/templates/.moai/docs/hook-stdin-fail-closed.md`, **When** 문서를 읽으면, **Then** (a) Claude Stop 의 moai 자체 상한 N=8, (b) 상한을 넘은 뒤 무의견 응답이 이어진다는 것, (c) 파싱 성공과 만료만이 셈을 되돌린다는 것, (d) 셈 기록 위치, (e) 호스트 상한 값과 무관하다는 것을 각각 서술한 문장이 있다. 로컬 사본 `.moai/docs/hook-stdin-fail-closed.md` 도 같은 내용이다. 판정은 리뷰어가 (a)~(e) 를 항목별로 표시하는 방식이며, 기계 검사로는 `grep -c "N=8\|8회" <두 파일>` 이 각각 1 이상인지만 본다.

### AC-SPC-013 — 코드 주석이 측정과 일치한다 (REQ-SPC-013) — 차단

**Given** run-phase 완료 트리, **When** 다음을 실행하면, **Then** 기대대로다.
- `grep -c "has not been measured" internal/codexadapter/stop_cap.go` → `0`, exit 1 (현재 `1`, exit 0 — 원장 L2)
- `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN` 줄이 moai 자체 상한(N=8)과 200 주입 조건을 함께 언급한다(리뷰어 판정).

### AC-SPC-014 — 선행 SPEC 정정 (F6) — 차단, plan-phase 에서 충족

**Given** 이 카드의 plan-phase 커밋 트리, **When** `.moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` 를 읽으면, **Then** (a) 프론트매터 `version: "0.4.3"` 과 `status: completed`, (b) HISTORY 에 0.4.3 행, (c) REQ-HSF-010 에 「SPEC-HOOK-STOP-PARSE-CAP-001 REQ-SPC-010 이 Claude 하네스 문구를 대체」 표시, (d) §F.2 「Stop 루프 — Claude」 행이 t1230·t1272 측정(무도구 조건, 200 주입 조건)을 인용, (e) §D 의 「Claude 측정 안 함」 문장이 정정됨. plan.md Q8 도 측정 결과로 정정됨. 기계 검사: `grep -c "SPEC-HOOK-STOP-PARSE-CAP-001" .moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` → 3 이상.

### AC-SPC-015 — LIVE 재측정 (M5) — 회귀 가드, 운영자 예산

**Given** plan.md M5 의 상한 선언과 `exec` 형 래퍼, **When** A(`CAP=200`)·B(unset) 두 팔을 각 1회 실행하면, **Then** 두 팔 모두 Stop 호출 수가 9 이상 12 이하이고 `result` 가 `error_max_turns` 가 아니다. 관측 항목(판정 아님): 모델의 훅·설정 편집 시도 횟수. 실행하지 못하면 Gap 으로 보고하며 run-phase 완료를 막지 않는다.

---

## §D 선행 SPEC 테스트 갱신 대상

- `internal/cli/hook_stdin_failclosed_test.go` 의 기대 사유(현재 `:815` 의 `want := "fail-closed: hook stdin could not be parsed as JSON (" + ...`)는 Claude 하네스 기대값을 REQ-SPC-010 리터럴로 바꾸고, Codex 하네스 기대값은 그대로 둔다. 선행 SPEC AC-HSF-001(d)·(e4) 와 AC-HSF-013 Claude 모드 행이 이 갱신의 영향을 받는다. 선행 SPEC 의 acceptance.md 문구는 고치지 않는다(spec.md §D).

---

## §E 증거 원장 (plan-phase, 트리 `e464fd5d0`)

L1 — 현재 동작, 파싱 실패 Stop 1회. 격리 프로젝트는 세션 scratchpad 의 `proj/`, 바이너리는 이 워크트리 HEAD 로 빌드한 `moai-t1272`.

```
명령: CLAUDE_PROJECT_DIR=<scratchpad>/proj <scratchpad>/moai-t1272 hook stop < <scratchpad>/bad.txt
       (칸반·팩토리·상한 변수 unset 뒤 실행, bad.txt 내용 "not json{")
stdout: {"decision":"block","reason":"fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)"}
stderr: moai hook Stop: invalid stdin JSON (hook: invalid JSON input: invalid character 'o' in literal null (expecting 'u')) on Stop, harness claude; answered fail-closed
        codex-adapter: dropped "stdin-parse-fail-closed" on Stop (9 bytes): stdin parse failure on a decision-bearing event answered with a fail-closed deny
exit: 0
```

이 명령은 리다이렉트를 쓰므로 단일 호출 형식이 아니다 — 재현 가능한 명령 기록이지 release-blocking RED 셀이 아니다. 9번째 호출의 동작은 이 판독과 코드 판독(`answerStdinParseFailure` 에 상태가 없음)으로 추론했고 9회를 연달아 실행하지는 않았다. AC-SPC-002·004 의 RED 는 run-phase TDD RED 단계에서 테스트 출력 원문으로 관측한다.

L2 — AC-SPC-013 RED-now.

```
명령: grep -c "has not been measured" internal/codexadapter/stop_cap.go
stdout: 1
exit: 0
```

L3 — AC-SPC-010 의 새 문구 부재.

```
명령: grep -c "notify" internal/cli/hook_stdin_failclosed.go
stdout: 0
exit: 1
```

L4 — AC-SPC-005 정적 부분이 현재 공허하게 초록인 근거.

```
명령: grep -c "StopHookBlockCap" internal/cli/hook_stdin_failclosed.go
stdout: 0
exit: 1
```

---

## §F 완료 정의 (Definition of Done)

- 「차단」 AC(001~004, 006~014) 가 run-phase 트리에서 모두 PASS 이고, 각 AC 의 명령·출력 원문·트리 SHA 가 progress.md §E.2 에 있다.
- AC-SPC-002·004·010 의 필수 RED 변이가 관측됐다(변이를 넣은 실행의 실패 출력 원문).
- `go test ./internal/cli/... ./internal/codexadapter/...` 가 통과하고, `golangci-lint run` 에 새 지적이 없다. 전체 스위트 판정은 develop CI 가 한다.
- `GOOS=windows GOARCH=amd64 go build ./...` 가 통과한다(부모 pid 판독이 플랫폼 태그를 요구하는지 확인).
- 회귀 가드 AC(005, 015)는 결과 또는 Gap 이 보고돼 있다.
