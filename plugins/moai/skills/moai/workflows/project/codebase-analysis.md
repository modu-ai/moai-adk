---
description: "Project Phase 3/1.5/2 — Codebase analysis via Explore subagent, two-stage deep interview for existing projects (clarity-scored Stage A + mandatory extended-axes Stage B), and user confirmation"
user-invocable: false
metadata:
  parent: moai-workflow-project
  phase: "Phase 3/1.5/2: Codebase Analysis and User Confirmation"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Phase 3: Codebase Analysis (Existing Projects Only)

[HARD] Delegate codebase analysis to the Explore subagent.

[SOFT] Use the `ultrathink` keyword for comprehensive analysis (activates Adaptive Thinking on Opus 4.7+).

Analysis Objectives passed to Explore agent:

- Project Structure: Main directories, entry points, architectural patterns
- Technology Stack: Languages, frameworks, key dependencies
- Core Features: Main functionality and business logic locations
- Build System: Build tools, package managers, scripts

Expected Output from Explore agent:

- Primary Language detected
- Framework identified
- Architecture Pattern (MVC, Clean Architecture, Microservices, etc.)
- Key Directories mapped (source, tests, config, docs)
- Dependencies cataloged with purposes
- Entry Points identified

Execution Modes:

- Fresh Documentation: When .moai/project/ is empty, generate all three files
- Update Documentation: When docs exist, read existing, analyze for changes, ask user which files to regenerate

---

## Phase 4: Deep Interview (Existing Projects Only)

Purpose: After codebase analysis, gather user intent and context that cannot be inferred from the code alone. Questions are informed by the analysis results from Phase 1.

[HARD] All questions MUST use AskUserQuestion in user's conversation_language.
[HARD] During the interview, the agent MUST NOT generate documentation or write files. The sole output is `.moai/project/interview.md`.

### Two-Stage Interview Structure

The interview is a **two-stage** procedure. The two stages are governed by DIFFERENT
rules and must never be conflated — conflating them is what leaves the extended-axes
round uncollected:

| Stage | Rounds | Governed by | Terminates when |
|-------|--------|-------------|-----------------|
| **Stage A** — clarity-scored adaptive discovery | Round 1 … `project.max_rounds` (3) | `clarity_threshold` (4, entry floor), sufficiency exit (≥ 8), abandon floor (≤ 3), `project.max_rounds` cap | early exit, abandon, OR the cap is reached |
| **Stage B** — mandatory extended-axes round | Round 4 (exactly one) | nothing — Stage B is EXEMPT from clarity scoring, from the Stage A early-exit skip, from the abandon path, and from `project.max_rounds` | after its four axes are collected |

**Stage B always runs**, on every Stage A exit route, before documentation generation.
`project.max_rounds` caps **Stage A only** — it is NOT a cap on the interview as a whole.

### Stage A: Adaptive Clarity-Scored Discovery (mirrors `plan/clarity-interview.md`)

Stage A is clarity-driven, NOT fixed-length. It scores accumulated answer clarity on a
0-10 scale and adapts the round count, reusing the adaptive mechanism defined in
`.claude/skills/moai/workflows/plan/clarity-interview.md` (the SAME 0-10 scale
semantics — do NOT invent a divergent rubric):

- **Entry floor**: `clarity_threshold` (4, from `.moai/config/sections/interview.yaml`)
  is the interview ENTRY floor — the clarity band at/above which the interview runs.
  It is NOT the early-exit target.
- **Additional rounds**: while the accumulated clarity score is below the sufficiency
  target (≥ 8) and above the abandon floor (≤ 3), run one or more additional Stage A
  rounds, up to `project.max_rounds` (3, from `interview.yaml`).
- **Early exit (sufficiency)**: the Stage A early exit fires ONLY when BOTH of the
  following hold — (i) the accumulated clarity score reaches the sufficiency target
  (≥ 8), AND (ii) all four **required base fields** (`domain`, `goal`, `constraints`,
  `scope`) have been answered — where a field auto-populated from the Phase 3 codebase
  analysis counts as answered. When both hold before `max_rounds` rounds have run,
  terminate **Stage A** early (skip the remaining **Stage A** rounds) and proceed to
  the mandatory **Stage B** round. The early exit terminates Stage A ONLY — it NEVER
  skips Stage B.
- **Required-field gate**: while any required base field is still unanswered (neither
  elicited nor confidently auto-populated), Stage A continues (up to
  `project.max_rounds`) regardless of the accumulated clarity score. A high clarity
  score alone never satisfies the early exit.
- **Volunteered-field credit (this is what makes the early exit reachable)**: a required
  base field VOLUNTEERED in an EARLIER Stage A round — e.g. the user names their hard
  constraints and their scope boundary while answering the Round 1 ownership/purpose
  question — COUNTS as answered and MUST NOT be re-asked in its designated round (the
  same credit the Phase 3 auto-population already carries). Without this credit the four
  required fields could only ever complete at the LAST Stage A round (which IS
  `max_rounds`), so the early exit's "before `max_rounds` rounds have run" condition
  could never hold, and Stage A would silently degrade back into a fixed-length
  interview pinned at `max_rounds` — the very defect the adaptive loop exists to remove.
- **Abandon**: if the accumulated clarity score drops to ≤ 3 (the answers add no
  useful information), end **Stage A** early and proceed to **Stage B** with the
  best-available answers.
- **Cap is a hard stop**: on reaching `project.max_rounds` with a required base field
  still unanswered, record that field as absent and proceed to **Stage B**. Do NOT
  loop Stage A.
- Re-evaluate the clarity score after each Stage A round and display the round counter:
  "Interview round {N}/{max_rounds}".

**Stage A base-field coverage mapping** — Stage A's rounds, between them, elicit all
four **required base fields** (`domain`, `goal`, `constraints`, `scope`). For this
existing-project host, a field confidently inferred from the Phase 3 codebase analysis
is AUTO-POPULATED: it counts as answered, is NOT re-asked, and is marked as
auto-populated below:

| Required base field | Elicited by | Auto-populated? |
|---|---|---|
| `domain` | AUTO-POPULATED from the Phase 3 codebase analysis (detected stack, architecture, entry points); asked in Stage A Round 1 ONLY when it cannot be confidently inferred | yes (when inferable) |
| `goal` | Stage A Round 1 (Ownership, Purpose, and Goal) | no — always elicited |
| `constraints` | Stage A Round 2 (Constraints and Non-Goals) | no — always elicited |
| `scope` | Stage A Round 3 (Scope, Boundaries, and Documentation Priority) | no — always elicited |

**Round 1: Ownership, Purpose, and Goal** — elicits `goal` (confirms auto-populated `domain`)

Topic: Who maintains this project, what problem domain does it sit in, and what is the primary goal going forward?

Present via AskUserQuestion with exactly 4 options based on Phase 3 detected project type:
- Option 1 (Recommended): Active product being developed further: This codebase is actively developed and the documentation should reflect its current trajectory and roadmap.
- Option 2: Legacy system being maintained: The codebase is stable and the documentation should reflect its current state for maintenance and onboarding.
- Option 3: System being refactored or migrated: Major structural changes are planned and documentation should reflect the target state.
- Option 4: Type your own answer: Enter a custom response to describe the ownership context.

Capture the one-line project goal / success condition, recorded as the `goal` field. The `domain` field is auto-populated from the Phase 3 analysis (a short slug such as cli-tooling or web-api); ask for it here ONLY when the analysis could not confidently infer it.

**Round 2: Constraints and Non-Goals** — elicits `constraints`

Topic: What are the known constraints, technical debts, or things this project intentionally does NOT do?

Present via AskUserQuestion with exactly 4 options informed by Phase 3 analysis findings:
- Option 1 (Recommended): No known critical constraints: Document the codebase as-is without constraint annotations.
- Option 2: Performance or scalability constraints exist: There are known bottlenecks or scaling limits that should be documented.
- Option 3: Security or compliance constraints exist: Specific security requirements or compliance rules affect the architecture.
- Option 4: Type your own answer: Describe the specific constraints or non-goals for this project.

Capture the hard constraints (performance, security, compatibility, technology limits, or "none known"), recorded as the `constraints` field.

**Round 3: Scope, Boundaries, and Documentation Priority** — elicits `scope`

Topic: What is in scope versus explicitly out of scope for this project, and which aspect must the documentation capture most accurately?

Present via AskUserQuestion with exactly 4 options:
- Option 1 (Recommended): Architecture and module boundaries: Prioritize documenting how the system is structured and how modules interact.
- Option 2: Technology stack and dependencies: Prioritize the frameworks, libraries, and their versions for onboarding.
- Option 3: Core business logic and data flow: Prioritize documenting what the system does and how data moves through it.
- Option 4: Type your own answer: Specify what is in scope, what is explicitly out of scope, and what should be documented with highest fidelity.

[HARD] The 4 options above resolve the DOCUMENTATION-PRIORITY answer only — they do NOT by themselves yield the scope boundary (only the free-form Option 4 would, and it is not the recommended option). **Then elicit the in-scope / out-of-scope boundary as a SEPARATE AskUserQuestion** (`project.questions_per_round: 3` allows a second question in this round): ask what is explicitly IN scope and what is explicitly OUT of scope for this project. Record that boundary summary as the `scope` field — `scope` MUST be elicited, and MUST NOT be inferred from the documentation-priority option. Skipping this second question leaves `scope` ABSENT, which propagates an empty `scope` into `harness-spec.yaml` and forces `harness-build-entry.md` Phase 3 to re-ask it — reopening the exact information leak this workflow exists to close. The documentation-priority answer is retained as additional context for Phase 3.

### Stage B: Mandatory Extended-Axes Round

**Round 4: Verification, Surfaces, and Sharing (extended axes)**

[HARD] **Stage B always runs.** It executes unconditionally after Stage A terminates —
by early exit, by abandon (clarity ≤ 3), OR by reaching the `project.max_rounds` cap —
and before documentation generation. Stage B is:

- **EXEMPT from `project.max_rounds`** — Round 4 is not counted against the cap.
- **EXEMPT from the Stage A early-exit skip** — the early exit skips the remaining
  Stage A rounds only, never this round.
- **EXEMPT from the Stage A abandon path** — an abandoned Stage A still proceeds here.
- **EXEMPT from clarity scoring** — Round 4 runs regardless of clarity score.

Why: the four extended axes are clarity-independent **factual collection**, not
ambiguity resolution. There is nothing to "score" about whether the project has a UI
or which test command it runs — so subjecting them to the Stage A clarity loop, or to
its round cap, would be a category error and would leave them uncollected.

Topic: How is the project verified, what does it surface, what does it integrate with, and who runs it? Elicit these four axes (later recorded into `harness-spec.yaml` — see `doc-generation.md`). For existing projects, PRE-FILL each axis from the Phase 3 codebase analysis where it can be inferred (e.g., detected test command → `verification`; a detected web framework → `has-ui`; detected DB/API dependencies → `external_systems`); an axis confidently inferred from analysis counts as answered and is NOT re-asked. Collect ALL remaining (un-inferred or ambiguous) axes in a single AskUserQuestion call carrying up to 4 questions — one question per remaining axis, each question with up to 4 options — replacing the former per-axis separate calls. Merging changes the number of blocking round-trips only: each axis keeps its own question, options, and description quality unchanged:

- **Verification method** — the test / e2e command or verification method (e.g., `go test ./...`, `pytest`, an e2e suite, or "manual verification"). Recorded as the `verification` field.
- **UI surface** — whether the project has a user-facing UI or is headless: `has-ui` (web / desktop / mobile front-end) vs `headless` (CLI / API / library / service). Recorded as the `ui_surface` field.
- **External systems** — the databases, APIs, or services the project integrates with (e.g., PostgreSQL, Redis, a payment API, an external microservice), or "none". Recorded as the `external_systems` field.
- **Team-sharing intent** — whether the project is `solo` (single maintainer) or `team-shared` (multiple contributors). Recorded as the `team_sharing` field.

Stage B still RUNS even when every axis is auto-populated — it confirms the inferred values rather than skipping. An axis the user declines or cannot answer is recorded as an explicit empty value; the round having RUN is what makes that a legitimate empty rather than an uncollected one.

**Output:** Write all answers to `.moai/project/interview.md` with this structure:

```
# Project Interview

## Stage A Round 1: Ownership, Purpose, and Goal
Question: {question asked}
Answer: {user's answer}
Domain: {domain slug — mark "(auto-populated from codebase analysis)" when inferred}
Goal: {one-line goal}

## Stage A Round 2: Constraints and Non-Goals
Question: {question asked}
Answer: {user's answer}
Constraints: {list, or none known}

## Stage A Round 3: Scope, Boundaries, and Documentation Priority
Question: {question asked}
Answer: {user's answer}
Scope: {in-scope / out-of-scope summary}

## Stage B Round 4: Verification, Surfaces, and Sharing
Verification: {verification method / test command}
UI surface: {has-ui | headless}
External systems: {list, or none}
Team sharing: {solo | team-shared}
```

A Stage A round that did not run (because Stage A exited early or abandoned) records its base field as absent rather than omitting the section.

Pass `interview.md` to Phase 5 (User Confirmation) and Phase 6 (Documentation Generation) as additional context. Documentation agents MUST read interview.md before generating files.

---

## Phase 5: User Confirmation

Present analysis summary via AskUserQuestion.

Display in user's conversation_language:

- Detected Language
- Framework
- Architecture
- Key Features list

Options:

- Proceed with documentation generation (Recommended): MoAI will generate product.md, structure.md, and tech.md based on the analysis above. You can review and edit the documents afterwards.
- Review specific analysis details first: See a detailed breakdown of each detected component before generating documents. Useful if you want to correct any misdetected frameworks or features.
- Cancel and adjust project configuration: Stop the process and make changes to your project setup. Choose this if the analysis looks significantly incorrect.

If "Review details": Provide detailed breakdown, allow corrections.
If "Proceed": Continue to Phase 6 (see `doc-generation.md`).
If "Cancel": Exit with guidance.
