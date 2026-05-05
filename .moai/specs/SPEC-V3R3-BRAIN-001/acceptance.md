---
id: SPEC-V3R3-BRAIN-001
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P1
labels: [brain, ideation, workflow, handoff, claude-design, v3r3, acceptance]
issue_number: null
phase: "v3.0.0 — Phase 8 — Brain Workflow Introduction"
breaking: false
harness: standard
---

# Acceptance Criteria: SPEC-V3R3-BRAIN-001

This document defines the Given-When-Then scenarios that verify SPEC-V3R3-BRAIN-001 is correctly implemented. Each scenario maps to one or more EARS requirements (REQ-BRAIN-001..012).

---

## Scenario 1: Happy Path — Complete 7-Phase Execution

**Maps to**: REQ-BRAIN-001, REQ-BRAIN-005, REQ-BRAIN-009

**Given**:
- A fresh project with `.moai/brain/` directory present (created by `moai init`)
- `.moai/project/brand/brand-voice.md` exists with non-empty content
- No prior IDEA-NNN directories exist

**When**:
- User invokes `/moai brain "I want to build a habit-tracking mobile app for senior citizens"`

**Then**:
- The system creates `.moai/brain/IDEA-001/` directory
- Phase 1 Discovery executes: AskUserQuestion is called at least once with ToolSearch preload (verified via debug log)
- Phase 2 Diverge produces 5-15 divergent concepts (in-memory, not persisted as standalone file)
- Phase 3 Research produces `.moai/brain/IDEA-001/research.md` with at least 3 cited sources from WebSearch and at least 1 from Context7 (parallel call evidence in tool-use stream)
- Phase 4 Converge produces `.moai/brain/IDEA-001/ideation.md` containing a "Lean Canvas" section with all 9 blocks (Problem, Customer Segments, Unique Value Proposition, Solution, Channels, Revenue Streams, Cost Structure, Key Metrics, Unfair Advantage)
- Phase 5 Critical Evaluation appends an "Evaluation Report" section to `ideation.md`
- Phase 6 Proposal produces `.moai/brain/IDEA-001/proposal.md` containing a `### SPEC Decomposition Candidates` heading followed by 2-10 bullet entries matching grammar `- SPEC-{DOMAIN}-{NUM}: {scope}`
- Phase 7 Handoff produces all 5 files under `.moai/brain/IDEA-001/claude-design-handoff/`: `prompt.md`, `context.md`, `references.md`, `acceptance.md`, `checklist.md`
- `prompt.md` contains 5 sections: Goal, References, Brand Voice, Acceptance, Out-of-scope
- `prompt.md` contains NO MoAI-specific tokens (no `SPEC-`, no `.moai/`, no `manager-`)
- After Phase 7, AskUserQuestion is invoked with 3 options (a) proceed to project (Recommended), (b) review manually, (c) regenerate
- The workflow exits cleanly with status code 0

**Verification commands**:
```bash
ls .moai/brain/IDEA-001/
test -f .moai/brain/IDEA-001/research.md
test -f .moai/brain/IDEA-001/ideation.md
test -f .moai/brain/IDEA-001/proposal.md
test -d .moai/brain/IDEA-001/claude-design-handoff/
ls .moai/brain/IDEA-001/claude-design-handoff/ | wc -l   # expect 5
grep -q "### SPEC Decomposition Candidates" .moai/brain/IDEA-001/proposal.md
grep -E "^- SPEC-[A-Z][A-Z0-9]+-[0-9]{3}: " .moai/brain/IDEA-001/proposal.md | wc -l   # expect 2-10
grep -q "Lean Canvas" .moai/brain/IDEA-001/ideation.md
grep -v "SPEC-\|\.moai/\|manager-" .moai/brain/IDEA-001/claude-design-handoff/prompt.md  # exit 0 means no matches
```

---

## Scenario 2: Brand Context Absent — Graceful Default-Voice Fallback

**Maps to**: REQ-BRAIN-006

**Given**:
- A fresh project with `.moai/brain/` directory present
- `.moai/project/brand/` directory does NOT exist (or `brand-voice.md` is missing)
- No prior IDEA-NNN directories exist

**When**:
- User invokes `/moai brain "I want a developer productivity dashboard"`
- User completes Phases 1-6 normally

**Then**:
- Phase 7 Handoff produces all 5 files under `.moai/brain/IDEA-001/claude-design-handoff/`
- `claude-design-handoff/prompt.md` contains a section header "Brand Voice (default — please customize)"
- `claude-design-handoff/context.md` contains a placeholder warning at the top: a clearly visible note instructing the user to populate `.moai/project/brand/brand-voice.md` and regenerate, OR to manually edit the prompt before pasting into claude.com Design
- The workflow exits with status code 0 (graceful, not error)
- Before Phase 7, AskUserQuestion offers user the option to run brand interview first (option) or accept default (option) — Recommended option is "accept default and continue" to keep workflow short

**Verification commands**:
```bash
grep -q "Brand Voice (default — please customize)" .moai/brain/IDEA-001/claude-design-handoff/prompt.md
grep -qi "brand-voice.md" .moai/brain/IDEA-001/claude-design-handoff/context.md
test ! -d .moai/project/brand/   # confirm absent precondition
```

---

## Scenario 3: Research Phase — WebSearch Failure Handled Gracefully

**Maps to**: REQ-BRAIN-003

**Given**:
- A fresh project with `.moai/brain/` directory present
- WebSearch tool is unavailable or returns network error (simulate via test harness)
- Context7 MCP is available

**When**:
- User invokes `/moai brain "I want a code review automation tool"`
- Phase 3 Research executes

**Then**:
- The system attempts WebSearch and Context7 in parallel (single-message tool calls)
- WebSearch returns error or empty results
- Context7 returns valid documentation
- The system does NOT abort Phase 3 — it produces `.moai/brain/IDEA-001/research.md` with:
  - A "Sources" section listing all successfully-fetched sources (Context7 results)
  - A "Research Limitations" section noting that WebSearch was unavailable
- The workflow continues to Phase 4 (Converge) without crash
- The workflow exits with status code 0

**Verification commands**:
```bash
test -f .moai/brain/IDEA-001/research.md
grep -qi "research limitations\|websearch.*unavailable\|websearch.*failed" .moai/brain/IDEA-001/research.md
test -f .moai/brain/IDEA-001/ideation.md   # Phase 4 still executed
test -f .moai/brain/IDEA-001/proposal.md   # Phase 6 still executed
```

---

## Scenario 4: SPEC Decomposition List Parseable by `/moai plan --from-brain`

**Maps to**: REQ-BRAIN-004, REQ-BRAIN-007 (downstream integration)

**Given**:
- A completed brain workflow has produced `.moai/brain/IDEA-001/proposal.md`
- The proposal.md contains a `### SPEC Decomposition Candidates` section with the following entries:
  ```
  - SPEC-AUTH-001: User authentication and session management
  - SPEC-API-001: REST API for habit CRUD operations
  - SPEC-NOTIF-001: Push notification subsystem
  ```
- `/moai project --from-brain IDEA-001` has been run, producing `product.md`/`structure.md`/`tech.md`

**When**:
- User invokes `/moai plan` (no arguments)

**Then**:
- The plan workflow detects `.moai/brain/IDEA-001/proposal.md` exists
- The plan workflow parses the "SPEC Decomposition Candidates" section using the documented grammar
- AskUserQuestion is invoked with the 3 candidate SPECs as options (plus an "other / custom" option)
- The Recommended option is the first candidate (SPEC-AUTH-001) per AskUserQuestion convention
- User selects one candidate
- The plan workflow proceeds to generate `.moai/specs/SPEC-AUTH-001/{spec.md, plan.md, acceptance.md}` populated with brain-derived context (proposal scope copied as preliminary requirements)
- No errors thrown for the parser regardless of bullet-list whitespace variations

**Edge case extension**:
- If proposal.md contains a malformed entry like `- AUTH-001: missing prefix` (not matching grammar), the plan workflow surfaces it as a WARNING (not error) in the output and excludes it from the AskUserQuestion options.

**Verification commands**:
```bash
# After /moai plan execution:
test -d .moai/specs/SPEC-AUTH-001/
grep -q "habit\|authentication" .moai/specs/SPEC-AUTH-001/spec.md   # brain context propagated
# Parser robustness test:
echo "- AUTH-001: malformed" >> .moai/brain/IDEA-001/proposal.md
# Re-run /moai plan; expect warning in output, no error
```

---

## Scenario 5: Language Neutrality — Tech-Stack-Agnostic Output

**Maps to**: REQ-BRAIN-008, REQ-BRAIN-011

**Given**:
- A fresh project with `.moai/brain/` directory present
- No `.moai/project/tech.md` exists yet (so no tech-stack precedent)

**When**:
- User invokes `/moai brain "I want to build a Python data pipeline tool for ETL workflows"`
- The user's idea explicitly mentions "Python"

**Then**:
- Phase 1 Discovery clarifies the idea WITHOUT asking "which Python web framework" or similar tech-stack-specific questions — questions focus on user persona, problem space, success metrics
- Phase 3 Research includes general ETL ecosystem references but does NOT prioritize Python-specific tools over alternatives — at least 2 of the cited sources discuss tools written in non-Python languages (e.g., dbt, Apache NiFi, Talend, Pentaho)
- Phase 6 `proposal.md` does NOT contain phrases like "FastAPI", "Django", "Flask", "Pandas", "Airflow", "PySpark", or any specific Python library/framework name in the requirements section. The "Solution" block of Lean Canvas describes capabilities (e.g., "high-throughput ETL transformation engine") not implementations (e.g., "Pandas-based engine")
- The "SPEC Decomposition Candidates" section uses neutral SPEC IDs (e.g., `SPEC-PIPELINE-001`, `SPEC-API-001`) — NOT Python-specific (e.g., `SPEC-FASTAPI-001`)
- Phase 7 `prompt.md` does NOT mention any specific frontend or backend technology
- The user can subsequently choose any of the 16 supported MoAI-ADK languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) at the `/moai project` or `/moai plan` stage without conflict

**Verification commands**:
```bash
# Tech-stack mention probe (should produce zero matches in proposal.md):
grep -iE "fastapi|django|flask|pandas|airflow|pyspark|numpy|sqlalchemy" .moai/brain/IDEA-001/proposal.md | wc -l   # expect 0
# Or only in "User mentioned" / "Original idea" verbatim quote section if echoed:
grep -iE "fastapi|django|flask" .moai/brain/IDEA-001/proposal.md
# If matches, verify they're inside a "Original Idea (verbatim)" block, not in requirements/solution sections
# SPEC-decomposition language-neutral check:
grep -E "^- SPEC-(FASTAPI|DJANGO|FLASK|PANDAS|PYSPARK)-" .moai/brain/IDEA-001/proposal.md   # expect 0 matches
# Handoff prompt language-neutral check:
grep -iE "fastapi|django|flask|react|vue|angular|svelte" .moai/brain/IDEA-001/claude-design-handoff/prompt.md | wc -l   # expect 0
```

---

## Scenario 6: AskUserQuestion Enforcement — No Prose Questions

**Maps to**: REQ-BRAIN-012

**Given**:
- A fresh project
- Debug logging enabled (capture tool-use stream)

**When**:
- User invokes `/moai brain "I have a vague idea about productivity tools"`
- The idea is intentionally vague (clarity score ≤ 3) to trigger Discovery rounds

**Then**:
- Phase 1 Discovery is triggered
- Every question directed at the user is invoked via `AskUserQuestion` tool call
- Each `AskUserQuestion` call is preceded by `ToolSearch(query: "select:AskUserQuestion")` in the same turn (deferred-tool preload)
- The agent's output text body contains NO interrogative prose ending in `?` (no "What do you think?", "Which approach do you prefer?", "Should we proceed?")
- All option labels in AskUserQuestion calls are in the user's `conversation_language` (Korean, per `.moai/config/sections/language.yaml`)
- The first option of each AskUserQuestion has `(권장)` suffix (Korean) or `(Recommended)` (English)
- Each AskUserQuestion has at most 4 questions, each with at most 4 options
- After max 5 rounds (REQ-BRAIN-002 cap), Discovery terminates and proceeds to Phase 2

**Verification approach**:
```bash
# Inspect debug log for tool-use stream:
grep -E "AskUserQuestion|ToolSearch" .claude/logs/session-*.log | head -20
# Verify ToolSearch immediately precedes each AskUserQuestion (within same turn)
# Verify no orchestrator output contains a free-form prose question:
# (manual review of session transcript — should not see any "?-ending" sentence in agent prose addressed to user)
```

---

## Edge Cases

### Edge Case 1: Multiple IDEAs in Same Project

**Given**: `.moai/brain/IDEA-001/` and `.moai/brain/IDEA-002/` already exist

**When**: User runs `/moai brain "third idea"`

**Then**: System creates `.moai/brain/IDEA-003/` (auto-increment from max existing)

### Edge Case 2: Mid-Workflow Interrupt

**Given**: Phase 3 Research is in progress, partial `research.md` written

**When**: User cancels (Ctrl+C or session closes)

**Then**:
- `.moai/brain/IDEA-001/` directory exists with partial files
- Re-running `/moai brain "<same idea>"` SHOULD prompt user via AskUserQuestion: (a) resume from last completed phase (Recommended), (b) restart from Phase 1, (c) abort
- Resume option uses existing files; restart option creates IDEA-002

### Edge Case 3: User Rejects All Phase 7 Options

**Given**: REQ-BRAIN-009 AskUserQuestion at Phase 7 exit

**When**: User selects option (c) "regenerate handoff package with adjustments"

**Then**:
- AskUserQuestion follow-up asks what to adjust (brand voice, references, acceptance criteria)
- Phase 7 re-executes with user-specified adjustments
- All 5 handoff files are overwritten
- After regeneration, REQ-BRAIN-009 AskUserQuestion is re-invoked

### Edge Case 4: proposal.md Decomposition Section Empty

**Given**: For very small ideas, proposal.md contains 0-1 SPEC candidates instead of 2-10

**When**: Phase 6 generates proposal.md

**Then**:
- If 0 candidates: Phase 6 emits a warning in the output and includes a "SPEC Decomposition Candidates: TBD — idea may be too small for SPEC-driven workflow, consider direct `/moai plan` instead" placeholder section
- If 1 candidate: Acceptable, no warning
- The grammar `- SPEC-{DOMAIN}-{NUM}: {scope}` is enforced for any candidates present

### Edge Case 5: Concurrent `/moai brain` Invocations

**Given**: User somehow invokes `/moai brain` twice in parallel sessions

**When**: Both sessions try to create IDEA-NNN directory

**Then**:
- Out of scope for this SPEC's testing — concurrency-safety of IDEA-NNN auto-increment is documented as a known limitation
- v0.2 may add file lock; for v0.1, document in proposal.md output that concurrent invocations are not supported

---

## Definition of Done (Acceptance-Level)

For this SPEC to be considered DONE, ALL of the following must hold:

- [ ] All 6 numbered scenarios above pass (Scenario 1-6)
- [ ] All 5 edge cases produce expected behavior (manual verification or test cases)
- [ ] All 12 EARS requirements (REQ-BRAIN-001..012) traced to at least one scenario
- [ ] Frontmatter validation: 9 required fields present in all 4 plan artifacts (research.md, spec.md, plan.md, acceptance.md)
- [ ] `created_at` and `updated_at` are `2026-05-04` (NOT `created` / `updated`)
- [ ] `labels` is YAML array (not comma-separated string)
- [ ] `version` is quoted string (`"0.1.0"`, not `0.1`)
- [ ] `priority` is `P1` (uppercase P-prefix accepted enum)
- [ ] No implementation code in spec.md, plan.md, acceptance.md (only WHAT/WHY/HOW-AT-PLAN-LEVEL)

---

## EARS Requirement Coverage Matrix

| Requirement | Type | Scenario(s) Covering | Priority |
|-------------|------|---------------------|----------|
| REQ-BRAIN-001 | Event-Driven | #1 | P1 |
| REQ-BRAIN-002 | Event-Driven | #6 (5-round cap) | P1 |
| REQ-BRAIN-003 | Event-Driven | #1, #3 | P1 |
| REQ-BRAIN-004 | Event-Driven | #1, #4 | P1 |
| REQ-BRAIN-005 | Event-Driven | #1 (no MoAI tokens check) | P1 |
| REQ-BRAIN-006 | Optional | #1 (with brand), #2 (without brand) | P2 |
| REQ-BRAIN-007 | Event-Driven | #4 (downstream) | P1 |
| REQ-BRAIN-008 | Ubiquitous | #5 | P1 |
| REQ-BRAIN-009 | Event-Driven | #1, edge case #3 | P1 |
| REQ-BRAIN-010 | Unwanted | #1 (no auto-project) | P1 |
| REQ-BRAIN-011 | Unwanted | #5 | P1 |
| REQ-BRAIN-012 | Unwanted | #6 | P1 |

All 12 requirements covered by at least one scenario.

---

## References

- `spec.md` (sibling) — EARS requirements, file deliverables, exclusions
- `plan.md` (sibling) — implementation phases, risk analysis, MX tag plan
- `research.md` (sibling) — codebase context, decisions, risk surface
- `.claude/rules/moai/core/askuser-protocol.md` — AskUserQuestion + ToolSearch enforcement (Scenario 6)
- `.claude/rules/moai/development/coding-standards.md` — frontmatter schema validation

---

**Status**: audit-ready
