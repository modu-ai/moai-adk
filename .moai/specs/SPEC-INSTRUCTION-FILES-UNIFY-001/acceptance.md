# SPEC-INSTRUCTION-FILES-UNIFY-001 — acceptance criteria

Every criterion names a command and the output that decides it. Where a criterion touches a
deployed file it names **both mirrors** — a grep proving one mirror clean establishes
nothing about the other, and for `AGENTS.md` the two mirrors have different filenames
(design.md §D.2), which is itself a way the one-mirror mistake gets made.

Where a failure mode is silent (a skipped import, a truncated tail), the criterion asserts
a **positive indicator**. The absence of an error is never the passing condition.

## §D AC Matrix

### D.1 File structure

**AC-IFU-001** — Given the template tree, When
`ls internal/template/templates/AGENTS.local.md` and
`grep -c '^/AGENTS\.local\.md$' .gitignore internal/template/templates/.gitignore` run,
Then the first exits non-zero with "No such file or directory" and the second reports `1`
for each path. (REQ-IFU-003)

**AC-IFU-002** — Given the deployed `CLAUDE.md` in both mirrors, When
`grep -n '^@AGENTS\.md$\|^@AGENTS\.local\.md$' CLAUDE.md internal/template/templates/CLAUDE.md`
runs, Then each file reports exactly two matches and the `@AGENTS.md` line number is lower
than the `@AGENTS.local.md` line number. (REQ-IFU-002)

**AC-IFU-003** — Given the deployed contract in both mirrors, When
`grep -c '^@' AGENTS.md internal/template/templates/AGENTS.md.tmpl` runs, Then each reports
`0` — the neutral contract imports nothing, and in particular imports no local file into
Codex's discovered chain. (REQ-IFU-016)

### D.2 The `.tmpl` invariant

**AC-IFU-004** — Given the template tree, When the **Codex-discovered filename set** is
scanned recursively:

```
find internal/template/templates -type f \
  \( -name 'AGENTS.md' -o -name 'AGENTS.override.md' -o -name 'CLAUDE.local.md' \) -print
```

Then it prints **nothing** (empty output, exit `0`); and when `ls
internal/template/templates/AGENTS.md.tmpl` runs, Then it exits `0`; and when the
nested-chain guard test runs (`go test ./internal/config/ -run TestContractByte -v`), Then
it exits `0`.

[HARD] The assertion is on the **filename set**, recursively, not on a path the criterion
already expects to exist. Merely *using* `…/AGENTS.md.tmpl` as a measurement location in
other criteria asserts nothing: were the mirror renamed back to `AGENTS.md`, those criteria
would keep passing against the renamed file — a guard repaired into vacuity. `-maxdepth 1`
is likewise insufficient, because Codex walks the chain and a discovered filename in any
nested template directory is in scope.

The name list is the set Codex keys on. It is a closed list in this criterion by
construction; if `fallback_filenames` is ever configured with additional names, this
criterion's list is extended in the same change.

Measured failure this prevents (card t925, commit `703598937`): the root contract merged
with the mirror, the mirror's last section dropped entirely, the preceding section cut mid
table row — no warning, exit `0`, stderr empty. (REQ-IFU-024)

### D.3 Byte ceilings — two distinct limits

**AC-IFU-005** — Per-file ceiling. Given each deployed contract document, When
`wc -c < AGENTS.md` and `wc -c < internal/template/templates/AGENTS.md.tmpl` run, Then each
value is `<= 24576` (`CodexContractByteCeiling`,
`internal/config/token_budget_guard.go:101`). (REQ-IFU-019)

**AC-IFU-006** — Nested-sum budget. Given the repository, When the nested-chain guard is
run (`go test ./internal/config/ -run TestNestedChainBudget -v`), Then it exits `0` with
every reported chain sum `<= 32768`. The guard names no path and walks for the filename
Codex keys on, so a new instruction file anywhere in the tree is in scope. Raising
`project_doc_max_bytes` is not an accepted remedy (REQ-IFU-025). (REQ-IFU-025)

**AC-IFU-007** — Given this repository's migrated local file, When `wc -m < AGENTS.local.md`
runs, Then the value is `< 40000`. (REQ-IFU-021)

### D.4 Mirror reconciliation

**AC-IFU-008** — Given both mirrors, When
`diff <(grep '^## ' AGENTS.md) <(grep '^## ' internal/template/templates/AGENTS.md.tmpl)`
runs, Then it exits `0` with empty output. This compares the **section set** only: the two
files diverge by 46 intentional lines, so a content-level diff would fail by design.
(REQ-IFU-018)

**AC-IFU-009** — Given the reconciled contract, When
`grep -in 'personal.*~/\.codex/AGENTS\.md.*consumed\|narrowing what the project' AGENTS.md internal/template/templates/AGENTS.md.tmpl`
runs, Then it reports no matches; and when `grep -c 'truncat' AGENTS.md` runs, Then it
reports `>= 1`. (REQ-IFU-017)

### D.5 Codex read order

**AC-IFU-010** — Given a fixture project carrying both `AGENTS.local.md` (sentinel `ALPHA`)
and `CLAUDE.local.md` (sentinel `BETA`), When the launcher assembles
`developer_instructions`, Then the assembled value contains `ALPHA` and does not contain
`BETA`. Verified by `go test ./internal/cli/ -run TestCodexLocalInstructions -v`, exit `0`.
(REQ-IFU-006, REQ-IFU-010)

**AC-IFU-011** — Given a fixture project carrying only `CLAUDE.local.md`, When the launcher
assembles `developer_instructions`, Then the assembled value contains that file's content
AND a deprecation advisory naming `moai migrate local-instructions`, AND the provenance
preamble contains the literal string `CLAUDE.local.md` rather than a substituted name. Same
test target, exit `0`. (REQ-IFU-007, REQ-IFU-008)

> **Recorded debt — this criterion decides two distinct outcomes under one verdict.** A
> failure does not say which of them occurred: (a) the deprecation advisory is missing, so
> the user is never told to migrate; or (b) the provenance preamble names the wrong file.
> These are different defects at different severities. **A wrong preamble filename makes
> the fallback invisible in exactly the situation the advisory exists to surface, and that
> is the quieter and more dangerous of the two this criterion hides.**
>
> Were the Tier L criterion ceiling not binding, this would split into one criterion for
> fallback content + advisory (REQ-IFU-007) and one for the preamble literal filename
> (REQ-IFU-008) — one per requirement, as the traceability table would prefer. The merge is
> a budget compromise at 25/25, not tidiness. Recorded here so the split decision is made by
> whoever owns the scope rather than absorbed silently.

**AC-IFU-012** — Given the updated link test, When
`go test ./internal/cli/ -run TestCodexContractLink -v` runs, Then it exits `0` with the
assertions inverted to: executing imports of `AGENTS.local.md` in `CLAUDE.md` = 1, and in
`AGENTS.md` = 0. Both halves are asserted; dropping the `AGENTS.md` half would permit the
neutral contract to pull a local file into the discovered chain. (REQ-IFU-002, REQ-IFU-016)

### D.6 Migration

**AC-IFU-013** — Given a fixture project with `CLAUDE.local.md` present and
`AGENTS.local.md` absent, When `moai migrate local-instructions` runs, Then it exits `0`,
`AGENTS.local.md` exists with byte-identical content to the original, `CLAUDE.local.md` no
longer exists at the project root, and a copy exists under the backup directory.
(REQ-IFU-009, REQ-IFU-010)

**AC-IFU-014** — Given a fixture project with BOTH local files present, When
`moai migrate local-instructions` runs, Then it exits non-zero without modifying either
file, naming the coexistence as the refusal reason. (REQ-IFU-010)

**AC-IFU-015** — Given a fixture project with `CLAUDE.local.md` present, When `moai update`
runs and then `moai doctor` runs, Then both exit `0`, `CLAUDE.local.md` is byte-identical
across the whole sequence (sha256 captured before and after), and each command's stdout
contains the migration advisory. (REQ-IFU-011, REQ-IFU-012)

> **Recorded debt — this criterion decides three distinct outcomes under one verdict.** A
> failure does not say which: (a) `moai update` mutated the local file; (b) `moai update`
> lacks the advisory; (c) `moai doctor` lacks the advisory. (a) is a data-integrity failure
> — the user's own instruction file was altered by a command that must not touch it —
> while (b) and (c) are UX failures. Folding a data-integrity outcome in with two UX
> outcomes is the part worth objecting to.
>
> Were the ceiling not binding, this would split into the sha256 immutability assertion
> (REQ-IFU-011) and the advisory-parity assertion across both commands (REQ-IFU-012).
> Recorded as budget compromise, not tidiness.

### D.7 Guard and learner surfaces

**AC-IFU-016** — Given the frozen set, When
`grep -n 'frozenInstructionFiles = ' internal/hook/pre_tool.go` runs, Then the line lists
all four of `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md`; and
`go test ./internal/hook/ -run TestHarnessFrozen` exits `0` with a case per new entry.
(REQ-IFU-013)

**AC-IFU-017** — Given the curator dispatch table, When
`go test ./internal/harness/curator/ -run TestDispatch` runs, Then it exits `0` and the
Tier-3 entry's `Path` is `AGENTS.local.md`. (REQ-IFU-014)

**AC-IFU-018** — Given the neutrality guard, When `go test ./internal/template/ -run TestNeutrality`
and `.github/workflows/template-neutrality-check.yaml` run over a template fixture containing
the literal `AGENTS.local.md`, Then both pass; and over a fixture containing
`CLAUDE.local.md`, Then the guard fails naming C5. (REQ-IFU-015)

### D.8 Import resolution — the silent-skip gate

**AC-IFU-019** — Given a fixture project whose `AGENTS.local.md` carries a unique sentinel
and whose `CLAUDE.md` carries `@AGENTS.local.md`, When a headless `claude -p` session runs
in that project asking for the sentinel, Then the response contains it. This asserts the
import RESOLVED; a grep for the directive line does not, because M0-1 establishes an
unresolved import is silently skipped with exit `0` and empty stderr. (REQ-IFU-023)

**AC-IFU-020** — Given a fixture project with NO `AGENTS.local.md`, When `moai init` runs
and a headless `claude -p` session then starts, Then `moai init` exits `0`, the deployed
`CLAUDE.md` still contains the literal line `@AGENTS.local.md`, and the session exits `0`.
(REQ-IFU-004)

### D.9 Run-phase measurements

**AC-IFU-021** — Real-worktree ancestor discovery. Given a fixture repository with
`AGENTS.local.md` and `CLAUDE.local.md` at its root, each carrying a distinct sentinel, and
a **real linked worktree** created by `git worktree add` (so `.git` in that tree is a file,
not a directory), When a headless `claude -p` session runs with cwd inside that worktree
and is asked for both sentinels, Then the observed values are recorded in `progress.md`
§E.2 with the command and verbatim output. **This criterion passes on a recorded result,
not on a particular result** — it exists to settle the question, and a negative outcome
discharges it exactly as a positive one does. The M0 P7 line-3 observation is a
synthetic-fixture observation and is not evidence for either answer. The outcome routes per
design.md §A.5. (REQ-IFU-023; gates the worktree leg)

**AC-IFU-022** — Codex discovery of `AGENTS.local.md`, with truncation detection. This is
the **Codex filename-discovery** question, distinct from AC-IFU-021's Claude-side ancestor
walk; neither criterion covers the other.

Given a fixture project carrying `AGENTS.md`, `CLAUDE.md`, and `AGENTS.local.md`, where
`AGENTS.local.md` holds `LOCAL_HEAD` at its first line and `LOCAL_TAIL` as its **final
line**, and `AGENTS.md` likewise ends with `CONTRACT_TAIL`, When the assembled
model-visible prompt input is rendered:

```
cd <fixture> && codex debug prompt-input
```

Then the rendered JSON is read for all four tokens. **Verb availability was verified in
this run against `codex-cli 0.157.0`**: `codex debug --help` lists `prompt-input` —
"Render the model-visible prompt input list as JSON". The verb is inherited from the t925
commit message but was not taken on trust; the criterion records the version it was
confirmed against, and a run under a different codex-cli version re-confirms before
relying on it.

Rendering the prompt input is preferred over prompting a live session: it reports what was
assembled rather than what a model chose to repeat, so a model declining to echo a token
cannot be mistaken for the token's absence.

The decision rule:

- `LOCAL_HEAD` absent → Codex does not discover `AGENTS.local.md`; the design's budget
  arithmetic holds.
- `LOCAL_HEAD` present → Codex DOES discover it, the content is counted twice, and the
  budget arithmetic is wrong; the finding blocks and is escalated.
- `CONTRACT_TAIL` or `LOCAL_TAIL` absent while its head token is present → **silent tail
  truncation occurred**, and the run is a FAILURE regardless of exit code.

A completed run proves nothing here: the documented failure is a silent tail cut with exit
`0` and empty stderr, so the tail sentinels are the only positive indicator. This is the
same shape AC-IFU-019 uses for import resolution, anchored at the tail instead of the body.
Command, codex-cli version, and verbatim output recorded in `progress.md` §E.2 whichever
way it comes out. (REQ-IFU-025)

### D.10 Documentation, this repository, and CI

**AC-IFU-023** — Given the docs-site, When `grep -lc 'AGENTS.local.md'` is run against each
of the six named pages in `docs-site/content/{ko,en,ja,zh}/`, Then all four locales report a
match for every page, and the per-page `## ` section count is equal across the four
locales. (REQ-IFU-020)

**AC-IFU-024** — Given the migrated repository-local file, When its §0 is read, Then it
names `AGENTS.local.md` as the file whose canonical copy it discriminates, and
`grep -c 'CLAUDE.local.md' AGENTS.local.md` reports only historical-reference occurrences,
each within a sentence marking it as a retired filename. (REQ-IFU-022)

**AC-IFU-025** — Given the whole change, When `make build && go test ./...` runs in CI on
the PR head, Then it exits `0`. (all)

---

## §D.1 Severity

Blocking: AC-IFU-002, -004, -005, -006, -008, -010, -012, -013, -014, -016, -019, -021,
-022, -025. All others are must-fix before close but do not block a milestone boundary.

AC-IFU-004 and AC-IFU-006 are blocking together because they guard the same measured
failure from two sides — the filename that causes discovery, and the byte sum discovery
produces.

## §D.2 Traceability

REQ→AC: 001→015, 002→002/012, 003→001, 004→020, 005→019, 006→010, 007→011, 008→011,
009→013, 010→013/014, 011→015, 012→015, 013→016, 014→017, 015→018, 016→003/012, 017→009,
018→008, 019→005, 020→023, 021→007, 022→024, 023→019/021, 024→004, 025→006/022.
Every REQ-IFU-001..025 is covered.

## §D.3 Definition of Done

All 25 criteria pass; every deployed-file criterion verified separately against both
mirrors; CI green on the PR head; and AC-IFU-021 and AC-IFU-022 recorded in `progress.md`
§E.2 with command and verbatim output **whichever way they come out**.

## §D.4 Forward-looking checks

The worktree leg is deliberately NOT asserted by any criterion here. AC-IFU-021 settles the
question; it does not presuppose the answer.

[HARD] The deferral in this section is scoped to the **Claude-side ancestor walk** only. It
does NOT cover the Codex filename-discovery question, which is not deferred: that is
AC-IFU-022, a mandatory run-phase measurement with its own command and decision rule. The
two are different mechanisms in different harnesses and neither answer implies the other.

If ancestor discovery is confirmed for `AGENTS.local.md` in a real worktree, **the finding
and its criterion TRANSFER to card t1219 or a successor SPEC — they do not land in this
criterion set.** Asserting the leg works here, while t1219 holds an open defect against the
same mechanism, would assert half a behaviour (design.md §A.3, §A.5). This is a scope
boundary rather than a capacity workaround: it would hold at any criterion count. If it is not confirmed, the degradation to Option 1 is recorded as a
known limitation and is not retried.
