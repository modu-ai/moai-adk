# SPEC-TODO-ANALYSIS-001 — 수용 기준

## §A 이 문서의 규율

Tier M — AC 상한 16, 실제 16.

각 AC 는 제목 바로 아래 첫 불릿에 **검증 대상 REQ** 를 적는다. run-phase 의 PASS/FAIL 행렬이 REQ 로 귀속돼야 어느 요구가 실제로 증명됐는지 문서 안에서 읽힌다. REQ-TA-001 ~ REQ-TA-015 전부 최소 1개의 AC 에 지칭된다.

이 계보의 선행 카드들이 **잡으려던 실패에서 그대로 통과한 기준 5개**를 냈다. 재발 방지를 위해 각 AC는 두 가지를 함께 적는다:

1. **붉게 만드는 잘못된 구현** — 이 기준이 실제로 무엇을 잡는가.
2. **기능이 통째로 없을 때 어떻게 되는가** — 부재를 주장하는 기준("아무것도 접히지 않았다")은 기능이 아예 없어도 통과하는 것이 기본값이다. 그래서 모든 부재 주장은 **같은 픽스처 안의 양성 대조**와 짝지어져 있거나, **기능 없이는 Given 자체를 만들 수 없어** 셋업에서 실패하도록 짜였다.

모든 AC는 `t.TempDir()` 픽스처 위에서 돌고, 실제 운영자 큐를 건드리지 않는다.

## §B 사전 조건

- `MOAI_TODO_ROOT` 상당의 테스트 주입 seam(또는 기존 `resolveTodoQueueRoot` 의 테스트 경로)으로 큐 파일을 임시 디렉터리에 고정한다.
- 파일 불변 주장은 `sha256(backlog.json)` 의 전후 비교로 판정한다. "보기에 같다"는 근거가 아니다.

## §C 수용 기준

### AC-TA-001 — 정확 중복은 거절되고 파일은 바이트 동일하다

- **검증**: REQ-TA-001, REQ-TA-003
- **Given** 큐에 `t1 = "Fix the flaky gate"` (state `queued`) 하나가 있고, `sha256` 을 기록해 둔 상태에서
- **When** `moai todo add "  fix   the FLAKY gate "` 를 실행하면
- **Then** exit code ≠ 0, stderr 가 `t1` 과 그 텍스트 접두를 담고, `sha256` 이 실행 전과 동일하다.
- **붉게 만드는 구현**: 정규화를 안 하는 비교(대소문자/공백 차이로 `none` 판정 → 카드가 추가됨), 거절하면서도 파일을 다시 쓰는 구현(`sha256` 변화).
- **기능 부재 시**: 카드가 추가되고 exit 0 → 실패. 부재 주장이 아니므로 안전하다.

### AC-TA-002 — `--force` 는 추가하되 그 사실을 기록한다

- **검증**: REQ-TA-004
- **Given** AC-TA-001 과 같은 픽스처에서
- **When** `moai todo add "fix the flaky gate" --force` 를 실행하면
- **Then** exit 0, `items` 길이가 정확히 1 증가하고, `findings` 에 `{relation: "duplicate-forced", related_id: "t1", source: "mechanical"}` 이 **정확히 1건** 존재한다.
- **붉게 만드는 구현**: `--force` 가 소견 기록을 건너뛰는 구현(강제된 중복이 흔적 없이 남음), 소견을 2건 쓰는 구현.
- **기능 부재 시**: `--force` 플래그가 없어 파싱 에러이거나 `findings` 가 없어 실패.

### AC-TA-003 — 근접 중복은 양쪽 카드를 그대로 두고 기록만 남긴다 (양성 + 음성 대조)

- **검증**: REQ-TA-001, REQ-TA-005
- **Given** 큐에 `t1 = "Rework the auth middleware error paths"` 가 있고 `t1` 의 5필드 스냅샷을 떠 둔 상태에서
- **When** ① `moai todo add "Rework auth middleware error paths"` 를 실행하고, 이어서 ② `moai todo add "Add a Windows CI matrix job"` 를 실행하면
- **Then** ①의 새 카드 `text` 가 인자와 바이트 동일하고, `t1` 의 5필드가 스냅샷과 바이트 동일하며, `near-duplicate` 소견이 score ≥ 0.80 으로 **1건** 생기고 — ②의 카드에 대해서는 소견이 **0건** 이다.
- **붉게 만드는 구현**: 근접 중복을 흡수/편집하는 구현(`t1` 또는 새 카드의 텍스트 변화), 임계값 없이 모든 쌍에 소견을 다는 구현(②가 0건이 아님), 임계값이 너무 높아 ①을 놓치는 구현.
- **기능 부재 시**: ①의 소견이 0건이라 실패. 음성 대조(②=0건)만 있었다면 부재 시에도 통과했을 것이고, 그 짝짓기가 이 AC의 요점이다.

### AC-TA-004 — `relate` 는 소견만 남기고, `unrelate` 는 소견만 지운다 (양방향)

- **검증**: REQ-TA-008, REQ-TA-009
- **Given** `t1`~`t4` 가 queued 이고, `t3`↔`t4` 를 지칭하는 `contains` 소견이 **이미 1건** 있으며, `items` 배열 전체의 JSON 스냅샷을 떠 둔 상태에서
- **When** ① `moai todo relate t2 t1 --relation absorbs --note "t2 covers t1"` 를 실행하고, 이어서 ② `moai todo unrelate <①이 만든 소견의 index>` 를 실행하면
- **Then** 두 실행 모두 exit 0 이고 — ① 후 `findings` 길이가 2 이며 그중 1건이 `{subject_id: "t2", related_id: "t1", relation: "absorbs", source: "agent"}` 이고, ② 후 `findings` 길이가 **정확히 1** 이며 남은 1건이 `t3`↔`t4` 이고 `{t2, t1, absorbs, agent}` 를 지칭하는 소견이 0건이며, `items` 배열이 ①·② 두 시점 모두 스냅샷과 **바이트 동일**하다.
- **붉게 만드는 구현**: `absorbs` 를 실제 흡수로 해석해 `t1` 을 drop 하거나 `t2` 텍스트에 병합하는 구현 — 독트린이 이름으로 금지한 바로 그 동작; `unrelate` 가 소견을 지우면서 카드까지 건드리는 구현(②의 `items` 바이트 동일 주장에서 실패); 지목한 것이 아니라 다른 소견을 지우는 구현(남은 1건이 `t3`↔`t4` 임을 주장하므로 잡힌다); 과잉 회수로 둘 다 지우는 구현(길이 정확히 1 주장에서 실패). 길이를 양방향으로 못박으므로 과소·과잉 둘 다 걸린다.
- **기능 부재 시**: `relate` / `unrelate` 동사가 없어 exit ≠ 0 → 실패.

### AC-TA-005 — 네 의미 관계 중 어느 것도 큐를 바꾸지 않는다

- **검증**: REQ-TA-009, REQ-TA-010
- **Given** 카드 3장, 그리고 `contains` / `absorbs` / `replaces` / `conflicts` 소견을 각각 1건씩 `relate` 로 기록한 뒤 파일의 `sha256` 을 기록한 상태에서
- **When** `moai todo list`, `moai todo next`, `moai todo why t1`, `moai todo list --json` 을 순서대로 실행하면
- **Then** 네 명령 모두 exit 0 이고, 실행 후 `sha256` 이 기록값과 동일하다 (읽기 명령은 아무것도 쓰지 않는다).
- **붉게 만드는 구현**: 소견을 보고 카드를 정리하는 구현, 읽기 경로에서 재분석하며 파일을 갱신하는 구현.
- **기능 부재 시**: `relate` 가 없어 **Given 을 구성할 수 없다** → 셋업에서 실패한다. 부재로 인한 위양성이 구조적으로 불가능하다.

### AC-TA-006 — 분석은 순서를 절대 건드리지 않는다

- **검증**: REQ-TA-009
- **Given** 큐가 `[t1, t2, t3]` 순서인 상태에서
- **When** 카드 5장을 추가하되 그중 2장은 근접 중복, 1장은 `--force` 정확 중복으로 넣으면
- **Then** 앞 3개 위치의 id 순열이 정확히 `[t1, t2, t3]` 이고, 새 카드 5장이 추가된 순서 그대로 꼬리에 붙어 있다.
- **붉게 만드는 구현**: 순서 정리를 구현한 것 전부 — 중복 카드를 원본 옆으로 모으는 구현 포함.
- **기능 부재 시**: `--force` 가 없어 3번째 add 가 실패 → 셋업에서 실패.

### AC-TA-007 — 읽기는 멱등이다

- **검증**: REQ-TA-010
- **Given** 소견이 달린 큐에서
- **When** 중간에 아무 명령 없이 `moai todo list --json` 을 연속 두 번 실행하면
- **Then** 두 stdout 바이트가 동일하고, 파일 `sha256` 이 두 실행 전후로 모두 동일하다.
- **붉게 만드는 구현**: 읽기 시 분석하는 구현, 소견에 매 실행 타임스탬프를 새로 찍는 구현 — "두 `list` 호출 사이에 큐가 스스로 재조직되는" 형태를 정확히 잡는다.
- **기능 부재 시**: `findings` 키가 없어 Given 을 만들 수 없다 → 셋업 실패.

### AC-TA-008 — `done` 은 소견을 함께 회수한다

- **검증**: REQ-TA-007
- **Given** `t1`↔`t2` 소견 1건과 `t3`↔`t4` 소견 1건, 총 2건이 있는 상태에서
- **When** `moai todo done t1` 을 실행하면
- **Then** `findings` 길이가 정확히 1이고, 남은 1건이 `t3`↔`t4` 이며, `t1` 을 어느 위치로든 지칭하는 소견이 0건이다.
- **붉게 만드는 구현**: 회수를 안 해 소견이 사라진 카드를 가리키는 구현, 과잉 회수로 무관한 소견까지 지우는 구현(길이가 0이 됨). 길이를 정확히 주장하므로 양방향으로 잡힌다.
- **기능 부재 시**: Given 구성 불가 → 셋업 실패.

### AC-TA-009 — `list` 는 소견과 함께 "그 다음에 뭘 할지"를 보여준다

- **검증**: REQ-TA-011
- **Given** `t5` 가 `t1` 을 지칭하는 `near-duplicate` 소견을 갖는 상태에서
- **When** `moai todo list` 를 실행하면
- **Then** 출력에 `t5` 행 아래 들여쓰기된 줄이 있고, 그 줄이 `near-duplicate`, `t1`, `mechanical`, 그리고 리터럴 `moai todo drop` 또는 `moai todo edit` 문자열을 모두 담는다.
- **붉게 만드는 구현**: 소견을 `--json` 에만 싣고 사람이 보는 `list` 에는 안 싣는 구현 — 운영자가 볼 수 없는 분석은 이 SPEC의 전제를 깨뜨린다.
- **기능 부재 시**: 해당 줄이 없어 실패.

### AC-TA-010 — 소견 없는 카드는 "없다"고 말한다

- **검증**: REQ-TA-012
- **Given** 소견이 하나도 없는 `t1` 에 대해
- **When** `moai todo why t1` 를 실행하면
- **Then** exit 0 이고 stdout 이 **비어 있지 않으며** 명시적인 no-findings 문구를 담는다.
- **붉게 만드는 구현**: 아무것도 출력하지 않는 구현 — 무소견과 크래시가 운영자에게 동일하게 보이는 상태.
- **기능 부재 시**: `why` 동사가 없어 exit ≠ 0 → 실패.

### AC-TA-011 — `machine-only` 표시는 붙고, 또 걷힌다 (양방향)

- **검증**: REQ-TA-013
- **Given** `t5`↔`t1` 기계 `near-duplicate` 소견만 있고 같은 쌍에 에이전트 소견이 없는 상태에서
- **When** ① `moai todo list` 를 실행하고, ② 기계 소견과 **반대 방향**인 `moai todo relate t1 t5 --relation replaces` 후 다시 `moai todo list` 를 실행하면
- **Then** ①의 해당 소견 줄은 `machine-only` 를 담고, ②의 같은 줄은 담지 않는다.
- **붉게 만드는 구현**: 표시가 아예 없는 구현(①에서 실패), 한 번 붙으면 절대 안 걷히는 구현(②에서 실패), 쌍을 **순서 있게** 비교해 `{t1, t5}` 를 `{t5, t1}` 과 다른 쌍으로 보는 구현(②에서 표시가 안 걷혀 실패 — REQ-TA-013 의 무순서 비교 규정을 이 방향 선택이 검증한다). 한쪽만 주장했다면 다른 쪽 구현이 통과했을 것이다.
- **기능 부재 시**: ①에서 실패.

### AC-TA-012 — 기존 파일이 그대로 읽히고, 항목 스키마는 정확히 5필드로 남는다

- **검증**: REQ-TA-006
- **Given** `findings` 키가 없는 이 기능 이전 형식의 `backlog.json` 픽스처에서
- **When** `moai todo list --json` 을 실행하고, 이어서 `moai todo add "new card"` 를 실행한 뒤 그 결과 파일을 `map[string]json.RawMessage` 로 열어 `items[]` **각 원소의 키 집합을 직접 세면**
- **Then** 첫 실행이 exit 0 이고 기존 항목 5필드가 바이트 동일하게 왕복하며 `findings` 가 빈 배열로 렌더되고 — 결과 파일의 **모든** `items[]` 원소의 키 집합이 정확히 `{id, text, added_at, spec_id, state}` 5개이며(초과도 누락도 없음), 최상위에는 `findings` 가 가산적으로 추가돼 있다.
- **붉게 만드는 구현**: `findings` 부재를 에러로 취급하는 구현; 항목에 새 필드를 추가해 REQ-TODO-013 의 5필드 계약을 깨는 구현.
- **왜 키 집합을 직접 세는가**: Go 의 `encoding/json` 은 **모르는 필드를 조용히 무시**하므로, "구 스키마 구조체로 디코딩이 에러 없이 성공한다"는 판정은 항목에 필드가 *추가된* 위반을 구조적으로 잡지 못한다(삭제·개명만 부분적으로 잡는다). 키 집합을 세거나 `json.Decoder` + `DisallowUnknownFields()` 를 **항목 단위로** 걸어야 잡힌다. 같은 옵션을 최상위 레코드에 걸면 가산적 `findings` 때문에 **올바른 구현에서 붉어지므로** 쓰지 않는다.
- **기능 부재 시**: `findings` 가 빈 배열로도 안 나와 실패.

### AC-TA-013 — 새 동사 전부가 headless 다

- **검증**: REQ-TA-015
- **Given** stdin 이 닫힌 비대화형 셸에서
- **When** `add`, `analyze`, `relate`, `unrelate`, `why`, `list` 를 각각 실행하고, `moai todo --help` 를 확인하고, `internal/cli/todo_test.go` 의 기존 `todoPromptGuard` 를 **todo 계열 비테스트 소스 전부**(`internal/cli/todo*.go` — `todo.go`, `todo_drop.go`, `todo_edit_move.go`, 그리고 신규 `todo_relate.go` / `todo_why.go`)에 대해 돌리면
- **Then** 여섯 명령이 블로킹 없이 `0` / `1` / `2` 중 하나의 exit code 로 끝나고(REQ-TODO-014 의 계약 그대로), `--help` 가 `analyze` / `relate` / `unrelate` / `why` 를 모두 나열하며, 가드가 어느 파일에서도 위반을 보고하지 않고, **같은 가드가 합성 위반 `x := AskUserQuestion()` 에 대해서는 위반을 보고한다**(기존 `TestTodoCmd_NoAskUserQuestion` 의 음성 대조를 그대로 유지).
- **붉게 만드는 구현**: 대화형 확인을 넣은 구현(블로킹 또는 EOF 에러 — REQ-TODO-014 위반); 신규 동사 파일을 가드 대상에서 빼는 구현 — 가드가 `todo.go` 하나만 읽으면 `todo_relate.go` 에 들어간 프롬프트를 못 잡는다. 이 AC 가 요구하는 것은 가드의 **존재**가 아니라 **범위 확장**이다.
- **쓰지 않는 기준 (폐기)**: `grep -rn "AskUserQuestion" internal/cli/` 결과 0건은 기준으로 쓰지 않는다. 실측 **317건 / 82파일**(`f1bc39310`)이고 전부 부재를 주장하는 주석과 테스트 이름이므로, 올바른 구현에서도 절대 통과할 수 없다.
- **기능 부재 시**: `--help` 에 새 동사가 없어 실패 — 가드 무위반 주장 단독이면 부재 시에도 통과했을 것이므로 `--help` 양성 대조가 이 AC를 지탱한다.

### AC-TA-014 — 독트린 개정이 로컬·템플릿 쌍둥이에 함께 착지한다

- **검증**: REQ-TA-014
- **Given** 개정 대상 4개 파일(로컬 2 + 템플릿 미러 2)에 대해
- **When** `diff` 로 쌍둥이를 비교하고 본문을 grep 하면
- **Then** `todo.md` 쌍과 `kanban-dispatch.md` 쌍이 각각 바이트 동일하고, 양쪽 모두 분석 경계 절을 담으며 그 절이 "records" 와 "never folds" 에 해당하는 문구를 모두 포함한다. 그리고 기존 [HARD] 금지 문장(추론된 우선순위·tidy-up·접어 넣기·오래돼 보임)이 **삭제되지 않고 남아 있다**.
- **붉게 만드는 구현**: 로컬만 고치고 미러를 빼먹은 구현(`moai update` 가 다음 실행에서 되돌린다), 충돌을 없애려고 기존 금지 문장을 흐리거나 지운 구현 — 후자는 이 카드가 명시적으로 경계한 실패다.
- **기능 부재 시**: 경계 절 grep 이 0건 → 실패.

### AC-TA-015 — 정확 중복 거절은 에이전트 없이도 작동한다

- **검증**: REQ-TA-001, REQ-TA-003
- **Given** 에이전트가 관여하지 않는 순수 CLI 실행에서, `t1 = "Fix the auth bug"` 가 있는 큐에 대해
- **When** ① `moai todo add "fix the auth bug"` 를 실행하고, ② `moai todo add "Repair broken login"` 를 실행하면
- **Then** ①은 거절되고(exit ≠ 0), ②는 exit 0 으로 추가되며 어떤 소견도 생기지 않는다.
- **붉게 만드는 구현**: 기계 계층을 에이전트 의존으로 만든 구현(①이 통과함), 텍스트가 다른 의미 중복까지 CLI에서 잡으려 한 구현(②가 거절됨 — 이 SPEC의 범위 밖이자 위험한 오탐).
- **기능 부재 시**: ①이 통과 → 실패. ②의 음성 주장은 ①의 양성 주장과 한 AC 안에 묶여 있어 단독으로 통과할 수 없다.

### AC-TA-016 — `analyze` 재실행은 소견을 다시 쌓지 않는다

- **검증**: REQ-TA-002, REQ-TA-009
- **Given** 근접 중복 쌍이 최소 1건 성립하는 카드 3장짜리 큐에서(소견은 아직 0건), `items` 배열의 JSON 스냅샷을 떠 둔 상태에서
- **When** `moai todo analyze` 를 중간에 아무 명령 없이 **연속 두 번** 실행하면
- **Then** 첫 실행 후 `findings` 길이가 **0보다 크고**, 두 번째 실행 후 그 길이가 첫 실행 후와 **정확히 동일**하며, `{subject_id, related_id, relation, source}` 가 같은 소견이 2건 이상인 경우가 0건이고, `items` 배열이 두 실행 전후 모두 스냅샷과 바이트 동일하다.
- **붉게 만드는 구현**: 재실행마다 무조건 append 하는 구현(두 번째 실행 후 길이가 증가 — `list` 가 같은 소견으로 덮이고 운영자가 소견을 전부 무시하게 되는 실패로 직행한다), `analyze` 가 카드를 append·remove·reorder·edit 하는 구현(`items` 바이트 동일 주장에서 실패).
- **기능 부재 시**: `analyze` 동사가 없어 exit ≠ 0 → 실패. 첫 실행 후 "길이 > 0" 주장이 양성 대조라, 소견을 아예 안 만드는 구현이 "길이 불변"만으로 통과하는 것을 막는다.

## §D 품질 게이트

- `go test ./internal/cli/... ./internal/kanban/...` 통과. `internal/cli` 단독 실행이 ~336초이므로 `-timeout 900s`. **전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다** — 전 패키지 판정은 CI 몫.
- `go vet ./...` / `golangci-lint run` 통과.
- `GOOS=windows go vet ./internal/cli/... ./internal/kanban/...` — 크로스 플랫폼 컴파일 확인(단, 이것은 컴파일만 증명하며 Windows 동작 근거가 아니다).
- 신규 코드 커버리지 85% 이상.
- `make build` 후 템플릿 미러 parity 확인.
- 템플릿 중립성: `internal/template/templates/` 에 들어가는 문구에 SPEC ID / REQ 토큰 / 내부 날짜 / 커밋 SHA 가 없을 것.

## §E Definition of Done

- AC-TA-001 ~ AC-TA-016 전부 PASS, 각 PASS 가 실제로 실행된 명령의 출력에 귀속될 것.
- 독트린 4개 파일 개정 + 쌍둥이 바이트 동일.
- `.moai/state/kanban/backlog.json` (실제 운영자 큐)이 이 SPEC의 어떤 단계에서도 수정되지 않았을 것 — 작업 전후 `sha256` 동일로 확인.
