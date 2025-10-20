---
name: {command-name}
description: {Comprehensive one-line description with context}
argument-hint: [{param1}] [{param2}] [{options}]
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Task
  - Bash(git:*)
  - Bash({specific-pattern}:*)
---

# 📋 {Command Title}

{Comprehensive 2-3 sentence description of command's purpose, integration with MoAI-ADK workflow, and key benefits}

## 🎯 커맨드 목적

{Detailed multi-paragraph explanation covering:
- What problem this command solves
- How it fits into the larger workflow
- When to use this command vs alternatives
- What makes this command unique/valuable}

## 📋 실행 흐름 (2-Phase 구조)

### ⚙️ Phase 0: 환경 분석 (선택적)

**목적**: {Pre-execution analysis purpose}

**실행**:
```bash
# {Environment check description}
{command-1}

# {Prerequisites verification}
{command-2}
```

**검증**:
- [ ] {Prerequisite 1 checked}
- [ ] {Prerequisite 2 verified}

---

### 📊 Phase 1: {Planning/Analysis Phase}

**목적**: {Detailed purpose of planning phase}

**자동 처리**:
- {Auto-task 1 that happens without user input}
- {Auto-task 2 that system handles}
- {Auto-task 3 performed automatically}

**실행 단계**:

#### 1.1 {First Sub-Step}
```bash
# {Detailed explanation}
{command-or-action}
```

**산출물**:
- {Output 1 with format specification}
- {Output 2 with expected structure}

#### 1.2 {Second Sub-Step}
```bash
{commands}
```

**산출물**:
- {Intermediate output description}

#### 1.3 {사용자 확인}

**AskUserQuestion 시점**: {When user confirmation is needed}

**확인 내용**:
```typescript
AskUserQuestion({
  questions: [{
    question: "{What to ask user}?",
    header: "{Short header}",
    options: [
      { label: "진행", description: "Phase 2 실행" },
      { label: "수정", description: "{What modification means}" },
      { label: "중단", description: "작업 취소" }
    ],
    multiSelect: false
  }]
})
```

**Phase 1 산출물 (최종)**:
- {Complete output 1 from planning}
- {Complete output 2 ready for execution}
- {User-approved plan}

---

### 🚀 Phase 2: {Execution Phase}

**목적**: {Detailed purpose of execution phase}

**사전 조건**:
- [ ] Phase 1 완료 및 사용자 승인
- [ ] {Additional precondition 1}
- [ ] {Additional precondition 2}

**실행 단계**:

#### 2.1 {First Execution Step}
```bash
# {What this does}
{execution-command-1}

# {Next action}
{execution-command-2}
```

**실시간 진행 상황**:
```
{Progress indicator format}
[▓▓▓▓▓▓▓░░░] {percentage}% - {current-action}
```

#### 2.2 {Second Execution Step}
```bash
{commands-with-explanations}
```

#### 2.3 {품질 검증}
```bash
# {Validation check 1}
{validation-command-1}

# {Validation check 2}
{validation-command-2}
```

**검증 기준**:
- [ ] {Quality criterion 1 with threshold}
- [ ] {Quality criterion 2 with expected value}
- [ ] {Quality criterion 3 with pass/fail}

**Phase 2 최종 산출물**:
```{format}
{example-final-output-structure}
```

## 🔗 연관 에이전트

### Primary Agent
- **{agent-name}** ({Icon} {Persona})
  - **전문 영역**: {Expertise}
  - **호출 시점**: {When invoked}
  - **역할**: {What agent does in this command}

### Secondary Agents
- **{agent-2}** ({Icon} {Role}) - {Integration scenario}
- **{agent-3}** ({Icon} {Role}) - {When used}

## 💡 사용 예시

### 기본 사용
```bash
/{command-name} {basic-example}
```

### 고급 사용
```bash
# {Advanced use case 1}
/{command-name} {param1} --{option1}={value1}

# {Advanced use case 2}
/{command-name} {param1} {param2} --{flag}
```

### 실전 시나리오

#### 시나리오 1: {Common Workflow}
```bash
# Step 1: {What user does first}
/{command-name} "{example-input}"

# Result: {What happens}
# Next: {What to do next}
```

#### 시나리오 2: {Edge Case}
```bash
# When {special condition}
/{command-name} {special-params}

# Handles: {How command adapts}
```

## 명령어 인수 상세

| 인수/옵션 | 타입 | 필수 | 기본값 | 설명 |
|----------|------|------|--------|------|
| `{param1}` | {type} | ✅ | - | {Detailed description of param1} |
| `{param2}` | {type} | ⚠️ | {default} | {Detailed description of param2} |
| `--{option1}` | {type} | ⚠️ | {default} | {What this option controls} |
| `--{flag}` | boolean | ⚠️ | false | {When to use this flag} |

**인수 검증**:
- {Validation rule 1}
- {Validation rule 2}

## ⚠️ 금지 사항

**절대 하지 말아야 할 작업**:

- ❌ {Prohibited action 1 with explanation}
- ❌ {Prohibited action 2 with reason}
- ❌ {Prohibited action 3 with alternative}

**사용해야 할 표현**:

- ✅ {Recommended practice 1}
- ✅ {Recommended practice 2}

## 🚨 에러 처리

### 일반적인 오류

| 에러 메시지 | 원인 | 해결 방법 |
|-----------|------|----------|
| `{Error 1}` | {Root cause} | {Step-by-step solution} |
| `{Error 2}` | {What triggers it} | {How to fix} |
| `{Error 3}` | {Condition} | {Resolution} |

### 복구 절차

1. **{Recovery Step 1}**: {What to do first}
2. **{Recovery Step 2}**: {Next action}
3. **{Fallback}**: {Last resort if all fails}

## ✅ 성공 기준

**커맨드 실행 후 확인 사항**:

- [ ] {Success criterion 1 with verification method}
- [ ] {Success criterion 2 with expected outcome}
- [ ] {Success criterion 3 with deliverable}

**품질 게이트**:
```bash
# {Quality check 1}
{verification-command-1}

# {Quality check 2}
{verification-command-2}
```

## 📋 다음 단계

**권장 워크플로우**:

1. **즉시 실행**: {What to do right after command completes}
2. **검증**: {How to verify results}
3. **다음 커맨드**: `/{next-command}` - {Why this is next}

**대안 경로**:
- {Alternative path 1 if condition X}
- {Alternative path 2 if condition Y}

## 🔄 통합 워크플로우

### MoAI-ADK 워크플로우 위치

```
/{prev-command} → /{command-name} → /{next-command}
                        ↓
                {Connected agents/tasks}
```

### 다른 커맨드와의 관계

| 커맨드 | 관계 | 실행 순서 |
|--------|------|----------|
| `/{related-1}` | {Relationship} | {Before/After/Parallel} |
| `/{related-2}` | {Relationship} | {Sequence} |

## 📊 성능 메트릭

- **평균 실행 시간**: {Expected duration}
- **메모리 사용량**: {Expected memory}
- **생성 파일 수**: {Expected file count}
- **API 호출**: {Expected external calls}

## 🎓 베스트 프랙티스

### 1. {Practice Category 1}

**권장**:
```bash
# {Good example}
/{command-name} {recommended-usage}
```

**비권장**:
```bash
# {Bad example - why to avoid}
/{command-name} {anti-pattern}
```

### 2. {Practice Category 2}

**팁**: {Helpful tip or trick}

### 3. {Practice Category 3}

**주의**: {Important consideration}

## 🔗 관련 리소스

### 관련 커맨드
- `/{command-1}` - {Description and relation}
- `/{command-2}` - {Description and when to use}

### 관련 에이전트
- `@agent-{agent-1}` - {How it supports this command}
- `@agent-{agent-2}` - {Integration point}

### 문서
- **SPEC**: {Link to specification}
- **가이드**: {Link to detailed guide}
- **예제**: {Link to examples}

## 📝 커맨드 출력 예시

**성공 케이스**:
```
✅ {Command Name} 완료

📊 실행 결과:
- {Result metric 1}: {value}
- {Result metric 2}: {value}
- {Result metric 3}: {value}

📁 생성된 파일:
- {File 1}: {Description}
- {File 2}: {Description}

📋 다음 단계:
- {Next step 1}
- {Next step 2}
```

**에러 케이스**:
```
❌ {Command Name} 실패

🔍 에러 상세:
- 유형: {Error type}
- 위치: {Where error occurred}
- 메시지: {Error message}

💡 해결 방법:
1. {Solution step 1}
2. {Solution step 2}

📞 추가 도움: {Where to get help}
```

---

**Template Level**: Full
**Best For**: Production MoAI-ADK workflows, enterprise automation
**Features**: 2-phase structure, quality gates, comprehensive error handling, integration
**Estimated Setup Time**: 30-45 minutes
**Maintenance**: Regular updates recommended as workflows evolve

---

이 커맨드는 {workflow-domain}의 표준 자동화를 제공합니다.
