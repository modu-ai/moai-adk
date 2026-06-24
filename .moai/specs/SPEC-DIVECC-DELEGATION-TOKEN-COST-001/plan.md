# Implementation Plan — SPEC-DIVECC-DELEGATION-TOKEN-COST-001

> Tier S plan. Pairs with plan.md of SPEC-DIVECC-EXTENSION-COST-LADDER-001 (N2).

## §A. Context

- **Tier**: S (doc/doctrine-only; primary single-file edit `moai.md` §4 + template mirror; optional local-only `CLAUDE.local.md` §16 edit; < 300 LOC prose; < 5 files).
- **Primary run-phase target (REQUIRED)**: `.claude/output-styles/moai/moai.md` §4 (Delegation Decision). Add the token-cost signal (Skill = current-context cheap; Agent = isolated-context ~7×) as a 4th weighing consideration + the "prefer Skill injection unless isolation is genuinely needed" directive.
- **Template mirror (REQUIRED)**: `moai.md` IS template-distributed → mirror the §4 edit to `internal/template/templates/.claude/output-styles/moai/moai.md`, run `make build`, verify neutrality (AC-DTC-004). The signal content (generic mechanism cost classification + public paper citation) is acceptable content class.
- **Secondary run-phase target (OPTIONAL, local-only)**: `CLAUDE.local.md` §16 (4-question self-check) — MAY add the same signal for maintainer coherence. **NOT mirrored** to the template tree (CLAUDE.local.md is a forbidden-mirror class) — see §B grounding + AC-DTC-006.
- **PRESERVE**: existing §4 three weighing questions, Forced Delegation Table, Volume Triggers, Allowed Direct Execution. The token-cost signal is ADDITIVE (REQ-DTC-007).
- **Target-pinning correction (from grounding)**: the ROADMAP said "CLAUDE.md §16"; grounding pinned it to `CLAUDE.local.md` §16 (CLAUDE.md §16 is Context Search, unrelated) + `moai.md` §4 as the canonical template-managed surface. Recorded in spec.md §B.2.

## §B. Known Issues (filtered to Tier S relevance)

- **Template neutrality (CLAUDE.local.md §15/§25)**: the `moai.md` §4 mirror must carry NO forbidden internal-content class — the paper citation (arXiv:2604.14228) is acceptable (public source); but NO internal SPEC ID / REQ token / commit SHA / internal date / `CLAUDE.local.md` reference may appear in the mirrored §4 prose. AC-DTC-004 gates this.
- **CLAUDE.local.md is never mirrored**: if the run-phase makes the optional §16 edit, it MUST NOT be copied to the template tree. CLAUDE.local.md references are themselves a forbidden content class (C5). AC-DTC-006 verifies the §16 edit has no template-tree counterpart.
- **B6 — spec-lint Out of Scope heading**: spec.md §F uses `### Out of Scope — <topic>` H3 sub-headings — satisfies `OutOfScopeRule`. (Plan-phase concern, handled.)
- **B10 — scope discipline**: touch ONLY `moai.md` §4 (+ template mirror) + optionally `CLAUDE.local.md` §16. Do NOT touch the N2 ladder surface (`agent-authoring.md`), hook scripts, skill bodies, plugins, or MCP config.

## §C. Pre-flight (run-phase)

```bash
# 1. Confirm §4 structure still matches plan grounding (3 questions + table + triggers)
sed -n '/## 4. Delegation Decision/,/## 5. Checkpoint/p' .claude/output-styles/moai/moai.md
# 2. Confirm template mirror exists
ls internal/template/templates/.claude/output-styles/moai/moai.md
# 3. Confirm CLAUDE.local.md §16 still has the 4 self-check questions
sed -n '/## 16. 오케스트레이터 자가 점검/,/## 17\./p' CLAUDE.local.md
```

## §D. Constraints (DO NOT VIOLATE)

- Doc-only. No Go code change, no hook/skill/plugin/MCP behavior change.
- The ~7× figure is the paper's claim — attribute it, never present as moai-tree measurement (REQ-DTC-006).
- `moai.md` §4 edit MUST be mirrored + `make build` + neutrality verified (REQ-DTC-004).
- `CLAUDE.local.md` §16 edit (if made) MUST NOT be mirrored (REQ-DTC-005).
- Do NOT touch the N2 surface `agent-authoring.md` (REQ-DTC-008).
- Conventional Commits; `🗿 MoAI` trailer; NO `--no-verify` / `--amend` / force-push.

## §E. Self-Verification (run-phase deliverable)

Run-phase manager-develop reports the AC-DTC-001..007 PASS/FAIL matrix with verbatim grep output, plus:
- `git show --stat <run-commit>` confirming the changed files: `moai.md` + its template mirror (+ optional `CLAUDE.local.md` §16, no template counterpart), and that `agent-authoring.md` is NOT among them (AC-DTC-007).
- `go test ./internal/template/... -run TestTemplateNeutralityAudit` PASS (AC-DTC-004).

## §F. Milestones (priority-ordered, no time estimates)

- **M1 — Add token-cost signal to `moai.md` §4**: a 4th weighing consideration / decision note carrying the Skill-injects-current-context vs Agent-spawns-isolated-context (~7×) asymmetry + the "prefer Skill injection unless isolation is genuinely needed" directive + paper attribution. Satisfies REQ-DTC-001..003, REQ-DTC-006, REQ-DTC-007.
- **M2 — Mirror + build**: copy the §4 edit into the template tree, `make build`, verify `TestTemplateNeutralityAudit` (AC-DTC-004).
- **M3 — Optional `CLAUDE.local.md` §16 edit** (REQ-DTC-005): add the same signal to the local-only 4-question self-check for maintainer coherence. NOT mirrored. Skip if `moai.md` §4 suffices.
- **M4 — Self-verify + commit + push**: run the AC grep matrix, commit (`docs(SPEC-DIVECC-DELEGATION-TOKEN-COST-001): M1 delegation token-cost signal`), push to main (Hybrid Trunk Tier S).

## §G. Anti-Patterns to avoid

- Presenting ~7× as moai-measured (violates REQ-DTC-006 + verification-claim-integrity).
- Mirroring a `CLAUDE.local.md` edit into the template tree (violates template isolation — CLAUDE.local.md is forbidden-mirror class).
- Changing the meaning of the existing quality/independence/bias weighing (violates REQ-DTC-007 — additive only).
- Editing the N2 surface `agent-authoring.md` (out of scope — N2 owns it).

## §H. Cross-References

- spec.md §B.2 (pinned targets + ROADMAP correction), §C (REQ-DTC-001..008), §D (AC matrix), §G (cross-references).
- SPEC-DIVECC-EXTENSION-COST-LADDER-001 (N2) — thematic pair.
- `.moai/docs/template-internal-isolation-doctrine.md` — template neutrality gate for M2 + the CLAUDE.local.md forbidden-mirror rule.
