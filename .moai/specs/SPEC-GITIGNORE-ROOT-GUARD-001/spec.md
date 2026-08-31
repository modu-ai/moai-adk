---
id: SPEC-GITIGNORE-ROOT-GUARD-001
title: "Root .gitignore regression guard: the generated-artifact rules are asserted on both surfaces, not only the embedded template (card t377)"
version: "0.1.0"
status: completed
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/template"
lifecycle: spec-anchored
tags: "gitignore, template-first, regression-guard, parity, t377"
tier: S
---

# SPEC-GITIGNORE-ROOT-GUARD-001 — Root `.gitignore` Regression Guard

## §A Context and Problem

Card t373 closed a hole: `.moai/project/graph/` (written by `moai graph build`) was neither
tracked nor ignored, so a single `git add -A` could sweep a six-figure-line generated index into a
commit. The fix landed the rule on **two** surfaces — `internal/template/templates/.gitignore`
(what every deployed project receives) and this repository's own root `.gitignore` — plus a guard,
`TestEmbeddedGitignoreCoversGeneratedArtifacts`.

**The guard covers only one of the two surfaces.** It reads the embedded filesystem. A regression
that removes the rule from the repository's root `.gitignore` leaves the guard green.

This is not a hypothesis. It was reproduced in the t373 worktree (HEAD `ecdc2921c`), by removing
the rule from the root file only and leaving the template and the embedded FS untouched:

```
$ grep -c "^\.moai/project/graph/$" .gitignore
0
$ git check-ignore -v .moai/project/graph/edges.jsonl ; echo $?
1                                    # the artifact is committable again — the hole is reopened

$ go test ./internal/template/ -run TestEmbeddedGitignoreCoversGeneratedArtifacts -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	0.467s
```

The guard **passes** while the hole is open. It cannot distinguish the repaired state from the
regressed one on this axis. (Mutant restored; `git status --porcelain` returned to zero tracked
modifications.)

### Why this is a live risk — stated as measured, after two wrong rationales

This SPEC has now discarded **two** justifications that did not survive checking. Both are recorded
rather than deleted, because the pattern (a plausible `moai update` story that the code does not
support) is the thing a later reader is most likely to reach for again.

**Discarded rationale 1 — deletion.** The card's first-drafted reason was that `moai update`'s
`CleanMoaiManagedPaths` would delete the file. False: `ManagedCleanTargets` is seven entries, all
under `.claude/`, with `.moai/config` removed separately. The root `.gitignore` is in neither.

**Discarded rationale 2 — the `.sh` / `.sh.tmpl` precedent, cited with the risk direction
inverted.** An earlier draft of this SPEC cited CLAUDE.local.md §2.3's pair-drift incident (an edit
landing on one side of a two-surface pair; a later `moai update` silently reverting the deployed
side) and asserted "the shape is the same here". **It is the opposite shape.** `moai update` does
touch the root `.gitignore` — by **merge**, not deletion: `plan.go:47` assigns it `EntryMerge`, the
template-sync path backs it up before deploy and merges after (`update_template_sync.go:459`,
`:510` → `MergeGitignoreFile`, `internal/cli/update/merge/merge.go:57-97`), producing template
content ∪ user-only lines. In the `.sh` incident `moai update` was the aggressor; here it is a
**repairer** — a root-only rule loss is healed the next time `moai update` runs, because the rule
is present on the template side and merge is a union.

**The risk that actually remains, stated without borrowing anyone else's incident.** The window
between a root-only regression and the next `moai update` is **unbounded**: nothing schedules an
update, and a repository maintainer may go a long time without running one. Inside that window the
hole is fully open — one `git add -A` commits the generated artifact, and the commit is not undone
by the later merge. The regression is also **silent while it lasts**: `git status` shows the
artifact as an ordinary untracked file, and the existing guard stays green (§A). So the loss is
recoverable but the damage inside the window is not, and nothing announces the window has opened.

That is a weaker claim than "the file gets reverted", and it is the one the code supports.

## §B Measurement — what the guard can and cannot be

Three measurements are recorded below, each pinned to the tree it was taken on. They are cumulative
rather than superseding: §B.1 is the state the plan-audit iterations judged against, §B.2 is what
the base moved to, and §B.3 is the current tree after this SPEC's own run phase.

### §B.1 — plan-phase baseline, HEAD `3f03d9c36`

Measured in this worktree at HEAD `3f03d9c36` (`== origin/develop` at the time), over non-comment
non-blank lines. **Preserved verbatim**: this is the measurement both plan-audit iterations were
written against, so the later figures below extend it rather than replace it.

```
$ grep -vE '^\s*#|^\s*$' .gitignore | sort -u | wc -l                                   → 177
$ grep -vE '^\s*#|^\s*$' internal/template/templates/.gitignore | sort -u | wc -l        → 135
$ comm -23 <root> <template> | wc -l   # rules only in root                              → 44
$ comm -13 <root> <template> | wc -l   # rules only in template                          →  2
```

### §B.2 — merged base, HEAD `9328a5242`

`origin/develop` advanced after the plan-phase measurement: card t375 added
`.moai/observability/*.jsonl` to **both** the root and the template `.gitignore`. Re-measured on
that merged tree (HEAD `9328a5242`), same four commands:

```
root                                                                                    → 178
template                                                                                → 136
root-only                                                                               →  44
template-only                                                                           →   2
```

One rule was added on each surface, so both totals rose by one while the divergence sets were
unchanged. The counts moved; the argument resting on them did not.

### §B.3 — post-run tree, HEAD `9a8a99667` (current)

This SPEC's own run phase added `**/.mink/auth/` to the root `.gitignore` (REQ-GRG-007). That
brought the rule level with the template and so removed it from the template-only set — this
measurement is the one that reflects what the card changed. Measured on the current tree,
HEAD `9a8a99667`:

```
$ grep -vE '^\s*#|^\s*$' .gitignore | sort -u | wc -l                                   → 179
$ grep -vE '^\s*#|^\s*$' internal/template/templates/.gitignore | sort -u | wc -l        → 136
$ comm -23 <root> <template> | wc -l   # rules only in root                              → 44
$ comm -13 <root> <template> | wc -l   # rules only in template                          →  1
```

The sole remaining template-only rule is `.agents/skills/moai*`, which the root's broader
`.agents/` rule (`.gitignore:130`) already covers.

**A full rule-set parity assertion is therefore not available.** The two files diverge by **46 rules
on the plan-phase baseline tree `3f03d9c36`** (§B.1: 44 root-only + 2 template-only) and by **45 on
the current tree `9a8a99667`** (§B.3: 44 + 1) — by design: the root carries repository-specific entries (`.codex/`, `/i18n-validator`,
`docs-site/public/`, `.moai/logs/`, harness runtime files, …) that have no business shipping to a
user project. A test asserting set equality would fail on the day it was written.

Byte equality is unavailable for a second, independent reason: the two files are **deliberately
different in their comments**. The root file's block cites the offending commit
(`59a622c5a, +180,178`); the template's does not, because template content must stay free of this
repository's internal development state (CLAUDE.local.md §25 neutrality).

**The doctrinal ground holds; a mechanical-enforcement claim does not, and is not made.** §25
forbids commit SHAs in template content, and that is the reason the two comment blocks differ. An
earlier draft additionally asserted the prohibition was "enforced by
`template-neutrality-check.yaml`". That was **written without observation and is false for this
SHA**: the leak guard's short-SHA class matches `\b[0-9a-f]{7,8}` with a trailing boundary, and
`59a622c5a` is nine characters, so it slips past. Measured by planting the root file's exact
wording into the template `.gitignore`: `TestTemplateNoInternalContentLeak` (strict) and
`TestTemplateNeutralityAudit` both passed. Re-planting with an 8-character SHA fired
`class=S2-short-sha-sentence-final` and failed — so the walker does scan `templates/.gitignore` and
the class does work; only this SHA's length evades it. Nothing in this SPEC depends on the
enforcement claim, and the byte-equality exclusion rests on the doctrine, not on a guard.

**The repository has already solved this exact problem once, for a different file class.** Rule
files mirrored into the template are guarded on two tiers:

- `internal/template/rule_template_mirror_test.go` — byte-identity, for pairs that are genuinely
  identical.
- `internal/template/sanitized_pair_parity_test.go` (`TestSanitizedPairParity`, sentinel
  `SANITIZED_PAIR_PARITY_DRIFT`) — for pairs **deliberately excluded from byte-parity** because the
  template copy is §25-sanitized. It normalizes away the intentionally-divergent tokens (SPEC-IDs,
  REQ/AC tokens, commit SHAs, dates) from both copies and then compares content, tolerating the
  rewording that sanitization legitimately produces. Its own header states the hazard it closes:
  *"a real (non-token) doctrine change to the LOCAL copy can silently fail to propagate to the
  template mirror — users receive stale doctrine, and nothing catches it."*

That is the same hazard this SPEC addresses, on a different file. The precedent therefore does not
merely rule byte-parity out; it establishes the shape of the correct answer — **compare declared
content after setting aside the intended divergence**, rather than compare bytes.

**The `.gitignore` pair is guarded by neither tier.** Both existing guards are scoped to rule files
under `.claude/rules/moai/**`; `.gitignore` sits at the repository root and in the template root,
outside both scopes. It is a two-surface pair with the same drift hazard and **zero** coverage —
which is why the t373 mutation passed.

This SPEC's design is the narrow member of that family: rather than compare all content, assert a
**declared set of rule lines** on both surfaces. The declared set is the thing the project has
decided must exist everywhere; everything else about either file stays free to differ.


### The `**/.mink/auth/` rule — in scope by operator judgment, deliberately outside the guard

**This subsection's premise describes a state this card has since consumed — read it in two tenses.**
At the plan-phase baseline (§B.1, HEAD `3f03d9c36`) there were 2 template-only rules: one covered by
a broader root rule (`.agents/skills/moai*` ⊂ `.agents/`), and `**/.mink/auth/`, which was not
covered by anything:

```
# on HEAD 3f03d9c36 — plan-phase state, the premise this subsection argues from
$ grep -c "mink" .gitignore    → 0
```

`**/.mink/auth/` — commented in the template as MINK agent credentials, plaintext API keys — was
present in every deployed project and **absent from this repository's own `.gitignore`**. Measured
state at the time: no `.mink` directory existed here (`ls -d .mink` → no such file) and no `mink`
path was tracked, so nothing was exposed.

**Current state (HEAD `9a8a99667`)**: the rule is now on both surfaces, because this card put it on
the root — `.gitignore:182` alongside the template's `internal/template/templates/.gitignore:169`.
It is consequently no longer template-only, which is why §B.3 measures 1 template-only rule where
§B.1 measured 2. The gap this subsection describes is closed by REQ-GRG-007, not still open.

An earlier draft placed this out of scope on scope-discipline grounds. **Operator judgment relayed
2026-08-31 put the one-line rule in scope, while holding the guard's own scope unchanged.** Both
halves are load-bearing and are recorded separately:

- **In scope**: add `**/.mink/auth/` to the repository's root `.gitignore`, bringing it level with
  the template. One line; no template change (the template already carries it). *Landed in the run
  phase — `.gitignore:182` on HEAD `9a8a99667`.*
- **NOT in scope**: adding it to `generatedArtifactIgnoreRules`. The declared list is the set of
  **generated-artifact** rules the guard asserts on both surfaces; a credential rule is a different
  class, and whether credential rules should also be guarded is a separate judgment. Widening the
  list here would widen what the guard must be right about, in the same change that is trying to
  establish the guard.

The distinction matters beyond this rule: it is what keeps "add a missing rule" and "extend the
guard's contract" from becoming the same act by default.

## §C Requirements

Modality is `SHALL`, matching this repository's measured convention — over `.moai/specs/*/spec.md`,
`The <subject> <modality>` statements use `shall`/`SHALL` **1536** times against `MUST` **107**.
An earlier draft used `MUST` throughout; the substitution is convention alignment, not a change of
force.

- **REQ-GRG-001** — The guard SHALL assert the generated-artifact rule set against the repository's
  root `.gitignore` in addition to the embedded template filesystem.
- **REQ-GRG-002** — The source SHALL hold exactly one definition of the declared rule set. Two
  copies, one per surface, would reproduce the drift the guard exists to prevent.
- **REQ-GRG-003** — The guard SHALL compare discrete rule lines — never bytes, and never full
  rule-set equality — for the two reasons measured in §B.
- **REQ-GRG-004** — The guard SHALL NOT fire on the two files' intended differences: divergent
  comments (including this repository's commit-SHA citation) and the 46 rules that legitimately
  exist on one side only.
- **REQ-GRG-005** — A failure message SHALL name which surface is missing which rule, so the reader
  is not left to diff two files by hand.
- **REQ-GRG-006** — The guard's test function SHALL be named
  `TestGitignoreDeclaredRulesOnBothSurfaces`. The name is pinned here because the acceptance
  criteria select the test by exact name: a rename that is not made here breaks the selector
  loudly, rather than silently selecting nothing (§D preamble).
- **REQ-GRG-007** — The root `.gitignore` SHALL carry `**/.mink/auth/`, and that rule SHALL NOT be
  added to the declared generated-artifact set (§B). The rule closes a credential-visibility gap
  between this repository and every project deployed from the template; the exclusion keeps the
  guard's contract to generated artifacts, which is what its criteria are written against.

## §D Acceptance Criteria

**Selector discipline (this SPEC's own subject, applied to itself).** Every criterion below selects
the guard by **exact anchored name**, `-run '^TestGitignoreDeclaredRulesOnBothSurfaces$'`. A prefix
selector is prohibited here: an unanchored `-run TestGitignore` selects the pre-existing, unrelated
`TestGitignore_IgnoresSkillMirrorOnly` and **does not select this guard at all** — measured on this
tree, HEAD `3f03d9c36`:

```
$ go test ./internal/template/ -run TestGitignore -count=1 -v
=== RUN   TestGitignore_IgnoresSkillMirrorOnly
--- PASS: TestGitignore_IgnoresSkillMirrorOnly (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.477s
```

An earlier draft of this section used exactly that selector, which made AC-GRG-001 green before any
work existed — the empty-sweep defect this SPEC was written to eliminate, reproduced inside the
SPEC. It is recorded rather than quietly corrected, because a reader who does not know it happened
here is likely to write the prefix form again.

- **AC-GRG-001** — With the tree unmodified, the guard passes AND is actually selected.
  Verify: `go test ./internal/template/ -run '^TestGitignoreDeclaredRulesOnBothSurfaces$' -count=1 -v`
  → the output contains `--- PASS: TestGitignoreDeclaredRulesOnBothSurfaces`. The `--- PASS:` line
  is the load-bearing part, and its necessity is measured rather than assumed. A zero-selection run
  on this tree (the guard does not exist yet, so the anchored selector currently matches nothing)
  prints:

  ```
  $ go test ./internal/template/ -run '^TestGitignoreDeclaredRulesOnBothSurfaces$' -count=1 -v
  testing: warning: no tests to run
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/template	0.472s [no tests to run]
  $ …same without -v
  ok  	github.com/modu-ai/moai-adk/internal/template	0.248s [no tests to run]
  ```

  Three things separate that from a real pass, and only the third is what this AC relies on: the
  ` [no tests to run]` suffix on the `ok` line (present in **both** forms), the
  `testing: warning:` line under `-v`, and the **absence of any `--- PASS:` line**. An earlier
  draft of this AC asserted that "a bare `ok` is also what a zero-selection run prints" — that is
  **false**, as the measurement above shows, and it was written without running the command inside
  the very section whose subject is not doing that. The instrument was never affected (requiring
  `--- PASS:` is strictly stronger than the rationale that was given for it); the reasoning was.
- **AC-GRG-002** (mutation, root surface) — Removing a declared rule from the root `.gitignore`
  only, leaving template and embedded FS untouched, turns the guard RED.
  Verify: plant, run the AC-GRG-001 command, observe `--- FAIL`, restore, observe
  `git status --porcelain .gitignore` empty. RED-now on the pre-implementation tree is already
  recorded: today's guard passes this mutant (§A), which is the defect.
- **AC-GRG-003** (mutation, template surface) — Removing a declared rule from the template source
  **and re-embedding via `make build`** turns the guard RED. Preserves the property t373
  established: the guard reads the embedded FS, so it catches a template edit that skipped
  `make build`.
  Verify: plant in `internal/template/templates/.gitignore`, run `make build`, run the AC-GRG-001
  command, observe `--- FAIL`, restore, `make build`, observe PASS.
  **RED status: not yet observed for this guard.** The equivalent RED was observed in card t373 for
  the single-surface guard (template rule removed + `make build` → FAIL). Observing it for this
  guard is a run-phase obligation, not a plan-phase claim.
- **AC-GRG-004** (false-positive proof — asserts a NON-firing, so it carries its own command) —
  With both surfaces correct, the guard passes **while the two files genuinely diverge**.
  Verify, in this order: (1) re-measure the divergence
  (`comm -23`/`comm -13` over the two sorted rule sets) and record both counts; (2) confirm the
  root-only count is non-zero — a pass proves nothing if divergence has since gone to zero;
  (3) run the AC-GRG-001 command and observe `--- PASS`.
  **Known limit, stated rather than papered over**: a mutant that skips its assertions whenever
  divergence is non-zero would satisfy this criterion while violating REQ-GRG-004. AC-GRG-002 is
  what closes that mutant — the two criteria must both hold, and neither is sufficient alone.
- **AC-GRG-005** (preservation check, not a flip — it holds today and must keep holding) — The
  declared rule set has exactly one declaration site.
  Verify: `grep -rn '^var generatedArtifactIgnoreRules' internal/template/ --include='*.go'` →
  exactly one line; record the file path it names, because AC-GRG-007(b) is evaluated against
  that same file. Anchoring to the declaration syntax is required: the unanchored form counts every
  occurrence (declaration + each use + comment mentions) and returned `3` on this tree, so it can
  never equal one and is unsatisfiable as a criterion.
- **AC-GRG-006** — The failure message names the surface and the rule.
  Verify: in the AC-GRG-002 and AC-GRG-003 mutant runs, the failure output contains BOTH the
  missing rule string (`.moai/project/graph/`) AND a surface token distinguishing the two files
  (`root` / `embedded`). Both are literal string checks, not a reader's judgement.

- **AC-GRG-007** (both halves, because either alone is the wrong outcome) —
  (a) `git check-ignore -v .mink/auth/token` exits 0 and names the root `.gitignore`; AND
  (b) `grep -c 'mink' <the file AC-GRG-005 located>` → `0`, showing the rule did NOT enter the
  declared set. **The target is the file AC-GRG-005 names, not a path written here.** Today that is
  `internal/template/embed_gitignore_generated_test.go`, but REQ-GRG-002 constrains the *count* of
  declaration sites, not their *location*: if the run phase moves the declaration, a hard-coded
  path would grep a file that no longer holds the declared set and pass green while observing
  nothing — the same vacuity D1 exhibited, relocated.
  RED-now on this tree: (a) currently fails — `grep -c "mink" .gitignore` → `0`, so the path is not
  ignored; (b) currently passes and must keep passing, so it is a preservation check rather than a
  flip.

### Requirement ↔ criterion map

Stated explicitly so no requirement rests on an implicit reading:

| Requirement | Asserted by |
|---|---|
| REQ-GRG-001 (root surface asserted) | AC-GRG-002 (root mutant → RED) |
| REQ-GRG-002 (one declaration site) | AC-GRG-005 |
| REQ-GRG-003 (rule lines, not bytes, not full equality) | AC-GRG-004 — the guard passes while the comment blocks differ (so not bytes) and while 46 rules diverge (so not full-set equality); both conditions are re-measured by the criterion rather than assumed |
| REQ-GRG-004 (no false fire on intended divergence) | AC-GRG-004, with AC-GRG-002 closing its stated mutant |
| REQ-GRG-005 (failure names surface + rule) | AC-GRG-006 |
| REQ-GRG-006 (pinned test name) | AC-GRG-001 — the anchored selector fails loudly if the name moves |
| REQ-GRG-007 (mink rule added, not declared) | AC-GRG-007 (a) and (b) — both halves required |

## §E Scope

### In Scope

- One test file under `internal/template/`, extending the existing declared-rule mechanism to a
  second surface. No production code.
- One line added to the repository's root `.gitignore`: `**/.mink/auth/` (REQ-GRG-007), which the
  template already carries. It is **not** added to the declared generated-artifact set — see §B for
  why the two acts are kept separate.

### Out of Scope

Recorded rather than dropped — each is a decision someone should make, not a thing forgotten:

- **Whether credential rules belong in the guard's declared set.** The `**/.mink/auth/` rule is
  added (above) but not declared; extending the guard's contract from generated artifacts to
  credential rules is a separate judgment with its own criteria.
- The four paths t373 left open by decision (`.moai/observability/`,
  `.moai/project/navigator/` artifacts) or by deliberate visibility
  (`.moai/chain/`, `.moai/.migrate-tx-*.json`). This SPEC changes no ignore policy; it only asserts
  that whatever is declared exists on both surfaces.
- Any general two-file parity mechanism for `.gitignore` beyond the declared set — ruled out by the
  §B measurement, not deferred for effort.

**Explicitly not claimed**: this guard protects the *declared* rules. A generated path nobody has
declared is invisible to it, exactly as it was to t373's guard. The declared list grows by human
edit; no mechanism discovers new generated paths.
