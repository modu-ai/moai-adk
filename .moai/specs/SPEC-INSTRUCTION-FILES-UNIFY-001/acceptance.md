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
> **[HARD] Anchor both ends, on the pattern *and* on the asserted line.** The rule above is
> necessary and was not sufficient: a pattern anchored only at the head (`'^TestFoo'`) matches
> every sibling whose name merely *starts* with `TestFoo`, and the asserted string
> `--- PASS: TestFoo` is then satisfied as a **substring** of `--- PASS: TestFoo_Bar (0.00s)`.
> A criterion built that way passes against an unimplemented tree, which is the same vacuity
> in a new shape. So every `-run` pattern here is anchored at **both** ends (`'^TestFoo$'`, or
> an alternation of both-end-anchored symbols), and every asserted `--- PASS:` line carries
> the delimiter Go prints after the test name — a single space, written as
> `` `--- PASS: TestFoo ` `` — so no longer symbol can satisfy it.
>
> **[HARD] The class is judged by a `moai spec lint` rule, not by a command written here.**
> The plan-audit of commit `653e5357295cfbb2eb3831a982871b6a9b90a00c` found this defect surviving
> in `AC-IFU-010`, `AC-IFU-012`, `AC-IFU-016` and `AC-IFU-025` here and in `AC-IFU-011 [REF]` of
> the sibling SPEC, *after* the five instances the prior audit enumerated had been repaired — the
> repair reached the instances and not the generator. Two successive attempts to record the
> generator as a prose `grep` pipeline then failed in their own right, and the third audit
> identified the cause as structural rather than as two authoring mistakes: a command living as
> text inside a document cannot record that it ran, cannot carry a positive control, keeps its
> scope inside the command rather than in its declaration, and sits in the same file it inspects.
>
> Pattern judgment therefore leaves prose, and the medium moves to **card t1269** — the
> `VacuousAssertionRule` in `internal/spec/lint_vacuous_assertion.go`, registered in the rule slice
> in `internal/spec/lint.go`, with two-arm fixtures. A `Rule` there receives one `*SPECDoc` per
> `Check`, so the whole-tree scope defect that sank the prose form is structurally impossible,
> and the rule runs in CI via `.github/workflows/spec-lint.yml`. (The prose glob reached every SPEC
> directory in the tree — `ls -d .moai/specs/SPEC-*/ | wc -l` → 938 at time of writing, a count that
> drifts with every SPEC added — where the declaration named two.)
>
> [HARD] **The claim below is forward-looking, and t1269 has not landed.** What this SPEC claims is
> that **its criteria will pass that rule once t1269 lands** — not that they pass a check today. No
> such check runs against this file at present, and reading this block as citing a live one is the
> same unobserved-claim shape the removal of the two prose commands exists to remove. Until t1269
> lands, the anchoring rule above is a `[HARD]` authoring obligation on whoever writes a criterion
> here, enforced by review rather than mechanically.
>
> **Where the removed material went.** The eleven lines that stood here — two `grep` pipelines and
> their explanatory block — are not deleted; they are **transferred to t1269**, which reimplements
> their intent as a rule that can record that it ran, carry a two-arm fixture, declare its scope
> outside the pattern, and live in a file other than the one it inspects. §D.3.1 carries the
> accounting, and the Definition-of-Done item that required the two commands to be re-run at close
> went with them (its obligation is t1269's CI job, not a close step here).
> **The general defect is wider than `go test`.** Any assertion satisfiable by something other
> than the thing under test is vacuous: a `grep -c` whose pattern also matches this SPEC's own
> prose, a count satisfied by an unrelated file, a `--- PASS:` satisfied by a sibling. When
> writing or repairing a criterion, the question is not "does this command pass when the work is
> done" but "can this command pass when it is **not**".
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
Codex's discovered chain. (REQ-IFU-018)

> **[HARD] This is a declared proxy on the neutrality clause, not its assertion.** `REQ-IFU-018`'s
> second clause constrains **clauses**; this criterion counts **imports**. It is retained because an
> import of a harness-specific local file (`@CLAUDE.local.md`) would itself be a harness-restricted
> construction, so a zero import count is a **necessary** condition of clause neutrality — but it is
> not sufficient, and it is already satisfied before any work (measured 2026-09-26 on this commit:
> `0` / `0`). **No criterion here positively asserts clause-level neutrality.** That gap is a named
> debt item in §D.3; it is stated rather than papered over, because the declared-proxy note is what
> keeps the §D.2 row honest and a proxy read as coverage is exactly the defect the retirement of
> `REQ-IFU-016` was meant to close. `REQ-IFU-018`'s section-set clause is asserted by `AC-IFU-008`.

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
> while §D.2 claimed `001→015`; `AC-IFU-015 [REF]` asserted nothing about `AGENTS.md` being
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

Then the output contains `--- PASS: TestCodexContractByteCeiling ` and does not contain
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

Then the output contains `--- PASS: TestCodexNestedTemplateDiscoveryBudget `, does not contain
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
reports `>= 1`; and when `grep -c 'project instruction files only' AGENTS.md` and
`grep -c 'project instruction files only' internal/template/templates/AGENTS.md.tmpl` run as
two separate commands, Then each reports `>= 1` — the replacement statement, written with
that phrase unbroken on one line. This third clause is RED at plan time (t1270, tree
`20c73990d`: both commands print `0` and exit `1`), and the superseded sentence cannot
satisfy it. (REQ-IFU-017)

### D.5 Codex read order

**AC-IFU-010** — Given a fixture project carrying both `AGENTS.local.md` (sentinel `ALPHA`)
and `CLAUDE.local.md` (sentinel `BETA`), When the launcher assembles
`developer_instructions`, Then the assembled value contains **both** sentinels and `ALPHA`
appears **before** `BETA` — as does `AGENTS.local.md` before `CLAUDE.local.md` in the two
provenance preambles. Verified by

```
go test ./internal/cli/ -run '^TestCodexLocalInstructions_AgentsLocalReadFirst$|^TestCodexLocalInstructions_DualFileMatrix$' -v
```

whose output must contain **both** `--- PASS: TestCodexLocalInstructions_AgentsLocalReadFirst `
and `--- PASS: TestCodexLocalInstructions_DualFileMatrix `, and must not contain
`no tests to run`. (REQ-IFU-006)

> **[HARD] v0.3.1 repair — this criterion was vacuous and its assertion contradicted its own
> requirement.** Two separate defects, both found by the plan-audit of `653e53572`:
>
> (a) **Vacuity.** The pattern was `'^TestCodexLocalInstructions'`, head-anchored only, and
> `TestCodexLocalInstructions` does not exist as a symbol. Measured 2026-09-26 against
> `653e53572`: `go test ./internal/cli/ -run '^TestCodexLocalInstructions' -v | grep -c --
> '--- PASS: TestCodexLocalInstructions'` → **74** — the pattern runs 22 pre-existing sibling
> tests and the asserted string is a substring of each of their PASS lines, none of which
> asserts a read order. The criterion passed against a tree whose order is the **opposite** of
> what it requires.
>
> (b) **Wrong assertion.** The criterion required `BETA` to be **absent**, i.e. exclusive
> precedence. `REQ-IFU-006` — its own and only cited requirement — requires the launcher to
> "read `AGENTS.local.md` **ahead of** `CLAUDE.local.md`", and design.md §C states the change is
> "the iteration order". Both files are still read and concatenated; asserting `BETA` absent
> would have failed a correct implementation and demanded a behaviour neither the requirement
> nor the design asks for. The assertion is now ordering, which is what `REQ-IFU-006` says.
>
> **The two named symbols, and why both.** `TestCodexLocalInstructions_AgentsLocalReadFirst`
> **does not exist yet and this criterion requires its creation in M2** (the `AC-IFU-016`
> pattern): it is the test that carries the ALPHA/BETA ordering assertion, and its `--- PASS:`
> line is what distinguishes "written and passing" from "still absent".
> `TestCodexLocalInstructions_DualFileMatrix` **does** exist
> (`internal/cli/codex_local_instructions_test.go`) and today asserts the **old** order —
> it builds its expectation by iterating `[]string{"CLAUDE.local.md", codexLocalInstructionName}`,
> so it passes only while `CLAUDE.local.md` comes first. Naming it here is deliberate: M2 must
> invert that expectation, and without this clause the change could land with a new passing
> test beside an old passing test asserting the opposite. A repository cannot hold both.
> `TestCodexLocalInstructions_LargeBodySlicesAndFreshRead` carries the same order expectation
> and M2 updates it in the same change; it is not named in the assertion because
> `AC-IFU-025`'s CI full-suite run is what catches it, and naming every affected test here
> would duplicate that.
>
> v0.3.0: this criterion previously also cited the no-coexistence requirement. That requirement
> moved to `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`, where `AC-IFU-013 [REF]` and `AC-IFU-014 [REF]` assert it
> directly, so the citation went with it; the read-order assertion here is `REQ-IFU-006`'s
> alone. The assertion itself is unchanged.
>
> The sibling requirement is deliberately named in prose rather than by its id token here: the
> §D.2 verification command greps this file for `REQ-IFU-` tokens, so a bare id in a note about
> a *removed* citation would make that command report a phantom coverage row — the command has
> to stay true to be worth running.

**AC-IFU-012** — Given the updated link test, When

```
go test ./internal/cli/ -run '^TestCodexContractLinkCreation$|^TestCodexContractLink_LocalImportMatrix$' -v
```

runs, Then its output contains **both** `--- PASS: TestCodexContractLinkCreation ` and
`--- PASS: TestCodexContractLink_LocalImportMatrix `, and does not contain `no tests to run`. The
second symbol is the decision rule for the half that requires work: `TestCodexContractLink_LocalImportMatrix`
is a test **M3 creates**, and it asserts both directions in one place — executing imports of
`AGENTS.local.md` in `CLAUDE.md` = 1, and executing imports of `AGENTS.local.md` in `AGENTS.md` = 0.
Dropping the `AGENTS.md` direction would permit the neutral contract to pull a local file into
Codex's discovered chain; dropping the `CLAUDE.md` direction would leave the import unreached.
(REQ-IFU-002)

> **[HARD] v0.3.2 repair — a blocking criterion dischargeable by an already-green command.** The
> prior form named only `TestCodexContractLinkCreation` and then stated in prose that "the test
> asserts both of" the two directions. The first half is already true: measured 2026-09-26 on this
> commit, `go test ./internal/cli/ -run '^TestCodexContractLinkCreation$' -v | grep -c -- '--- PASS: TestCodexContractLinkCreation '`
> → **1**. The second half is a claim about test *content* with no invocation deciding it, and the
> `CLAUDE.md` direction is genuinely absent from the test as it stands: read at this commit,
> `internal/cli/codex_contract_link_test.go` asserts `codexTestExecImports(…codexClaudeRelPath, codexLinkAgentsDirective)`
> — that is `@AGENTS.md` in `CLAUDE.md`, a **different directive** — alongside the `AGENTS.md`
> direction of this criterion. So the criterion could be reported discharged on a green command
> while the work it names was undone.
>
> The repair follows `AC-IFU-010`'s pattern one criterion earlier: name a `$`-anchored symbol the
> implementation must create and assert its delimited `--- PASS:` line, so the criterion cannot go
> green until the assertion exists. Confirmed absent at this commit:
> `grep -rc 'func TestCodexContractLink_LocalImportMatrix' --include='*_test.go' .` → **0
> declarations**, so the criterion fails correctly against the unimplemented tree today.
>
> The citation also dropped `REQ-IFU-016`. This criterion measures import counts, which is not the
> noun that requirement constrained (clauses); the requirement itself is now retired, its substance
> absorbed as a clause of `REQ-IFU-018` (spec.md §C.4). Removing a citation the criterion does not
> earn is part of the same repair — a citation that survives without substance is what let
> `REQ-IFU-016` read as covered while nothing tested it.

> **[HARD] v0.3.1 repair — vacuous pattern, and one asserted half does not exist yet.** The
> pattern was `'^TestCodexContractLink'`, head-anchored only; no symbol
> `TestCodexContractLink` exists. Measured 2026-09-26 against `653e53572`:
> `go test ./internal/cli/ -run '^TestCodexContractLink' -v | grep -c -- '--- PASS:
> TestCodexContractLink'` → **11**, all from pre-existing sub-tests of
> `TestCodexContractLinkCreation`. The declared symbol is
> `TestCodexContractLinkCreation` (`internal/cli/codex_contract_link_test.go`).
>
> **The `CLAUDE.md` half is an assertion M2 adds to that test; it is not one the test carries
> today.** Read at `653e53572`, the test asserts `codexTestExecImports(…codexClaudeRelPath,
> codexLinkAgentsDirective) != 1` — that is `@AGENTS.md` in `CLAUDE.md`, a different directive —
> and `codexTestExecImports(…codexAgentsRelPath, codexTestLocalImportDirective) != 0`, which is
> the `AGENTS.md` half of this criterion. So the `AGENTS.md` half exists and the
> `CLAUDE.md`/`@AGENTS.local.md` half must be written. The previous wording ("the assertions are
> inverted to") read as though both were already in place, which is how a criterion that
> requires new work comes to pass before the work is done.

### D.6 Guard and learner surfaces

**AC-IFU-016** — Given the frozen set, When
`grep -n 'frozenInstructionFiles = ' internal/hook/pre_tool.go` runs, Then the line lists
all four of `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md`; and When

```
go test ./internal/hook/ -run '^TestFrozenInstructionFiles$' -v
```

runs, Then its output contains `--- PASS: TestFrozenInstructionFiles ` and does not contain
`no tests to run`, with one sub-case per entry in the set. (REQ-IFU-013)

> v0.3.1 repair (the D9 class sweep): the pattern was head-anchored only and the asserted line
> carried no delimiter. No `TestFrozenInstructionFiles*` symbol exists today, so this criterion
> was **not** vacuous in the measured sense — it fails correctly against the current tree — but
> it was vacuous by construction: once M2 creates the guard, any later
> `TestFrozenInstructionFiles_Something` would have satisfied both clauses while the guard
> itself was deleted. Anchored for the class, not for an observed failure.

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

runs, Then its output contains **both** `--- PASS: TestSurfaceForTier_Tier3 ` and
`--- PASS: TestPrepareTierDispatch_Tier3 `, does not contain `no tests to run`, and the
Tier-3 entry's `Path` is `AGENTS.local.md`. Both symbols are declared in
`internal/harness/curator/dispatch_test.go`. (REQ-IFU-014)

> v0.3.0 repair: the criterion named `TestDispatch`, which matches no test in that package.

**AC-IFU-018** — Given the neutrality guard, When

```
go test ./internal/template/agentemit/ -run '^TestNeutralityByInheritance$' -v
```

and `.github/workflows/template-neutrality-check.yaml` run over a template fixture containing
the literal `AGENTS.local.md`, Then the test output contains
`--- PASS: TestNeutralityByInheritance `, does not contain `no tests to run`, and the workflow
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
unresolved import is silently skipped with exit `0` and empty stderr. The probe is also
carried as a durable, re-runnable check: the run phase creates
`TestClaudeImportResolution_AgentsLocalSentinel` in `./internal/cli/` (0 declarations at
t1270, tree `20c73990d`), and
`go test ./internal/cli/ -run '^TestClaudeImportResolution_AgentsLocalSentinel$' -v` must
print `--- PASS: TestClaudeImportResolution_AgentsLocalSentinel ` and must not print
`no tests to run`; where the `claude` binary or its credentials are absent the test skips,
and a skipped run does not discharge this clause. (REQ-IFU-023)

**AC-IFU-020** — Given a fixture project with NO `AGENTS.local.md`, When `moai init` runs
and a headless `claude -p` session then starts, Then `moai init` exits `0`, the deployed
`CLAUDE.md` still contains the literal line `@AGENTS.local.md`, and the session exits `0`;
and When `moai update` then runs on the same fixture, still with no `AGENTS.local.md`, Then
`moai update` exits `0` and the re-deployed `CLAUDE.md` still contains the literal line
`@AGENTS.local.md` — recorded as a separate result from the `init` run. (REQ-IFU-004)

**AC-IFU-027** — Launcher invariance. Given the fixture of `AC-IFU-019` (an
`AGENTS.local.md` carrying a unique sentinel, reached only through `CLAUDE.md`'s
`@AGENTS.local.md` import), When the same sentinel request is issued under **each** of
`moai cc`, bare `claude`, and `moai glm` in headless mode, Then every one of the three
responses contains the sentinel; the same fixture also carries an `AGENTS.md` with a second,
distinct sentinel reached only through `CLAUDE.md`'s `@AGENTS.md` import, and every one of
the three responses contains that sentinel too — six results, because the two files travel
different imports and one establishes nothing about the other; and When
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
`--- PASS: TestAlwaysLoadedTokenBudget ` — with the trailing space, so that
`TestAlwaysLoadedTokenBudget_OverBudgetFails` (a real sibling in the same file) cannot satisfy
the clause on its own — declared in `internal/config/token_budget_guard_test.go`. (all)

> v0.3.1 repair (the D9 class sweep): the asserted line carried no delimiter, so deleting the
> real guard while keeping `_OverBudgetFails` would have left this clause passing. This criterion
> has no `-run` pattern to anchor — it reads a full-suite CI run — so the delimiter on the
> asserted line is the whole fix.

> The always-loaded clause is added at v0.3.0 and costs no criterion slot. Root `AGENTS.md`
> is itself inside the always-loaded surface that guard measures, and this SPEC moves that
> total from two directions — REQ-IFU-002 thins `CLAUDE.md`, REQ-IFU-018 may fold template-only
> sections into `AGENTS.md`. Both named ceilings (`AC-IFU-005` per-file 24,576, `AC-IFU-006`
> nested-sum 32,768) are different limits with different owners, so neither would catch it.
> Card t1175 is concurrently retuning that same budget; without this clause an
> always-loaded overrun would arrive during M3 — the milestone that rewrites the always-loaded surface — with no criterion and no owner. (v0.3.1: this note said M4, a milestone the B1/B2 carve deleted; plan.md §E now carries M1-M3 only.)

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

[HARD] **A row holds only if the citation matches AND the nouns match.** Naming the requirement id
in the criterion's body is necessary and is not sufficient. A row is valid only when the **observable
noun the criterion measures** is the noun the cited requirement **constrains** — or, where the
criterion measures a proxy, when the criterion's own body states why the proxy establishes the
requirement (`AC-IFU-003`'s declared-proxy note is the model — and note that it defers to no positive
assertion, which is why it names that absence rather than implying coverage).

Without this clause the derivation rule asks only whether the citation exists, so a table can be
citation-faithful and substance-empty — and the more faithfully the rule is followed, the more
legitimate that silence looks. The plan-audit of `4eb5405dc` found exactly that: `REQ-IFU-016`
constrained *clauses* while both citing criteria measured *import counts*, so the requirement was
tested by nothing while every mechanical check in this SPEC passed. The set-difference command below
cannot see this shape — the citation is present — which is why the clause is `[HARD]` rather than
delegated to a command.

**The enumeration method, recorded so a later reader can re-run it rather than re-derive it.** For
each requirement: write down the observable noun it constrains; write down the noun each citing
criterion actually measures; where they differ it is a proxy; for a proxy, test **both** implications
— *would a correct implementation satisfy the criterion*, and *would satisfying the criterion
establish the requirement* — and measure anything tree-dependent rather than reasoning about it.
This pass is a close obligation (§D.3) and is what the row-by-row re-derivation means.

[HARD] **Derivation is one-way, and the table is never the record.** The criterion bodies are
authoritative; this table is their projection. On any disagreement between a row and the cited
criterion's own citation line, the **body wins** — the repair is to re-derive the row from the
body (or, where the body's citation is itself wrong, to fix the citation and then re-derive).
Editing the row to agree with a body it misreports is prohibited, and so is reading this table
to learn what a criterion covers: read the criterion.

[HARD] **The verification command below does NOT check this table, and must not be cited as
though it did.** It compares two id **sets** — the ids cited anywhere in this file against the
ids declared in `spec.md` — so it is silent on pairing: a row naming two ids that both exist
elsewhere in the file passes it. That is exactly the v0.2.0 defect shape named above, and it is
why the derivation rule is `[HARD]` rather than delegated to the command. Re-deriving every row
from its criterion's citation line is a **manual step at close**, recorded as done; the command
is a necessary check on set membership and nothing more.

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
| REQ-IFU-017 | AC-IFU-009 |
| REQ-IFU-018 | AC-IFU-008 (section set), AC-IFU-003 (declared proxy on the neutrality clause) |
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

Close items below are what a guard that held once does not discharge.

- **The §D.2 table re-derived row by row, with the noun comparison recorded pair by pair.** For
  each of the 15 requirements, open **every** criterion citing it and write down, as a row in
  `progress.md` §E.2: the requirement id, the criterion id, the observable noun the requirement
  constrains, the noun the criterion measures, and the verdict — `match`, or `proxy` with the
  criterion's own stated reason. Re-derive each table row from the criterion's own citation line
  while doing it, and record that it was re-derived rather than diffed.

  [HARD] **The pair list is the artifact; "the pass was done" is not.** This item is discharged by a
  table in `progress.md` §E.2 and by nothing else. A reader must be able to tell a completed pass
  from an unstarted one without asking the agent that ran it, so the table carries:

  - **one row per (requirement, criterion) pair** — every pair, including the ones that matched;
  - a **declared pair total** stated above the table, and a row count that equals it, so
    completeness is checkable by arithmetic rather than taken on trust;
  - all **15** requirement ids present in the id column — a requirement cited by no criterion is a
    row reading `no citing criterion`, never an absent row, because an absent row and an unread
    requirement look identical;
  - for every `proxy` verdict, the criterion's own stated reason **quoted**, not summarized.

  A summary sentence with no table does not discharge this item, and neither does a table of
  mismatches only: "no mismatches over 23 pairs" and "no mismatches over the 4 pairs I got to" are
  the same sentence. This is the same failure shape as `4eb5405dc`'s recorded "both return nothing",
  written while the commands had never been run — a claim no reader could check.

  The obligation exists because hand enumeration of this class has already proved incomplete twice in
  this document — iter-1 enumerated five instances of the vacuity class, iter-2 found four more
  survivors after that repair, and iter-3 did not state whether it read every pair. That is why this
  pass stayed inside the SPEC's scope instead of becoming debt: an unmechanizable class of unknown
  size carried as debt is precisely what survived three audits. The §D.2 verification command checks
  neither pairing nor nouns; this step is the only thing that does.

Conditional item — **if `AC-IFU-021` confirms ancestor discovery for `AGENTS.local.md` in a
real linked worktree**, the finding and its verbatim evidence are handed to card **t1219**,
and the handoff is recorded in `progress.md` §E.2 naming what was transferred and where.
Without this item the obligation rested on prose in three files and on the run-phase agent
remembering it (plan-audit D7); design.md §A.5's confirmed branch requires the transfer, so
Done requires evidence the transfer happened.

### §D.3.1 Named debt and deferrals

This SPEC closes its plan phase at the three-iteration plan-audit cap by operator-approved scope
reduction plus debt (spec.md HISTORY v0.3.2). [HARD] Debt is named item by item; a debt list whose
members are not named is not a debt list, it is a disclaimer.

**Debt carried into the run phase (inside this SPEC's scope, not verified by a criterion):**

1. **No criterion positively asserts clause-level harness neutrality.** `REQ-IFU-018`'s neutrality
   clause is asserted by `AC-IFU-008` on its section-set half and by `AC-IFU-003` as a declared
   proxy on the neutrality half; nothing measures a clause. The work is owned — `plan.md` M3 names
   the one offending paragraph — and the verification is not. A run-phase agent closing M3 records
   the before/after text of that paragraph in `progress.md` §E.2 in place of a criterion's output.
2. **The anchoring rule is enforced by review until t1269's rule lands.** The head
   block of this file states this plainly rather than implying a live check. A criterion added or
   edited during the run phase is exactly where the class returns, and nothing mechanical will
   catch it in the interim.
3. **CI never executes `TestClaudeImportResolution_AgentsLocalSentinel`, so `AC-IFU-019`'s
   durable check has local-run evidence only.** The test needs a live `claude` with credentials
   and skips without one. CI runs the suite as `go test -json … ./...`
   (`.github/workflows/ci.yml:229` in the `test` job, `:311` in the race job), so the skip is
   recorded in the event stream, and `scripts/ci-census/test-census.sh` lists it as a
   `SKIPPED TEST` row — but the census is a reporter that exits `0` by contract
   (`test-census.sh:52-54`), and no step fails, warns, or annotates on a skip. A green CI run is
   therefore indistinguishable from a run in which this check did not execute, and a regression
   in Claude Code's import resolution would pass CI. The check stays live only while someone runs
   it locally with `claude` present; `plan.md` M3 owns that run and records its
   `--- PASS: TestClaudeImportResolution_AgentsLocalSentinel ` line in `progress.md` §E.2. This
   is accepted as debt, not closed by this SPEC.

**Deferred to card t1270 (out of this SPEC's plan scope; coordinates recorded here so t1270 can
pick them up without re-deriving them):**

4. **D21 (t1270) — `AC-IFU-009` asserts the removal of the superseded budget sentence and not the presence
   of its replacement.** `REQ-IFU-017`'s first half ("shall state that the budget is charged
   against project instruction files only") is satisfiable by deleting the old sentence and writing
   nothing. The repair is a third clause greping both mirrors for the replacement phrasing; the
   existing `grep -c 'truncat'` clause discharges the requirement's **second** half, not its first.
5. **D22 (t1270) — three criteria narrower than the requirement they cite.** `AC-IFU-020` asserts
   `moai init` where `REQ-IFU-004` names `init` **and** `update`, and `update` is the likelier
   regression surface because it re-deploys over an existing tree. `AC-IFU-027` asserts the
   `AGENTS.local.md` sentinel where `REQ-IFU-005` names both `AGENTS.md` **and**
   `AGENTS.local.md` reached; the two travel different imports, so one establishes nothing about
   the other, and separating them needs a second sentinel. `REQ-IFU-023` says the system "shall
   **carry** a mechanical check", which reads as a durable automated guard, while `AC-IFU-019` and
   `AC-IFU-021` are one-off headless `claude -p` measurements — a narrowing of the requirement is
   the likely resolution, since a `claude -p` probe needs a live model and cannot run in CI.

   **Closed by t1270 (spec.md HISTORY v0.3.3).** D21: `AC-IFU-009` gained a third clause greping
   both mirrors separately for `project instruction files only`, measured RED at tree `20c73990d`.
   D22: `AC-IFU-020` gained a `moai update` leg on the same fixture; `AC-IFU-027` gained an
   `AGENTS.md` sentinel reached through `@AGENTS.md` under all three launchers; and `AC-IFU-019`
   gained a durable Go test, `TestClaudeImportResolution_AgentsLocalSentinel`, rather than a
   narrowed `REQ-IFU-023`. That test needs a live `claude` and skips without one, so CI will show
   it skipped, not passed — the check survives the run phase as a re-runnable local gate, and a
   skipped run discharges nothing (named debt item 3). No criterion id was added; the count stays at 20.
6. **The `moai spec lint` anchoring rule itself — card t1269.** `VacuousAssertionRule` in
   `internal/spec/lint_vacuous_assertion.go`, registered in the rule slice in
   `internal/spec/lint.go`, with two-arm fixtures — one arm carrying a conformant criterion that
   must pass, one carrying each unanchored shape that must be caught. [HARD] **t1269 has not landed,
   so item 2 above holds until it does.** This entry is also the destination of record for the two
   removed `grep` pipelines and the Definition-of-Done item that required them: the removals are
   transfers to t1269, and a reader who finds those eleven lines gone should arrive here.

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
