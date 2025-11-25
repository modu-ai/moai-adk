# 📚 Skill Factory Parallel Analysis - Step-by-Step Detailed Guide

**Goal**: Use skill-factory agents to analyze existing skill folders in parallel and derive improvement items

---

## 🎯 Overall Flow (3 Steps)

```
Step 1️⃣  → Select analysis targets (Tier representative skills)
         ↓
Step 2️⃣  → Execute parallel agent analysis (simultaneous execution)
         ↓
Step 3️⃣  → Integrate results and create report
```

---

## Step 1️⃣: Select Analysis Targets

### Objectives
✅ Select representative skills for each tier
✅ Finalize 4 skills to analyze
✅ Define analysis criteria

### Execution Method

#### 1-1) Understand Skill Folder Structure
```bash
# Command
find .claude/skills -type d -name "moai-*" | sort

# Output example
.claude/skills/moai-foundation-trust
.claude/skills/moai-foundation-tags
.claude/skills/moai-alfred-tag-scanning
.claude/skills/moai-domain-backend
.claude/skills/moai-lang-python
... (50+ skills)
```

#### 1-2) Categorize by Tier
```
📦 Foundation Tier (6 skills)
  ├─ moai-foundation-trust       ← Selected ⭐
  ├─ moai-foundation-tags
  ├─ moai-foundation-specs
  ├─ moai-foundation-ears
  ├─ moai-foundation-git
  └─ moai-foundation-langs

📦 Alfred Tier (11 skills)
  ├─ moai-alfred-tag-scanning    ← Selected ⭐
  ├─ moai-alfred-code-reviewer
  ├─ moai-alfred-debugger-pro
  ├─ moai-alfred-language-detection
  └─ ... (7 more)

📦 Domain Tier (10 skills)
  ├─ moai-domain-backend         ← Selected ⭐
  ├─ moai-domain-frontend
  ├─ moai-domain-web-api
  ├─ moai-domain-security
  └─ ... (6 more)

📦 Language Tier (23 skills)
  ├─ moai-lang-python            ← Selected ⭐
  ├─ moai-lang-typescript
  ├─ moai-lang-go
  ├─ moai-lang-rust
  └─ ... (19 more)
```

#### 1-3) Selection Criteria
| Criteria | Description | Application |
|----------|-------------|-------------|
| **Importance** | Impact on overall project | Foundation > Alfred > Domain > Language |
| **Complexity** | Document size and concept scope | Higher increases analysis value |
| **Connectivity** | Relationship with other skills | Clear integration points |

### Result: Selected Skills (4 skills)
```
✅ moai-foundation-trust      (Foundation, core quality principles)
✅ moai-alfred-tag-scanning   (Alfred, tracking system)
✅ moai-domain-backend        (Domain, architecture)
✅ moai-lang-python           (Language, latest standards)
```

---

## Step 2️⃣: Execute Parallel Agent Analysis

### Objectives
✅ Analyze 4 skills **simultaneously**
✅ Generate completion scores for each skill
✅ Specify improvement items

### Execution Method

#### 2-1) Execute Parallel Analysis Agents

**Important**: Call the following 4 Tasks in **the same message** at once
(This causes Claude Code to execute them in parallel)

```python
# Pseudocode: Parallel execution of 4 agents

Agent 1: Task(
  subagent_type="general-purpose",
  description="Foundation: Trust skill analysis",
  prompt="""
  File: /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-foundation-trust/SKILL.md

  Analysis items:
  1. Metadata (name, version, description)
  2. Document structure (sections, headings)
  3. Core content (TRUST 5 principles)
  4. Code examples presence
  5. Completion score (0-100)

  Return in JSON format
  """
)

Agent 2: Task(
  subagent_type="general-purpose",
  description="Alfred: Tag-scanning skill analysis",
  prompt="..." # Same structure, different file
)

Agent 3: Task(
  subagent_type="general-purpose",
  description="Domain: Backend skill analysis",
  prompt="..." # Same structure, different file
)

Agent 4: Task(
  subagent_type="general-purpose",
  description="Language: Python skill analysis",
  prompt="..." # Same structure, different file
)
```

#### 2-2) Analysis Items for Each Agent

Each agent analyzes the following items:

```json
{
  "skill_name": "Skill name",
  "category": "Tier name",
  "metadata": {
    "name": "...",
    "description": "...",
    "allowed_tools": ["..."],
    "auto_load": "...",
    "trigger_cues": "..."
  },
  "structure": {
    "sections": ["Metadata", "What it does", ...],
    "total_lines": 100,
    "has_yaml_frontmatter": true,
    "has_code_examples": true
  },
  "content_score": 75,
  "findings": [
    "✅ Strength 1",
    "✅ Strength 2",
    "⚠️ Weakness 1",
    "❌ Critical issue"
  ],
  "recommendations": [
    "Improvement item 1",
    "Improvement item 2"
  ]
}
```

#### 2-3) Benefits of Parallel Execution

| Aspect | Sequential | Parallel | Improvement |
|--------|------------|----------|-------------|
| **Time Required** | 60 minutes | 15 minutes | ⬇️ 75% reduction |
| **System Efficiency** | 1 agent | 4 simultaneous | ⬆️ 4x |
| **Analysis Consistency** | Time bias | Simultaneity guaranteed | ⬆️ High |
| **Cross-Tier Patterns** | Limited | Comprehensive | ⬆️ Strong |

#### 2-4) Agent Execution Confirmation

Each agent signals completion with the following:

```bash
✅ Agent 1 Complete: Foundation Trust analysis (JSON returned)
✅ Agent 2 Complete: Alfred Tag-scanning analysis (JSON returned)
✅ Agent 3 Complete: Domain Backend analysis (JSON returned)
✅ Agent 4 Complete: Language Python analysis (JSON returned)

# Wait until all agents complete simultaneously
```

### Execution Result Examples

#### Agent 1 Output (Foundation Trust)
```json
{
  "skill_name": "moai-foundation-trust",
  "category": "Foundation",
  "content_score": 75,
  "findings": [
    "✅ Perfect YAML metadata",
    "✅ Clear TRUST 5 principles",
    "⚠️ Lack of specific verification commands"
  ],
  "recommendations": [
    "Add language-specific TRUST verification command matrix",
    "Provide automated verification script templates"
  ]
}
```

#### Agent 2 Output (Alfred Tag-scanning)
```json
{
  "skill_name": "moai-alfred-tag-scanning",
  "category": "Alfred",
  "content_score": 68,
  "findings": [
    "✅ Clear CODE-FIRST principle",
    "❌ Missing template file (templates/tag-inventory-template.md)",
    "⚠️ Examples are boilerplate"
  ],
  "recommendations": [
    "[CRITICAL] Create missing template file",
    "[HIGH] Detail how-it-works algorithm"
  ]
}
```

#### Agents 3 & 4 (Omitted, similar structure)

---

## Step 3️⃣: Integrate Results and Create Report

### Objectives
✅ Integrate 4 analysis results
✅ Discover tier-by-tier patterns
✅ Derive actionable recommendations

### Execution Method

#### 3-1) Score Aggregation
```
Foundation (Trust):     75/100
Alfred (Tag-scanning):  68/100
Domain (Backend):       75/100
Language (Python):      85/100
────────────────────────────
Average:                75.75/100  (B+ grade)
```

#### 3-2) Cross-Tier Pattern Analysis

**Common Strengths** (All skills):
```
✅ Standardized metadata structure (YAML frontmatter)
✅ Consistent documentation system (13 standard sections)
✅ Explicit connections to other skills
✅ Foundation principle compliance
```

**Common Weaknesses** (All skills):
```
❌ Lack of code examples (theory-focused)
❌ Insufficient workflow guides (tool listings)
❌ Missing error handling guides
❌ No templates/scripts provided
❌ Low practicality of examples
```

#### 3-3) Improvement Priority Matrix

```
        High Impact
            ▲
            │
    TAG-    │  TRUST
    scanning│  ★    Python
    ★★★    │  ★★   (Complete)
            │
    Backend │
    ★★     │
            │
    ────────┼──────────► Low Effort
          Low   High

Legend:
★★★ = High impact, low effort (Immediate action)
★★  = Medium impact, effort needed (Planned)
★   = Low impact but optional (When available)
```

#### 3-4) Action Plan Development

##### Phase 1: Urgent (1 week)
```
Priority 1: TAG-scanning missing template creation
  File: templates/tag-inventory-template.md
  Content: TAG inventory samples + normal/broken TAG examples
  Effort: 2-3 hours
  ROI: Very high (completion 68→85)

Priority 2: Trust verification commands addition
  Items: Language-specific TRUST verification command matrix
  Examples: Python pytest, Go go test, Rust cargo test
  Effort: 2-3 hours
  ROI: High (applies to all projects)

Priority 3: Backend code examples addition
  Count: 5 examples (1 per language)
  Examples: Python FastAPI, Go Gin, Node.js Express
  Effort: 3-4 hours
  ROI: Medium (applies to backend projects only)
```

##### Phase 2: In Progress (2 weeks)
```
- Add 5 real TAG-scanning use cases
- Trust automated verification script templates
- Backend security section creation (JWT, RBAC, Secrets)
```

##### Phase 3: Ongoing (1 month)
```
- Add error handling guides to all skills
- Integrate CI/CD pipeline examples
- Write verification tests for each skill
```

#### 3-5) Report Creation

Organize results into a Markdown document:

```markdown
# Parallel Analysis Report

## Executive Summary
- Analysis targets: 4 skills (Foundation, Alfred, Domain, Language)
- Average score: 75.75/100 (B+)
- Overall assessment: Structurally solid but needs practical guide enhancement

## Detailed Analysis Results
### 1. Foundation - Trust (75/100)
- Strengths: Perfect YAML structure, clear TRUST 5 principles
- Weaknesses: Lack of verification commands, abstract examples

### 2. Alfred - Tag-scanning (68/100)
- Strengths: Clear CODE-FIRST principle
- Weaknesses: Missing template file, undocumented algorithm

... (continued)

## Cross-Tier Patterns
- Common strengths: Metadata standardization, documentation system
- Common weaknesses: Lack of code examples, insufficient workflows

## Action Plan
Phase 1 (1 week): Urgent - Templates, commands, code examples
Phase 2 (2 weeks): In progress - Use cases, security, scripts
Phase 3 (1 month): Ongoing - Error handling, CI/CD, tests
```

### File Creation
```bash
File: .claude/skills/moai-skill-factory/PARALLEL-ANALYSIS-REPORT.md
Size: About 600-800 lines
Content: Comprehensive analysis, tier-by-tier evaluation, improvements, action plan
```

---

## 🔄 Actual Execution Example

### Command Line Execution (Pseudocode)
```bash
# Step 1: Check skill folders
$ ls -la .claude/skills/ | grep moai-

# Step 2: Execute 4 agents in parallel
# (Call simultaneously in Claude Code)

# Step 3: Collect results
Agent 1: ✅ Foundation Trust complete (JSON)
Agent 2: ✅ Alfred Tag-scanning complete (JSON)
Agent 3: ✅ Domain Backend complete (JSON)
Agent 4: ✅ Language Python complete (JSON)

# Step 4: Create report
$ echo "Integrating parallel analysis results..."
$ cat > PARALLEL-ANALYSIS-REPORT.md << EOF
# Analysis Report
(Integrate 4 analysis results)
EOF
```

---

## 📊 Analysis Metrics (Understanding)

### Completion Score Interpretation
```
90-100: ⭐⭐⭐ Model case (production-ready)
80-89:  ⭐⭐   Excellent (minor improvements needed)
70-79:  ⭐    Needs improvement (has key content but insufficient)
60-69:  ⚠️    Incomplete (structure only, lacks content)
<60:    🔴    Incomplete (needs rework)
```

### Score Components
```
Metadata:        20% (structural foundation)
Document structure: 15% (standard compliance)
Content depth:    25% (concept explanation)
Code examples:    20% (practicality)
Workflow:        10% (usage guide)
Error handling:   10% (recovery procedures)
────────────────────
Total:           100%
```

---

## 🎓 Key Learning Points

### Strengths of Parallel Analysis
1. **Time Efficiency**: Sequential 60 minutes → Parallel 15 minutes (4x improvement)
2. **Consistency**: Simultaneous evaluation with same criteria
3. **Pattern Discovery**: Cross-Tier comparison reveals common weaknesses
4. **Prioritization**: Impact analysis for ROI optimization

### Role of Skill-factory Agents
- YAML structure validation
- Document standard compliance evaluation
- Objective completion score calculation
- Specific improvement item identification and prioritization

### Execution of Improvement Recommendations
1. **Code Examples**: Minimum 3 per language
2. **Templates**: Provide pyproject.toml, pytest.ini references
3. **Workflows**: RED→GREEN→REFACTOR TDD cycle
4. **Error Handling**: Common error pattern cataloging

---

## ✅ Completion Checklist

- [ ] Step 1: Selection of 4 analysis target skills complete
- [ ] Step 2: Parallel execution of 4 agents complete
- [ ] Step 3: Result integration and report creation complete
- [ ] Action Plan Phase 1 review complete
- [ ] Improvement item priority confirmation complete

---

**Next Steps**: Review PARALLEL-ANALYSIS-REPORT.md → Execute Phase 1 Action Plan

