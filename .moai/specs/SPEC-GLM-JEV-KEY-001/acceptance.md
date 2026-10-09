---
id: SPEC-GLM-JEV-KEY-001
title: "Acceptance criteria — glm --key flag and moai jev command"
created: 2026-10-09
---

# SPEC-GLM-JEV-KEY-001 — acceptance.md

## §A. Card Completion Criteria Map

Card t1613 defines three completion criteria; every one maps to at least one AC below.

| Card criterion | ACs |
|---|---|
| (a) both key commands' help output verified | AC-GJK-001, AC-GJK-002, AC-GJK-003 |
| (b) save result verified on disk | AC-GJK-004, AC-GJK-005 |
| (c) legacy `moai glm setup <key>` preserved (existing glm tests stay green) | AC-GJK-006, AC-GJK-007, AC-GJK-008 |
| Invariants (mode-0600 storage, masking, constant ownership, refusal shapes, input validation) | AC-GJK-009 … AC-GJK-015 |

## §B. Edge Cases

- `--key` with a value containing spaces / dotenv-special characters (`"`, `\`, `$`) → stored escaped by the owning package's `EscapeValue`, read back identically by `Load` (round-trip, package-level).
- `--key` with an embedded CR/LF (Go literal `"first\nsecond"`) → refused before write on BOTH new paths; the existing credential file is preserved byte-for-byte (AC-GJK-014/015). Legacy `setup` keeps today's first-line-only read-back behavior unchanged (known limitation, plan.md §B4).
- `--key=<value>` single-token spelling → accepted by the glm scan (§D.1 of plan.md).
- Credential of ≤4 characters → jev confirmation discloses NO part of it (REQ-JEVC-020 floor), glm `maskAPIKey` returns `****`.
- Pre-existing credential file at 0644 → tightened to 0600 on save (AC-GJK-012).
- Repeated saves → last writer wins, same file, no duplicate lines (package `Save` overwrites).

## §C. Quality Gate Criteria

- Scoped test families green: `go test ./internal/cli/ -run 'Test(GLM|Glm|Jev|Key|Root)'` — uppercase `TestGLM` covers the existing 116-case `TestGLM*` family, `TestKey` covers `TestKeyFormsShareStorageLastWriterWins`, and `TestGlm`/`TestJev`/`TestRoot` cover the new cases; `go test ./internal/glmcred/ ./internal/jevcred/` (CI owns the full suite). A `[no tests to run]` line in the selector's output is an empty sweep — a failure, never a pass (verification-completeness §1.1).
- `go vet ./internal/cli/` clean; gofmt clean on touched files.
- No new env-var literal outside owning packages (AC-GJK-013).
- TRUST 5: Secured = no full-key disclosure anywhere (AC-GJK-004/005 asserts output; REQ-GJK-010 binds all surfaces); Tested = every AC maps to a named RED-first test in plan.md §F.

## §D. AC Matrix

### §D.1 Acceptance criteria (Given-When-Then)

All scenarios run with the credential location redirected to `t.TempDir()` via the existing seams (plan.md §D.6); no test touches the real `$HOME`.

**AC-GJK-001** (High, card-a) — root help lists the new command.
Given the built CLI, When `moai --help` runs, Then the output contains a `jev` command entry.

**AC-GJK-002** (High, card-a) — glm help documents the flag.
Given the built CLI, When `moai glm --help` runs, Then the output contains `--key` with save-oriented wording.

**AC-GJK-003** (High, card-a) — jev help documents the flag.
Given the built CLI, When `moai jev --help` runs, Then the output contains `--key`.

**AC-GJK-004** (Critical, card-b, glm) — flag save lands on disk.
Given a redirected home, When the root command runs with args `["glm", "--key", "test-key-1234567890"]`, Then `<home>/.moai/.env.glm` exists at mode 0600 containing `GLM_API_KEY="test-key-1234567890"`, stdout contains `GLM API key stored (`, and the full key string appears in neither stdout nor stderr.

**AC-GJK-005** (Critical, card-b, jev) — flag save lands on disk.
Given a redirected home, When the root command runs with args `["jev", "--key", "tsk-cred-1234567890"]`, Then `<home>/.moai/.env.typesafe` exists at mode 0600 containing `TYPESAFE_API_KEY="tsk-cred-1234567890"`, the confirmation discloses at most the final four characters, and the full credential appears in neither stdout nor stderr.

**AC-GJK-006** (Critical, card-c) — legacy setup preserved.
Given the existing internal/cli glm test family unmodified, When that family runs, Then every test is green (no change to `runGLMSetup` or to the manual routing order).

**AC-GJK-007** (High) — both forms write the same storage, last writer wins.
Given a redirected home, When `glm --key A` then `glm setup B` run, Then the file contains `B`; and When `glm setup A` then `glm --key B` run, Then the file contains `B`.

**AC-GJK-008** (High) — routing precedence.
Given the `--key` scan exists, When `moai glm setup K2` runs, Then the value is stored through the setup path (scan never intercepts a routed subcommand).

**AC-GJK-009** (High) — extra-argument refusal.
When `moai glm --key K -f` runs (or `--key K` followed by any other token, e.g. `status`), Then the command exits non-zero with a usage error naming the conflict and no credential file is written.

**AC-GJK-010** (Medium) — bare jev invocation.
When `moai jev` runs without `--key`, Then help text prints, the process exits 0, and no credential file is created.

**AC-GJK-011** (High) — empty / missing value.
When `moai jev --key ""` runs, or `moai glm --key` runs with no following value, Then the command errors and nothing is stored.

**AC-GJK-012** (High) — mode tightening.
Given a pre-existing credential file at mode 0644, When a save runs through either command, Then the file mode is 0600 afterwards. (Asserted at package level by the existing glmcred/jevcred Save tests; CLI-level assertion optional.)

**AC-GJK-013** (Medium) — constant ownership, scoped to this card's ADDED lines.
Given the implementation commits exist (a diff with no added lines sweeps nothing and asserts nothing), When `git diff "$(git merge-base develop HEAD)..HEAD" -- internal/cli/ ':(exclude)**/*_test.go' | grep '^+' | grep "TYPESAFE_API_KEY\|GLM_API_KEY"` runs, Then it yields 0 rows — no added NON-TEST line spells the credential names as literals. Test files are excluded: `_test.go` is a sanctioned literal area (AGENTS.local.md hardcoding allowance) and the new AC-004/005 tests legitimately assert dotenv content strings. The 3 pre-existing rows in files this card never touches (`glm_tools.go:6` comment, `mcp_audit.go:30,32`) sit outside the change range and are out of scope: a whole-tree 0-row verdict is permanently red and proves nothing about this card. The merge-base form (not a literal pinned SHA) is the repo's measured rule for "what did THIS card change" (gitflow-lane-protocol §8).

**AC-GJK-014** (High) — jev newline value refused, stored credential preserved.
Given a redirected home holding a stored credential, When the root command runs with args `["jev", "--key", "first\nsecond"]` (embedded LF), Then the command exits non-zero with a validation error and `.env.typesafe` remains byte-for-byte identical to before the attempt.

**AC-GJK-015** (High) — glm `--key` newline value refused, stored key preserved.
Given a redirected home holding a stored key, When the root command runs with args `["glm", "--key", "first\nsecond"]`, Then the command exits non-zero with a validation error and `.env.glm` remains byte-for-byte identical to before the attempt.

### §D.2 Severity summary

Critical: AC-GJK-004, AC-GJK-005, AC-GJK-006 (the card's own completion criteria). High: AC-GJK-001/002/003, AC-GJK-007..009, AC-GJK-011, AC-GJK-012, AC-GJK-014, AC-GJK-015. Medium: AC-GJK-010, AC-GJK-013.

### §D.3 Traceability

REQ-GJK-001→AC-004/007/012 · REQ-GJK-002→AC-004 · REQ-GJK-003→AC-009 · REQ-GJK-004/005→AC-011 · REQ-GJK-006→AC-008/006 · REQ-GJK-007→AC-005/012 · REQ-GJK-008→AC-005 · REQ-GJK-009→AC-010 · REQ-GJK-010→AC-004/005 · REQ-GJK-011→AC-012 · REQ-GJK-012→AC-014/015 · card criteria (a)→AC-001..003, (b)→AC-004/005, (c)→AC-006..008.

### §D.4 Indirect verification

`moai doctor`'s Jev check and the web console read the same file through `jevcred.Load` — after AC-GJK-005, a doctor run on that home reports the credential configured. This is an optional manual probe; the AC evidence is the on-disk file.

### §D.5 Closure gates

- All §D.1 ACs green with the §C gate commands run in this run phase (verbatim output into progress.md §E.2).
- Card criteria (a)/(b)/(c) each explicitly signed off in the completion report.
- No `moai glm` launch-path behavior change (existing family green = evidence).

### §D.6 Forward-looking checks

- t1612 can build bare-`moai jev` guidance on the help seam (AC-GJK-010) without re-plumbing storage.
- The §D.4 doctor probe doubles as the regression tripwire if a later card ever splits jev storage away from `jevcred`.

### §D.7 Definition of Done

Every §D.1 AC observed green in this run phase + §C gates clean + the three card criteria verified + sync-phase close (`docs(SPEC-GLM-JEV-KEY-001): sync-phase artifacts` carrying the `completed` transition).
