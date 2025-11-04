# MoAI-ADK Command Documentation: Practical Examples

This file shows **real examples** from existing MoAI-ADK commands with annotations explaining the standards.

---

## Example 1: YAML Frontmatter (from 1-plan.md)

### The Standard

```yaml
---
name: alfred:1-plan
description: "Define specifications and create development branch"
argument-hint: Title 1 Title 2 ... | SPEC-ID modifications
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Grep
  - Glob
  - TodoWrite
  - Bash(git:*)
  - Bash(gh:*)
  - Bash(rg:*)
  - Bash(mkdir:*)
---
```

### Why Each Field Matters

| Field | Value | Explanation |
|-------|-------|-------------|
| **name** | `alfred:1-plan` | Unique identifier used by Claude Code to invoke this command |
| **description** | "Define specifications..." | Single sentence, present tense, describes user benefit |
| **argument-hint** | `Title 1 Title 2 ... \| SPEC-ID...` | Shows format: positional args OR pipe-separated alternatives |
| **allowed-tools** | Array of patterns | Whitelist specific commands (Bash(git:*) not Bash(*)) |

### Anti-Pattern (❌ WRONG)

```yaml
---
name: 1-plan  # Missing "alfred:" prefix
description: "Creates specifications"  # Passive voice
argument-hint: "title" # Vague, no options shown
allowed-tools:
  - Bash(*)  # TOO BROAD! Should be Bash(git:*)
---
```

---

## Example 2: Batched Questions (from 0-project.md, lines 226-350)

### The Pattern

```python
AskUserQuestion(
    questions=[
        {
            "question": "Which language would you like to use for the project initialization and documentation?",
            "header": "Language",
            "multiSelect": false,
            "options": [
                {
                    "label": "🌍 English",
                    "description": "All dialogs and documentation in English"
                },
                {
                    "label": "🇰🇷 한국어",
                    "description": "All dialogs and documentation in Korean"
                }
            ]
        },
        {
            "question": "In which language should Alfred's sub-agent prompts be written?",
            "header": "Agent Prompt Language",
            "multiSelect": false,
            "options": [
                {
                    "label": "🌐 English (Global Standard)",
                    "description": "All sub-agent prompts in English for global consistency"
                },
                {
                    "label": "🗣️ Selected Language (Localized)",
                    "description": "All sub-agent prompts in the language you selected above"
                }
            ]
        },
        {
            "question": "How would you like to be called in our conversations?",
            "header": "Nickname",
            "multiSelect": false,
            "options": [
                {
                    "label": "Enter custom nickname",
                    "description": "Type your preferred name using the 'Other' option below"
                }
            ]
        }
    ]
)
```

### Why This Is Batched (Good UX)

**Before batching** (❌ 3 separate calls):
```
Turn 1: AskUserQuestion(Q1) → User selects
Turn 2: AskUserQuestion(Q2) → User selects
Turn 3: AskUserQuestion(Q3) → User selects
TOTAL: 3 turns (inefficient)
```

**After batching** (✅ 1 call):
```
Turn 1: AskUserQuestion([Q1, Q2, Q3]) → User selects all
TOTAL: 1 turn (66% reduction!)
```

### Response Processing (lines 228-294)

```markdown
**Q1: Language Selection**

Selected option → `.moai/config.json` storage:

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**Q2: Agent Prompt Language Selection**

```json
{
  "language": {
    "agent_prompt_language": "localized"
  }
}
```

**Q3: Nickname Input**

```json
{
  "user": {
    "nickname": "GOOS"
  }
}
```
```

**Key observation**: Each Q3 response maps to a specific config section!

---

## Example 3: Phase A/B Pattern (from 2-run.md, lines 99-195)

### Overview Section (Narrative)

```markdown
## 🔍 STEP 1: SPEC analysis and execution plan establishment

STEP 1 consists of **two independent phases** to provide flexible workflow based on task complexity:

### 📋 STEP 1 Workflow Overview

```
┌─────────────────────────────────────────────────────────────┐
│ STEP 1: SPEC Analysis & Planning                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Phase A (OPTIONAL)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ 🔍 Explore Agent                        │               │
│  │ • Browse existing codebase              │               │
│  │ • Find similar implementations          │               │
│  │ • Identify patterns & architecture      │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (exploration results)                              │
│                    ↓                                        │
│  Phase B (REQUIRED)                                         │
│  ┌─────────────────────────────────────────┐               │
│  │ ⚙️ implementation-planner Agent         │               │
│  │ • Analyze SPEC requirements             │               │
│  │ • Design execution strategy             │               │
│  │ • Create implementation plan            │               │
│  │ • Request user approval                 │               │
│  └─────────────────────────────────────────┘               │
│                    ↓                                        │
│          (user approval via AskUserQuestion)                │
│                    ↓                                        │
│              PROCEED TO STEP 2                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Points**:
- **Phase A is optional** - Skip if you don't need to explore existing code
- **Phase B is required** - Always runs to analyze SPEC and create execution plan
- **Results flow forward** - Exploration results (if any) are passed to implementation-planner
```

### Phase A: When to Use (Guidance)

```markdown
### 🔍 Phase A: Codebase Exploration (OPTIONAL)

**Use the Explore agent when you need to understand existing code before planning.**

#### When to use Phase A:

- ✅ Need to understand existing code structure/patterns
- ✅ Need to find similar function implementations for reference
- ✅ Need to understand project architectural rules
- ✅ Need to check libraries and versions being used

#### How to invoke Explore agent:

```
Invoking the Task tool (Explore agent):
- subagent_type: "Explore"
- description: "Explore existing code structures and patterns"
- prompt: "SPEC-$ARGUMENTS와 관련된 기존 코드를 탐색해주세요:
 - 유사한 기능 구현 코드 (src/)
 - 참고할 테스트 패턴 (tests/)
 - 아키텍처 패턴 및 디자인 패턴
 - 현재 라이브러리 및 버전 (package.json, requirements.txt)
 상세도 수준: medium"
```

**Note**: If you skip Phase A, proceed directly to Phase B.
```

### Phase B: Complete Task Tool (Code Example)

```markdown
### ⚙️ Phase B: Execution Planning (REQUIRED)

**Call the implementation-planner agent to analyze SPEC and establish execution strategy.**

This phase is **always required** regardless of whether Phase A was executed.

#### How to invoke implementation-planner:

```
Task tool call:
- subagent_type: "implementation-planner"
- description: "SPEC analysis and establishment of execution strategy"
- prompt: "$ARGUMENTS의 SPEC을 분석하고 실행 계획을 수립해주세요.
 다음을 포함해야 합니다:
 1. SPEC 요구사항 추출 및 복잡도 평가
 2. 라이브러리 및 도구 선택 (WebFetch 사용)
 3. TAG 체인 설계
 4. 단계별 실행 계획
 5. 위험 요소 및 대응 계획
 6. 행동 계획을 작성하고 `AskUserQuestion 도구`로 사용자와 다음 단계를 확인합니다
 (선택사항) 탐색 결과: $EXPLORE_RESULTS"
```

**Note**: If Phase A was executed, pass the exploration results via `$EXPLORE_RESULTS` variable.
```

### Key Observations

1. **Phase A**: Optional, clear when to use/skip
2. **Phase B**: Required, shows complete Task tool invocation
3. **Data flow**: Phase A results → `$EXPLORE_RESULTS` → Phase B
4. **Decision point**: Clear that Phase A is optional but Phase B always runs

---

## Example 4: Complete Decision Point (from 2-run.md, lines 234-294)

### Setup (Narrative)

```markdown
## Implementation Strategy Approval

After the execution plan is ready, Alfred uses `AskUserQuestion tool` to obtain explicit user approval before proceeding to TDD implementation.
```

### Complete AskUserQuestion (Code)

```python
AskUserQuestion(
    questions=[
        {
            "question": "Implementation plan is ready. How would you like to proceed?",
            "header": "Implementation Approval",
            "multiSelect": false,
            "options": [
                {
                    "label": "✅ Proceed with TDD",
                    "description": "Start RED → GREEN → REFACTOR cycle"
                },
                {
                    "label": "🔍 Research First",
                    "description": "Invoke Explore agent to study existing code patterns"
                },
                {
                    "label": "🔄 Modify Strategy",
                    "description": "Request changes to implementation approach"
                },
                {
                    "label": "⏸️ Postpone",
                    "description": "Save plan and return later"
                }
            ]
        }
    ]
)
```

### Response Processing (Mapped Actions)

```markdown
**Response Processing**:

- **"✅ Proceed with TDD"** (`answers["0"] === "✅ Proceed with TDD"`) → Execute Phase 2
  - Proceed directly to STEP 2 (TDD implementation)
  - Invoke tdd-implementer agent with approved plan
  - Begin RED phase (write failing tests)
  - Display: "🔴 Starting RED phase..."

- **"🔍 Research First"** (`answers["0"] === "🔍 Research First"`) → Run exploration first
  - Invoke Explore agent to analyze existing codebase
  - Pass exploration results to implementation-planner
  - Re-generate plan with research insights
  - Re-present plan for approval
  - Display: "🔍 Codebase exploration complete. Plan updated."

- **"🔄 Modify Strategy"** (`answers["0"] === "🔄 Modify Strategy"`) → Revise plan
  - Collect strategy modification requests from user
  - Update implementation plan with changes
  - Re-present for approval (recursive)
  - Display: "🔄 Plan modified. Please review updated strategy."

- **"⏸️ Postpone"** (`answers["0"] === "⏸️ Postpone"`) → Save and resume later
  - Save plan to `.moai/specs/SPEC-{ID}/plan.md`
  - Commit with message "plan(spec): Save implementation plan for SPEC-{ID}"
  - User can resume with `/alfred:2-run SPEC-{ID}`
  - Display: "⏸️ Plan saved. Resume with `/alfred:2-run SPEC-{ID}`"
```

### Key Observations

1. **Exact string matching**: `answers["0"] === "exact label"`
2. **Action description**: What happens for each response
3. **Concrete next steps**: "Execute Phase 2", "Invoke agent", etc.
4. **User communication**: "Display:" messages
5. **All 4 options mapped**: No orphaned options

---

## Example 5: Comparison Table (from 1-plan.md, lines 396-400)

### Pattern: Decision Matrix

```markdown
| 선택지 | 저장값 | 동작 | `/alfred:1-plan` 시 | 팀 협업 영향 |
|--------|--------|------|-------------------|-----------|
| 📋 Feature Branch + PR | `"feature_branch"` | 매 SPEC마다 feature/SPEC-{ID} 브랜치 생성 → PR 리뷰 → develop 병합 | 1. 브랜치 자동 생성<br>2. PR 템플릿 생성<br>3. 리뷰자 설정<br>4. Merge 후 삭제 | ✅ 최적: 팀 리뷰, 코드 추적, 감사 이력 완벽<br>⚠️ 약간의 workflow 오버헤드 |
| 🔄 Direct Commit to Develop | `"develop_direct"` | develop 브랜치에 직접 커밋 (브랜치 생성 생략) | 1. 브랜치 생성 생략<br>2. 직접 develop 커밋<br>3. conflict 시 사용자 수동 해결 | ✅ 빠름: 프로토타입, 개인 프로젝트 적합<br>❌ 팀 리뷰 불가, 이력 추적 어려움 |
| 🤔 Decide per SPEC | `"per_spec"` | SPEC마다 git-manager가 워크플로우 선택 요청 | 1. AskUserQuestion으로 사용자 선택 요청<br>2. 선택에 따라 1번 또는 2번 경로 실행 | 🔀 유연함: SPEC 특성에 따라 선택 가능<br>⚠️ 매번 결정 필요한 오버헤드 |
```

### Why Tables Work Well

1. **Side-by-side comparison** - Easy to read differences
2. **Decision making** - User can compare pros/cons
3. **Concrete values** - Shows what gets saved (`"feature_branch"`)
4. **Impact summary** - Shows consequences in team context

---

## Example 6: Narrative + Code Balance (from 1-plan.md)

### Pure Narrative Section (Lines 28-80)

```markdown
## 💡 Planning philosophy: "Always make a plan first and then proceed."

`/alfred:1-plan` is a general-purpose command that **creates a plan**, rather than simply "creating" a SPEC document.

### 3 main scenarios

#### Scenario 1: Creating a Plan (Primary Method) ⭐
```bash
/alfred:1-plan "User authentication function"
→ Refine idea
→ Requirements specification using EARS syntax
→ Create feature/SPEC-XXX branch
→ Create Draft PR
```

#### Scenario 2: Brainstorming
```bash
/alfred:1-plan "Payment system improvement idea"
→ Organizing and structuring ideas
→ Deriving requirements candidates
→ Technical review and risk analysis
```

#### Scenario 3: Improve existing SPEC
```bash
/alfred:1-plan "SPEC-AUTH-001 Security Enhancement"
→ Analyze existing plan
→ Establish improvement direction
→ Create new version plan
```
```

**Analysis**:
- 70% narrative explaining the "why" and "when"
- 30% code showing command examples (real invocations)
- Emoji marking primary approach (⭐)
- Clear user intent demonstrated in each scenario

### Code-Heavy Section (Lines 189-236)

```markdown
#### How to invoke spec-builder:

```
Call the Task tool:
- subagent_type: "spec-builder"
- description: "Analyze the plan and establish a plan"
- prompt: """당신은 spec-builder 에이전트입니다.

언어 설정:
- 대화_언어: {{CONVERSATION_LANGUAGE}}
- 언어명: {{CONVERSATION_LANGUAGE_NAME}}

중요 지시사항:
SPEC 문서는 이중 언어 구조를 따라야 합니다 (사용자 언어 + 영어 요약):

[...more instructions...]

작업:
프로젝트 문서를 분석하여 SPEC 후보자를 제시해주세요.
분석 모드로 실행하며, 다음을 포함해야 합니다:
1. product/structure/tech.md의 심층 분석
2. SPEC 후보자 식별 및 우선순위 결정
3. EARS 구조 설계
4. 사용자 승인 대기

사용자 입력: $ARGUMENTS
(선택사항) 탐색 결과: $EXPLORE_RESULTS"""
```
```

**Analysis**:
- 30% code showing exact Task tool invocation
- Variables preserved: `{{CONVERSATION_LANGUAGE}}`
- Complete prompt shown (copy-paste ready)
- Multi-line format with triple backticks
- Special variable documentation: `$ARGUMENTS`, `$EXPLORE_RESULTS`

---

## Key Takeaways

### ✅ Patterns That Work

1. **YAML first** - Define metadata before content
2. **70/30 split** - Balance narrative with code
3. **Phase A/B** - Clear optional vs required
4. **Batched questions** - Group related user interactions
5. **Complete invocations** - Copy-paste ready code
6. **Response mapping** - Exact string matching shown
7. **Tables** - Side-by-side comparisons
8. **ASCII diagrams** - Visual flow representation
9. **Real examples** - Actual command invocations
10. **Explicit language** - Clear next steps for user

### ❌ Patterns That Fail

1. ❌ Incomplete code examples (users can't copy-paste)
2. ❌ Pseudo-code mixed with real syntax (confusion)
3. ❌ AskUserQuestion without response mapping
4. ❌ Sequential questions instead of batching
5. ❌ Vague next steps ("you can proceed...")
6. ❌ Inconsistent emoji usage
7. ❌ Sections >200 lines without breaks
8. ❌ Placeholder values instead of real examples
9. ❌ Missing optional/required phase distinction
10. ❌ Code without explaining WHY or WHEN

