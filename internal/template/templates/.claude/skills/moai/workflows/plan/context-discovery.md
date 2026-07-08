---
description: "Plan Phase 0/1A/0.3 — Brain proposal detection, project exploration, and clarity evaluation before deep research begins"
user-invocable: false
metadata:
  parent: moai-workflow-plan
  phase: "Phase Step 0 / Phase 1A / Phase 0.3: Context Discovery and Clarity Evaluation"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

## Brain Context Auto-Detection

<!-- Verifies: SPEC Decomposition Candidates surfaced to user via AskUserQuestion -->
<!-- Verifies: /moai plan detects proposal.md and presents SPEC candidates -->

When `/moai plan` is invoked (with or without arguments), perform this pre-execution check:

### Step 0: Brain Proposal Detection

1. **Scan** for `.moai/brain/IDEA-*/proposal.md` files (Glob: `.moai/brain/IDEA-[0-9]*/proposal.md`).
2. If any proposal.md files are found:
   a. Read the most recent file (highest IDEA-NNN number by directory name).
   b. Parse the `### SPEC Decomposition Candidates` section using grammar:
      ```
      Grammar: ^- SPEC-[A-Z][A-Z0-9]+-[0-9]{3}: .+$
      ```
   c. Collect all matching entries as `brain_candidates`.
   d. Non-matching entries emit a WARNING in output (but do NOT error out — defensive parser).

3. If `brain_candidates` is non-empty AND user did not provide a specific SPEC title in $ARGUMENTS:
   - Surface candidates via AskUserQuestion (per `askuser-protocol.md`):
     ```
     ToolSearch(query: "select:AskUserQuestion")
     AskUserQuestion({
       questions: [{
         header: "Brain 워크플로우 SPEC 후보",
         question: "Brain 워크플로우에서 생성된 SPEC 분해 후보가 있습니다. 어느 것을 계획하시겠습니까?",
         options: [
           { label: "<first candidate> (권장)", description: "Brain IDEA에서 자동 감지된 첫 번째 후보" },
           { label: "<second candidate>", description: "..." },
           ...up to 4 options total (use "직접 입력" as last option for custom SPEC title)
         ]
       }]
     })
     ```
   - User selection becomes the SPEC title for Phase 1B.
   - [HARD] NEVER auto-create SPECs from candidates — user MUST select explicitly.

4. If user provided a specific SPEC title OR selected "직접 입력": proceed normally to Phase 1A.

5. If no brain candidates found: skip this check, proceed normally.

**Defensive Parser Rules**:
- Entries matching the grammar are offered as candidates.
- Entries NOT matching (e.g., `- AUTH-001: missing prefix`, `- SPEC-001: missing domain`) emit:
  `[WARNING] Skipped malformed brain candidate: "<entry>" — expected format: - SPEC-{DOMAIN}-{NNN}: {scope}`
- Parser warnings do NOT block plan execution.
- Maximum 9 candidates surfaced (AskUserQuestion option limit: 4 per question, minus "직접 입력").

---

## Phase Sequence

### Phase 1A: Project Exploration (Optional)

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
- Report comprehensive results for Phase 1B context

### Phase 0.3: Clarity Evaluation (Conditional)

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

**Next phase:** Read `workflows/plan/clarity-interview.md` to continue with Phase 0.3.1 Deep Interview Loop.
