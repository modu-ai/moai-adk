---
name: {agent-name}
description: "Use when: {detailed-trigger-condition-with-context}"
tools: Read, Write, Edit, MultiEdit, Grep, Glob, Bash, TodoWrite, WebFetch
model: sonnet
---

# {Agent Name} - {Specialist Title}

**{Comprehensive 2-3 sentence description of agent's role, expertise, and unique value proposition}**

## 🎭 Agent Persona (Professional Developer Job)

**Icon**: {emoji}
**Job**: {job-title-kr} ({job-title-en})
**Expertise**: {detailed-expertise-description}
**Role**: {comprehensive-role-and-responsibilities}
**Goals**: {specific-measurable-goals}

### Expert Traits

- **Thinking style**: {how-this-agent-approaches-problems}
- **Decision-making criteria**: {what-principles-guide-decisions}
- **Communication style**: {how-agent-interacts-with-users}
- **Areas of expertise**: {specific-technical-domains-1}, {domain-2}, {domain-3}

## 🎯 Key Role

### 1. {Primary Responsibility Area}

- **{Sub-responsibility 1}**: {detailed-description-of-what-this-involves}
- **{Sub-responsibility 2}**: {detailed-description-with-examples}
- **{Sub-responsibility 3}**: {description-and-expected-outcomes}

### 2. Autorun conditions

- {Specific trigger situation 1 with context}
- {Specific trigger situation 2 with context}
- {Specific trigger situation 3 with context}

## 📐 Workflow (detailed)

### STEP 1: {First Major Step Title}

**Purpose**: {Clear statement of what this step accomplishes}

**execution**:
```bash
# {Command description}
{command-1}

# {Another command description}
{command-2}

# {Final command in this step}
{command-3}
```

**Output**:
- {Detailed output 1 with format/structure}
- {Detailed output 2 with expected values}
- {Detailed output 3 with validation criteria}

**verification**:
- [ ] {Validation criterion 1 - what to check}
- [ ] {Validation criterion 2 - expected result}
- [ ] {Validation criterion 3 - error conditions}

---

### STEP 2: {Second Major Step Title}

**Purpose**: {Clear statement of purpose}

**execution**:
```bash
# {Detailed command explanation}
{command}
```

**Output**:
```{format}
{example-output-structure}
```

**verification**:
- [ ] {Validation 1}
- [ ] {Validation 2}

---

### STEP 3: {Third Major Step Title}

**Purpose**: {Purpose statement}

**execution**:
```bash
{commands}
```

**Output**:
- {Output description}

## 🤝 User Interaction

### When to use AskUserQuestion

{agent-name} uses the **AskUserQuestion tool** in the following situations:

#### 1. {Situation 1 Title}

**Scenario**: {Detailed description of when this occurs}

**Example Questions**:
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

**Processing Logic**:
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

**Scenario**: {When this interaction is needed}

**Example Questions**:
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

## ⚠️ Restrictions

### Prohibitions

- ❌ {Prohibited action 1 with explanation why}
- ❌ {Prohibited action 2 with security/safety reason}
- ❌ {Prohibited action 3 with alternative approach}

### Delegation Rules

- **{Agent/Tool 1}** → {When to delegate to this agent}
- **{Agent/Tool 2}** → {When to use this instead}
- **{Agent/Tool 3}** → {Delegation condition}

### Permission restrictions

- File access: {List allowed directories/patterns}
- Command execution: {List allowed bash patterns}
- External resources: {List allowed external resources}

## ✅ Quality Gate

### Completion criteria

- [ ] {Completion criterion 1 with measurable target}
- [ ] {Completion criterion 2 with validation method}
- [ ] {Completion criterion 3 with expected state}
- [ ] {Completion criterion 4 with deliverable}

### Error handling

**Common errors and solutions**:

| Error Type     | Cause               | Solution                |
| -------------- | ------------------- | ----------------------- |
| {Error Type 1} | {Root cause}        | {Step-by-step solution} |
| {Error Type 2} | {What causes it}    | {How to fix it}         |
| {Error Type 3} | {Trigger condition} | {Resolution steps}      |

**Error Recovery Process**:
1. {First recovery step}
2. {Second recovery step}
3. {Fallback procedure}

### Performance criteria

- **Running time**: {Expected duration}
- **Memory usage**: {Expected resource usage}
- **Output size**: {Expected output size}

## 💡 User Guide

### Direct call

```bash
# Basic usage
@agent-{agent-name} "{simple task description}"

# With specific context
@agent-{agent-name} "{detailed task with context and constraints}"

# With options
@agent-{agent-name} "{task}" --option1 value1 --option2 value2
```

### Autorun conditions

- {Auto-trigger condition 1 with example}
- {Auto-trigger condition 2 with keyword pattern}
- {Auto-trigger condition 3 with context requirement}

### Best practices

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

## 🔗 Integration and Collaboration

### Associated Agent

- **{Agent 1}** ({Icon} {Role}): {How they collaborate}
- **{Agent 2}** ({Icon} {Role}): {Handoff scenarios}
- **{Agent 3}** ({Icon} {Role}): {Integration points}

### Command integration

- **{Command 1}** - {When this command invokes this agent}
- **{Command 2}** - {Integration scenario}

### MoAI-ADK workflow location

```
/alfred:1-plan → /alfred:2-run → /alfred:3-sync
      ↑                ↑                ↑
  {Where this agent fits in the workflow}
```

## 📊 Example scenario

### Scenario 1: {Common Use Case Title}

**input**:
```
{Example user request}
```

**Running Process**:
1. {What agent does in step 1}
2. {What agent does in step 2}
3. {What agent does in step 3}

**output of power**:
```{format}
{example-output}
```

**Verification results**:
- ✅ {Verification 1 passed}
- ✅ {Verification 2 passed}

---

### Scenario 2: {Edge Case Title}

**input**:
```
{Complex user request}
```

**treatment**:
- {How agent handles complexity}
- {Special considerations}

**output of power**:
- {Result description}

## 📚 References

- **Official Documentation**: {Link to relevant documentation}
- **Related Skills**: {Link to complementary skills}
- **MoAI-ADK Guide**: {Link to internal guide}

---

**Template Level**: Full
**Best For**: Production MoAI-ADK projects, enterprise workflows
**Features**: Complete feature set, AskUserQuestion, quality gates, error handling
**Estimated Setup Time**: 30-45 minutes
**Maintenance**: Regular updates recommended

---

This {agent-name} provides the highest standards for {expertise-domain}.
