# Interactive Discovery: TUI-Based Skill Requirement Gathering

This guide covers using **moai-core-ask-user-questions** Skill to engage users through interactive dialogue and clarify Skill requirements.

---

## Why Interactive Discovery?

### The Problem: Vague Requests

```
❌ User says: "Create a Skill for testing"
   Questions: Which language? Which framework?
   Framework? Testing framework? Unit/integration/e2e?
   Target audience? Experience level?
   Result: Ambiguous Skill that doesn't match intent
```

### The Solution: Structured TUI Surveys

```
✅ skill-factory says: "Let me ask a few questions..."
   ↓
   Survey 1: Programming language (Python/TypeScript/Go/etc.)
   Survey 2: Testing scope (Unit/integration/e2e/all)
   Survey 3: Target experience (Beginner/intermediate/advanced)
   Survey 4: Must-have features (Mocking/fixtures/CI integration)
   ↓
   Result: Crystal-clear Skill requirements
```

---

## Interactive Discovery Process

### Phase 0a: Problem Definition

**Goal**: Understand what problem this Skill solves.

**TUI Survey Pattern 1: Domain Selection**

```
┌─ SKILL FACTORY: INTERACTIVE DISCOVERY ────────────────────┐
│                                                             │
│ What problem domain does this Skill address?              │
│                                                             │
│ ▶ Debugging & Troubleshooting                             │
│   Performance Analysis & Optimization                     │
│   Code Quality & Best Practices                           │
│   Testing & Quality Assurance                             │
│   Infrastructure & DevOps                                │
│   Data Processing & Transformation                       │
│   Security & Cryptography                                │
│   Other / Custom: [text input]                           │
│                                                             │
│ [↑↓ Navigate] [Enter: Select] [Esc: Cancel]              │
└─────────────────────────────────────────────────────────────┘
```

**When to use this survey**:
- First question to narrow scope
- Establishes technology/domain context
- Helps with WebSearch queries later

**Follow-up question**:
```
┌─ SKILL FACTORY: FRAMEWORK/TECHNOLOGY ─────────────────────┐
│                                                             │
│ Within [Domain], which technology/framework?              │
│                                                             │
│ ▶ Python (Flask/Django/FastAPI/etc.)                     │
│   TypeScript (Express/Next.js/NestJS/etc.)               │
│   Go (Standard library/Gin/etc.)                         │
│   Rust (Axum/Actix/etc.)                                │
│   Java (Spring/Quarkus/etc.)                            │
│   Other: [text input]                                   │
│                                                             │
│ [↑↓ Navigate] [Enter: Select]                            │
└─────────────────────────────────────────────────────────────┘
```

**Outcome**: `selected_domain` + `selected_tech` captured

---

### Phase 0b: Scope Clarification

**Goal**: Define what's in scope vs. out of scope.

**TUI Survey Pattern 2: Feature Priority (Multi-select)**

```
┌─ SKILL FACTORY: FEATURE PRIORITY ──────────────────────────┐
│                                                             │
│ Which features are MUST-HAVE? (Select multiple)          │
│                                                             │
│ ☑ Core concepts & fundamentals                          │
│ ☑ Practical examples & patterns                         │
│ ☐ Error handling & debugging                            │
│ ☑ Performance optimization                              │
│ ☐ Security best practices                               │
│ ☐ Deployment & CI/CD integration                        │
│ ☐ Testing strategies                                    │
│ ☐ Monitoring & observability                            │
│                                                             │
│ [↑↓ Navigate] [Space: Toggle] [Enter: Continue]          │
└────────────────────────────────────────────────────────────┘
```

**Selection capture**:
```python
must_have_features = [
    "core_concepts",
    "practical_examples",
    "performance_optimization"
]
```

**Outcome**: Feature priority list

---

### Phase 0c: Target Audience

**Goal**: Understand who will use this Skill.

**TUI Survey Pattern 3: Experience Level**

```
┌─ SKILL FACTORY: TARGET AUDIENCE ──────────────────────────┐
│                                                             │
│ Who is the PRIMARY audience?                              │
│                                                             │
│   Beginners (< 1 year experience)                        │
│   Intermediate (1-3 years)                               │
│ ▶ Advanced (3+ years)                                    │
│   All levels (mixed audience)                            │
│   Hiring managers / Non-technical                       │
│                                                             │
│ [↑↓ Navigate] [Enter: Select]                            │
└────────────────────────────────────────────────────────────┘
```

**Outcome**: `target_audience_level` captured

---

### Phase 0d: Special Considerations

**Goal**: Capture any unique requirements or constraints.

**TUI Survey Pattern 4: Version & Maturity**

```
┌─ SKILL FACTORY: FRAMEWORK VERSION ────────────────────────┐
│                                                             │
│ Which version of [Framework]?                             │
│                                                             │
│   Latest stable (auto-detect)                            │
│ ▶ Latest LTS (Long-term support)                         │
│   Specific version: [text input]                         │
│   "Support multiple versions (e.g., 3.9+)"              │
│                                                             │
│ [↑↓ Navigate] [Enter: Select]                            │
└────────────────────────────────────────────────────────────┘
```

**Outcome**: `framework_version` specified

---

### Phase 0e: Requirements Summary & Confirmation

After all surveys, present summary for user confirmation:

```
┌─ SKILL FACTORY: REVIEW YOUR SELECTIONS ───────────────────┐
│                                                             │
│ ✓ Problem Domain: Performance Analysis & Optimization   │
│ ✓ Technology: Python + Profiling Tools                  │
│ ✓ Must-Have: Core concepts, examples, optimization tips│
│ ✓ Audience: Advanced (3+ years)                         │
│ ✓ Framework: Python 3.12 LTS                           │
│                                                             │
│ Ready to begin research and Skill creation?              │
│                                                             │
│  [✓ Yes, proceed] [← Go back, modify]                   │
└────────────────────────────────────────────────────────────┘
```

**Outcome**: Confirmed `SkillRequirements` charter

---

### Phase 0f: Version Awareness & Stability Matrix (NEW)

**Goal**: Capture version requirements to ensure all Skills reference latest stable versions (as of 2025-10-22).

**CRITICAL**: All Skills MUST specify exact tool/framework versions. Outdated version references create stale content.

**TUI Survey Pattern 5: Stable Version Requirements**

```
┌─ SKILL FACTORY: VERSION MATRIX ───────────────────────────────┐
│                                                                 │
│ For [Framework], which tool versions should we reference?    │
│ (Defaults shown: 2025-10-22 stable versions)                 │
│                                                                 │
│ Framework Versions:                                            │
│ • Python:     3.13.1 (latest)  ▶ 3.12.7 (LTS)                │
│ • Node.js:    23.0.0  (latest)  ▶ 22.11.0 (LTS)              │
│ • Go:         1.23.3  (latest)  ▶ 1.22.x (older)             │
│ • Rust:       1.82.0  (latest)  ▶ MSRV: 1.70+                │
│ • Java:       23.0.1  (latest)  ▶ 21.0.x (LTS)               │
│                                                                 │
│ Tool Versions (Python ecosystem example):                      │
│ ☑ pytest 8.4.2 (testing framework)                           │
│ ☑ ruff 0.13.1 (linting/formatting)                           │
│ ☑ uv 0.9.3 (package manager)                                 │
│ ☑ mypy 1.8.0 (type checking)                                 │
│ ☑ FastAPI 0.115.0 (web framework)                            │
│ ☑ Pydantic 2.7.0 (data validation)                           │
│                                                                 │
│ [↑↓ Navigate] [Space: Toggle] [Enter: Continue]              │
│                                                                 │
│ Want to customize versions? [Yes] / [No, use defaults]        │
└─────────────────────────────────────────────────────────────────┘
```

**Version Selection Capture**:
```python
framework_versions = {
    "python": "3.13.1",
    "tools": [
        "pytest:8.4.2",
        "ruff:0.13.1",
        "uv:0.9.3",
        "mypy:1.8.0"
    ]
}
```

**Available Stable Version Matrix** (referenced as source of truth):

| Framework   | Latest | LTS        | EOL Date      | Ref                                    |
| ----------- | ------ | ---------- | ------------- | -------------------------------------- |
| **Python**  | 3.13.1 | 3.12.7     | 3.12: 2028-10 | [python.org](https://python.org)       |
| **Node.js** | 23.0.0 | 22.11.0    | 22.x: 2025-10 | [nodejs.org](https://nodejs.org)       |
| **Go**      | 1.23.3 | 1.22.x     | 1.22: 2025-02 | [golang.org](https://golang.org)       |
| **Rust**    | 1.82.0 | MSRV: 1.70 | N/A           | [rust-lang.org](https://rust-lang.org) |
| **Java**    | 23.0.1 | 21.0.x     | 21: 2028-09   | [oracle.com](https://oracle.com)       |

**When This Survey Is Needed**:
- ✅ Creating a new Skill (always)
- ✅ Updating existing Skill (quarterly version checks)
- ✅ Migrating to new major version (breaking changes detected)
- ✅ Security updates required (security advisory found)

**Related Document**: `.moai/memory/VERSION-TRACKING.md` — Comprehensive version matrix for all 23 languages and 50+ frameworks

**Outcome**: `stable_versions` captured with explicit tool/framework version numbers

---

### Phase 0g: Final Confirmation with Version Summary

Enhanced confirmation step that includes version matrix:

```
┌─ SKILL FACTORY: FINAL REVIEW ────────────────────────────────┐
│                                                                 │
│ ✓ Problem Domain: Testing & Quality Assurance               │
│ ✓ Technology: TypeScript + Vitest                           │
│ ✓ Must-Have: Core, patterns, mocking, optimization          │
│ ✓ Audience: Intermediate (1-3 years)                        │
│ ✓ Framework Version: Vitest 2.0.5 (latest)                 │
│ ✓ Tool Versions:                                             │
│   • TypeScript: 5.3.x                                        │
│   • Vitest: 2.0.5                                            │
│   • Biome: 1.7.x (linter/formatter)                         │
│   • Node.js: 22.11.0 (LTS recommended)                      │
│                                                                 │
│ ✓ Research Focus: Latest 2025 patterns with these versions   │
│                                                                 │
│ Ready to research and generate Skill?                         │
│ (This will reference current stable versions)                │
│                                                                 │
│  [✓ Yes, proceed] [← Go back, modify versions]              │
└─────────────────────────────────────────────────────────────────┘
```

**Outcome**: Confirmed `SkillRequirements` charter with **explicit version matrix**

---

## TUI Survey Design Best Practices

### 1. Progressive Narrowing

```
Survey 1: Broad domain (5-8 options)
   ↓ User selects one
Survey 2: Specific tech within domain (5-8 options)
   ↓ User selects one or more
Survey 3: Features/scope (multi-select from pool)
   ↓ User checks boxes
Survey 4: Advanced options (if needed, 3-4 options)
   ↓ User selects one
Survey 5: Confirmation summary
   ↓ User confirms or goes back
```

### 2. Sensible Defaults

```
❌ Bad (requires user to know exact answer):
  □ How many lines of code per section?
  □ Which specific exception types?
  □ What is the exact API response format?

✅ Good (offers reasonable defaults):
  □ Code size: Small (under 50 lines) / Medium / Large
  □ Exception handling: Basic (try/except) / Advanced (custom types)
  □ API responses: JSON (default) / XML / Protocol Buffers
```

### 3. Custom Input Fallback

```
┌─ Survey ──────────────────────────────────────────────────┐
│                                                             │
│ ▶ Option 1 (predefined)                                 │
│   Option 2 (predefined)                                 │
│   Option 3 (predefined)                                 │
│   Other / Custom: [________________]  ← For unknown cases│
│                                                             │
└───────────────────────────────────────────────────────────┘
```

### 4. Multi-Select for Complex Requirements

```
Question: "Which testing patterns are important?"
(Users can select multiple)

☑ Unit testing
☑ Integration testing
☐ End-to-end testing
☑ Mocking & fixtures
☐ Performance testing
```

**Result**: More accurate feature prioritization

---

## Example: Complete Interactive Discovery Flow

### Scenario: "I want a Skill for TypeScript testing"

**Step 1: Problem Domain** → User selects "Testing & Quality Assurance"

**Step 2: Technology** → User selects "TypeScript"

**Step 3: Testing Framework** → User selects "Vitest"

**Step 4: Must-Have Features**
- Core concepts ✓
- Unit testing patterns ✓
- Mocking with Vitest ✓
- Performance optimization ✓
- CI/CD integration ✗

**Step 5: Audience** → "Intermediate (1-3 years)"

**Step 6: Framework Version** → "Latest stable"

**Step 7: Confirmation**
```
✓ Domain: Testing & Quality Assurance
✓ Technology: TypeScript + Vitest
✓ Features: Core, patterns, mocking, optimization
✓ Audience: Intermediate
✓ Version: Vitest latest stable

Ready to research and generate Skill?
```

**Result**: skill-factory now has precise requirements for web research and Skill generation

---

## Failure Modes & Recovery

### 🔴 User Provides No/Invalid Selection

```
User: [Presses Escape 3 times]
system: "It seems you'd like to cancel. Should I stop?"

Recovery:
1. Offer explicit yes/no confirmation
2. Or: Restart with fresh survey
3. Or: Use defaults and proceed
```

### 🟡 User Selects Contradictory Options

```
User selects:
  Domain: "Data Processing"
  Technology: "Kubernetes"
  ↑ Mismatch! Kubernetes is DevOps, not data processing

Detection: skill-factory notices mismatch
Response: "I noticed you selected Data Processing but Kubernetes
 is typically DevOps-related. Did you mean... Kafka or Spark?"
```

### 🟢 User Enters Custom Value

```
User selects: "Other / Custom"
skill-factory: "What specific technology would you like?"
User: "MicroPython for IoT devices"
Result: WebSearch targets "MicroPython IoT best practices 2025"
```

---

## Integration with WebSearch Phase

### From Interactive Discovery → To WebSearch

```
TUI Survey Captures:
├─ domain: "Testing & Quality Assurance"
├─ technology: "TypeScript"
├─ framework: "Vitest"
├─ features: ["core", "patterns", "mocking", "optimization"]
├─ audience: "intermediate"
└─ version: "latest-stable"

↓

skill-factory builds WebSearch queries:
1. "TypeScript Vitest latest 2025 best practices"
2. "Vitest async testing patterns modern"
3. "TypeScript testing mocking strategies"
4. "Vitest performance optimization official docs"
5. "TypeScript Vitest breaking changes migration"

↓

WebSearch findings INFORM the Skill
(See WEB-RESEARCH.md for details)
```

---

## Effective Question Design

### DO: Ask Open-Ended Questions

```
✅ "What specific problem do you want to solve?"
   (Allows detailed answer)

❌ "Do you want testing?"
   (Yes/No is too limiting)
```

### DO: Provide Context

```
✅ "Target audience: who will use this?
   (Beginner developers might need more examples)
   - Beginners
   - Intermediate
   - Advanced"

❌ "Target audience?"
   (Too vague, user doesn't know what you need)
```

### DO: Use Sensible Grouping

```
✅ Survey 1: Pick ONE (narrowing)
   Survey 2: Pick ONE (further narrowing)
   Survey 3: Pick MULTIPLE (features/scope)
   Survey 4: Confirmation

❌ Survey 1: Pick MULTIPLE (confusing)
   Survey 2: Pick ONE (conflicting)
```

---

## Capturing Requirements for WebSearch

Map survey answers to research queries:

| Survey Answer | WebSearch Topic | Example Query                       |
| ------------- | --------------- | ----------------------------------- |
| Python 3.12   | Latest version  | "Python 3.12 best practices 2025"   |
| Async testing | Modern pattern  | "async/await testing patterns 2025" |
| Performance   | Optimization    | "performance profiling tools 2025"  |
| Security      | Compliance      | "OWASP security patterns 2025"      |

---

## Related Resources

- [SKILL.md](SKILL.md) — Main Skill framework
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Using discoveries in research
- [skill-factory.md](../../agents/alfred/skill-factory.md) — Sub-Agent orchestration
- `moai-core-ask-user-questions` Skill — TUI implementation

---

**Version**: 0.1.0
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Skills + skill-factory
