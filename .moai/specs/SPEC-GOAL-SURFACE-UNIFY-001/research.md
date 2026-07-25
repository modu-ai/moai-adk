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

> **Scope note.** Thin by design: the baseline measurement plus the two-command distinction. All numbers below are cited from executed commands with observed output, per the no-unobserved-claim invariant.

## §A Baseline Measurement

Executed in the worktree at `origin/main` = `e306e21a9`, branch `feat/SPEC-GOAL-SURFACE-UNIFY-001`, working tree clean.

### §A.1 Local surface

```bash
grep -rlF '`/goal' .claude/ | sort
grep -roF '`/goal' .claude/ | wc -l
```

Observed: **13 files, 175 occurrences.**

| Occurrences | File |
|---:|---|
| 64 | `.claude/rules/moai/workflow/goal-directive.md` |
| 35 | `.claude/rules/moai/workflow/session-handoff.md` |
| 19 | `.claude/rules/moai/workflow/session-handoff-examples.md` |
| 12 | `.claude/skills/moai/workflows/run.md` |
| 9 | `.claude/skills/moai/workflows/harness-builder.md` |
| 9 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` |
| 9 | `.claude/output-styles/moai/moai.md` |
| 6 | `.claude/skills/moai/workflows/goal.md` |
| 5 | `.claude/rules/moai/workflow/native-invocation-model.md` |
| 2 | `.claude/skills/moai/workflows/moai.md` |
| 2 | `.claude/skills/moai/SKILL.md` |
| 2 | `.claude/rules/moai/workflow/dynamic-workflows.md` |
| 1 | `.claude/skills/moai/workflows/harness-build-entry.md` |

### §A.2 Template surface

```bash
grep -rlF '`/goal' internal/template/templates/ | sort
grep -roF '`/goal' internal/template/templates/ | wc -l
```

Observed: **14 files, 176 occurrences** — one file and one occurrence more than the local `.claude/`-scoped result.

### §A.3 The 14th pair (scope correction)

The extra template file is `internal/template/templates/CLAUDE.md`. Its local counterpart is the repository-root `CLAUDE.md`, which a `.claude/`-scoped `grep -rl` cannot reach.

```bash
grep -nF '`/goal' CLAUDE.md
grep -nF '`/goal' internal/template/templates/CLAUDE.md
```

Observed — identical line on both sides:

```
41:- ⑤ **Execute → verify → iterate** — run the plan, verify against acceptance criteria, iterate; when a goal is armed (`/goal`, `/moai goal`), the goal evaluator is the termination judge.
```

**Corrected total scope: 28 existing files** (14 local + 14 template), against the 26 stated in the delegation brief.

### §A.4 Mirror parity

```bash
for f in <the 13 .claude/** files>; do
  t="internal/template/templates/$f"
  diff -q "$f" "$t" >/dev/null 2>&1 && echo "same   $f" || echo "DIFF   $f"
done
diff -q CLAUDE.md internal/template/templates/CLAUDE.md >/dev/null 2>&1 && echo "same   CLAUDE.md" || echo "DIFF   CLAUDE.md"
```

Observed: **13 × `same`, 0 × `DIFF`** for the `.claude/**` pairs, and **`DIFF   CLAUDE.md`** for the 14th.

The `CLAUDE.md` divergence is expected, not a defect: the template copy is neutralized per the template internal-content isolation policy. Byte-identity must therefore be asserted over 13 pairs only — see `design.md` §B.2.

### §A.5 Slash-command wrapper inventory

```bash
ls .claude/commands/moai/*.md | wc -l
ls internal/template/templates/.claude/commands/moai/*.md.tmpl | wc -l
```

Observed: **14** and **14**. Present: clean, codemaps, e2e, feedback, fix, gate, harness, loop, mx, plan, project, review, run, sync. `goal.md` is absent from both trees.

`design` is deliberately wrapper-less — its workflow file exists but it is not registered in the SKILL.md router (it runs via the `manager-design` agent). The missing-wrapper count is therefore exactly **one**.

### §A.6 Sibling wrapper shape (read, not assumed)

`.claude/commands/moai/run.md` — 4 frontmatter fields, single-line body:

```
---
description: Implement SPEC requirements using DDD/TDD methodology
argument-hint: "SPEC-XXX [--team] [--resume SPEC-XXX]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: run $ARGUMENTS
```

The template variant is identical except that `description` carries the 4-locale conditional:

```
description: {{if eq .ConversationLanguage "ko"}}…{{else if eq .ConversationLanguage "ja"}}…{{else if eq .ConversationLanguage "zh"}}…{{else}}…{{end}}
```

Enforcement read from `internal/template/commands_audit_test.go`: `description` and `allowed-tools` required; `allowed-tools` must be a CSV string not a YAML array; body under 20 non-empty lines; body must contain `Skill(` or `subagent`; `tools` / `disallowed-tools` / `model` are rejected as deprecated. The test walks the **embedded** FS, which is why `make build` is load-bearing (see `acceptance.md` AC-GSU-016).

### §A.7 Verb surface (for `argument-hint`)

Read from `.claude/skills/moai/workflows/goal.md` § Verbs: `/moai goal "<condition>"` (register + arm), `status [--all]`, `clear`. A fourth verb `resume` is documented in the same section as **deferred and NOT delivered** — `moai goal --help` lists only `arm` / `status` / `clear`, because `clear` deletes the state file rather than tombstoning it.

Note a pre-existing inconsistency, in scope for M3 to reconcile: `.claude/skills/moai/SKILL.md:157` still advertises four verbs including `resume`. The `argument-hint` authored in M5 must follow `workflows/goal.md` (the delivered surface), not `SKILL.md`.

---

## §B The Two-Command Distinction

The refactor rests entirely on this distinction, which is documented and classified in the existing doctrine rather than inferred.

| | native `/goal` | `/moai goal` |
|---|---|---|
| Owner | Claude Code built-in TUI command | MoAI |
| Invocation model | **HUMAN-ONLY** — no `[Skill]` / `[Workflow]` marker, not exposed through the `Skill` tool | **PROGRAMMATIC** — orchestrator arms it directly |
| Evaluator | Session-scoped prompt-based Stop hook (Haiku by default) | `moai hook stop-goal` Stop-hook evaluator |
| State | Runtime-internal | `.moai/state/goal/<session-id>.json` (per-session) |
| Survives `/clear` | No — `/clear` removes an active goal | State file persists; re-armable programmatically |
| Can the model set it? | **No** | **Yes** |

Source, verbatim from `native-invocation-model.md` § Classification Matrix:

> `/goal` | HUMAN-ONLY | No `[Skill]`/`[Workflow]` marker — commands reference: `/goal [condition|clear] Set a goal`. Not exposed through the `Skill` tool

And its Axis B illustration, which is the standing justification for `/moai goal`:

> `/moai goal` now reimplements the HUMAN-ONLY `/goal` capability programmatically (a per-session condition-declared agentic loop evaluated by the `stop-goal` Stop hook) — it is the worked example that proves the Axis B justification path.

**Implication for scope.** Because the classification is a factual statement about Claude Code, and because it is the justification for `/moai goal` existing, retiring the *emission path* does not falsify it. `native-invocation-model.md` therefore keeps its native-`/goal` references — it is a retention surface, not an emission surface. See `plan.md` §A.2 for the full two-surface retention list and `acceptance.md` AC-GSU-013 for the over-deletion guard.

---

## §C Arm-Only Property (verified, not assumed)

Read from `workflows/goal.md` § Verbs and § Safety Invariants:

- `/moai goal "<condition>"` writes `.moai/state/goal/<session-id>.json` (atomic temp+rename) and the Stop hook picks it up on the next turn-end.
- The evaluator "only decides whether the turn continues; it never pre-approves irreversible actions."
- No verb starts work. Arming records state and blocks turn-end; it does not dispatch a phase.

Consequence: a paste-ready resume whose Block 5 carried only a goal-arming directive would arm a condition against an idle session and burn turns to the ceiling with nothing to converge. This is the mechanical basis for REQ-GSU-009 (Block 5 keeps `/moai run`).

Cross-check against `session-handoff.md` Block 5 as it stands:

> Where the next SPEC declares a machine-verifiable end-state, the `Run:` line MAY carry `/moai goal "<condition>"` … the post-paste native-`/goal` follow-up block is then demoted to an optional variant.

The permission is granted; the relationship to `/moai run` is never defined. That undefined relationship is the gap M2 closes.

---

## §D Progression-Mode Axis Already Exists

Read from `workflows/goal.md`:

- Line 59 — `## Progression Mode (Autonomous / Semi-autonomous) — chosen at Implementation Kickoff Approval`
- Line 80 — the selected mode "is persisted in goal state as `progression_mode` (default `autonomous` when the user declines to choose)"
- § Safety Invariants #1 — "Implementation Kickoff Approval is mandatory in both modes. The progression-mode axis is a post-approval progression CHOICE, not a relaxation of the gate."

The mechanism, the persistence field, the gate placement, and the safety invariant are all already present. **Net new development for W2: zero.** W2 is the codification of a relationship statement, which is why it carries no implementation milestone of its own.

---

## §E Environment Notes

- **`awk` portability.** macOS ships BWK `awk`, which has no 3-arg `match()`. A first draft of the AC-GSU-008 judgment command used `match($0, /…/, m)` and failed with `awk: syntax error at source line 1`. The delivered command uses `grep -o` plus `sed -n` range extraction instead.
- **`grep -c` vs `grep -o | wc -l`.** These differ materially on this corpus: `session-handoff.md` has 12 *lines* but 18 *occurrences* of `post-paste`. All `acceptance.md` counts use the occurrence form.
- **Detector literal.** The open form `` `/goal `` was chosen over `` `/goal` `` after comparing both: the closed form misses `` `/goal ac_converge` `` and `` `/goal`-turn ``. Open-form and closed-form counts are not comparable (38 vs 32 across the M4 file set).
- **URL-path false positives.** In `docs-site/`, an unbackticked `/goal` search matches link paths such as `/en/cli-reference/goal`. A first pass returned 58 files / 32 `claude-code` pages; re-measuring with the backticked detector returned **50 / 28**, which reconciles with the D5 brief. All docs-site counts in §G use the backticked form. `cli-reference/loop.md` (×4) was in the inflated set solely on a URL match and carries no command reference.

---

## §F Go Emission-Path Measurement (D4)

```bash
grep -rn '"/goal\|/goal ' --include='*.go' internal/ cmd/ pkg/ | grep -v '_test.go'
```

### §F.1 The 8 emission literals (confirmed exactly as briefed)

| Location | Count | What it is |
|---|---:|---|
| `internal/hook/handoff_inject_render.go:88,99,110,121` | 4 | `goalPrefix: "  • /goal "` in the ko / ja / zh / en blocks of `handoffLocaleStrings()` |
| `internal/harness/v4manifest/schema.go:15` | 1 | `PrimitiveGoal = "/goal"` |
| `internal/harness/v4manifest/runner_template.go:13,87` | 2 | doc comment + `case "/goal":` dispatch arm |
| `internal/cli/handoff.go:104` | 1 | flag help string |

### §F.2 A fourth retention surface the brief mis-classified

The same sweep also returned three lines the D4 brief placed in the do-not-touch set on the grounds that they are "the `/moai goal` implementation, not the native command":

```
internal/goal/evaluate.go:74:  NativeGoalActive bool // set when the runtime signals an active native /goal
internal/goal/evaluate.go:135: // Step 4 (checked before ceiling so a native /goal always wins): yield.
internal/goal/evaluate.go:137: return Verdict{Yielded: true, Reason: "native /goal active — stop-goal yields (no double-block)"}, false
```

These are references **to the native command**, implementing the yield invariant recorded at `workflows/goal.md` § Safety Invariants #4. The outcome the brief wanted (do not touch) is right; the stated reason is not, and the distinction is load-bearing: an AC phrased "zero native-`/goal` in non-test Go" would force their deletion and silently remove the no-double-block guarantee. Pinned by AC-GSU-027.

Observed pins: `grep -coF 'NativeGoalActive' internal/goal/evaluate.go` → `2`; `grep -coF 'native /goal' internal/goal/evaluate.go` → `3`.

### §F.3 The "17 implementation references" figure is not reproducible

The brief cited 17 other Go references. That number could not be reproduced from any pattern tried; the nearest attempt —

```bash
grep -rn 'internal/goal\|\.moai/state/goal\|stop-goal\|moai goal' --include='*.go' internal/ cmd/ | grep -v '_test.go' | wc -l
```

— returned **35** non-test lines. Because the figure is pattern-dependent and unattributable, no AC is built on it. The control AC (AC-GSU-027) instead pins two *specific, reproducible* symbols in the one file where over-deletion would cause harm, which is the guarantee the brief actually asked for.

### §F.4 `PrimitiveGoal` occupancy precondition — measured 0

```bash
ls .claude/commands/harness/**/manifest.json .claude/workflows/*.js
grep -rl '"/goal"' .claude/commands/harness/ .claude/workflows/
```

Observed: 2 manifests (`oss-docs`, `release-update`) and 5 workflow scripts exist; **zero** declare the `/goal` primitive. Symbol usage of `PrimitiveGoal` is confined to `schema.go:15,25` and a `types.go:90` doc comment, plus two test files.

**Consequence: a hard rename is safe; back-compat accepting both tokens is not required.** REQ-GSU-022 still requires the run phase to re-verify, because occupancy could change between plan and run; AC-GSU-024 re-measures it as half of a compound.

### §F.5 Renderer has no test

```bash
ls internal/hook/handoff_inject_render*_test.go
```

Observed: `no matches found`. This is why M7 is `cycle_type: tdd` and why AC-GSU-023 is the RED gate.

---

## §G Documentation Measurement (D5)

Backticked detector throughout (see §E on URL false positives).

```bash
grep -rlF '`/goal' docs-site/content/ | sort
```

Observed: **50 files** — 28 under `claude-code/`, 22 on the MoAI surface. Reconciles with the D5 brief's 28 / 22.

### §G.1 Affected — 13 files (brief said 9)

| Group | Files | Occurrences |
|---|---:|---:|
| `docs-site/content/*/advanced/autonomous-loops.md` | 4 | 52 |
| `docs-site/content/*/cli-reference/handoff.md` | 4 | 4 |
| `docs-site/content/*/advanced/self-evolving.md` | 4 | 8 |
| `.moai/docs/autonomous-workflow-strategy.md` | 1 | 25 |
| **Total** | **13** | **89** |

**`self-evolving.md` reclassified from retain to affected.** The brief grouped it with the factual-contrast pages. Sampling all four locales shows both references are MoAI-surface primitive naming, not a contrast statement:

```
en:40: … Recorded fields include routing decisions, gate evidence, `/moai loop` / `/goal` convergence trajectories, and subagent delegation results.
en:97: - [Autonomous Continuation Loops](…) — `/moai loop` / `/goal` convergence trajectories integrated into Loop 0 observation
```

`/goal` is named as one of MoAI's own convergence primitives whose trajectories the routing ledger records. Under W1 that reads `/moai goal`. The same two lines appear in ko / ja / zh.

### §G.2 Retain — with the per-surface rationale

| Group | Files | Occurrences | Rationale |
|---|---:|---:|---|
| `docs-site/content/*/claude-code/**` | 28 | 80 | Documents Claude Code's own feature; the statements are factually true and MoAI's emission change does not falsify them |
| `docs-site/content/*/cli-reference/goal.md` | 4 | 12 | Factual contrast — verified at `ko:9`: states `/moai goal` is the programmatic counterpart of the HUMAN-ONLY native command |
| `docs-site/content/*/utility-commands/moai-goal.md` | 4 | 4 | Factual contrast — verified at `ko:14` |
| `docs-site/content/*/advanced/hooks-reference.md` | 2 (en, ko) | 2 | Stop-hook table dual mention `` `/goal`/`/moai goal` ``, consistent with the retained yield invariant (§F.2) |
| `.moai/research/*.md` | 3 | 4 | Historical archives; retroactively editing a past record is inappropriate |

This is the **third retention surface** in the four-surface set (§F.2 supplies the fourth) — same test as `native-invocation-model.md`: the sentence stays true after retirement.

### §G.3 Locale asymmetry — brief premise corrected

```bash
ls docs-site/content/*/advanced/hooks-reference.md
```

Observed: **four files** (en, ja, ko, zh) — the page exists in all four locales, contrary to the brief's "exists only in `en` and `ko`". What differs is which locales carry the reference:

```
en:170 | `Stop` | `handle-stop-goal.sh` | goal engine — evaluates the `/goal`/`/moai goal` … |
ko:170 | `Stop` | `handle-stop-goal.sh` | goal 엔진 — `/goal`/`/moai goal` 자율 지속 조건 평가 |
ja     | (none)
zh     | (none)
```

So there is no page to create; the gap is a pre-existing four-locale **content** gap inside pages that already exist. The brief's instruction (do not manufacture symmetry) is preserved as REQ-GSU-027; only its stated premise is corrected.
