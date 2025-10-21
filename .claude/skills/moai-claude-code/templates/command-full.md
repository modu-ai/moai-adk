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

## 🎯 Command Purpose

{Detailed multi-paragraph explanation covering:
- What problem this command solves
- How it fits into the larger workflow
- When to use this command vs alternatives
- What makes this command unique/valuable}

## 📋 Execution flow (2-Phase structure)

### ⚙️ Phase 0: Environmental Analysis (Optional)

**Purpose**: {Pre-execution analysis purpose}

**execution**:
```bash
# {Environment check description}
{command-1}

# {Prerequisites verification}
{command-2}
```

**verification**:
- [ ] {Prerequisite 1 checked}
- [ ] {Prerequisite 2 verified}

---

### 📊 Phase 1: {Planning/Analysis Phase}

**Purpose**: {Detailed purpose of planning phase}

**Automatic processing**:
- {Auto-task 1 that happens without user input}
- {Auto-task 2 that system handles}
- {Auto-task 3 performed automatically}

**Execution Steps**:

#### 1.1 {First Sub-Step}
```bash
# {Detailed explanation}
{command-or-action}
```

**Output**:
- {Output 1 with format specification}
- {Output 2 with expected structure}

#### 1.2 {Second Sub-Step}
```bash
{commands}
```

**Output**:
- {Intermediate output description}

#### 1.3 {User Confirmation}

**AskUserQuestion timing**: {When user confirmation is needed}

**Confirmation**:
```typescript
AskUserQuestion({
  questions: [{
    question: "{What to ask user}?",
    header: "{Short header}",
    options: [
{ label: "Proceed", description: "Execute Phase 2" },
 { label: "Modify", description: "{What modification means}" },
 { label: "Abort", description: "Cancel operation" }
    ],
    multiSelect: false
  }]
})
```

**Phase 1 deliverable (final)**:
- {Complete output 1 from planning}
- {Complete output 2 ready for execution}
- {User-approved plan}

---

### 🚀 Phase 2: {Execution Phase}

**Purpose**: {Detailed purpose of execution phase}

**Prerequisites**:
- [ ] Phase 1 completed and user approved
- [ ] {Additional precondition 1}
- [ ] {Additional precondition 2}

**Execution Steps**:

#### 2.1 {First Execution Step}
```bash
# {What this does}
{execution-command-1}

# {Next action}
{execution-command-2}
```

**Real-time progress**:
```
{Progress indicator format}
[▓▓▓▓▓▓▓░░░] {percentage}% - {current-action}
```

#### 2.2 {Second Execution Step}
```bash
{commands-with-explanations}
```

#### 2.3 {Quality Verification}
```bash
# {Validation check 1}
{validation-command-1}

# {Validation check 2}
{validation-command-2}
```

**Verification Criteria**:
- [ ] {Quality criterion 1 with threshold}
- [ ] {Quality criterion 2 with expected value}
- [ ] {Quality criterion 3 with pass/fail}

**Phase 2 final output**:
```{format}
{example-final-output-structure}
```

## 🔗 Associated Agent

### Primary Agent
- **{agent-name}** ({Icon} {Persona})
- **Expertise**: {Expertise}
 - **When invoked**: {When invoked}
 - **Role**: {What agent does in this command}

### Secondary Agents
- **{agent-2}** ({Icon} {Role}) - {Integration scenario}
- **{agent-3}** ({Icon} {Role}) - {When used}

## 💡 Example of use

### Default Enabled
```bash
/{command-name} {basic-example}
```

### Advanced Use
```bash
# {Advanced use case 1}
/{command-name} {param1} --{option1}={value1}

# {Advanced use case 2}
/{command-name} {param1} {param2} --{flag}
```

### Real-world scenarios

#### Scenario 1: {Common Workflow}
```bash
# Step 1: {What user does first}
/{command-name} "{example-input}"

# Result: {What happens}
# Next: {What to do next}
```

#### Scenario 2: {Edge Case}
```bash
# When {special condition}
/{command-name} {special-params}

# Handles: {How command adapts}
```

## Command argument details

| Arguments/Options | Type | Required | default | Description |
|----------|------|------|--------|------|
| `{param1}` | {type} | ✅ | - | {Detailed description of param1} |
| `{param2}` | {type} | ⚠️ | {default} | {Detailed description of param2} |
| `--{option1}` | {type} | ⚠️ | {default} | {What this option controls} |
| `--{flag}` | boolean | ⚠️ | false | {When to use this flag} |

**Argument Validation**:
- {Validation rule 1}
- {Validation rule 2}

## ⚠️ Prohibitions

**What you should never do**:

- ❌ {Prohibited action 1 with explanation}
- ❌ {Prohibited action 2 with reason}
- ❌ {Prohibited action 3 with alternative}

**Expressions to use**:

- ✅ {Recommended practice 1}
- ✅ {Recommended practice 2}

## 🚨 Error handling

### Common errors

| error message | Cause | Solution |
|-----------|------|----------|
| `{Error 1}` | {Root cause} | {Step-by-step solution} |
| `{Error 2}` | {What triggers it} | {How to fix} |
| `{Error 3}` | {Condition} | {Resolution} |

### Recovery Procedure

1. **{Recovery Step 1}**: {What to do first}
2. **{Recovery Step 2}**: {Next action}
3. **{Fallback}**: {Last resort if all fails}

## ✅ Success Criteria

**Check points after executing the command**:

- [ ] {Success criterion 1 with verification method}
- [ ] {Success criterion 2 with expected outcome}
- [ ] {Success criterion 3 with deliverable}

**Quality Gate**:
```bash
# {Quality check 1}
{verification-command-1}

# {Quality check 2}
{verification-command-2}
```

## 📋 Next steps

**Recommended Workflow**:

1. **Execute immediately**: {What to do right after command completes}
2. **Verification**: {How to verify results}
3. **Next command**: `/{next-command}` - {Why this is next}

**Alternative Path**:
- {Alternative path 1 if condition X}
- {Alternative path 2 if condition Y}

## 🔄 Integrated Workflow

### MoAI-ADK workflow location

```
/{prev-command} → /{command-name} → /{next-command}
                        ↓
                {Connected agents/tasks}
```

### Relationship with other commands

| command | relationship | Execution order |
|--------|------|----------|
| `/{related-1}` | {Relationship} | {Before/After/Parallel} |
| `/{related-2}` | {Relationship} | {Sequence} |

## 📊 Performance Metrics

- **Average execution time**: {Expected duration}
- **Memory usage**: {Expected memory}
- **Number of files created**: {Expected file count}
- **API calls**: {Expected external calls}

## 🎓 Best Practices

### 1. {Practice Category 1}

**Recommended**:
```bash
# {Good example}
/{command-name} {recommended-usage}
```

**Not recommended**:
```bash
# {Bad example - why to avoid}
/{command-name} {anti-pattern}
```

### 2. {Practice Category 2}

**Tip**: {Helpful tip or trick}

### 3. {Practice Category 3}

**Caution**: {Important consideration}

## 🔗 Related Resources

### Related commands
- `/{command-1}` - {Description and relation}
- `/{command-2}` - {Description and when to use}

### Related Agents
- `@agent-{agent-1}` - {How it supports this command}
- `@agent-{agent-2}` - {Integration point}

### document
- **SPEC**: {Link to specification}
- **Guide**: {Link to detailed guide}
- **Examples**: {Link to examples}

## 📝 Command output example

**Success Case**:
```
✅ {Command Name} completed

📊 Execution result:
- {Result metric 1}: {value}
- {Result metric 2}: {value}
- {Result metric 3}: {value}

📁 Files generated:
- {File 1}: {Description}
- {File 2}: {Description}

📋 Next steps:
- {Next step 1}
- {Next step 2}
```

**Error Case**:
```
❌ {Command Name} failed

🔍 Error details:
- Type: {Error type}
- Location: {Where error occurred}
- Message: {Error message}

💡 Solution:
1. {Solution step 1}
2. {Solution step 2}

📞 Additional help: {Where to get help}
```

---

**Template Level**: Full
**Best For**: Production MoAI-ADK workflows, enterprise automation
**Features**: 2-phase structure, quality gates, comprehensive error handling, integration
**Estimated Setup Time**: 30-45 minutes
**Maintenance**: Regular updates recommended as workflows evolve

---

This command provides standard automation for {workflow-domain}.
