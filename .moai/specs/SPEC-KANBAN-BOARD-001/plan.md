---
id: SPEC-KANBAN-BOARD-001
title: "Implementation plan — six-column kanban board model with a single-origin board state store"
version: "0.4.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, plan, milestones, state-store, single-origin, sole-writer, atomicity"
tier: L
---

## §A. Context

The scope, the decided design, and the requirement set are `spec.md`. This file is the ordered path through them, plus the pre-flight measurements and the anti-patterns the run-phase implementer needs in hand before touching a file.

Milestones are ordered by **decision-reversibility**: the choices most likely to change on contact — the state store's location and resolution, then the card record and the consistency table — come first, so a wrong turn is discovered while it is still cheap. The mechanical work (template mirror, catalog regeneration, full verification) is deliberately last.

---

## §B. Known issues and unlanded ground

Eight facts make this plan conditional. All eight were verified, not assumed.

1. **`SPEC-KANBAN-RENAME-001` has not landed.** It exists in this worktree as an untracked directory and is not on `origin/main`. Every identifier this plan names is post-rename. M0 gates on it (REQ-KB-002).
2. **The pre-commit gate blocks commits.** The repository's `.git/hooks/pre-commit` invokes `moai gate`, which exits non-zero on pre-existing ast-grep findings. The documented bypass is `SKIP_MOAI_PRECOMMIT=1`. That defect is tracked separately and is not fixed here — it is recorded so an implementer does not diagnose it a second time.
3. **The board state store is gitignored, which means it is invisible to every git-based check.** `.gitignore` ignores `.moai/state/` (line 275) and `**/.moai/state/` (line 207). Nothing about the board's runtime state will appear in a diff, a `git status`, or CI. Every check on it must read the file, not the repository.
4. **Both lock substrates exist, and only one of them is cross-process.** `internal/spec/lock.go` (with `lock_unix.go` / `lock_windows.go`) is the flock-plus-atomic-create family REQ-KB-019 reuses. `internal/lockfile/lockfile_windows.go` is a `map[string]*sync.Mutex` whose own package comment records that writes across **different** OS processes are not protected and forbids upgrading it. Reaching for the second because it is simpler produces a lock that passes every same-process test and holds nothing in production.
5. **The atomic-write primitive already exists.** `internal/atomicfile` provides `Replace` (write-temp-then-rename) with the Windows handling already written, and `internal/verify/store.go` `Save` is a working same-directory example (`os.CreateTemp(dir, ...)` into the target's own directory, then rename). REQ-KB-018 reuses both shapes rather than authoring a third.
6. **The lock family REQ-KB-019 reuses does not clean up after a dead holder, and the gap is platform-asymmetric.** `internal/spec/lock_windows.go`'s own header states that stale-lock detection is a post-MVP enhancement and that cleanup is manual. On Unix the substrate holds `flock(2)` on an open descriptor, which the kernel releases on process exit — measured, `.moai/state/` carries fourteen orphaned zero-length `spec-close-*.lock` files that block nothing. On Windows the artifact **is** the lock and the block is permanent. An implementer developing on macOS will not reproduce this; REQ-KB-023 exists because of it.
7. **The resolution REQ-KB-005 reuses is unexported and returns a boolean.** Measured: `isPrimaryCheckout(projectDir string) (bool, error)` in package `hook`. There is no exported path-resolving helper, so "reuse it" is not directly executable — REQ-KB-005 takes the extraction branch of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-018` into `internal/core/git`, and the extracted symbol returns a **path**. Read that package's `doc.go` before writing the extraction: the sibling's audit caught it naming a package whose declared purpose did not fit.
8. **`.moai/state/kanban/` is the sibling's, not this SPEC's.** `SPEC-KANBAN-RENAME-001` `REQ-KR-009` puts the session record there, resolved per-tree through `internal/factory/record.go`'s `stateDirSegments`. Board state is `.moai/state/kanban-board/`, resolved at the primary checkout, through **its own** constant. Reusing the record's constant lands the board in each worktree, which is AP-1 reached by following an instruction.

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
#    Expect at plan time: the first command prints nothing, the second
#    prints internal/factory — i.e. the gate HALTS.
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

# 11. The path collision REQ-KB-005 resolves. Expect the sibling's per-tree
#     record constant, and NO existing entry named kanban or kanban-board.
grep -n 'stateDirSegments' internal/factory/record.go
ls .moai/state/ | grep -c '^kanban'

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
| Board path is `.moai/state/kanban-board/` | measured, §B.8 | its **own** path-segment constant; `.moai/state/kanban/` is `REQ-KR-009`'s and is neither reused nor amended |
| `status` is read from the card's branch | measured, §B / `spec.md` §A.4a, §A.4b | branch resolved via `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003` by **observing the card's worktree**; no branch name is derived here and the branch set is not searched; a card with **no live worktree** reads the primary checkout |
| Contract consumptions are not declared edges | `spec.md` §A.4a | `REQ-KW-003` and `REQ-KS-006` are consumed in requirement text; both siblings stay in `related_specs:`. `dependencies:` declares what must **land** first, and neither of these must |
| Role is read, never derived | `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` | the runtime half of REQ-KB-017 consumes the role **declaration**; deriving a role from a session identifier or launch label is forbidden |

**On the env-constants row.** It binds, and this SPEC's surface is expected to have nothing to bind *to*: the board model introduces no environment variable — the topology, the thresholds, and the entry switch all belong to `SPEC-KANBAN-BOOTSTRAP-001`. It therefore carries no requirement of its own here. Recorded explicitly so the absence reads as a scope boundary rather than as a dropped constraint; **where** the run-phase does introduce an environment name, this row binds it unchanged.

---

## §E. Settled decisions and what each one costs

Five decisions are settled and are not re-opened at run-phase. A run-phase that finds cause to re-open one records the reason as a blocker rather than choosing differently in place.

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
- Persist board state beneath that root's `.moai/state/kanban-board/`, through **this package's own** path-segment constant — not `internal/factory/record.go`'s, which resolves per-tree and belongs to the session record (REQ-KB-005).
- Distinguish an **absent** state file (empty board, dispatch permitted, directory created by the sole writer) from an **unreadable** one (unknown, dispatch refused), and check the two against each other rather than separately (REQ-KB-021, `spec.md` §D.13).
- Record the creating process in the board-wide lock artifact and provide the bounded clearing operation, conditioned on that process being positively observed absent **and** on a re-read of the recorded identity immediately before the unlink, aborting on mismatch. Construct the release-and-recreate interleaving ahead of that re-read; a clear that completes through it is the defect. Do not code this as though it were atomic — `unlink(2)` takes a path and resolves it at call time, so a residual window survives the re-read and is documented rather than closed (REQ-KB-023, `spec.md` §A.7(3a), §D.12).
- Establish the absence half: no board-state path anywhere resolves relative to the current working tree, with the positive control of `spec.md` §D.1 run once and recorded (REQ-KB-005).
- Implement the column as a recorded value with no derivation path — no helper that "recovers" a missing column by inference (REQ-KB-006).
- Establish that the board writes no SPEC frontmatter: no `column` field, no `status` write, no enum extension, no matrix transition, with the scan and its positive control of `spec.md` §D.3 (REQ-KB-007).
- Establish the sole writer: exactly one role, the `lead`, writes the board state file, enforced rather than documented, with the scan and positive control of `spec.md` §D.5. Elect nothing — the role's assignment is `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004` (REQ-KB-017).
- Make every board write atomic through `internal/atomicfile`, with the temporary file created in the **target's own directory** — never the system temp directory, which can cross a device boundary and degrade the rename into a copy. Check the same-directory property statically and the atomicity property from a concurrent reader, per `spec.md` §D.7 (REQ-KB-018).
- Serialize every board mutation beneath a **board-wide** advisory lock held across the whole read-modify-write, reusing `internal/spec/lock.go` and adding no third mechanism. Exercise it with **separate processes** per `spec.md` §D.6 — a goroutine test passes against an in-process mutex and measures nothing (REQ-KB-019).
- Tolerate a partially-written or unreadable state file in the safe direction — unknown, dispatch refused, never an empty board — and provide the bounded recovery out of that state: a reconstruction or replacement by the sole writer under the board-wide lock, invoked as an explicit operator-visible act and never taken by the read path (REQ-KB-013, `spec.md` §D.8). Bound it as REQ-KB-022 defines — the state file alone, one invocation, one verdict, no retry — and where a card cannot be reconstructed, record that durably and surface it rather than presenting the replacement as the board that was lost.
- Enforce the sole-writer refusal against the caller's **declared role**, read through `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`. Derive no role from a session identifier or a launch label, and define no declaration mechanism here (REQ-KB-017).

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
- **AP-24 — Reusing `internal/factory/record.go`'s path constant for the board.** `REQ-KR-009` instructs reuse of that constant for the **session record**, and it resolves per-tree. Following that instruction for board state lands one board per worktree — AP-1 arrived at by obeying a sibling requirement, which is what makes it hard to see. The board has its own constant and its own directory (REQ-KB-005, §B.8).
- **AP-25 — Clearing a stale lock by reading the identity and then unlinking.** Correct in isolation, and it opens the window: the owner may release and a live process re-acquire between the two steps, after which the clear unlinks a valid lock and two writers enter the critical section. The clearing operation would then cause the concurrency the lock exists to prevent. Re-read the recorded identity immediately before the unlink and abort on mismatch — and do not then describe the result as atomic, which is AP-29 (REQ-KB-023).

- **AP-28 — Picking one of a card's branches.** A card keeps every branch it ever carried; measured, 3 of the 29 SPEC identifiers on branches carry two or more, and one carries `draft`, `in-progress` and `completed` at once (`research.md` §O.2). Every tiebreak an implementer reaches for is wrong: by stage, because the type prefix is not a stage ladder and two of the three cards carry no stage triple at all; by recency, because `REQ-KW-019` names recency among its explicit refusals; by any rule whatever, because that requirement refuses selection as such. Resolve by **observing the card's worktree**, fall back to the primary checkout where none is live, and search for nothing (REQ-KB-020, `spec.md` §A.4b).

- **AP-28a — Keying the fallback on whether a branch exists.** The near-miss of AP-28, and the v0.3.0 text's actual defect. Nothing deletes a card's branches, so "the card has no branch" is a condition that essentially never becomes true, and an implementation keyed on it searches the branch set for every disposed card — reading `draft` off a retained `plan/` branch for a card sitting in `done`. Key on **worktree liveness**, which is the thing that actually changes (REQ-KB-020).

- **AP-29 — Describing the narrowed race as closed.** Distinct from AP-25, and the more durable error: AP-25 omits the re-read, while this one performs it and then reasons as though the artifact could no longer change. It cannot be closed at this layer, and an implementer who believes it has been stops looking for the consequences. State the residual where the requirement states it (REQ-KB-023, `spec.md` §A.7(3a)).
- **AP-26 — Treating "replacement of the state file" as licence to write an empty board.** It satisfies "the board leaves the unknown state" perfectly, and it is AP-8 reached through the door marked *explicit*. What could not be recovered is recorded and surfaced (REQ-KB-022).
- **AP-27 — Promoting a sibling into `dependencies:` because a requirement cites it.** `REQ-KB-020` cites `REQ-KW-003` and `REQ-KB-017` cites `REQ-KS-006`; neither is a declared edge, and both absences are decisions. `dependencies:` declares what must **land** first, and a rule readable from a sibling document today imposes no such ordering. This exact promotion was made and reversed within v0.3.0 — it closed a cycle against `SPEC-KANBAN-WORKTREE-001`'s landing dependency on this SPEC (`spec.md` §A.4a). The mirror-image error is worse and is also forbidden: resolving such a cycle by deleting the **sibling's** edge would drop a real prerequisite and leave no record of why (`spec.md` §C).
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
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the role declaration M1's sole-writer refusal reads. Consumed by name; not a `dependencies:` entry, because that sibling already depends on this one.
- `SPEC-KANBAN-RENAME-001` `REQ-KR-009` — the session-record path §B.8 separates from and does not amend.
