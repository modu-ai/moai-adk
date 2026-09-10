---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.3.0
created: 2026-07-25
updated: 2026-07-27
author: manager-spec
priority: HIGH
phase: "v3.1.0"
module: doctrine
lifecycle: spec-anchored
tags: "goal, doctrine, session-handoff, slash-command, template-mirror"
tier: L
---

## §A Context

Doctrine and documentation refactor. Three user-approved work items (W1 emission-surface unification, W2 goal-presentation relocation, W3 wrapper creation) executed as ONE Tier L SPEC because W1 and W2 both edit `session-handoff.md`, `goal-directive.md`, and `run.md` — splitting them would make two SPECs contend for the same files (decision D2).

Workspace: isolated worktree, branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, based at `origin/main` = `e306e21a9` (decision D1 — a branch switch in the main checkout is blocked by 59 dirty tracked files that differ from `origin/main`).

### §A.1 Scope correction against the delegation brief

The brief measured 13 local + 13 template = 26 files. A `.claude/`-scoped `grep -rl` missed one pair: the root `CLAUDE.md` and its mirror `internal/template/templates/CLAUDE.md`, both carrying one native-`/goal` occurrence on line 41 (the §2 stage-⑤ clause `when a goal is armed (\`/goal\`, \`/moai goal\`)`).

Post-split scope: **15 local doctrine + 15 template mirrors + 5 Go + 2 new = 37 run-phase paths** (`plan.md` §F.1 is the arithmetic SSOT). The 13 public-documentation paths moved to `SPEC-GOAL-DOCS-RETIRE-001`.

The `CLAUDE.md` pair is NOT byte-identical (intentional — the template copy is neutralized per CLAUDE.local.md §25), so it is excluded from the byte-parity loop and verified by the specific edited clause instead (REQ-GSU-018).

### §A.2 Retention register — three retained native-`/goal` surfaces (guard against over-deletion)

Deleting every native-`/goal` mention would erase the subject the prohibition governs. This table is the **canonical retention register** for this SPEC, and `spec.md` REQ-GSU-004 binds to it by reference rather than restating it — so the two cannot drift apart (the failure mode that produced audit findings D2 and N1).

The membership test is `design.md` §B.3's: **does the sentence become false when MoAI stops emitting native `/goal`?** If it stays true, it is retention.

| # | Surface | Layer | Retained content | Why | Guard |
|---|---|---|---|---|---|
| 1 | `goal-directive.md` § Native `/goal` Prohibition | doctrine | The prohibition rationale: native `/goal` is HUMAN-ONLY, the model cannot invoke it, so the pipeline does not emit it | A prohibition needs its subject | AC-GSU-003, AC-GSU-002 (section-carved) |
| 2 | `native-invocation-model.md` § Classification Matrix + Axis B | doctrine | The `/goal` HUMAN-ONLY classification row and the Axis B worked illustration | A factual statement about Claude Code, and the justification for `/moai goal` existing | AC-GSU-013 |
| 3 | `internal/goal/evaluate.go` | Go | `NativeGoalActive` field, the step-4 yield branch, the verdict reason string | Implements interoperation *with* native `/goal` — `stop-goal` yields so the two evaluators do not double-block. A safety invariant, not an emission | AC-GSU-027 |

**Count re-derived, not edited (audit finding N1).** Iteration 1 recorded six rows. Three of those were documentation-layer surfaces — `docs-site/content/*/claude-code/**` (28 pages), the `autonomous-loops.md` native sections, and `.moai/docs/autonomous-workflow-strategy.md` — and all three left with the sync-phase scope when it was split to `SPEC-GOAL-DOCS-RETIRE-001`. The count is therefore **6 − 3 = 3**, derived from post-split layer membership rather than by editing the previous number. Derivation:

```bash
# Retention rows whose surface is doctrine-or-Go (i.e. remains in this SPEC):
#   goal-directive.md prohibition      → doctrine  ✓
#   native-invocation-model.md matrix  → doctrine  ✓
#   internal/goal/evaluate.go yield    → Go        ✓
#   docs-site/*/claude-code/**         → docs      → SPEC-GOAL-DOCS-RETIRE-001
#   autonomous-loops.md native sects   → docs      → SPEC-GOAL-DOCS-RETIRE-001
#   .moai/docs/autonomous-workflow-strategy.md → docs → SPEC-GOAL-DOCS-RETIRE-001
# => 3
```

Five surfaces must agree on this number: `spec.md` REQ-GSU-004, this register, `plan.md` §D D3, `design.md` §B.3's heading, and `acceptance.md` §D. Verified in §E.

**Correction to the D4 brief (M7 framing).** The delegation brief framed the non-M7 Go references as "the `/moai goal` implementation, not the native command" and said not to touch them. That framing is right about the outcome but wrong about the reason for row 3: `internal/goal/evaluate.go` lines 74, 135-137 are references *to the native command*. They are retained because they are a retention surface, not because they are implementation identifiers — which matters, because an AC written as "zero native-`/goal` in non-test Go" would force their deletion and silently remove the no-double-block guarantee. AC-GSU-027 pins them explicitly.

---

## §B Known Issues / Risks

- **B1 — Mirror parity is CI-enforced for only 2 of the 15 pairs; AC-GSU-019 is the sole guard for the other 13 (S4).** `internal/template/rule_template_mirror_test.go` enforces byte-identity for five named paths, of which only `workflow/session-handoff.md` is in this SPEC's set; `internal/template/output_styles_audit_test.go` separately byte-parity-enforces `.claude/output-styles/moai`, covering `moai.md`. The remaining 13 pairs — including `goal-directive.md`, every `skills/moai/**` mirror, `moai-meta-harness/SKILL.md`, and `CLAUDE.md` — have **no CI parity gate**. This raises AC-GSU-019's importance rather than lowering it: a mirror left unedited fails no test and ships as silent drift. Both sides of every pair must be edited in the same commit.
- **B2 — Template neutrality is a hard CI gate.** `.github/workflows/template-neutrality-check.yaml` + `internal/template/internal_content_leak_test.go` reject a SPEC ID, REQ/AC token, internal date, or commit SHA under `internal/template/templates/`. The M6 sweep must copy *neutral* doctrine text, never SPEC-annotated text.
- **B3 — `make build` is required for W3, not optional.** Templates are embedded via `//go:embed all:templates`; a `.md.tmpl` added without `make build` is invisible to `TestCommandsThinPattern`, which walks the embedded FS. This is exactly what AC-GSU-016 detects.
- **B4 — Section-name churn cascades.** M1 renames `goal-directive.md`'s title and adds two new headings that M2/M3/M4 cross-reference. M1 must land first or the later milestones cite headings that do not exist yet.
- **B5 — Self-check item counts are stated inline.** `session-handoff.md` says "10 items" and `moai.md` §8 says "12 items". Removing the `/goal` follow-up item makes them 9 and 11. A stale count is a silent inconsistency (REQ-GSU-007).
- **B6 — `session-handoff.md` carries an undefined relationship.** Its Block 5 spec says the `Run:` line MAY carry `/moai goal "<condition>"` but never defines that directive's relationship to `/moai run`. Left as-is, the arm-only property (which would spin idle turns) stays unstated. Closed in M2, same milestone that owns the file.
- **B7 — Do not use `git add -A`.** The worktree is clean now, but `make build` regenerates catalog hashes; stage by explicit pathspec.

---

## §C Pre-flight

```bash
git rev-parse --short HEAD && git branch --show-current && git status --porcelain | wc -l
go test ./internal/template/ 2>&1 | tail -3
grep -rohF '`/goal' .claude/ | wc -l
```

Expected at entry: `e306e21a9` / `feat/SPEC-GOAL-SURFACE-UNIFY-001` / `0` dirty; `ok` on the template suite; `175` local occurrences.

---

## §D Constraints

- **D1 [HARD]** All writes stay inside the worktree. Nothing is written to the main checkout's `.moai/specs/`.
- **D2 [HARD]** `goal-directive.md` is rewritten **in place** — filename unchanged, so zero cross-reference paths need updating.
- **D3 [HARD]** Exactly the **three** surfaces in §A.2 retain native-`/goal` references. Every other occurrence becomes `/moai goal` or is removed with its structure. Before recording or satisfying any sweep-style AC, test it against **each** of the six — a sweep phrased broadly enough to satisfy itself by deleting a retention surface is rejected.
- **D4 [HARD]** No file has two milestone owners (see §F ownership map).
- **D5 [HARD]** No time estimates. Priority labels and phase ordering only.
- **D6 [HARD]** Template edits carry no SPEC ID, REQ/AC token, internal date, commit SHA, or audit citation.
- **D7** Match each file's existing prose register and heading conventions; this is a refactor of wording and structure, not a rewrite of voice.

---

## §E Self-Verification

Each milestone is verified by its own ACs in `acceptance.md`. Two gates sit outside the AC matrix:

- **Held-in** — every AC transitions from its recorded FAIL baseline to PASS.
- **Held-out** — the existing guard suite still passes: `go test ./internal/template/` (mirror parity, neutrality, command audit) and `go test ./...`. These already pass at baseline, so they are gates, **not ACs** — an AC that already passes is not an AC.

---

## §F Milestones

Ordered by dependency and decision-reversibility: the decisions most likely to change (doctrine shape, structural removal, user-facing flow) lead; the mechanical sweeps and propagation trail. Milestone number equals execution order.

### M1 — `goal-directive.md`: rewrite in place as `/moai goal` doctrine

**Priority High.** The doctrine SSOT. Every other milestone cross-references headings created here, so it lands first (B4).

Owns 1 file:
- `.claude/rules/moai/workflow/goal-directive.md`

Work:
1. Retitle so the doctrine subject is `/moai goal`; rewrite § What It Is, § Writing an Effective Condition, § Proactive Recommendation Triggers (T1-T4), and § MoAI Integration Notes onto the `/moai goal` surface.
2. Add `## Native \`/goal\` Prohibition` — the single retained rationale section (REQ-GSU-004). All residual native-`/goal` references in this file live inside it.
3. Add `### Goal-Presentation Timing` — the W2 codification home: the arm-only property, presentation at the Implementation Kickoff Approval progression-mode axis, the Kickoff-remains-required invariant, and the rejected `/moai goal --run` alternative with its reason (REQ-GSU-008, -010, -011, -012).
4. Repoint the § Guardrails bullet that currently delegates resume-context goal handling to the `session-handoff.md` Post-Paste section (removed in M2) onto the Block 5 rule.
5. **Relocate the comparison table's native row into the prohibition section (S11).** § Comparing Autonomous-Continuation Approaches carries a `| /goal | The previous turn finishes | A fresh model confirms the condition is met |` row. By the §A.2 membership test that row is retention-class — but AC-GSU-002 requires zero native-`/goal` occurrences *outside* the prohibition section. The resolution is relocation, not deletion: the native row moves into `## Native \`/goal\` Prohibition` (where the comparison belongs anyway, since the prohibition's whole point is to distinguish the two commands), and the comparison table keeps `/moai goal`, `/moai loop`, `/loop`, and the Stop-hook rows. This is a deliberate design decision, not an AC-forced deletion.

ACs: AC-GSU-001 .. AC-GSU-005.

### M2 — `session-handoff.md` + examples + render surface: remove the two-step mechanism, close the Block 5 gap

**Priority High.** The largest structural decision. Also closes B6 in the same milestone that owns the file, avoiding a second pass over `session-handoff.md`.

Owns 3 files:
- `.claude/rules/moai/workflow/session-handoff.md`
- `.claude/rules/moai/workflow/session-handoff-examples.md`
- `.claude/output-styles/moai/moai.md` (the §8 render surface)

Work:
1. Delete `## Post-Paste /goal Follow-up Block` from `session-handoff.md` in full, plus the Block 1 "Post-paste `/goal` follow-up" bullet, the § Output Surface item (2) conditional, the § Auto-Memory Integration item-2 clause, and the § Cross-references entry.
2. Delete the `Post-paste /goal instruction line` row from the Localization Table (and its ja/zh columns in `session-handoff-examples.md`).
3. Delete the follow-up-block item from § Pre-emit self-check and update the stated count 10 → 9 (REQ-GSU-007).
4. In `session-handoff-examples.md`: remove the goal-first bootstrap variant's dependence on the removed section, update the Paste-Time Activation Matrix so the class-(d) row no longer routes a MoAI-emitted goal line, and drop the two anti-pattern bullets that describe the retired mechanism.
5. In `moai.md` §8: remove the post-paste follow-up block clause, update the emission-time save clause's `--goal` condition wording, and update the stated self-check count 12 → 11.
6. **Decided, not deferred — the Paste-Time Activation Matrix classification (audit iteration 2 flagged judgment call).** `session-handoff.md:163` reads "(d) user-only TUI commands (`/goal`, `/effort`, `/clear`) fire ONLY as a standalone user message. A `/goal` line is class (d)". By the §A.2 membership test this is a *classification* statement that stays true after retirement — yet AC-GSU-031's target `0` forces the token out. Resolution: **keep the classification, drop the `/goal` token from it.** The class-(d) enumeration keeps `/effort` and `/clear`, and the routing sentence naming a MoAI-emitted goal line is removed. No unique content is lost, because the native-`/goal` classification survives canonically in retention register row 2 (`native-invocation-model.md`) and in row 1's prohibition section. The run phase implements this decision rather than rediscovering it.
7. **Close B6**: state in the Block 5 spec that `/moai goal` is arm-only and therefore Block 5's single primary action stays `/moai run`; cross-reference `workflows/goal.md` § Progression Mode (REQ-GSU-008, -009).

ACs: AC-GSU-006 .. AC-GSU-011.

### M3 — `workflows/goal.md` + `native-invocation-model.md`: W2 wiring and retention semantics

**Priority High.** Completes the bidirectional linkage M2 opened, and guards retention register row 2 against over-deletion.

Owns 2 files:
- `.claude/skills/moai/workflows/goal.md`
- `.claude/rules/moai/workflow/native-invocation-model.md`

Work:
1. `workflows/goal.md`: repoint the § Cross-references entry off the removed Post-Paste section onto `session-handoff.md` § Canonical Format Block 5; add a back-reference from § Progression Mode to that Block 5 rule so the W2 relationship is discoverable from either side; confirm the § Verbs surface matches what M5's `argument-hint` will claim (REQ-GSU-015).
2. `native-invocation-model.md`: retain the `/goal` HUMAN-ONLY classification row and the Axis B illustration verbatim; add a cross-reference to `goal-directive.md` § Native `/goal` Prohibition so a reader arriving at the classification learns the pipeline does not emit it.

ACs: AC-GSU-012 .. AC-GSU-014.

### M4 — Emission-path sweep

**Priority Medium.** Mechanical once M1-M3 fix the doctrine shape and heading names.

Owns 9 files:
- `.claude/skills/moai/workflows/run.md`
- `.claude/skills/moai/workflows/harness-builder.md`
- `.claude/skills/moai/workflows/harness-build-entry.md`
- `.claude/skills/moai/workflows/moai.md`
- `.claude/skills/moai/SKILL.md`
- `.claude/skills/moai-meta-harness/SKILL.md` — **added at audit iteration 1 (D6)**: line 51 documents the `PrimitiveGoal` manifest token M7 renames. `harness-builder.md:107,199` document the same token and were already owned, so the omission was an inconsistency, not a deliberate exclusion. Leaving it would let this SPEC introduce doc-vs-code drift it does not own
- `.claude/rules/moai/workflow/orchestration-mode-selection.md`
- `.claude/rules/moai/workflow/dynamic-workflows.md`
- `CLAUDE.md` (root, §2 stage ⑤)

**Mixed-form lines (S8).** `orchestration-mode-selection.md:18,145,204,205` each carry BOTH a backticked `` `/goal` `` and an unbackticked `(/goal ac_converge)`. Swapping only the backticked token leaves the unbackticked residue. Same shape at `run.md:126` and `harness-builder.md:149`, which `plan.md` already instructs changing. AC-GSU-031's union detector is the guard.

**Graceful-degradation rewrite must be re-derived, not transposed (S10).** `goal-directive.md` states native `/goal` "Requires Claude Code v2.1.139 or later, an accepted workspace trust dialog, and hooks enabled". That v2.1.139 floor is a property of the **native** command's Stop-hook wrapper, NOT of `moai hook stop-goal`. Transposing it would assert a version requirement `/moai goal` does not have. M4 re-derives the condition for `stop-goal` (hooks enabled; no version floor) and MUST NOT copy the version number across.

Work: switch every emission-path reference to `/moai goal`. Notable per-file specifics:
- `run.md` — § Run-phase Autonomy heading, the `ac_converge` set instruction, the graceful-degradation bullet (native runtime-version / hooks-disabled conditions become the `stop-goal` hook's own availability conditions), and the § Cross-references entry.
- `orchestration-mode-selection.md` — §B.1b auto-mode pairing, the §C.3 preferences-drained row, the §C.3 blocker-report subheading, and two § F cross-references.
- `harness-builder.md` — the ACTIVATE-phase convergence instruction and the § Primitive Mapping row.
- `CLAUDE.md` — stage ⑤ becomes `when a goal is armed (\`/moai goal\`)`.

ACs: AC-GSU-015.

### M5 — W3 wrapper creation

**Priority Medium.** Independent of M1-M4; ordered here so the single `make build` runs once, after all local doctrine edits settle.

Owns 2 new files:
- `internal/template/templates/.claude/commands/moai/goal.md.tmpl` (template source — written FIRST per CLAUDE.local.md §2)
- `.claude/commands/moai/goal.md` (local)

Work: author the thin wrapper matching the real shape of the `run.md` / `run.md.tmpl` sibling pair — 4 frontmatter fields (`description`, `argument-hint`, `allowed-tools: Skill`) and a one-line `Use Skill("moai") with arguments: goal $ARGUMENTS` body, under 20 non-empty body lines. The template variant carries the 4-locale `{{if eq .ConversationLanguage ...}}` description form. `argument-hint` names the delivered verbs only. Then run `make build`.

ACs: AC-GSU-016 .. AC-GSU-018.

### M6 — Template mirror sweep

**Priority Medium.** Strictly last: it propagates the settled local text.

Owns 15 existing mirrors:
- `internal/template/templates/.claude/rules/moai/workflow/{goal-directive,session-handoff,session-handoff-examples,native-invocation-model,orchestration-mode-selection,dynamic-workflows}.md`
- `internal/template/templates/.claude/skills/moai/{SKILL.md,workflows/goal.md,workflows/run.md,workflows/moai.md,workflows/harness-builder.md,workflows/harness-build-entry.md}`
- `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` — **added at audit iteration 1 (D6)**; the pair is byte-identical at baseline (`diff -q` → `same`)
- `internal/template/templates/.claude/output-styles/moai/moai.md`
- `internal/template/templates/CLAUDE.md`

Work: re-sync each mirror so the 13 `.claude/**` pairs return to byte-identity, and the `CLAUDE.md` pair carries the identical edited stage-⑤ clause without asserting whole-file identity. Copy neutral text only (D6).

ACs: AC-GSU-019 .. AC-GSU-021.

### M7 — Go emission paths (`cycle_type: tdd`)

**Priority High.** Runs after M6 so the doctrine is settled, but it is the milestone with the highest user impact: without it, the doctrine would declare that the pipeline never emits native `/goal` while the Go auto-resume injector keeps printing `• /goal <cond> ← enter this line to restore autonomous continuation` in four languages. The retirement would be inert exactly at the point of user contact.

**[HARD] M7 is the only milestone with `cycle_type: tdd`.** M1-M6 are documentation edits with no test cycle. M7 changes Go behaviour in a file that has **no test** (`ls internal/hook/handoff_inject_render*_test.go` → no matches), so it follows RED-GREEN-REFACTOR: author the four-locale renderer assertion, observe it fail, then change the renderer. The run phase must route M7 accordingly and must not fold it into the documentation cycle.

Owns 5 files, 8 literals:

| File | Literals | Change |
|---|---:|---|
| `internal/hook/handoff_inject_render.go` | 4 (lines 88, 99, 110, 121) | `goalPrefix: "  • /goal "` → `/moai goal` in each of the ko / ja / zh / en blocks of `handoffLocaleStrings()`. The paired `goalSuffix` restoration-guidance wording is updated with it |
| `internal/harness/v4manifest/schema.go` | 1 (line 15) | `PrimitiveGoal = "/goal"` → `"/moai goal"` |
| `internal/harness/v4manifest/runner_template.go` | 2 (lines 13, 87) | doc comment + `case "/goal":` dispatch arm |
| `internal/cli/handoff.go` | 1 (line 104) | flag **help string** only |
| `internal/harness/v4manifest/runner_template_test.go` | 0 (test fixture) | **added at audit iteration 1 (D3)** — see below |

**[HARD] M7 breaks an existing passing test unless it updates it (D3).** `runner_template_test.go:19` reads `{PrimitiveGoal, \`case "/goal"\`}` and asserts both `strings.Contains(RunnerTemplate, tc.marker)` and `strings.Contains(RunnerTemplate, tc.primitive)`. Verified: the suite passes at baseline (`go test ./internal/harness/v4manifest/` → `ok`). Once M7 sets `PrimitiveGoal = "/moai goal"` and rewrites the dispatch arm, the hard-coded `marker` literal `case "/goal"` is absent and the subtest FAILS — breaking this SPEC's own §C held-out gate. M7 therefore updates the marker to `` `case "/moai goal"` `` in the same change. Note this is invisible to AC-GSU-027, whose third component excludes `_test.go`, so it would read `0` while the suite is red; AC-GSU-030 is the guard.

Four constraints:

1. **Renderer test first (REQ-GSU-021).** The test asserts rendered output for all four locales and must fail before the change (AC-GSU-023).
2. **`PrimitiveGoal` occupancy precondition (REQ-GSU-022).** Measured now: **0** — no harness manifest or workflow script declares the `/goal` primitive (`grep -rl '"/goal"' .claude/commands/harness/ .claude/workflows/` → 0 files, across 2 manifests and 5 workflow scripts). A hard rename is therefore safe and **back-compat is not required**. The run phase re-verifies before changing, because occupancy could change between plan and run; AC-GSU-024 pins it at 0 as part of a compound.
3. **`--goal` flag NAME is a CLI contract (REQ-GSU-023).** `moai handoff save --goal` is invoked from the session-handoff doctrine itself. `StringVar(&goal, "goal", …)` keeps its `"goal"` name; only the help string changes. AC-GSU-026 asserts both halves so a rename cannot pass.
4. **The v4manifest test fixture moves with the token (REQ-GSU-028, D3).** Verified by AC-GSU-030, which pins the package suite at exit 0 *and* the updated marker literal.

ACs: AC-GSU-022 .. AC-GSU-027, AC-GSU-030.

### §F.1 Ownership map (no file owned twice)

| Milestone | Phase | Cycle | Paths | Composition |
|---|---|---|---:|---|
| M1 | run | docs | 1 | 1 local doctrine |
| M2 | run | docs | 3 | 3 local doctrine |
| M3 | run | docs | 2 | 2 local doctrine |
| M4 | run | docs | 9 | 9 local doctrine (incl. `moai-meta-harness/SKILL.md`, D6) |
| M5 | run | docs | 2 | 2 new (1 template source + 1 local) |
| M6 | run | docs | 15 | 15 template mirrors (incl. the `moai-meta-harness` mirror, D6) |
| M7 | run | **tdd** | 5 | 5 Go (4 emission files + `runner_template_test.go`, D3) |
| **Total** | | | **37** | all run-phase |

**Canonical path total: 37** in the milestone-ownership sense (1+3+2+9+2+15+5) — every path in this table appears in exactly one row, and this SPEC has no sync-phase paths (see §F.2). The **actual run-phase path union across M1-M7 is 40**: three paths are touched as a mechanical or requirement-mandated consequence of the table above without being independently enumerated as owned rows, and were not caught by this table at plan time:

| Extra path | Why not enumerated at plan time | Why it is still in scope |
|---|---|---|
| `internal/template/catalog.yaml` | Regenerated mechanically by `make build` (M5's own action), not hand-edited — plan time enumerated the hand-authored `.md.tmpl`, not `make build`'s regeneration side effect | AC-GSU-016 requires the wrapper to be embedded-FS-reachable, which is exactly what `make build` regenerating this file proves |
| `internal/hook/handoff_inject_render_test.go` | The milestone brief listed M7 as "5 Go files" (4 emission files + `runner_template_test.go`) without separately calling out the *new* test file AC-GSU-033 requires M7 to author | AC-GSU-021 / AC-GSU-033 mandate this file's existence as the RED-then-GREEN regression guard |
| `internal/hook/handoff_inject_cover_test.go` | A pre-existing lockstep fixture pinning the retired `"/goal "` token; not itself an edit target of any REQ | M7 recorded this as a scope deviation (§E.2 "M7 scope deviation") — `"/moai goal "` does not contain the substring `"/goal "`, so the fixture's assertion went red and required a one-line pin update |

Derivation of the 37-row column, so the number is computed rather than carried: the doctrine layer is **15 local + 15 template = 30 files** (M1-M4 own the 15 local; M6 owns their 15 mirrors), plus 5 Go (M7) and 2 new (M5). Since audit iteration 1: M4 gained `moai-meta-harness/SKILL.md` (D6) → 9 local and M6 gained its mirror → 15 template; M7 gained `runner_template_test.go` (D3) → 5 Go, counted as an existing file rather than "new". Since audit iteration 2 the 13 sync paths left with the scope reduction, taking the total from 50 to 37. The 3 additional paths above (§H sync-phase note) bring the actual run-phase touch-count to 40; they are enumeration gaps at plan time, not scope creep — none required a new REQ or AC.

### §F.2 Sync-phase scope — MOVED to `SPEC-GOAL-DOCS-RETIRE-001`

The public-documentation scope (13 files) that this section previously carried was split out after plan-audit iteration 2 emitted STOP. It now lives in `SPEC-GOAL-DOCS-RETIRE-001`, which declares `depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]` because its `cli-reference/handoff.md` work mirrors the Go help string that M7 rewrites.

Nothing in this SPEC is sync-phase scoped any longer: M1-M7 are all run-phase. The moved requirement and criterion identifiers are registered at `spec.md` §B.7.

## §G Anti-Patterns

- **Blind find-and-replace of `/goal` → `/moai goal`.** This is three distinct edit kinds (emission swap, structural removal, rationale retention). A global replace would corrupt the prohibition section and the classification matrix, and would leave the obsolete two-step mechanism describing a problem that no longer exists.
- **Deleting every native-`/goal` mention.** Erases the subject the prohibition governs and falsifies nothing — the classification remains true.
- **Editing a local file without its mirror in the same commit.** Breaks the B1 CI gate.
- **Adding the wrapper without `make build`.** The `.md.tmpl` stays outside the embedded FS; AC-GSU-016 fails.
- **Leaving a stated self-check count stale after removing an item.** A silent inconsistency (B5).
- **Putting a bare `/moai goal` line in Block 5 as the single primary action.** Arm-only; the session would spin idle turns to the ceiling.

---

## §H Cross-References

- `spec.md` §B — 25 REQs (REQ-GSU-001..024 + 028).
- `acceptance.md` — 30 ACs (AC-GSU-001..027 + 030, 031, 033) with recorded baselines, plus the REQ↔AC traceability matrix (§F: 25/25 REQs covered).
- `design.md` — doctrine-surface boundary and SSOT/render-surface parity.
- `research.md` — baseline commands and observed outputs.
- `CLAUDE.local.md` §2 / §25 / §27.3.
