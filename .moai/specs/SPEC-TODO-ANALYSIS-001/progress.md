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

M1(독트린)·M2(스키마)는 `e04801047`·`2c63f2ac1` 로 이미 착지했다. 이번 run 은 M3(기계 분석기 + `add` 통합)·M4(에이전트 동사 + 가시성)·M5(스킬 표면 갱신 + 미러 + `make build`) 를 냈다.

### 착지한 코드

| 마일스톤 | 파일 | 내용 |
|---|---|---|
| M3 | `internal/kanban/backlog_analysis.go` (신규) | `NormalizeCardText`(NFC→trim→공백축약→case-fold) · `TokenSetJaccard` · `ClassifyCardText` · `BacklogNearDuplicateThreshold = 0.80`. 저장소 비의존 순수 함수 |
| M3 | `internal/cli/todo_analysis.go` (신규) | `appendAnalyzedCard` — 분석·판정·append 가 **한 `Mutate` 콜백 안**. 거절은 콜백 error 반환(파일 불변). `analyze` 동사. `todoFindingLine` |
| M3 | `internal/cli/todo.go` | `add --force`, `add`/`add --pick` 를 분석 경유 append 로 교체, `list` 에 소견 줄 렌더, 신규 4동사 등록 |
| M4 | `internal/cli/todo_relate.go` (신규) | `relate` / `unrelate`. 콜백이 `rec.Items` 를 **존재 확인용 읽기로만** 참조 — 카드 쓰기 경로가 코드에 없다 |
| M4 | `internal/cli/todo_why.go` (신규) | `why <n>`, 무소견 시 명시 문구 |
| M4 | `internal/cli/todo_test.go` | `todoPromptGuard` 스캔 범위를 `todo.go` 하나 → `todo*.go` 비테스트 전부(glob). 스캔 파일 수 ≥ 2 양성 대조 추가 |
| M5 | `.claude/skills/moai/workflows/todo.md` + 템플릿 미러 | 명령 표에 `--force`/`analyze`/`relate`/`unrelate`/`why` 5행, `--json` 레코드 예시에 `findings` 배열 + 필드 설명 |

### 실행한 명령과 관측 출력

| 명령 | 관측 |
|---|---|
| `go test ./internal/cli/ ./internal/kanban/ -count=1 -timeout 900s` | `ok internal/cli 569.642s` / `ok internal/kanban 28.490s`, `TEST_EXIT=0` |
| `go vet ./...` | `VET_EXIT=0` |
| `GOOS=windows go vet ./internal/cli/... ./internal/kanban/...` | `WINVET_EXIT=0` (컴파일만 증명 — Windows 동작 근거 아님) |
| `golangci-lint run ./internal/cli/... ./internal/kanban/...` | `0 issues.`, `LINT_EXIT=0` |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` | `ok internal/template 40.827s` |
| `make build` | `catalog.yaml updated successfully (12899 bytes)` → `bin/moai` |
| `diff` 미러 대조 | `todo.md` 쌍 · `kanban-dispatch.md` 쌍 각각 바이트 동일 |

증거 파일(추적됨): `.moai/reports/t119/run-verify-batch.txt`, `.moai/reports/t119/run-coverage-new-files.txt`. (`.moai/state/` 는 gitignore 대상이라 감사 시점에 해소되지 않는다 — 그래서 리포트로 옮겼다.)

### 커버리지 (신규 코드)

`go tool cover -func` 함수 평균 — 신규 CLI 3파일 **92.3%** (11 함수), 신규 kanban 1파일 **97.7%** (4 함수). 둘 다 §D 의 85% 문턱 위.

### 빌드된 바이너리로 확인한 실제 동작

임시 git 저장소(`mktemp -d` + `CLAUDE_PROJECT_DIR`)에서 `bin/moai` 로 직접 실행해 관측한 것:

- 정확 중복 `add` → exit 1, stderr 가 `t1 already holds this card ("Rework the auth middleware error paths")`.
- `list` 가 카드 아래 들여쓴 소견 줄을 렌더: `↳ near-duplicate t2 (mechanical, score 0.83, machine-only) — moai todo drop t1 | moai todo edit t1 "<text>"`.
- `relate t1 t2 --relation replaces` 후 같은 줄에서 `machine-only` 가 사라지고, `unrelate 2` 후 다시 붙는다(무순서 쌍 비교).
- `why t3`(소견 없음) → `t3: no findings`.
- `analyze` 연속 2회 → 둘 다 `analyzed 3 pairs, recorded 0 findings` (이미 `add` 가 기록한 소견이 튜플 중복으로 걸러짐).

### 계획 대비 벗어난 점

- **점수 렌더 조건화** (계획에 없던 판단): 에이전트 소견은 측정값이 없는데 `score 0.00` 으로 렌더돼, 측정된 비유사도처럼 읽혔다. `source == mechanical` 일 때만 점수를 찍도록 좁혔다.
- **`add --pick` 도 분석을 탄다**: 계획 §E M3 는 `runTodoAddAppend` / `runTodoAddPick` 를 함께 적었고, 그대로 두 경로 모두 `appendAnalyzedCard` 를 쓴다. `--pick` 전용 AC 는 없다(미검증 아님 — 같은 함수를 공유하므로 `add` 의 AC 가 그 경로의 판정 로직을 덮지만, `--pick` **호출 경로 자체**의 회귀 테스트는 없다).
- **M5 의 "미러 parity" 는 M1 이 이미 만족**했고, 이번에 추가한 스킬 표면 갱신분도 같은 파일 쌍에 함께 복사했다. Go 코드는 템플릿 미러 대상이 아니다.

## §E.3 Run-phase Audit-Ready Signal

### AC PASS/FAIL 행렬

각 PASS 는 위 전량 통과(`TEST_EXIT=0`)에 귀속된다. 테스트가 없는 두 건(AC-TA-014, AC-TA-013 의 `--help` 부분은 테스트가 있고 미러 부분은 `diff`)은 명령 출력에 귀속했다.

| AC | 판정 | 귀속 |
|---|---|---|
| AC-TA-001 정확 중복 거절 + 파일 불변 | PASS | `TestTodoAddRefusesExactDuplicate` (sha256 전후 대조) |
| AC-TA-002 `--force` 기록 | PASS | `TestTodoAddForceAdmitsAndRecords` |
| AC-TA-003 근접 중복 양성+음성 | PASS | `TestTodoAddNearDuplicateRecordsOnly` |
| AC-TA-004 `relate`/`unrelate` 양방향 | PASS | `TestTodoRelateAndUnrelateTouchNoCard` (items 바이트 동일 2회, 생존 소견 지목) |
| AC-TA-005 네 관계가 큐를 안 바꿈 | PASS | `TestSemanticRelationsChangeNothing` (`relate` 없이는 Given 구성 불가) |
| AC-TA-006 순서 불변 | PASS | `TestTodoAnalysisNeverReordersQueue` |
| AC-TA-007 읽기 멱등 | PASS | `TestTodoListJSONIsIdempotent` |
| AC-TA-008 `done` 소견 회수 | PASS | `TestTodoDoneReclaimsFindings` (M2 착지분) |
| AC-TA-009 `list` 소견 + 다음 행동 | PASS | `TestTodoListShowsFindingAndNextStep` |
| AC-TA-010 무소견 명시 | PASS | `TestTodoWhySaysNothingFound` |
| AC-TA-011 `machine-only` 양방향 | PASS | `TestMachineOnlyMarkAppearsAndClears` (역방향 `relate` 로 무순서 비교 검증) |
| AC-TA-012 구 스키마 왕복 + 5필드 | PASS | `TestTodoLegacyRecordRoundTrips` (M2 착지분) |
| AC-TA-013 신규 동사 headless | PASS | `TestTodoNewVerbsAreHeadless` (stdin 닫고 6동사, `--help` 4동사 양성 대조) + `TestTodoCmd_NoAskUserQuestion` (glob 범위 + 합성 위반 음성 대조) |
| AC-TA-014 독트린 미러 | PASS | `diff -q` 두 쌍 무출력; 경계 절 grep(`records\|never folds`) 양쪽 12히트; 기존 [HARD] 금지 4구절(inferred priority / tidy-up / fold one card into another / looks stale) 양쪽 전부 잔존 — `fold one card into another` 는 줄바꿈에 걸쳐 있어 `never to` 를 뺀 구절로 대조 |
| AC-TA-015 에이전트 없이 작동 | PASS | `TestTodoExactRefusalWorksWithoutAgent` |
| AC-TA-016 `analyze` 멱등 | PASS | `TestTodoAnalyzeRerunIsIdempotent` (첫 실행 길이 > 0 양성 대조 포함) |

### 미검증 (Gaps)

- **실제 운영자 큐의 sha256 이 작업 전후로 달라졌다.** 전 `f42517a2…` → 후 `e5a2ca84…`. 이 SPEC 의 코드가 건드린 것이 아님을 두 가지로 확인했다: 파일에 `findings` 키가 없고(이 기능이 한 번도 쓴 적 없음), 늘어난 항목은 리드가 이번 세션에 추가한 `t177`·`t178` 이다. 그래도 **"작업 전후 동일"이라는 DoD 문구 자체는 만족하지 못했다** — 다른 세션이 같은 큐에 쓰는 팩토리 환경에서 그 문구는 원래 성립할 수 없다. 이 SPEC 이 큐를 손상시키지 않았다는 주장은 위 두 관측에 귀속되며, sha256 동일성에는 귀속되지 않는다.
- **`add --pick` 경로의 회귀 테스트 없음** — 판정 로직은 `add` 와 같은 함수를 공유하지만 그 호출 경로를 지나는 테스트가 없다.
- **임계값 0.80 을 실제 큐에서 재측정하지 않았다.** `plan.md §D.2` 는 실제 큐에 `analyze` 를 돌려 소견 건수가 카드 수의 20% 를 넘는지 보라고 했는데, 운영자 큐에서 `analyze` 를 돌리는 것은 라이브 큐 쓰기라 하지 않았다. 임계값은 초기값 그대로다.
- **Windows 동작 미검증** — `GOOS=windows go vet` 은 컴파일만 증명한다.
- **전체 스위트(`go test ./...`) 미실행** — 로컬 금지 규율에 따라 대상 2패키지 + `internal/template` 만 돌렸다. 전 패키지 판정은 CI 몫이다.

### 잔여 위험

- `AC-TA-003` 픽스처의 Jaccard 0.833 은 임계 0.80 대비 여유 0.033 이다. 정규화에 다섯 번째 단계가 들어가면 이 픽스처가 임계를 넘나든다 — `TestNormalizeCardText` 가 파이프라인 길이를 못박아 그 변경을 붉게 만든다.
- `analyze` 는 O(n²) 쌍 비교다. 현재 큐 규모(수십 장)에서는 무시할 수 있으나 수천 장에서는 다르다.
- 소견은 카드가 `done` 될 때만 회수된다. `drop` 된 카드의 소견은 남는다(카드가 파일에 남으므로 가리키는 대상은 있다) — 의도한 동작이지만 소견 목록이 길어지는 경로다.

## §E.4 Sync-phase Audit-Ready Signal

### 착지한 문서

| 표면 | 내용 |
|---|---|
| `CHANGELOG.md` `[Unreleased]` | `### Added` 6불릿 — 분석이 기록이지 변형이 아니라는 점, 정확 중복 거절과 `--force`, 근접 중복이 의도적으로 텍스트만 본다는 한계, 네 관계 동사, `why` / `list` 가시성, 가산적 `findings` 배열 |
| `docs-site/content/{ko,en,ja,zh}/utility-commands/moai-todo.md` | 신규 `## 자동 분석` 절 + 상태 파일 JSON 에 `findings` + 필드 설명 행 + CLI 표 5행 + bash 예시 5개. ko 정본 → en/ja/zh 파생 |
| `.moai/specs/SPEC-TODO-ANALYSIS-001/spec.md` | frontmatter `status: draft → completed`, `version 0.2.0 → 1.0.0`, §F 이력 행 추가 |

### 실행한 명령과 관측 출력

| 명령 | 관측 |
|---|---|
| `hugo --gc --minify` (docs-site) | `Total in 2256 ms`, WARN 0줄, 4로케일 페이지 183/181/181/181, `public/index.html` 생성 |
| `wc -l` 4로케일 | 215 / 215 / 215 / 215 (동일) |
| `grep -c "^## "` 4로케일 | 9 / 9 / 9 / 9 (절 수 동일) |
| 이모지 스캔 (`\U0001F300-\U0001FAFF`) | 4로케일 모두 0 |
| mermaid 방향 스캔 (`flowchart LR` / `graph LR`) | 4로케일 모두 0 (TD-only 유지) |

### 3-phase close

- plan → `f1bc39310` (draft), run → `ebd828b5d`, sync → 이 커밋. `status: completed` 는 이 sync 커밋에 실린다.
- **run-phase 에서 `status: in-progress` 로의 중간 전이를 남기지 않았다.** 한 레인이 run 과 sync 를 연속으로 수행했고 그 사이에 커밋이 하나뿐이라, draft → completed 로 한 번에 넘어간다. 3-phase 계약의 문서 흔적이라는 면에서는 결손이며, 여기 기록해 둔다.

### 미검증 (Gaps)

- **번역 3로케일은 사람이 검수하지 않았다.** ko 정본을 기준으로 파생했고, 사실·수치·코드 블록·명령 문법은 기계적으로 동일하게 유지했지만 문장의 자연스러움은 검증되지 않았다.
- **docs-site 링크 검사기를 돌리지 않았다.** 새 절은 외부 링크를 추가하지 않았고 기존 상호참조도 건드리지 않았으므로 링크 표면은 변하지 않았다 — `hugo` 빌드 무경고가 근거의 전부다.
- **배포 확인 없음** — Vercel 은 머지 후에 빌드한다. 로컬 `hugo` 빌드가 통과했다는 것뿐이다.
