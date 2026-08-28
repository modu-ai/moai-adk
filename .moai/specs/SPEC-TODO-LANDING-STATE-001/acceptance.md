# SPEC-TODO-LANDING-STATE-001 — Acceptance Criteria

Scope: **half A only** — the discriminator (`spec.md` §A.5). The criteria that verified landing
storage, the recording verb, an observed commit SHA, and the live SPEC-status read moved to card
t359 with the design they tested.

Every AC below names the **observed RED** — the command and the wrong output it produced **before**
the change — and the **expected GREEN**, so a later reader can re-run both. An AC whose RED was not
actually observed says so in its RED cell rather than asserting one, and is classified as a
regression-guard rather than as release-blocking (§C).

Commands run from the card worktree root (`.claude/worktrees/t331`) unless stated.

---

## §A Standing preconditions

### A.1 Baseline tree, and what a tree SHA does not pin [HARD]

Every RED cell whose subject is a **file** was measured at `3de2f85a2` and re-opened at its address
at HEAD `11426a128`; the pins hold across both because no cited source file changed between them
(`git diff --name-only 3de2f85a2 HEAD` returns only this SPEC's own artifacts and the audit report).
A RED carried forward without re-measurement is a carry-over, not a baseline
(`verification-claim-integrity.md` §2).

A RED cell whose subject is **`origin/main` or `origin/develop`** is not pinned by a tree SHA at all:
those refs advance independently of any tree, and two of this SPEC's figures already moved
(`0 329 → 0 349`; t322's develop count `5 → 24`). Such a cell states the command, the two refs' own
SHAs, and the instant — `origin/main` = `48239c7dc`, `origin/develop` = `c6aa61346`, observed
**2026-08-28T13:15Z** — and is **re-measured** by a later reader, never re-cited
(`verification-completeness.md` §4).

### A.2 Exit-code reading [HARD]

`| wc -l` reports `0` for a command that failed. Every count-shaped assertion reads the exit code
separately or runs under `set -o pipefail`.

### A.3 Verification isolation [HARD]

Suites run scoped to the affected packages. An environment-scrubbed run inside this worktree is one
compound invocation (`unset … && go test …`), never a separate `unset`.

### A.4 The three misjudgement inputs

These are test inputs, not illustrations. Each AC that repairs a discriminator names which one it is
established against. All three bear on the **detection** surfaces this SPEC changes.

| Input | Shape | Measured fact (ref probes: `origin/develop` = `c6aa61346`, 2026-08-28T13:15Z) |
|---|---|---|
| **squash-blind** | t200 | `git merge-base --is-ancestor 7fc161b36 origin/develop ; echo $?` → `1`; content landed as `294b4b6ab` (PR #1612), same probe → `0` |
| **false not-started** | t338 | `git log origin/develop --perl-regexp --grep='\bt338\b' --oneline \| wc -l` → `0`. Only the zero-commit half is measured here; the "work existed, blocked by a dead session's stale `index.lock`" account is operator testimony this SPEC did **not** verify and does not rest on |
| **silent pass** | F6 | `todo done --require-landed` on an unanswerable query exits 0 and prints `done <id>`, byte-identical to a satisfied guard — established by `internal/cli/todo.go:417-431` @ `3de2f85a2` (two `return nil` paths, neither writing stdout), not by the tests, which establish the exit code |

The squash-blind input is the reason a landed answer must rest on the ref's own history rather than
on ancestry of a card-branch SHA. It bears on the **discriminator** only; recording a SHA anywhere is
card t359's scope and no criterion below asserts one.

---

## §B AC Matrix

Ten criteria against a Tier M ceiling of 16. The six that verified landing storage, the recording
verb, an observed commit SHA, and the live SPEC-status read left with card t359 (`spec.md` §A.5,
§D); nothing below asserts a persisted fact.

| AC | Verifies | Scenario (Given / When / Then) | Observed RED | Expected GREEN |
|---|---|---|---|---|
| **AC-TLS-001** | REQ-TLS-001 | **Given** a project whose `git_strategy.worktree_base_branch` is `develop` **when** the landed check runs for a card named only by `origin/develop` **then** it answers `landed`. | Ref probe @ `origin/main` `48239c7dc` / `origin/develop` `c6aa61346`, 2026-08-28T13:15Z: `git log origin/main --perl-regexp --grep='\bt293\b' --oneline \| wc -l` → `0`, same query on `origin/develop` → `9`. So the check answers `not landed` and `todo pr` renders `no-link` for a card that shipped. Same shape for t310 (`0` / `6`) and t322 (`0` / `24`). | Unit test over the resolver: configured `develop` ⇒ ref `origin/develop`; a t293-shaped fixture answers `landed`. |
| **AC-TLS-002** | REQ-TLS-001 | **Given** a project whose `worktree_base_branch` is empty **when** the landed check runs **then** it uses `origin/main`, byte-identically to `3de2f85a2`. | **RED cell: none, and deliberately so** — this is the behaviour being preserved, so it is a **regression-guard**, not a release-blocking criterion (§C). A criterion with no RED is unadopted as a release gate; classified honestly rather than dressed with an invented failure. | Table test: empty ⇒ `origin/main`; the existing `todo pr` golden output over an unconfigured fixture is unchanged. |
| **AC-TLS-003** | REQ-TLS-002 | **Given** a project configured on `develop` **when** `moai todo done tX --require-landed` refuses **then** the refusal names `origin/develop`. | The refusal names `origin/main` unconditionally — the string is built from the constant (`internal/cli/todo.go:426-428` @ `3de2f85a2`), and `internal/cli/todo_undone_test.go:277-278` asserts it against `kanban.LandedRef`. | Refusal text asserted against the **resolved** ref; the existing test updated to resolve rather than to read a constant. |
| **AC-TLS-004** | REQ-TLS-003, REQ-TLS-004 | **Given** a resolved ref that does not exist **when** the landed check runs **then** it answers `unknown`, and the answer sits in a field a caller must read. | `git log origin/no-such-ref --perl-regexp --grep='\bt331\b' --oneline ; echo rc=$?` → `fatal: ambiguous argument 'origin/no-such-ref': unknown revision or path not in the working tree.` / `rc=128`. `Landed` maps that to `(false, err)` (`internal/kanban/prlink_landed.go:76-79` @ `3de2f85a2`), so a caller that drops the error reads `not landed`. | Three-valued answer; a caller that ignores the answer field does not compile. Test asserts `unknown` ≠ `not-landed` for the unresolvable ref. |
| **AC-TLS-005** | REQ-TLS-004 | **Given** an unresolvable landed ref **when** `moai todo pr` renders **then** the outcome is distinct from `no-link` and names the ref. | `ResolveCardPRLink` returns the `no-link` outcome with the error alongside (`internal/kanban/prlink.go:179-181` @ `3de2f85a2`); `runTodoPR` collects the card as degraded (`internal/cli/todo_pr.go:141-144`), prints a stderr note (`:147-151`), and prints the same five-column row (`:175-176`). The rendered row for a degraded card is identical to a genuinely unstarted one. | A fifth outcome kind rendered; golden output distinguishes the two rows. |
| **AC-TLS-006** | REQ-TLS-005, REQ-TLS-006 | **Given** `--require-landed` **when** the query is unanswerable **then** stdout differs from the satisfied case. | Indistinguishability rests on the **code**: `todoRequireLanded` (`internal/cli/todo.go:417-431` @ `3de2f85a2`) returns `nil` at `:423` (unanswerable, after a stderr note) and at `:430` (landed), writing stdout on neither path, so `done` prints the same bytes. The tests establish the **exit code**, one invocation, verbatim: `go test ./internal/cli/ -count=1 -run 'TestTodoDone_RequireLanded\|TestTodoDone_NoLandingQueryWithoutTheFlag' -v` → `--- PASS: TestTodoDone_RequireLandedRefusesWhenNotLanded (0.13s)` / `--- PASS: TestTodoDone_RequireLandedProceedsWhenInconclusive (0.13s)` / `--- PASS: TestTodoDone_NoLandingQueryWithoutTheFlag (0.20s)` / `ok github.com/modu-ai/moai-adk/internal/cli 1.395s`. Note `TestTodoDone_NoLandingQueryWithoutTheFlag` (`:306-335`) never inspects stdout — it asserts `err == nil` and a subprocess count. | A new test captures stdout for all three answers (`landed` / `not-landed` / `unknown`) and asserts three distinct verdict tokens. `TestTodoDone_RequireLandedProceedsWhenInconclusive`'s stdout prefix assert (`:297`) is updated deliberately, not incidentally. |
| **AC-TLS-007** | REQ-TLS-006 | **Given** `--require-landed` and an `unknown` answer **when** `done` runs **then** the card is archived and the exit code is 0. | **RED cell: none** — a policy-preservation criterion, classified as a **regression-guard** (§C). `TestTodoDone_RequireLandedProceedsWhenInconclusive` PASSes today (verbatim above) and must keep passing on its policy assertions. | That test still passes on archive + exit code after M3 rewrites its stdout expectation. |
| **AC-TLS-008** | REQ-TLS-008, REQ-TLS-009 | **Given** a seeded queue **when** every landing-detection path runs — `todo pr`, `todo pr --json`, the landed path, and the degraded path — **then** the queue directory is byte-identical afterwards and no card's `state` has moved. | **RED established by mutation, observed, and reverted** — see §D.1 below for the mutant, the command, and its verbatim failing output. The plain source-sweep form this criterion used to carry (grep for `Items[i].State`, `ArchiveCard`, `Drop`) is **rejected**: the planted mutant satisfies all three greps (each returns rc 1, zero matches) while dropping every card in the queue. | `TestTodoPR_QueueDirUnchanged` (`internal/cli/todo_pr_test.go:142-199` @ `3de2f85a2`) stays green across all four sub-cases after the change, extended with a landed-and-`completed` fixture. GREEN verbatim, this run: `--- PASS: TestTodoPR_QueueDirUnchanged (0.50s)` with all four sub-cases PASS. |
| **AC-TLS-009** | REQ-TLS-010, REQ-TLS-011 | **When** both `todo.md` copies are read **then** they carry the resolved ref, the `unknown` outcome, the stdout verdict token, the unweakened [HARD] operator-only rule, and the stated remaining limit. | Both copies @ `3de2f85a2` describe four outcomes and a two-valued landed check (`.claude/skills/moai/workflows/todo.md:51` for the four outcomes, `:59-67` for the opt-in limit note, `:53-57` for the [HARD] rule). The mirror is byte-identical: `diff` of the two copies is empty, 241 lines each (measured this run). | Both files updated in one change; `make build` exit 0; a grep proves the embedded bundle carries the edit; `diff` of the two copies stays empty. |
| **AC-TLS-010** | REQ-TLS-007 | **Given** a card that is `picked` with zero commits (the t338 shape) **when** `todo pr` renders **then** the row is distinguishable from a `queued`, never-started card. | `git log origin/develop --perl-regexp --grep='\bt338\b' --oneline \| wc -l` → `0` (`origin/develop` `c6aa61346`, 2026-08-28T13:15Z), so its S2 cell reads exactly like C1's. And the row cannot carry the difference: `todo pr` prints five columns — `CardID, Kind, PRs, Confidence, text` (`internal/cli/todo_pr.go:175-176` @ `3de2f85a2`) — none of which is the queue `state`. | The rendered row carries the queue `state` beside the link outcome, so `queued`+`no-link` and `picked`+`no-link` are different rows. **Stated limit, asserted as part of the criterion**: this distinguishes *picked with no commits* from *never started*; it does **not** claim to detect work that produced no commit. |

---

## §C Severity classification

- **Release-blocking**: AC-TLS-001, 003, 004, 005, 006, 008, 009, 010.
- **Regression-guard (must stay green; a failure is a contract break, but there is no RED to flip)**:
  AC-TLS-002, AC-TLS-007. Both are no-change assertions and say so in their RED cell rather than
  claiming a failure that was never observed.

Every release-blocking criterion above carries a RED that was actually observed — seven from
measurement, and AC-TLS-008 from a planted mutant (§D.1).

## §D Indirect verification accepted

- The misjudgement inputs are **historical**: t200's squash and t293/t310/t322's develop landings
  cannot be re-created live. Each criterion therefore asserts against a **fixture built from the
  measured shape**, with the measurement quoted in the RED cell so the fixture's fidelity is
  auditable.
- t338's "work existed but produced no commit" account is **operator testimony, not a measurement**
  this SPEC made. Only its zero-commit half is measured, and AC-TLS-010 states inside the criterion
  that the tooling does not claim to detect work that produced no commit.
- Windows behaviour arrives via CI job read-back; the local obligation stops at
  `GOOS=windows go vet ./internal/...` compile evidence.

### D.1 AC-TLS-008's RED, established by mutation

The criterion guards a non-goal (`spec.md` §B.4), so it has no natural failing input: today nothing
in the detection path writes. Under `verification-completeness.md` §2 that makes it unadopted until
a failure has been observed, and its previous source-sweep form additionally failed the **mutant
probe**. Both were closed by planting a mutant, observing the criterion go red, and reverting.

**The mutant** — a state write inside `runTodoPR`, reached through a differently-named helper and
using a pointer alias rather than the indexed-field assignment the old sweep scanned for:

```go
// in runTodoPR, before the outcome loop:
_ = reconcileQueueAfterLandingSweep(only)

func reconcileQueueAfterLandingSweep(_ string) error {
	return newTodoStore().Mutate(func(r *kanban.BacklogRecord) error {
		for i := range r.Items {
			p := &r.Items[i]
			p.State = kanban.BacklogStateDropped
		}
		return nil
	})
}
```

**The old form does not catch it.** All three of its greps return zero matches against the mutated
file (`rc=1` each): `grep -rn 'Items\[i\]\.State' internal/cli/todo_pr.go`, `grep -rn 'ArchiveCard'
internal/cli/todo_pr.go`, `grep -rn '\.Drop(' internal/cli/todo_pr.go`. A criterion built on those
greps passes while every card in the queue is silently dropped.

**The behavioural form does catch it.** `go test ./internal/cli/ -count=1 -run
'TestTodoPR_QueueDirUnchanged'`, verbatim (first sub-case; all four failed identically):

```
--- FAIL: TestTodoPR_QueueDirUnchanged (0.50s)
    --- FAIL: TestTodoPR_QueueDirUnchanged/linked_and_ambiguous (0.13s)
        todo_pr_test.go:177: queue directory changed across the invocation
            before:
            backlog.db d07c3917937fef96b68570b6053c9a48228119931bad2678a8760bc0fdb7a008
            ...
            after:
            backlog.db b85f9d65aa25e38c9da8795e484ac90afa1aabecf44b164009999e7391611daa
        todo_pr_test.go:193: backlog.lock mtime moved: ... — the read verb took the lock
```

**Reverted, and green again.** `internal/cli/todo_pr.go` restored (SHA-1 `a80ca6bd` before and
after the probe; `git status --short` shows no production-code change), and the same command returns
`--- PASS: TestTodoPR_QueueDirUnchanged (0.50s)` with all four sub-cases PASS.

The criterion is therefore adopted on its behavioural form: the whole-queue-directory digest, which
a state write cannot evade by renaming its helper or by dropping to raw SQL, because both change the
bytes it hashes.

## §E Closure gates

1. Every release-blocking AC PASSes with an evidence path resolvable from `progress.md` §E.2.
2. Both `todo.md` copies updated and `make build` run in the same change (Template-First), and their
   `diff` stays empty.
3. The subprocess census tests pass, or their budget change is stated and justified in `progress.md`.
4. `TestTodoPR_QueueDirUnchanged` green (AC-TLS-008) — the read-only ruling survived.
5. Nothing under `internal/kanban/backlog_*.go` is modified: this SPEC persists nothing, so a diff
   touching the store is a scope breach into card t359, not an implementation detail.

## §F Definition of Done

- The landed question is asked about the configured integration branch, and an unconfigured project
  is byte-identical to `3de2f85a2`.
- `landed`, `not-landed`, and `unknown` are three distinct answers, and `unknown` is visible on both
  the `todo pr` surface and the `todo done` stdout — so "the guard passed" and "the guard did not
  run" are no longer the same bytes.
- A `todo pr` row shows the queue `state` beside the link outcome.
- No path in the change closes, moves, drops, or re-states a card on the tooling's own authority,
  and that is asserted behaviourally rather than by name-matching.
- The doctrine and its template mirror say all of the above, including what the check still cannot
  answer: that a `landed` answer does not mean the card's last step landed.
