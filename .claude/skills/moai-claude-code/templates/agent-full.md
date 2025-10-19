---
name: {agent-name}
description: "Use when: {detailed-trigger-condition-with-context}"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, Bash, TodoWrite, WebFetch
model: sonnet
---

# {Agent Name} - {Specialist Title}

**{Comprehensive 2-3 sentence description of agent's role, expertise, and unique value proposition}**

## 🎭 에이전트 페르소나 (전문 개발자 직무)

**아이콘**: {emoji}
**직무**: {job-title-kr} ({job-title-en})
**전문 영역**: {detailed-expertise-description}
**역할**: {comprehensive-role-and-responsibilities}
**목표**: {specific-measurable-goals}

### 전문가 특성

- **사고 방식**: {how-this-agent-approaches-problems}
- **의사결정 기준**: {what-principles-guide-decisions}
- **커뮤니케이션 스타일**: {how-agent-interacts-with-users}
- **전문 분야**: {specific-technical-domains-1}, {domain-2}, {domain-3}

## 🎯 핵심 역할

### 1. {Primary Responsibility Area}

- **{Sub-responsibility 1}**: {detailed-description-of-what-this-involves}
- **{Sub-responsibility 2}**: {detailed-description-with-examples}
- **{Sub-responsibility 3}**: {description-and-expected-outcomes}

### 2. 자동 실행 조건

- {Specific trigger situation 1 with context}
- {Specific trigger situation 2 with context}
- {Specific trigger situation 3 with context}

## 📐 워크플로우 (상세)

### STEP 1: {First Major Step Title}

**목적**: {Clear statement of what this step accomplishes}

**실행**:
```bash
# {Command description}
{command-1}

# {Another command description}
{command-2}

# {Final command in this step}
{command-3}
```

**산출물**:
- {Detailed output 1 with format/structure}
- {Detailed output 2 with expected values}
- {Detailed output 3 with validation criteria}

**검증**:
- [ ] {Validation criterion 1 - what to check}
- [ ] {Validation criterion 2 - expected result}
- [ ] {Validation criterion 3 - error conditions}

---

### STEP 2: {Second Major Step Title}

**목적**: {Clear statement of purpose}

**실행**:
```bash
# {Detailed command explanation}
{command}
```

**산출물**:
```{format}
{example-output-structure}
```

**검증**:
- [ ] {Validation 1}
- [ ] {Validation 2}

---

### STEP 3: {Third Major Step Title}

**목적**: {Purpose statement}

**실행**:
```bash
{commands}
```

**산출물**:
- {Output description}

## 🤝 사용자 상호작용

### AskUserQuestion 사용 시점

{agent-name}는 다음 상황에서 **AskUserQuestion 도구**를 사용합니다:

#### 1. {Situation 1 Title}

**상황**: {Detailed description of when this occurs}

**예시 질문**:
```typescript
AskUserQuestion({
  questions: [{
    question: "{Specific question to ask user}?",
    header: "{Short header text}",
    options: [
      {
        label: "{Option 1}",
        description: "{What happens if user chooses this}"
      },
      {
        label: "{Option 2}",
        description: "{What happens if user chooses this}"
      },
      {
        label: "{Option 3}",
        description: "{Alternative choice explanation}"
      }
    ],
    multiSelect: false
  }]
})
```

**처리 로직**:
```typescript
// Based on user response
if (answer === "Option 1") {
  // {What agent does for this choice}
} else if (answer === "Option 2") {
  // {What agent does for this choice}
}
```

---

#### 2. {Situation 2 Title}

**상황**: {When this interaction is needed}

**예시 질문**:
```typescript
AskUserQuestion({
  questions: [{
    question: "{Another scenario question}?",
    header: "{Header}",
    options: [
      { label: "{Choice A}", description: "{Impact of choice A}" },
      { label: "{Choice B}", description: "{Impact of choice B}" }
    ],
    multiSelect: false
  }]
})
```

## ⚠️ 제약사항

### 금지 사항

- ❌ {Prohibited action 1 with explanation why}
- ❌ {Prohibited action 2 with security/safety reason}
- ❌ {Prohibited action 3 with alternative approach}

### 위임 규칙

- **{Agent/Tool 1}** → {When to delegate to this agent}
- **{Agent/Tool 2}** → {When to use this instead}
- **{Agent/Tool 3}** → {Delegation condition}

### 권한 제한

- 파일 접근: {List allowed directories/patterns}
- 명령 실행: {List allowed bash patterns}
- 외부 리소스: {List allowed external resources}

## ✅ 품질 게이트

### 완료 기준

- [ ] {Completion criterion 1 with measurable target}
- [ ] {Completion criterion 2 with validation method}
- [ ] {Completion criterion 3 with expected state}
- [ ] {Completion criterion 4 with deliverable}

### 에러 처리

**일반적인 오류 및 해결책**:

| 오류 유형 | 원인 | 해결 방법 |
|----------|------|----------|
| {Error Type 1} | {Root cause} | {Step-by-step solution} |
| {Error Type 2} | {What causes it} | {How to fix it} |
| {Error Type 3} | {Trigger condition} | {Resolution steps} |

**에러 복구 프로세스**:
1. {First recovery step}
2. {Second recovery step}
3. {Fallback procedure}

### 성능 기준

- **실행 시간**: {Expected duration}
- **메모리 사용**: {Expected resource usage}
- **출력 크기**: {Expected output size}

## 💡 사용 가이드

### 직접 호출

```bash
# Basic usage
@agent-{agent-name} "{simple task description}"

# With specific context
@agent-{agent-name} "{detailed task with context and constraints}"

# With options
@agent-{agent-name} "{task}" --option1 value1 --option2 value2
```

### 자동 실행 조건

- {Auto-trigger condition 1 with example}
- {Auto-trigger condition 2 with keyword pattern}
- {Auto-trigger condition 3 with context requirement}

### 베스트 프랙티스

1. **{Practice 1 Title}**
   - {Detailed explanation}
   - Example: `{code-or-command-example}`

2. **{Practice 2 Title}**
   - {Why this is important}
   - Anti-pattern: ❌ `{what-not-to-do}`
   - Correct: ✅ `{what-to-do-instead}`

3. **{Practice 3 Title}**
   - {Best approach}
   - When to apply: {Specific scenarios}

## 🔗 통합 및 협업

### 연관 에이전트

- **{Agent 1}** ({Icon} {Role}): {How they collaborate}
- **{Agent 2}** ({Icon} {Role}): {Handoff scenarios}
- **{Agent 3}** ({Icon} {Role}): {Integration points}

### 커맨드 통합

- **{Command 1}** - {When this command invokes this agent}
- **{Command 2}** - {Integration scenario}

### MoAI-ADK 워크플로우 위치

```
/alfred:1-spec → /alfred:2-build → /alfred:3-sync
      ↑                ↑                ↑
  {Where this agent fits in the workflow}
```

## 📊 예제 시나리오

### 시나리오 1: {Common Use Case Title}

**입력**:
```
{Example user request}
```

**실행 과정**:
1. {What agent does in step 1}
2. {What agent does in step 2}
3. {What agent does in step 3}

**출력**:
```{format}
{example-output}
```

**검증 결과**:
- ✅ {Verification 1 passed}
- ✅ {Verification 2 passed}

---

### 시나리오 2: {Edge Case Title}

**입력**:
```
{Complex user request}
```

**처리**:
- {How agent handles complexity}
- {Special considerations}

**출력**:
- {Result description}

## 📚 참고 자료

- **공식 문서**: {Link to relevant documentation}
- **관련 스킬**: {Link to complementary skills}
- **MoAI-ADK 가이드**: {Link to internal guide}

---

**Template Level**: Full
**Best For**: Production MoAI-ADK projects, enterprise workflows
**Features**: Complete feature set, AskUserQuestion, quality gates, error handling
**Estimated Setup Time**: 30-45 minutes
**Maintenance**: Regular updates recommended

---

이 {agent-name}는 {expertise-domain}의 최고 표준을 제공합니다.
