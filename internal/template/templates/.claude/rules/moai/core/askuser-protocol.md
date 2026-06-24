---
description: Canonical reference for AskUserQuestion-only interaction protocol, ToolSearch deferred-tool preload procedure, and Socratic interview standards
globs:
---

# AskUserQuestion Protocol — Canonical Reference

> This file is the **single source of truth** for AskUserQuestion interaction rules.
> Cross-referenced by: CLAUDE.md §8, moai-constitution.md §MoAI Orchestrator, agent-common-protocol.md §User Interaction Boundary, output-styles/moai/moai.md §3/§10.

---

## Channel Monopoly

**AskUserQuestion is the only user-facing question channel.** The MoAI orchestrator MUST route every user-facing question through an `AskUserQuestion` tool invocation. Free-form interrogative prose in the response body is **prohibited** as a question channel.

Applies to all orchestrator turns involving:
- Clarification questions when user intent is ambiguous (Stage 1 Clarify)
- Preference and decision questions ("Which approach?", "Continue or abort?")
- Socratic interview rounds during Context-First Discovery (CLAUDE.md §7 Rule 5)
- Branch and workflow selection
- Conflict resolution (merge strategy, rollback confirmation, etc.)

**Exceptions** (free-form prose questions permitted ONLY when):
- `AskUserQuestion` is technically unavailable — this should not occur in normal orchestrator operation
- The expression is a statement of status that happens to end with a question mark, not a genuine request for a decision

**Anti-pattern (NEVER repeat)**:
```
# Wrong — free-form prose question in response body
다음 진행 방향을 알려주세요:
- A: 지금 즉시 시작
- B: PR ready + 세션 종료
```

**Correct pattern**: Always use `AskUserQuestion`. See §Free-form Circumvention Prohibition for the "Other" option mechanism.

---

## ToolSearch Preload Procedure

`AskUserQuestion` is a **deferred tool** in Claude Code. Its JSON schema is NOT loaded into the active context at agent initialization time. Attempting to invoke `AskUserQuestion` without first selecting it results in `InputValidationError: tool not in schema`.

### Mandatory Preload Step

Immediately before **every** `AskUserQuestion` call, the orchestrator MUST invoke:

```
ToolSearch(query: "select:AskUserQuestion")
```

This loads the tool schema into the active context and makes the subsequent `AskUserQuestion` call valid.

### General Rule for Deferred Tools

Any deferred tool requires a `ToolSearch` select preload before invocation. The pattern generalizes:

```
ToolSearch(query: "select:<tool>[,<tool>...]")
```

- Single tool: `ToolSearch(query: "select:AskUserQuestion")`
- Multiple tools: `ToolSearch(query: "select:AskUserQuestion,TaskCreate")`

### Preload Sequence

```
[Turn N]
Step 1: ToolSearch(query: "select:AskUserQuestion")   ← preload deferred schema
Step 2: AskUserQuestion({ questions: [...] })           ← now valid to invoke
```

Never reverse or omit Step 1. The preload applies per-turn — if a new turn begins and `AskUserQuestion` will be called again, preload again.

---

## Socratic Interview Structure

When a Stage 1 Clarify trigger is satisfied (see §Ambiguity Triggers and Exceptions), the orchestrator conducts a **Socratic interview** through sequential `AskUserQuestion` rounds.

### Structural Constraints (all mandatory)

1. **Round limit**: Maximum 4 questions per `AskUserQuestion` call (Claude Code hard limit)
2. **Option limit**: Maximum 4 options per question (Claude Code hard limit)
3. **First option label**: MUST carry the `(권장)` (Korean) or `(Recommended)` (English) suffix to signal the recommended choice
4. **Language**: All question text, option labels, and option descriptions MUST be in the user's `conversation_language` (read from `.moai/config/sections/language.yaml`)
5. **Round progression**: Each subsequent round MUST narrow ambiguity by building on previous answers — repeating the same question is prohibited
6. **Termination condition**: Rounds continue until intent clarity reaches 100%; the interview MUST NOT end prematurely
7. **Pre-execution confirmation**: After clarity is achieved, consolidate findings into a brief report and obtain **explicit final confirmation** via `AskUserQuestion` before irreversible actions

### Round Structure Example

```
Round 1: ToolSearch → AskUserQuestion (scope questions)
Round 2: ToolSearch → AskUserQuestion (approach questions, built on Round 1 answers)
...
Final:   ToolSearch → AskUserQuestion (confirmation: "Proceed with this plan?")
```

---

## Option Description Standards

Every option in an `AskUserQuestion` call MUST have a `description` field populated with sufficient detail for the user to evaluate implications and trade-offs **without consulting external context**.

### Required description content

Each option description MUST include:

1. **Immediate result**: What happens immediately if this option is selected
2. **Side effects and risks**: Any follow-on consequences, risks, or irreversibility
3. **Quantitative information** (where applicable): Token cost, latency, file count, etc. (e.g., "saves ~30K tokens", "modifies 5 files")

### Bias Prevention

**bias prevention rule**: Option descriptions MUST use neutral, factual language — no persuasive or deprecating tone
- The recommendation signal is conveyed **exclusively** through the `(권장)` / `(Recommended)` label suffix on the first option
- Descriptions must not phrase the recommended option more favorably or the non-recommended options more negatively than the facts justify

**Anti-pattern**: Writing a description that says "This is the best approach because..." — that is bias. State facts only.

---

## Recommendation Placement Principles

AskUserQuestion의 `(권장)` 라벨은 **사용자가 통계적으로 다수 선택한 합리적 기본값**(선호 메모리에서 관측)에 근거해야 하며, 단순히 시스템이 밀고 싶은 정책 기본값이어서는 안 된다. 본 절은 추천 배치 5원칙을 정의한다.

### 1. 발화 시점 — 정보이익 정렬 (Fisher 정보 I=p(1−p))

**Where** 오케스트레이터가 다가오는 결정에 대해 불확실성 p를 추정하고, **When** p ≈ 0.5 (Fisher 정보 I=p(1−p) 최대, 결정 경계)이면, 오케스트레이터는 해당 질문을 AskUserQuestion으로 발화해야 한다. **While** p가 0 또는 1에 가까우면(거의 확정), 오케스트레이터는 통계적 다수 옵션으로 자동 처리하고 질문을 생략한다.

- p 추정(초기 휴리스틱): 동일 도메인의 관측된 다수 선택 비율. cold-start(관측 부족)는 p ≈ 0.5로 취급해 발화.
- 근거: just-in-time 결정경계 질문 원칙 (Fisher 정보 I=p(1−p)는 p=0.5에서 최대).

### 2. 질문 순서 — 정보이익 내림차순

**Where** 하나의 AskUserQuestion 호출에 여러 질문이 배치되면, 오케스트레이터는 각 질문의 추정 정보이익을 내림차순으로 정렬한다 (가장 높은 정보이익 질문이 첫 번째).

- 근거: 높은 정보이익 질문을 먼저 배치하면 사용자가 낮은 가치 질문을 만나기 전에 핵심 의사결정을 완료할 수 있다.

### 3. 추천 옵션 — 통계적 다수 합리적 기본값 (cold-start 공개 의무)

**The recommended option**(첫 옵션, `(권장)` 라벨)은 선호 메모리에서 관측된 **통계적 다수 합리적 기본값**이어야 한다. 시스템이 밀고 싶은 정책 기본값이 아니어야 한다.

**Where** 충분한 관측이 존재하지 않으면(cold-start, 관측 부족), 오케스트레이터는 기존 정적 기본값으로 폴백하고 옵션 description에 **"based on static default, N observations needed for personalization"** (또는 동등한 `conversation_language` 자연어 표현)을 공개해야 한다.

- 근거: 기본값 효과는 합리적 기본값에서 성립; 시스템 밀어넣기는 자율성 침식 위험. cold-start 공개는 미관측 추천 금지를 만족한다.

### 4. 전제조건 서술 — 추천 성립 조건 명시

**추천 옵션의 `description`**은 추천이 성립하는 전제조건을 서술해야 한다. 사용자가 전제 위반 시 추천을 즉시 거부할 수 있도록.

- 형식 권장: `"Recommended when <precondition>"` (en) 또는 동등한 `conversation_language` 표현 — 전제 위반 시 거부가 자명한(trivial) 형태.
- 근거: 투명성 + 쉬운 opt-out 번들. 전제가 서술되지 않은 추천은 기형적 설계이다.

### 5. 적응형 추천 강도 — 숙련도 기반 자동 분기

**Where** 오케스트레이터가 고숙련도(전문가)를 추정하면(세션 카운트 / 의사결정 일관성 / 명시적 자가 평가 중 ≥1), 오케스트레이터는 **약 추천 강도**(info-centric, 자율성 우선 — `(권장)` 라벨 override 없이 inferred preference를 공개만)를 적용한다.

**Where** 저숙련도(일반 사용자)로 추정되면, 오케스트레이터는 **강 추천 강도**(기본값-like — `(권장)` 라벨 + 투명한 이유)를 적용한다.

- cold-start 보호: 숙련도 추정이 불가능한 초기는 neutral 강도로 처리 (inferred preference 기반 `(권장)` 배치 없음).
- 근거: 전문가에게 강 추천은 info-centric 작업에서 자율성 침식; 일반 사용자에게 약 추천은 결정 피로 가중. 자동 분기가 양쪽을 모두 만족한다.

---

## Preview Field Standards

The `preview` field on each `AskUserQuestion` option renders a multi-line content block in a monospace box alongside the option list. When ANY option in a question has a `preview`, the Claude Code TUI auto-switches to side-by-side layout (vertical option list on the left, focused option's preview on the right).

This field complements `description` — it does NOT replace it. `description` carries the prose explanation that arrives with every option; `preview` carries the visual artifact (table / mockup / snippet) that benefits from side-by-side comparison.

### When to Use (SHOULD)

Apply `preview` when options carry **structural or quantitative differences** that benefit from visual side-by-side comparison:

- Sprint entry SPEC selection (Tier / Scope / Files / Risk comparison)
- Workflow branching decisions (cost / latency / risk trade-offs)
- Migration strategy selection (rollback path / performance / scope deltas)
- Architecture decision (component layout / dependency graph variants)
- Tier classification (Tier S minimal / Tier M standard / Tier L thorough envelope comparison)

### When NOT to Use

Omit `preview` when labels and descriptions already suffice:

- Simple yes/no confirmations
- PR merge approval
- Single-decision-point confirmations after the orchestrator has already laid out the structural context in prose
- Permission grants (e.g., "allow Bash?", "allow Write?")
- Continue / Abort prompts at a checkpoint gate

### Constraint: Single-Select Only

`preview` is rendered ONLY when `multiSelect: false`. The Claude Code TUI silently drops the `preview` field when `multiSelect: true`. Do not combine — if multi-select is required, fall back to richer `description` text instead.

### Constraint: Scroll Limitation (Issue #33062)

The Claude Code TUI preview pane is currently NOT scrollable. Content exceeding the visible window is truncated with an "N lines hidden" indicator, and arrow keys only navigate between options on the left (not within the preview pane). Mitigation guidelines (best-effort, not enforced):

- Keep preview content under ~12 visible lines
- Place the most decision-relevant information in the first 6 lines
- For longer artifacts (full SPEC body, large diff), condense to a metadata table in `preview` and surface the full content via a follow-up message after selection

Reference: `https://github.com/anthropics/claude-code/issues/33062`

### Format Freedom

`preview` content renders as markdown inside a monospace box. The author may use any visual format that fits the comparison:

- **Compact metadata table** (one `key: value` per line) — preferred for option-set comparison; allows visual scanning of deltas when the same key set appears across all options
- **ASCII art mockup** — UI layouts, architecture diagrams, component boundaries
- **Code snippet** (fenced or unfenced) — implementation variants, configuration examples
- **Mixed** — metadata table plus a small diagram, when both contribute to the decision

When options carry comparable metadata, prefer a consistent key set across all options' previews so the user can visually scan the deltas. When options are fundamentally different in shape (e.g., "implement now" vs "ASCII mockup of UI"), format freedom is acceptable even if it sacrifices direct comparability.

### Bias Prevention Inheritance

The bias prevention rule from §Option Description Standards applies equally to `preview` content:

- The recommendation signal is conveyed **exclusively** by the `(권장)` / `(Recommended)` label suffix on the first option
- Preview content MUST use neutral, factual language — no persuasive framing, no decorations privileging one option
- Do not visually inflate the recommended option's preview (no larger box, no extra emoji, no longer body)

### Worked Example

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "Sprint 8 entry SPEC를 선택해주세요.",
    header: "Sprint 8",
    multiSelect: false,
    options: [
      {
        label: "SPEC-V3R6-SPEC-ID-VALIDATION-001 (권장)",
        description: "manager-spec body에 SPEC ID regex pre-write self-check 추가. Sprint 7 TMC-001 plan-phase L51 도출 원천 해소.",
        preview: "Tier:    S (minimal)\nScope:   manager-spec.md body + regex pre-write check\nFiles:   1-2 edit\nRisk:    Low — agent body 수정, 동작 변경 없음\nL51 origin: Sprint 7 TMC-001 plan-phase 도출"
      },
      {
        label: "SPEC-V3R6-CATALOG-FRONTMATTER-AUDIT-001",
        description: "frontmatter schema audit + lint rule 확장. §24 namespace align 후속.",
        preview: "Tier:    M (standard)\nScope:   internal/spec/lint.go + catalog.yaml\nFiles:   3-5 edit\nRisk:    Med — lint rule 확장은 cascade 가능\nOrigin:  §24 namespace align 후속"
      },
      {
        label: "SPEC-V3R6-CLI-INTEGRATION-001",
        description: "CLI subcommand integration test 추가. moai cli regression 방지.",
        preview: "Tier:    M (standard)\nScope:   cmd/moai + internal/cli integration tests\nFiles:   5-8 edit\nRisk:    Med — sandbox env 의존성 추가 가능\nOrigin:  CI 회귀 방지 SHOULD-FIX"
      }
    ]
  }]
})
```

Note how each option's `preview` uses the same key set (`Tier`/`Scope`/`Files`/`Risk`/`Origin`-or-`L51 origin`), allowing the user to scan deltas vertically when navigating the option list.

### Cross-references

- Claude Code SDK documentation: `toolConfig.askUserQuestion.previewFormat` (`"markdown"` | `"html"`). The Claude Code native TUI auto-renders the `preview` field without explicit `previewFormat` config.
- Constraint origin: GitHub issue `anthropics/claude-code#33062` (preview pane scroll limitation).
- Related rule: §Option Description Standards (description is always required; preview is additive).

---

## Orchestrator–Subagent Boundary

The `AskUserQuestion` interaction channel is **asymmetric** by design.

### Orchestrator Obligations

The MoAI orchestrator (main session) MUST:
- Use `AskUserQuestion` as the exclusive channel for all user-facing questions
- Preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")` before each call
- Collect all necessary user preferences **before** delegating to subagents
- On receiving a blocker report from a subagent: run an `AskUserQuestion` round with the user, inject the user's responses into a fresh subagent prompt, and re-delegate

See `.claude/rules/moai/core/askuser-protocol.md` (this file) for the complete preload sequence.

### Subagent Prohibitions

Subagents invoked via `Agent()` operate in isolated, stateless contexts and CANNOT interact with users directly:
- [ZONE:Frozen] [HARD] Subagents MUST NOT invoke `AskUserQuestion`
- [ZONE:Frozen] [HARD] Subagents MUST NOT output free-form prose questions directed at the user
- [ZONE:Frozen] [HARD] Subagents MUST NOT embed AskUserQuestion call syntax in their response body

### Blocker Report Format

When a subagent requires user input that was not provided in the spawn prompt, it MUST return a structured blocker report instead of attempting to interact with the user:

```markdown
## Missing Inputs

The following parameters are required but were not provided:

| Parameter | Type | Expected Values | Rationale |
|-----------|------|-----------------|-----------|
| [name]    | [type] | [values]      | [why needed] |

**Blocker**: Cannot proceed without the above inputs. Please re-delegate with these values injected into the prompt.
```

### Re-delegation Procedure

```
[Orchestrator receives blocker report]
Step 1: ToolSearch(query: "select:AskUserQuestion")
Step 2: AskUserQuestion — ask user for the missing inputs
Step 3: Construct fresh subagent prompt with user's answers injected
Step 4: Re-delegate to subagent
```

---

## Ambiguity Triggers and Exceptions

This section is the **single source of truth** for Stage 1 Clarify trigger conditions. Both `CLAUDE.md §7 Rule 5` and `CLAUDE.md §8 Ambiguity Triggers` cross-reference this definition.

### The Four Triggers (any one activates Stage 1)

1. **Pronoun or demonstrative without clear referent**: "this", "that", "it", "the previous one" — the referent cannot be unambiguously determined from context
2. **Multi-interpretable action verb without specified scope**: "clean up", "process", "improve", "fix" — the action could apply to multiple different implementations
3. **Unclear boundaries**: How far to go, how much to change, which files are in scope, where to stop
4. **Potential conflict with existing state**: Uncommitted changes, in-progress branches, overlapping work that the request might conflict with

### The Five Exceptions (Stage 1 is skipped)

1. Single-line typo or formatting fix — scope is self-evident
2. Bug fix with explicit reproduction provided — the reproducer defines scope
3. Direct file read when the path is explicitly specified — no interpretation needed
4. Command invocation with all required arguments provided — no ambiguity
5. Continuation of previously confirmed work in the same session — intent already established

### First-Action Sequence After Trigger

```
Trigger detected
  → Step 1: ToolSearch(query: "select:AskUserQuestion")   [deferred tool preload]
  → Step 2: Compose AskUserQuestion round (≤4 Q, ≤4 options, (권장) first, conversation_language)
  → Step 3: Send AskUserQuestion, collect responses
  → Step 4: Assess intent clarity (100% required)
  → Step 5: If <100%: go to Step 1 with narrowed questions
             If 100%: consolidate report → final confirmation → execute
```

---

## Free-form Circumvention Prohibition

Free-form interrogative prose in the response body MUST NOT be used as a substitute for `AskUserQuestion` — always use AskUserQuestion.

### Why this matters

`AskUserQuestion` automatically appends an **"Other"** option to every question set. This means:
- Users who prefer free-form answers can select "Other" and type their response
- The orchestrator does NOT need to produce free-form questions to support free-form answers
- Structured options via `AskUserQuestion` are faster and less error-prone than prose for most users

### The "Other" Mechanism

When the orchestrator constructs an `AskUserQuestion` round that does not exhaustively cover all possibilities, the "Other" option is automatically available. This covers:
- Edge cases not anticipated in the option list
- User preferences that do not fit the provided options
- Free-form elaboration on a structured choice

### Prohibited Patterns

```
# Prohibited — free-form question in prose
"Which direction would you like to proceed?"

# Prohibited — markdown list as options in prose
- **A**: Run SPEC immediately
- **B**: Review first
- **C**: Abort

# Prohibited — inline question at end of response paragraph
"I've completed the changes. Should I create a PR now?"
```

### Correct Pattern

```
# Correct — always through AskUserQuestion
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "다음 단계를 선택하세요.",
    header: "진행 방향",
    options: [
      { label: "PR 즉시 생성 (권장)", description: "현재 변경사항으로 PR을 생성합니다. CI가 자동 실행됩니다." },
      { label: "검토 후 PR", description: "변경사항을 먼저 검토하고 PR을 생성합니다." },
      { label: "중단", description: "현재 작업을 중단하고 상태를 보존합니다." }
    ]
  }]
})
```

## Non-ASCII Tool-Call Encoding

The `AskUserQuestion` payload — `question`, `header`, and every option `label` / `description` / `preview` — routinely carries text in the user's `conversation_language`. For Korean, Japanese, Chinese, and other multi-byte scripts, this text MUST be written as **native UTF-8 directly** in the tool-call JSON. Hand-authored `\uXXXX` escape sequences are **PROHIBITED**.

### Failure Mode

A malformed escape — a stray space inside the sequence, a truncated code point, or a half-written `\u` — corrupts the JSON so the `questions` array is parsed as a bare string instead of a list of objects. The call is rejected with `Invalid tool parameters` / `InputValidationError`, and the orchestrator's clarification round silently fails on its first attempt.

### Root-Cause Mechanism

The corruption is not random; it follows a three-step chain documented across LLM tool-call runtimes:

1. **Serialization escaping.** A serialization layer emitting JSON with `ensure_ascii`-style escaping converts multi-byte characters into `\uXXXX` sequences (native CJK text becomes a run of `\uXXXX` code points) when a prior tool call or result is recorded into the conversation history.
2. **Prompt pollution.** That escaped form is fed back into the next inference turn, so the model sees literal `\uXXXX` sequences in its own context instead of native characters.
3. **Mimicry failure.** The model imitates the escape format for its next tool call but cannot reliably reproduce the exact code points, emitting plausible-looking but corrupted escapes (the stray-space / truncated forms above).

The corrective lever is step 1: keep multi-byte text as native UTF-8 in every tool call so the context is never seeded with `\uXXXX` runs.

### Directive and Recovery

- **Preventive (always):** write all `conversation_language` text as native UTF-8 in the tool-call JSON. Never hand-escape a non-ASCII character.
- **Recovery (on failure):** if a call is rejected with `Invalid tool parameters` and the payload contained non-ASCII text, re-issue the identical call with the text rewritten as native UTF-8 — do not try to "repair" the escape sequence.

### Scope Note

This is a model-output discipline, not a project-code defect: a correct JSON serializer (for example Go's `encoding/json`) already preserves multi-byte UTF-8 and never emits `ensure_ascii`-style escapes, so it cannot be the pollution source. The discipline binds the orchestrator's own tool-call construction.

---

Version: 1.1.0
Classification: Canonical Reference — do not duplicate content; cross-reference this file instead.
