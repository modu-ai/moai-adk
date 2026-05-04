# /moai design — Hybrid Design Workflow

## SPEC Reference

SPEC-AGENCY-ABSORB-001: REQ-ROUTE-001 through REQ-ROUTE-008, REQ-FALLBACK-001 through REQ-FALLBACK-003, REQ-BRIEF-001 through REQ-BRIEF-003, REQ-DETECT-003
SPEC-V3R3-DESIGN-PIPELINE-001: REQ-DPL-005, REQ-DPL-008 (Phase 2 — Workflow Routing)

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

### Check 3: Brand Context Loader (REQ-DPL-008)

After brand files are confirmed present, load and cache brand context:
1. Read `.moai/project/brand/brand-voice.md` → cache as `brand_voice`
2. Read `.moai/project/brand/visual-identity.md` → cache as `visual_identity`
3. Read `.moai/project/brand/target-audience.md` → cache as `target_audience`

**Brand-conflict warning** (design constitution §3.1 — brand wins on conflict):
- After loading, compare token definitions in `.moai/design/tokens.json` (if exists) against `visual_identity` color/typography values.
- If mismatch detected: output warning "Brand conflict detected: token values differ from visual-identity.md. Brand context takes precedence." and list conflicting keys.
- Proceed regardless — brand values are authoritative; downstream agents must use `brand_voice`/`visual_identity` over stale tokens.

### Check 4: Previous Path Selection (REQ-DPL-005 — idempotency)

If `.moai/design/path-selection.json` exists (written by `internal/design/pipeline`):
- Surface previous selection via AskUserQuestion:
  - Option 1 (Recommended): "Resume [Path X] — continue from last session"
  - Option 2: "Select new path — override previous selection"
- If user selects Resume: skip Phase 1, jump directly to the corresponding Phase (A/B1/B2).
- If user selects new path: overwrite `path-selection.json` after Phase 1.

---

## Phase 1: Route Selection (REQ-ROUTE-002, REQ-ROUTE-003, REQ-ROUTE-006, REQ-DPL-005)

Use AskUserQuestion to present the three design paths.

**Option order** (CLAUDE.md §8: recommended-first rule):

Option 1 (Recommended): Path A (Claude Design)
- "Claude Design handoff bundle import (most stable, recommended for new users)"
- Requirements: Claude.ai Pro, Max, Team, or Enterprise subscription
- Output: Design tokens, component manifests, static assets from Claude Design session

Option 2: Path B1 (Figma)
- "Figma file via dynamic figma-extractor (requires Figma credentials)"
- Requirements: Figma API token in environment or `.moai/config/sections/design.yaml`
- Output: Extracted design tokens and component specs from Figma file

Option 3: Path B2 (Pencil)
- "Pencil .pen files via dynamic pencil-mcp (requires .pen files in project)"
- Requirements: `.pen` files present in `.moai/design/` or project root
- Output: Design artifacts rendered from Pencil files via pencil-mcp

**Subscription override** (REQ-ROUTE-006): When `subscription.tier: "pro-or-below"` in user.yaml or user states no Claude Design access:
- Swap Option 1 ↔ Option 2 (B1 becomes recommended).
- Add to Path A description: "Requires Claude.ai Pro or higher subscription."

**After selection**: Write `path-selection.json` via `internal/design/pipeline.WritePathSelection` with:
- `path`: "A" | "B1" | "B2"
- `brand_context_loaded`: true (since Check 3 completed)
- `spec_id`: current SPEC-ID or empty string
- `ts`: current UTC timestamp
- `session_id`: `${CLAUDE_SESSION_ID}`

**No-response handling** (REQ-ROUTE-007): Re-present up to 3 times. After 3 failures, output "Selection not confirmed. Resume with `/moai design` when ready." and stop.

---

## Brain Handoff Bundle Auto-Detection

<!-- Verifies REQ-BRAIN-005: brain output (claude-design-handoff/) consumed by /moai design --path A -->

When `/moai design --path A` is invoked WITHOUT a `--bundle` argument:

**Step 0: Scan for brain handoff bundles**

1. Glob for `.moai/brain/IDEA-*/claude-design-handoff/prompt.md` (indicates a completed brain Phase 7 output).
2. Collect all matching IDEA directories as `brain_bundles` (sorted by IDEA number descending — newest first).
3. If `brain_bundles` is non-empty AND no `--bundle` argument was provided:

```
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    header: "Brain 워크플로우 핸드오프 번들 감지됨",
    question: "Brain 워크플로우에서 생성된 Claude Design 핸드오프 패키지를 발견했습니다. 어떻게 진행하시겠습니까?",
    options: [
      {
        label: "Brain 핸드오프 패키지 사용 (권장)",
        description: ".moai/brain/IDEA-NNN/claude-design-handoff/ 의 prompt.md를 Claude Design에 붙여넣기하세요. 완료 후 다운로드한 번들 경로를 입력합니다."
      },
      {
        label: "수동으로 번들 경로 입력",
        description: "이미 Claude Design에서 디자인을 완료하고 번들을 다운로드한 경우 선택하세요."
      }
    ]
  }]
})
```

4. If user selects "Brain 핸드오프 패키지 사용":
   - Display the path to the prompt.md: `.moai/brain/IDEA-NNN/claude-design-handoff/prompt.md`
   - Output instructions: "Open the prompt.md file, copy its contents, and paste into Claude Design at https://claude.ai/design"
   - Wait for user to complete the Claude Design session and download the bundle.
   - Proceed to Step A2 (collect bundle path from user).

5. If `brain_bundles` is empty OR `--bundle` was provided: skip this step, proceed directly to Phase A.

---

## Phase A: Claude Design Import Path (REQ-ROUTE-004)

When Path A (Claude Design) is selected:

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
- AskUserQuestion: "Would you like to switch to Path B1 (Figma) or Path B2 (Pencil)?"
- If yes: return to Phase 1.
- If no: stop and wait for user to provide a corrected bundle path.

---

## Phase B1: Figma Extractor Path (REQ-DPL-005)

When Path B1 (Figma) is selected:

Step B1-1: Validate Figma credentials:
- Check `FIGMA_API_TOKEN` env var or `.moai/config/sections/design.yaml` `figma.api_token`.
- If missing: output "Figma API token required. Set FIGMA_API_TOKEN or configure design.yaml." and stop.

Step B1-2: Generate figma-extractor meta-harness:
- Invoke `moai-meta-harness` with target="figma-extractor".
- Meta-harness generates a dynamic agent skill for Figma file extraction.
- Pass: Figma file URL (collected via AskUserQuestion), brand context (`visual_identity`).

Step B1-3: Brand context enforcement:
- After extraction, compare extracted color/typography tokens against `visual_identity`.
- If conflict: apply brand values and log overrides ("Overriding Figma token `<key>` with brand value `<val>`").
- Proceed to Phase B-Common.

---

## Phase B2: Pencil MCP Path (REQ-DPL-005)

When Path B2 (Pencil) is selected:

Step B2-1: Verify .pen files:
- Glob `.moai/design/*.pen` and `*.pen`.
- If none found: output "No .pen files found. Place Pencil files in `.moai/design/` or project root." and stop.

Step B2-2: Generate pencil-mcp meta-harness:
- Invoke `moai-meta-harness` with target="pencil-mcp".
- Meta-harness generates a dynamic agent skill for Pencil file rendering.
- Pass: `.pen` file paths, brand context (`visual_identity`).

Step B2-3: Brand context enforcement:
- After rendering, compare rendered tokens against `visual_identity`.
- If conflict: apply brand values and log overrides.
- Proceed to Phase B-Common.

---

## Phase B-Common: Shared Code-Based Design Steps

After Phase B1 or B2 completes token extraction:

Step BC-1: Load design context:
- Check `.moai/design/` exists. If absent: skip, log "design docs not initialized".
- Invoke `moai-workflow-design-context` skill with `dir=".moai/design"`.
- Receive consolidated context block (token-capped per REQ-5 algorithm).
- Prepend context block to downstream subagent prompts.

Step BC-2: Generate BRIEF (REQ-BRIEF-001, REQ-BRIEF-002, REQ-BRIEF-003):
- Invoke `manager-spec` in BRIEF generation mode.
- Required BRIEF sections: `## Goal`, `## Audience`, `## Brand`
- Auto-inject brand content from brand files if Brand section is empty.
- If brand files missing: halt with `BRIEF_SECTION_INCOMPLETE`.

Step BC-3: Load code-based design skills:
- Load `moai-domain-copywriting`
- Load `moai-domain-brand-design`
- Load `moai-workflow-gan-loop`

Step BC-4: Delegate to `expert-frontend`:
- Prompt includes: BRIEF, brand context summary, loaded skill references, `.moai/config/sections/design.yaml`.
- `expert-frontend` generates copy (JSON) and design tokens concurrently.

Step BC-5: Proceed to Phase C (quality gate).

---

## Phase C: Quality Gate (REQ-ROUTE-008)

After Path A or B1/B2 produces design artifacts:

Step C1: Invoke `moai-workflow-gan-loop`:
- Pass: BRIEF, design artifacts, copy JSON, `.moai/config/sections/design.yaml`
- Loop: Builder-Evaluator iterations (max 5) until `pass_threshold` (0.75) met.

Step C2: On loop PASS:
- Output evaluation report summary.
- Proceed to optional E2E testing.

Step C3: On loop FAIL (iterations exhausted):
- Present failure report via AskUserQuestion with three options:
  1. Accept current output (force-pass)
  2. Adjust criteria and restart loop
  3. Switch design approach (return to Phase 1)

Step C4: Optional E2E testing (when Playwright or claude-in-chrome MCP available):
- Run `/moai e2e` on generated design output.
- Surface interaction failures as blocking issues.

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

If any section is missing or empty, `manager-spec` must return `BRIEF_SECTION_INCOMPLETE`.

---

## Thin Command Routing

The `/moai design` slash command file is a thin routing wrapper. All logic lives here in this file:

```
Use Skill("moai") with arguments: design $ARGUMENTS
```

---

## Custom Harness Extension (Optional)

@.moai/harness/design-extension.md

*(이 파일은 `/moai project --harness`로 생성되며 Q13 답변이 "Advanced"일 때만 만들어집니다. 파일이 없으면 자동으로 skip됩니다.)*
