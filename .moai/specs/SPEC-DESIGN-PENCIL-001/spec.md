---
id: SPEC-DESIGN-PENCIL-001
version: 0.2.0
status: draft
created_at: 2026-04-20
updated_at: 2026-04-21
author: moai-adk-go
priority: Medium
labels: [design, pencil, mcp, batch-design, wireframe, skill, workflow-extension]
issue_number: null
depends_on: [SPEC-DESIGN-DOCS-001, SPEC-DESIGN-ATTACH-001]
related_specs: []
---

# SPEC-DESIGN-PENCIL-001: moai-workflow-pencil-integration Skill + /moai design Phase B2.6 Pencil Path

## HISTORY

- 2026-04-21 v0.2.0: plan-auditor iteration 1 FAIL 후속 수정. Frontmatter 표준(version/created_at/labels 추가), REQ↔AC 1:1 traceability 확립(9개 AC 신규 추가), 각 REQ에 Rationale 부착, REQ-001/REQ-013 경계 분리(파일/폴더 vs MCP 런타임), AC-8 fallback 의미 명확화(Path B 내부 continuation), REQ-006/013/016 `then` 명시 추가, REQ-015 ↔ Error Code Table retry 용어 통일, DSL Grammar 서브섹션 추가, `<id>` 정의 및 AC-6 regex 강화, REQ-015에 `PENCIL_BATCH_FAILED` 반환 명시, AC-4 "documented" 기준 구체화.
- 2026-04-20 v0.1.0: SPEC 최초 작성.

## Background

### 사용자 요구사항

사용자는 `/moai design` 워크플로우가 `.moai/design/pencil-plan.md` 파일, `.pen` 파일, 그리고 Pencil MCP 서버 가용성을 자동으로 감지하여 batch operation을 자동 실행하기를 원한다. moai-studio 프로젝트에서는 이미 이 패턴을 사용 중이며, `pencil-redesign-plan.md`는 315줄 분량의 문서로 약 8개의 batch 작업을 기술하고 있다. 현재 `/moai design` 워크플로우에는 이 자동화가 통합되어 있지 않아, Pencil 기반 와이어프레임 작업이 수작업으로 이루어지고 있다.

### Pencil MCP 생태계

Pencil MCP는 deferred tool 목록에 포함되어 있어 ToolSearch를 통해 동적으로 로드해야 하는 지연 로딩 대상이다. 제공되는 주요 tool 목록은 다음과 같다:

- `mcp__pencil__batch_design` — design operation batch 실행 (I = insert, M = move, R = remove)
- `mcp__pencil__batch_get` — 속성 조회
- `mcp__pencil__get_editor_state` — Pencil 연결 상태 확인
- `mcp__pencil__get_screenshot` — 스크린샷 export
- `mcp__pencil__snapshot_layout` — layout 문제 스캔 (`problemsOnly: true` 모드 지원)
- `mcp__pencil__find_empty_space_on_canvas` — 배치 보조
- `mcp__pencil__open_document` — `.pen` 문서 열기
- 기타 style / variable 관련 tool

모든 `mcp__pencil__*` 호출 이전에 ToolSearch로 schema를 선행 로딩해야 한다.

### pencil-plan.md DSL Grammar

moai-studio는 `pencil-plan.md`에서 JavaScript 유사 DSL을 사용한다. 본 SPEC의 parser는 다음 3개 연산만을 acceptable 문법으로 간주한다:

```
I(parentId, { type: <"frame"|"text"|"rect"|"ellipse"|"line"|"group">, ...props })   // Insert
M(nodeId, parentId, index)                                                          // Move: 0-based index
R(nodeId)                                                                           // Remove
```

DSL 규칙:

- 각 연산은 개행(newline) 경계로 구분되며 한 줄에 하나의 연산만 존재
- `parentId`, `nodeId`는 double-quoted string literal
- `index`는 비음수 정수
- `{...props}`는 JSON-serializable object literal (Pencil의 `batch_design` schema에 따름)
- 주석은 `//` line comment 및 `/* ... */` block comment 허용
- Batch 경계는 `## Batch <N>` 형식의 Markdown heading으로 구분 (N은 1부터 시작하는 양의 정수)

이 문법 범위 밖의 토큰은 `PENCIL_PLAN_SYNTAX_ERROR`로 처리한다. 연산은 batch 단위로 구성되며, batch 1개당 최대 25개 operation을 포함한다. batch와 batch 사이에는 screenshot + layout scan으로 무결성을 검증한다.

### 기존 워크플로우 참조

`.claude/skills/moai/workflows/design.md`에는 이미 Phase B (code-based) 경로가 존재한다. 본 SPEC에서는 Phase B2.5 (SPEC-DESIGN-ATTACH-001에서 도입된 context loading) 직후, Phase B3 (BRIEF 생성) 이전에 Phase B2.6을 삽입한다. 이 Phase는 Pencil 전제조건이 충족될 때에만 활성화된다. Phase B2.6은 Path B 내부의 sub-phase이므로, 이 Phase 실패 시의 "fallback" 의미는 "Path B 경로 선택 재진입"이 아니라 "Phase B2.6의 나머지 작업을 건너뛰고 Phase B3로 진행"임을 명확히 한다.

## Requirements (EARS)

### REQ-PENCIL-001 (File/Folder Precondition Check)

`WHEN /moai design workflow reaches Phase B2.6, the system SHALL verify the following file/folder preconditions only: .moai/design/pencil-plan.md exists AND at least one .pen file exists in .moai/design/ or project root.`

**Rationale**: MCP 런타임 가용성은 REQ-PENCIL-013의 ToolSearch 단계에서만 판정하여 책임 경계를 명확히 분리한다. Phase B2.6 진입 여부는 파일/폴더 시그널만으로 충분하다.

### REQ-PENCIL-002 (Graceful Skip)

`WHEN any file/folder precondition in REQ-PENCIL-001 fails, the system SHALL skip Phase B2.6 gracefully without user-visible error and proceed to Phase B3.`

**Rationale**: Pencil 경로는 선택적 확장 기능이므로 파일 부재가 워크플로우 실패로 이어지면 안 된다. Quiet skip이 사용자 의도에 부합한다.

### REQ-PENCIL-003 (Skill Invocation)

`WHEN all file/folder preconditions are met, the system SHALL invoke moai-workflow-pencil-integration skill and wait for its completion (success or structured error) before proceeding to Phase B3.`

**Rationale**: Skill이 생성하는 와이어프레임 산출물은 Phase B3의 BRIEF 작성 시점에 참조 가능해야 하므로 동기 대기가 필수다.

### REQ-PENCIL-004 (ToolSearch Preloading)

`WHEN the skill starts, it SHALL invoke ToolSearch with query "mcp__pencil" to load Pencil MCP tool schemas before issuing any mcp__pencil__* call.`

**Rationale**: Pencil MCP는 deferred tool이므로 schema 선행 로드 없이 호출하면 tool dispatch가 실패한다.

### REQ-PENCIL-005 (Editor State Verification)

`WHEN Pencil tool schemas are loaded, the skill SHALL call mcp__pencil__get_editor_state and verify the response references the expected .pen file name detected in REQ-PENCIL-001.`

**Rationale**: 사용자가 의도한 `.pen` 문서와 Pencil 앱이 실제로 열고 있는 문서가 일치하는지 확인하지 않으면, 잘못된 캔버스에 변경이 적용되는 치명적 결과가 발생할 수 있다.

### REQ-PENCIL-006 (Connection Failure Handling)

`IF get_editor_state returns a different document or reports a connection failure, then the skill SHALL return PENCIL_CONNECTION_FAILED and guide the user to restart Claude Desktop or open the correct .pen file.`

**Rationale**: 연결 불일치는 사용자의 Pencil 앱 상태 이슈이므로 skill이 자가 복구를 시도하는 것은 부적절하다. 명확한 사용자 행동 지시가 최선이다.

### REQ-PENCIL-007 (Batch Parsing Order)

`WHEN the skill parses pencil-plan.md, it SHALL extract batch operations in declaration order (Batch 1 through Batch N) and preserve this order in execution.`

**Rationale**: Pencil 레이아웃은 이전 batch의 노드 ID에 의존하는 경우가 많아 순서 역전 시 reference error가 발생한다.

### REQ-PENCIL-008 (Batch Size Limit)

`WHEN a batch contains more than 25 operations, the skill SHALL split the batch into sub-batches of at most 25 operations each before calling batch_design, preserving original operation order.`

**Rationale**: Pencil MCP의 `batch_design` endpoint는 batch당 25 op 제한을 가지며 moai-studio 실측에서 동일 제약이 확인되었다.

### REQ-PENCIL-009 (Layout Verification)

`WHEN each batch completes, the skill SHALL call mcp__pencil__snapshot_layout with problemsOnly=true and attach any reported layout issues to the orchestrator-visible result.`

**Rationale**: 레이아웃 overlap/overflow 문제는 시각적으로만 드러나므로 자동 검증 없이는 downstream phase에서 뒤늦게 발견된다.

### REQ-PENCIL-010 (Halt on Layout Issue)

`WHEN a batch completes with layout issues reported by REQ-PENCIL-009, the skill SHALL NOT proceed to the next batch and SHALL return a structured error containing the problematic frame IDs.`

**Rationale**: 문제 있는 batch 위에 후속 batch를 쌓으면 오류가 복합되어 디버깅이 어려워진다. Fail-fast가 필수다.

### REQ-PENCIL-011 (Screenshot Archival)

`WHEN each batch completes without layout issues, the skill SHALL call mcp__pencil__get_screenshot and save PNG to .moai/design/screenshots/frame-<rootFrameId>-<ISO8601-timestamp>.png, where <rootFrameId> is the root frame node id returned by batch_design for that batch.`

**Rationale**: 시간 순 정렬과 batch 추적을 모두 가능하게 하려면 frame id와 ISO8601 timestamp 조합이 필요하다. `<rootFrameId>` 정의가 명확해야 downstream reproducibility가 보장된다.

### REQ-PENCIL-012 (Summary Report)

`WHEN all batches complete successfully, the skill SHALL write a summary report to .moai/design/pencil-run-<ISO8601-timestamp>.md listing applied batches, screenshot paths saved, and any warnings collected during execution.`

**Rationale**: 세션 종료 후 사용자가 어떤 batch가 적용되었는지 복기할 수 있어야 하며, 이 summary는 후속 SPEC의 근거 자료로도 활용된다.

### REQ-PENCIL-013 (MCP Runtime Unavailability)

`IF ToolSearch with query "mcp__pencil" returns no matching tool schemas at runtime, then the skill SHALL return PENCIL_MCP_UNAVAILABLE and point the user to Pencil MCP setup documentation.`

**Rationale**: REQ-PENCIL-001의 파일/폴더 체크는 MCP 런타임 상태를 알 수 없으므로 ToolSearch 결과가 MCP 가용성의 단일 진실원이다. REQ-001과 책임 경계가 분리된다.

### REQ-PENCIL-014 (Progress Updates)

`WHILE batch execution is in progress, the skill SHALL emit progress updates via TaskUpdate for each batch start and batch completion event.`

**Rationale**: 수십 개 batch를 실행하는 경우 사용자가 진행 상황을 추적할 수 있어야 timeout 오해와 불필요한 중단을 방지할 수 있다.

### REQ-PENCIL-015 (Retry on Batch Failure)

`WHEN batch_design returns an error for a given batch, the skill SHALL attempt exactly 1 retry (for a total of 2 attempts) with the same batch; if the retry also fails, then the skill SHALL return PENCIL_BATCH_FAILED with the failing batch index.`

**Rationale**: Transient 네트워크 오류는 1회 재시도로 회복되는 비율이 높지만, 그 이상은 근본 원인 조사가 필요하므로 unlimited retry는 오히려 해롭다. Error code 귀속을 REQ 자체에 명시해 traceability를 확보한다.

### REQ-PENCIL-016 (DSL Syntax Error Collection)

`IF pencil-plan.md contains tokens outside the DSL Grammar that the parser cannot interpret, then the skill SHALL collect all such errors (not halt on the first) and SHALL return PENCIL_PLAN_SYNTAX_ERROR with a list of { line_number, offending_text } entries.`

**Rationale**: 사용자가 한 번의 실행으로 전체 문법 오류를 확인할 수 있어야 반복 수정 cycle이 줄어든다. First-fail halt는 사용자 경험을 저하시킨다.

## Acceptance Criteria

### AC-1: Skill 파일 존재 (REQ-PENCIL-003)

`internal/template/templates/.claude/skills/moai-workflow-pencil-integration/SKILL.md` 파일이 존재하고, 유효한 YAML frontmatter를 가진다.

### AC-2: Phase B2.6 삽입 (REQ-PENCIL-001, REQ-PENCIL-002)

`.claude/skills/moai/workflows/design.md` 파일은 Phase B2.5와 Phase B3 사이에 `### Phase B2.6: Pencil Path (Conditional)` heading을 가진 섹션을 포함하고, 해당 섹션은 파일/폴더 precondition 체크 및 graceful skip 로직을 기술한다.

### AC-3: Skill allowed-tools 명시 (REQ-PENCIL-004)

Skill frontmatter의 `allowed-tools` 필드는 다음 tool 목록을 모두 포함한다: `mcp__pencil__batch_design, mcp__pencil__get_editor_state, mcp__pencil__snapshot_layout, mcp__pencil__get_screenshot, mcp__pencil__open_document, ToolSearch, Read, Write`.

### AC-4: Error Code 문서화 (REQ-PENCIL-006, REQ-PENCIL-013, REQ-PENCIL-015, REQ-PENCIL-016)

Skill 본문에 다음 4개 error code가 각각 trigger 조건 1줄과 recovery 절차 1줄 이상을 포함하는 표 또는 동등한 구조로 문서화되어 있다: `PENCIL_MCP_UNAVAILABLE`, `PENCIL_CONNECTION_FAILED`, `PENCIL_PLAN_SYNTAX_ERROR`, `PENCIL_BATCH_FAILED`.

### AC-5: Batch Split 동작 (REQ-PENCIL-008)

하나의 batch에 30개 operation이 포함된 pencil-plan.md를 처리할 때, skill은 이를 25 + 5 두 개의 sub-batch로 split하여 원래 연산 순서를 보존하며 실행한다.

### AC-6: Screenshot 아카이브 (REQ-PENCIL-011)

성공적인 실행 후, `.moai/design/screenshots/` 디렉토리에 `frame-[A-Za-z0-9_\-]+-[0-9]{8}T[0-9]{6}Z\.png` 패턴과 일치하는 PNG 파일이 최소 1개 이상 존재하며, `<id>` 부분은 해당 batch의 root frame node id와 일치한다.

### AC-7: 조용한 Skip (REQ-PENCIL-002)

`.moai/design/pencil-plan.md`가 없거나 `.pen` 파일이 존재하지 않을 때, `/moai design` Phase B2.6은 stdout/stderr 및 사용자 대화 표면에 에러 메시지 없이 skip되고 Phase B3로 즉시 진행한다.

### AC-8: Connection Failure 시 Phase B3 진행 (REQ-PENCIL-006)

Pencil 연결이 실패하면, skill은 `PENCIL_CONNECTION_FAILED`를 반환하고 orchestrator는 워크플로우 자체를 중단하지 않는다. Phase B2.6은 내부에서 종료되며 Phase B3 (BRIEF 생성)으로 진행한다. 여기서 "fallback"은 Path B 경로 재선택이 아니라 Path B 내부에서 Phase B2.6의 남은 작업을 생략하고 Phase B3로의 continuation을 의미한다.

### AC-9: Template Test 통과 (전체 REQ 통합 검증)

`go test ./internal/template/...` 명령이 성공적으로 통과한다.

### AC-10: Skill Invocation 동기 대기 (REQ-PENCIL-003)

Phase B2.6이 skill을 호출한 후 skill이 반환(성공 또는 structured error)하기 전까지 Phase B3가 시작되지 않음을, workflow trace 또는 skill invocation log로 검증할 수 있다.

### AC-11: ToolSearch 선행 로드 (REQ-PENCIL-004)

Skill 실행 trace에서 첫 번째 `mcp__pencil__*` 호출보다 먼저 `ToolSearch(query="mcp__pencil")` 호출이 기록된다.

### AC-12: Editor State 일치 검증 (REQ-PENCIL-005)

Skill은 `get_editor_state` 응답이 REQ-PENCIL-001에서 감지된 `.pen` 파일 이름을 포함하는지 확인하는 분기를 포함하며, 불일치 시 REQ-PENCIL-006의 error path로 진입한다.

### AC-13: Batch 순서 보존 (REQ-PENCIL-007)

pencil-plan.md에 Batch 1, Batch 2, Batch 3이 선언된 경우, skill은 `batch_design`을 정확히 해당 순서로 호출하며, 호출 순서는 실행 trace로 검증 가능하다.

### AC-14: Layout 검증 호출 (REQ-PENCIL-009)

각 batch 완료 직후, skill 실행 trace에는 `mcp__pencil__snapshot_layout` 호출이 `problemsOnly=true` 인자와 함께 기록된다.

### AC-15: Layout Issue 발생 시 Halt (REQ-PENCIL-010)

Layout issue가 감지된 batch 이후 다음 batch의 `batch_design` 호출이 실행 trace에 존재하지 않으며, skill은 문제 frame id가 포함된 structured error를 반환한다.

### AC-16: Summary Report 생성 (REQ-PENCIL-012)

전체 batch 성공 실행 직후 `.moai/design/pencil-run-[0-9]{8}T[0-9]{6}Z\.md` 패턴의 파일이 정확히 1개 생성되며, 파일 본문에는 applied batch 수, 저장된 screenshot 경로 리스트, warning 목록 섹션이 포함된다.

### AC-17: MCP 미가용 시 Error Code 반환 (REQ-PENCIL-013)

Pencil MCP 서버가 등록되지 않은 환경에서 `ToolSearch(query="mcp__pencil")`가 빈 결과를 반환하면, skill은 `PENCIL_MCP_UNAVAILABLE`을 반환하고 setup 문서 경로를 결과에 포함한다.

### AC-18: Progress Update Emit (REQ-PENCIL-014)

3개 이상의 batch가 포함된 pencil-plan.md 실행 시, 각 batch마다 최소 1회 이상의 `TaskUpdate` 호출이 실행 trace에 기록된다.

### AC-19: Batch Retry 및 Error Code 반환 (REQ-PENCIL-015)

`batch_design` mock이 2회 연속 실패하도록 설정된 테스트 시나리오에서, skill은 정확히 2회 시도(최초 1회 + 재시도 1회) 후 `PENCIL_BATCH_FAILED`를 반환하고, 반환 payload에는 실패한 batch index가 포함된다.

### AC-20: DSL Syntax Error 수집 (REQ-PENCIL-016)

DSL Grammar를 벗어난 토큰을 3개 이상 포함한 pencil-plan.md를 처리할 때, skill은 첫 오류에서 중단하지 않고 모든 오류를 수집하여 `PENCIL_PLAN_SYNTAX_ERROR` 반환 payload에 `{line_number, offending_text}` entry 3개 이상을 포함한다.

## Skill Frontmatter Template

`internal/template/templates/.claude/skills/moai-workflow-pencil-integration/SKILL.md`의 YAML frontmatter는 다음과 같은 형식을 가진다:

```yaml
---
name: moai-workflow-pencil-integration
description: >
  Detects .pen files and pencil-plan.md in .moai/design/, loads Pencil MCP,
  executes batch operations with layout verification, archives screenshots.
  Invoked from /moai design Phase B2.6 when preconditions are met.
license: Apache-2.0
compatibility: Designed for Claude Code with Pencil MCP
allowed-tools: ToolSearch, Read, Write, Glob, mcp__pencil__batch_design, mcp__pencil__get_editor_state, mcp__pencil__snapshot_layout, mcp__pencil__get_screenshot, mcp__pencil__open_document, mcp__pencil__find_empty_space_on_canvas
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "pencil, design, mcp, batch, wireframe"

triggers:
  keywords: ["pencil", ".pen", "batch design", "wireframe"]
  phases: ["design"]
---
```

## Error Code Table

| Code | Trigger | Recovery |
|---|---|---|
| PENCIL_MCP_UNAVAILABLE | ToolSearch(query="mcp__pencil")가 빈 결과 반환 (REQ-PENCIL-013) | Phase B2.6 skip, Phase B3로 진행; 사용자에게 Pencil MCP setup 문서 안내 |
| PENCIL_CONNECTION_FAILED | get_editor_state 실패 또는 잘못된 문서 반환 (REQ-PENCIL-006) | 사용자에게 Claude Desktop 재시작 또는 올바른 .pen 열기 안내 후 Phase B3로 진행 |
| PENCIL_PLAN_SYNTAX_ERROR | DSL Grammar 이탈 토큰 탐지 (REQ-PENCIL-016) | 모든 오류를 line 단위로 수집하여 반환, halt |
| PENCIL_BATCH_FAILED | batch_design 최초 호출 + 1회 재시도(총 2회 시도) 실패 (REQ-PENCIL-015) | Halt, 실패 batch index를 사용자에게 보고 |

## Scope

### IN SCOPE (구축 대상)

- `moai-workflow-pencil-integration` skill 신규 생성
- `/moai design` 워크플로우의 Phase B2.6 (Pencil Path) 삽입
- 4개 error code 정의 및 처리 로직
- Screenshot 아카이브 디렉토리 구조 (`.moai/design/screenshots/`)
- pencil-plan.md DSL parser (Batch heading 인식, I/M/R 연산 파싱, 25-op split)
- Summary report 생성 (`pencil-run-<ISO8601-timestamp>.md`)
- Layout verification (`snapshot_layout` with `problemsOnly: true`)
- Batch 단위 TaskUpdate 진행 상황 emit

### Exclusions (What NOT to Build)

- pencil-plan.md 내용 authoring (사용자 책임)
- Pencil MCP 서버 자체 구현 (upstream 책임)
- Claude Design bundle import (기존 `moai-workflow-design-import` 책임)
- `.pen` 파일 편집기 UI 제공
- Pencil 외 다른 디자인 tool (Figma, Sketch 등) 통합
- GAN Loop과 Pencil Path의 병합 (별도 SPEC에서 다룸)

## Risks

### R-1: Pencil MCP 업그레이드로 인한 DSL parser 깨짐

- 위험도: Medium
- 영향: DSL 문법이 변경되면 기존 pencil-plan.md 파일이 parsing 실패
- 완화책: `get_editor_state`를 통한 version check 선행, 호환되지 않는 버전 감지 시 명확한 에러 메시지 제공

### R-2: 대용량 pencil-plan.md로 인한 timeout

- 위험도: Medium
- 영향: 수십 개 batch가 있는 plan은 실행 시간이 길어져 timeout 가능성
- 완화책: REQ-PENCIL-014에 따른 batch 단위 TaskUpdate 진행 상황 emit으로 사용자가 진행 상황을 실시간 확인 가능. batch 단위 checkpoint를 통해 재시작 시 중복 실행 방지

### R-3: Screenshot 저장소 공간 팽창

- 위험도: Low
- 영향: 반복 실행 시 `.moai/design/screenshots/` 디렉토리가 계속 증가
- 완화책: `.pen` 파일당 최근 5회 실행분만 보관하는 retention 정책 적용, 그 이상은 자동 cleanup. cleanup 로직은 본 skill 실행 종료 시 수행
