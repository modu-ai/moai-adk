auditor-model: claude-opus-5-5[1m]

# SPEC Review Report: SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001

Card: t1259 · Iteration: 4 (dispatch-time re-audit, ordered by the lead after the iter3 ceiling) · Tier: L (threshold 0.85)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1259` · Branch `WT-local-instructions` · Subject HEAD `a73c78d1a` (clean tree)
Date: 2026-09-28

**Verdict: FAIL**
**Overall Score: 0.84** (iter3 0.88) — **STOP signal** (score regression; see Recommendation)

Reasoning context ignored per M1 Context Isolation. Inputs read: all five artifacts, the three prior iteration
reports, `premise-20260928.md`, `verdict.md`, and the parent SPEC's `progress.md` §M1a.

Scope of this pass: (1) is the SPEC implementable now against the current tree, (2) did the landed
dependencies invalidate any requirement or criterion, (3) is iter3's D3-residual closed. Short answer:
M1, M2 and M4 are implementable as written; **M3 (REQ-IFU-021/022) is not** — the landed parent turned its
premise from "verify before relying" into a known negative, and its two criteria measure a file the
repository does not commit. D3-residual is closed.

---

## 1. Is the SPEC implementable now? — measured

| Premise | Command | Observed (this run, HEAD `a73c78d1a`) | Holds? |
|---|---|---|---|
| Parent landed | `grep '^status:' .moai/specs/SPEC-INSTRUCTION-FILES-UNIFY-001/spec.md` | `status: completed` | yes |
| t1175 landed | same, `SPEC-ALWAYS-LOADED-DIET-002` | `status: completed` | yes |
| Loop reads `AGENTS.local.md` first | `grep -n codexLocalInstructionName internal/cli/codex_launcher.go` | `126: for _, name := range []string{codexLocalInstructionName, codexClaudeLocalName}` | yes |
| Preamble site (REQ-IFU-008) | `grep -n 'source: %s' internal/cli/codex_launcher.go` | `137:` | yes |
| AC-IFU-029 regression guard green | `unset MOAI_KANBAN… && go test ./internal/cli/ -count=1 -run '^TestCodexLocalInstructions_DualFileMatrix$' -v` | `--- PASS: TestCodexLocalInstructions_DualFileMatrix (0.01s)`, `ok … 1.347s`, `no tests to run` count 0 | yes |
| AC-IFU-011 RED-now | same with `'^TestCodexLocalInstructions_FallbackAdvisory$'` | `testing: warning: no tests to run` / `ok … [no tests to run]` | yes (red for the right reason: symbol absent) |
| Report streams (AC-IFU-030) | `grep -n 'out := cmd.OutOrStdout()' internal/cli/{update,doctor}.go` | `doctor.go:74`, `update.go:153`; `doctor.go:80 printer.New(printer.WithWriters(out, cmd.ErrOrStderr()))` | yes |
| Verb precedent | `ls internal/cli/migrate_agency*.go` | `migrate_agency.go` 25,790 B, `migrate_agency_test.go` 30,580 B, idempotent test present | yes |
| No verb yet | `grep -rn local-instructions internal/cli/*.go \| grep -v _test` | (none) | premise live |
| 24 docs files + baseline shape | per-file `grep -c '^## '` over ko/en/ja/zh | all 24 exist; claude-md-guide 10/18/18/18, quickstart 14/10/10/10, rest equal — matches `acceptance.md:326-331` | yes |
| Before-value (AC-IFU-007) | `git show develop:CLAUDE.local.md \| wc -m` and `origin/develop` | `45810` both | yes |
| Traceability | `acceptance.md` §D.2 diff command | empty, exit 0; `grep -c '^\*\*AC-IFU-[0-9]\{3\}\*\*'` → 10 | yes |

The relevant surfaces did not move between this branch's base and current `origin/develop`:
`git diff --stat e9577de4f refs/remotes/origin/develop -- internal/cli/codex_launcher.go internal/cli/codex_contract.go internal/cli/update.go internal/cli/doctor.go internal/cli/migrate*.go CLAUDE.local.md CLAUDE.md AGENTS.md docs-site/content .gitignore` → empty (origin/develop is 16 commits ahead).

---

## 2. What the landed dependencies invalidated

### D1 — `AC-IFU-007` / `AC-IFU-024` measure a file this repository does not commit (blocking)

```
$ git check-ignore -v AGENTS.local.md CLAUDE.local.md
.gitignore:275:/AGENTS.local.md	AGENTS.local.md
$ git ls-files AGENTS.local.md CLAUDE.local.md
CLAUDE.local.md
```

`AGENTS.local.md` is ignored at `.gitignore:275` (introduced `6c647bbe2`, 2026-09-10). `CLAUDE.local.md` is
tracked despite its own ignore rule. `REQ-IFU-021` (`spec.md:128-133`) defines the migrating copy as **the
copy committed on `develop`**, and `spec.md:141-147` leans on `CLAUDE.local.md` §0.2 — *an uncommitted working
copy may never be cited as canonical*. Yet both criteria read the working copy:

- `acceptance.md:245` — `wc -m < AGENTS.local.md`
- `acceptance.md:296` — `grep -n -C1 'CLAUDE.local.md' AGENTS.local.md`

**Mutant (writable, so the criterion is not adoptable):** run the M1 verb in the lane, trim the result to
≤ 39,999, commit. Git records the deletion of the tracked `CLAUDE.local.md`; the new `AGENTS.local.md` is
ignored and never staged. Both criteria pass on the working file; the merge lands on `develop` with **no**
maintainer local-instruction file at all, and every lane that branches afterwards loses it. Nothing in the
SPEC says the migrated file must be force-added (`grep -rn -i 'add -f\|force-add\|gitignor' <SPEC dir>` →
only `plan.md:118`, "git-tracked despite being gitignored", about the *old* file).

Severity **major** · Class **blocking** (a stated criterion is satisfiable while its requirement is violated;
contradicts the SPEC's own §0.2 citation rule).

### D2 — M3's lane-reception premise is now a known negative, and nothing gates it (blocking)

The parent landed with this in the canonical contract of this tree:

```
$ grep -n 'worktree session does not receive' AGENTS.md
262:project and is skipped silently, so a worktree session does not receive `AGENTS.local.md`.
```

and the measurement behind it, parent `progress.md` §M1a (`claude 2.1.283`, real nested worktree in this
repository's geometry): `LOCAL_AGENTS_TOKEN = NOT_PRESENT` in the nested worktree, `LOCAL_CLAUDE_TOKEN = QUEBEC2`
(reached by directory-ancestor walk) — "known limitation — a worktree session does not receive
`AGENTS.local.md` content".

This repository's lanes are worktree sessions (this audit runs in one). Today they receive `CLAUDE.local.md`.
`REQ-IFU-021` replaces it with `AGENTS.local.md`, and `REQ-IFU-010a` removes the original. The SPEC's own
plan states why this matters — `plan.md:133-134`: *"Verify the parent SPEC's M1a worktree answer holds for
this repository's own worktrees before relying on it; the lanes read this file."* — but:

- the answer is no longer pending: it landed, negative, for the untracked geometry;
- the force-tracked geometry D1 requires (file present at each worktree root) is **unmeasured** — M1a's
  fixture was untracked and primary-root-only, and AGENTS.md:262 states the negative without that carve-out;
- the check lives only as a plan bullet — no criterion, no stop condition — which is the exact shape
  `acceptance.md:527-529` rejects: *"a duty recorded only inside a criterion's own note is a duty the close
  will not perform."*

The v0.2.5 dispatch repair (`a73c78d1a`) recorded the dependencies as met (`plan.md:44-51`) without
reconciling this consequence. Executing M3 as specified can strip the maintainer doctrine from every lane
with all ten criteria green. The Out-of-Scope rationale at `spec.md:200-203` ("keeps a worktree session from
carrying up to four local-instruction loads") assumes lanes still load the file; post-migration the count may
be zero, not four.

Severity **major** · Class **blocking** (the dispatch asked precisely whether a landed dependency invalidated a
requirement; for REQ-IFU-021/022 it did).

---

## 3. iter3 D3-residual — RESOLVED

```
$ grep -rn "PR head" .moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/
```

Every remaining hit classified: `acceptance.md:444-445`, `plan.md:172`, `progress.md:106,250,458,470-473,509,518`,
`spec.md:24,26` are past-tense history or repair records; `spec.md:39` is the close-condition statement quoting
the search string; `acceptance.md:523` is the negation ("— not from a PR head"). No live normative hit.
`acceptance.md:522-523` and `plan.md:150-160` now name the `origin/develop` head carrying the lane's merge SHA,
and §D.3 carries the four close-time duties (`acceptance.md:531-550`), including the `no tests to run` marker
read and the recording home for the docs-i18n reading. The four sites iter3 named are all repaired.

---

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: nine ids, zero-padded, no duplicates (`spec.md:87-134`). The gaps
  (`001~006`, `013~019`, `023~025`) are the declared carve footprint, justified at `spec.md:80-83`; judged the
  same way as iter1-3. `REQ-IFU-010a/010b` are sub-clauses of one id (`spec.md:110-114`).
- [PASS] MP-2 GEARS (requirement layer only): 007 Where-form (`spec.md:87`); 008, 009, 020, 021, 022
  ubiquitous; 010 unwanted "shall not" + 010a/010b Where-forms; 011 "shall not … ; where …, it shall"; 012
  When-form (`spec.md:117`). ACs are Given-When-Then and were not graded here.
- [PASS] MP-3 frontmatter: all 12 canonical fields at `spec.md:2-13`, `version: "0.2.5"` quoted,
  `status: draft`, `priority: P1`, `lifecycle: spec-anchored`, no rejected aliases; tree-built lint reports
  0 errors.
- [N/A] MP-4: single-language Go CLI scope; the four docs locales are the locale axis, not programming languages.
- [PASS] MP-5 D7: references resolve to `SPEC-INSTRUCTION-FILES-UNIFY-001` (completed),
  `SPEC-ALWAYS-LOADED-DIET-002` (completed); none retired/superseded/archived. `SPEC-V3R3-DOCS-PARITY-001`
  not found → SHOULD (N4 below), not BLOCKING.
- [PASS] MP-6 D8: `grep -c syscall spec.md` → 0.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no match (exit 1).

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | `spec.md:34-36` contradicts itself and `plan.md:44-51` (D3); stale t1269 and carve notes (N3, N5). |
| Completeness | 0.90 | 0.75-1.0 | All sections present; Out of Scope with H3 + bullets (`spec.md:180-211`). M3 lacks its reception gate (D2). |
| Testability | 0.75 | 0.75 | Most criteria are binary and anchored; `AC-IFU-007`/`024` admit a mutant (D1); M3 reception untested (D2). |
| Traceability | 0.95 | 1.0 | §D.2 diff empty; one-to-one mapping (`acceptance.md:482-493`); minor: stale v0.2.3 repair claim (N2). |

Harmonic mean: `4 / (1/0.80 + 1/0.90 + 1/0.75 + 1/0.95)` = `4 / 4.7470` = **0.843 → 0.84**.

## Defects Found

D1. premise-M3-commit — `acceptance.md:245`, `:296` (`.gitignore:275`) — `AC-IFU-007`/`AC-IFU-024` read the working-tree `AGENTS.local.md`, which the repository ignores; REQ-IFU-021's canonical copy is the `develop`-committed one; a mutant that never stages the file passes both — Severity: major — Class: blocking — Required fix: measure from the committed tree (`git show HEAD:AGENTS.local.md | wc -m` on the lane's merge commit; same `git show` source for the `grep -n -C1` read, which also fails loudly if the file was never committed), add to M3 that the migrated file is force-added (`git add -f AGENTS.local.md`) because `.gitignore:275` ignores it, and assert the committed tree no longer carries `CLAUDE.local.md`.

D2. premise-M3-reception — `plan.md:133-134`, `spec.md:128-135`, `spec.md:200-203` — the landed contract (`AGENTS.md:262`) and the parent's §M1a measurement say worktree sessions do not receive `AGENTS.local.md`; this repository's lanes are worktree sessions; the SPEC holds this as an ungated plan bullet — Severity: major — Class: blocking — Required fix, one of: (a) add a criterion measuring reception of the migrated file in a real linked worktree of this repository in the force-tracked geometry (marker token, positive control, M1a method) and make M3 conditional on a present result; or (b) carve REQ-IFU-021/022 (M3) out of this SPEC into a follow-up blocked on worktree reception (t1219 or a new card), leaving M1/M2/M4 runnable now. Update `spec.md:200-203` to match either way.

D3. stale-status-para — `spec.md:34-36` — "no plan-audit has run against this SPEC yet" and "the M2 sequencing dependency … is unchanged", followed by a dangling "One", contradict `spec.md:36-38` (three audits ran) and `plan.md:44-51` (dependencies met) — the gate a run reader checks first — Severity: minor — Class: blocking (internal inconsistency on a gate) — Required fix: delete the two stale clauses and the fragment.

N2. stale-repair-claim — `acceptance.md:386` vs `spec.md:26` — HISTORY v0.2.3 says `strict=false` was corrected to `docs-i18n-check.yml:75`; the criterion body still cites `:71-74` (measured: line 75 carries `strict=false`, 71 is the `elif`) — Severity: minor — Class: optional — Required fix: change `:71-74` to `:71-75` or cite by symbol.

N3. stale-t1269 — `acceptance.md:29-34` — says t1269 "has not landed"; it has (`171634ee7`, `internal/spec/lint_vacuous_assertion.go`). A tree-built `moai spec lint` (see Gaps) emits one `VacuousTestAssertion` warning at `acceptance.md:15`, the deliberate bad-form illustration — Severity: minor — Class: optional — Required fix: restate the block as live, and either reword line 15's example or record the warning as an accepted false positive.

N4. D7-SHOULD — `acceptance.md:547` — `SPEC-V3R3-DOCS-PARITY-001` is called "an owner already", but no such directory exists under `.moai/specs/` (the workflow comment at `docs-i18n-check.yml:74` says "or equivalent") — Severity: minor — Class: optional — Required fix: say the pointer names a planned route, not an existing SPEC, or name the real owner.

N5. stale-carve-note — `acceptance.md:41-48` — "Tier M, 16/16, currently 7 criteria" is present-tense and wrong (Tier L, 10) — Severity: minor — Class: optional — Required fix: mark as the v0.1.x carve-time state.

N6. attribution-sha — `spec.md:95`, `:168`, `acceptance.md:138`, `:275`, `research.md:41`, `:93` — v0.2.5 readings are attributed to `develop = origin/develop = 37dc766b9`, which is not an ancestor of this branch (`git merge-base --is-ancestor 37dc766b9 HEAD` → 1); the premise report measured `e9577de4f`. Values reproduce identically at both (45,810; `:137`) — Severity: minor — Class: optional — Required fix: cite the commit actually measured.

N7. both-present-launcher — `spec.md:87-88`, `design.md:90-91` — the loop concatenates both files when both exist (`codex_launcher.go:126-138`); REQ-IFU-007 scopes the advisory to "only `CLAUDE.local.md`", while design shape 1 would emit whenever `CLAUDE.local.md` is in the read set. Either is defensible; the SPEC does not say which — Severity: minor — Class: optional — Required fix: state the both-present launcher behaviour in one sentence.

N8. pre-satisfied-page — `acceptance.md:310-313` — `advanced/codex-dual-harness.md` already carries `AGENTS.local.md` in all four locales (count 1 each), so that page passes the grep half with no M4 work; the other 20 files are 0 and the criterion as a whole stays red — Severity: minor — Class: optional — Required fix: record the pre-counts beside the `## ` baseline.

## Regression Check

- iter3 D3-residual — RESOLVED (§3 above).
- iter3 N2 (`:71-74`, 19 vs 20) — PARTIALLY RESOLVED: the count is fixed (`acceptance.md:405` "all 19 workflow files"); the line citation is not (N2 here).
- iter3 optional DOCS-PARITY pointer — applied (`acceptance.md:545-550`), but the target does not exist (N4).
- iter3 N1 — still resolved (AC-IFU-011 marker fires, DualFileMatrix passes; §1 table).

## Recommendation

**FAIL, with a STOP signal and a scope-reduction proposal.** The score fell 0.88 → 0.84. The drop is not
revision churn; it is a premise that changed underneath the SPEC when the parent landed. That is the case the
LEAN clause exists for, so the proposal is scope reduction rather than a fifth pass:

1. **Split M3 out.** REQ-IFU-021 and REQ-IFU-022 (with AC-IFU-007, AC-IFU-024) move to a follow-up SPEC or card
   blocked on worktree reception of `AGENTS.local.md` (D2 option b). What remains — M1 verb and advisories,
   M2 launcher advisory, M4 docs — is implementable on the current tree today, measured in §1.
2. If the operator prefers to keep M3 here, apply D1 in full and D2 option (a) before run.
3. Either way, fix D3 (a three-clause delete). N2-N8 may travel with it.

A confirming check after the split is mechanical: the §D.2 diff stays empty over the reduced set, `spec.md:34-36`
no longer mentions an unmet dependency, and `grep -n 'wc -m < AGENTS.local.md' acceptance.md` returns nothing.

## Gaps

- Lint provenance: the installed `moai` (`v3.2.0-rc.16`, commit `a8a9b9376`) is a strict ancestor of HEAD with
  8 later commits under `internal/spec/`, so its clean result is not cited. The tree-built binary
  (`go build -o <scratch>/moai-tree ./cmd/moai`, HEAD `a73c78d1a`) reported `0 error(s), 1 warning(s)`
  (N3). The exit status of that run was not captured.
- The force-tracked worktree geometry (D2) was not measured here; it is the measurement D2's option (a) asks for.
- No `gh` reads; `develop` protection and Vercel configuration were not re-read (they are close-time duties).
- Cross-model audit not run: no `audit_model` / gate is configured in `.moai/config/sections/`.
- `mcp__moai__spec_audit` (project_root = this worktree): one INFO `EraAutoDetected` (H-5), no drift.

## Residual risk

- If D2 option (a) is chosen and reception succeeds, the nested-worktree ancestor walk still loads the
  primary checkout's copy as well (the t1219 duplicate) — which copy the primary holds depends on the primary
  sitting on `main`, where the migrated file will not exist until a release.

🗿 MoAI
