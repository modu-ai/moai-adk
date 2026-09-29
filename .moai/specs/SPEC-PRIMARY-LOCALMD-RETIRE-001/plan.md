# plan.md — SPEC-PRIMARY-LOCALMD-RETIRE-001

> Stateless artifact (no `status:` frontmatter per the schema SSOT § Artifact Statelessness).
> Tier M. Ordering of milestones follows decision-reversibility: the primary-side
> preserve-and-retire sequence (highest change-likelihood, least reversible) leads;
> mechanical verification closes.

## §A Context

### A.1 Situation

Card t1317 retires the primary checkout's legacy `CLAUDE.local.md`. The card body (Korean,
SSOT) fixes the procedure: confirm the t1290 reception gate, then retire per §0.4's
"별도 전환" clause — preserve under `.moai/state/`, do not delete outright — as ONE policy
commit, strictly after t1279's D-scope landing. t1279's §① measured the stake:
+20,855 tokens/session in every nested worktree session via upward traversal (probe 3:
41,354 first-turn input tokens vs probe 5: 62,209 tokens, `.moai/reports/t1279/verdict.md`).

### A.2 Verified preconditions (baseline-attributed)

All measured against THIS tree at `68e37864a`, this session. Commands + outputs recorded in
research.md §B; the worktree-session guard refused re-measurement of the primary side
(`git -C` cross-tree redirect denied), so primary-side pre-state is carried
delegation-verified evidence and re-verified at run entry (M1 step 1).

| # | Precondition | Evidence |
|---|--------------|----------|
| 1 | t1279 D-scope landed | `grep '^status:' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/spec.md` → `status: completed` (observed this run) |
| 2 | t1290 merged | `git merge-base --is-ancestor c13cee6d5 HEAD` → true (observed this run) |
| 3 | Reception gate passed | `SPEC-LOCAL-INSTR-RECEPTION-001/progress.md` closing entry: AC-LIR-001~009 all pass (2026-09-28, audit-read 646ae8302); AC-IFU-007 `git show HEAD:AGENTS.local.md \| wc -m` → 37061 ≤ 39,999 |
| 4 | Worktree geometry | `AGENTS.local.md` 35,897 B present; `CLAUDE.local.md` absent (observed this run, `ls`) |
| 5 | `.moai/state/` gitignored | `git check-ignore .moai/state/retired/x` → ignored (observed this run) |
| 6 | Primary pre-state | ` M CLAUDE.local.md` permanent status; tracked on `main` (delegation-verified; re-check at M1) |

### A.3 Design decisions (card-inherent choices take the recommended option)

| # | Decision | Choice | Reasoning |
|---|----------|--------|-----------|
| 1 | Where the ONE policy commit lands | local `main` in the primary checkout (recommended, taken) | `main` is the only branch tracking `CLAUDE.local.md`; committing the deletion there removes the working-copy load AND makes the deprecated committed copy unreachable — closing the §0.3 hazard permanently. Commits to the already-checked-out branch are permitted (`AGENTS.md` §2). Documented consequence: `0 N` divergence vs `origin/main` until lead batch-push (REQ-PLR-008). |
| 2 | Preserve target | `.moai/state/retired/CLAUDE.local.md`, byte copy + sha256 + provenance in progress.md (recommended, taken) | `.moai/state/` is gitignored machine-local scratch by design (verified this run); nothing canonical is lost — canonical home is develop's `AGENTS.local.md`, main's copy recoverable from git history. |
| 3 | Develop-side docs refresh | Same WT branch, minimal §0.4 revision (recommended, taken) | The 보존 sentence becomes false after conversion; leaving it would leave doctrine contradicting the tree. Scope held to ONE sentence-level revision; §0.1–§0.3 untouched (REQ-PLR-011). |
| 4 | Load-elimination AC shape | Mechanical absence REQUIRED + session probe OPTIONAL (recommended, taken) | A probe session spawn may be impractical mid-run; mechanical absence (`test ! -e` + porcelain empty) is binary and sufficient to prove the traversal source is gone, since upward traversal is file-presence-driven. Choice recorded in progress.md per REQ-PLR-010. |
| 5 | Run-phase entry preconditions as ACs | AC-PLR-007 (recommended, taken) | Ordering gates are cheap to re-check and the card marks them [HARD]; re-affirming at run entry converts the card's sequencing prose into binary gates. |

### A.4 PRESERVE list (scope discipline)

- `AGENTS.local.md` §0.1–§0.3 and every other section — ONLY the §0.4 sentence is edited.
- `SPEC-LOCAL-INSTR-RECEPTION-001/spec.md` frontmatter (the `status: draft` gap is
  out-of-scope observation).
- `internal/contract/frozen.go` and its tests (name list is correct as-is).
- All 10 `.moai/docs/*.md` files citing CLAUDE.local.md §-anchors.
- `origin/main`, `origin/develop` — no push from this lane.
- Every other working-tree file in the primary checkout — the policy commit stages
  `CLAUDE.local.md` by explicit pathspec only (REQ-PLR-005).

## §B Known Issues (filtered to applicable categories per the delegation template)

- **B2 Cross-SPEC policy conflict** — `AGENTS.local.md` §0.4 currently MANDATES preserving
  the primary copy ("별도 전환 전까지 보존한다"). This SPEC reverses that mandate by
  completing the conversion it reserves; the reversal is IN SCOPE (REQ-PLR-011) and must
  land in the same batch so the doctrine never contradicts the tree for long. No other
  retired/superseded SPEC governs the primary-side file.
- **B8 Working-tree hygiene** — the primary checkout is shared. Re-read
  `git status --short` immediately before staging; explicit pathspec only; never
  `git add -A` / `git commit -a` (`AGENTS.md` §2 sweep prohibition).
- **B11 AskUserQuestion prohibited** — any blocker (ordering gate failure, preservation
  mismatch, staleness divergence) returns a structured blocker report (REQ-PLR-012).
- **bash risk discipline** — the primary-side act is a state-mutating git operation in a
  shared checkout; every command single-purpose, no compound chains across the commit
  boundary, staleness re-read (REQ-PLR-006) between observation and commit.

## §C Pre-flight (run-phase, before M1)

```bash
# 1. Worktree + branch + HEAD (this tree)
git rev-parse --show-toplevel && git branch --show-current && git rev-parse --short HEAD

# 2. Ordering gate 1 — t1279 completed
grep '^status:' .moai/specs/SPEC-SESSION-DOUBLELOAD-001/spec.md   # expect: status: completed

# 3. Ordering gate 2 — develop-side reception end-state
git ls-files -- AGENTS.local.md CLAUDE.local.md                   # expect: AGENTS.local.md only
git merge-base --is-ancestor c13cee6d5 HEAD && echo GATE-OK

# 4. Primary pre-state (ExitWorktree round-trip; single commands, no compounds)
#    git branch --show-current        → expect: main
#    git status --porcelain -- CLAUDE.local.md  → expect: " M CLAUDE.local.md"
#    git ls-files -- CLAUDE.local.md  → expect: CLAUDE.local.md
```

Any expectation failing → stop, blocker report (REQ-PLR-012).

## §D Constraints (DO NOT VIOLATE)

- PRESERVE list §A.4 verbatim.
- Forbidden commands: `git restore CLAUDE.local.md` (any target), `git reset --hard`,
  `git stash`, branch switches in the primary checkout, `--no-verify`, `--amend`,
  force-push, `git add -A` / `git add .` / `git commit -a`, `git push` from this lane.
- Required: staleness re-read immediately before the primary-side commit (REQ-PLR-006);
  Conventional Commit subject; commit body names card t1317 + SPEC id;
  `Authored-By-Agent: manager-develop` trailer; `🗿 MoAI` trailer.
- The primary-side act is the ExitWorktree → act → EnterWorktree round-trip; the session
  never edits primary working files from inside the worktree (worktree-session guard
  refuses cross-tree git anyway — observed this session).

## §E Self-Verification

Per `manager-develop-prompt-template.md` §E with VCI 5-section format per item. Minimum
deliverables: E1 AC PASS/FAIL matrix (AC-PLR-001..008, commands + verbatim outputs + HEAD
SHA attribution); E6 branch HEAD + commit SHAs (local `main` new commit, WT branch commits,
unpushed counts — no push performed, per REQ-PLR-008); E7 blocker report if any.
Go-code E2/E3/E4/E5 items are NOT applicable — this SPEC's diff surface is one primary-side
file deletion plus one doctrine sentence revision; CI on the develop-side batch is the
cross-tree verdict (git-flow lane protocol §4).

## §F Milestones (priority-ordered by decision-reversibility, no time estimates)

- **M1 — Primary-side preserve + retire (the ONE policy commit).** Priority High.
  ExitWorktree → primary checkout → re-verify pre-state (branch `main`, ` M` marker,
  tracked) → capture working copy + sha256 → write `.moai/state/retired/CLAUDE.local.md` →
  `cmp` byte-identity → re-read HEAD + branch (staleness) → `git rm CLAUDE.local.md`
  (explicit pathspec) → re-read `git status --short` → commit (ONE commit, message naming
  card t1317 + SPEC-PRIMARY-LOCALMD-RETIRE-001) → post-state checks (AC-PLR-003/004) →
  EnterWorktree back. Covers REQ-PLR-003..009; AC-PLR-001..004.
- **M2 — Develop-side §0.4 doctrine refresh.** Priority High. In this worktree: revise
  `AGENTS.local.md` line 37 sentence per REQ-PLR-011 (Korean register, §0.1–§0.3 intact) →
  commit on the WT branch. Covers REQ-PLR-011; AC-PLR-006.
- **M3 — Ordering-gate records + load-elimination evidence + closure prep.** Priority
  Medium. Record AC-PLR-007 re-affirmation outputs in progress.md; run AC-PLR-005 mechanical
  absence + decide probe optional-or-executed; AC-PLR-008 divergence report (`0 N` count on
  local `main`); SPEC frontmatter `status: in-progress` transition (manager-develop, M1
  commit). Covers REQ-PLR-001/002/008/010/012; AC-PLR-005/007/008.

## §G Anti-Patterns (named)

- Restoring the deprecated copy "to check something" (REQ-PLR-007; the 2026-09-07 incident).
- Sweep-staging the primary checkout because "the tree is mine" (B8; `AGENTS.md` §2).
- Pushing local `main` from the lane to "finish the job" (REQ-PLR-008; lead-batched).
- Editing §0.1–§0.3 while "in the file anyway" (PRESERVE list; scope discipline).
- Re-measuring primary pre-state from inside the worktree via `git -C` (the guard refuses;
  observed this session — use the ExitWorktree round-trip).

## §H Cross-References

- spec.md §C (REQ-PLR-001..012) · acceptance.md §D (AC-PLR-001..008) · research.md §B/§C
  (evidence + dependency sweep) · progress.md §E.1 (plan-phase signal) / §E.2-§E.4 (run/sync
  placeholders, owned by manager-develop / manager-docs).
- `.moai/reports/t1279/verdict.md` §① · card body (t1317) · `AGENTS.local.md` §0.
