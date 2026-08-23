# SPEC-TODO-ANALYSIS-001 — plan-audit iteration 2 개정 요약

- 대상: `.moai/specs/SPEC-TODO-ANALYSIS-001/{spec,plan,acceptance,progress}.md`
- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t119`, branch `WT-todo-auto-analysis`, 개정 시작 HEAD `f1bc39310` (커밋·푸시 없음)
- 근거 문서: `.moai/reports/t119/plan-audit.md` (iteration 1, FAIL 0.71 / 임계 0.80)
- 성격: **개정**이지 재작성이 아니다. §B.1 의 현저성 기준, Tier M 사이징, 논증 구조, 6개 Out of Scope 절은 그대로 두고 결함만 걷어냈다.
- 규모: REQ 15 → **15** (불변), AC 15 → **16** (Tier M 상한 각 16)

---

## 1. 결함별 개정 내역 (blocking 7 + D7, optional 5)

### D1 (critical, blocking) — AC-TA-013 의 grep 기준이 올바른 구현에서 붉어진다

- **바꾼 것**: `grep -rn "AskUserQuestion" internal/cli/` → 0건 기준을 **폐기**하고, 같은 패키지에 이미 확립된 정본 관용구로 교체했다 — `internal/cli/todo_test.go` 의 `todoPromptGuard` 를 **todo 계열 비테스트 소스 전부**(`internal/cli/todo*.go`: `todo.go`, `todo_drop.go`, `todo_edit_move.go`, 신규 `todo_relate.go` / `todo_why.go`)에 대해 돌리고, `TestTodoCmd_NoAskUserQuestion` 의 음성 대조(`x := AskUserQuestion()` 를 가드가 잡을 것)를 그대로 유지한다.
- AC 본문에 "**쓰지 않는 기준 (폐기)**" 절을 남겨, 폐기된 grep 형태가 왜 통과 불가능한지(실측 317건 / 82파일, 전부 부재를 주장하는 주석·테스트 이름) 다음 독자가 다시 채택하지 못하게 했다.
- **붉게 만드는 구현**을 "가드의 존재"가 아니라 "가드가 `todo.go` 하나만 읽는 구현"으로 못박아, 이 AC 가 요구하는 것이 신설이 아니라 **범위 확장**임을 분명히 했다.
- **테스트 줄번호는 인용하지 않았다.** 감사와 지시문은 `todo_test.go:451` / `:461` / `:469` 를 인용했지만, 줄번호는 D11 이 지적한 바로 그 드리프트 결함 유형이다. 대신 함수명(`todoPromptGuard`)과 테스트명(`TestTodoCmd_NoAskUserQuestion`)으로 지칭했다 — 둘 다 greppable 하고 위치 변경에 견딘다.
- 착지: `acceptance.md` AC-TA-013 전체 · `plan.md §E M4` (가드 범위 확장을 M4 산출로 명시)

### D2 (major, blocking) — 현저성 기준이 자기가 허용한 행위에 적용되지 않는다

- **바꾼 것**: `spec.md §B.1` 표의 "정확 중복 add 를 거절" 행을 무조건 `현저: 예` 에서 **`호출자가 exit code 또는 stderr 를 읽을 때에만`** 으로 조건화하고, 판정을 `자동 허용` → `조건부 자동 허용` 으로 바꿨다.
- 조건이 성립하지 않는 경로(스크립트·파이프)는 **명시적 잔여 위험**으로 선언했다. `§B.1` 말미에 전용 문단을 두고, `manager-lead` 완화가 L2 경로에만 성립하는 반면 `§B.3` 은 에이전트 없는 L1 단독 작동을 요구하며 `AC-TA-015` 가 그것을 강제한다는 충돌을 지면에 올렸다.
- **REQ 를 신설하지 않은 이유**를 함께 적었다: CLI 는 호출자가 자기 종료 코드를 읽게 만들 수 없다. 이 경로를 덮는다고 주장하는 REQ 는 증명할 수 없는 주장이 되고, 이 SPEC 이 세운 기준(증명되지 않는 것을 허용으로 적지 않는다)에 스스로 걸린다. 감사는 REQ 신설과 잔여 위험 선언 중 택일을 허용했고, 후자가 정직한 답이다. 부수 효과로 REQ 예산이 15 로 남아 상한 16 에 여유 1 이 유지된다.
- `--force` 가 이 경우의 탈출구가 **아니라는** 점(의도적 중복 입력용)도 명시했다.
- 착지: `spec.md §B.1` (표 1행 + 잔여 위험 문단) · `plan.md §D.3` (완화가 닿지 않는 경로 불릿 추가, `spec.md §B.1` 로 상호참조)

### D3 (major, blocking) — 어느 AC 도 REQ 를 지칭하지 않는다

- **바꾼 것**: 16개 AC 전부에 제목 바로 아래 `- **검증**: REQ-TA-…` 첫 불릿을 넣었다. `acceptance.md §A` 에 이 규율과 그 이유(run-phase PASS/FAIL 행렬의 REQ 귀속)를 한 문단으로 적었다.
- 커버리지 실측: `REQ-TA-001` ~ `REQ-TA-015` **15개 전부** 최소 1개 AC 에 지칭된다 (`grep -o 'REQ-TA-[0-9][0-9][0-9]' acceptance.md | sort | uniq -c` 로 확인 — 001:4, 002:1, 003:2, 004:1, 005:1, 006:1, 007:1, 008:1, 009:4, 010:2, 011:1, 012:1, 013:2, 014:1, 015:2).
- 착지: `acceptance.md` 전체 + `§A`

### D4 (major, blocking) — `analyze` 를 검증하는 AC 가 없고 재실행 멱등성이 미정의

- **REQ 쪽**: `REQ-TA-002` 의 `analyze` 분기에 비중복 절을 추가 — 같은 `{subject_id, related_id, relation, source}` 튜플을 두 번 append 하지 않으며, 연속 두 번 실행이 `findings` 배열 길이를 바꾸지 않는다. GEARS 패턴은 Event-driven 그대로 유지(복합 When + shall not).
- **AC 쪽**: `AC-TA-016` 신설. 연속 2회 `analyze` 후 길이 불변 + 동일 튜플 중복 0건 + `items` 바이트 동일. **양성 대조**로 "첫 실행 후 길이 > 0" 을 함께 주장해, 소견을 아예 만들지 않는 구현이 "길이 불변"만으로 통과하는 것을 막았다.
- 착지: `spec.md` REQ-TA-002 · `acceptance.md` AC-TA-016 (신설) · `plan.md §E M3` 산출 목록

### D5 (major, blocking) — `unrelate` 를 검증하는 AC 가 없다

- **바꾼 것**: 신규 AC 대신 `AC-TA-004` 를 `relate` + `unrelate` **양방향**으로 확장했다. Tier M AC 상한이 16 이고 D4 가 이미 1칸을 쓰므로, 두 미검증 분기를 상한 안에서 모두 덮는 유일한 배치다. 두 동사가 같은 픽스처·같은 스냅샷을 공유하므로 분리보다 오히려 강한 주장이 된다.
- Given 에 무관한 `t3`↔`t4` 소견을 **미리 1건** 깔았다. 소견이 1건뿐이면 "길이 −1"과 "지목한 그 건이 사라졌다"가 같은 주장이 되어 오지목 구현을 못 잡는다. 2건 → 1건으로 만들면 남은 1건이 `t3`↔`t4` 임을 주장할 수 있고, 과소 회수·과잉 회수·오지목이 모두 걸린다.
- Then: 두 실행 모두 exit 0 · ① 후 길이 2 · ② 후 길이 **정확히 1** · 남은 1건이 `t3`↔`t4` · `{t2,t1,absorbs,agent}` 0건 · `items` 배열이 ①② 두 시점 모두 스냅샷과 **바이트 동일**.
- 착지: `acceptance.md` AC-TA-004 (제목 포함 전면 개정) · `plan.md §E M4`

### D6 (major, blocking) — headless exit 계약을 좁혀 옮겨 적었다

- **바꾼 것**: `0/1` → **`0/1/2`**, 두 곳 모두.
- 원문 재측정: `.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:63` REQ-TODO-014 — "structured stdout for `--json`, human-readable stderr, **exit 0/1/2**, no AskUserQuestion anywhere in the package".
- 착지: `spec.md` REQ-TA-015 · `acceptance.md` AC-TA-013 Then 절

### D7 (major, blocking-classed) — AC-TA-012 가 선언한 위반을 잡지 못한다

- **바꾼 것**: 판정 수단을 "구 스키마 구조체로 디코딩하면 에러 없이 성공" 에서 **`items[]` 각 원소의 키 집합이 정확히 `{id, text, added_at, spec_id, state}` 5개** 주장으로 교체했다(`map[string]json.RawMessage` 로 열어 직접 센다). 제목도 "새 파일이 옛 로더에서 그대로 읽힌다" → "**항목 스키마는 정확히 5필드로 남는다**" 로 바꿔, AC 가 실제로 증명하는 것에 이름을 맞췄다.
- AC 본문에 "**왜 키 집합을 직접 세는가**" 절을 넣어 Go 의 `encoding/json` 이 미지 필드를 조용히 무시한다는 사실과, `DisallowUnknownFields()` 를 **항목 단위로** 걸면 동등하지만 **최상위 레코드에 걸면 가산적 `findings` 때문에 올바른 구현에서 붉어진다**는 함정을 함께 적었다. 지시문은 두 형태를 택일로 제시했으나, 최상위 적용은 D1 과 같은 "올바른 구현에서 실패하는 기준" 이 되므로 경고를 명시하는 편이 안전하다.
- 착지: `acceptance.md` AC-TA-012 전체

### D10 (minor severity, blocking class) — 운영자의 완화안을 인용 없이 반박했다

- **바꾼 것**: `§B.1` 에 카드 t119 본문을 **원문 그대로** 인용하는 블록을 넣었다:
  > 오판 1건이 카드를 조용히 삼킬 수 있으므로 드롭은 되돌릴 수 있어야 한다(undrop 동사 존재).

  (출처 재확인: `moai todo list --json` 읽기 전용 실행으로 t119 본문에서 그대로 추출)
- 이어 한 문단으로 반박했다 — 완화안은 위험을 **정확히 지목**하며, 갈리는 지점은 식별이 아니라 **발동 조건**이다. `undrop` 은 운영자가 "있어야 할 카드가 없다"고 먼저 알아채야 손이 가고, 흡수·재정렬은 하필 그 알아챔을 없애는 변형이다. 즉 완화안이 요구하는 전제를 변형 자체가 무너뜨린다. 완화안이 **틀린 것이 아니라 이 두 동작에만 닿지 않는다**는 형태로 썼고, `undrop` 존치를 명시해 운영자의 결정 중 살아남는 부분을 지면에 남겼다.
- 착지: `spec.md §B.1` (인용 블록 + 반박 문단, 기준 선언 앞)

### D8 (optional) — "바이트 단위로 동일하다" 가 실제보다 강하다

- **바꾼 것**: 표의 순서 정리 행을 "재정렬된 큐는 … 바이트 단위로 동일하다" → "**운영자의 기본 화면에서 구별되지 않는다**" 로 바꾸고, 본문에 정확한 사정을 적었다 — 파일에는 단조 증가 `t<N>` id 와 `added_at` 이 남으므로 **재정렬이 일어났다는 사실 자체는 기계적으로 탐지된다**; 탐지되지 않는 것은 *누가·왜* 이고, 운영자의 평상시 화면에는 애초에 흔적이 안 뜬다.
- 재측정으로 근거를 보강: `internal/cli/todo.go:309` 의 사람용 렌더는 `id`·`state`·`text` **3열**뿐이며 `AddedAt` 은 `:271` 생성 시점에만 등장한다.
- **결론은 그대로다** — 순서 정리 기각은 정정 후에도 성립한다.
- 착지: `spec.md §B.1` (표 1행 + 본문 문단)

### D9 (optional) — 기준에 범위 한정어가 없어 기존 동작까지 건다

- **바꾼 것**: 기준 문장을 "**운영자가 저작한 카드의 내용·순서·존재**에 대한 자동 변경은 …" 으로 한정하고, 바로 아래에 한정의 이유를 적었다 — `normalizeBacklogRecord`(`internal/kanban/backlog_store.go:326`, `last_seq` 인상은 `:333-335`)의 내부 불변식 복구는 카드의 내용·순서·존재를 바꾸지 않으므로 대상이 아니다.
- 착지: `spec.md §B.1` (기준 인용 블록 + 후속 문단)

### D11 (optional) — 인용 줄번호 2건

- `todo_edit_move.go:17` → **`:5-7`**. 감사는 `:5` 를 제안했으나 재측정 결과 문장이 5–7행에 걸쳐 있어 범위로 적었다.
- `todo_edit_move.go:110` → **`:99`** (감사와 일치).
- 인용 **문구 자체는 두 건 모두 변경 없음**.
- 착지: `spec.md §A.2`, `§B.1`

### D12 (optional, 판단 위임 — 채택) — plan.md M2 의 참조 갱신 계획

- **바꾼 것**: "`spec.md §E` 참조를 이 SPEC으로 갱신만 한다" 를 삭제하고, **그대로 둔다**로 뒤집었다. 재측정 확인: `backlog_store.go:46` 의 `(spec.md §E out-of-scope)` 는 `SPEC-KANBAN-TODO-CLI-001/spec.md:92` "No version bump and no new per-item fields" 를 가리키는 정확한 인용이다. 이 SPEC 은 per-item 필드를 추가하지 않으므로 문장도 참조도 여전히 참이고, 돌리면 근거 없는 곳을 가리키게 된다. 해당 주석 변경이 M2 범위 밖임을 명시했다.
- 착지: `plan.md §E M2`

### D13 (optional, 판단 위임 — 채택) — `same pair` 방향성 미정의

- **REQ 쪽**: `REQ-TA-013` 에 "pairs are compared **unordered**, so `{a, b}` and `{b, a}` are the same pair" 삽입.
- **AC 쪽**: 신규 AC 없이 `AC-TA-011` ②의 방향을 뒤집었다 — `relate t5 t1` → **`relate t1 t5`** (기계 소견은 `{subject: t5, related: t1}`). 이 한 번의 방향 교환으로 ②가 `unreviewed` 해제와 무순서 비교를 **동시에** 검증하고, "쌍을 순서 있게 비교하는 구현" 이 붉게 만드는 구현 목록에 추가된다. AC 예산 소모 0.
- 착지: `spec.md` REQ-TA-013 · `acceptance.md` AC-TA-011

---

## 2. 부수 개정 (결함 목록 밖, 정합성 유지용)

| 변경 | 이유 | 위치 |
|---|---|---|
| frontmatter `version: "0.1.0"` → `"0.2.0"`, `updated: 2026-08-22` → `2026-08-23` | 본문 실질 개정 반영 | `spec.md` frontmatter |
| `§F 이력` 에 0.2.0 행 추가 | 결함 반영 이력 기록 | `spec.md §F` |
| `§A` 의 "AC 상한 16, 실제 15" → "실제 16" | AC-TA-016 신설 반영 | `acceptance.md §A` |
| `§E DoD` 의 "AC-TA-001 ~ AC-TA-015" → "~ AC-TA-016" | 동일 | `acceptance.md §E` |
| `§B` 에 plan-audit 의 Tier M 유지 판정 기록 | `plan.md §B` 가 감사 판정을 명시적으로 요청했고 답이 나왔다 | `plan.md §B` |
| M3 산출에 AC-TA-016 추가 + `analyze` 멱등성 한 줄 | D4 반영 | `plan.md §E M3` |
| M4 의 `todo_relate.go` 서술에 `unrelate` 명시 + 가드 범위 확장 항목 추가 | D1·D5 반영 | `plan.md §E M4` |
| `runTodoList` 항목에 "쌍은 무순서 비교" 명시 | D13 반영 | `plan.md §E M4` |

---

## 3. 검증 (실행한 것 / 실행하지 않은 것)

**실행한 것** — 감사가 인용한 수치는 전부 트리에서 재측정했다:

| 확인 | 명령 | 결과 |
|---|---|---|
| D1 grep 실측 | `grep -rn "AskUserQuestion" internal/cli/ \| wc -l` / `grep -rln … \| wc -l` | **317** / **82** — 감사와 일치 |
| 정본 관용구 존재 | `grep -rn "todoPromptGuard\|NoAskUserQuestion" internal/cli/*.go` | `todo_test.go` 에 테스트·가드·음성 대조 모두 실재 |
| D6 원문 | `grep -n "exit 0/1/2" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md` | `:63` — 감사와 일치 |
| D11 위치 | `cat -n internal/cli/todo_edit_move.go` | `:5-7` / `:99` |
| D9 위치 | `cat -n internal/kanban/backlog_store.go` (315-360) | `normalizeBacklogRecord` `:326`, `last_seq` 인상 `:333-335` |
| D8 렌더 열 | `grep -n "AddedAt\|Fprintf(out" internal/cli/todo.go` | 렌더 `:309` = `id\tstate\ttext`, `AddedAt` 은 `:271` 뿐 |
| D12 참조 대상 | `grep -n "per-item" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md` | `:92` "No version bump and no new per-item fields" |
| D10 원문 | `moai todo list --json` (읽기 전용) | 카드 t119 본문에서 완화 문장 그대로 추출 |
| 개정 후 REQ/AC 수 | `grep -c '^- \*\*REQ-TA-' spec.md` / `grep -c '^### AC-TA-' acceptance.md` | **15** / **16** (Tier M 상한 각 16) |
| 개정 후 REQ 커버리지 | `grep -o 'REQ-TA-[0-9][0-9][0-9]' acceptance.md \| sort \| uniq -c` | 001~015 **전량 ≥ 1** |
| Out of Scope 절 보존 | `grep -c '^### Out of Scope — ' spec.md` | **6** (불변) |
| 폐기 기준 잔존 | `grep -rn 'exit 0/1' .moai/specs/SPEC-TODO-ANALYSIS-001/` | `0/1` 단독 잔존 0건 (`0/1/2` 만) |

**실행하지 않은 것 (Gaps)**:

- `go test` 를 **한 건도 돌리지 않았다**. plan-phase 라 구현이 없고, 로컬 전체·대형 스위트 금지 규율(`internal/cli` 단독 ~336초)을 따랐다. 이 보고서는 구현 동작에 대한 어떤 주장도 하지 않는다 — 코드 관련 주장은 전부 정적 읽기와 grep 에 귀속된다.
- `moai todo` 쓰기 동사(`add`/`edit`/`move`/`drop`/`done`/`relate`)를 실행하지 않았다. 실제 운영자 큐는 읽기 전용(`list --json`)으로만 접근했다.
- 커밋·푸시하지 않았고, `internal/` 과 `.claude/` 를 건드리지 않았다.

**잔여 위험**:

- 새로 넣은 AC-TA-016(`analyze` 멱등성)과 AC-TA-004 의 `unrelate` 절이 **실제로 붉어지는지**는 run-phase 전까지 알 수 없다. 기준의 형태만 정했고 동작은 미검증이다.
- AC-TA-003 의 Jaccard 0.833 은 임계 0.80 대비 여유 **0.033** 뿐이다(감사 산술, 이번 개정에서 픽스처 변경 없음). 정규화 단계에 불용어 제거 같은 변형이 들어가면 이 픽스처가 임계를 넘나들 수 있다.
- D2 의 스크립트·파이프 경로는 **해소가 아니라 선언**이다. 감사가 허용한 두 선택지 중 잔여 위험 선언을 택했으므로, 재감사가 REQ 신설을 요구하면 이 지점이 다시 열린다.
- REQ 15 / AC 16 은 Tier M 상한 16 에 AC 여유가 **0** 이다. iteration 2 이후 새 AC 요구가 나오면 기존 AC 병합 또는 Tier L 승격이 필요하다.

---

## 4. 변경 파일

| 파일 | 성격 |
|---|---|
| `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` | §A.2 인용 정정, §B.1 전면 개정(인용·반박·범위 한정·조건화·잔여 위험·정정), REQ-TA-002 / 013 / 015 개정, frontmatter version·updated, §F 이력 |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/acceptance.md` | §A 규율 추가, 16 AC 전부에 REQ 태그, AC-TA-004 / 011 / 012 / 013 개정, AC-TA-016 신설, §E DoD |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/plan.md` | §B Tier 판정 기록, §D.3 완화 미도달 경로, §E M2 참조 갱신 항목 삭제, M3 / M4 산출·서술 |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/progress.md` | §E.1 에 iteration 2 신호(결함별 해소 표 · 재측정 수치 대조 표 · Gaps). §E.2~§E.4 는 placeholder 그대로 |
