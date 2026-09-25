---
id: SPEC-HOOK-STDIN-FAILCLOSED-001
title: "Acceptance — 훅 stdin 파싱 실패의 결정 이벤트 fail-closed"
version: "0.4.0"
created: 2026-09-24
author: manager-spec (card t1152)
---

# Acceptance — SPEC-HOOK-STDIN-FAILCLOSED-001

모든 기준은 명령 출력이나 파일 판독으로 관측 가능하다. 각 PASS 는 실행한 명령과 원문 출력을, 주장하는 실행에서, 측정 대상 트리에 대해 남긴다(`verification-claim-integrity.md` §2).

0.2.0 개정: plan-audit 1회차(`.moai/reports/t1152/plan-audit.md`)의 D1(관측 이벤트 출력의 이벤트별 정정, AC-003 관측 술어), D4(depth 형태 추가), D5(AC-003 정적 검사를 AST 검사로 교체), D6(형태별 카나리), D7(사유 필수 요소 단언), D3(탈출 장치 출처 제약 음성 기준 AC-HSF-010 신설), D11(줄 번호) 을 반영했다.

0.3.0 개정: 운영자 판정 Q1(A1 — Codex Stop 면제, 탈출 장치 메커니즘 (i))·Q3(`runAgentHook` 포함)과 Q2 측정을 반영했다. AC-HSF-002·003·006·007·008·010 을 개정하고, AC-HSF-011(Codex Stop 면제), AC-HSF-012(`runAgentHook` 결정 매핑), AC-HSF-013(`runAgentHook` 관측 매핑)을 신설했다.

0.3.1 개정: plan-audit 2회차(`.moai/reports/t1152/plan-audit-iter2.md`)의 N1~N3 을 반영했다. N1 — AC-HSF-006 (p1)·(p2) 가 정적 판정만 검증하며 「추가 후 제거」 잔여 위험(spec.md §F.2)을 인증하지 않는다는 한정을 붙였다. N2 — AC-HSF-012 에 유효한 stdin + 잘못된 `--harness` 의 거부를 새 계약으로 단언하고, §1 의 REQ-HSF-005 행에 AC-HSF-012 를 더했다. N3 — AC-HSF-003(b1) 을 `runAgentHook` 의 파싱 실패 분기까지 넓히고 필수 RED 변이 (c4) 를 추가했다. AC 수는 그대로다(13 + GATE).

0.3.2 개정: plan-audit 3회차(`.moai/reports/t1152/plan-audit-iter3.md`)의 N10~N12 를 반영했다. N10 — AC-HSF-013 에 관측 매핑 action(`x-verification`·`x-completion`)의 잘못된 `--harness` 거부를 단언하고 필수 RED 변이를 더했으며, §1 의 REQ-HSF-005 행에 AC-HSF-013 을 더했다. N11 — AC-HSF-003(b1) 이 `IsDecisionBearing` 호출의 존재만이 아니라 그 결과가 fail-closed 분기를 고르는 조건식에 쓰이는지를 보게 하고, 인자 조건을 AST 로 판정되는 부정 조건으로 좁혔으며, 필수 RED 변이 (c5) 를 추가했다. N12 — AC-HSF-004·012·013 의 「stdin 보다 먼저」 증거에 스파이 `HookProtocol` 의 `ReadInput` 호출 0 회를 직접 관측으로 더했다(stderr 경고 부재는 대리 관측으로 남긴다). §E 의 Pre-flight 5 완료 조건을 새 판정 규칙에 맞췄다. AC 수는 그대로다(13 + GATE).

0.3.3 개정: plan-audit 4회차(`.moai/reports/t1152/plan-audit-iter4.md`)의 N13·N14 를 반영해 §E 의 Pre-flight 5 완료 조건에 삭제 뒤 기록 요건과 두 경로 확정 전제를 더했다. AC 수는 그대로다(13 + GATE).

0.3.4 개정: Implementation Kickoff Approval 을 받은 사실(2026-09-24, t1152 레인 운영자 답)에 맞춰 §E 의 Kickoff 조건 문장을 고쳤다. AC 수는 그대로다(13 + GATE).

0.4.0 개정: Pre-flight 5 결과 「유지한다」(`.moai/reports/t1152/preflight5.md`)와 운영자 재판정(2026-09-25 — 탈출 장치 없이 결정 이벤트는 언제나 거부)을 반영했다. AC-HSF-006·010 을 삭제했다 — 번호는 감사 보고서 인용을 살리려고 「삭제됨」으로 남긴다. AC-HSF-001·007·011·012 의 「탈출 장치 꺼짐」 전제와 탈출 적용 키 언급을 걷어냈다. AC-HSF-001(e4) 를 「카나리·복구 절차 문구 부재」로 고쳤고, REQ-HSF-001 의 실행 시점 스위치 금지 절을 잡는 AST 검사 AC-HSF-003(b3) 와 필수 RED 변이 (c6) 을 더했다. §E 의 Pre-flight 5 완료 조건을 「0.4.0 에서 종료」로 바꿨다. 살아 있는 AC 는 11 + GATE 다.

## §0 공통 입력 — 「파손 4 형태」와 카나리

각 형태는 **서로 다른 카나리 문자열**을 품는다. 카나리는 테스트 실행 시 `CANARY-<FORM>-` 뒤에 `crypto/rand` 로 만든 16진 문자 64개를 붙여 만든다(형태마다 70바이트 이상, 서로 다름). 64바이트 이상으로 잡는 이유: `encoding/json` 오류 메시지는 입력의 한 글자나 숫자 리터럴 정도를 인용할 수 있어(깊이 초과 오류는 `invalid character '['` 로 관측됨), 짧은 표식은 합법적 오류 인용과 구별되지 않는다.

| 이름 | 내용 | 파싱 실패 원인 |
|---|---|---|
| malformed | `{"broken":"<CANARY-MALFORMED>` | 닫히지 않은 문자열 |
| truncated | `tool_input.note` 에 `<CANARY-TRUNCATED>` 를 담은 유효한 결정 이벤트 페이로드를, 카나리 **뒤**에서 잘라 앞부분만 넣는다(닫는 중괄호 없음) | 입력 조기 종료 |
| oversize | `{"hook_event_name":"<이벤트>","tool_input":{"note":"<CANARY-OVERSIZE>","content":"` + 6 MiB 의 `a` + `"}}` | 5 MiB 상한(`internal/hook/protocol.go:44`)에서 잘려 파싱 실패 |
| depth | `{"hook_event_name":"<이벤트>","tool_input":{"note":"<CANARY-DEPTH>","x":` + `[` 10001개 + `]` 10001개 + `}}` (약 20 KB) | `encoding/json` 중첩 깊이 상한 초과 — `normalizeHookInput` 첫 단계의 `json.Unmarshal`(`internal/hook/normalize.go:85-88`)에서 발생. 대조: 깊이 9000 은 같은 타입으로 파싱된다(spec.md §A.3 측정) |

depth 형태는 run 착수 시 대조군을 함께 잰다: 같은 구성의 깊이 9000 입력은 `ReadInput` 이 오류를 돌려주지 **않아야** 한다. 대조군이 실패하면 depth 형태의 결과를 판정에 쓰지 않고 원인을 먼저 보고한다.

「fail-closed 쌍」은 (하네스, 결정 이벤트) 가운데 REQ-HSF-013 의 면제 술어가 거짓인 쌍이다(현재 Claude 4 + Codex 3 = 7). 테스트는 이 집합을 `DecisionBearingEvents()` 와 그 술어로 계산하며, 이벤트 이름을 손으로 적지 않는다. 「면제 쌍」은 술어가 참인 (Codex, Stop) 하나다.

「결정 이벤트」는 `codexadapter.DecisionBearingEvents()` 의 원소, 「관측 이벤트」는 `runHookEvent` 에 연결된 하위 명령 중 그 밖의 것(착지 시점 기준 22개)이다. 목록을 테스트 안에 손으로 적지 않고 두 원천(`DecisionBearingEvents()`, `hook.go` 의 하위 명령 표)에서 계산한다.

「거부 필드」는 spec.md §B.4 의 열이다: PreToolUse `hookSpecificOutput.permissionDecision == "deny"`, PermissionRequest `hookSpecificOutput.decision.behavior == "deny"`, Stop·UserPromptSubmit 최상위 `decision == "block"`.

## §1 요구사항 → 기준 대응과 수리 전 트리의 기대

두 칸 원칙(`verification-completeness.md` §2): 각 기준에 수리 전 트리에서 무엇이 보여야 하는지(RED 이유)와, 무엇이 그것을 뒤집는지를 함께 적는다. 수리 전 관측은 plan.md Pre-flight 4 의 기준선 커밋이 담는다.

| REQ | AC | 수리 전 트리(Pre-flight 기준선)에서의 기대 | 무엇이 뒤집는가 |
|---|---|---|---|
| REQ-HSF-001 | AC-HSF-001, AC-HSF-002, AC-HSF-003(b3)·(c6) | RED — stdout 이 `{}` 라 (c) 가 실패. (b3) 은 수리 전에도 녹색일 수 있으므로(스위치가 없다) 변이 (c6) 으로 RED 를 따로 관측 | M1 |
| REQ-HSF-002 | AC-HSF-003 | RED — 관측 집합이 공집합이라 (a) 가 실패. AST 검사는 수리 전에도 녹색일 수 있으므로 중복 `switch` 변이로 RED 를 따로 관측 | M1 |
| REQ-HSF-003 | AC-HSF-002 | RED — (c) 실패 | M1 |
| REQ-HSF-004 | AC-HSF-001 | RED — (c) 실패 | M1 |
| REQ-HSF-005 | AC-HSF-004, AC-HSF-012, AC-HSF-013 | RED — `pre-tool` 은 `--harness bogus` 에서 stdin 경고가 먼저 나온다(AC-004). `moai hook agent` 는 `--harness` 를 읽지 않아 유효한 stdin + `--harness bogus` 가 결정 매핑 action(AC-012)과 관측 매핑 action(AC-013) 모두에서 exit 0 으로 성공한다 — 둘 다 신규 거부, Pre-flight 4(c) | M3, M1b |
| REQ-HSF-006 | AC-HSF-005 | 녹색(특성 기준) — RED 는 관측 이벤트 출력을 바꾸는 변이로 관측 | 뒤집히지 않아야 함 |
| REQ-HSF-007 | — (AC-HSF-006·010 삭제됨) | 삭제됨 (0.4.0, Pre-flight 5 결과) | — |
| REQ-HSF-008 | AC-HSF-007 | RED — stderr 한 줄·기록 모두 없다 | M2 |
| REQ-HSF-009 | AC-HSF-008 | RED — `decision == "block"` 이 아니다 | M1 |
| REQ-HSF-010 | AC-HSF-001(e), AC-HSF-002(e) | RED — 사유 자체가 없다 | M1 |
| REQ-HSF-011 | AC-HSF-007(c), AC-HSF-011(d) | 수리 전에는 기록이 없어 공허하게 녹색이다 — 원문을 기록에 싣는 변이로 RED 를 관측 | M2 |
| REQ-HSF-012 | AC-HSF-011 | RED — stdout 은 이미 `{}` 지만 면제 stderr 줄·면제 기록이 없어 (c) 가 실패 | M1 |
| REQ-HSF-013 | AC-HSF-003(a)·(c3), AC-HSF-011 | RED — 술어가 없다. 술어가 늘 거짓이 되는 변이는 AC-011 을, 디스패처가 술어를 건너뛰는 변이는 AC-003(a) 를 빨갛게 만든다 | M1 |
| REQ-HSF-014 | AC-HSF-012, AC-HSF-003(b1)·(c4) | RED — `foo-validation` 등이 `{}` 를 낸다(plan-audit 의 codex 백엔드 재현, Pre-flight 4(c) 에서 다시 잰다) | M1b |
| REQ-HSF-015 | AC-HSF-013 | 녹색(특성 기준) — 관측 매핑 action 의 출력을 바꾸는 변이로 RED 를 관측 | 뒤집히지 않아야 함 |
| REQ-HSF-016 | — (AC-HSF-010 삭제됨) | 삭제됨 (0.4.0, Pre-flight 5 결과) | — |

## §2 기준

### AC-HSF-001 — Claude 하네스: 결정 이벤트 × 파손 4 형태 → 거부

**Given** `--harness` 미지정, 실제 `hook.NewProtocol()`, 호출 횟수를 세는 스파이 레지스트리, **When** 각 결정 이벤트의 하위 명령에 파손 4 형태를 각각 stdin 으로 넣고 `RunE` 를 실행하면, **Then** (a) `RunE` 가 nil 을 돌려준다(exit 0), (b) 디스패치 횟수가 0 이다, (c) stdout 이 JSON 값 정확히 하나이며 `{}` 도 빈 문자열도 아니다, (d) 그 바이트가 `Render(ev, Lookup(HarnessClaude, ev, DecisionFatalError).Outcome, <사유>)` 와 같고 그 이벤트의 거부 필드(§0)를 갖는다, (e) 사유 필드가 비어 있지 않으며 (e1) `fail-closed`, (e2) 원인이 stdin 파싱 실패라는 고정 문구(run-phase 상수), (e3) 운영자 문서 식별자(run-phase 상수)를 담고, (e4) 어떤 카나리도, 복구 절차 문구(`disableAllHooks`)도 담지 않는다(REQ-HSF-010 — 복구 절차는 운영자 문서에만). 4 × 4 = 16 경우 전부. (e2)·(e3) 은 테스트가 리터럴이 아니라 구현의 상수를 참조해 단언한다.

### AC-HSF-002 — Codex 하네스: 결정 이벤트 × 파손 4 형태 → 거부

**Given** `--harness codex`, 나머지는 AC-HSF-001 과 같은 조건, **When** Codex 의 fail-closed 쌍(PreToolUse, PermissionRequest, UserPromptSubmit — §0 에서 술어로 계산) × 파손 4 형태 = 12 경우를 실행하면, **Then** (a)~(c) 가 성립하고, (d) stdout 바이트가 `codexadapter.TranslateCodex(ev, DecisionFatalError, <구현이 넘긴 원인 문자열>)` 의 출력과 같고 거부 필드를 가지며, (e) AC-HSF-001(e1)~(e4) 가 같은 방식으로 성립한다. PermissionRequest 는 Codex EventTable 에서 미적응(U)이지만 하위 명령은 존재하므로 같은 기준을 적용한다. Codex 의 Stop 은 면제 쌍이므로 이 기준이 아니라 AC-HSF-011 을 따른다(운영자 판정 A1, plan.md §B.1 Q1).

### AC-HSF-003 — 결정 이벤트 목록은 하나뿐이다

**(a) 동작 등가.** **Given** 착지된 구현, **When** 테스트가 `runHookEvent` 에 연결된 모든 하위 명령(26개)에 malformed 를 넣고, 두 하네스 각각에서 「stdout 이 JSON 으로 파싱되고 §0 의 거부 필드 가운데 하나를 갖는 이벤트 집합」을 관측하면, **Then** Claude 모드에서 관측된 집합은 `codexadapter.DecisionBearingEvents()` 와 원소 단위로 같고, Codex 모드에서 관측된 집합은 `{ev ∈ DecisionBearingEvents() | 면제 술어(Codex, ev) 가 거짓}` 과 원소 단위로 같다 — 기대 집합은 두 원천(집합 함수와 술어)으로 계산하며 이벤트 이름을 테스트에 적지 않는다. 빈 stdout(worktree-create·worktree-remove)과 `{}` 는 거부 필드가 없으므로 집합에 들지 않는다. 술어가 모든 쌍에 거짓을 돌려주도록 퇴화하면 이 기준은 공허하게 통과할 수 있으므로, (Codex, Stop) 의 실제 출력은 AC-HSF-011 이 따로 고정한다.

**(b) AST 검사.** `internal/cli` 의 테스트가 아닌 Go 파일 전부를 `go/parser` 로 읽는 테스트(또는 같은 규칙의 `ast-grep` 규칙)가 다음을 단언한다:
- (b1) 두 진입점 `runHookEvent` 와 `runAgentHook` **각각**에서, `ReadInput` 호출 결과의 오류 분기(`if err != nil` 블록) 안에, 또는 그 블록이 직접 부르는 같은 패키지 함수 안에(두 진입점이 공유하는 파싱 실패 처리 함수라면 그 함수 안에) `codexadapter.IsDecisionBearing` 호출이 1개 이상 있고, 그 호출의 인자는 `hook.EventX` 형태의 선택자 리터럴이 아닌 식이다(의도는 「그 진입점이 결정한 이벤트 값 — `runAgentHook` 에서는 action 매핑의 결과」이지만, 그 의미 조건은 AST 로 판정되지 않으므로 검사하는 것은 이 부정 조건이다). **그리고** 그 호출의 결과가 — 직접, 또는 그 결과를 담은 지역 변수를 거쳐 — fail-closed 출력을 쓰는 분기를 고르는 `if` 조건식이나 `switch` 의 태그·case 식에 쓰인다. 결과를 `_` 에 대입하거나, 어떤 분기 조건식에도 쓰이지 않는 변수에 담는 형태는 이 조건을 채우지 못한다. 같은 범위 안에 네 결정 이벤트 식별자(`EventPreToolUse`, `EventPermissionRequest`, `EventStop`, `EventUserPromptSubmit`)와의 `==`·`!=` 비교나 그 식별자를 case 식으로 둔 `switch` 가 없다. 두 진입점 가운데 하나라도 조건을 채우지 못하면 (b1) 은 실패다.
- (b3) (0.4.0, REQ-HSF-001 의 금지 절) (b1) 과 같은 범위 — 두 진입점의 `ReadInput` 오류 분기 블록과, 그 블록이 직접 부르는 `internal/cli` 패키지 함수의 본문 — 안에 `os.Getenv`·`os.LookupEnv`·`os.Environ` 호출과 파일 판독 호출(`os.ReadFile`·`os.Open`)이 없다. 파싱 실패의 출력이 환경이나 설정 파일에 따라 갈리는 경로가 이 범위에 없다는 뜻이다. 한계: 다른 패키지 함수 안으로 옮긴 스위치는 이 검사가 보지 못한다 — 그 잔여는 리뷰 몫이며, spec.md §F.1 의 3(c) 가 run 착수 시 t1099 흡수 트리에서 이 범위가 이미 조건을 어기는지를 먼저 잰다.
- (b2) 원소 타입이 `hook.EventType` 인 슬라이스·배열 리터럴, 키가 `hook.EventType` 인 맵 리터럴, `switch` 문의 case 식 전체, `||` 로 이은 `== hook.EventX` 비교 사슬 가운데 **어느 것도** 네 결정 이벤트 식별자(`EventPreToolUse`, `EventPermissionRequest`, `EventStop`, `EventUserPromptSubmit`) 중 2개 이상을 포함하지 않는다. 하위 명령 표(`hook.go:53-78`, 원소가 구조체인 리터럴)는 원소 타입이 `hook.EventType` 이 아니므로 대상이 아니다.
- 기준선: 이 트리에서 네 식별자가 `internal/cli` 비테스트 파일에 나타나는 곳은 하위 명령 표 4행(`hook.go:54`·`:57`·`:62`·`:63`)과 `runAgentHook` 매핑 switch 의 대입 2행(`event = hook.EventPreToolUse`, `hook.go:472`·`:479`)뿐이다(2026-09-24, `grep -rnE 'hook\.Event(PreToolUse|PermissionRequest|Stop|UserPromptSubmit)\b' internal/cli --include='*.go' | grep -v _test.go` → 6행, 0.3.0 에서 HEAD `44dfc25fd` 로 다시 잼; `fabc33812` 에서도 6행 — 0.2.0 기록). 매핑 switch 의 case 식은 이벤트 식별자가 아니라 접미사 검사라 (b2) 의 대상이 아니다. run 착수 시 t1099 흡수 후 트리에서 다시 잰다.

**(c) 필수 RED 변이 여섯.** run 증거에 다음을 각각 한 번 적용해 빨개지는 것을 관측하고 되돌린다: (c1) 구현의 분류 호출 결과를 `false` 로 바꾸는 변이 → (a) 가 실패한다. (c2) `internal/cli` 비테스트 파일에 `switch event { case hook.EventPreToolUse: … case hook.EventStop: … }` 처럼 **case 를 여러 줄로 나눈** 중복 목록을 넣는 변이 → (b2) 가 실패한다. (c3) 디스패처가 면제 술어 호출을 건너뛰게(술어 결과를 무시하게) 하는 변이 → Codex 모드의 관측 집합에 Stop 이 들어가 (a) 가 실패한다. (c4) `runAgentHook` 의 파싱 실패 분기가 `IsDecisionBearing` 대신 `event == hook.EventPreToolUse` 한 번 비교로 fail-closed 여부를 정하는 변이 → (b1) 이 실패한다. 이 변이는 (a)(`runHookEvent` 의 하위 명령만 관측), (b2)(결정 이벤트 식별자 1개), AC-HSF-012·013(도달 가능한 매핑이 PreToolUse·PostToolUse·SubagentStop 뿐이라 동작이 같다)을 모두 통과하므로, (b1) 이 `runAgentHook` 까지 보지 않으면 REQ-HSF-014 의 「판정은 `IsDecisionBearing` 한 곳」 절을 잡는 기준이 없다. plan-audit 에서 codex 백엔드는 줄 단위 grep 이 이 변이를 놓친다는 것을 재현했다(`duplicate-switch-was-not-detected`) — (b) 가 grep 이 아니라 AST 여야 하는 이유다. (c5) `runAgentHook` 의 파싱 실패 분기에 `_ = codexadapter.IsDecisionBearing(event)` 를 남긴 채, fail-closed 여부는 다른 실제 술어 — 예: `event != hook.EventPostToolUse && event != hook.EventSubagentStop` — 로 고르는 변이 → (b1) 의 「결과가 분기 조건식에 쓰인다」 절이 실패한다. 이 변이는 호출 존재·인자 형태·네 결정 식별자 금지를 모두 채우고, 비교 대상이 관측 이벤트 식별자라 (b2) 도 통과하며, 도달 가능한 매핑에서 동작이 같아 AC-HSF-012·013 도 통과한다 — 결과 사용 조건만이 잡는다. (c6) (0.4.0) 공유 파싱 실패 처리 함수(또는 한 진입점의 오류 분기)에 `if os.Getenv("<임의 이름>") != "" { 종전 기본 출력 }` 을 넣는 변이 → (b3) 이 실패한다. 이 변이는 변수를 설정하지 않은 모든 동작 기준(AC-HSF-001·002·012)을 통과하고, 이름을 모르는 스위치는 동작 테스트로 켤 수 없으므로 구조 검사만이 잡는다.

### AC-HSF-004 — 하네스 모드는 stdin 보다 먼저 판정된다

**Given** 착지된 구현, **When** `--harness bogus` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료이며, stderr 에 stdin 파싱 경고가 나오지 않는다. **And** 같은 실행을 `deps.HookProtocol` 을 `ReadInput` 호출 횟수를 세는 스파이로 바꿔 다시 하면 그 호출 횟수가 0 이다 — 판정이 읽기보다 앞섰다는 직접 관측이다. stderr 경고 부재는 stdin 을 읽고 오류를 보류하는 구현도 통과하는 대리 관측이므로 단독 근거로 쓰지 않는다. **When** `--harness codex` 와 malformed stdin 으로 `pre-tool` 을 실행하면, **Then** 출력이 AC-HSF-002 의 Codex 형태다 — 수리 전 트리에서는 같은 입력이 `{}` 였다(Pre-flight 기준선).

### AC-HSF-005 — 관측 이벤트는 `6a3603274` 동작을 유지한다 (특성 기준)

**Given** 착지된 구현, **When** 관측 이벤트 하위 명령 22개 각각에 파손 4 형태를 넣고 두 하네스 모드(미지정, `codex`)로 실행하면(22 × 4 × 2 = 176 경우), **Then** 모든 경우에서 `RunE` 가 nil, 디스패치 0 회, stderr 에 `invalid stdin JSON` 경고 한 줄이며, stdout 은 이벤트별 현재 출력과 바이트 단위로 같다 — worktree-create·worktree-remove 는 **빈 stdout(0 바이트)**, 나머지 20개는 정확히 `{}`. 빈 stdout 인 이유는 `writeHookOutput` 이 두 worktree 이벤트에서 `input == nil` 이면 아무것도 쓰지 않기 때문이다(`internal/cli/hook.go:389-392`). **And** 같은 176 경우의 출력이 Pre-flight 기준선 커밋에서 잰 수리 전 출력과 같다. TaskCreated·Notification 은 HOI 게이트보다 파싱이 앞서므로 게이트 설정과 무관하게 같은 결과여야 한다.

### AC-HSF-006 — 삭제됨 (0.4.0, Pre-flight 5 결과)

탈출 장치의 정당한 활성화 경로(셸 환경 변수)를 검증하던 기준이다. 탈출 장치가 없어져(REQ-HSF-007 삭제) 대상이 사라졌다. 번호는 감사 보고서 인용을 위해 비워 두며, 0.3.4 판 본문은 git 이력(커밋 `39ee312cf` 의 이 파일)에 있다.

### AC-HSF-007 — 관측 가능성과 페이로드 비노출

**Given** `CLAUDE_PROJECT_DIR` 를 `t.TempDir()` 로 둔 환경, **When** fail-closed 쌍 7 × 파손 4 형태(28 경우)와 `runAgentHook` 결정 매핑 action 4종 × 하네스 2 × 파손 4 형태(32 경우)마다 fail-closed 를 한 번씩 발화시키면, **Then** 각 경우에 (a) stderr 에 이벤트 이름·하네스 모드·파싱 오류 원인·`fail-closed` 를 담은 한 줄이 있고, (b) `.moai/logs/codex-adapter.jsonl`(또는 t1099 착지 시점의 `DiagnosticSinkRel`)에 `RecordDiscards` 형식의 행이 한 건 추가되며, 그 행의 키는 핸들러 fault 키 `hook-fault` 와 다른 파싱 실패 전용 키이고, 길이 필드는 디스패처가 관측한 stdin 바이트 수와 같다(oversize 는 5 MiB 상한 `5 << 20` 바이트, 나머지는 입력 길이), (c) 그 경우에 쓰인 카나리가 stdout·stderr·기록 파일 **어디에도** 나타나지 않는다 — 네 형태 모두에 대한 원문 비노출. **And** 필수 RED 변이: 기록의 Reason 이나 stderr 줄에 페이로드 앞부분을 싣는 변이를 한 번 적용해 (c) 가 실패하는 것을 관측하고 되돌린다.

### AC-HSF-008 — Stop 의 거부는 `stop_hook_active` 에 의존하지 않는다

**Given** `{"hook_event_name":"Stop","stop_hook_active":true,` (닫히지 않은 JSON) 과 `{"hook_event_name":"Stop","stop_hook_active":false,` 두 입력, **When** `stop` 하위 명령을 두 하네스에서 실행하면, **Then** Claude 쪽 두 경우는 `decision == "block"` 이고, Codex 쪽 두 경우는 면제 쌍이므로 stdout 이 정확히 `{}` 이며 AC-HSF-011 의 면제 stderr 줄·기록이 남는다 — 어느 쪽도 `stop_hook_active` 값에 따라 달라지지 않는다. **And** Stop fail-closed 코드 경로에 Claude 호스트 차단 상한 의존과 그 근거 등급(「저장소 독트린, 미측정」)을 적은 `@MX:WARN` 이 있다(`grep -n '@MX:WARN' <해당 파일>` 로 확인). **And** plan.md §B.1 Q2 에 판정과 증거 경로(`.moai/reports/t1152/q2-codex-stop-cap.md`)가 기록돼 있고, progress.md 에 Pre-flight 3 의 codex-cli 버전 기록과 추적 항목 Q8 의 측정 결과(또는 미측정 사유)가 있다.

### AC-HSF-009 — t1099 전제 재확인이 기록돼 있다

**Given** run 착수 시점, **Then** progress.md §E.2 에 spec.md §F.1 의 `git diff fabc33812 <착지 병합 커밋> -- …` 명령과 그 결과 요약이 있고, (a) `DecisionBearingEvents()` 원소, (b) fatal_error 열 Outcome, (c) `writeCodexFailClosed` 서명과 기록 키, (d) 네 이벤트의 `Render` 형태, (e) `Discard` 필드와 stderr 미러 형식 각각에 대해 「동일」 또는 「변경 → spec 개정 커밋 SHA」가 적혀 있다.

### AC-HSF-010 — 삭제됨 (0.4.0, Pre-flight 5 결과)

모델이 쓸 수 있는 설정면으로는 탈출 장치가 켜지지 않고 판독 실패면 인정되지 않음을 검증하던 기준이다(REQ-HSF-007·016). 두 요구가 삭제돼 대상이 사라졌다. 탈출 장치가 없다는 사실 자체는 AC-HSF-003(b3)·(c6) 이 구조로 잡는다. 번호는 비워 두며, 0.3.4 판 본문은 git 이력(커밋 `39ee312cf` 의 이 파일)에 있다.

### AC-HSF-011 — Codex 하네스의 Stop 은 면제된다

**Given** `--harness codex`, 실제 `hook.NewProtocol()`, 스파이 레지스트리, `CLAUDE_PROJECT_DIR` 는 `t.TempDir()`, **When** `stop` 하위 명령에 파손 4 형태를 각각 넣고 실행하면, **Then** (a) `RunE` 가 nil(exit 0), 디스패치 0 회, (b) stdout 이 정확히 `{}`, (c) stderr 에 이벤트·하네스 모드·파싱 오류 원인과 면제 사유(호스트에 Stop 연속 차단 상한이 없음)를 담은 한 줄이 있고, 영속 기록에 면제 전용 키(파싱 실패 fail-closed 키·`hook-fault` 와 다름)의 행이 한 건 생긴다, (d) 카나리가 stdout·stderr·기록 어디에도 없다. **And** 같은 입력을 `--harness` 미지정으로 실행하면 AC-HSF-001 대로 `decision == "block"` 이다. **And** 면제 여부는 `internal/codexadapter` 의 술어 하나에서 나온다: 필수 RED 변이 — 술어가 늘 거짓을 돌려주게 하는 변이를 한 번 적용해 (b)·(c) 가 실패(Codex Stop 이 차단으로 바뀜)하는 것을 관측하고 되돌린다. **And** 술어 선언부 주석에 Q2 증거 경로(`.moai/reports/t1152/q2-codex-stop-cap.md`), codex-cli 버전, 관측 범위, 「Codex 가 상한을 도입하면 여기를 다시 본다」가 있다(파일 판독으로 확인).

### AC-HSF-012 — `runAgentHook`: 결정 매핑 action 은 fail-closed 한다

**Given** 실제 `hook.NewProtocol()`, 스파이 레지스트리, **When** `moai hook agent` 에 결정 이벤트로 매핑되는 action 4종 — `x-validation`, `x-pre-transformation`, `x-pre-implementation`, 미지 action `foo` — 을 각각 파손 4 형태와 두 하네스 모드(미지정, `codex`)로 실행하면(4 × 4 × 2 = 32 경우), **Then** 모든 경우에서 (a) `RunE` 가 nil(exit 0), (b) 디스패치 0 회, (c) stdout 이 같은 하네스·같은 이벤트(PreToolUse)에 대해 AC-HSF-001(d)·002(d) 가 요구하는 바이트와 같고 거부 필드를 가지며, (d) AC-HSF-001(e1)~(e4) 가 성립하고, (e) AC-HSF-007 의 stderr 한 줄·파싱 실패 전용 키의 기록 한 건·카나리 비노출이 성립한다. **And** `--harness bogus` 와 malformed stdin 으로 `x-validation` 을 실행하면 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료이며 stderr 에 stdin 파싱 경고가 없으며, AC-HSF-004 와 같은 스파이 `HookProtocol` 로 잰 `ReadInput` 호출 횟수가 0 이다(매핑·하네스 판정이 stdin 보다 앞섰다는 직접 관측). **And** 유효한 stdin(`{}`)과 `--harness bogus` 로 `x-validation` 을 실행해도 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료다. 이것은 보존 동작이 아니라 이 SPEC 이 `runAgentHook` 에 새로 도입하는 거부다(REQ-HSF-005) — 수리 전 트리에서는 같은 입력이 exit 0 으로 성공한다(plan-audit 2회차에서 codex 백엔드가 재현, Pre-flight 4(c) 기준선에서 다시 잰다). 수리 전 트리에서 결정 매핑 action 은 `{}` 를 낸다(Pre-flight 4(c)).

### AC-HSF-013 — `runAgentHook`: 관측 매핑 action 은 현재 동작을 유지한다 (특성 기준)

**Given** 착지된 구현, **When** 관측 이벤트로 매핑되는 action 4종 — `x-verification`, `x-post-transformation`, `x-post-implementation`(PostToolUse), `x-completion`(SubagentStop) — 을 파손 4 형태와 두 하네스 모드로 실행하면(32 경우), **Then** 모든 경우에서 `RunE` 가 nil, 디스패치 0 회, stderr 에 `moai hook agent <action>: invalid stdin JSON` 경고 한 줄, stdout 이 정확히 `{}` 이며, Pre-flight 4(c) 기준선의 수리 전 출력과 바이트 단위로 같다. 이 보존은 유효한 하네스 모드(미지정, `codex`)에 대한 것이다. **And** 유효한 stdin(`{}`)과 `--harness bogus` 로 관측 매핑 action `x-verification` 과 `x-completion` 을 각각 실행하면, 오류가 `invalid --harness value` 를 담고 0 이 아닌 종료이며, AC-HSF-004 와 같은 스파이 `HookProtocol` 로 잰 `ReadInput` 호출 횟수가 0 이다. 이것은 보존이 아니라 REQ-HSF-005 가 agent 호출 전체에 새로 도입하는 거부다 — 수리 전 트리에서는 같은 입력이 exit 0 으로 성공한다(Pre-flight 4(c) 기준선). 필수 RED 변이 둘: (1) 관측 매핑 action 에도 fail-closed 를 적용하는 변이로 보존 단언이 실패하는 것, (2) `--harness` 판독을 결정 매핑 action(`IsDecisionBearing(event)` 가 참인 경우)에서만 하는 변이로 이 거부 단언이 실패하는 것을 각각 한 번 관측하고 되돌린다 — (2) 는 AC-HSF-012 를 통과하므로 이 단언만이 잡는다.

### AC-HSF-GATE — 품질 게이트

**Given** 착지된 구현, **When** `go test ./internal/cli/... -run 'Hook|Stdin|FailClosed|AgentHook' -count=1`, `go test ./internal/codexadapter/... -count=1`, `go vet ./internal/cli/... ./internal/codexadapter/...`, `golangci-lint run ./internal/cli/... ./internal/codexadapter/...` 를 실행하면, **Then** 모두 0 으로 끝나며, `-run` 선택자가 실제로 이 SPEC 의 테스트를 1개 이상 실행했다는 것을 `-v` 출력의 `=== RUN` 행 수로 함께 보인다(빈 선택은 통과가 아니다). 전 패키지 판정은 develop push 의 CI 가 내린다.

## §D 갱신이 필요한 기존 테스트 (이 트리 `60017eb83` 에서 측정)

| 테스트 | 위치 | 현재 단언 | 이 SPEC 이후 |
|---|---|---|---|
| `TestRunHookEvent_MalformedStdinGraceful` | `internal/cli/hook_protocol_fix_test.go:60` (입력 `:71`, 단언 `:95-124`) | `pre-tool` 에 `{"broken` → nil, 디스패치 0, stdout 이 유효한 JSON 하나. 주석(`:3-5`, `:57-59`)은 「default output」을 약속 | 단언 자체는 거부 JSON 도 유효한 JSON 이므로 **그대로 통과할 수 있다** — 바로 그래서 갱신이 필요하다. 이 테스트는 수리 전후를 가르지 못한다. 거부 형태를 단언하도록 바꾸고, 주석의 「default output」을 결정 이벤트에 한해 고친다. 관측 이벤트 쪽 의도는 AC-HSF-005 가 이어받는다 |
| `TestRunHookEvent_ReadInputError` | `internal/cli/coverage_test.go:63` | `post-tool` + mock `ReadInput` 오류 → nil, `HookProtocol.WriteOutput` 으로 기본 출력 | `post-tool` 은 관측 이벤트이므로 **변경 불필요**. 다만 결정 이벤트 경로는 `HookProtocol.WriteOutput` 을 거치지 않을 수 있으므로, 이 테스트를 결정 이벤트로 복제하지 말 것 |
| `TestRunAgentHook_ReadInputError` | `internal/cli/misc_coverage_test.go:355` (action `test-validation` → PreToolUse, 주석 `:372-374`) | mock `ReadInput` 오류 → `RunE` 가 nil 이면 통과. 주석은 「warn + default output + exit 0」을 약속 | 단언(`err == nil`)은 fail-closed 후에도 **그대로 통과한다** — 수리 전후를 가르지 못한다. 결정 매핑 action 이므로 주석을 고치고, stdout 을 잡아 거부 형태를 단언하도록 바꾼다. 관측 매핑 쪽 의도는 AC-HSF-013 이 이어받는다 |
| `TestRunAgentHook_AllActionSuffixes` | `internal/cli/coverage_fixes_test.go:810` (실행 `:848`, 매핑 단언 `:852`) | 파싱 **성공** 후 action → 이벤트 매핑이 맞는지(`capturedEvent`) | 매핑을 `ReadInput` 앞으로 옮겨도 단언은 그대로다 — 변경 불필요. 이동 후에도 통과하는지만 확인한다 |
| t1099 의 `TestHookFaultInjection` | `fabc33812:internal/cli/hook_fault_injection_test.go:137` | 핸들러 fault 경로(파싱 성공 후) | 영향 없음을 확인해야 한다 — 특히 plan.md M1 이 `writeCodexFailClosed` 서명을 넓히면, 기존 fault 경로 호출부가 같은 기록 키(`hook-fault`)와 Reason 을 유지하는지 이 테스트로 확인한다 |

`internal/hook` 패키지의 `ReadInput` 테스트는 프로토콜 계층을 바꾸지 않으므로 영향이 없다. 위 표 밖에서 결정 이벤트에 파손 stdin 을 넣고 `{}` 를 단언하는 테스트는 이 트리에서 찾지 못했다(`grep -rn 'broken\|invalid JSON\|not valid json' internal/cli/*_test.go` 결과 판독). run 착수 시 t1099 흡수 후 트리에서 같은 검색을 다시 한다.

## §E 완료 정의

- 살아 있는 AC(AC-HSF-001~005, 007~009, 011~013 — 006·010 은 0.4.0 에서 삭제)와 GATE 가 명령·원문 출력과 함께 progress.md §E.2 에 기록됐다. 필수 RED 변이(AC-HSF-003(c1)·(c2)·(c3)·(c4)·(c5)·(c6), AC-HSF-007, AC-HSF-011, AC-HSF-013 의 (1)·(2))의 실패 출력도 함께 기록됐다.
- Pre-flight 기준선 재현 커밋이 수리 커밋보다 앞선다.
- plan.md Q2·Q1·Q3 판정이 기록됐다(0.3.0, plan.md §B.1). Q3 포함에 따른 REQ·AC 추가(0.3.0)가 run 착수 전에 착지했다. Implementation Kickoff Approval 이 progress.md §E.1 「Kickoff 판정 (2026-09-24)」에 기록돼 있고, t1099 착지·흡수 뒤 구현 전에 한 전제 재확인(spec.md §F.1 의 2·3)이 progress.md 에 있다.
- depth 형태의 대조군(깊이 9000 파싱 성공)이 기록됐다.
- plan.md §C Pre-flight 5(호스트의 설정 `env` 전파·유지 측정)는 0.4.0 에서 종료됐다 — 결과 「유지한다」, 운영자 재판정으로 탈출 장치를 없앴다(spec.md HISTORY 0.4.0, 증거 `.moai/reports/t1152/preflight5.md`). 판정·실행 수·운영자 재판정은 progress.md §E.1·§E.2 에 있다. run 에서 다시 재지 않는다 — 측정 대상(탈출 장치의 출처 제약)이 더는 없다.
- `internal/hook` 패키지 diff 가 0 이다.
