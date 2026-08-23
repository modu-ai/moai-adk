# SPEC-TODO-ANALYSIS-001 — 구현 계획

## §A 맥락

`moai todo` 큐에 분석 계층을 넣는다. 코드 변경보다 **독트린 결정**이 먼저이고, 그 결정이 뒤집히면 아래 마일스톤 대부분이 무의미해진다. 그래서 마일스톤은 되돌리기 어려운 순서, 즉 **바뀔 가능성이 큰 결정부터** 배치했다.

## §B Tier 판정

**Tier M.**

| 축 | 측정 | Tier M 범위 |
|---|---|---|
| 영향 파일 | 12~14 (Go 6~8, 독트린 2, 템플릿 미러 2, 테스트 포함) | 5~15 |
| LOC | 500~800 추정 (분석기 ~150, 스키마 ~80, 신규 동사 4개 ~250, 테스트 ~300) | 300~1000 |
| 산출물 | spec.md + plan.md + acceptance.md | 3 files |
| plan-audit PASS 임계 | 0.80 | |

**Tier L이 아닌 이유**: 파일 수와 LOC 모두 M 범위 안이고, 새 서브시스템이 아니라 기존 `internal/cli/todo*.go` + `internal/kanban/backlog_store.go` 의 확장이다. 락 substrate, 원자적 쓰기, id 발행, 테스트 하네스가 전부 이미 있다.

**M의 상단이라는 단서**: 독트린 [HARD] 절 개정이 들어 있어 "constitutional" 판정으로 L에 걸릴 여지가 있다. 그러나 이 SPEC의 개정은 금지를 **완화하는 것이 아니라 무엇이 허용되는지 이름 붙여 좁히는 것**이므로(§F M1), 헌법적 재설계가 아니라 기존 규칙의 명세화로 본다. plan-audit 이 이 판단에 동의하지 않으면 Tier L 승격(+ design.md, research.md)이 정당한 지적이다.

**plan-audit iteration 1 판정: Tier M 유지.** 근거 — 개정 대상이 `moai-constitution.md` 가 아니라 워크플로 규칙 1 + 스킬 1이고, 방식이 기존 [HARD] 문장을 한 글자도 지우지 않는 추가-only 이며(AC-TA-014 가 잔존을 주장), 영향 파일·LOC 두 축 모두 M 범위 안이다.

## §C 이미 있는 것 / 새로 만드는 것

| 있는 것 | 위치 | 재사용 방식 |
|---|---|---|
| 락 기반 read-modify-write | `internal/kanban/backlog_store.go` `Mutate` | 분석은 append와 **같은** `Mutate` 콜백 안에서 돈다 |
| 거절 시 파일 불변 | `Mutate` 콜백에서 error 반환 | 정확 중복 거절이 그대로 이 계약을 쓴다 |
| 가산적 최상위 필드 선례 | `last_seq` | `findings` 배열이 같은 방식 |
| `--expect` 오지목 방어 | `next`/`edit`/`drop` | `relate` 도 같은 형태 |
| 텍스트 마커 기록 선례 | `drop` 의 `[DROPPED — …] ` | 스키마 없이 기록하는 대안 — 채택하지 않음(§D) |

## §D 되돌리기 어려운 결정 3개 (먼저 검토받아야 할 것)

### D.1 소견을 어디에 저장할 것인가

- **채택**: 최상위 `findings` 배열 (가산적, `last_seq` 선례).
- **기각 1 — 항목당 새 필드**: REQ-TODO-013 의 5필드 계약을 깬다. 구 로더 호환이 무너진다.
- **기각 2 — 텍스트 마커** (`drop` 방식): 관계는 **쌍**의 속성이라 한 카드의 텍스트에 넣으면 반대쪽에서 안 보이고, 카드 텍스트를 오염시켜 `--expect` 접두 비교를 망가뜨린다.
- **기각 3 — 별도 파일**: 두 파일의 원자성을 맞춰야 한다. 하나의 `Mutate` 안에서 끝나지 않는다.

### D.2 임계값 0.80 (token-set Jaccard)

되돌리기 쉬운 축이지만 **틀리면 손해가 비대칭**이다. 너무 낮으면 소견 소음이 `list` 를 덮어 운영자가 전부 무시하게 되고, 그러면 이 기능은 있으나 마나가 된다. 너무 높으면 놓치지만, 놓친 중복은 지금과 같은 상태일 뿐이다.

**따라서 높은 쪽으로 틀리는 것을 기본으로 한다.** 0.80 은 초기값이며, 실제 큐(현재 카드 수십 장)에 `analyze` 를 돌려 소견 건수를 재고 조정한다. 소견 수가 카드 수의 20%를 넘으면 임계값이 낮은 것으로 본다.

### D.3 정확 중복을 "거절"할 것인가 "허용하고 기록"할 것인가

- **채택: 거절 + `--force` 탈출구.** 거절은 아무것도 파괴하지 않고(큐 불변), 운영자에게 즉시 보인다(id 대신 에러). §B.1 의 현저성 기준을 통과하는 유일한 변형이다.
- **위험**: 자동화 스크립트가 add 실패를 무시하면 카드가 조용히 유실된다. 완화 — exit code ≠ 0 + stderr 에 충돌 id 명시. `manager-lead` 는 add 결과를 읽으므로(카드 id를 받아야 디스패치가 되므로) 실패를 놓칠 수 없다.
- **완화가 닿지 않는 경로**: 위 완화는 L2(에이전트) 경로에서만 성립한다. `spec.md §B.3` 은 에이전트 없는 순수 CLI 실행에서도 L1 이 작동할 것을 요구하고 `AC-TA-015` 로 강제하는데, 그 경로의 호출자(스크립트·파이프)가 exit code 와 stderr 를 둘 다 버리면 완화가 발동하지 않는다. CLI 는 호출자가 자기 종료 코드를 읽게 만들 수 없으므로 이 SPEC 은 보완 장치를 만들지 않고 `spec.md §B.1` 에 **명시적 잔여 위험**으로 선언했다. §B.1 표의 "정확 중복 거절" 행이 무조건 "현저: 예"가 아니라 "호출자가 exit code 또는 stderr 를 읽을 때에만"으로 조건화된 이유가 이것이다.

## §E 마일스톤 (되돌리기 어려운 순서)

### M1 — 독트린 개정 (우선순위: High)

가장 먼저 하고, 가장 먼저 리뷰받는다. 여기서 뒤집히면 M2~M4가 통째로 무효다.

대상 4파일 (로컬 + `internal/template/templates/` 미러):

- `.claude/skills/moai/workflows/todo.md`
- `.claude/rules/moai/workflow/kanban-dispatch.md`

**개정 방향 — 금지를 흐리지 않고 좁힌다.** 기존 [HARD] 문장은 한 글자도 지우지 않는다. 그 아래에 경계 절을 **추가**한다:

> 분석은 자동으로 돌고 **기록한다**. 기록은 카드를 바꾸지 않는다.
> 분석이 일으키는 변형은 정확히 하나 — 정규화 후 텍스트가 동일한 카드의 **입력 거절**이며, 이는 기존 카드를 건드리지 않고 큐 파일을 바이트 단위로 그대로 둔다.
> 분석은 카드를 접어 넣거나, 재정렬하거나, 버리거나, 고치지 않는다. `contains` / `absorbs` / `replaces` / `conflicts` 는 기록 외에 아무 일도 일으키지 않는다.

`kanban-dispatch.md` 에는 리드의 의무 한 줄을 더한다: 리드는 소견을 **붙일 수 있고, 그것에 따라 행동해서는 안 된다**.

산출: AC-TA-014.

**주의 — 레인 충돌**: 다른 레인(t170, TODO-ENABLE-FLAG)이 같은 `workflows/todo.md` 를 건드릴 수 있다. 나중에 착지하는 쪽이 충돌을 소유한다. 이 SPEC의 개정은 기존 문장을 **추가만** 하므로 텍스트 충돌 표면이 작다.

### M2 — 소견 레코드 스키마 (우선순위: High)

- `internal/kanban/backlog_store.go`: `BacklogFinding` 타입 + `BacklogRecord.Findings []BacklogFinding` (가산적, `omitempty` 아님 — 항상 배열로 렌더).
- `normalizeBacklogRecord` 에 nil → 빈 슬라이스 정규화 추가.
- `done` 경로에 소견 회수(REQ-TA-007).
- 구 스키마 왕복 테스트 (AC-TA-012).

`backlog_store.go:46` 주석의 "no per-item field may ever be added (spec.md §E out-of-scope)" 는 **문장도 참조도 그대로 둔다**. 문장은 여전히 참이고(이 변경은 최상위 필드다), 그 `spec.md §E` 는 SPEC-KANBAN-TODO-CLI-001 `spec.md:92` 의 "No version bump and no new per-item fields" 를 가리키는 정확한 인용이다 — 이 SPEC 으로 돌리면 근거가 없는 곳을 가리키게 된다. 이 파일의 주석 변경은 M2 범위 밖이다.

산출: AC-TA-008, AC-TA-012.

### M3 — 기계 분석기 + `add` 통합 (우선순위: High)

- 새 파일 `internal/kanban/backlog_analysis.go`: 정규화(NFC → trim → 공백 축약 → case-fold), 토큰화, Jaccard, 분류(`exact`/`near`/`none`). 순수 함수, 저장소 비의존 → 테이블 테스트로 단독 검증.
- `internal/cli/todo.go` `runTodoAddAppend` / `runTodoAddPick`: `Mutate` 콜백 안에서 분석 → `exact` 면 error 반환(파일 불변), `near` 면 append + 소견.
- `--force` 플래그.
- 새 동사 `moai todo analyze` — 전체 큐 재분석, 기록만.

산출: AC-TA-001 ~ AC-TA-003, AC-TA-006, AC-TA-015, AC-TA-016.

`analyze` 는 소견을 **재실행해도 다시 쌓지 않는다**(REQ-TA-002) — 같은 `{subject_id, related_id, relation, source}` 튜플은 두 번 append 되지 않는다. AC-TA-016 이 이것을 잡는다.

### M4 — 에이전트 계층 동사 + 운영자 가시성 (우선순위: Medium)

- 새 파일 `internal/cli/todo_relate.go`: `relate` / `unrelate`. `relate` 는 두 id 존재 확인 + 소견 append 만, `unrelate` 는 지목된 소견 1건 remove 만; 카드 필드 쓰기 경로가 **코드에 존재하지 않는다** (AC-TA-004 가 두 동사를 한 픽스처에서 양방향으로 잡고, AC-TA-005 가 네 관계 전부를 잡는다).
- `internal/cli/todo_why.go` 또는 기존 파일 확장: `why <n>`.
- `runTodoList` 에 소견 줄 렌더 + `machine-only` 표시(쌍은 **무순서** 비교 — REQ-TA-013) + `--json` 에 `findings` 포함.
- `internal/cli/todo_test.go` 의 기존 `todoPromptGuard` 스캔 대상을 `todo.go` 하나에서 **`internal/cli/todo*.go` 비테스트 파일 전부**로 넓힌다(음성 대조는 그대로). AC-TA-013 이 요구하는 것은 가드의 신설이 아니라 범위 확장이다.

산출: AC-TA-004, AC-TA-005, AC-TA-007, AC-TA-009 ~ AC-TA-011, AC-TA-013.

### M5 — 템플릿 미러 + 빌드 (우선순위: Medium)

`make build` (임베드 재생성) → 미러 parity 확인 → 대상 패키지 테스트.

## §F 기술적 접근의 핵심

### F.1 분석은 append와 같은 락 안에서 돈다

분석을 `Mutate` **밖**에서 하면 분석 시점과 append 시점 사이에 다른 세션의 add 가 끼어들어 중복이 통과한다 — SPEC-KANBAN-TODO-CLI-001 이 죽인 read-modify-write 레이스의 재발이다. 분석·판정·append 가 한 콜백 안에서 끝나야 한다.

### F.2 거절 경로는 기존 계약을 그대로 쓴다

`Mutate` 콜백이 error 를 반환하면 아무것도 쓰이지 않는다. `edit`/`drop`/`next --expect` 가 이미 이 계약 위에 서 있고, 정확 중복 거절도 같은 자리에 얹힌다. 새 불변식을 발명하지 않는다.

### F.3 `relate` 에는 카드를 쓰는 코드가 없다

"관계에 따라 행동하지 않는다"를 문서가 아니라 **코드 구조**로 강제한다. `todo_relate.go` 의 `Mutate` 콜백은 `rec.Findings` 에만 append 하고 `rec.Items` 를 읽기 전용으로만 참조한다(존재 확인). 흡수를 하려면 새 코드를 써야 한다 — 실수로는 못 한다.

## §G 위험

| 위험 | 영향 | 완화 |
|---|---|---|
| plan-audit 이 독트린 개정을 거부 | M2~M5 전부 무효 | M1을 먼저 내고 단독 리뷰. 기존 [HARD] 문장 보존이 협상 카드 |
| 임계값 오설정으로 소견 소음 | 운영자가 소견을 전부 무시 → 기능 무력화 | D.2 — 높은 쪽으로 기본, 실제 큐에서 재측정 |
| t170 레인과 `todo.md` 충돌 | 머지 충돌 | 추가-only 개정, 나중에 착지하는 쪽이 소유 |
| 자동화가 add 실패를 무시해 카드 유실 | 운영자 의도 손실 | exit ≠ 0 + stderr 충돌 id. 리드는 add 의 id 출력을 읽어야 디스패치가 되므로 실패를 놓칠 수 없다 |
| 테스트가 실제 운영자 큐를 건드림 | 라이브 카드 손상 (오늘 실제로 1건 발생) | 전 테스트 `t.TempDir()` + 작업 전후 실제 큐 `sha256` 대조 (DoD) |
| `internal/cli` 스위트 ~336초 | 검증 지연 | `-timeout 900s`, 대상 패키지만. 전체 스위트는 CI 몫 |

## §H 상호참조

- `spec.md §B` — 가역성 + 현저성 판정 기준, 이 계획의 순서를 결정한 논거.
- `acceptance.md` — 각 AC 를 붉게 만드는 잘못된 구현.
- SPEC-KANBAN-TODO-CLI-001 — 락 substrate, 5필드 계약, headless 계약.
- `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — 템플릿 중립성 콘텐츠 클래스.
