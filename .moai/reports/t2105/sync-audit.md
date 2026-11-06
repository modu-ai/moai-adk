auditor-model: claude-opus-5-5

# Sync audit — card t2105 (Class B, one SPEC frontmatter change)

verdict: PASS-WITH-DEBT
audited_sha: 00cc223852b9c9c0957b9ca4b55cff049049ac11

- Score: 90/100 (weighted: Functionality 90 x0.40, Security 100 x0.25, Craft 80 x0.20, Consistency 90 x0.15)
- Base: 9cf8a1fc14. Audit mode: Claude-only, in-session (no cross-model / paid MCP tool called, per dispatch).
- Iteration: 1 (first audit of this card).
- Rules applied: verification-claim-integrity.md §1.1 surface 3 (defect claim needs the tool), §2.2 (tool provenance), §3.1 (refusal = Gap); spec-frontmatter-schema.md § Optional Fields (`lint.skip` "only for documented debt"), § Non-transition frontmatter corrections.

- debt: D1 dispose_in=sync Reword the frontmatter comment (and record in progress.md) so it states the real cause: 55b20438fd is a squash merge (#1930, one parent) whose body carries both `Authored-By-Agent: manager-spec` (plan sub-commits, first) and `Authored-By-Agent: manager-develop` on the sub-commit that says "spec.md status draft -> in-progress"; the rule reads only the FIRST trailer, so the warning is a rule false positive on squash commits, not an ownership violation. Open a follow-up card for the rule (attribute the trailer of the sub-commit that carries the status change, or treat multi-trailer squash bodies as unmeasured).
- debt: D2 dispose_in=sync Remove `lint.skip: [OwnershipTransitionInvalid]` from SPEC-FACTORY-COMPLETION-RECOVERY-001 at that SPEC's next status transition (its sync commit); the skip returns before the git lookup, so until it is removed a genuine future mis-owned transition on this SPEC is silent.

## Findings (structured defect-list)

- F1 [Medium] [optional, carried as debt D1] [in-diff] `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md:15` (comment) and the HEAD commit message — the debt is characterized as "the draft -> in-progress commit carries Authored-By-Agent: manager-spec (the matrix expects manager-develop)". Literally true of the squash commit's first trailer, but the squash body shows the transition sub-commit was `manager-develop` (the expected owner). The recorded debt is a rule false positive, not a history defect. Confidence high (reproduced below). Not blocking: the mechanism chosen is still the right one and the gate outcome is correct; the documentation of the debt is what is wrong. Required fix: D1.
- F2 [Low] [optional] [in-diff] HEAD 00cc223852 carries no `Authored-By-Agent:` trailer, so the claim that manager-spec authored the edit is not witnessed in git. No rule requires a trailer for a non-transition edit; recorded for traceability only.
- F3 [Low] [optional, carried as debt D2] [in-diff] The skip disables the whole rule for this SPEC (early return at `internal/spec/lint_ownership.go:378`), not just the one historical record; the exit condition is prose and nothing enforces it.
- F4 [Info] [optional] [outside] The same first-trailer-wins parse may also mis-attribute other squash commits (e.g. the advisory SPEC-PROGRESS-RECORD-IO-001 record, commit 5ef49c4 — single parent aef730d091, trailer list not inspected: see Gaps). Not part of this card.

No blocking finding. Must-pass dimensions (Functionality, Security) pass.

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | three-SPEC gate `baseline: OK ... 0 non-advisory tracked`, exit 0 (E2); skip honored as Info `OwnershipTransitionSkipped` |
| Security (25%) | 100 | PASS | frontmatter-only change; no code, secret, input surface |
| Craft (20%) | 80 | PASS | correct mechanism, exit condition stated; debt record mis-states the cause (F1) |
| Consistency (15%) | 90 | PASS | nested `lint:`/`skip:` shape matches SPEC-V3R3-CI-AUTONOMY-001 precedent; no `LintSkipFlatForm` warning |

## Claim

1. The diff is exactly one file, frontmatter only; `status`, `version`, body, HISTORY untouched; baseline file not in the diff.
2. After the edit the three-SPEC baseline gate reads OK with 0 non-advisory warnings; only SPEC-FACTORY-COMPLETION-RECOVERY-001's warning was non-advisory; the other two are advisory because their status is terminal (completed, superseded).
3. The skip is honored by the current rule and by t2087's changed rule path (140c1ca406).
4. The lane's root-cause description is incomplete: the transition was authored by manager-develop inside a squash merge; the lint attributes it from the squash's first trailer (F1).

## Evidence

E1 — diff (`git -C <tree> diff 9cf8a1fc14..HEAD --stat`):
```
 .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md | 6 +++++-
 1 file changed, 5 insertions(+), 1 deletion(-)
```
Full diff: `-updated: 2026-11-02` / `+updated: 2026-11-06`, and after `tier: L`:
```
+# t2105 documented debt: the draft -> in-progress commit 55b20438fd carries Authored-By-Agent: manager-spec (the matrix expects manager-develop); history is not rewritten. Remove this skip once the next status transition by its owner lands.
+lint:
+  skip:
+    - OwnershipTransitionInvalid
```
`.moai/spec-lint-baseline.json` not in the diff; its content: `"rules": { "TierArtifactMissing": 4 }`, tree_sha d5638ba81.
Precedent (`grep -n -B2 -A4 "^lint:" .moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md`): `23:lint:` / `24-  skip:` / `25-    - StatusGitConsistency` — same nested shape.

E2 — `moai spec lint SPEC-PROGRESS-RECORD-IO-001 SPEC-FACTORY-COMPLETION-RECOVERY-001 SPEC-LSPMCP-001 --baseline .moai/spec-lint-baseline.json` (exit 0):
```
WARNING   OwnershipTransitionInvalid  .../SPEC-PROGRESS-RECORD-IO-001/spec.md  1  ... transition "implemented" → "completed" expected owner "orchestrator|manager-docs" but commit 5ef49c46... (Authored-By-Agent: manager-spec) maps to "manager-spec"
INFO      OwnershipTransitionSkipped  .../SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md  1  ... ownership transition check skipped via lint.skip: [OwnershipTransitionInvalid] opt-out
WARNING   OwnershipTransitionInvalid  .../SPEC-LSPMCP-001/spec.md  1  ... transition "archived" → "superseded" expected owner "manager-spec" but commit 94d9fccf... (Authored-By-Agent: manager-develop) maps to "manager-develop"
0 error(s), 2 warning(s)
baseline: OK — .moai/spec-lint-baseline.json
  decreased: TierArtifactMissing recorded 4 -> current 0 (-4)
  inventory: 2 warning(s) total (advisory included), 0 non-advisory tracked across 1 recorded rule(s)
```
`--baseline` with `--json` is refused by the tool (exit 3: "cannot use --baseline with --json"); the JSON form was run without `--baseline` (exit 0): both `OwnershipTransitionInvalid` findings carry `"advisory":true`; the FACTORY-COMPLETION-RECOVERY entry is `"severity":"info","code":"OwnershipTransitionSkipped"`. No `ParseFailure` and no `LintSkipFlatForm` finding, so the frontmatter parses and the nested skip binds.

E3 — why advisory (source, this tree): `internal/spec/lint.go:319-323` demotes when `GrandfatheredEra` OR `TerminalStatus`; `applyEraDemotion` (lint.go:424-425) marks every Warning `Advisory = true` on a demoted SPEC; `CompareBaseline` (`lint_baseline.go:191`) counts only `!f.Advisory`. `grep` of the two other SPECs: PROGRESS-RECORD-IO-001 `status: completed`, LSPMCP-001 `status: superseded` — both terminal, so demoted. FACTORY-COMPLETION-RECOVERY-001 is `status: in-progress`, created 2026-10-06, phase v3.2.0 — not terminal, not grandfathered, so its warning was non-advisory. (`eraDemotableCodes` only demotes ERRORs of two codes; it is not the path here — the warning-advisory branch is.) Skip handling: `lint_ownership.go:377-388` returns Info `OwnershipTransitionSkipped` before the git lookup, and `lint.go:313` `applylintSkip` also drops any finding with a listed code.

E4 — commit facts:
- `git show -s --format=%P 55b20438fd` → one parent `2a1898ea22…` (squash merge, subject ends "(#1930)").
- Trailers in its body (grep of `git show -s --format=%H%n%s%n%b 55b20438fd`): first at body line 10 `Authored-By-Agent: manager-spec`; mixed manager-spec / manager-develop / manager-docs thereafter.
- Body lines 48-61: sub-commit "M2-B run consolidation operation and worker-liveness predicate" states "spec.md status draft -> in-progress (first run-phase M2-B commit)." followed by `Authored-By-Agent: manager-develop`.
- `git show 55b20438fd -- .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md` contains `-status: draft` / `+status: in-progress` (and version 0.1.0 → 0.3.13).
- Parser: `parseAuthoredByAgent` uses `FindStringSubmatch` (first match only) — `lint_ownership.go:260-266`. Hence manager-spec is reported.

E5 — t2087 overlap (read-only): `git show 140c1ca406:internal/spec/lint_ownership.go` keeps `hasOwnershipSkipOptOut` and the early Info return ahead of `cachedOwnershipTransition`; lookup now goes through `batchedOwnershipTransition` with per-SPEC fallback. `git grep … 140c1ca406 -- internal/spec/lint.go` shows `applylintSkip(ruleFindings, doc.LintSkip)` still at line 313 and the ownership rule still on the per-doc path (not a `crossSPECRule`). The opt-out is honored under t2087.

E6 — HEAD message (`git show -s --format=%H%n%s%n%b HEAD`): no `Authored-By-Agent:` trailer (F2). `git status --short` empty.

## Baseline-attribution

- Tree: worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2105`, HEAD 00cc223852b9c9c0957b9ca4b55cff049049ac11, clean; base 9cf8a1fc14.
- Judging build (§2.2): installed `moai` v3.2.0-rc.42, `v3.2.0-1344-g646955327e`, built 2026-11-06T05:06:19Z. Not a build from this tree; whether 646955327e is an ancestor of HEAD was not checked (Gap). The diff touches no Go code, and the rule source read in E3 is this tree's, so the lint behavior relied on is the tree's code path; the gate outputs in E2 are from the installed build.

## Gaps

- Pre-edit control not re-run by the auditor: reproducing the `EXCEEDED +1` reading would need the pre-edit SPEC in the tree (checkout/restore is outside this audit's permitted actions). Accepted from the lane's record only; the source path (E3) and the squash trailer order (E4) make it consistent.
- Lane claim 4 (whole corpus ~32 s, `baseline: OK`, control EXCEEDED +1 under t2087-merged scratch binary) is lane-measured only; not reproduced (no merge/build permitted).
- Ancestry of the installed build commit 646955327e vs HEAD not checked.
- A compound `git log` command with `%(trailers:…)` for commit 5ef49c4 was refused by the worktree-isolation guard; fallback was `git show -s --format=%P` only (parent count). Trailer list of 5ef49c4 not inspected (F4 left unconfirmed).
- No test suite run (dispatch constraint); no Go code changed.
- Cross-model audit not run by instruction; no codex receipt exists for this audit.

## Residual-risk

- While the skip stands, a real mis-owned transition on this SPEC would not be reported (D2).
- The first-trailer-wins parse will keep producing false `OwnershipTransitionInvalid` on squash merges whose first sub-commit is a plan commit; each will tempt another per-SPEC skip unless the rule is fixed (D1 follow-up).
- Whole-corpus gate result rests on the lane's run against an unmerged t2087 combination; after t2087 lands, the corpus gate should be re-read once.
- Scratch outputs under the session scratchpad are not cited; the deciding lines are copied above.
