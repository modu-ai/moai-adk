# t2053 migration flow — self-audit of left-over states

Author: manager-develop (implementer of the t2053 fixes). Tree: `WT-update-progress` at `78aa44f36a`
(J1 fix `299c6061d4` included). Not committed.

## Method and evidence legend

I read `internal/cli/worktree/migrate.go`, `movelist.go`, `classify.go`,
`internal/cli/update_worktree_migration.go` and `internal/hook/worktree_registry.go`, then walked the
flow one step boundary at a time and asked: if the process (or one git call) stops here, what is on disk,
and what does each entry path do with it?

Evidence column:

- **T** — a test I wrote and ran in this card exercises that exact state and path (pass observed).
- **R** — read from the code only; I did not execute it. Treat as a hypothesis until someone runs it.
- **R?** — depends on an assumption I did not verify (named in the row).

Entry paths: **APPLY** (`moai worktree migrate --apply`), **UPDATE** (`runUpdateConsolidation`, called by
`moai update`), **ROLLBACK** (`migrate --rollback <journal>`), **VERIFY** (`migrate --verify`),
**RERUN** (a later APPLY or UPDATE over the state left by an earlier run).

Verdict vocabulary: **CONVERGES** (state ends consistent), **REPORTED** (ends non-zero or prints a
failure/pending line, state left for the next entry), **SILENT** (exit 0 and nothing flags it — a false
success), **N/A** (path cannot reach the state).

Tree T moves from source S to destination D. "Registration" = git's `worktrees/<id>/gitdir`;
"state-file" = the hook worktree registry entry keyed by path; "card" = factory card `worktree_path`.

## Flow, as implemented

1. `planMigrationFor`: `git worktree list --porcelain` (bounded) → per-tree skip reasons.
2. `migrateOneTree`: `captureManifestChecked(S)` → `EnsureInstanceToken(S)` (token file in admin dir) →
   `journal.pending` (whole-file rewrite) → move-list `pending` intent line → `forwardGitMove(S→D)` →
   `RekeyWorktreeEntry(S→D)` → `rebindCardForPath(S→D)` → `captureManifestChecked(D)` + compare →
   move-list `moved` line.
3. Engine: `journal.settle` (whole-file rewrite) after each tree.
4. Any failure after the forward move → `reverseAfterFailure` → `m6ReverseMove` (identity check, repair
   if forward move was unregistered, `reverseGitMove`, inverse re-key, inverse rebind, `reversed` line).
5. ROLLBACK: reverse over journal steps (identity, manifest for `moved` steps, repair, reverse move,
   `rekeyAndRebindBack`, `restored` line).
6. VERIFY: reads every `*.json` in `.moai/state/worktree-migrate/`.

## State table

### A. Before the forward move

| # | Left-over state | How reached | APPLY / UPDATE | ROLLBACK | VERIFY | RERUN | Gap |
|---|---|---|---|---|---|---|---|
| A1 | Planning failed (list timeout/non-zero/unparsable) | git list fails | REPORTED: engine returns the error, nothing touched (**T**, F1) | N/A | N/A | next run replans | none |
| A2 | Manifest capture timed out | git call over `gitCommandTimeout` | skipped with reason `manifest unavailable`; counted `skipped`, exit 0 (**R**) | N/A | N/A | retried by next run | skip is exit 0 by design; acceptable, but not distinguishable from "ineligible" in the exit code |
| A3 | Token written, nothing else | kill after `EnsureInstanceToken` | token is idempotent; no visible state | N/A | N/A | CONVERGES (token reused) (**R**) | none |
| A4 | Journal file truncated/partial | kill (or ENOSPC) inside `journalWrite` = `os.WriteFile` (truncate then write) | n/a after kill | REPORTED: `parse the journal` error — but the recovery handle for ALL steps in that journal, including trees already `moved` earlier in the run, is lost (**R**) | SILENT: `json.Unmarshal` failure → `continue`; the broken journal is invisible, exit 0 (**R**) | RERUN ignores journals | **G1, G2** |
| A5 | Journal `pending` written, intent line failed | move-list write error | tree marked `failed`, nothing moved, journal settled `failed` (**R**) | converged: dest absent, source carries ident, registered | N/A | retried | none |
| A6 | Journal `pending` + intent line `pending`, no move | kill between intent line and `git worktree move` | n/a after kill | CONVERGES to "nothing to do" after J1 (dest absent, source ident + registered) (**T**, converged-step guard) | pending steps are not examined (see G3) | RERUN re-plans T as eligible and moves it; stale pending journal and line stay behind (**R**) | **G3**, **G9** (MovedBy protects a tree that never moved) |

### B. The forward `git worktree move`

| # | Left-over state | How reached | APPLY / UPDATE | ROLLBACK | VERIFY | RERUN | Gap |
|---|---|---|---|---|---|---|---|
| B1 | Dir at D, registration still names S (S `prunable`), state-file/card name S | move killed after rename, before registration update; process continued (timeout error) | REPORTED→CONVERGES: `reverseAfterFailure` repairs D's registration then reverses; if repair fails → `failed`, tree stays at D (**T**, H1 apply tests) | CONVERGES via repair-then-reverse; repair failure → pending + non-zero (**T**, H1 rollback tests) | SILENT: pending/failed steps skipped (**R**, G3) | SILENT: S is listed `prunable` → `skip "Git lists this tree as prunable"`; D is not registered so the planner never sees it; exit 0 (**R**) | **G3**, **G4** |
| B2 | Same as B1 but the whole process was killed | SIGKILL / power loss | n/a | CONVERGES, only if the operator knows to run `--rollback <that journal>` (**T** for the repair mechanics) | SILENT (G3) | SILENT (G4) | **G4**, **G5** (no auto-discovery of pending journals) |
| B3 | Dir at D, registration at D, state-file/card still name S | forward move completed in git; process killed or timed out before `RekeyWorktreeEntry` | timeout path (process alive): dest exists → reverse → CONVERGES (**R**) | CONVERGES by moving back (pending step reverses unconditionally; "no entry keyed to" tolerated) (**R**) | SILENT (G3) | **SILENT**: D is a registered home-root tree; `p.destination` for D resolves to D itself → `skip "destination already exists"`; state-file and card keep naming S forever (**R?** — requires `MOAI_HOME` path to contain `/.moai/worktrees/` for the planner's path filter to even list D) | **G4**, **G6** |
| B4 | Forward move timed out, dest absent | killed before rename | `failed` row, `failed` ML line; tree untouched (**R**) | converged (nothing to do) | N/A | retried | none |

### C. After the forward move

| # | Left-over state | How reached | APPLY / UPDATE | ROLLBACK | VERIFY | RERUN | Gap |
|---|---|---|---|---|---|---|---|
| C1 | State-file re-key failed (not "no entry") | registry lock/IO error | reverse order runs → `reversed`, exit non-zero (**R**) | n/a | `reversed` skipped | tree eligible again | none |
| C2 | Card rebind failed | factory DB locked/IO | reverse order → `reversed` (**R**) | — | — | — | none |
| C3 | Reverse succeeded in git but the INVERSE re-key or rebind failed | `m6ReverseMove` ends with `_ = hook.RekeyWorktreeEntryInverse(...)` and `_ = rebindCardForPath(...)` | **SILENT inside a reported row**: row says `reversed`, `reversed` ML line released the `MovedBy` protection, but state-file and/or card still name D, which no longer exists (**R**) | a later `--rollback` of that journal sees dest absent, source registered → runs the idempotent inverses → CONVERGES (**T**, J1 shape) | skipped | RERUN: forward re-key finds "no entry keyed to S" (tolerated); card stays bound to D (**R**) | **G7** |
| C4 | Manifest mismatch or capture timeout after the move | content changed, or git timed out | reverse order (**T**, existing tests + timeout test) | — | — | — | none |
| C5 | `moved` ML line failed to write | IO error | row `moved`, run returns `move-list writes failed` error (**R**) | `moved` step with manifest verified normally | intact | tree already at D | REPORTED, none |
| C6 | `journal.settle` flush failed | IO error | engine aborts the loop and returns the error; remaining trees unprocessed; on-disk step still `pending` although tree is `moved`/`reversed` (**R**) | pending reverses WITHOUT the manifest check — a moved tree with later user changes is moved back anyway (directory move, content intact) (**R**) | pending skipped (G3) | tree at D already; skipped as "destination already exists" | **G3**; minor |

### D. The reverse move (`m6ReverseMove`)

| # | Left-over state | How reached | APPLY / UPDATE | ROLLBACK | VERIFY | RERUN | Gap |
|---|---|---|---|---|---|---|---|
| D1 | Destination identity differs | foreign tree at D | `refused` ML line, outcome `failed`, tree stays; reason text says "the reverse move failed, the tree stands at D" (misleading but non-silent) (**R**) | `refused` + non-zero (**T**, existing test) | — | — | none |
| D2 | `TreeIdent(D)` returns an ERROR | unreadable `.git` pointer/token | **reverse proceeds anyway** (`err == nil && mismatch` is the only refusal) (**R**) | refuses (`terr != nil` ⇒ mismatch) (**R**) | — | — | **G8** (inconsistent identity rule) |
| D3 | Reverse move failed; tree stays at D, registration/state/card consistent at D | git error or timeout, no rename | `failed`, `pending` ML line (D→S orientation) (**T**) | CONVERGES (moves back) (**T**) | SILENT: `failed` skipped (G3) | SILENT: D "destination already exists" (the tree sits in the target location, so this is migrated-by-accident; ML last line stays `pending`) | **G3**, **G10** |
| D4 | Reverse rename done, error returned, registration OK | killed after rename+registration | `treeMovedBack` → true → `reversed` (**T**, F2) | same (**T**, G2) | — | — | none |
| D5 | Reverse rename done, registration still names D, repair fails or hangs | the J1 trigger | `failed` + `pending` line; tree at S, S unregistered (**T**, G1) | CONVERGES via repair (**T**, J1) ; repair failure → non-zero, no `restored` (**T**) | SILENT (G3) | SILENT: git lists D as `prunable` (skip), S is not registered so the planner never sees the tree at all; exit 0 (**R**) | **G3**, **G4** |
| D6 | Reverse killed mid-flight (process dead) | SIGKILL | n/a | as D3–D5 depending on what git finished (**R**) | SILENT | SILENT | **G4**, **G5** |

### E. ROLLBACK itself

| # | Left-over state | How reached | ROLLBACK next run | Gap |
|---|---|---|---|---|
| E1 | Killed after some steps restored | SIGKILL mid-loop | restored steps are "converged" (dest absent, source ident, registered); remaining steps restored; summary `restored N` counts only the new ones (**R**) | none |
| E2 | Step `moved` with manifest mismatch | tree edited after the move | `refused` line + non-zero; step stays (**T**, existing) | none |
| E3 | Source path now occupied by another tree | operator action | reverse move fails; `treeMovedBack` rejects (source ident differs) → pending + non-zero (**R**) | none |
| E4 | Step has empty `ident` (hand-written/legacy journal) | not produced by this engine | dest present: reverses without identity check; dest absent: silently skipped, which can hide a stranded S (**R**) | **G11** (low: no producer in this engine) |
| E5 | Rollback never marks the journal consumed | by design | a rolled-back journal keeps its `moved` steps; VERIFY then reports them `absent` (non-zero) until the journal file is removed (**R**) | **G12** (false FAILURE, opposite sign) |
| E6 | Lock held | another mover | REPORTED: `migration lock … held`, moves nothing; UPDATE degrades this to a warning (**R**) | none |

### F. VERIFY

| # | Behaviour | Gap |
|---|---|---|
| F1 | Examines only steps with `outcome == "" or "moved"`. `pending` (killed mid-move), `failed` (reverse failed, tree stranded) and `reversed` steps are never examined, and a repo whose only damage is a stranded tree prints `0 recorded move(s) intact, 0 mismatched` and exits 0 (**R**) | **G3** |
| F2 | Invalid/unreadable journal files are skipped with `continue`, no output (**R**) | **G2** |
| F3 | "Intact" = directory exists + `TreeIdent` equals the journal ident. It does not check git's registration, the state-file entry or the card binding (**R**) | **G13** |
| F4 | Verifies the whole journal history, so a tree that was legitimately removed later (`worktree done`) is reported `absent` forever (**R**) | covered by G12 |

## Entry-path summary (does every path converge or report?)

| Path | Converges | Reports a failure | False success exists? |
|---|---|---|---|
| APPLY (process alive) | yes for every failure after the forward move (reverse order), with the H1/J1 repairs | non-zero exit on `failed`/`reversed` | only C3 (inverse re-key/rebind swallowed) |
| UPDATE | same engine | prints `Worktree migration failed …` and returns **nil** (design AC-WHR-019(j): update never fails) — reported in text only | stranded states are printed as `skipped` on later updates, never `failed` (G4) |
| ROLLBACK | yes for B1–B3, D3–D6 and A6 when the operator supplies the right journal | non-zero on any unconfirmed step (after J1) | E4 only |
| VERIFY | no (read-only) | only for `moved` steps | **yes — G2, G3** |
| RERUN | no: the planner never reads journals or pending move-list lines | none | **yes — G4, G5, G6** |

## Gaps (ranked by what a user can be left with while every command exits 0)

- **G4 (high) — RERUN does not converge or flag a killed-mid-flight tree.** B1/B2/B3/D3/D5/D6 leave the
  tree at an unregistered or half-registered location. `planMigrationFor` sees either a `prunable`
  record (skip) or a registered home-root tree whose destination equals itself (skip "destination
  already exists"). Neither is `failed`; exit 0. R?: whether D is even listed depends on the home root
  path containing `/.moai/worktrees/` (the planner's path filter); in a custom `MOAI_HOME` the tree is
  invisible to RERUN altogether. Evidence is code reading only.
- **G5 (high) — no entry path discovers pending journals or pending move-list lines on its own.**
  Convergence requires the operator to find the right `…/worktree-migrate/<ts>.json` and pass it to
  `--rollback`. APPLY/UPDATE start without a pre-flight "unsettled steps exist" check. After a kill,
  the next `moai update` proceeds as if nothing happened.
- **G3 (high) — VERIFY ignores `pending`, `failed` and `reversed` steps** (F1). A stranded tree yields a
  clean verify.
- **G2 (medium) — VERIFY silently skips unparsable journals** (F2), and ROLLBACK reports a parse error
  without telling the operator which tree it concerns.
- **G1 (medium) — journal rewrite is not atomic** (`os.WriteFile` truncates). A kill inside a flush
  corrupts the recovery handle for every step in that run, including trees already moved. Fix shape:
  write a temp file in the same directory and `rename`.
- **G7 (medium) — `m6ReverseMove` swallows inverse re-key/rebind errors** (C3) and reports `reversed`,
  releasing `MovedBy` protection while the state-file/card may name a path that no longer exists.
  ROLLBACK already surfaces the same errors (J1 helper); the apply-time reverse should do the same
  (record `pending`, count the tree failed).
- **G6 (medium) — a tree that git moved but whose process died before re-key/rebind** (B3) has no
  forward-completion path; only rollback exists. A rerun cannot finish the migration or fix the state.
- **G8 (low) — identity rule differs**: apply-time reverse proceeds when `TreeIdent(D)` errors, rollback
  refuses (D2).
- **G9 (low, safe direction) — `MovedBy` protects a tree whose latest ML line is `pending` and which is
  registered at either end**, so a tree that never moved (A6) stays protected from cleanup until a later
  `moved`/`reversed`/`restored` line lands.
- **G10 (low) — a tree stranded at D with consistent registration (D3) is exactly the target layout**;
  RERUN reports it `skipped`, and the last ML line stays `pending`. Harmless to the tree, but the record
  never settles.
- **G11 (low) — empty-ident steps** (E4) bypass the identity check and can be skipped silently.
- **G12 (low) — VERIFY after ROLLBACK or after a legitimate `worktree done`** reports a false FAILURE
  (journals are never marked consumed).
- **G13 (low) — VERIFY checks directory + identity only**, not registration/state-file/card (F3).

## What I did not analyse

- Concurrent movers other than the migration lock (the lock is an OS flock held for the process
  lifetime; its behaviour on NFS/remote filesystems is not examined).
- Interactions with `moai worktree done|clean|sweep` between a kill and a rerun, beyond the `MovedBy`
  guard read above.
- Windows path-spelling behaviour of `canonicalTreePath` for a missing path (macOS `/var` vs
  `/private/var` was handled by the tests' real-path fixtures; a missing path falls back to
  `filepath.Clean`, so a non-canonical spelling of a missing source could make `registeredAt` answer
  false). **R?**, unverified.
- Journal filename collision when two runs start within the same second (`20060102T150405.json`):
  the second would overwrite the first. Unlikely under the lock; not tested.
- The CLI-side changes of this card (update/init progress, `init_gitdetect` timeout) — outside the
  migration flow.
