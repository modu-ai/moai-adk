# /moai design — Hybrid Design Workflow

## SPEC Reference

SPEC-AGENCY-ABSORB-001: REQ-ROUTE-001 through REQ-ROUTE-008, REQ-FALLBACK-001 through REQ-FALLBACK-003, REQ-BRIEF-001 through REQ-BRIEF-003, REQ-DETECT-003

---

## Phase 0: Pre-flight Checks

Before presenting the route selection, perform these checks in order:

### Check 1: Existing .agency/ detection (REQ-DETECT-003)

If `.agency/` directory exists AND `.moai/project/brand/` does not exist:
- Output warning before route selection: "agency data detected — run `moai migrate agency` to migrate your brand context first."
- Include `moai migrate agency --dry-run` as the preview command.
- Continue to route selection (do not block).

### Check 2: Brand context existence (REQ-ROUTE-001)

Check whether `.moai/project/brand/` contains the three brand files:
- `brand-voice.md`
- `visual-identity.md`
- `target-audience.md`

If any file is missing or contains `_TBD_` markers:
- Skip route selection.
- Propose brand interview: "Brand context is incomplete. Run the brand interview to define your brand voice, visual identity, and audience before designing."
- Invoke `manager-spec` with BRIEF interview mode to populate the missing files.
- After interview completes, resume from Phase 0 Check 2.

If partial brand context exists (some files present, some missing):
- Output "Incomplete brand context: missing `<filenames>`."
- Offer to complete only the missing files via targeted interview.

---

## Phase 1: Route Selection (REQ-ROUTE-002, REQ-ROUTE-003, REQ-ROUTE-006)

Use AskUserQuestion to present the two paths.

**Default option order** (Pro/Max/Team/Enterprise subscription assumed):

Option 1 (Recommended): Claude Design import
- "Work in Claude.ai/design to create your design, then export a handoff bundle. Claude Code imports the bundle automatically."
- Requirements: Claude.ai Pro, Max, Team, or Enterprise subscription
- Output: Design tokens, component manifests, static assets from your Claude Design session

Option 2: Code-based brand design
- "Generate design tokens, component specs, and layout from your brand identity files using moai-domain-brand-design skill."
- Requirements: Complete `.moai/project/brand/visual-identity.md`
- Output: Design tokens JSON, component specifications, layout grid, same artifact structure as path A

**Subscription override** (REQ-ROUTE-006): When the user has declared `subscription.tier: "pro-or-below"` in `.moai/config/sections/user.yaml` or explicitly states they do not have Claude Design access:
- Reverse the option order: code-based path becomes Option 1 (Recommended)
- Add to Claude Design option description: "Requires Claude.ai Pro or higher subscription."
- Do not disable the Claude Design option — keep it available for future subscription upgrades.

**No-response handling** (REQ-ROUTE-007): If the user does not select an option, re-present the question. Maximum 3 re-presentations. After 3 failed attempts, output "Selection not confirmed. Resume with `/moai design` when ready." and stop without closing the session.

---

## Phase A: Claude Design Import Path (REQ-ROUTE-004)

When path A (Claude Design) is selected:

Step A1: Guide the user to Claude.ai:
- Output: "Open https://claude.ai/design in your browser."
- Output: "Describe your design brief to Claude Design."
- Output: "When complete, use the Export or Share menu to download a handoff bundle (ZIP format)."
- Output: "Save the bundle to your local filesystem."

Step A2: Collect bundle path:
- AskUserQuestion: "What is the local file path to the downloaded handoff bundle?"
- Validate that the path ends in `.zip` or `.html`.

Step A3: Invoke `moai-workflow-design-import` skill:
- Pass: bundle file path, project brief, `.moai/config/sections/design.yaml`
- Expected output: `.moai/design/tokens.json`, `.moai/design/components.json`, `.moai/design/assets/`

Step A4: On import success:
- Proceed to Phase C (common quality gate).
- Load `moai-workflow-gan-loop` and pass the imported design artifacts.

Step A5: On import failure:
- Present the error code and message from `moai-workflow-design-import`.
- AskUserQuestion: "Would you like to switch to path B (code-based design)?"
- If yes: proceed to Phase B.
- If no: stop and wait for user to provide a corrected bundle path.

---

## Phase B: Code-Based Design Path (REQ-ROUTE-005)

When path B (code-based) is selected:

Step B1: Load skills:
- Load `moai-domain-copywriting`
- Load `moai-domain-brand-design`
- Load `moai-workflow-gan-loop`

Step B2: Load brand context:
- Read `.moai/project/brand/brand-voice.md`
- Read `.moai/project/brand/visual-identity.md`
- Read `.moai/project/brand/target-audience.md`

### Phase B2.5: Load .moai/design/ Context

1. Check .moai/design/ exists. If absent: skip, log "design docs not initialized".
2. Check design_docs.auto_load_on_design_command. If false: skip (user may invoke standalone).
3. Read README.md for attach rules (if present).
4. Invoke moai-workflow-design-context skill with dir=".moai/design".
5. Receive consolidated context block (Markdown, token-capped per REQ-5 algorithm).
6. Prepend context block to the orchestrator's next subagent prompt (expert-frontend or moai-domain-brand-design).
7. Proceed to Phase B3 (BRIEF generation).

### Phase B2.6: Pencil Path (Conditional)

Executes after Phase B2.5 and before Phase B3. This phase is conditional: it activates only when Pencil file/folder signals are present. It does not block Phase B3 on error.

#### Precondition Check (REQ-PENCIL-001, REQ-PENCIL-002)

Check both conditions:
1. `.moai/design/pencil-plan.md` exists.
2. At least one `.pen` file exists in `.moai/design/` or the project root (use Glob: `.moai/design/*.pen` and `*.pen`).

If either condition is not met: skip Phase B2.6 silently (no user-visible error message, no stderr output) and proceed directly to Phase B3 (graceful skip per REQ-PENCIL-002).

#### Skill Invocation (REQ-PENCIL-003)

When both preconditions are met:
- Invoke `moai-workflow-pencil-integration` skill synchronously.
- Wait for the skill to return (success or structured error) before proceeding to Phase B3.
- Phase B3 MUST NOT start until the skill returns.

#### Error Handling (AC-8)

When the skill returns a structured error (`PENCIL_MCP_UNAVAILABLE`, `PENCIL_CONNECTION_FAILED`, `PENCIL_PLAN_SYNTAX_ERROR`, or `PENCIL_BATCH_FAILED`):
- Log the error code and message for the session record.
- Do NOT abort the overall `/moai design` workflow.
- Continue to Phase B3 immediately.
- "Fallback" here means continuation within Path B — not a return to Phase 1 route selection.

Exception: `PENCIL_PLAN_SYNTAX_ERROR` and `PENCIL_BATCH_FAILED` are halting errors within the skill itself. The skill returns them to the orchestrator, and the orchestrator continues to Phase B3 after logging them.

Step B3: Generate BRIEF (REQ-BRIEF-001, REQ-BRIEF-002, REQ-BRIEF-003):
- Invoke `manager-spec` in BRIEF generation mode.
- Required BRIEF sections: `## Goal`, `## Audience`, `## Brand`
- If Brand section is empty: auto-inject key content from the three brand files with source citation (`> source: .moai/project/brand/<filename>`)
- If brand files are missing: halt with `BRIEF_SECTION_INCOMPLETE` and request brand interview.

Step B4: Delegate to `expert-frontend`:
- Prompt includes:
  - The BRIEF document
  - Brand context summary from the three brand files
  - Reference to loaded skills: `moai-domain-copywriting`, `moai-domain-brand-design`
  - Design parameter reference: `.moai/config/sections/design.yaml`
- `expert-frontend` generates copy (JSON format) and design tokens concurrently.

Step B5: Proceed to Phase C (common quality gate).

---

## Phase C: Quality Gate (REQ-ROUTE-008)

After either path A or path B produces design artifacts:

Step C1: Invoke `moai-workflow-gan-loop`:
- Pass: BRIEF, design artifacts, copy JSON, `.moai/config/sections/design.yaml`
- Loop executes Builder-Evaluator iterations (max 5) until `pass_threshold` (0.75) is met or iterations are exhausted.

Step C2: On loop PASS:
- Output evaluation report summary.
- Proceed to optional E2E testing.

Step C3: On loop FAIL (iterations exhausted):
- Present failure report via AskUserQuestion with three options:
  1. Accept current output (force-pass)
  2. Adjust criteria and restart loop
  3. Switch design approach (restart from Phase 1)

Step C4: Optional E2E testing (when Playwright or claude-in-chrome MCP available):
- Run `/moai e2e` on the generated design output.
- Surface any interaction failures as blocking issues.

---

## BRIEF Section Requirements (REQ-BRIEF-001)

When `manager-spec` generates the BRIEF document for a design task, it must include:

```markdown
## Goal
<What the design must accomplish. Specific outcome, not vague intent.>

## Audience
<Who will use or see the design. Reference target-audience.md persona names.>

## Brand
<Visual and verbal identity constraints. Auto-injected from brand files if not provided.>
> source: .moai/project/brand/brand-voice.md
> source: .moai/project/brand/visual-identity.md
> source: .moai/project/brand/target-audience.md
```

If any section is missing or empty, `manager-spec` must return `BRIEF_SECTION_INCOMPLETE` and refuse to generate the BRIEF.

---

## Thin Command Routing

The `/moai design` slash command file is a thin routing wrapper. All logic lives here in this file. The command file contains only:

```
Use Skill("moai") with arguments: design $ARGUMENTS
```

This document is the authoritative source for the design subcommand workflow logic.
