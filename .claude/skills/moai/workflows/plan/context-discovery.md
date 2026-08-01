---
description: "Plan Phase 2/3 — Project exploration and clarity evaluation before deep research begins"
user-invocable: false
metadata:
  parent: moai-workflow-plan
  phase: "Phase 2 / Phase 3: Context Discovery and Clarity Evaluation"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Phase Sequence

### Phase 2: Project Exploration (Optional)

Agent: Explore subagent (read-only codebase analysis)

When to run:
- User provides vague or unstructured request
- Need to discover existing files and patterns
- Unclear about current project state

When to skip:
- User provides clear SPEC title (e.g., "Add authentication module")
- Resume scenario with existing SPEC context

Tasks for the Explore subagent:
- If .moai/project/codemaps/ exists: Use as architecture baseline to accelerate exploration
- Find relevant files by keywords from user request
- Locate existing SPEC documents in .moai/specs/
- Identify implementation patterns and dependencies
- Discover project configuration files
- Read target directories in depth — understand deeply how each module works, its intricacies and side effects
- Study cross-module interactions in great detail — trace data flow through the system
- Go through related test files to understand expected behavior and edge cases
- Report comprehensive results for Phase 8 context

### Phase 3: Clarity Evaluation (Conditional)

Purpose: Evaluate how clearly the user's request is specified before beginning deep research. A vague request produces a weaker SPEC; this phase detects vagueness early and gathers missing context through a structured interview.

**Skip conditions (any one is sufficient):**
- `--skip-interview` flag is present in $ARGUMENTS
- Input matches `resume SPEC-XXX` pattern (resuming an existing draft)
- Input contains 5 or more distinct technical keywords (e.g., framework names, file paths, function names, domain terms)
- `interview.enabled: false` in `.moai/config/sections/interview.yaml`

**Clarity Scoring (1-10):**

Evaluate the user's input against five dimensions:

1. Technical keyword count: 2+ points for 3-4 keywords; 1 point for 1-2; 0 for none
2. Action verbs specificity: "add CRUD endpoints for user profile" scores higher than "improve the app"
3. File or module mentions: explicit file paths or module names each add 1 point
4. Generic nouns penalty: deduct 1 point for each vague noun like "system", "feature", "thing"
5. Scope boundary clarity: a defined boundary ("only the POST /users endpoint, no auth changes") adds 2 points

**Score-to-rounds mapping:**

| Clarity Score | Interview Rounds |
|---|---|
| 1-3 | 0 (request too vague — ask one broad clarification question instead) |
| 4-6 | 2 rounds maximum |
| 7-10 | 5 rounds maximum |

Log the score: "Clarity score: {N}/10 — proceeding with {M} interview round(s)."

If score is 1-3: Preload `ToolSearch(query: "select:AskUserQuestion")`, then use a single AskUserQuestion asking for a clearer description, then re-evaluate. Do not enter the full interview loop.

---

**Next phase:** Read `workflows/plan/clarity-interview.md` to continue with Phase 4 Deep Interview Loop.
