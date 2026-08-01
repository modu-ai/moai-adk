---
description: "Detailed reference for AskUserQuestion recommendation-placement principles and the preview field"
paths: "**/askuser-protocol.md"
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
