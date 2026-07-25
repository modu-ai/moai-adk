---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: plan
module: doctrine
lifecycle: spec-anchored
tags: [goal, doctrine, session-handoff, slash-command, template-mirror]
tier: L
---

## §A Context

Doctrine and documentation refactor. Three user-approved work items (W1 emission-surface unification, W2 goal-presentation relocation, W3 wrapper creation) executed as ONE Tier L SPEC because W1 and W2 both edit `session-handoff.md`, `goal-directive.md`, and `run.md` — splitting them would make two SPECs contend for the same files (decision D2).

Workspace: isolated worktree, branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, based at `origin/main` = `e306e21a9` (decision D1 — a branch switch in the main checkout is blocked by 59 dirty tracked files that differ from `origin/main`).

### §A.1 Scope correction against the delegation brief

The brief measured 13 local + 13 template = 26 files. A `.claude/`-scoped `grep -rl` missed one pair: the root `CLAUDE.md` and its mirror `internal/template/templates/CLAUDE.md`, both carrying one native-`/goal` occurrence on line 41 (the §2 stage-⑤ clause `when a goal is armed (\`/goal\`, \`/moai goal\`)`).

Corrected scope: **28 existing files** (14 local + 14 template) + **2 new** wrapper files = **30 paths**.

The `CLAUDE.md` pair is NOT byte-identical (intentional — the template copy is neutralized per CLAUDE.local.md §25), so it is excluded from the byte-parity loop and verified by the specific edited clause instead (REQ-GSU-018).

### §A.2 Four retained native-`/goal` surfaces (guard against over-deletion)

Deleting every native-`/goal` mention would erase the subject the prohibition governs. Exactly **four** surfaces retain their references, for categorically different reasons:

| Surface | Retained content | Why | Guard |
|---|---|---|---|
| `goal-directive.md` § Native `/goal` Prohibition | The prohibition rationale: native `/goal` is HUMAN-ONLY, the model cannot invoke it, so the pipeline does not emit it | Edit kind 3 — the prohibition needs its subject | AC-GSU-003 |
| `native-invocation-model.md` § Classification Matrix + Axis B | The `/goal` HUMAN-ONLY classification row and the Axis B worked illustration | A factual statement about Claude Code, and the *justification* for `/moai goal` existing. Retiring the emission path does not make it false | AC-GSU-013 |
| `docs-site/content/*/claude-code/**` (28 pages) | Documentation of Claude Code's own `/goal` feature | Same logic one layer out: these describe the upstream product, not MoAI's pipeline | AC-GSU-029 |
| `internal/goal/evaluate.go` | `NativeGoalActive` field, the step-4 yield branch, the verdict reason string | Implements interoperation *with* native `/goal` — `stop-goal` yields so the two evaluators do not double-block. This is a safety invariant, not an emission | AC-GSU-027 |

This is why `native-invocation-model.md` is owned by M3 (retention semantics) and not M4 (mechanical emission sweep) — its edit adds a cross-reference, it does not swap tokens.

**Correction to the D4 brief.** The delegation brief framed the non-M7 Go references as "the `/moai goal` implementation, not the native command" and said not to touch them. That framing is right about the outcome but wrong about the reason for four of them: `internal/goal/evaluate.go` lines 74, 135-137 are references *to the native command*, implementing the yield invariant. They are retained because they are a retention surface, not because they are implementation identifiers — which matters, because an AC written as "zero native-`/goal` in non-test Go" would force their deletion and silently remove the no-double-block guarantee. AC-GSU-027 pins them explicitly.

---

## §B Known Issues / Risks

- **B1 — Mirror parity is a hard CI gate.** `internal/template/rule_template_mirror_test.go` asserts byte-identity for SSOT mirrors. Both sides of each pair must be edited in the same commit, or CI fails.
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
- **D3 [HARD]** Exactly the two surfaces in §A.2 retain native-`/goal` references. Every other occurrence becomes `/moai goal` or is removed with its structure.
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
6. **Close B6**: state in the Block 5 spec that `/moai goal` is arm-only and therefore Block 5's single primary action stays `/moai run`; cross-reference `workflows/goal.md` § Progression Mode (REQ-GSU-008, -009).

ACs: AC-GSU-006 .. AC-GSU-011.

### M3 — `workflows/goal.md` + `native-invocation-model.md`: W2 wiring and retention semantics

**Priority High.** Completes the bidirectional linkage M2 opened, and guards the two retained surfaces against over-deletion.

Owns 2 files:
- `.claude/skills/moai/workflows/goal.md`
- `.claude/rules/moai/workflow/native-invocation-model.md`

Work:
1. `workflows/goal.md`: repoint the § Cross-references entry off the removed Post-Paste section onto `session-handoff.md` § Canonical Format Block 5; add a back-reference from § Progression Mode to that Block 5 rule so the W2 relationship is discoverable from either side; confirm the § Verbs surface matches what M5's `argument-hint` will claim (REQ-GSU-015).
2. `native-invocation-model.md`: retain the `/goal` HUMAN-ONLY classification row and the Axis B illustration verbatim; add a cross-reference to `goal-directive.md` § Native `/goal` Prohibition so a reader arriving at the classification learns the pipeline does not emit it.

ACs: AC-GSU-012 .. AC-GSU-014.

### M4 — Emission-path sweep

**Priority Medium.** Mechanical once M1-M3 fix the doctrine shape and heading names.

Owns 8 files:
- `.claude/skills/moai/workflows/run.md`
- `.claude/skills/moai/workflows/harness-builder.md`
- `.claude/skills/moai/workflows/harness-build-entry.md`
- `.claude/skills/moai/workflows/moai.md`
- `.claude/skills/moai/SKILL.md`
- `.claude/rules/moai/workflow/orchestration-mode-selection.md`
- `.claude/rules/moai/workflow/dynamic-workflows.md`
- `CLAUDE.md` (root, §2 stage ⑤)

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

Owns 14 existing mirrors:
- `internal/template/templates/.claude/rules/moai/workflow/{goal-directive,session-handoff,session-handoff-examples,native-invocation-model,orchestration-mode-selection,dynamic-workflows}.md`
- `internal/template/templates/.claude/skills/moai/{SKILL.md,workflows/goal.md,workflows/run.md,workflows/moai.md,workflows/harness-builder.md,workflows/harness-build-entry.md}`
- `internal/template/templates/.claude/output-styles/moai/moai.md`
- `internal/template/templates/CLAUDE.md`

Work: re-sync each mirror so the 13 `.claude/**` pairs return to byte-identity, and the `CLAUDE.md` pair carries the identical edited stage-⑤ clause without asserting whole-file identity. Copy neutral text only (D6).

ACs: AC-GSU-019 .. AC-GSU-021.

### M7 — Go emission paths (`cycle_type: tdd`)

**Priority High.** Runs after M6 so the doctrine is settled, but it is the milestone with the highest user impact: without it, the doctrine would declare that the pipeline never emits native `/goal` while the Go auto-resume injector keeps printing `• /goal <cond> ← enter this line to restore autonomous continuation` in four languages. The retirement would be inert exactly at the point of user contact.

**[HARD] M7 is the only milestone with `cycle_type: tdd`.** M1-M6 are documentation edits with no test cycle. M7 changes Go behaviour in a file that has **no test** (`ls internal/hook/handoff_inject_render*_test.go` → no matches), so it follows RED-GREEN-REFACTOR: author the four-locale renderer assertion, observe it fail, then change the renderer. The run phase must route M7 accordingly and must not fold it into the documentation cycle.

Owns 4 files, 8 literals:

| File | Literals | Change |
|---|---:|---|
| `internal/hook/handoff_inject_render.go` | 4 (lines 88, 99, 110, 121) | `goalPrefix: "  • /goal "` → `/moai goal` in each of the ko / ja / zh / en blocks of `handoffLocaleStrings()`. The paired `goalSuffix` restoration-guidance wording is updated with it |
| `internal/harness/v4manifest/schema.go` | 1 (line 15) | `PrimitiveGoal = "/goal"` → `"/moai goal"` |
| `internal/harness/v4manifest/runner_template.go` | 2 (lines 13, 87) | doc comment + `case "/goal":` dispatch arm |
| `internal/cli/handoff.go` | 1 (line 104) | flag **help string** only |

Three constraints:

1. **Renderer test first (REQ-GSU-021).** The test asserts rendered output for all four locales and must fail before the change (AC-GSU-023).
2. **`PrimitiveGoal` occupancy precondition (REQ-GSU-022).** Measured now: **0** — no harness manifest or workflow script declares the `/goal` primitive (`grep -rl '"/goal"' .claude/commands/harness/ .claude/workflows/` → 0 files, across 2 manifests and 5 workflow scripts). A hard rename is therefore safe and **back-compat is not required**. The run phase re-verifies before changing, because occupancy could change between plan and run; AC-GSU-024 pins it at 0 as part of a compound.
3. **`--goal` flag NAME is a CLI contract (REQ-GSU-023).** `moai handoff save --goal` is invoked from the session-handoff doctrine itself. `StringVar(&goal, "goal", …)` keeps its `"goal"` name; only the help string changes. AC-GSU-026 asserts both halves so a rename cannot pass.

ACs: AC-GSU-022 .. AC-GSU-027.

### §F.1 Ownership map (no file owned twice)

| Milestone | Phase | Cycle | Local doctrine | Template | Go | New |
|---|---|---|---:|---:|---:|---:|
| M1 | run | docs | 1 | 0 | 0 | 0 |
| M2 | run | docs | 3 | 0 | 0 | 0 |
| M3 | run | docs | 2 | 0 | 0 | 0 |
| M4 | run | docs | 8 | 0 | 0 | 0 |
| M5 | run | docs | 0 | 0 | 0 | 2 |
| M6 | run | docs | 0 | 14 | 0 | 0 |
| M7 | run | **tdd** | 0 | 0 | 4 | 1 (test) |
| Sync set | **sync** | docs | 0 | 0 | 0 | 0 |
| **Total** | | | **14** | **14** | **4** | **3** |

Run-phase paths: 14 + 14 + 4 + 3 = **35**. Plus the 13 sync-phase paths (§F.2) = **48 paths**. Every path appears in exactly one row.

### §F.2 Sync-phase scope (owner `manager-docs`, NOT a run-phase milestone)

Per approved decision D5, the public and internal documentation surface is updated in the **sync phase**, not the run phase. Owner: `manager-docs`. 13 files:

| Group | Files | Occurrences | Why affected |
|---|---:|---:|---|
| `docs-site/content/{en,ja,ko,zh}/advanced/autonomous-loops.md` | 4 | 52 | The public render surface of the doctrine M1 rewrites — carries a dedicated native-`/goal` section plus the three-primitive comparison table |
| `docs-site/content/{en,ja,ko,zh}/cli-reference/handoff.md` | 4 | 4 | The `--goal` row mirrors the Go help string M7 changes |
| `docs-site/content/{en,ja,ko,zh}/advanced/self-evolving.md` | 4 | 8 | Names `/goal` as a MoAI convergence primitive whose trajectories the routing ledger records (lines 40, 97) |
| `.moai/docs/autonomous-workflow-strategy.md` | 1 | 25 | Internal strategy doc built on native `/goal` as one of three engines |

**Correction to the D5 brief: 13 files, not 9.** The brief classified `advanced/self-evolving.md` (×4) as retain, in the "MoAI-surface factual-contrast" group. Sampling all four locales shows both of its references are MoAI-surface *primitive naming* — `` `/moai loop` / `/goal` convergence trajectories `` — not a factual statement contrasting the two commands. Under W1 those read `/moai goal`. The genuinely factual-contrast pages (`cli-reference/goal.md`, `utility-commands/moai-goal.md`) are retained as the brief specified; verified by sampling `ko/cli-reference/goal.md:9` and `ko/utility-commands/moai-goal.md:14`.

Sync-phase retention set — do NOT modify (AC-GSU-029 pins each count):

| Group | Files | Occurrences | Why retained |
|---|---:|---:|---|
| `docs-site/content/*/claude-code/**` | 28 | 80 | Documents Claude Code's own feature; statements stay factually true |
| `docs-site/content/*/cli-reference/goal.md` | 4 | 12 | States that `/moai goal` is the programmatic counterpart of the HUMAN-ONLY native command — the justification for `/moai goal` existing |
| `docs-site/content/*/utility-commands/moai-goal.md` | 4 | 4 | Same factual contrast |
| `docs-site/content/*/advanced/hooks-reference.md` | 2 (en, ko) | 2 | Stop-hook table dual mention `` `/goal`/`/moai goal` ``, consistent with the retained yield invariant |
| `.moai/research/*.md` | 3 | 4 | Historical research archives; retroactively editing a past record is inappropriate |

**Locale-asymmetry correction to the D5 brief.** The brief stated `advanced/hooks-reference.md` "exists only in `en` and `ko`" and instructed the sync phase not to create `ja`/`zh` pages. The page **exists in all four locales** (`ls docs-site/content/*/advanced/hooks-reference.md` → 4 files). What differs is which locales *carry the reference*: `en:170` and `ko:170` have the Stop-hook row mentioning `/goal`; the `ja` and `zh` copies have no such line. So there is no page to create — the asymmetry is a pre-existing four-locale **content** gap inside existing pages. Per REQ-GSU-027 the sync phase flags it and does not close it; the instruction's intent (do not manufacture symmetry) holds, its stated premise does not.

---

## §G Anti-Patterns

- **Blind find-and-replace of `/goal` → `/moai goal`.** This is three distinct edit kinds (emission swap, structural removal, rationale retention). A global replace would corrupt the prohibition section and the classification matrix, and would leave the obsolete two-step mechanism describing a problem that no longer exists.
- **Deleting every native-`/goal` mention.** Erases the subject the prohibition governs and falsifies nothing — the classification remains true.
- **Editing a local file without its mirror in the same commit.** Breaks the B1 CI gate.
- **Adding the wrapper without `make build`.** The `.md.tmpl` stays outside the embedded FS; AC-GSU-016 fails.
- **Leaving a stated self-check count stale after removing an item.** A silent inconsistency (B5).
- **Putting a bare `/moai goal` line in Block 5 as the single primary action.** Arm-only; the session would spin idle turns to the ceiling.

---

## §H Cross-References

- `spec.md` §B — REQ-GSU-001..018.
- `acceptance.md` — 21 ACs with recorded baselines.
- `design.md` — doctrine-surface boundary and SSOT/render-surface parity.
- `research.md` — baseline commands and observed outputs.
- `CLAUDE.local.md` §2 / §25 / §27.3.
