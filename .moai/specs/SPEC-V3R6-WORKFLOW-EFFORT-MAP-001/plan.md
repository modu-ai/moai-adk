# plan.md — SPEC-V3R6-WORKFLOW-EFFORT-MAP-001

> Implementation plan for the purpose-driven model+effort SSOT. Derived from `spec.md` §A–§J. Use **Milestones** (M1..M5) — NOT Rounds. This SPEC does not cross the ≥30-task SSE-stall threshold.

## §A. Context

This plan operationalizes `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` — a Tier M doctrine+config+script SPEC that creates a unified "agent purpose → (model, effort)" SSOT across three currently-disjoint surfaces:

1. Claude Code dynamic Workflow `agent()` calls (`.claude/workflows/*.js`)
2. Agent Teams `role_profiles` (`.moai/config/sections/workflow.yaml` + template mirror)
3. Session-handoff resume message Block 1 ultracode re-set line (`.claude/rules/moai/workflow/session-handoff.md` + output-styles render surface + both template mirrors)

The depth is **doctrine + script template only** (decision #1). No Go source under `internal/` is modified. The `effort` YAML field is declarative metadata consumed at runtime by the orchestrator (LLM) and workflow-script authors.

## §B. Known Issues (pre-existing — NOT absorbed)

The working tree is dirty. The following changes are **unrelated** to this SPEC and MUST NOT be absorbed into any commit this SPEC produces (REQ-WEM-014):

- 5 modified `.moai/config/sections/*.yaml` files (`git-convention.yaml`, `language.yaml`, `llm.yaml`, `quality.yaml`, `user.yaml`)
- Untracked directories: `.moai/design/web-console-handoff/`, `.moai/docs/harness-delivery-strategy.md`, `.moai/reports/sync-audit/`, `.moai/reports/worktree-rescue-*`

The implementer MUST scope every `git add` to explicitly the SPEC's target paths (enumerated in §C). A `git add -A` or `git add .` is PROHIBITED.

## §C. Pre-flight (target path enumeration)

The exact paths this SPEC is allowed to touch:

**Doctrine SSOT (source-of-truth):**
- `.claude/rules/moai/workflow/dynamic-workflows.md`
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/output-styles/moai/moai.md`

**Template mirrors (must be edited in same commit as SSOT — byte-parity):**
- `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`
- `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
- `internal/template/templates/.claude/output-styles/moai/moai.md`

**Config YAML (local + template mirror):**
- `.moai/config/sections/workflow.yaml`
- `internal/template/templates/.moai/config/sections/workflow.yaml`

**Workflow script:**
- `.claude/workflows/codemaps-extract.js`

**SPEC artifacts (this directory):**
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/spec.md` (already created)
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/plan.md` (this file)
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/acceptance.md`
- `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/progress.md`

**Total file count:** 13 files = 3 doctrine SSOT + 3 template mirrors + 2 config (local + template) + 1 workflow script + 4 SPEC artifacts (spec.md / plan.md / acceptance.md / progress.md). Of these, the 4 SPEC artifacts are created/owned by plan-phase (this directory); the remaining 9 are run-phase edit targets.

## §D. Constraints (carry into run-phase)

1. **No `make build`.** The doctrine/YAML/JS/MD surfaces are not embedded via a code path that requires regeneration for THIS SPEC's edits. (Note: `internal/template/templates/` IS the embed source, but `rule_template_mirror_test.go` reads files directly; `make build` regenerates `embedded.go` which is NOT load-bearing for the mirror-parity assertion. The iter-0 draft's reference to `embedded_mirror_test.go` was a phantom citation — that file does not exist.)
2. **Byte-parity invariants** on 3 mirror pairs (REQ-WEM-009).
3. **Template neutrality** in `internal/template/templates/` content (REQ-WEM-013): no internal SPEC IDs, REQ tokens, commit SHAs, `feedback_` refs, Audit citations.
4. **Session-handoff diet constraints** (AP-D-002/AP-D-004): ultracode re-set line is ONE line, no history/lesson/directive-escalation prose.
5. **7-role schema lock**: add `effort` field only; do NOT add/remove role keys.
6. **Scope discipline**: every `git add` MUST be explicit-path; no `git add -A`/`.` (REQ-WEM-014).

## §E. Self-Verification (run-phase §E.1 skeleton — populated by manager-develop)

The run-phase implementer (manager-develop) populates `progress.md` §E.2/§E.3 per the canonical manager-develop §E deliverable template. The plan-phase §E.1 audit-ready signal is emitted in `progress.md`.

## §F. Milestones

### M1 — SSOT taxonomy authoring (dynamic-workflows.md + template mirror)

**Scope:**
- Add a "Purpose-driven model+effort selection" subsection to `.claude/rules/moai/workflow/dynamic-workflows.md`. The subsection:
  - Cites the official effort guidance URL (`https://platform.claude.com/docs/en/build-with-claude/effort`).
  - Reproduces the §D taxonomy table from `spec.md` (8 purpose rows).
  - Documents the workflow `agent()` opt schema `{model, effort, agentType, isolation, phase, schema, label}` and the inheritance semantics (omit model → inherit main-loop; omit effort → inherit session).
  - States REQ-WEM-002: workflow-script authors SHALL set `effort` explicitly per taxonomy.
  - Gives the codemaps-extract.js pattern as a worked example (read-only-extract → `effort: 'low'`).
- Apply the IDENTICAL edit to `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`.
- Verify byte-parity: `diff` returns 0 lines.

**AC bindings:** AC-WEM-001, AC-WEM-002, AC-WEM-009b (dynamic-workflows pair — AC-enforced via diff), AC-WEM-010, AC-WEM-011 (template neutrality).

**Risks:** The dynamic-workflows.md file is already ~180 lines; adding the taxonomy subsection may push it past a readability threshold. Mitigation: keep the subsection compact (taxonomy table + 3 short paragraphs + 1 code example); do NOT duplicate the official effort guidance prose verbatim — cite it.

### M2 — codemaps-extract.js effort fix

**Scope:**
- Edit `.claude/workflows/codemaps-extract.js` line 53: change `agent(PROMPT(pkg), { label: ..., phase: 'Extract', agentType: 'Explore' })` to include `effort: 'low'`.
- Add a 1-line comment above the call documenting the purpose→effort mapping (`// read-only-extract purpose → effort: 'low' per dynamic-workflows.md § Purpose-driven model+effort selection`).
- Update the file header comment if needed to reference the new doctrine section.

**AC bindings:** AC-WEM-003.

**Risks:** None material. The `effort` opt is officially documented; Claude Code will accept `low` for a workflow subagent. No determinism impact (the script body still calls no wall-clock/random).

### M3 — role_profiles effort field (workflow.yaml local + template mirror)

**Scope:**
- Add an `effort` key to each of the 7 `role_profiles` entries in `.moai/config/sections/workflow.yaml`. Values per §D taxonomy:
  - `researcher`: `low` (read-only codebase exploration — official subagent recommendation)
  - `analyst`: `medium` (requirements analysis — balanced reasoning)
  - `architect`: `xhigh` (architecture decisions — intelligence-sensitive deep reasoning)
  - `implementer`: `xhigh` (code implementation — coding/agentic)
  - `tester`: `high` (test creation — coding but typically more mechanical than implementer)
  - `designer`: `high` (UI/UX design with MCP tools — design work but tool-augmented)
  - `reviewer`: `xhigh` (code review — verify-judge purpose, intelligence-sensitive)
- Apply the IDENTICAL 7 values to `internal/template/templates/.moai/config/sections/workflow.yaml`.
- Note: the two workflow.yaml files are NOT byte-parity today (119 diff lines from pre-existing local-only comments + structural differences). The constraint is that the 7 `effort` values MUST be byte-aligned across both files; the rest of the local-only divergence is left untouched (out of scope per REQ-WEM-014).
- Add a 1-line comment in both files above the `role_profiles` block: `# effort: declarative metadata consumed by orchestrator/workflow-script authors; no Go code reads this field (per SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 REQ-WEM-006).`
  - **Template neutrality note:** the template mirror's comment MUST omit the SPEC ID — use `# effort: declarative metadata consumed by orchestrator/workflow-script authors; no Go code reads this field.` instead (REQ-WEM-013).

**AC bindings:** AC-WEM-004, AC-WEM-005, AC-WEM-006 (git-diff confirms no Go file under internal/ modified — struct-field-absence is supporting evidence only), AC-WEM-011 (template neutrality on the comment), AC-WEM-012 (dirty-tree exclusion).

**Risks:**
- The local workflow.yaml has divergent inline comments referencing internal SPEC IDs (e.g., `SPEC-V3R2-WF-003`, `SPEC-V3R5-WORKFLOW-OPT-001`). These are PRE-EXISTING and out of scope; do NOT remove them. The template mirror is the neutrality-governed file; the local file is allowed to carry internal refs (it is not user-distributable).
- The 119-line diff is not eliminated by this SPEC; only the 7 `effort` lines need to match. Run `diff <(grep -A1 'effort:' local) <(grep -A1 'effort:' template)` post-edit to verify value-alignment.

### M4 — session-handoff Block 1 ultracode conditional line

**Scope:**
- Edit `.claude/rules/moai/workflow/session-handoff.md` `Field-by-Field Specification` Block 1 entry (currently a single bullet at line 76). Add a sub-bullet:
  - "Block 1 also carries a purpose-conditional `/effort ultracode` re-set line, emitted ONLY when the next SPEC's plan declares workflow fan-out (dynamic Workflow or Agent Teams). The line sits immediately after `ultrathink.`. Per dynamic-workflows.md, ultracode is NOT restored by the `ultrathink.` opener — it must be explicitly re-issued after `/clear` when the resumed session needs auto-orchestration. When the next SPEC does NOT need workflow fan-out, the ultracode line is omitted."
- Update the Canonical Format fenced block (around line 30) to show the conditional line in a comment or side-note (do NOT make it look unconditional — the example must show the `ultrathink.` line as required and the ultracode line as conditional).
- Update the Anti-Patterns general-hygiene list to add: "Omitting the `/effort ultracode` re-set line when the next SPEC needs workflow fan-out — the resumed session silently drops to non-ultracode effort."
- Apply the IDENTICAL edits to `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`.
- Verify byte-parity: `diff` returns 0 lines.

**AC bindings:** AC-WEM-007, AC-WEM-009a (session-handoff pair — CI-enforced via `rule_template_mirror_test.go`).

**Risks:** The session-handoff.md file is the most heavily cross-referenced rule in MoAI (output-styles/moai/moai.md §8 mirrors it bidirectionally). Any edit MUST be reflected in moai.md §8 in the SAME commit (M5). Mitigation: M4 and M5 land in the same run-phase commit.

### M5 — output-styles render surface parity + final verification

**Scope:**
- Edit `.claude/output-styles/moai/moai.md` §8 (the canonical render surface for the session-handoff template) to match the session-handoff.md Block 1 edit from M4. Per session-handoff.md `Cross-references`: "Before committing any edit to the Localization Table, the 6-block skeleton, the cut-line marker spec, or the Pre-emit self-check labels in THIS file, verify the parity check against the render surface."
- Apply the IDENTICAL edit to `internal/template/templates/.claude/output-styles/moai/moai.md`.
- Verify byte-parity on all 3 mirror pairs (PRIMARY check for AC-WEM-009a session-handoff is the CI test `rule_template_mirror_test.go`; the `diff` commands below are the PRIMARY check for AC-WEM-009b dynamic-workflows + output-styles, which are NOT in the Go allowlist):
  ```bash
  # AC-WEM-009a (CI-enforced): go test ./internal/template/ -run TestRuleTemplateMirror
  # AC-WEM-009b (AC-enforced via diff — these 2 pairs have NO CI allowlist coverage):
  diff .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md | wc -l   # expect 0
  diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md | wc -l                              # expect 0
  # AC-WEM-009a corroboration (also re-verified by the CI test):
  diff .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md | wc -l       # expect 0
  ```
- Verify codemaps-extract.js: `grep -n "effort:" .claude/workflows/codemaps-extract.js` returns the `low` line.
- Verify role_profiles 7-key effort in both YAML files (use the corrected awk range — `role_profiles:` is the LAST child of `team:`, bounded by the next top-level `workflow:` sibling `token_budget:` at 4-space indent; the iter-0 `/,/^        patterns:/` end-anchor was structurally broken because `patterns:` PRECEDES `role_profiles:` in the file):
  ```bash
  awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml | grep -cE '^            [a-z_]+:'    # expect 7 role keys
  awk '/^        role_profiles:/,/^    token_budget:/' .moai/config/sections/workflow.yaml | grep -c 'effort:'                    # expect 7
  awk '/^        role_profiles:/,/^    token_budget:/' internal/template/templates/.moai/config/sections/workflow.yaml | grep -c 'effort:'  # expect 7
  ```
- Verify decision #1 invariant (no Go source under `internal/` modified by this SPEC's commits) — the PRIMARY check is git-diff-based (the iter-0 `grep '\.Effort'` was vacuous: `RoleProfileEntry` has no `Effort` field so no Go code could reference it regardless):
  ```bash
  git diff --name-only origin/main -- internal/ | grep '\.go$'    # expect: 0 lines (no Go files touched by this SPEC)
  # SUPPORTING evidence only (struct-field-absence — proves the field cannot couple even if a reader existed):
  grep -A6 'type RoleProfileEntry struct' internal/config/types.go | grep -i 'effort'    # expect: 0 matches
  ```
- Verify template neutrality (no SPEC IDs / REQ tokens / SHAs in template mirror edits):
  ```bash
  grep -rE 'SPEC-V3R6-WORKFLOW-EFFORT-MAP|REQ-WEM-' internal/template/templates/.claude/rules/moai/workflow/{dynamic-workflows,session-handoff}.md internal/template/templates/.claude/output-styles/moai/moai.md internal/template/templates/.moai/config/sections/workflow.yaml
  # expect: 0 matches
  ```

**AC bindings:** AC-WEM-008, AC-WEM-009a (CI-enforced session-handoff pair), AC-WEM-009b (AC-enforced dynamic-workflows + output-styles pairs), AC-WEM-003, AC-WEM-004, AC-WEM-005, AC-WEM-006, AC-WEM-011.

**Risks:** The moai.md render-surface file is ~53KB. The §8 section is the most complex part (carries the Localization Table, the 6-block skeleton, the Pre-emit self-check). Mitigation: limit the edit to the Block 1 field-spec bullet and the self-check list; do NOT touch the Localization Table (no locale changes needed — `ultracode` and `ultrathink.` are keyword tokens preserved verbatim across locales).

## §G. Anti-Patterns (run-phase traps)

- **AP-WEM-001 — git add -A / git add .** Absorbs the 5 dirty-tree YAML files + untracked dirs. Mitigation: every `git add` is explicit-path.
- **AP-WEM-002 — Editing SSOT without editing template mirror in same commit.** Breaks byte-parity. Enforcement is SPLIT: session-handoff.md parity IS CI-enforced by `rule_template_mirror_test.go` (the real mirror test, allowlist covers session-handoff.md only); dynamic-workflows.md and output-styles/moai/moai.md parity are NOT CI-enforced and MUST be verified by `diff`/`cmp` in AC-WEM-009b. (The iter-0 draft cited a phantom `embedded_mirror_test.go` — that file does not exist; the real file is `rule_template_mirror_test.go`.) Mitigation: each milestone lists BOTH paths and requires same-commit edit.
- **AP-WEM-003 — Leaking SPEC ID / REQ token into template mirror.** Violates template neutrality (REQ-WEM-013). Mitigation: when the SSOT carries `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` or `REQ-WEM-NNN`, the template mirror carries generic prose only.
- **AP-WEM-004 — Embedding history/lesson prose in the ultracode re-set line.** Violates session-handoff diet constraints (AP-D-002/AP-D-004). Mitigation: the line is literally `/effort ultracode` — one line, no annotation.
- **AP-WEM-005 — Adding/removing role keys.** Violates 7-role schema lock (team-protocol.md). Mitigation: only add the `effort` field to existing keys.
- **AP-WEM-006 — Modifying Go under `internal/`.** Violates decision #1. Mitigation: the implementer greps `internal/` post-edit to confirm no Go file was touched by this SPEC's commits.
- **AP-WEM-007 — Touching the Localization Table in moai.md §8.** Unnecessary; `ultracode`/`ultrathink.` are keyword tokens preserved verbatim across locales. Mitigation: limit moai.md edit to the field-spec bullet and self-check list.

## §H. Cross-References

- `spec.md` §A–§J — requirements, taxonomy, AC matrix.
- `acceptance.md` — Given-When-Then scenarios for AC-WEM-001..008, 009a, 009b, 010..012 (13 ACs total).
- `progress.md` — §E.1 plan-phase audit-ready signal (populated); §E.2-§E.5 skeletons for run/sync/Mx phases.
- `.claude/rules/moai/development/sprint-round-naming.md` — Milestones (M1..M5) are the correct unit for this SPEC (sub-30 tasks, no SSE-stall split needed → NO Rounds).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` is owned by manager-spec (this plan-phase commit); `draft → in-progress` is owned by manager-develop (M1 run-phase commit).

## §I. Path Forward (post-plan-phase)

1. plan-auditor independent audit (target ≥ 0.80; this SPEC is Tier M, threshold 0.80).
2. Implementation Kickoff Approval (human gate per §19.1 of CLAUDE.local.md).
3. `/moai run SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` (manager-develop, cycle_type=ddd — this is brownfield doctrine editing, DDD characterization-first applies).
4. `/moai sync SPEC-V3R6-WORKFLOW-EFFORT-MAP-001` (manager-docs).
5. Mx-phase close (orchestrator-direct or manager-docs).
