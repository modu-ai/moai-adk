---
name: {skill-name}
description: {Capability + trigger phrases (<=1024 chars, aim for <=200)}
allowed-tools:                     # optional; remove if full tool access is acceptable
  - Read
  - Bash
  - Write
---

# {Skill Title}

> {One-sentence compelling summary of value proposition}

---

## 🎯 Purpose of this skill

{Comprehensive explanation covering:
- Problem statement and context
- How this skill addresses the problem
- Unique value proposition
- Integration with broader workflows}

**Problem**: {Detailed problem description with examples}
**Solution**: {Comprehensive solution approach}
**Impact**: {Measurable benefits and improvements}

---

## 🏗️ MoAI-ADK integration

### Alfred auto-selection conditions

Alfred automatically activates this skill under the following conditions:

- {Specific automatic trigger condition 1 with context}
- {Specific automatic trigger condition 2 with keywords}
- {Specific automatic trigger condition 3 with workflow state}

### Workflow location

```
/alfred:1-plan → /alfred:2-run → /alfred:3-sync
                                        ↑
Automatically activate this skill
                                  ({when activated})
```

**Integration Point**:
- **Phase**: {Which phase of MoAI-ADK workflow}
- **Trigger**: {What triggers automatic invocation}
- **Role**: {What this skill contributes to the workflow}

---

## 📋 Core features

### 1. {Major Feature 1 Name}

{Detailed multi-paragraph description of this feature}

**How ​​to implement**:
```{language}
# {Implementation detail 1}
{code-example-1}

# {Implementation detail 2}
{code-example-2}
```

**Output**:
- **{Output 1}**: {Detailed description with format}
- **{Output 2}**: {Description with validation criteria}
- **{Output 3}**: {Description with usage notes}

**verification**:
```bash
# {Verification method}
{verification-command}
```

---

### 2. {Major Feature 2 Name}

{Comprehensive feature description}

**Algorithm**:
1. {Step 1 of algorithm}
2. {Step 2 with details}
3. {Step 3 and expected outcome}

**Implementation example**:
```{language}
{detailed-code-example}
```

---

### 3. {Major Feature 3 Name}

{Feature description with use cases}

**Use Scenario**:
- **{Scenario A}**: {When and why to use}
- **{Scenario B}**: {Alternative use case}

---

## 💡 Usage Pattern

### Pattern 1: Manual call

**Example User Request**:
```
"Please execute {skill-name}"
"{natural-language-trigger-phrase}"
```

**Alfred Action**:
1. {What Alfred does in step 1}
2. {What Alfred does in step 2}
3. {Final action and result}

---

### Pattern 2: Automatic activation

**Trigger condition**: {When automatic activation occurs}

**Alfred detection scenario**:
```
User: "{example-user-request}"
→ Alfred Analysis: {how Alfred recognizes this needs the skill}
→ Autoplay: {what happens automatically}
→ Result: {what user receives}
```

---

### Pattern 3: Command integration

**Related command**: `/{command-name}`

**Integrated Flow**:
```
Run /{command-name}
  ↓
{When skill is invoked during command}
  ↓
Automatically call this skill
  ↓
{What skill contributes}
  ↓
Continue command
```

---

## ⚙️ Settings and configuration

### Configuration file location

Configure in `.moai/config.json`:

```json
{
  "{skill-config-section}": {
    "{option1}": {default-value},
    "{option2}": {default-value},
    "{option3}": {
      "{sub-option1}": {value},
      "{sub-option2}": {value}
    }
  }
}
```

### Setting option details

| Options     | Type   | default     | Required | Description                 |
| ----------- | ------ | ----------- | -------- | --------------------------- |
| `{option1}` | {type} | `{default}` | ✅/⚠️      | {Comprehensive description} |
| `{option2}` | {type} | `{default}` | ⚠️        | {What this controls}        |
| `{option3}` | {type} | `{default}` | ⚠️        | {Usage notes}               |

### Environment variables (optional)

```bash
# {Environment variable 1}
export {VAR_NAME_1}="{value}"

# {Environment variable 2}
export {VAR_NAME_2}="{value}"
```

---

## 📁 Directory Structure

```
.claude/skills/{skill-name}/
├── SKILL.md # Main skill definition (this file)
├── reference.md # Detailed reference document
├── examples.md # Collection of practical examples
├── scripts/ # Utility script
│   ├── {helper-1}.py
│   └── {helper-2}.py
└── templates/ # template file
    ├── {template-1}.txt
    └── {template-2}.json
```

### Additional file descriptions

- **reference.md**: {What additional documentation it contains}
- **examples.md**: {What examples are provided}
- **scripts/**: {What utility scripts do}
- **templates/**: {What templates are included}

---

## ✅ Verification Checklist

### Validation before execution

- [ ] {Pre-execution check 1}
- [ ] {Pre-execution check 2}
- [ ] {Pre-execution check 3}

### Verify after execution

- [ ] {Post-execution validation 1 with criteria}
- [ ] {Post-execution validation 2 with expected state}
- [ ] {Post-execution validation 3 with deliverable}
- [ ] {Check MoAI-ADK workflow integration}

### Verification command

```bash
# {Validation script 1}
uv run .claude/skills/{skill-name}/scripts/validate.py

# {Validation check 2}
{verification-command}

# {Integration test}
{integration-test-command}
```

---

## 🚨 Error handling

### Error classification

#### 1. {Error Category 1}

**Symptom**: {How this error manifests}

**cause**:
- {Possible cause 1}
- {Possible cause 2}

**Solution**:
```bash
# {Solution step 1}
{command-1}

# {Solution step 2}
{command-2}
```

---

#### 2. {Error Category 2}

**Symptom**: {Error description}

**Debugging**:
```bash
# {How to debug}
{debug-command}
```

**correction**:
1. {Fix step 1}
2. {Fix step 2}

---

### Logging and Debugging

**Log Location**: `{log-file-path}`

**Log level settings**:
```bash
# {How to enable debug logging}
{logging-config-command}
```

**Check log**:
```bash
# {How to view logs}
tail -f {log-file-path}
```

---

## 🔗 Related agents/commands

### Related commands

- **/{command-1}** - {How this skill supports the command}
- **/{command-2}** - {Integration point}

### Associated Agent

- **@agent-{agent-1}** - {How they work together}
- **@agent-{agent-2}** - {Collaboration scenario}

### Related skills

- **{skill-1}** - {Complementary functionality}
- **{skill-2}** - {When to use together}

---

## 📊 Performance and Metrics

### Performance characteristics

- **Execution time**: {Typical execution time}
- **Memory usage**: {Expected memory usage}
- **Disk I/O**: {File operations count}
- **Network**: {External API calls if any}

### Optimization tips

1. **{Optimization 1}**: {How to improve performance}
2. **{Optimization 2}**: {Configuration tweak}
3. **{Optimization 3}**: {Best practice}

---

## 🎓 Best Practices

### 1. {Practice Category 1}

**Recommendation**:
```{language}
# {Good practice example}
{recommended-code}
```

**What to avoid**:
```{language}
# {Anti-pattern example}
{avoid-this-code}
```

**Reason**: {Why this is best practice}

---

### 2. {Practice Category 2}

**Tip**: {Helpful tip}

**example**:
```bash
{example-of-best-practice}
```

---

### 3. {Practice Category 3}

**Caution**: {Important consideration}

---

## 📖 Practical examples

### Example 1: {Common Use Case}

**Purpose**: {What this example demonstrates}

**input**:
```{format}
{example-input}
```

**execution**:
```bash
{commands-to-run}
```

**output of power**:
```{format}
{example-output}
```

**Explanation**: {What happened and why}

---

### Example 2: {Advanced Use Case}

**Purpose**: {Advanced scenario}

**Scenario**: {Detailed scenario description}

**avatar**:
```{language}
{implementation-code}
```

**Result**: {What you achieve}

---

### Example 3: {Edge Case}

**Scenario**: {Unusual but important scenario}

**How ​​to handle**: {How skill handles this}

---

## 🔧 Customization

### Extension points

Areas where you can customize this skill to fit your project:

1. **{Extension Point 1}**
- File: `{file-to-modify}`
 - How to modify: {How to customize}

2. **{Extension Point 2}**
- Settings: `{config-key}`
 - Options: {Available options}

### Plugin system (advanced)

```python
# {How to create plugins for this skill}
{plugin-example-code}
```

---

## 📚 References

### Official Documentation
- **Claude Code Skills**: https://docs.claude.com/en/docs/claude-code/skills
- **{Related Doc}**: {URL}

### MoAI-ADK Resources
- **Development Guide**: `.moai/memory/development-guide.md`
- **SPEC Metadata**: `.moai/memory/spec-metadata.md`

### Community
- **GitHub Issues**: {Link}
- **Discussion**: {Link}

---

## 🔄 Update log

### v1.0.0 (Initial)
- {Feature 1 introduced}
- {Feature 2 implemented}
- {Initial release notes}

---

**Template Level**: Full
**Best For**: Production MoAI-ADK integration, enterprise workflows
**Features**:
- Alfred auto-selection
- Workflow integration
- Detailed settings
- Verification automation
- Error handling
- Performance optimization

**Directory Structure**: Full (SKILL.md + reference.md + examples.md + scripts/ + templates/)
**Estimated Setup Time**: 45-60 minutes
**Maintenance**: Regular updates as workflow evolves
**Support**: Full MoAI-ADK integration support

---

This skill provides the highest level of automation in {domain}.
