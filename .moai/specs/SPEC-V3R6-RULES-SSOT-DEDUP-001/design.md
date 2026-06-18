# Design — SPEC-V3R6-RULES-SSOT-DEDUP-001

## §1. Design Goal

Decide, for each duplicated doctrine block, the single SSOT owner and the pointer form each
non-owner adopts; design the zone-registry reduction that does NOT break the `moai constitution`
CLI; and fix the three contradictions (timeout, alwaysLoad, max_teammates) to a single canonical
value before any de-dup edit relocates them.

## §2. SSOT-Owner Decision Table

Selection principle: the owner is the file whose **primary purpose** IS the block (the file the
audit already named as the canonical surface, or the most-cross-referenced surface). Every other
occurrence becomes a pointer carrying only its locally-unique delta.

| # | Duplicated block | SSOT owner (decision) | Non-owners → pointer | Owner-selection rationale |
|---|------------------|-----------------------|----------------------|---------------------------|
| T1 | AskUserQuestion enforcement (preload, Socratic, options, boundary) | `core/askuser-protocol.md` | agent-common-protocol §User Interaction Boundary (keep subagent-prohibition delta); moai-constitution §MoAI Orchestrator (keep orchestrator-obligation delta); CLAUDE.md §8 = forward-item, NOT edited | askuser-protocol.md is the file whose entire purpose is the AskUserQuestion contract; both copies already say "canonical reference: askuser-protocol.md" |
| T2 | Hook JSON config + path-quoting + timeout table | `core/hooks-system.md` | settings-management §Hooks Configuration (keep StatusLine-no-env-var delta) | hooks-system.md is the dedicated hook-event SSOT; settings-management is a settings file that happens to restate hooks |
| T3 | Context-window per-model threshold table (1M=50%, 200K=90%) | `workflow/context-window-management.md` | settings-management; zone-registry CONST-V3R5-022; skill-ab-testing | context-window-management.md is "the authoritative SSOT for the numeric thresholds"; session-handoff already defers correctly |
| T4 | 7-command verification batch block | `core/agent-common-protocol.md` §Parallel Execution | verification-batch-pattern.md (keep grouping taxonomy + re-sync sentinel) | the HARD batching obligation + the verbatim 7-cmd example live in agent-common-protocol; verification-batch-pattern owns only the *rationale/taxonomy* |
| T5 | Status Transition Ownership Matrix (7 rows + Authored-By-Agent) | `development/spec-frontmatter-schema.md` | lifecycle-sync-gate.md (keep close-subject-full-ID one-liner) | spec-frontmatter-schema is the schema SSOT both files already cite as "§ Status Transition Ownership Matrix" |
| T6 | Team mechanics (Role Matrix, Mailbox v2, spawn-wrapper) + 5+1+1 composition | `team-protocol.md` (mechanics) + `team-pattern-cookbook.md` (compositions) | worktree-integration §Team Protocol → pointer; agent-teams-pattern.md → folded into cookbook + DELETED | two-owner split (mechanics vs patterns) is the natural seam; the 4th + 3rd copies are redundant |
| T7 | Skill frontmatter/schema table | `development/skill-authoring.md` | skill-writing-craft (schema table → pointer); skill-ab-testing (frontmatter/PD checklist → pointer) | skill-authoring is the schema SSOT; writing-craft owns prose, ab-testing owns A/B method |
| T8 | Per-domain tool-whitelist + escalation/transition matrices | `development/agent-patterns.md` (tool-whitelist) | agent-authoring (already points here — preserve); orchestrator-templates (residual only) | agent-patterns.md holds the per-domain whitelist table; agent-authoring.md L302 already references it |

## §3. Pointer Form (canonical)

Every pointer follows: `> Canonical: see \`<owner-file>\` § <Anchor> for <block name>. This file owns only <local delta or nothing>.`
This satisfies AC-SSD-012 (names owner file + anchor; one-hop resolution).

## §4. Per-file Mirror Class (drives AC choice — from B3)

[CORRECTED per plan-auditor D1/D2 — the prior version mislabeled 14 non-enrolled files as
"byte-parity / `TestRuleTemplateMirrorDrift`".] Live ground truth: `TestRuleTemplateMirrorDrift`
enrolls ONLY the `workflowOptMirroredPaths` allowlist; of the 20 in-scope files, **only
`hooks-system.md` is enrolled**. Asserting that test for a non-enrolled file is VACUOUSLY green.
zone-registry is NOT byte-identical today (13-line CONST-V3R6-001 delta). The four classes:

| Class | File(s) | AC verification (the ONLY valid clause for this class) |
|-------|---------|--------------------------------------------------------|
| **byte-parity allowlist** (test-enrolled — drift genuinely caught) | `hooks-system.md` ONLY | `go test ... -run TestRuleTemplateMirrorDrift` PASS |
| **byte-identical-by-discipline** (NOT enrolled — identical today, MUST stay identical) | moai-constitution, settings-management, context-window-management, spec-frontmatter-schema, team-protocol, team-pattern-cookbook, worktree-integration, skill-authoring, skill-writing-craft, skill-ab-testing, agent-patterns, agent-authoring, orchestrator-templates | explicit per-file `diff -q <deployed> <mirror>` exit 0 (NOT `TestRuleTemplateMirrorDrift` — that is vacuous for these) |
| **§25-sanitized** (NOT byte-identical — leak-stripped) | agent-common-protocol (17 tok), verification-batch-pattern (2 tok), agent-teams-pattern (1 tok → DELETED) | `TestTemplateNoInternalContentLeak` PASS + scoped `diff` (only the de-dup edit changed); deleted file → absence in both trees |
| **§25-divergent registry** (NOT enrolled, NOT in leak-comment list, but DIVERGES 13 lines) | zone-registry | `TestTemplateNoInternalContentLeak` PASS + `diff` confirming ONLY the de-dup edit changed and the pre-existing 13-line delta is unchanged (NOT `diff -q` exit-0 — that FALSE-FAILS on the legitimate delta) |
| **no mirror** (deployed-only) | lifecycle-sync-gate | `test ! -e <mirror-path>`; no mirror edit |

Run-phase MUST re-confirm each file's allowlist membership before choosing the AC, since the
allowlist (`workflowOptMirroredPaths` + `lateBranchMirroredPaths`) is the ground truth. The
acceptance.md "Mirror-class verification convention" header carries this same corrected table as the
binding AC source.

## §5. File-Deletion Design (T6 / agent-teams-pattern.md)

Deletion is the highest-blast-radius operation. Procedure (4 steps, ordered):
1. **Fold** the 5+1+1 composition into `team-pattern-cookbook.md` as a 6th pattern (deployed +
   mirror) — including the "When to Spawn", Composition Reference table, File Ownership Map, Spawn
   Order. Strip any internal SPEC-ID (`SPEC-V3R5-WORKFLOW-OPT-001`) per §25 when folding into the
   mirror.
2. **Re-point referrers**: grep `agent-teams-pattern` repo-wide (`.claude/`, `internal/template/templates/`,
   plus the Go-test comment at `rule_template_mirror_test.go:66`). Rewrite each prose referrer to the
   cookbook 6th pattern; for the Go-test comment, update the comment text (the file is being removed
   from the §25-leak universe by deletion — confirm the test does not iterate over it as a required path).
3. **Delete** both `.claude/rules/moai/workflow/agent-teams-pattern.md` and its template mirror.
4. **Verify** absence + no dangling referrer (AC-SSD-008).

## §6. Contradiction Reconciliations (resolve BEFORE relocating)

### §6.1 PostToolUse timeout (5s vs 10s) — TWO surfaces in hooks-system.md must agree (D7)
- **Finding**: hooks-system.md is **internally inconsistent** — its JSON example (L244) shows
  PostToolUse `timeout: 10` + `async: true`, but its timeout TABLE (L323) groups
  `SessionStart, PreCompact, PreToolUse, PostToolUse` together under a **5s** default (and the next
  table row lists Stop under 10s while the JSON example shows Stop at 5s). settings-management.md
  (L211) says "10s + async (was 60s before v2.16.0)"; CLAUDE.local.md §7 says MoAI policy default 5s.
- **Reconciliation (canonical)**: these are NOT contradictory once framed — **5s is the default for
  synchronous blocking hooks** (SessionStart, PreToolUse, etc.); **PostToolUse is the documented
  exception at 10s + `async: true`** because its LSP/AST/MX validations run in the background (the
  10s is a per-run background ceiling, not a blocking wait). The edit MUST touch **BOTH surfaces in
  hooks-system.md**: (1) the L323 timeout TABLE — move PostToolUse out of the 5s synchronous row into
  a 10s+async exception row, so the table stops contradicting its own JSON example; (2) the JSON
  example is already correct (10s+async) and stays. settings-management.md adopts the same single
  statement via its pointer. No numeric value changes; the framing is unified across all three
  surfaces (hooks-system table + hooks-system JSON + settings-management pointer) so none reads as a
  bare contradiction. (Editing only the JSON example and leaving the L323 table is the D7 defect —
  the table is the surface a reader consults for the policy default.)

### §6.2 alwaysLoad version (v2.1.119 vs v2.1.121)
- **Finding**: settings-management.md presents both `v2.1.119+` (line 38) and `v2.1.121+` (line 89)
  as the version where `alwaysLoad` was added.
- **Reconciliation (canonical)**: pick **one** introduction version as the canonical "added in"
  value and state the other (if it refers to a later *extension* of the field, e.g. the v2.1.121
  PostToolUse `updatedToolOutput` extension) as a distinct, separately-labeled change — NOT a second
  "introduction" version. Run-phase reads the surrounding text to determine which is the true
  introduction vs a later extension, then states the introduction version once. If both genuinely
  claim introduction, the EARLIER (`v2.1.119`) is canonical for `alwaysLoad` introduction.

### §6.3 max_teammates default 10 vs 3-5 ceiling
- **Finding**: team-protocol.md L114 rejects teams exceeding `max_teammates` (default 10); the 3-5
  ceiling (Anthropic "start with 3-5 teammates") is stated everywhere else.
- **Reconciliation (canonical)**: these operate at two layers — `max_teammates=10` is the **hard
  mechanical reject ceiling** (a config safety limit), while **3-5 is the recommended starting team
  size** (a doctrine guideline). Reconcile by stating both explicitly in team-protocol.md: "the
  recommended starting size is 3-5 teammates (Anthropic guidance); `max_teammates` (default 10) is
  the hard mechanical reject ceiling, not a target." This removes the apparent contradiction without
  changing the config default.

## §7. zone-registry Reduction Design (REQ-SSD-009, CLI-gated)

Driven entirely by the B4 / research.md finding. Two independent sub-reductions:

### §7.1 `clause:` bloat (CONSTRAINED — verbatim-substring requirement)
- The `validate` drift check (`validator.go:264`) requires `normalizeWhitespace(clause)` to be a
  substring of `normalizeWhitespace(stripCodeFences(source))`. A 3-5 word non-verbatim label FAILS
  this for every reduced entry.
- **Design**: reduce each `clause:` to a **still-verbatim shorter excerpt** of the source HARD
  clause (e.g. take the first sentence / the operative MUST-clause rather than the full paragraph).
  This shrinks bloat while keeping `validate` green. Measure the line-count delta as evidence.
- **If** verbatim-shorter-excerpt does not meaningfully reduce the 1019-line bloat (because many
  clauses are already short), the **full structural reduction to {id, zone, file, anchor, label}
  label-only entries is BLOCKED** by the CLI dependency and is deferred to follow-up SPEC
  `SPEC-V3R6-CONST-VALIDATE-LABEL-001` (make `validator.go` resolve clause text from the source
  anchor instead of requiring it inline, OR make the drift check skip label-only entries). This SPEC
  records the blocker note rather than forcing an unsafe change.

### §7.2 `paths:` over-broad trigger (UNCONSTRAINED — safe to narrow)
- Current: `paths: ".claude/**,.moai/specs/**,.claude/rules/**"` — loads the 1019-line registry
  (~13K tokens) on every `.moai/specs/**` edit (i.e., every SPEC authoring turn).
- **Design**: the registry's purpose is HARD-clause codification of the *rules tree*; it does NOT
  need to load on routine SPEC-body edits under `.moai/specs/**`. Narrow `paths:` to
  `.claude/rules/**,.claude/agents/**` (the surfaces whose HARD clauses it codifies), dropping
  `.moai/specs/**` and the over-broad `.claude/**`. Run-phase MUST verify no rule that genuinely
  needs the registry on a `.moai/specs/**` edit relies on this (grep for cross-references that fire
  during SPEC authoring); if one exists, keep `.moai/specs/**`. Default: narrow, verify, revert the
  narrow if a real dependency surfaces.

### §7.3 Why a separate milestone (M_Z)
zone-registry is the only target where the deployed-↔-source edit can break a CLI command
(`validate`), so it is isolated with its own pre-edit gate (re-confirm B4) and its own
post-edit evidence (`moai constitution validate` clean). Bundling it with a body de-dup milestone
would couple a low-risk pointer edit to a CLI-breaking risk.

## §8. Risk-to-Design Traceability

| Risk (plan §B) | Design mitigation |
|----------------|-------------------|
| RK-1 (validate break) | §7.1 verbatim-excerpt-only + blocker note for full reduction |
| RK-2 (§25 diff fails) | §4 mirror-class table → leak-test ACs |
| RK-3/RK-4 (deletion race + dangling) | §5 4-step deletion procedure + §F dependency |
| RK-5 (delta dropped) | §2 each non-owner keeps named local delta; §3 pointer form |
| RK-6 (no-mirror file) | §4 lifecycle-sync-gate = no mirror edit |
| RK-7 (reconcile changes meaning) | §6 three reconciliations fix the canonical value first |
