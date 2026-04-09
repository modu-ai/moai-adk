# Acceptance Criteria: SPEC-AGENT-002

## AC-001: Common Rule 생성 및 로드

### Scenario 1: Rule 파일이 존재하고 공통 프로토콜을 포함

```gherkin
Given agent-common-protocol.md rule 파일이 templates/.claude/rules/moai/core/에 존재할 때
When 에이전트가 호출되면
Then Language Handling, Output Format, MCP Fallback 프로토콜이 컨텍스트에 로드되어야 한다
```

### Scenario 2: 기존 에이전트 본문에서 공통 섹션이 제거됨

```gherkin
Given 16개 MoAI 에이전트 정의가 수정되었을 때
When 각 에이전트 본문을 검사하면
Then "Language Handling" 섹션이 본문에 존재하지 않아야 한다
And "Output Format Rules" 섹션이 본문에 존재하지 않아야 한다
And "Essential Reference" 섹션이 본문에 존재하지 않아야 한다
```

## AC-002: Workflow Step 100% 보존

### Scenario 1: Workflow Step 수 보존

```gherkin
Given 축소 전 expert-backend 에이전트의 Workflow Steps가 6단계일 때
When 축소 후 본문을 검사하면
Then Workflow Steps가 정확히 6단계여야 한다
And 각 단계의 핵심 동작이 동일해야 한다
```

### Scenario 2: 모든 에이전트 워크플로우 보존

```gherkin
Given 16개 MoAI 에이전트 각각의 축소 전 Workflow Step 수가 기록되었을 때
When 축소 후 모든 에이전트를 검사하면
Then 각 에이전트의 Workflow Step 수가 축소 전과 동일해야 한다
```

## AC-003: Permission Mode 정렬

### Scenario 1: 분석 전용 에이전트 권한

```gherkin
Given manager-strategy 에이전트의 역할이 "Read-Only Analysis"일 때
When 에이전트 frontmatter를 검사하면
Then permissionMode가 "plan"이어야 한다
And tools에 Write, Edit가 포함되지 않아야 한다
```

### Scenario 2: 평가 전용 에이전트 권한

```gherkin
Given manager-quality 에이전트의 역할이 "Read-only evaluator"일 때
When 에이전트 frontmatter를 검사하면
Then permissionMode가 "plan"이어야 한다
And tools에 Write, Edit가 포함되지 않아야 한다
```

## AC-004: maxTurns 제거

```gherkin
Given 22개 에이전트 정의가 수정되었을 때
When 모든 에이전트의 YAML frontmatter를 검사하면
Then maxTurns 필드가 존재하지 않아야 한다
```

## AC-005: Agent Body 크기 감소

```gherkin
Given 16개 MoAI 에이전트 본문이 축소되었을 때
When 각 에이전트의 본문 라인 수를 측정하면
Then 평균 본문 크기가 120줄 이하여야 한다
And 최대 본문 크기가 200줄을 초과하지 않아야 한다
```

## AC-006: Hook Matcher MultiEdit 포함

```gherkin
Given manager-ddd와 manager-tdd의 tools에 MultiEdit가 포함될 때
When hooks의 PreToolUse/PostToolUse matcher를 검사하면
Then matcher 패턴이 "Write|Edit|MultiEdit"를 포함해야 한다
```

## AC-007: Stale Reference 제거

```gherkin
Given 모든 에이전트 정의가 수정되었을 때
When Grep으로 "Rule 1:", "Rule 3:", "Rule 5:", "Rule 6:" 패턴을 검색하면
Then 매칭 결과가 0건이어야 한다

When Grep으로 "support-claude", "workflow-docs", "expert-database", "design-uiux" 패턴을 검색하면
Then 매칭 결과가 0건이어야 한다

When Grep으로 "moai-core-trust-validation", "moai-essentials-review", "moai-ai-nano-banana" 패턴을 검색하면
Then 매칭 결과가 0건이어야 한다
```

## AC-008: Build 및 Test 통과

```gherkin
Given 모든 template 변경이 완료되었을 때
When make build를 실행하면
Then embedded.go가 정상 생성되어야 한다

When go test ./...를 실행하면
Then 모든 테스트가 통과해야 한다
```

## AC-009: Delegation Protocol 보존

```gherkin
Given 각 에이전트의 축소 전 Delegation Protocol이 기록되었을 때
When 축소 후 Delegation Protocol을 검사하면
Then 위임 대상 에이전트 목록이 축소 전과 동일해야 한다
And "When to delegate" 조건이 보존되어야 한다
```

## AC-010: Scope Boundaries 보존

```gherkin
Given 각 에이전트의 축소 전 IN SCOPE/OUT OF SCOPE 목록이 기록되었을 때
When 축소 후 Scope Boundaries를 검사하면
Then IN SCOPE 항목이 축소 전과 동일해야 한다
And OUT OF SCOPE 항목이 축소 전과 동일해야 한다
```

## Quality Gate

- [x] 모든 AC 시나리오 통과
- [x] make build 성공
- [x] go test ./... 통과
- [x] 로컬 .claude/agents/ 동기화 완료
