# Acceptance — SPEC-HOOK-STOP-PARSE-CAP-001

이 문서는 검증 층이다. 요구사항(GEARS)은 spec.md §C 에 있고, 여기서는 각 요구를 Given-When-Then 으로 이진 판정 가능한 형태로 옮긴다.

## §0 공통 조건

- **파손 stdin**: `not json{` (9바이트). 선행 SPEC 의 파손 4 형태 중 형식 불량 하나를 대표로 쓴다. 셈은 파싱 실패의 종류와 무관하므로 4 형태 전부를 돌릴 필요는 없다 — AC-SPC-001 만 4 형태를 모두 돌려 종류 무관성을 확인한다.
- **유효 stdin**: `{"hook_event_name":"Stop","session_id":"s","cwd":"<임시 디렉터리>"}` 처럼 `ReadInput` 이 오류 없이 받는 최소 페이로드.
- **열쇠 K**: 따로 적지 않으면 세션 열쇠다 — 테스트가 `t.Setenv("CLAUDE_CODE_SESSION_ID", <K>)` 로 정한다(그 테스트는 병렬로 돌리지 않는다). 새 열쇠 = 아직 쓰지 않은 세션 id. 프로세스 열쇠가 필요한 AC(AC-SPC-006 대조, AC-SPC-007)는 변수를 빈 값으로 두고 대체 해석의 조상 표를 주입한다(plan.md §C 6). 셈 기록 위치는 `t.TempDir()` 아래 프로젝트 루트로 돌린다.
- **기록 파일 이름 형식**: `^[0-9a-f]{64}\.json$` (plan.md §B.2).
- **Claude Stop 거부 바이트**: `Render(Stop, Lookup(HarnessClaude, Stop, DecisionFatalError).Outcome, <REQ-SPC-010 문자열>)` — 기대 사유는 테스트가 REQ-SPC-010 의 리터럴로 조립한다(구현이 넘긴 값을 되받지 않는다).
- **Codex 기대 사유**: `fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)` — 테스트가 이 리터럴을 Claude 기대 사유와 **별개로** 적는다(§D).
- **상한 해제 응답**: stdout 이 정확히 `{}` + 개행, `RunE` 가 nil.
- **분류**: 「차단」은 run-phase 완료를 막는 AC, 「회귀 가드」는 구현 뒤에만 의미 있는 판정이 나오는 AC 다(verification-completeness.md §2.1 의 판정 불가 처분).

---

## §0.1 요구사항 → 기준 대응 (두 칸: 수리 전 기대와 그것을 뒤집는 마일스톤)

| REQ | AC | 수리 전 트리(`e464fd5d0`, 코드 동일 `6c0da65d5`)에서의 기대와 그 이유 | 무엇이 뒤집는가 |
|---|---|---|---|
| REQ-SPC-001 | AC-SPC-001 | RED — (d) 셈 기록이 존재하지 않는다(셈이 없는 코드). (a)~(c) 는 거부 사유를 빼면 초록이지만 (c) 의 사유 바이트가 새 문구라 RED | M1 + M3 |
| REQ-SPC-002 | AC-SPC-001 | 위와 같다 | M2 + M3 |
| REQ-SPC-003 | AC-SPC-002 | RED — 9번째 호출도 거부를 낸다(상태 없는 분기, 원장 L1 + 코드 판독) | M3 |
| REQ-SPC-004 | AC-SPC-005 | 공허한 초록 — 셈이 없어 세 환경이 같다(원장 L4). 변이로만 RED 관측 | M3 뒤 변이 |
| REQ-SPC-005 | AC-SPC-003 | RED — (b) 지울 셈 기록이 애초에 없어 판정 불가, (c) 의 사유 바이트가 새 문구가 아니다, (e) 의 stderr 줄이 없다 | M3 |
| REQ-SPC-006 | AC-SPC-004 | RED — (b) 누적 9번째가 거부 | M3 |
| REQ-SPC-007 | AC-SPC-006 | RED — 셈 기록을 읽지 않으므로 대조(만료 전 → 해제)가 거부를 낸다 | M1 |
| REQ-SPC-008 | AC-SPC-007 (iv), AC-SPC-008 | AC-SPC-007 (iv) 는 stderr 줄이 없어 RED, AC-SPC-008 (ii) 는 9번째 해제가 없어 RED | M1 + M3 |
| REQ-SPC-009 | AC-SPC-009 | 초록 — 보존 기준이다. 사유를 하네스 공통으로 바꾸는 변이가 (ii) 를 빨갛게 해야 한다 | 보존(M2 에서 깨지지 않을 것) |
| REQ-SPC-010 | AC-SPC-010 | RED — 개정 전 사유(원장 L1, L3) | M2 |
| REQ-SPC-011 | AC-SPC-011 | RED — 9·10번째 상한 해제 호출 자체가 없다 | M3 |
| REQ-SPC-012 | AC-SPC-012 | RED — 문서에 moai 자체 상한 서술이 없다 | M4 또는 sync |
| REQ-SPC-013 | AC-SPC-013 | RED — 원장 L2 | M4 |
| REQ-SPC-014 | AC-SPC-007 (i)~(iii), AC-SPC-016 | RED — 이 경로는 세션 id 변수도 세션 소유자 해석도 읽지 않는다(원장 L5). 셈 기록 자체가 없으므로 (i)~(iii) 의 「같은 기록」 판정이 서지 않는다 | M1 (AC-SPC-016 은 M1 + M5a) |
| REQ-SPC-015 | AC-SPC-007 (v), AC-SPC-006 (d) | RED — 기록 파일이 만들어지지 않으므로 이름 형식 판정이 서지 않는다 | M1 |

AC-SPC-014 는 선행 SPEC 정정(F6)을, AC-SPC-015 는 M5b·M5c LIVE 측정을 다룬다 — 둘 다 이 SPEC 의 REQ 에 직접 대응하지 않는다. AC-SPC-016 은 REQ-SPC-014 의 1순위가 실경로에서 실제로 작동하는지를 재는 측정 AC 다.

---

## §1 AC 목록

### AC-SPC-001 — N 이하에서는 차단이 유지된다 (REQ-SPC-001·002) — 차단

**Given** 셈 기록이 없는 열쇠 K, Claude 하네스, 호출 횟수를 세는 스파이 레지스트리, **When** 같은 K 로 `moai hook stop` 에 파손 stdin 을 8번 연달아 넣으면(파손 4 형태를 두 바퀴 돌려 8회), **Then** 8번 모두 (a) `RunE` 가 nil, (b) 디스패치 0회, (c) stdout 이 §0 의 Claude Stop 거부 바이트와 같고, (d) k번째 호출 뒤 셈 기록의 연속 횟수가 k 다.

### AC-SPC-002 — N 을 넘으면 상한 해제 응답이 나온다 (REQ-SPC-003) — 차단

**Given** AC-SPC-001 의 8회를 마친 K, **When** 파손 stdin 으로 9번째·10번째 호출을 하면, **Then** 두 호출 모두 (a) stdout 이 상한 해제 응답, (b) 디스패치 0회, (c) stderr 에 이벤트 `Stop`·하네스 `claude`·연속 횟수(9, 10)·N(8)·열쇠 종류(세션)를 담은 줄이 정확히 하나, (d) 영속 기록 한 건의 키가 `stdin-parse-fail-closed`·`stdin-parse-exempt`·`hook-fault` 셋 모두와 다르고 두 호출에서 같은 값이며, (e) 셈 기록의 연속 횟수가 각각 9, 10 이다.
**필수 RED 변이**: N 비교를 `>` 에서 `>=` 로 바꾼 구현은 8번째 호출에서 `{}` 를 내 AC-SPC-001(c) 가 빨개진다. N 판정을 없앤 구현(현재 코드)은 9번째에서 거부를 내 (a) 가 빨개진다.

### AC-SPC-003 — 파싱에 성공한 Stop 이 셈을 되돌린다 (REQ-SPC-005) — 차단

**Given** 파손 stdin 으로 8회를 마친 K, **When** 같은 K 로 유효 stdin 의 Stop 을 한 번 부른 뒤 파손 stdin 으로 한 번 더 부르면, **Then** (a) 유효 호출에서 디스패치가 1회 일어나고, (b) 유효 호출 직후 K 의 셈 기록 파일이 존재하지 않으며, (c) 이어진 파손 호출의 stdout 이 Claude Stop 거부 바이트이고 셈 기록의 연속 횟수가 1 이다.
**변형 1**: 10회(상한 해제 상태) 뒤 유효 Stop 한 번, 파손 한 번 → 마지막 호출이 거부를 낸다.
**변형 2 (없는 기록)**: 상태 영역 디렉터리조차 없는 프로젝트 루트에서 새 열쇠로 유효 Stop 을 한 번 부르면 (d) 호출 뒤 `.moai/state/stop-parse-cap/` 디렉터리가 존재하지 않는다.
**변형 3 (삭제 실패)**: K 의 기록 삭제가 실패하도록 주입하면(삭제 이음매, 또는 상태 영역 디렉터리의 쓰기 권한 제거 — 후자는 root 로 도는 러너에서 건너뛰고 그 사실을 보고한다) (e) 유효 호출의 디스패치와 stdout 이 삭제 성공 때와 같고, stderr 에 삭제 실패 사유를 담은 줄이 정확히 하나 있다.

### AC-SPC-004 — 도구 사용은 셈을 되돌리지 않는다 (REQ-SPC-006) — 차단

**Given** 셈 기록이 없는 K, **When** 같은 K 로 다음 순서를 실행하면 — 파손 Stop 5회, 유효 PreToolUse 1회, 파손 PreToolUse 1회, 유효 PostToolUse 1회, 파손 PostToolUse 1회, 파손 Stop 4회 — **Then** (a) 끼어든 네 호출 전후로 셈 기록의 연속 횟수가 5 로 같고, (b) 마지막 파손 Stop 4회 가운데 앞의 3회(누적 6·7·8)는 거부, 4번째(누적 9)는 상한 해제 응답이며, (c) 파손 PreToolUse 는 선행 SPEC 대로 fail-closed 거부를 낸다(상한 없음).
**필수 RED 변이**: 모든 이벤트의 파싱 성공에서 셈을 지우는 구현은 (b) 의 9번째가 거부가 되어 빨개진다.

### AC-SPC-005 — 호스트 상한 값과 무관하다 (REQ-SPC-004) — 회귀 가드

**Given** 같은 열쇠 규칙, **When** AC-SPC-001·002 의 10회 순서를 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 가 (i) unset, (ii) `8`, (iii) `200` 인 세 환경에서 각각 새 열쇠로 실행하면(테스트는 병렬로 돌리지 않는다 — 환경 변수를 바꾸므로), **Then** 세 환경의 stdout 바이트 열 10개가 서로 같다. 그리고 정적 검사로, 정확히 두 파일 `internal/cli/hook_stdin_failclosed.go` 와 `internal/cli/hook_stop_parse_cap.go`(plan.md M1 이 정한 셈 기록 구현 파일)에 `EnvClaudeCodeStopHookBlockCap` 식별자와 `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP` 리터럴이 없다 — `grep -c` 가 두 파일 모두 0. 이 두 파일 밖(예: 그 식별자를 정당하게 쓰는 `internal/cli/launcher_blockcap_infinite.go`)은 검사 대상이 아니다. 둘째 파일이 없으면 이 정적 검사는 판정 불가로 보고한다(0 을 통과로 읽지 않는다).
**RED 경로**: 현재 코드는 상태가 없어 이 AC 가 공허하게 초록이다(원장 L4). 의미 있는 RED 는 run-phase 에서 변이로 관측한다 — 환경 변수가 200 일 때 N 을 200 으로 바꾸는 변이 구현이 (iii) 의 9번째에서 거부를 내 빨개져야 한다.

### AC-SPC-006 — 만료된 셈 기록과 정리 범위 (REQ-SPC-007) — 차단

만료 시간은 60분이다(spec.md §B).

**Given** K 의 셈 기록이 연속 횟수 8, 마지막 갱신 시각이 61분 전인 상태, 그리고 다른 열쇠 K2 의 셈 기록이 역시 61분 전 갱신인 상태, **When** K 로 파손 Stop 을 한 번 부르면, **Then** (a) stdout 이 거부(상한 해제가 아님), (b) K 의 셈 기록 연속 횟수가 1, (c) K2 의 셈 기록 파일이 지워져 있다.
**대조 — 해제 상태 이어받기(spec.md §F 첫 행을 고정)**: 프로세스 열쇠 P 의 기록이 연속 횟수 9(해제 상태), 마지막 갱신 59분 전이면, 같은 P 로 온 파손 Stop 이 **첫 호출에서** 상한 해제 응답을 낸다(횟수 10). 새 호스트가 같은 pid 를 받아 이 기록을 이어받는 경우의 동작이 바로 이것이며, 이 대조는 그 동작을 진술된 잔여 위험으로 고정한다.
**(d) 정리 범위**: 상태 영역에 다음을 함께 두고 K 로 파손 Stop 을 한 번 부르면 — ① 이름 형식에 맞고 61분 전 갱신인 K2 기록(일반 파일), ② 형식에 맞지 않는 이름의 파일 `123.json`·`notes.txt`(내용은 61분 전 갱신 기록), ③ 형식에 맞는 이름의 심볼릭 링크로, 대상이 상태 영역 **밖**의 61분 전 갱신 기록인 것, ④ 형식에 맞는 이름의 일반 파일로 내용이 `{` 인 것 — 호출 뒤 ①만 지워지고 ②·③(링크 자체와 그 대상 파일 모두)·④ 는 그대로이며 ③ 의 대상 파일 바이트가 바뀌지 않았다.
**(e) 링크로 바뀐 상태 영역**: 상태 영역 디렉터리 자리가 다른 디렉터리를 가리키는 심볼릭 링크이고 그 대상 디렉터리에 형식에 맞는 61분 전 기록이 있을 때, 파손 Stop 을 부르면 대상 디렉터리의 기록이 지워지지 않고 바뀌지도 않으며, 응답은 REQ-SPC-008 의 거부 + stderr 적용 실패 줄이다.

### AC-SPC-007 — 셈 열쇠 해석 (REQ-SPC-014·015·008) — 차단

- **(i) 세션 열쇠 우선**: `CLAUDE_CODE_SESSION_ID=sess-A` 로 파손 Stop 9회 → 9번째가 상한 해제 응답이고 stderr 줄의 열쇠 종류가 세션이다. 이어 `sess-B` 로 파손 Stop 1회 → 거부이며 `sess-B` 의 기록 연속 횟수가 1 이다(열쇠별로 셈이 갈린다). 같은 세션 id 인 채로 대체 해석의 조상 표를 서로 다른 호스트로 바꿔도 같은 기록을 쓴다.
- **(ii) 대체 경로로 떨어지는 조건**: 변수가 (a) 없음, (b) 빈 문자열, (c) 공백만(`"   "`)인 세 경우 각각에서, 조상 표가 호스트 100 을 가리키게 주입하고 파손 Stop 9회 → 9번째가 상한 해제이고 stderr 줄의 열쇠 종류가 프로세스다.
- **(iii) 합성 프로세스 트리 — 끼어든 셸을 건너뛴다**: 변수를 비우고, 조상 표를 두 모양으로 차례로 주입한다 — 트리 1: moai 400 → bash 300 → bash 200 → 호스트 100(명령 이름 `claude`), 트리 2: moai 401 → bash 201 → 호스트 100. 트리 1 로 파손 Stop 한 번, 트리 2 로 파손 Stop 한 번을 부르면 두 호출이 **같은 셈 기록**을 쓰며 그 연속 횟수가 2 다. 이 시험은 해석을 상수로 대신하지 않고 실제 조상 탐색을 합성 표 위에서 돌린다(plan.md §C 6).
  **필수 RED 변이**: 대체 경로를 원시 부모 pid 로 바꾼 구현은 트리 1 에서 300, 트리 2 에서 201 을 열쇠로 삼아 기록이 둘로 갈리고 각 연속 횟수가 1 이 되어 빨개진다.
- **(iv) 열쇠 없음 → 차단 유지**: 변수를 비우고 조상 표가 해석 실패(`(0, false)`)를 내게 주입한 뒤 파손 Stop 을 12번 부르면, 12번 모두 Claude Stop 거부 바이트이고, 상태 영역에 셈 기록 파일이 하나도 생기지 않으며, 매 호출의 stderr 에 셈을 적용하지 못했다는 줄이 있다.
- **(v) 파일 이름과 경로 가둠**: 세션 id `../../escape`, `a/b`, `a\b`, `..`, 제어 문자 `\x01`·`\x7f` 가 든 값, 그리고 NUL 이 든 값(환경 변수로는 NUL 을 넣을 수 없으므로 열쇠→파일 이름 함수에 직접 넣는다)마다 — 만들어지는 기록 파일 이름이 `^[0-9a-f]{64}\.json$` 에 맞고, 상태 영역의 부모 디렉터리와 `t.TempDir()` 루트에 새 항목이 생기지 않으며, 기록·stderr 어디에도 세션 id 원문이 없다. 그리고 세션 id `"123"` 과 프로세스 열쇠 123 이 서로 다른 파일 이름을 낸다.

### AC-SPC-008 — 셈 기록을 쓰거나 해석할 수 없으면 차단을 유지한다 (REQ-SPC-008) — 차단

**Given** (i) 상태 영역 디렉터리 자리에 일반 파일이 있어 기록을 만들 수 없는 프로젝트 루트, (ii) K 의 셈 기록 내용이 `{` 인 프로젝트 루트, (iii) K 의 기록 자리가 다른 파일을 가리키는 심볼릭 링크이거나 디렉터리인 프로젝트 루트, **When** 각각 파손 Stop 을 10번 부르면, **Then** (i) 에서는 10번 모두 거부이고 매 호출 stderr 에 적용 실패 줄이 있으며, (ii) 에서는 첫 호출이 거부이고 그 뒤 셈 기록이 해석 가능한 내용(연속 횟수 1)으로 바뀌어 있으며 9번째 호출에서 처음 상한 해제 응답이 나오고, (iii) 에서는 10번 모두 거부이고 매 호출 stderr 에 적용 실패 줄이 있으며 링크의 대상 파일 바이트가 바뀌지 않았다.

### AC-SPC-009 — Codex 하네스는 바뀌지 않는다 (REQ-SPC-009) — 차단

**Given** `--harness codex`, **When** (i) 파손 Stop 을 12번, (ii) 파손 stdin 으로 Codex 의 나머지 결정 이벤트 세 개(PreToolUse·PermissionRequest·UserPromptSubmit)를 각각 1번, (iii) `moai hook agent x-validation --harness codex` 를 파손 stdin 으로 1번 부르면, **Then** (i) 12번 모두 stdout `{}` 에 면제 키 `stdin-parse-exempt` 기록(선행 SPEC AC-HSF-011 과 같은 결과)이고 상태 영역에 셈 기록 파일이 생기지 않으며, (ii) 각 이벤트 ev 의 stdout 바이트가 `TranslateCodex(ev, DecisionFatalError, <§0 의 Codex 기대 사유 리터럴>)` 와 같고, (iii) stdout 바이트가 그 action 이 대응하는 이벤트에 대한 같은 식의 결과와 같다 — Codex 사유는 개정 전 문자열 그대로이며, (ii)·(iii) 의 어떤 호출도 셈 기록 파일을 만들지 않는다.
**필수 RED 변이**: 사유를 하네스 공통으로 새 문구로 바꾼 구현은 (ii)·(iii) 의 네 호출 모두에서 빨개진다. 기대값은 테스트 리터럴이므로 구현의 사유 함수를 바꿔도 함께 움직이지 않는다(§D).

### AC-SPC-010 — Claude 하네스 사유 문구 (REQ-SPC-010) — 차단

**Given** `--harness` 미지정, **When** 네 결정 이벤트(PreToolUse·PermissionRequest·Stop·UserPromptSubmit)와 `moai hook agent x-validation` 을 파손 stdin 으로 각각 부르면(Stop 은 새 열쇠로 1회), **Then** 다섯 출력 모두 (a) 그 이벤트의 거부 필드에 담긴 사유가 REQ-SPC-010 의 리터럴과 바이트 단위로 같고, (b) 사유에 `disableAllHooks` 와 `moai update` 가 없으며, (c) 사유가 `fail-closed: ` 로 시작하고 `(.moai/docs/hook-stdin-fail-closed.md)` 로 끝난다.
**필수 RED 변이**: 개정 전 사유를 내는 구현(현재 코드)은 (a) 에서 빨개진다 — 원장 L1 의 출력이 그 실패의 모양이다.

### AC-SPC-011 — 페이로드 원문이 새지 않는다 (REQ-SPC-011) — 차단

**Given** 파손 stdin 에 카나리 문자열 `CANARY_7f3a` 를 넣은 페이로드(`{"tool_input":"CANARY_7f3a"` 처럼 잘린 JSON), **When** 같은 열쇠로 10번 불러 9·10번째에서 상한 해제 응답이 나오게 하면, **Then** 10번의 stdout·stderr·영속 기록 파일·셈 기록 파일 어디에도 `CANARY_7f3a` 가 없다.

### AC-SPC-012 — 운영자 문서 (REQ-SPC-012) — 차단

**Given** 템플릿 원본 `internal/template/templates/.moai/docs/hook-stdin-fail-closed.md`, **When** 문서를 읽으면, **Then** (a) Claude Stop 의 moai 자체 상한 N=8, (b) 상한을 넘은 뒤 무의견 응답이 이어진다는 것, (c) 파싱 성공과 만료(60분)만이 셈을 되돌린다는 것, (d) 셈 열쇠의 선택 규칙과 셈 기록 위치, (e) 호스트 상한 값과 무관하다는 것을 각각 서술한 문장이 있다. 로컬 사본 `.moai/docs/hook-stdin-fail-closed.md` 도 같은 내용이다. 판정은 리뷰어가 (a)~(e) 를 항목별로 표시하는 방식이며, 기계 검사로는 `grep -cE "N=8|8회" <두 파일>` 이 각각 1 이상인지만 본다.

### AC-SPC-013 — 코드 주석이 측정과 일치한다 (REQ-SPC-013) — 차단

**Given** run-phase 완료 트리, **When** 다음을 실행하면, **Then** 기대대로다.
- `grep -c "has not been measured" internal/codexadapter/stop_cap.go` → `0`, exit 1 (현재 `1`, exit 0 — 원장 L2)
- `internal/cli/hook_stdin_failclosed.go` 의 Stop 루프 `@MX:WARN` 줄이 moai 자체 상한(N=8)과 200 주입 조건을 함께 언급한다(리뷰어 판정).

### AC-SPC-014 — 선행 SPEC 정정 (F6) — 차단, plan-phase 에서 충족

**Given** 이 카드의 plan-phase 커밋 트리, **When** `.moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` 를 읽으면, **Then** (a) 프론트매터 `version: "0.4.3"` 과 `status: completed`, (b) HISTORY 에 0.4.3 행, (c) REQ-HSF-010 에 「SPEC-HOOK-STOP-PARSE-CAP-001 REQ-SPC-010 이 Claude 하네스 문구를 대체」 표시, (d) §F.2 「Stop 루프 — Claude」 행이 t1230·t1272 측정(무도구 조건, 200 주입 조건)을 인용, (e) §D 의 「Claude 측정 안 함」 문장이 정정됨. plan.md Q8 과 M1 의 Stop 경로 `@MX:WARN` 지시문(「미측정 독트린」)도 측정 결과로 정정됨. 기계 검사: `grep -c "SPEC-HOOK-STOP-PARSE-CAP-001" .moai/specs/SPEC-HOOK-STDIN-FAILCLOSED-001/spec.md` → 3 이상.

### AC-SPC-015 — LIVE 재측정 (M5b·M5c) — 회귀 가드, 운영자 예산

**Given** plan.md M5 의 상한 선언과 배포 래퍼 사슬(설치된 `settings.json` Stop 항목 + `handle-stop.sh`, `PATH` 맨 앞의 `moai` 대리 스크립트가 워크트리 빌드 바이너리로 `exec`), **When** (i) A(`CAP=200`)·(ii) B(unset) 두 팔을 각 1회, 대리 스크립트가 파손 stdin 을 넣도록 실행하면, **Then** 두 팔 모두 Stop 호출 수가 9 이상 12 이하이고 `result` 가 `error_max_turns` 가 아니다. (iii) M5c 대화형 세션에서 `/clear` 전후 두 Stop 의 측정 로그를 비교해 `CLAUDE_CODE_SESSION_ID` 가 바뀌었는지를 기록한다(관측 — 판정 아님. 바뀌면 의도대로 셈이 이월되지 않고, 바뀌지 않으면 spec.md §F 「`/clear` 뒤의 이월」이 확인된 잔여 위험이 된다). 관측 항목(판정 아님): 모델의 훅·설정 편집 시도 횟수. 실행하지 못하면 Gap 으로 보고하며 run-phase 완료를 막지 않는다.

### AC-SPC-016 — 훅 환경의 세션 id 존재를 실경로에서 잰다 (REQ-SPC-014) — 차단

**Given** plan.md M5a 의 상한 선언(1회, `--max-turns 2`, `timeout -k 10 120`)과 AC-SPC-015 와 같은 배포 래퍼 사슬(대리 스크립트는 유효 stdin 을 그대로 통과), **When** 워크트리 빌드 바이너리로 `claude -p` 를 1회 실행해 Stop 이 한 번 이상 불리면, **Then** (a) 대리 스크립트의 측정 로그에 `hook stop` 호출마다 `CLAUDE_CODE_SESSION_ID` 의 상태(`set-nonempty` / `set-empty` / `unset`)가 한 줄씩 남아 있고 — 이 로그는 `exec` 직전의 같은 프로세스에서 쓰였으므로 `moai hook stop` 프로세스의 환경이다 — (b) 같은 실행의 훅 stderr 로그에 moai 가 남긴 열쇠 종류가, 상태가 `set-nonempty` 이면 세션, 아니면 프로세스다. 두 상태 중 어느 쪽이 관측되든 (a)·(b) 가 서면 이 AC 는 통과이며, 관측값은 progress.md §E.2 에 원문으로 남기고 spec.md §A.4 의 「관측하지 않았다」 문장을 D-NEW-1 로 갱신할 입력이 된다.
**RED 경로**: 현재 코드는 열쇠 종류를 stderr 에 남기지 않으므로 (b) 가 서지 않는다(원장 L5). M1 이 열쇠 종류 출력을 넣고, M5a 가 관측한다. 실행하지 못하면 Gap 이 아니라 blocker 로 보고한다.

---

## §D 선행 SPEC 테스트 갱신 대상

Claude 하네스 사유를 바이트로 비교하는 곳은 셋이다. 모두 Claude 기대값을 REQ-SPC-010 리터럴로 바꾸고, Codex 기대값은 그대로 둔다.

- `internal/cli/hook_stdin_failclosed_test.go` — 현재 공유 도우미 `expectedFailClosedReason()`(`:202`)가 Claude 기대값(`expectedClaudeFailClosed`, `:212` Render)과 Codex 기대값(`expectedCodexFailClosed`, `:221` TranslateCodex)을 **함께** 만든다. 이 도우미를 하네스별 리터럴 두 개(Claude 용, Codex 용)로 나눈다 — 하나만 새 문구로 바꾸면 Codex 기대값도 따라 움직여 기존 스위트가 Codex 변경을 잡지 못한다. `:815` 의 `want := "fail-closed: hook stdin could not be parsed as JSON (" + ...` 는 조립 검사이므로 하네스를 명시해 갈라 적는다. 선행 SPEC AC-HSF-001(d)·(e4) 와 AC-HSF-013 Claude 모드 행이 이 갱신의 영향을 받는다.
- `internal/cli/hook_protocol_fix_test.go:127` — PreToolUse 거부 사유를 `expectedFailClosedReason()` 과 비교한다. Claude 리터럴과 비교하도록 바꾼다.
- `internal/cli/misc_coverage_test.go:382` — `moai hook agent` 의 PreToolUse 거부 사유를 같은 도우미와 비교한다. Claude 리터럴과 비교하도록 바꾼다.

선행 SPEC 의 acceptance.md 문구는 고치지 않는다(spec.md §D).

---

## §E 증거 원장 (plan-phase)

L1~L4 는 트리 `e464fd5d0`, L5 는 트리 `6c0da65d5`(`e464fd5d0` 이후 문서 커밋만 있어 코드 동일)에서 쟀다.

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

L5 — AC-SPC-007·016 RED-now: 이 경로는 세션 id 변수도 세션 소유자 해석도 읽지 않는다.

```
명령: grep -c "EnvClaudeCodeSessionID" internal/cli/hook_stdin_failclosed.go
stdout: 0
exit: 1

명령: grep -c "ResolveOwnerPID" internal/cli/hook_stdin_failclosed.go
stdout: 0
exit: 1
```

---

## §F 완료 정의 (Definition of Done)

- 「차단」 AC(001~004, 006~014, 016) 가 run-phase 트리에서 모두 PASS 이고, 각 AC 의 명령·출력 원문·트리 SHA 가 progress.md §E.2 에 있다.
- AC-SPC-002·004·007·009·010 의 필수 RED 변이가 관측됐다(변이를 넣은 실행의 실패 출력 원문).
- `go test ./internal/cli/... ./internal/codexadapter/... ./internal/session/... ./internal/config/...` 가 통과하고, `golangci-lint run` 에 새 지적이 없다. 전체 스위트 판정은 develop CI 가 한다.
- `GOOS=windows GOARCH=amd64 go build ./...` 가 통과한다(세션 소유자 해석의 Windows 경로가 함께 컴파일되는지 확인).
- 회귀 가드 AC(005, 015)는 결과 또는 Gap 이 보고돼 있다.
