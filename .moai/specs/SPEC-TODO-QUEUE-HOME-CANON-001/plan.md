# plan.md — SPEC-TODO-QUEUE-HOME-CANON-001

## A. Context

Card t658 (Tier M, Class C). After t657 merged the queue DATA into the home SQLite
store, the CODE/DOCS layer must be consolidated so the divergence cannot recur.
The plan-phase survey on tree b66789479 (spec.md §1) established that path
RESOLUTION is already single-seam; the residual work is statement convergence
(foreman skill), a machine-verifiable single-resolution guard, and a recorded
disposition ledger for every legacy-JSON reader.

Key survey correction to the card premise: `internal/statusline/backlog.go:50`
no longer reads `backlog.json` directly — the direct read was removed by t306
(SPEC-TODO-SQLITE-001 M3+M4) and t510 (SPEC-STATE-ANCHOR-001 M1/M2); the file now
delegates to `kanban.BacklogCountsForRoot` (spec.md §1.1 A14). Recorded here so
the run phase does not chase a stale premise.

## B. Known Issues

- `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md:95`
  (= local copy, verified byte-identical): queue-watch script pins
  `d=.moai/state/todo` — a project-local path the resolution no longer uses for
  standard git-repo projects; the monitor is a silent no-op.
- `cmd/t657-merge/main.go:51` builds a `backlog.json` path outside the seam —
  boundary exception (t835), must be exempted from the REQ-003 guard.
- Adoption machinery (`resolveStateDir` legacy candidates, `queueExists`,
  `relocateQueueArtifacts`) intentionally reads legacy locations — keep until
  SPEC-TODO-QUEUE-HOME-MERGE-001 M5; the guard must not flag `internal/kanban`
  itself.

## C. Pre-flight

- [ ] Re-verify survey rows A14/B8 on the run-phase HEAD (`git log -- internal/statusline/backlog.go` + read the file) — confirm the direct JSON read is still absent.
- [ ] Re-run `diff -q` local vs template foreman SKILL.md before editing.
- [ ] Confirm `modernc.org/sqlite` remains a direct dependency (go.mod).
- [ ] Confirm no new `.moai/state/todo` canonical claims appeared between plan and run (`grep -rn "state/todo" internal/template/templates/`).

## D. Constraints

- Zero resolution-behavior change (spec.md §3).
- Template-First cycle: edit template source → `make build` → verify in local copy
  → commit (CLAUDE.local.md §2). Local `.claude/skills/moai-kanban-foreman/SKILL.md`
  must end byte-identical to the template mirror.
- Catalog hashes regenerated after template edits
  (`go run ./internal/template/scripts/gen-catalog-hashes.go --all`).
- Guard scope: `internal/` + `pkg/` only; `internal/kanban` exempt (it IS the seam);
  `cmd/` exempt (B6 boundary, t835).
- Affected-package tests only; no `go test ./...` locally (CLAUDE.local.md §4/§6).
- t704 owns `internal/template/templates/.moai/docs/todo-queue-storage.md`; do not
  edit that file in this run.

## E. Self-Verification

Planned §E.2 evidence at run-phase close:
1. `go test ./internal/kanban/... -run 'TestQueuePath|TestDocs'` (new guards) — output captured.
2. `go test` for each other affected package — output captured.
3. `diff -q` local vs template foreman SKILL.md — exit 0.
4. `git diff --stat` showing catalog.yaml hash refresh alongside the skill edit.
5. Grep proof: `grep -rn "state/todo" internal/template/templates/.claude/skills/` → 0 hits.
6. Mutation check: reverting the skill edit makes the new guard test FAIL (run once, then restore).

## F. Milestones (priority-ordered by decision reversibility — no time estimates)

### M1 — Disposition ledger finalization (Priority High)

Confirm or amend the B1-B9 dispositions (spec.md §1.2) against the run-phase HEAD.
Every production legacy-JSON reader ends with `converge` or a recorded
justification. Decision-dense: do FIRST, because it fixes the guard's exemption
list. No open clarifications at authoring — B6 handled as recorded exception;
re-confirm B6 in M1 if t835's scope changes before run-phase.

### M2 — Foreman skill canonical-path correction (Priority High)

Correct the queue-watch script in
`internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` to
reference the canonical home location (`~/.moai/db/<project-key>/todo`) as resolved
by `kanban.StateDirForRoot`, keeping the script's checksum-based change detection
intact. Regenerate catalog hashes; sync the local copy (byte-identical). Update the
description line if it names the watched path.

### M3 — Single-resolution guard tests (Priority High)

Add to `internal/kanban` (or the most fitting package) two machine-verifiable tests:
1. Docs-match-resolution test: parse the foreman skill's watch path from the
   TEMPLATE source tree and assert it equals `StateDirForRoot`'s output for a
   git-repo `t.TempDir()` fixture (positive control) and differs from the
   project-local path for that fixture.
2. Repo-scope seam test: walk `internal/` + `pkg/` non-test `.go` files; assert no
   file outside `internal/kanban` constructs a literal `backlog.json` queue path
   (string-literal scan, mechanical, modeled on
   `internal/web/todo_queue_read_test.go`). Exempt `cmd/` per D.
Both tests must FAIL on a mutated input (mutation check, E.6).

### M4 — Verification sweep + ledger closure (Priority Medium)

Run the affected-package suites, the parity diff, the catalog check, and the E.1-E.6
evidence capture; populate `progress.md` §E.2/§E.3; close the disposition ledger
(M1) with any run-phase discoveries adjudicated under REQ-006.

## G. Anti-Patterns

- Editing `internal/template/templates/.moai/docs/todo-queue-storage.md` (t704's).
- Touching `cmd/t657-merge` (t835's merge-core boundary).
- Widening the guard to flag `internal/kanban`'s own adoption machinery (B2/B4 are
  keep-until-M5 by design).
- Removing the temporary-origin or home-unresolvable fallbacks (REQ-005).
- Local-copy-first edit of the skill (Template-First violation).

## H. Cross-References

- SPEC-TODO-SQLITE-001 — the SQLite store cutover (t306).
- SPEC-WEB-TODO-QUEUE-001 — root relocation + pure/adopting split (A1/A2).
- SPEC-TODO-HOME-TEMP-GUARD-001 — the temporary-origin guard (REQ-005 retention).
- SPEC-TODO-QUEUE-HOME-MERGE-001 — t657 data merge; M5 retirement (out of scope).
- SPEC-STATE-ANCHOR-001 — statusline board-root seam (A14 convergence).
- SPEC-DOCS-TODO-TEMP-GUARD-001 — completed docs-site correction (t575).
- Sibling cards: t835 (merge verifier/dedup), t705 (temp-git anomaly), t706
  (factory queue truth), t704 (todo-queue-storage.md defect).
