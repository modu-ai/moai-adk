# SPEC-UPDATE-MIRROR-HEAL-001 — Implementation Plan

Card: t520 · Branch: `WT-update-mirror-heal` · Base: `origin/develop` = `0b1e27877`
Evidence path: `.moai/reports/t520/`

## §A Context

`moai update` on a version-matched project skips `Deploy` entirely
(`internal/cli/update_template_sync.go:617-622`), and both `.agents/skills` producers live inside
`Deploy`. A deleted mirror is therefore permanent for that project. Full grounding: `spec.md` §1.

## §B Known issues and inherited claims

- **B1 — inherited, not measured here.** The 24 → 0 → 37 scratch-project sequence is card t498's
  measurement. It is cited as prior evidence and is never presented as this card's own.
- **B2 — coordinate correction on the record.** The card's `internal/cli/update/update.go:509` does
  not exist on this tree; the real cause is `update_template_sync.go:617-622`, and
  `update.go:518`'s `syncSkipped` block is downstream. See `spec.md` §1.
- **B3 — gate coverage of the inherited case is unestablished.** t498's run records neither the
  scratch project's `--agent` value nor its `template_version` stamp. M1's reproduction constructs
  the stamped precondition explicitly rather than inheriting that assumption.
- **B5 — both intake-proposed gates are refuted.** `.agents/` fails on deletion-survival; `.codex/`
  fails because `init.go:191-193` gives a `--agent claude` project no `.codex/` while the mirror is
  created unconditionally. The chosen predicate is the deploy-version stamp (`spec.md` §3.5), and
  the rejections are recorded in `spec.md` §3.2-§3.4 so they are not re-walked.
- **B4 — doctor text goes stale on landing.** `doctor_codex.go:457-461` asserts a routine update
  does not restore the mirror. True today; false after M3. Handled in M4.

- **B6 — git cannot witness the `.agents/` axis.** `.agents/` is gitignored
  (`git check-ignore -v .agents/skills/foo` → `.gitignore:133:.agents/`, rc 0; control
  `internal/cli/update.go` → rc 1), so any `git diff -- .agents` reads empty even after a create.
  All C-3 evidence is filesystem snapshots (`acceptance.md` §D.0). The wider `.gitignore` divergence
  belongs to card t522 and is deliberately not pursued here.
- **B7 — the gate's converse is false and is accepted, not fixed.** `system.yaml` is written inside
  the walk while `mirrorSkills` runs only after both walk guards, and init swallows deploy failure
  (`initializer.go:225-230`), so a stamped-but-never-mirrored project can exist. Branch analysis and
  the decision: `spec.md` §3.6. Boundary pinned by AC-UMH-015/016.

## §C Pre-flight

- Confirm the two mirror producers and the early return still sit where `spec.md` §1 places them
  (`git grep -n mirrorSkills internal/template/deployer.go`).
- Confirm no production caller disables mirroring (`grep -rn "WithSkillMirror" internal | grep -v _test`).
- Capture the C-3 baseline snapshot **before the first edit**:
  `find .agents ... > .moai/reports/t520/agents-before.txt` per `acceptance.md` AC-UMH-013 (a
  snapshot taken afterwards witnesses nothing).

## §D Constraints

C-1 no creation in unwired projects · C-2 early return preserved · C-3 isolated reproduction only ·
C-4 report/repair boundary · C-5 fail-open. Bodies in `spec.md` §5.

## §E Self-verification

Package-scoped tests only: `go test ./internal/cli/... ./internal/template/...`. No local full
suite (`CLAUDE.local.md` §4/§6); the full-suite verdict comes from CI on the develop push.
Evidence exported to `.moai/reports/t520/` before citation.

## §F Milestones

Ordered by decision-reversibility: the two open design decisions and the new public seam come
first; the mechanical wiring and text reconciliation come last.

### M1 — Gate predicate: fix the mirror-introducing version constant

The single decision the rest of the card rests on. Introduce one exported constant naming the first
release whose deploy creates a mirror, with its value fixed from the release record
(`CHANGELOG.md` `[3.1.3] - 2026-08-24`, "A skill mirror at `.agents/skills`"; `skill_mirror.go`
added `9c94c6b7a`, 2026-08-22) rather than guessed, and a comment citing both. Comparison reuses
`compareVersionLoose` (`internal/cli/update_version.go:234`) rather than a new comparator.

Surfaced for review: whether the constant lives in `internal/cli` next to the gate or in
`internal/template` next to the producer. The producer is the thing whose behavior it describes.

### M2 — Reproduction harness (RED) and the fixture matrix

A failing `internal/cli` test that builds a project under `t.TempDir()` **by running the deploy
path** (so the stamp and the mirror are products of the same run — never a hand-written stamp,
which is how 0.2.0 bypassed the §3.6 question), records the mirror entry set, deletes `.agents`,
runs the update path at version match, and asserts restoration. Fails on the current tree.

Fixtures, all in this milestone because they are one helper's parameter space:

| Fixture | Pins |
|---|---|
| deployed, then `.agents` deleted | AC-UMH-001/002 |
| stamp below the constant / stamp absent | AC-UMH-003/004 |
| healthy mirror | AC-UMH-006 |
| non-symlink occupant | AC-UMH-007 |
| extra locally-authored skill | AC-UMH-008 |
| stamped-but-partial (stamp written, walk failed, no skills) | AC-UMH-015/016 — the §3.6 boundary |
| `3.1.3` / `v3.1.3` / `3.2.0-rc.0` / `3.1.2` / `dev` / absent | AC-UMH-017 |

Also in this milestone: the snapshot helper AC-UMH-003/004/006/013 share (relative path + type +
size + symlink target). It must be shown able to see a change at all — its control is stated in
each criterion.

### M3 — Repair seam and its scope set

The new function: given a project root, restore the Path A symlink entries for the
**template-shipped** skill names present under `.claude/skills` (`spec.md` §4 S2), reusing the
existing `mirrorOneSkill` semantics so the non-symlink-occupancy skip (REQ-UMH-005) and the
idempotent already-correct branch come for free rather than being reimplemented. The scope set is
built in two steps, and both are mandatory:

1. **Derive** the candidate names from the embedded template FS — not from a directory listing of
   `.claude/skills`, which would readmit S1's locally-authored entries (REQ-UMH-007).
2. **Filter** the candidates by canonical-target existence before linking — drop any name whose
   `.claude/skills/<name>` is absent (REQ-UMH-010).

The 0.3.0 text stopped after step 1, which contradicted §3.6/AC-UMH-015: read literally it forbade
the intersection, and the producer supplies none of its own (`os.Symlink` succeeds against a missing
target, and `srcDir` is stat-ed only on the copy path). That conflict is resolved **in favour of
§3.6/AC-UMH-015** — derivation source and existence filter are different steps, so neither principle
is given up — because the alternative reading produces 16 dangling links on the partial-deploy
fixture, and a dangling entry is damage rather than repair (it is a `moai doctor` finding in its own
right). Reusing `mirrorOneSkill` for the per-entry semantics stays correct; the filter sits above it,
in the repair pass, not inside the shared producer.

Decision surfaced for review: whether the seam is an exported `template` package entry point taking
an explicit skill set, or an option on the existing deployer. The former keeps `Deploy` untouched.

**Run-phase note 1 (from plan-audit iteration 3).** The dangling-entry hazard REQ-UMH-010 addresses
is **Path A only**. Path B files are ordinary template writes that never touch the mirror producer,
so no existence filter applies to them — do not extend the filter to Path B on the assumption that
the two paths share a failure mode.

**Run-phase note 2 (from plan-audit iteration 3) — decide, do not discover.** If the M3 seam is
implemented as a *deployer option*, that same option is also live on the **deploy-time** path, where
`Deploy` already runs the producer. Choosing that shape therefore changes deploy behavior as well as
repair behavior. Make that a deliberate decision with its deploy-side effect stated, or pick a shape
whose blast radius is the repair pass alone.

### M4 — Path B restoration and call-site wiring

Restore the 16 template-published `.agents/skills/moai-*/SKILL.md` files from the embedded FS,
restore-missing-only (never overwriting an existing file at a published-skill path — the
`ProtectedSkips` rule already encoded in `DeployResult`). Wire the whole pass as a best-effort,
existence-gated call beside `refreshCodexWiringBestEffort` at `internal/cli/update.go:507` —
before the `syncSkipped` return, leaving the early return at `update_template_sync.go:617-622`
untouched (C-2).

### M5 — Doctor guidance reconciliation

Update the mirror-absent detail text at `doctor_codex.go:457-461` so it no longer asserts that a
routine update cannot restore the mirror (REQ-UMH-008). Textual only; no check, finding, or score
changes (C-4). Note the coupling to `doctor_codex_test.go`'s `testMirrorRedeployDirective` /
`testMirrorDetailPhrase` anchors — both may need updating in step.

### M6 — Template-First mirror check

Confirm whether any file changed under `internal/template/templates/**`; if so, `make build` and
verify the emitted artifacts, per `CLAUDE.local.md` §2. Expected: no template-source change (this
card is Go-side only), and the milestone exists to make that expectation checked rather than
assumed.

## §G Anti-patterns

- Deleting or relocating the early return "so `Deploy` just runs" — violates C-2 and re-imposes a
  full deploy on every up-to-date update.
- Gating on `.agents/` existence — closed exactly when repair is needed (`spec.md` §3.2).
- Gating on `.codex/` wiring — closes repair permanently for claude-only projects, which are the
  majority mirror-holding population (`spec.md` §3.3).
- Gating on the manifest — mirror entries are untracked by contract, and the witness measures 0
  against a 436 control (`spec.md` §3.4 R3).
- Repairing in any MoAI project — creates `.agents/` for the pre-3.1.3 population that never had
  one (`spec.md` §3.4 R4).
- Mirroring every directory under `.claude/skills` — invents entries no deploy produced
  (`spec.md` §4 S1).
- Reproducing the defect by deleting `.agents` in this repository (C-3).
- Making the doctor repair, or the repair report — collapses the C-4 boundary.

## §H Cross-references

`spec.md` · `acceptance.md` · SPEC-CODEX-MIRROR-DOCTOR-001 · SPEC-CODEX-COMMAND-SKILLS-001 ·
SPEC-CODEX-WIRING-001 (REQ-CW-009).
