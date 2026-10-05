# plan.md — SPEC-AGENTS-WORKTREE-ROW-001

## §A Context

- Card t1071 (Class C, plan-phase delegation 2026-09-22). Landing tree: worktree
  `.claude/worktrees/t1071`, branch `WT-cross-harness-row`, base `cd99336bf` (local develop).
- The SPEC's §A carries the re-measured evidence (C1–C8) this plan builds on; every coordinate
  below is from the landing tree, measured 2026-09-22 in this run.
- Scope surfaces: exactly two files — the root contract and the template mirror. Card t1072 owns
  `worktree-integration.md` and `session-handoff-examples.md`; this plan never opens them for
  editing.

### §A.1 Scope surfaces (decision-reversibility order)

1. Root `AGENTS.md` — capability table (`:21-25`): highest change-likelihood surface, because
   the row's wording is a new contract clause every agent session reads.
2. `internal/template/templates/AGENTS.md.tmpl` — capability table (`:21-29`; seven rows — the
   fork carries four rows the root copy lacks, per plan-audit D1): same clause, fork copy;
   judged independently.
3. `internal/template/templates/AGENTS.md.tmpl` — `## 11. moai CLI Verbs` (`:293-303`): the most
   mechanical of the three edits (a table row in a stable inventory).
4. Regeneration (`make build`) and guard measurement: mechanical, last.

## §B Known Issues relevant to this card

- **B-marker-blind lint (t1057/t1020)**: spec-lint's REQ collection needs leading `- ` list
  markers; this SPEC's §C uses them, and the plan-phase lint below verifies collection fires.
- **Comment-hit citation (this card, C6)**: the dispatch's `:16-17` was a comment. All
  coordinates in this plan were re-anchored to code declarations (`:324`, `:272-299`) before
  authoring.
- **Fork misjudgment**: the mirror's `:291-305` Verbs table and the root's absent Verbs table
  are intentional divergence (C7); no "mirror parity" repair is attempted, and none is owed.
- **Line-number decay**: the `:21-25` / `:293-303` coordinates pin this tree (`cd99336bf`);
  edits above them in future trees shift them. The machine guards key on row-shape patterns, not
  line numbers, for exactly this reason.

## §C Pre-flight

Run before M1 (all read-only, batched):

```bash
# 1. Tree + branch re-read
git rev-parse --short HEAD && git branch --show-current
#    expect: cd99336bf (or the plan-phase commits this SPEC itself lands) / WT-cross-harness-row

# 2. Negative controls (pre-edit gap proof — measured 2026-09-22, all three zero)
grep -c '^| worktree-entry |' AGENTS.md                                   # expect 0
grep -c '^| worktree-entry |' internal/template/templates/AGENTS.md.tmpl  # expect 0
grep -c '^| `moai codex`' internal/template/templates/AGENTS.md.tmpl      # expect 0

# 3. Positive controls (pattern-shape proof — measured 2026-09-22, both one)
grep -c '^| question-channel |' AGENTS.md                                 # expect 1
grep -c '^| `moai init' internal/template/templates/AGENTS.md.tmpl        # expect 1
```

## §D Constraints (DO NOT VIOLATE)

- PRESERVE: `worktree-integration.md`, `session-handoff-examples.md` (t1072), all Go source,
  every launcher diagnostic string and threshold.
- Forbidden: `git add -A` (stage by explicit pathspec only); editing any file outside
  `AGENTS.md`, `internal/template/templates/AGENTS.md.tmpl`, and this SPEC's own directory;
  re-running the *creation* verb `agents-emit` by hand (C2→C3 emission is make-driven).
- Commit trailer: `Authored-By-Agent: manager-spec` + `Card: t1071` + `🗿 MoAI`, contiguous, in
  that order. Do NOT push.

## §E Self-Verification deliverables (run-phase)

Each item reports command + verbatim output + tree SHA, per the attribution discipline:

- **E1** AC matrix PASS/FAIL from acceptance.md §D.
- **E2** Guard triple at exactly `1` each (the same three grep patterns as §C.2, re-run
  post-edit) and both positive controls still `1`.
- **E3** `make build` exit 0 (regeneration discharged; `agents-emit-check` inside it green).
- **E4** `go run ./cmd/moai spec lint SPEC-AGENTS-WORKTREE-ROW-001` exit 0 with REQ collection
  firing (nonzero REQ-attributed findings or clean, as measured — recorded verbatim).
- **E5** Working tree after commit: `git status --porcelain` clean except gitignored residue.
- **E6** Scope diff: `git diff --stat <base>..HEAD` touching only the two contract files plus
  this SPEC directory.

## §F Milestones

- **M1 (Priority High)** — root `AGENTS.md` capability table: append the `worktree-entry` row
  after the `design-sync` row (`:25`), keeping the three-column form. Row wording per REQ-AWR-001
  (entry via `moai codex -w` stated, creation-impossible limit stated, one line).
- **M2 (Priority High)** — template mirror: (a) the same row appended to its capability table
  after the table's last row (`:29` — the mirror's table has seven rows, so appending at the
  end is the position its structure wants; a `:25` insertion would land mid-table); (b) a
  `moai codex` verb row appended to `## 11. moai CLI Verbs` (`:303`), wording per REQ-AWR-003.
  The two mirror edits are one unit — same file, same milestone.
- **M3 (Priority Medium)** — `make build` (regeneration per REQ-AWR-005; Template-First cycle),
  then the E2 guard batch and E4 lint, all as one read-only verification batch.
- **M4 (Priority Medium)** — progress.md plan-phase evidence section populated (coordinates,
  guard outputs, lint provenance); single commit on `WT-cross-harness-row` with the card trailers;
  no push.

Dependency order: M1 ∥ M2 are independent file edits; M3 depends on both; M4 depends on M3.

## §G Anti-Patterns to avoid here

- Do not "fix" the mirror's Verbs table to match the root by *adding a Verbs table to the root* —
  C7 shows the root intentionally carries none.
- Do not reword the existing three rows while adding the fourth (scope discipline; the diff
  should read as one added row per table plus one added verb row).
- Do not cite the dispatch's `:16-17` / `:293-302` coordinates in any artifact without the
  re-measured correction attached (C4/C6).

## §H Cross-references

- `SPEC-CODEX-LAUNCHER-001` — owns the launcher surface this card documents (REQ-CL-* in the
  launcher file header).
- Card t1072 — sibling doc surfaces (`worktree-integration.md`,
  `session-handoff-examples.md`).
- `.moai/docs/template-internal-isolation-doctrine.md` §25 — the neutrality classes the new
  rows must respect (no internal SPEC ids, no commit SHAs in the distributed rows).
