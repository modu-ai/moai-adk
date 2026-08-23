# SPEC-TODO-ANALYSIS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (spec.md + plan.md + acceptance.md). 근거는 `plan.md §B`. **plan-audit iteration 1 이 Tier M 유지를 판정**했다(개정 대상이 헌법이 아니라 워크플로 규칙 1 + 스킬 1, 추가-only, 파일·LOC 두 축 모두 M 범위).
- REQ 15 / AC 16 (Tier M 상한 각 16). iteration 1 대비 REQ 불변, AC +1 — `analyze` 재실행 멱등성 AC 신설(AC-TA-016), `unrelate` 는 신규 AC 대신 AC-TA-004 를 `relate`+`unrelate` 양방향으로 확장해 흡수(상한 16 안에서 두 미검증 분기를 모두 덮기 위한 선택).
- SPEC ID regex check: 실행됨, `PASS` (`internal/spec/lint.go` `specIDPattern` 대응 패턴).
- ID 충돌 검사: `.moai/specs/SPEC-TODO-*` 0건.
- Out of Scope: 6개 `### Out of Scope — <topic>` 소제목, 각 `-` 불릿 보유 (불변).
- 독트린 충돌 입장: `spec.md §B` — 가역성 단독이 아니라 **가역성 + 현저성**을 자동 변형의 허가 기준으로 세우고, 순서 정리와 카드 흡수는 현저성 미달로 기각, 정확 중복 입력 거절과 소견 기록만 자동 허용.
- 상태: plan-audit iteration 2 대기.

### iteration 2 — 해소한 결함 (blocking 7 + D7, optional 4)

`f1bc39310` 기준 `.moai/reports/t119/plan-audit.md` 의 결함 목록에 대한 델타. 인용한 수치는 이번 개정 시점에 **재측정**했다.

| 결함 | 해소 방식 | 착지 위치 |
|---|---|---|
| D1 (critical) | `grep -rn "AskUserQuestion" internal/cli/` → 0건 기준을 **폐기**하고, 같은 패키지의 정본 관용구(`internal/cli/todo_test.go` `todoPromptGuard` / `TestTodoCmd_NoAskUserQuestion`, 음성 대조 포함)를 **신규 동사 파일까지 범위 확장**하는 형태로 교체 | `acceptance.md` AC-TA-013 · `plan.md §E M4` |
| D2 | §B.1 표의 "정확 중복 거절" 행을 무조건 "현저: 예"에서 **"호출자가 exit code 또는 stderr 를 읽을 때에만"** 으로 조건화하고, 완화가 닿지 않는 스크립트·파이프 경로를 **명시적 잔여 위험으로 선언**(REQ 신설 아님 — CLI 는 호출자가 자기 종료 코드를 읽게 만들 수 없다) | `spec.md §B.1` · `plan.md §D.3` |
| D3 | 16개 AC 전부에 `- **검증**: REQ-TA-…` 첫 불릿 추가. REQ-TA-001~015 전량 최소 1회 지칭 | `acceptance.md` 전체 |
| D4 | REQ-TA-002 에 재실행 비중복 절 추가(같은 `{subject_id, related_id, relation, source}` 튜플 재append 금지) + 이를 붉게 만드는 AC-TA-016 신설(연속 2회 실행 후 `findings` 길이 불변, 첫 실행 후 길이 > 0 양성 대조) | `spec.md` REQ-TA-002 · `acceptance.md` AC-TA-016 |
| D5 | `unrelate` 를 AC-TA-004 에 흡수: `items` 스냅샷 바이트 동일 + `findings` 길이 정확히 −1 + 남은 1건이 무관한 `t3`↔`t4` 임을 주장(지목한 그 건이 지워졌음을 양방향으로 못박음) | `acceptance.md` AC-TA-004 |
| D6 | exit 계약 `0/1` → `0/1/2` 정정. 원문 재측정: `SPEC-KANBAN-TODO-CLI-001/spec.md:63` REQ-TODO-014 = "exit 0/1/2" | `spec.md` REQ-TA-015 · `acceptance.md` AC-TA-013 |
| D7 | AC-TA-012 의 판정 수단을 "구 스키마 구조체 디코딩 성공"에서 **`items[]` 각 원소의 키 집합 정확히 5개** 주장으로 교체 + Go 의 `encoding/json` 이 미지 필드를 무시한다는 이유를 AC 본문에 명시(최상위 `DisallowUnknownFields` 는 가산적 `findings` 때문에 올바른 구현에서 붉어지므로 금지) | `acceptance.md` AC-TA-012 |
| D10 | §B.1 에 카드 t119 의 완화안을 **원문 그대로 인용**("오판 1건이 카드를 조용히 삼킬 수 있으므로 드롭은 되돌릴 수 있어야 한다(undrop 동사 존재)")하고 한 문단으로 반박 — 완화안이 틀린 것이 아니라 흡수·재정렬에만 **닿지 않는다**(발동 전제인 "누군가 알아챈다"를 변형 자체가 무너뜨린다), `undrop` 은 그대로 존치 | `spec.md §B.1` |
| D8 (optional) | "바이트 단위로 동일하다" → 파일에는 단조 `t<N>` id 와 `added_at` 이 남아 **재정렬 발생 자체는 기계적으로 탐지되며**, 탐지되지 않는 것은 *누가·왜* 이고 운영자 기본 화면에는 흔적이 안 뜬다로 정정. 재측정: `internal/cli/todo.go:309` 의 사람용 렌더는 `id`·`state`·`text` 3열뿐(`AddedAt` 은 `:271` 생성 시점에만 등장) | `spec.md §B.1` |
| D9 (optional) | 현저성 기준에 **"운영자가 저작한 카드의 내용·순서·존재"** 범위 한정 추가. 재측정: `normalizeBacklogRecord`(`internal/kanban/backlog_store.go:326`)가 `:333-335` 에서 `last_seq` 를 조용히 끌어올린다 — 카드의 내용·순서·존재를 바꾸지 않으므로 기준 대상 아님을 본문에 명시 | `spec.md §B.1` |
| D11 (optional) | 인용 줄번호 정정. 재측정: "Nothing here infers…" = `todo_edit_move.go:5-7`(감사는 `:5-6`, 실제 문장은 5–7행에 걸침), "Order is the only thing…" = `:99`. 인용 문구 자체는 변경 없음 | `spec.md §A.2`, `§B.1` |
| D12 (optional) | M2 의 참조 갱신 항목 삭제. 재측정: `backlog_store.go:46` 의 `(spec.md §E out-of-scope)` 는 `SPEC-KANBAN-TODO-CLI-001/spec.md:92` "No version bump and no new per-item fields" 를 가리키는 정확한 인용이며, 이 SPEC 으로 돌리면 근거 없는 곳을 가리키게 된다 | `plan.md §E M2` |
| D13 (optional) | REQ-TA-013 에 "쌍은 **무순서** 비교" 명시 + AC-TA-011 ②를 기계 소견과 **반대 방향**(`relate t1 t5`)으로 바꿔 한 단계 안에서 검증 | `spec.md` REQ-TA-013 · `acceptance.md` AC-TA-011 |

### 재측정한 수치 (감사 주장 대조)

| 항목 | 감사 기재 | 재측정 (`f1bc39310`, 이번 개정 시점) | 판정 |
|---|---|---|---|
| `grep -rn "AskUserQuestion" internal/cli/` | 317건 / 82파일 | **317건 / 82파일** | 일치 |
| REQ-TODO-014 exit 계약 | `0/1/2` (`spec.md:63`) | **`0/1/2`, `spec.md:63`** | 일치 |
| `grep -c "REQ-TA-" acceptance.md` (개정 전) | 0 | **0** → 개정 후 **16개 AC 전부 태그** | 해소 |
| "Nothing here infers…" 위치 | `:5-6` | **`:5-7`** (문장이 5–7행에 걸침) | 감사보다 1행 넓게 정정 |
| "Order is the only thing…" 위치 | `:99` | **`:99`** | 일치 |
| `normalizeBacklogRecord` 위치 | `:326` | **`:326`** (`last_seq` 인상은 `:333-335`) | 일치 |
| `backlog_store.go:46` 참조 대상 | SPEC-KANBAN-TODO-CLI-001 §E | **`SPEC-KANBAN-TODO-CLI-001/spec.md:92`** | 일치 (갱신 대상 아님) |
| `moai todo list` 사람용 렌더 열 | `added_at` 미출력 | **`id`·`state`·`text` 3열** (`todo.go:309`) | 일치 |
| REQ / AC 수 | 15 / 15 | **15 / 16** (Tier M 상한 각 16) | 상한 내 |

### 미검증 (Gaps)

- `go test` 를 한 건도 돌리지 않았다 — plan-phase 이고 구현이 없으며, 로컬 전체·대형 스위트 금지 규율을 따랐다. 코드 관련 주장은 전부 **정적 읽기 + grep** 에 귀속된다.
- AC-TA-016(`analyze` 멱등성)과 AC-TA-004 의 `unrelate` 절이 **실제로 붉어지는지**는 run-phase 전까지 알 수 없다. 기준의 형태만 정했고 동작은 미검증이다.
- AC-TA-003 의 Jaccard 0.833 은 임계 0.80 대비 여유 0.033 뿐이다(감사 산술 재측정값, 이번 개정에서 픽스처 변경 없음). 정규화 단계에 불용어 제거 같은 변형이 들어가면 이 픽스처가 임계를 넘나들 수 있다 — 구현 시 경계값임을 인지할 것.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
