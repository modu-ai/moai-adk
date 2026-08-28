# Implementation Plan — SPEC-TODO-LANDING-STATE-001

Card t331. Authored at tree `3de2f85a2` (worktree `.claude/worktrees/t331`, branch
`WT-card-landing-state`). Every citation below carries that tree SHA.

Milestones are ordered by **decision-reversibility**, not by execution convenience: the decisions
most likely to change on review come first (the ref-resolution semantics, the storage shape, the
user-visible stdout contract), and the mechanical work comes last.

---

## §A Tier classification

**Tier M.**

| Signal | Measurement @ `3de2f85a2` |
|---|---|
| Packages touched | 3 (`internal/kanban`, `internal/cli`, plus a read of `internal/config`) |
| Production call sites of the constant being replaced | 6 (`grep -rn 'LandedRef' --include='*.go' internal/`) |
| New storage objects | 1 table + 1 top-level record field, both purely additive |
| New CLI verb | 1 (`todo land`), plus one changed stdout line on `todo done` and one new outcome kind on `todo pr` |
| Doctrine files | 2 (the skill doc and its template mirror) |
| Migration required | none (REQ-TLS-017) |

Not Tier S: it changes a user-visible CLI contract and a persisted schema. Not Tier L: no new
subsystem, no cross-cutting refactor, and the blast radius is measured rather than estimated.

---

## §B Constraints carried into run-phase

1. **[HARD] Template-First.** `.claude/skills/moai/workflows/todo.md` has a mirror at
   `internal/template/templates/.claude/skills/moai/workflows/todo.md` @ `3de2f85a2`. Edit the
   template, run `make build`, then sync. A doc edited only in `.claude/` is deleted by the next
   `moai update` (CLAUDE.local.md §2.3).
2. **[HARD] Verification scope.** Run the affected packages
   (`go test ./internal/kanban/... ./internal/cli/...`), then push and read CI for the full-suite
   verdict. Do not run `go test ./...` locally.
3. **[HARD] `internal/cli` timeout floor.** That suite has been measured at 336s standalone; use a
   600s Bash timeout.
4. **[HARD] Exit-code reading.** A count piped through `wc -l` reports 0 for a command that failed;
   read the exit code separately or use `set -o pipefail`.
5. **[HARD] No prompting.** `internal/cli/todo.go:20` @ `3de2f85a2` records the subagent boundary —
   this command surface never prompts. The new verb inherits it.
6. **Subprocess census.** `internal/cli/todo_undone_test.go:311-334` and the `todo pr` census assert
   exact subprocess counts. Any new `git` call must be routed through `todoRunCommand`
   (`internal/cli/todo_pr.go:56-66` @ `3de2f85a2`) and its count budget stated, or those tests go red
   for the right reason and must be updated deliberately, not silently.

---

## §C Open questions for the Implementation Kickoff gate

- **[NEEDS CLARIFICATION: landing-verdict token spelling]** REQ-TLS-009 requires a stdout landing
  verdict on every `todo done`. Two shapes are viable: a suffix on the existing line
  (`done t331 landing=landed`) or a second line. The suffix changes the exact stdout string that
  `TestTodoDone_RequireLandedProceedsWhenInconclusive` (`internal/cli/todo_undone_test.go:296` @
  `3de2f85a2`) asserts with `strings.HasPrefix(stdout, "done t1")` — a prefix assert, so a suffix
  survives it. A second line does too. The choice affects any operator script parsing `todo done`.
  Recommendation: suffix, because it keeps one line per act.
- **[NEEDS CLARIFICATION: exit code on `unknown`]** REQ-TLS-012 forbids refusing, so the exit code
  stays 0 in the current reading. An alternative is a distinct non-zero-but-non-fatal code. That
  would change the contract for every unattended caller and is **not** recommended; it is named here
  so the gate can rule on it rather than have it decided inside a milestone.
- **[NEEDS CLARIFICATION: which verb records the observation]** §B.4 requires an operator-issued
  verb. `moai todo land <id>` is proposed (shaped after `todo relate`: writes a record, touches no
  card). The alternative — recording as a side effect of `todo done` — is weaker, because it records
  the landing only at the moment the operator no longer needs it.

---

## §D Milestones

### M1 — the ref-resolution semantics (highest change likelihood)

The decision most likely to be revised on review, because it changes what every existing consumer's
question means.

- Replace `const LandedRef` (`internal/kanban/prlink_landed.go:28` @ `3de2f85a2`) with a resolver:
  a `LandedRefFor(projectRoot string) string` returning `origin/<worktree_base_branch>` when
  configured and `origin/main` when not (REQ-TLS-001, REQ-TLS-002).
- Reuse `config.LoadWorktreeBaseBranch` (`internal/config/loader_worktree_base.go:28` @ `3de2f85a2`).
  Do not re-parse the YAML.
- Thread the resolved ref through `LandedGrepArgs` and `GitLandedQuerier`
  (`internal/kanban/prlink_landed.go:39-81` @ `3de2f85a2`) as a field, not a package global — the
  package is a pure function of its inputs by ruling (`internal/kanban/prlink.go:17-19` @
  `3de2f85a2`) and must stay so.
- Update the six user-facing text sites (REQ-TLS-003) so they name the resolved ref.

**Reversibility note.** Empty configuration must be byte-identical to today. That is the property
that makes this milestone safe to land before the rest.

### M2 — the three-valued answer and the `unknown` outcome (user-visible contract)

- Replace the `(bool, error)` return of `Landed` with a three-valued answer (REQ-TLS-005..007).
- Map an unresolvable ref to `unknown` rather than to `false` (REQ-TLS-006). The probe at
  `3de2f85a2` — `git log origin/no-such-ref … ; rc=128` — is the input.
- Add the `unknown` outcome kind to the resolver (`internal/kanban/prlink.go:31-49` @ `3de2f85a2`),
  keeping the existing four intact, and render it distinctly in `todo pr` (REQ-TLS-008).
- Keep the resolver a pure function; the querier stays the only I/O seam.

### M3 — the stdout landing verdict on `todo done` (user-visible contract)

- Emit exactly one landing-verdict token on stdout per invocation, including without the flag
  (REQ-TLS-009..011), resolving the §C spelling question first.
- Preserve the permissive policy on `unknown` (REQ-TLS-012) — this milestone changes the report, not
  the decision.
- Update `internal/cli/todo_undone_test.go:287-302` @ `3de2f85a2` deliberately: its current PASS is
  the RED baseline this SPEC exists to move (acceptance.md AC-TLS-006).

### M4 — landing observations: storage shape (persisted schema)

- Add a `landings` top-level array to `BacklogRecord` (`internal/kanban/backlog_store.go:186-193` @
  `3de2f85a2`), following the `findings` / `archived` precedent exactly: absent in older files,
  always rendered as an array, never null (REQ-TLS-013, REQ-TLS-016).
- Add a `landings` table to `backlogDDL` (`internal/kanban/backlog_sqlite.go:101-135` @ `3de2f85a2`)
  as `CREATE TABLE IF NOT EXISTS`. No `ALTER TABLE`, no CHECK change, no `schema_version` bump
  (REQ-TLS-017).
- Leave `BacklogItem`'s five fields untouched (REQ-TODO-013).
- Confirm the downgrade story: a binary predating this change opens the queue and leaves the rows
  intact (REQ-TLS-018).

### M5 — the recording verb

- `moai todo land <id>` — resolve the ref, query it, and record one observation under the lock
  (REQ-TLS-014, REQ-TLS-023).
- The observation names the commit **as it exists on the resolved ref** (REQ-TLS-015). The t200
  measurement in spec.md §A.3 is the guard: `7fc161b36` must never be recorded.
- The verb changes no card (REQ-TLS-021) — the `todo relate` shape.
- `todo pr` is not touched (REQ-TLS-024).

### M6 — live SPEC status on the read surface

- Render the card's live SPEC status through `kanban.ReadCardStatus`
  (`internal/kanban/status_read.go:99` @ `3de2f85a2`), carrying its `Source` tag (REQ-TLS-019).
- An unresolved status renders as unresolved, never as a status value (REQ-TLS-020).
- This is a read-path addition on `todo pr`; keep it inside the existing per-invocation cost budget
  (the status read is a `git show` blob read, no network).

### M7 — doctrine and template mirror (mechanical)

- Update both copies of `todo.md` (REQ-TLS-025): the new verb, the `unknown` outcome, the stdout
  verdict token, and the restated [HARD] operator-only rule.
- Restate the remaining limit (REQ-TLS-026): `landed` means the ref's history names the card, not
  that the card's last step landed.
- `make build`, then verify the embedded bundle carries the edit.

---

## §E Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The ref correction changes `todo pr`'s answers for **every** card at once; a lead reading the new output may over-trust it. | REQ-TLS-026 forces the remaining limit into the doctrine. spec.md §A.5 records that the ref correction closes the false negative and leaves the false positive open. |
| R2 | A downstream project with `worktree_base_branch` empty silently keeps the old (wrong-for-them) behaviour. | Deliberate: the neutral-empty ruling is inherited from `internal/config/types.go:164-174` @ `3de2f85a2`. `moai doctor` already surfaces an unresolvable configured branch. |
| R3 | Adding a `git` call to a read path breaks the subprocess census tests. | §B.6: route through `todoRunCommand` and state the budget. Treat a census failure as a design signal, not a test to relax. |
| R4 | The recorded SHA is taken from the wrong ref after a later ref reconfiguration, making an old observation unresolvable. | REQ-TLS-014 records the ref **with** the SHA, so an observation is self-describing and a stale one is detectable rather than misleading. |
| R5 | Scope creep into automatic closure. | §B.4 non-goal + §D exclusions + AC-TLS-013, which asserts the absence mechanically rather than by convention. |
| R6 | The `unknown` outcome is added to the JSON shape of `todo pr --json`, breaking a consumer that switches exhaustively on four kinds. | The kind field is documented as a closed set; adding a fifth value is a contract change. Name it in the sync-phase notes and check `internal/web` for consumers before landing M2. |

---

## §F Cross-references

- spec.md §A.6 — the state table this plan implements.
- acceptance.md — the RED baselines, each measured at `3de2f85a2`.
- `SPEC-TODO-DESTRUCTIVE-GUARD-001` §B.2 — the inherited t330/t331 boundary.
- `SPEC-WORKTREE-BASEREF-001` — the config resolver reused in M1.
