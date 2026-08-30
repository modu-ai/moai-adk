---
id: SPEC-KANBAN-BOARD-001
title: "Implementation plan — six-column kanban board model with a single-origin board state store"
version: "0.6.1"
created: 2026-08-10
updated: 2026-08-14
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, plan, milestones, state-store, single-origin, sole-writer, atomicity, role-carrier, dispatch-rule"
tier: L
---

## §A. Context

The scope, the decided design, and the requirement set are `spec.md`. This file is the ordered path through them, plus the pre-flight measurements and the anti-patterns the run-phase implementer needs in hand before touching a file.

Milestones are ordered by **decision-reversibility**: the choices most likely to change on contact — the state store's location and resolution, then the card record and the consistency table — come first, so a wrong turn is discovered while it is still cheap. The mechanical work (template mirror, catalog regeneration, full verification) is deliberately last.

---

## §B. Known issues and unlanded ground

Eleven facts make this plan conditional. All eleven were verified, not assumed; items 1 and 8 were re-measured at v0.5.0 and both had changed.

1. **`SPEC-KANBAN-RENAME-001` HAS landed** — reversed at v0.5.0, and the gate that was written to halt now passes. Measured on the base branch at `c55c61aa5`: `git ls-tree -d --name-only origin/main internal/kanban` prints `internal/kanban`, and the same query for `internal/factory` prints nothing — the exact pair `AC-KB-001` requires, and `internal/kanban/` on disk carries `bootstrap.go`, `record.go`, `revision.go` and their tests. Through v0.4.0 this item read "has not landed… exists in this worktree as an untracked directory", which was true then. **The prerequisite SPEC's own frontmatter still reads `status: in-progress`, and that is not a landing signal** — it is lifecycle bookkeeping lagging the merge. The observation this plan relies on is the base-branch query, which is what `acceptance.md` §A.1 rule 6 makes authoritative precisely because a status field and a filesystem predicate can both be satisfied by something that has not landed. M0 still runs the gate (REQ-KB-002); it is now expected to **pass** rather than halt, and an M0 that halts is evidence the base branch moved, not that this note is stale.
2. **The pre-commit gate blocks commits.** The repository's `.git/hooks/pre-commit` invokes `moai gate`, which exits non-zero on pre-existing ast-grep findings. The documented bypass is `SKIP_MOAI_PRECOMMIT=1`. That defect is tracked separately and is not fixed here — it is recorded so an implementer does not diagnose it a second time.
3. **The board state store is gitignored, which means it is invisible to every git-based check.** `.gitignore` ignores `.moai/state/` (line 275) and `**/.moai/state/` (line 207). Nothing about the board's runtime state will appear in a diff, a `git status`, or CI. Every check on it must read the file, not the repository.
4. **Both lock substrates exist, and only one of them is cross-process.** `internal/spec/lock.go` (with `lock_unix.go` / `lock_windows.go`) is the flock-plus-atomic-create family REQ-KB-019 reuses. `internal/lockfile/lockfile_windows.go` is a `map[string]*sync.Mutex` whose own package comment records that writes across **different** OS processes are not protected and forbids upgrading it. Reaching for the second because it is simpler produces a lock that passes every same-process test and holds nothing in production.
5. **The atomic-write primitive already exists.** `internal/atomicfile` provides `Replace` (write-temp-then-rename) with the Windows handling already written, and `internal/verify/store.go` `Save` is a working same-directory example (`os.CreateTemp(dir, ...)` into the target's own directory, then rename). REQ-KB-018 reuses both shapes rather than authoring a third.
6. **The lock family REQ-KB-019 reuses does not clean up after a dead holder, and the gap is platform-asymmetric.** `internal/spec/lock_windows.go`'s own header states that stale-lock detection is a post-MVP enhancement and that cleanup is manual. On Unix the substrate holds `flock(2)` on an open descriptor, which the kernel releases on process exit — measured, `.moai/state/` carries fourteen orphaned zero-length `spec-close-*.lock` files that block nothing. On Windows the artifact **is** the lock and the block is permanent. An implementer developing on macOS will not reproduce this; REQ-KB-023 exists because of it.
7. **The resolution REQ-KB-005 reuses is unexported and returns a boolean.** Measured: `isPrimaryCheckout(projectDir string) (bool, error)` in package `hook`. There is no exported path-resolving helper, so "reuse it" is not directly executable — REQ-KB-005 takes the extraction branch of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018` into `internal/core/git`, and the extracted symbol returns a **path**. Read that package's `doc.go` before writing the extraction: the sibling's audit caught it naming a package whose declared purpose did not fit.
8. **`.moai/state/kanban/` is the sibling's, not this SPEC's — and it now has two occupants, neither of them the board.** Re-measured at v0.5.0: `SPEC-KANBAN-RENAME-001` `REQ-KR-009` puts the session record there, resolved per-tree through `internal/kanban/record.go`'s `stateDirSegments = []string{".moai", "state", "kanban"}` — the post-rename file and value; v0.3.0 and v0.4.0 cited `internal/factory/record.go` and `{".moai", "state", "factory"}`, both of which are now stale. The directory **exists** and holds the session records plus `backlog.json`, the queue `/moai todo` writes; `.moai/state/kanban-board/` is still absent. Board state is `.moai/state/kanban-board/`, resolved at the primary checkout, through **its own** constant. Reusing the record's constant lands the board in each worktree, which is AP-1 reached by following an instruction — and there are now two per-tree stores under that name for an implementer to land beside.

9. **Two landed surfaces state part of this SPEC's model, and each disagrees with it in one place.** `.claude/rules/moai/workflow/kanban-dispatch.md` is always-loaded and reaches every session regardless of what has been read. It agrees on the six ordered columns — **including `backlog` as one of them** — the ownerless `backlog` and `done`, the column→role map, the operator-act entry, and the evidence-read completion rule. It **disagrees** on exactly one point: it re-derives column position from SPEC `status` after a `/clear`, which REQ-KB-006 forbids. The **second** disagreement is with `/moai todo`, not with the rule — `todo.md` § Boundaries states "**Not a board.** Column position lives with the lead and the SPEC status, not in this file", where `backlog` is a column here. (Through v0.5.0 this item filed both against the rule and additionally claimed the rule corroborates "only the lead records a transition"; measured, the rule states neither that phrase nor the "Not a board" text, and its nearest support for the former is scoped to the `sync → done` step. Both corrected at v0.6.0, `spec.md` §A.9.) Neither disagreement is resolved by this plan: editing a shipped always-loaded rule is not in this SPEC's scope, and the rule's own boundaries mark the derivation as interim pending the store this SPEC defines. The implementer inherits the consequence rather than the fix — on the day REQ-KB-006 lands, a shipped always-loaded rule contradicts it, and that is a reportable finding, not a licence to implement the derivation.

10. **The operator-admission mechanism now exists and is not this SPEC's.** `/moai todo` (`.claude/skills/moai/workflows/todo.md`, `.claude/commands/moai/todo.md`), state at `.moai/state/kanban/backlog.json`. It closes the hole `SPEC-KANBAN-BOOTSTRAP-001` `plan.md` §B.4 recorded. Cited, not owned, not verified here (`spec.md` §A.10).

11. **`SPEC-KANBAN-BOOTSTRAP-001` is outside the current scope, so `REQ-KS-006` has no landing date.** That is what makes REQ-KB-025 necessary rather than optional: REQ-KB-017's runtime refusal reads a role declaration, and a refusal with nothing to read admits every write. The contract stays the sibling's, and it is consumed **whole and by reference** — REQ-KB-025 deliberately carries no list of its properties, because the v0.5.0 form did carry one, enumerated three of the contract's four clause groups, and thereby licensed a lead-only carrier that satisfied every observation and could not serve the routing `REQ-KS-019` performs. Read the contract from `REQ-KS-006`, never from a summary (`spec.md` §A.8, §D.14).

  Two consequences the implementer inherits. **The failure mode is directional**: a session-private carrier is the cheapest thing satisfying REQ-KB-017, and every board-side test passes on it; the observations that fail are the two cross-session directions in `AC-KB-017`'s carrier half, which must both be run. **The cross-SPEC uniqueness cannot be checked here**: that sibling declares this SPEC in its own `dependencies:` and so has not landed at M1, which is why the adoption obligation is landed on `REQ-KS-006` and decided by `AC-KS-030` rather than asserted in this SPEC's prose — and why a green uniqueness scan at M1 clears this SPEC's implementation only.

---

## §C. Pre-flight (M0 — run these before any edit)

Each is a command with a recorded expectation, so a drift shows up as a mismatch rather than as a surprise mid-milestone.

```bash
# 1. The rename prerequisite, read from the BASE BRANCH and not from the
#    working tree. REQ-KB-002 requires the rename to have LANDED; `test -d`
#    reads the local tree and is satisfied by an uncommitted rename — which
#    is precisely the state this worktree is in, so the working-tree form
#    passes on the one tree it was written to fail on (acceptance.md §A.1
#    rule 6, AC-KB-001).
#    Expect as re-measured at v0.5.0: the first command prints
#    internal/kanban, the second prints NOTHING — i.e. the gate PASSES.
#    (Through v0.4.0 the expectation was the reverse and the gate halted.)
#    The prerequisite SPEC's own frontmatter still reads status:
#    in-progress; that is bookkeeping lag, not a landing signal, and this
#    query is the observation that decides (§B.1).
git ls-tree -d --name-only origin/main internal/kanban
git ls-tree -d --name-only origin/main internal/factory

# 2. The single-origin discriminant. Run this from BOTH checkouts — the
#    primary case is the one that breaks, and it is the case the board
#    actually resolves toward.
#    Expect from the worktree: git-dir under .git/worktrees/, git-common-dir
#    the primary .git (already absolute).
#    Expect from the primary:  the BARE form returns the relative ".git",
#    while both absolute forms return the full path. This is why the bare
#    form is never used alone (spec.md §A.3, REQ-KB-005).
git rev-parse --git-dir
git rev-parse --git-common-dir                          # relative in primary
git rev-parse --path-format=absolute --git-common-dir   # absolute everywhere
git rev-parse --absolute-git-dir                        # fallback probe
git --version                                           # 2.31+ gates the flag

# 2b. The resolution M1 must reproduce, in the branch_guard.go form: one
#     probe, both paths absolute, in argument order. The board root is the
#     parent of the second line.
git rev-parse --path-format=absolute --git-dir --git-common-dir

# 3. The existing consumer of that same discriminant, so M1 reuses it
#    rather than writing a second probe. Expect the primary probe near
#    line 178 and the older-git fallback near line 190.
grep -n 'path-format=absolute\|absolute-git-dir\|git-common-dir' internal/hook/branch_guard.go

# 4. The status enum, read from the schema rather than from memory.
grep -n 'Valid values:' .claude/rules/moai/development/spec-frontmatter-schema.md

# 4b. The status census the compatibility table must admit. Expect 17 files
#     carrying `planned`, and 42 carrying one of the three out-of-lifecycle
#     terminals that are not board cards at all (spec.md §A.2).
grep -rlE '^status: planned\s*$' .moai/specs/ | wc -l
grep -rlE '^status: (archived|superseded|rejected)\s*$' .moai/specs/ | wc -l

# 5. The gitignore lines that make a worktree's state private. Expect both.
grep -n '\.moai/state' .gitignore

# 6. The two lock substrates, so M1 selects the cross-process one on
#    measured grounds rather than on convenience (§B.4).
ls internal/spec/lock*.go internal/lockfile/
sed -n '/in-process-mutex limitation/,/Windows users run solo mode/p' internal/lockfile/lockfile_windows.go

# 7. The atomic-write primitive and a working same-directory caller (§B.5).
sed -n '1,22p' internal/atomicfile/replace.go
grep -n 'CreateTemp\|os.Rename' internal/verify/store.go

# 8. The sole-writer absence this revision repairs, re-measured. Expect 0
#    before M1 lands the requirement, and a non-zero count after.
grep -rc 'sole writer\|single writer' .moai/specs/SPEC-KANBAN-BOARD-001/spec.md

# 9. The shape of the symbol REQ-KB-005 reuses. Expect an UNEXPORTED name
#    returning (bool, error) — which is why the requirement extracts rather
#    than calls (§B.7).
grep -n 'func isPrimaryCheckout' internal/hook/branch_guard.go

# 9b. The extraction target's declared purpose and its import direction.
#     Expect: "Git repository operations for MoAI-ADK", and NO match for
#     internal/hook or internal/cli. Read this before writing the extraction;
#     do not select a target by name resemblance.
head -3 internal/core/git/doc.go
grep -rn 'internal/hook\|internal/cli' internal/core/git/ | wc -l

# 10. The branch-side status read REQ-KB-020 depends on. Expect a status
#     line, from the primary checkout, with NO checkout and NO fetch — refs
#     are shared across every checkout of one repository. Substitute any
#     live branch that carries a SPEC.
git show "$(git -C . rev-parse --abbrev-ref HEAD):.moai/specs/SPEC-KANBAN-BOARD-001/spec.md" | grep -m1 '^status:'

# 11. The path collision REQ-KB-005 resolves, re-measured at v0.5.0 after
#     the rename landed. Expect the sibling's per-tree record constant at
#     its POST-rename home and value —
#       internal/kanban/record.go:
#         var stateDirSegments = []string{".moai", "state", "kanban"}
#     — the directory PRESENT with two per-tree occupants (session records
#     plus backlog.json, which /moai todo writes), and kanban-board still
#     ABSENT. Earlier revisions cited internal/factory/record.go and the
#     "factory" segment; both are stale (§B.8).
grep -n 'stateDirSegments' internal/kanban/record.go
ls -1 .moai/state/ | grep -E '^kanban(-board)?$'
ls -1 .moai/state/kanban/ 2>/dev/null

# 11b. The surfaces that landed after v0.4.0 and that this SPEC cites
#      without owning (§B.9, §B.10). Expect all three present, and expect
#      the dispatch rule to contain BOTH the agreements spec.md §A.9 cites
#      and the "re-derived from SPEC status" clause it reports as a
#      disagreement — read it before implementing REQ-KB-006, not after.
ls .claude/rules/moai/workflow/kanban-dispatch.md \
   .claude/skills/moai/workflows/todo.md \
   .claude/commands/moai/todo.md
grep -n 're-derived from SPEC status' .claude/rules/moai/workflow/kanban-dispatch.md

# 11c. The role-declaration carrier REQ-KB-025 conditions on. Expect NO
#      landed carrier — SPEC-KANBAN-BOOTSTRAP-001 is out of scope and
#      REQ-KS-006 fixes none by its own terms — which is the branch of
#      REQ-KB-025 that applies. Re-run this before M1 rather than trusting
#      it: if a carrier HAS appeared, M1 adopts it and defines none.
grep -rn 'REQ-KS-006' .moai/specs/SPEC-KANBAN-BOOTSTRAP-001/spec.md | head -3
grep -rln 'declared role\|role declaration' internal/ 2>/dev/null | head

# 12. The stale-lock gap REQ-KB-023 repairs, in the substrate's own words,
#     plus the Unix counter-case. Expect the "post-MVP enhancement / manual
#     del" header, a flock release on Close in the Unix impl, and a non-zero
#     count of orphaned artifacts that are inert on this platform (§B.6).
sed -n '1,8p' internal/spec/lock_windows.go
grep -n 'Close releases the flock' internal/spec/lock_unix.go
ls .moai/state/spec-close-*.lock 2>/dev/null | wc -l

# 13. The fallback-forcing pattern AC-KB-002's second half reuses. Expect the
#     package-level execCommand indirection and its comment recording that
#     direct invocation of the fallback is a vacuous pass.
grep -n 'var execCommand\|direct invocation' internal/hook/branch_guard.go internal/hook/branch_guard_test.go
```

### C.1 Two shell conventions this plan inherits

Both were paid for once already in the predecessor and are non-negotiable here.

**(a) Never read `$?` after a pipe.** `cmd | tail` makes `$?` the exit status of `tail`. Redirect to a log, read the exit code from the command itself, then count failures over the whole file:

```bash
go test ./... > /tmp/kanban-board-test.log 2>&1
rc=$?
fails=$(grep -c '^FAIL' /tmp/kanban-board-test.log)
echo "rc=$rc fails=$fails"
```

Counting `^FAIL` over the whole log — not over a tail — is what makes the count trustworthy: a tail truncates exactly the region a long suite puts its failures in.

**(b) Never iterate an undefined shell array, and lint per file.** Write the literal list:

```bash
for f in spec.md plan.md acceptance.md design.md research.md progress.md; do
  moai spec lint ".moai/specs/SPEC-KANBAN-BOARD-001/$f"
done
```

A `*.md` glob is unsatisfiable for any multi-artifact SPEC in this repository: `DuplicateSPECIDRule.CheckAll` treats each path as a separate SPEC, so siblings fail with either `ParseFailure` or `DuplicateSPECID`.

---

## §D. Constraints

| Constraint | Source | Consequence here |
|---|---|---|
| Template-First | `CLAUDE.local.md` §2 | template source edited first, then `make build`, then commit the regenerated `catalog.yaml` |
| Mirror delta preservation | measured, not assumed | a sanitized pair becoming byte-identical is a **failure** |
| Template content neutrality | `CLAUDE.local.md` §25 | no SPEC ID, REQ/AC token, internal date, or SHA under `internal/template/templates/` |
| Full test suite | prior run-phase miss | `go test ./...`, never an affected-packages subset |
| Env constants | `CLAUDE.local.md` §14 | names in `internal/config/envkeys.go`; no inline literals |
| Status enum immutability | frontmatter schema SSOT | no new status value; no matrix row reassigned; the board writes no frontmatter at all |
| Post-rename identifiers | `SPEC-KANBAN-RENAME-001` | no occurrence of `factory` in anything authored |
| No third lock mechanism | measured, §B.4 | REQ-KB-019 reuses `internal/spec/lock.go`; `internal/lockfile`'s in-process Windows fallback is neither used nor "upgraded" |
| No second atomic-write primitive | measured, §B.5 | REQ-KB-018 reuses `internal/atomicfile`; the temp file is created in the **target's own directory** |
| No line-number citation in normative text | v0.3.0 repair | requirements cite symbols (`isPrimaryCheckout`); line anchors live in `research.md` under its re-measurement discipline |
| Board path is `.moai/state/kanban-board/` | measured, §B.8 | its **own** path-segment constant; `.moai/state/kanban/` is `REQ-KR-009`'s — now carrying two per-tree occupants — and is neither reused nor amended |
| One role declaration, never two | `spec.md` §A.8, REQ-KB-025 | the carrier is established here only where `REQ-KS-006` has not landed, adopted where it has, and must satisfy that contract **whole — read it there, not from this row** (AP-33); a second declaration is a failure even where it agrees |
| The dispatch rule is cited, not edited | `spec.md` §A.9, §C | `.claude/rules/moai/workflow/kanban-dispatch.md` is a shipped always-loaded rule; its two disagreements with this SPEC are reported, and neither this plan nor its run-phase edits that file |
| `status` is read from the card's branch | measured, §B / `spec.md` §A.4a, §A.4b | branch resolved via `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` by **observing the card's worktree**; no branch name is derived here and the branch set is not searched; a card with **no live worktree** reads the primary checkout |
| Contract consumptions are not declared edges | `spec.md` §A.4a | `REQ-KW-003` and `REQ-KS-006` are consumed in requirement text; both siblings stay in `related_specs:`. `dependencies:` declares what must **land** first, and neither of these must |
| Role is read, never derived | `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` | the runtime half of REQ-KB-017 consumes the role **declaration**; deriving a role from a session identifier or launch label is forbidden |

**On the env-constants row.** It binds, and this SPEC's surface is expected to have nothing to bind *to*: the board model introduces no environment variable — the topology, the thresholds, and the entry switch all belong to `SPEC-KANBAN-BOOTSTRAP-001`. It therefore carries no requirement of its own here. Recorded explicitly so the absence reads as a scope boundary rather than as a dropped constraint; **where** the run-phase does introduce an environment name, this row binds it unchanged.

---

## §E. Settled decisions and what each one costs

Six decisions are settled and are not re-opened at run-phase. A run-phase that finds cause to re-open one records the reason as a blocker rather than choosing differently in place.

- **D6 — The role declaration's carrier is borrowed, conditionally, and there is exactly one of it.** Settled at v0.5.0 and **repaired at v0.6.0**, forced by a scope change rather than by a defect in the design: `SPEC-KANBAN-BOOTSTRAP-001` left the current scope, so `REQ-KS-006` — the contract `REQ-KB-017`'s runtime refusal reads — has no landing date, and a refusal with nothing to read admits every write. The split is contract versus carrier, and `REQ-KS-006` itself draws it by fixing no carrier. **What is borrowed**: the carrier alone, and only where the sibling has not landed; where it has, this SPEC adopts and defines none. **What is not**: the contract's definition, the argument grounding it (one role ↔ two-or-more labels), the role set and its election, and the **mechanisms** of dispatch routing and quorum accounting — all still that sibling's, none restated. The **key** those mechanisms read is a different thing from the mechanisms, and it is part of what the carrier must carry.

  **What v0.6.0 repaired, because the first form of this decision quietly narrowed the thing it borrowed.** v0.5.0 stated the binding clause as *resolvable from a session that is not the `lead`*, singular, and REQ-KB-025 enumerated three properties. That is workers-reading-lead — the direction `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` need — and it silently dropped `REQ-KS-006`'s lead-reading-workers clause, the key on which `REQ-KS-019` selects a dispatch target and `REQ-KS-012` accounts quorum. A lead-only carrier satisfied all three enumerated properties and could not route. The decision now consumes the contract **whole and by reference**, carries no enumeration anywhere, and is checked in **both** cross-session directions.

  **What the implementer inherits**: read `REQ-KS-006` itself, never a summary (AP-33); run both directions of `AC-KB-017`'s carrier half (AP-30); and note that this SPEC's uniqueness scan clears its own implementation only — the cross-SPEC half is `AC-KS-030`'s, because the sibling declares this SPEC in its `dependencies:` and so cannot have landed when the scan runs. That is also why the adoption obligation was moved out of this SPEC's prose and onto `REQ-KS-006`, where the implementer who could break it is reading. Lands in **M1** with the sole-writer trio it serves.

- **D1 — The board state has one origin: the primary checkout.** The predecessor's `column`-in-frontmatter design is rejected on three measured grounds (`spec.md` §A.3): a worktree's `.moai/state/` is private, `.moai/specs/` is tracked so a write is a commit against a branch-protected `main` (2,528 tracked files measured), and a card's branch forked before the lead's later writes so a merge restores a stale column. The replacement resolves the primary root as the parent of the **absolute** common git directory, obtained by reusing `internal/hook/branch_guard.go`'s `isPrimaryCheckout` — **not** by taking the parent of the bare `--git-common-dir`, which returns a relative `.git` in the primary checkout and is the defect v0.2.0 repairs. **Consequences the implementer inherits**: no `column` frontmatter field, so no schema registration and no template-mirror edit to the schema; column history is not in git and that loss is accepted; the two stores are disjoint and neither writes into the other. Two further consequences arrive at v0.3.0: the reuse is an **extraction** into `internal/core/git` returning a path, because the existing symbol is unexported and returns a boolean (§B.7); and the board's directory is `.moai/state/kanban-board/`, because `.moai/state/kanban/` is `REQ-KR-009`'s per-tree session record (§B.8). Lands in **M1**.

- **D4 — The board reads a card's `status` from the card's branch.** Settled at v0.3.0, and it is the repair with the widest blast radius: v0.2.0 named no read location at all, and the only location an implementer would reach for — the tree they are standing in — reports `draft` for the entire interval a card spends in `run`, which `REQ-KB-008` then refuses to dispatch. The branch is resolved by `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` (observation plus the SPEC identifier the branch carries), never by a name derived here; a card with **no live worktree** reads the primary checkout; and the read observes **committed** state only. **Refined at v0.4.0**, because *the* card's branch was the wrong singular: cards keep every branch they ever carried, so the source is selected on worktree liveness rather than on branch existence and the branch set is never searched (`spec.md` §A.4b, AP-28 / AP-28a). **What the implementer inherits**: a blob read (`git show <ref>:<path>`) with no checkout and no fetch, and a **contract** dependency on the sibling that is deliberately *not* declared in `dependencies:` — the identification rule is readable from that document today, nothing of that SPEC's need land first, and the reverse edge would close a cycle against its landing dependency on this one (`spec.md` §A.4a). Lands in **M2**, with the compatibility table it feeds.

- **D5 — The lock gets a stale-artifact exit, and the recovery gets a bound.** Two repairs of one shape: something that refuses with no way out. `REQ-KB-023` records the creating process in the board-wide lock artifact and provides a clearing operation conditioned on that process being **positively observed absent** *and* on a re-read of the recorded identity immediately before the unlink — the second clause is not decoration, it is the check-then-unlink race that would otherwise let the repair admit two writers. It narrows that race rather than closing it, and v0.4.0 says so in the requirement: no portable unlink takes a handle, so the residual is inherited rather than solved (`spec.md` §A.7(3a)). `REQ-KB-021` splits **absent** (an empty board; the writer creates the directory) from **unreadable** (unknown; refuse). `REQ-KB-022` defines what "bounded" means and forbids an unrecorded replacement, which would be the empty board §A.6 rejects wearing the word *explicit*. **What the implementer inherits**: the Unix substrate makes none of this reproducible locally (§B.6). Lands in **M1**.

- **D2 — WIP 2 with one coder session: the second card enters `run` and waits, unheld.** Admission is never gated on a free session; gating it would make the effective WIP the session count, which is the conflation REQ-KB-010 forbids arriving by another route. The unheld state is one field serving two causes — a card waiting for a session, and a card whose holder was released by the worktree sibling. Lands in **M2**.

- **D3 — One writer, atomic writes, board-wide exclusion.** Restored at v0.2.0 after the SPEC split deleted the ownership half of the predecessor's `REQ-KM-044` along with its rejected storage half (`spec.md` §A.7). The three are separable and each fails on its own: the `lead` is the sole writer and the implementation enforces it (REQ-KB-017); every write is a same-directory temp plus atomic rename through `internal/atomicfile` (REQ-KB-018); every board mutation holds a **board-wide** lock across the whole read-modify-write, from `internal/spec/lock.go` (REQ-KB-019). **What the implementer inherits**: WIP-2 of REQ-KB-009 is enforceable *only* beneath the board-wide lock — a card-scoped lock lets two concurrent transitions of two different cards each observe the bound satisfied and each write, landing at WIP 3; the sibling's `REQ-KW-013` card lock is untouched and stays correct for holder assignment; and the lock's test must span **separate processes**, because the in-process substrate named in §B.4 would pass a goroutine test while holding nothing. Lands in **M1**, ahead of the WIP admission it makes sound.

### E.1 One correction carried from the predecessor

The predecessor recorded a measurement establishing that an unknown frontmatter key produces no `moai spec lint` finding, and concluded that its new `column` field needed a schema-document registration but no lint rule. **That whole line of reasoning is withdrawn**, because there is no new frontmatter field. It is noted here so a reviewer who remembers the probe does not go looking for the registration it justified.

---

## §F. Milestones

Four milestones, M0 through M3.

### M0 — Preflight and prerequisite gate

Run every command in §C and record its output. Halt on the unlanded rename prerequisite (REQ-KB-002), decided by §C command 1 **against `origin/main`** — no working-tree predicate stands in for a landed-on-base-branch claim anywhere in this plan, and if one is found, it is repaired before it is run (`acceptance.md` §A.1 rule 6). Record the measured primary-root path from §C command 2 — M1's resolution test compares against it. Record the extraction target's `doc.go` and import direction from §C command 9b **before** M1 writes the extraction.

Nothing is edited in M0. Its only output is measurement.

### M1 — The state store, its resolution, its writer, and what the board must not write

The least reversible decision in practice, so it goes first. D1 and D3 both land here — D3 ahead of M2's WIP admission, which is unsound without it.

- **Extract** `isPrimaryCheckout`'s resolution into `internal/core/git` as a symbol returning the **path**, and re-point the existing `internal/hook` caller at it without changing that caller's contract. The symbol is unexported and returns a boolean today, so it cannot simply be called; copying it is forbidden, which leaves extraction (REQ-KB-005, `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018`). Confirm the target's `doc.go` first (§C command 9b).
- Resolve the board root as the parent of the **absolute** common git directory through that extracted symbol, with its older-git fallback intact. The bare `--git-common-dir` form is never used alone: it returns a relative `.git` in the primary checkout (REQ-KB-005, positive half).
- Persist board state beneath that root's `.moai/state/kanban-board/`, through **this package's own** path-segment constant — not `internal/kanban/record.go`'s `stateDirSegments`, which resolves per-tree and belongs to the session record. Note that after the rename this constant lives in the **same package** the board is being written into and already names `{".moai", "state", "kanban"}`, so reusing it is one import away rather than one package away (REQ-KB-005, AP-24).
- Distinguish an **absent** state file (empty board, dispatch permitted, directory created by the sole writer) from an **unreadable** one (unknown, dispatch refused), and check the two against each other rather than separately (REQ-KB-021, `spec.md` §D.13).
- Record the creating process in the board-wide lock artifact and provide the bounded clearing operation, conditioned on that process being positively observed absent **and** on a re-read of the recorded identity immediately before the unlink, aborting on mismatch. Construct the release-and-recreate interleaving ahead of that re-read; a clear that completes through it is the defect. Do not code this as though it were atomic — `unlink(2)` takes a path and resolves it at call time, so a residual window survives the re-read and is documented rather than closed (REQ-KB-023, `spec.md` §A.7(3a), §D.12).
- Establish the absence half: no board-state path anywhere resolves relative to the current working tree, with the positive control of `spec.md` §D.1 run once and recorded (REQ-KB-005).
- Implement the column as a recorded value with no derivation path — no helper that "recovers" a missing column by inference (REQ-KB-006).
- Establish that the board writes no SPEC frontmatter: no `column` field, no `status` write, no enum extension, no matrix transition, with the scan and its positive control of `spec.md` §D.3 (REQ-KB-007).
- Establish the sole writer: exactly one role, the `lead`, writes the board state file, enforced rather than documented, with the scan and positive control of `spec.md` §D.5. Elect nothing — the role's assignment is `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004` (REQ-KB-017).
- Make every board write atomic through `internal/atomicfile`, with the temporary file created in the **target's own directory** — never the system temp directory, which can cross a device boundary and degrade the rename into a copy. Check the same-directory property statically and the atomicity property from a concurrent reader, per `spec.md` §D.7 (REQ-KB-018).
- Serialize every board mutation beneath a **board-wide** advisory lock held across the whole read-modify-write, reusing `internal/spec/lock.go` and adding no third mechanism. Exercise it with **separate processes** per `spec.md` §D.6 — a goroutine test passes against an in-process mutex and measures nothing (REQ-KB-019).
- Tolerate a partially-written or unreadable state file in the safe direction — unknown, dispatch refused, never an empty board — and provide the bounded recovery out of that state: a reconstruction or replacement by the sole writer under the board-wide lock, invoked as an explicit operator-visible act and never taken by the read path (REQ-KB-013, `spec.md` §D.8). Bound it as REQ-KB-022 defines — the state file alone, one invocation, one verdict, no retry — and where a card cannot be reconstructed, record that durably and surface it rather than presenting the replacement as the board that was lost.
- Enforce the sole-writer refusal against the caller's **declared role**, read through `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`. Derive no role from a session identifier or a launch label, and restate no part of that contract (REQ-KB-017).
- **Establish the carrier that refusal reads**, because the sibling defining the contract is out of scope and fixes no carrier by its own terms — re-checking §C command 11c first, and adopting rather than defining if one has appeared since. **Read `REQ-KS-006` and satisfy it whole; do not work from a summary, including the one this plan would give you.** The clause an implementation drops for free is the dispatch-selection and quorum-accounting key — the direction in which the **lead reads a worker's** declaration — because nothing the board itself does exercises it. The other direction, workers reading the lead, is exercised by two of the worktree sibling's gates. Both are checked in `AC-KB-017`'s carrier half, and a carrier passing only one is the fork REQ-KB-025 exists to prevent. Define exactly one declaration surface, with the absence scan and positive control of `spec.md` §D.14; note that scan clears **this SPEC's** implementation only, the cross-SPEC half being decided on the sibling by `AC-KS-030` (REQ-KB-025, `spec.md` §A.8).

### M2 — The card, the columns, the table, and admission

The board model proper, on top of a store whose location is now settled. D2 lands here.

- Define the six-column closed enumeration, ordered, with no constructor accepting a value outside the set (REQ-KB-003).
- Define the card record — SPEC identifier, column, holder, last-transition instant — with an unheld card carrying an empty holder rather than a synthesized one (REQ-KB-004).
- Read each card's `status` from the card's **branch** by blob read, without a checkout, resolving the branch by **observing the card's worktree** through `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` — deriving no branch name here and **searching** the branch set for none, since a card retains every branch it ever carried and 3 of the 29 SPEC identifiers on branches currently carry two or more, disagreeing (`research.md` §O.2). Fall back to the primary checkout where the card has **no live worktree**, which is the selector — not the absence of a branch, which never arrives. Observe committed state only (REQ-KB-020, `spec.md` §A.4b). This lands **before** the table below, which is unusable without it: with a primary-side read, every card in `run` pairs as `(run, draft)` and the table refuses it.
- Render a card whose source resolves to no single tree — a worktree reporting no branch, or a `REQ-KW-019` refusal — with `status` **unresolved**: in its recorded column, not dispatchable, candidates surfaced, nothing written, and no enum member substituted. The default an implementation acquires for free here is the zero value, which reports `draft` and dispatches (REQ-KB-024, `spec.md` §A.4b).
- Implement the `(column, status)` compatibility table of `spec.md` §A.4 **as revised at v0.2.0**: `planned` admitted in `backlog` and `plan` and nowhere else, `completed` admitted in `sync`, and `archived` / `superseded` / `rejected` producing no card at all. Legal pairings accepted; an illegal pairing marking the card inconsistent, reporting it not dispatchable, and repairing neither field. Exercise **every** pairing including the illegal rows (REQ-KB-008, `spec.md` §D.2).
- Implement WIP-2 admission beneath the board-wide lock of M1 — never beneath a card lock — with a named refusal and an unchanged board on refusal, plus the conditionality control of `spec.md` §D.4 (REQ-KB-009, REQ-KB-019).
- Keep the WIP limit and the session count independent, deriving neither from the other (REQ-KB-010).
- Admit into `run` ungated by session availability, with unheld-in-`run` as a legal steady state (REQ-KB-011).
- Report a card in `backlog` or `done` as not dispatchable (REQ-KB-012).

### M3 — Template mirror, neutrality, and verification

Mechanical, and therefore last.

- Template source first, then `make build`, then commit the regenerated `catalog.yaml`; re-measure each touched pair's classification and preserve its delta (REQ-KB-014).
- Neutrality on every authored template file, with the repository's guard as the verdict (REQ-KB-015).
- Full suite per §C.1(a), and per-file `spec lint` per §C.1(b) (REQ-KB-016).
- Every acceptance criterion in `acceptance.md` executed, with its command and verbatim output recorded in `progress.md` §E.2.

---

## §G. Anti-patterns

- **AP-1 — Resolving the board state relative to the current tree.** `os.Getwd`, the session's own repository root, or any tree-relative anchor produces one board per worktree. The whole of D1 is the fix; this is the failure it is a fix *for*, and it is silent — six sessions each see a coherent board, and none of them is the board.
- **AP-2 — Re-introducing a `column` frontmatter field.** Rejected with measured grounds (`spec.md` §A.3). A "small" frontmatter write is a commit against a branch-protected `main`, six times per card.
- **AP-3 — Deriving the column from `status`.** Decided against, twice now. A helper that recovers a missing column by inference re-introduces the rejected design under another name, and it is blind at `review` exactly as the original was.
- **AP-4 — Repairing a column/status disagreement.** Silently rewriting one to match the other means either the board writing a status transition it does not own, or the board overwriting what another actor observed. An illegal pairing blocks and surfaces (REQ-KB-008).
- **AP-5 — Conflating the WIP limit with the session count.** They are two knobs (REQ-KB-010). Clamping one to the other silently is the shape this most often takes.
- **AP-6 — Gating `run` admission on a free session.** It reads as prudence and is the WIP/session-count conflation arriving by the back door (D2, REQ-KB-011).
- **AP-7 — Inventing a held or blocked column.** The unheld state already exists and serves both of its causes (`spec.md` §A.5). A seventh column to express it is a new column for a state the board already has.
- **AP-8 — Treating an unreadable state file as an empty board.** The empty-board path already exists, which is exactly why this is tempting. It admits cards past a WIP limit whose contents are unknown (REQ-KB-013).
- **AP-16 — Taking the parent of the bare `--git-common-dir`.** It is correct from every worktree and wrong in the primary checkout, where the bare form returns the relative `.git`. The failure is asymmetric and therefore easy to ship: every worktree test passes, and the one checkout the single-origin design points at is the one that breaks. Reuse the `branch_guard.go` probe (§F M1).
- **AP-17 — Letting a non-`lead` session write the board.** The rule existing only in prose is exactly how it was lost once already (`spec.md` §A.7). "Enforced, not documented" is the requirement; a comment above a write path does not satisfy it (REQ-KB-017).
- **AP-18 — Guarding a board mutation with a card-scoped lock.** The card lock is already there, already correct for holder assignment, and reaching for it is the natural move — which is what makes this the most likely way WIP-2 silently fails. Two concurrent transitions of two *different* cards each take a different lock, each read one card in `run`, and both write (REQ-KB-019).
- **AP-19 — Testing the lock with goroutines.** An in-process mutex passes a same-process test perfectly and protects nothing across the OS processes the sessions actually are. `internal/lockfile/lockfile_windows.go` is the repository's own worked example of that gap. Separate processes, or the criterion is not met (`spec.md` §D.6).
- **AP-20 — Writing the temp file to the system temp directory.** The rename still happens, so the code looks atomic and the test asserting "a rename occurred" passes. Across a device boundary the rename degrades to a copy, and the torn write returns. Same directory as the target (REQ-KB-018).
- **AP-21 — Recovering from the unknown state on read.** A read path that quietly repairs an unreadable board destroys the evidence of whatever killed the writer, and does it on a file whose contents are by definition unknown. Recovery is an explicit act somebody performs; the read path keeps reporting unknown (REQ-KB-013, `spec.md` §D.8).
- **AP-22 — Refusing without an exit.** The mirror image of AP-8, and the defect v0.1.0 actually shipped: a board that refuses every dispatch forever after one killed write is failing permanently, not safely.
- **AP-23 — Reading a card's `status` out of the tree you are standing in.** The natural move, and wrong for the whole interval that matters: the transition is on the card's branch and the worktree survives until both pull requests merge, so the primary-side copy reads `draft` while the card sits in `run`, and `REQ-KB-008` refuses to dispatch — every card, on the normal path. Read the branch (REQ-KB-020, `spec.md` §A.4a).
- **AP-24 — Reusing `internal/kanban/record.go`'s path constant for the board.** `REQ-KR-009` instructs reuse of that constant for the **session record**, and it resolves per-tree. Following that instruction for board state lands one board per worktree — AP-1 arrived at by obeying a sibling requirement, which is what makes it hard to see. It is harder to see since the rename: the constant now lives in the same package the board is being written into and names the same segment (`{".moai", "state", "kanban"}`), so reaching for it is one import away rather than one package away, and the directory already holds two per-tree stores to land beside. The board has its own constant and its own directory (REQ-KB-005, §B.8).
- **AP-25 — Clearing a stale lock by reading the identity and then unlinking.** Correct in isolation, and it opens the window: the owner may release and a live process re-acquire between the two steps, after which the clear unlinks a valid lock and two writers enter the critical section. The clearing operation would then cause the concurrency the lock exists to prevent. Re-read the recorded identity immediately before the unlink and abort on mismatch — and do not then describe the result as atomic, which is AP-29 (REQ-KB-023).

- **AP-28 — Picking one of a card's branches.** A card keeps every branch it ever carried; measured, 3 of the 29 SPEC identifiers on branches carry two or more, and one carries `draft`, `in-progress` and `completed` at once (`research.md` §O.2). Every tiebreak an implementer reaches for is wrong: by stage, because the type prefix is not a stage ladder and two of the three cards carry no stage triple at all; by recency, because `REQ-KW-019` names recency among its explicit refusals; by any rule whatever, because that requirement refuses selection as such. Resolve by **observing the card's worktree**, fall back to the primary checkout where none is live, and search for nothing (REQ-KB-020, `spec.md` §A.4b).

- **AP-28a — Keying the fallback on whether a branch exists.** The near-miss of AP-28, and the v0.3.0 text's actual defect. Nothing deletes a card's branches, so "the card has no branch" is a condition that essentially never becomes true, and an implementation keyed on it searches the branch set for every disposed card — reading `draft` off a retained `plan/` branch for a card sitting in `done`. Key on **worktree liveness**, which is the thing that actually changes (REQ-KB-020).

- **AP-29 — Describing the narrowed race as closed.** Distinct from AP-25, and the more durable error: AP-25 omits the re-read, while this one performs it and then reasons as though the artifact could no longer change. It cannot be closed at this layer, and an implementer who believes it has been stops looking for the consequences. State the residual where the requirement states it (REQ-KB-023, `spec.md` §A.7(3a)).
- **AP-26 — Treating "replacement of the state file" as licence to write an empty board.** It satisfies "the board leaves the unknown state" perfectly, and it is AP-8 reached through the door marked *explicit*. What could not be recovered is recorded and surfaced (REQ-KB-022).
- **AP-27 — Promoting a sibling into `dependencies:` because a requirement cites it.** `REQ-KB-020` cites `REQ-KW-003` and `REQ-KB-017` cites `REQ-KS-006`; neither is a declared edge, and both absences are decisions. `dependencies:` declares what must **land** first, and a rule readable from a sibling document today imposes no such ordering. This exact promotion was made and reversed within v0.3.0 — it closed a cycle against `SPEC-KANBAN-WORKTREE-001`'s landing dependency on this SPEC (`spec.md` §A.4a). The mirror-image error is worse and is also forbidden: resolving such a cycle by deleting the **sibling's** edge would drop a real prerequisite and leave no record of why (`spec.md` §C).
- **AP-30 — Building a role carrier that serves one direction.** Two shapes, one failure. A **session-private** carrier is the cheapest thing satisfying REQ-KB-017's runtime refusal, since the board only ever asks what role *its own caller* occupies. A **lead-only** carrier is the next cheapest and is subtler: it satisfies the board, and it satisfies the workers-read-lead direction that `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` need, so it passes every observation those two SPECs contribute — and it cannot serve the direction `REQ-KS-019` routes on, where the **lead reads a worker's** declaration. Both break in *another SPEC's* criteria, the kind of failure least likely to be caught by the SPEC that caused it. Run both cross-session directions in the test, or the surviving one is never exercised (REQ-KB-025, `spec.md` §D.14).

- **AP-31 — Defining a second role declaration because one is "close enough".** The failure mode of a borrow, and it is invisible on the day it is made: two declarations that agree are indistinguishable from one, until one of them changes and nothing in either document says which is authoritative. Where `REQ-KS-006` has landed, adopt its carrier and define none; where it has not, define one and expect a later landing to adopt it — that expectation is not a hope, it is `REQ-KS-006`'s own widened obligation, decided by `AC-KS-030`. Re-run §C command 11c before writing rather than trusting this plan's measurement (REQ-KB-025).

- **AP-33 — Satisfying a summary of `REQ-KS-006` instead of `REQ-KS-006`.** The v0.5.0 form of REQ-KB-025 enumerated three of that contract's four clause groups and declared itself a restatement of none; the missing one was the dispatch-selection and quorum-accounting key, and the enumeration licensed a carrier that passed everything and served nothing. Any list of the contract's properties — in a requirement, in this plan, in a code comment — is a narrowing waiting to happen. Cite the contract; do not summarize it (REQ-KB-025, `spec.md` §A.8).

- **AP-32 — Implementing the dispatch rule's column derivation because it is always-loaded.** `.claude/rules/moai/workflow/kanban-dispatch.md` states that column position is "re-derived from SPEC status after a clear", and it reaches every session whether or not this SPEC has been read — so an implementer who trusts the rule over the SPEC writes exactly the design REQ-KB-006 forbids and §A.2 rejects, blind at the `run`/`review` collision the explicit column exists to resolve. The rule's own boundaries mark it as an interim stance pending the store this SPEC defines, and that framing is stated nowhere the rule's reader will see it. The contradiction is a reportable finding, not a licence (`spec.md` §A.9, REQ-KB-006).

- **AP-9 — Extending the status enum to fit the board.** The `review` column has no counterpart; that is a fact about the vocabulary, not a defect in it.
- **AP-10 — Asserting the compatibility table on its legal rows only.** An implementation that accepts every pairing passes such a test perfectly (`spec.md` §D.2).
- **AP-11 — An absence check with no positive control.** A scan that has never been shown to fire is indistinguishable from a broken command, and both report success.
- **AP-12 — Reimplementing an existing guard's regex.** A guard reimplemented without its exemption list is a false-failure machine. Run the guard; treat its exit code as the verdict.
- **AP-13 — `cmd | tail` then `$?`.** §C.1(a).
- **AP-14 — Testing the affected packages only.** §D, REQ-KB-016.
- **AP-15 — Absorbing a sibling's scope.** Stall detection, holder release, worktree lifecycle, dispatch, and topology each look like a two-line addition from inside the board model. Each belongs to a named sibling (`spec.md` §C); adding one here re-creates the predecessor.

---

## §H. Out of Scope

### Out of Scope — carried from spec.md §C

- The worktree lifecycle, stall detection, holder release, and assignment exclusion (`SPEC-KANBAN-WORKTREE-001`); preflight beyond the rename gate, topology, bootstrap, configuration, quorum, dispatch, backend selection, and the coder-session chain (`SPEC-KANBAN-BOOTSTRAP-001`); a `column` frontmatter field and its schema registration; status-enum extension; a seventh column; any board rendering surface. `spec.md` §C is the authority; this heading exists so the exclusion travels with the plan rather than being one file away.

### Out of Scope — plan-phase deliverables

- Any code, test, or template edit. This SPEC's plan phase produces artifacts only; M0's measurements are read-only.

---

## §I. Cross-references

- `spec.md` — scope, requirements, verification surfaces.
- `design.md` — the decisions this plan executes, each with its rejected alternatives and the measured reason for rejection (Tier L artifact, added at v0.2.0).
- `research.md` — the measurements underlying both, each with its command and observed output so a later reader can re-run it (Tier L artifact, added at v0.2.0).
- `acceptance.md` — the criteria and the command that decides each.
- `progress.md` — the phase evidence record.
- `SPEC-KANBAN-WORKTREE-001` — a `related_specs:` entry, deliberately **not** a `dependencies:` one (AP-27). `REQ-KW-003` supplies the branch D4 reads, as a contract consumption discharged by citation; `REQ-KW-005` and `REQ-KW-007` are why the primary-side `status` is stale; `REQ-KW-014` is the shape D5's lock repair ports; `REQ-KW-018` is the extraction disposition D1 takes. Its own `dependencies:` names this SPEC on a landing need, and that edge is the one that stays declared.
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the role-declaration **contract** M1's sole-writer refusal reads. Consumed by name; not a `dependencies:` entry, because that sibling already depends on this one and, since v0.5.0, because it is out of scope and REQ-KB-025 exists to remove the landing prerequisite a declared edge would assert. Its **carrier**, which that requirement deliberately fixes not at all, is D6's borrow.
- `SPEC-KANBAN-RENAME-001` `REQ-KR-009` — the session-record path §B.8 separates from and does not amend. Landed; its constant is now `internal/kanban/record.go`'s.
- `.claude/rules/moai/workflow/kanban-dispatch.md` — the always-loaded dispatch rule §B.9 cites on six points and reports two disagreements against. Not edited by this plan.
- `.claude/skills/moai/workflows/todo.md` — `/moai todo`, the operator-admission mechanism §B.10 cites and does not own.
