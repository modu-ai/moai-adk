# SPEC-LOCAL-INSTR-RECEPTION-001 — implementation plan

> Authored by card t1290 (plan phase), worktree `.claude/worktrees/t1290`, branch
> `WT-agents-local-migration`, HEAD `514ac7abe` at authoring. Tier L (constitutional: the file
> it migrates is the always-loaded maintainer doctrine of every lane in this repository), so the
> artifact set is 5 files + progress.md.

## §A Context

This SPEC exists because plan-audit iter4 (`.moai/reports/t1259/plan-audit-iter4.md`, D2,
blocking) found the parent's M3 unimplementable: the landed contract states a worktree session
does not receive `AGENTS.local.md` (`AGENTS.md:262`), the parent-of-parent measured exactly that
negative on a real nested worktree (`SPEC-INSTRUCTION-FILES-UNIFY-001/progress.md` §M1a), and
the parent held its only reception check as an ungated plan bullet — the shape its own
acceptance.md rejects. This SPEC moves that M3 here and makes its first milestone the reception
proof the parent never authored.

Scope: 8 requirements (`REQ-LIR-001`~`006`, plus moved `REQ-IFU-021`/`022`), 11 acceptance
criteria. The mechanism decision (force-track, `.worktreeinclude` fallback) lives in design.md;
this plan carries the milestone order and the measurement recipes.

## §B Blocking dependencies

**[HARD] The parent SPEC's M1 migration verb must land before this SPEC's M2 runs.** This
SPEC's migration is performed **with** `moai migrate local-instructions` — performing it by hand
forfeits the one real-file test of the verb (parent plan.md M3 rationale, carried here). The
parent is `status: draft` as of this writing (read in this tree, `514ac7abe`); its card is
t1259. This dependency is recorded in prose, not `depends_on:` frontmatter, matching the
parent's own convention: this SPEC's M1 (reception) is independent of the verb and must not be
serialized behind the parent's whole lifecycle.

**[HARD] The parent SPEC's owning card (t1259) owes the split-recording edit.** Because this
lane has no write authority over the parent's artifacts (its branch lives in another worktree),
the excision is recorded here and reported; the parent needs, at its next edit:

1. `spec.md` — remove `REQ-IFU-021` and `REQ-IFU-022` (§C.3), renumber nothing; add a HISTORY
   row recording the move to `SPEC-LOCAL-INSTR-RECEPTION-001` (citing iter4 D2); replace the
   `### Out of Scope — the worktree duplicate-load` paragraph (spec.md:200-203, whose
   "up to four local-instruction loads" rationale iter4 found inverted) with a pointer to this
   SPEC's reception gate.
2. `plan.md` — delete §D M3 (and the "Verify the parent SPEC's M1a worktree answer…" bullet it
   contained — that answer landed, negative, and is now this SPEC's M1); re-map the milestone
   table row.
3. `acceptance.md` — remove `AC-IFU-007` and `AC-IFU-024` rows from §D.1 and §D.2; the criteria
   live here now.

**t1175 (rules diet) has landed** (iter4 §1 measured `status: completed`) — no longer blocking.

## §C Constraints

- **Reception before migration, always.** M2 does not start while any of `AC-LIR-001`~`005` is
  unproven. A blocker at M1 stops the SPEC; it does not promote a hand-wave.
- **The lane does not push.** Integration is a lead-granted window (`moai integration acquire`),
  push is the lead's batch.
- **Template-First for every template-tree edit.** `internal/template/templates/` first, then
  `make build`, then root copies. Content neutrality: no card ids, SPEC ids, or internal dates
  in template content.
- **Committed-tree reads.** Every content measurement after M2 reads `git show HEAD:…`, never
  the working copy (iter4 D1: the working copy is gitignored and can silently pass while the
  merge lands nothing).
- **Never `go test ./...` locally**; affected packages only, CI is the full-suite verdict.
- **No `AskUserQuestion` from this lane**; blockers are reports to the lead.

## §D Milestones

| Milestone | Content | Gate |
|---|---|---|
| **M1** | Reception gate — probe built, negative observed, then reception proven per entry path | none (this is the gate) |
| **M2** | The migration proper — verb-run content move, committed-tree criteria | [HARD] M1 all green |
| **M3** | Documentation reconciliation + close duties | M2 landed |

### M1 — the reception gate (first act of the SPEC)

Sequence, in this order:

1. **RED observation (probe failability, `AC-LIR-005`).** Before any fixture exists, run the
   Claude probe (recipe below) from a real linked worktree of this repository and record
   `LOCAL_AGENTS_TOKEN = NOT_PRESENT` with the contract control present. This is the observed
   red that makes every later green meaningful. Expected today (the plan-phase expectation, not
   yet a run-phase observation): local token absent — the file does not exist in the tree
   (measured: `ls AGENTS.local.md` → No such file, `514ac7abe`) and no discovery path reaches a
   file that is not there.
2. **Fixture lands.** Create `AGENTS.local.md` containing a probe token line
   (`LOCAL_AGENTS_TOKEN = <value>`) plus the §0 skeleton, and force-add it
   (`git add -f AGENTS.local.md` — `.gitignore:275` ignores it) on this lane's branch. This is
   the reception instrument, not the migration: the real content move is M2's, and
   REQ-LIR-004's gate is what M2 owes, not M1.
3. **Reception legs, one per entry path** (`AC-LIR-001`~`004`): a fresh `moai cc -w` worktree; a
   `claude --worktree` worktree; `EnterWorktree` re-entry into an M1-created tree; a
   `moai codex -w` child. **Before probing, each Claude-side leg lands the fixture on the
   probe tree's own branch and commits it there.** Launcher-created trees branch from
   `origin/HEAD` = `origin/develop` (`worktree.baseRef` unset in both settings files), whose
   HEAD does not contain the fixture commit (measured: `git cat-file -e
   HEAD:AGENTS.local.md` fails on `origin/develop`) — a tree probed before the landing prints
   control-present + `LOCAL_AGENTS_TOKEN = NOT_PRESENT`, which is AC-LIR-005(b)'s own
   negative output, not a measurement. Inside the probe tree:
   `git show <lane-branch>:AGENTS.local.md > AGENTS.local.md && git add -f AGENTS.local.md
   && git commit -m "fixture: reception probe"` (force-add — `.gitignore:275` ignores it);
   where the launcher supports creating the tree from a given branch, create from the lane
   branch instead of landing afterwards. The `EnterWorktree` leg re-enters the `moai cc -w`
   tree, whose branch already carries the landed commit. The Codex leg reads the original
   project root directly (design.md §C) and needs no per-tree landing. Each Claude-side leg
   then runs the token probe from the new session's cwd; the Codex leg runs the Codex probe.
   The tracked-contract token (design.md §E) must read present in every leg — a leg where
   both tokens are absent measured a broken probe, not a reception failure.
4. **Negative control (`AC-LIR-005`, second half).** Create one worktree from a base commit
   without the file — its HEAD lacks the fixture, the deliberate opposite of the Claude
   legs' file-presence precondition — and observe `NOT_PRESENT` while the fixture-carrying
   legs read present — the probe can fail, and fails for the right reason.
5. **Mechanism verdict.** If any Claude-side leg is red, design.md §C's fallback ladder applies
   (`.worktreeinclude` entry, Template-First, then re-probe). If the gate cannot close after
   the ladder, return a blocker report with the measured matrix; do not proceed to M2.

**RED-now expectations (the negative, stated before M1 runs).** The probe prints
`LOCAL_AGENTS_TOKEN = NOT_PRESENT` today, for the right stated reason: the file does not exist
anywhere in this tree (measured at `514ac7abe`), `AGENTS.md:262` states the import-skip for the
worktree geometry, and §M1a measured `NOT_PRESENT` for a gitignored `AGENTS.local.md` in a real
nested worktree (`claude 2.1.283`, fixture `PAPA6`). The green path: the force-tracked fixture
makes the file present at the root of every worktree whose checkout contains the fixture
commit — which is why M1 step 3 lands the fixture on each probe tree's own branch before
probing — and the `@AGENTS.local.md` import
in the tracked `CLAUDE.md` resolves in-project — the one link M1 measures that no prior
measurement covers.

**Probe recipe (Claude side)** — the M1a pattern, adapted to this repository:

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
  MOAI_KANBAN_SETTINGS_INJECTED && timeout 180 claude -p "Answer only from your loaded \
instructions, run no tools. Two lines: (1) the exact version line of AGENTS.md, or ABSENT \
if AGENTS.md is not loaded; (2) the exact value of LOCAL_AGENTS_TOKEN, or NOT_PRESENT if \
absent." --model claude-haiku-4-5-20251001
```

Run from the probe session's cwd — only after the leg's file-presence precondition has
passed (`git -C <probe-tree> show HEAD:AGENTS.local.md >/dev/null` exits 0; M1 step 3's
landing is what makes it pass). Read on four signals: the file-presence precondition, exit
0, the version line present (control), and the token value line. The version line is the
control token (design.md §E); `ABSENT` on line 1 means the leg measured nothing, and so does
a failed precondition.

**Probe recipe (Codex side, `AC-LIR-004`)** — primary form:
`moai codex -w <probe-name> -- exec "<same two-line token question>"`; record the child's
answer verbatim. If the `-w` child cannot be driven non-interactively in this form, the first
execution pins the working invocation (launcher flag surface for a non-interactive run, or a
payload dump of the `developer_instructions` the launcher builds for a `-w` launch) and records
it verbatim in progress.md §E.2 before any pass is claimed. The AC is the token's presence in
the child session's model-visible input — the launcher's §8 claim is the hypothesis, the probe
is the verdict.

**Negative-control recipe** — `git worktree add /tmp/wtprobe-neg <base-without-file>` from a
scratch path, probe from there, remove the tree after. (A measurement fixture in `/tmp`, not a
card worktree; the launcher-mandatory rule governs card work, not probe instrumentation.)

### M2 — the migration proper (gated on M1)

- **[HARD] Entry condition:** `AC-LIR-001`~`005` all PASS in progress.md §E.2 with evidence
  paths. Any open reception leg = no M2.
- Re-measure the before-value at this milestone:
  `git show develop:CLAUDE.local.md | wc -m` (45,810 measured at `514ac7abe` — an expectation,
  not a bound; it moved 44,381 → 44,740 → 45,810 across three readings).
- Run `moai migrate local-instructions` on the develop-committed copy (the canonical copy per
  §0.1 — the lane's branch carries it after the develop absorb).
- The migrated file is force-added (`git add -f AGENTS.local.md`); `AC-IFU-007` reads the
  committed tree: `git show HEAD:AGENTS.local.md | wc -m` ≤ 39,999.
- §0 rewrite names `AGENTS.local.md`; `AC-IFU-024` reads it from the committed tree:
  `git show HEAD:AGENTS.local.md | grep -n -C1 'AGENTS.local.md'`.
- Operational procedure relocates to `.moai/docs/`; rules stay, substance unchanged.
- `AC-LIR-007` postcondition: `git ls-files AGENTS.local.md CLAUDE.local.md` lists exactly the
  former; `git show HEAD:CLAUDE.local.md` exits non-zero.
- Re-run the M1 probe once against the migrated geometry (the fixture is gone; the real file
  must still receive) — reception survives the content swap or M2 is not done.

### M3 — documentation reconciliation and close

- `REQ-LIR-006` / `AC-LIR-009`: update `AGENTS.md` §8's import-skip sentence and `CLAUDE.md`
  §18 to the measured behavior. Template-First: `AGENTS.md.tmpl` + template `CLAUDE.md`, then
  `make build`, then root copies; `agents-emit` if any agent file moves (none expected);
  content-neutral wording.
- Close duties: traceability diff (acceptance.md §D.2), `moai spec lint` clean, evidence paths
  under `.moai/reports/t1290/`, and the reception evidence retained — the probe outputs are
  this SPEC's load-bearing artifacts and outlive the milestone.

## §E Self-verification

Per-milestone, affected surfaces only. Every probe leg is read on four signals (file-presence
precondition, exit 0, control token present, local token line) — a leg failing the
precondition or missing the control measured nothing. Every
`go test`-shaped check (none planned in M1-M3; the verb's tests belong to the parent) would
carry the `-v` + `--- PASS:` + no-`no tests to run` discipline. At close: the §D.2
traceability diff, plus re-read of `git ls-files` postcondition on the merge head.

## §F Anti-patterns

- Starting M2 because the fixture probe went green on one entry path. All four legs, or no
  migration.
- Trusting `AGENTS.md` §8's `-w` claim without the Codex probe. It is the hypothesis M1 tests,
  not the result M1 reports.
- Measuring the migrated content from the working copy. Gitignored file, tracked-by-force:
  only `git show HEAD:` is evidence (iter4 D1).
- Deleting `CLAUDE.local.md` from the working tree before the committed tree carries
  `AGENTS.local.md` — a mid-milestone state where a crash leaves no file at all.
- Putting the reception verdict in a criterion note instead of a criterion — the exact shape
  iter4 rejected in the parent ("a duty recorded only inside a criterion's own note is a duty
  the close will not perform").
- Hand-editing the root `AGENTS.md`/`CLAUDE.md` without the template mirror (the next
  `moai update` reverts it), or leaking card/SPEC ids into template wording.
- Carrying the before-value forward instead of re-measuring. It moved three times in four days.

## §G Cross-references

- design.md — the mechanism decision and the fallback ladder M1 step 5 applies.
- research.md — every premise this plan cites, with command, output, and tree SHA.
- `.moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/plan.md` §D M3 — the milestone this SPEC
  took over, and the bullet that became this M1.
- `.claude/rules/moai/development/verification-completeness.md` — the two-cell adoption
  discipline acceptance.md applies.
