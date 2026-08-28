# Acceptance Criteria — SPEC-TODO-DESTRUCTIVE-GUARD-001

## §A Standing preconditions

### A.1 Verification isolation [HARD]

The live queue in this repository is in active use by six concurrent lanes plus the lead. **No criterion below may be verified against the live queue.** Every runnable check runs in an isolated repository:

```bash
REPO=$(mktemp -d) && cd "$REPO" && git init -q && mkdir -p .moai/state/todo
```

A criterion executed against the repository's own `.moai/` is void, and any evidence produced that way is inadmissible.

### A.2 Exit-code reading [HARD]

Every criterion asserting a non-zero exit reads it as:

```bash
out=$(cmd 2>&1); rc=$?
```

Never `$?` after a pipe — that reports the pipe's status, and has already produced one false `rc=0` reading in this project.

### A.3 Byte-identity comparison

"Byte-identical" is asserted against the output of `moai todo export-json`, captured before and compared after, in the same isolated repository.

There is only ever one live engine to compare. `openEngine` (`internal/kanban/backlog_store.go:437-455`) migrates a JSON-only queue under the lock and then falls through to `openBacklogEngine(backlogSQLitePath(...))` on every path, and `internal/cli/todo_export.go:3-11` records that the swap is one-way with no engine-selecting knob. So no criterion below may compare "the record file itself on a legacy backend" — after the first mutating verb, no such live backend exists. `export-json` is the single serialization surface, which is also why it is the one that must carry the archive (REQ-TDG-015).

### A.4 Base-tree meaningfulness

Each refusal criterion below states the behaviour at the base tree (`812ee01fc`) that makes the check meaningful. A check passing identically before and after the change proves nothing.

---

## §B Criteria

### B.1 Reversal

**AC-TDG-001 — `undone` restores the card**
*Given* an isolated queue holding a card `t1`,
*When* `moai todo done t1` runs and then `moai todo undone t1` runs,
*Then* both exit `rc=0` and `moai todo list` names `t1` again.
*Base tree*: `undone` does not exist — `moai todo undone t1` exits non-zero with an unknown-command error. The base verb surface is the fifteen registered at `internal/cli/todo.go:137-141`: `add list done next unpick edit move drop undrop analyze relate unrelate why pr export-json`. Assert this by reading that `AddCommand` call or by counting `moai todo --help` subcommands — **not** by transcribing a list from the doctrine table, which omits `why`.

**AC-TDG-002 — the round trip is byte-identical, findings included**
*Given* an isolated queue holding cards `t1` and `t2` and a recorded finding naming `t1` (`moai todo relate t1 contains t2` or equivalent),
*When* the serialized queue is captured, `done t1` runs, `undone t1` runs, and the queue is serialized again,
*Then* the two serializations are byte-identical.
*Base tree*: `done t1` discards the card and calls `RemoveFindingsNaming("t1")` (`internal/cli/todo.go:347`), so no restore exists and the finding is unrecoverable. This criterion is the one that fails if a restore recovers the card row but loses its findings.

**AC-TDG-003 — `done` retains rather than discards**
*Given* an isolated queue holding `t1`,
*When* `done t1` runs,
*Then* the archive contains the `t1` row, verifiable without `undone` — by querying the archive table in `backlog.db`, and equivalently by reading the archived top-level field in `moai todo export-json` output.
*Base tree*: no archive table and no archive field exist; the row is gone from the record entirely.

### B.2 Storage

**AC-TDG-004 — the enum and the per-item contract are unchanged**
*Given* the implemented tree,
*When* `grep -n "CHECK (state IN" internal/kanban/backlog_sqlite.go` runs and the `BacklogState` constants and `BacklogItem` fields are read,
*Then* the CHECK still reads `CHECK (state IN ('queued','picked','dropped'))`, `BacklogState` still carries exactly three values, and `BacklogItem` still carries exactly five fields.
*Base tree*: identical — this criterion is a **freeze assertion**, and it is meaningful precisely because the natural implementation violates it. It fails if the fourth-state design is built instead.

**AC-TDG-005 — the schema version is not bumped, and downgrade survives**
*Given* a database created by the implemented binary and carrying at least one archived row,
*When* the stamped `schema_version` is read and a binary built from `812ee01fc` opens the same database,
*Then* the stamp reads `"1"` and the older binary opens it with `rc=0` rather than `ErrBacklogCorrupt`.
*Base tree*: not applicable in the forward direction; the reverse direction is the regression this guards, and `ensureSchema` (`backlog_sqlite.go:251-253`) aborts on any mismatch including a newer stamp.

**AC-TDG-006 — the archive survives the migration from a JSON-only queue**
*Given* an isolated repository holding **only** a legacy `backlog.json` with cards `t1` and `t2` and no `backlog.db`,
*When* `moai todo done t1` runs — which migrates the store under the lock before mutating — and then `moai todo undone t1` runs,
*Then* `backlog.db` now exists, both commands exit `rc=0`, and the `export-json` serialization is byte-identical across the round trip per A.3.
*Base tree*: `undone` does not exist.
*Note*: this criterion deliberately does **not** compare a live JSON engine against a live SQLite engine. Per A.3 no such pair is reachable — the first mutating verb migrates the store, and both arms would then run the same engine. What is being asserted is that a queue arriving in the legacy format ends up with a working archive, which is the reachable form of the "both backends" concern.

**AC-TDG-007 — archived rows are invisible to live-queue readers**
*Given* an isolated queue holding `t1` and `t2`, with `t1` archived by `done`,
*When* `moai todo list`, `moai todo next`, `moai todo why t1`, and `moai todo analyze` run,
*Then* `list`, `next` and `analyze` do not name `t1`; `moai todo why t1` emits exactly `t1: no findings`; and re-adding `t1`'s exact text is **not** reported as a duplicate of the archived card.
*Why `why` needs an exact-output assertion*: `internal/cli/todo_why.go:34-35` prints `"%s: no findings\n"`, echoing the argument id back. A grep for `t1` therefore matches even when `why` is reporting nothing at all, so "output does not name `t1`" is unsatisfiable for this verb and would fail a correct implementation. Assert the exact line instead.
*Base tree*: `t1` is absent from all readers because it was destroyed; this criterion asserts the same observable outcome now holds for a *retained* row — the two are distinguished by AC-TDG-003 running in the same repository.
*Scope*: `export-json` is deliberately **excluded** from this list and is asserted in the opposite direction by AC-TDG-015. It is the downgrade route, not a live-queue reader.

### B.3 Guards

**AC-TDG-008 — `--expect` refuses a mismatch**
*Given* an isolated queue where `t1` has text `alpha work`,
*When* `out=$(moai todo done t1 --expect beta 2>&1); rc=$?` runs,
*Then* `rc` is non-zero, `out` names the observed prefix, and the serialized queue is byte-identical to before the call.
*Base tree*: `done` accepts no `--expect` flag — the call exits non-zero with an unknown-flag error rather than a mismatch refusal, so the assertion must check the error names the observed prefix, not merely that `rc != 0`.

**AC-TDG-009 — `--require-landed` refuses on positive evidence**
*Given* an isolated repository whose integration ref contains no commit naming `t1`,
*When* `out=$(moai todo done t1 --require-landed 2>&1); rc=$?` runs,
*Then* `rc` is non-zero, `out` names the ref the answer is about, and the queue is byte-identical.
*Base tree*: the flag does not exist. Note the limit recorded in `spec.md` §A.4 — a card whose *run* commit names it reports as landed; this criterion asserts the refusal path only, not phase awareness.

**AC-TDG-010 — the flag is opt-in and costs nothing when absent**
*Given* an isolated queue and a `git` shim on `PATH` that records every invocation,
*When* `moai todo done t1` runs **without** `--require-landed`,
*Then* the shim records zero landing-query invocations and the command exits `rc=0`.
*Base tree*: `done` runs no landing query, so the recorded count is zero — this criterion **must** be paired with AC-TDG-009 in the same suite, since alone it passes identically at base.

**AC-TDG-011 — refusal writes nothing**
*Given* an isolated queue,
*When* each refusal path runs — absent id, `--expect` mismatch, `--require-landed` refusal, `undone` of an id never archived, `undone` into a reissued id —
*Then* every one exits non-zero and leaves the serialized queue byte-identical.
*Base tree*: the absent-id path already holds (`todo.go:351-353`); the four others do not exist and must be brought under the same guarantee.

### B.4 Boundaries

**AC-TDG-012 — nothing prompts**
*Given* the implemented tree,
*When* `done` and `undone` are exercised across every success, refusal, and error path with stdin closed (`< /dev/null`),
*Then* no invocation blocks and every one terminates with a determinate exit code.
*Base tree*: holds for `done`; this criterion extends the guarantee to the new verb and the new flags, preserving the `SUBAGENT BOUNDARY` discipline at `internal/cli/todo.go:20`.

**AC-TDG-013 — a reissued id refuses restore**
*Given* an isolated queue where `t1` was archived by `done` and a later `add` has reissued the id `t1` to a different card,
*When* `out=$(moai todo undone t1 2>&1); rc=$?` runs,
*Then* `rc` is non-zero, `out` names the collision, and the queue is byte-identical — the live `t1` is not overwritten.
*Base tree*: `undone` does not exist. Whether the id allocator can in fact reissue `t1` is itself worth confirming during M2 against `rec.LastSeq` normalization (`backlog_store.go:601-603`); if it provably cannot, this criterion is satisfied by a unit test on the restore guard rather than an end-to-end reissue.

**AC-TDG-014 — doctrine lands in both todo.md paths**
*Given* the implemented tree,
*When* `cmp .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md` runs, and each file is grepped for `undone`,
*Then* `cmp` exits `rc=0` and both files name the verb.
*Base tree*: the two files are byte-identical (13709 bytes, `cmp` clean at `812ee01fc`) and neither names `undone`, so the grep half fails at base while the `cmp` half passes — both halves are required.

### B.5 Downgrade and archive lifecycle

**AC-TDG-015 — the export carries the archive, and discloses what a downgrade costs**
*Given* an isolated queue holding live card `t2` and archived card `t1`,
*When* the two streams are captured **separately** and the written `backlog.json` is parsed:

```bash
err=$(moai todo export-json 2>&1 >/dev/null); rc=$?
```

*Then* `rc=0`, the parsed JSON contains `t1` under the archive field **and** `t2` under `items`, and **`err`** names the downgrade consequence — that a release predating the archive discards archived rows on its first write.
*Stream separation is load-bearing, not cosmetic*: `internal/cli/todo.go:20-22` contracts "one structured stdout line out, human-readable errors on stderr", and `internal/cli/todo_export.go:81-82` is that stdout line — the surface agents parse. A merged `2>&1` capture would pass an implementation that printed the disclosure to stdout, which REQ-TDG-015 forbids and which would pollute a machine-read stream. Do not simplify this back to a single capture.
*Base tree*: `export-json` marshals the whole record (`internal/cli/todo_export.go:69`), so the inclusion half would pass trivially were the field present; the **disclosure half is what fails at base**, and both halves are required. Run this criterion paired with AC-TDG-007, which asserts the opposite direction for live-queue readers — together they show the split is deliberate rather than an accident of marshalling.

**AC-TDG-016 — restore empties the archive entry**
*Given* an isolated queue where `t1` has been archived by `done`,
*When* `moai todo undone t1` runs and the archive is then read directly (archive table, or the archive field in `export-json`),
*Then* the archive no longer contains `t1`, and a second `moai todo undone t1` exits non-zero leaving the queue byte-identical.
*Base tree*: neither the archive nor `undone` exists. This criterion is implied by AC-TDG-002 — a retained entry would break byte-identity — and is asserted separately so the intended lifecycle is stated rather than inferred from a round-trip that could pass for the wrong reason.

---

## §C Definition of Done

- [ ] All 16 criteria pass in an isolated repository per §A.1.
- [ ] `go test ./internal/kanban/... ./internal/cli/...` passes (600s timeout floor on `internal/cli`); full-suite verdict read from CI, not run locally.
- [ ] `go vet ./...` clean.
- [ ] `cmp` parity confirmed on the todo.md pair, after `make build`.
- [ ] `schema_version` confirmed still `"1"` (AC-TDG-005).
- [ ] The `LandedRef` pre-flight in `plan.md` §C has been performed and its outcome recorded in `progress.md` §E.2.

Acceptance-criterion count: 16 (Tier M ceiling: 16).
