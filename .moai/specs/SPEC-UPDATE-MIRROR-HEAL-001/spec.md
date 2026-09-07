---
id: SPEC-UPDATE-MIRROR-HEAL-001
title: "moai update repairs a deleted .agents/skills codex mirror"
version: "0.5.0"
status: in-progress
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "update, codex, skill-mirror, repair, self-heal"
tier: M
---

# SPEC-UPDATE-MIRROR-HEAL-001 — `moai update` repairs a deleted `.agents/skills` codex mirror

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-07 | Initial plan-phase draft (card t520). Grounded in code read on this tree (`0b1e27877`) plus one inherited measurement from card t498. Chose the `.codex/` wiring-existence gate. |
| 0.5.0 | 2026-09-07 | Plan-audit iteration 3 (PASS 0.91) follow-ups. **R3-3**: REQ-UMH-010's code coordinates demoted to a dated reference (tree `0b1e27877`, 2026-09-07) with a re-measure command — kept, not deleted, because a coordinate is what lets the next reader re-measure. **R3-2 (judgment, no change)**: REQ-UMH-010 sits between 008 and 009 in emission order; the numbering is deliberately **not** rearranged — a final-stage renumber is exactly the act that produced iteration 2's miscount (N2), so the ordering cost is accepted over the churn risk. `acceptance.md` 0.4.0 carries the R3-1 repair; this file's requirement text is unchanged by it. |
| 0.4.0 | 2026-09-07 | Plan-audit iteration 2 repair (FAIL 0.84; iteration-1 blockers confirmed closed, these are defects introduced by that repair). **N1**: 0.3.0 §3.6 ground ② described the existence intersection as an existing property of the producer. It is not — measured: `skill_mirror.go` Lstats only the *mirror* path (`:206`), never `srcDir`, which is used solely by the copy fallback (`:247`), and `os.Symlink` (`:168`) succeeds against a missing target. Absent a new check, an empty-`.claude/skills` fixture yields **16 dangling links reported as `MirrorModeSymlink`**, not 0 entries. Promoted to **REQ-UMH-010** (a requirement on new work, no longer an observation), and `plan.md` M3's conflicting last sentence rewritten. **N4**: §3.1 K4 still forbade what §3.6 deliberately decided to do; corrected in place with the superseded wording retained. |
| 0.3.0 | 2026-09-07 | Plan-audit iteration 1 repair (FAIL 0.81, `.moai/reports/t520/plan-audit.md`). **D1**: AC-UMH-013 rested on `git diff -- .agents`, which cannot fail — `.agents/` is gitignored (`git check-ignore -v .agents/skills/foo` → `.gitignore:133:.agents/`, rc 0; control `internal/cli/update.go` → rc 1), so the guard is re-written onto a filesystem snapshot and added to AC-UMH-014's mutant list; every other absence-as-PASS criterion gains a control. **D2**: the gate's converse ("stamp ≥ 3.1.3 ⟹ a mirror-creating deploy completed") is false — `system.yaml` is written *inside* the walk while `mirrorSkills` runs only after both walk guards, and init swallows deploy failure as a warning (`initializer.go:225-230`). Both branches evaluated in new §3.6; branch (나) chosen with argument, and the boundary pinned by AC-UMH-015/016 rather than left unverified. **D3**: version-fixture matrix added (`v` prefix, pre-release, non-numeric). **D4**: §4's `deployedSkills` description corrected against the code comment (recorded *before* any skip branch, not "the set the walk wrote"). |
| 0.2.0 | 2026-09-07 | §3 rewritten. The lead's measurement refuted the `.codex/` gate: `init.go:191-193` returns before wiring for `--agent claude`, while mirror creation is unconditional, so "mirror present, `.codex/` absent" is the **default** state of a claude-only project and a `.codex/` gate would close repair permanently for that population. Both originally-proposed candidates are now recorded as rejected, and a new predicate (deploy-version stamp) is derived from code with a four-constraint conformance table. Coordinate corrected: `refreshCodexWiringBestEffort` is `update.go:507`, not `:517`. |

## §1 Context

Codex CLI does not scan `.claude/skills`; it scans `<repo>/.agents/skills`. That directory is the
only path by which a MoAI skill catalog is reachable from Codex, and it is populated by two
mechanisms, both of which live inside the template `Deploy` call:

- **Path A — runtime symlink mirror.** `internal/template/deployer.go:297-299` calls
  `d.mirrorSkills(projectRoot, deployedSkills)` as the last step of `Deploy`, creating
  `.agents/skills/<name>` → `../../.claude/skills/<name>` per skill this run deployed. Fail-open,
  and deliberately not manifest-tracked (`internal/template/skill_mirror.go` header).
- **Path B — template-direct publication.** 16 real files ship in the embedded template tree at
  `internal/template/templates/.agents/skills/moai-*/SKILL.md` (measured:
  `find internal/template/templates/.agents/skills -name SKILL.md | wc -l` → `16`; the same 16 names
  are pinned in `internal/template/published_skills.go`), deployed by the ordinary template walk —
  that is, also inside `Deploy`.

When the mirror is deleted, `moai update` cannot restore it. The causal early return is
`internal/cli/update_template_sync.go:617-622`, inside `runTemplateSyncWithProgress`: on version
match without `--force` it prints the already-up-to-date pill and returns `(true, nil)` **without
calling `runTemplateSyncWithReporter`**, which is what calls `Deploy`. No `Deploy`, no mirror — via
either path.

**Coordinate corrections (recorded, not silently adopted).**

1. The card's `internal/cli/update/update.go:509` does not exist on this tree. The real cause is
   `update_template_sync.go:617-622`; the `if syncSkipped { return nil }` block that does exist —
   `internal/cli/update.go:518` — is a **downstream consequence**, short-circuiting the post-sync
   follow-up phase only after the sync step has already declined to deploy.
2. The architectural precedent `refreshCodexWiringBestEffort` is called at
   `internal/cli/update.go:507` (its explaining comment at `:500-506`), not `:517`. Its substance is
   unaffected; only the citation is corrected.

**Inherited measurement (card t498, NOT measured by this SPEC's author).** In an isolated scratch
project: `moai init` → mirror entries 24 → delete `.agents` → `moai update --yes` rc=0 → 0 →
`moai update --templates-only --force` → 37. Treated as t498's evidence; §6's reproduction
re-establishes the shape independently rather than resting on those numbers.

**Consequence for card t503.** t503 added Path B behind the same gate as Path A. It therefore did
not fix this defect; it added a second artifact that a version-matched update also cannot restore.
One repair seam must cover both paths — REQ-UMH-004.

## §2 Requirements (GEARS)

**REQ-UMH-001 (event-driven, capability-gated).**
Where the project's recorded deploy version is one whose deploy creates a skill mirror, When
`moai update` completes its template-sync step — including the version-matched step that deployed
nothing — the update command shall run a mirror-repair pass over `.agents/skills`.

**REQ-UMH-002 (unwanted).**
Where the project records no deploy version, or records one that predates skill-mirror creation,
the update command shall not create, modify, or remove any path under `.agents/`.

**REQ-UMH-003 (ubiquitous).**
The template-sync step shall retain its version-match early return unchanged, in its present
position, with its present `(skipped, err)` contract.

**REQ-UMH-004 (event-driven).**
When the mirror-repair pass runs, it shall restore both mirror-producing artifacts: the per-skill
symlink entries of Path A and the template-published `SKILL.md` files of Path B.

**REQ-UMH-005 (unwanted).**
The mirror-repair pass shall not overwrite, replace, or remove an entry that already occupies a
mirror path and is not a MoAI-created symlink.

**REQ-UMH-006 (state-driven).**
While the mirror-repair pass encounters a failure, the update command shall emit a warning and
continue, and shall not return that failure as an update error.

**REQ-UMH-007 (ubiquitous).**
The mirror-repair pass shall mirror only skills the template ships, and shall not create a mirror
entry for a locally-authored skill directory that no deploy produced.

**REQ-UMH-008 (event-driven).**
When the repair pass lands, the `moai doctor` mirror-absent guidance shall be reconciled with the
new behavior, because its present detail text asserts as fact that "a routine `moai update` on a
version-matched project does not restore it" (`internal/cli/doctor_codex.go:457-461`) — a statement
this card makes false.

**REQ-UMH-010 (unwanted).**
The mirror-repair pass shall not create a mirror entry whose canonical target directory
(`.claude/skills/<name>`) does not exist on disk.

This is a requirement on **new** work, not a property of the producer.

> **Dated reference — measured on tree `0b1e27877`, 2026-09-07.** The three coordinates below were
> each verified at that commit and are recorded so a later reader can **re-measure**, not so they can
> be trusted unmeasured. Line numbers drift; the anchors are the function and symbol names.
> Re-measure with `grep -n "Lstat\|os.Symlink\|srcDir" internal/template/skill_mirror.go`.

At that tree: `mirrorOneSkill` Lstats only the mirror path (`skill_mirror.go:206`); `srcDir` is
referenced solely by the copy fallback (`:247`); and `os.Symlink` (`:168`) succeeds even when its
target is absent, so the existing producer would emit dangling links reported as
`MirrorModeSymlink`. A dangling entry is
damage, not repair — `moai doctor` raises it as its own finding
(`doctor_codex.go` dangling branch) — so the repair pass must not manufacture the very state the
diagnostic warns about. The check is a **filter applied before linking**, which leaves REQ-UMH-007's
derivation source (the embedded template FS) untouched: names are derived from the template, then
filtered by target existence.

**REQ-UMH-009 (unwanted).**
The reproduction and verification of this defect shall not read, write, or delete this repository's
own `.agents` state.

## §3 Existence-gate decision — which predicate defines "this project already had a mirror"

This is the SPEC's substantive decision. Two candidates were proposed at card intake; **both are
refuted by measurement**, and the record below exists so the next reader does not re-walk either.

### §3.1 The four constraints a predicate must satisfy

| # | Constraint | Why |
|---|---|---|
| K1 | Survives deletion of `.agents/` | That deletion **is** the state being repaired; evidence living inside it is gone exactly when needed |
| K2 | Independent of the Codex opt-in | Mirror creation is unconditional, so opt-in state does not correlate with mirror possession |
| K3 | Not the manifest | `skill_mirror.go`'s header states mirror entries are deliberately untracked (directory-symlink hashing fails EISDIR) |
| K4 | Excludes the population that predates the mirror feature | C-1: an update must not **create** `.agents/` in a project the feature never targeted — i.e. one whose last deploy was performed by a binary with no mirror step |

> **K4 wording superseded (0.4.0).** K4 previously read "excludes the never-had-a-mirror population
> … a project **no deploy ever gave one**". §3.6 then deliberately decided to repair one class of
> never-had-a-mirror project (stamped-but-partial), which made the old wording false inside its own
> document. The row above narrows K4 to the population it actually protects — the pre-feature one —
> and the original phrasing is kept here so the change is legible: what moved is not the constraint's
> purpose but the recognition that "never had a mirror" and "the feature never targeted it" are
> different sets, and only the second is what C-1 defends.

### §3.2 Rejected candidate R1 — `.agents/` exists

Fails **K1**, decisively. The measured case deletes the whole directory (t498: 24 → 0), so the gate
is closed at precisely the moment repair is needed. A gate that is closed in the failure state is an
obstruction, not a gate.

### §3.3 Rejected candidate R2 — Codex wiring files exist (`.codex/hooks.json` / `config.toml`)

Fails **K2**. Three measurements on this tree:

1. `internal/cli/init.go:191-193` (`wireCodexUnlessClaude`) returns before `codexwiring.Wire` when
   the resolved wiring is `agentWiringClaude` — so a `--agent claude` project gets **no `.codex/`
   at all**.
2. `wiringFilesExist` (`internal/codexwiring/codexwiring.go:96-103`) stats exactly
   `.codex/hooks.json` and `.codex/config.toml`, nothing else.
3. `WithSkillMirror` has no production caller: `grep -rn "WithSkillMirror" internal` returns three
   hits — the option definition, its doc comment, and a comment at `deployer.go:91` — with the
   remaining matches in tests. The zero value keeps mirroring ON, so **every** `Deploy` creates a
   mirror.

Therefore "mirror present, `.codex/` absent" is not an edge case: it is the **default state of every
claude-only project**. Gating on `.codex/` would close repair permanently for that population,
violating C-1's intent in the mirror image of R1's failure. The doctor precedent (mirror inspected
in the WIRED branch only, REQ-CMD-008) argues for *silence* toward claude-only users, which is a
weaker claim than *refusing them repair*, and it does not rescue R2.

### §3.4 Candidate set and conformance

| # | Predicate | K1 survives deletion | K2 opt-in independent | K3 not manifest | K4 excludes never-had-mirror |
|---|---|---|---|---|---|
| R1 | `.agents/` exists | **NO** | yes | yes | yes |
| R2 | `.codex/` wiring files exist | yes | **NO** | yes | partial (accidentally) |
| R3 | Manifest carries `.agents/skills` entries | yes | yes | **NO** | partial |
| R4 | `.claude/skills` exists (i.e. any MoAI project) | yes | yes | yes | **NO** |
| **R5** | **`moai.template_version` ≥ the mirror-introducing version** | **yes** | **yes** | **yes** | **yes** |

**R3 rejected on K3 plus measurement.** Path A entries are never tracked by contract. Path B files
*are* ordinary template files and do flow through `Track`, but the witness is measured absent here:
`grep -c '\.agents/skills' .moai/manifest.json` → `0`, against the control
`grep -c '\.claude/skills' .moai/manifest.json` → `436` (a non-empty control, so the `0` is a
measurement rather than a broken selector). A manifest witness would additionally be blind to any
project last deployed before t503 shipped Path B.

**R4 rejected on K4.** Every MoAI project has `.claude/skills`, including projects last deployed
before the mirror existed; R4 would create `.agents/` for a population that never had one.

### §3.5 Chosen: R5 — the recorded deploy version stamp

`.moai/config/sections/system.yaml` carries `moai.template_version`, written from the binary version
at deploy time (`internal/template/templates/.moai/config/sections/system.yaml.tmpl:9` →
`template_version: "{{.Version}}"`) and read by `plan.GetProjectConfigVersion`
(`internal/cli/update/plan/plan.go:286-326`), which returns `"0.0.0"` when the file or the field is
absent. The predicate:

> Repair when the project's recorded `template_version` is greater than or equal to the first
> release whose deploy creates a skill mirror.

- **K1** — the stamp lives under `.moai/`, untouched by deleting `.agents/`.
- **K2** — written on every deploy regardless of `--agent`.
- **K3** — not the manifest.
- **K4** — the never-had-a-mirror population is real and distinguishable: `skill_mirror.go` was
  added `9c94c6b7a` (2026-08-22) and the mirror shipped in the **`[3.1.3] - 2026-08-24`** CHANGELOG
  section ("A skill mirror at `.agents/skills` for codex-cli, derived from the run rather than
  hand-listed"). A project last deployed by ≤ 3.1.2 carries a stamp below that and is excluded; the
  absent/unparseable case degrades to `"0.0.0"`, which is also excluded — a safe default that
  additionally makes the pass inert outside a MoAI project.

The version-introducing constant is named once in code (a single exported constant, one place to
change) and its value is fixed at run phase from the release record above rather than guessed.
A loose comparator already exists for reuse: `compareVersionLoose`, `internal/cli/update_version.go:234`.

**Why this is not R4 in disguise.** In the early-return case the stamp is *equal* to the running
binary's version by construction (that equality is what triggers the return), so on a ≥ 3.1.3 binary
the gate is satisfied precisely because the project's last deploy was performed by a mirror-creating
binary. R4 asserts nothing about that history; R5 reads it.

**Accepted cost, stated explicitly.** A user who deletes `.agents/` deliberately, as an opt-out, on
a ≥ 3.1.3 project will see it restored by the next update. This SPEC treats deletion as damage
rather than as a signal, because deletion is indistinguishable from the failure mode being repaired.
An explicit opt-out marker is named in §7 as out of scope and is the correct home for that need.

**Gap in the inherited measurement.** t498's run does not record its scratch project's `--agent`
value or its stamp, so it does not by itself establish gate coverage. §6's reproduction constructs
the precondition by *running a deploy* rather than by hand-writing a stamp (§3.6 D2 repair).

### §3.6 The gate's converse is false — evaluated, and accepted with a pinned boundary

The gate needs "stamp ≥ 3.1.3 ⟹ a mirror-creating deploy completed for this project". That
implication is **false**, measured on this tree:

1. `system.yaml` is written **inside** the template walk (`initializer.go:220` comment: "Templates
   include `.moai/config/sections/*.yaml`"; the stamp source is
   `system.yaml.tmpl:9 template_version: "{{.Version}}"`).
2. `mirrorSkills` runs **only after both walk guards** — `deployer.go:286-298`:
   `if walkErr != nil { return }`, `if deployErr != nil { return }`, *then* the mirror call.
3. A failed deploy is **non-fatal at init** — `initializer.go:225-230`: the error is appended to
   `result.Warnings` and Init continues (the manifest step is Step 5, reached regardless).

So a walk that writes `system.yaml` and then fails leaves a project stamped ≥ 3.1.3 that never had
a mirror. Two branches were evaluated.

**Branch (가) — add a second conjunct witnessing walk completion. Rejected: no such witness exists,
and the nearest approximation is actively counter-productive.**

| Candidate conjunct | Verdict |
|---|---|
| Mirror outcome recorded under `.moai/` | Does not exist. `DeployResult.SkillMirrors` is consumed only by `internal/mirrornotice/notice.go:55` for display; nothing persists it (`grep -rn "SkillMirrors" internal \| grep -v _test` → 7 hits, all producer, display, or accessor). |
| `.moai/manifest.json` `version` / `deployed_at` | Written at Init **Step 5** (`initializer.go:645-660`), which is reached after the Step-3 non-fatal swallow — so it is written just as readily when the deploy failed. Not a completion witness. |
| Manifest carries a `.claude/skills/**` entry | Narrows but does not close the hole (a walk can fail after the skills subtree), adds the manifest dependency K3 cautions against, **and denies repair to exactly the population that most needs it** — see the argument below. |

**Two externally-supplied arguments, recorded on both sides.** The lead offered one argument for
each branch, explicitly as input rather than as a decision. Both are recorded; neither is the basis
of the choice.

- *For (나):* "the existence-gate contract binds the opt-in axis, not the deploy-success axis
  (`wire.go:51-56`); a project whose deploy failed **did** run init or update — the opt-in existed
  and the execution failed — so mirror creation is completion of a failed run, not new
  provisioning."
- *For (가):* "that project never had a mirror, so the word *repair* does not apply; the card is
  self-**healing**, and making something that never existed is scope escape."

**Verification of the (나) argument's premise, since it rests on one.** The premise is that
`wire.go`'s contract protects the opt-in axis rather than the success axis. Read verbatim
(`internal/codexwiring/wire.go:47-51`): *"File existence is the user's standing opt-in — a
`--agent claude` (or flag-absent) init left no wiring behind, and an update must not create any."*
The comment **attributes the absence to a flag choice** and forbids creation on that basis, so it
supports the opt-in reading — but it is **silent on the failed-run case**, which is a different
cause of absence than the one it names. Two things therefore limit its force, and both are stated
rather than smoothed: it is silent on the case at hand, and it governs *codex wiring*, a gate this
SPEC deliberately does **not** use (§3.3 R2). Its force here is analogical corroboration, not
authority.

**Branch (나) — accept, with the boundary pinned. Chosen — on the two measured grounds below, not
on the supplied argument.**

C-1's wording is "REPAIR of a project that already had the mirror, not creation", and the population
it names is the one that *predates the feature*. A stamped-but-partial project is not that
population: a mirror-creating deploy **was** run against it and did not finish. Repairing it is
completing an interrupted deploy, not creating something for a project the feature never targeted —
which is why R5's version boundary, not this case, is what C-1 actually turns on.

Two further facts settle it in the same direction:

- **Without the pass, that project is stuck forever.** The partial deploy wrote the stamp, so every
  later `moai update` version-matches and never deploys. The repair pass is the only path by which
  such a project ever gets its mirror; a conjunct that excluded it would harden the trap.
- **The pass is bounded on Path A — by REQ-UMH-010, which this SPEC adds.** With the
  target-existence filter, a deploy that failed before the skills subtree yields zero Path A
  entries, so the repair cannot invent skills the project does not have.

  **This was stated wrongly in 0.3.0 and the correction matters.** That version presented the bound
  as an existing property of the producer. It is not: `mirrorOneSkill` never stats `srcDir` on the
  symlink path (`skill_mirror.go:206` Lstats the *mirror* path; `:247` is the copy fallback) and
  `os.Symlink` (`:168`) succeeds against a missing target — so the unmodified producer would create
  **16 dangling links** on this fixture, reported as `MirrorModeSymlink`. The bound is therefore a
  requirement placed on new work (REQ-UMH-010), verified by AC-UMH-015, and the acceptance argument
  above depends on that requirement being implemented rather than inherited.

**Weighing the (가) argument fairly.** "It never had a mirror, so *repair* is the wrong word" is
correct about the word and does not settle the act. The question the gate must answer is which
population C-1 protects, and the measured answer is the pre-3.1.3 one — a population R5 excludes on
evidence (§3.5 K4). Against that, the stuck-forever property is a measured consequence of this
tree's own control flow, not a preference: excluding the partial project makes the defect permanent
for it. Where a naming objection meets a measured trap, the measurement decides.

The residual risk this leaves is therefore narrow and is stated rather than hidden: on a
stamped-but-partial project the pass will restore the 16 Path B published files even though that
project never held them. That is file creation from templates in a project whose deploy was
interrupted, and it is accepted on the "completing an interrupted deploy" reading above.
**Acceptance does not mean unverified**: AC-UMH-015 and AC-UMH-016 construct the stamped-but-partial
state and pin both halves of the behavior, so a future change that alters it fails a test rather
than passing silently.

## §4 Repair-scope decision — which skills the pass mirrors

`mirrorSkills` is driven by `deployedSkills` — which is **not** "the set the walk wrote". The code
records a skill name *before any skip branch* (`deployer.go:216-223`, comment: "Record the skill
this file belongs to, **before any skip branch**: a re-deploy whose files all already exist still
owes its mirror"), so the set is "template-shipped skill names the walk **encountered**", including
those whose files were skipped as already-present or user-owned. (Corrected from 0.2.0, which
asserted the opposite; the false premise sat directly above the M3 implementation decision.)

That correction *strengthens* S2 below rather than disturbing it: the producer's own set is
membership-based, not write-based, so a repair pass reproducing it must enumerate template-shipped
names rather than observe writes. A repair pass running without a deploy has no such set and must
supply one.

- **S1 — every directory under `.claude/skills`.** Rejected: it would create mirror entries for
  locally-authored skills no deploy produced. `moai doctor` declines to raise an unmirrored finding
  for exactly this reason (`codexMirrorObservations` comment: the denominator "is not observable
  from here").
- **S2 — the template-shipped skill names that are present under `.claude/skills`.** Chosen: it
  restores exactly what a deploy would have created, and nothing else. This is REQ-UMH-007.

## §5 Constraints

- **C-1 (no creation).** Nothing under `.agents/` is created in a project that never had a mirror
  (REQ-UMH-002, gate K4).
- **C-2 (early return preserved).** The version-match early return at
  `update_template_sync.go:617-622` is a performance optimization and is neither removed, relocated,
  nor weakened. The repair pass runs beside it, at the `update.go:507` position (REQ-UMH-003).
- **C-3 (isolation).** Reproduction happens only in an isolated temporary project (`t.TempDir` or a
  scratch dir). This repository's own `.agents` state is an observation subject for sibling cards
  t498 and t510 and is not read, written, or deleted by this card's work (REQ-UMH-009).
- **C-4 (report vs repair boundary).** `moai doctor` (SPEC-CODEX-MIRROR-DOCTOR-001, card t498)
  **reports** mirror state and repairs nothing — `inspectSkillMirror` is read-only by requirement
  (REQ-CMD-002). This card **repairs** and adds no diagnostic row. The doctor never gains a write;
  the repair pass never gains a doctor finding. Their one coupling is textual: REQ-UMH-008.
- **C-5 (fail-open).** Mirror creation is fail-open by contract in `skill_mirror.go`; the repair
  pass inherits that stance (REQ-UMH-006).

## §6 Verification approach

A Go test in `internal/cli` builds a project under `t.TempDir()` with a stamped
`.moai/config/sections/system.yaml`, deploys, deletes `.agents`, runs the update path at version
match, and asserts restoration — plus the two negative cases (stamp below the mirror-introducing
version, and stamp absent) asserting the tree is left byte-identical. No shell reproduction against
this repository. Full criteria: `acceptance.md`.

## §7 Exclusions

### Out of Scope — an explicit mirror opt-out

- Adding a config key, marker file, or flag by which a user declares "do not maintain a mirror in
  this project" (§3.5 accepted cost). Deletion is not read as that declaration, and no new signal is
  introduced here.

### Out of Scope — changing the early return

- Removing, relocating, or conditionalizing the version-match early return at
  `update_template_sync.go:617-622`.
- Changing the `(skipped bool, err error)` contract of `runTemplateSyncWithProgress` or the
  downstream `syncSkipped` block at `update.go:518`.

### Out of Scope — the diagnostic surface

- Adding, removing, or re-scoring any `moai doctor` check or finding. The only doctor change in
  scope is the textual reconciliation required by REQ-UMH-008.
- Making `inspectSkillMirror` (or any doctor code path) write, repair, or remove anything.

### Out of Scope — mirror lifecycle beyond restoration

- Removing mirror entries whose canonical skill was renamed or retired (`skill_mirror.go`'s header
  assigns that to the clean path).
- Registering mirror entries in the manifest (deliberately untracked; hashing a directory symlink
  fails EISDIR).
- Changing the symlink-vs-copy fallback behavior, or the relative link body.

### Out of Scope — this repository's own state

- Creating, restoring, or deleting `.agents/` in this repository (C-3).

## §8 Cross-references

- `SPEC-CODEX-MIRROR-DOCTOR-001` (card t498) — the reporting surface; boundary at C-4.
- `SPEC-CODEX-COMMAND-SKILLS-001` (card t503) — introduced Path B; §1 records why it did not fix
  this defect.
- `SPEC-CODEX-WIRING-001` (REQ-CW-009) — the existence-gate contract whose *placement* precedent
  (`update.go:507`) this SPEC reuses, and whose *predicate* it deliberately does not (§3.3).
