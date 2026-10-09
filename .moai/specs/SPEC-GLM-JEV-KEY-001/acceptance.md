---
id: SPEC-GLM-JEV-KEY-001
title: "Acceptance criteria — glm --key flag and moai jev command"
created: 2026-10-09
updated: 2026-10-10
---

# SPEC-GLM-JEV-KEY-001 — acceptance.md

## §A. Card Completion Criteria Map

Card t1613 defines three completion criteria; every one maps to at least one AC below.

| Card criterion | ACs |
|---|---|
| (a) both key commands' help output verified | AC-GJK-001, AC-GJK-002, AC-GJK-003 |
| (b) save result verified on disk | AC-GJK-004, AC-GJK-005 |
| (c) legacy `moai glm setup <key>` preserved (existing glm tests stay green) | AC-GJK-006, AC-GJK-007, AC-GJK-008 |
| Invariants (mode-0600 storage, masking, constant ownership, refusal shapes, input validation) | AC-GJK-009 … AC-GJK-016 |

Card-level gates: AC-GJK-001, AC-GJK-002, and AC-GJK-003 verify card criterion (a) directly, and AC-GJK-013 verifies the credential-name constraint of plan.md §D.5 directly. Each of these four is a card-level gate that carries no REQ of its own (§D.3). Every other AC traces to at least one REQ.

## §B. Edge Cases

- `--key` with a value containing spaces / dotenv-special characters (`"`, `\`, `$`) → stored escaped by the owning package's `EscapeValue`, read back identically by `Load` (round-trip, package-level).
- `--key` with an embedded line break (Go literals `"first\nsecond"` LF, `"first\rsecond"` CR, `"first\r\nsecond"` CRLF) → refused before write on BOTH new paths; the existing credential file is preserved byte-for-byte (AC-GJK-014, AC-GJK-015). Legacy `setup` keeps today's first-line-only read-back behavior unchanged (known limitation, plan.md §B4).
- `--key=<value>` single-token spelling → accepted by the glm scan, dash-leading value included (`--key=-f` is stored; plan.md §D.1); jev refuses a dash-leading value in this spelling (`--key=-f`, REQ-GJK-013, AC-GJK-011); a normal jev `--key=<value>` is stored (AC-GJK-005).
- A separated `--key` value whose trimmed form begins with `-` (`--key -f`, `--key --help`) → refused with a usage error and nothing stored (AC-GJK-011). `moai glm --key --help` is answered by the help precedence instead (REQ-GJK-003): it prints help and stores nothing.
- Credential of ≤4 characters → jev confirmation discloses NO part of it (REQ-JEVC-020 floor), glm `maskAPIKey` returns `****`.
- Pre-existing credential file at 0644 → tightened to 0600 on save (AC-GJK-012).
- Repeated saves → last writer wins, same file, no duplicate lines (package `Save` overwrites).

## §C. Quality Gate Criteria

- Preservation verdict (criterion c, front line — the card modifies `glm.go`, so the PACKAGE scope is the preservation evidence): `go test ./internal/cli/` green, including the pre-existing glm-family tests the focused selector does not sweep (measured misses by name shape: glm_test 11, glm_new_test 35, glm_compat 4, glm_team 6, glm_persist_gate 5, mcp_glm_parse 2 — 63 tests). Focused new-family run (secondary, honestly scoped): `go test ./internal/cli/ -run 'Test(GLM|Glm|Jev|Key|Root)'` sweeps 145 name-matching functions, NOT the whole glm family; `go test ./internal/glmcred/ ./internal/jevcred/` (CI owns the full suite). A `[no tests to run]` line in the selector's output is an empty sweep — a failure, never a pass (verification-completeness §1.1).
- `go vet ./internal/cli/` clean; gofmt clean on touched files.
- No new env-var literal outside owning packages (AC-GJK-013).
- TRUST 5: Secured = no full-key disclosure anywhere (AC-GJK-004, AC-GJK-005, AC-GJK-009, AC-GJK-011, AC-GJK-014, AC-GJK-015 assert output; REQ-GJK-010 binds all surfaces); Tested = every AC maps to a named RED-first test in plan.md §F.

## §D. AC Matrix

### §D.1 Acceptance criteria (Given-When-Then)

All scenarios run with the credential location redirected to `t.TempDir()` via the existing seams (plan.md §D.6); no test touches the real `$HOME`.

**AC-GJK-001** (High, card-a; card-level gate, §A) — root help lists the new command.
Given the built CLI, When `moai --help` runs, Then the output contains a `jev` command entry.

**AC-GJK-002** (High, card-a; card-level gate, §A) — glm help documents the flag.
Given the built CLI, When `moai glm --help` runs, Then the output contains the exact string `Store the GLM API key and exit (same storage as 'moai glm setup <key>')`, which is the usage string registered for the `--key` flag at internal/cli/glm.go:154.

**AC-GJK-003** (High, card-a; card-level gate, §A) — jev help documents the flag.
Given the built CLI, When `moai jev --help` runs, Then the output contains `--key`.

**AC-GJK-004** (Critical, card-b, glm) — flag save lands on disk.
Given a redirected home, When the root command runs with args `["glm", "--key", "test-key-1234567890"]`, Then `<home>/.moai/.env.glm` exists at mode 0600 containing `GLM_API_KEY="test-key-1234567890"`, stdout contains `GLM API key stored (`, the full key string appears in neither stdout nor stderr, and the launch path is never reached (the test stubs the launch step and records zero invocations).

**AC-GJK-005** (Critical, card-b, jev) — flag save lands on disk.
Given a redirected home, When the root command runs with args `["jev", "--key", "tsk-cred-1234567890"]`, Then `<home>/.moai/.env.typesafe` exists at mode 0600 containing `TYPESAFE_API_KEY="tsk-cred-1234567890"`, the confirmation discloses at most the final four characters, and the full credential appears in neither stdout nor stderr. The normal `--key=<value>` spelling stores the same way: when the root command runs with args `["jev", "--key=tsk-cred-1234567890"]`, Then `<home>/.moai/.env.typesafe` holds `TYPESAFE_API_KEY="tsk-cred-1234567890"` at mode 0600, and the full credential appears in neither stdout nor stderr.
Short-credential scenario: Given a redirected home, When `moai jev --key Q` runs, `moai jev --key Qz9K` runs, and `moai jev --key Qz9Kx` runs, Then the 1- and 4-character confirmations disclose no part of their credential, the 5-character confirmation discloses only `z9Kx`, and in the 4- and 5-character runs the full credential appears in neither stdout nor stderr.

**AC-GJK-006** (Critical, card-c) — legacy setup preserved.
Given the existing internal/cli glm test family unmodified, When the package-scope run `go test ./internal/cli/` executes (the front-line evidence: the card modifies `glm.go`, and this sweep includes the pre-existing glm-family tests whose names the focused selector does not match), Then it is green with no change to `runGLMSetup` or to the manual routing order.

**AC-GJK-007** (High) — both forms write the same storage, last writer wins.
Given a redirected home, When `glm --key A` then `glm setup B` run, Then the file contains `B`; and When `glm setup A` then `glm --key B` run, Then the file contains `B`.

**AC-GJK-008** (High) — routing precedence.
Given the `--key` scan exists, When `moai glm setup K2` runs, Then the value is stored through the setup path (scan never intercepts a routed subcommand).

**AC-GJK-009** (High) — extra-argument refusal and help precedence.
Given a redirected home with no credential file, When `moai glm --key test-key-1234567890 -f` runs (or `--key test-key-1234567890` followed by any other token, e.g. `status`), Then the command exits non-zero with a usage error naming the conflict, the full key appears in neither stdout nor stderr, and no credential file is written. Also, When `moai glm --key test-key-1234567890 --help` runs, Then help text is printed with exit 0 (not a usage error), no credential file is written, the full key appears in neither stdout nor stderr, and the launch path records zero invocations (REQ-GJK-003 help precedence).

**AC-GJK-010** (Medium) — bare jev invocation.
When `moai jev` runs without `--key`, Then help text prints, the process exits 0, and no credential file is created.

**AC-GJK-011** (High) — empty, missing, or flag-shaped value.
Given a redirected home whose target credential file already holds `stored-1234567890`: When `moai jev --key ""` runs, Then the command fails with jev's empty-credential error, `.env.typesafe` stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr; When `moai glm --key ""` runs (empty after trim), Then the command fails with the setup path's `empty API key` error (REQ-GJK-005), `.env.glm` stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr; When `moai glm --key` runs with no following value, Then the command fails with the missing-value usage error (REQ-GJK-004), `.env.glm` stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr. Flag-shaped value (REQ-GJK-004, REQ-GJK-013): When `moai glm --key -f` runs, and When `moai glm --key -` runs, and When `moai glm --key " -f"` runs, Then each command fails with a usage error naming `--key` (the trimmed value is tested), `.env.glm` stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr; When `moai jev --key -f` runs, When `moai jev --key --help` runs, When `moai jev --key -` runs, and When `moai jev --key " -f"` runs, Then each command fails with a usage error refusing a dash-leading credential, the literal `-`, `-f`, or `--help` is not stored, `.env.typesafe` stays byte-for-byte unchanged, and the stored value appears in neither stdout nor stderr. `=` spelling (REQ-GJK-004, REQ-GJK-013): When `moai glm --key=-f` runs, Then stdout contains `GLM API key stored (` and `.env.glm` holds `GLM_API_KEY="-f"`; When `moai jev --key=-f` runs, Then the command fails with a usage error and `.env.typesafe` stays byte-for-byte unchanged.

**AC-GJK-012** (High) — mode tightening.
Given a pre-existing credential file at mode 0644, When a save runs through either command, Then the file mode is 0600 afterwards. Binary evidence: `TestSave_NarrowsExisting0644to0600` in `internal/glmcred/glmcred_test.go` and `TestSave_NarrowsExisting0644to0600` in `internal/jevcred/jevcred_test.go`; the AC is green when both tests pass.

**AC-GJK-013** (Medium; card-level gate, §A) — constant ownership, scoped to this card's ADDED lines.
Given the implementation commits exist — established by separate plain commands, never assumed: `git merge-base origin/main HEAD; echo "merge_base_exit=$?"` prints the base on its own line followed by `merge_base_exit=0`, AND `git diff --stat origin/main...HEAD -- internal/cli/` is non-empty (a bare pipeline cannot distinguish a zero-match exit from a git-failure exit, so the premise is itself executable; the three-dot range is the same diff as merge-base..HEAD and avoids the `$(git merge-base …)` range that the worktree-session guard refuses, `worktree-integration-ops.md` authoring rule 3), When `git diff origin/main...HEAD -- internal/cli/ ':(exclude)**/*_test.go' | grep '^+' | grep "TYPESAFE_API_KEY\|GLM_API_KEY"` runs, Then it yields 0 rows — no added NON-TEST line spells the credential names as literals. Test files are excluded: `_test.go` is a sanctioned literal area (AGENTS.local.md hardcoding allowance) and the new AC-GJK-004 and AC-GJK-005 tests legitimately assert dotenv content strings. The 3 pre-existing rows in files this card never touches (`glm_tools.go:6` comment, `mcp_audit.go:30,32`) sit outside the change range and are out of scope: a whole-tree 0-row verdict is permanently red and proves nothing about this card. The merge-base form (not a literal pinned SHA) is the repo's measured rule for "what did THIS card change" (gitflow-lane-protocol §8). The card base is main (AGENTS.md §3); the local gitflow-lane-protocol §1 and §8 still name develop, and that conflict is an operator decision left open.

**AC-GJK-014** (High) — jev newline value refused, stored credential preserved.
Given a redirected home holding a stored credential in `.env.typesafe`, When the root command runs with args `["jev", "--key", V]` for each embedded line-break value V in turn (Go literals `"first\nsecond"` LF, `"first\rsecond"` CR, and `"first\r\nsecond"` CRLF), Then for every V the command exits non-zero with a validation error, `.env.typesafe` remains byte-for-byte identical to before the attempt, and the full value V (line break included) appears in neither stdout nor stderr.

**AC-GJK-015** (High) — glm `--key` newline value refused, stored key preserved.
Given a redirected home holding a stored key in `.env.glm`, When the root command runs with args `["glm", "--key", V]` for each embedded line-break value V in turn (Go literals `"first\nsecond"` LF, `"first\rsecond"` CR, and `"first\r\nsecond"` CRLF), Then for every V the command exits non-zero with a validation error, `.env.glm` remains byte-for-byte identical to before the attempt, and the full value V (line break included) appears in neither stdout nor stderr.

**AC-GJK-016** (High) — post-`--` tokens are child passthrough, never scanned.
Given a redirected home, the env test seam `MOAI_TEST_GLM_KEY` set to a test key K, and the redirects in plan.md §D.6 (`findProjectRootFn` redirected to an empty `t.TempDir()`, `launchEffortPrefsFn` returning no effort, so that the launcher appends no `--settings` or `--effort` pair), with `launchClaudeFunc` stubbed to record its arguments, When `moai glm -- --key K` runs, Then no credential file is created and the arguments recorded by the launch seam `launchClaudeFunc` equal `["--", "--key", K]`, the post-`--` tokens unmodified. The separator is part of the expected list because the launcher keeps `--` (launcher.go:879-882). A whole-args scan would create a credential file or record different arguments, so the AC fails that mutant.

### §D.2 Severity summary

Critical: AC-GJK-004, AC-GJK-005, AC-GJK-006 (the card's own completion criteria). High: AC-GJK-001, AC-GJK-002, AC-GJK-003, AC-GJK-007, AC-GJK-008, AC-GJK-009, AC-GJK-011, AC-GJK-012, AC-GJK-014, AC-GJK-015, AC-GJK-016. Medium: AC-GJK-010, AC-GJK-013.

### §D.3 Traceability

Card-level gates (no REQ; see §A): AC-GJK-001, AC-GJK-002, AC-GJK-003, AC-GJK-013.

REQ-GJK-001→AC-GJK-004, AC-GJK-007, AC-GJK-012 · REQ-GJK-002→AC-GJK-004 · REQ-GJK-003→AC-GJK-009, AC-GJK-016 · REQ-GJK-004→AC-GJK-011 (dash-leading separated value; `=` spelling) · REQ-GJK-005→AC-GJK-011 · REQ-GJK-006→AC-GJK-006, AC-GJK-008 · REQ-GJK-007→AC-GJK-005, AC-GJK-012 · REQ-GJK-008→AC-GJK-005 · REQ-GJK-009→AC-GJK-010 · REQ-GJK-010→AC-GJK-004, AC-GJK-005, AC-GJK-009, AC-GJK-011, AC-GJK-014, AC-GJK-015 · REQ-GJK-011→AC-GJK-012 · REQ-GJK-012→AC-GJK-014, AC-GJK-015 · REQ-GJK-013→AC-GJK-011 (dash-leading jev refusal, separated and `=` spellings) · card criteria (a)→AC-GJK-001, AC-GJK-002, AC-GJK-003; (b)→AC-GJK-004, AC-GJK-005; (c)→AC-GJK-006, AC-GJK-007, AC-GJK-008.

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
