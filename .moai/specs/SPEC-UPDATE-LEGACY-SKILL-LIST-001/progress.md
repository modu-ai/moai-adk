# SPEC-UPDATE-LEGACY-SKILL-LIST-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

### Baseline

- **Code baseline and worktree HEAD** `9a6b6c854`, branch `main`.
- **Baseline re-validated at run-phase entry.** `origin/main` advanced to `f5dba46c3` (#1259) before the feature branch was cut. `git diff --stat 9a6b6c854 f5dba46c3` touches only `README.{md,ko,ja,zh}.md` (4 files, +28/-4). Every file this SPEC's ACs measure — `internal/cli/update_archive.go`, `internal/cli/update.go`, `internal/template/skills_manifest.go`, `internal/runtime/audit_cache.go` — is byte-identical across that range (`git diff --quiet` exits 0 for each), and `.moai/archive/` is untouched. Every baseline recorded against `9a6b6c854` therefore holds unchanged at `f5dba46c3`, which is the feature branch's base.
- **Run-phase location**: isolated worktree `.claude/worktrees/legacy-skill-list` on branch `feat/SPEC-UPDATE-LEGACY-SKILL-LIST-001`, cut from `origin/main` at `f5dba46c3`. The primary checkout's branch state is untouched per `main-checkout-branch-guard.md`.
- `git rev-list --count --left-right origin/main...HEAD` → `0	0` (synced, no divergence).
- Working tree clean apart from one untracked report HTML (`.moai/reports/moai-pi-pack-final-design-20260729.html`), which this SPEC does not touch.
- Every artifact naming a baseline names `9a6b6c854`.

### v0.1.0 (initial plan-phase authoring)

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- Every `file:line`, count, and command carried in the delegation brief was independently re-executed before being written into an artifact. The full verification table is plan.md §C.
- **Two brief claims corrected by measurement**:
  - Reference counts: the brief cited backend 46 / database 25 / frontend 17. Measured, in **backend / database / frontend** order: file counts `56 / 30 / 26`, occurrence counts `68 / 30 / 34`. The conclusion — these are among the most-referenced skills in the repo — is unchanged and strengthened. (v0.2.0 correction: the v0.1.0 line stated the occurrence triple as `68 / 34 / 30`, transposing frontend and database relative to the file-count ordering used in the same sentence. spec.md §A uses backend/frontend/database order — `56 / 26 / 30` and `68 / 34 / 30` — and plan.md §C uses backend/database/frontend — `56 / 30 / 26` and `68 / 30 / 34`; both were internally consistent. Only this line mixed the two orderings.)
  - Inventory parity: "31 local `moai-*` dirs" is 30 `moai-*` **plus** the bare `moai` unified-skill directory. A `find` diff of the template tree against the local `moai-*` glob showed the only difference is the bare `moai`, confirming 31 = 31 = 31 with that precision.
- **One additional defect found during verification and folded into scope-adjacent evidence**: `internal/template/skills_removal_test.go` was already narrowed from 16 entries to 9, carrying the in-source note "Some skills from the original 16-skill removal list still exist in the template tree" and naming `moai-domain-backend` / `-frontend` / `-database` among those still present. The revival was therefore noticed in the template package on 2026-04-28 and the parallel correction to `legacySkillIDs` never happened. Recorded as spec.md §A Defect 4. That comment's second line is itself now stale (four of the IDs it lists as "still exist" have since been removed) — recorded as an out-of-scope observation, not scope.
- **Positional-index risk resolved by measurement, not assumption**: the brief flagged `legacySkillIDs[0..2]` and `[:5]`. The full set is `[0]`, `[1]`, `[2]`, `[3]`, `[:5]`, `[:8]` — max index 3, max slice bound 8. All remain in range at 13 entries, so shrinking the list breaks no existing test. Confirmed by a green baseline run.
- The 3-element intersection was reproduced by an executed in-package probe, not carried over from the brief. Probe removed; `git status --porcelain` confirmed the tree returned to its prior state.

### Observed at authoring time (2026-07-31, HEAD `9a6b6c854`)

Pre-fix baselines for every AC (full command text in acceptance.md §B):

| AC | Observed pre-fix baseline |
|----|---------------------------|
| AC-LSL-001 | list entry count `16` |
| AC-LSL-002 | offender count in slice body `3` |
| AC-LSL-003 | `lists the 16 skill IDs removed in BC-V3R3-007` present, count `1` |
| AC-LSL-004 | `func TestLegacySkillIDsNotEmbedded` matches `0` — the `-run` selector would have been vacuous |
| AC-LSL-005 | guard absent; intersection measured by probe → `legacySkillIDs=16 embedded=30 overlap=3 [moai-domain-backend moai-domain-frontend moai-domain-database]` |
| AC-LSL-006 | `go test ./internal/cli/ -run 'TestArchive\|TestRestoreSkill\|TestSkipSyncNoArchive' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli  0.776s`; `All16` occurrences 2 and 2 |
| AC-LSL-007 | `--- SKIP` count `0` (guard file absent) |
| AC-LSL-008 | three offender archive paths tracked → `3` |
| AC-LSL-009 | `git ls-files .moai/archive/skills/v2.16/` → `7`; four genuine dirs → `4`; `grep -c archive .gitignore` → no match (exit 1) |
| AC-LSL-010 | `func TestArchiveLegacySkills_ContinuesAfterFailure` matches `0` |
| AC-LSL-011 | pre-fix, all three in-loop returns precede the `total:` emission, so a failure suppresses the summary |
| AC-LSL-012 | in-loop `return archived, fmt.Errorf` count `3` (lines 302, 305, 320); `errors.Join` count `0` |
| AC-LSL-013 | `"archive: "` count `1`; `total: %d skills archived` count `2` |
| AC-LSL-014 | `go vet ./internal/cli/ ./internal/template/` exit `0`, no output; `go test ./internal/cli/... ./internal/template/...` all `ok`; `gofmt -l internal/cli/update_archive.go` **lists the file** (pre-existing single-hunk import-order deviation); `git status --porcelain internal/template/templates/` → `0` |

Additional verified anchors:

- `git log -S"moai-domain-backend" -- internal/cli/update_archive.go` → exactly one commit, `ec0e9e257`.
- `git ls-tree -r --name-only 74bae50f4^ -- .claude/skills/moai-domain-backend` → `SKILL.md`, `references/examples.md`, `references/reference.md` — confirming the archived copies hold only the post-revival consolidated `SKILL.md`, not the deleted legacy body.
- Live vs archived `SKILL.md` md5 differ for all three (table in spec.md §A Defect 2).
- Catalog registration reproduced at lines 157/159 (backend), 162/164 (database), 220/222 (frontend).
- The 13 retained IDs are absent from **both** the template tree and the local `.claude/skills/` tree — verified individually, so the corrected list is genuinely all-legacy.

### v0.2.0 (plan-audit iteration-1 revision — M1-M4 + S1-S10)

```yaml
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.79
    threshold: 0.80
    dimensions: {clarity: 0.85, completeness: 0.85, testability: 0.65, traceability: 0.85}
    must_pass: 7/7
    resolved: [M1, M2, M3, M4, S1, S2, S3, S4, S5, S6, S7, S8, S9, S10, S11]
    deferred: []
    report: .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/audit.md
```

The auditor's baseline re-verification reproduced 28 of 31 recorded values exactly and found zero defect-locus mis-attributions. The FAIL was on four localized defects, all resolved below. Every reproduction was re-run independently before its correction was written.

**M1 — `plan.md §A` D1's rationale was refuted.** Rewritten. The v0.1.0 text claimed the six siblings were "mid plan-audit revision … in flight on `plan/epic-update-config-audit`" and that editing from a different branch would collide. Measured at `9a6b6c854`: `git diff --stat main plan/epic-update-config-audit` is empty and `git rev-list --count --left-right main...plan/epic-update-config-audit` → `0	17` — the branch is fully merged and both this SPEC and the siblings live on `main`, so no collision is possible. The replacement rationale rests on two measured facts instead: the roster inconsistency (below) and the **plan-artifact-hash cost** — per `spec-workflow.md` § Report Persistence, `internal/runtime/audit_cache.go` `planArtifactNames` hashes `{spec.md, plan.md, acceptance.md, tasks.md}`, so editing a sibling's HISTORY denominator changes `spec.md`, invalidates its cached PASS, and forces a full Phase 1 plan-audit re-execution. That is a mechanical cost, not cosmetic churn.

**Epic roster — ordinal dropped rather than corrected (resolves S11).** `SPEC-UPDATE-YAML-PRESERVE-001` **is** an Epic member — `SPEC-UPDATE-DATA-SURVIVAL-001/progress.md:40` and `SPEC-UPDATE-DOC-DRIFT-001/progress.md:121` both name it among the remaining Epic SPECs — yet it carries no ordinal, while the other six self-number `1 of 6` … `6 of 6` with no gaps. The roster is 7 members with 6 numbered. "Epic SPEC 7 of 7" would have been wrong on both the ordinal and the denominator, so this SPEC adopts **no ordinal**, and renumbering the siblings is recorded as out of scope.

**Two contradictory readings of the sibling audit state, and the discriminator.** The audit report asserted the siblings hold iteration-2 PASS verdicts; the revision brief asserted no PASS verdict exists anywhere for this Epic and cited `ls .moai/reports/plan-audit/ | grep -E "^SPEC-(UPDATE|CONFIG)-"` returning nothing. Re-measured directly:

```
$ ls .moai/reports/plan-audit/*epic-update-config-iter2.md   → 6 files, all dated Jul 31 15:09
$ for f in …; do grep -m1 -E 'Verdict' "$f"; done
  CONFIG-KEY-HONESTY-001    PASS 0.81      CONFIG-TIER-PERSIST-001   PASS 0.82
  UPDATE-CI-GUARD-001       PASS 0.85      UPDATE-DATA-SURVIVAL-001  PASS 0.84
  UPDATE-DOC-DRIFT-001      PASS 0.82      UPDATE-REINSTALL-LOOP-002 PASS 0.88

$ ls .moai/reports/plan-audit/ | grep -E "^SPEC-(UPDATE|CONFIG)-"   → exit 1, no match
```

Both facts hold simultaneously: **the reports exist (6/6 PASS at iteration 2), and the brief's grep genuinely returns nothing** — because this shell aliases `ls` to long format, so every listing line begins with `-rw-r--r--@` and `^SPEC-` can never match. The brief's refutation was a measurement artifact, not a finding. Separately, the brief is right that no iteration-2 verdict was backfilled into any sibling's `progress.md` (each still records only `iteration_1: verdict: FAIL`) — the PASS lives only in the report files. Both observations are now in plan.md §A D1, along with the `ls | grep '^name'` trap, which had already produced one garbled directory diff during initial authoring.

**M2 — `AC-LSL-004(b)` / `AC-LSL-005` / `AC-LSL-007` were mutually unsatisfiable.** Fixed by anchoring. Reproduced on a throwaway module carrying the exact shape the SPEC mandates (parent + one passing subtest + two skips):

```
--- PASS: TestLegacySkillIDsNotEmbedded (0.00s)
    --- PASS: TestLegacySkillIDsNotEmbedded/production (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_error (0.00s)
    --- SKIP: TestLegacySkillIDsNotEmbedded/manifest_empty (0.00s)

unanchored 'grep -c -- "--- PASS: Test…"'  → 2   ← AC expected 1: FAILS
anchored   'grep -cE "^--- PASS: Test…"'   → 1   ← passes
'grep -c -- "--- SKIP"'                     → 2   ← AC-LSL-007 passes either way
```

The FAIL direction was verified separately (one failing subtest): unanchored `2`, anchored `1`, offending ID named. Both ACs now use `^`-anchored patterns, and plan.md §E M2 additionally pins the required test shape — two independent defences rather than one.

**M3 — `AC-LSL-014(b)` was blind to a test-compile failure.** Replaced with an exit-code assertion plus a `[build failed]` marker check. Reproduced on a throwaway module with one broken `_test.go`:

```
$ go build ./...                                  → exit 0   (go build skips _test.go)
$ go test ./... -count=1                          → exit 1
  # vac/pkg [vac/pkg.test]
  pkg/a_test.go:9:6: undefined: undefinedSymbol
  FAIL	vac/pkg [build failed]
  FAIL
$ go test … | grep -c -E '^(FAIL|---) '           → 0   ← the retired AC passed vacuously
```

The pattern required a space after `FAIL`, but Go emits a TAB, and a compile failure produces no `--- ` line at all. This SPEC adds two new test files to `internal/cli`, so the blindness was live. New baseline measured on the real tree: `go test ./internal/cli/... ./internal/template/... -count=1` → **exit=0**, `grep -c 'build failed'` → **0**, `grep -c -- '--- FAIL'` → **0**.

**M4 — `§C.1` step 4 used a destructive, non-discriminating revert.** Replaced with a new `§C.0` mechanism section built on `go test -overlay`, which substitutes file content at compile time and never writes to the working tree. Both of the auditor's grounds were confirmed: `internal/cli/update_archive.go` carries M1 **and** M4, so `git checkout -- <path>` could discard the whole fix, and a `git status --porcelain` post-condition cannot detect that (a destroyed fix produces the same clean tree the check expects). The command is also on the primary-checkout forbidden list in `main-checkout-branch-guard.md`. The auditor's proposed `git restore --source=<sha>` was **not** adopted — it is the same class of working-tree write. Mechanism proven before adoption:

```
baseline (no overlay)   → --- PASS
with overlay (mutated)  → --- FAIL, "intersection non-empty: LIVE"
grep -c 'LIVE' <source> → 0        ← working tree provably untouched
```

`§C.0` adds three mandatory post-conditions — P1 (file unmodified), **P2 (the fix is still present: 13 entries)**, P3 (mutation never reached the repo) — and P2 is the discriminating check the retired procedure lacked. `§C.4` (a git-index property, not a compile-time one) was rebuilt on `git rm --cached` + `git reset --`, which stages and unstages without touching file content. Because nothing is written, the auditor's suggested "M1 must be committed before falsification" precondition is unnecessary and was not added; plan.md §E M2 records that as the reason the overlay form was chosen.

**SHOULD-FIX disposition — all 11 resolved, none deferred.**

| Item | Resolution |
|------|-----------|
| S1 | `AC-LSL-011`'s `grep -c '"total:"'` (which would reject `strings.Contains(output, "total: ")`) replaced with `grep -cF 'total:'`. Both new test filenames pinned in plan.md §E M2 / M4, since five ACs reference them by path. |
| S2 | `REQ-LSL-016` promoted from a reviewer-read obligation to its own mechanical criterion, **AC-LSL-016**, requiring a named `success_count_excludes_failures` subtest whose PASS line is asserted. The remaining AC-LSL-010 obligations are now explicitly labelled non-binary structural properties. |
| S3 | `AC-LSL-009(d)` grep scoped to `^\.?moai/archive` and the `grep -c` exit-1-on-zero-match behaviour documented, with the count captured into a variable so the AC cannot abort a `set -e` script. |
| S4 | `REQ-LSL-012` rewritten as a single-subject prohibition ("shall not add a `.gitignore` rule for the archive tree"). Its former `Where`-branch was unsatisfiable — `AC-LSL-009(d)` requires zero matches, so the primary clause was dead. |
| S5 | `REQ-LSL-008` restated with the guard as subject in event-detected GEARS form ("**When** the guard is evaluated against the pre-correction 16-entry list, the guard shall fail and shall name the offending IDs"). `REQ-LSL-012`'s compound two-subject shape resolved by S4. |
| S6 | Frontmatter adopted the Epic's canonical style: `version: "0.2.0"` (quoted), `priority: P1`, `phase: "v3.0.2"` — matching five of six siblings. |
| S7 | The transposed occurrence triple corrected in place, above, with the ordering made explicit. |
| S8 | **AC-LSL-015** added, pinning `archiveVersion = "v2.16"` at source level. AC-LSL-008/009 exercise the archive only as git paths and would not catch a constant change; M4 is the only milestone editing the file the constant lives in. |
| S9 | plan.md §E M4 step 2 now states `continue` semantics explicitly, with the reason: the two drift-backup sites sit above `archiveSkill`, so a literal `return → append` substitution without `continue` falls through into `os.Rename` for an entry whose backup parent could not be created. |
| S10 | **REQ-LSL-018** added (§D.5 Regression containment); `AC-LSL-006` now traces to it instead of to a plan milestone. |
| S11 | Resolved by dropping the ordinal — see the Epic roster paragraph above. The auditor flagged this as unverifiable; it is now verified, and the answer invalidated the denominator it was checking. |

AC count 14 → 16.

### v0.3.0 (plan-audit iteration-2 revision — MF-1..MF-3 + SF-a..SF-h)

```yaml
plan_audit:
  iteration_2:
    verdict: FAIL
    score: 0.81
    threshold: 0.80
    must_pass: 7/7
    trend: 0.79 -> 0.81 (rising; no STOP / scope-reduction triggered)
    note: >
      Aggregate cleared the threshold; the FAIL was driven by one defect that
      made the Definition of Done unachievable (MF-1), not by the score.
    report: .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/audit.md (## Iteration 2)
  iteration_3:
    verdict: PASS
    score: 0.89
    threshold: 0.80
    must_pass: pass
    must_fix: 0
    should_fix: 6
    trend: 0.79 -> 0.81 -> 0.89 (monotonic rise)
    note: >
      Final allowed iteration. All three iteration-2 MUST-FIX confirmed
      genuinely closed (not reworded); SF-a..SF-h all eight resolved; both
      carried gaps closed. No escalation needed — PASS is unconditional, so
      PASS-with-debt / scope-reduction / user-override are all inapplicable.
      The 6 SHOULD-FIX are non-blocking and need no re-audit; fold them into a
      v0.3.1 editorial pass or the sync phase.
    report: .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/audit.md (## Iteration 3)

plan_status: audit-ready
plan_complete_at: 2026-07-31
```

**Open SHOULD-FIX from iteration 3 (non-blocking, no re-audit required):** D1 — `plan.md:11` says "Three measured facts" where the MF-2 correction left two legs; D2 — a "one unanchored occurrence remains" statement where there are two (both in prose, so the conclusion holds); D3 — the lint absence-claim recorded in Gap 1 below (already corrected in place); D4/D5 — §A rule 3a applies itself unevenly and lacks a parent-FAIL row; D6 — `must_pass: 7/7` notation.

**Section-authorship note (ledger closure).** The v0.3.0 artifact revision was performed by `manager-spec`, which terminated mid-write on a provider weekly-rate-limit error while composing this progress.md section. `spec.md`, `plan.md`, and `acceptance.md` had already been written and are complete; this section was completed by the orchestrator. Every claim below is attributed to a command the orchestrator ran against the post-revision tree — no claim is carried over from the terminated agent's report.

| MUST-FIX | Resolution | Orchestrator verification |
|----------|------------|---------------------------|
| MF-1 — the anchoring defect class recurred between `AC-LSL-010(b)` and the newly-added `AC-LSL-016(b)` | Swept rather than point-fixed. A standing `[HARD]` rule was added at acceptance.md §A 3a requiring anchored patterns for any count-based criterion over `go test` output, with the recurrence itself named in the rule text so the class cannot silently return. | `grep -c '\^--- PASS\|\^--- SKIP\|\^--- FAIL' acceptance.md` → `8` anchored patterns. Exactly one unanchored `--- PASS:` occurrence remains, and it sits inside §A rule 3a's own explanatory prose as the counter-example — not in an executable criterion. |
| MF-2 — D1 leg (ii) replaced a false premise with a differently-false premise | The mechanism argument was removed entirely rather than replaced a third time. plan.md §A D1 now carries a retraction table recording both prior rationales and how each failed, plus the generalized rule: **when a decision is a scope statement, state the scope and stop.** | Three legs independently re-verified: `planArtifactNames` exists at `internal/runtime/audit_cache.go:63-68` hashing `{acceptance.md, plan.md, spec.md, tasks.md}`; `internal/runtime/audit_cache.go:73-74` states verbatim that cache entries "do not persist across separate /moai run invocations"; `spec-workflow.md:319` sets skip-eligibility condition 2 at score ≥ 0.90 while the six siblings score 0.81-0.88. The claimed cost is therefore zero, and the retraction is correct. |
| MF-3 — `AC-LSL-015(b)` carried `EXPECT: unchanged from baseline` with no number, over a prose baseline naming 3 regions where the measured count is 6 | Numeric `EXPECT: 6` supplied with the full line list (73, 287, 298, 311, 328, 347) and an inline note recording that the v0.2.0 text violated §A rule 4 and undercounted by naming three regions for six sites. | Read at acceptance.md `### AC-LSL-015`; `EXPECT: 6` / `BASELINE (2026-07-31, 9a6b6c854): 6` present with the enumerated line list including `dryRunArchiveLegacySkills`. |

Structural counts after the revision, measured directly: REQ `19` unique IDs, NFR `5`, AC `16`. Every `REQ-LSL-*` and `NFR-LSL-*` identifier appearing in spec.md is literally cited in acceptance.md — zero uncovered. `REQ-LSL-018` was split into 018/019 so that `AC-LSL-006(a)` and `(c)` each trace to a single requirement.

**SF-a..SF-h**: spec.md's v0.3.0 HISTORY row records all eight as resolved. The orchestrator verified the two the iteration-2 report singled out as latent instances of the MF-1 class (the `AC-LSL-006` existence guard and the §C.0 P2 post-condition) are addressed in the artifact text; the remaining six are recorded on the authoring agent's claim and were **not** independently re-measured by the orchestrator. Treat them as claimed-not-verified until iteration 3 audits them.

**Gaps — both closed at iteration 3. Retained with their history because one of them was an instance of this SPEC's own target defect class.**

1. **Lint — CLOSED, and the gap statement that preceded it was itself defective.** `moai spec lint` with no argument lints the whole catalogue and does not complete under a bound (iteration-2 auditor: `exit=124` at 200s; orchestrator: `exit=143` at 120s). From that, plus a single rejected flag (`moai spec lint --spec <ID>` → `Unknown flag: --spec`), the orchestrator wrote here that "no cheap per-SPEC narrowing is available." **That was an unobserved absence claim.** `moai spec lint --help` carries `USAGE: moai spec lint [spec.md...] [--flags]` — a positional file argument. Measured: `moai spec lint .moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/spec.md` → `✓ No findings — all SPEC documents are valid` in **0.37s**, and `--strict` on the same path is equally clean. A directory path is rejected (`ParseFailure … is a directory`); the argument is a file. So the v0.3.0 artifacts ARE lint-clean, verified, and per-SPEC narrowing was cheap all along. Recorded verbatim rather than silently corrected: declaring a capability absent because one flag spelling failed, without reading `--help`, is the same claim-integrity failure (`verification-claim-integrity.md` §1.1 surface 3 — a defect/absence claim requires the domain tool to have been run) that this SPEC exists to remove from `legacySkillIDs`. It appeared here, in this SPEC's own evidence record, written by the orchestrator.
2. **Sibling write-coupling — CLOSED.** No sibling Epic SPEC modifies `internal/cli/update_archive.go`. The only candidate, `SPEC-UPDATE-DATA-SURVIVAL-001` M2, edits registry code and `dirs.go`; `update_archive.go` is a *scan* target for it, not an edit target. The coupling is read-coupling, not write-coupling, so the parallelism claim in plan.md is now measured rather than asserted.

**Landing-order consequence (measured, iteration 3).** This SPEC SHOULD land **before** `SPEC-UPDATE-DATA-SURVIVAL-001`. That SPEC's M2 ships a registry hard-coding `update_archive.go:92` / `:304` together with a static-scan drift guard; this SPEC's M1 shifts those lines by three and M4 restructures the enclosing function. Landing second turns that guard red on arrival. The Epic run-order note below reached the same conclusion by inference; it is now backed by measurement.

### Lint and isolation, observed after authoring

- `moai spec lint` → `0 error(s), 62 warning(s)`; `moai spec lint | grep -i 'LEGACY-SKILL-LIST'` → no match (exit 1). Zero findings for this SPEC; the 62 warnings all belong to other, grandfathered SPECs. Re-run after the v0.2.0 revision: unchanged (`0 error(s), 62 warning(s)`, still no finding for this SPEC), and `moai spec lint --strict | grep -i 'LEGACY-SKILL-LIST'` → no match (exit 1) — zero findings under strict as well.
- Post-revision structural checks: `REQ-LSL-001..018` and `NFR-LSL-001..005` sequential with no gaps; `### AC-LSL-001` … `### AC-LSL-016` sequential with no gaps; every REQ and NFR ID literally cited in acceptance.md (zero uncovered); `grep -c 'NEEDS CLARIFICATION' plan.md` → `0`; the only `git checkout --` / `git restore --source` occurrences in acceptance.md are the four inside the §C.0 prohibition prose, none in an executable step.
- `grep -c '^### Out of Scope —' spec.md` → `6`, each with `-` bullets, satisfying the `OutOfScopeRule` shape.
- `git status --porcelain` shows exactly one new path, `.moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/` (untracked), plus the pre-existing untracked report HTML. No file outside this SPEC directory was created, edited, or deleted; nothing was committed, branched, or pushed.

### Open clarification markers

None, at v0.1.0 and at v0.2.0. Every decision the plan needed — guard placement, degradation semantics, aggregate-error shape, `.gitignore` scope, Epic ordinal, and the v0.2.0 additions (falsification mechanism, test-shape pinning, `continue` semantics) — was resolvable from observed evidence and is recorded as a decision with its rejected alternative in plan.md §A or acceptance.md §C.0.

S11 was the one item the auditor could not verify. It is now verified (YAML-PRESERVE is an Epic member, named in two sibling `progress.md` files) and required no user input, so it did not become a clarification marker.

### Epic run order

This SPEC declares `related_specs`, not `depends_on`, so its run-phase dependency pre-flight is trivially satisfied.

**Parallelism: measured, and narrower than v0.2.0 claimed.** The v0.2.0 text asserted "no sibling touches `internal/cli/update_archive.go`. It may therefore run in parallel." That was asserted, not measured. Measured at `9a6b6c854`:

```
$ grep -rn 'update_archive' <the seven sibling dirs>
SPEC-UPDATE-DOC-DRIFT-001/spec.md:299        update_archive.go:339-353  (dryRunArchiveLegacySkills)
SPEC-UPDATE-DOC-DRIFT-001/plan.md:290        update_archive.go:339-353  "M1 evidence sites"
SPEC-UPDATE-REINSTALL-LOOP-002/acceptance.md:238   update_archive.go:351
SPEC-UPDATE-DATA-SURVIVAL-001/plan.md:116          update_archive.go archiveSkill        (:92)
SPEC-UPDATE-DATA-SURVIVAL-001/plan.md:117          update_archive.go archiveLegacySkills (:304)
SPEC-UPDATE-DATA-SURVIVAL-001/acceptance.md:161-162  same two rows
SPEC-UPDATE-DATA-SURVIVAL-001/plan.md:108          names update_archive.go in a scope list
```

All four cited line numbers resolve exactly today: `:92` → `_ = os.RemoveAll(dstDir)`, `:304` → `if err := os.Rename(dstDir, backupDir)`, `:339` → `func dryRunArchiveLegacySkills(`, `:351` → the `[dry-run] total:` Pill.

What this establishes and what it does not:

- **Established: a live line-number coupling.** M1 removes three list entries, shifting every line below the slice up by 3 — which moves `:92`, `:304`, `:339-353`, and `:351`. M4 restructures `archiveLegacySkills`, moving `:304` further and reshaping the function `SPEC-UPDATE-DATA-SURVIVAL-001`'s M2 static source-scan guard is specified to scan.
- **Not established: whether any sibling *modifies* the file.** The sibling plan text cites these sites as registry rows and evidence anchors; it does not state an edit intent either way, and reading the plans does not settle it. **This is carried forward as an unverified gap**, not resolved by assertion in either direction.

Revised claim: this SPEC's *scope* is disjoint from every sibling's (no sibling milestone lists `update_archive.go` as an edit target), but its *line numbers* are not — three siblings pin line anchors inside the file this SPEC restructures. Running in parallel is therefore acceptable only with the shared-checkout commit discipline (pathspec-scoped commits) **plus** the expectation that whichever SPEC lands second re-measures its own line citations. Sequencing this SPEC before `SPEC-UPDATE-DATA-SURVIVAL-001` would avoid the re-measurement entirely and is the lower-risk order.

## §E.2 Run-phase Evidence

Implemented in worktree `.claude/worktrees/legacy-skill-list` on branch `feat/SPEC-UPDATE-LEGACY-SKILL-LIST-001`, cut from `origin/main` at `f5dba46c3`. TDD order: M2's guard test written first and observed RED against the pre-M1 16-entry list, then M1 turned it GREEN.

### AC matrix (every row is a command the implementer ran)

| AC | Result | Observed |
|----|--------|----------|
| AC-LSL-001 list entry count | PASS | `13` (was 16) |
| AC-LSL-002 offenders in slice body | PASS | `0` (was 3) |
| AC-LSL-003 stale "16 skill IDs" comment | PASS | `0` (was 1) |
| AC-LSL-004(a) guard test exists | PASS | `1` |
| AC-LSL-004(b) parent PASS, `^`-anchored | PASS | `1` |
| AC-LSL-004(c) derives from embedded manifest | PASS | `1`; real-skill literals in the file: `0` |
| AC-LSL-005 falsification via `-overlay` | PASS | anchored `--- FAIL` `1`; failure names `moai-domain-backend` `1`; injected ID in the real source `0` (tree untouched) |
| AC-LSL-006(a-i) pre-existing suite parent PASS | **PASS at a corrected EXPECT of 20** — see below | `20` |
| AC-LSL-006(a-ii) FAIL count | PASS | `0` |
| AC-LSL-006(b) positional indices in range | PASS | max index `3`, max bound `8`, plus a new `[:3]` — all < 13 |
| AC-LSL-006(c) `All16` naming consistency | PASS | `2` and `2` (kept, not renamed — consistent) |
| AC-LSL-007(a)(b) degradation subtests SKIP | PASS | subtests present `2`; indent-anchored `--- SKIP` `2` |
| AC-LSL-008 wrong archive files untracked | PASS | `0` (was 3) |
| AC-LSL-009(a)(b)(c)(d) genuine four preserved | PASS | tracked `4`, exactly the four expected dirs, byte-diff `0`, `.gitignore` rule `0` |
| AC-LSL-010(a)(b) non-aborting loop test | PASS | exists `1`; parent PASS anchored `1` |
| AC-LSL-011 `total:` emitted on failure | PASS | `--- FAIL` `0`; `total:` literal in test source `3` |
| AC-LSL-013 output literals preserved | PASS | `"archive: "` `1`; `total: %d skills archived` `2` |
| AC-LSL-014(a) vet | PASS | `go vet ./internal/cli/` exit `0` |
| AC-LSL-014(b) suite green, build-failure-aware | PASS | `go test ./internal/cli/ -count=1` exit `0`; `build failed` `0`; `--- FAIL` `0` |
| AC-LSL-014(c) gofmt | PASS | `gofmt -l` on the three touched files: empty |
| AC-LSL-015(a)(b) archive scheme untouched | PASS | `archiveVersion = "v2.16"` `1`; archive path joins `6` (unchanged by M4) |
| AC-LSL-016(a)(b) success-only count | PASS | subtest PASS, indent-anchored `1` |

Cross-platform build: `go build ./...` exit `0` on darwin, `GOOS=windows GOARCH=amd64` exit `0`, `GOOS=linux GOARCH=amd64` exit `0`.

### AC-LSL-006(a-i): EXPECT corrected 19 → 20 against measurement

The AC's stated EXPECT of `19` is defective as written, and the defect is in the AC, not the implementation. Its rationale argued the count is stable because "the top-level test count does NOT vary with list length (only subtest counts do), so this number is stable across M1's 16→13 shrink." That reasoning is correct for M1 but incomplete: **M4 adds `TestArchiveLegacySkills_ContinuesAfterFailure`, whose name matches the AC's own selector prefix `TestArchiveLegacySkills_`**, so the same SPEC that authored the AC also raises the number the AC pins.

Verified as an addition, not a substitution: all 19 pre-existing parent tests are present in the run and the single delta is the M4 test. Enumerated names are in the run log. A rename or a dropped prefix would have shown as a missing name, which the AC's own CONSTRAINT clause anticipates; that failure mode did not occur.

### Regression found and fixed: a 7th consumer the SPEC's inventory missed

`TestDryRunArchive` (`internal/cli/update_dry_run_test.go`) failed after M1 with `output missing skill ID moai-domain-backend … got: [dry-run] total: 0 skills archived`. Its fixture hard-coded the three removed IDs as string literals, so after M1 it seeded skills that `dryRunArchiveLegacySkills` — which walks only `legacySkillIDs` — never considers.

This file was absent from the SPEC's 6-file consumer inventory because that inventory was built by grepping for the identifier `legacySkillIDs`, and this consumer coupled to the list by **duplicating its contents as literals** instead of referencing it. AC-LSL-006's selector likewise does not match `TestDryRunArchive`, so the AC passed while the regression was live. The class is: a text-identifier search cannot find a consumer that copied the values instead of referencing the name.

Fixed by deriving the fixture from `legacySkillIDs[:3]`, which both repairs the regression and removes the duplication that hid it. Six further literal uses of the same IDs remain in `migrate_restore_skill_test.go` and `update_archive_test.go`; those are NOT defects — `archiveSkill(root, id)` and `restoreSkill(root, id, force)` take the ID as a parameter and never consult the list, so the literals are ordinary fixture names. Left untouched per scope discipline.

_<pending run-phase>_

- `pre_fix_commit:` _<pending — capture `git rev-parse HEAD` at run-phase entry, before M1's first implementation commit.>_

  Scope of this field (clarified at v0.2.0): it records the tree the AC **baselines** were measured against, so a run-phase reader can confirm the pre-fix values in acceptance.md §B still describe the starting tree. It does **not** bind acceptance.md §C.1-§C.4 — the v0.1.0 wording said it did, which was incoherent, since §C is a post-fix falsification procedure. With the §C.0 overlay mechanism no commit SHA is needed by §C at all: nothing is reverted, so nothing needs a restore source.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: sync-complete (manager-docs sync-phase, single sync commit 3-phase close per Status Transition Ownership Matrix)
- sync_complete_at: 2026-07-31
- sync_commit_sha: 1eb6b2463 (the squash-merge commit of sync PR #1262 on `main` — `docs(SPEC-UPDATE-LEGACY-SKILL-LIST-001): sync-phase close — 3-phase plan→run→sync (#1262)`. A commit cannot carry its own hash, so the sync commit wrote a `pending-backfill-*` placeholder and a follow-up commit on the sync branch backfilled the branch-local SHA `f56b8189a`; because this repo squash-merges (Route B), `f56b8189a` does not survive into `main` history (`git merge-base --is-ancestor f56b8189a origin/main` → false) and a citation to it would not resolve. This third commit therefore records the **merge** SHA, matching the sibling convention — `SPEC-ENVKEY-ANTHROPIC-SSOT-001` records `4d0df6563` and `SPEC-CLIFIX-HYGIENE-001` records `316fe3e84`, both verified `main` ancestors. Per the SHA-placeholder backfill exemption of spec-frontmatter-schema.md § Forbidden ownership crossings.)
- b12_self_test_a (pre-emission grep): PASS — `grep -c 'SPEC-UPDATE-LEGACY-SKILL-LIST-001' CHANGELOG.md` → 0 before appending (no duplicate from a parallel BATCH-SYNC session)
- b12_self_test_b (AC count match): PASS — 16 acceptance criteria, 16/16 PASS, and the CHANGELOG entry states 16/16
- b12_self_test_c (file path verification): PASS — every path claimed in the CHANGELOG entry verified present in `git show --name-only 005d800af` (12 files, +2884 / −734)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Fixed` — SPEC-UPDATE-LEGACY-SKILL-LIST-001 entry, first in section (Tier M, P1, M1–M4)
- frontmatter_status_transitions: spec.md `in-progress → completed` atomic on this single sync commit; `updated: 2026-07-31` refreshed. plan.md / acceptance.md / audit.md bodies untouched (body edits are manager-spec-owned)
- run_phase_pr: #1260 (merge commit `005d800af`, auto-merged 2026-07-31 — legacySkillIDs 16 → 13 + embedded-manifest cross-check guard + wrong-archive removal + non-aborting archive loop)
- ci_rollup: 26 checks — 20 SUCCESS, 6 SKIPPED, 0 FAILURE
- canary_compliance_check: N/A — this SPEC defines no forward-looking policy that its own sync would test
- note: §25 template neutrality N/A — `git show --name-only 005d800af | grep -c '^internal/template/templates/'` → 0; the merge touched `internal/cli` Go sources, SPEC artifacts, and `.moai/archive/skills/v2.16/` deletions only, so no neutrality scan applies. This sync commit carries only the frontmatter transition + CHANGELOG entry + this §E.4 block (no code changes, no README, no docs-site — the user-visible effect is a spurious warning ceasing, for which CHANGELOG is the correct surface)
- known_follow_ups: (a) `moai migrate restore-skill --force` would overwrite a live skill with the stale archive (`migrate_restore_skill.go:68-83`); (b) wrong archives already created in downstream user projects are not cleaned up; (c) `moai-meta-harness` retirement (SKILL.md self-declares DEPRECATED while rules and tests still reference it) — unrelated, surfaced during the same investigation
