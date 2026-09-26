# SPEC-INSTRUCTION-FILES-UNIFY-001 — acceptance criteria

Every criterion names a command and the output that decides it. Where a criterion touches a
deployed file it names **both mirrors** — a grep proving one mirror clean establishes
nothing about the other, and for `AGENTS.md` the two mirrors have different filenames
(design.md §D.2), which is itself a way the one-mirror mistake gets made.

Where a failure mode is silent (a skipped import, a truncated tail), the criterion asserts
a **positive indicator**. The absence of an error is never the passing condition.

> **[HARD] The positive-indicator rule binds every `go test` invocation here.** `go test -run`
> takes an unanchored regexp and exits `0` when it matches nothing, printing `no tests to run`
> — so naming a pattern that matches no test produces a criterion that cannot fail. Every
> criterion below therefore (1) names the test's **symbol** as it is declared in source, (2)
> runs with `-v`, and (3) asserts the literal `--- PASS: <TestName>` appears in the output.
> Asserting exit `0` alone is not sufficient and is not accepted. The plan-audit of commit
> `1140bcd1d` found five criteria failing exactly this way; three of them named a guard that
> already existed under a different symbol.
>
> **Symbols, not line numbers.** Test and source locations are cited by symbol. Any line
> number that appears is illustrative of where the symbol stood on 2026-09-26 against base
> develop `553e224f3`, and is not part of the assertion — `origin/develop` is ~93 commits
> ahead of this branch and has already moved one of them (card t1224 moved
> `frozenInstructionFiles` by two lines while leaving the symbol intact).

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

**AC-IFU-026** — Deployment of the contract itself. Given a fixture project created by
`moai init`, When `ls AGENTS.md` runs in it and then `sha256sum AGENTS.md` is captured, and
When the file is then overwritten with a sentinel line and `moai update` is run and the hash
re-captured, Then `AGENTS.md` exists after `moai init`, the post-`update` hash equals the
original, and `grep -c '<sentinel>' AGENTS.md` reports `0` — the file is deployed and is
replaced rather than preserved, which is what makes it a non-user-edited surface. Both the
template source (`internal/template/templates/AGENTS.md.tmpl`) and the deployed result are
named: the deployer strips the `.tmpl` suffix, so a template-side `ls` asserts nothing about
what the user receives. (REQ-IFU-001)

> Added at v0.3.0. The plan-audit of `1140bcd1d` found `REQ-IFU-001` cited by no criterion
> while §D.2 claimed `001→015`; `AC-IFU-015` asserted nothing about `AGENTS.md` being
> deployed, and it has since moved to the sibling SPEC.

### D.2 The `.tmpl` invariant

**AC-IFU-004** — Given the template tree, When the **Codex-discovered filename set** is
scanned recursively:

```
find internal/template/templates -type f \
  \( -name 'AGENTS.md' -o -name 'AGENTS.override.md' -o -name 'CLAUDE.local.md' \) -print
```

Then it prints **nothing** (empty output, exit `0`); and when `ls
internal/template/templates/AGENTS.md.tmpl` runs, Then it exits `0`; and when the per-file
ceiling guard runs

```
go test ./internal/config/ -run '^TestCodexContractByteCeiling$' -v
```

Then the output contains `--- PASS: TestCodexContractByteCeiling` and does not contain
`no tests to run`. The symbol is declared in `internal/config/token_budget_guard_test.go`.

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

> v0.3.0 repair: the third clause named `TestContractByte`, which matches no test in that
> package, so it exited `0` against an unimplemented tree. The real symbol is
> `TestCodexContractByteCeiling`.

### D.3 Byte ceilings — two distinct limits

**AC-IFU-005** — Per-file ceiling. Given each deployed contract document, When
`wc -c < AGENTS.md` and `wc -c < internal/template/templates/AGENTS.md.tmpl` run, Then each
value is `<= 24576` (`CodexContractByteCeiling`, declared in
`internal/config/token_budget_guard.go`). (REQ-IFU-019)

**AC-IFU-006** — Nested-sum budget. Given the repository, When the nested-chain guard is run

```
go test ./internal/config/ -run '^TestCodexNestedTemplateDiscoveryBudget$' -v
```

Then the output contains `--- PASS: TestCodexNestedTemplateDiscoveryBudget`, does not contain
`no tests to run`, and every reported chain sum is `<= 32768`. The symbol is declared in
`internal/config/token_budget_guard_test.go`. The guard names no path and walks for the
filename Codex keys on, so a new instruction file anywhere in the tree is in scope. Raising
`project_doc_max_bytes` is not an accepted remedy (REQ-IFU-025). (REQ-IFU-025)

> v0.3.0 repair, and the priority one. This criterion named `TestNestedChainBudget` — no such
> test exists, so the **sole** mechanical enforcement of REQ-IFU-025 exited `0` and could not
> fail. REQ-IFU-025's ceiling is the exact limit whose breach produced card t925's measured
> silent mid-table-row truncation. The real symbol is
> `TestCodexNestedTemplateDiscoveryBudget`.

### D.4 Mirror reconciliation

**AC-IFU-008** — Given both mirrors, When
`diff <(grep '^## ' AGENTS.md) <(grep '^## ' internal/template/templates/AGENTS.md.tmpl)`
runs, Then it exits `0` with empty output. This compares the **section set** only: the two
files diverge intentionally (spec.md §C.4 — measured 2026-09-26 against `553e224f3` as 57
template-only and 17 root-only lines), so a content-level diff would fail by design.
(REQ-IFU-018)

**AC-IFU-009** — Given the reconciled contract, When
`grep -in 'personal.*~/\.codex/AGENTS\.md.*consumed\|narrowing what the project' AGENTS.md internal/template/templates/AGENTS.md.tmpl`
runs, Then it reports no matches; and when `grep -c 'truncat' AGENTS.md` runs, Then it
reports `>= 1`. (REQ-IFU-017)

### D.5 Codex read order

**AC-IFU-010** — Given a fixture project carrying both `AGENTS.local.md` (sentinel `ALPHA`)
and `CLAUDE.local.md` (sentinel `BETA`), When the launcher assembles
`developer_instructions`, Then the assembled value contains `ALPHA` and does not contain
`BETA`. Verified by
`go test ./internal/cli/ -run '^TestCodexLocalInstructions' -v`, whose output must contain
`--- PASS: TestCodexLocalInstructions` and must not contain `no tests to run`.
(REQ-IFU-006)

> v0.3.0: this criterion previously also cited the no-coexistence requirement. That requirement
> moved to `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`, where `AC-IFU-013` and `AC-IFU-014` assert it
> directly, so the citation went with it; the read-order assertion here is `REQ-IFU-006`'s
> alone. The assertion itself is unchanged.
>
> The sibling requirement is deliberately named in prose rather than by its id token here: the
> §D.2 verification command greps this file for `REQ-IFU-` tokens, so a bare id in a note about
> a *removed* citation would make that command report a phantom coverage row — the command has
> to stay true to be worth running.

**AC-IFU-012** — Given the updated link test, When
`go test ./internal/cli/ -run '^TestCodexContractLink' -v` runs, Then its output contains
`--- PASS: TestCodexContractLink`, does not contain `no tests to run`, and the assertions are
inverted to: executing imports of `AGENTS.local.md` in `CLAUDE.md` = 1, and in `AGENTS.md` =
0. Both halves are asserted; dropping the `AGENTS.md` half would permit the neutral contract
to pull a local file into the discovered chain. (REQ-IFU-002, REQ-IFU-016)

### D.6 Guard and learner surfaces

**AC-IFU-016** — Given the frozen set, When
`grep -n 'frozenInstructionFiles = ' internal/hook/pre_tool.go` runs, Then the line lists
all four of `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md`; and When

```
go test ./internal/hook/ -run '^TestFrozenInstructionFiles' -v
```

runs, Then its output contains `--- PASS: TestFrozenInstructionFiles` and does not contain
`no tests to run`, with one sub-case per entry in the set. (REQ-IFU-013)

> **[HARD] This test does not exist yet, and the criterion is written to require its
> creation.** Verified 2026-09-26: `grep -rn 'HARNESS_FROZEN\|frozenInstruction'
> internal/hook/*_test.go` returns no match — nothing in `internal/hook` currently guards the
> `frozenInstructionFiles` symbol. (`TestAnalyzeSession_SafetyFiltersFrozenFiles` in
> `reflective_write_test.go` filters learner *proposals*; it does not read this set.) The
> v0.2.0 draft named `TestHarnessFrozen` as though it existed, which made the criterion exit
> `0` against the current tree. The run phase creates `TestFrozenInstructionFiles`; the
> `--- PASS:` assertion is what distinguishes "created and passing" from "still absent".

**AC-IFU-017** — Given the curator dispatch table, When

```
go test ./internal/harness/curator/ -run '^TestSurfaceForTier_Tier3$|^TestPrepareTierDispatch_Tier3$' -v
```

runs, Then its output contains **both** `--- PASS: TestSurfaceForTier_Tier3` and
`--- PASS: TestPrepareTierDispatch_Tier3`, does not contain `no tests to run`, and the
Tier-3 entry's `Path` is `AGENTS.local.md`. Both symbols are declared in
`internal/harness/curator/dispatch_test.go`. (REQ-IFU-014)

> v0.3.0 repair: the criterion named `TestDispatch`, which matches no test in that package.

**AC-IFU-018** — Given the neutrality guard, When

```
go test ./internal/template/agentemit/ -run '^TestNeutralityByInheritance$' -v
```

and `.github/workflows/template-neutrality-check.yaml` run over a template fixture containing
the literal `AGENTS.local.md`, Then the test output contains
`--- PASS: TestNeutralityByInheritance`, does not contain `no tests to run`, and the workflow
passes; and over a fixture containing `CLAUDE.local.md`, Then the guard fails naming C5.
(REQ-IFU-015)

> v0.3.0 repair, two-part. The package was `./internal/template/` without `...`, which does
> not walk sub-packages, so the guard was never reached; and the symbol is
> `TestNeutralityByInheritance`, declared in `internal/template/agentemit/golden_test.go`.
> `-run TestNeutrality` would have matched it by substring had the package been right — both
> halves had to be wrong for this to pass vacuously, and both were.

### D.7 Import resolution — the silent-skip gate

**AC-IFU-019** — Given a fixture project whose `AGENTS.local.md` carries a unique sentinel
and whose `CLAUDE.md` carries `@AGENTS.local.md`, When a headless `claude -p` session runs
in that project asking for the sentinel, Then the response contains it. This asserts the
import RESOLVED; a grep for the directive line does not, because M0-1 establishes an
unresolved import is silently skipped with exit `0` and empty stderr. (REQ-IFU-023)

**AC-IFU-020** — Given a fixture project with NO `AGENTS.local.md`, When `moai init` runs
and a headless `claude -p` session then starts, Then `moai init` exits `0`, the deployed
`CLAUDE.md` still contains the literal line `@AGENTS.local.md`, and the session exits `0`.
(REQ-IFU-004)

**AC-IFU-027** — Launcher invariance. Given the fixture of `AC-IFU-019` (an
`AGENTS.local.md` carrying a unique sentinel, reached only through `CLAUDE.md`'s
`@AGENTS.local.md` import), When the same sentinel request is issued under **each** of
`moai cc`, bare `claude`, and `moai glm` in headless mode, Then every one of the three
responses contains the sentinel; and When
`git diff --stat` is taken over the launcher sources for the whole change
(`internal/cli/cc*.go`, `internal/cli/glm*.go`), Then no launcher passes an
instruction-file path or an instruction argument of its own — the import path is the only
mechanism. Three separate invocations with three separate results; one launcher passing
establishes nothing about the other two, which is the same two-mirror discipline applied to
the launcher axis. (REQ-IFU-005)

> Added at v0.3.0. The plan-audit found `REQ-IFU-005` cited by no criterion while §D.2
> claimed `005→019`; `AC-IFU-019` asserts one sentinel through one import and says nothing
> about launcher invariance, which is the substance of the requirement.

### D.8 Run-phase measurements

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

Given a fixture project carrying `AGENTS.md`, `CLAUDE.md`, and `AGENTS.local.md`, where the
**four** sentinels are each defined:

| Token | Placement |
|---|---|
| `CONTRACT_HEAD` | the **first line** of the fixture's `AGENTS.md` |
| `CONTRACT_TAIL` | the **final line** of the fixture's `AGENTS.md` |
| `LOCAL_HEAD` | the **first line** of `AGENTS.local.md` |
| `LOCAL_TAIL` | the **final line** of `AGENTS.local.md` |

When the assembled model-visible prompt input is rendered:

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

The decision rule, **positive control first**:

- `CONTRACT_HEAD` absent → **the run is INCONCLUSIVE and no other branch may be read.**
  `AGENTS.md` is discovered by filename, so a correct render always carries it; its absence
  means the render itself failed (wrong cwd, a fixture whose `AGENTS.md` was never written, a
  codex-cli whose output shape changed). Re-run after fixing the fixture; do not record a
  conclusion.
- `CONTRACT_HEAD` present and `LOCAL_HEAD` absent → Codex does not discover
  `AGENTS.local.md`; the design's budget arithmetic holds.
- `CONTRACT_HEAD` present and `LOCAL_HEAD` present → Codex DOES discover it, the content is
  counted twice, and the budget arithmetic is wrong; the finding blocks and is escalated.
- `CONTRACT_TAIL` absent while `CONTRACT_HEAD` is present, or `LOCAL_TAIL` absent while
  `LOCAL_HEAD` is present → **silent tail truncation occurred**, and the run is a FAILURE
  regardless of exit code.

> v0.3.0 repair (D5). The v0.2.0 text said "all four tokens" while defining three, and its
> truncation branch read "its head token is present" — for `CONTRACT_TAIL` that head token
> was the undefined fourth, so the truncation branch was **inoperable for `AGENTS.md`**, the
> one file Codex certainly does discover. Worse, with no positive control a render that
> captured nothing at all resolved to "`LOCAL_HEAD` absent → the budget arithmetic holds" —
> reinstating, one level up, exactly the inference-from-absence the tail sentinels exist to
> eliminate. `CONTRACT_HEAD` is now defined and is the control.

A completed run proves nothing here: the documented failure is a silent tail cut with exit
`0` and empty stderr, so the tail sentinels are the only positive indicator. This is the
same shape AC-IFU-019 uses for import resolution, anchored at the tail instead of the body.
Command, codex-cli version, and verbatim output recorded in `progress.md` §E.2 whichever
way it comes out. (REQ-IFU-025)

### D.9 Whole-change CI

**AC-IFU-025** — Given the whole change, When `make build && go test ./...` runs in CI on
the PR head, Then it exits `0`; and Then the run includes
`--- PASS: TestAlwaysLoadedTokenBudget` (declared in
`internal/config/token_budget_guard_test.go`). (all)

> The always-loaded clause is added at v0.3.0 and costs no criterion slot. Root `AGENTS.md`
> is itself inside the always-loaded surface that guard measures, and this SPEC moves that
> total from two directions — REQ-IFU-002 thins `CLAUDE.md`, REQ-IFU-018 may fold template-only
> sections into `AGENTS.md`. Both named ceilings (`AC-IFU-005` per-file 24,576, `AC-IFU-006`
> nested-sum 32,768) are different limits with different owners, so neither would catch it.
> Card t1175 is concurrently retuning that same budget; without this clause an
> always-loaded overrun would arrive during M4 with no criterion and no owner.

---

## §D.1 Severity

Blocking: AC-IFU-002, -004, -005, -006, -008, -010, -012, -016, -019, -021, -022, -025,
-026, -027. All others are must-fix before close but do not block a milestone boundary.

AC-IFU-004 and AC-IFU-006 are blocking together because they guard the same measured
failure from two sides — the filename that causes discovery, and the byte sum discovery
produces.

`AC-IFU-026` and `AC-IFU-027` are blocking because each is the only criterion covering its
requirement; that is what their absence cost the v0.2.0 draft.

## §D.2 Traceability

[HARD] **This table is derived from the citation line in each criterion's own body, not
asserted independently of it.** A mapping appears here only if the cited criterion names that
requirement in its own text. The v0.2.0 table claimed two mappings (`001→015`, `005→019`) that
the cited criteria contradicted, and closed with a coverage claim nothing had verified — an
unobserved coverage claim under `verification-claim-integrity.md` §1.1 surface 3.

| REQ | Covered by |
|---|---|
| REQ-IFU-001 | AC-IFU-026 |
| REQ-IFU-002 | AC-IFU-002, AC-IFU-012 |
| REQ-IFU-003 | AC-IFU-001 |
| REQ-IFU-004 | AC-IFU-020 |
| REQ-IFU-005 | AC-IFU-027 |
| REQ-IFU-006 | AC-IFU-010 |
| REQ-IFU-013 | AC-IFU-016 |
| REQ-IFU-014 | AC-IFU-017 |
| REQ-IFU-015 | AC-IFU-018 |
| REQ-IFU-016 | AC-IFU-003, AC-IFU-012 |
| REQ-IFU-017 | AC-IFU-009 |
| REQ-IFU-018 | AC-IFU-008 |
| REQ-IFU-019 | AC-IFU-005 |
| REQ-IFU-023 | AC-IFU-019, AC-IFU-021 |
| REQ-IFU-024 | AC-IFU-004 |
| REQ-IFU-025 | AC-IFU-006, AC-IFU-022 |

Verification command, re-run at close rather than remembered — the union of the ids cited in
criterion bodies must equal the union of the ids declared in `spec.md`:

```
diff <(grep -o 'REQ-IFU-[0-9]\{3\}' acceptance.md | sort -u) \
     <(grep -o '^- \*\*REQ-IFU-[0-9]\{3\}' spec.md | grep -o 'REQ-IFU-[0-9]\{3\}' | sort -u)
```

Empty output is the passing condition. The nine requirements transferred to
`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` are absent from both sides, so their absence is not a
gap here.

## §D.3 Definition of Done

All 20 criteria pass; every deployed-file criterion verified separately against both
mirrors; every test-invoking criterion's `--- PASS: <TestName>` line captured rather than
its exit code alone; the §D.2 verification command run and its empty output recorded; CI
green on the PR head; and AC-IFU-021 and AC-IFU-022 recorded in `progress.md` §E.2 with
command and verbatim output **whichever way they come out**.

Conditional item — **if `AC-IFU-021` confirms ancestor discovery for `AGENTS.local.md` in a
real linked worktree**, the finding and its verbatim evidence are handed to card **t1219**,
and the handoff is recorded in `progress.md` §E.2 naming what was transferred and where.
Without this item the obligation rested on prose in three files and on the run-phase agent
remembering it (plan-audit D7); design.md §A.5's confirmed branch requires the transfer, so
Done requires evidence the transfer happened.

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
boundary rather than a capacity workaround: it would hold at any criterion count. If it is
not confirmed, the degradation to Option 1 is recorded as a known limitation and is not
retried.
