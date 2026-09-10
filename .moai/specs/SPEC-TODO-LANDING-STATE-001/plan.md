# Implementation Plan — SPEC-TODO-LANDING-STATE-001 (half A)

Card t331. Authored at tree `3de2f85a2` (worktree `.claude/worktrees/t331`, branch
`WT-card-landing-state`), re-verified at HEAD `11426a128` (0.2.0) and again at HEAD `45cff0f59`
(0.3.0 delta fix); no cited source file changed across any of the three, so every `@ 3de2f85a2` pin
below still resolves.

**Scope: half A — the discriminator.** The landing-evidence half (storage, the recording verb, an
observed commit on `todo pr`, the live SPEC-status read) is card **t359**, which depends on this one.
See `spec.md` §A.5 and §D.

Milestones are ordered by **decision-reversibility**, not by execution convenience: the decisions
most likely to change on review come first (the ref-resolution semantics, then the two user-visible
output contracts), and the mechanical work comes last.

---

## §A Tier classification

**Tier M — 11 requirements against a ceiling of 16, 10 acceptance criteria against a ceiling of 16.**

| Signal | Measurement |
|---|---|
| Packages touched | 3 (`internal/kanban`, `internal/cli`, plus a read of `internal/config`) |
| Production call sites of the constant being replaced | **7** (`grep -rn 'LandedRef' --include='*.go' internal/` @ `3de2f85a2` → 12 lines: 1 comment, 1 declaration, 7 production uses, 2 test lines) |
| Files expected to change | 6-8: `prlink_landed.go`, `prlink.go`, `todo_pr.go`, `todo.go`, two test files, two `todo.md` copies |
| New storage objects | **none** — this SPEC persists nothing (`spec.md` §B.2) |
| New CLI verb | none. One new outcome kind on `todo pr`, one new column on its row, one changed stdout line on `todo done` |
| Doctrine files | 2 (the skill doc and its template mirror) |
| Migration required | none — nothing is stored |

Not Tier S: more than 5 files, and it changes a user-visible CLI contract on two surfaces. Not
Tier L: no new subsystem, no persisted schema, no cross-cutting refactor, and the blast radius is
measured rather than estimated. The prior draft's 26 requirements were the symptom of two SPECs in
one document, not of a tier misjudgement; with the storage half moved to t359 the count sits inside
the Tier M ceiling with room to spare.

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
6. **Subprocess census.** `internal/cli/todo_undone_test.go:310-334` and the `todo pr` census assert
   exact subprocess counts. Any new `git` call must be routed through `todoRunCommand`
   (`internal/cli/todo_pr.go:57-65` @ `3de2f85a2`) and its count budget stated, or those tests go red
   for the right reason and must be updated deliberately, not silently.
7. **[HARD] Template neutrality on the mirror edit.** REQ-TLS-010 requires editing
   `internal/template/templates/.claude/skills/moai/workflows/todo.md`, and changes under that path
   trigger `.github/workflows/template-neutrality-check.yaml`, which rejects SPEC IDs, REQ tokens,
   internal dates, and commit SHAs in template content. This SPEC's own vocabulary is full of all
   four, so the mirror edit must be written in neutral prose: describe the resolved ref, the
   `unknown` outcome, and the verdict token without naming `SPEC-TODO-LANDING-STATE-001`,
   `REQ-TLS-*`, `3de2f85a2`, or any date.
8. **[HARD] Persist nothing.** Nothing under `internal/kanban/backlog_*.go` is modified. A diff
   touching the store is a scope breach into card t359 (acceptance.md §E gate 5).

---

## §C Questions raised at plan-phase, and how each was settled

No clarification marker remains open in this artifact. Two of the three raised at 0.1.0 were
resolvable from the sources and are ruled below; the third left with the storage half.

**1. The landing-verdict token — RULED, and the ruling has two parts with different sources.**

*Part (a), the token's spelling — decided by the SPEC.* `spec.md` §A.6 Table 2 writes the answer the
same way in every row (`landing=landed`, `landing=not-landed`, `landing=unknown`), so the marker
recorded a choice already made elsewhere in the same document. The spelling is inherited, not chosen
here.

*Part (b), the token's placement — a plan-phase design ruling, not a derivation.* Table 2 decides
nothing about **where** the token sits, and `REQ-TLS-005` ("exactly one landing-verdict token on
stdout") is satisfied by a suffix and by a second line equally, so nothing in the SPEC constrains it.
The plan rules the suffix form — `done <id> landing=<verdict>` — on one stated reason: **one line per
act**, because a second line would give an operator script two records for one event. That is a
design preference held at plan-phase and is labelled as such; it is open to reversal on review at no
cost to any requirement. (An earlier draft presented the placement as something "the SPEC had in fact
decided", which overclaimed its source.)

Mechanically the suffix survives the one existing assertion:
`TestTodoDone_RequireLandedProceedsWhenInconclusive` asserts
`strings.HasPrefix(stdout, "done t1")` (`internal/cli/todo_undone_test.go:297` @ `3de2f85a2`), and a
suffix does not disturb a prefix.

**2. The exit code on `unknown` — RULED: it stays 0.**
Settled by two sources rather than by preference. REQ-TLS-006 forbids refusing on `unknown`, and
`spec.md` §D excludes "reversing the proceed-on-unanswerable policy" outright; a distinct non-zero
code is that reversal wearing different clothes, because every unattended caller treats non-zero as
refusal. The repair is on the reporting axis, not the policy axis (`spec.md` §B.5) — the stdout
token is what makes the two outcomes distinguishable, and it does that without touching the exit
code any caller already depends on.

**3. Which verb records the observation — RETIRED, not answered.**
The question only existed for the storage half and left with it (card t359). Nothing in this SPEC
records anything, so there is no verb to choose.

---

## §D Milestones

### M1 — the ref-resolution semantics (highest change likelihood)

The decision most likely to be revised on review, because it changes what every existing consumer's
question means.

- Replace `const LandedRef` (`internal/kanban/prlink_landed.go:28` @ `3de2f85a2`) with a resolver:
  a `LandedRefFor(projectRoot string) string` returning `origin/<worktree_base_branch>` when
  configured and `origin/main` when not (REQ-TLS-001).
- Reuse `config.LoadWorktreeBaseBranch` (`internal/config/loader_worktree_base.go:28` @ `3de2f85a2`).
  Do not re-parse the YAML.
- Thread the resolved ref through `LandedGrepArgs` and `GitLandedQuerier`
  (`internal/kanban/prlink_landed.go:39-81` @ `3de2f85a2`) as a field, not a package global — the
  package is a pure function of its inputs by ruling (`internal/kanban/prlink.go:18-21` @
  `3de2f85a2`) and must stay so.
- Update the **seven** user-facing text sites (REQ-TLS-002) so they name the resolved ref:
  `prlink_landed.go:44`, `:78`; `todo_pr.go:75`, `:87`; `todo.go:357`, `:399`, `:428`.
- Update the four **doc comments** that hardcode `origin/main` in prose, which the `LandedRef` grep
  does not find and which silently mislead the next reader once the ref resolves:
  `prlink_landed.go:26-27`, `:52`, `prlink.go:31`, `prlink.go:42` @ `3de2f85a2`.

**Reversibility note.** Empty configuration must be byte-identical to today. That is the property
that makes this milestone safe to land before the rest.

### M2 — the three-valued answer and the `unknown` outcome (user-visible contract)

- Replace the `(bool, error)` return of `Landed` with a three-valued answer (REQ-TLS-003).
- Map an unresolvable ref to `unknown` rather than to `false` (REQ-TLS-004). The probe
  `git log origin/no-such-ref … ; rc=128` is the input.
- Add the `unknown` outcome kind to the resolver (`internal/kanban/prlink.go:30-48` @ `3de2f85a2`),
  keeping the existing four intact, and render it distinctly in `todo pr` (REQ-TLS-004).
- Keep the resolver a pure function; the querier stays the only I/O seam. `Landed` keeps returning a
  boolean fact and no commit (`SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10, enforced at
  `prlink_landed.go:62-67` @ `3de2f85a2`) — the answer's *arity* changes, not what it may carry.

### M3 — the stdout landing verdict on `todo done` (user-visible contract)

- Emit exactly one landing-verdict token on stdout per invocation, including without the flag
  (REQ-TLS-005), in the `done <id> landing=<verdict>` form ruled in §C.
- Preserve the permissive policy on `unknown` (REQ-TLS-006) — this milestone changes the report, not
  the decision, and the exit code stays 0 (§C).
- Update `internal/cli/todo_undone_test.go:287-303` @ `3de2f85a2` deliberately: its current PASS is
  the RED baseline this SPEC exists to move (acceptance.md AC-TLS-006).

### M4 — the queue state on the `todo pr` row

- Add the card's queue `state` to the rendered row (REQ-TLS-007), so a `queued` card with no link and
  a `picked` card with no link stop reading alike (`spec.md` §A.6 C1 vs C2).
- The row is printed at `internal/cli/todo_pr.go:175-176` @ `3de2f85a2` — five columns today
  (`CardID, Kind, PRs, Confidence, text`), none of them the state. The value is already in hand:
  `runTodoPR` holds `rec.Items` and builds a text map from it at `:167-169`.
- **This is a contract change on a machine-readable surface, and is stated as one: the row goes from
  five tab-separated columns to six.** The new `state` column sits **between `Confidence` and
  `text`** — appended after every fixed-width-ish field and before the free-text tail, so a consumer
  splitting on tabs and reading the last field still reads the card text. It is nonetheless a column
  count change: a consumer doing `cut -f5` gets the state where it used to get the text. Same class
  as R6's fifth outcome kind, and it goes into the same sync-phase note.
- No new subprocess, no new query. This is a column over data already loaded.

### M5 — doctrine and template mirror (mechanical)

- Update both copies of `todo.md` (REQ-TLS-010): the resolved ref, the `unknown` outcome, the stdout
  verdict token, and the restated [HARD] operator-only rule. Write the mirror in neutral prose
  (§B.7) — no SPEC id, no REQ token, no date, no commit SHA.
- Restate the remaining limit **inside the `todo pr` outcome list** (REQ-TLS-011, both clauses):
  (a) `landed` means the resolved ref's history names the card, not that the card's last step landed;
  (b) `unknown` means the question could not be asked and is not evidence of not-landed. The existing
  `--require-landed` limit note (`todo.md:59-67` @ `3de2f85a2`) stays unweakened — the delta is that
  the limit stops being stated only there. See acceptance.md §D.2 for the measured gap.
- `make build`, then verify the embedded bundle carries the edit and that `diff` of the two copies
  is empty.

---

## §E Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The ref correction changes `todo pr`'s answers for **every** card at once; a lead reading the new output may over-trust it. | REQ-TLS-011 forces the remaining limit into the doctrine. `spec.md` §A.5 records that the ref correction closes the false negative and leaves the false positive open for card t359. |
| R2 | A downstream project with `worktree_base_branch` empty silently keeps the old (wrong-for-them) behaviour. | Deliberate: the neutral-empty ruling is inherited from `internal/config/types.go:164-174` @ `3de2f85a2`. `moai doctor` already surfaces an unresolvable configured branch. |
| R3 | Adding a `git` call to a read path breaks the subprocess census tests. | §B.6: route through `todoRunCommand` and state the budget. Treat a census failure as a design signal, not a test to relax. M4 adds no subprocess at all. |
| R4 | Scope creep into automatic closure — made *more* attractive by this SPEC, because after it every develop-landed card starts answering `landed`. | `spec.md` §B.4 non-goal + §D exclusions + AC-TLS-008, which asserts the absence **behaviourally** (whole-queue byte identity) rather than by name-matching, and whose RED was established by planting a mutant that evades every name-based check (acceptance.md §D.1). |
| R5 | Scope creep into card t359's storage axis mid-implementation ("we are already here, just add the column"). | acceptance.md §E gate 5 and §B.8: nothing under `internal/kanban/backlog_*.go` is modified, so the breach is visible in the diff rather than argued about in review. |
| R6 | The `unknown` outcome is added to the JSON shape of `todo pr --json`, breaking a consumer that switches exhaustively on the four kinds. | **The in-repo half is measured and is zero, not deferred.** `grep -rn 'PRLink' --include='*.go' .` excluding tests and `internal/kanban/prlink*` returns matches in exactly one file — `internal/cli/todo_pr.go:135`, `:141`, `:176`, `:181-182` — and it formats `o.Kind` with `%s` (`:175-176`) rather than switching on it, so there is no exhaustive in-repo switch to break. `internal/web` exists (108 Go files) and contains **zero** `PRLink` references (`grep -rn 'PRLink' internal/web` → rc 1). The residual risk is external consumers of `todo pr --json`, which cannot be greped and is genuinely open: name the new kind in the sync-phase notes. Note also that the "closed set, switch exhaustively" wording belongs to `PRLinkConfidence` (`prlink.go:50-51` @ `3de2f85a2`), not to `PRLinkKind`, whose comment (`:30-33`) says only "one of the four … outcome kinds" — a fifth value is still a contract change, but a weaker one than the earlier draft implied. **The M4 column-count change (five tab-separated columns to six) is the same risk on the plain-text surface and rides the same mitigation: name BOTH the new outcome kind and the new column in the sync-phase notes.** |

---

## §F Cross-references

- spec.md §A.6 — the state table this plan implements, on the shipped surfaces S1 / S2 / S4. The S3
  (hand ancestry) column is recorded there, not implemented here: this SPEC changes nothing about it
  and no milestone below touches it.
- acceptance.md — the RED baselines; §D.1 carries the planted-mutant RED behind AC-TLS-008.
- `SPEC-TODO-DESTRUCTIVE-GUARD-001` §B.2 — the inherited t330/t331 boundary.
- `SPEC-WORKTREE-BASEREF-001` — the config resolver reused in M1.
- **Card t359** — the landing-evidence half; depends on this SPEC landing first (`spec.md` §A.5).
