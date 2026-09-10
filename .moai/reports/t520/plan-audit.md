# SPEC Review Report: SPEC-UPDATE-MIRROR-HEAL-001

Card: t520 · Branch: `WT-update-mirror-heal` · Tree audited: `0b1e27877` (worktree `.claude/worktrees/t520`)
Iteration: 1/2 (Tier M ceiling)
Auditor: plan-auditor (Claude-only; no `audit_model` key present under `.moai/config/` — `grep -rn "audit_model" .moai/config/` → no output, so no cross-model fan-out was run)

**Verdict: FAIL**
**Overall Score: 0.81** (Tier M PASS threshold 0.80)

The aggregate score alone clears the Tier M threshold. The FAIL is driven by three
**blocking**-classified findings (D1-D3), not by the score and not by the volume of the optional
list. Stated explicitly so the routing is unambiguous: fix D1-D3, then the verdict is revisited on
that delta.

Reasoning context ignored per M1 Context Isolation. The spawning prompt supplied a list of
already-verified code facts; every fact this report leans on was re-measured in this session
against this tree, and where my measurement diverges from the supplied one it is recorded as such
(see D4, D7).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -n '^\*\*REQ-'` over `spec.md` → REQ-UMH-001..009 at
  L69, 74, 78, 82, 86, 90, 94, 98, 104. Sequential, no gaps, no duplicates, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — judged against the nine `REQ-UMH-*`
  entries in `spec.md` §2, **not** against the `AC-UMH-*` entries in `acceptance.md` (the
  verification layer is Given-When-Then by design and is graded in Group 4). REQ-001 is the GEARS
  compound form (`spec.md:70-72`, "Where the project's recorded deploy version … When `moai update`
  completes its template-sync step … shall run"); REQ-002/005/007/009 are canonical negatives
  ("shall not create, modify, or remove any path under `.agents/`", `spec.md:76`); REQ-003/007 are
  ubiquitous; REQ-004/008 event-driven; REQ-006 state-driven ("While the mirror-repair pass
  encounters a failure …", `spec.md:91`). No informal modality, no Given/When/Then presented as a
  requirement.
- **[PASS] MP-3 YAML frontmatter validity** — `spec.md:1-15` carries all 12 canonical fields with
  correct types: `id` (matches the domain-namespaced pattern), `title` quoted,
  `version: "0.2.0"` quoted semver, `status: draft` (enum), `created`/`updated` ISO `2026-09-07`,
  `author`, `priority: P1`, `phase: "v3.2.0 target"` (a release target — **not** one of the
  prohibited lifecycle tokens `plan`/`run`/`sync`/`mx`), `module: "internal/cli"`,
  `lifecycle: spec-anchored`, `tags` comma-separated string. Plus optional `tier: M`. No rejected
  snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) appears.
- **[N/A] MP-4 language neutrality** — single-language SPEC. `module: "internal/cli"`; the subject is
  Go code in this repository's own binary, not template-bound or multi-language tooling. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — verb executed. Referenced SPEC IDs resolved and
  their `status:` read: `SPEC-CODEX-MIRROR-DOCTOR-001` → `completed`;
  `SPEC-CODEX-COMMAND-SKILLS-001` → `completed`; `SPEC-CODEX-WIRING-001` → `completed`;
  `SPEC-CODEX-SKILLS-CANONICAL-001` (named via commit `9c94c6b7a`'s subject) → `completed`. None is
  `retired` / `superseded` / `archived`, so no reconciliation clause is owed. No referenced SPEC is
  missing from `.moai/specs/`.
- **[N/A] MP-6 D8 cross-platform discipline** — verb executed:
  `grep -rn "syscall" .moai/specs/SPEC-UPDATE-MIRROR-HEAL-001/` → no output. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-UPDATE-MIRROR-HEAL-001/`
  → no output. `plan.md` exists; `research.md` is absent (Tier M does not carry one), so the sweep
  ran over the artifacts that exist.

Tier budget: 9 requirements ≤ 16 and 14 acceptance criteria ≤ 16 — both within the Tier M ceilings,
which apply independently.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | Requirements read unambiguously. Two premises do not: `spec.md:209-210` describes `deployedSkills` as "the set the walk actually wrote this run", which the code contradicts in an explicit comment (D4) — a reader could implement wrote-vs-saw semantics differently; and K4's label at `spec.md:120` ("Excludes the never-had-a-mirror population") is not the predicate R5 actually tests (D8). |
| Completeness | 1.00 | 1.0 | HISTORY (`spec.md:19-24`), Context/WHY (§1), WHAT (§2 requirements + §3 decision), HOW (`plan.md` §F M1-M6), acceptance (`acceptance.md`, 14 criteria), and five `### Out of Scope — <topic>` H3 sub-headings at `spec.md:247, 253, 260, 266, 274`, each carrying specific `-` bullets. Frontmatter complete (MP-3). |
| Testability | 0.75 | 0.75 | Every criterion names a command and an expected observable (`acceptance.md:7-10` fixes the convention). One criterion is nonetheless incapable of failing (AC-UMH-013, D1) and two more are green on the pre-implementation tree with no mutant assigned (AC-UMH-005 clause 1, AC-UMH-006 — D5). The AC-UMH-014 mutation probe is real and correctly scoped for the four guards it does name. |
| Traceability | 0.75 | 0.75 | All nine REQs are covered: 001→AC-001, 002→AC-003/004, 003→AC-005, 004→AC-002, 005→AC-006/007, 006→AC-009, 007→AC-008, 008→AC-011, 009→AC-013. Three ACs trace to non-REQ anchors instead: AC-UMH-010 → "§3.5 K4", AC-UMH-012 → "C-4", AC-UMH-014 → "verification integrity" (`acceptance.md:25, 27, 29`). Each is a legitimate criterion, but none names a `REQ-XXX`. |

Aggregate = arithmetic mean of the four = 0.8125 → 0.81.

---

## Defects Found

**D1. AC-UMH-013 cannot fail — the only mechanical guard on constraint C-3 is vacuous.**
`acceptance.md:137-143` — Severity: **critical** — Class: **blocking**.
The criterion verifies REQ-UMH-009 / C-3 ("this repository's `.agents/` is untouched") with
`git diff --name-only 0b1e27877..HEAD -- .agents` → expected no output. `.agents/` is gitignored at
this repository's root, so nothing the run phase does to it can ever appear in that diff:

```
$ git check-ignore -v .agents/skills/foo
.gitignore:133:.agents/	.agents/skills/foo     (exit 0)
$ ls -d .agents
ls: .agents: No such file or directory
```

`.gitignore:133` is a bare `.agents/` rule; only `internal/template/templates/.agents/` is
re-included (`.gitignore:137`). The repository additionally has no root `.agents/` at all, so the
pathspec matches nothing today and would still match nothing after a violation. The stated control
(`git diff --name-only 0b1e27877..HEAD | wc -l` → non-zero) is a control **inside the same cut**: it
establishes the diff is non-empty, never that this pathspec is capable of matching. And the SPEC's
own vacuity brake does not reach here — AC-UMH-014's mutant list (`acceptance.md:146-149`) names
AC-003, AC-004, AC-007, AC-012 and omits AC-013.
*What would settle it*: run the guard's own mutant — create `.agents/skills/x` at the repository
root, then run the verification command. It will print nothing, i.e. report PASS while the forbidden
thing has happened.
*Required fix*: verify C-3 by filesystem, not by `git diff`. Record `find .agents -print 2>&1 | sort`
(expected: `find: .agents: No such file or directory`) before the first run-phase commit and again at
close, and assert the two are identical; keep a `git status --porcelain --ignored -- .agents`
reading alongside it. Then add AC-013 to the AC-UMH-014 mutant set with the mutant above.

**D2. The gate's load-bearing premise is asserted, not tested — a ≥3.1.3 stamp does not imply a
mirror-creating deploy ever completed.**
`spec.md:180` (K2 row) and `spec.md:182-187` (K4 row) — Severity: **major** — Class: **blocking**.
The forward claim verifies. `template_version` is written by the ordinary template walk —
`internal/template/templates/.moai/config/sections/system.yaml.tmpl:9` renders
`template_version: "{{.Version}}"` from `WithVersion(version.GetVersion())`
(`internal/core/project/initializer.go:414`), deployed and manifest-tracked by the same walk that
mirrors (`internal/template/deployer.go`, `m.Track(destRelPath, manifest.TemplateManaged, …)`), with
no `--agent` conditional anywhere on the config sections. So "written on every deploy regardless of
`--agent`" is TRUE as measured.

The property the gate needs is the **converse**, and it is false on a reachable path.
`mirrorSkills` runs only after the walk returns cleanly — in `DeployWithResult` the
`if walkErr != nil { return result, walkErr }` and `if deployErr != nil { return result, deployErr }`
guards both precede the `if !d.skillMirrorDisabled { result.SkillMirrors = d.mirrorSkills(...) }`
block. Meanwhile `internal/core/project/initializer.go:226-230` treats a failed `deployTemplates` as
**non-fatal**: it appends a warning and continues. A deploy that writes `system.yaml` early in the
walk and errors later therefore leaves a project stamped ≥ 3.1.3 that has never held a mirror —
precisely the population C-1 says must not have `.agents/` created for it. R5 will create it there.
No criterion exercises this: AC-UMH-001 hand-writes `system.yaml` under `t.TempDir()`
(`acceptance.md:34-36`), which tests the gate's arithmetic while bypassing the question of what a
real deploy leaves behind.
*What would settle it*: a test that deploys with an injected mid-walk failure after
`.moai/config/sections/system.yaml` has been written, then asserts (a) the stamp is present and
≥ 3.1.3 and (b) `.agents/` is absent — and then states which side of C-1 that project is on.
*Required fix*: either record it in §3.5 as a second accepted cost, with the same explicitness the
deliberate-deletion cost already gets, or add an AC. Do not leave K2/K4 reading as unconditional
properties; K2's sentence is currently a premise claim carrying no measurement.

**D3. The `v`-prefix and degradation paths are load-bearing and unexercised.**
`acceptance.md:33-67` (AC-001/003/004) against `spec.md:171-191` — Severity: **major** — Class:
**blocking**.
This repository's own stamp is `v`-prefixed — `.moai/config/sections/system.yaml:45` reads
`template_version: v3.1.3` — and §3.5's degradation argument rests entirely on the comparator
tolerating that. The comparator does:
`internal/cli/update_version.go` `stripVersionPrefix` strips `go-v` then `v`, and `parseSemverLoose`
truncates each segment at the first `-` or `+` and stops at the first non-digit. So `v3.1.3` compares
equal to `3.1.3`, and `3.2.0-rc.0` compares as `3.2.0`. None of that is asserted anywhere. The three
gate fixtures are: stamp equal to the binary (AC-001), stamp below the constant (AC-003), stamp
absent (AC-004). A prefixed stamp, a pre-release stamp, and a non-numeric stamp are all absent.
The third matters most: `parseSemverLoose("dev")` yields `[0,0,0]`, so a dev-built binary's project
closes the gate — the safe direction, but currently an unasserted accident rather than a stated
property.
*What would settle it*: table-driven cases over `{"v3.1.3", "3.1.3", "3.1.3-rc.1", "v3.2.0-rc.0",
"3.1.2", "dev", ""}` asserting open/closed per input.
*Required fix*: extend AC-UMH-003/004 (or add AC-UMH-015) with that table, and name `dev`/unparseable
→ closed as an intended property rather than leaving it emergent.

**D4. `spec.md` §4 states the opposite of what the code says about the mirror set.**
`spec.md:209-210` — Severity: **major** — Class: **optional** (the conclusion survives; the premise
does not).
§4 opens: "`mirrorSkills` is driven by `deployedSkills`, the set the walk actually wrote this run
(`internal/template/deployer.go:152-222`)". The code records the skill name **before** the skip
branch, and says so in a comment written for exactly this misreading: "Record the skill this file
belongs to, before any skip branch: a re-deploy whose files all already exist still owes its mirror."
The existing-file protection sits ~30 lines further down, at the `protectedScope :=
isPublishedSkillPath(destRelPath)` line. The set is therefore "skills the walk **saw**", not "wrote".
The S2 choice happens to be exactly what the deployer does, so the decision is right — but an
implementer reading §4 literally could reproduce wrote-this-run semantics and under-restore on a
project whose skills are all already present.
*Required fix*: correct the sentence and the line citation, and note that S2 is chosen *because* it
matches the deployer's saw-not-wrote behavior rather than in spite of it.

**D5. Two criteria are green on the pre-implementation tree with no mutant assigned.**
`acceptance.md:69-79` (AC-005) and `acceptance.md:80-85` (AC-006) — Severity: **minor** — Class:
**blocking** (cheap to fix; it is the same hazard AC-014 was written for).
AC-UMH-005's first clause reads `internal/cli/update_template_sync.go` and asserts the version-match
guard and its `return true, nil` are present — true right now, before any implementation, so that
clause can never be red. Its second clause (no template write, mirror still repaired) is the part
that goes red, and the criterion does not say so. AC-UMH-006 (healthy project is a byte-identical
no-op) is likewise satisfied by a repair pass that does nothing at all. Per
`.claude/rules/moai/development/verification-completeness.md` §2, each needs a RED-now cell stating
which clause is red and why.
*Required fix*: split AC-005 into a preservation clause explicitly classified as a regression guard
and a behavioral clause carrying the RED; add AC-006 to the AC-UMH-014 mutant set (mutant: make the
repair pass rewrite existing entries unconditionally — the snapshot must flip red).

**D6. §3.4's R3 manifest measurement is under-determined; its adjacent `Track` claim is unverified in
the SPEC (and is in fact true).**
`spec.md:159-163` — Severity: **minor** — Class: **optional**.
The witness `grep -c '\.agents/skills' .moai/manifest.json` → `0` against a `.claude/skills` control
of `436` is presented as establishing that mirror entries are untracked. This repository has **no
root `.agents/` directory at all** (`ls -d .agents` → `No such file or directory`), so the `0` is
equally explained by "no mirror was ever deployed into this tree" — a control inside the same cut.
R3's rejection still stands on K3 (Path A entries are untracked by contract, per `skill_mirror.go`'s
header), so the decision is unaffected. Separately, §3.4 asserts flatly that Path B files "do flow
through `Track`". I verified it: every written file, published-skill paths included, reaches
`m.Track(destRelPath, manifest.TemplateManaged, templateHash)` in `deployer.go`. Record it as
measured rather than asserted.

**D7. §3.3's grep tally is internally inconsistent.**
`spec.md:137-140` — Severity: **minor** — Class: **optional**.
"`grep -rn "WithSkillMirror" internal` returns three hits — the option definition, its doc comment,
and a comment at `deployer.go:91` — with the remaining matches in tests." Three hits cannot have
remaining matches. Measured: 5 hits — `deployer.go:91` (comment), `skill_mirror.go:122` (doc
comment), `skill_mirror.go:125` (definition), `skill_mirror_fallback_test.go:234` and `:338` (test
call sites). The substance — no production caller, so mirroring is on for every `Deploy` — verifies.

**D8. K4's label is not the predicate R5 tests, and the ≥-side boundary has no criterion.**
`spec.md:120` and `spec.md:193-196` — Severity: **minor** — Class: **optional**.
K4 is written as "Excludes the never-had-a-mirror population". R5 tests something narrower: "was last
deployed by a mirror-capable binary". A project stamped ≥ 3.1.3 that legitimately holds no mirror is
*created* for, not *restored* to — which is almost certainly the wanted behavior, and is what D2's
partial-deploy case lands in. §3.5's "Why this is not R4 in disguise" paragraph gestures at the
distinction without resolving the label, and AC-003/004 only cover the below/absent side.
*Required fix*: restate K4 as "excludes projects whose last deploy predates mirror creation", and say
in one sentence that a ≥-constant project without a mirror is deliberately repaired.

**D9. `plan.md` M4 mis-attributes restore-missing-only to an existing rule.**
`plan.md:84-86` — Severity: **minor** — Class: **optional**.
M4 says the never-overwrite behavior is "the `ProtectedSkips` rule already encoded in
`DeployResult`". The deployer's rule is manifest-provenance-driven: at the `protectedScope` branch an
existing file that is tracked `UserModified`/`UserCreated` is skipped, an **untracked** existing file
is recorded `UserCreated` and skipped, and a `TemplateManaged` one is overwritten. Restore-missing-only
is strictly stricter than that. The stricter rule is the correct reading of REQ-UMH-005 (a Path B
`SKILL.md` is a regular file and is never "a MoAI-created symlink"), so the design is right — but the
plan should present it as a new rule resembling the existing one, not as reuse. Note the consequence
worth stating in §7: under that rule a stale or corrupted published `SKILL.md` is never refreshed by
the repair pass.

**D10. §1 names one of two identical early returns.**
`spec.md:42-46` — Severity: **minor** — Class: **optional**.
`update_template_sync.go:617-622` is correctly identified as the causal return on the `moai update`
path (`update.go:494` is the sole call site of `runTemplateSyncWithProgress`). That function then
delegates at its tail to `runTemplateSyncWithReporter`, which carries its **own** version-match early
return at `update_template_sync.go:95-105` (`packageVersion == projectVersion && !forceBackup` →
`return nil`). On the update path the outer one fires first, so the inner is unreachable there; but
AC-UMH-005 pins only the outer, and a future edit could remove the pinned guard while the unpinned one
preserves the behavior. Worth one sentence in §1 and one clause in AC-005.
Note also, because it bears on §3.5's equality argument: the early return compares with **exact string
equality**, not `compareVersionLoose`. That is what makes "the stamp is equal to the running binary's
version by construction" true — the SPEC's claim is sound — but it also means the gate predicate
(loose) and the skip predicate (exact) are different comparators over the same field, which D3's
prefixed-stamp fixtures should pin.

---

## Constraint coverage (operator-named C1-C4)

| Constraint | Encoded as | Testable AC | Verdict |
|---|---|---|---|
| C1 existence-gate preserved (repair, not creation) | REQ-UMH-002 (`spec.md:74-76`), C-1 (`spec.md:222-223`) | AC-UMH-003, AC-UMH-004, both with pre/post walk snapshots and both in the AC-014 mutant set | **PASS on the below/absent side**; the ≥-side boundary is uncovered (D2, D8) |
| C2 early return not deleted or relocated | REQ-UMH-003 (`spec.md:78-80`), C-2 (`spec.md:224-226`), anti-pattern (`plan.md:107-108`) | AC-UMH-005 | PASS, with the RED-clause defect at D5 and the second-return gap at D10 |
| C3 reproduction isolated; repo `.agents` untouched | REQ-UMH-009 (`spec.md:104-106`), C-3 (`spec.md:227-229`), `plan.md:34`, `acceptance.md:3-5` | AC-UMH-013 | **FAIL — the criterion cannot fail (D1)** |
| C4 doctor REPORTS vs this card REPAIRS | C-4 (`spec.md:230-233`) states it explicitly and names both sides | AC-UMH-012 (doctor stays read-only, in the mutant set) + AC-UMH-011 (text reconciliation, with a control) | PASS. Verified the target text exists: `internal/cli/doctor_codex.go` `codexMirrorObservations` renders "…a routine `moai update` on a version-matched project does not restore it…", so AC-011's `grep -c "does not restore it"` → `0` is a real post-condition and its `"mirror absent"` control is a live selector |

## Author-declared gaps — adjudication

1. **t498's unrecorded scratch-project preconditions.** Genuinely bounded. `plan.md:14-21` (B1, B3)
   refuses to present the 24→0→37 sequence as this card's evidence, and M2 (`plan.md:63-70`)
   constructs the stamped precondition explicitly. It really does construct it — AC-UMH-001 writes
   `system.yaml` under `t.TempDir()` rather than inheriting. The residue is D2: constructing the
   stamp by hand is what leaves the stamp's real provenance untested.
2. **Deliberate deletion indistinguishable from the failure mode.** A legitimate scope boundary. The
   cost is stated in the SPEC's own voice (`spec.md:198-201`), the opt-out is named as out of scope
   with a topic heading and a specific bullet (`spec.md:247-251`), and no requirement silently
   depends on the opt-out existing. Not an unshipped requirement.
3. **Seam shape and constant's home package left open.** Correctly bounded — both are surfaced for
   review inside the milestones that own them (`plan.md:60-61`, `plan.md:79-80`) and neither is load
   bearing for any AC. Note that M1's framing ("fix the constant … rather than guessed") is already
   settled by AC-UMH-010, which fixes the value at `3.1.3`; the milestone's remaining freedom is the
   package, not the value.
4. **Path B manifest behavior inferred.** Partially. See D6: the blindness caveat is hedged, but "do
   flow through `Track`" is stated flatly. Nothing load-bearing rests on it — R3 is rejected on K3,
   which is contract, not measurement — so the exposure is a wording defect rather than a design one.

## Boundary claims I verified independently

- **3.1.3 is the right release boundary.** `9c94c6b7a` is real, dated **2026-08-22**, subject
  `feat(SPEC-CODEX-SKILLS-CANONICAL-001): M1 derive the skill mirror set from the run`.
  `CHANGELOG.md` `## [3.1.3] - 2026-08-24` carries "**A skill mirror at `.agents/skills`** for
  codex-cli, derived from the run rather than hand-listed". The preceding release is
  `## [3.1.2] - 2026-08-21`, which predates the commit. The mirror could not have shipped in 3.1.2,
  and it is named in 3.1.3 — no off-by-one.
- **Repair placement.** `internal/cli/update.go:507` is `refreshCodexWiringBestEffort(out, …)`, and
  the `if syncSkipped { … return nil }` block begins at `:514`. The precedent sits where the SPEC
  says, before the skip return, and `update.go:494` is the only call into the sync path — so a pass
  wired there runs on every `moai update`, `--templates-only` included (that flag only skips the
  binary update, `update.go:153`).
- **Mirroring is unconditional.** `DeployWithResult` runs `mirrorSkills` whenever
  `!d.skillMirrorDisabled`, and the only non-test reference to the disabling option is its own
  definition and two comments (D7). No `--agent` value reaches it.

## Recommendation

FAIL. Fix order:

1. **D1** — replace AC-UMH-013's `git diff` verification with a filesystem pre/post assertion, and add
   AC-013 to AC-UMH-014's mutant set. This is the cheapest and the most important: C-3 is an
   operator-named constraint whose only mechanical guard currently reports PASS in the presence of a
   violation.
2. **D3** — add the comparator fixture table (prefixed / pre-release / non-numeric / empty). One AC,
   table-driven, and it closes the degradation path the gate's safety argument depends on.
3. **D2** — decide and record: is a ≥ 3.1.3 project with no mirror repaired deliberately (then say so
   in §3.5 and fix K4's label per D8), or is the partial-deploy stamp a case the gate must exclude
   (then a criterion is owed)? Either answer is defensible; leaving K2/K4 as unconditional properties
   is not.
4. **D5** — split AC-UMH-005's clauses and give AC-UMH-006 a mutant. Small, and it makes the SPEC's
   own vacuity discipline complete rather than selective.

D4 and D6-D10 are optional: surface them to the orchestrator and let it decide. D4 is the one I would
still fix even as optional — a false premise about `deployedSkills` sits directly upstream of the M3
implementation decision, and correcting it costs one sentence.

Not a defect, recorded so it is not re-litigated: the `Given … When … Then …` shape of every
`AC-UMH-*` entry is the correct verification-layer format and was graded under Group 4, never against
GEARS. The GEARS judgment in MP-2 was made against the `REQ-UMH-*` layer in `spec.md` §2 only.

---
---

# Round 2 — re-audit of the iteration-1 repair

Artifacts re-read at `spec.md` 0.3.0 / `acceptance.md` 0.2.0 / revised `plan.md` / revised
`progress.md`. Same tree: `0b1e27877`. Round 1 above is preserved unaltered — the audit history is
itself evidence.

**Verdict: FAIL**
**Overall Score: 0.84** (Tier M threshold 0.80) — up from 0.81; no score regression, so no STOP
signal fires.

Iteration 2 of 2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`). A FAIL here is the
ceiling, so it escalates: the orchestrator owes the operator a choice between PASS-with-debt
(accept the findings below as documented debt and enter run phase), scope-reduction, or an explicit
override to iterate further. My own recommendation is at the end.

Reasoning context ignored per M1 Context Isolation. The coordinator's message summarised what the
author reports having repaired; none of it is treated as fact. Every claim below was read out of the
files or measured against the code in this session.

## Round-1 findings — disposition

| R1 finding | Status | Evidence read this session |
|---|---|---|
| **D1** vacuous C-3 guard | **CLOSED** | `acceptance.md:163-176` replaces `git diff` with a filesystem snapshot, and `acceptance.md:195` adds the AC-013 mutant row ("create `.agents/skills/mutant-probe` in **this** repository … it MUST report a difference"). I walked the mutant: with `.agents` absent both snapshots are empty; after the probe the directory exists, so the `test ! -e .agents` clause fails and the guard flips red. The guard can now fail. `§D.0` rule 1 (`acceptance.md:17-22`) documents the git form as rejected evidence, and I re-measured its basis — `git check-ignore -v .agents/skills/foo` → `.gitignore:133:.agents/` rc 0; control `git check-ignore -v internal/cli/update.go` → rc 1. Both readings in `§D.0` are correct. Carry-over defect in the *new* form: N3. |
| **D2** gate converse asserted, not tested | **ADDRESSED** — the analysis is real and its citations hold | New `§3.6` evaluates both branches. I verified every load-bearing citation: `deployer.go:286` `if walkErr != nil`, `:289` `if deployErr != nil`, `:297-298` the mirror call — the ordering claim is exact. `initializer.go:225-230` swallows a failed `deployTemplates` into `result.Warnings` — verbatim. Step 3 is at `:220`, `initManifest` is invoked at `:319-323` — so the manifest step is genuinely downstream of the swallow and is not a completion witness. `grep -rn "SkillMirrors" internal --include='*.go' \| grep -v _test.go` → **exactly 7 hits**, all producer (`deployer.go:298`), display (`mirrornotice/notice.go:55`), or accessor/field (`skill_mirror.go:67,68,84,98,111`). Branch (가)'s rejection ground — no persisted witness exists — is sound. **But the ground the acceptance rests on is now the defect: see N1.** |
| **D3** version degradation untested | **CLOSED** | `acceptance.md:224-243` AC-UMH-017, six rows. Each row matches the comparator I read in round 1: `stripVersionPrefix` strips `go-v`/`v`; `parseSemverLoose` truncates at `-`/`+` and stops at the first non-digit, so `dev` → `[0,0,0]`. `3.2.0-rc.0` → open and `dev` → closed are both correct as written, and the criterion says explicitly that it converts emergent behavior into asserted behavior. |
| **D4** §4 premise false | **CLOSED, and correctly re-derived** | `spec.md` §4 now states the set is "template-shipped skill names the walk **encountered**", quoting the comment. I re-read `internal/template/deployer.go` and the comment is verbatim as quoted. The author's further claim — that this *strengthens* S2, because a membership-based producer means a repair pass must enumerate template names rather than observe writes — is correct as reasoning. It is also precisely what collides with the existence-intersection in N1: enumerating names is half the rule, and the artifacts now disagree about the other half. |
| **D5** AC-005/006 green-at-arrival | **CLOSED** | `acceptance.md:94-103` makes the behavioral half mandatory and says in the criterion body that "the source-read half alone would pass on a tree where the guard exists but is unreachable". `acceptance.md:190` gives AC-006 a mutant row. |
| D6 (R3 measurement under-determined) | **OPEN** | `spec.md` §3.4 unchanged — the `0` against `436` control still stands without noting that this tree has no root `.agents/` at all. Optional in round 1; still optional. |
| D7 (grep tally "three hits") | **OPEN** | §3.3 unchanged; the self-contradicting sentence is still there. Optional. |
| D8 (K4 label) | **OPEN — and escalated, see N4** | The K4 row still reads "an update must not **create** `.agents/` in a project no deploy ever gave one", which is now in direct contradiction with §3.6's chosen branch. |
| D9 (M4 ProtectedSkips attribution) | **OPEN** | `plan.md:113` unchanged. Optional. |
| D10 (second early return) | **OPEN** | §1 unchanged. Optional. |

Three of three blocking findings from round 1 are closed on the evidence. That is real progress and
I am recording it as such.

## New findings — introduced by the repair, never audited before

**N1. The repair's scope rule contradicts itself across three layers, and the property the §3.6
acceptance rests on has no requirement behind it.**
`plan.md:98-107` (M3) against `spec.md` §4 S2 and `acceptance.md:204-213` (AC-UMH-015) —
Severity: **critical** — Class: **blocking** — Origin: **new material**.

§3.6's second ground for accepting the stamped-but-partial population is that the pass is
self-limiting: "it mirrors only template-shipped skills that are **present under `.claude/skills`**,
so a deploy that failed before the skills subtree yields zero Path A entries — the repair cannot
invent skills the project does not have." AC-UMH-015 pins exactly that ("**zero** Path A mirror
entries are created, because the scope set (S2) is intersected with the canonical skill directories
that actually exist").

`plan.md` M3, the instruction the run phase actually implements from, ends: "**The scope set is
derived from the embedded template FS, not from a directory listing of `.claude/skills`.**" Read
literally, that forbids the intersection AC-015 requires. The sentence was plainly written to reject
S1 (do not list `.claude/skills` and mirror everything), but after the D4 correction the same
sentence now reads as rejecting the existence check too — and D4's own "strengthening" argument
("enumerate template-shipped names rather than observe writes") pushes it further that way.

This is not a wording quibble, because nothing else stops the failure. I read
`internal/template/skill_mirror.go` end to end: `mirrorOneSkill` computes `srcDir` and **never
stats it**. It `Lstat`s the *mirror* path, `MkdirAll`s the mirror dir, then calls
`d.symlink(want, mirrorPath)`. `os.Symlink` succeeds against a non-existent target, so a
template-shipped name with no canonical directory yields a **dangling** link reported as
`MirrorModeSymlink` — a silent success. The existence intersection is therefore not inherited from
the producer M3 says to reuse; it must be built by the caller, and M3 currently tells the
implementer not to build it.

Compounding it: no requirement carries the property. REQ-UMH-007 (`spec.md:95-97`) reads "shall
mirror only skills the template ships, and shall not create a mirror entry for a locally-authored
skill directory that no deploy produced" — that is the local-only direction. The other direction (a
template-shipped name whose canonical directory is absent) appears only in §4 prose, §3.6 prose, and
AC-UMH-015, which the §D matrix traces to "§3.6" rather than to a REQ. So the safety ground of the
card's biggest design concession is requirement-less.

*What would settle it*: on the AC-015 fixture, a pass built to M3's literal instruction creates 16
Path A dangling links; a pass built to S2 creates zero. The two artifacts predict different
observable outcomes for the same fixture, which is the definition of a contradiction rather than a
difference of emphasis.
*Required fix*: (a) rewrite M3's last sentence to state both halves — the name universe comes from
the embedded FS, and it is then intersected with the canonical directories that exist; (b) add the
second clause to REQ-UMH-007 so the property is a requirement, not only a criterion; (c) state in
M3 that `mirrorOneSkill` performs no source-existence check, so the intersection is the caller's
obligation and reuse does not supply it.

**N2. AC-UMH-014's "one mutant probes two guards" justification is false, and the coverage audit
that produced it miscounts its own matrix.**
`acceptance.md:251-261` (§D.2) and `acceptance.md:47` — Severity: **major** — Class: **blocking** —
Origin: **new material**.

Two defects, one mechanism.

*The count.* §D.2 states "the matrix marks 8 absence guards (003, 004, 006, 007, 008, 011, 012,
013)". The matrix marks **nine**: those eight plus AC-UMH-015, which carries **yes** at
`acceptance.md:47`. The corrective introduced for round-1 D1 is explicitly structural — "the mutant
list in AC-UMH-014 is derived from the 'Absence guard?' column … and the two MUST be checked against
each other" — and at its first application the derivation is off by one. A rule stated as mechanical
and then applied by hand with a wrong count has not replaced the hand-picking it was written to
replace.

*The justification.* §D.2 excuses the missing row: AC-015's "forbidden outcome (inventing entries
from the template list alone) is the same mutation as AC-UMH-008's scope widening, and one mutant
probing two guards is recorded here rather than duplicated." The two mutants are not the same; they
are opposite directions of the same intersection:

- AC-008's mutant (`acceptance.md:192`) is "widen the scope set to every directory under
  `.claude/skills` (S1)". On AC-015's fixture `.claude/skills` is empty or absent, so S1 widening
  yields **zero** entries and **AC-015 stays green**.
- AC-015's forbidden outcome is enumerating template names *without* intersecting existing
  directories. On AC-008's fixture that mutant excludes `local-only-skill` (it is not a template
  name), so **AC-008 stays green**.

Neither mutant flips the other guard. The shared-mutant claim is not a compression of coverage, it
is an absence of coverage with a reason attached — the same shape (a guard omitted from the mutant
list) that made round 1's D1 possible.
*Required fix*: correct §D.2's count to nine, and give AC-UMH-015 its own mutant row — "drop the
existence intersection from the scope set; the pass then creates dangling entries on the
stamped-but-partial fixture and AC-UMH-015 flips red". That mutant is also the direct probe for N1.

**N3. AC-UMH-013's snapshot command does not run on this platform, and the criterion's stated
control does not distinguish what it claims to distinguish.**
`acceptance.md:163-176` — Severity: **major** — Class: **blocking** (one-line fixable) — Origin:
**new material**.

The command is `find .agents -mindepth 0 -maxdepth 3 \( -type l -printf '%p L %l\n' -o -printf
'%p %y\n' \) 2>/dev/null | sort`. `-printf` is a GNU findutils primary. Stock macOS `find` rejects
it:

```
$ /usr/bin/find .agents -maxdepth 3 -printf '%p %y\n'
find: -printf: unknown primary or operator          (rc 1)
```

It appeared to work when I first ran it because `find` in this agent's shell is a function that
routes to a bundled `bfs` (`which -a find` shows the function ahead of `/usr/bin/find`). Outside
this shell — a plain developer terminal on darwin, or CI's `bash` — the primary fails, `2>/dev/null`
swallows the error, and both snapshot files come out empty. The evidence command then asserts
nothing, which is the hidden-stderr-produces-all-zeros hazard in a criterion written to remove a
different instance of the same hazard.

Second, the criterion claims the fallback distinguishes two things it does not: "so 'both files
empty' is distinguishable from 'the snapshot command silently failed'". `test ! -e .agents` tests
*directory presence*, not *command success*. With `.agents` absent and `find` broken, every reading
is byte-identical to the healthy case — the two states are not distinguished. What the clause
actually buys is that a violation which creates `.agents` is still caught (by `test ! -e`, not by
the snapshot), which is worth having and is not what the sentence says.
*Required fix*: use a portable formulation — `ls -laR .agents 2>&1` or
`find .agents -mindepth 0 -maxdepth 3 -exec ls -ld {} \; 2>&1`, with stderr **kept** rather than
discarded so a failed invocation is visible in the artifact; and restate the control sentence to
claim only what it does (a create is caught by the existence assertion).

**N4. K4's row now contradicts §3.6's chosen branch.**
`spec.md` K4 row against `spec.md` §3.6 — Severity: **major** — Class: **blocking** — Origin:
round-1 D8, **escalated by new material**.

K4 still reads: "Excludes the never-had-a-mirror population — C-1: an update must not **create**
`.agents/` in a project no deploy ever gave one." §3.6 now deliberately decides to create
`.agents/` in a project no deploy ever gave one (a stamped-but-partial project), arguing that "a
stamped-but-partial project is not that population". Both sentences are in the same document, and
one of them is false. In round 1 this was a labelling imprecision I marked optional; §3.6 converted
it into an internal contradiction, which is a CN-1 consistency defect.
*Required fix*: rewrite the K4 row as "excludes projects whose last deploy predates mirror
creation", and let §3.6 own the stamped-but-partial case explicitly. The conformance table's verdict
for R5 does not change; only the criterion's wording does.

**N5. The run-phase question the author reports leaving open is not recorded in any artifact.**
`progress.md:19`, `plan.md` M2 — Severity: **minor** — Class: **blocking** (recording it is the
whole fix) — Origin: **new material**. Reported as a **Gap**, not as an approved boundary.

The author reports deferring one question — whether AC-UMH-015/016's stamped-but-partial fixture
can be built with existing test seams or needs a new injection point. I looked for it:
`grep -rn "injection\|inject\|seam\|reachab"` over `plan.md` and `progress.md` returns only the M3
seam-shape decision. `progress.md:19` lists exactly three carried gaps — constant home package, seam
shape, and B3 — and this is not among them. I cannot adjudicate whether a deferral is a legitimate
scope boundary when the artifacts do not contain it; a question that exists only in a report to the
coordinator is not carried into the run phase by anything.

On the merits, so the recording is not merely bureaucratic: the deferral looks acceptable, but for a
reason the SPEC should state. AC-015/016 assert the *pass's* behavior on a given project **state**,
and that state (stamp present, no mirror, no `.claude/skills`) is directly constructible by hand —
no failing deploy needs to be injected. AC-UMH-001's "MUST NOT write the stamp by hand" rule is
scoped to AC-001's own fixture and does not bind AC-015. What a hand-built state does **not**
establish is that a real partial deploy produces it — the same bypass D2 objected to, one level
down. That residual is small and is fair run-phase work; it just has to be written down.
*Required fix*: add one line to `progress.md`'s carried-gaps list stating the question and the
above disposition (hand-built state is sufficient for AC-015/016; whether a real partial-deploy
fixture is also built is a run-phase call).

## Assessments the coordinator asked for by name

**Is the author's downgrade of the `wire.go` premise sound, or an under-valuation?** Sound — and it
is the better epistemics, not modesty. I read `internal/codexwiring/wire.go:47-51` verbatim: "File
existence is the user's standing opt-in — a `--agent claude` (or flag-absent) init left no wiring
behind, and an update must not create any." The comment attributes absence to a *flag choice* and
forbids creation on that basis; it says nothing about absence caused by a failed run. Quoting it as
authority over the failed-run case would be transporting a rule across the axis it was written for —
the same move that makes "a reference exists, therefore the referent is live" unsound. And it
governs codex wiring, the gate this SPEC deliberately rejected in §3.3 R2, so using it as authority
here would also mean borrowing force from a mechanism the SPEC declined to adopt. "Analogical
corroboration, not authority" is the correct weight. The decision correctly does not rest on it.

**Do §3.6's two acceptance grounds hold in code?** Ground ① (stuck-forever) holds: the early return
fires on exact string equality of the stamp with the running binary's version
(`update_template_sync.go:617-622`, `packageVersion == projectVersion && !forceUpdate`), so a
partial deploy that wrote the stamp does make every later same-version update skip deploy
permanently. Ground ② (Path A self-limiting) does **not** hold as a property of the existing
producer — `mirrorOneSkill` performs no source-existence check and will create a dangling link — so
it holds only if the new pass implements the intersection, which is exactly what N1 says is
contradicted by `plan.md` M3 and unbacked by any REQ. Ground ② is a requirement on the work, written
as though it were an observation about the code.

**Are AC-UMH-015/016 mechanically verifiable, and is the Path B residual risk honestly stated?**
Both are verifiable — each names a test, and their fixture is constructible (see N5). The Path B
residual is stated honestly and in the SPEC's own voice: `spec.md` §3.6 says the pass "will restore
the 16 Path B published files even though that project never held them … That is file creation from
templates in a project whose deploy was interrupted", and AC-UMH-016 asserts the creation rather
than tolerating it, with the reason spelled out ("so 'accepted' does not decay into 'unverified'").
I have no complaint about that disclosure. My complaint is N4: §3.6 makes this decision while K4
still asserts the opposite rule two sections earlier.

## Category Scores (round 2)

| Dimension | R1 | R2 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | 0.75 | The D4 correction removed one false premise and the new §3.6 is unusually legible, but two internal contradictions replaced it: M3-vs-S2/AC-015 (N1) and K4-vs-§3.6 (N4). Both are places where two artifacts predict different behavior. |
| Completeness | 1.00 | 1.00 | 1.0 | All sections retained; §3.6, §D.0, §D.2 added; five `### Out of Scope — <topic>` H3 sub-headings still present at `spec.md:336, 342, 349, 355, 363`; frontmatter at 0.3.0 still carries all 12 canonical fields with `phase: "v3.2.0 target"`. |
| Testability | 0.75 | 0.85 | 0.75-1.0 | Materially better: no criterion is now incapable of failing, absence-as-PASS criteria carry controls (§D.0 rule 2), AC-005's behavioral half is mandatory, AC-017 adds the comparator matrix. Held below 1.0 by N3 (the C-3 evidence command does not run on stock darwin, and its control claim is false) and N2 (one guard's mutant is missing behind a false sharing claim). |
| Traceability | 0.75 | 0.75 | 0.75 | All nine REQs remain covered. Six of seventeen ACs now trace to non-REQ anchors (010→§3.5 K4, 012→C-4, 014→verification integrity, 015/016→§3.6, 017→§3.5/D3) — mostly legitimate design anchors, but N1's intersection property having no REQ at all is a real coverage hole rather than a bookkeeping one. |

Aggregate = (0.75 + 1.00 + 0.85 + 0.75) / 4 = 0.8375 → **0.84**.

## Must-Pass Results (re-checked at 0.3.0 / 0.2.0)

- **[PASS] MP-1** — `grep -n '^\*\*REQ-'` → REQ-UMH-001..009 at L70, 75, 79, 83, 87, 91, 95, 99, 105.
  Nine, sequential, no gaps or duplicates. No requirement was added by the repair.
- **[PASS] MP-2** — requirement layer only. The nine REQ bodies are unchanged from 0.2.0 and were
  judged GEARS-conformant in round 1; re-read, still conformant. The seventeen `AC-UMH-*` entries are
  Given-When-Then by design and were graded in Group 4, never here.
- **[PASS] MP-3** — frontmatter re-read at `spec.md:1-15`: all 12 canonical fields, `version: "0.3.0"`
  quoted, dates ISO, no snake_case alias, plus optional `tier: M`.
- **[N/A] MP-4** — single-language SPEC (`module: "internal/cli"`).
- **[PASS] MP-5** — re-run: all four referenced SPECs (`SPEC-CODEX-MIRROR-DOCTOR-001`,
  `SPEC-CODEX-COMMAND-SKILLS-001`, `SPEC-CODEX-WIRING-001`, `SPEC-CODEX-SKILLS-CANONICAL-001`) are
  `completed`. No new SPEC-ID reference was introduced by the repair.
- **[N/A] MP-6** — `grep -rn "syscall"` over the SPEC directory → no output.
- **[PASS] MP-7** — `grep -rn 'NEEDS CLARIFICATION'` → no output.

Tier budget: 9 requirements ≤ 16; **17 acceptance criteria > 16** — the Tier M acceptance-criterion
ceiling is exceeded by one. Recorded here rather than as a numbered finding because it is a budget
signal, not a defect: the ceilings apply independently and the intent is to prompt a tier-up or a
split. My reading is that this SPEC is at the top of Tier M and the seventeenth criterion is earning
its place; if the operator disagrees, the cheapest resolution is folding AC-UMH-016 into AC-UMH-015
(one fixture, two assertions), not deleting a criterion.

## Recommendation

FAIL at the Tier M iteration ceiling. Fix order, all small:

1. **N1** — three edits: M3's last sentence states both halves of the scope rule; REQ-UMH-007 gains
   the absent-canonical-directory clause; M3 notes that `mirrorOneSkill` does not stat its source, so
   reuse does not supply the intersection. Without this the run phase can implement the SPEC as
   written and produce sixteen dangling symlinks on the very fixture AC-UMH-015 exists to forbid.
2. **N2** — count nine, and give AC-UMH-015 its own mutant row (drop the existence intersection).
   That row is also N1's probe, so one edit buys both.
3. **N4** — one sentence in the K4 row so the document stops contradicting its own §3.6.
4. **N3** — replace `-printf` with a portable listing, keep stderr, and restate the control claim to
   what it actually establishes.
5. **N5** — one line in `progress.md`'s carried-gaps list.

Because this is the ceiling iteration, the orchestrator escalates rather than iterating silently. If
asked for my view: **PASS-with-debt is defensible here provided N1 and N2 are fixed first** — they
are the two that let a wrong implementation pass a green test suite, and both are text edits
measured in lines, not decisions. N3, N4 and N5 are safe to carry as documented debt into the run
phase. Scope reduction is not warranted: the SPEC is not too large, it is one contradiction away
from being implementable, and the repair round demonstrably raised its evidence discipline rather
than lowering it.

---
---

# Round 3 — re-audit of the iteration-2 repair

Artifacts re-read at `spec.md` 0.4.0 / `acceptance.md` 0.3.0 / revised `plan.md` / revised
`progress.md`. Same tree: `0b1e27877`. Rounds 1 and 2 above are preserved unaltered.

This round exceeds the Tier M iteration ceiling of 2 and runs on an explicit operator override. Its
stated reason is a recurrence pattern rather than an open finding: each of the two prior repairs
introduced new blocking defects, one of which (N1) would have produced damage — sixteen dangling
symlinks — had the run phase implemented the SPEC as written. The primary target of this round is
therefore what the third repair introduced, not whether N1-N5 closed.

**Verdict: PASS**
**Overall Score: 0.91** (Tier M threshold 0.80). Trajectory 0.81 → 0.84 → 0.91; no regression at any
step.

All seven must-pass criteria PASS or are N/A. Three new findings are recorded below; none is
blocking, and I say so on evidence rather than on the absence of evidence — each was executed, not
inspected. Because a PASS starts implementation immediately, I have stated separately whether
`plan.md` M3 is executable as an instruction: it is.

Reasoning context ignored per M1 Context Isolation. The coordinator's message reported what the
author claims to have repaired, and separately reported two facts the coordinator measured. I
re-measured both rather than inheriting them, and both hold.

## Round-2 findings — disposition

| R2 finding | Status | Evidence read or executed this session |
|---|---|---|
| **N1** scope rule self-contradictory; safety property requirement-less | **CLOSED — and closed on the right side** | `REQ-UMH-010` now exists (`spec.md` §2): "The mirror-repair pass shall not create a mirror entry whose canonical target directory (`.claude/skills/<name>`) does not exist on disk." `plan.md` M3 is rewritten into two numbered mandatory steps — **derive** candidates from the embedded template FS (REQ-UMH-007), then **filter** by canonical-target existence before linking (REQ-UMH-010) — and states which reading lost and why. §3.6 ground ② no longer claims an inherited property: it now reads "The pass is bounded on Path A — **by REQ-UMH-010, which this SPEC adds**". Every code coordinate the new requirement cites is exact: `skill_mirror.go:168` `return os.Symlink(oldname, newname)`, `:206` `if info, err := os.Lstat(mirrorPath); err == nil {`, `:247` `if err := d.mirrorCopy(srcDir, mirrorPath); err != nil {`. |
| **N2** false sharing claim + miscount | **CLOSED** | Recounted myself, not recalled: matrix rows marked `yes` are 003, 004, 006, 007, 008, 011, 012, 013, 015 → **9**; AC-UMH-014's mutant table carries **9** rows with the same identifiers. The exemption is retracted in §D.2 and — the part that matters — the retraction argues the inversion correctly, matching my own round-2 analysis in both directions (S1 widening leaves 015 green on an empty-skills fixture; dropping the existence filter leaves 008 green because `local-only-skill` is not a template name). The new standing rule ("any future sharing claim must name the single mutant and argue that **both** guards go red under it") is the correct generalization rather than a patch. |
| **N3** `-printf` inoperative; false control claim | **CLOSED in substance; one new defect inside the fix — R3-1** | I executed the new recipe rather than reading it. See below. |
| **N4** K4 contradicts §3.6 | **CLOSED** | K4 now reads "Excludes the population that predates the mirror feature … one whose last deploy was performed by a binary with no mirror step", and the superseded wording is preserved in a block quote directly beneath it with the reason the two sets differ. The document no longer contradicts itself. |
| **N5** deferral existed only in a report | **CLOSED** | `progress.md` §E.1 now carries it as its own bullet, including the sentence that makes it self-enforcing: "Recorded here because a boundary that lives only in a report is not an approved boundary." |
| D6, D7, D9, D10 (round-1 optional) | **OPEN, unchanged** | Still optional; the orchestrator's call, and it has not been overtaken by anything in this repair. |

**N3 verified by execution, not inspection.** The coordinator measured the *surface* (no `-printf`,
no `2>/dev/null` in the executable recipe) and explicitly left the functional question to me. I ran
the recipe with the mutant, in `/tmp` so this repository's `.agents/` was never touched (C-3 binds
the auditor too):

```
$ test -e .agents; echo "agents_exists_rc=$?" > before.txt     # → agents_exists_rc=1
$ mkdir -p .agents/skills/mutant-probe
$ test -e .agents; echo "agents_exists_rc=$?" > after.txt      # → agents_exists_rc=0
$ find .agents -print | LC_ALL=C sort >> after.txt
$ diff before.txt after.txt; echo "diff_rc=$?"
1c1,4
< agents_exists_rc=1
---
> agents_exists_rc=0
> .agents
> .agents/skills
> .agents/skills/mutant-probe
diff_rc=1
```

The guard flips red under its own named mutant, and every primary used (`test`, `echo`, `find
-print`, `-type l`, `readlink`, `printf`, `sort`, `diff`) is POSIX and runs on stock darwin. The
criterion also names the mutant explicitly in AC-UMH-014's table. That is the round-2 finding closed
on execution.

## New findings — introduced by this repair

**R3-1. AC-UMH-013's `find_rc` assertion measures the wrong command, and its value never reaches the
artifact.**
`acceptance.md` AC-UMH-013 step 2 — Severity: **minor** — Class: **optional (fix during run phase)**
— Origin: **new material, iteration-3 repair**.

The criterion's headline promise is "**every capture command's exit status asserted**, so a command
that failed to run cannot present as an empty snapshot", and step 2 implements it as:

```
find .agents -print | LC_ALL=C sort >> .moai/reports/t520/agents-<phase>.txt
then echo "find_rc=$?" — which MUST be 0
```

`$?` after a pipeline is the exit status of the **last** command in it, so `find_rc` reports
`sort`, not `find`. Executed:

```
$ find /nonexistent-xyz -print | LC_ALL=C sort > out.txt; echo "find_rc_as_written=$?"
bfs: error: /nonexistent-xyz: No such file or directory.
find_rc_as_written=0
$ wc -l < out.txt
       0
```

A failed `find` yields `find_rc=0` and an empty snapshot — precisely the vacuity class §D.0 rule 4
was written to close, reappearing inside the assertion meant to close it. Secondarily, the `echo`
carries no `>>` redirect, so the asserted value is printed to the terminal and never lands in
`agents-<phase>.txt`; the evidence artifact does not contain the status it claims to record.

Why this is **not** blocking, stated so the classification is checkable rather than asserted: the
guard still catches its mutant, because step 1 (`test -e .agents; echo …`) is outside any pipeline
and is what the diff above actually flipped; step 2 only runs when `.agents` exists, which is never
on this tree; and the mandatory control (run the same recipe against `.claude/skills` and assert a
non-zero line count) would catch a globally-broken `find`. The defect is a false sentence in a
criterion, not a hole in a guard.
*Required fix (one line, during run phase)*: split the pipeline —
`find .agents -print > tmp; echo "find_rc=$?" >> agents-<phase>.txt; LC_ALL=C sort tmp >> agents-<phase>.txt`
— so the asserted status belongs to `find` and lands in the artifact.

**R3-2. `REQ-UMH-010` is placed between 008 and 009.**
`spec.md` §2 — Severity: **minor** — Class: **optional** — Origin: **new material**.
Document order in the requirements section is now 001-008, **010**, 009. MP-1 still **passes**: the
set is complete, with no gap, no duplicate and consistent zero-padding, and I am not manufacturing a
must-pass failure out of an ordering nit. But §2 is the one section a reader scans linearly to
convince themselves nothing is missing, and a non-monotonic sequence is exactly what defeats that
scan. Move the block below REQ-UMH-009.

**R3-3. `REQ-UMH-010` carries implementation detail and moving coordinates inside §2.**
`spec.md` §2 — Severity: **minor** — Class: **optional** — Origin: **new material**.
The normative sentence is clean GEARS (a canonical negative). The three explanatory paragraphs
beneath it name a function (`mirrorOneSkill`), a stdlib call (`os.Symlink`), and three file:line
coordinates. All three coordinates are correct today — I verified `:168`, `:206`, `:247`
individually — which is the point: they are correct *now*, they sit in the most stable section of
the document, and nothing re-measures them when `skill_mirror.go` moves. This is the same
moving-coordinate hazard the SPEC's own evidence discipline is otherwise careful about. The
rationale is genuinely useful, so the fix is relocation rather than deletion: keep the `shall not`
sentence in §2 and move the measured justification to §4, which already reasons about the producer.

## The five judgments the coordinator asked for

**1. Did this repair introduce new defects?** Yes — three, all above, all minor, none capable of
producing damage. This is a materially different profile from the two prior rounds: iteration 1
introduced a criterion that could not fail, iteration 2 introduced a contradiction that would have
produced sixteen dangling symlinks, and iteration 3 introduced a mis-scoped `$?`, a misplaced
requirement block, and some rationale in the wrong section. The recurrence pattern that justified
this round is real and is still visible — but its severity has collapsed, which is the signal that
matters for deciding whether to keep iterating.

**2. Is `REQ-UMH-010` a verifiable requirement or more prose?** Verifiable, and verified three ways.
It is a testable negative ("shall not create a mirror entry whose canonical target directory does
not exist on disk"). AC-UMH-015 covers it — the §D matrix now names it (`§3.6 / REQ-UMH-010`) — and
asserts something a wrong implementation fails: "the count is zero **and** no entry under
`.agents/skills` is a symlink whose `os.Stat` fails … because the failure mode here is not 'too many
entries' but *dangling* ones". And it has its own mutant row: "remove the REQ-UMH-010
target-existence filter — on the empty-skills fixture the pass must then create dangling entries,
flipping AC-UMH-015 red". A requirement, a criterion that fails without it, and a mutant that proves
the criterion can fail.

**3. Is the new AC-013 recipe darwin-executable, does the mutant turn it red, and does the text name
that mutant?** Yes, yes, and yes — all three executed above rather than reasoned about. The residual
is R3-1, which touches the recipe's self-description and not its ability to catch the mutant.

**4. Is declining to consolidate the 17 criteria justified?** Yes. The Tier M budget is a signal to
tier up or split, not a mandate to delete coverage, and no coverage was cut to fit. The author names
the one genuine consolidation candidate (AC-003 + AC-004 — both gate-closed no-ops over the same
fixture family, differing only in the stamp value) and declines it for a reason I find stronger than
the saving: renumbering the mutant table during a final repair is precisely the operation that
produced N2's miscount one round earlier. Identifying the candidate, recording it, and declining it
with a risk argument is the correct handling of a budget signal.

**5. Is the carried gap safe to defer to run phase?** Yes, now that it is recorded — and the
recording is what changed, which is why my round-2 answer was "Gap" and this one is "yes".
AC-UMH-015/016 assert the pass's behavior on a project **state**, and that state (stamp present, no
mirror, empty or absent `.claude/skills`) is directly constructible by hand; AC-UMH-001's
"MUST NOT write the stamp by hand" rule is scoped to its own fixture. What a hand-built state does
not establish is that a real partial deploy produces it, and that is a fixture-fidelity question the
run phase can answer with the code in front of it. `progress.md` now carries it, names it as the one
place the §3.6 pin could prove expensive, and cannot lose it silently.

## Is `plan.md` M3 executable as an implementation instruction?

Yes, and I checked this specifically because a PASS starts implementation. M3 now gives the scope
set as two numbered steps, each bound to a named requirement, plus the filter's placement:

1. derive candidates from the embedded template FS, not a listing of `.claude/skills` (REQ-UMH-007);
2. filter by canonical-target existence before linking (REQ-UMH-010);

with "the filter sits above it, in the repair pass, not inside the shared producer". The two steps
are genuinely compatible rather than a relocated contradiction — they act on different axes, one
choosing the name universe and the other applying a membership test to it — so the round-2 conflict
is resolved rather than moved. M3 also records which reading lost and why, so a future reader cannot
re-derive the discarded one from the surviving sentence.

Two notes for the run phase, neither a defect:

- **Path B never touches the producer**, so the dangling hazard is Path A-only: the published
  `SKILL.md` files are restored from the embedded FS and have no canonical-directory dependency.
  REQ-UMH-010 correctly does not constrain them, which is also why AC-UMH-016 can assert that the 16
  files *are* created on the partial fixture without contradicting REQ-UMH-010.
- **The seam-shape decision is still open** (exported `template` entry point vs. an option on the
  existing deployer), and M3 rightly surfaces it. Under either shape the filter is applied by the
  caller when it builds the set, so REQ-UMH-010 holds — but if the deployer-option shape is chosen,
  the same option sits on the deploy-time path, and the SPEC says nothing about whether the filter
  would then also change deploy-time behavior. Deploy-time dangling is pre-existing and out of scope
  here; the run phase should just be deliberate about it rather than discovering it.

## Category Scores (round 3)

| Dimension | R1 | R2 | R3 | Evidence |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | 0.90 | Both internal contradictions are gone: M3-vs-§3.6/AC-015 resolved into two compatible steps, K4-vs-§3.6 corrected with the superseded wording preserved. What remains is presentational, not semantic — the out-of-order requirement block (R3-2) and rationale sitting in the requirements section (R3-3). |
| Completeness | 1.00 | 1.00 | 1.00 | Ten requirements, seventeen criteria, five `### Out of Scope — <topic>` H3 sub-headings, frontmatter at 0.4.0 with all 12 canonical fields, and a HISTORY row per repair round naming what changed and why. |
| Testability | 0.75 | 0.85 | 0.90 | The mutant probe set is complete for the first time (9 guards, 9 rows, counted not recalled), the C-3 recipe is portable and demonstrably red under its mutant, and REQ-UMH-010 arrives with both a criterion and a mutant. Held below 1.0 solely by R3-1's mis-scoped exit status. |
| Traceability | 0.75 | 0.75 | 0.85 | All ten REQs covered, including the new one (AC-UMH-015 → `§3.6 / REQ-UMH-010`). Five of seventeen criteria still anchor to design sections rather than requirements (010→§3.5 K4, 012→C-4, 014→verification integrity, 016→§3.6, 017→§3.5/D3); each is a legitimate design or meta anchor, and the round-2 hole — a load-bearing property with no requirement at all — is closed. |

Aggregate = (0.90 + 1.00 + 0.90 + 0.85) / 4 = 0.9125 → **0.91**.

## Must-Pass Results (re-checked at 0.4.0 / 0.3.0)

- **[PASS] MP-1** — REQ-UMH-001..010, ten entries, no gap, no duplicate, consistent padding. Document
  order is non-monotonic (R3-2) but that is not a gap or a duplicate and does not fail this criterion.
- **[PASS] MP-2** — requirement layer only. REQ-UMH-010 is a canonical GEARS negative ("The
  mirror-repair pass shall not create a mirror entry whose …"); the other nine are unchanged from
  0.2.0 and were judged conformant in round 1. The seventeen `AC-UMH-*` entries are Given-When-Then
  by design and were graded in Group 4, never here.
- **[PASS] MP-3** — `spec.md:4` `version: "0.4.0"`, all 12 canonical fields present with correct
  types, `phase: "v3.2.0 target"`, no snake_case alias, optional `tier: M`.
- **[N/A] MP-4** — single-language SPEC (`module: "internal/cli"`).
- **[PASS] MP-5** — the four referenced SPECs remain `completed`; the repair introduced no new
  SPEC-ID reference.
- **[N/A] MP-6** — no `syscall` in the SPEC directory.
- **[PASS] MP-7** — no `[NEEDS CLARIFICATION` marker in `plan.md`; `research.md` is absent at Tier M.

Budget: 10 requirements ≤ 16; 17 criteria against a guide of 16 — recorded as a signal and
adjudicated in judgment 4 above, not counted as a defect.

## Recommendation

**PASS.** Proceed to run phase under the operator's auto-advance approval.

Carry three items into the run phase, all optional and all one-line:

1. **R3-1** — split AC-UMH-013's step-2 pipeline so `find_rc` belongs to `find` and lands in the
   artifact. Do this before the first C-3 capture, since the "before" snapshot is taken at the start
   of the run phase and a snapshot taken afterwards witnesses nothing.
2. **R3-2** — move the `REQ-UMH-010` block below `REQ-UMH-009`.
3. **R3-3** — relocate REQ-UMH-010's measured rationale to §4, keeping the `shall not` sentence in §2.

Round-1 optional findings D6, D7, D9 and D10 remain open and remain the orchestrator's call; none
blocks implementation.

I am recording the closure plainly, because a round run past the ceiling has no obligation to find
something: the two defects that could have caused damage — a guard that could not fail, and a scope
rule that would have produced sixteen dangling symlinks — are both closed, each verified by
execution rather than by reading the author's account of it. What this repair introduced is a
misplaced `$?`, a misordered block, and some rationale in the wrong section. That is the signal the
override was bought to obtain.
