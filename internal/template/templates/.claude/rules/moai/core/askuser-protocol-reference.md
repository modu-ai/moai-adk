---
description: "Detailed reference for AskUserQuestion recommendation-placement principles and the preview field"
paths: "**/askuser-protocol.md,**/.claude/skills/moai/workflows/*.md,**/.claude/output-styles/moai/*.md"
---

# AskUserQuestion Protocol — Reference Detail

> Detail companion to `askuser-protocol.md` (the SSOT). That file carries the binding
> rules; this file carries the reasoning, the worked example, and the constraint
> catalogue. Loaded only when `askuser-protocol.md` itself is being edited — read it
> directly when composing a question that needs these details.

## Recommendation Placement Principles

> This section defines the policy SSOT for recommendation placement (emission timing / question ordering / recommended-option rationale / precondition statement / adaptive strength).

The AskUserQuestion `(Recommended)` label (locale token `(권장)` in Korean) MUST be grounded in the **statistically-majority rational default the user has selected** (observed in preference memory), NOT merely a policy default the system wants to push. This section defines the five principles of recommendation placement.

### 1. Emission timing — information-gain alignment (Fisher information I=p(1−p))

**Where** the orchestrator estimates uncertainty p for an upcoming decision, **When** p ≈ 0.5 (Fisher information I=p(1−p) is maximal — the decision boundary), the orchestrator MUST emit that question via AskUserQuestion. **While** p is close to 0 or 1 (nearly certain), the orchestrator auto-resolves to the statistical-majority option and omits the question.

- p estimation (initial heuristic): the observed majority-selection ratio in the same domain. Cold-start (observations < N) is treated as p ≈ 0.5 to trigger emission.
- Rationale: the just-in-time decision-boundary question principle (Murphy "Probabilistic Machine Learning" Ch.3 — Fisher information I=p(1−p) is maximal at p=0.5).

### 2. Question ordering — descending information gain

**Where** multiple questions are placed in a single AskUserQuestion call, the orchestrator orders them by estimated information gain in descending order (the highest-information-gain question first).

- Rationale: placing higher-information-gain questions first lets the user complete the core decisions before encountering lower-value questions.

### 3. Recommended option — statistical-majority rational default (cold-start disclosure obligation)

**The recommended option** (the first option, carrying the `(Recommended)` / `(권장)` label) MUST be the **statistical-majority rational default** observed in preference memory. It MUST NOT be a policy default the system wants to push.

**Where** sufficient observations do not exist (cold-start, observations < N), the orchestrator MUST fall back to the existing static default and disclose in the option description **"based on static default, N observations needed for personalization"** (or the equivalent natural-language expression in `conversation_language`).

- Rationale: the default effect (d≈0.55) holds for rational defaults; system-pushing risks autonomy erosion. Cold-start disclosure satisfies the no-unobserved-recommendation rule (`verification-claim-integrity.md §1.1 surface 3`).

### 4. Precondition statement — make the recommendation's holding conditions explicit

**The recommended option's `description`** MUST state the preconditions under which the recommendation holds, so the user can immediately reject it when a precondition is violated.

- Recommended format: `"Recommended when <precondition>"` (en) or the equivalent `conversation_language` expression — a form where rejection on precondition violation is trivial.
- Rationale: transparency + easy opt-out bundling. A recommendation whose preconditions are unstated is a malformed design.

### 5. Adaptive recommendation strength — proficiency-based automatic branching

**Where** the orchestrator estimates high proficiency (expert) — session count ≥ threshold, OR decision consistency, OR explicit self-assessment (any one of the three) — the orchestrator applies **weak recommendation strength** (info-centric, autonomy-first — discloses the inferred preference WITHOUT overriding via the `(Recommended)` label).

**Where** low proficiency (general user) is estimated, the orchestrator applies **strong recommendation strength** (default-like — `(Recommended)` label + transparent rationale).

- Cold-start protection: when proficiency estimation is impossible (early, session count < threshold), apply neutral strength (no `(Recommended)` placement based on inferred preference).
- Rationale: strong recommendation to an expert erodes autonomy in info-centric work; weak recommendation to a general user adds decision fatigue. Automatic branching satisfies both.
- Proficiency-estimation detail: design.md §A.4.

### Cross-reference

- Information-gain rationale for emission timing / question ordering: design.md §B.2 (documenting both sides of conflicting evidence).
- Autonomy buffer of the statistical-majority recommendation: this section §3 + §5 (adaptive strength) + recovery-control toggle (requirements-owned, out of this section's scope).
- Precondition statement and transparency: `verification-claim-integrity.md §1.1 surface 3` (no unobserved-inference claim).

> The recommendation placement principles above are evidence-based.

---

## Preview Field Standards

The `preview` field on each `AskUserQuestion` option renders a multi-line content block in a monospace box alongside the option list. When ANY option in a question has a `preview`, the Claude Code TUI auto-switches to side-by-side layout (vertical option list on the left, focused option's preview on the right).

This field complements `description` — it does NOT replace it. `description` carries the prose explanation that arrives with every option; `preview` carries the visual artifact (table / mockup / snippet) that benefits from side-by-side comparison.

### When to Use (SHOULD)

Apply `preview` when options carry **structural or quantitative differences** that benefit from visual side-by-side comparison:

- Epic entry SPEC selection (Tier / Scope / Files / Risk comparison)
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
    question: "Select the Epic 8 entry SPEC.",
    header: "Epic 8",
    multiSelect: false,
    options: [
      {
        label: "SPEC-V3R6-SPEC-ID-VALIDATION-001 (Recommended)",
        description: "Add a SPEC ID regex pre-write self-check to the manager-spec body.",
        preview: "Tier:    S (minimal)\nScope:   manager-spec.md body + regex pre-write check\nFiles:   1-2 edit\nRisk:    Low — agent body edit, no behavior change\n"
      },
      {
        label: "SPEC-V3R6-CATALOG-FRONTMATTER-AUDIT-001",
        description: "Frontmatter schema audit + lint rule extension.",
        preview: "Tier:    M (standard)\nScope:   internal/spec/lint.go + catalog.yaml\nFiles:   3-5 edit\nRisk:    Med — lint rule extension can cascade\nOrigin:  frontmatter schema audit follow-up"
      },
      {
        label: "SPEC-V3R6-CLI-INTEGRATION-001",
        description: "Add CLI subcommand integration tests. Prevents moai CLI regressions.",
        preview: "Tier:    M (standard)\nScope:   cmd/moai + internal/cli integration tests\nFiles:   5-8 edit\nRisk:    Med — may add sandbox env dependency\nOrigin:  CI regression prevention SHOULD-FIX"
      }
    ]
  }]
})
```

Note how each option's `preview` uses the same key set (`Tier`/`Scope`/`Files`/`Risk`/`Origin`), allowing the user to scan deltas vertically when navigating the option list.

### Cross-references

- Claude Code SDK documentation: `toolConfig.askUserQuestion.previewFormat` (`"markdown"` | `"html"`). The Claude Code native TUI auto-renders the `preview` field without explicit `previewFormat` config.
- Constraint origin: GitHub issue `anthropics/claude-code#33062` (preview pane scroll limitation).
- Related rule: §Option Description Standards (description is always required; preview is additive).

---

## Non-ASCII Tool-Call Encoding detail

> Relocated verbatim from `askuser-protocol.md` § Non-ASCII Tool-Call Encoding to keep the always-loaded file within its size budget. The directive, the failure mode, the recovery procedure, and the 3-item pre-emit self-check remain inline there.

### Root-Cause Mechanism

The corruption is not random; it follows a three-step chain documented across LLM tool-call runtimes:

1. **Serialization escaping.** A serialization layer emitting JSON with `ensure_ascii`-style escaping converts multi-byte characters into `\uXXXX` sequences (native CJK text becomes a run of `\uXXXX` code points) when a prior tool call or result is recorded into the conversation history.
2. **Prompt pollution.** That escaped form is fed back into the next inference turn, so the model sees literal `\uXXXX` sequences in its own context instead of native characters.
3. **Mimicry failure.** The model imitates the escape format for its next tool call but cannot reliably reproduce the exact code points, emitting plausible-looking but corrupted escapes (the stray-space / truncated forms above).

The corrective lever is step 1: keep multi-byte text as native UTF-8 in every tool call so the context is never seeded with `\uXXXX` runs.

### Self-Reinforcing Pollution Loop (why one failure recurs)

This failure is **not** an isolated one-off — it is self-reinforcing, and that is why it "keeps happening" rather than failing once and clearing. The Root-Cause Mechanism above is a loop, not a line: once a single `\uXXXX` run is seeded into the conversation context (step 2, prompt pollution), the model sees escaped text in its own context and mimics that format on the *next* tool call too (step 3), re-seeding fresh corruption. Left unbroken, one malformed call becomes a run of malformed calls.

Breaking the loop requires more than retrying the one rejected call:

- **Do not carry the corrupted form forward.** After a recovery, the very next tool call carrying non-ASCII text is the highest-risk moment — the polluted context is still in view. Re-author that payload as native UTF-8 from the intended source text (the user's actual words), NOT by transcribing the `\uXXXX` sequence you can see in context.
- **Recovery is per-payload, not per-call-type.** The clean-up applies to Bash, Write / Edit, and every subsequent multi-byte tool call in the turn — not only the `AskUserQuestion` that first failed.
- **Persistent recurrence → reset the context.** If native-UTF-8 re-authoring still yields repeated `InputValidationError` on non-ASCII payloads within the same session, the context is saturated with `\uXXXX` runs. Escalate to a `/clear` (per `context-window-management.md` § Context Window Targets) with a paste-ready resume message, so the next session starts from an un-polluted context. This is the last-resort loop-break, not the first response.

### Scope Note

This is a model-output discipline, not a project-code defect: a correct JSON serializer (for example Go's `encoding/json`) already preserves multi-byte UTF-8 and never emits `ensure_ascii`-style escapes, so it cannot be the pollution source. The discipline binds the orchestrator's own construction of every tool call — `AskUserQuestion`, Bash, Write / Edit, and any other tool whose JSON payload carries non-ASCII text — not just clarification rounds. The `AskUserQuestion` case is the origin example; a corrupted `\uXXXX` escape in a Bash command or a Write payload fails the same way.

## Blind Spot Pass

The **Blind Spot Pass** is an OPTIONAL pre-plan Discovery technique for surfacing the user's **unknown-unknowns**: read-only reconnaissance by `Agent(Explore)`, with findings surfaced to the user through the orchestrator's `AskUserQuestion` channel.

- **When**: the user is working in an **unfamiliar** domain (new subsystem, unfamiliar design/library territory) AND the orchestrator suspects unknown-unknowns — SHOULD run **before plan-phase entry**, before authoring the SPEC. The trigger is a judgment call, NOT an automatic gate; in a familiar domain with no suspected unknown-unknowns, the pass is skipped with no forced overhead.
- **Mechanism**: (1) spawn `Agent(Explore)` in **read-only** mode to scan the relevant domain (subsystem, library surface, integration points); (2) surface the likely unknown-unknowns through a single `AskUserQuestion` round so the user can react before the plan is authored.
- **Subagent boundary (preserved)**: `Agent(Explore)` — and any subagent — **does not prompt the user** directly; findings surface only through the orchestrator's channel. A subagent that lacks input returns a blocker report; it never asks the user.

---

## Relocated bodies — always-loaded stub sections

> Relocated from `askuser-protocol.md` to keep the always-loaded file within its size budget. Every binding clause stays inline there; these are the non-binding bodies it points at.

### General Rule for Deferred Tools

Any deferred tool requires a `ToolSearch` select preload before invocation. The pattern generalizes: `ToolSearch(query: "select:<tool>[,<tool>...]")` — single (`select:AskUserQuestion`) or multiple (`select:AskUserQuestion,TaskCreate`).

**Preload sequence** (per turn — if a new turn begins and `AskUserQuestion` will be called again, preload again; never reverse or omit Step 1):

```
[Turn N]
Step 1: ToolSearch(query: "select:AskUserQuestion")   ← preload deferred schema
Step 2: AskUserQuestion({ questions: [...] })           ← now valid to invoke
```

---

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

### The Unknowns 4-Quadrant Lens

Classify the ambiguity by **user blind spot** (Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns):

- **Known-Knowns** — stated + confirmed facts. No clarification needed
- **Known-Unknowns** — gaps the user is aware of. Resolve via a Socratic interview round (§ Socratic Interview Structure)
- **Unknown-Knowns** — constraints implicit in the codebase the user has not surfaced. Resolve via `Agent(Explore)` read-only reconnaissance, then confirm with the user
- **Unknown-Unknowns** — risks neither side has articulated. When suspected (unfamiliar domain/subsystem/design territory), run a Blind Spot Pass (§ Blind Spot Pass) before plan-phase entry

### First-Action Sequence After Trigger

```
Trigger detected
  → Step 1: ToolSearch(query: "select:AskUserQuestion")   [deferred tool preload]
  → Step 2: Compose AskUserQuestion round (≤4 Q, ≤4 options, (권장) first under recommendation_mode: push — withheld under pull, conversation_language)
  → Step 3: Send AskUserQuestion, collect responses
  → Step 4: Assess intent clarity (100% required)
  → Step 5: If <100%: go to Step 1 with narrowed questions
             If 100%: consolidate report → final confirmation → execute
```

---

### Directive and Recovery

- **Preventive (always):** write all `conversation_language` text as native UTF-8 in the tool-call JSON — this binds **every** tool call carrying multi-byte text, not only `AskUserQuestion` but Bash commands, Write / Edit content arguments, and any other tool-call payload. Never hand-escape a non-ASCII character.
- **Recovery (on failure):** if a call is rejected with `Invalid tool parameters` and the payload contained non-ASCII text, re-issue the identical call with the text rewritten as native UTF-8 — do not try to "repair" the escape sequence. Do not carry the corrupted form forward; re-author the next non-ASCII payload from the intended source text, not by transcribing the `\uXXXX` run visible in context. Persistent recurrence within a session → escalate to `/clear` with a paste-ready resume (last-resort loop-break).

### Pre-Emit Self-Check (before any tool call carrying non-ASCII text) — 3 items

- [ ] Is every `conversation_language` string in this payload written as native UTF-8 characters (한글 / 日本語 / 中文), with **zero** hand-authored `\uXXXX` sequences?
- [ ] Am I authoring this text from the intended source meaning, not transcribing an escaped `\uXXXX` run visible in my own context?
- [ ] If a prior call in this turn already failed with `Invalid tool parameters` on non-ASCII text, have I re-authored — not repaired — this payload, and am I watching for a saturated context that warrants `/clear`?

### Pre-emit self-check (report-before-ask) — 5 items

- [ ] Does the user's latest message request a report / analysis / explanation rather than a decision? If yes, this turn ends with the report — defer this AskUserQuestion to a later turn.
- [ ] Do this question's options derive from investigation results? If yes, does a substantive report precede this call in the same turn?
- [ ] Is every codename / identifier appearing in the options explained in the preceding report?
- [ ] Do the findings live in the response body (not only inside option previews)?
- [ ] If a report was promised earlier in the task, has it actually been rendered?

---

Version: 1.3.0
Classification: Canonical Reference — do not duplicate content; cross-reference this file instead.
