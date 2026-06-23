---
description: "Project Phase 1/1.5/2 — Codebase analysis via Explore subagent, 3-round deep interview for existing projects, and user confirmation"
user-invocable: false
metadata:
  parent: moai-workflow-project
  phase: "Phase 1/1.5/2: Codebase Analysis and User Confirmation"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Phase 1: Codebase Analysis (Existing Projects Only)

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

## Phase 1.5: Deep Interview (Existing Projects Only)

Purpose: After codebase analysis, gather user intent and context that cannot be inferred from the code alone. Questions are informed by the analysis results from Phase 1.

[HARD] All questions MUST use AskUserQuestion in user's conversation_language.
[HARD] During the interview, the agent MUST NOT generate documentation or write files. The sole output is `.moai/project/interview.md`.

**Interview Rounds (3 rounds maximum, configured in `.moai/config/sections/interview.yaml`):**

**Round 1: Ownership and Purpose**

Topic: Who maintains this project and what is the primary goal going forward?

Present via AskUserQuestion with exactly 4 options based on Phase 1 detected project type:
- Option 1 (Recommended): Active product being developed further: This codebase is actively developed and the documentation should reflect its current trajectory and roadmap.
- Option 2: Legacy system being maintained: The codebase is stable and the documentation should reflect its current state for maintenance and onboarding.
- Option 3: System being refactored or migrated: Major structural changes are planned and documentation should reflect the target state.
- Option 4: Type your own answer: Enter a custom response to describe the ownership context.

**Round 2: Constraints and Non-Goals**

Topic: What are the known constraints, technical debts, or things this project intentionally does NOT do?

Present via AskUserQuestion with exactly 4 options informed by Phase 1 analysis findings:
- Option 1 (Recommended): No known critical constraints: Document the codebase as-is without constraint annotations.
- Option 2: Performance or scalability constraints exist: There are known bottlenecks or scaling limits that should be documented.
- Option 3: Security or compliance constraints exist: Specific security requirements or compliance rules affect the architecture.
- Option 4: Type your own answer: Describe the specific constraints or non-goals for this project.

**Round 3: Documentation Priority**

Topic: What is the most important aspect to capture accurately in the documentation?

Present via AskUserQuestion with exactly 4 options:
- Option 1 (Recommended): Architecture and module boundaries: Prioritize documenting how the system is structured and how modules interact.
- Option 2: Technology stack and dependencies: Prioritize the frameworks, libraries, and their versions for onboarding.
- Option 3: Core business logic and data flow: Prioritize documenting what the system does and how data moves through it.
- Option 4: Type your own answer: Specify what should be documented with highest fidelity.

**Output:** Write all answers to `.moai/project/interview.md` with this structure:

```
# Project Interview

## Round 1: Ownership and Purpose
Question: {question asked}
Answer: {user's answer}

## Round 2: Constraints and Non-Goals
Question: {question asked}
Answer: {user's answer}

## Round 3: Documentation Priority
Question: {question asked}
Answer: {user's answer}
```

Pass `interview.md` to Phase 2 (User Confirmation) and Phase 3 (Documentation Generation) as additional context. Documentation agents MUST read interview.md before generating files.

---

## Phase 2: User Confirmation

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
If "Proceed": Continue to Phase 3 (see `doc-generation.md`).
If "Cancel": Exit with guidance.
